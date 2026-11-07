package solid

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/Gomobra/PixTui/packages/core"
	"github.com/Gomobra/PixTui/packages/core/renderables"
)

// ----- Reactive primitives -----

// Signal holds a reactive value and a list of effects that depend on it.
type Signal struct {
	mu     sync.RWMutex
	value  interface{}
	effects []*Effect
}

// NewSignal creates a new signal with the given initial value.
func NewSignal(initialValue interface{}) *Signal {
	return &Signal{value: initialValue}
}

// Value returns the current signal value and registers the calling effect
// as a dependency if an effect is active.
func (s *Signal) Value() interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if active := getActiveEffect(); active != nil {
		s.mu.RUnlock()
		s.mu.Lock()
		// avoid duplicate registration
		for _, e := range s.effects {
			if e == active {
				s.mu.Unlock()
				return s.value
			}
		}
		s.effects = append(s.effects, active)
		s.mu.Unlock()
		s.mu.RLock()
		return s.value
	}
	return s.value
}

// Set updates the signal value and schedules all dependent effects for re-run.
func (s *Signal) Set(newValue interface{}) {
	s.mu.Lock()
	old := s.value
	s.value = newValue
	effects := append([]*Effect{}, s.effects...)
	s.mu.Unlock()

	if reflect.DeepEqual(old, newValue) {
		return
	}
	for _, e := range effects {
		e.schedule()
	}
}

// Effect represents a reactive effect that re-runs when its tracked signals change.
type Effect struct {
	mu     sync.Mutex
	fn     func()
	deps   []*Signal
	scheduled bool
	running bool
}

// newEffect creates an effect and immediately runs it once.
func newEffect(fn func()) *Effect {
	e := &Effect{fn: fn, scheduled: true}
	e.run()
	return e
}

func (e *Effect) run() {
	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		return
	}
	e.running = true
	setActiveEffect(e)
	// Clear previous deps so we re-collect them on this run.
	e.deps = nil
	e.mu.Unlock()

	e.fn()

	setActiveEffect(nil)
	e.mu.Lock()
	e.running = false
	e.mu.Unlock()
}

func (e *Effect) schedule() {
	e.mu.Lock()
	if e.scheduled {
		e.mu.Unlock()
		return
	}
	e.scheduled = true
	e.mu.Unlock()
	// Re-run asynchronously (next tick style) to allow batching.
	go e.run()
}

// trackSignal registers a signal as a dependency of the active effect.
func trackSignal(s *Signal) {
	if active := getActiveEffect(); active != nil {
		s.mu.RLock()
		for _, d := range active.deps {
			if d == s {
				s.mu.RUnlock()
				return
			}
		}
		s.mu.RUnlock()
		s.mu.Lock()
		// register effect on signal
		found := false
		for _, ef := range s.effects {
			if ef == active {
				found = true
				break
			}
		}
		if !found {
			s.effects = append(s.effects, active)
		}
		s.mu.Unlock()
		active.mu.Lock()
		active.deps = append(active.deps, s)
		active.mu.Unlock()
	}
}

// ----- Active effect context (for CreateEffect/CreateMemo) -----

var (
	activeEffectMu sync.Mutex
	activeEffect   *Effect
)

func setActiveEffect(e *Effect) {
	activeEffectMu.Lock()
	activeEffect = e
	activeEffectMu.Unlock()
}

func getActiveEffect() *Effect {
	activeEffectMu.Lock()
	defer activeEffectMu.Unlock()
	return activeEffect
}

// CreateSignal creates a reactive signal and returns (value, setter).
// The setter triggers re-render of any effects that read the signal.
func CreateSignal(initialValue interface{}) (interface{}, func(interface{})) {
	sig := NewSignal(initialValue)
	return sig.Value(), func(newValue interface{}) {
		sig.Set(newValue)
	}
}

// CreateEffect runs an effect and re-runs it when its tracked signals change.
func CreateEffect(effect func()) {
	newEffect(effect)
}

// memoSlot holds a memoized value and its dependencies.
type memoSlot struct {
	value interface{}
	deps  []interface{}
}

// createMemoInternal creates or retrieves a memoized value for the current fiber.
// If the fiber has seen this memo slot before and deps haven't changed, the
// cached value is returned. Otherwise the memo function is re-invoked.
func createMemoInternal(f *fiber, memo func() interface{}, deps []interface{}) interface{} {
	f.mu.Lock()
	idx := f.hooksIdx
	f.hooksIdx++

	var slot *memoSlot
	if idx < len(f.memos) {
		slot = &f.memos[idx]
	} else {
		f.memos = append(f.memos, memoSlot{})
		slot = &f.memos[idx]
	}

	if !depsEqual(slot.deps, deps) {
		slot.value = memo()
		slot.deps = copyDeps(deps)
	}
	value := slot.value
	f.mu.Unlock()
	return value
}

// CreateMemo returns a memoized value, recomputed only when its dependencies change.
// When called outside of a component fiber, the memo function is invoked once
// and the result is cached.
func CreateMemo(memo func() interface{}) interface{} {
	if f := getCurrentFiberForSolid(); f != nil {
		return createMemoInternal(f, memo, nil)
	}
	return memo()
}

