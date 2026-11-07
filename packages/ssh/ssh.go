package ssh

import (
	"bufio"
	"bytes"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/Gomobra/PixTui/packages/core"
)

// Identity describes who is connected to a session and how they authenticated.
type Identity struct {
	// Method is one of "none", "password", or "publickey".
	Method string
	// Username is the client-supplied login name.
	Username string
	// Fingerprint is the SHA256 fingerprint of the authenticated public key (publickey only).
	Fingerprint string
	// PublicKey is the authenticated public key (publickey only).
	PublicKey ssh.PublicKey
}

// SessionHandler is invoked once per accepted shell/exec session.
type SessionHandler func(session *Session) error

// Middleware wraps a session handler. Call next() to hand control to the next
// link in the chain (ultimately the handler). Code executed after next()
// returns runs as teardown for that link.
type Middleware func(session *Session, next func() error) error

// Server is an SSH server that serves PixTui applications over SSH.
type Server struct {
	addr        string
	handler     SessionHandler
	middlewares []Middleware
	sessions    map[string]*Session
	running     bool
	mu          sync.Mutex
	listener    net.Listener
	config      *ssh.ServerConfig

	// Authentication wiring.
	noAuth                  bool
	passwordAuth            func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error)
	publicKeyAuth           func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error)
	keyboardInteractiveAuth func(conn ssh.ConnMetadata, challenge ssh.KeyboardInteractiveChallenge) (*ssh.Permissions, error)

	// Authorized keys for publickey authentication.
	authorizedKeys map[string]ssh.PublicKey

	// Resource limits.
	maxSessionsPerConn int
	maxSessionsGlobal  int
	activeSessions     int

	// Optional per-session timeouts.
	idleTimeout time.Duration
	maxTimeout  time.Duration

	// onError is the single error sink for runtime errors.
	onError func(error)
}

// NewServer creates a new SSH server bound to addr. The handler may be nil.
// Use Use() to add middleware and Listen() to start accepting connections.
func NewServer(addr string, handler SessionHandler) *Server {
	config := &ssh.ServerConfig{
		ServerVersion: "SSH-2.0-pixtui",
	}
	return &Server{
		addr:               addr,
		handler:            handler,
		sessions:           make(map[string]*Session),
		config:             config,
		maxSessionsPerConn: 1,
		maxSessionsGlobal:  100,
	}
}

// AddHostKey adds an SSH host key to the server. Nil keys are ignored.
func (s *Server) AddHostKey(key ssh.Signer) {
	if key == nil {
		return
	}
	s.config.AddHostKey(key)
}

// AddHostKeyPEM parses a PEM-encoded private key (RSA, ECDSA, or ED25519) and
// adds it as a host key.
func (s *Server) AddHostKeyPEM(pemBlock []byte) error {
	key, err := ssh.ParsePrivateKey(pemBlock)
	if err != nil {
		return fmt.Errorf("ssh: failed to parse host key PEM: %w", err)
	}
	s.AddHostKey(key)
	return nil
}

// AddHostKeyFile reads a private key from path and adds it as a host key.
func (s *Server) AddHostKeyFile(path string) error {
	key, err := LoadHostKey(path)
	if err != nil {
		return err
	}
	s.AddHostKey(key)
	return nil
}

// GenerateHostKey generates a new host key of the given algorithm
// ("rsa", "ecdsa", or "ed25519") and adds it to the server.
func (s *Server) GenerateHostKey(algorithm string) error {
	key, err := generateHostKey(algorithm)
	if err != nil {
		return err
	}
	s.AddHostKey(key)
	return nil
}

// SetPasswordAuth sets the password authentication callback.
func (s *Server) SetPasswordAuth(cb func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error)) {
	s.passwordAuth = cb
	s.config.PasswordCallback = func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
		perm, err := cb(conn, password)
		if err != nil {
			return nil, err
		}
		return annotatePermissions(perm, "password", conn.User(), "", nil), nil
	}
}

// SetPublicKeyAuth sets the public key authentication callback.
func (s *Server) SetPublicKeyAuth(cb func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error)) {
	s.publicKeyAuth = cb
	s.config.PublicKeyCallback = func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
		perm, err := cb(conn, key)
		if err != nil {
			return nil, err
		}
		return annotatePermissions(perm, "publickey", conn.User(), sshFingerprint(key), key.Marshal()), nil
	}
}

