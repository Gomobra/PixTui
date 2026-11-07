package keymap

import (
	"sync"
)

// ---------------------------------------------------------------------------
// Testing utilities (mirrors @pixtui/keymap/src/testing)
// ---------------------------------------------------------------------------

// TestKeyModifierOptions are optional modifiers for a test key event.
type TestKeyModifierOptions struct {
	Ctrl    bool
	Shift   bool
	Meta    bool
	Super   bool
	Hyper   bool
}

// TestHostMetadataOptions are options for building test host metadata.
type TestHostMetadataOptions struct {
	Platform        string
	PrimaryModifier string
	Modifiers       map[string]string
}

// CreateTestKeymapHostOptions are options for creating a test host.
type CreateTestKeymapHostOptions struct {
	Metadata any
}

// TestDiagnosticCapture captures warnings and errors emitted by a keymap.
type TestDiagnosticCapture struct {
	warnings       []string
	errors         []string
	warningEvents  []WarningEvent
	errorEvents    []ErrorEvent
	mu             sync.Mutex
}

// WarningEvent is a warning event emitted by the keymap.
type WarningEvent struct {
	Message string
	Code    string
}

// ErrorEvent is an error event emitted by the keymap.
type ErrorEvent struct {
	Message string
	Code    string
	Err     string
}

// takeWarnings returns a snapshot of captured warnings and clears the buffer.
func (c *TestDiagnosticCapture) takeWarnings() ([]string, []WarningEvent) {
	c.mu.Lock()
	defer c.mu.Unlock()
	w := append([]string{}, c.warnings...)
	e := append([]WarningEvent{}, c.warningEvents...)
	c.warnings = nil
	c.warningEvents = nil
	return w, e
}

// takeErrors returns a snapshot of captured errors and clears the buffer.
func (c *TestDiagnosticCapture) takeErrors() ([]string, []ErrorEvent) {
	c.mu.Lock()
	defer c.mu.Unlock()
	er := append([]string{}, c.errors...)
	ee := append([]ErrorEvent{}, c.errorEvents...)
	c.errors = nil
	c.errorEvents = nil
	return er, ee
}

// clear clears all captured diagnostics.
func (c *TestDiagnosticCapture) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.warnings = nil
	c.errors = nil
	c.warningEvents = nil
	c.errorEvents = nil
}

// dispose removes the diagnostic listeners from the keymap.
func (c *TestDiagnosticCapture) dispose() {
	// Listeners are GC'd when the capture is dropped.
}

// TestKeymapTarget is a test target with a parent chain for focus activation.
type TestKeymapTarget struct {
	parent      *TestKeymapTarget
	isDestroyed bool
	id          string
}

// NewTestKeymapTarget creates a new test target.
func NewTestKeymapTarget(id string) *TestKeymapTarget {
	return &TestKeymapTarget{id: id}
}

// Append appends a child target and returns it.
func (t *TestKeymapTarget) Append(child *TestKeymapTarget) *TestKeymapTarget {
	child.parent = t
	return child
}

// ID returns the target id.
func (t *TestKeymapTarget) ID() string { return t.id }

// TestKeymapEvent is a test key event implementing KeymapEvent.
type TestKeymapEvent struct {
	name     string
	ctrl     bool
	shift    bool
	meta     bool
	super    bool
	hyper    bool
	prevented bool
	stopped   bool
}

// NewTestKeymapEvent creates a new test key event.
func NewTestKeymapEvent(name string, modifiers ...TestKeyModifierOptions) *TestKeymapEvent {
	e := &TestKeymapEvent{name: name}
	if len(modifiers) > 0 {
		e.ctrl = modifiers[0].Ctrl
		e.shift = modifiers[0].Shift
		e.meta = modifiers[0].Meta
		e.super = modifiers[0].Super
		e.hyper = modifiers[0].Hyper
	}
	return e
}

func (e *TestKeymapEvent) Name() string           { return e.name }
func (e *TestKeymapEvent) Ctrl() bool             { return e.ctrl }
func (e *TestKeymapEvent) Shift() bool            { return e.shift }
func (e *TestKeymapEvent) Meta() bool             { return e.meta }
func (e *TestKeymapEvent) Super() bool            { return e.super }
func (e *TestKeymapEvent) Hyper() bool            { return e.hyper }
func (e *TestKeymapEvent) PreventDefault()        { e.prevented = true }
func (e *TestKeymapEvent) StopPropagation()       { e.stopped = true }
func (e *TestKeymapEvent) PropagationStopped() bool { return e.stopped }

