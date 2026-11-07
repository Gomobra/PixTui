package keymap

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// KeymapEvent is the interface a key event must satisfy so the keymap can read
// its modifiers and consume it.
type KeymapEvent interface {
	Name() string
	Ctrl() bool
	Shift() bool
	Meta() bool
	Super() bool
	Hyper() bool
	PreventDefault()
	StopPropagation()
	PropagationStopped() bool
}

// HostMetadata describes the host platform and modifier capabilities.
type HostMetadata struct {
	Platform        string
	PrimaryModifier string // "ctrl" | "super" | "unknown"
	Modifiers       map[string]string
}

// KeymapHost is the embedding host the keymap plugs into.
type KeymapHost[TTarget any, TEvent KeymapEvent] struct {
	Metadata           HostMetadata
	RootTarget         TTarget
	IsDestroyed        func() bool
	GetFocusedTarget   func() TTarget
	GetParentTarget    func(TTarget) TTarget
	IsTargetDestroyed  func(TTarget) bool
	OnKeyPress         func(func(TEvent)) func()
	OnKeyRelease       func(func(TEvent)) func()
	OnFocusChange      func(func(TTarget)) func()
	OnDestroy          func(func()) func()            // optional
	OnRawInput         func(func(string) bool) func() // optional
	CreateCommandEvent func() TEvent
	OnTargetDestroy    func(TTarget, func()) func()
}

// ---------------------------------------------------------------------------
// Key strokes, matches and sequences
// ---------------------------------------------------------------------------

// KeyStroke is a normalized key stroke.
type KeyStroke struct {
	Name  string
	Ctrl  bool
	Shift bool
	Meta  bool
	Super bool
	Hyper bool
}

// MatchID returns a stable match key for the stroke.
func (s KeyStroke) MatchID() string {
	return fmt.Sprintf("key:%s:%d:%d:%d:%d:%d",
		s.Name, boolInt(s.Ctrl), boolInt(s.Shift), boolInt(s.Meta), boolInt(s.Super), boolInt(s.Hyper))
}

