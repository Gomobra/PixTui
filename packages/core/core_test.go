package core

import (
	"testing"
)

func TestRenderableCreate(t *testing.T) {
	ctx := &RenderContext{
		Width:  80,
		Height: 24,
	}
	r := NewRenderable(ctx, "test", RenderableOptions[Renderable]{})
	if r.ID() != "test" {
		t.Errorf("expected ID 'test', got '%s'", r.ID())
	}
	if !r.IsVisible() {
		t.Error("expected visible to be true")
	}
	if r.IsDestroyed() {
		t.Error("expected not destroyed")
	}
}

func TestRenderableDestroy(t *testing.T) {
	ctx := &RenderContext{
		Width:  80,
		Height: 24,
	}
	r := NewRenderable(ctx, "test", RenderableOptions[Renderable]{})
	r.Destroy()
	if !r.IsDestroyed() {
		t.Error("expected destroyed to be true")
	}
}

func TestRenderableSetPosition(t *testing.T) {
	ctx := &RenderContext{
		Width:  80,
		Height: 24,
	}
	r := NewRenderable(ctx, "test", RenderableOptions[Renderable]{})
	r.SetPosition(10, 20)
	// Position is set internally, just verify no panic
}

func TestRenderableSetZIndex(t *testing.T) {
	ctx := &RenderContext{
		Width:  80,
		Height: 24,
	}
	r := NewRenderable(ctx, "test", RenderableOptions[Renderable]{})
	r.SetZIndex(5)
}

func TestRenderableSetVisible(t *testing.T) {
	ctx := &RenderContext{
		Width:  80,
		Height: 24,
	}
	r := NewRenderable(ctx, "test", RenderableOptions[Renderable]{})
	r.SetVisible(false)
	if r.IsVisible() {
		t.Error("expected visible to be false")
	}
	r.SetVisible(true)
	if !r.IsVisible() {
		t.Error("expected visible to be true")
	}
}

func TestRenderableSetOpacity(t *testing.T) {
	ctx := &RenderContext{
		Width:  80,
		Height: 24,
	}
	r := NewRenderable(ctx, "test", RenderableOptions[Renderable]{})
	r.SetOpacity(0.5)
}

func TestRenderableAddChild(t *testing.T) {
	ctx := &RenderContext{
		Width:  80,
		Height: 24,
	}
	parent := NewRenderable(ctx, "parent", RenderableOptions[Renderable]{})
	child := NewRenderable(ctx, "child", RenderableOptions[Renderable]{})
	parent.AddChild(child)
	children := parent.Children()
	if len(children) != 1 {
		t.Errorf("expected 1 child, got %d", len(children))
	}
	cr, ok := children[0].(*Renderable)
	if !ok {
		t.Errorf("expected child to be *Renderable")
	}
	if cr.ID() != "child" {
		t.Errorf("expected child ID 'child', got '%s'", cr.ID())
	}
	if cr.Parent() != parent {
		t.Error("expected child parent to be parent")
	}
}

func TestRenderableRemoveChild(t *testing.T) {
	ctx := &RenderContext{
		Width:  80,
		Height: 24,
	}
	parent := NewRenderable(ctx, "parent", RenderableOptions[Renderable]{})
	child := NewRenderable(ctx, "child", RenderableOptions[Renderable]{})
	parent.AddChild(child)
	parent.RemoveChild(child)
	children := parent.Children()
	if len(children) != 0 {
		t.Errorf("expected 0 children, got %d", len(children))
	}
	if child.Parent() != nil {
		t.Error("expected child parent to be nil")
	}
}

func TestRootRenderable(t *testing.T) {
	ctx := &RenderContext{
		Width:  80,
		Height: 24,
	}
	root := NewRootRenderable(ctx, "root")
	if root.ID() != "root" {
		t.Errorf("expected ID 'root', got '%s'", root.ID())
	}
}

func TestNativeBuffer(t *testing.T) {
	buf := NewNativeBuffer(10, 5)
	if buf.Width() != 10 {
		t.Errorf("expected width 10, got %d", buf.Width())
	}
	if buf.Height() != 5 {
		t.Errorf("expected height 5, got %d", buf.Height())
	}
	buf.Clear()
	buf.SetCell(0, 0, 'A', 0, 0, 0)
	access := buf.Access()
	if access.Char[0] != 'A' {
		t.Errorf("expected char 'A', got %c", access.Char[0])
	}
}

func TestCliRendererCreate(t *testing.T) {
	r := NewCliRenderer(80, 24)
	if r == nil {
		t.Fatal("expected non-nil renderer")
	}
	w, h := r.Dimensions()
	if w != 80 || h != 24 {
		t.Errorf("expected dimensions 80x24, got %dx%d", w, h)
	}
}

func TestCliRendererSetDimensions(t *testing.T) {
	r := NewCliRenderer(80, 24)
	r.SetDimensions(100, 50)
	w, h := r.Dimensions()
	if w != 100 || h != 50 {
		t.Errorf("expected dimensions 100x50, got %dx%d", w, h)
	}
}

func TestRender(t *testing.T) {
	renderer := NewCliRenderer(80, 24)
	root := NewRootRenderable(renderer.Context(), "root")
	err := renderer.Render(root, 0.016)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRenderWithChildren(t *testing.T) {
	renderer := NewCliRenderer(80, 24)
	ctx := renderer.Context()
	root := NewRootRenderable(ctx, "root")
	child := NewRenderable(ctx, "child", RenderableOptions[Renderable]{})
	root.AddChild(child)
	err := renderer.Render(root, 0.016)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCapture(t *testing.T) {
	renderer := NewCliRenderer(80, 24)
	root := NewRootRenderable(renderer.Context(), "root")
	renderer.Render(root, 0.016)
	access := renderer.Capture()
	if access.Width != 80 || access.Height != 24 {
		t.Errorf("expected 80x24, got %dx%d", access.Width, access.Height)
	}
}
