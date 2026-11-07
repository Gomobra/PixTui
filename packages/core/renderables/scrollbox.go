package renderables

import (
	"math"
	"sync"

	"github.com/Gomobra/PixTui/packages/core"
	"github.com/Gomobra/PixTui/packages/native"
)

// ScrollBarOrientation identifies the scrollbar axis.
type ScrollBarOrientation int

const (
	ScrollBarVertical   ScrollBarOrientation = 1
	ScrollBarHorizontal ScrollBarOrientation = 0
)

// ScrollUnit is the unit for scroll deltas.
type ScrollUnit string

const (
	ScrollUnitAbsolute ScrollUnit = "absolute"
	ScrollUnitViewport ScrollUnit = "viewport"
	ScrollUnitLine     ScrollUnit = "line"
)

// ScrollAcceleration computes scroll steps from wheel events.
type ScrollAcceleration interface {
	Tick(now int64) float64
	Reset()
}

// LinearScrollAccel is a simple linear scroll acceleration.
type LinearScrollAccel struct{}

func (LinearScrollAccel) Tick(now int64) float64 { return 1.0 }
func (LinearScrollAccel) Reset()                  {}

// ScrollBoxOptions contains options for ScrollBoxRenderable.
type ScrollBoxOptions struct {
	core.RenderableOptions[core.Renderable]
	RootOptions            *ScrollBoxOptions
	WrapperOptions         *BoxOptions
	ViewportOptions        *BoxOptions
	ContentOptions         *BoxOptions
	ScrollbarOptions       *ScrollBarPartOptions
	VerticalScrollbarOptions *ScrollBarPartOptions
	HorizontalScrollbarOptions *ScrollBarPartOptions
	StickyScroll           bool
	StickyStart            *string
	ScrollX                bool
	ScrollY                bool
	ScrollAcceleration     ScrollAcceleration
	ViewportCulling        bool
}

// ScrollBarPartOptions configure a scrollbar part.
type ScrollBarPartOptions struct {
	Visible        *bool
	ShowArrows     *bool
	TrackOptions   *ScrollBarTrackOptions
	ArrowOptions   *ScrollBarArrowOptions
	OnChange       func(position int)
	Width          *int
	Height         *int
	PaddingLeft    *int
	PaddingTop     *int
	Position       *string
	ScrollStep     *float64
}

// ScrollBarTrackOptions configure scrollbar track colors.
type ScrollBarTrackOptions struct {
	ForegroundColor *core.RGBA
	BackgroundColor *core.RGBA
}

// ScrollBarArrowOptions configure scrollbar arrows.
type ScrollBarArrowOptions struct {
	Attributes       uint32
	ForegroundColor  *core.RGBA
	BackgroundColor  *core.RGBA
	ArrowChars       map[string]string
}

// ScrollBoxWrapper is the inner rect minus the vertical bar gutter.
type ScrollBoxWrapper struct {
	box             *ScrollBoxRenderable
	backgroundColor *core.RGBA
}

func (w *ScrollBoxWrapper) BackgroundColor() *core.RGBA { return w.backgroundColor }
func (w *ScrollBoxWrapper) SetBackgroundColor(c *core.RGBA) { w.backgroundColor = c; w.box.applyScrollOptions() }

// ScrollBoxViewport is the culling window with padding in cells.
type ScrollBoxViewport struct {
	box           *ScrollBoxRenderable
	backgroundColor *core.RGBA
	padding       [4]int
	onSizeChange  func()
}

func (v *ScrollBoxViewport) BackgroundColor() *core.RGBA { return v.backgroundColor }
func (v *ScrollBoxViewport) SetBackgroundColor(c *core.RGBA) { v.backgroundColor = c; v.box.applyScrollOptions() }
func (v *ScrollBoxViewport) Padding() [4]int { return v.padding }
func (v *ScrollBoxViewport) SetPadding(left, top, right, bottom int) {
	v.padding = [4]int{left, top, right, bottom}
	v.box.applyScrollOptions()
}
func (v *ScrollBoxViewport) OnSizeChange() func() { return v.onSizeChange }
func (v *ScrollBoxViewport) SetOnSizeChange(f func()) { v.onSizeChange = f; v.box.refreshScrollHook() }
func (v *ScrollBoxViewport) EmitResize() { if v.onSizeChange != nil { v.onSizeChange() } }

// ScrollBoxContent is the hidden content node that moves with the scroll offset.
type ScrollBoxContent struct {
	box              *ScrollBoxRenderable
	backgroundColor  *core.RGBA
	styles           map[string]interface{}
	viewportCulling  bool
	onSizeChange     func()
}