// solidComponentRegistry maps component names to their render functions.
var solidComponentRegistry = make(map[string]func(props map[string]interface{}) SolidComponent)

// RegisterComponent registers a component function under a name, mirroring the
// @opentui/solid extend() API used by framework packages like @opentui/qrcode.
func RegisterComponent(name string, fn func(props map[string]interface{}) SolidComponent) {
	solidComponentRegistry[name] = fn
}

// RegisteredComponent returns a previously registered component, or nil if none
// is registered under the given name.
func RegisteredComponent(name string) (func(props map[string]interface{}) SolidComponent, bool) {
	fn, ok := solidComponentRegistry[name]
	return fn, ok
}

// ----- Component model -----

// SolidComponent represents a Solid component node.
type SolidComponent struct {
	Type  string
	Props map[string]interface{}
	Fn    func(props map[string]interface{}) SolidComponent
}

// CreateComponent creates a Solid component from a render function.
func CreateComponent(fn func(props map[string]interface{}) SolidComponent) SolidComponent {
	return SolidComponent{Fn: fn}
}

// Box creates a Box Solid component.
func Box(props map[string]interface{}) SolidComponent {
	return SolidComponent{Type: "box", Props: props}
}

// Text creates a Text Solid component.
func Text(props map[string]interface{}) SolidComponent {
	return SolidComponent{Type: "text", Props: props}
}

// Input creates an Input Solid component.
func Input(props map[string]interface{}) SolidComponent {
	return SolidComponent{Type: "input", Props: props}
}

// Select creates a Select Solid component.
func Select(props map[string]interface{}) SolidComponent {
	return SolidComponent{Type: "select", Props: props}
}

// ScrollBox creates a ScrollBox Solid component.
func ScrollBox(props map[string]interface{}) SolidComponent {
	return SolidComponent{Type: "scrollbox", Props: props}
}

// Textarea creates a Textarea Solid component.
func Textarea(props map[string]interface{}) SolidComponent {
	return SolidComponent{Type: "textarea", Props: props}
}

// Slider creates a Slider Solid component.
func Slider(props map[string]interface{}) SolidComponent {
	return SolidComponent{Type: "slider", Props: props}
}

// TabSelect creates a TabSelect Solid component.
func TabSelect(props map[string]interface{}) SolidComponent {
	return SolidComponent{Type: "tabselect", Props: props}
}

// Image creates an Image Solid component.
func Image(props map[string]interface{}) SolidComponent {
	return SolidComponent{Type: "image", Props: props}
}

// Video creates a Video Solid component.
func Video(props map[string]interface{}) SolidComponent {
	return SolidComponent{Type: "video", Props: props}
}

// Code creates a Code Solid component.
func Code(props map[string]interface{}) SolidComponent {
	return SolidComponent{Type: "code", Props: props}
}

// Markdown creates a Markdown Solid component.
func Markdown(props map[string]interface{}) SolidComponent {
	return SolidComponent{Type: "markdown", Props: props}
}

// Diff creates a Diff Solid component.
func Diff(props map[string]interface{}) SolidComponent {
	return SolidComponent{Type: "diff", Props: props}
}

// LineNumber creates a LineNumber Solid component.
func LineNumber(props map[string]interface{}) SolidComponent {
	return SolidComponent{Type: "line-number", Props: props}
}

// ScrollBar creates a ScrollBar Solid component.
func ScrollBar(props map[string]interface{}) SolidComponent {
	return SolidComponent{Type: "scrollbar", Props: props}
}

// TextTable creates a TextTable Solid component.
func TextTable(props map[string]interface{}) SolidComponent {
	return SolidComponent{Type: "text-table", Props: props}
}

// ASCIIFont creates an ASCIIFont Solid component.
func ASCIIFont(props map[string]interface{}) SolidComponent {
	return SolidComponent{Type: "ascii-font", Props: props}
}

// FrameBuffer creates a FrameBuffer Solid component.
func FrameBuffer(props map[string]interface{}) SolidComponent {
	return SolidComponent{Type: "frame-buffer", Props: props}
}

// EditBuffer creates an EditBuffer Solid component.
func EditBuffer(props map[string]interface{}) SolidComponent {
	return SolidComponent{Type: "edit-buffer", Props: props}
}

// ----- Context -----

// Context holds a value that can be provided to a component tree and consumed
// by descendants via useContext.
type Context struct {
	defaultValue interface{}
}

// CreateContext creates a new context with the given default value. The
// default value is returned by useContext when no provider is found in the
// fiber tree.
func CreateContext(defaultValue interface{}) *Context {
	return &Context{defaultValue: defaultValue}
}

// Provider returns a Solid component that provides the given context value to
// its descendants. The value is read from props["value"] and children from
// props["children"].
func (c *Context) Provider(props map[string]interface{}) SolidComponent {
	value := props["value"]
	return CreateComponent(func(p map[string]interface{}) SolidComponent {
		_ = p
		f := getCurrentFiberForSolid()
		if f != nil {
			f.mu.Lock()
			if f.contextValues == nil {
				f.contextValues = make(map[*Context]interface{})
			}
			f.contextValues[c] = value
			f.mu.Unlock()
		}
		if children, ok := props["children"].([]SolidComponent); ok && len(children) > 0 {
			return children[0]
		}
		return SolidComponent{}
	})
}

