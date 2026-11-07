package react

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/Gomobra/PixTui/packages/core"
	"github.com/Gomobra/PixTui/packages/core/renderables"
	"github.com/Gomobra/PixTui/packages/plugin"
)

// ComponentFunc is the signature of a function component.
type ComponentFunc func(props map[string]interface{}) ReactElement

// ReactElement describes a node in the React element tree.
// Type is a string for host elements ("box", "text", ...) or a
// ComponentFunc for user-defined components.
type ReactElement struct {
	Type     interface{}
	Props    map[string]interface{}
	Key      string
	Ref      *Ref
	Children []ReactElement
}

// Ref is a mutable ref object whose .Current property holds a value.
type Ref struct {
	Current interface{}
}

// JSXElement represents a JSX-like element in Go. Since Go has no native JSX
// syntax, this struct is used with a builder pattern to construct element trees
// in a JSX-like manner.
type JSXElement struct {
	Type     interface{}
	Props    JSXProps
	Key      string
	Ref      *Ref
	Children []JSXElement
}

// JSXProps represents JSX element properties. It is a map of string keys to
// arbitrary values, mirroring the props object in React/TypeScript.
type JSXProps map[string]interface{}

// Fragment is the special element type used for Fragment elements. A Fragment
// renders its children without producing a host renderable of its own.
var Fragment = "fragment"

// componentRegistry maps component names to their render functions.
var componentRegistry = make(map[string]ComponentFunc)

// RegisterComponent registers a function component under a name, mirroring the
// @pixtui/react extend() API used by framework packages like @pixtui/qrcode.
func RegisterComponent(name string, fn ComponentFunc) {
	componentRegistry[name] = fn
}

// RegisteredComponent returns a previously registered component, or nil if none
// is registered under the given name.
func RegisteredComponent(name string) (ComponentFunc, bool) {
	fn, ok := componentRegistry[name]
	return fn, ok
}

// CreateJSXElement creates a JSX host element.
func CreateJSXElement(type_ string, props JSXProps, children ...JSXElement) JSXElement {
	if props == nil {
		props = JSXProps{}
	}
	return JSXElement{
		Type:     type_,
		Props:    props,
		Children: children,
	}
}

// CreateJSXComponent creates a JSX element rendered by a function component.
func CreateJSXComponent(fn ComponentFunc, props JSXProps, children ...JSXElement) JSXElement {
	if props == nil {
		props = JSXProps{}
	}
	return JSXElement{
		Type:     fn,
		Props:    props,
		Children: children,
	}
}

// CreateJSXFragment creates a Fragment element that wraps children without
// producing a host renderable.
func CreateJSXFragment(children ...JSXElement) JSXElement {
	return JSXElement{
		Type:     Fragment,
		Props:    JSXProps{},
		Children: children,
	}
}

// ToReactElement converts a JSXElement to a ReactElement for use with the
// ReactRenderer.
func (je JSXElement) ToReactElement() ReactElement {
	children := make([]ReactElement, len(je.Children))
	for i, c := range je.Children {
		children[i] = c.ToReactElement()
	}
	props := map[string]interface{}(je.Props)
	return ReactElement{
		Type:     je.Type,
		Props:    props,
		Key:      je.Key,
		Ref:      je.Ref,
		Children: children,
	}
}

// JSXBuilder provides a fluent builder pattern for constructing JSX elements.
type JSXBuilder struct {
	el JSXElement
}

// NewJSXBuilder creates a new JSXBuilder for the given element type.
func NewJSXBuilder(type_ string) *JSXBuilder {
	return &JSXBuilder{
		el: JSXElement{
			Type:  type_,
			Props: JSXProps{},
		},
	}
}

// NewJSXComponentBuilder creates a new JSXBuilder for a function component.
func NewJSXComponentBuilder(fn ComponentFunc) *JSXBuilder {
	return &JSXBuilder{
		el: JSXElement{
			Type:  fn,
			Props: JSXProps{},
		},
	}
}

// NewJSXFragmentBuilder creates a new JSXBuilder for a Fragment.
func NewJSXFragmentBuilder() *JSXBuilder {
	return &JSXBuilder{
		el: JSXElement{
			Type:  Fragment,
			Props: JSXProps{},
		},
	}
}

// Prop sets a single property on the builder and returns the builder for chaining.
func (b *JSXBuilder) Prop(key string, value interface{}) *JSXBuilder {
	b.el.Props[key] = value
	return b
}

// Props sets multiple properties on the builder and returns the builder for chaining.
func (b *JSXBuilder) Props(props JSXProps) *JSXBuilder {
	for k, v := range props {
		b.el.Props[k] = v
	}
	return b
}

// Key sets the element key and returns the builder for chaining.
func (b *JSXBuilder) Key(key string) *JSXBuilder {
	b.el.Key = key
	return b
}

// Ref sets the element ref and returns the builder for chaining.
func (b *JSXBuilder) Ref(ref *Ref) *JSXBuilder {
	b.el.Ref = ref
	return b
}

// Child adds a single child JSXElement and returns the builder for chaining.
func (b *JSXBuilder) Child(child JSXElement) *JSXBuilder {
	b.el.Children = append(b.el.Children, child)
	return b
}

// Children adds multiple child JSXElements and returns the builder for chaining.
func (b *JSXBuilder) Children(children ...JSXElement) *JSXBuilder {
	b.el.Children = append(b.el.Children, children...)
	return b
}

// Build returns the constructed JSXElement.
func (b *JSXBuilder) Build() JSXElement {
	return b.el
}

// ToReactElement converts the built JSXElement to a ReactElement.
func (b *JSXBuilder) ToReactElement() ReactElement {
	return b.el.ToReactElement()
}

// CreateElement creates a React host element.
func CreateElement(type_ string, props map[string]interface{}, children ...ReactElement) ReactElement {
	if props == nil {
		props = map[string]interface{}{}
	}
	return ReactElement{
		Type:     type_,
		Props:    props,
		Children: children,
	}
}

// CreateComponent creates a React element rendered by a function component.
func CreateComponent(fn ComponentFunc, props map[string]interface{}, children ...ReactElement) ReactElement {
	if props == nil {
		props = map[string]interface{}{}
	}
	return ReactElement{
		Type:     fn,
		Props:    props,
		Children: children,
	}
}

// CreateRef creates a ref.
func CreateRef() *Ref {
	return &Ref{}
}

// Context is a React context for distributing values to the tree.
type Context struct {
	DefaultValue interface{}
}

// CreateContext creates a context with the given default value.
func CreateContext(defaultValue interface{}) *Context {
	return &Context{DefaultValue: defaultValue}
}

// contextProviderType is the element type marker for context providers.
type contextProviderType struct {
	ctx *Context
}

// ContextProvider creates a provider element that supplies a context value
// to its descendants.
func ContextProvider(ctx *Context, value interface{}, children ...ReactElement) ReactElement {
	return ReactElement{
		Type:     &contextProviderType{ctx: ctx},
		Props:    map[string]interface{}{"value": value},
		Children: children,
	}
}

// isHostElement reports whether the element is a host element.
func isHostElement(el ReactElement) bool {
	_, ok := el.Type.(string)
	return ok
}