func (c *ScrollBoxContent) BackgroundColor() *core.RGBA { return c.backgroundColor }
func (c *ScrollBoxContent) SetBackgroundColor(bg *core.RGBA) { c.backgroundColor = bg; c.box.applyScrollOptions() }
func (c *ScrollBoxContent) ViewportCulling() bool { return c.viewportCulling }
func (c *ScrollBoxContent) SetViewportCulling(v bool) { c.viewportCulling = v; c.box._viewportCulling = v }
func (c *ScrollBoxContent) OnSizeChange() func() { return c.onSizeChange }
func (c *ScrollBoxContent) SetOnSizeChange(f func()) { c.onSizeChange = f; c.box.refreshScrollHook() }
func (c *ScrollBoxContent) EmitResize() { if c.onSizeChange != nil { c.onSizeChange() } }
func (c *ScrollBoxContent) SetStyle(key string, value interface{}) {
	if c.styles == nil { c.styles = make(map[string]interface{}) }
	if value == nil { delete(c.styles, key) } else { c.styles[key] = value }
	c.box.applyScrollOptions()
}

// ScrollBoxBar is a scrollbar: track, thumb and arrows.
type ScrollBoxBar struct {
	box               *ScrollBoxRenderable
	orientation       ScrollBarOrientation
	axis              int
	visible           bool
	showArrows        bool
	trackForeground   *core.RGBA
	trackBackground   *core.RGBA
	thickness         int
	gap               int
	overlay           bool
	onChange          func(position int)
	arrowOptions      ScrollBarArrowOptions
	policy            string
	scrollStep        float64
}

func (b *ScrollBoxBar) Axis() int { return b.axis }
func (b *ScrollBoxBar) Orientation() ScrollBarOrientation { return b.orientation }
func (b *ScrollBoxBar) Visible() bool { return b.visible }
func (b *ScrollBoxBar) SetVisible(v bool) { b.visible = v; b.box.applyScrollOptions() }
func (b *ScrollBoxBar) ShowArrows() bool { return b.showArrows }
func (b *ScrollBoxBar) SetShowArrows(v bool) { b.showArrows = v; b.box.applyScrollArrows(b.axis) }
func (b *ScrollBoxBar) OnChange() func(int) { return b.onChange }
func (b *ScrollBoxBar) SetOnChange(f func(int)) { b.onChange = f; b.box.refreshScrollHook() }
func (b *ScrollBoxBar) EmitChange(position int) { if b.onChange != nil { b.onChange(position) } }
func (b *ScrollBoxBar) ScrollPosition() int {
	state := b.box.scrollState()
	return int(state.Offset[b.axis])
}
func (b *ScrollBoxBar) SetScrollPosition(v int) { b.box.scrollAxis(b.axis, v, false) }
func (b *ScrollBoxBar) ScrollSize() int {
	state := b.box.scrollState()
	if b.axis == 1 { return int(state.Content[3]) }
	return int(state.Content[2])
}
func (b *ScrollBoxBar) ViewportSize() int {
	state := b.box.scrollState()
	if b.axis == 1 { return int(state.Viewport[3]) }
	return int(state.Viewport[2])
}
func (b *ScrollBoxBar) ScrollBy(delta float64, unit ScrollUnit) {
	amount := delta
	if unit == ScrollUnitViewport {
		vs := b.ViewportSize()
		if vs > 0 { amount = delta * float64(vs) / 2.0 }
	}
	b.box.scrollAxis(b.axis, int(amount), true)
}

// BarPress represents an active bar press (arrow or track).
type BarPress struct {
	axis   int
	offset int
	arrow  int
}

// ScrollBoxRenderable is a scrollable container whose scroll state is native.
type ScrollBoxRenderable struct {
	*core.Renderable
	options ScrollBoxOptions

	wrapper       *ScrollBoxWrapper
	viewport      *ScrollBoxViewport
	content       *ScrollBoxContent
	horizontalBar *ScrollBoxBar
	verticalBar   *ScrollBoxBar

	_stickyScroll    bool
	_stickyStart     *string
	scrollAccel      ScrollAcceleration
	axes             [2]bool
	_viewportCulling bool
	state            core.NativeSceneScrollState
	stateRevision    uint64
	scrollHook       bool
	configured       bool
	barPress         *BarPress

	autoScrollMouseX     int
	autoScrollMouseY     int
	autoScrollThresholdV int
	autoScrollThresholdH int
	autoScrollSpeedSlow  int
	autoScrollSpeedMed   int
	autoScrollSpeedFast  int
	isAutoScrolling      bool
	cachedAutoScrollSpeed int
	autoScrollAccumX     float64
	autoScrollAccumY     float64
	scrollAccumX         float64
	scrollAccumY         float64

	layoutCallbacks []struct {
		frame    uint64
		callback func()
	}
	anchorToken interface{}
	mu          sync.RWMutex
}

const (
	autoScrollThresholdV = 3
	autoScrollThresholdH = 3
	autoScrollSpeedSlow  = 6
	autoScrollSpeedMed   = 36
	autoScrollSpeedFast  = 72
)