// String renders the stroke as a human readable key sequence, e.g. "ctrl+x".
func (s KeyStroke) String() string {
	parts := make([]string, 0, 6)
	if s.Ctrl {
		parts = append(parts, "ctrl")
	}
	if s.Shift {
		parts = append(parts, "shift")
	}
	if s.Meta {
		parts = append(parts, "meta")
	}
	if s.Super {
		parts = append(parts, "super")
	}
	if s.Hyper {
		parts = append(parts, "hyper")
	}
	name := s.Name
	if name == "return" {
		name = "enter"
	}
	parts = append(parts, name)
	return strings.Join(parts, "+")
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// KeyStrokeInput is the input form of a single stroke.
type KeyStrokeInput struct {
	Name  string
	Ctrl  bool
	Shift bool
	Meta  bool
	Super bool
	Hyper bool
}

// normalizeKeyStroke normalizes a key stroke input.
func normalizeKeyStroke(input KeyStrokeInput) KeyStroke {
	return KeyStroke{
		Name:  strings.ToLower(strings.TrimSpace(input.Name)),
		Ctrl:  input.Ctrl,
		Shift: input.Shift,
		Meta:  input.Meta,
		Super: input.Super,
		Hyper: input.Hyper,
	}
}

// normalizeEventKeyStroke converts a KeymapEvent into a normalized stroke.
func normalizeEventKeyStroke[TEvent KeymapEvent](event TEvent) KeyStroke {
	return KeyStroke{
		Name:  strings.ToLower(strings.TrimSpace(event.Name())),
		Ctrl:  event.Ctrl(),
		Shift: event.Shift(),
		Meta:  event.Meta(),
		Super: event.Super(),
		Hyper: event.Hyper(),
	}
}

// KeySequencePart is one part of a key sequence.
type KeySequencePart struct {
	Stroke      KeyStroke
	Display     string
	Match       string
	TokenName   string
	PatternName string
	PayloadKey  string
}

// KeyMatch is a string key used to match strokes.
type KeyMatch = string

// createKeySequencePart builds a KeySequencePart from a stroke input.
func createKeySequencePart(input KeyStrokeInput, opts ...func(*KeySequencePart)) KeySequencePart {
	stroke := normalizeKeyStroke(input)
	part := KeySequencePart{
		Stroke:  stroke,
		Display: stroke.String(),
		Match:   stroke.MatchID(),
	}
	for _, opt := range opts {
		opt(&part)
	}
	return part
}

func cloneKeySequencePart(part KeySequencePart) KeySequencePart {
	p := part
	p.Stroke = part.Stroke
	return p
}

func cloneKeySequence(parts []KeySequencePart) []KeySequencePart {
	out := make([]KeySequencePart, len(parts))
	for i, p := range parts {
		out[i] = cloneKeySequencePart(p)
	}
	return out
}

func stringifyKeyStroke(input interface{}, opts ...map[string]string) string {
	preferDisplay := false
	if len(opts) > 0 {
		if v, ok := opts[0]["preferDisplay"]; ok && v == "true" {
			preferDisplay = true
		}
	}
	switch v := input.(type) {
	case KeySequencePart:
		if preferDisplay && v.Display != "" {
			return v.Display
		}
		return v.Stroke.String()
	case KeyStroke:
		return v.String()
	case KeyStrokeInput:
		return normalizeKeyStroke(v).String()
	default:
		return fmt.Sprintf("%v", input)
	}
}

func stringifyKeySequence(parts []KeySequencePart, opts ...map[string]string) string {
	separator := ""
	if len(opts) > 0 {
		if v, ok := opts[0]["separator"]; ok {
			separator = v
		}
	}
	out := make([]string, len(parts))
	for i, p := range parts {
		out[i] = stringifyKeyStroke(p, opts...)
	}
	return strings.Join(out, separator)
}

// ---------------------------------------------------------------------------
// Commands
// ---------------------------------------------------------------------------

// CommandResult is the result of running a command: a bool, nil or *RunCommandResult.
type CommandResult any

// RunCommandResult is a structured command result.
type RunCommandResult struct {
	OK      bool
	Reason  string // "not-found"|"inactive"|"disabled"|"invalid-args"|"rejected"|"error"
	Command *Command
}

// CommandHandler runs a command.
type CommandHandler[TTarget any, TEvent KeymapEvent] func(ctx CommandContext[TTarget, TEvent]) CommandResult

// CommandContext is passed to command handlers.
type CommandContext[TTarget any, TEvent KeymapEvent] struct {
	Keymap  any
	Event   TEvent
	Focused TTarget
	Target  TTarget
	Data    map[string]any
	Command *Command
	Input   string
	Payload any
}

// Command is a named command with a run handler.
type Command struct {
	Name   string
	Run    CommandHandler[any, KeymapEvent]
	Fields map[string]any
}

// BindingCommand is either a command name or a direct handler.
type BindingCommand[TTarget any, TEvent KeymapEvent] struct {
	Kind int // BindingCommandName or BindingCommandHandler
	Name string
	Run  CommandHandler[any, KeymapEvent]
}

const (
	BindingCommandNone    = 0
	BindingCommandName    = 1
	BindingCommandHandler = 2
)

// BindingEvent is the event phase a binding fires on.
type BindingEvent = string

const (
	BindingEventPress   BindingEvent = "press"
	BindingEventRelease BindingEvent = "release"
)

// Binding is a source binding declared on a layer.
type Binding[TTarget any, TEvent KeymapEvent] struct {
	Key   string // key sequence string
	Cmd   BindingCommand[TTarget, TEvent]
	Event BindingEvent
	// PreventDefault consumes the event when the binding runs.
	PreventDefault bool
	// Fallthrough continues to later matching bindings after running.
	Fallthrough bool
	// Fields are custom binding fields compiled by binding field compilers.
	Fields map[string]any
}

// ---------------------------------------------------------------------------
// Layers
// ---------------------------------------------------------------------------

// TargetMode selects how a layer target is matched against focus.
type TargetMode = string

const (
	TargetModeFocus       TargetMode = "focus"
	TargetModeFocusWithin TargetMode = "focus-within"
)

// Layer is a group of bindings and commands.
type Layer[TTarget any, TEvent KeymapEvent] struct {
	Target     TTarget
	Priority   int
	Bindings   []Binding[TTarget, TEvent]
	Commands   []Command
	TargetMode TargetMode
	Fields     map[string]any
}

// ---------------------------------------------------------------------------
// Active bindings and keys (query results)
// ---------------------------------------------------------------------------

// ActiveBinding is a resolved, active binding.
type ActiveBinding[TTarget any, TEvent KeymapEvent] struct {
	Sequence       []KeySequencePart
	Command        BindingCommand[TTarget, TEvent]
	CommandAttrs   map[string]any
	Attrs          map[string]any
	Event          BindingEvent
	PreventDefault bool
	Fallthrough    bool
}

// ActiveKey is an active key presentation.
type ActiveKey[TTarget any, TEvent KeymapEvent] struct {
	Stroke       KeyStroke
	Display      string
	TokenName    string
	Bindings     []ActiveBinding[TTarget, TEvent]
	BindingAttrs map[string]any
	CommandAttrs map[string]any
	Command      BindingCommand[TTarget, TEvent]
	Continues    bool
}

// ActiveKeyOptions control what active keys include.
type ActiveKeyOptions struct {
	IncludeBindings bool
	IncludeMetadata bool
}

// ---------------------------------------------------------------------------
// Command queries
// ---------------------------------------------------------------------------

// CommandQuery filters commands.
type CommandQuery[TTarget any, TEvent KeymapEvent] struct {
	Visibility string // "reachable"|"active"|"registered"
	Focused    TTarget
	Namespace  string
	Search     string
	SearchIn   []string
	Filter     func(*Command) bool
	Limit      int
}

// CommandBindingsQuery queries bindings for commands.
type CommandBindingsQuery[TTarget any] struct {
	Visibility string
	Focused    TTarget
	Commands   []string
}

// CommandEntry is a command together with its bindings in a query projection.
type CommandEntry[TTarget any, TEvent KeymapEvent] struct {
	Command  *Command
	Bindings []ActiveBinding[TTarget, TEvent]
}

// ---------------------------------------------------------------------------
// Hooks / intercepts
// ---------------------------------------------------------------------------

// HookName is a hook/intercept name.
type HookName = string

const (
	HookState         HookName = "state"
	HookPendingSeq    HookName = "pendingSequence"
	HookDispatch      HookName = "dispatch"
	InterceptKey      HookName = "key"
	InterceptKeyAfter HookName = "key:after"
	InterceptRaw      HookName = "raw"
)

// ---------------------------------------------------------------------------
// Diagnostics
// ---------------------------------------------------------------------------

// ActiveBindingInfo is a diagnostics entry for an active binding.
type ActiveBindingInfo[TTarget any, TEvent KeymapEvent] struct {
	Sequence   string
	Command    string
	LayerOrder int
	Priority   int
	Event      BindingEvent
}

// FeaturePack is a named group of bindings that can be enabled/disabled together.
type FeaturePack[TTarget any, TEvent KeymapEvent] struct {
	Name     string
	Bindings []Binding[TTarget, TEvent]
	Enabled  bool
	layer    *registeredLayer[TTarget, TEvent]
	keymap   *Keymap[TTarget, TEvent]
}

// Enable enables the feature pack.
func (p *FeaturePack[TTarget, TEvent]) Enable() {
	p.keymap.mu.Lock()
	defer p.keymap.mu.Unlock()
	if p.Enabled || p.layer != nil {
		return
	}
	p.Enabled = true
	p.layer = p.keymap.registerLayerLocked(Layer[TTarget, TEvent]{
		Priority: 100,
		Bindings: p.Bindings,
		Fields:   map[string]any{"featurePack": p.Name},
	})
}

// Disable disables the feature pack.
func (p *FeaturePack[TTarget, TEvent]) Disable() {
	p.keymap.mu.Lock()
	defer p.keymap.mu.Unlock()
	if !p.Enabled || p.layer == nil {
		return
	}
	p.Enabled = false
	p.keymap.unregisterLayerLocked(p.layer)
	p.layer = nil
}

// ---------------------------------------------------------------------------
// Internal registered layer
// ---------------------------------------------------------------------------

type registeredLayer[TTarget any, TEvent KeymapEvent] struct {
	order      int
	target     TTarget
	targetMode TargetMode
	priority   int
	requires   []struct {
		Name  string
		Value any
	}
	matchers           []runtimeMatcher
	fields             map[string]any
	attrs              map[string]any
	commands           []commandState[TTarget, TEvent]
	sourceBindings     []Binding[TTarget, TEvent]
	bindings           []bindingState[TTarget, TEvent]
	root               *sequenceNode[TTarget, TEvent]
	hasTokenBindings   bool
	activeKeyBlocked   bool
	activeCmdViewBlock bool
	offTargetDestroy   func()
}

type commandState[TTarget any, TEvent KeymapEvent] struct {
	command *Command
	fields  map[string]any
	attrs   map[string]any
}

type runtimeMatcher struct {
	source    string
	match     func() bool
	subscribe func(func()) func()
	dispose   func()
}

type bindingState[TTarget any, TEvent KeymapEvent] struct {
	binding          *Binding[TTarget, TEvent]
	sequence         []KeySequencePart
	command          BindingCommand[TTarget, TEvent]
	event            BindingEvent
	sourceTarget     TTarget
	sourceLayerOrder int
	bindingIndex     int
	requires         []struct {
		Name  string
		Value any
	}
	matchers       []runtimeMatcher
	attrs          map[string]any
	run            CommandHandler[any, KeymapEvent]
	preventDefault bool
	Fallthrough    bool
}

type sequenceNode[TTarget any, TEvent KeymapEvent] struct {
	parent            *sequenceNode[TTarget, TEvent]
	depth             int
	stroke            *KeyStroke
	match             string
	patternName       string
	children          map[string]*sequenceNode[TTarget, TEvent]
	patternChildren   []*sequenceNode[TTarget, TEvent]
	bindings          []bindingState[TTarget, TEvent]
	reachableBindings []bindingState[TTarget, TEvent]
}

// ---------------------------------------------------------------------------
// Pending sequence
// ---------------------------------------------------------------------------

type pendingSequenceCapture[TTarget any, TEvent KeymapEvent] struct {
	layer    *registeredLayer[TTarget, TEvent]
	binding  *bindingState[TTarget, TEvent]
	index    int
	parts    []KeySequencePart
	patterns []pendingSequencePatternCapture
}

type pendingSequencePatternCapture struct {
	name       string
	payloadKey string
	values     []any
	parts      []KeySequencePart
}

type pendingSequenceState[TTarget any, TEvent KeymapEvent] struct {
	captures []pendingSequenceCapture[TTarget, TEvent]
}

// ---------------------------------------------------------------------------
// State
// ---------------------------------------------------------------------------

type State[TTarget any, TEvent KeymapEvent] struct {
	mu                             sync.RWMutex
	order                          int
	tokens                         map[string]resolvedKeyToken
	patterns                       map[string]resolvedSequencePattern[TEvent]
	layerFields                    map[string]layerFieldCompiler
	layerBindingsTransformers      []layerBindingsTransformer[TTarget, TEvent]
	bindingExpanders               []bindingExpander[TTarget, TEvent]
	bindingParsers                 []bindingParser[TEvent]
	bindingTransformers            []bindingTransformer[TTarget, TEvent]
	bindingFields                  map[string]bindingFieldCompiler
	commandTransformers            []commandTransformer[TTarget, TEvent]
	commandFields                  map[string]commandFieldCompiler
	eventMatchResolvers            []eventMatchResolver[TEvent]
	disambiguationResolvers        []disambiguationResolver[TTarget, TEvent]
	keyHooks                       []keyHookRegistration[TTarget, TEvent]
	keyAfterHooks                  []keyAfterHookRegistration[TTarget, TEvent]
	rawHooks                       []rawHookRegistration
	hookID                         int64
	layers                         map[*registeredLayer[TTarget, TEvent]]bool
	sortedLayers                   []*registeredLayer[TTarget, TEvent]
	activeLayersCacheVersion       int
	activeLayersCacheFocused       TTarget
	activeLayersCache              []*registeredLayer[TTarget, TEvent]
	activeKeyCacheBlockers         int
	activeCommandViewCacheBlockers int
	commandResolvers               []commandResolver[TTarget, TEvent]
	pending                        *pendingSequenceState[TTarget, TEvent]
	data                           map[string]any
	dataVersion                    int
	readonlyDataVersion            int
	readonlyData                   map[string]any
	cacheVersion                   int
	derivedVersion                 int
	stateChangeDepth               int
	stateChangePending             bool
	flushingStateChange            bool
	usedWarningKeys                map[string]bool
}

type resolvedKeyToken struct {
	stroke KeyStroke
	match  string
}

type resolvedSequencePattern[TEvent KeymapEvent] struct {
	name       string
	display    string
	payloadKey string
	match      string
	min        int
	max        int
	matcher    func(TEvent) *patternMatch
	finalize   func([]any) any
}

type patternMatch struct {
	value   any
	display string
}

type layerFieldCompiler func(any, layerFieldCompilerContext)
type layerFieldCompilerContext struct {
	fieldName  string
	require    func(string, any)
	attr       func(string, any)
	activeWhen func(func() bool)
}

type bindingFieldCompiler func(any, bindingFieldCompilerContext)
type bindingFieldCompilerContext struct {
	fieldName  string
	require    func(string, any)
	attr       func(string, any)
	activeWhen func(func() bool)
}

type commandFieldCompiler func(any, commandFieldCompilerContext)
type commandFieldCompilerContext struct {
	fieldName  string
	require    func(string, any)
	attr       func(string, any)
	activeWhen func(func() bool)
}

type layerBindingsTransformer[TTarget any, TEvent KeymapEvent] func([]Binding[TTarget, TEvent], layerBindingsTransformerContext[TTarget, TEvent]) ([]Binding[TTarget, TEvent], bool)
type layerBindingsTransformerContext[TTarget any, TEvent KeymapEvent] struct {
	Layer            Layer[TTarget, TEvent]
	ValidateBindings func(interface{}) map[string]string
}

type bindingExpander[TTarget any, TEvent KeymapEvent] func(bindingExpanderContext) ([]bindingExpansion, bool)
type bindingExpanderContext struct {
	Input    string
	Displays []string
	Layer    map[string]any
}
type bindingExpansion struct {
	Key      string
	Displays []string
}

type bindingParser[TEvent KeymapEvent] func(bindingParserContext[TEvent]) *bindingParserResult
type bindingParserContext[TEvent KeymapEvent] struct {
	Input              string
	Index              int
	Layer              map[string]any
	Tokens             map[string]resolvedKeyToken
	Patterns           map[string]resolvedSequencePattern[TEvent]
	NormalizeTokenName func(string) string
	CreateMatch        func(string) string
	ParseObjectKey     func(KeyStrokeInput, ...func(*KeySequencePart)) KeySequencePart
}
type bindingParserResult struct {
	Parts         []KeySequencePart
	NextIndex     int
	UsedTokens    []string
	UnknownTokens []string
}

type bindingTransformer[TTarget any, TEvent KeymapEvent] func(*Binding[TTarget, TEvent], bindingTransformerContext[TTarget, TEvent])
type bindingTransformerContext[TTarget any, TEvent KeymapEvent] struct {
	Layer        map[string]any
	ParseKey     func(interface{}) KeySequencePart
	Add          func(*Binding[TTarget, TEvent])
	SkipOriginal func()
}

type commandTransformer[TTarget any, TEvent KeymapEvent] func(*Command, commandTransformerContext[TTarget, TEvent])
type commandTransformerContext[TTarget any, TEvent KeymapEvent] struct {
	Layer        Layer[TTarget, TEvent]
	Add          func(*Command)
	SkipOriginal func()
}

type eventMatchResolver[TEvent KeymapEvent] func(TEvent, eventMatchResolverContext) []string
type eventMatchResolverContext struct {
	ResolveKey func(interface{}) string
}

type disambiguationResolver[TTarget any, TEvent KeymapEvent] func(disambiguationContext[TTarget, TEvent]) *disambiguationDecision[TTarget, TEvent]

type disambiguationDecision[TTarget any, TEvent KeymapEvent] struct {
	Action  string // "run-exact"|"continue-sequence"|"clear"|"defer"
	Handler func(deferredDisambiguationDecision) *deferredDisambiguationDecision
}

type deferredDisambiguationDecision struct {
	Action string // "run-exact"|"continue-sequence"|"clear"
}

type disambiguationContext[TTarget any, TEvent KeymapEvent] struct {
	Event         TEvent
	Focused       TTarget
	Sequence      []KeySequencePart
	Stroke        KeySequencePart
	Exact         []ActiveBinding[TTarget, TEvent]
	Continuations []ActiveKey[TTarget, TEvent]
	getData       func(string) any
	setData       func(string, any)
	runExact      func() *disambiguationDecision[TTarget, TEvent]
	continueSeq   func() *disambiguationDecision[TTarget, TEvent]
	clear         func() *disambiguationDecision[TTarget, TEvent]
	deferFn       func(func(deferredDisambiguationDecision) *deferredDisambiguationDecision) *disambiguationDecision[TTarget, TEvent]
}

type deferredDisambiguationContext[TTarget any, TEvent KeymapEvent] struct {
	Sequence    []KeySequencePart
	Focused     TTarget
	runExact    func() *deferredDisambiguationDecision
	continueSeq func() *deferredDisambiguationDecision
	clear       func() *deferredDisambiguationDecision
}

type commandResolver[TTarget any, TEvent KeymapEvent] func(string, commandResolverContext[TTarget, TEvent]) *Command
type commandResolverContext[TTarget any, TEvent KeymapEvent] struct {
	Input      string
	Payload    any
	SetInput   func(string)
	SetPayload func(any)
	GetCommand func(string) *Command
}

type keyHookRegistration[TTarget any, TEvent KeymapEvent] struct {
	id       int64
	Listener func(keyInputContext[TTarget, TEvent])
	Config   struct {
		Priority int
		Release  bool
	}
}

type keyAfterHookRegistration[TTarget any, TEvent KeymapEvent] struct {
	id       int64
	Listener func(keyAfterInputContext[TTarget, TEvent])
	Config   struct {
		Priority int
		Release  bool
	}
}

type rawHookRegistration struct {
	id       int64
	Listener func(rawInputContext)
	Config   struct{ Priority int }
}

type keyInputContext[TTarget any, TEvent KeymapEvent] struct {
	Event   TEvent
	SetData func(string, any)
	GetData func(string) any
	Consume func(...func(consumeOptions))
}

type consumeOptions struct {
	PreventDefault  bool
	StopPropagation bool
}

type keyAfterInputContext[TTarget any, TEvent KeymapEvent] struct {
	Event     TEvent
	EventType BindingEvent
	Focused   TTarget
	Handled   bool
	Reason    string
	Sequence  []KeySequencePart
	Pending   []KeySequencePart
	SetData   func(string, any)
	GetData   func(string) any
	Consume   func(...func(consumeOptions))
}

type rawInputContext struct {
	Sequence string
	Stop     func()
}

// ---------------------------------------------------------------------------
// Keymap
// ---------------------------------------------------------------------------

// Keymap is the full key binding engine.
type Keymap[TTarget any, TEvent KeymapEvent] struct {
	host  KeymapHost[TTarget, TEvent]
	state *State[TTarget, TEvent]
	mu    sync.RWMutex

	listenerMu sync.RWMutex

	keypressListener      func(TEvent)
	keyreleaseListener    func(TEvent)
	rawListener           func(string) bool
	focusedTargetListener func(TTarget)

	listeners  map[HookName]map[int64]func(any)
	listenerID int64
	warnKeys   map[string]bool

	getPendingSequence func() []KeySequencePart
	getActiveKeys      func(ActiveKeyOptions) []ActiveKey[TTarget, TEvent]
}

// NewKeymap creates a new keymap bound to a host.
func NewKeymap[TTarget any, TEvent KeymapEvent](host KeymapHost[TTarget, TEvent]) (*Keymap[TTarget, TEvent], error) {
	if host.IsDestroyed != nil && host.IsDestroyed() {
		return nil, errors.New("keymap: cannot create a keymap for a destroyed host")
	}

	s := &State[TTarget, TEvent]{
		tokens:          make(map[string]resolvedKeyToken),
		patterns:        make(map[string]resolvedSequencePattern[TEvent]),
		layerFields:     make(map[string]layerFieldCompiler),
		bindingFields:   make(map[string]bindingFieldCompiler),
		commandFields:   make(map[string]commandFieldCompiler),
		layers:          make(map[*registeredLayer[TTarget, TEvent]]bool),
		sortedLayers:    []*registeredLayer[TTarget, TEvent]{},
		data:            make(map[string]any),
		readonlyData:    make(map[string]any),
		usedWarningKeys: make(map[string]bool),
	}

	k := &Keymap[TTarget, TEvent]{
		host:      host,
		state:     s,
		listeners: make(map[HookName]map[int64]func(any)),
		warnKeys:  make(map[string]bool),
	}

	k.keypressListener = func(ev TEvent) { k.handleKeyEvent(ev, false) }
	k.keyreleaseListener = func(ev TEvent) { k.handleKeyEvent(ev, true) }
	k.rawListener = func(seq string) bool { return k.handleRawSequence(seq) }
	k.focusedTargetListener = func(focused TTarget) { k.handleFocusedTargetChange(focused) }

	if host.OnKeyPress != nil {
		host.OnKeyPress(k.keypressListener)
	}
	if host.OnKeyRelease != nil {
		host.OnKeyRelease(k.keyreleaseListener)
	}
	if host.OnRawInput != nil {
		host.OnRawInput(k.rawListener)
	}
	if host.OnFocusChange != nil {
		host.OnFocusChange(k.focusedTargetListener)
	}
	if host.OnDestroy != nil {
		host.OnDestroy(func() { k.cleanup() })
	}

	k.getPendingSequence = k.getPendingSequenceState
	k.getActiveKeys = k.getActiveKeysState

	// Default binding parser: splits the sequence on whitespace, treating each
	// token as a key stroke (e.g. "ctrl+x ctrl+c").
	k.state.bindingParsers = append(k.state.bindingParsers, func(ctx bindingParserContext[TEvent]) *bindingParserResult {
		rest := ctx.Input[ctx.Index:]
		i := 0
		for i < len(rest) && (rest[i] == ' ' || rest[i] == '\t' || rest[i] == '\n') {
			i++
		}
		if i >= len(rest) {
			return nil
		}
		start := i
		for i < len(rest) && rest[i] != ' ' && rest[i] != '\t' && rest[i] != '\n' {
			i++
		}
		token := rest[start:i]
		name := token
		ctrl, shift, meta, super, hyper := false, false, false, false, false
		parts := strings.Split(token, "+")
		if len(parts) > 1 {
			name = parts[len(parts)-1]
			for _, m := range parts[:len(parts)-1] {
				switch strings.ToLower(m) {
				case "ctrl":
					ctrl = true
				case "shift":
					shift = true
				case "meta":
					meta = true
				case "super":
					super = true
				case "hyper":
					hyper = true
				}
			}
		}
		part := createKeySequencePart(KeyStrokeInput{Name: name, Ctrl: ctrl, Shift: shift, Meta: meta, Super: super, Hyper: hyper})
		return &bindingParserResult{
			Parts:      []KeySequencePart{part},
			NextIndex:  ctx.Index + i,
			UsedTokens: []string{token},
		}
	})

	return k, nil
}

// HostMetadata returns the host metadata.
func (k *Keymap[TTarget, TEvent]) HostMetadata() HostMetadata {
	return k.host.Metadata
}

// HasPendingSequence reports whether a key sequence is pending.
func (k *Keymap[TTarget, TEvent]) HasPendingSequence() bool {
	k.state.mu.RLock()
	defer k.state.mu.RUnlock()
	return k.state.pending != nil
}

// GetPendingSequence returns the pending key sequence parts.
func (k *Keymap[TTarget, TEvent]) GetPendingSequence() []KeySequencePart {
	return k.getPendingSequence()
}

func (k *Keymap[TTarget, TEvent]) getPendingSequenceState() []KeySequencePart {
	k.state.mu.RLock()
	defer k.state.mu.RUnlock()
	return k.collectSequencePartsFromPending(k.state.pending)
}

// ClearPendingSequence clears any pending sequence.
func (k *Keymap[TTarget, TEvent]) ClearPendingSequence() {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.setPendingSequenceLocked(nil)
}

// PopPendingSequence pops the first part of the pending sequence.
func (k *Keymap[TTarget, TEvent]) PopPendingSequence() bool {
	k.mu.Lock()
	defer k.mu.Unlock()
	return k.popPendingSequenceLocked()
}

// CreateKeyMatcher creates a matcher for a key.
func (k *Keymap[TTarget, TEvent]) CreateKeyMatcher(key interface{}) func(interface{}) bool {
	part := parseSingleKeyPart(key, k.state)
	match := part.Match
	return func(input interface{}) bool {
		if input == nil {
			return false
		}
		return resolveKeyMatchInput(input) == match
	}
}

// ParseKeySequence parses a key sequence string into parts.
func (k *Keymap[TTarget, TEvent]) ParseKeySequence(key interface{}) []KeySequencePart {
	return parseKeySequenceInput(key, k.state)
}

// FormatKey formats a key sequence into a string.
func (k *Keymap[TTarget, TEvent]) FormatKey(key interface{}, opts ...map[string]string) string {
	return stringifyKeySequence(parseKeySequenceInput(key, k.state), opts...)
}

// GetCommands returns commands matching the query.
func (k *Keymap[TTarget, TEvent]) GetCommands(query *CommandQuery[TTarget, TEvent]) []*Command {
	k.state.mu.RLock()
	defer k.state.mu.RUnlock()
	entries := k.getFilteredCommandEntriesLocked(query)
	out := make([]*Command, len(entries))
	for i, e := range entries {
		out[i] = e.commandState.command
	}
	return out
}

// GetCommandBindings returns bindings grouped by command.
func (k *Keymap[TTarget, TEvent]) GetCommandBindings(query CommandBindingsQuery[TTarget]) map[string][]ActiveBinding[TTarget, TEvent] {
	k.state.mu.RLock()
	defer k.state.mu.RUnlock()
	return k.getCommandBindingsLocked(query)
}

// RunCommand runs a named command.
func (k *Keymap[TTarget, TEvent]) RunCommand(cmd string, options *RunCommandOptions[TTarget, TEvent]) *RunCommandResult {
	k.mu.Lock()
	defer k.mu.Unlock()
	return k.runCommandLocked(cmd, options)
}

// DispatchCommand dispatches a command to the active layer chain.
func (k *Keymap[TTarget, TEvent]) DispatchCommand(cmd string, options *RunCommandOptions[TTarget, TEvent]) *RunCommandResult {
	k.mu.Lock()
	defer k.mu.Unlock()
	return k.dispatchCommandLocked(cmd, options)
}

// On registers a hook listener.
func (k *Keymap[TTarget, TEvent]) On(name HookName, fn func(any)) func() {
	k.listenerMu.Lock()
	k.listenerID++
	id := k.listenerID
	if k.listeners[name] == nil {
		k.listeners[name] = make(map[int64]func(any))
	}
	k.listeners[name][id] = fn
	k.listenerMu.Unlock()
	return func() {
		k.listenerMu.Lock()
		delete(k.listeners[name], id)
		k.listenerMu.Unlock()
	}
}

// InterceptKey registers a key intercept.
func (k *Keymap[TTarget, TEvent]) InterceptKey(fn func(keyInputContext[TTarget, TEvent]), priority int, release bool) func() {
	k.state.mu.Lock()
	k.state.hookID++
	id := k.state.hookID
	k.state.keyHooks = append(k.state.keyHooks, keyHookRegistration[TTarget, TEvent]{
		id:       id,
		Listener: fn,
		Config: struct {
			Priority int
			Release  bool
		}{Priority: priority, Release: release},
	})
	k.state.mu.Unlock()
	return func() {
		k.state.mu.Lock()
		out := []keyHookRegistration[TTarget, TEvent]{}
		for _, h := range k.state.keyHooks {
			if h.id != id {
				out = append(out, h)
			}
		}
		k.state.keyHooks = out
		k.state.mu.Unlock()
	}
}

// InterceptKeyAfter registers a key:after intercept.
func (k *Keymap[TTarget, TEvent]) InterceptKeyAfter(fn func(keyAfterInputContext[TTarget, TEvent]), priority int, release bool) func() {
	k.state.mu.Lock()
	k.state.hookID++
	id := k.state.hookID
	k.state.keyAfterHooks = append(k.state.keyAfterHooks, keyAfterHookRegistration[TTarget, TEvent]{
		id:       id,
		Listener: fn,
		Config: struct {
			Priority int
			Release  bool
		}{Priority: priority, Release: release},
	})
	k.state.mu.Unlock()
	return func() {
		k.state.mu.Lock()
		out := []keyAfterHookRegistration[TTarget, TEvent]{}
		for _, h := range k.state.keyAfterHooks {
			if h.id != id {
				out = append(out, h)
			}
		}
		k.state.keyAfterHooks = out
		k.state.mu.Unlock()
	}
}

// InterceptRaw registers a raw input intercept.
func (k *Keymap[TTarget, TEvent]) InterceptRaw(fn func(rawInputContext), priority int) func() {
	k.state.mu.Lock()
	k.state.hookID++
	id := k.state.hookID
	k.state.rawHooks = append(k.state.rawHooks, rawHookRegistration{
		id:       id,
		Listener: fn,
		Config:   struct{ Priority int }{Priority: priority},
	})
	k.state.mu.Unlock()
	return func() {
		k.state.mu.Lock()
		out := []rawHookRegistration{}
		for _, h := range k.state.rawHooks {
			if h.id != id {
				out = append(out, h)
			}
		}
		k.state.rawHooks = out
		k.state.mu.Unlock()
	}
}

// RegisterLayer registers a layer and returns a release function.
func (k *Keymap[TTarget, TEvent]) RegisterLayer(layer Layer[TTarget, TEvent]) func() {
	rl := k.compileLayer(layer, false, -1)
	if rl == nil {
		return func() {}
	}
	k.addLayerLocked(rl)
	return func() {
		k.mu.Lock()
		k.unregisterLayerLocked(rl)
		k.mu.Unlock()
	}
}

// CreateSlot creates a layer slot.
func (k *Keymap[TTarget, TEvent]) CreateSlot() *layerSlot[TTarget, TEvent] {
	k.mu.Lock()
	defer k.mu.Unlock()
	slot := &layerSlot[TTarget, TEvent]{order: k.state.order}
	k.state.order++
	return slot
}

// ReplaceLayer replaces the layer in a slot.
func (k *Keymap[TTarget, TEvent]) ReplaceLayer(slot *layerSlot[TTarget, TEvent], layer *Layer[TTarget, TEvent]) *registeredLayer[TTarget, TEvent] {
	k.mu.Lock()
	defer k.mu.Unlock()
	if layer == nil {
		if slot.layer != nil {
			k.unregisterLayerLocked(slot.layer)
			slot.layer = nil
		}
		return nil
	}
	rl := k.compileLayer(*layer, false, slot.order)
	if rl == nil {
		return nil
	}
	if slot.layer != nil {
		k.unregisterLayerLocked(slot.layer)
	}
	slot.layer = rl
	k.addLayerLocked(rl)
	return rl
}

// layerSlot is a mount position for replaceLayer.
type layerSlot[TTarget any, TEvent KeymapEvent] struct {
	order int
	layer *registeredLayer[TTarget, TEvent]
}

// cleanup releases all resources.
func (k *Keymap[TTarget, TEvent]) cleanup() {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.setPendingSequenceLocked(nil)
	for layer := range k.state.layers {
		k.detachReactiveMatchersLayer(layer)
		if layer.offTargetDestroy != nil {
			layer.offTargetDestroy()
			layer.offTargetDestroy = nil
		}
	}
	k.state.layers = make(map[*registeredLayer[TTarget, TEvent]]bool)
	k.state.sortedLayers = []*registeredLayer[TTarget, TEvent]{}
}

func (k *Keymap[TTarget, TEvent]) attachReactiveMatchersLayer(rl *registeredLayer[TTarget, TEvent]) {
	for _, m := range rl.matchers {
		if m.subscribe != nil {
			m.dispose = m.subscribe(func() {
				k.queueStateChange()
			})
		}
	}
}

func (k *Keymap[TTarget, TEvent]) detachReactiveMatchersLayer(rl *registeredLayer[TTarget, TEvent]) {
	for _, m := range rl.matchers {
		if m.dispose != nil {
			m.dispose()
			m.dispose = nil
		}
	}
}

func (k *Keymap[TTarget, TEvent]) attachReactiveMatchersBinding(b *bindingState[TTarget, TEvent]) {
	for _, m := range b.matchers {
		if m.subscribe != nil {
			m.dispose = m.subscribe(func() {
				k.queueStateChange()
			})
		}
	}
}

func (k *Keymap[TTarget, TEvent]) detachReactiveMatchersBinding(b *bindingState[TTarget, TEvent]) {
	for _, m := range b.matchers {
		if m.dispose != nil {
			m.dispose()
			m.dispose = nil
		}
	}
}

// ---------------------------------------------------------------------------
// Layer compilation and registration
// ---------------------------------------------------------------------------

func (k *Keymap[TTarget, TEvent]) registerLayerLocked(layer Layer[TTarget, TEvent]) *registeredLayer[TTarget, TEvent] {
	rl := k.compileLayer(layer, false, -1)
	if rl == nil {
		return nil
	}
	k.addLayerLocked(rl)
	return rl
}

func (k *Keymap[TTarget, TEvent]) compileLayer(layer Layer[TTarget, TEvent], ownCommands bool, slotOrder int) *registeredLayer[TTarget, TEvent] {
	if any(layer.Target) != nil && k.host.IsTargetDestroyed != nil && k.host.IsTargetDestroyed(layer.Target) {
		k.emitError("destroyed-layer-target", nil, "keymap: cannot register a layer for a destroyed target")
		return nil
	}

	sourceBindings := layer.Bindings
	for _, t := range k.state.layerBindingsTransformers {
		next, ok := t(sourceBindings, layerBindingsTransformerContext[TTarget, TEvent]{
			Layer:            layer,
			ValidateBindings: func(b interface{}) map[string]string { return map[string]string{} },
		})
		if !ok {
			continue
		}
		sourceBindings = next
	}

	commands := layer.Commands
	for _, t := range k.state.commandTransformers {
		transformed := []Command{}
		for _, cmd := range commands {
			tc := cmd
			keepOriginal := true
			t(&tc, commandTransformerContext[TTarget, TEvent]{
				Layer:        layer,
				Add:          func(c *Command) { transformed = append(transformed, *c) },
				SkipOriginal: func() { keepOriginal = false },
			})
			if keepOriginal {
				transformed = append(transformed, tc)
			}
		}
		commands = transformed
	}
	_ = ownCommands

	order := slotOrder
	if order < 0 {
		order = k.state.order
		k.state.order++
	}

	bindingStates := k.compileBindings(sourceBindings, order, layer.Fields)
	if len(bindingStates) == 0 && len(commands) == 0 {
		return nil
	}

	rl := &registeredLayer[TTarget, TEvent]{
		order:          order,
		target:         layer.Target,
		targetMode:     k.normalizeTargetMode(layer),
		priority:       layer.Priority,
		commands:       []commandState[TTarget, TEvent]{},
		sourceBindings: sourceBindings,
		bindings:       bindingStates,
		root:           buildSequenceTree(bindingStates, k.state.patterns),
	}
	for i := range commands {
		rl.commands = append(rl.commands, commandState[TTarget, TEvent]{command: &commands[i]})
	}

	k.attachReactiveMatchersLayer(rl)
	for i := range rl.bindings {
		k.attachReactiveMatchersBinding(&rl.bindings[i])
	}

	if any(rl.target) != nil && k.host.OnTargetDestroy != nil {
		target := rl.target
		rl.offTargetDestroy = k.host.OnTargetDestroy(target, func() {
			k.mu.Lock()
			k.unregisterLayerLocked(rl)
			k.mu.Unlock()
		})
	}

	k.queueStateChange()
	return rl
}

func (k *Keymap[TTarget, TEvent]) addLayerLocked(rl *registeredLayer[TTarget, TEvent]) {
	k.state.layers[rl] = true
	if rl.order < k.state.order-1 {
		sorted := make([]*registeredLayer[TTarget, TEvent], 0, len(k.state.layers))
		for l := range k.state.layers {
			sorted = append(sorted, l)
		}
		sort.Slice(sorted, func(i, j int) bool { return sorted[i].order < sorted[j].order })
		k.state.layers = make(map[*registeredLayer[TTarget, TEvent]]bool)
		for _, l := range sorted {
			k.state.layers[l] = true
		}
	}
	k.state.sortedLayers = append(k.state.sortedLayers, rl)
	sort.Slice(k.state.sortedLayers, func(i, j int) bool {
		if k.state.sortedLayers[j].priority != k.state.sortedLayers[i].priority {
			return k.state.sortedLayers[j].priority > k.state.sortedLayers[i].priority
		}
		return k.state.sortedLayers[j].order > k.state.sortedLayers[i].order
	})

	if len(rl.commands) > 0 {
		k.ensureValidPendingSequenceLocked()
	}
}

func (k *Keymap[TTarget, TEvent]) unregisterLayerLocked(rl *registeredLayer[TTarget, TEvent]) {
	if !k.state.layers[rl] {
		return
	}
	delete(k.state.layers, rl)
	sorted := make([]*registeredLayer[TTarget, TEvent], 0, len(k.state.layers))
	for l := range k.state.layers {
		sorted = append(sorted, l)
	}
	k.state.sortedLayers = sorted
	sort.Slice(k.state.sortedLayers, func(i, j int) bool {
		if k.state.sortedLayers[j].priority != k.state.sortedLayers[i].priority {
			return k.state.sortedLayers[j].priority > k.state.sortedLayers[i].priority
		}
		return k.state.sortedLayers[j].order > k.state.sortedLayers[i].order
	})

	k.detachReactiveMatchersLayer(rl)
	for i := range rl.bindings {
		k.detachReactiveMatchersBinding(&rl.bindings[i])
	}
	if rl.offTargetDestroy != nil {
		rl.offTargetDestroy()
		rl.offTargetDestroy = nil
	}

	if k.state.pending != nil {
		for _, c := range k.state.pending.captures {
			if c.layer == rl {
				k.setPendingSequenceLocked(nil)
				break
			}
		}
	}
	k.queueStateChange()
}

func (k *Keymap[TTarget, TEvent]) normalizeTargetMode(layer Layer[TTarget, TEvent]) TargetMode {
	if layer.TargetMode != "" {
		if any(layer.Target) == nil {
			panic(fmt.Sprintf("keymap: targetMode %q requires a target", layer.TargetMode))
		}
		return layer.TargetMode
	}
	if any(layer.Target) != nil {
		return TargetModeFocusWithin
	}
	return ""
}

// ---------------------------------------------------------------------------
// Binding compilation
// ---------------------------------------------------------------------------

func (k *Keymap[TTarget, TEvent]) compileBindings(bindings []Binding[TTarget, TEvent], sourceLayerOrder int, fields map[string]any) []bindingState[TTarget, TEvent] {
	out := []bindingState[TTarget, TEvent]{}
	for bindingIndex, binding := range bindings {
		expandedKeys := []string{binding.Key}
		for _, exp := range k.state.bindingExpanders {
			next := []string{}
			for _, cand := range expandedKeys {
				if result, ok := exp(bindingExpanderContext{Input: cand, Layer: fields}); ok {
					for _, e := range result {
						next = append(next, e.Key)
					}
				} else {
					next = append(next, cand)
				}
			}
			expandedKeys = next
		}

		for _, expKey := range expandedKeys {
			parts := parseKeySequenceInput(expKey, k.state)
			if len(parts) == 0 {
				continue
			}
			transformed := []*Binding[TTarget, TEvent]{&binding}
			for _, t := range k.state.bindingTransformers {
				extra := []*Binding[TTarget, TEvent]{}
				keepOriginal := true
				tb := binding
				t(&tb, bindingTransformerContext[TTarget, TEvent]{
					Layer: fields,
					ParseKey: func(in interface{}) KeySequencePart {
						return parseSingleKeyPart(in, k.state)
					},
					Add:          func(nb *Binding[TTarget, TEvent]) { extra = append(extra, nb) },
					SkipOriginal: func() { keepOriginal = false },
				})
				if keepOriginal {
					transformed = append(transformed, &tb)
				}
				transformed = append(transformed, extra...)
			}

			for _, tb := range transformed {
				event := normalizeBindingEvent(tb.Event)
				sequence := cloneKeySequence(parts)
				if len(sequence) == 0 {
					continue
				}
				if event == BindingEventRelease && len(sequence) > 1 {
					k.emitError("binding-compile-error", nil, "keymap: release bindings only support a single key stroke")
					continue
				}
				cmd := normalizeBindingCommand(tb.Cmd)
				bs := bindingState[TTarget, TEvent]{
					binding:          tb,
					sequence:         sequence,
					command:          cmd,
					event:            event,
					sourceTarget:     layerTarget(binding),
					sourceLayerOrder: sourceLayerOrder,
					bindingIndex:     bindingIndex,
					preventDefault:   tb.PreventDefault,
					Fallthrough:      tb.Fallthrough,
					attrs:            map[string]any{},
				}
				if cmd.Kind == BindingCommandHandler && cmd.Run != nil {
					bs.run = cmd.Run
				}
				out = append(out, bs)
			}
		}
	}
	return out
}

func layerTarget[TTarget any, TEvent KeymapEvent](b Binding[TTarget, TEvent]) TTarget {
	var zero TTarget
	return zero
}

func normalizeBindingEvent(event BindingEvent) BindingEvent {
	if event == "" || event == BindingEventPress {
		return BindingEventPress
	}
	if event == BindingEventRelease {
		return BindingEventRelease
	}
	return BindingEventPress
}

func normalizeBindingCommand[TTarget any, TEvent KeymapEvent](cmd BindingCommand[TTarget, TEvent]) BindingCommand[TTarget, TEvent] {
	return cmd
}

// ---------------------------------------------------------------------------
// Conditions
// ---------------------------------------------------------------------------

func (k *Keymap[TTarget, TEvent]) matchesConditions(requires []struct {
	Name  string
	Value any
}, matchers []runtimeMatcher) bool {
	for _, r := range requires {
		v, ok := k.state.data[r.Name]
		if !ok || v != r.Value {
			return false
		}
	}
	for _, m := range matchers {
		if m.match == nil || !m.match() {
			return false
		}
	}
	return true
}

func (k *Keymap[TTarget, TEvent]) matchesBindingConditions(bs *bindingState[TTarget, TEvent], focused TTarget) bool {
	if !k.matchesConditions(bs.requires, bs.matchers) {
		return false
	}
	return k.isBindingVisible(bs, focused)
}

func (k *Keymap[TTarget, TEvent]) isBindingVisible(bs *bindingState[TTarget, TEvent], focused TTarget) bool {
	if bs.command.Kind == BindingCommandNone || bs.run != nil {
		return true
	}
	view := k.getActiveCommandView(focused)
	if _, ok := view.reachableByName[bs.command.Name]; ok {
		return true
	}
	return k.getFallbackResolvedCommand(bs.command.Name, focused) != nil
}

func (k *Keymap[TTarget, TEvent]) bindingCommandAttrs(bs *bindingState[TTarget, TEvent], focused TTarget) map[string]any {
	if bs.command.Kind == BindingCommandNone {
		return nil
	}
	view := k.getActiveCommandView(focused)
	if e, ok := view.reachableByName[bs.command.Name]; ok {
		return e.commandState.attrs
	}
	fb := k.getFallbackResolvedCommand(bs.command.Name, focused)
	if fb != nil {
		return fb.commandState.attrs
	}
	return nil
}

// ---------------------------------------------------------------------------
// Runtime data
// ---------------------------------------------------------------------------

func (k *Keymap[TTarget, TEvent]) GetData(name string) any {
	k.state.mu.RLock()
	defer k.state.mu.RUnlock()
	return k.state.data[name]
}

func (k *Keymap[TTarget, TEvent]) SetData(name string, value any) {
	k.mu.Lock()
	defer k.mu.Unlock()
	if value == nil {
		if _, ok := k.state.data[name]; !ok {
			return
		}
		delete(k.state.data, name)
		k.state.dataVersion++
		k.ensureValidPendingSequenceLocked()
		k.queueStateChange()
		return
	}
	old := k.state.data[name]
	if old == value {
		return
	}
	k.state.data[name] = value
	k.state.dataVersion++
	k.ensureValidPendingSequenceLocked()
	k.queueStateChange()
}

func (k *Keymap[TTarget, TEvent]) GetReadOnlyData() map[string]any {
	k.state.mu.RLock()
	defer k.state.mu.RUnlock()
	if k.state.readonlyDataVersion == k.state.dataVersion {
		return k.state.readonlyData
	}
	copy := make(map[string]any, len(k.state.data))
	for kk, v := range k.state.data {
		copy[kk] = v
	}
	k.state.readonlyData = copy
	k.state.readonlyDataVersion = k.state.dataVersion
	return copy
}

// ---------------------------------------------------------------------------
// Activation / active keys
// ---------------------------------------------------------------------------

func (k *Keymap[TTarget, TEvent]) getActiveKeysState(opts ActiveKeyOptions) []ActiveKey[TTarget, TEvent] {
	k.state.mu.RLock()
	defer k.state.mu.RUnlock()
	focused := k.getFocusedTarget()
	pending := k.ensureValidPendingSequenceLockedReturn()
	var activeLayers []*registeredLayer[TTarget, TEvent]
	if pending != nil {
		activeLayers = []*registeredLayer[TTarget, TEvent]{}
	} else {
		activeLayers = k.getActiveLayers(focused)
	}
	if pending != nil {
		return k.collectActiveKeysFromPending(pending.captures, opts, focused)
	}
	return k.collectActiveKeysAtRoot(activeLayers, opts, focused)
}

func (k *Keymap[TTarget, TEvent]) collectActiveKeysFromPending(captures []pendingSequenceCapture[TTarget, TEvent], opts ActiveKeyOptions, focused TTarget) []ActiveKey[TTarget, TEvent] {
	activeKeys := map[string]*activeKeyState[TTarget, TEvent]{}
	stopped := map[string]bool{}
	opts2 := activeOptionsForCaptures(captures, k.state.patterns)
	seen := map[string]bool{}
	for _, opt := range opts2 {
		match := opt.part.Match
		if seen[match] {
			continue
		}
		seen[match] = true
		k.collectActiveKeyOption(opt, opts2, activeKeys, stopped, opts, focused)
	}
	return materializeActiveKeys(activeKeys, opts, focused)
}

type activeKeyState[TTarget any, TEvent KeymapEvent] struct {
	stroke         KeyStroke
	display        string
	tokenName      string
	continues      bool
	firstBinding   *bindingState[TTarget, TEvent]
	commandBinding *bindingState[TTarget, TEvent]
	bindings       []bindingState[TTarget, TEvent]
}

func (k *Keymap[TTarget, TEvent]) collectActiveKeyOption(opt sequenceActiveOption[TTarget, TEvent], siblings []sequenceActiveOption[TTarget, TEvent], activeKeys map[string]*activeKeyState[TTarget, TEvent], stopped map[string]bool, opts ActiveKeyOptions, focused TTarget) {
	if stopped[opt.part.Match] {
		return
	}
	selection := k.selectActiveKeyOption(opt, siblings, opts, focused)
	if selection == nil {
		return
	}
	existing := activeKeys[opt.part.Match]
	if existing == nil {
		activeKeys[opt.part.Match] = &activeKeyState[TTarget, TEvent]{
			stroke:         opt.part.Stroke,
			display:        selection.display,
			tokenName:      selection.tokenName,
			continues:      selection.continues,
			firstBinding:   selection.firstBinding,
			commandBinding: selection.commandBinding,
			bindings:       selection.bindings,
		}
	} else {
		if existing.firstBinding == nil && selection.firstBinding != nil {
			existing.firstBinding = selection.firstBinding
		}
		if existing.commandBinding == nil && selection.commandBinding != nil {
			existing.commandBinding = selection.commandBinding
		}
		if selection.continues {
			existing.continues = true
		}
		if len(selection.bindings) > 0 {
			if existing.bindings == nil {
				existing.bindings = append([]bindingState[TTarget, TEvent]{}, selection.bindings...)
			} else {
				existing.bindings = append(existing.bindings, selection.bindings...)
			}
		}
	}
	if selection.stop {
		stopped[opt.part.Match] = true
	}
}

type sequenceActiveOption[TTarget any, TEvent KeymapEvent] struct {
	part      KeySequencePart
	binding   *bindingState[TTarget, TEvent]
	index     int
	exact     bool
	continues bool
}

func activeOptionsForCaptures[TTarget any, TEvent KeymapEvent](captures []pendingSequenceCapture[TTarget, TEvent], patterns map[string]resolvedSequencePattern[TEvent]) []sequenceActiveOption[TTarget, TEvent] {
	out := []sequenceActiveOption[TTarget, TEvent]{}
	for _, c := range captures {
		if !captureHasMinimum(c, patterns) {
			continue
		}
		index := c.index + 1
		if index >= len(c.binding.sequence) {
			continue
		}
		part := c.binding.sequence[index]
		out = append(out, sequenceActiveOption[TTarget, TEvent]{
			part:      part,
			binding:   c.binding,
			index:     index,
			exact:     index == len(c.binding.sequence)-1,
			continues: index < len(c.binding.sequence)-1,
		})
	}
	return out
}

func captureHasMinimum[TTarget any, TEvent KeymapEvent](c pendingSequenceCapture[TTarget, TEvent], patterns map[string]resolvedSequencePattern[TEvent]) bool {
	part := c.binding.sequence[c.index]
	if part.PatternName == "" {
		return true
	}
	p := patterns[part.PatternName]
	if p.matcher == nil {
		return true
	}
	count := 0
	if len(c.patterns) > 0 {
		last := c.patterns[len(c.patterns)-1]
		if last.name == part.PatternName {
			count = len(last.values)
		}
	}
	return count >= p.min
}

func (k *Keymap[TTarget, TEvent]) selectActiveKeyOption(opt sequenceActiveOption[TTarget, TEvent], siblings []sequenceActiveOption[TTarget, TEvent], opts ActiveKeyOptions, focused TTarget) *activeKeySelection[TTarget, TEvent] {
	matching := []sequenceActiveOption[TTarget, TEvent]{}
	for _, s := range siblings {
		if s.part.Match == opt.part.Match {
			matching = append(matching, s)
		}
	}
	exactBindings := []*bindingState[TTarget, TEvent]{}
	for _, m := range matching {
		if m.exact {
			exactBindings = append(exactBindings, m.binding)
		}
	}
	selected := k.selectActiveBindings(toBindingPtrSlice(bindingStateSlice(exactBindings)), focused)
	continues := false
	for _, m := range matching {
		if m.continues {
			continues = true
		}
	}
	if !continues && selected == nil {
		return nil
	}
	display := ""
	tokenName := ""
	for _, m := range matching {
		if display == "" {
			display = m.part.Display
			tokenName = m.part.TokenName
		}
	}
	return &activeKeySelection[TTarget, TEvent]{
		display:        display,
		tokenName:      tokenName,
		continues:      continues,
		firstBinding:   nil,
		commandBinding: nil,
		bindings:       nil,
		stop:           continues || (selected != nil && selected.stop),
	}
}

type activeKeySelection[TTarget any, TEvent KeymapEvent] struct {
	display        string
	tokenName      string
	continues      bool
	firstBinding   *bindingState[TTarget, TEvent]
	commandBinding *bindingState[TTarget, TEvent]
	bindings       []bindingState[TTarget, TEvent]
	stop           bool
}

func (k *Keymap[TTarget, TEvent]) selectActiveBindings(bindings []*bindingState[TTarget, TEvent], focused TTarget) *activeBindingSelect[TTarget, TEvent] {
	selected := []*bindingState[TTarget, TEvent]{}
	var commandBinding *bindingState[TTarget, TEvent]
	stop := false
	for _, b := range bindings {
		if !k.matchesBindingConditions(b, focused) {
			continue
		}
		selected = append(selected, b)
		if b.command.Kind == BindingCommandNone && b.run == nil {
			continue
		}
		if commandBinding == nil {
			commandBinding = b
		}
		if !b.Fallthrough {
			stop = true
			break
		}
	}
	if len(selected) == 0 {
		return nil
	}
	return &activeBindingSelect[TTarget, TEvent]{bindings: selected, commandBinding: commandBinding, stop: stop}
}

type activeBindingSelect[TTarget any, TEvent KeymapEvent] struct {
	bindings       []*bindingState[TTarget, TEvent]
	commandBinding *bindingState[TTarget, TEvent]
	stop           bool
}

func materializeActiveKeys[TTarget any, TEvent KeymapEvent](activeKeys map[string]*activeKeyState[TTarget, TEvent], opts ActiveKeyOptions, focused TTarget) []ActiveKey[TTarget, TEvent] {
	out := []ActiveKey[TTarget, TEvent]{}
	for _, st := range activeKeys {
		if st.commandBinding == nil && !st.continues {
			continue
		}
		ak := ActiveKey[TTarget, TEvent]{
			Stroke:    st.stroke,
			Display:   st.display,
			TokenName: st.tokenName,
			Continues: st.continues,
		}
		if st.commandBinding != nil {
			ak.Command = st.commandBinding.command
		}
		out = append(out, ak)
	}
	return out
}

func (k *Keymap[TTarget, TEvent]) collectActiveKeysAtRoot(activeLayers []*registeredLayer[TTarget, TEvent], opts ActiveKeyOptions, focused TTarget) []ActiveKey[TTarget, TEvent] {
	activeKeys := map[string]*activeKeyState[TTarget, TEvent]{}
	stopped := map[string]bool{}
	for _, layer := range activeLayers {
		if (len(layer.root.children) == 0 && len(layer.root.patternChildren) == 0) || !k.matchesConditionsLayer(layer) {
			continue
		}
		k.collectActiveKeyNodes(layer.root.children, activeKeys, stopped, opts, focused)
		k.collectActiveKeyNodesSlice(layer.root.patternChildren, activeKeys, stopped, opts, focused)
	}
	return materializeActiveKeys(activeKeys, opts, focused)
}

func (k *Keymap[TTarget, TEvent]) collectActiveKeyNodes(children map[string]*sequenceNode[TTarget, TEvent], activeKeys map[string]*activeKeyState[TTarget, TEvent], stopped map[string]bool, opts ActiveKeyOptions, focused TTarget) {
	for _, node := range children {
		if node.match == "" || stopped[node.match] || node.stroke == nil {
			continue
		}
		selection := k.selectActiveKeyNode(node, opts, focused)
		if selection == nil {
			continue
		}
		existing := activeKeys[node.match]
		if existing == nil {
			activeKeys[node.match] = &activeKeyState[TTarget, TEvent]{
				stroke:         *node.stroke,
				display:        selection.display,
				tokenName:      selection.tokenName,
				continues:      selection.continues,
				firstBinding:   selection.firstBinding,
				commandBinding: selection.commandBinding,
				bindings:       selection.bindings,
			}
		} else {
			if existing.firstBinding == nil && selection.firstBinding != nil {
				existing.firstBinding = selection.firstBinding
			}
			if existing.commandBinding == nil && selection.commandBinding != nil {
				existing.commandBinding = selection.commandBinding
			}
			if selection.continues {
				existing.continues = true
			}
			if len(selection.bindings) > 0 {
				if existing.bindings == nil {
					existing.bindings = append([]bindingState[TTarget, TEvent]{}, selection.bindings...)
				} else {
					existing.bindings = append(existing.bindings, selection.bindings...)
				}
			}
		}
		if selection.stop {
			stopped[node.match] = true
		}
	}
}

func (k *Keymap[TTarget, TEvent]) collectActiveKeyNodesSlice(children []*sequenceNode[TTarget, TEvent], activeKeys map[string]*activeKeyState[TTarget, TEvent], stopped map[string]bool, opts ActiveKeyOptions, focused TTarget) {
	for _, node := range children {
		if node.match == "" || stopped[node.match] || node.stroke == nil {
			continue
		}
		selection := k.selectActiveKeyNode(node, opts, focused)
		if selection == nil {
			continue
		}
		existing := activeKeys[node.match]
		if existing == nil {
			activeKeys[node.match] = &activeKeyState[TTarget, TEvent]{
				stroke:         *node.stroke,
				display:        selection.display,
				tokenName:      selection.tokenName,
				continues:      selection.continues,
				firstBinding:   selection.firstBinding,
				commandBinding: selection.commandBinding,
				bindings:       selection.bindings,
			}
		} else {
			if existing.firstBinding == nil && selection.firstBinding != nil {
				existing.firstBinding = selection.firstBinding
			}
			if existing.commandBinding == nil && selection.commandBinding != nil {
				existing.commandBinding = selection.commandBinding
			}
			if selection.continues {
				existing.continues = true
			}
			if len(selection.bindings) > 0 {
				if existing.bindings == nil {
					existing.bindings = append([]bindingState[TTarget, TEvent]{}, selection.bindings...)
				} else {
					existing.bindings = append(existing.bindings, selection.bindings...)
				}
			}
		}
		if selection.stop {
			stopped[node.match] = true
		}
	}
}

func (k *Keymap[TTarget, TEvent]) selectActiveKeyNode(node *sequenceNode[TTarget, TEvent], opts ActiveKeyOptions, focused TTarget) *activeKeySelection[TTarget, TEvent] {
	continues := len(node.children) > 0 || len(node.patternChildren) > 0
	if !continues {
		bindings := node.bindings
		if k.needsExactBindingPrefilter(node) {
			bindings = k.collectMatchingBindings(node.bindings, focused)
		}
		selected := k.selectActiveBindings(toBindingPtrSlice(bindings), focused)
		if selected == nil {
			return nil
		}
		display, tokenName := k.getPartPresentation(node.bindings, node.depth-1)
		return &activeKeySelection[TTarget, TEvent]{
			display:        display,
			tokenName:      tokenName,
			continues:      false,
			firstBinding:   nil,
			commandBinding: selected.commandBinding,
			bindings:       bindingStatePtrSliceToSlice(selected.bindings),
			stop:           selected.stop,
		}
	}
	reachable := k.collectMatchingBindings(node.reachableBindings, focused)
	if len(reachable) == 0 {
		return nil
	}
	display, tokenName := k.getPartPresentation(reachable, node.depth-1)
	return &activeKeySelection[TTarget, TEvent]{
		display:        display,
		tokenName:      tokenName,
		continues:      true,
		firstBinding:   nil,
		commandBinding: nil,
		bindings:       nil,
		stop:           true,
	}
}

func (k *Keymap[TTarget, TEvent]) needsExactBindingPrefilter(node *sequenceNode[TTarget, TEvent]) bool {
	for _, b := range node.bindings {
		if len(b.matchers) > 0 {
			return true
		}
	}
	return false
}

func (k *Keymap[TTarget, TEvent]) collectMatchingBindings(bindings []bindingState[TTarget, TEvent], focused TTarget) []bindingState[TTarget, TEvent] {
	out := []bindingState[TTarget, TEvent]{}
	for _, b := range bindings {
		if k.matchesBindingConditions(&b, focused) {
			out = append(out, b)
		}
	}
	return out
}

func (k *Keymap[TTarget, TEvent]) getPartPresentation(bindings []bindingState[TTarget, TEvent], partIndex int) (string, string) {
	for _, b := range bindings {
		if partIndex >= 0 && partIndex < len(b.sequence) {
			part := b.sequence[partIndex]
			return part.Display, part.TokenName
		}
	}
	return "", ""
}

func bindingStateSlice[TTarget any, TEvent KeymapEvent](ptrSlice []*bindingState[TTarget, TEvent]) []bindingState[TTarget, TEvent] {
	out := []bindingState[TTarget, TEvent]{}
	for _, p := range ptrSlice {
		if p != nil {
			out = append(out, *p)
		}
	}
	return out
}

func bindingStatePtrSliceToSlice[TTarget any, TEvent KeymapEvent](ptrSlice []*bindingState[TTarget, TEvent]) []bindingState[TTarget, TEvent] {
	return bindingStateSlice(ptrSlice)
}

func toBindingPtrSlice[TTarget any, TEvent KeymapEvent](bindings []bindingState[TTarget, TEvent]) []*bindingState[TTarget, TEvent] {
	out := []*bindingState[TTarget, TEvent]{}
	for i := range bindings {
		out = append(out, &bindings[i])
	}
	return out
}

func (k *Keymap[TTarget, TEvent]) matchesConditionsLayer(rl *registeredLayer[TTarget, TEvent]) bool {
	return k.matchesConditions(rl.requires, rl.matchers)
}

func (k *Keymap[TTarget, TEvent]) getActiveLayers(focused TTarget) []*registeredLayer[TTarget, TEvent] {
	if k.state.activeLayersCacheVersion == k.state.cacheVersion && any(k.state.activeLayersCacheFocused) == any(focused) {
		return k.state.activeLayersCache
	}
	k.state.activeLayersCacheVersion = k.state.cacheVersion
	k.state.activeLayersCacheFocused = focused
	k.state.activeLayersCache = getActiveLayersForFocused(k.state.sortedLayers, k.host, focused)
	return k.state.activeLayersCache
}

func (k *Keymap[TTarget, TEvent]) getFocusedTarget() TTarget {
	if k.host.IsDestroyed != nil && k.host.IsDestroyed() {
		var zero TTarget
		return zero
	}
	if k.host.GetFocusedTarget != nil {
		return k.host.GetFocusedTarget()
	}
	var zero TTarget
	return zero
}

func getActiveLayersForFocused[TTarget any, TEvent KeymapEvent](layers []*registeredLayer[TTarget, TEvent], host KeymapHost[TTarget, TEvent], focused TTarget) []*registeredLayer[TTarget, TEvent] {
	out := []*registeredLayer[TTarget, TEvent]{}
	path := getActivationPath(host, focused)
	for _, layer := range layers {
		if isLayerActiveForFocused(host, layer, focused, path) {
			out = append(out, layer)
		}
	}
	return out
}

func getActivationPath[TTarget any, TEvent KeymapEvent](host KeymapHost[TTarget, TEvent], focused TTarget) map[any]bool {
	path := map[any]bool{}
	current := focused
	for {
		if any(current) == nil {
			break
		}
		path[any(current)] = true
		var next TTarget
		if host.GetParentTarget != nil {
			next = host.GetParentTarget(current)
		}
		if any(next) == nil || any(next) == any(current) {
			break
		}
		current = next
	}
	return path
}

func isLayerActiveForFocused[TTarget any, TEvent KeymapEvent](host KeymapHost[TTarget, TEvent], layer *registeredLayer[TTarget, TEvent], focused TTarget, path map[any]bool) bool {
	if any(layer.target) == nil {
		return true
	}
	if host.IsTargetDestroyed != nil && host.IsTargetDestroyed(layer.target) {
		return false
	}
	if layer.targetMode == TargetModeFocus {
		return any(layer.target) == any(focused)
	}
	return path[any(layer.target)]
}

// ---------------------------------------------------------------------------
// Pending sequence management
// ---------------------------------------------------------------------------

func (k *Keymap[TTarget, TEvent]) setPendingSequenceLocked(next *pendingSequenceState[TTarget, TEvent]) {
	if k.state.pending == nil && next == nil {
		return
	}
	if k.state.pending != nil && next != nil && pendingSequenceEqual(k.state.pending, next) {
		return
	}
	k.state.pending = next
	k.notifyHook(HookPendingSeq, k.collectSequencePartsFromPending(next))
	k.queueStateChange()
}

func (k *Keymap[TTarget, TEvent]) ensureValidPendingSequenceLocked() *pendingSequenceState[TTarget, TEvent] {
	return k.ensureValidPendingSequenceLockedReturn()
}

func (k *Keymap[TTarget, TEvent]) ensureValidPendingSequenceLockedReturn() *pendingSequenceState[TTarget, TEvent] {
	pending := k.state.pending
	if pending == nil {
		return nil
	}
	focused := k.getFocusedTarget()
	filtered := []pendingSequenceCapture[TTarget, TEvent]{}
	for _, c := range pending.captures {
		if !k.state.layers[c.layer] {
			continue
		}
		if !k.matchesConditionsLayer(c.layer) {
			continue
		}
		if !k.matchesBindingConditions(c.binding, focused) {
			continue
		}
		if !captureHasContinuations(c.binding, c.index, k.state.patterns) {
			continue
		}
		filtered = append(filtered, c)
	}
	if len(filtered) == 0 {
		k.setPendingSequenceLocked(nil)
		return nil
	}
	if len(filtered) != len(pending.captures) {
		k.setPendingSequenceLocked(&pendingSequenceState[TTarget, TEvent]{captures: filtered})
	}
	return k.state.pending
}

func (k *Keymap[TTarget, TEvent]) popPendingSequenceLocked() bool {
	pending := k.ensureValidPendingSequenceLockedReturn()
	if pending == nil {
		return false
	}
	if len(pending.captures) == 0 || len(pending.captures[0].parts) <= 1 {
		k.setPendingSequenceLocked(nil)
		return true
	}
	next := []pendingSequenceCapture[TTarget, TEvent]{}
	for _, c := range pending.captures {
		pc := popCapture(c)
		if pc == nil {
			continue
		}
		next = append(next, *pc)
	}
	if len(next) == 0 {
		k.setPendingSequenceLocked(nil)
		return true
	}
	k.setPendingSequenceLocked(&pendingSequenceState[TTarget, TEvent]{captures: next})
	return true
}

func popCapture[TTarget any, TEvent KeymapEvent](c pendingSequenceCapture[TTarget, TEvent]) *pendingSequenceCapture[TTarget, TEvent] {
	if len(c.parts) == 0 || len(c.parts) <= 1 {
		return nil
	}
	index := c.index - 1
	return &pendingSequenceCapture[TTarget, TEvent]{layer: c.layer, binding: c.binding, index: index, parts: c.parts[:len(c.parts)-1], patterns: c.patterns}
}

func captureHasContinuations[TTarget any, TEvent KeymapEvent](b *bindingState[TTarget, TEvent], index int, patterns map[string]resolvedSequencePattern[TEvent]) bool {
	if index >= len(b.sequence) {
		return false
	}
	return index+1 < len(b.sequence)
}

func (k *Keymap[TTarget, TEvent]) collectSequencePartsFromPending(pending *pendingSequenceState[TTarget, TEvent]) []KeySequencePart {
	if pending == nil || len(pending.captures) == 0 {
		return []KeySequencePart{}
	}
	first := pending.captures[0]
	if len(first.parts) == 0 {
		return []KeySequencePart{}
	}
	out := []KeySequencePart{}
	for i := 0; i < len(first.parts); i++ {
		display := ""
		tokenName := ""
		hasDisplayConflict := false
		hasTokenConflict := false
		for _, c := range pending.captures {
			if i >= len(c.parts) {
				continue
			}
			part := c.parts[i]
			if display == "" {
				display = part.Display
				tokenName = part.TokenName
			}
			if !hasDisplayConflict && display != part.Display {
				hasDisplayConflict = true
			}
			if !hasTokenConflict && tokenName != part.TokenName {
				hasTokenConflict = true
			}
		}
		if display == "" || hasDisplayConflict {
			display = first.parts[i].Stroke.String()
		}
		if hasTokenConflict {
			tokenName = ""
		}
		out = append(out, KeySequencePart{
			Stroke:    first.parts[i].Stroke,
			Display:   display,
			Match:     first.parts[i].Match,
			TokenName: tokenName,
		})
	}
	return out
}

func pendingSequenceEqual[TTarget any, TEvent KeymapEvent](a, b *pendingSequenceState[TTarget, TEvent]) bool {
	if len(a.captures) != len(b.captures) {
		return false
	}
	for i := range a.captures {
		if a.captures[i].layer != b.captures[i].layer || a.captures[i].binding != b.captures[i].binding || a.captures[i].index != b.captures[i].index {
			return false
		}
		if len(a.captures[i].parts) != len(b.captures[i].parts) {
			return false
		}
		for j := range a.captures[i].parts {
			if a.captures[i].parts[j].Match != b.captures[i].parts[j].Match {
				return false
			}
		}
	}
	return true
}

// ---------------------------------------------------------------------------
// Sequence tree
// ---------------------------------------------------------------------------

func buildSequenceTree[TTarget any, TEvent KeymapEvent](bindings []bindingState[TTarget, TEvent], patterns map[string]resolvedSequencePattern[TEvent]) *sequenceNode[TTarget, TEvent] {
	root := &sequenceNode[TTarget, TEvent]{children: map[string]*sequenceNode[TTarget, TEvent]{}}
	for _, b := range bindings {
		if b.event != BindingEventPress {
			continue
		}
		node := root
		for _, part := range b.sequence {
			var child *sequenceNode[TTarget, TEvent]
			if part.PatternName != "" {
				found := false
				for _, pc := range node.patternChildren {
					if pc.patternName == part.PatternName {
						child = pc
						found = true
						break
					}
				}
				if !found {
					child = &sequenceNode[TTarget, TEvent]{parent: node, depth: node.depth + 1, stroke: &part.Stroke, match: part.Match, patternName: part.PatternName, children: map[string]*sequenceNode[TTarget, TEvent]{}}
					node.patternChildren = append(node.patternChildren, child)
				}
			} else {
				var ok bool
				child, ok = node.children[part.Match]
				if !ok {
					child = &sequenceNode[TTarget, TEvent]{parent: node, depth: node.depth + 1, stroke: &part.Stroke, match: part.Match, children: map[string]*sequenceNode[TTarget, TEvent]{}}
					node.children[part.Match] = child
				}
			}
			child.reachableBindings = append(child.reachableBindings, b)
			node = child
		}
		node.bindings = append(node.bindings, b)
	}
	return root
}

// ---------------------------------------------------------------------------
// Dispatch
// ---------------------------------------------------------------------------

// RunCommandOptions are options for running a command.
type RunCommandOptions[TTarget any, TEvent KeymapEvent] struct {
	Event          TEvent
	Focused        TTarget
	Target         TTarget
	IncludeCommand bool
	Payload        any
}

// DispatchKey feeds a key event into the keymap, running the dispatch pipeline.
// release indicates whether this is a key release event.
func (k *Keymap[TTarget, TEvent]) DispatchKey(event TEvent, release bool) {
	k.handleKeyEvent(event, release)
}

func (k *Keymap[TTarget, TEvent]) handleKeyEvent(event TEvent, release bool) {
	afterHooks := k.getKeyAfterHooks(release)
	var afterFocused TTarget
	if afterHooks != nil {
		afterFocused = k.getFocusedTarget()
	}
	ctx := keyInputContext[TTarget, TEvent]{
		Event:   event,
		SetData: func(name string, value any) { k.SetData(name, value) },
		GetData: func(name string) any { return k.GetData(name) },
		Consume: func(opts ...func(consumeOptions)) {
			preventDefault := true
			stopPropagation := true
			if len(opts) > 0 && opts[0] != nil {
				o := consumeOptions{}
				opts[0](o)
				preventDefault = o.PreventDefault
				stopPropagation = o.StopPropagation
			}
			if preventDefault {
				event.PreventDefault()
			}
			if stopPropagation {
				event.StopPropagation()
			}
		},
	}
	for _, h := range k.state.keyHooks {
		if h.Config.Release != release {
			continue
		}
		h.Listener(ctx)
		if event.PropagationStopped() {
			if afterHooks != nil {
				k.emitKeyAfter(afterHooks, event, release, afterFocused, &keyDispatchOutcome{handled: true, reason: "intercept-consumed"})
			}
			return
		}
	}
	if release {
		outcome := k.dispatchReleaseLayers(event)
		if afterHooks != nil {
			k.emitKeyAfter(afterHooks, event, release, afterFocused, outcome)
		}
		return
	}
	outcome := k.dispatchLayers(event)
	if afterHooks != nil {
		k.emitKeyAfter(afterHooks, event, release, afterFocused, outcome)
	}
}

type keyDispatchOutcome struct {
	handled  bool
	reason   string
	sequence []KeySequencePart
}

func (k *Keymap[TTarget, TEvent]) getKeyAfterHooks(release bool) []func(keyAfterInputContext[TTarget, TEvent]) {
	out := []func(keyAfterInputContext[TTarget, TEvent]){}
	for _, h := range k.state.keyAfterHooks {
		if h.Config.Release == release {
			out = append(out, h.Listener)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func (k *Keymap[TTarget, TEvent]) emitKeyAfter(hooks []func(keyAfterInputContext[TTarget, TEvent]), event TEvent, release bool, focused TTarget, outcome *keyDispatchOutcome) {
	seq := []KeySequencePart{}
	if outcome != nil {
		seq = outcome.sequence
	}
	eventType := BindingEventPress
	if release {
		eventType = BindingEventRelease
	}
	ctx := keyAfterInputContext[TTarget, TEvent]{
		Event:     event,
		EventType: eventType,
		Focused:   focused,
		Handled:   outcome.handled,
		Reason:    outcome.reason,
		Sequence:  seq,
		Pending:   k.getPendingSequence(),
		SetData:   func(name string, value any) { k.SetData(name, value) },
		GetData:   func(name string) any { return k.GetData(name) },
		Consume: func(opts ...func(consumeOptions)) {
			preventDefault := true
			stopPropagation := true
			if len(opts) > 0 && opts[0] != nil {
				o := consumeOptions{}
				opts[0](o)
				preventDefault = o.PreventDefault
				stopPropagation = o.StopPropagation
			}
			if preventDefault {
				event.PreventDefault()
			}
			if stopPropagation {
				event.StopPropagation()
			}
		},
	}
	for _, h := range hooks {
		h(ctx)
	}
}

func (k *Keymap[TTarget, TEvent]) dispatchReleaseLayers(event TEvent) *keyDispatchOutcome {
	focused := k.getFocusedTarget()
	layers := k.getActiveLayers(focused)
	matchKeys := k.resolveEventMatchKeys(event)
	outcome := &keyDispatchOutcome{handled: false, reason: "no-match"}
Outer:
	for _, mk := range matchKeys {
		for _, layer := range layers {
			if len(layer.bindings) == 0 {
				continue
			}
			if !k.matchesConditionsLayer(layer) {
				continue
			}
			handled, stop := k.runReleaseBindings(layer, mk, event, focused)
			if handled {
				outcome.handled = true
			}
			if stop {
				break Outer
			}
		}
	}
	return outcome
}

func (k *Keymap[TTarget, TEvent]) runReleaseBindings(layer *registeredLayer[TTarget, TEvent], strokeKey string, event TEvent, focused TTarget) (handled bool, stop bool) {
	for _, b := range layer.bindings {
		if b.event != BindingEventRelease {
			continue
		}
		if len(b.sequence) == 0 || b.sequence[0].Match != strokeKey {
			continue
		}
		if !k.matchesBindingConditions(&b, focused) {
			continue
		}
		bh := k.runBinding(layer, &b, event, focused, nil)
		if bh {
			handled = true
			if !b.Fallthrough {
				return true, true
			}
		}
	}
	return handled, false
}

func (k *Keymap[TTarget, TEvent]) dispatchLayers(event TEvent) *keyDispatchOutcome {
	focused := k.getFocusedTarget()
	pending := k.ensureValidPendingSequenceLockedReturn()
	matchKeys := k.resolveEventMatchKeys(event)
	if pending != nil {
		return k.dispatchPendingSequence(pending, matchKeys, event, focused)
	}
	layers := k.getActiveLayers(focused)
	outcome := &keyDispatchOutcome{handled: false, reason: "no-match"}
	for _, mk := range matchKeys {
		res := k.dispatchFromRoot(layers, []string{mk}, event, focused)
		if res.handled {
			outcome.handled = true
			return outcome
		}
	}
	return outcome
}

func (k *Keymap[TTarget, TEvent]) dispatchPendingSequence(pending *pendingSequenceState[TTarget, TEvent], matchKeys []string, event TEvent, focused TTarget) *keyDispatchOutcome {
	advanced := []pendingSequenceCapture[TTarget, TEvent]{}
	for _, c := range pending.captures {
		adv := advanceSequenceCapture(c, matchKeys, event, k.state.patterns)
		if adv == nil {
			continue
		}
		advanced = append(advanced, *adv)
	}
	if len(advanced) == 0 {
		outcome := &keyDispatchOutcome{handled: false, reason: "sequence-miss"}
		k.setPendingSequenceLocked(nil)
		return outcome
	}
	return k.dispatchPendingCaptures(advanced, false, event, focused)
}

func (k *Keymap[TTarget, TEvent]) dispatchPendingCaptures(captures []pendingSequenceCapture[TTarget, TEvent], handledExact bool, event TEvent, focused TTarget) *keyDispatchOutcome {
	outcome := &keyDispatchOutcome{handled: false, reason: "no-match"}
	for i := 0; i < len(captures); i++ {
		c := captures[i]
		cont := k.collectContinuationCaptures(captures, i, c)
		if len(cont) > 0 {
			if handledExact {
				continue
			}
			exact := k.collectExactCaptures(captures, c)
			resolved := k.tryResolvePendingAmbiguity(captures, i, cont, exact, event, focused, handledExact)
			if resolved != nil {
				return resolved
			}
			k.setPendingSequenceLocked(&pendingSequenceState[TTarget, TEvent]{captures: cont})
			return &keyDispatchOutcome{handled: false, reason: "sequence-pending"}
		}
		if !captureIsExact(c, k.state.patterns) {
			continue
		}
		exact := k.collectExactCaptures(captures, c)
		if len(exact) == 0 {
			continue
		}
		res := k.runCaptureBindings(exact[0].layer, exact, event, focused, nil)
		if res.handled {
			outcome.handled = true
			if res.stop {
				k.setPendingSequenceLocked(nil)
				return outcome
			}
		}
	}
	k.setPendingSequenceLocked(nil)
	return outcome
}

func (k *Keymap[TTarget, TEvent]) collectContinuationCaptures(captures []pendingSequenceCapture[TTarget, TEvent], startIndex int, prefix pendingSequenceCapture[TTarget, TEvent]) []pendingSequenceCapture[TTarget, TEvent] {
	out := []pendingSequenceCapture[TTarget, TEvent]{}
	for i, cand := range captures {
		if i < startIndex {
			continue
		}
		if !captureHasContinuations(cand.binding, cand.index, k.state.patterns) {
			continue
		}
		if sameParts(cand.parts, prefix.parts) {
			out = append(out, cand)
		}
	}
	return out
}

func (k *Keymap[TTarget, TEvent]) collectExactCaptures(captures []pendingSequenceCapture[TTarget, TEvent], prefix pendingSequenceCapture[TTarget, TEvent]) []pendingSequenceCapture[TTarget, TEvent] {
	out := []pendingSequenceCapture[TTarget, TEvent]{}
	for _, c := range captures {
		if c.layer != prefix.layer {
			continue
		}
		if !captureIsExact(c, k.state.patterns) {
			continue
		}
		if !sameParts(c.parts, prefix.parts) {
			continue
		}
		out = append(out, c)
	}
	return out
}

func (k *Keymap[TTarget, TEvent]) tryResolvePendingAmbiguity(advanced []pendingSequenceCapture[TTarget, TEvent], index int, continuation, exact []pendingSequenceCapture[TTarget, TEvent], event TEvent, focused TTarget, handledExact bool) *keyDispatchOutcome {
	if len(k.state.disambiguationResolvers) == 0 || len(continuation) == 0 {
		return nil
	}
	exactBindings := []*bindingState[TTarget, TEvent]{}
	for _, c := range exact {
		exactBindings = append(exactBindings, c.binding)
	}
	hasCmd := false
	for _, b := range exactBindings {
		if b.command.Kind != BindingCommandNone || b.run != nil {
			hasCmd = true
		}
	}
	if !hasCmd {
		return nil
	}
	seq := k.collectSequencePartsFromPending(&pendingSequenceState[TTarget, TEvent]{captures: continuation})
	ctx := k.buildDisambiguationContext(event, focused, seq, exactBindings, continuation)
	decision := k.resolveDisambiguation(ctx)
	if decision == nil {
		return &keyDispatchOutcome{handled: false, reason: "sequence-pending"}
	}
	switch decision.Action {
	case "run-exact":
		if len(exact) > 0 {
			res := k.runCaptureBindings(exact[0].layer, exact, event, focused, nil)
			return &keyDispatchOutcome{handled: res.handled, reason: "binding-handled", sequence: seq}
		}
	case "continue-sequence":
		return &keyDispatchOutcome{handled: false, reason: "sequence-pending"}
	case "clear":
		k.setPendingSequenceLocked(nil)
		return &keyDispatchOutcome{handled: false, reason: "sequence-cleared"}
	}
	return &keyDispatchOutcome{handled: false, reason: "sequence-pending"}
}

func (k *Keymap[TTarget, TEvent]) dispatchFromRoot(layers []*registeredLayer[TTarget, TEvent], matchKeys []string, event TEvent, focused TTarget) *keyDispatchOutcome {
	outcome := &keyDispatchOutcome{handled: false, reason: "no-match"}
	for i := 0; i < len(layers); i++ {
		layer := layers[i]
		if layer == nil || !k.matchesConditionsLayer(layer) {
			continue
		}
		captures := k.collectRootCaptures(layer, matchKeys, event, focused)
		if len(captures) == 0 {
			continue
		}
		continuation := []pendingSequenceCapture[TTarget, TEvent]{}
		for _, c := range captures {
			if captureHasContinuations(c.binding, c.index, k.state.patterns) {
				continuation = append(continuation, c)
			}
		}
		if len(continuation) > 0 {
			exact := []pendingSequenceCapture[TTarget, TEvent]{}
			for _, c := range captures {
				if captureIsExact(c, k.state.patterns) {
					exact = append(exact, c)
				}
			}
			resolved := k.tryResolveRootAmbiguity(layers, i, matchKeys, continuation, exact, event, focused)
			if resolved != nil {
				return resolved
			}
			k.setPendingSequenceLocked(&pendingSequenceState[TTarget, TEvent]{captures: continuation})
			return &keyDispatchOutcome{handled: false, reason: "sequence-pending"}
		}
		exact := []pendingSequenceCapture[TTarget, TEvent]{}
		for _, c := range captures {
			if captureIsExact(c, k.state.patterns) {
				exact = append(exact, c)
			}
		}
		res := k.runCaptureBindings(layer, exact, event, focused, nil)
		if res.handled {
			outcome.handled = true
			if res.stop {
				return outcome
			}
		}
	}
	return outcome
}

func (k *Keymap[TTarget, TEvent]) tryResolveRootAmbiguity(layers []*registeredLayer[TTarget, TEvent], layerIndex int, matchKeys []string, continuation, exact []pendingSequenceCapture[TTarget, TEvent], event TEvent, focused TTarget) *keyDispatchOutcome {
	if len(k.state.disambiguationResolvers) == 0 || len(continuation) == 0 {
		return nil
	}
	exactBindings := []*bindingState[TTarget, TEvent]{}
	for _, c := range exact {
		exactBindings = append(exactBindings, c.binding)
	}
	seq := k.collectSequencePartsFromPending(&pendingSequenceState[TTarget, TEvent]{captures: continuation})
	ctx := k.buildDisambiguationContext(event, focused, seq, exactBindings, continuation)
	decision := k.resolveDisambiguation(ctx)
	if decision == nil {
		return &keyDispatchOutcome{handled: false, reason: "sequence-pending"}
	}
	switch decision.Action {
	case "run-exact":
		if len(exact) > 0 {
			res := k.runCaptureBindings(exact[0].layer, exact, event, focused, nil)
			return &keyDispatchOutcome{handled: res.handled, reason: "binding-handled", sequence: seq}
		}
	case "continue-sequence":
		return &keyDispatchOutcome{handled: false, reason: "sequence-pending"}
	case "clear":
		k.setPendingSequenceLocked(nil)
		return &keyDispatchOutcome{handled: false, reason: "sequence-cleared"}
	}
	return &keyDispatchOutcome{handled: false, reason: "sequence-pending"}
}

type bindingRunResult struct {
	handled bool
	stop    bool
}

func (k *Keymap[TTarget, TEvent]) runCaptureBindings(layer *registeredLayer[TTarget, TEvent], captures []pendingSequenceCapture[TTarget, TEvent], event TEvent, focused TTarget, payload any) bindingRunResult {
	handled := false
	for _, c := range captures {
		b := c.binding
		if !k.matchesBindingConditions(b, focused) {
			continue
		}
		bh := k.runBinding(layer, b, event, focused, payload)
		if bh {
			handled = true
			if !b.Fallthrough {
				return bindingRunResult{handled: true, stop: true}
			}
		}
	}
	return bindingRunResult{handled: handled, stop: false}
}

func (k *Keymap[TTarget, TEvent]) collectRootCaptures(layer *registeredLayer[TTarget, TEvent], matchKeys []string, event TEvent, focused TTarget) []pendingSequenceCapture[TTarget, TEvent] {
	out := []pendingSequenceCapture[TTarget, TEvent]{}
	best := -1
	for _, b := range layer.bindings {
		if b.event != BindingEventPress {
			continue
		}
		c := advanceSequenceBinding(&b, 0, []KeySequencePart{}, matchKeys, event, k.state.patterns)
		if c == nil {
			continue
		}
		c.layer = layer
		prio := capturePriority(c, matchKeys)
		if best == -1 || prio < best {
			best = prio
			out = []pendingSequenceCapture[TTarget, TEvent]{}
		}
		if prio == best {
			out = append(out, *c)
		}
	}
	return out
}

func capturePriority[TTarget any, TEvent KeymapEvent](c *pendingSequenceCapture[TTarget, TEvent], matchKeys []string) int {
	if len(c.parts) == 0 {
		return len(matchKeys)
	}
	part := c.parts[len(c.parts)-1]
	if part.PatternName != "" {
		return len(matchKeys)
	}
	for i, mk := range matchKeys {
		if mk == part.Match {
			return i
		}
	}
	return len(matchKeys)
}

func advanceSequenceBinding[TTarget any, TEvent KeymapEvent](b *bindingState[TTarget, TEvent], index int, parts []KeySequencePart, matchKeys []string, event TEvent, patterns map[string]resolvedSequencePattern[TEvent]) *pendingSequenceCapture[TTarget, TEvent] {
	if index >= len(b.sequence) {
		return nil
	}
	part := b.sequence[index]
	if part.PatternName != "" {
		p := patterns[part.PatternName]
		if p.matcher == nil {
			return nil
		}
		m := p.matcher(event)
		if m == nil {
			return nil
		}
		eventPart := KeySequencePart{
			Stroke: normalizeEventKeyStroke(event),
			Display: func() string {
				if m.display != "" {
					return m.display
				}
				return fmt.Sprintf("%v", m.value)
			}(),
			Match:       part.Match,
			PatternName: part.PatternName,
			PayloadKey:  part.PayloadKey,
		}
		newParts := append([]KeySequencePart{}, parts...)
		newParts = append(newParts, eventPart)
		return &pendingSequenceCapture[TTarget, TEvent]{binding: b, index: index, parts: newParts}
	}
	found := false
	for _, mk := range matchKeys {
		if mk == part.Match {
			found = true
			break
		}
	}
	if !found {
		return nil
	}
	newParts := append([]KeySequencePart{}, parts...)
	newParts = append(newParts, part)
	return &pendingSequenceCapture[TTarget, TEvent]{binding: b, index: index, parts: newParts}
}

func advanceSequenceCapture[TTarget any, TEvent KeymapEvent](c pendingSequenceCapture[TTarget, TEvent], matchKeys []string, event TEvent, patterns map[string]resolvedSequencePattern[TEvent]) *pendingSequenceCapture[TTarget, TEvent] {
	currentPart := c.binding.sequence[c.index]
	if currentPart.PatternName != "" {
		p := patterns[currentPart.PatternName]
		if p.matcher != nil {
			m := p.matcher(event)
			if m != nil {
				eventPart := KeySequencePart{
					Stroke: normalizeEventKeyStroke(event),
					Display: func() string {
						if m.display != "" {
							return m.display
						}
						return fmt.Sprintf("%v", m.value)
					}(),
					Match:       currentPart.Match,
					PatternName: currentPart.PatternName,
					PayloadKey:  currentPart.PayloadKey,
				}
				newParts := append([]KeySequencePart{}, c.parts...)
				newParts = append(newParts, eventPart)
				return &pendingSequenceCapture[TTarget, TEvent]{layer: c.layer, binding: c.binding, index: c.index, parts: newParts, patterns: c.patterns}
			}
		}
		if !captureHasMinimum(c, patterns) {
			return nil
		}
		return advanceSequenceBinding(c.binding, c.index+1, c.parts, matchKeys, event, patterns)
	}
	return advanceSequenceBinding(c.binding, c.index+1, c.parts, matchKeys, event, patterns)
}

func captureIsExact[TTarget any, TEvent KeymapEvent](c pendingSequenceCapture[TTarget, TEvent], patterns map[string]resolvedSequencePattern[TEvent]) bool {
	return c.index == len(c.binding.sequence)-1 && captureHasMinimum(c, patterns)
}

func sameParts(a, b []KeySequencePart) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Match != b[i].Match {
			return false
		}
	}
	return true
}

func (k *Keymap[TTarget, TEvent]) runBinding(layer *registeredLayer[TTarget, TEvent], b *bindingState[TTarget, TEvent], event TEvent, focused TTarget, payload any) bool {
	if b.run != nil {
		result := k.executeCommandHandler(b.run, event, focused, layer.target, payload)
		if result == "rejected" {
			return false
		}
		k.applyBindingEventEffects(b, event)
		return true
	}
	if b.command.Kind != BindingCommandName {
		return false
	}
	chain := k.getResolvedCommandChain(b.command.Name, focused)
	for _, entry := range chain {
		result := k.executeCommandHandler(entry.command.Run, event, focused, entry.target, payload)
		if result == "rejected" {
			continue
		}
		k.applyBindingEventEffects(b, event)
		return true
	}
	return false
}

func (k *Keymap[TTarget, TEvent]) executeCommandHandler(run CommandHandler[any, KeymapEvent], event TEvent, focused TTarget, target TTarget, payload any) string {
	if run == nil {
		return "handled"
	}
	ctx := CommandContext[any, KeymapEvent]{
		Keymap:  k,
		Event:   event,
		Focused: focused,
		Target:  target,
		Data:    k.GetReadOnlyData(),
		Command: nil,
		Payload: payload,
	}
	result := run(ctx)
	switch r := result.(type) {
	case *RunCommandResult:
		if !r.OK {
			return "rejected"
		}
	case bool:
		if !r {
			return "rejected"
		}
	case nil:
		// handled
	}
	return "handled"
}

func (k *Keymap[TTarget, TEvent]) applyBindingEventEffects(b *bindingState[TTarget, TEvent], event TEvent) {
	if !b.preventDefault {
		return
	}
	event.PreventDefault()
	event.StopPropagation()
}

// ---------------------------------------------------------------------------
// Command catalog
// ---------------------------------------------------------------------------

type layerCommandEntry[TTarget any, TEvent KeymapEvent] struct {
	layer        *registeredLayer[TTarget, TEvent]
	commandState commandState[TTarget, TEvent]
}

type commandView[TTarget any, TEvent KeymapEvent] struct {
	entries      []layerCommandEntry[TTarget, TEvent]
	chainsByName map[string][]layerCommandEntry[TTarget, TEvent]
}

type activeCommandView[TTarget any, TEvent KeymapEvent] struct {
	layers          []*registeredLayer[TTarget, TEvent]
	entries         []layerCommandEntry[TTarget, TEvent]
	reachable       []layerCommandEntry[TTarget, TEvent]
	reachableByName map[string]layerCommandEntry[TTarget, TEvent]
	chainsByName    map[string][]layerCommandEntry[TTarget, TEvent]
}

func (k *Keymap[TTarget, TEvent]) getCommandView() commandView[TTarget, TEvent] {
	entries := []layerCommandEntry[TTarget, TEvent]{}
	chains := map[string][]layerCommandEntry[TTarget, TEvent]{}
	for _, layer := range k.state.sortedLayers {
		for _, cs := range layer.commands {
			e := layerCommandEntry[TTarget, TEvent]{layer: layer, commandState: cs}
			entries = append(entries, e)
			chains[cs.command.Name] = append(chains[cs.command.Name], e)
		}
	}
	return commandView[TTarget, TEvent]{entries: entries, chainsByName: chains}
}

func (k *Keymap[TTarget, TEvent]) getActiveCommandView(focused TTarget) activeCommandView[TTarget, TEvent] {
	layers := k.getActiveLayers(focused)
	entries := []layerCommandEntry[TTarget, TEvent]{}
	reachable := []layerCommandEntry[TTarget, TEvent]{}
	reachableByName := map[string]layerCommandEntry[TTarget, TEvent]{}
	chains := map[string][]layerCommandEntry[TTarget, TEvent]{}
	for _, layer := range layers {
		if len(layer.commands) == 0 || !k.matchesConditionsLayer(layer) {
			continue
		}
		for _, cs := range layer.commands {
			e := layerCommandEntry[TTarget, TEvent]{layer: layer, commandState: cs}
			entries = append(entries, e)
			chains[cs.command.Name] = append(chains[cs.command.Name], e)
			if _, ok := reachableByName[cs.command.Name]; !ok {
				reachableByName[cs.command.Name] = e
				reachable = append(reachable, e)
			}
		}
	}
	return activeCommandView[TTarget, TEvent]{layers: layers, entries: entries, reachable: reachable, reachableByName: reachableByName, chainsByName: chains}
}

func (k *Keymap[TTarget, TEvent]) getFilteredCommandEntriesLocked(query *CommandQuery[TTarget, TEvent]) []layerCommandEntry[TTarget, TEvent] {
	visibility := "reachable"
	if query != nil && query.Visibility != "" {
		visibility = query.Visibility
	}
	var entries []layerCommandEntry[TTarget, TEvent]
	switch visibility {
	case "registered":
		view := k.getCommandView()
		entries = make([]layerCommandEntry[TTarget, TEvent], len(view.entries))
		for i, e := range view.entries {
			entries[i] = layerCommandEntry[TTarget, TEvent]{layer: e.layer, commandState: commandState[TTarget, TEvent]{command: e.commandState.command}}
		}
	case "active":
		focused := query.Focused
		if any(focused) == nil {
			focused = k.getFocusedTarget()
		}
		entries = k.getActiveCommandView(focused).entries
	default:
		focused := query.Focused
		if any(focused) == nil {
			focused = k.getFocusedTarget()
		}
		entries = k.getActiveCommandView(focused).reachable
	}
	if query == nil {
		return entries
	}
	out := []layerCommandEntry[TTarget, TEvent]{}
	for _, e := range entries {
		if query.Namespace != "" {
			if ns, ok := e.commandState.fields["namespace"]; ok {
				if fmt.Sprintf("%v", ns) != query.Namespace {
					continue
				}
			} else {
				continue
			}
		}
		if query.Search != "" {
			if !k.commandMatchesSearch(e.commandState, query.Search) {
				continue
			}
		}
		if query.Filter != nil && !query.Filter(e.commandState.command) {
			continue
		}
		out = append(out, e)
		if query.Limit > 0 && len(out) >= query.Limit {
			break
		}
	}
	return out
}

func (k *Keymap[TTarget, TEvent]) commandMatchesSearch(cs commandState[TTarget, TEvent], search string) bool {
	low := strings.ToLower(search)
	if strings.Contains(strings.ToLower(cs.command.Name), low) {
		return true
	}
	for _, v := range cs.fields {
		if strings.Contains(strings.ToLower(fmt.Sprintf("%v", v)), low) {
			return true
		}
	}
	return false
}

func (k *Keymap[TTarget, TEvent]) getCommandBindingsLocked(query CommandBindingsQuery[TTarget]) map[string][]ActiveBinding[TTarget, TEvent] {
	out := map[string][]ActiveBinding[TTarget, TEvent]{}
	for _, cmd := range query.Commands {
		out[cmd] = []ActiveBinding[TTarget, TEvent]{}
	}
	if len(out) == 0 {
		return out
	}
	focused := query.Focused
	if any(focused) == nil {
		focused = k.getFocusedTarget()
	}
	view := k.getActiveCommandView(focused)
	for _, layer := range view.layers {
		if len(layer.bindings) == 0 || !k.matchesConditionsLayer(layer) {
			continue
		}
		for _, b := range layer.bindings {
			if b.command.Kind != BindingCommandName {
				continue
			}
			if _, ok := out[b.command.Name]; !ok {
				continue
			}
			if k.matchesBindingConditions(&b, focused) {
				out[b.command.Name] = append(out[b.command.Name], k.createActiveBinding(&b, k.bindingCommandAttrs(&b, focused)))
			}
		}
	}
	return out
}

func (k *Keymap[TTarget, TEvent]) createActiveBinding(b *bindingState[TTarget, TEvent], attrs map[string]any) ActiveBinding[TTarget, TEvent] {
	return ActiveBinding[TTarget, TEvent]{
		Sequence:       cloneKeySequence(b.sequence),
		Command:        b.command,
		CommandAttrs:   attrs,
		Attrs:          b.attrs,
		Event:          b.event,
		PreventDefault: b.preventDefault,
		Fallthrough:    b.Fallthrough,
	}
}

func (k *Keymap[TTarget, TEvent]) getResolvedCommandChain(command string, focused TTarget) []struct {
	target  TTarget
	command *Command
} {
	view := k.getActiveCommandView(focused)
	out := []struct {
		target  TTarget
		command *Command
	}{}
	for _, e := range view.chainsByName[command] {
		out = append(out, struct {
			target  TTarget
			command *Command
		}{e.layer.target, e.commandState.command})
	}
	return out
}

func (k *Keymap[TTarget, TEvent]) getFallbackResolvedCommand(command string, focused TTarget) *layerCommandEntry[TTarget, TEvent] {
	return nil
}

func (k *Keymap[TTarget, TEvent]) runCommandLocked(cmd string, options *RunCommandOptions[TTarget, TEvent]) *RunCommandResult {
	if options == nil {
		options = &RunCommandOptions[TTarget, TEvent]{}
	}
	focused := options.Focused
	if any(focused) == nil {
		focused = k.getFocusedTarget()
	}
	chain := k.getResolvedCommandChain(cmd, focused)
	for _, entry := range chain {
		result := k.executeCommandHandler(entry.command.Run, options.Event, focused, entry.target, options.Payload)
		if result == "rejected" {
			continue
		}
		return &RunCommandResult{OK: true, Command: entry.command}
	}
	return &RunCommandResult{OK: false, Reason: "not-found"}
}

func (k *Keymap[TTarget, TEvent]) dispatchCommandLocked(cmd string, options *RunCommandOptions[TTarget, TEvent]) *RunCommandResult {
	return k.runCommandLocked(cmd, options)
}

// ---------------------------------------------------------------------------
// Event matching
// ---------------------------------------------------------------------------

func (k *Keymap[TTarget, TEvent]) resolveEventMatchKeys(event TEvent) []string {
	if len(k.state.eventMatchResolvers) == 0 {
		stk := normalizeEventKeyStroke(event)
		return []string{stk.MatchID()}
	}
	out := []string{}
	seen := map[string]bool{}
	for _, r := range k.state.eventMatchResolvers {
		resolved := r(event, eventMatchResolverContext{ResolveKey: func(in interface{}) string {
			switch v := in.(type) {
			case KeyStrokeInput:
				return normalizeKeyStroke(v).MatchID()
			case KeyStroke:
				return v.MatchID()
			}
			return ""
		}})
		for _, c := range resolved {
			if !seen[c] {
				seen[c] = true
				out = append(out, c)
			}
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// Disambiguation
// ---------------------------------------------------------------------------

func (k *Keymap[TTarget, TEvent]) buildDisambiguationContext(event TEvent, focused TTarget, sequence []KeySequencePart, exactBindings []*bindingState[TTarget, TEvent], continuation []pendingSequenceCapture[TTarget, TEvent]) disambiguationContext[TTarget, TEvent] {
	exactActives := []ActiveBinding[TTarget, TEvent]{}
	for _, b := range exactBindings {
		exactActives = append(exactActives, k.createActiveBinding(b, k.bindingCommandAttrs(b, focused)))
	}
	continues := []ActiveKey[TTarget, TEvent]{}
	for _, c := range continuation {
		if c.index+1 < len(c.binding.sequence) {
			part := c.binding.sequence[c.index+1]
			continues = append(continues, ActiveKey[TTarget, TEvent]{Stroke: part.Stroke, Display: part.Display, Continues: true})
		}
	}
	ctx := disambiguationContext[TTarget, TEvent]{
		Event:         event,
		Focused:       focused,
		Sequence:      sequence,
		Exact:         exactActives,
		Continuations: continues,
	}
	ctx.runExact = func() *disambiguationDecision[TTarget, TEvent] {
		return &disambiguationDecision[TTarget, TEvent]{Action: "run-exact"}
	}
	ctx.continueSeq = func() *disambiguationDecision[TTarget, TEvent] {
		return &disambiguationDecision[TTarget, TEvent]{Action: "continue-sequence"}
	}
	ctx.clear = func() *disambiguationDecision[TTarget, TEvent] {
		return &disambiguationDecision[TTarget, TEvent]{Action: "clear"}
	}
	ctx.deferFn = func(h func(deferredDisambiguationDecision) *deferredDisambiguationDecision) *disambiguationDecision[TTarget, TEvent] {
		return &disambiguationDecision[TTarget, TEvent]{Action: "defer", Handler: h}
	}
	return ctx
}

func (k *Keymap[TTarget, TEvent]) resolveDisambiguation(ctx disambiguationContext[TTarget, TEvent]) *disambiguationDecision[TTarget, TEvent] {
	for _, r := range k.state.disambiguationResolvers {
		result := r(ctx)
		if result == nil {
			continue
		}
		return result
	}
	return nil
}

// ---------------------------------------------------------------------------
// Raw input
// ---------------------------------------------------------------------------

func (k *Keymap[TTarget, TEvent]) handleRawSequence(sequence string) bool {
	stopped := false
	ctx := rawInputContext{Sequence: sequence, Stop: func() { stopped = true }}
	for _, h := range k.state.rawHooks {
		h.Listener(ctx)
		if stopped {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Hooks / notify
// ---------------------------------------------------------------------------

func (k *Keymap[TTarget, TEvent]) notifyHook(name HookName, value any) {
	k.listenerMu.RLock()
	lst := k.listeners[name]
	k.listenerMu.RUnlock()
	for _, l := range lst {
		l(value)
	}
}

func (k *Keymap[TTarget, TEvent]) queueStateChange() {
	k.state.mu.Lock()
	k.state.derivedVersion++
	k.state.cacheVersion++
	k.state.mu.Unlock()
	k.notifyHook(HookState, nil)
}

func (k *Keymap[TTarget, TEvent]) emitWarning(code, message string) {
	k.notifyHook("warning", map[string]string{"code": code, "message": message})
}

func (k *Keymap[TTarget, TEvent]) emitError(code string, err error, message string) {
	k.notifyHook("error", map[string]string{"code": code, "message": message, "err": fmt.Sprintf("%v", err)})
}

func (k *Keymap[TTarget, TEvent]) warnOnce(key, code, message string) {
	k.mu.Lock()
	defer k.mu.Unlock()
	if k.warnKeys[key] {
		return
	}
	k.warnKeys[key] = true
	k.emitWarning(code, message)
}

// ---------------------------------------------------------------------------
// Diagnostics
// ---------------------------------------------------------------------------

// Diagnostics returns all active bindings across active layers.
func (k *Keymap[TTarget, TEvent]) Diagnostics() []ActiveBindingInfo[TTarget, TEvent] {
	k.mu.RLock()
	defer k.mu.RUnlock()
	focused := k.getFocusedTarget()
	layers := k.getActiveLayers(focused)
	out := []ActiveBindingInfo[TTarget, TEvent]{}
	for _, layer := range layers {
		for _, b := range layer.bindings {
			if !k.matchesBindingConditions(&b, focused) {
				continue
			}
			cmdName := ""
			if b.command.Kind == BindingCommandName {
				cmdName = b.command.Name
			}
			out = append(out, ActiveBindingInfo[TTarget, TEvent]{
				Sequence:   stringifyKeySequence(b.sequence),
				Command:    cmdName,
				LayerOrder: layer.order,
				Priority:   layer.priority,
				Event:      b.event,
			})
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// Feature packs
// ---------------------------------------------------------------------------

// RegisterFeaturePack registers a feature pack of bindings.
func (k *Keymap[TTarget, TEvent]) RegisterFeaturePack(name string, bindings []Binding[TTarget, TEvent]) *FeaturePack[TTarget, TEvent] {
	return &FeaturePack[TTarget, TEvent]{
		Name:     name,
		Bindings: bindings,
		keymap:   k,
	}
}

// ---------------------------------------------------------------------------
// Parser / compiler primitives
// ---------------------------------------------------------------------------

func parseKeySequenceInput[TTarget any, TEvent KeymapEvent](key interface{}, state *State[TTarget, TEvent]) []KeySequencePart {
	switch v := key.(type) {
	case string:
		return parseBindingSequenceString(v, state)
	case KeyStrokeInput:
		return []KeySequencePart{createKeySequencePart(v)}
	default:
		return []KeySequencePart{}
	}
}

func parseBindingSequenceString[TTarget any, TEvent KeymapEvent](key string, state *State[TTarget, TEvent]) []KeySequencePart {
	if len(state.bindingParsers) == 0 {
		return []KeySequencePart{}
	}
	parts := []KeySequencePart{}
	index := 0
	for index < len(key) {
		matched := false
		for _, p := range state.bindingParsers {
			res := p(bindingParserContext[TEvent]{
				Input:              key,
				Index:              index,
				Layer:              map[string]any{},
				Tokens:             state.tokens,
				Patterns:           state.patterns,
				NormalizeTokenName: func(s string) string { return strings.ToLower(strings.TrimSpace(s)) },
				CreateMatch:        func(id string) string { return id },
				ParseObjectKey: func(in KeyStrokeInput, opts ...func(*KeySequencePart)) KeySequencePart {
					return createKeySequencePart(in, opts...)
				},
			})
			if res != nil {
				if res.NextIndex <= index || res.NextIndex > len(key) {
					break
				}
				parts = append(parts, res.Parts...)
				index = res.NextIndex
				matched = true
				break
			}
		}
		if !matched {
			if key[index] == ' ' || key[index] == '\t' || key[index] == '\n' {
				index++
				continue
			}
			parts = append(parts, createKeySequencePart(KeyStrokeInput{Name: string(key[index])}))
			index++
		}
	}
	return parts
}

func parseSingleKeyPart[TTarget any, TEvent KeymapEvent](key interface{}, state *State[TTarget, TEvent]) KeySequencePart {
	if si, ok := key.(KeyStrokeInput); ok {
		return createKeySequencePart(si)
	}
	parts := parseKeySequenceInput(key, state)
	if len(parts) == 1 {
		return parts[0]
	}
	return KeySequencePart{}
}

func resolveKeyMatchInput(input interface{}) string {
	switch v := input.(type) {
	case KeySequencePart:
		return v.Match
	case KeyStroke:
		return v.MatchID()
	case KeyStrokeInput:
		return normalizeKeyStroke(v).MatchID()
	}
	return ""
}

func (k *Keymap[TTarget, TEvent]) handleFocusedTargetChange(focused TTarget) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.setPendingSequenceLocked(nil)
	k.queueStateChange()
}