// isComponentElement reports whether the element is a function component.
func isComponentElement(el ReactElement) bool {
	_, ok := el.Type.(ComponentFunc)
	return ok
}

// isProviderElement reports whether the element is a context provider.
func isProviderElement(el ReactElement) bool {
	_, ok := el.Type.(*contextProviderType)
	return ok
}

// isFragmentElement reports whether the element is a Fragment.
func isFragmentElement(el ReactElement) bool {
	s, ok := el.Type.(string)
	return ok && s == Fragment
}

// componentName returns a stable name for a component function value.
func componentName(fn interface{}) string {
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		return ""
	}
	return runtimeFuncName(v.Pointer())
}

func runtimeFuncName(ptr uintptr) string {
	f := runtime.FuncForPC(ptr)
	if f == nil {
		return fmt.Sprintf("<func %#x>", ptr)
	}
	return f.Name()
}

// hookSlot identifies a hook slot on a fiber.
type hookSlot int

const (
	useStateSlot hookSlot = iota
	useEffectSlot
	useMemoSlot
	useCallbackSlot
	useRefSlot
)

// fiber is an internal node of the React fiber tree.
type fiber struct {
	renderer *ReactRenderer

	element ReactElement
	typ     interface{}
	key     string
	props   map[string]interface{}

	parent   *fiber
	children []*fiber

	dom         *core.Renderable // non-nil for host elements
	domAttached bool

	mu sync.Mutex

	// hook state
	state   []interface{}
	refs    []*Ref
	memos   []memoSlot
	cbs     []callbackSlot
	effects []effectSlot

	stateIdx int
	refIdx int
	memoIdx int
	cbIdx int
	effectIdx int

	// context values visible to this subtree
	context map[*Context]interface{}
}

type memoSlot struct {
	value interface{}
	deps  []interface{}
}

type callbackSlot struct {
	value func()
	deps  []interface{}
}

type effectSlot struct {
	fn      func() func()
	deps    []interface{}
	run     bool
	cleanup func()
}

// ReactRenderer renders React element trees to PixTui renderables using
// an incremental reconciler (fiber-based diffing) and a working hooks
// subsystem.
type ReactRenderer struct {
	ctx *core.RenderContext

	mu sync.RWMutex

	root        *fiber
	rootElement ReactElement
	lastBuffer  *core.NativeBuffer
	lastDelta   float64

	renderScheduled      bool
	rendering            bool
	pendingEffects       []pendingEffect
	pendingLayoutEffects []pendingLayoutEffect

	// Plugin support (React-specific plugins)
	pluginContext        *ReactPluginContext
	hostElementFactories map[string]ReactPluginHostElementFactory
	registeredPlugins    []ReactPlugin

	// Integration with the core plugin package (packages/plugin/plugin.go)
	pluginCtx *plugin.Context
}

type pendingEffect struct {
	fiber *fiber
	slot  int
}

type pendingLayoutEffect struct {
	fiber *fiber
	slot  int
}

// NewReactRenderer creates a React renderer bound to the given context.
func NewReactRenderer(ctx *core.RenderContext) *ReactRenderer {
	return &ReactRenderer{
		ctx:                  ctx,
		hostElementFactories: make(map[string]ReactPluginHostElementFactory),
	}
}

// SetPluginContext binds a core plugin.Context to the React renderer, allowing
// React plugins to use the core plugin registry hooks (registerRenderable,
// registerCommand, onEvent, emitEvent).
func (r *ReactRenderer) SetPluginContext(pc *plugin.Context) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.pluginCtx = pc
}

// PluginContext returns the bound core plugin.Context, if any.
func (r *ReactRenderer) PluginContext() *plugin.Context {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.pluginCtx
}

// RegisterPlugin registers a React plugin with the renderer. The plugin receives
// a ReactPluginContext through which it can register custom host elements,
// hooks, or modifiers.
func (r *ReactRenderer) RegisterPlugin(p ReactPlugin) error {
	r.mu.Lock()
	if r.pluginContext == nil {
		r.pluginContext = &ReactPluginContext{
			Renderer:  r,
			CoreCtx:   r.ctx,
			PluginCtx: r.pluginCtx,
		}
		r.hostElementFactories = make(map[string]ReactPluginHostElementFactory)
	}
	ctx := r.pluginContext
	r.mu.Unlock()

	if err := p.Setup(ctx); err != nil {
		return fmt.Errorf("failed to setup plugin %q: %w", p.Name(), err)
	}

	r.mu.Lock()
	r.registeredPlugins = append(r.registeredPlugins, p)
	r.mu.Unlock()
	return nil
}

// RegisteredPlugins returns the list of registered React plugins.
func (r *ReactRenderer) RegisteredPlugins() []ReactPlugin {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]ReactPlugin, len(r.registeredPlugins))
	copy(out, r.registeredPlugins)
	return out
}

// RegisterHostElement registers a custom host element type. Factories registered
// here take precedence over built-in elements in createHostRenderable.
func (r *ReactRenderer) RegisterHostElement(typeName string, factory ReactPluginHostElementFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hostElementFactories[typeName] = factory
}

// EmitPluginEvent emits an event through the core plugin context, if bound.
// This integrates React plugin events with the core plugin system's
// onEvent/emitEvent hooks.
func (r *ReactRenderer) EmitPluginEvent(event string, data interface{}) {
	r.mu.Lock()
	pc := r.pluginCtx
	r.mu.Unlock()
	if pc != nil && pc.EmitEvent != nil {
		pc.EmitEvent(event, data)
	}
}

// OnPluginEvent registers an event handler through the core plugin context,
// if bound. This integrates React plugin events with the core plugin system's
// onEvent/emitEvent hooks.
func (r *ReactRenderer) OnPluginEvent(event string, handler plugin.EventHandler) {
	r.mu.Lock()
	pc := r.pluginCtx
	r.mu.Unlock()
	if pc != nil && pc.OnEvent != nil {
		pc.OnEvent(event, handler)
	}
}

// paints the result into the buffer. Only changed subtrees are recreated.
func (r *ReactRenderer) Render(element ReactElement, buffer *core.NativeBuffer, deltaTime float64) {
	r.mu.Lock()
	r.rootElement = element
	r.lastBuffer = buffer
	r.lastDelta = deltaTime
	r.renderScheduled = true
	r.mu.Unlock()

	r.renderLoop()
}

// ScheduleRender requests a re-render of the root element using the last
// buffer and delta time. Safe to call from event handlers / setState.
func (r *ReactRenderer) ScheduleRender() {
	r.mu.Lock()
	if r.renderScheduled {
		r.mu.Unlock()
		return
	}
	r.renderScheduled = true
	r.mu.Unlock()

	r.renderLoop()
}