// useContext returns the value of the nearest context provider for the given
// context in the current fiber's ancestor chain, falling back to the
// context's default value.
func useContext(ctx *Context) interface{} {
	f := getCurrentFiberForSolid()
	for f != nil {
		f.mu.Lock()
		if f.contextValues != nil {
			if val, ok := f.contextValues[ctx]; ok {
				f.mu.Unlock()
				return val
			}
		}
		parent := f.parent
		f.mu.Unlock()
		f = parent
	}
	return ctx.defaultValue
}

// ----- Fiber tree (internal reconciler) -----

type effectSlot struct {
	fn      func()
	deps    []interface{}
	run     bool
	cleanup func()
}

type fiber struct {
	renderer   *SolidRenderer
	component  SolidComponent
	props      map[string]interface{}
	parent     *fiber
	children   []*fiber

	dom          *core.Renderable
	domAttached  bool
	mu sync.Mutex

	signals []*Signal
	memos   []memoSlot
	effects []effectSlot
	hooksIdx int

	// contextValues maps a Context to the value provided by this fiber.
	contextValues map[*Context]interface{}
}

// SolidRenderer renders Solid component trees to core renderables.
type SolidRenderer struct {
	mu sync.Mutex

	ctx *core.RenderContext

	root *fiber

	rootComponent SolidComponent
	lastBuffer  *core.NativeBuffer
	lastDelta   float64

	renderScheduled bool
	rendering       bool
	pendingEffects  []pendingEffect
}

type pendingEffect struct {
	fiber *fiber
	slot  int
}

// NewSolidRenderer creates a new Solid renderer bound to the given context.
func NewSolidRenderer(ctx *core.RenderContext) *SolidRenderer {
	return &SolidRenderer{ctx: ctx}
}

// Render reconciles the component tree and paints it into the buffer.
func (r *SolidRenderer) Render(component SolidComponent, buffer *core.NativeBuffer, deltaTime float64) {
	r.mu.Lock()
	r.rootComponent = component
	r.lastBuffer = buffer
	r.lastDelta = deltaTime
	r.renderScheduled = true
	r.mu.Unlock()

	r.renderLoop()
}

// ScheduleRender requests a re-render using the last buffer and delta time.
func (r *SolidRenderer) ScheduleRender() {
	r.mu.Lock()
	if r.renderScheduled {
		r.mu.Unlock()
		return
	}
	r.renderScheduled = true
	r.mu.Unlock()
	r.renderLoop()
}

func (r *SolidRenderer) renderLoop() {
	for {
		r.mu.Lock()
		if r.rendering {
			r.mu.Unlock()
			return
		}
		if !r.renderScheduled {
			r.mu.Unlock()
			return
		}
		r.rendering = true
		r.renderScheduled = false
		component := r.rootComponent
		buffer := r.lastBuffer
		deltaTime := r.lastDelta
		if buffer == nil {
			r.rendering = false
			r.mu.Unlock()
			return
		}
		buffer.Clear()
		oldRoot := r.root
		root := r.reconcileFiber(oldRoot, nil, component)
		r.root = root
		pendingEffects := r.pendingEffects
		r.pendingEffects = nil
		r.rendering = false
		r.mu.Unlock()

		r.paintFiber(root, nil, buffer, deltaTime)
		r.runEffects(pendingEffects)
	}
}

func (r *SolidRenderer) reconcileFiber(old *fiber, parent *fiber, comp SolidComponent) *fiber {
	if old != nil && sameComponentType(old, comp) {
		return r.updateFiber(old, comp)
	}
	if old != nil {
		r.unmountFiber(old)
	}
	return r.mountFiber(parent, comp)
}

func (r *SolidRenderer) mountFiber(parent *fiber, comp SolidComponent) *fiber {
	f := &fiber{
		renderer:   r,
		component:  comp,
		props:      comp.Props,
		parent:     parent,
	}

	if comp.Fn != nil {
		// Custom component: run its function and reconcile the output.
		r.withFiber(f, func() {
			f.hooksIdx = 0
			output := comp.Fn(comp.Props)
			if output.Fn != nil || output.Type != "" {
				child := r.reconcileFiber(nil, f, output)
				f.children = []*fiber{child}
			}
		})
		return f
	}

	// Host element: create the renderable now.
	dom := createHostRenderable(comp, r.ctx)
	f.dom = dom
	if parent != nil && parent.dom != nil {
		parent.dom.AddChild(dom)
		f.domAttached = true
	}
	if len(comp.Props) > 0 && comp.Props["children"] != nil {
		if children, ok := comp.Props["children"].([]SolidComponent); ok {
			f.children = r.reconcileChildren(f, children)
		}
	}

	return f
}

