package renderables

import (
	"testing"

	"github.com/Gomobra/PixTui/packages/core"
)

func newTestCtx() *core.RenderContext {
	return &core.RenderContext{
		Width:  80,
		Height: 24,
	}
}

func TestBoxRenderable(t *testing.T) {
	ctx := newTestCtx()
	r := NewBoxRenderable(ctx, BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      "test-box",
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Test",
	})
	if r.ID() != "test-box" {
		t.Errorf("expected ID 'test-box', got '%s'", r.ID())
	}
	if !r.IsVisible() {
		t.Error("expected visible")
	}
	buf := core.NewNativeBuffer(80, 24)
	r.RenderSelf(buf, 0)
	// Just verify no panic
}

func TestTextRenderable(t *testing.T) {
	ctx := newTestCtx()
	r := NewTextRenderable(ctx, TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      "test-text",
			Visible: boolPtr(true),
		},
		Content: "Hello World",
	})
	if r.Content() != "Hello World" {
		t.Errorf("expected content 'Hello World', got '%s'", r.Content())
	}
	buf := core.NewNativeBuffer(80, 24)
	r.RenderSelf(buf, 0)
}

func TestTextRenderableSetContent(t *testing.T) {
	ctx := newTestCtx()
	r := NewTextRenderable(ctx, TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{ID: "test"},
		Content:           "initial",
	})
	r.SetContent("updated")
	if r.Content() != "updated" {
		t.Errorf("expected 'updated', got '%s'", r.Content())
	}
}

func TestInputRenderable(t *testing.T) {
	ctx := newTestCtx()
	r := NewInputRenderable(ctx, InputOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{ID: "test-input"},
		Value:             "type here",
		Placeholder:       "placeholder",
	})
	if r.Value() != "type here" {
		t.Errorf("expected value 'type here', got '%s'", r.Value())
	}
	buf := core.NewNativeBuffer(80, 24)
	r.RenderSelf(buf, 0)
}

func TestSelectRenderable(t *testing.T) {
	ctx := newTestCtx()
	r := NewSelectRenderable(ctx, SelectOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{ID: "test-select"},
		Options:           []string{"a", "b", "c"},
		Selected:          1,
	})
	buf := core.NewNativeBuffer(80, 24)
	r.RenderSelf(buf, 0)
}

func TestScrollBoxRenderable(t *testing.T) {
	ctx := newTestCtx()
	r := NewScrollBoxRenderable(ctx, ScrollBoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{ID: "test-scrollbox"},
	})
	buf := core.NewNativeBuffer(80, 24)
	r.RenderSelf(buf, 0)
}

func TestTextareaRenderable(t *testing.T) {
	ctx := newTestCtx()
	r := NewTextareaRenderable(ctx, TextareaOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{ID: "test-textarea"},
		Value:             "multiline\ntext",
	})
	buf := core.NewNativeBuffer(80, 24)
	r.RenderSelf(buf, 0)
}

func TestSliderRenderable(t *testing.T) {
	ctx := newTestCtx()
	r := NewSliderRenderable(ctx, SliderOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{ID: "test-slider"},
		Min:               0,
		Max:               100,
		Value:             50,
	})
	buf := core.NewNativeBuffer(80, 24)
	r.RenderSelf(buf, 0)
}

func TestTabSelectRenderable(t *testing.T) {
	ctx := newTestCtx()
	r := NewTabSelectRenderable(ctx, TabSelectOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{ID: "test-tabselect"},
		Tabs:              []string{"tab1", "tab2"},
		Selected:          0,
	})
	buf := core.NewNativeBuffer(80, 24)
	r.RenderSelf(buf, 0)
}

func TestCodeRenderable(t *testing.T) {
	ctx := newTestCtx()
	r := NewCodeRenderable(ctx, CodeOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{ID: "test-code"},
		Code:              "fmt.Println(\"hi\")",
		Language:          "go",
	})
	buf := core.NewNativeBuffer(80, 24)
	r.RenderSelf(buf, 0)
}

func TestMarkdownRenderable(t *testing.T) {
	ctx := newTestCtx()
	r := NewMarkdownRenderable(ctx, MarkdownOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{ID: "test-markdown"},
		Content:           "# Heading\n\nSome text",
	})
	buf := core.NewNativeBuffer(80, 24)
	r.RenderSelf(buf, 0)
}

func TestDiffRenderable(t *testing.T) {
	ctx := newTestCtx()
	r := NewDiffRenderable(ctx, DiffOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{ID: "test-diff"},
		Content:           "-old\n+new",
	})
	buf := core.NewNativeBuffer(80, 24)
	r.RenderSelf(buf, 0)
}

func TestLineNumberRenderable(t *testing.T) {
	ctx := newTestCtx()
	r := NewLineNumberRenderable(ctx, "test-linenumber", 1)
	buf := core.NewNativeBuffer(80, 24)
	r.RenderSelf(buf, 0)
}

