package core

import (
	"fmt"
	"sync"

	"github.com/Gomobra/PixTui/packages/native"
)

// PaintHook is a hook called during paint: (buffer, deltaTime)
type PaintHook func(buffer *NativeBuffer, deltaTime float64)

// Hook names used across the renderable lifecycle.
var hookNames = []string{
	"renderSelf",
	"onResize",
	"onLayoutResize",
	"onUpdate",
	"renderBefore",
	"renderAfter",
	"onLifecyclePass",
}

// RenderableOptions, LayoutEvents, RenderableEvents, Position, LayoutOptions are defined in types.go

// NativeBuffer wraps native.OptimizedBuffer for the core layer.
type NativeBuffer struct {
	mu    sync.RWMutex
	inner *native.OptimizedBuffer
}

// NewNativeBuffer creates a new NativeBuffer.
func NewNativeBuffer(width, height int) *NativeBuffer {
	return &NativeBuffer{
		inner: native.NewOptimizedBuffer(width, height),
	}
}

// Width returns the buffer width.
func (b *NativeBuffer) Width() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.inner.Width()
}

// Height returns the buffer height.
func (b *NativeBuffer) Height() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.inner.Height()
}

// SetCell sets a cell at (x, y).
func (b *NativeBuffer) SetCell(x, y int, ch uint32, fg, bg uint16, attrs uint32) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.inner.SetCell(x, y, ch, fg, bg, attrs)
}

// Clear clears the buffer.
func (b *NativeBuffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.inner.Clear()
}

// Access returns a snapshot of the buffer.
func (b *NativeBuffer) Access() native.BufferAccess {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.inner.Access()
}

// FillRect fills a rectangular region.
func (b *NativeBuffer) FillRect(x, y, w, h int, ch uint32, fg, bg uint16, attrs uint32) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.inner.FillRect(x, y, w, h, ch, fg, bg, attrs)
}

// GetCell returns the cell at (x, y).
func (b *NativeBuffer) GetCell(x, y int) (uint32, uint16, uint16, uint32, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.inner.GetCell(x, y)
}

// NativeRenderableIntegration describes how a renderable integrates with the native scene.
type NativeRenderableIntegration struct {
	Kind  native.SceneNodeKind
	Body  string // "host" or custom
	BeforeAfter bool
}

// Renderable is the base class for all UI renderable nodes.
// Mirrors packages/core/src/Renderable.ts with Yoga layout integration.
type Renderable struct {
	mu        sync.RWMutex
	id        string
	num       int
	ctx       *RenderContext
	options   RenderableOptions[Renderable]
	children  []RenderableIface
	parent    *Renderable

	width      float32
	height     float32
	x, y       float32
	zIndex     int
	visible    bool
	buffered   bool
	live       bool
	liveCount  int
	opacity    float64
	focusable  bool
	focused    bool
	destroyed  bool
	dirty      bool

	// Native scene integration
	nativeScene *NativeScene
	nodeHandle  *SceneNodeHandle
	nativeKind  native.SceneNodeKind
	yogaFreed   bool

	// Yoga layout node for this renderable.
	yogaNode *native.YogaNode

	// Layout (Yoga-backed via native scene)
	positionType string
	overflow     string
	position     Position
	flexShrink   float32

	// Children in layout order
	childrenInLayoutOrder []RenderableIface

	// Hooks
	renderBefore func(*Renderable, *NativeBuffer, float64)
	renderAfter  func(*Renderable, *NativeBuffer, float64)
	onUpdate     func()
	onResize     func()
	onLayoutResize func()
	onLifecyclePass func()

	// Event listeners
	onMouse       func(*MouseEvent)
	onMouseDown   func(*MouseEvent)
	onMouseUp     func(*MouseEvent)
	onMouseMove   func(*MouseEvent)
	onMouseDrag   func(*MouseEvent)
	onMouseDragEnd func(*MouseEvent)
	onMouseDrop   func(*MouseEvent)
	onMouseOver   func(*MouseEvent)
	onMouseOut    func(*MouseEvent)
	onMouseScroll func(*MouseEvent)
	onPaste       func(*PasteEvent)
	onKeyDown     func(*KeyEvent)

	// Focus/keyboard
	keypressHandler func(*KeyEvent)
	pasteHandler    func(*PasteEvent)

	// Animation
	animationFPS         *float64
	stopAnimationTick    func()
	animationDrawnFrame  int
	liveCountChanged     func(previous int)

	// Frame buffer for buffered renderables
	frameBuffer *NativeBuffer

	// Native integration
	nativeIntegration NativeRenderableIntegration
}

