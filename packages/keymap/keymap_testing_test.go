package keymap

import (
	"testing"
)

func TestCreateTestKeymapHost(t *testing.T) {
	host := NewTestKeymapHost(nil)
	if host.RootTarget.ID() != "root" {
		t.Errorf("expected root target id 'root', got %q", host.RootTarget.ID())
	}
	if host.Metadata.Platform != "unknown" {
		t.Errorf("expected default platform 'unknown', got %q", host.Metadata.Platform)
	}
}

func TestCreateTestKeymapHostWithMetadata(t *testing.T) {
	host := NewTestKeymapHost(&CreateTestKeymapHostOptions{Metadata: HostMetadata{
		Platform:        "linux",
		PrimaryModifier: "ctrl",
	}})
	if host.Metadata.Platform != "linux" {
		t.Errorf("expected platform 'linux', got %q", host.Metadata.Platform)
	}
}

func TestTestKeymapEvent(t *testing.T) {
	e := NewTestKeymapEvent("ctrl+x", TestKeyModifierOptions{Ctrl: true})
	if e.Name() != "ctrl+x" {
		t.Errorf("expected name 'ctrl+x', got %q", e.Name())
	}
	if !e.Ctrl() {
		t.Error("expected ctrl to be true")
	}
	if e.Shift() {
		t.Error("expected shift to be false")
	}
	e.PreventDefault()
	if !e.prevented {
		t.Error("expected preventDefault to be called")
	}
}

func TestTestKeymapTargetAppend(t *testing.T) {
	root := NewTestKeymapTarget("root")
	child := root.Append(NewTestKeymapTarget("child"))
	if child.parent != root {
		t.Error("expected child.parent to be root")
	}
}

func TestCreateTestKeymap(t *testing.T) {
	harness := CreateTestKeymap(nil)
	if harness.Keymap == nil {
		t.Fatal("expected keymap to be created")
	}
	if harness.Host == nil {
		t.Fatal("expected host to be created")
	}
	if harness.Root == nil {
		t.Fatal("expected root target")
	}
	if harness.Diagnostics == nil {
		t.Fatal("expected diagnostics capture")
	}
	harness.Cleanup()
	if !harness.Host.isDestroyed {
		t.Error("expected host to be destroyed after cleanup")
	}
}

func TestTestDiagnosticCapture(t *testing.T) {
	harness := CreateTestKeymap(nil)
	defer harness.Cleanup()

	// Emit a warning through the keymap.
	harness.Keymap.emitWarning("test-warning", "this is a warning")
	warnings, _ := harness.Diagnostics.takeWarnings()
	if len(warnings) != 1 || warnings[0] != "this is a warning" {
		t.Errorf("expected 1 warning 'this is a warning', got %v", warnings)
	}

	// Emit an error.
	harness.Keymap.emitError("test-error", nil, "this is an error")
	errors, _ := harness.Diagnostics.takeErrors()
	if len(errors) != 1 || errors[0] != "this is an error" {
		t.Errorf("expected 1 error 'this is an error', got %v", errors)
	}

	// Clear.
	harness.Diagnostics.clear()
	warnings, _ = harness.Diagnostics.takeWarnings()
	if len(warnings) != 0 {
		t.Errorf("expected 0 warnings after clear, got %d", len(warnings))
	}
}

func TestTestKeymapHostPress(t *testing.T) {
	host := NewTestKeymapHost(nil)
	var pressed []string
	host.OnKeyPress(func(e *TestKeymapEvent) {
		pressed = append(pressed, e.Name())
	})

	host.Press("a")
	host.Press("b", TestKeyModifierOptions{Ctrl: true})
	if len(pressed) != 2 || pressed[0] != "a" || pressed[1] != "b" {
		t.Errorf("expected pressed ['a','b'], got %v", pressed)
	}
}

func TestRegisterDefaultKeysNoop(t *testing.T) {
	harness := CreateTestKeymap(&CreateTestKeymapOptions{DefaultKeys: true})
	defer harness.Cleanup()
	if harness.Keymap.HostMetadata().Platform != "unknown" {
		t.Error("expected default metadata platform unknown")
	}

	// Default keys should register bindings that show up in diagnostics.
	diag := harness.Keymap.Diagnostics()
	t.Logf("diagnostics: %+v", diag)
	if len(diag) == 0 {
		t.Error("expected default keys to register diagnostics")
	}
	// Emacs binding ctrl+a should be present.
	found := false
	for _, d := range diag {
		if d.Sequence == "ctrl+a" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected emacs ctrl+a binding in diagnostics")
	}
	// Vim binding h should be present.
	found = false
	for _, d := range diag {
		if d.Sequence == "h" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected vim h binding in diagnostics")
	}
	// Dead binding ctrl+q should be present.
	found = false
	for _, d := range diag {
		if d.Sequence == "ctrl+q" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected dead binding ctrl+q in diagnostics")
	}
	// Leader binding (displayed as "spaceg" since spaces are removed in stringify).
	found = false
	for _, d := range diag {
		if d.Sequence == "spaceg" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected leader binding spaceg in diagnostics")
	}
}

func TestDefaultKeysChord(t *testing.T) {
	harness := CreateTestKeymap(&CreateTestKeymapOptions{DefaultKeys: true})
	defer harness.Cleanup()

	// Use ctrl+x ctrl+c chord (unambiguous, unlike gg vs g).
	harness.Host.Press("x", TestKeyModifierOptions{Ctrl: true})
	if !harness.Keymap.HasPendingSequence() {
		t.Fatal("expected pending sequence after ctrl+x")
	}
	harness.Host.Press("c", TestKeyModifierOptions{Ctrl: true})
	if harness.Keymap.HasPendingSequence() {
		t.Error("expected sequence to clear after ctrl+c")
	}
}
