package core

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/Gomobra/PixTui/packages/native"
)

// NativeSessionScheduler schedules work on a later event-loop turn.
type NativeSessionScheduler interface {
	Now() int64
	Schedule(callback func(), delayMs int) func()
}

// monotonicScheduler implements NativeSessionScheduler using the Go clock.
type monotonicScheduler struct{}

func (monotonicScheduler) Now() int64 { return time.Now().UnixNano() }

func (monotonicScheduler) Schedule(callback func(), delayMs int) func() {
	timer := time.AfterFunc(time.Duration(delayMs)*time.Millisecond, callback)
	return func() { timer.Stop() }
}

// defaultScheduler is the shared default scheduler.
var defaultScheduler NativeSessionScheduler = monotonicScheduler{}

// NativeSessionDriverOptions configures a NativeSession driver.
type NativeSessionDriverOptions struct {
	Context         *NativeContextOptions
	Output          *NativeSessionOptions
	OutputBufferSize int
	CloseTimeoutMs  int
	Scheduler       NativeSessionScheduler
}

// completion is a one-shot future.
type completion struct {
	mu     sync.Mutex
	done   chan struct{}
	err    error
	closed bool
}

func newCompletion() *completion { return &completion{done: make(chan struct{})} }

func (c *completion) promise() <-chan struct{} { return c.done }

func (c *completion) resolve() {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	c.closed = true
	close(c.done)
	c.mu.Unlock()
}

func (c *completion) reject(err error) {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	c.err = err
	c.closed = true
	close(c.done)
	c.mu.Unlock()
}

// terminalTransition tracks an in-flight terminal setup/suspend/resume.
type terminalTransition struct {
	kind string
	wait *completion
}

// scheduledTask tracks a pending scheduled turn.
type scheduledTask struct {
	deadlineNs *int64
	cancel     func()
}

// outputState tracks an in-flight sink write.
type outputState struct {
	ticket   *NativeOutputTicket
	completed bool
	failed    bool
}

// stdoutPostState tracks a stdout writer copy in flight.
type stdoutPostState struct {
	since int64
	polls int
}

// NativeSession owns one checked Session and optionally its Context, driving
// output to an io.Writer sink. Mirrors packages/core/src/NativeSession.ts.
type NativeSession struct {
	mu              sync.Mutex
	lib             RenderLib
	resourceCtx     *ResourceContext
	nativeSession   *native.Session
	sink            io.Writer
	sinkWrite       func([]byte) (int, error)
	buffer          []byte

	scheduler NativeSessionScheduler
	maxWriteBytes int64
	maxAtomicWriteBytes int64
	closeTimeoutNs int64

	closeWait          *completion
	idleWait           *completion
	presentationWait   *completion
	imageUploadsPending int

	transition        *terminalTransition
	scheduled         *scheduledTask
	output            *outputState
	stdoutPost        *stdoutPostState
	blocked           bool
	sinkErrorSeen     bool
	detached          bool
	closeDeadlineNs   int64
	closeCapNs        int64
	closeLate         error
	closeRequested    bool
	closeInterruption error
	stopped           bool
	exiting           bool
	finishing         bool
	disposed          bool
	sessionErr        error

	detachedSessions map[*NativeSession]struct{}
}