// NewRenderable creates a new base Renderable.
func NewRenderable(ctx *RenderContext, id string, opts RenderableOptions[Renderable]) *Renderable {
	r := &Renderable{
		id:         id,
		ctx:        ctx,
		options:    opts,
		width:      0,
		height:     0,
		x:          0,
		y:          0,
		zIndex:     0,
		visible:    true,
		buffered:   false,
		live:       false,
		liveCount:  0,
		opacity:    1.0,
		focusable:  false,
		focused:    false,
		positionType: "relative",
		overflow:   "visible",
		flexShrink: 1,
	}

	if opts.ID != "" {
		r.id = opts.ID
	}
	if opts.Width != nil {
		r.width = 0 // will be set via Yoga
	}
	if opts.Height != nil {
		r.height = 0
	}
	if opts.ZIndex != nil {
		r.zIndex = *opts.ZIndex
	}
	if opts.Visible != nil {
		r.visible = *opts.Visible
	}
	if opts.Buffered != nil {
		r.buffered = *opts.Buffered
	}
	if opts.Live != nil {
		r.live = *opts.Live
	}
	if opts.Opacity != nil {
		r.opacity = *opts.Opacity
	}
	if opts.AnimationFPS != nil {
		r.animationFPS = opts.AnimationFPS
	}

	// Set defaults from options
	if opts.RenderBefore != nil {
		r.renderBefore = func(r *Renderable, b *NativeBuffer, dt float64) { opts.RenderBefore(r, b, dt) }
	}
	if opts.RenderAfter != nil {
		r.renderAfter = func(r *Renderable, b *NativeBuffer, dt float64) { opts.RenderAfter(r, b, dt) }
	}
	if opts.OnMouse != nil {
		r.onMouse = func(e *MouseEvent) { opts.OnMouse(r, e) }
	}
	if opts.OnMouseDown != nil {
		r.onMouseDown = func(e *MouseEvent) { opts.OnMouseDown(r, e) }
	}
	if opts.OnMouseUp != nil {
		r.onMouseUp = func(e *MouseEvent) { opts.OnMouseUp(r, e) }
	}
	if opts.OnMouseMove != nil {
		r.onMouseMove = func(e *MouseEvent) { opts.OnMouseMove(r, e) }
	}
	if opts.OnMouseDrag != nil {
		r.onMouseDrag = func(e *MouseEvent) { opts.OnMouseDrag(r, e) }
	}
	if opts.OnMouseDragEnd != nil {
		r.onMouseDragEnd = func(e *MouseEvent) { opts.OnMouseDragEnd(r, e) }
	}
	if opts.OnMouseDrop != nil {
		r.onMouseDrop = func(e *MouseEvent) { opts.OnMouseDrop(r, e) }
	}
	if opts.OnMouseOver != nil {
		r.onMouseOver = func(e *MouseEvent) { opts.OnMouseOver(r, e) }
	}
	if opts.OnMouseOut != nil {
		r.onMouseOut = func(e *MouseEvent) { opts.OnMouseOut(r, e) }
	}
	if opts.OnMouseScroll != nil {
		r.onMouseScroll = func(e *MouseEvent) { opts.OnMouseScroll(r, e) }
	}
	if opts.OnPaste != nil {
		r.onPaste = func(e *PasteEvent) { opts.OnPaste(r, e) }
	}
	if opts.OnKeyDown != nil {
		r.onKeyDown = func(k *KeyEvent) { opts.OnKeyDown(k) }
	}
	// OnSizeChange handled by RenderableOptions in types.go if present

	// Create native scene node
	if ctx != nil && ctx.NativeScene != nil {
		node, err := ctx.NativeScene.CreateNode(r, native.SceneNodeKindCustom)
		if err == nil {
			r.SetNativeNode(node)
		}
		ctx.NativeScene.RegisterLifecyclePass(r)
		r.SetNativeScene(ctx.NativeScene, native.SceneNodeKindCustom)
	}

	// Create Yoga layout node if native scene with Yoga engine is available.
	if ctx != nil && ctx.NativeScene != nil {
		yoga := ctx.NativeScene.GetYoga()
		if yoga != nil {
			r.yogaNode = yoga.CreateNode()
			// Apply default flex shrink of 1 (Yoga default), matching Renderable default.
			r.yogaNode.SetFlexShrink(1)
		}
	}

	return r
}