// SetKeyboardInteractiveAuth sets the keyboard-interactive authentication
// callback (RFC 4256). The challenge callback is used to pose questions to
// the client and collect one or more answers. The returned Permissions are
// annotated with the "keyboard-interactive" method so the session identity
// can be recovered after the handshake.
//
// Note: x/crypto/ssh's KeyboardInteractiveChallenge is a function used to
// prompt the client, not a plain challenge string.
func (s *Server) SetKeyboardInteractiveAuth(cb func(conn ssh.ConnMetadata, challenge ssh.KeyboardInteractiveChallenge) (*ssh.Permissions, error)) {
	s.keyboardInteractiveAuth = cb
	s.config.KeyboardInteractiveCallback = func(conn ssh.ConnMetadata, client ssh.KeyboardInteractiveChallenge) (*ssh.Permissions, error) {
		perm, err := cb(conn, client)
		if err != nil {
			return nil, err
		}
		return annotatePermissions(perm, "keyboard-interactive", conn.User(), "", nil), nil
	}
}

// SetAuthorizedKeys configures the set of public keys accepted for
// publickey authentication. Keys are matched by their wire-format marshaled
// bytes, and any previously configured authorized keys are replaced. This
// installs authorizedKeysAuth as the server's PublicKeyCallback.
//
// Note: ssh.ServerConfig exposes a single PublicKeyCallback, so calling
// SetPublicKeyAuth after this method (or vice versa) replaces whichever
// callback was installed last.
func (s *Server) SetAuthorizedKeys(keys []ssh.PublicKey) {
	if s.authorizedKeys == nil {
		s.authorizedKeys = make(map[string]ssh.PublicKey, len(keys))
	} else {
		clear(s.authorizedKeys)
	}
	for _, key := range keys {
		if key == nil {
			continue
		}
		s.authorizedKeys[string(key.Marshal())] = key
	}
	s.config.PublicKeyCallback = s.authorizedKeysAuth
}

// LoadAuthorizedKeys reads an OpenSSH authorized_keys file from path and
// configures its public keys for publickey authentication. Each non-empty,
// non-comment line is parsed with ssh.ParseAuthorizedKey; lines that fail to
// parse are skipped. See SetAuthorizedKeys for matching semantics.
func (s *Server) LoadAuthorizedKeys(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("ssh: failed to read authorized_keys file %q: %w", path, err)
	}
	var keys []ssh.PublicKey
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(line))
		if err != nil {
			continue
		}
		keys = append(keys, key)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("ssh: failed to read authorized_keys file %q: %w", path, err)
	}
	s.SetAuthorizedKeys(keys)
	return nil
}

// authorizedKeysAuth is the PublicKeyCallback used when authorized keys have
// been configured via SetAuthorizedKeys or LoadAuthorizedKeys. It succeeds
// only when the presented public key matches a configured authorized key.
func (s *Server) authorizedKeysAuth(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
	if key == nil {
		return nil, fmt.Errorf("ssh: empty public key")
	}
	if s.authorizedKeys == nil {
		return nil, fmt.Errorf("ssh: no authorized keys configured")
	}
	if _, ok := s.authorizedKeys[string(key.Marshal())]; !ok {
		return nil, fmt.Errorf("ssh: public key not authorized for user %q", conn.User())
	}
	return annotatePermissions(nil, "publickey", conn.User(), sshFingerprint(key), key.Marshal()), nil
}

// SetNoAuth allows unauthenticated access.
func (s *Server) SetNoAuth() {
	s.noAuth = true
	s.config.NoClientAuth = true
}

// SetIdleTimeout disconnects a session after the given duration with no client
// input. Zero disables the idle timeout.
func (s *Server) SetIdleTimeout(d time.Duration) {
	s.idleTimeout = d
}

// SetMaxTimeout disconnects a session after the given absolute lifetime,
// regardless of activity. Zero disables the max timeout.
func (s *Server) SetMaxTimeout(d time.Duration) {
	s.maxTimeout = d
}

// SetSessionLimits configures the maximum sessions per connection and globally.
func (s *Server) SetSessionLimits(perConn, global int) {
	s.maxSessionsPerConn = perConn
	s.maxSessionsGlobal = global
}

// SetErrorHandler sets the runtime error sink. Errors from handlers, middleware,
// and the transport are reported here.
func (s *Server) SetErrorHandler(cb func(error)) {
	s.onError = cb
}