// renderLoop processes queued renders. It is re-entered whenever a state
// update schedules a new render during or after the current commit.
func (r *ReactRenderer) renderLoop() {
	for {
		// Acquire write lock to claim the render and copy input data.
		r.mu.Lock()
		if r.rendering {
			// A render is already in progress on this or another goroutine;
			// the in-progress render will pick up the scheduled flag.
			r.mu.Unlock()
			return
		}
		if !r.renderScheduled {
			r.mu.Unlock()
			return
		}
		r.rendering = true
		r.renderScheduled = false
		element := r.rootElement
		buffer := r.lastBuffer
		deltaTime := r.lastDelta
		if buffer == nil {
			r.rendering = false
			r.mu.Unlock()
			return
		}
		buffer.Clear()
		oldRoot := r.root
		r.mu.Unlock()

		// Copy plugin data under read lock so createHostRenderable doesn't need
		// to lock renderer.mu (which is NOT held during reconcile).
		factories := make(map[string]ReactPluginHostElementFactory, len(r.hostElementFactories))
		var reactPluginCtx *ReactPluginContext
		r.mu.RLock()
		for k, v := range r.hostElementFactories {
			factories[k] = v
		}
		reactPluginCtx = r.pluginContext
		r.mu.RUnlock()

		// Reconcile without holding any lock on renderer.mu.
		root := r.reconcileFiber(oldRoot, nil, element, nil, factories, reactPluginCtx)

		// Acquire write lock to publish results.
		r.mu.Lock()
		r.root = root
		pendingEffects := r.pendingEffects
		r.pendingEffects = nil
		pendingLayoutEffects := r.pendingLayoutEffects
		r.pendingLayoutEffects = nil
		r.rendering = false
		r.mu.Unlock()

		// Run layout effects synchronously before paint (useLayoutEffect semantics).
		r.runLayoutEffects(pendingLayoutEffects)

		// Commit/paint phase: walk the fiber tree, create/update renderables
		// and paint them into the buffer.
		r.paintFiber(root, nil, buffer, deltaTime)

		// Run passive effects after paint (useEffect semantics).
		r.runEffects(pendingEffects)
	}
}

func (r *ReactRenderer) withFiber(f *fiber, fn func()) {
	setCurrentFiber(f)
	fn()
	setCurrentFiber(nil)
}

// reconcileFiber updates or mounts a single fiber position.
func (r *ReactRenderer) reconcileFiber(old *fiber, parent *fiber, el ReactElement, ctxVals map[*Context]interface{}, factories map[string]ReactPluginHostElementFactory, reactPluginCtx *ReactPluginContext) *fiber {
	if old != nil && sameFiberType(old, el) {
		return r.updateFiber(old, el, ctxVals, factories, reactPluginCtx)
	}
	if old != nil {
		r.unmountFiber(old)
	}
	return r.mountFiber(parent, el, ctxVals, factories, reactPluginCtx)
}

func (r *ReactRenderer) mountFiber(parent *fiber, el ReactElement, ctxVals map[*Context]interface{}, factories map[string]ReactPluginHostElementFactory, reactPluginCtx *ReactPluginContext) *fiber {
	f := &fiber{
		renderer: r,
		element:  el,
		typ:      el.Type,
		key:      el.Key,
		props:    el.Props,
		parent:   parent,
		context:  ctxVals,
	}

	switch {
	case isProviderElement(el):
		provider := el.Type.(*contextProviderType)
		merged := make(map[*Context]interface{}, len(ctxVals)+1)
		for k, v := range ctxVals {
			merged[k] = v
		}
		if v, ok := el.Props["value"]; ok {
			merged[provider.ctx] = v
		}
		f.context = merged
		if len(el.Children) > 0 {
			f.children = r.reconcileChildren(f, el.Children, merged, factories, reactPluginCtx)
		}
	case isComponentElement(el):
		// No host renderable; the rendered output becomes the child.
		r.withFiber(f, func() {
			f.stateIdx = 0
	f.refIdx = 0
	f.memoIdx = 0
	f.cbIdx = 0
	f.effectIdx = 0
			output := el.Type.(ComponentFunc)(el.Props)
			if output.Type != nil {
				f.children = []*fiber{r.reconcileFiber(nil, f, output, f.context, factories, reactPluginCtx)}
			}
		})
	case isFragmentElement(el):
		// Fragment: no renderable of its own, just reconcile children.
		f.dom = nil
		if len(el.Children) > 0 {
			f.children = r.reconcileChildren(f, el.Children, f.context, factories, reactPluginCtx)
		}
	default:
		// Host element: create the renderable now.
		dom := createHostRenderable(el, r.ctx, factories, reactPluginCtx)
		f.dom = dom
		if parent != nil && parent.dom != nil {
			parent.dom.AddChild(dom)
			f.domAttached = true
		}
		if el.Ref != nil {
			el.Ref.Current = dom
		}
		if len(el.Children) > 0 {
			f.children = r.reconcileChildren(f, el.Children, f.context, factories, reactPluginCtx)
		}
	}

	return f
}

func (r *ReactRenderer) updateFiber(f *fiber, el ReactElement, ctxVals map[*Context]interface{}, factories map[string]ReactPluginHostElementFactory, reactPluginCtx *ReactPluginContext) *fiber {
	oldProps := f.props
	f.element = el
	f.props = el.Props
	f.key = el.Key
	f.context = ctxVals

	switch {
	case isProviderElement(el):
		provider := el.Type.(*contextProviderType)
		merged := make(map[*Context]interface{}, len(ctxVals)+1)
		for k, v := range ctxVals {
			merged[k] = v
		}
		if v, ok := el.Props["value"]; ok {
			merged[provider.ctx] = v
		}
		f.context = merged
		if len(el.Children) > 0 {
			f.children = r.reconcileChildren(f, el.Children, merged, factories, reactPluginCtx)
		} else if len(f.children) > 0 {
			f.children = r.reconcileChildren(f, nil, merged, factories, reactPluginCtx)
		}
	case isComponentElement(el):
		r.withFiber(f, func() {
			f.stateIdx = 0
	f.refIdx = 0
	f.memoIdx = 0
	f.cbIdx = 0
	f.effectIdx = 0
			output := el.Type.(ComponentFunc)(el.Props)
			var oldChild *fiber
			if len(f.children) > 0 {
				oldChild = f.children[0]
			}
			child := r.reconcileFiber(oldChild, f, output, f.context, factories, reactPluginCtx)
			f.children = []*fiber{child}
		})
	case isFragmentElement(el):
		// Fragment: reconcile children without a renderable.
		f.dom = nil
		if len(el.Children) > 0 {
			f.children = r.reconcileChildren(f, el.Children, f.context, factories, reactPluginCtx)
		} else if len(f.children) > 0 {
			f.children = r.reconcileChildren(f, nil, f.context, factories, reactPluginCtx)
		}
	default:
		propsChanged := !deepEqualProps(oldProps, el.Props)
		if propsChanged && f.dom != nil {
			if p := f.dom.Parent(); p != nil {
				p.RemoveChild(f.dom)
			}
			dom := createHostRenderable(el, r.ctx, factories, reactPluginCtx)
			f.dom = dom
			f.domAttached = false
			if p := f.parent; p != nil && p.dom != nil {
				p.dom.AddChild(dom)
				f.domAttached = true
			}
		}
		if el.Ref != nil && f.dom != nil {
			el.Ref.Current = f.dom
		}
		if len(el.Children) > 0 {
			f.children = r.reconcileChildren(f, el.Children, f.context, factories, reactPluginCtx)
		} else if len(f.children) > 0 {
			f.children = r.reconcileChildren(f, nil, f.context, factories, reactPluginCtx)
		}
	}

	return f
}

