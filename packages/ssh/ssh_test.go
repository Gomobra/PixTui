package ssh

import (
	"io"
	"testing"

	"github.com/Gomobra/PixTui/packages/core"
)

// mockChannel implements ssh.Channel for testing without a real connection.
type mockChannel struct {
	closed bool
	data   []byte
}

func (m *mockChannel) Read(data []byte) (int, error)   { return 0, io.EOF }
func (m *mockChannel) Write(data []byte) (int, error) {
	m.data = append(m.data, data...)
	return len(data), nil
}
func (m *mockChannel) Close() error                      { m.closed = true; return nil }
func (m *mockChannel) CloseWrite() error                 { return nil }
func (m *mockChannel) SendRequest(string, bool, []byte) (bool, error) {
	return false, nil
}
func (m *mockChannel) Stderr() io.ReadWriter             { return nil }
func (m *mockChannel) SetNonBlocking(bool) error         { return nil }

func TestServerCreate(t *testing.T) {
	server := NewServer("127.0.0.1:2222", nil)
	if server == nil {
		t.Fatal("expected non-nil server")
	}
}

func TestServerAddHostKey(t *testing.T) {
	server := NewServer("127.0.0.1:2222", nil)
	server.AddHostKey(nil)
}

func TestSessionCreate(t *testing.T) {
	server := NewServer("127.0.0.1:2222", nil)
	session := &Session{
		id:             "test-session",
		server:         server,
		resizeListeners: make(map[int]func(int, int)),
	}
	if session.ID() != "test-session" {
		t.Errorf("expected ID 'test-session', got '%s'", session.ID())
	}
}

func TestSessionAdapter(t *testing.T) {
	server := NewServer("127.0.0.1:2222", nil)
	session := &Session{id: "test", server: server}
	adapter := session.NativeSession()
	if adapter == nil {
		t.Fatal("expected non-nil adapter")
	}
}

func TestSessionSetWindowSize(t *testing.T) {
	server := NewServer("127.0.0.1:2223", nil)
	session := &Session{
		id:             "test-resize",
		server:         server,
		resizeListeners: make(map[int]func(int, int)),
	}
	session.SetWindowSize(120, 40)
	cols, rows := session.GetWindowSize()
	if cols != 120 || rows != 40 {
		t.Errorf("expected window size 120x40, got %dx%d", cols, rows)
	}
}

func TestSessionSetWindowSizeResizeListener(t *testing.T) {
	server := NewServer("127.0.0.1:2224", nil)
	session := &Session{
		id:             "test-listener",
		server:         server,
		resizeListeners: make(map[int]func(int, int)),
	}
	called := false
	session.OnResize(func(cols, rows int) {
		if cols == 100 && rows == 30 {
			called = true
		}
	})
	session.SetWindowSize(100, 30)
	if !called {
		t.Error("expected resize listener to be called")
	}
}

func TestSessionWriteFrame(t *testing.T) {
	server := NewServer("127.0.0.1:2225", nil)
	session := &Session{
		id:     "test-frame",
		server: server,
		channel: &mockChannel{},
	}
	renderer := core.NewCliRenderer(80, 24)
	err := session.WriteFrame(renderer)
	if err != nil {
		t.Errorf("expected no error writing frame, got %v", err)
	}
}

func TestServerListenClose(t *testing.T) {
	server := NewServer("127.0.0.1:2226", nil)
	err := server.Listen()
	if err != nil {
		t.Fatalf("expected Listen to succeed, got %v", err)
	}
	err = server.Close()
	if err != nil {
		t.Errorf("expected Close to succeed, got %v", err)
	}
}