// GetYogaNode returns the Yoga layout node for this renderable.
func (r *Renderable) GetYogaNode() *native.YogaNode {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.yogaNode
}

// ID returns the renderable ID.
func (r *Renderable) ID() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.id
}

// Context returns the render context.
func (r *Renderable) Context() *RenderContext {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.ctx
}

// AddChild adds a child renderable.
func (r *Renderable) AddChild(child RenderableIface) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if cr, ok := child.(*Renderable); ok {
		cr.parent = r
	}
	r.children = append(r.children, child)
	r.childrenInLayoutOrder = append(r.childrenInLayoutOrder, child)
	r.markDirty()

	// Also add child to Yoga node parent for layout integration.
	if r.yogaNode != nil {
		if cr, ok := child.(*Renderable); ok && cr.yogaNode != nil {
			r.yogaNode.AddChild(cr.yogaNode)
		}
	}
}

// RenderableIface is implemented by all renderable types.
type RenderableIface interface {
	RenderSelf(*NativeBuffer, float64)
}

// RemoveChild removes a child renderable.
func (r *Renderable) RemoveChild(child RenderableIface) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, c := range r.children {
		if c == child {
			r.children = append(r.children[:i], r.children[i+1:]...)
			if cr, ok := c.(*Renderable); ok {
				cr.parent = nil
			}
			// Remove from layout order
			for j, lc := range r.childrenInLayoutOrder {
				if lc == child {
					r.childrenInLayoutOrder = append(r.childrenInLayoutOrder[:j], r.childrenInLayoutOrder[j+1:]...)
					break
				}
			}
			r.markDirty()

			// Also remove from Yoga node parent.
			if r.yogaNode != nil {
				if cr, ok := child.(*Renderable); ok && cr.yogaNode != nil {
					r.yogaNode.RemoveChild(cr.yogaNode)
				}
			}
			return
		}
	}
}

// Children returns the child renderables.
func (r *Renderable) Children() []RenderableIface {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]RenderableIface, len(r.children))
	copy(result, r.children)
	return result
}

// Parent returns the parent renderable.
func (r *Renderable) Parent() *Renderable {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.parent
}

// GetX returns the renderable X position.
func (r *Renderable) GetX() float32 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.x
}

// GetY returns the renderable Y position.
func (r *Renderable) GetY() float32 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.y
}

// GetWidth returns the renderable width.
func (r *Renderable) GetWidth() float32 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.width
}

// GetHeight returns the renderable height.
func (r *Renderable) GetHeight() float32 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.height
}

// SetWidth sets the renderable width and propagates to Yoga node.
func (r *Renderable) SetWidth(w float32) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.width = w
	if r.yogaNode != nil {
		r.yogaNode.SetWidth(w)
	}
}

// SetHeight sets the renderable height and propagates to Yoga node.
func (r *Renderable) SetHeight(h float32) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.height = h
	if r.yogaNode != nil {
		r.yogaNode.SetHeight(h)
	}
}

// SetPosition sets the renderable position.
func (r *Renderable) SetPosition(x, y float32) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.x = x
	r.y = y
}

// SetZIndex sets the z-index.
func (r *Renderable) SetZIndex(z int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.zIndex = z
}

// SetVisible sets visibility.
func (r *Renderable) SetVisible(v bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.visible = v
}

// SetOpacity sets opacity.
func (r *Renderable) SetOpacity(o float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.opacity = o
}

// IsVisible returns visibility.
func (r *Renderable) IsVisible() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.visible
}

// IsDestroyed returns whether the renderable is destroyed.
func (r *Renderable) IsDestroyed() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.destroyed
}

// Destroy destroys the renderable.
func (r *Renderable) Destroy() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.destroyed {
		return
	}
	r.destroyed = true
	r.children = nil
	r.childrenInLayoutOrder = nil
	r.parent = nil

	// Free Yoga node.
	if r.yogaNode != nil {
		r.yogaFreed = true
	}

	// Destroy native node if exists
	if r.nativeScene != nil && r.nodeHandle != nil {
		r.nativeScene.DestroyNode(r)
	}
}