// reconcileChildren diffs a list of new children against a fiber's current
// children, keyed by element key (falling back to index for keyless children),
// reusing, updating, mounting or unmounting fibers as needed.
func (r *ReactRenderer) reconcileChildren(parent *fiber, newChildren []ReactElement, ctxVals map[*Context]interface{}, factories map[string]ReactPluginHostElementFactory, reactPluginCtx *ReactPluginContext) []*fiber {
	old := parent.children

	// Index old children by key.
	oldByKey := make(map[string]*fiber, len(old))
	for _, c := range old {
		if c.key != "" {
			oldByKey[c.key] = c
		}
	}

	used := make(map[*fiber]bool, len(old))
	next := make([]*fiber, 0, len(newChildren))

	oldIdx := 0 // scan position for keyless reconciliation
	for i, el := range newChildren {
		key := el.Key
		if key == "" {
			key = fmt.Sprintf("@%d", i)
		}

		var reuse *fiber
		if c, ok := oldByKey[key]; ok && !used[c] && sameFiberType(c, el) {
			reuse = c
		} else if len(key) > 0 && key[0] == '@' {
			// Keyless: consume old children in order until a type match.
			for oldIdx < len(old) {
				c := old[oldIdx]
				oldIdx++
				if used[c] || c.key != "" {
					continue
				}
				if sameFiberType(c, el) {
					reuse = c
					break
				}
				r.unmountFiber(c)
			}
		}

		if reuse != nil {
			used[reuse] = true
			next = append(next, r.updateFiber(reuse, el, ctxVals, factories, reactPluginCtx))
			continue
		}
		next = append(next, r.mountFiber(parent, el, ctxVals, factories, reactPluginCtx))
	}

	// Unmount any old children that were not reused.
	for _, c := range old {
		if !used[c] {
			r.unmountFiber(c)
		}
	}

	parent.children = next
	return next
}

func (r *ReactRenderer) unmountFiber(f *fiber) {
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

// paintFiber walks the fiber tree and paints every host renderable into the
// buffer. Parent renderables are painted before children.
func (r *ReactRenderer) paintFiber(f *fiber, parentDom *core.Renderable, buffer *core.NativeBuffer, deltaTime float64) {
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

func (r *ReactRenderer) scheduleEffect(f *fiber, slot int) {
	// Called during render while r.mu is held; no locking needed here to
	// avoid a self-deadlock on the non-reentrant sync.Mutex.
	r.pendingEffects = append(r.pendingEffects, pendingEffect{fiber: f, slot: slot})
}

func (r *ReactRenderer) scheduleLayoutEffect(f *fiber, slot int) {
	// Called during render while r.mu is held; no locking needed here to
	// avoid a self-deadlock on the non-reentrant sync.Mutex.
	r.pendingLayoutEffects = append(r.pendingLayoutEffects, pendingLayoutEffect{fiber: f, slot: slot})
}

func (r *ReactRenderer) runEffects(pendingEffects []pendingEffect) {
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
		if fn != nil {
			cleanup := fn()
			pe.fiber.mu.Lock()
			eff.cleanup = cleanup
			pe.fiber.mu.Unlock()
		}
	}
}

func (r *ReactRenderer) runLayoutEffects(pendingEffects []pendingLayoutEffect) {
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
		if fn != nil {
			cleanup := fn()
			pe.fiber.mu.Lock()
			eff.cleanup = cleanup
			pe.fiber.mu.Unlock()
		}
	}
}

// ----- React hooks (fiber-bound) -----

func (f *fiber) useState(initial interface{}) (interface{}, func(interface{})) {
	f.mu.Lock()
	idx := f.stateIdx
	f.stateIdx++

	if idx >= len(f.state) {
		value := initial
		if init, ok := initial.(func() interface{}); ok {
			value = init()
		}
		f.state = append(f.state, value)
	}
	value := f.state[idx]
	f.mu.Unlock()

	setValue := func(newValue interface{}) {
		var resolved interface{} = newValue
		if updater, ok := newValue.(func(interface{}) interface{}); ok {
			f.mu.Lock()
			resolved = updater(f.state[idx])
			f.mu.Unlock()
		}
		f.mu.Lock()
		f.state[idx] = resolved
		f.mu.Unlock()
		f.renderer.ScheduleRender()
	}
	return value, setValue
}

func (f *fiber) useEffect(effect func() func(), deps []interface{}) {
	f.mu.Lock()
	idx := f.effectIdx
	f.effectIdx++

	var eff *effectSlot
	if idx < len(f.effects) {
		eff = &f.effects[idx]
	} else {
		f.effects = append(f.effects, effectSlot{})
		eff = &f.effects[idx]
	}

	// Jadwalkan effect hanya saat mounting pertama kali atau ketika
	// dependensi berubah. Ini mencerminkan React useEffect: effect berjalan
	// sekali setelah commit render dan dijalankan ulang hanya jika deps
	// berubah. Fungsi effect (dan cleanup-nya yang tertunda) hanya diganti
	// ketika effect benar-benar dijalankan ulang, sehingga render yang tidak
	// mengubah deps adalah no-op dan closure yang tersimpan dipertahankan.
	shouldRun := !eff.run || !depsEqual(eff.deps, deps)
	if shouldRun {
		eff.fn = effect
		eff.deps = copyDeps(deps)
		r := f.renderer
		r.scheduleEffect(f, idx)
	}
	f.mu.Unlock()
}

func (f *fiber) useLayoutEffect(effect func() func(), deps []interface{}) {
	f.mu.Lock()
	idx := f.effectIdx
	f.effectIdx++

	var eff *effectSlot
	if idx < len(f.effects) {
		eff = &f.effects[idx]
	} else {
		f.effects = append(f.effects, effectSlot{})
		eff = &f.effects[idx]
	}

	// Layout effect berjalan sinkron setelah commit render, sebelum paint
	// (useLayoutEffect semantics). Seperti useEffect berjalan sekali saat
	// mount dan dijalankan ulang hanya jika dependensi berubah.
	shouldRun := !eff.run || !depsEqual(eff.deps, deps)
	if shouldRun {
		eff.fn = effect
		eff.deps = copyDeps(deps)
		r := f.renderer
		r.scheduleLayoutEffect(f, idx)
	}
	f.mu.Unlock()
}

// useEffectEvent returns a stable callback that always invokes the latest
// handler. The handler reference is updated via a layout effect after each
// commit, and the returned callback is memoized with an empty dependency
// array so its identity never changes. This mirrors React's useEffectEvent.
// Useful for event handlers passed to effects with empty dependency arrays
// or memoized child components.
func (f *fiber) useEffectEvent(handler func()) func() {
	handlerRef := f.useRef(handler)

	f.useLayoutEffect(func() func() {
		handlerRef.Current = handler
		return nil
	}, nil)

	return f.useCallback(func() {
		if fn, ok := handlerRef.Current.(func()); ok && fn != nil {
			fn()
		}
	}, nil)
}