// NewScrollBoxRenderable creates a new ScrollBoxRenderable.
func NewScrollBoxRenderable(ctx *core.RenderContext, opts ScrollBoxOptions) *ScrollBoxRenderable {
	rootOpts := core.RenderableOptions[core.Renderable]{
		Width:           opts.Width,
		Height:          opts.Height,
		Visible:         opts.Visible,
		ZIndex:          opts.ZIndex,
		ForegroundColor: opts.ForegroundColor,
		BackgroundColor: opts.BackgroundColor,
		Attributes:      opts.Attributes,
		RenderBefore:    opts.RenderBefore,
		RenderAfter:     opts.RenderAfter,
		OnMouse:         opts.OnMouse,
		OnMouseDown:     opts.OnMouseDown,
		OnMouseUp:       opts.OnMouseUp,
		OnMouseMove:     opts.OnMouseMove,
		OnMouseDrag:     opts.OnMouseDrag,
		OnMouseDragEnd:  opts.OnMouseDragEnd,
		OnMouseDrop:     opts.OnMouseDrop,
		OnMouseOver:     opts.OnMouseOver,
		OnMouseOut:      opts.OnMouseOut,
		OnMouseScroll:   opts.OnMouseScroll,
		OnPaste:         opts.OnPaste,
		OnKeyDown:       opts.OnKeyDown,
	}
	base := core.NewRenderable(ctx, opts.ID, rootOpts)

	r := &ScrollBoxRenderable{
		Renderable:        base,
		options:           opts,
		_stickyScroll:     opts.StickyScroll,
		_stickyStart:      opts.StickyStart,
		scrollAccel:       opts.ScrollAcceleration,
		_viewportCulling:  opts.ViewportCulling,
		axes:              [2]bool{opts.ScrollX, opts.ScrollY},
		state:             core.NativeSceneScrollState{},
		wrapper:           &ScrollBoxWrapper{box: nil},
		viewport:          &ScrollBoxViewport{box: nil},
		content:           &ScrollBoxContent{box: nil},
		horizontalBar:     &ScrollBoxBar{orientation: ScrollBarHorizontal, axis: 0, thickness: 1},
		verticalBar:       &ScrollBoxBar{orientation: ScrollBarVertical, axis: 1, thickness: 1},
		autoScrollThresholdV: autoScrollThresholdV,
		autoScrollThresholdH: autoScrollThresholdH,
		autoScrollSpeedSlow:  autoScrollSpeedSlow,
		autoScrollSpeedMed:   autoScrollSpeedMed,
		autoScrollSpeedFast:  autoScrollSpeedFast,
	}
	r.wrapper.box = r
	r.viewport.box = r
	r.content.box = r
	r.horizontalBar.box = r
	r.verticalBar.box = r

	if r.scrollAccel == nil {
		r.scrollAccel = LinearScrollAccel{}
	}
	if !r._viewportCulling {
		r._viewportCulling = true
	}

	if opts.ScrollbarOptions != nil {
		r.applyBarOptions(r.verticalBar, opts.ScrollbarOptions)
		r.applyBarOptions(r.horizontalBar, opts.ScrollbarOptions)
	}
	if opts.VerticalScrollbarOptions != nil {
		r.applyBarOptions(r.verticalBar, opts.VerticalScrollbarOptions)
	}
	if opts.HorizontalScrollbarOptions != nil {
		r.applyBarOptions(r.horizontalBar, opts.HorizontalScrollbarOptions)
	}

	r.configured = true
	r.applyScrollOptions()
	r.applyScrollArrows(0)
	r.applyScrollArrows(1)
	r.refreshScrollHook()

	return r
}

func (r *ScrollBoxRenderable) applyBarOptions(bar *ScrollBoxBar, o *ScrollBarPartOptions) {
	if o == nil { return }
	if o.Visible != nil { bar.visible = *o.Visible }
	if o.ShowArrows != nil { bar.showArrows = *o.ShowArrows }
	if o.TrackOptions != nil {
		if o.TrackOptions.ForegroundColor != nil { bar.trackForeground = o.TrackOptions.ForegroundColor }
		if o.TrackOptions.BackgroundColor != nil { bar.trackBackground = o.TrackOptions.BackgroundColor }
	}
	if o.ArrowOptions != nil {
		bar.arrowOptions = *o.ArrowOptions
	}
	if o.OnChange != nil { bar.onChange = o.OnChange }
	if o.Width != nil && bar.axis == 0 { bar.thickness = *o.Width }
	if o.Height != nil && bar.axis == 1 { bar.thickness = *o.Height }
	if o.PaddingLeft != nil && bar.axis == 1 { bar.gap = *o.PaddingLeft }
	if o.PaddingTop != nil && bar.axis == 0 { bar.gap = *o.PaddingTop }
	if o.Position != nil { bar.overlay = *o.Position == "absolute" }
	if o.ScrollStep != nil { bar.scrollStep = *o.ScrollStep }
}

func (r *ScrollBoxRenderable) applyScrollOptions() {
	if !r.configured { return }
	ns := r.GetNativeScene()
	if ns == nil { return }
	padding := r.viewport.Padding()
	_ = ns.SetScroll(r.Renderable, core.NativeSceneScrollOptions{
		Axes:    0,
		Follow:  0,
		Culling: r._viewportCulling,
		Padding: [4]float32{float32(padding[0]), float32(padding[1]), float32(padding[2]), float32(padding[3])},
	})
	r.RequestRender()
}

