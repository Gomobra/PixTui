package core

import (
	"fmt"
	"sync"
	"time"

	"github.com/Gomobra/PixTui/packages/core/post"
	"github.com/Gomobra/PixTui/packages/native"
)

// RendererControlState represents the renderer control state.
type RendererControlState int

const (
	RendererControlStateActive RendererControlState = iota
	RendererControlStateExplicitSuspended
)

// CliRenderer is the central orchestrator of the rendering pipeline.
// Mirrors packages/core/src/renderer.ts with full frame stepping, paint recording, commit.
type CliRenderer struct {
	mu        sync.Mutex
	ctx       *RenderContext
	nativeCtx *native.ResourceContext
	session   *native.Session
	terminal  *native.Terminal
	buffer    *NativeBuffer

	width, height int
	frameID       uint64
	targetFPS     float64
	maxFPS        float64
	dirty         bool
	running       bool
	rendering     bool

	events RendererEvents
	controlState RendererControlState

	root *RootRenderable
	nativeScene *NativeScene
	lifecyclePasses *NativeLifecyclePasses

	postProcessor *post.PostProcessor
}

// NewCliRenderer creates a new CliRenderer
func NewCliRenderer(width, height int) *CliRenderer {
	nativeCtx := native.NewResourceContext()
	session, err := nativeCtx.CreateSession()
	if err != nil {
		panic(fmt.Sprintf("failed to create session: %v", err))
	}
	session.SetDimensions(width, height)

	nativeScene := NewNativeScene(nativeCtx, session)
	lifecyclePasses := NewNativeLifecyclePasses()

	r := &CliRenderer{
		nativeCtx:      nativeCtx,
		session:        session,
		terminal:       native.NewTerminal(),
		buffer:         NewNativeBuffer(width, height),
		width:          width,
		height:         height,
		targetFPS:      30,
		maxFPS:         60,
		controlState:   RendererControlStateActive,
		nativeScene:    nativeScene,
		lifecyclePasses: lifecyclePasses,
		postProcessor:  post.NewPostProcessor(),
	}

	r.ctx = &RenderContext{
		Width:            width,
		Height:           height,
		FrameID:          0,
		Clock:            SystemClock{},
		RequestRender:    func() { r.dirty = true },
		WidthMethod:      WidthMethodUnicode,
		Capabilities:     &TerminalCapabilities{},
		NativeScene:      nativeScene,
		LifecyclePasses:  lifecyclePasses,
	}

	return r
}

// SetDimensions sets the renderer dimensions
func (r *CliRenderer) SetDimensions(width, height int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.width = width
	r.height = height
	r.buffer.inner.Resize(width, height)
	r.session.SetDimensions(width, height)
	if r.ctx != nil {
		r.ctx.Width = width
		r.ctx.Height = height
	}
	if r.events.Resize != nil {
		r.events.Resize(width, height)
	}
}

// Dimensions returns the current dimensions
func (r *CliRenderer) Dimensions() (int, int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.width, r.height
}

// Buffer returns the render buffer
func (r *CliRenderer) Buffer() *NativeBuffer {
	return r.buffer
}

// Context returns the render context
func (r *CliRenderer) Context() *RenderContext {
	return r.ctx
}

// NativeSession returns the native session
func (r *CliRenderer) NativeSession() *native.Session {
	return r.session
}

// NativeContext returns the native context
func (r *CliRenderer) NativeContext() *native.ResourceContext {
	return r.nativeCtx
}

// NativeScene returns the native scene
func (r *CliRenderer) NativeScene() *NativeScene {
	return r.nativeScene
}

// PostProcessor returns the post-processor used to apply post-processing
// effects (filters, matrices, effects) to the render buffer after each frame.
func (r *CliRenderer) PostProcessor() *post.PostProcessor {
	return r.postProcessor
}

// AddPostEffect adds a post-processing effect to the renderer's pipeline.
func (r *CliRenderer) AddPostEffect(e post.PostEffect) {
	r.postProcessor.AddEffect(e)
}

// RemovePostEffect removes a post-processing effect from the renderer's pipeline.
func (r *CliRenderer) RemovePostEffect(e post.PostEffect) {
	r.postProcessor.RemoveEffect(e)
}

// Render performs a render pass over the renderable tree
func (r *CliRenderer) Render(root *RootRenderable, deltaTime float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.rendering = true
	defer func() { r.rendering = false }()

	r.frameID++
	r.ctx.FrameID = r.frameID
	r.dirty = false

	r.buffer.Clear()

	r.root = root

	if root != nil {
		// Compute Yoga layout for the entire tree before rendering.
		r.computeLayout(root.Renderable, float32(r.width), float32(r.height))
		r.renderRenderable(root, deltaTime)
	}

	// Run post-processing pipeline on the buffer after rendering the tree.
	r.postProcessor.Run(r.buffer, deltaTime)

	if r.events.Frame != nil {
		r.events.Frame(CliRendererFrameEvent{
			FrameID:  r.frameID,
			DeltaTime: deltaTime,
		})
	}

	return nil
}