// Use appends a middleware link to the chain. Middleware runs in registration
// order: the first registered link is the outermost.
func (s *Server) Use(mw Middleware) *Server {
	s.middlewares = append(s.middlewares, mw)
	return s
}

// Addr returns the address the server is listening on.
func (s *Server) Addr() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.addr
}

// Listen binds and starts accepting SSH connections.
func (s *Server) Listen() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("ssh: server already running")
	}
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		s.mu.Unlock()
		return fmt.Errorf("ssh: failed to listen: %w", err)
	}
	s.listener = listener
	s.addr = listener.Addr().String()
	s.running = true
	s.mu.Unlock()

	go s.acceptLoop()
	return nil
}

// Close stops accepting connections and closes every live session.
func (s *Server) Close() error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = false
	listener := s.listener
	sessions := s.sessions
	s.sessions = make(map[string]*Session)
	s.mu.Unlock()

	if listener != nil {
		listener.Close()
	}
	for _, session := range sessions {
		session.Close()
	}
	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			s.mu.Lock()
			running := s.running
			s.mu.Unlock()
			if !running {
				return
			}
			continue
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	// Perform the SSH handshake (key exchange + authentication).
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, s.config)
	if err != nil {
		conn.Close()
		return
	}
	// Global requests are not used by this server.
	go ssh.DiscardRequests(reqs)

	identity := identityFromConn(sshConn, s.noAuth)
	remoteAddr := sshConn.RemoteAddr()

	var connSessions int32

	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		if s.maxSessionsPerConn > 0 &&
			atomic.LoadInt32(&connSessions) >= int32(s.maxSessionsPerConn) {
			newChannel.Reject(ssh.RejectionReason(2), "per-connection session limit reached")
			continue
		}

		if s.maxSessionsGlobal > 0 {
			s.mu.Lock()
			if s.activeSessions >= s.maxSessionsGlobal {
				s.mu.Unlock()
				newChannel.Reject(ssh.RejectionReason(2), "global session limit reached")
				continue
			}
			s.activeSessions++
			s.mu.Unlock()
		}
		atomic.AddInt32(&connSessions, 1)

		channel, requests, err := newChannel.Accept()
		if err != nil {
			atomic.AddInt32(&connSessions, -1)
			s.mu.Lock()
			s.activeSessions--
			s.mu.Unlock()
			continue
		}

		session := newSession(sshConn, channel, identity, remoteAddr)
		session.armTimeouts(s.idleTimeout, s.maxTimeout)

		release := func() {
			atomic.AddInt32(&connSessions, -1)
			s.mu.Lock()
			s.activeSessions--
			s.mu.Unlock()
		}
		session.OnClose(release)

		s.mu.Lock()
		s.sessions[session.id] = session
		s.mu.Unlock()

		go s.handleSession(session, channel, requests)
	}
}

func (s *Server) handleSession(session *Session, channel ssh.Channel, requests <-chan *ssh.Request) {
	defer func() {
		session.Close()
		s.mu.Lock()
		delete(s.sessions, session.id)
		s.activeSessions--
		s.mu.Unlock()
	}()

	var shellStarted bool
	for req := range requests {
		switch req.Type {
		case "pty-req":
			term, cols, rows, ok := parsePtyReq(req.Payload)
			if ok {
				session.setPty(term, cols, rows)
			}
			req.Reply(true, nil)
		case "window-change":
			cols, rows, ok := parseWindowChange(req.Payload)
			if ok {
				session.resize(cols, rows)
			}
			req.Reply(true, nil)
		case "env":
			key, val, ok := parseEnv(req.Payload)
			if ok {
				session.setEnv(key, val)
			}
			req.Reply(true, nil)
		case "shell":
			if !shellStarted {
				shellStarted = true
				req.Reply(true, nil)
				go s.runHandler(session)
			} else {
				req.Reply(true, nil)
			}
		case "exec":
			if !shellStarted {
				shellStarted = true
				req.Reply(true, nil)
				go s.runHandler(session)
			} else {
				req.Reply(true, nil)
			}
		case "subsystem":
			// Subsystems (e.g. SFTP) are not supported.
			req.Reply(false, nil)
		default:
			req.Reply(false, nil)
		}
	}
}

func (s *Server) runHandler(session *Session) {
	err := s.runMiddleware(session)
	if err != nil && s.onError != nil {
		s.onError(err)
	}
	session.End()
}