func (f *fiber) useMemo(memo func() interface{}, deps []interface{}) interface{} {
	f.mu.Lock()
	idx := f.memoIdx
	f.memoIdx++

	var slot *memoSlot
	isNew := false
	if idx < len(f.memos) {
		slot = &f.memos[idx]
	} else {
		f.memos = append(f.memos, memoSlot{})
		slot = &f.memos[idx]
		isNew = true
	}

	if isNew || !depsEqual(slot.deps, deps) {
		if memo != nil {
			slot.value = memo()
		}
		slot.deps = copyDeps(deps)
	}
	value := slot.value
	f.mu.Unlock()
	return value
}

func (f *fiber) useCallback(callback func(), deps []interface{}) func() {
	f.mu.Lock()
	idx := f.cbIdx
	f.cbIdx++

	var slot *callbackSlot
	isNew := false
	if idx < len(f.cbs) {
		slot = &f.cbs[idx]
	} else {
		f.cbs = append(f.cbs, callbackSlot{})
		slot = &f.cbs[idx]
		isNew = true
	}

	if isNew || !depsEqual(slot.deps, deps) {
		if callback != nil {
			slot.value = callback
		}
		slot.deps = copyDeps(deps)
	}
	value := slot.value
	f.mu.Unlock()
	return value
}

func (f *fiber) useRef(initial interface{}) *Ref {
	f.mu.Lock()
	idx := f.refIdx
	f.refIdx++

	if idx >= len(f.refs) {
		f.refs = append(f.refs, &Ref{Current: initial})
	}
	ref := f.refs[idx]
	f.mu.Unlock()
	return ref
}

func (f *fiber) useContext(c *Context) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.context != nil {
		if v, ok := f.context[c]; ok {
			return v
		}
	}
	return c.DefaultValue
}

// ----- Package-level hooks -----

var (
	currentFiberMu sync.Mutex
	currentFiber   *fiber
)

func setCurrentFiber(f *fiber) {
	currentFiberMu.Lock()
	currentFiber = f
	currentFiberMu.Unlock()
}

func getCurrentFiber() *fiber {
	currentFiberMu.Lock()
	defer currentFiberMu.Unlock()
	return currentFiber
}

// UseState returns the current state and a setter for the currently
// rendering fiber. Outside of a render it returns the initial value and a
// no-op setter.
func UseState(initial interface{}) (interface{}, func(interface{})) {
	if f := getCurrentFiber(); f != nil {
		return f.useState(initial)
	}
	return initial, func(interface{}) {}
}

// UseEffect schedules an effect to run after the current render commit
// (passive, after paint), re-running only when the dependencies change.
// The effect function may return a cleanup function that is run before the
// effect re-runs or when the component unmounts. Outside of a render it runs
// immediately.
func UseEffect(effect func() func(), deps []interface{}) {
	if f := getCurrentFiber(); f != nil {
		f.useEffect(effect, deps)
		return
	}
	effect()
}

// UseLayoutEffect schedules an effect to run synchronously after the render
// commit, before paint (layout phase). It re-runs only when the dependencies
// change. The effect function may return a cleanup function. Outside of a
// render it runs immediately.
func UseLayoutEffect(effect func() func(), deps []interface{}) {
	if f := getCurrentFiber(); f != nil {
		f.useLayoutEffect(effect, deps)
		return
	}
	effect()
}

// UseEffectEvent returns a stable callback that always calls the latest
// version of the provided handler. This prevents unnecessary re-renders and
// effect re-runs while ensuring the callback always has access to the latest
// props and state.
//
// Useful for event handlers that need to be passed to effects with empty
// dependency arrays or memoized child components. Outside of a render it
// returns the handler itself.
func UseEffectEvent(handler func()) func() {
	if f := getCurrentFiber(); f != nil {
		return f.useEffectEvent(handler)
	}
	return handler
}

// UseMemo returns a memoized value, recomputed only when deps change.
func UseMemo(memo func() interface{}, deps []interface{}) interface{} {
	if f := getCurrentFiber(); f != nil {
		return f.useMemo(memo, deps)
	}
	return memo()
}

// UseCallback returns a memoized callback, recreated only when deps change.
func UseCallback(callback func(), deps []interface{}) func() {
	if f := getCurrentFiber(); f != nil {
		return f.useCallback(callback, deps)
	}
	return callback
}

// UseRef returns a stable ref object for the lifetime of the fiber.
func UseRef(initial interface{}) *Ref {
	if f := getCurrentFiber(); f != nil {
		return f.useRef(initial)
	}
	return &Ref{Current: initial}
}

// UseContext returns the context value for the current fiber, or the context
// default if no provider is in scope.
func UseContext(c *Context) interface{} {
	if f := getCurrentFiber(); f != nil {
		return f.useContext(c)
	}
	return c.DefaultValue
}

// Hooks is retained for API compatibility; package-level hooks are the
// canonical entry point.
type Hooks struct{}

// NewHooks creates a hooks container.
func NewHooks() *Hooks {
	return &Hooks{}
}

// ----- Host element creation -----

func createHostRenderable(el ReactElement, ctx *core.RenderContext, factories map[string]ReactPluginHostElementFactory, reactPluginCtx *ReactPluginContext) *core.Renderable {
	// Check plugin-registered host element factories first.
	factory, hasFactory := factories[el.Type.(string)]

	if hasFactory && factory != nil {
		dom, err := factory(el, ctx, reactPluginCtx)
		if err != nil {
			// Fall back to default renderable on error.
			return core.NewRenderable(ctx, el.Key, core.RenderableOptions[core.Renderable]{})
		}
		if dom != nil {
			return dom
		}
	}

	switch el.Type.(string) {
	case "box":
		return createBox(el, ctx)
	case "text":
		return createText(el, ctx)
	case "input":
		return createInput(el, ctx)
	case "select":
		return createSelect(el, ctx)
	case "scrollbox":
		return createScrollBox(el, ctx)
	case "textarea":
		return createTextarea(el, ctx)
	case "slider":
		return createSlider(el, ctx)
	case "tabselect":
		return createTabSelect(el, ctx)
	case "image":
		return createImage(el, ctx)
	case "video":
		return createVideo(el, ctx)
	case "code":
		return createCode(el, ctx)
	case "markdown":
		return createMarkdown(el, ctx)
	case "diff":
		return createDiff(el, ctx)
	case "line-number":
		return createLineNumber(el, ctx)
	case "scrollbar":
		return createScrollBar(el, ctx)
	case "texttable":
		return createTextTable(el, ctx)
	case "ascii-font":
		return createASCIIFont(el, ctx)
	case "frame-buffer":
		return createFrameBuffer(el, ctx)
	case "edit-buffer":
		return createEditBuffer(el, ctx)
	// Additional types from the TypeScript component catalogue.
	case "span":
		return createSpan(el, ctx)
	case "link":
		return createLink(el, ctx)
	case "line-break":
		return createLineBreak(el, ctx)
	default:
		return core.NewRenderable(ctx, el.Key, core.RenderableOptions[core.Renderable]{})
	}
}