func (r *ScrollBoxRenderable) applyScrollArrows(axis int) {
	if !r.configured { return }
	r.RequestRender()
}

func (r *ScrollBoxRenderable) StickyScroll() bool { return r._stickyScroll }
func (r *ScrollBoxRenderable) SetStickyScroll(v bool) { r._stickyScroll = v; r.applyScrollOptions() }
func (r *ScrollBoxRenderable) StickyStart() *string { return r._stickyStart }
func (r *ScrollBoxRenderable) SetStickyStart(v *string) { r._stickyStart = v; r.applyScrollOptions() }
func (r *ScrollBoxRenderable) ScrollX() bool { return r.axes[0] }
func (r *ScrollBoxRenderable) SetScrollX(v bool) { r.axes[0] = v; r.applyScrollOptions() }
func (r *ScrollBoxRenderable) ScrollY() bool { return r.axes[1] }
func (r *ScrollBoxRenderable) SetScrollY(v bool) { r.axes[1] = v; r.applyScrollOptions() }

func (r *ScrollBoxRenderable) ScrollTop() int { return int(r.scrollState().Offset[1]) }
func (r *ScrollBoxRenderable) SetScrollTop(v int) { r.scrollAxis(1, v, false) }
func (r *ScrollBoxRenderable) ScrollLeft() int { return int(r.scrollState().Offset[0]) }
func (r *ScrollBoxRenderable) SetScrollLeft(v int) { r.scrollAxis(0, v, false) }
func (r *ScrollBoxRenderable) ScrollWidth() int { return int(r.scrollState().Content[2]) }
func (r *ScrollBoxRenderable) ScrollHeight() int { return int(r.scrollState().Content[3]) }
func (r *ScrollBoxRenderable) ViewportCulling() bool { return r._viewportCulling }
func (r *ScrollBoxRenderable) SetViewportCulling(v bool) { r._viewportCulling = v; r.applyScrollOptions() }

func (r *ScrollBoxRenderable) scrollState() core.NativeSceneScrollState {
	return r.state
}

func (r *ScrollBoxRenderable) scrollAxis(axis, value int, relative bool) {
	var x, y float64
	if axis == 0 { x = float64(value) } else { y = float64(value) }
	relX := axis != 0 || relative
	relY := axis != 1 || relative

	// Compute new offset with clamping
	maxX := int(r.state.Content[2] - r.state.Viewport[2])
	maxY := int(r.state.Content[3] - r.state.Viewport[3])
	if maxX < 0 { maxX = 0 }
	if maxY < 0 { maxY = 0 }

	beforeX, beforeY := int(r.state.Offset[0]), int(r.state.Offset[1])
	newX, newY := beforeX, beforeY

	if axis == 0 {
		if relative {
			newX = beforeX + value
		} else {
			newX = value
		}
		if newX < 0 { newX = 0 }
		if newX > maxX { newX = maxX }
	} else {
		if relative {
			newY = beforeY + value
		} else {
			newY = value
		}
		if newY < 0 { newY = 0 }
		if newY > maxY { newY = maxY }
	}

	// Update state
	r.state.Offset[0] = float64(newX)
	r.state.Offset[1] = float64(newY)
	r.stateRevision++

	// Call native scene ScrollTo for side effects (stub, but keep for integration)
	if ns := r.GetNativeScene(); ns != nil {
		_, _ = ns.ScrollTo(r.Renderable, x, y, relX, relY, false)
	}

	moved := 0
	if newX != beforeX { moved |= 1 }
	if newY != beforeY { moved |= 2 }
	if moved != 0 { r.RequestRender() }
	r.notifyOffsets(r.state, moved)
}

func (r *ScrollBoxRenderable) notifyOffsets(state core.NativeSceneScrollState, moved int) {
	for _, axis := range []int{1, 0} {
		if (moved & (axis + 1)) == 0 { continue }
		if axis == 1 { r.verticalBar.EmitChange(int(state.Offset[1])) } else { r.horizontalBar.EmitChange(int(state.Offset[0])) }
	}
}

func (r *ScrollBoxRenderable) listening() bool {
	return r.verticalBar.onChange != nil ||
		r.horizontalBar.onChange != nil ||
		r.content.onSizeChange != nil ||
		r.viewport.onSizeChange != nil ||
		len(r.layoutCallbacks) > 0
}

func (r *ScrollBoxRenderable) refreshScrollHook(adding ...bool) {
	if !r.configured { return }
	add := false
	if len(adding) > 0 { add = adding[0] } else { add = r.listening() }
	if add == r.scrollHook { return }
	r.scrollHook = add
}

func (r *ScrollBoxRenderable) notifyScroll() {
	state := r.scrollState()
	r.notifyOffsets(state, 0)
}

func (r *ScrollBoxRenderable) FrameID() uint64 {
	ctx := r.Context()
	if ctx != nil { return ctx.FrameID }
	return 0
}

func (r *ScrollBoxRenderable) isDestroyed() bool { return r.IsDestroyed() }