func (s *Server) runMiddleware(session *Session) error {
	if s.handler == nil {
		return nil
	}
	chain := func(sess *Session) error { return s.handler(sess) }
	for i := len(s.middlewares) - 1; i >= 0; i-- {
		mw := s.middlewares[i]
		prev := chain
		chain = func(sess *Session) error {
			return mw(sess, func() error { return prev(sess) })
		}
	}
	return chain(session)
}

// removeSession drops a session from the live session registry.
func (s *Server) removeSession(session *Session) {
	s.mu.Lock()
	delete(s.sessions, session.id)
	s.mu.Unlock()
}

// Session represents an SSH session tied to an PixTui application.
type Session struct {
	id         string
	conn       *ssh.ServerConn
	channel    ssh.Channel
	server     *Server
	remoteAddr string
	mu         sync.Mutex
	closed     bool

	term   string
	cols   int
	rows   int
	hasPty bool

	identity Identity
	env      map[string]string
	context  map[string]any

	idleTimer   *time.Timer
	maxTimer    *time.Timer
	idleTimeout time.Duration
	maxTimeout  time.Duration

	// Listeners keyed by an incrementing ID so callers can unregister by
	// returned closure (function values are not comparable in Go).
	resizeListeners map[int]func(int, int)
	resizeNextID    int
	closeListeners  map[int]func()
	closeNextID     int
}

// newSession creates a session bound to an authenticated SSH connection.
func newSession(sshConn *ssh.ServerConn, channel ssh.Channel, identity Identity, remoteAddr net.Addr) *Session {
	return &Session{
		id:              fmt.Sprintf("%x", sshConn.SessionID()),
		conn:            sshConn,
		channel:         channel,
		remoteAddr:      remoteAddr.String(),
		identity:        identity,
		env:             make(map[string]string),
		context:         make(map[string]any),
		term:            "xterm-256color",
		cols:            80,
		rows:            24,
		resizeListeners: make(map[int]func(int, int)),
		closeListeners:  make(map[int]func()),
	}
}

// ID returns the session ID.
func (s *Session) ID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.id
}

// Write sends data to the client over the SSH channel.
func (s *Session) Write(data []byte) (int, error) {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return 0, io.ErrClosedPipe
	}
	ch := s.channel
	s.mu.Unlock()
	n, err := ch.Write(data)
	s.resetIdleLocked()
	return n, err
}

// Read reads client input from the SSH channel.
func (s *Session) Read(data []byte) (int, error) {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return 0, io.ErrClosedPipe
	}
	ch := s.channel
	s.mu.Unlock()
	n, err := ch.Read(data)
	s.resetIdleLocked()
	return n, err
}

// Close closes the session channel and fires close listeners. Idempotent.
func (s *Session) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	ch := s.channel
	if s.idleTimer != nil {
		s.idleTimer.Stop()
	}
	if s.maxTimer != nil {
		s.maxTimer.Stop()
	}
	listeners := make([]func(), 0, len(s.closeListeners))
	for _, cb := range s.closeListeners {
		listeners = append(listeners, cb)
	}
	s.closeListeners = make(map[int]func())
	s.mu.Unlock()

	if ch != nil {
		ch.Close()
	}
	for _, cb := range listeners {
		cb()
	}
	return nil
}

// End force-closes this session.
func (s *Session) End() error {
	return s.Close()
}

// GetWindowSize returns the current terminal window size (columns, rows).
func (s *Session) GetWindowSize() (int, int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cols, s.rows
}

// SetWindowSize updates the terminal window size and notifies resize listeners.
func (s *Session) SetWindowSize(cols, rows int) {
	s.resize(cols, rows)
}

// GetEnv returns an environment variable value supplied by the client.
func (s *Session) GetEnv(key string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.env[key]
}

// RemoteAddress returns the client socket address.
func (s *Session) RemoteAddress() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.remoteAddr
}

// Identity returns who is connected and how they authenticated.
func (s *Session) Identity() Identity {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.identity
}

// Context returns the per-session context bag shared with middleware.
func (s *Session) Context() map[string]any {
	return s.context
}

// GetContext returns a context value.
func (s *Session) GetContext(key string) (any, bool) {
	s.mu.Lock()
	v, ok := s.context[key]
	s.mu.Unlock()
	return v, ok
}

// SetContext sets a context value.
func (s *Session) SetContext(key string, value any) {
	s.mu.Lock()
	s.context[key] = value
	s.mu.Unlock()
}