// TestKeymapHost embeds the concrete KeymapHost type so it can be passed
// directly to NewKeymap, while adding test-only state (focused target,
// listeners, etc.).
type TestKeymapHost struct {
	KeymapHost[*TestKeymapTarget, *TestKeymapEvent]

	focusedTarget *TestKeymapTarget
	isDestroyed   bool

	keyPressListeners  map[int64]func(*TestKeymapEvent)
	keyReleaseListeners map[int64]func(*TestKeymapEvent)
	focusListeners     map[int64]func(*TestKeymapTarget)
	destroyListeners   map[int64]func()
	rawListeners       map[int64]func(string) bool
	targetDestroy      map[*TestKeymapTarget]map[int64]func()
	nextID             int64
	mu                 sync.Mutex
}

// DEFAULT_TEST_HOST_METADATA is the default metadata for a test host.
var DEFAULT_TEST_HOST_METADATA = HostMetadata{
	Platform:        "unknown",
	PrimaryModifier: "unknown",
	Modifiers: map[string]string{
		"ctrl":  "unknown",
		"shift": "unknown",
		"meta":  "unknown",
		"super": "unknown",
		"hyper": "unknown",
	},
}

// NewTestKeymapHost creates a new test host with all KeymapHost callbacks wired.
func NewTestKeymapHost(options *CreateTestKeymapHostOptions) *TestKeymapHost {
	h := &TestKeymapHost{
		KeymapHost: KeymapHost[*TestKeymapTarget, *TestKeymapEvent]{
			RootTarget: NewTestKeymapTarget("root"),
		},
		keyPressListeners:  make(map[int64]func(*TestKeymapEvent)),
		keyReleaseListeners: make(map[int64]func(*TestKeymapEvent)),
		focusListeners:     make(map[int64]func(*TestKeymapTarget)),
		destroyListeners:   make(map[int64]func()),
		rawListeners:       make(map[int64]func(string) bool),
		targetDestroy:      make(map[*TestKeymapTarget]map[int64]func()),
	}
	if options != nil && options.Metadata != nil {
		switch m := options.Metadata.(type) {
		case HostMetadata:
			h.Metadata = m
		case TestHostMetadataOptions:
			h.Metadata = normalizeTestHostMetadataOptions(m)
		}
	}
	if h.Metadata.Modifiers == nil && h.Metadata.Platform == "" && h.Metadata.PrimaryModifier == "" {
		h.Metadata = DEFAULT_TEST_HOST_METADATA
	}

	h.IsDestroyed = func() bool { return h.isDestroyed }
	h.GetFocusedTarget = func() *TestKeymapTarget {
		if h.focusedTarget != nil && h.focusedTarget.isDestroyed {
			return nil
		}
		return h.focusedTarget
	}
	h.GetParentTarget = func(t *TestKeymapTarget) *TestKeymapTarget {
		if t == nil {
			return nil
		}
		return t.parent
	}
	h.IsTargetDestroyed = func(t *TestKeymapTarget) bool {
		if t == nil {
			return false
		}
		return t.isDestroyed
	}

	h.OnKeyPress = func(l func(*TestKeymapEvent)) func() {
		h.mu.Lock()
		h.nextID++
		id := h.nextID
		h.keyPressListeners[id] = l
		h.mu.Unlock()
		return func() {
			h.mu.Lock()
			delete(h.keyPressListeners, id)
			h.mu.Unlock()
		}
	}
	h.OnKeyRelease = func(l func(*TestKeymapEvent)) func() {
		h.mu.Lock()
		h.nextID++
		id := h.nextID
		h.keyReleaseListeners[id] = l
		h.mu.Unlock()
		return func() {
			h.mu.Lock()
			delete(h.keyReleaseListeners, id)
			h.mu.Unlock()
		}
	}
	h.OnFocusChange = func(l func(*TestKeymapTarget)) func() {
		h.mu.Lock()
		h.nextID++
		id := h.nextID
		h.focusListeners[id] = l
		h.mu.Unlock()
		return func() {
			h.mu.Lock()
			delete(h.focusListeners, id)
			h.mu.Unlock()
		}
	}
	h.OnDestroy = func(l func()) func() {
		h.mu.Lock()
		h.nextID++
		id := h.nextID
		h.destroyListeners[id] = l
		h.mu.Unlock()
		return func() {
			h.mu.Lock()
			delete(h.destroyListeners, id)
			h.mu.Unlock()
		}
	}
	h.OnRawInput = func(l func(string) bool) func() {
		h.mu.Lock()
		h.nextID++
		id := h.nextID
		h.rawListeners[id] = l
		h.mu.Unlock()
		return func() {
			h.mu.Lock()
			delete(h.rawListeners, id)
			h.mu.Unlock()
		}
	}
	h.CreateCommandEvent = func() *TestKeymapEvent { return NewTestKeymapEvent("synthetic") }
	h.OnTargetDestroy = func(target *TestKeymapTarget, l func()) func() {
		h.mu.Lock()
		if h.targetDestroy[target] == nil {
			h.targetDestroy[target] = make(map[int64]func())
		}
		h.nextID++
		id := h.nextID
		h.targetDestroy[target][id] = l
		h.mu.Unlock()
		return func() {
			h.mu.Lock()
			delete(h.targetDestroy[target], id)
			if len(h.targetDestroy[target]) == 0 {
				delete(h.targetDestroy, target)
			}
			h.mu.Unlock()
		}
	}
	return h
}