func (r *ScrollBoxRenderable) AfterLayout(callback func()) {
	if r.isDestroyed() { return }
	r.layoutCallbacks = append(r.layoutCallbacks, struct {
		frame    uint64
		callback func()
	}{frame: r.FrameID(), callback: callback})
	r.refreshScrollHook()
	r.RequestRender()
}

func (r *ScrollBoxRenderable) Anchor(node *core.Renderable) func() {
	if r.isDestroyed() { return func() {} }
	token := struct{}{}
	r.anchorToken = token
	return func() {
		if r.anchorToken != token || r.isDestroyed() { return }
		r.anchorToken = nil
	}
}

func (r *ScrollBoxRenderable) PrimaryAxis() string { return "column" }

func (r *ScrollBoxRenderable) OnUpdate(deltaTime float64) {
	r.handleAutoScroll(deltaTime)
}

func (r *ScrollBoxRenderable) ScrollBy(delta interface{}, unit ScrollUnit) {
	switch d := delta.(type) {
	case float64:
		r.verticalBar.ScrollBy(d, unit)
	case int:
		r.verticalBar.ScrollBy(float64(d), unit)
	case map[string]interface{}:
		if y, ok := d["y"]; ok {
			switch vy := y.(type) {
			case float64: r.verticalBar.ScrollBy(vy, unit)
			case int: r.verticalBar.ScrollBy(float64(vy), unit)
			}
		}
		if r.isDestroyed() { return }
		if x, ok := d["x"]; ok {
			switch vx := x.(type) {
			case float64: r.horizontalBar.ScrollBy(vx, unit)
			case int: r.horizontalBar.ScrollBy(float64(vx), unit)
			}
		}
	}
}

func (r *ScrollBoxRenderable) ScrollChildIntoView(child interface{}, options map[string]interface{}) {
	var node *core.Renderable
	switch c := child.(type) {
	case string:
		node = r.findDescendantById(c)
	case *core.Renderable:
		node = c
	}
	if node == nil || r.isDestroyed() { return }
	r.RequestRender()
}

func (r *ScrollBoxRenderable) findDescendantById(id string) *core.Renderable {
	for _, child := range r.Children() {
		cr, ok := child.(*core.Renderable)
		if !ok { continue }
		if cr.ID() == id { return cr }
		if found := findDescendant(cr, id); found != nil { return found }
	}
	return nil
}

func findDescendant(parent *core.Renderable, id string) *core.Renderable {
	for _, child := range parent.Children() {
		cr, ok := child.(*core.Renderable)
		if !ok { continue }
		if cr.ID() == id { return cr }
		if found := findDescendant(cr, id); found != nil { return found }
	}
	return nil
}

func (r *ScrollBoxRenderable) ScrollTo(position interface{}) {
	switch p := position.(type) {
	case float64:
		r.scrollAxis(1, int(p), false)
	case int:
		r.scrollAxis(1, p, false)
	case map[string]interface{}:
		if y, ok := p["y"]; ok {
			switch vy := y.(type) {
			case float64: r.scrollAxis(1, int(vy), false)
			case int: r.scrollAxis(1, vy, false)
			}
		}
		if r.isDestroyed() { return }
		if x, ok := p["x"]; ok {
			switch vx := x.(type) {
			case float64: r.scrollAxis(0, int(vx), false)
			case int: r.scrollAxis(0, vx, false)
			}
		}
	}
}

func (r *ScrollBoxRenderable) scrollByWheel(delta float64, vertical bool) {
	axis := 1
	if !vertical { axis = 0 }
	sign := float64(1)
	if delta < 0 { sign = -1 }
	state := r.scrollState()
	position := state.Offset[axis]
	max := state.Max[axis]
	if sign < 0 && position <= 0 { return }
	if sign > 0 && position >= max { return }

	rem := &r.scrollAccumY
	if axis == 0 { rem = &r.scrollAccumX }
	if *rem*sign < 0 { *rem = 0 }
	now := int64(0)
	*rem += sign * delta * r.scrollAccel.Tick(now)
	step := int(math.Trunc(*rem))
	if step != 0 { r.scrollAxis(axis, int(position)+step, false) }
	kept := float64(r.scrollState().Offset[axis]) - position == float64(step)
	if kept { *rem -= float64(step) } else { *rem = 0 }
}