// RequestRender requests a re-render.
func (r *Renderable) RequestRender() {
	if r.ctx != nil && r.ctx.RequestRender != nil {
		r.ctx.RequestRender()
	}
}

// RenderSelf renders the renderable into the buffer. Override in subclasses.
func (r *Renderable) RenderSelf(buffer *NativeBuffer, deltaTime float64) {
}

// RenderBefore runs before painting.
func (r *Renderable) RenderBefore(buffer *NativeBuffer, deltaTime float64) {
	if r.renderBefore != nil {
		r.renderBefore(r, buffer, deltaTime)
	}
}

// RenderAfter runs after painting.
func (r *Renderable) RenderAfter(buffer *NativeBuffer, deltaTime float64) {
	if r.renderAfter != nil {
		r.renderAfter(r, buffer, deltaTime)
	}
}

// OnUpdate is called on update.
func (r *Renderable) OnUpdate() {
	if r.onUpdate != nil {
		r.onUpdate()
	}
}

// OnResize is called on resize.
func (r *Renderable) OnResize() {
	if r.onResize != nil {
		r.onResize()
	}
}

// OnLifecyclePass is called on lifecycle pass.
func (r *Renderable) OnLifecyclePass() {
	if r.onLifecyclePass != nil {
		r.onLifecyclePass()
	}
}

func (r *Renderable) String() string {
	return fmt.Sprintf("Renderable{id=%s, visible=%v, zIndex=%d}", r.id, r.visible, r.zIndex)
}

// markDirty marks the renderable as needing re-render.
func (r *Renderable) markDirty() {
	r.dirty = true
}

// IsDirty returns whether the renderable is dirty.
func (r *Renderable) IsDirty() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.dirty
}

// setDirty sets the dirty flag.
func (r *Renderable) setDirty(dirty bool) {
	r.mu.Lock()
	r.dirty = dirty
	r.mu.Unlock()
}

// RootRenderable is the root renderable of the tree.
type RootRenderable struct {
	*Renderable
}

// NewRootRenderable creates a new root renderable.
func NewRootRenderable(ctx *RenderContext, id string) *RootRenderable {
	return &RootRenderable{
		Renderable: NewRenderable(ctx, id, RenderableOptions[Renderable]{}),
	}
}

// SetNativeScene attaches a native scene to this renderable.
func (r *Renderable) SetNativeScene(ns *NativeScene, kind native.SceneNodeKind) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nativeScene = ns
	r.nativeKind = kind
}

// GetNativeScene returns the native scene.
func (r *Renderable) GetNativeScene() *NativeScene {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.nativeScene
}

// CreateNativeNode creates a native scene node for this renderable.
func (r *Renderable) CreateNativeNode(kind native.SceneNodeKind) (*SceneNodeHandle, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.nativeScene == nil {
		return nil, fmt.Errorf("no native scene attached")
	}
	handle, err := r.nativeScene.CreateNode(r, kind)
	if err != nil {
		return nil, err
	}
	r.nodeHandle = handle
	r.nativeKind = kind
	return handle, nil
}

// GetNativeNode returns the native scene node handle.
func (r *Renderable) GetNativeNode() *SceneNodeHandle {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.nodeHandle
}

// DestroyNativeNode destroys the native scene node.
func (r *Renderable) DestroyNativeNode() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.nativeScene == nil {
		return nil
	}
	return r.nativeScene.DestroyNode(r)
}

// SetNativeNode sets the native scene node handle.
func (r *Renderable) SetNativeNode(handle *SceneNodeHandle) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nodeHandle = handle
}

// Focus focuses the renderable.
func (r *Renderable) Focus() {
	r.mu.Lock()
	if r.destroyed || r.focused || !r.focusable {
		r.mu.Unlock()
		return
	}
	r.focused = true
	r.mu.Unlock()

	if r.ctx != nil && r.ctx.NativeScene != nil && r.nodeHandle != nil {
		r.ctx.NativeScene.SetFocus(r, true)
	}

	if r.ctx != nil && r.ctx.FocusRenderable != nil {
		r.ctx.FocusRenderable(r)
	}

	r.RequestRender()
}