func normalizeTestHostMetadataOptions(o TestHostMetadataOptions) HostMetadata {
	m := HostMetadata{
		Platform:        o.Platform,
		PrimaryModifier: o.PrimaryModifier,
		Modifiers:       map[string]string{},
	}
	for k, v := range DEFAULT_TEST_HOST_METADATA.Modifiers {
		m.Modifiers[k] = v
	}
	for k, v := range o.Modifiers {
		m.Modifiers[k] = v
	}
	return m
}

func (h *TestKeymapHost) Focus(target *TestKeymapTarget) {
	h.focusedTarget = target
	h.mu.Lock()
	listeners := make([]func(*TestKeymapTarget), 0, len(h.focusListeners))
	for _, l := range h.focusListeners {
		listeners = append(listeners, l)
	}
	h.mu.Unlock()
	for _, l := range listeners {
		l(target)
	}
}

func (h *TestKeymapHost) Press(name string, modifiers ...TestKeyModifierOptions) *TestKeymapEvent {
	e := NewTestKeymapEvent(name, modifiers...)
	h.mu.Lock()
	listeners := make([]func(*TestKeymapEvent), 0, len(h.keyPressListeners))
	for _, l := range h.keyPressListeners {
		listeners = append(listeners, l)
	}
	h.mu.Unlock()
	for _, l := range listeners {
		l(e)
	}
	return e
}

func (h *TestKeymapHost) Release(name string, modifiers ...TestKeyModifierOptions) *TestKeymapEvent {
	e := NewTestKeymapEvent(name, modifiers...)
	h.mu.Lock()
	listeners := make([]func(*TestKeymapEvent), 0, len(h.keyReleaseListeners))
	for _, l := range h.keyReleaseListeners {
		listeners = append(listeners, l)
	}
	h.mu.Unlock()
	for _, l := range listeners {
		l(e)
	}
	return e
}

func (h *TestKeymapHost) Raw(sequence string) bool {
	handled := false
	h.mu.Lock()
	listeners := make([]func(string) bool, 0, len(h.rawListeners))
	for _, l := range h.rawListeners {
		listeners = append(listeners, l)
	}
	h.mu.Unlock()
	for _, l := range listeners {
		if l(sequence) {
			handled = true
		}
	}
	return handled
}

func (h *TestKeymapHost) DestroyTarget(target *TestKeymapTarget) {
	target.isDestroyed = true
	h.mu.Lock()
	listeners := h.targetDestroy[target]
	delete(h.targetDestroy, target)
	h.mu.Unlock()
	for _, l := range listeners {
		l()
	}
}

func (h *TestKeymapHost) Destroy() {
	if h.isDestroyed {
		return
	}
	h.isDestroyed = true
	h.mu.Lock()
	listeners := make([]func(), 0, len(h.destroyListeners))
	for _, l := range h.destroyListeners {
		listeners = append(listeners, l)
	}
	h.mu.Unlock()
	for _, l := range listeners {
		l()
	}
}

// TestKeymapHarness bundles a keymap, its host, root target, and diagnostics.
type TestKeymapHarness struct {
	Keymap     *Keymap[*TestKeymapTarget, *TestKeymapEvent]
	Host       *TestKeymapHost
	Root       *TestKeymapTarget
	Diagnostics *TestDiagnosticCapture
}

// Cleanup destroys the host and disposes diagnostics.
func (h *TestKeymapHarness) Cleanup() {
	h.Host.Destroy()
	h.Diagnostics.dispose()
}

// CreateTestKeymapHost creates a test keymap host.
func CreateTestKeymapHost(options *CreateTestKeymapHostOptions) *TestKeymapHost {
	return NewTestKeymapHost(options)
}