func (r *SolidRenderer) updateFiber(f *fiber, comp SolidComponent) *fiber {
	f.component = comp
	f.props = comp.Props

	if comp.Fn != nil {
		r.withFiber(f, func() {
			f.hooksIdx = 0
			var oldChild *fiber
			if len(f.children) > 0 {
				oldChild = f.children[0]
			}
			output := comp.Fn(comp.Props)
			child := r.reconcileFiber(oldChild, f, output)
			f.children = []*fiber{child}
		})
		return f
	}

	propsChanged := !deepEqualProps(f.props, comp.Props)
	if propsChanged && f.dom != nil {
		if p := f.dom.Parent(); p != nil {
			p.RemoveChild(f.dom)
		}
		dom := createHostRenderable(comp, r.ctx)
		f.dom = dom
		f.domAttached = false
		if p := f.parent; p != nil && p.dom != nil {
			p.dom.AddChild(dom)
			f.domAttached = true
		}
	}
	if len(comp.Props) > 0 && comp.Props["children"] != nil {
		if children, ok := comp.Props["children"].([]SolidComponent); ok {
			f.children = r.reconcileChildren(f, children)
		}
	}

	return f
}

func (r *SolidRenderer) reconcileChildren(parent *fiber, newChildren []SolidComponent) []*fiber {
	old := parent.children
	next := make([]*fiber, 0, len(newChildren))
	for i, el := range newChildren {
		var reuse *fiber
		if i < len(old) {
			c := old[i]
			if sameComponentType(c, el) {
				reuse = c
			}
		}
		if reuse != nil {
			next = append(next, r.updateFiber(reuse, el))
			continue
		}
		next = append(next, r.mountFiber(parent, el))
	}
	parent.children = next
	return next
}

func (r *SolidRenderer) unmountFiber(f *fiber) {
	if f == nil {
		return
	}
	for _, c := range f.children {
		r.unmountFiber(c)
	}
	f.children = nil
	if f.dom != nil {
		if p := f.dom.Parent(); p != nil {
			p.RemoveChild(f.dom)
		}
		f.dom = nil
		f.domAttached = false
	}
	for _, e := range f.effects {
		if e.cleanup != nil {
			e.cleanup()
		}
	}
	f.effects = nil
}

func (r *SolidRenderer) paintFiber(f *fiber, parentDom *core.Renderable, buffer *core.NativeBuffer, deltaTime float64) {
	if f == nil {
		return
	}

	target := parentDom
	if f.dom != nil {
		if !f.domAttached && parentDom != nil {
			parentDom.AddChild(f.dom)
			f.domAttached = true
		}
		f.dom.RenderBefore(buffer, deltaTime)
		f.dom.RenderSelf(buffer, deltaTime)
		f.dom.RenderAfter(buffer, deltaTime)
		target = f.dom
	}

	for _, c := range f.children {
		r.paintFiber(c, target, buffer, deltaTime)
	}
}

func (r *SolidRenderer) scheduleEffect(f *fiber, slot int) {
	r.pendingEffects = append(r.pendingEffects, pendingEffect{fiber: f, slot: slot})
}

func (r *SolidRenderer) runEffects(pendingEffects []pendingEffect) {
	for _, pe := range pendingEffects {
		pe.fiber.mu.Lock()
		if pe.slot >= len(pe.fiber.effects) {
			pe.fiber.mu.Unlock()
			continue
		}
		eff := &pe.fiber.effects[pe.slot]
		if eff.cleanup != nil {
			cleanup := eff.cleanup
			eff.cleanup = nil
			pe.fiber.mu.Unlock()
			cleanup()
			pe.fiber.mu.Lock()
		}
		fn := eff.fn
		eff.run = true
		pe.fiber.mu.Unlock()
		fn()
	}
}

func (r *SolidRenderer) withFiber(f *fiber, fn func()) {
	setCurrentFiberForSolid(f)
	fn()
	setCurrentFiberForSolid(nil)
}

// ----- Current fiber context (for hooks) -----

var (
	currentFiberMu sync.Mutex
	currentFiber   *fiber
)

func setCurrentFiberForSolid(f *fiber) {
	currentFiberMu.Lock()
	currentFiber = f
	currentFiberMu.Unlock()
}

func getCurrentFiberForSolid() *fiber {
	currentFiberMu.Lock()
	defer currentFiberMu.Unlock()
	return currentFiber
}

// ----- Package-level hooks -----

// UseSignal returns the current signal value and setter for the current fiber.
func UseSignal(initial interface{}) (interface{}, func(interface{})) {
	if f := getCurrentFiberForSolid(); f != nil {
		f.mu.Lock()
		idx := f.hooksIdx
		f.hooksIdx++
		if idx >= len(f.signals) {
			sig := NewSignal(initial)
			f.signals = append(f.signals, sig)
		}
		sig := f.signals[idx]
		f.mu.Unlock()
		return sig.Value(), func(newValue interface{}) {
			sig.Set(newValue)
			f.renderer.ScheduleRender()
		}
	}
	return initial, func(interface{}) {}
}