// Blur blurs the renderable.
func (r *Renderable) Blur() {
	r.mu.Lock()
	if !r.focused {
		r.mu.Unlock()
		return
	}
	r.focused = false
	r.mu.Unlock()

	if r.ctx != nil && r.ctx.NativeScene != nil && r.nodeHandle != nil {
		r.ctx.NativeScene.SetFocus(r, false)
	}

	if r.ctx != nil && r.ctx.BlurRenderable != nil {
		r.ctx.BlurRenderable(r)
	}

	r.RequestRender()
}

// IsFocused returns whether the renderable is focused.
func (r *Renderable) IsFocused() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.focused
}

// SetFocusable sets whether the renderable can be focused.
func (r *Renderable) SetFocusable(f bool) {
	r.mu.Lock()
	r.focusable = f
	r.mu.Unlock()
}

// IsFocusable returns whether the renderable can be focused.
func (r *Renderable) IsFocusable() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.focusable
}

// HasSelection returns whether the renderable has a selection.
func (r *Renderable) HasSelection() bool {
	return false
}

// OnSelectionChanged is called when selection changes.
func (r *Renderable) OnSelectionChanged(start, end int) bool {
	return false
}

// GetSelectedText returns the selected text.
func (r *Renderable) GetSelectedText() string {
	return ""
}

// ShouldStartSelection checks if selection should start at (x, y).
func (r *Renderable) ShouldStartSelection(x, y int) bool {
	return false
}

// SetAnimationFPS sets the animation FPS.
func (r *Renderable) SetAnimationFPS(fps float64) {
	r.mu.Lock()
	r.animationFPS = &fps
	r.mu.Unlock()
	r.RequestRender()
}

// GetAnimationFPS returns the animation FPS.
func (r *Renderable) GetAnimationFPS() *float64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.animationFPS
}

// GetFlexDirection returns the flex direction.
func (r *Renderable) GetFlexDirection() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.positionType
}

// SetFlexDirection sets the flex direction on the renderable and its Yoga node.
func (r *Renderable) SetFlexDirection(dir string) {
	r.mu.Lock()
	r.positionType = dir
	r.mu.Unlock()
	if r.yogaNode != nil {
		r.yogaNode.SetFlexDirection(parseFlexDirection(dir))
	}
	r.RequestRender()
}

// GetOverflow returns the overflow mode.
func (r *Renderable) GetOverflow() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.overflow
}

// SetOverflow sets the overflow mode.
func (r *Renderable) SetOverflow(overflow string) {
	r.mu.Lock()
	r.overflow = overflow
	r.mu.Unlock()
	r.RequestRender()
}

// GetPosition returns the position.
func (r *Renderable) GetPosition() Position {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.position
}

// SetPosition2 sets the position (overloaded).
func (r *Renderable) SetPosition2(pos Position) {
	r.mu.Lock()
	r.position = pos
	r.mu.Unlock()
	r.RequestRender()
}

// GetComputedLayout returns the computed layout from native scene / Yoga.
func (r *Renderable) GetComputedLayout() Layout {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.nativeScene != nil && r.nodeHandle != nil {
		return r.nativeScene.GetLayout(r.nodeHandle)
	}
	return Layout{Width: r.width, Height: r.height}
}

// GetScreenX returns the screen X position.
func (r *Renderable) GetScreenX() float32 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.x
}

// GetScreenY returns the screen Y position.
func (r *Renderable) GetScreenY() float32 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.y
}

// Layout represents computed layout geometry.
type Layout struct {
	Left, Top, Right, Bottom, Width, Height float32
}

// ---- Yoga layout setter implementations ----

// SetFlexDirection sets the flex direction on the Yoga node.
func (r *Renderable) SetYogaFlexDirection(dir native.FlexDirection) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.yogaNode != nil {
		r.yogaNode.SetFlexDirection(dir)
	}
	r.RequestRender()
}

// SetJustifyContent sets the main-axis distribution on the Yoga node.
func (r *Renderable) SetJustifyContent(jc string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.yogaNode != nil {
		r.yogaNode.SetJustifyContent(parseJustifyContent(jc))
	}
	r.RequestRender()
}

// SetAlignItems sets the default cross-axis alignment on the Yoga node.
func (r *Renderable) SetAlignItems(ai string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.yogaNode != nil {
		r.yogaNode.SetAlignItems(parseAlignItems(ai))
	}
	r.RequestRender()
}

// SetAlignSelf sets the per-node cross-axis alignment override on the Yoga node.
func (r *Renderable) SetAlignSelf(as string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.yogaNode != nil {
		r.yogaNode.SetAlignSelf(parseAlignSelf(as))
	}
	r.RequestRender()
}

