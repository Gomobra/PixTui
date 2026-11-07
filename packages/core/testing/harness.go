package testing

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/Gomobra/PixTui/packages/core"
	"github.com/Gomobra/PixTui/packages/native"
)

// TestHarness provides setup/teardown for a test environment backed by a CliRenderer.
// It exposes the renderer, render context, and native buffer for use in tests.
type TestHarness struct {
	renderer *core.CliRenderer
	context  *core.RenderContext
	buffer   *core.NativeBuffer
	width    int
	height   int
}

// NewTestHarness creates a new TestHarness with the given dimensions.
// The renderer is constructed with a fresh native session and an empty buffer.
func NewTestHarness(width, height int) *TestHarness {
	renderer := core.NewCliRenderer(width, height)
	return &TestHarness{
		renderer: renderer,
		context:  renderer.Context(),
		buffer:   renderer.Buffer(),
		width:    width,
		height:   height,
	}
}

// Renderer returns the underlying CliRenderer.
func (h *TestHarness) Renderer() *core.CliRenderer {
	return h.renderer
}

// Context returns the RenderContext associated with the renderer.
func (h *TestHarness) Context() *core.RenderContext {
	return h.context
}

// Buffer returns the NativeBuffer associated with the renderer.
func (h *TestHarness) Buffer() *core.NativeBuffer {
	return h.buffer
}

// Render performs a render pass over the given root renderable with the specified delta time.
func (h *TestHarness) Render(root *core.RootRenderable, deltaTime float64) {
	if err := h.renderer.Render(root, deltaTime); err != nil {
		panic(fmt.Sprintf("test harness render failed: %v", err))
	}
}

// Capture returns a snapshot of the current frame buffer.
func (h *TestHarness) Capture() native.BufferAccess {
	return h.renderer.Capture()
}

// Close releases all resources held by the harness (terminal state, native session, context).
func (h *TestHarness) Close() error {
	return h.renderer.Destroy()
}

// Settle runs `turns` host turns, allowing scheduled work and goroutines to progress.
// Mirrors the TypeScript settle() helper.
func (h *TestHarness) Settle(turns int) {
	for i := 0; i < turns; i++ {
		runtime.Gosched()
	}
}

// SettleUntil runs host turns until `predicate` holds.
// Returns an error if the condition is not met within `maxTurns` turns,
// so a stalled renderer fails fast.
func (h *TestHarness) SettleUntil(predicate func() bool, maxTurns int) error {
	for turn := 0; turn < maxTurns; turn++ {
		if predicate() {
			return nil
		}
		h.Settle(1)
	}
	if !predicate() {
		return fmt.Errorf("condition not met within %d host turns", maxTurns)
	}
	return nil
}

// ServiceReadyFrames runs host turns until the renderer is idle: not rendering and no
// pending ready-frame or output-idle retry work. Mirrors the TypeScript
// serviceReadyFrames() helper, which reads renderer scheduler internals
// (rendering, cancelReadyFrame, outputIdleRenderScheduled).
//
// In the Go port rendering is synchronous and there is no output-idle retry, so the
// pending-frame flags are always false; the predicate reduces to the renderer not being
// mid-render. A stalled renderer fails fast after maxTurns.
func (h *TestHarness) ServiceReadyFrames(maxTurns int) error {
	return h.SettleUntil(func() bool {
		return !h.renderer.IsRendering()
	}, maxTurns)
}

// ForceRenderStatus replaces the Session frame commit status.
// `next` picks the status of each commit: Presented commits normally; any other
// status consumes the frame without output, as the native Session does under pressure.
// Returns a restore function.
func (h *TestHarness) ForceRenderStatus(next func() native.SessionRenderStatus) func() {
	return h.renderer.NativeSession().SetRenderStatusFunc(next)
}

// HeldOutputIdle holds every Session idle() wait, so a skipped frame's output-idle retry
// runs only on Release(). Mirrors the TypeScript holdOutputIdle() helper.
type HeldOutputIdle struct {
	mu       sync.Mutex
	session  *native.Session
	released bool
	calls    int
	releaseCh chan struct{}
	restore   func()
}

// newHeldOutputIdle creates a HeldOutputIdle that overrides the session's idle wait.
func newHeldOutputIdle(session *native.Session) *HeldOutputIdle {
	h := &HeldOutputIdle{
		session:   session,
		releaseCh: make(chan struct{}),
	}
	idle := func() {
		h.mu.Lock()
		h.calls++
		if h.released {
			h.mu.Unlock()
			return
		}
		ch := h.releaseCh
		h.mu.Unlock()
		if ch != nil {
			<-ch
		}
	}
	h.restore = session.SetIdleFunc(idle)
	return h
}

// Release resolves the held wait; later idle() calls pass through until Hold().
func (h *HeldOutputIdle) Release() error {
	h.mu.Lock()
	if h.released {
		h.mu.Unlock()
		return nil
	}
	h.released = true
	ch := h.releaseCh
	h.releaseCh = nil
	h.mu.Unlock()
	if ch != nil {
		close(ch)
	}
	// Yield twice to mirror the TypeScript double microtask flush.
	runtime.Gosched()
	runtime.Gosched()
	return nil
}

// Hold makes subsequent idle() calls block until Release() is called.
func (h *HeldOutputIdle) Hold() {
	h.mu.Lock()
	h.released = false
	h.releaseCh = make(chan struct{})
	h.mu.Unlock()
}

// Calls returns the number of idle() calls since the hold started.
func (h *HeldOutputIdle) Calls() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.calls
}

// Restore restores the original session idle function.
func (h *HeldOutputIdle) Restore() {
	h.restore()
}

// HoldOutputIdle installs an idle override on the renderer's session and returns a
// HeldOutputIdle controller. Use Hold() to block idle waits and Release() to let them proceed.
func (h *TestHarness) HoldOutputIdle() *HeldOutputIdle {
	return newHeldOutputIdle(h.renderer.NativeSession())
}

// ProcessListenerCounts returns a snapshot of process listener counts for renderer-related events.
// In the Go port there are no OS process listeners, so this returns an empty map.
// It is provided for API parity with the TypeScript reference.
func (h *TestHarness) ProcessListenerCounts() map[string]int {
	return make(map[string]int)
}

// AssertRendererReleased waits for the renderer to be destroyed and checks that it released
// everything it acquired: the native context/session and the root renderable.
func (h *TestHarness) AssertRendererReleased() error {
	if !h.renderer.IsDestroyed() {
		return fmt.Errorf("renderer is not destroyed")
	}
	if h.renderer.NativeContext().Status() != native.ContextStatusClosed {
		return fmt.Errorf("native context is not closed")
	}
	if h.context.NativeScene != nil && !h.context.NativeScene.IsDestroyed() {
		return fmt.Errorf("native scene is not destroyed")
	}
	return nil
}

var _ = fmt.Sprintf