// NewNativeSession creates a new NativeSession writing to sink.
func NewNativeSession(sink io.Writer, options *NativeSessionDriverOptions, owner *NativeSession) (*NativeSession, error) {
	if options == nil {
		options = &NativeSessionDriverOptions{}
	}
	size := options.OutputBufferSize
	if size <= 0 {
		size = 65536
	}
	timeout := options.CloseTimeoutMs
	if timeout <= 0 {
		timeout = 1000
	}
	output := options.Output
	if output == nil {
		output = &NativeSessionOptions{
			ChunkSize:      65536,
			SpanCapacity:   128,
			MaxBytes:       8388608,
			ControlCapacity: 4096,
		}
	}
	if size > 0xffffffff {
		return nil, fmt.Errorf("NativeSession outputBufferSize must be a positive u32")
	}
	if timeout < 0 || timeout > 0x7fffffff {
		return nil, fmt.Errorf("NativeSession closeTimeoutMs must be an integer from 0 to 2147483647")
	}

	sinkWrite, ok := sink.(interface{ Write([]byte) (int, error) })
	if !ok {
		return nil, errors.New("NativeSession sink is not writable")
	}

	buffer := make([]byte, size)

	s := &NativeSession{
		lib:                 noopRenderLib{},
		resourceCtx:         NewResourceContext(),
		sink:                sink,
		sinkWrite:           sinkWrite.Write,
		buffer:              buffer,
		maxWriteBytes:       output.MaxBytes,
		closeTimeoutNs:      int64(timeout) * 1_000_000,
		closeWait:           newCompletion(),
		detachedSessions:    make(map[*NativeSession]struct{}),
	}
	s.scheduler = options.Scheduler
	if s.scheduler == nil {
		s.scheduler = defaultScheduler
	}

	if owner != nil {
		if err := owner.checkOpen(); err != nil {
			return nil, err
		}
		s.resourceCtx = owner.resourceCtx
	}

	nativeSession, err := s.resourceCtx.inner.CreateSession()
	if err != nil {
		s.stopped = true
		s.resourceCtx.Destroy()
		return nil, err
	}
	s.nativeSession = nativeSession
	s.maxAtomicWriteBytes = int64(nativeSession.ID())

	if owner != nil {
		owner.mu.Lock()
		owner.detachedSessions[s] = struct{}{}
		owner.mu.Unlock()
	}

	return s, nil
}

// Context returns the native context handle.
func (s *NativeSession) Context() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.resourceCtx.context
}

// RenderLib returns the render library bound to this session.
func (s *NativeSession) RenderLib() RenderLib { return s.lib }

// Session returns the underlying native session.
func (s *NativeSession) Session() *native.Session { return s.nativeSession }

// Disposed reports whether native teardown is complete.
func (s *NativeSession) Disposed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.disposed
}

// Error returns the session error, if any.
func (s *NativeSession) Error() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sessionErr
}

// Open reports whether the session is active and not closing.
func (s *NativeSession) Open() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sessionErr == nil && !s.closeRequested && !s.exiting
}

// UsesOutput reports whether this session writes to the given sink.
func (s *NativeSession) UsesOutput(sink io.Writer) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sink == sink
}

// AttachRenderer attaches a renderer to the session.
func (s *NativeSession) AttachRenderer(options NativeSessionRendererOptions) error {
	s.mu.Lock()
	if s.sessionErr != nil || s.stopped || s.closeRequested {
		s.mu.Unlock()
		return errors.New("NativeSession is not open")
	}
	s.mu.Unlock()
	s.lib.SessionAttachRenderer(s.resourceCtx.context, s.nativeSession.ID(), options)
	return nil
}

// CreateDetached creates a detached child session sharing this session's context.
func (s *NativeSession) CreateDetached(options NativeSessionRendererOptions) (*NativeSession, error) {
	s.mu.Lock()
	if s.sessionErr != nil || s.stopped || s.closeRequested {
		s.mu.Unlock()
		return nil, errors.New("NativeSession is not open")
	}
	s.mu.Unlock()

	child, err := NewNativeSession(&discardWriter{}, &NativeSessionDriverOptions{
		Output:           &NativeSessionOptions{ChunkSize: 4096, SpanCapacity: 2, MaxBytes: 8192, ControlCapacity: 0},
		OutputBufferSize: 4096,
		Scheduler:        s.scheduler,
	}, s)
	if err != nil {
		return nil, err
	}
	if err := child.AttachRenderer(options); err != nil {
		child.Dispose()
		return nil, err
	}
	return child, nil
}

// SetupTerminal sets up the terminal.
func (s *NativeSession) SetupTerminal(options NativeSessionTerminalOptions) error {
	return s.changeTerminal("setup", options)
}

// Suspend suspends the terminal.
func (s *NativeSession) Suspend() error { return s.changeTerminal("suspend", NativeSessionTerminalOptions{}) }

// Resume resumes the terminal.
func (s *NativeSession) Resume() error { return s.changeTerminal("resume", NativeSessionTerminalOptions{}) }

// Render commits a frame.
func (s *NativeSession) Render(force bool, frame NativeSceneFrameRequest) NativeSessionRenderStatus {
	s.mu.Lock()
	if s.sessionErr != nil || s.stopped || s.closeRequested {
		s.mu.Unlock()
		return NativeSessionRenderStatusClosed
	}
	s.mu.Unlock()
	result := s.lib.SessionFrameCommit(s.resourceCtx.context, s.nativeSession.ID(), frame, force)
	if result == NativeSessionRenderStatusPending {
		s.schedule(nil)
	}
	return result
}

