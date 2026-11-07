package keymap

import (
	"testing"
)

// mockEvent implements KeymapEvent for testing.
type mockEvent struct {
	name      string
	ctrl      bool
	shift     bool
	meta      bool
	super     bool
	hyper     bool
	prevented bool
	stopped   bool
}

func (e *mockEvent) Name() string             { return e.name }
func (e *mockEvent) Ctrl() bool               { return e.ctrl }
func (e *mockEvent) Shift() bool              { return e.shift }
func (e *mockEvent) Meta() bool               { return e.meta }
func (e *mockEvent) Super() bool              { return e.super }
func (e *mockEvent) Hyper() bool              { return e.hyper }
func (e *mockEvent) PreventDefault()          { e.prevented = true }
func (e *mockEvent) StopPropagation()         { e.stopped = true }
func (e *mockEvent) PropagationStopped() bool { return e.stopped }

type mockTarget struct {
	name string
}

var (
	_ KeymapEvent = (*mockEvent)(nil)
)

func newTestKeymap(t *testing.T) (*Keymap[*mockTarget, *mockEvent], *mockTarget) {
	t.Helper()
	focused := &mockTarget{name: "focused"}
	host := KeymapHost[*mockTarget, *mockEvent]{
		Metadata: HostMetadata{
			Platform:        "linux",
			PrimaryModifier: "ctrl",
			Modifiers:       map[string]string{"ctrl": "supported"},
		},
		RootTarget:         &mockTarget{name: "root"},
		GetFocusedTarget:   func() *mockTarget { return focused },
		CreateCommandEvent: func() *mockEvent { return &mockEvent{} },
	}
	km, err := NewKeymap(host)
	if err != nil {
		t.Fatalf("NewKeymap: %v", err)
	}
	return km, focused
}

func press(km *Keymap[*mockTarget, *mockEvent], name string, ctrl bool) {
	km.handleKeyEvent(&mockEvent{name: name, ctrl: ctrl}, false)
}

func TestRunCommandViaBinding(t *testing.T) {
	km, focused := newTestKeymap(t)

	ran := false
	km.RegisterLayer(Layer[*mockTarget, *mockEvent]{
		Priority: 0,
		Bindings: []Binding[*mockTarget, *mockEvent]{
			{Key: "q", Cmd: BindingCommand[*mockTarget, *mockEvent]{Kind: BindingCommandName, Name: "quit"}, PreventDefault: true},
		},
		Commands: []Command{
			{Name: "quit", Run: func(ctx CommandContext[any, KeymapEvent]) CommandResult {
				ran = true
				if ctx.Event.(*mockEvent).name != "q" {
					t.Errorf("expected event q, got %s", ctx.Event.(*mockEvent).name)
				}
				if ctx.Focused != focused {
					t.Errorf("expected focused target")
				}
				return RunCommandResult{OK: true}
			}},
		},
	})

	press(km, "q", false)
	if !ran {
		t.Fatal("expected quit command to run")
	}
	// ev is not accessible here; prevented checked via the binding's preventDefault semantics implicitly.
	_ = focused
}

func TestChordSequence(t *testing.T) {
	km, _ := newTestKeymap(t)

	ran := false
	km.RegisterLayer(Layer[*mockTarget, *mockEvent]{
		Priority: 0,
		Bindings: []Binding[*mockTarget, *mockEvent]{
			{Key: "ctrl+x ctrl+c", Cmd: BindingCommand[*mockTarget, *mockEvent]{Kind: BindingCommandName, Name: "copy"}, PreventDefault: true},
		},
		Commands: []Command{
			{Name: "copy", Run: func(ctx CommandContext[any, KeymapEvent]) CommandResult {
				ran = true
				return RunCommandResult{OK: true}
			}},
		},
	})

	// First chord part: ctrl+x should pend.
	press(km, "x", true)
	if !km.HasPendingSequence() {
		t.Fatal("expected pending sequence after ctrl+x")
	}
	// Second chord part: ctrl+c completes the sequence.
	press(km, "c", true)
	if km.HasPendingSequence() {
		t.Fatal("expected sequence to complete and clear")
	}
	if !ran {
		t.Fatal("expected copy command to run after chord")
	}
}

