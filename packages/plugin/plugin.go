package plugin

import (
	"fmt"
	"sort"
	"sync"
)

// Plugin is the interface that all PixTui plugins must implement.
type Plugin interface {
	// Name returns the unique plugin name.
	Name() string
	// Version returns the plugin version string.
	Version() string
	// Init initializes the plugin with the given context.
	Init(ctx *Context) error
	// Shutdown cleans up plugin resources.
	Shutdown() error
}

// Descriptor is an optional interface plugins may implement to provide rich
// metadata. When a plugin implements Descriptor, the Manager uses it to populate
// the plugin info returned by PluginInfo and to resolve dependencies.
type Descriptor interface {
	// Descriptor returns the plugin's metadata.
	Descriptor() PluginInfo
}

// PluginInfo describes a plugin, including its name, version, description and
// the plugins it depends on. A zero-value PluginInfo has empty dependencies.
type PluginInfo struct {
	// Name is the unique plugin name.
	Name string
	// Version is the plugin version string.
	Version string
	// Description is a human-readable summary of what the plugin does.
	Description string
	// Author is the plugin author or publisher.
	Author string
	// Homepage is a URL to the plugin's homepage or repository.
	Homepage string
	// Dependencies lists the names of plugins this plugin requires to be
	// registered before it can be activated.
	Dependencies []string
}

// LifecycleState tracks where a plugin is in its lifecycle.
type LifecycleState int

const (
	// StateRegistered means the plugin has been registered but not yet
	// initialized.
	StateRegistered LifecycleState = iota
	// StateInitialized means Init has been called successfully.
	StateInitialized
	// StateActive means the plugin is enabled and active.
	StateActive
	// StateShutdown means Shutdown has been called.
	StateShutdown
	// StateError means registration or initialization failed.
	StateError
)

// String returns the lifecycle state as a human-readable string.
func (s LifecycleState) String() string {
	switch s {
	case StateRegistered:
		return "registered"
	case StateInitialized:
		return "initialized"
	case StateActive:
		return "active"
	case StateShutdown:
		return "shutdown"
	case StateError:
		return "error"
	default:
		return fmt.Sprintf("state(%d)", int(s))
	}
}

// HookPriority orders hook handlers for the same event. Higher priority hooks
// run before lower priority ones.
type HookPriority int

const (
	// HookPriorityHigh runs before normal and low hooks.
	HookPriorityHigh HookPriority = 3
	// HookPriorityNormal is the default priority.
	HookPriorityNormal HookPriority = 2
	// HookPriorityLow runs after normal and high hooks.
	HookPriorityLow HookPriority = 1
)

// HookRegistration records a single event hook with its priority and a unique
// registration ID, which is used to unregister the hook.
type HookRegistration struct {
	// ID uniquely identifies this hook registration within the manager.
	ID uint64
	// Priority is the hook's priority; higher runs first.
	Priority HookPriority
	// Handler is invoked when the event is emitted.
	Handler EventHandler
}

// CommandInfo wraps a registered command handler with optional description.
type CommandInfo struct {
	// Handler handles the plugin command.
	Handler CommandHandler
	// Description is a human-readable description of the command.
	Description string
}

// Context provides plugins with access to PixTui core services.
type Context struct {
	// RegisterRenderable registers a custom renderable type.
	RegisterRenderable func(typeName string, factory RenderableFactory)
	// RegisterCommand registers a command handler.
	RegisterCommand func(name string, handler CommandHandler)
	// RegisterCommandWithDesc registers a command handler with a description.
	RegisterCommandWithDesc func(name, description string, handler CommandHandler)
	// OnEvent registers an event handler.
	OnEvent func(event string, handler EventHandler)
	// RegisterHook registers a prioritized event hook. Hooks run after plain
	// OnEvent handlers, ordered by descending priority.
	RegisterHook func(event string, priority HookPriority, handler EventHandler)
	// EmitEvent emits an event to all registered handlers.
	EmitEvent func(event string, data interface{})
}

// RenderableFactory creates a new renderable from plugin-provided options.
type RenderableFactory func(options map[string]interface{}) (interface{}, error)

// CommandHandler handles a plugin command.
type CommandHandler func(args []string) (string, error)