func createBox(el ReactElement, ctx *core.RenderContext) *core.Renderable {
	opts := renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      getString(el.Props, "id", ""),
			Width:   strPtr(getString(el.Props, "width", "")),
			Height:  strPtr(getString(el.Props, "height", "")),
			Visible: boolPtr(getBool(el.Props, "visible", true)),
			ZIndex:  intPtr(getInt(el.Props, "zIndex", 0)),
		},
		BackgroundColor:    hexPtrColor(getString(el.Props, "backgroundColor", "transparent")),
		BorderStyle:        getString(el.Props, "borderStyle", "single"),
		Border:             getBool(el.Props, "border", false),
		BorderColor:        hexPtrColor(getString(el.Props, "borderColor", "#FFFFFF")),
		ShouldFill:         getBool(el.Props, "shouldFill", false),
		Title:              getString(el.Props, "title", ""),
		TitleColor:         hexPtrColor(getString(el.Props, "titleColor", "")),
		TitleAlignment:     getString(el.Props, "titleAlignment", "left"),
		BottomTitle:        getString(el.Props, "bottomTitle", ""),
		Focusable:          getBool(el.Props, "focusable", false),
		FocusedBorderColor: hexPtrColor(getString(el.Props, "focusedBorderColor", "#00AAFF")),
	}
	r := renderables.NewBoxRenderable(ctx, opts)
	return r.Renderable
}

func createText(el ReactElement, ctx *core.RenderContext) *core.Renderable {
	opts := renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      getString(el.Props, "id", ""),
			Width:   strPtr(getString(el.Props, "width", "")),
			Height:  strPtr(getString(el.Props, "height", "")),
			Visible: boolPtr(getBool(el.Props, "visible", true)),
			ZIndex:  intPtr(getInt(el.Props, "zIndex", 0)),
		},
		Content: getString(el.Props, "content", ""),
	}
	r := renderables.NewTextRenderable(ctx, opts)
	return r.Renderable
}

func createInput(el ReactElement, ctx *core.RenderContext) *core.Renderable {
	opts := renderables.InputOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      getString(el.Props, "id", ""),
			Width:   strPtr(getString(el.Props, "width", "")),
			Height:  strPtr(getString(el.Props, "height", "")),
			Visible: boolPtr(getBool(el.Props, "visible", true)),
			ZIndex:  intPtr(getInt(el.Props, "zIndex", 0)),
		},
		Value:       getString(el.Props, "value", ""),
		Placeholder: getString(el.Props, "placeholder", ""),
	}
	r := renderables.NewInputRenderable(ctx, opts)
	return r.Renderable
}

func createSelect(el ReactElement, ctx *core.RenderContext) *core.Renderable {
	opts := renderables.SelectOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      getString(el.Props, "id", ""),
			Width:   strPtr(getString(el.Props, "width", "")),
			Height:  strPtr(getString(el.Props, "height", "")),
			Visible: boolPtr(getBool(el.Props, "visible", true)),
			ZIndex:  intPtr(getInt(el.Props, "zIndex", 0)),
		},
		Options:  getStringSlice(el.Props, "options"),
		Selected: getInt(el.Props, "selected", 0),
	}
	r := renderables.NewSelectRenderable(ctx, opts)
	return r.Renderable
}

func createScrollBox(el ReactElement, ctx *core.RenderContext) *core.Renderable {
	opts := renderables.ScrollBoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      getString(el.Props, "id", ""),
			Width:   strPtr(getString(el.Props, "width", "")),
			Height:  strPtr(getString(el.Props, "height", "")),
			Visible: boolPtr(getBool(el.Props, "visible", true)),
			ZIndex:  intPtr(getInt(el.Props, "zIndex", 0)),
		},
	}
	r := renderables.NewScrollBoxRenderable(ctx, opts)
	return r.Renderable
}

func createTextarea(el ReactElement, ctx *core.RenderContext) *core.Renderable {
	opts := renderables.TextareaOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      getString(el.Props, "id", ""),
			Width:   strPtr(getString(el.Props, "width", "")),
			Height:  strPtr(getString(el.Props, "height", "")),
			Visible: boolPtr(getBool(el.Props, "visible", true)),
			ZIndex:  intPtr(getInt(el.Props, "zIndex", 0)),
		},
		Value: getString(el.Props, "value", ""),
	}
	r := renderables.NewTextareaRenderable(ctx, opts)
	return r.Renderable
}

func createSlider(el ReactElement, ctx *core.RenderContext) *core.Renderable {
	opts := renderables.SliderOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      getString(el.Props, "id", ""),
			Width:   strPtr(getString(el.Props, "width", "")),
			Height:  strPtr(getString(el.Props, "height", "")),
			Visible: boolPtr(getBool(el.Props, "visible", true)),
			ZIndex:  intPtr(getInt(el.Props, "zIndex", 0)),
		},
		Min:   getFloat(el.Props, "min", 0),
		Max:   getFloat(el.Props, "max", 100),
		Value: getFloat(el.Props, "value", 0),
	}
	r := renderables.NewSliderRenderable(ctx, opts)
	return r.Renderable
}

func createTabSelect(el ReactElement, ctx *core.RenderContext) *core.Renderable {
	opts := renderables.TabSelectOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      getString(el.Props, "id", ""),
			Width:   strPtr(getString(el.Props, "width", "")),
			Height:  strPtr(getString(el.Props, "height", "")),
			Visible: boolPtr(getBool(el.Props, "visible", true)),
			ZIndex:  intPtr(getInt(el.Props, "zIndex", 0)),
		},
		Tabs:     getStringSlice(el.Props, "tabs"),
		Selected: getInt(el.Props, "selected", 0),
	}
	r := renderables.NewTabSelectRenderable(ctx, opts)
	return r.Renderable
}

func createImage(el ReactElement, ctx *core.RenderContext) *core.Renderable {
	opts := renderables.ImageOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      getString(el.Props, "id", ""),
			Width:   strPtr(getString(el.Props, "width", "")),
			Height:  strPtr(getString(el.Props, "height", "")),
			Visible: boolPtr(getBool(el.Props, "visible", true)),
			ZIndex:  intPtr(getInt(el.Props, "zIndex", 0)),
		},
		Path:     getString(el.Props, "path", ""),
		Protocol: core.ImageRenderProtocolAuto,
	}
	r := renderables.NewImageRenderable(ctx, opts)
	return r.Renderable
}

func createVideo(el ReactElement, ctx *core.RenderContext) *core.Renderable {
	opts := renderables.VideoOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      getString(el.Props, "id", ""),
			Width:   strPtr(getString(el.Props, "width", "")),
			Height:  strPtr(getString(el.Props, "height", "")),
			Visible: boolPtr(getBool(el.Props, "visible", true)),
			ZIndex:  intPtr(getInt(el.Props, "zIndex", 0)),
		},
		AutoPlay: getBool(el.Props, "autoPlay", false),
	}
	r := renderables.NewVideoRenderable(ctx, opts)
	return r.Renderable
}

func createCode(el ReactElement, ctx *core.RenderContext) *core.Renderable {
	opts := renderables.CodeOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      getString(el.Props, "id", ""),
			Width:   strPtr(getString(el.Props, "width", "")),
			Height:  strPtr(getString(el.Props, "height", "")),
			Visible: boolPtr(getBool(el.Props, "visible", true)),
			ZIndex:  intPtr(getInt(el.Props, "zIndex", 0)),
		},
		Code:     getString(el.Props, "code", ""),
		Language: getString(el.Props, "language", ""),
	}
	r := renderables.NewCodeRenderable(ctx, opts)
	return r.Renderable
}

