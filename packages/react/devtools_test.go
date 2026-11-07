package react

import (
	"testing"

	"github.com/Gomobra/PixTui/packages/core"
)

func TestDevToolsEnabled(t *testing.T) {
	ctx := &core.RenderContext{Width: 80, Height: 24}
	renderer := NewReactRenderer(ctx)
	devtools := NewDevTools(renderer)

	if devtools.Enabled() {
		t.Error("expected devtools to be disabled by default")
	}

	devtools.Enable()
	if !devtools.Enabled() {
		t.Error("expected devtools to be enabled")
	}

	devtools.Disable()
	if devtools.Enabled() {
		t.Error("expected devtools to be disabled")
	}
}

func TestDevToolsFiberTree(t *testing.T) {
	ctx := &core.RenderContext{Width: 80, Height: 24}
	renderer := NewReactRenderer(ctx)
	devtools := NewDevTools(renderer)
	devtools.Enable()

	buffer := core.NewNativeBuffer(80, 24)
	element := CreateElement("box", map[string]interface{}{"title": "root"},
		CreateElement("text", map[string]interface{}{"content": "child"}))
	renderer.Render(element, buffer, 0)

	tree := devtools.FiberTree()
	if tree == nil {
		t.Fatal("expected non-nil fiber tree")
	}
	if tree.Type != "box" {
		t.Errorf("expected root type 'box', got %s", tree.Type)
	}
	if len(tree.Children) != 1 {
		t.Errorf("expected 1 child, got %d", len(tree.Children))
	}
	if tree.Children[0].Type != "text" {
		t.Errorf("expected child type 'text', got %s", tree.Children[0].Type)
	}
	if tree.ChildCount != 1 {
		t.Errorf("expected childCount 1, got %d", tree.ChildCount)
	}
}

func TestDevToolsDisabledReturnsNil(t *testing.T) {
	ctx := &core.RenderContext{Width: 80, Height: 24}
	renderer := NewReactRenderer(ctx)
	devtools := NewDevTools(renderer)
	// not enabled

	if devtools.FiberTree() != nil {
		t.Error("expected nil fiber tree when disabled")
	}
}

func TestDevToolsInspect(t *testing.T) {
	ctx := &core.RenderContext{Width: 80, Height: 24}
	renderer := NewReactRenderer(ctx)
	devtools := NewDevTools(renderer)
	devtools.Enable()

	buffer := core.NewNativeBuffer(80, 24)
	renderer.Render(CreateElement("text", map[string]interface{}{"content": "hi"}), buffer, 0)

	out := devtools.Inspect()
	if out == "" {
		t.Error("expected non-empty inspect output")
	}
	if out == "<devtools disabled>" {
		t.Error("expected inspect output, not disabled message")
	}
}