// UseEffect schedules an effect on the current fiber.
func UseEffect(effect func(), deps []interface{}) {
	if f := getCurrentFiberForSolid(); f != nil {
		f.mu.Lock()
		idx := f.hooksIdx
		f.hooksIdx++
		var eff *effectSlot
		if idx < len(f.effects) {
			eff = &f.effects[idx]
		} else {
			f.effects = append(f.effects, effectSlot{})
			eff = &f.effects[idx]
		}
		shouldRun := !eff.run || !depsEqual(eff.deps, deps)
		if shouldRun {
			r := f.renderer
			r.scheduleEffect(f, idx)
		}
		eff.deps = copyDeps(deps)
		f.mu.Unlock()
		return
	}
	effect()
}

// UseMemo returns a memoized value on the current fiber.
func UseMemo(memo func() interface{}, deps []interface{}) interface{} {
	if f := getCurrentFiberForSolid(); f != nil {
		f.mu.Lock()
		idx := f.hooksIdx
		f.hooksIdx++
		var slot *memoSlot
		if idx < len(f.memos) {
			slot = &f.memos[idx]
		} else {
			f.memos = append(f.memos, memoSlot{})
			slot = &f.memos[idx]
		}
		if !depsEqual(slot.deps, deps) {
			slot.value = memo()
			slot.deps = copyDeps(deps)
		}
		value := slot.value
		f.mu.Unlock()
		return value
	}
	return memo()
}

// ----- Host element creation -----

func createHostRenderable(comp SolidComponent, ctx *core.RenderContext) *core.Renderable {
	switch comp.Type {
	case "box":
		return createBox(comp, ctx)
	case "text":
		return createText(comp, ctx)
	case "input":
		return createInput(comp, ctx)
	case "select":
		return createSelect(comp, ctx)
	case "scrollbox":
		return createScrollBox(comp, ctx)
	case "textarea":
		return createTextarea(comp, ctx)
	case "slider":
		return createSlider(comp, ctx)
	case "tabselect":
		return createTabSelect(comp, ctx)
	case "image":
		return createImage(comp, ctx)
	case "video":
		return createVideo(comp, ctx)
	case "code":
		return createCode(comp, ctx)
	case "markdown":
		return createMarkdown(comp, ctx)
	case "diff":
		return createDiff(comp, ctx)
	case "line-number":
		return createLineNumber(comp, ctx)
	case "scrollbar":
		return createScrollBar(comp, ctx)
	case "text-table":
		return createTextTable(comp, ctx)
	case "ascii-font":
		return createASCIIFont(comp, ctx)
	case "frame-buffer":
		return createFrameBuffer(comp, ctx)
	case "edit-buffer":
		return createEditBuffer(comp, ctx)
	default:
		id := ""
		if comp.Props != nil {
			if s, ok := comp.Props["id"].(string); ok {
				id = s
			}
		}
		return core.NewRenderable(ctx, id, core.RenderableOptions[core.Renderable]{})
	}
}

func createBox(comp SolidComponent, ctx *core.RenderContext) *core.Renderable {
	opts := renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      getString(comp.Props, "id", ""),
			Width:   strPtr(getString(comp.Props, "width", "")),
			Height:  strPtr(getString(comp.Props, "height", "")),
			Visible: boolPtr(getBool(comp.Props, "visible", true)),
			ZIndex:  intPtr(getInt(comp.Props, "zIndex", 0)),
		},
		BackgroundColor:    rgbaPtr(renderables.ParseHex(getString(comp.Props, "backgroundColor", "transparent"))),
		BorderStyle:        getString(comp.Props, "borderStyle", "single"),
		Border:             getBool(comp.Props, "border", false),
		BorderColor:        rgbaPtr(renderables.ParseHex(getString(comp.Props, "borderColor", "#FFFFFF"))),
		ShouldFill:         getBool(comp.Props, "shouldFill", false),
		Title:              getString(comp.Props, "title", ""),
		TitleColor:         rgbaPtrOpt(renderables.ParseHex(getString(comp.Props, "titleColor", ""))),
		TitleAlignment:     getString(comp.Props, "titleAlignment", "left"),
		BottomTitle:        getString(comp.Props, "bottomTitle", ""),
		Focusable:          getBool(comp.Props, "focusable", false),
		FocusedBorderColor: rgbaPtr(renderables.ParseHex(getString(comp.Props, "focusedBorderColor", "#00AAFF"))),
	}
	r := renderables.NewBoxRenderable(ctx, opts)
	return r.Renderable
}

func createText(comp SolidComponent, ctx *core.RenderContext) *core.Renderable {
	opts := renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      getString(comp.Props, "id", ""),
			Width:   strPtr(getString(comp.Props, "width", "")),
			Height:  strPtr(getString(comp.Props, "height", "")),
			Visible: boolPtr(getBool(comp.Props, "visible", true)),
			ZIndex:  intPtr(getInt(comp.Props, "zIndex", 0)),
		},
		Content: getString(comp.Props, "content", ""),
	}
	r := renderables.NewTextRenderable(ctx, opts)
	return r.Renderable
}