// SetFlexWrap sets the wrapping behavior on the Yoga node.
func (r *Renderable) SetFlexWrap(wrap string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.yogaNode != nil {
		r.yogaNode.SetFlexWrap(parseFlexWrap(wrap))
	}
	r.RequestRender()
}

// SetFlexGrow sets the flex grow factor on the Yoga node.
func (r *Renderable) SetFlexGrow(f float32) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.yogaNode != nil {
		r.yogaNode.SetFlexGrow(f)
	}
	r.RequestRender()
}

// SetFlexShrink sets the flex shrink factor on the Yoga node.
func (r *Renderable) SetFlexShrink(f float32) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.flexShrink = f
	if r.yogaNode != nil {
		r.yogaNode.SetFlexShrink(f)
	}
	r.RequestRender()
}

// SetFlexBasis sets the flex basis (preferred main size) on the Yoga node.
func (r *Renderable) SetFlexBasis(b float32) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.yogaNode != nil {
		r.yogaNode.SetFlexBasis(b)
	}
	r.RequestRender()
}

// SetMinMax sets min and max constraints on the Yoga node.
func (r *Renderable) SetMinMax(minW, minH, maxW, maxH float32) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.yogaNode != nil {
		r.yogaNode.SetMinMax(minW, minH, maxW, maxH)
	}
	r.RequestRender()
}

// SetMargin sets all margins (top, right, bottom, left) on the Yoga node.
func (r *Renderable) SetMargin(t, rig, b, l float32) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.yogaNode != nil {
		r.yogaNode.SetMargin(t, rig, b, l)
	}
	r.RequestRender()
}

// SetPadding sets all padding (top, right, bottom, left) on the Yoga node.
func (r *Renderable) SetPadding(t, rig, b, l float32) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.yogaNode != nil {
		r.yogaNode.SetPadding(t, rig, b, l)
	}
	r.RequestRender()
}

// SetGap sets row and column gaps on the Yoga node.
func (r *Renderable) SetGap(row, col float32) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.yogaNode != nil {
		r.yogaNode.SetGap(row, col)
	}
	r.RequestRender()
}

// ---- Yoga string-to-enum parsing helpers ----

func parseFlexDirection(dir string) native.FlexDirection {
	switch dir {
	case "row":
		return native.FlexDirectionRow
	case "row-reverse":
		return native.FlexDirectionRowReverse
	case "column-reverse":
		return native.FlexDirectionColumnReverse
	case "column", "":
		return native.FlexDirectionColumn
	default:
		return native.FlexDirectionColumn
	}
}

func parseJustifyContent(jc string) native.JustifyContent {
	switch jc {
	case "center":
		return native.JustifyCenter
	case "flex-end":
		return native.JustifyFlexEnd
	case "space-between":
		return native.JustifySpaceBetween
	case "space-around":
		return native.JustifySpaceAround
	case "space-evenly":
		return native.JustifySpaceEvenly
	case "flex-start", "":
		return native.JustifyFlexStart
	default:
		return native.JustifyFlexStart
	}
}

func parseAlignItems(ai string) native.AlignItems {
	switch ai {
	case "flex-start":
		return native.AlignFlexStart
	case "center":
		return native.AlignCenter
	case "flex-end":
		return native.AlignFlexEnd
	case "stretch":
		return native.AlignStretch
	case "baseline":
		return native.AlignBaseline
	case "auto", "":
		return native.AlignAuto
	default:
		return native.AlignStretch
	}
}

func parseAlignSelf(as string) native.AlignSelf {
	switch as {
	case "flex-start":
		return native.AlignSelfFlexStart
	case "center":
		return native.AlignSelfCenter
	case "flex-end":
		return native.AlignSelfFlexEnd
	case "stretch":
		return native.AlignSelfStretch
	case "baseline":
		return native.AlignSelfBaseline
	case "auto", "":
		return native.AlignSelfAuto
	default:
		return native.AlignSelfAuto
	}
}

func parseFlexWrap(wrap string) native.FlexWrap {
	switch wrap {
	case "wrap":
		return native.FlexWrapWrap
	case "wrap-reverse":
		return native.FlexWrapReverse
	case "nowrap", "":
		return native.FlexWrapNone
	default:
		return native.FlexWrapNone
	}
}

var _ = fmt.Sprintf