// OnResize registers a callback invoked after the client resizes the PTY.
// The returned function unregisters the callback.
func (s *Session) OnResize(cb func(cols, rows int)) func() {
	s.mu.Lock()
	id := s.resizeNextID
	s.resizeNextID++
	s.resizeListeners[id] = cb
	s.mu.Unlock()
	return func() {
		s.mu.Lock()
		delete(s.resizeListeners, id)
		s.mu.Unlock()
	}
}

// OnClose registers a callback invoked once when the session closes. Late
// subscribers are invoked immediately if the session is already closed.
func (s *Session) OnClose(cb func()) func() {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		cb()
		return func() {}
	}
	id := s.closeNextID
	s.closeNextID++
	s.closeListeners[id] = cb
	s.mu.Unlock()
	return func() {
		s.mu.Lock()
		delete(s.closeListeners, id)
		s.mu.Unlock()
	}
}

// NewRenderer creates a CliRenderer sized to the client's PTY. Resize events
// from the renderer are forwarded back to the session.
func (s *Session) NewRenderer() *core.CliRenderer {
	s.mu.Lock()
	cols, rows := s.cols, s.rows
	s.mu.Unlock()

	r := core.NewCliRenderer(cols, rows)
	r.SetResizeHandler(func(w, h int) {
		s.resize(w, h)
	})
	return r
}

// WriteFrame captures the renderer's current buffer and writes it to the
// client as an ANSI screen redraw.
func (s *Session) WriteFrame(r *core.CliRenderer) error {
	access := r.Capture()
	width, height := access.Width, access.Height
	if width <= 0 || height <= 0 {
		return nil
	}

	var buf bytes.Buffer
	buf.WriteString("\x1b[H") // cursor home

	lastFG, lastBG := uint16(0xff), uint16(0xff)
	lastAttrs := uint32(0xffffffff)

	for y := 0; y < height; y++ {
		base := y * width
		buf.WriteString("\x1b[0m") // reset attributes at row start
		lastFG, lastBG, lastAttrs = 0xff, 0xff, 0xffffffff

		for x := 0; x < width; x++ {
			i := base + x
			ch := access.Char[i]
			fg := access.Fg[i]
			bg := access.Bg[i]
			attrs := access.Attributes[i]

			if fg != lastFG || bg != lastBG || attrs != lastAttrs {
				code := sgrCode(fg, bg, attrs)
				if code != "" {
					buf.WriteString("\x1b[")
					buf.WriteString(code)
					buf.WriteString("m")
				}
				lastFG, lastBG, lastAttrs = fg, bg, attrs
			}

			if ch == 0 {
				ch = ' '
			}
			buf.WriteRune(rune(ch))
		}
		buf.WriteString("\r\n")
	}

	_, err := s.Write(buf.Bytes())
	return err
}

// NativeSession returns a CliRenderer-compatible session adapter.
func (s *Session) NativeSession() *SessionAdapter {
	return &SessionAdapter{session: s}
}

func (s *Session) setPty(term string, cols, rows int) {
	s.mu.Lock()
	if term != "" {
		s.term = term
	}
	if cols > 0 {
		s.cols = cols
	}
	if rows > 0 {
		s.rows = rows
	}
	s.hasPty = true
	c, r := s.cols, s.rows
	listeners := make([]func(int, int), 0, len(s.resizeListeners))
	for _, cb := range s.resizeListeners {
		listeners = append(listeners, cb)
	}
	s.mu.Unlock()

	for _, cb := range listeners {
		cb(c, r)
	}
}

func (s *Session) resize(cols, rows int) {
	s.mu.Lock()
	if cols > 0 {
		s.cols = cols
	}
	if rows > 0 {
		s.rows = rows
	}
	c, r := s.cols, s.rows
	listeners := make([]func(int, int), 0, len(s.resizeListeners))
	for _, cb := range s.resizeListeners {
		listeners = append(listeners, cb)
	}
	s.mu.Unlock()

	for _, cb := range listeners {
		cb(c, r)
	}
}

func (s *Session) setEnv(key, val string) {
	s.mu.Lock()
	if s.env == nil {
		s.env = make(map[string]string)
	}
	s.env[key] = val
	s.mu.Unlock()
}

// armTimeouts arms the optional idle and max session timeouts.
func (s *Session) armTimeouts(idle, max time.Duration) {
	s.mu.Lock()
	s.idleTimeout = idle
	s.maxTimeout = max
	s.mu.Unlock()

	if idle > 0 {
		s.resetIdleLocked()
	}
	if max > 0 {
		s.maxTimer = time.AfterFunc(max, func() {
			s.Close()
		})
	}
}