func (r *ScrollBoxRenderable) RenderSelf(buffer *core.NativeBuffer, deltaTime float64) {
	if r.options.RenderBefore != nil {
		r.options.RenderBefore(r.Renderable, buffer, deltaTime)
	}

	state := r.scrollState()
	offsetX := int(state.Offset[0])
	offsetY := int(state.Offset[1])
	contentW := int(state.Content[2])
	contentH := int(state.Content[3])
	viewportW := int(state.Viewport[2])
	viewportH := int(state.Viewport[3])
	if viewportW == 0 {
		viewportW = int(r.GetWidth())
	}
	if viewportH == 0 {
		viewportH = int(r.GetHeight())
	}

	// Render children with scroll offset and culling
	children := r.Children()
	for _, child := range children {
		cr, ok := child.(interface {
			GetX() float32
			GetY() float32
			GetWidth() float32
			GetHeight() float32
			IsVisible() bool
			SetPosition(float32, float32)
		})
		if !ok || !cr.IsVisible() {
			continue
		}
		childW := int(cr.GetWidth())
		childH := int(cr.GetHeight())
		childX := int(cr.GetX())
		childY := int(cr.GetY())

		// Track content size from children if not explicitly set
		if childX+childW > contentW {
			contentW = childX + childW
		}
		if childY+childH > contentH {
			contentH = childY + childH
		}

		// Compute child position relative to scroll offset
		renderX := childX - offsetX
		renderY := childY - offsetY

		// Culling: skip children completely outside viewport
		if renderX+childW <= 0 || renderX >= viewportW || renderY+childH <= 0 || renderY >= viewportH {
			continue
		}

		// Temporarily adjust child position for rendering into parent buffer
		cr.SetPosition(float32(renderX), float32(renderY))
		child.RenderSelf(buffer, deltaTime)
		cr.SetPosition(float32(childX), float32(childY))
	}

	// Render scrollbar thumb
	r.renderScrollbarThumb(buffer, offsetX, offsetY, viewportW, viewportH, contentW, contentH)

	if r.options.RenderAfter != nil {
		r.options.RenderAfter(r.Renderable, buffer, deltaTime)
	}
}

func (r *ScrollBoxRenderable) renderScrollbarThumb(buffer *core.NativeBuffer, offsetX, offsetY, viewportW, viewportH, contentW, contentH int) {
	// Resolve thumb color from bar options, falling back to renderable defaults.
	thumbFg := r.verticalBar.trackForeground
	if thumbFg == nil {
		thumbFg = r.options.ForegroundColor
	}
	thumbBg := r.verticalBar.trackBackground
	if thumbBg == nil {
		thumbBg = r.options.BackgroundColor
	}

	// Vertical scrollbar on right edge
	if contentH > viewportH {
		thumbH := viewportH * viewportH / contentH
		if thumbH < 1 {
			thumbH = 1
		}
		thumbStartY := int(float64(offsetY) / float64(contentH) * float64(viewportH))
		if thumbStartY < 0 {
			thumbStartY = 0
		}
		thumbEnd := thumbStartY + thumbH
		if thumbEnd > viewportH {
			thumbEnd = viewportH
		}
		x := viewportW - 1
		for y := thumbStartY; y < thumbEnd; y++ {
			if x >= 0 && x < buffer.Width() && y >= 0 && y < buffer.Height() {
				writeCell(buffer, x, y, '█', thumbFg, thumbBg, 0)
			}
		}
	}

	// Horizontal scrollbar on bottom edge
	if contentW > viewportW {
		thumbW := viewportW * viewportW / contentW
		if thumbW < 1 {
			thumbW = 1
		}
		thumbStartX := int(float64(offsetX) / float64(contentW) * float64(viewportW))
		if thumbStartX < 0 {
			thumbStartX = 0
		}
		thumbEnd := thumbStartX + thumbW
		if thumbEnd > viewportW {
			thumbEnd = viewportW
		}
		y := viewportH - 1
		for x := thumbStartX; x < thumbEnd; x++ {
			if x >= 0 && x < buffer.Width() && y >= 0 && y < buffer.Height() {
				writeCell(buffer, x, y, '█', thumbFg, thumbBg, 0)
			}
		}
	}
}

func (r *ScrollBoxRenderable) OnMouseEvent(event *core.MouseEvent) {
	if r.handleBarMouse(event) { return }
	if event.Pressed {
		r.updateAutoScroll(event.X, event.Y)
	} else {
		r.stopAutoScroll()
	}
}

func (r *ScrollBoxRenderable) handleBarMouse(event *core.MouseEvent) bool {
	press := r.barPress
	if press != nil && event.Pressed {
		if press.arrow == 0 { r.dragThumb(press, event) }
		return true
	}
	if press != nil && !event.Pressed {
		r.endBarPress()
		return true
	}
	if event.Button != 0 { return false }
	state := r.scrollState()
	_ = state
	for _, axis := range []int{1, 0} {
		bar := r.verticalBar
		if axis == 0 { bar = r.horizontalBar }
		if !bar.visible { continue }
		r.barPress = &BarPress{axis: axis, offset: 0, arrow: 0}
		r.pressTrack(axis, event)
		return true
	}
	return false
}

func (r *ScrollBoxRenderable) pressArrow(axis int, arrow int, part *ScrollBoxBar) {
	press := &BarPress{axis: axis, arrow: arrow}
	r.barPress = press
	part.ScrollBy(float64(arrow)*0.5, ScrollUnitViewport)
}

func (r *ScrollBoxRenderable) thumbPointer(axis int, event *core.MouseEvent) (at, length, start, size int) {
	state := r.scrollState()
	l := 0
	if axis == 1 { l = int(state.Viewport[3]) } else { l = int(state.Viewport[2]) }
	along := 0
	if axis == 1 { along = event.Y } else { along = event.X }
	if along < 0 { along = 0 }
	if along > l { along = l }
	return along * 2, l * 2, 0, l
}