// createInput creates an Input renderable.
func createInput(comp SolidComponent, ctx *core.RenderContext) *core.Renderable {
	opts := renderables.InputOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      getString(comp.Props, "id", ""),
			Width:   strPtr(getString(comp.Props, "width", "")),
			Height:  strPtr(getString(comp.Props, "height", "")),
			Visible: boolPtr(getBool(comp.Props, "visible", true)),
			ZIndex:  intPtr(getInt(comp.Props, "zIndex", 0)),
		},
		Value:      getString(comp.Props, "value", ""),
		Placeholder: getString(comp.Props, "placeholder", ""),
	}
	r := renderables.NewInputRenderable(ctx, opts)
	return r.Renderable
}

// createSelect creates a Select renderable.
func createSelect(comp SolidComponent, ctx *core.RenderContext) *core.Renderable {
	opts := renderables.SelectOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      getString(comp.Props, "id", ""),
			Width:   strPtr(getString(comp.Props, "width", "")),
			Height:  strPtr(getString(comp.Props, "height", "")),
			Visible: boolPtr(getBool(comp.Props, "visible", true)),
			ZIndex:  intPtr(getInt(comp.Props, "zIndex", 0)),
		},
		Options:  getStringSlice(comp.Props, "options"),
		Selected: getInt(comp.Props, "selected", 0),
	}
	r := renderables.NewSelectRenderable(ctx, opts)
	return r.Renderable
}

// createScrollBox creates a ScrollBox renderable.
func createScrollBox(comp SolidComponent, ctx *core.RenderContext) *core.Renderable {
	opts := renderables.ScrollBoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      getString(comp.Props, "id", ""),
			Width:   strPtr(getString(comp.Props, "width", "")),
			Height:  strPtr(getString(comp.Props, "height", "")),
			Visible: boolPtr(getBool(comp.Props, "visible", true)),
			ZIndex:  intPtr(getInt(comp.Props, "zIndex", 0)),
		},
	}
	r := renderables.NewScrollBoxRenderable(ctx, opts)
	return r.Renderable
}

// createTextarea creates a Textarea renderable.
func createTextarea(comp SolidComponent, ctx *core.RenderContext) *core.Renderable {
	opts := renderables.TextareaOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      getString(comp.Props, "id", ""),
			Width:   strPtr(getString(comp.Props, "width", "")),
			Height:  strPtr(getString(comp.Props, "height", "")),
			Visible: boolPtr(getBool(comp.Props, "visible", true)),
			ZIndex:  intPtr(getInt(comp.Props, "zIndex", 0)),
		},
		Value: getString(comp.Props, "value", ""),
	}
	r := renderables.NewTextareaRenderable(ctx, opts)
	return r.Renderable
}

// createSlider creates a Slider renderable.
func createSlider(comp SolidComponent, ctx *core.RenderContext) *core.Renderable {
	opts := renderables.SliderOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      getString(comp.Props, "id", ""),
			Width:   strPtr(getString(comp.Props, "width", "")),
			Height:  strPtr(getString(comp.Props, "height", "")),
			Visible: boolPtr(getBool(comp.Props, "visible", true)),
			ZIndex:  intPtr(getInt(comp.Props, "zIndex", 0)),
		},
		Min:   getFloat(comp.Props, "min", 0),
		Max:   getFloat(comp.Props, "max", 100),
		Value: getFloat(comp.Props, "value", 0),
	}
	r := renderables.NewSliderRenderable(ctx, opts)
	return r.Renderable
}

// createTabSelect creates a TabSelect renderable.
func createTabSelect(comp SolidComponent, ctx *core.RenderContext) *core.Renderable {
	opts := renderables.TabSelectOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      getString(comp.Props, "id", ""),
			Width:   strPtr(getString(comp.Props, "width", "")),
			Height:  strPtr(getString(comp.Props, "height", "")),
			Visible: boolPtr(getBool(comp.Props, "visible", true)),
			ZIndex:  intPtr(getInt(comp.Props, "zIndex", 0)),
		},
		Tabs:     getStringSlice(comp.Props, "tabs"),
		Selected: getInt(comp.Props, "selected", 0),
	}
	r := renderables.NewTabSelectRenderable(ctx, opts)
	return r.Renderable
}

// createImage creates an Image renderable.
func createImage(comp SolidComponent, ctx *core.RenderContext) *core.Renderable {
	opts := renderables.ImageOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      getString(comp.Props, "id", ""),
			Width:   strPtr(getString(comp.Props, "width", "")),
			Height:  strPtr(getString(comp.Props, "height", "")),
			Visible: boolPtr(getBool(comp.Props, "visible", true)),
			ZIndex:  intPtr(getInt(comp.Props, "zIndex", 0)),
		},
		Path:     getString(comp.Props, "path", ""),
		Protocol: core.ImageRenderProtocolAuto,
	}
	r := renderables.NewImageRenderable(ctx, opts)
	return r.Renderable
}

// createVideo creates a Video renderable.
func createVideo(comp SolidComponent, ctx *core.RenderContext) *core.Renderable {
	opts := renderables.VideoOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      getString(comp.Props, "id", ""),
			Width:   strPtr(getString(comp.Props, "width", "")),
			Height:  strPtr(getString(comp.Props, "height", "")),
			Visible: boolPtr(getBool(comp.Props, "visible", true)),
			ZIndex:  intPtr(getInt(comp.Props, "zIndex", 0)),
		},
		AutoPlay: getBool(comp.Props, "autoPlay", false),
	}
	r := renderables.NewVideoRenderable(ctx, opts)
	return r.Renderable
}