func createMarkdown(el ReactElement, ctx *core.RenderContext) *core.Renderable {
	opts := renderables.MarkdownOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      getString(el.Props, "id", ""),
			Width:   strPtr(getString(el.Props, "width", "")),
			Height:  strPtr(getString(el.Props, "height", "")),
			Visible: boolPtr(getBool(el.Props, "visible", true)),
			ZIndex:  intPtr(getInt(el.Props, "zIndex", 0)),
		},
		Content: getString(el.Props, "content", ""),
	}
	r := renderables.NewMarkdownRenderable(ctx, opts)
	return r.Renderable
}

func createDiff(el ReactElement, ctx *core.RenderContext) *core.Renderable {
	opts := renderables.DiffOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      getString(el.Props, "id", ""),
			Width:   strPtr(getString(el.Props, "width", "")),
			Height:  strPtr(getString(el.Props, "height", "")),
			Visible: boolPtr(getBool(el.Props, "visible", true)),
			ZIndex:  intPtr(getInt(el.Props, "zIndex", 0)),
		},
		Content: getString(el.Props, "content", ""),
	}
	r := renderables.NewDiffRenderable(ctx, opts)
	return r.Renderable
}

func createLineNumber(el ReactElement, ctx *core.RenderContext) *core.Renderable {
	r := renderables.NewLineNumberRenderable(ctx, getString(el.Props, "id", ""), getInt(el.Props, "startLine", 1))
	return r.Renderable
}

func createScrollBar(el ReactElement, ctx *core.RenderContext) *core.Renderable {
	opts := renderables.ScrollBarOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      getString(el.Props, "id", ""),
			Width:   strPtr(getString(el.Props, "width", "")),
			Height:  strPtr(getString(el.Props, "height", "")),
			Visible: boolPtr(getBool(el.Props, "visible", true)),
			ZIndex:  intPtr(getInt(el.Props, "zIndex", 0)),
		},
		Vertical:   getBool(el.Props, "vertical", true),
		ThumbSize:  getFloat(el.Props, "thumbSize", 3),
		ThumbStart: getFloat(el.Props, "thumbStart", 0),
	}
	r := renderables.NewScrollBarRenderable(ctx, opts)
	return r.Renderable
}

func createTextTable(el ReactElement, ctx *core.RenderContext) *core.Renderable {
	opts := renderables.TextTableOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      getString(el.Props, "id", ""),
			Width:   strPtr(getString(el.Props, "width", "")),
			Height:  strPtr(getString(el.Props, "height", "")),
			Visible: boolPtr(getBool(el.Props, "visible", true)),
			ZIndex:  intPtr(getInt(el.Props, "zIndex", 0)),
		},
		Headers: getStringSlice(el.Props, "headers"),
		Rows:    getStringSliceSlice(el.Props, "rows"),
	}
	r := renderables.NewTextTableRenderable(ctx, opts)
	return r.Renderable
}

func createASCIIFont(el ReactElement, ctx *core.RenderContext) *core.Renderable {
	opts := renderables.ASCIIFontOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      getString(el.Props, "id", ""),
			Width:   strPtr(getString(el.Props, "width", "")),
			Height:  strPtr(getString(el.Props, "height", "")),
			Visible: boolPtr(getBool(el.Props, "visible", true)),
			ZIndex:  intPtr(getInt(el.Props, "zIndex", 0)),
		},
		Text: getString(el.Props, "text", ""),
		Font: getString(el.Props, "font", ""),
	}
	r := renderables.NewASCIIFontRenderable(ctx, opts)
	return r.Renderable
}

func createFrameBuffer(el ReactElement, ctx *core.RenderContext) *core.Renderable {
	opts := renderables.FrameBufferOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      getString(el.Props, "id", ""),
			Width:   strPtr(getString(el.Props, "width", "")),
			Height:  strPtr(getString(el.Props, "height", "")),
			Visible: boolPtr(getBool(el.Props, "visible", true)),
			ZIndex:  intPtr(getInt(el.Props, "zIndex", 0)),
		},
	}
	r := renderables.NewFrameBufferRenderable(ctx, opts)
	return r.Renderable
}

func createEditBuffer(el ReactElement, ctx *core.RenderContext) *core.Renderable {
	opts := renderables.EditBufferRenderableOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      getString(el.Props, "id", ""),
			Width:   strPtr(getString(el.Props, "width", "")),
			Height:  strPtr(getString(el.Props, "height", "")),
			Visible: boolPtr(getBool(el.Props, "visible", true)),
			ZIndex:  intPtr(getInt(el.Props, "zIndex", 0)),
		},
		Value:       getString(el.Props, "value", ""),
		Placeholder: getString(el.Props, "placeholder", ""),
		ReadOnly:    getBool(el.Props, "readOnly", false),
	}
	r := renderables.NewEditBufferRenderable(ctx, opts)
	return r.Renderable
}

func createSpan(el ReactElement, ctx *core.RenderContext) *core.Renderable {
	// A span is an inline text element, rendered like Text.
	opts := renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      getString(el.Props, "id", ""),
			Width:   strPtr(getString(el.Props, "width", "")),
			Height:  strPtr(getString(el.Props, "height", "")),
			Visible: boolPtr(getBool(el.Props, "visible", true)),
			ZIndex:  intPtr(getInt(el.Props, "zIndex", 0)),
		},
		Content: getString(el.Props, "content", ""),
	}
	r := renderables.NewTextRenderable(ctx, opts)
	return r.Renderable
}

func createLink(el ReactElement, ctx *core.RenderContext) *core.Renderable {
	// A link is an inline text element with an href, rendered like Text.
	opts := renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      getString(el.Props, "id", ""),
			Width:   strPtr(getString(el.Props, "width", "")),
			Height:  strPtr(getString(el.Props, "height", "")),
			Visible: boolPtr(getBool(el.Props, "visible", true)),
			ZIndex:  intPtr(getInt(el.Props, "zIndex", 0)),
		},
		Content: getString(el.Props, "content", getString(el.Props, "href", "")),
	}
	r := renderables.NewTextRenderable(ctx, opts)
	return r.Renderable
}

func createLineBreak(el ReactElement, ctx *core.RenderContext) *core.Renderable {
	// A line break is rendered as a Text element containing a newline.
	opts := renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      getString(el.Props, "id", ""),
			Width:   strPtr(getString(el.Props, "width", "")),
			Height:  strPtr(getString(el.Props, "height", "")),
			Visible: boolPtr(getBool(el.Props, "visible", true)),
			ZIndex:  intPtr(getInt(el.Props, "zIndex", 0)),
		},
		Content: "\n",
	}
	r := renderables.NewTextRenderable(ctx, opts)
	return r.Renderable
}