func (r *ScrollBoxRenderable) pressTrack(axis int, event *core.MouseEvent) {
	at, length, start, size := r.thumbPointer(axis, event)
	inThumb := at >= start && at < start+size
	if !inThumb {
		state := r.scrollState()
		max := state.Max[axis]
		var pos float64
		if length == 0 { pos = 0 } else { pos = float64(at) / float64(length) * float64(max) }
		r.scrollAxis(axis, int(pos), false)
		if r.isDestroyed() { return }
		at, length, start, size = r.thumbPointer(axis, event)
	}
	r.barPress = &BarPress{axis: axis, offset: at - start}
}

func (r *ScrollBoxRenderable) dragThumb(press *BarPress, event *core.MouseEvent) {
	at, length, _, size := r.thumbPointer(press.axis, event)
	span := length - size
	if span < 0 { span = 0 }
	startPos := at - press.offset
	if startPos < 0 { startPos = 0 }
	if startPos > span { startPos = span }
	state := r.scrollState()
	max := state.Max[press.axis]
	var pos float64
	if span == 0 { pos = 0 } else { pos = float64(startPos) / float64(span) * float64(max) }
	r.scrollAxis(press.axis, int(pos), false)
}

func (r *ScrollBoxRenderable) endBarPress() {
	r.barPress = nil
}

func (r *ScrollBoxRenderable) HandleKeyPress(key *core.KeyEvent) bool {
	return false
}

func (r *ScrollBoxRenderable) resetScrollAccumulators() {
	r.scrollAccumX = 0
	r.scrollAccumY = 0
}

func (r *ScrollBoxRenderable) startAutoScroll(mouseX, mouseY int) {
	if r.isDestroyed() { return }
	r.stopAutoScroll()
	r.autoScrollMouseX = mouseX
	r.autoScrollMouseY = mouseY
	r.cachedAutoScrollSpeed = r.getAutoScrollSpeed(mouseX, mouseY)
	r.isAutoScrolling = true
	r.refreshScrollHook()
}

func (r *ScrollBoxRenderable) updateAutoScroll(mouseX, mouseY int) {
	r.autoScrollMouseX = mouseX
	r.autoScrollMouseY = mouseY
	r.cachedAutoScrollSpeed = r.getAutoScrollSpeed(mouseX, mouseY)
	scrollX := r.getAutoScrollDirectionX(mouseX)
	scrollY := r.getAutoScrollDirectionY(mouseY)
	if scrollX == 0 && scrollY == 0 {
		r.stopAutoScroll()
	} else if !r.isAutoScrolling {
		r.startAutoScroll(mouseX, mouseY)
	}
}

func (r *ScrollBoxRenderable) stopAutoScroll() {
	was := r.isAutoScrolling
	r.isAutoScrolling = false
	r.autoScrollAccumX = 0
	r.autoScrollAccumY = 0
	if was {
		r.refreshScrollHook()
	}
}

func (r *ScrollBoxRenderable) handleAutoScroll(deltaTime float64) {
	if !r.isAutoScrolling { return }
	scrollX := r.getAutoScrollDirectionX(r.autoScrollMouseX)
	scrollY := r.getAutoScrollDirectionY(r.autoScrollMouseY)
	scrollAmount := float64(r.cachedAutoScrollSpeed) * (deltaTime / 1000.0)
	scrolled := false

	if scrollX != 0 {
		r.autoScrollAccumX += float64(scrollX) * scrollAmount
		intX := int(math.Trunc(r.autoScrollAccumX))
		if intX != 0 {
			r.SetScrollLeft(r.ScrollLeft() + intX)
			r.autoScrollAccumX -= float64(intX)
			scrolled = true
		}
	}
	if scrollY != 0 {
		r.autoScrollAccumY += float64(scrollY) * scrollAmount
		intY := int(math.Trunc(r.autoScrollAccumY))
		if intY != 0 {
			r.SetScrollTop(r.ScrollTop() + intY)
			r.autoScrollAccumY -= float64(intY)
			scrolled = true
		}
	}
	if scrolled {
		ctx := r.Context()
		if ctx != nil && ctx.RequestSelectionUpdate != nil { ctx.RequestSelectionUpdate() }
	}
	if scrollX == 0 && scrollY == 0 {
		r.stopAutoScroll()
	}
}

func (r *ScrollBoxRenderable) getAutoScrollDirectionX(mouseX int) int {
	relX := mouseX - int(r.GetX())
	distToLeft := relX
	distToRight := int(r.GetWidth()) - relX
	if distToLeft <= r.autoScrollThresholdH {
		if r.ScrollLeft() > 0 { return -1 }
		return 0
	}
	if distToRight <= r.autoScrollThresholdH {
		maxScroll := r.ScrollWidth() - int(r.GetWidth())
		if maxScroll < 0 { maxScroll = 0 }
		if r.ScrollLeft() < maxScroll { return 1 }
		return 0
	}
	return 0
}