// CreateTestKeymapOptions are options for CreateTestKeymap.
type CreateTestKeymapOptions struct {
	Metadata   any
	DefaultKeys bool
}

// CreateTestKeymap creates a keymap bound to a test host, optionally with
// default keys registered.
func CreateTestKeymap(options *CreateTestKeymapOptions) *TestKeymapHarness {
	var metadata any
	if options != nil {
		metadata = options.Metadata
	}
	host := NewTestKeymapHost(&CreateTestKeymapHostOptions{Metadata: metadata})
	km, err := NewKeymap(host.KeymapHost)
	if err != nil {
		panic(err)
	}
	diagnostics := CaptureKeymapDiagnostics(km)

	harness := &TestKeymapHarness{
		Keymap:      km,
		Host:        host,
		Root:        host.RootTarget,
		Diagnostics: diagnostics,
	}

	if options != nil && options.DefaultKeys {
		RegisterDefaultKeys(km, host.RootTarget)
		host.Focus(host.RootTarget)
	}

	return harness
}

// CaptureKeymapDiagnostics captures warnings and errors emitted by a keymap.
func CaptureKeymapDiagnostics(km *Keymap[*TestKeymapTarget, *TestKeymapEvent]) *TestDiagnosticCapture {
	c := &TestDiagnosticCapture{}
	km.On("warning", func(v any) {
		if ev, ok := v.(map[string]string); ok {
			c.mu.Lock()
			c.warnings = append(c.warnings, ev["message"])
			c.warningEvents = append(c.warningEvents, WarningEvent{Message: ev["message"], Code: ev["code"]})
			c.mu.Unlock()
		}
	})
	km.On("error", func(v any) {
		if ev, ok := v.(map[string]string); ok {
			c.mu.Lock()
			c.errors = append(c.errors, ev["message"])
			c.errorEvents = append(c.errorEvents, ErrorEvent{Message: ev["message"], Code: ev["code"], Err: ev["err"]})
			c.mu.Unlock()
		}
	})
	return c
}

// ---------------------------------------------------------------------------
// Default keys addon
// ---------------------------------------------------------------------------