// EventHandler handles a plugin event.
type EventHandler func(data interface{})

// Manager manages the lifecycle of plugins.
type Manager struct {
	mu      sync.RWMutex
	plugins map[string]Plugin
	infos   map[string]PluginInfo
	states  map[string]LifecycleState
	enabled map[string]bool
	ctx     *Context

	// renderableFactories stores registered renderable factories by type name.
	renderableFactories map[string]RenderableFactory
	// registeredCommands stores registered command handlers by name.
	registeredCommands map[string]CommandInfo
	// eventHandlers stores event handlers by event name.
	eventHandlers map[string][]EventHandler
	// hooks stores prioritized hook registrations by event name.
	hooks map[string][]HookRegistration
	// middleware wraps plugin registration, run in the order added.
	middleware []Middleware

	// hookID is a monotonically increasing counter used to assign unique IDs
	// to hook registrations.
	hookID uint64
}

// Lifecycle events emitted by the Manager.
const (
	// EventPluginRegistered is emitted after a plugin is registered and initialized.
	EventPluginRegistered = "plugin:registered"
	// EventPluginActivated is emitted after a plugin is activated.
	EventPluginActivated = "plugin:activated"
	// EventPluginEnabled is emitted when a plugin is enabled.
	EventPluginEnabled = "plugin:enabled"
	// EventPluginDisabled is emitted when a plugin is disabled.
	EventPluginDisabled = "plugin:disabled"
	// EventPluginShutdown is emitted when a plugin is shut down.
	EventPluginShutdown = "plugin:shutdown"
)

// NewManager creates a new PluginManager.
func NewManager() *Manager {
	m := &Manager{
		plugins:             make(map[string]Plugin),
		infos:               make(map[string]PluginInfo),
		states:              make(map[string]LifecycleState),
		enabled:             make(map[string]bool),
		renderableFactories: make(map[string]RenderableFactory),
		registeredCommands:  make(map[string]CommandInfo),
		eventHandlers:       make(map[string][]EventHandler),
		hooks:               make(map[string][]HookRegistration),
	}
	m.ctx = &Context{
		RegisterRenderable:     m.registerRenderable,
		RegisterCommand:        m.registerCommand,
		RegisterCommandWithDesc: m.registerCommandWithDesc,
		OnEvent:                m.onEvent,
		RegisterHook: func(event string, priority HookPriority, handler EventHandler) {
			m.registerHookLocked(event, priority, handler)
		},
		EmitEvent: m.emitEvent,
	}
	return m
}

// defaultInfo returns a PluginInfo for a plugin, using its Descriptor when
// available and falling back to its Name/Version.
func defaultInfo(p Plugin) PluginInfo {
	info := PluginInfo{
		Name:    p.Name(),
		Version: p.Version(),
	}
	if d, ok := p.(Descriptor); ok {
		if di := d.Descriptor(); di.Name != "" {
			info = di
		}
		if info.Name == "" {
			info.Name = p.Name()
		}
		if info.Version == "" {
			info.Version = p.Version()
		}
	}
	return info
}

// Unregister unregisters and shuts down a plugin.
func (m *Manager) Unregister(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	p, ok := m.plugins[name]
	if !ok {
		return &PluginError{Code: ErrCodeNotFound, Plugin: name, Err: fmt.Errorf("plugin %q not found", name)}
	}

	if err := p.Shutdown(); err != nil {
		return &PluginError{Code: ErrCodeShutdownFailed, Plugin: name, Err: fmt.Errorf("failed to shutdown plugin %q: %w", name, err)}
	}

	m.emitEventLocked(EventPluginShutdown, map[string]interface{}{"plugin": name})

	delete(m.plugins, name)
	delete(m.infos, name)
	delete(m.states, name)
	delete(m.enabled, name)
	return nil
}

// Visit calls f for every registered plugin. If f returns an error, iteration
// stops and that error is returned.
func (m *Manager) Visit(f func(p Plugin) error) error {
	m.mu.RLock()
	names := make([]string, 0, len(m.plugins))
	for name := range m.plugins {
		names = append(names, name)
	}
	m.mu.RUnlock()

	sort.Strings(names)
	for _, name := range names {
		p, ok := m.Get(name)
		if !ok {
			continue
		}
		if err := f(p); err != nil {
			return err
		}
	}
	return nil
}