func (s *Session) resetIdleLocked() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.idleTimeout <= 0 || s.closed {
		return
	}
	if s.idleTimer != nil {
		s.idleTimer.Stop()
	}
	s.idleTimer = time.AfterFunc(s.idleTimeout, func() {
		s.Close()
	})
}

// SessionAdapter adapts an SSH session for use as an io.ReadWriter.
type SessionAdapter struct {
	session *Session
}

// Write implements io.Writer for the adapter.
func (a *SessionAdapter) Write(data []byte) (int, error) {
	return a.session.Write(data)
}

// Read implements io.Reader for the adapter.
func (a *SessionAdapter) Read(data []byte) (int, error) {
	return a.session.Read(data)
}

// Close implements io.Closer for the adapter.
func (a *SessionAdapter) Close() error {
	return a.session.Close()
}

// ---------------------------------------------------------------------------
// Host key generation and loading
// ---------------------------------------------------------------------------

// GenerateHostKey generates a new host key of the given algorithm
// ("rsa", "ecdsa", or "ed25519").
func GenerateHostKey(algorithm string) (ssh.Signer, error) {
	return generateHostKey(algorithm)
}

func generateHostKey(algorithm string) (ssh.Signer, error) {
	switch algorithm {
	case "ed25519":
		_, priv, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("ssh: failed to generate ed25519 host key: %w", err)
		}
		signer, err := ssh.NewSignerFromKey(priv)
		if err != nil {
			return nil, fmt.Errorf("ssh: failed to create ed25519 signer: %w", err)
		}
		return signer, nil
	case "ecdsa":
		priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("ssh: failed to generate ecdsa host key: %w", err)
		}
		signer, err := ssh.NewSignerFromKey(priv)
		if err != nil {
			return nil, fmt.Errorf("ssh: failed to create ecdsa signer: %w", err)
		}
		return signer, nil
	case "rsa":
		priv, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, fmt.Errorf("ssh: failed to generate rsa host key: %w", err)
		}
		signer, err := ssh.NewSignerFromKey(priv)
		if err != nil {
			return nil, fmt.Errorf("ssh: failed to create rsa signer: %w", err)
		}
		return signer, nil
	default:
		return nil, fmt.Errorf("ssh: unsupported host key algorithm: %s", algorithm)
	}
}

// LoadHostKey reads a PEM private key from path and returns a host-key signer.
func LoadHostKey(path string) (ssh.Signer, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ssh: failed to read host key file %q: %w", path, err)
	}
	return ParseHostKeyPEM(data)
}

// ParseHostKeyPEM parses a PEM-encoded private key (RSA, ECDSA, or ED25519)
// into a host-key signer.
func ParseHostKeyPEM(pemBlock []byte) (ssh.Signer, error) {
	key, err := ssh.ParsePrivateKey(pemBlock)
	if err != nil {
		return nil, fmt.Errorf("ssh: failed to parse host key PEM: %w", err)
	}
	return key, nil
}

// sshFingerprint returns the OpenSSH-style SHA256 fingerprint of a public key.
func sshFingerprint(pub ssh.PublicKey) string {
	sum := sha256.Sum256(pub.Marshal())
	enc := base64.StdEncoding.EncodeToString(sum[:])
	enc = strings.TrimRight(enc, "=")
	return "SHA256:" + enc
}

// annotatePermissions stamps a ssh.Permissions with the authenticated identity
// so the connection handler can recover it after the handshake.
func annotatePermissions(perm *ssh.Permissions, method, username, fingerprint string, pubKey []byte) *ssh.Permissions {
	if perm == nil {
		perm = &ssh.Permissions{}
	}
	if perm.Extensions == nil {
		perm.Extensions = make(map[string]string)
	}
	perm.Extensions["pixtui.method"] = method
	perm.Extensions["pixtui.username"] = username
	if fingerprint != "" {
		perm.Extensions["pixtui.fingerprint"] = fingerprint
	}
	if pubKey != nil {
		perm.Extensions["pixtui.publickey"] = string(pubKey)
	}
	return perm
}