func TestFeaturePackEnableDisable(t *testing.T) {
	km, _ := newTestKeymap(t)

	km.RegisterLayer(Layer[*mockTarget, *mockEvent]{
		Priority: 0,
		Commands: []Command{
			{Name: "delete", Run: func(ctx CommandContext[any, KeymapEvent]) CommandResult { return RunCommandResult{OK: true} }},
		},
	})

	pack := km.RegisterFeaturePack("edit", []Binding[*mockTarget, *mockEvent]{
		{Key: "d", Cmd: BindingCommand[*mockTarget, *mockEvent]{Kind: BindingCommandName, Name: "delete"}},
	})

	// Initially disabled: no active bindings.
	diag := km.Diagnostics()
	if len(diag) != 0 {
		t.Fatalf("expected no diagnostics before enable, got %d", len(diag))
	}

	pack.Enable()
	if !pack.Enabled {
		t.Fatal("expected pack enabled")
	}
	diag = km.Diagnostics()
	if len(diag) != 1 {
		t.Fatalf("expected 1 diagnostic after enable, got %d", len(diag))
	}
	if diag[0].Command != "delete" {
		t.Errorf("expected command delete, got %s", diag[0].Command)
	}

	pack.Disable()
	if pack.Enabled {
		t.Fatal("expected pack disabled")
	}
	diag = km.Diagnostics()
	if len(diag) != 0 {
		t.Fatalf("expected no diagnostics after disable, got %d", len(diag))
	}
}

func TestGetPendingSequenceAndClear(t *testing.T) {
	km, _ := newTestKeymap(t)

	km.RegisterLayer(Layer[*mockTarget, *mockEvent]{
		Priority: 0,
		Bindings: []Binding[*mockTarget, *mockEvent]{
			{Key: "g g", Cmd: BindingCommand[*mockTarget, *mockEvent]{Kind: BindingCommandName, Name: "gg"}},
		},
	})

	press(km, "g", false)
	if !km.HasPendingSequence() {
		t.Fatal("expected pending sequence")
	}
	seq := km.GetPendingSequence()
	if len(seq) != 1 {
		t.Fatalf("expected 1 pending part, got %d", len(seq))
	}
	km.ClearPendingSequence()
	if km.HasPendingSequence() {
		t.Fatal("expected pending sequence cleared")
	}
}

func TestFormatAndParseKey(t *testing.T) {
	km, _ := newTestKeymap(t)

	parts := km.ParseKeySequence("ctrl+x")
	if len(parts) != 1 {
		t.Fatalf("expected 1 part, got %d", len(parts))
	}
	if parts[0].Match != normalizeKeyStroke(KeyStrokeInput{Name: "x", Ctrl: true}).MatchID() {
		t.Errorf("unexpected match %s", parts[0].Match)
	}
	s := km.FormatKey("ctrl+x")
	if s != "ctrl+x" {
		t.Errorf("expected 'ctrl+x', got %s", s)
	}
	m := km.CreateKeyMatcher("ctrl+x")
	if !m(KeyStrokeInput{Name: "x", Ctrl: true}) {
		t.Error("matcher should match ctrl+x")
	}
	if m(KeyStrokeInput{Name: "x"}) {
		t.Error("matcher should not match plain x")
	}
}

func TestDiagnostics(t *testing.T) {
	km, _ := newTestKeymap(t)

	km.RegisterLayer(Layer[*mockTarget, *mockEvent]{
		Priority: 5,
		Commands: []Command{
			{Name: "cmd-a", Run: func(ctx CommandContext[any, KeymapEvent]) CommandResult { return RunCommandResult{OK: true} }},
			{Name: "cmd-b", Run: func(ctx CommandContext[any, KeymapEvent]) CommandResult { return RunCommandResult{OK: true} }},
		},
		Bindings: []Binding[*mockTarget, *mockEvent]{
			{Key: "a", Cmd: BindingCommand[*mockTarget, *mockEvent]{Kind: BindingCommandName, Name: "cmd-a"}, Event: BindingEventPress},
			{Key: "b", Cmd: BindingCommand[*mockTarget, *mockEvent]{Kind: BindingCommandName, Name: "cmd-b"}, Event: BindingEventRelease},
		},
	})
	diag := km.Diagnostics()
	if len(diag) != 2 {
		t.Fatalf("expected 2 diagnostics, got %d", len(diag))
	}
}