// Require reports whether all declared dependencies of the named plugin are
// registered. Returns false if the plugin itself is not registered.
func (m *Manager) Require(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	info, ok := m.infos[name]
	if !ok {
		return false
	}
	for _, dep := range info.Dependencies {
		if _, exists := m.plugins[dep]; !exists {
			return false
		}
	}
	return true
}

// DependencyErrors returns the unsatisfied dependencies of the named plugin, if any.
func (m *Manager) DependencyErrors(name string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	info, ok := m.infos[name]
	if !ok {
		return nil
	}
	var missing []string
	for _, dep := range info.Dependencies {
		if _, exists := m.plugins[dep]; !exists {
			missing = append(missing, dep)
		}
	}
	return missing
}

// Shutdown shuts down all registered plugins.
func (m *Manager) Shutdown() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, p := range m.plugins {
		if err := p.Shutdown(); err != nil {
			return &PluginError{Code: ErrCodeShutdownFailed, Plugin: name, Err: fmt.Errorf("failed to shutdown plugin %q: %w", name, err)}
		}
		m.emitEventLocked(EventPluginShutdown, map[string]interface{}{"plugin": name})
		m.states[name] = StateShutdown
		m.enabled[name] = false
	}
	m.plugins = make(map[string]Plugin)
	m.infos = make(map[string]PluginInfo)
	m.states = make(map[string]LifecycleState)
	m.enabled = make(map[string]bool)
	return nil
}

// Get returns a registered plugin by name.
func (m *Manager) Get(name string) (Plugin, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.plugins[name]
	return p, ok
}

// List returns all registered plugin names.
func (m *Manager) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.plugins))
	for name := range m.plugins {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Count returns the number of registered plugins.
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.plugins)
}

// PluginInfo returns the metadata for a registered plugin.
func (m *Manager) PluginInfo(name string) (PluginInfo, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	info, ok := m.infos[name]
	return info, ok
}

// AllPluginInfo returns metadata for every registered plugin.
func (m *Manager) AllPluginInfo() []PluginInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	infos := make([]PluginInfo, 0, len(m.infos))
	for _, info := range m.infos {
		infos = append(infos, info)
	}
	sort.Slice(infos, func(i, j int) bool { return infos[i].Name < infos[j].Name })
	return infos
}

// State returns the lifecycle state of a plugin.
func (m *Manager) State(name string) (LifecycleState, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.states[name]
	return s, ok
}

// IsEnabled reports whether a plugin is currently enabled.
func (m *Manager) IsEnabled(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.enabled[name]
}

// Enable enables a registered plugin. Activating an already-active plugin is a
// no-op.
func (m *Manager) Enable(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.plugins[name]; !ok {
		return fmt.Errorf("plugin %q not found", name)
	}
	if m.enabled[name] {
		return nil
	}
	m.enabled[name] = true
	m.states[name] = StateActive
	m.emitEventLocked("plugin:enabled", map[string]interface{}{"plugin": name})
	return nil
}

// Disable disables a registered plugin without shutting it down. A disabled
// plugin remains registered but is treated as inactive.
func (m *Manager) Disable(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.plugins[name]; !ok {
		return fmt.Errorf("plugin %q not found", name)
	}
	if !m.enabled[name] {
		return nil
	}
	m.enabled[name] = false
	m.states[name] = StateInitialized
	m.emitEventLocked("plugin:disabled", map[string]interface{}{"plugin": name})
	return nil
}