// createCode creates a Code renderable.
func createCode(comp SolidComponent, ctx *core.RenderContext) *core.Renderable {
	opts := renderables.CodeOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      getString(comp.Props, "id", ""),
			Width:   strPtr(getString(comp.Props, "width", "")),
			Height:  strPtr(getString(comp.Props, "height", "")),
			Visible: boolPtr(getBool(comp.Props, "visible", true)),
			ZIndex:  intPtr(getInt(comp.Props, "zIndex", 0)),
		},
		Code:     getString(comp.Props, "code", ""),
		Language: getString(comp.Props, "language", ""),
	}
	r := renderables.NewCodeRenderable(ctx, opts)
	return r.Renderable
}

// createMarkdown creates a Markdown renderable.
func createMarkdown(comp SolidComponent, ctx *core.RenderContext) *core.Renderable {
	opts := renderables.MarkdownOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      getString(comp.Props, "id", ""),
			Width:   strPtr(getString(comp.Props, "width", "")),
			Height:  strPtr(getString(comp.Props, "height", "")),
			Visible: boolPtr(getBool(comp.Props, "visible", true)),
			ZIndex:  intPtr(getInt(comp.Props, "zIndex", 0)),
		},
		Content: getString(comp.Props, "content", ""),
	}
	r := renderables.NewMarkdownRenderable(ctx, opts)
	return r.Renderable
}

// createDiff creates a Diff renderable.
func createDiff(comp SolidComponent, ctx *core.RenderContext) *core.Renderable {
	opts := renderables.DiffOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      getString(comp.Props, "id", ""),
			Width:   strPtr(getString(comp.Props, "width", "")),
			Height:  strPtr(getString(comp.Props, "height", "")),
			Visible: boolPtr(getBool(comp.Props, "visible", true)),
			ZIndex:  intPtr(getInt(comp.Props, "zIndex", 0)),
		},
		Content: getString(comp.Props, "content", ""),
	}
	r := renderables.NewDiffRenderable(ctx, opts)
	return r.Renderable
}

// createLineNumber creates a LineNumber renderable.
func createLineNumber(comp SolidComponent, ctx *core.RenderContext) *core.Renderable {
	r := renderables.NewLineNumberRenderable(ctx, getString(comp.Props, "id", ""), getInt(comp.Props, "startLine", 1))
	return r.Renderable
}

// createScrollBar creates a ScrollBar renderable.
func createScrollBar(comp SolidComponent, ctx *core.RenderContext) *core.Renderable {
	opts := renderables.ScrollBarOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      getString(comp.Props, "id", ""),
			Width:   strPtr(getString(comp.Props, "width", "")),
			Height:  strPtr(getString(comp.Props, "height", "")),
			Visible: boolPtr(getBool(comp.Props, "visible", true)),
			ZIndex:  intPtr(getInt(comp.Props, "zIndex", 0)),
		},
		Vertical:   getBool(comp.Props, "vertical", true),
		ThumbSize:  getFloat(comp.Props, "thumbSize", 3),
		ThumbStart: getFloat(comp.Props, "thumbStart", 0),
	}
	r := renderables.NewScrollBarRenderable(ctx, opts)
	return r.Renderable
}

// createTextTable creates a TextTable renderable.
func createTextTable(comp SolidComponent, ctx *core.RenderContext) *core.Renderable {
	opts := renderables.TextTableOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      getString(comp.Props, "id", ""),
			Width:   strPtr(getString(comp.Props, "width", "")),
			Height:  strPtr(getString(comp.Props, "height", "")),
			Visible: boolPtr(getBool(comp.Props, "visible", true)),
			ZIndex:  intPtr(getInt(comp.Props, "zIndex", 0)),
		},
		Headers: getStringSlice(comp.Props, "headers"),
		Rows:    getStringSliceSlice(comp.Props, "rows"),
	}
	r := renderables.NewTextTableRenderable(ctx, opts)
	return r.Renderable
}

// createASCIIFont creates an ASCIIFont renderable.
func createASCIIFont(comp SolidComponent, ctx *core.RenderContext) *core.Renderable {
	opts := renderables.ASCIIFontOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      getString(comp.Props, "id", ""),
			Width:   strPtr(getString(comp.Props, "width", "")),
			Height:  strPtr(getString(comp.Props, "height", "")),
			Visible: boolPtr(getBool(comp.Props, "visible", true)),
			ZIndex:  intPtr(getInt(comp.Props, "zIndex", 0)),
		},
		Text:  getString(comp.Props, "text", ""),
		Font:  getString(comp.Props, "font", ""),
	}
	r := renderables.NewASCIIFontRenderable(ctx, opts)
	return r.Renderable
}

