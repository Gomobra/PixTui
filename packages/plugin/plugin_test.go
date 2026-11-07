package plugin

import (
	"testing"
)

// testPlugin is a minimal test plugin implementation.
type testPlugin struct {
	*BasePlugin
	initCalled   bool
	shutdownCalled bool
}

func newTestPlugin(name, version string) *testPlugin {
	return &testPlugin{
		BasePlugin: NewBasePlugin(name, version),
	}
}

func (p *testPlugin) Init(ctx *Context) error {
	p.initCalled = true
	return nil
}

func (p *testPlugin) Shutdown() error {
	p.shutdownCalled = true
	return nil
}

func TestPluginManagerRegister(t *testing.T) {
	m := NewManager()
	p := newTestPlugin("test-plugin", "1.0.0")
	if err := m.Register(p); err != nil {
		t.Fatalf("failed to register plugin: %v", err)
	}

	if m.Count() != 1 {
		t.Errorf("expected 1 plugin, got %d", m.Count())
	}

	got, ok := m.Get("test-plugin")
	if !ok {
		t.Fatal("expected to find registered plugin")
	}
	if got.Name() != "test-plugin" {
		t.Errorf("expected name 'test-plugin', got '%s'", got.Name())
	}
	if got.Version() != "1.0.0" {
		t.Errorf("expected version '1.0.0', got '%s'", got.Version())
	}
	if !p.initCalled {
		t.Error("expected Init to be called on registration")
	}
}

func TestPluginManagerDuplicateRegister(t *testing.T) {
	m := NewManager()
	p1 := newTestPlugin("dup-plugin", "1.0.0")
	p2 := newTestPlugin("dup-plugin", "2.0.0")

	if err := m.Register(p1); err != nil {
		t.Fatalf("failed to register first plugin: %v", err)
	}
	if err := m.Register(p2); err == nil {
		t.Error("expected error on duplicate plugin registration")
	}
}

func TestPluginManagerUnregister(t *testing.T) {
	m := NewManager()
	p := newTestPlugin("unreg-plugin", "1.0.0")
	if err := m.Register(p); err != nil {
		t.Fatalf("failed to register plugin: %v", err)
	}

	if err := m.Unregister("unreg-plugin"); err != nil {
		t.Fatalf("failed to unregister plugin: %v", err)
	}

	if m.Count() != 0 {
		t.Errorf("expected 0 plugins after unregister, got %d", m.Count())
	}
	if !p.shutdownCalled {
		t.Error("expected Shutdown to be called on unregister")
	}

	if _, ok := m.Get("unreg-plugin"); ok {
		t.Error("expected plugin to not be found after unregister")
	}
}

func TestPluginManagerUnregisterNonexistent(t *testing.T) {
	m := NewManager()
	if err := m.Unregister("nonexistent"); err == nil {
		t.Error("expected error when unregistering nonexistent plugin")
	}
}

func TestPluginManagerList(t *testing.T) {
	m := NewManager()
	m.Register(newTestPlugin("plugin-a", "1.0.0"))
	m.Register(newTestPlugin("plugin-b", "2.0.0"))
	m.Register(newTestPlugin("plugin-c", "3.0.0"))

	names := m.List()
	if len(names) != 3 {
		t.Fatalf("expected 3 plugin names, got %d", len(names))
	}
}

func TestPluginManagerShutdown(t *testing.T) {
	m := NewManager()
	p1 := newTestPlugin("shutdown-a", "1.0.0")
	p2 := newTestPlugin("shutdown-b", "2.0.0")
	m.Register(p1)
	m.Register(p2)

	if err := m.Shutdown(); err != nil {
		t.Fatalf("failed to shutdown: %v", err)
	}

	if m.Count() != 0 {
		t.Errorf("expected 0 plugins after shutdown, got %d", m.Count())
	}
	if !p1.shutdownCalled || !p2.shutdownCalled {
		t.Error("expected both plugins to be shutdown")
	}
}

func TestBasePlugin(t *testing.T) {
	p := NewBasePlugin("base", "0.1.0")
	if p.Name() != "base" {
		t.Errorf("expected name 'base', got '%s'", p.Name())
	}
	if p.Version() != "0.1.0" {
		t.Errorf("expected version '0.1.0', got '%s'", p.Version())
	}
	if err := p.Init(nil); err != nil {
		t.Errorf("expected Init to succeed, got %v", err)
	}
	if err := p.Shutdown(); err != nil {
		t.Errorf("expected Shutdown to succeed, got %v", err)
	}
}

func TestManagerHookPriorityOrdering(t *testing.T) {
	m := NewManager()
	var order []string
	m.RegisterHook("ordered-event", HookPriorityLow, func(data interface{}) {
		order = append(order, "low")
	})
	m.RegisterHook("ordered-event", HookPriorityHigh, func(data interface{}) {
		order = append(order, "high")
	})
	m.RegisterHook("ordered-event", HookPriorityNormal, func(data interface{}) {
		order = append(order, "normal")
	})

	m.EmitEvent("ordered-event", nil)

	if len(order) != 3 {
		t.Fatalf("expected 3 hooks to run, got %d: %v", len(order), order)
	}
	if order[0] != "high" || order[1] != "normal" || order[2] != "low" {
		t.Errorf("expected hooks to run high, normal, low; got %v", order)
	}
}

func TestManagerUnregisterHook(t *testing.T) {
	m := NewManager()
	unregister := m.RegisterHook("unregister-event", HookPriorityNormal, func(data interface{}) {})

	if got := m.GetHooks("unregister-event"); len(got) != 1 {
		t.Fatalf("expected 1 hook, got %d", len(got))
	}
	unregister()
	if got := m.GetHooks("unregister-event"); len(got) != 0 {
		t.Fatalf("expected 0 hooks after unregister, got %d", len(got))
	}
	// Unregistering again is a no-op and must not panic.
	unregister()
	if got := m.GetHooks("unregister-event"); len(got) != 0 {
		t.Fatalf("expected 0 hooks after double unregister, got %d", len(got))
	}
}

// hookPlugin registers a hook through the plugin Context during Init.
type hookPlugin struct {
	*BasePlugin
	ctxHookCalled bool
}

func (p *hookPlugin) Init(ctx *Context) error {
	ctx.RegisterHook("ctx-event", HookPriorityHigh, func(data interface{}) {
		p.ctxHookCalled = true
	})
	return nil
}

func (p *hookPlugin) Shutdown() error { return nil }

func TestContextRegisterHook(t *testing.T) {
	m := NewManager()
	p := &hookPlugin{BasePlugin: NewBasePlugin("ctx-hook-plugin", "1.0.0")}
	if err := m.Register(p); err != nil {
		t.Fatalf("failed to register plugin: %v", err)
	}

	m.EmitEvent("ctx-event", nil)
	if !p.ctxHookCalled {
		t.Error("expected context-registered hook to be invoked on emit")
	}
}
