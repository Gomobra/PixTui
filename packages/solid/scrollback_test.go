package solid

import (
	"testing"

	"github.com/Gomobra/PixTui/packages/core"
)

func TestCreateScrollbackWriter(t *testing.T) {
	writer := createScrollbackWriter(func(ctx ScrollbackRenderContext) SolidComponent {
		return Box(map[string]interface{}{
			"id":   "scrollback-box",
			"border": true,
			"title": "Snapshot",
		})
	}, WithWidth(80), WithHeight(24))

	ctx := ScrollbackRenderContext{
		Width:         80,
		Height:        24,
		RenderContext: &core.RenderContext{Width: 80, Height: 24},
	}
	snapshot := writer(ctx)
	if snapshot == nil {
		t.Fatal("expected non-nil snapshot")
	}
	if snapshot.Width != 80 {
		t.Errorf("expected width 80, got %d", snapshot.Width)
	}
	if snapshot.Height <= 0 {
		t.Errorf("expected positive height, got %d", snapshot.Height)
	}
	if snapshot.Root == nil {
		t.Error("expected non-nil root renderable")
	}
	snapshot.Teardown()
}

func TestScrollbackWriterDefaultDimensions(t *testing.T) {
	writer := createScrollbackWriter(func(ctx ScrollbackRenderContext) SolidComponent {
		return Text(map[string]interface{}{"content": "default"})
	})

	ctx := ScrollbackRenderContext{
		Width:         80,
		Height:        24,
		RenderContext: &core.RenderContext{Width: 80, Height: 24},
	}
	snapshot := writer(ctx)
	if snapshot == nil {
		t.Fatal("expected non-nil snapshot")
	}
	if snapshot.Width <= 0 {
		t.Errorf("expected positive width, got %d", snapshot.Width)
	}
}

func TestWriteSolidToScrollback(t *testing.T) {
	ctx := &core.RenderContext{Width: 80, Height: 24}
	renderer := NewSolidRenderer(ctx)

	snapshot, err := writeSolidToScrollback(renderer, func(c ScrollbackRenderContext) SolidComponent {
		return Box(map[string]interface{}{
			"id":   "write-test",
			"border": true,
		})
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snapshot == nil {
		t.Fatal("expected non-nil snapshot")
	}
	if snapshot.Width <= 0 {
		t.Errorf("expected positive width, got %d", snapshot.Width)
	}
}

func TestScrollbackOptions(t *testing.T) {
	opts := &ScrollbackWriterOptions{}
	WithWidth(120)(opts)
	WithHeight(40)(opts)
	WithStartOnNewLine(false)(opts)
	WithRowColumns(80)(opts)
	WithTrailingNewline(true)(opts)

	if opts.Width != 120 {
		t.Errorf("expected width 120, got %d", opts.Width)
	}
	if opts.Height != 40 {
		t.Errorf("expected height 40, got %d", opts.Height)
	}
	if opts.StartOnNewLine {
		t.Error("expected StartOnNewLine false")
	}
	if opts.RowColumns != 80 {
		t.Errorf("expected rowColumns 80, got %d", opts.RowColumns)
	}
	if !opts.TrailingNewline {
		t.Error("expected TrailingNewline true")
	}
}

var _ = core.NewRootRenderable