// ResolveOrder returns plugin names in an order that satisfies declared
// dependencies. Plugins without dependencies come first; plugins depending on
// unknown or missing plugins are placed after their available dependencies.
// Cycles are broken by falling back to registration order.
func (m *Manager) ResolveOrder() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := m.listLocked()
	if len(names) <= 1 {
		return names
	}

	index := make(map[string]int, len(names))
	for i, n := range names {
		index[n] = i
	}

	// Dependency graph: name -> dependencies that are registered.
	deps := make(map[string][]string, len(names))
	for _, n := range names {
		info, ok := m.infos[n]
		if !ok {
			continue
		}
		var avail []string
		for _, d := range info.Dependencies {
			if _, exists := m.plugins[d]; exists {
				avail = append(avail, d)
			}
		}
		deps[n] = avail
	}

	// Kahn's topological sort, breaking ties by registration order.
	incoming := make(map[string]int, len(names))
	for n := range deps {
		incoming[n] = len(deps[n])
	}
	// Initialize incoming for nodes with no deps entry.
	for _, n := range names {
		if _, ok := incoming[n]; !ok {
			incoming[n] = 0
		}
	}

	// Use a slice ordered by registration index as the ready queue.
	ready := make([]string, 0, len(names))
	for _, n := range names {
		if incoming[n] == 0 {
			ready = append(ready, n)
		}
	}

	result := make([]string, 0, len(names))
	for len(ready) > 0 {
		// Pick the ready node with the smallest registration index.
		bestPos := 0
		for i := 1; i < len(ready); i++ {
			if index[ready[i]] < index[ready[bestPos]] {
				bestPos = i
			}
		}
		n := ready[bestPos]
		ready = append(ready[:bestPos], ready[bestPos+1:]...)
		result = append(result, n)

		// Reduce incoming for plugins that depend on n.
		for _, dependent := range names {
			found := false
			for _, d := range deps[dependent] {
				if d == n {
					found = true
					break
				}
			}
			if !found {
				continue
			}
			incoming[dependent]--
			if incoming[dependent] == 0 {
				ready = append(ready, dependent)
			}
		}
	}

	// Append any nodes left over due to cycles, in registration order.
	emitted := make(map[string]bool, len(result))
	for _, n := range result {
		emitted[n] = true
	}
	for _, n := range names {
		if !emitted[n] {
			result = append(result, n)
		}
	}
	return result
}

// ---------------------------------------------------------------------------
// Internal registry hooks (used by Context)
// ---------------------------------------------------------------------------

func (m *Manager) registerRenderable(typeName string, factory RenderableFactory) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.renderableFactories[typeName] = factory
}

func (m *Manager) registerCommand(name string, handler CommandHandler) {
	m.registerCommandWithDesc(name, "", handler)
}

func (m *Manager) registerCommandWithDesc(name, description string, handler CommandHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.registeredCommands[name] = CommandInfo{Handler: handler, Description: description}
}

func (m *Manager) onEvent(event string, handler EventHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.eventHandlers[event] = append(m.eventHandlers[event], handler)
}

// RegisterHook registers a prioritized hook handler for an event. Hooks with a
// higher priority run before hooks with a lower priority. Hooks are invoked by
// EmitEvent in addition to plain OnEvent handlers.
//
// The returned function unregisters the hook. It is safe to call after the hook
// has already been removed or after the manager has been shut down.
func (m *Manager) RegisterHook(event string, priority HookPriority, handler EventHandler) func() {
	m.mu.Lock()
	defer m.mu.Unlock()
	id := m.registerHookLocked(event, priority, handler)

	return func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		m.unregisterHookLocked(event, id)
	}
}

// registerHookLocked registers a prioritized hook handler assuming the manager
// lock is already held. It returns the unique ID assigned to the registration.
func (m *Manager) registerHookLocked(event string, priority HookPriority, handler EventHandler) uint64 {
	m.hookID++
	id := m.hookID
	m.hooks[event] = append(m.hooks[event], HookRegistration{ID: id, Priority: priority, Handler: handler})
	sort.SliceStable(m.hooks[event], func(i, j int) bool {
		return m.hooks[event][i].Priority > m.hooks[event][j].Priority
	})
	return id
}

// unregisterHookLocked removes the hook registration with the given ID for an
// event, assuming the manager lock is already held. It is a no-op if no such
// registration exists.
func (m *Manager) unregisterHookLocked(event string, id uint64) {
	registrations := m.hooks[event]
	for i, reg := range registrations {
		if reg.ID == id {
			m.hooks[event] = append(registrations[:i], registrations[i+1:]...)
			if len(m.hooks[event]) == 0 {
				delete(m.hooks, event)
			}
			return
		}
	}
}