func TestScrollBarRenderable(t *testing.T) {
	ctx := newTestCtx()
	r := NewScrollBarRenderable(ctx, ScrollBarOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{ID: "test-scrollbar"},
		Vertical:          true,
		ThumbSize:         10,
		ThumbStart:        0,
	})
	buf := core.NewNativeBuffer(80, 24)
	r.RenderSelf(buf, 0)
}

func TestTextTableRenderable(t *testing.T) {
	ctx := newTestCtx()
	r := NewTextTableRenderable(ctx, TextTableOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{ID: "test-texttable"},
		Headers:           []string{"Col1", "Col2"},
		Rows:              [][]string{{"a", "b"}, {"c", "d"}},
	})
	buf := core.NewNativeBuffer(80, 24)
	r.RenderSelf(buf, 0)
}

func TestASCIIFontRenderable(t *testing.T) {
	ctx := newTestCtx()
	r := NewASCIIFontRenderable(ctx, ASCIIFontOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{ID: "test-ascii"},
		Text:              "HI",
		Font:              "standard",
	})
	buf := core.NewNativeBuffer(80, 24)
	r.RenderSelf(buf, 0)
}

func TestFrameBufferRenderable(t *testing.T) {
	ctx := newTestCtx()
	r := NewFrameBufferRenderable(ctx, FrameBufferOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{ID: "test-framebuffer"},
	})
	buf := core.NewNativeBuffer(80, 24)
	r.RenderSelf(buf, 0)
}

func TestEditBufferRenderable(t *testing.T) {
	ctx := newTestCtx()
	r := NewEditBufferRenderable(ctx, EditBufferRenderableOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{ID: "test-editbuffer"},
		Value:             "editable text",
		ReadOnly:          false,
	})
	buf := core.NewNativeBuffer(80, 24)
	r.RenderSelf(buf, 0)
}

func TestRenderableVisibilityToggle(t *testing.T) {
	ctx := newTestCtx()
	r := NewBoxRenderable(ctx, BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      "test-vis",
			Visible: boolPtr(true),
		},
	})
	if !r.IsVisible() {
		t.Error("expected visible initially")
	}
	r.SetVisible(false)
	if r.IsVisible() {
		t.Error("expected not visible")
	}
	r.SetVisible(true)
	if !r.IsVisible() {
		t.Error("expected visible again")
	}
}

func TestRenderableZIndex(t *testing.T) {
	ctx := newTestCtx()
	r := NewBoxRenderable(ctx, BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "test-zindex",
			ZIndex: intPtr(5),
		},
	})
	// ZIndex is stored internally; verify via options
	if r.options.ZIndex == nil || *r.options.ZIndex != 5 {
		t.Errorf("expected zIndex 5")
	}
}

func TestRenderableDestroy(t *testing.T) {
	ctx := newTestCtx()
	r := NewBoxRenderable(ctx, BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{ID: "test-destroy"},
	})
	r.Destroy()
	if !r.IsDestroyed() {
		t.Error("expected destroyed")
	}
}

func TestRenderableAddRemoveChild(t *testing.T) {
	ctx := newTestCtx()
	parent := NewBoxRenderable(ctx, BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{ID: "parent"},
	})
	child := NewTextRenderable(ctx, TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{ID: "child"},
		Content:           "child",
	})
	parent.AddChild(child.Renderable)
	if len(parent.Children()) != 1 {
		t.Errorf("expected 1 child, got %d", len(parent.Children()))
	}
	parent.RemoveChild(child.Renderable)
	if len(parent.Children()) != 0 {
		t.Errorf("expected 0 children after remove, got %d", len(parent.Children()))
	}
}

// Helper pointer functions (duplicated locally to avoid import issues)
func boolPtr(b bool) *bool { return &b }
func intPtr(i int) *int    { return &i }

// ----- Benchmarks -----

func BenchmarkBoxRender(b *testing.B) {
	ctx := newTestCtx()
	r := NewBoxRenderable(ctx, BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "bench-box",
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Bench",
	})
	buf := core.NewNativeBuffer(80, 24)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.RenderSelf(buf, 0)
	}
}

func BenchmarkTextRender(b *testing.B) {
	ctx := newTestCtx()
	r := NewTextRenderable(ctx, TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "bench-text",
			Visible: boolPtr(true),
		},
		Content: "Hello World",
	})
	buf := core.NewNativeBuffer(80, 24)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.RenderSelf(buf, 0)
	}
}

func BenchmarkInputRender(b *testing.B) {
	ctx := newTestCtx()
	r := NewInputRenderable(ctx, InputOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{ID: "bench-input"},
		Value:             "benchmark input",
	})
	buf := core.NewNativeBuffer(80, 24)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.RenderSelf(buf, 0)
	}
}