// SetScreen sets the alternate screen state.
func (s *NativeSession) SetScreen(alternate bool, width, height int, trailingOutput []byte) {
	s.mu.Lock()
	if s.sessionErr != nil || s.stopped || s.closeRequested {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()
	s.lib.SessionSetScreen(s.resourceCtx.context, s.nativeSession.ID(), alternate, width, height, trailingOutput)
	s.schedule(nil)
}

// Resize resizes the renderer.
func (s *NativeSession) Resize(width, height int) {
	s.mu.Lock()
	if s.sessionErr != nil || s.stopped || s.closeRequested {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()
	s.lib.SessionResizeRenderer(s.resourceCtx.context, s.nativeSession.ID(), width, height)
}

// Control sends a control command.
func (s *NativeSession) Control(command NativeSessionControl) {
	s.mu.Lock()
	if s.sessionErr != nil || s.stopped || s.closeRequested {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()
	_ = s.lib.SessionControl(s.resourceCtx.context, s.nativeSession.ID(), command)
	s.schedule(nil)
}

// WriteClipboard writes to the clipboard.
func (s *NativeSession) WriteClipboard(target int, bytes []byte) bool {
	s.mu.Lock()
	if s.sessionErr != nil || s.stopped || s.closeRequested {
		s.mu.Unlock()
		return false
	}
	s.mu.Unlock()
	accepted := s.lib.SessionClipboard(s.resourceCtx.context, s.nativeSession.ID(), target, bytes)
	if accepted {
		s.schedule(nil)
	}
	return accepted
}

// TriggerNotification triggers a desktop notification.
func (s *NativeSession) TriggerNotification(message, title string) bool {
	s.mu.Lock()
	if s.sessionErr != nil || s.stopped || s.closeRequested {
		s.mu.Unlock()
		return false
	}
	s.mu.Unlock()
	accepted := s.lib.SessionNotification(s.resourceCtx.context, s.nativeSession.ID(), message, title)
	if accepted {
		s.schedule(nil)
	}
	return accepted
}

// SetProgramStatus sets the program status.
func (s *NativeSession) SetProgramStatus(status ProgramStatus) bool {
	s.mu.Lock()
	if s.sessionErr != nil || s.stopped || s.closeRequested {
		s.mu.Unlock()
		return false
	}
	s.mu.Unlock()
	_ = status
	s.schedule(nil)
	return true
}

// GetCapabilities returns session capabilities.
func (s *NativeSession) GetCapabilities() NativeSessionCapabilities {
	s.mu.Lock()
	if s.sessionErr != nil || s.stopped || s.closeRequested {
		s.mu.Unlock()
		return NativeSessionCapabilities{}
	}
	s.mu.Unlock()
	return s.lib.SessionGetCapabilities(s.resourceCtx.context, s.nativeSession.ID())
}

// Write writes bytes to the session output.
func (s *NativeSession) Write(bytes []byte) bool {
	s.mu.Lock()
	if s.sessionErr != nil || s.stopped || s.closeRequested {
		s.mu.Unlock()
		return false
	}
	s.mu.Unlock()
	err := s.lib.SessionWrite(s.resourceCtx.context, s.nativeSession.ID(), bytes)
	if err != nil {
		return false
	}
	s.schedule(nil)
	return true
}

// Idle waits for accepted output and sink pressure to settle.
func (s *NativeSession) Idle() <-chan struct{} {
	s.mu.Lock()
	if s.stopped {
		ch := s.closeWait.promise()
		s.mu.Unlock()
		return ch
	}
	if s.idleWait != nil {
		ch := s.idleWait.promise()
		s.mu.Unlock()
		return ch
	}
	wait := newCompletion()
	s.idleWait = wait
	s.mu.Unlock()
	s.schedule(nil)
	return wait.promise()
}

// Close requests a graceful close of the session.
func (s *NativeSession) Close() <-chan struct{} {
	s.mu.Lock()
	if s.stopped || s.closeRequested {
		ch := s.closeWait.promise()
		s.mu.Unlock()
		return ch
	}
	s.closeRequested = true
	s.closeDeadlineNs = s.scheduler.Now() + s.closeTimeoutNs
	s.closeCapNs = s.scheduler.Now() + 10*s.closeTimeoutNs
	if s.transition != nil {
		s.closeInterruption = errors.New("NativeSession terminal transition interrupted by close")
		s.transition.wait.reject(s.closeInterruption)
		s.transition = nil
	}
	if s.sessionErr != nil {
		s.finish(s.sessionErr)
		s.mu.Unlock()
		return s.closeWait.promise()
	}
	s.lib.SessionClose(s.resourceCtx.context, s.nativeSession.ID())
	s.schedule(nil)
	s.mu.Unlock()
	return s.closeWait.promise()
}

// Dispose cancels the session without graceful close.
func (s *NativeSession) Dispose() {
	s.mu.Lock()
	s.finish(s.sessionErr)
	s.mu.Unlock()
}

// RestoreOnExit restores the terminal before process exit.
func (s *NativeSession) RestoreOnExit() bool {
	s.mu.Lock()
	if s.disposed || s.finishing {
		s.mu.Unlock()
		return false
	}
	s.exiting = true
	s.mu.Unlock()
	return true
}

// checkOpen returns an error if the session is not open.
func (s *NativeSession) checkOpen() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.sessionErr != nil {
		return s.sessionErr
	}
	if s.stopped || s.closeRequested {
		if s.closeLate != nil {
			return s.closeLate
		}
		return errors.New("NativeSession is closing or disposed")
	}
	return nil
}

func (s *NativeSession) changeTerminal(kind string, options NativeSessionTerminalOptions) error {
	s.mu.Lock()
	if s.sessionErr != nil || s.stopped || s.closeRequested {
		s.mu.Unlock()
		return errors.New("NativeSession is not open")
	}
	if s.transition != nil {
		if s.transition.kind == kind {
			ch := s.transition.wait.promise()
			s.mu.Unlock()
			<-ch
			return nil
		}
		s.mu.Unlock()
		return errors.New("NativeSession terminal transition is pending")
	}
	switch kind {
	case "setup":
		s.lib.SessionSetupTerminal(s.resourceCtx.context, s.nativeSession.ID(), options)
	case "suspend":
		s.lib.SessionSuspend(s.resourceCtx.context, s.nativeSession.ID())
	case "resume":
		s.lib.SessionResume(s.resourceCtx.context, s.nativeSession.ID())
	}
	wait := newCompletion()
	s.transition = &terminalTransition{kind: kind, wait: wait}
	s.mu.Unlock()
	s.schedule(nil)
	<-wait.promise()
	return nil
}

// schedule arranges a turn on the scheduler.
func (s *NativeSession) schedule(deadlineNs *int64) {
	s.mu.Lock()
	if s.exiting {
		if s.stopped && s.canDetach() {
			s.detach()
		}
		s.mu.Unlock()
		return
	}
	if s.stopped && (s.detached || !s.canDetach()) {
		s.mu.Unlock()
		return
	}
	if s.scheduled != nil {
		if s.scheduled.deadlineNs == nil {
			s.mu.Unlock()
			return
		}
		if deadlineNs != nil && *s.scheduled.deadlineNs <= *deadlineNs {
			s.mu.Unlock()
			return
		}
		if s.scheduled.cancel != nil {
			s.scheduled.cancel()
		}
	}
	task := &scheduledTask{deadlineNs: deadlineNs}
	s.scheduled = task
	var delayMs int
	if deadlineNs != nil {
		remaining := *deadlineNs - s.scheduler.Now()
		if remaining < 0 {
			remaining = 0
		}
		delayMs = int((remaining + 999999) / 1000000)
		if delayMs > 0x7fffffff {
			delayMs = 0x7fffffff
		}
	}
	cancel := s.scheduler.Schedule(func() {
		s.mu.Lock()
		if s.scheduled != task {
			s.mu.Unlock()
			return
		}
		s.scheduled = nil
		s.mu.Unlock()
		s.turn()
	}, delayMs)
	task.cancel = cancel
	s.mu.Unlock()
}

func (s *NativeSession) turn() {
	s.mu.Lock()
	if s.sessionErr != nil {
		sessionErr := s.sessionErr
		s.mu.Unlock()
		s.finish(sessionErr)
		return
	}
	now := s.scheduler.Now()

	if s.output != nil && s.output.completed {
		if s.output.ticket != nil {
			s.lib.SessionCompleteOutput(s.resourceCtx.context, s.nativeSession.ID(), s.output.ticket, true)
		}
		s.output = nil
		s.completePresentation()
	}

	var result *nativePumpResult
	if s.output == nil && s.stdoutPost == nil {
		result = s.lib.SessionPump(s.resourceCtx.context, s.nativeSession.ID(), now, 1)
	}
	s.mu.Unlock()

	if result == nil {
		s.mu.Lock()
		s.schedule(&now)
		s.mu.Unlock()
		return
	}

	switch result.status {
	case NativeSessionPumpStatusIdle:
		s.mu.Lock()
		if s.transition != nil {
			_ = s.transition
		}
		if !s.blocked && s.idleWait != nil {
			s.idleWait.resolve()
			s.idleWait = nil
		}
		s.mu.Unlock()
	case NativeSessionPumpStatusAgain:
		s.schedule(nil)
	case NativeSessionPumpStatusOutputPending:
		s.mu.Lock()
		if !s.blocked {
			s.writeOutput()
		}
		s.mu.Unlock()
	case NativeSessionPumpStatusWaitUntil:
		s.schedule(&result.deadlineNs)
	case NativeSessionPumpStatusClosed:
		s.finish(s.closeLate)
	}

	s.mu.Lock()
	if now != 0 {
		s.schedule(&now)
	}
	s.mu.Unlock()
}

type nativePumpResult struct {
	status     NativeSessionPumpStatus
	deadlineNs int64
}

func (s *NativeSession) writeOutput() {
	s.mu.Lock()
	s.output = nil
	s.blocked = false
	s.mu.Unlock()
	s.schedule(nil)
}

func (s *NativeSession) completePresentation() {
	s.mu.Lock()
	wait := s.presentationWait
	s.mu.Unlock()
	if wait == nil {
		return
	}
	state := s.lib.SessionGetRendererState(s.resourceCtx.context, s.nativeSession.ID())
	s.mu.Lock()
	if !state.FramePending {
		s.presentationWait.resolve()
		s.presentationWait = nil
	}
	s.mu.Unlock()
}

// NativeSessionPumpStatus is the status returned by the native output pump.
type NativeSessionPumpStatus int

const (
	NativeSessionPumpStatusIdle NativeSessionPumpStatus = iota
	NativeSessionPumpStatusAgain
	NativeSessionPumpStatusOutputPending
	NativeSessionPumpStatusWaitUntil
	NativeSessionPumpStatusClosed
)

func (s *NativeSession) canDetach() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.output != nil && !s.output.completed {
		return false
	}
	return true
}

func (s *NativeSession) detach() {
	s.mu.Lock()
	s.detached = true
	s.output = nil
	s.mu.Unlock()
}

func (s *NativeSession) finish(err error) {
	s.mu.Lock()
	if s.finishing {
		s.mu.Unlock()
		return
	}
	if s.disposed {
		s.schedule(nil)
		s.mu.Unlock()
		return
	}
	s.finishing = true
	s.stopped = true
	if s.sessionErr == nil {
		s.sessionErr = err
	}
	scheduled := s.scheduled
	s.scheduled = nil
	if scheduled != nil && scheduled.cancel != nil {
		scheduled.cancel()
	}
	if s.sessionErr != nil {
		s.lib.SessionCancel(s.resourceCtx.context, s.nativeSession.ID())
	}
	for child := range s.detachedSessions {
		child.Dispose()
	}
	if len(s.detachedSessions) == 0 {
		s.attachmentCleanup()
		s.lib.GetYogaHost().RunMutation(func() error {
			if !s.resourceCtx.Disposed() {
				s.lib.DestroySession(s.resourceCtx.context, s.nativeSession.ID())
			}
			s.disposed = true
			return nil
		})
	}
	if !s.disposed && s.sessionErr == nil {
		s.mu.Unlock()
		return
	}
	for _, wait := range []*completion{s.closeWait, s.idleWait, s.presentationWait} {
		if wait != nil {
			if s.sessionErr != nil {
				wait.reject(s.sessionErr)
			} else {
				wait.resolve()
			}
		}
	}
	s.idleWait = nil
	s.presentationWait = nil
	s.transition = nil
	s.finishing = false
	s.mu.Unlock()
}

func (s *NativeSession) attachmentCleanup() {}

// discardWriter is an io.Writer that discards all data.
type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) { return len(p), nil }

var _ = errors.New
var _ = fmt.Sprintf