// EmitEvent emits an event to all registered plain handlers and prioritized
// hooks for the event.
func (m *Manager) EmitEvent(event string, data interface{}) {
	m.emitEvent(event, data)
}

// GetHooks returns a copy of all hook registrations for an event, ordered by
// descending priority.
func (m *Manager) GetHooks(event string) []HookRegistration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	hooks := make([]HookRegistration, len(m.hooks[event]))
	copy(hooks, m.hooks[event])
	return hooks
}

func (m *Manager) emitEvent(event string, data interface{}) {
	m.mu.RLock()
	handlers := append([]EventHandler{}, m.eventHandlers[event]...)
	hooks := append([]HookRegistration{}, m.hooks[event]...)
	m.mu.RUnlock()

	// Plain handlers run first, then prioritized hooks.
	for _, h := range handlers {
		h(data)
	}
	for _, h := range hooks {
		h.Handler(data)
	}
}

// emitEventLocked emits an event assuming the manager lock is already held. It
// copies the handlers so callbacks cannot deadlock by re-entering the manager.
func (m *Manager) emitEventLocked(event string, data interface{}) {
	handlers := append([]EventHandler{}, m.eventHandlers[event]...)
	hooks := append([]HookRegistration{}, m.hooks[event]...)
	m.mu.Unlock()

	for _, h := range handlers {
		h(data)
	}
	for _, h := range hooks {
		h.Handler(data)
	}

	m.mu.Lock()
}

// GetRenderableFactory returns the renderable factory for the given type name, if registered.
func (m *Manager) GetRenderableFactory(typeName string) (RenderableFactory, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	factory, ok := m.renderableFactories[typeName]
	return factory, ok
}

// GetRegisteredCommand returns the command handler for the given name, if registered.
func (m *Manager) GetRegisteredCommand(name string) (CommandHandler, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	info, ok := m.registeredCommands[name]
	if !ok {
		return nil, false
	}
	return info.Handler, ok
}

// GetRegisteredCommandInfo returns the command info for the given name, if registered.
func (m *Manager) GetRegisteredCommandInfo(name string) (CommandInfo, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	info, ok := m.registeredCommands[name]
	return info, ok
}

// GetRegisteredCommands returns all registered command names.
func (m *Manager) GetRegisteredCommands() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.registeredCommands))
	for name := range m.registeredCommands {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// GetEventHandlers returns a copy of all event handlers registered for the given event.
func (m *Manager) GetEventHandlers(event string) []EventHandler {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]EventHandler{}, m.eventHandlers[event]...)
}

// Dependencies returns the declared dependencies of a plugin.
func (m *Manager) Dependencies(name string) ([]string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	info, ok := m.infos[name]
	if !ok {
		return nil, false
	}
	deps := make([]string, len(info.Dependencies))
	copy(deps, info.Dependencies)
	return deps, ok
}

// listLocked returns plugin names in insertion order. Callers must hold the
// manager lock.
func (m *Manager) listLocked() []string {
	names := make([]string, 0, len(m.plugins))
	for name := range m.plugins {
		names = append(names, name)
	}
	return names
}

// BasePlugin provides a base implementation for plugins with common fields.
type BasePlugin struct {
	name    string
	version string
}

// NewBasePlugin creates a new BasePlugin.
func NewBasePlugin(name, version string) *BasePlugin {
	return &BasePlugin{name: name, version: version}
}

// Name returns the plugin name.
func (p *BasePlugin) Name() string { return p.name }

// Version returns the plugin version.
func (p *BasePlugin) Version() string { return p.version }

// Init is a no-op base initialization.
func (p *BasePlugin) Init(ctx *Context) error { return nil }

// Shutdown is a no-op base shutdown.
func (p *BasePlugin) Shutdown() error { return nil }

// ---------------------------------------------------------------------------
// Error types
// ---------------------------------------------------------------------------

// ErrorCode categorizes plugin errors returned by the Manager.
type ErrorCode int