func (r *ScrollBoxRenderable) getAutoScrollDirectionY(mouseY int) int {
	relY := mouseY - int(r.GetY())
	distToTop := relY
	distToBottom := int(r.GetHeight()) - relY
	if distToTop <= r.autoScrollThresholdV {
		if r.ScrollTop() > 0 { return -1 }
		return 0
	}
	if distToBottom <= r.autoScrollThresholdV {
		maxScroll := r.ScrollHeight() - int(r.GetHeight())
		if maxScroll < 0 { maxScroll = 0 }
		if r.ScrollTop() < maxScroll { return 1 }
		return 0
	}
	return 0
}

func (r *ScrollBoxRenderable) getAutoScrollSpeed(mouseX, mouseY int) int {
	relX := mouseX - int(r.GetX())
	relY := mouseY - int(r.GetY())
	distToLeft := relX
	distToRight := int(r.GetWidth()) - relX
	distToTop := relY
	distToBottom := int(r.GetHeight()) - relY
	minDist := distToLeft
	if distToRight < minDist { minDist = distToRight }
	if distToTop < minDist { minDist = distToTop }
	if distToBottom < minDist { minDist = distToBottom }
	if minDist <= 1 { return r.autoScrollSpeedFast }
	if minDist <= 2 { return r.autoScrollSpeedMed }
	return r.autoScrollSpeedSlow
}

func (r *ScrollBoxRenderable) SetRootOptions(opts *ScrollBoxOptions) {
	if opts == nil { return }
	r.RequestRender()
}

func (r *ScrollBoxRenderable) SetWrapperOptions(opts *BoxOptions) {}
func (r *ScrollBoxRenderable) SetViewportOptions(opts *BoxOptions) {}
func (r *ScrollBoxRenderable) SetContentOptions(opts *BoxOptions) {}
func (r *ScrollBoxRenderable) SetScrollbarOptions(opts *ScrollBarPartOptions) {
	r.applyBarOptions(r.verticalBar, opts)
	r.applyBarOptions(r.horizontalBar, opts)
}
func (r *ScrollBoxRenderable) SetVerticalScrollbarOptions(opts *ScrollBarPartOptions) { r.applyBarOptions(r.verticalBar, opts) }
func (r *ScrollBoxRenderable) SetHorizontalScrollbarOptions(opts *ScrollBarPartOptions) { r.applyBarOptions(r.horizontalBar, opts) }

func (r *ScrollBoxRenderable) ScrollAcceleration() ScrollAcceleration { return r.scrollAccel }
func (r *ScrollBoxRenderable) SetScrollAcceleration(v ScrollAcceleration) { r.scrollAccel = v }

func (r *ScrollBoxRenderable) Overflow() string { return "hidden" }
func (r *ScrollBoxRenderable) SetOverflow(v string) {}

func (r *ScrollBoxRenderable) DestroySelf() {
	r.layoutCallbacks = nil
	r.endBarPress()
	r.stopAutoScroll()
	r.Renderable.Destroy()
}

var _ = core.TextAttributeNone
var _ = native.SceneNodeKindBox

// GetScrollY returns the vertical scroll position (alias for ScrollTop).
func (r *ScrollBoxRenderable) GetScrollY() int { return r.ScrollTop() }

// GetScrollX returns the horizontal scroll position (alias for ScrollLeft).
func (r *ScrollBoxRenderable) GetScrollX() int { return r.ScrollLeft() }

// SetContentSize sets the content dimensions and recomputes max scroll.
func (r *ScrollBoxRenderable) SetContentSize(w, h int) {
	r.mu.Lock()
	r.state.Content[2] = float64(w)
	r.state.Content[3] = float64(h)
	r.state.Max[0] = float64(w) - float64(int(r.GetWidth()))
	r.state.Max[1] = float64(h) - float64(int(r.GetHeight()))
	if r.state.Max[0] < 0 { r.state.Max[0] = 0 }
	if r.state.Max[1] < 0 { r.state.Max[1] = 0 }
	r.mu.Unlock()
	r.RequestRender()
	r.applyScrollToNative()
}

// SetViewportSize sets the viewport dimensions and recomputes max scroll.
func (r *ScrollBoxRenderable) SetViewportSize(w, h int) {
	r.mu.Lock()
	r.state.Viewport[2] = float64(w)
	r.state.Viewport[3] = float64(h)
	r.state.Max[0] = r.state.Content[2] - float64(w)
	r.state.Max[1] = r.state.Content[3] - float64(h)
	if r.state.Max[0] < 0 { r.state.Max[0] = 0 }
	if r.state.Max[1] < 0 { r.state.Max[1] = 0 }
	r.mu.Unlock()
	r.RequestRender()
	r.applyScrollToNative()
}

func (r *ScrollBoxRenderable) applyScrollToNative() {
	ns := r.GetNativeScene()
	if ns == nil { return }
	padding := r.viewport.Padding()
	_ = ns.SetScroll(r.Renderable, core.NativeSceneScrollOptions{
		Axes:    0,
		Follow:  0,
		Culling: r._viewportCulling,
		Padding: [4]float32{float32(padding[0]), float32(padding[1]), float32(padding[2]), float32(padding[3])},
	})
}