// hexPtrColor parses a CSS hex color string (#RRGGBB or #RRGGBBAA) into a
// *core.RGBA. It returns nil when the string cannot be parsed.
func hexPtrColor(s string) *core.RGBA {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "#") {
		s = s[1:]
	}
	var r, g, b, a uint8
	a = 255
	ok := true
	switch len(s) {
	case 6: // #RRGGBB
		if v, err := hexByte(s[0:2]); err == nil {
			r = v
		} else {
			ok = false
		}
		if v, err := hexByte(s[2:4]); err == nil {
			g = v
		} else {
			ok = false
		}
		if v, err := hexByte(s[4:6]); err == nil {
			b = v
		} else {
			ok = false
		}
	case 8: // #RRGGBBAA
		if v, err := hexByte(s[0:2]); err == nil {
			r = v
		} else {
			ok = false
		}
		if v, err := hexByte(s[2:4]); err == nil {
			g = v
		} else {
			ok = false
		}
		if v, err := hexByte(s[4:6]); err == nil {
			b = v
		} else {
			ok = false
		}
		if v, err := hexByte(s[6:8]); err == nil {
			a = v
		} else {
			ok = false
		}
	default:
		return nil
	}
	if !ok {
		return nil
	}
	return &core.RGBA{R: r, G: g, B: b, A: a}
}

// errInvalidHex is returned when a CSS-style hex color cannot be parsed.
var errInvalidHex = errors.New("invalid hex color")

// hexByte parses exactly two hexadecimal digits into a uint8.
func hexByte(s string) (uint8, error) {
	if len(s) != 2 {
		return 0, errInvalidHex
	}
	v, err := strconv.ParseUint(s, 16, 8)
	return uint8(v), err
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

func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }
func intPtr(i int) *int       { return &i }

// nativeImageProtocolAuto mirrors native.ImageProtocolAuto without importing
// native here (to avoid a circular-style dependency on the concrete protocol).
var nativeImageProtocolAuto = "auto"

// ----- Diffing helpers -----

func sameFiberType(a *fiber, b ReactElement) bool {
	aHost := isHostElement(ReactElement{Type: a.typ})
	bHost := isHostElement(b)
	if aHost != bHost {
		return false
	}
	if aHost {
		as, ok := a.typ.(string)
		if !ok {
			return false
		}
		bs, ok := b.Type.(string)
		if !ok {
			return false
		}
		return as == bs
	}
	return componentName(a.typ) == componentName(b.Type)
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

// ============================================================================
// React Plugin Support
// ============================================================================

// ReactPlugin is the interface that React plugins must implement. Plugins can
// register custom host elements, hooks, or modifiers through the
// ReactPluginContext provided during Setup.
type ReactPlugin interface {
	// Name returns the unique plugin name.
	Name() string
	// Setup initializes the plugin with the given ReactPluginContext.
	Setup(ctx *ReactPluginContext) error
	// Shutdown cleans up plugin resources.
	Shutdown() error
}

// ReactPluginHostElementFactory creates a custom host renderable from a React
// element. It is used by plugins to register custom host element types.
type ReactPluginHostElementFactory func(el ReactElement, ctx *core.RenderContext, pluginCtx *ReactPluginContext) (*core.Renderable, error)

// ReactPluginModifier is a function that transforms JSXProps before they are
// applied to a host element. Plugins can register modifiers to intercept and
// modify props.
type ReactPluginModifier func(props JSXProps) JSXProps

// ReactPluginHookHandler is a generic handler for plugin-registered hooks.
type ReactPluginHookHandler func(args ...interface{}) interface{}

// ReactPluginContext provides plugins with access to the React renderer, the
// core render context, the core plugin context, and element tree manipulation.
type ReactPluginContext struct {
	// Renderer is the ReactRenderer this plugin is registered with.
	Renderer *ReactRenderer
	// CoreCtx is the core RenderContext.
	CoreCtx *core.RenderContext
	// PluginCtx is the bound core plugin.Context (may be nil if not set via
	// ReactRenderer.SetPluginContext).
	PluginCtx *plugin.Context

	// RegisterHostElement registers a custom host element type. The factory is
	// called by createHostRenderable when an element of the given type is
	// encountered.
	RegisterHostElement func(typeName string, factory ReactPluginHostElementFactory)
	// RegisterModifier registers a prop modifier function. Modifiers are applied
	// to JSXProps before host element creation.
	RegisterModifier func(name string, modifier ReactPluginModifier)
	// RegisterHook registers a named hook handler.
	RegisterHook func(name string, handler ReactPluginHookHandler)
}

// reactPluginContextImpl is the internal implementation of ReactPluginContext.
type reactPluginContextImpl struct {
	renderer  *ReactRenderer
	coreCtx   *core.RenderContext
	pluginCtx *plugin.Context
	modifiers map[string]ReactPluginModifier
	hooks     map[string]ReactPluginHookHandler
}

// newReactPluginContext creates a new ReactPluginContext.
func newReactPluginContext(r *ReactRenderer) *reactPluginContextImpl {
	return &reactPluginContextImpl{
		renderer:  r,
		coreCtx:   r.ctx,
		pluginCtx: r.pluginCtx,
		modifiers: make(map[string]ReactPluginModifier),
		hooks:     make(map[string]ReactPluginHookHandler),
	}
}

func (c *reactPluginContextImpl) Renderer() *ReactRenderer     { return c.renderer }
func (c *reactPluginContextImpl) CoreCtx() *core.RenderContext { return c.coreCtx }
func (c *reactPluginContextImpl) PluginCtx() *plugin.Context   { return c.pluginCtx }

func (c *reactPluginContextImpl) RegisterHostElement(typeName string, factory ReactPluginHostElementFactory) {
	c.renderer.mu.Lock()
	defer c.renderer.mu.Unlock()
	c.renderer.hostElementFactories[typeName] = factory
}

func (c *reactPluginContextImpl) RegisterModifier(name string, modifier ReactPluginModifier) {
	c.renderer.mu.Lock()
	defer c.renderer.mu.Unlock()
	c.modifiers[name] = modifier
}

func (c *reactPluginContextImpl) RegisterHook(name string, handler ReactPluginHookHandler) {
	c.renderer.mu.Lock()
	defer c.renderer.mu.Unlock()
	c.hooks[name] = handler
}

// applyModifiers applies all registered modifiers to the given props.
type JSXResult struct {
	Props     JSXProps
	Modifiers []string
}

func applyModifiers(modifiers map[string]ReactPluginModifier, props JSXProps) JSXResult {
	result := JSXResult{Props: props}
	names := make([]string, 0, len(modifiers))
	for n := range modifiers {
		names = append(names, n)
	}
	sortStrings(names)
	for _, n := range names {
		if m, ok := modifiers[n]; ok {
			result.Props = m(result.Props)
			result.Modifiers = append(result.Modifiers, n)
		}
	}
	return result
}

func sortStrings(s []string) {
	for i := 0; i < len(s); i++ {
		for j := i + 1; j < len(s); j++ {
			if s[i] > s[j] {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}

// callPluginHook invokes a registered plugin hook by name, if any.
func callPluginHook(hooks map[string]ReactPluginHookHandler, name string, args ...interface{}) (interface{}, bool) {
	h, ok := hooks[name]
	if !ok {
		return nil, false
	}
	return h(args...), true
}