const (
	// ErrCodeInvalid means the plugin or its input is invalid.
	ErrCodeInvalid ErrorCode = iota + 1
	// ErrCodeAlreadyExists means a plugin with the same name is already registered.
	ErrCodeAlreadyExists
	// ErrCodeNotFound means a plugin was not found.
	ErrCodeNotFound
	// ErrCodeInitFailed means the plugin's Init failed.
	ErrCodeInitFailed
	// ErrCodeShutdownFailed means the plugin's Shutdown failed.
	ErrCodeShutdownFailed
	// ErrCodeDependencyUnsatisfied means a plugin's dependencies are not met.
	ErrCodeDependencyUnsatisfied
)

func (c ErrorCode) String() string {
	switch c {
	case ErrCodeInvalid:
		return "invalid"
	case ErrCodeAlreadyExists:
		return "already_exists"
	case ErrCodeNotFound:
		return "not_found"
	case ErrCodeInitFailed:
		return "init_failed"
	case ErrCodeShutdownFailed:
		return "shutdown_failed"
	case ErrCodeDependencyUnsatisfied:
		return "dependency_unsatisfied"
	default:
		return fmt.Sprintf("code(%d)", int(c))
	}
}

// PluginError is a structured error returned by the Manager.
type PluginError struct {
	// Code categorizes the error.
	Code ErrorCode
	// Plugin is the name of the plugin involved, if any.
	Plugin string
	// Err is the underlying error.
	Err error
}

func (e *PluginError) Error() string {
	if e.Plugin != "" {
		return fmt.Sprintf("plugin %q: %v", e.Plugin, e.Err)
	}
	return e.Err.Error()
}

// Unwrap returns the underlying error.
func (e *PluginError) Unwrap() error { return e.Err }

// ---------------------------------------------------------------------------
// Middleware
// ---------------------------------------------------------------------------

// Middleware wraps plugin registration. It receives the plugin and a function
// to proceed with registration; it may modify, reject, or short-circuit the
// registration by returning an error without calling next.
type Middleware func(p Plugin, next func(Plugin) error) error

// Use registers middleware that runs during plugin registration, in the order
// added. Middleware runs after the plugin's own Init has been called but before
// the plugin is marked active.
func (m *Manager) Use(mw Middleware) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.middleware = append(m.middleware, mw)
}

// ---------------------------------------------------------------------------
// Registration with middleware
// ---------------------------------------------------------------------------

// registerLocked performs the full registration pipeline assuming the manager
// lock is held. It runs all registered middleware in order; each middleware may
// call next to proceed to the following middleware, or return an error to
// short-circuit. The final next completes Init and marks the plugin active.
func (m *Manager) registerLocked(p Plugin) error {
	name := p.Name()
	if name == "" {
		return &PluginError{Code: ErrCodeInvalid, Err: fmt.Errorf("plugin name must not be empty")}
	}
	if _, exists := m.plugins[name]; exists {
		return &PluginError{Code: ErrCodeAlreadyExists, Plugin: name, Err: fmt.Errorf("plugin %q already registered", name)}
	}

	info := defaultInfo(p)
	m.infos[name] = info
	m.states[name] = StateRegistered
	m.enabled[name] = true

	var chain func(Plugin) error
	chain = func(current Plugin) error {
		return current.Init(m.ctx)
	}
	// Wrap middleware in reverse so the first registered middleware runs first.
	for i := len(m.middleware) - 1; i >= 0; i-- {
		mw := m.middleware[i]
		next := chain
		chain = func(current Plugin) error {
			return mw(current, next)
		}
	}

	if err := chain(p); err != nil {
		m.states[name] = StateError
		return &PluginError{Code: ErrCodeInitFailed, Plugin: name, Err: fmt.Errorf("failed to init plugin %q: %w", name, err)}
	}

	m.states[name] = StateInitialized
	if m.enabled[name] {
		m.states[name] = StateActive
	}
	m.plugins[name] = p

	m.emitEventLocked(EventPluginRegistered, map[string]interface{}{
		"plugin": name,
		"info":   info,
	})
	m.emitEventLocked(EventPluginActivated, map[string]interface{}{
		"plugin": name,
	})
	return nil
}

// Register registers a plugin with the manager. The plugin's Init is invoked
// with the manager context, followed by any registered middleware. If the plugin
// implements Descriptor, its metadata is recorded and its dependencies are tracked.
func (m *Manager) Register(p Plugin) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.registerLocked(p)
}