// computeLayout runs Yoga layout on the node and propagates computed geometry
// to the renderable and all descendants.
func (r *CliRenderer) computeLayout(node RenderableIface, width, height float32) {
	rn, ok := node.(*Renderable)
	if !ok || rn.yogaNode == nil {
		return
	}

	yoga := rn.ctx.NativeScene.GetYoga()
	if yoga == nil {
		return
	}
	yoga.CalculateLayout(rn.yogaNode, width, height)
	layout := rn.yogaNode.GetComputedLayout()
	rn.mu.Lock()
	rn.x = layout.Left
	rn.y = layout.Top
	rn.width = layout.Width
	rn.height = layout.Height
	rn.mu.Unlock()

	for _, child := range rn.children {
		r.computeLayout(child, width, height)
	}
}

// renderRenderable recursively renders a renderable and its children
func (r *CliRenderer) renderRenderable(node RenderableIface, deltaTime float64) {
	if node == nil {
		return
	}

	// node is RenderableIface; check if it's also a *Renderable for IsDestroyed/IsVisible
	if rn, ok := node.(*Renderable); ok {
		if rn.IsDestroyed() || !rn.IsVisible() {
			return
		}
	}

	node.RenderSelf(r.buffer, deltaTime)

	if rn, ok := node.(*Renderable); ok {
		for _, child := range rn.Children() {
			r.renderRenderable(child, deltaTime)
		}
	}
}

// Capture returns a snapshot of the current frame
func (r *CliRenderer) Capture() native.BufferAccess {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.buffer.Access()
}

// RequestRender requests a re-render
func (r *CliRenderer) RequestRender() {
	r.mu.Lock()
	r.dirty = true
	r.mu.Unlock()
}

// SetResizeHandler sets the resize event handler
func (r *CliRenderer) SetResizeHandler(handler func(width, height int)) {
	r.events.Resize = handler
}

// SetKeyHandler sets the key event handler
func (r *CliRenderer) SetKeyHandler(handler func(data []byte)) {
	r.events.Key = handler
}

// Start begins terminal rendering
func (r *CliRenderer) Start() error {
	if err := r.terminal.EnableRawMode(); err != nil {
		return err
	}
	if err := r.terminal.EnableAlternateScreen(); err != nil {
		return err
	}
	if err := r.terminal.HideCursor(); err != nil {
		return err
	}
	if err := r.terminal.EnableMouse(); err != nil {
		return err
	}
	r.running = true
	return nil
}

// Stop restores terminal state
func (r *CliRenderer) Stop() error {
	r.running = false
	r.terminal.ShowCursor()
	r.terminal.DisableMouse()
	r.terminal.DisableAlternateScreen()
	return r.terminal.DisableRawMode()
}

// WriteFrame writes the current buffer to the terminal
func (r *CliRenderer) WriteFrame() error {
	access := r.buffer.Access()
	var output []byte
	for y := 0; y < access.Height; y++ {
		for x := 0; x < access.Width; x++ {
			ch, _, _, _, _ := r.buffer.inner.GetCell(x, y)
			if ch == 0 {
				ch = ' '
			}
			output = append(output, []byte(string(rune(ch)))...)
		}
		output = append(output, '\n')
	}
	_, err := r.terminal.WriteOutput(output)
	return err
}

// Run starts the render loop
func (r *CliRenderer) Run(root *RootRenderable) error {
	if !r.running {
		if err := r.Start(); err != nil {
			return err
		}
	}

	ticker := time.NewTicker(time.Duration(1000/float64(r.targetFPS)) * time.Millisecond)
	defer ticker.Stop()

	lastTime := time.Now()

	for r.running {
		select {
		case <-ticker.C:
			now := time.Now()
			deltaTime := now.Sub(lastTime).Seconds()
			lastTime = now

			r.Render(root, deltaTime)
			if err := r.WriteFrame(); err != nil {
				return err
			}
		}
	}

	return nil
}

// Destroy cleans up all resources
func (r *CliRenderer) Destroy() error {
	r.Stop()
	if r.events.Destroy != nil {
		r.events.Destroy()
	}
	return r.nativeCtx.Close()
}

// IsDestroyed returns whether the renderer's native context has been closed.
func (r *CliRenderer) IsDestroyed() bool {
	return r.nativeCtx.Status() == native.ContextStatusClosed
}

// IsRendering returns whether the renderer is currently in the middle of a render pass.
// Rendering in the Go port is synchronous, so this is true only while Render() is
// executing; observers outside a render pass will see false.
func (r *CliRenderer) IsRendering() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rendering
}

var _ = fmt.Sprintf