// RegisterDefaultKeys registers a conservative set of default bindings
// (emacs-style movement + common editing) for testing purposes.
func RegisterDefaultKeys(km *Keymap[*TestKeymapTarget, *TestKeymapEvent], root *TestKeymapTarget) {
	// Emacs-style movement bindings.
	emacsBindings := []Binding[*TestKeymapTarget, *TestKeymapEvent]{
		{Key: "ctrl+a", Cmd: BindingCommand[*TestKeymapTarget, *TestKeymapEvent]{Kind: BindingCommandName, Name: "emacs-beginning-of-line"}},
		{Key: "ctrl+e", Cmd: BindingCommand[*TestKeymapTarget, *TestKeymapEvent]{Kind: BindingCommandName, Name: "emacs-end-of-line"}},
		{Key: "ctrl+f", Cmd: BindingCommand[*TestKeymapTarget, *TestKeymapEvent]{Kind: BindingCommandName, Name: "emacs-forward-char"}},
		{Key: "ctrl+b", Cmd: BindingCommand[*TestKeymapTarget, *TestKeymapEvent]{Kind: BindingCommandName, Name: "emacs-backward-char"}},
		{Key: "ctrl+p", Cmd: BindingCommand[*TestKeymapTarget, *TestKeymapEvent]{Kind: BindingCommandName, Name: "emacs-previous-line"}},
		{Key: "ctrl+n", Cmd: BindingCommand[*TestKeymapTarget, *TestKeymapEvent]{Kind: BindingCommandName, Name: "emacs-next-line"}},
	}
	// Neovim-style motion bindings.
	neovimBindings := []Binding[*TestKeymapTarget, *TestKeymapEvent]{
		{Key: "h", Cmd: BindingCommand[*TestKeymapTarget, *TestKeymapEvent]{Kind: BindingCommandName, Name: "vim-left"}},
		{Key: "j", Cmd: BindingCommand[*TestKeymapTarget, *TestKeymapEvent]{Kind: BindingCommandName, Name: "vim-down"}},
		{Key: "k", Cmd: BindingCommand[*TestKeymapTarget, *TestKeymapEvent]{Kind: BindingCommandName, Name: "vim-up"}},
		{Key: "l", Cmd: BindingCommand[*TestKeymapTarget, *TestKeymapEvent]{Kind: BindingCommandName, Name: "vim-right"}},
		{Key: "gg", Cmd: BindingCommand[*TestKeymapTarget, *TestKeymapEvent]{Kind: BindingCommandName, Name: "vim-goto-top"}},
		{Key: "G", Cmd: BindingCommand[*TestKeymapTarget, *TestKeymapEvent]{Kind: BindingCommandName, Name: "vim-goto-bottom"}},
	}
	// Dead bindings: consume keys without running a command.
	deadBindings := []Binding[*TestKeymapTarget, *TestKeymapEvent]{
		{Key: "ctrl+q", Cmd: BindingCommand[*TestKeymapTarget, *TestKeymapEvent]{Kind: BindingCommandNone}, PreventDefault: true},
	}
	// Leader prefix binding: space followed by another key.
	leaderBindings := []Binding[*TestKeymapTarget, *TestKeymapEvent]{
		{Key: "space g", Cmd: BindingCommand[*TestKeymapTarget, *TestKeymapEvent]{Kind: BindingCommandName, Name: "leader-goto"}},
	}
	// Chord bindings (multi-stroke).
	chordBindings := []Binding[*TestKeymapTarget, *TestKeymapEvent]{
		{Key: "ctrl+x ctrl+c", Cmd: BindingCommand[*TestKeymapTarget, *TestKeymapEvent]{Kind: BindingCommandName, Name: "emacs-copy"}},
	}

	allBindings := append(append(append(append(emacsBindings, neovimBindings...), deadBindings...), leaderBindings...), chordBindings...)
	km.RegisterLayer(Layer[*TestKeymapTarget, *TestKeymapEvent]{
		Target:      root,
		TargetMode:  TargetModeFocusWithin,
		Priority:    100,
		Bindings:    allBindings,
		Commands: []Command{
			{Name: "emacs-beginning-of-line", Run: func(ctx CommandContext[any, KeymapEvent]) CommandResult { return RunCommandResult{OK: true} }},
			{Name: "emacs-end-of-line", Run: func(ctx CommandContext[any, KeymapEvent]) CommandResult { return RunCommandResult{OK: true} }},
			{Name: "emacs-forward-char", Run: func(ctx CommandContext[any, KeymapEvent]) CommandResult { return RunCommandResult{OK: true} }},
			{Name: "emacs-backward-char", Run: func(ctx CommandContext[any, KeymapEvent]) CommandResult { return RunCommandResult{OK: true} }},
			{Name: "emacs-previous-line", Run: func(ctx CommandContext[any, KeymapEvent]) CommandResult { return RunCommandResult{OK: true} }},
			{Name: "emacs-next-line", Run: func(ctx CommandContext[any, KeymapEvent]) CommandResult { return RunCommandResult{OK: true} }},
			{Name: "emacs-copy", Run: func(ctx CommandContext[any, KeymapEvent]) CommandResult { return RunCommandResult{OK: true} }},
			{Name: "vim-left", Run: func(ctx CommandContext[any, KeymapEvent]) CommandResult { return RunCommandResult{OK: true} }},
			{Name: "vim-down", Run: func(ctx CommandContext[any, KeymapEvent]) CommandResult { return RunCommandResult{OK: true} }},
			{Name: "vim-up", Run: func(ctx CommandContext[any, KeymapEvent]) CommandResult { return RunCommandResult{OK: true} }},
			{Name: "vim-right", Run: func(ctx CommandContext[any, KeymapEvent]) CommandResult { return RunCommandResult{OK: true} }},
			{Name: "vim-goto-top", Run: func(ctx CommandContext[any, KeymapEvent]) CommandResult { return RunCommandResult{OK: true} }},
			{Name: "vim-goto-bottom", Run: func(ctx CommandContext[any, KeymapEvent]) CommandResult { return RunCommandResult{OK: true} }},
			{Name: "leader-goto", Run: func(ctx CommandContext[any, KeymapEvent]) CommandResult { return RunCommandResult{OK: true} }},
		},
		Fields:      map[string]any{"addon": "default-keys"},
	})
}

// ---------------------------------------------------------------------------
// React / Solid integration stubs
// ---------------------------------------------------------------------------

// RegisterKeymapReact is a placeholder for React framework integration.
// In the Go port this is a no-op; React bindings are provided by the
// @pixtui/keymap/react TypeScript entrypoint in the original repository.
func RegisterKeymapReact() {}

// RegisterKeymapSolid is a placeholder for Solid framework integration.
// In the Go port this is a no-op; Solid bindings are provided by the
// @pixtui/keymap/solid TypeScript entrypoint in the original repository.
func RegisterKeymapSolid() {}