func identityFromConn(sshConn *ssh.ServerConn, noAuth bool) Identity {
	id := Identity{
		Method:   "none",
		Username: sshConn.User(),
	}
	if perm := sshConn.Permissions; perm != nil {
		if method := perm.Extensions["pixtui.method"]; method != "" {
			id.Method = method
			id.Username = perm.Extensions["pixtui.username"]
			id.Fingerprint = perm.Extensions["pixtui.fingerprint"]
			if pk := perm.Extensions["pixtui.publickey"]; pk != "" {
				if key, err := ssh.ParsePublicKey([]byte(pk)); err == nil {
					id.PublicKey = key
				}
			}
		} else if noAuth {
			id.Method = "none"
		}
	}
	return id
}

// ---------------------------------------------------------------------------
// SSH protocol payload parsing
// ---------------------------------------------------------------------------

func parseString(data []byte) (string, []byte, error) {
	if len(data) < 4 {
		return "", nil, io.ErrUnexpectedEOF
	}
	n := int(binary.BigEndian.Uint32(data))
	data = data[4:]
	if len(data) < n {
		return "", nil, io.ErrUnexpectedEOF
	}
	return string(data[:n]), data[n:], nil
}

func parseUint32(data []byte) (uint32, []byte, error) {
	if len(data) < 4 {
		return 0, nil, io.ErrUnexpectedEOF
	}
	return binary.BigEndian.Uint32(data), data[4:], nil
}

// parsePtyReq decodes an SSH "pty-req" channel request payload:
// string term, uint32 col width, uint32 row height, uint32 pixel width,
// uint32 pixel height, string encoded modes.
func parsePtyReq(payload []byte) (term string, cols, rows int, ok bool) {
	rest := payload
	var err error

	term, rest, err = parseString(rest)
	if err != nil {
		return
	}
	var c uint32
	c, rest, err = parseUint32(rest)
	if err != nil {
		return
	}
	var r uint32
	r, rest, err = parseUint32(rest)
	if err != nil {
		return
	}
	// Skip pixel width/height (2 x uint32).
	if len(rest) < 8 {
		return
	}
	rest = rest[8:]
	// Skip encoded terminal modes string.
	if _, _, err = parseString(rest); err != nil {
		return
	}
	return term, int(c), int(r), true
}

// parseWindowChange decodes an SSH "window-change" payload:
// uint32 col width, uint32 row height, uint32 pixel width, uint32 pixel height.
func parseWindowChange(payload []byte) (cols, rows int, ok bool) {
	rest := payload
	var err error
	var c uint32
	c, rest, err = parseUint32(rest)
	if err != nil {
		return
	}
	var r uint32
	r, _, err = parseUint32(rest)
	if err != nil {
		return
	}
	return int(c), int(r), true
}

// parseEnv decodes an SSH "env" payload: string key, string value.
func parseEnv(payload []byte) (key, val string, ok bool) {
	key, rest, err := parseString(payload)
	if err != nil {
		return
	}
	val, _, err = parseString(rest)
	if err != nil {
		return
	}
	return key, val, true
}

// sgrCode builds an ANSI SGR parameter string from color/attribute values.
func sgrCode(fg, bg uint16, attrs uint32) string {
	const (
		attrBold          uint32 = 1 << 0
		attrDim           uint32 = 1 << 1
		attrItalic        uint32 = 1 << 2
		attrUnderline     uint32 = 1 << 3
		attrBlink         uint32 = 1 << 4
		attrInverse       uint32 = 1 << 5
		attrHidden        uint32 = 1 << 6
		attrStrikethrough uint32 = 1 << 7
	)

	var parts []string
	if attrs&attrBold != 0 {
		parts = append(parts, "1")
	}
	if attrs&attrDim != 0 {
		parts = append(parts, "2")
	}
	if attrs&attrItalic != 0 {
		parts = append(parts, "3")
	}
	if attrs&attrUnderline != 0 {
		parts = append(parts, "4")
	}
	if attrs&attrBlink != 0 {
		parts = append(parts, "5")
	}
	if attrs&attrInverse != 0 {
		parts = append(parts, "7")
	}
	if attrs&attrHidden != 0 {
		parts = append(parts, "8")
	}
	if attrs&attrStrikethrough != 0 {
		parts = append(parts, "9")
	}
	if fg != 0 {
		parts = append(parts, fmt.Sprintf("3%d", fg%8))
	}
	if bg != 0 {
		parts = append(parts, fmt.Sprintf("4%d", bg%8))
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, ";")
}

var _ io.ReadWriter = (*Session)(nil)