// createFrameBuffer creates a FrameBuffer renderable.
func createFrameBuffer(comp SolidComponent, ctx *core.RenderContext) *core.Renderable {
	opts := renderables.FrameBufferOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      getString(comp.Props, "id", ""),
			Width:   strPtr(getString(comp.Props, "width", "")),
			Height:  strPtr(getString(comp.Props, "height", "")),
			Visible: boolPtr(getBool(comp.Props, "visible", true)),
			ZIndex:  intPtr(getInt(comp.Props, "zIndex", 0)),
		},
	}
	r := renderables.NewFrameBufferRenderable(ctx, opts)
	return r.Renderable
}

// createEditBuffer creates an EditBuffer renderable.
func createEditBuffer(comp SolidComponent, ctx *core.RenderContext) *core.Renderable {
	opts := renderables.EditBufferRenderableOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      getString(comp.Props, "id", ""),
			Width:   strPtr(getString(comp.Props, "width", "")),
			Height:  strPtr(getString(comp.Props, "height", "")),
			Visible: boolPtr(getBool(comp.Props, "visible", true)),
			ZIndex:  intPtr(getInt(comp.Props, "zIndex", 0)),
		},
		Value:      getString(comp.Props, "value", ""),
		Placeholder: getString(comp.Props, "placeholder", ""),
		ReadOnly:   getBool(comp.Props, "readOnly", false),
	}
	r := renderables.NewEditBufferRenderable(ctx, opts)
	return r.Renderable
}

// ----- Prop accessors -----

func getString(props map[string]interface{}, key, defaultVal string) string {
	if val, ok := props[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return defaultVal
}

func getBool(props map[string]interface{}, key string, defaultVal bool) bool {
	if val, ok := props[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultVal
}

func getInt(props map[string]interface{}, key string, defaultVal int) int {
	if val, ok := props[key]; ok {
		if i, ok := val.(int); ok {
			return i
		}
		if f, ok := val.(float64); ok {
			return int(f)
		}
	}
	return defaultVal
}

func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }
func intPtr(i int) *int       { return &i }

// ----- Diffing helpers -----

func sameComponentType(a *fiber, b SolidComponent) bool {
	aHost := a.component.Fn == nil
	bHost := b.Fn == nil
	if aHost != bHost {
		return false
	}
	if aHost {
		return a.component.Type == b.Type
	}
	return true
}

func deepEqualProps(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		bv, ok := b[k]
		if !ok {
			return false
		}
		if !reflect.DeepEqual(v, bv) {
			return false
		}
	}
	return true
}

func depsEqual(a, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !reflect.DeepEqual(a[i], b[i]) {
			return false
		}
	}
	return true
}

func copyDeps(deps []interface{}) []interface{} {
	if deps == nil {
		return nil
	}
	out := make([]interface{}, len(deps))
	copy(out, deps)
	return out
}

// ----- RenderTest -----

// RenderTest renders a Solid component tree for testing and returns the root renderable.
func RenderTest(component SolidComponent) *core.RootRenderable {
	root := core.NewRootRenderable(&core.RenderContext{}, "root")
	renderer := NewSolidRenderer(root.Context())
	buffer := core.NewNativeBuffer(80, 24)
	renderer.Render(component, buffer, 0)
	return root
}

// BasicExample demonstrates a basic Solid TUI app.
func BasicExample() {
	// Example Solid TUI application
}

// rgbaPtr returns a pointer to a copy of the given core.RGBA.
func rgbaPtr(c core.RGBA) *core.RGBA {
	return &c
}

// rgbaPtrOpt returns a pointer to a copy of the given core.RGBA, or nil if the
// color is transparent black (i.e. no color specified).
func rgbaPtrOpt(c core.RGBA) *core.RGBA {
	if c.A == 0 && c.R == 0 && c.G == 0 && c.B == 0 {
		return nil
	}
	return &c
}

// getStringSlice extracts a []string from props under the given key.
func getStringSlice(props map[string]interface{}, key string) []string {
	if val, ok := props[key]; ok {
		if arr, ok := val.([]interface{}); ok {
			result := make([]string, len(arr))
			for i, v := range arr {
				if s, ok := v.(string); ok {
					result[i] = s
				}
			}
			return result
		}
	}
	return nil
}

// getStringSliceSlice extracts a [][]string from props under the given key.
func getStringSliceSlice(props map[string]interface{}, key string) [][]string {
	if val, ok := props[key]; ok {
		if arr, ok := val.([]interface{}); ok {
			result := make([][]string, len(arr))
			for i, v := range arr {
				if inner, ok := v.([]interface{}); ok {
					innerSlice := make([]string, len(inner))
					for j, v2 := range inner {
						if s, ok := v2.(string); ok {
							innerSlice[j] = s
						}
					}
					result[i] = innerSlice
				}
			}
			return result
		}
	}
	return nil
}

// getFloat extracts a float64 from props under the given key.
func getFloat(props map[string]interface{}, key string, defaultVal float64) float64 {
	if val, ok := props[key]; ok {
		if f, ok := val.(float64); ok {
			return f
		}
		if i, ok := val.(int); ok {
			return float64(i)
		}
	}
	return defaultVal
}

var _ = core.Renderable{}
var _ = fmt.Sprintf
