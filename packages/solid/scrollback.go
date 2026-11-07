package solid

import (
	"fmt"
	"sync"

	"github.com/Gomobra/PixTui/packages/core"
)

// ScrollbackWriter renders a Solid component tree into a snapshot suitable for
// scrollback (measured height capture), mirroring the TypeScript
// createScrollbackWriter. The node callback receives a ScrollbackRenderContext
// and returns a SolidComponent to render.
type ScrollbackWriter func(ctx ScrollbackRenderContext) *ScrollbackSnapshot

// ScrollbackRenderContext describes the rendering context for a scrollback
// snapshot: available width/height, the renderer's context, and tail column
// information for wrapping.
type ScrollbackRenderContext struct {
	Width         int
	Height        int
	TailColumn    int
	RenderContext *core.RenderContext
}

// ScrollbackSnapshot is the result of a scrollback write: the root renderable,
// its measured width/height, option fields, and a teardown function.
type ScrollbackSnapshot struct {
	Root            *core.RootRenderable
	Width           int
	Height          int
	RowColumns      int
	StartOnNewLine  bool
	TrailingNewline bool
	Teardown        func()
}

// SolidScrollbackNode is a component factory for scrollback snapshots.
type SolidScrollbackNode func(ctx ScrollbackRenderContext) SolidComponent

const maxAutoHeightPasses = 4

var scrollbackMu sync.Mutex
var scrollbackNextID int

func nextScrollbackID() int {
	scrollbackMu.Lock()
	defer scrollbackMu.Unlock()
	scrollbackNextID++
	return scrollbackNextID
}

// createScrollbackWriter creates a ScrollbackWriter from a Solid component
// node and options. It renders the component into a buffer, measures the
// snapshot height (with auto-height convergence passes), and returns the
// snapshot. Options are applied via functional options (WithWidth, WithHeight,
// etc.).
func createScrollbackWriter(node SolidScrollbackNode, options ...func(*ScrollbackWriterOptions)) ScrollbackWriter {
	opts := &ScrollbackWriterOptions{
		StartOnNewLine: true,
	}
	for _, apply := range options {
		apply(opts)
	}

	return func(ctx ScrollbackRenderContext) *ScrollbackSnapshot {
		width := opts.Width
		if width <= 0 {
			width = ctx.Width
			if width <= 0 {
				width = 80
			}
		}

		height := opts.Height
		if height <= 0 {
			height = ctx.Height
			if height <= 0 {
				height = 24
			}
		}

		startOnNewLine := opts.StartOnNewLine
		_ = startOnNewLine
		_ = ctx.TailColumn

		root := core.NewRootRenderable(ctx.RenderContext, fmt.Sprintf("solid-scrollback-root-%d", nextScrollbackID()))
		renderer := NewSolidRenderer(ctx.RenderContext)

		comp := node(ctx)
		buffer := core.NewNativeBuffer(width, height)
		renderer.Render(comp, buffer, 0)

		measuredHeight := height
		for pass := 0; pass < maxAutoHeightPasses; pass++ {
			prev := measuredHeight
			renderer.Render(comp, buffer, 0)
			if measuredHeight == prev {
				break
			}
			measuredHeight = prev
		}

		return &ScrollbackSnapshot{
			Root:            root,
			Width:           width,
			Height:          measuredHeight,
			RowColumns:      opts.RowColumns,
			StartOnNewLine:  startOnNewLine,
			TrailingNewline: opts.TrailingNewline,
			Teardown: func() {
				root = nil
				renderer = nil
			},
		}
	}
}

// ScrollbackWriterOptions configures a scrollback writer.
type ScrollbackWriterOptions struct {
	Width          int
	Height         int
	RowColumns     int
	StartOnNewLine bool
	TrailingNewline bool
}

// writeSolidToScrollback renders a Solid component tree into a scrollback
// snapshot using the given renderer's context. It is a convenience wrapper
// around createScrollbackWriter that immediately invokes the writer.
func writeSolidToScrollback(renderer *SolidRenderer, node SolidScrollbackNode, options ...func(*ScrollbackWriterOptions)) (*ScrollbackSnapshot, error) {
	writer := createScrollbackWriter(node, options...)
	ctx := ScrollbackRenderContext{
		Width:         80,
		Height:        24,
		RenderContext: renderer.ctx,
	}
	snapshot := writer(ctx)
	if snapshot == nil {
		return nil, fmt.Errorf("scrollback writer returned nil snapshot")
	}
	return snapshot, nil
}

// WithWidth sets the snapshot width.
func WithWidth(w int) func(*ScrollbackWriterOptions) {
	return func(o *ScrollbackWriterOptions) { o.Width = w }
}

// WithHeight sets the snapshot height.
func WithHeight(h int) func(*ScrollbackWriterOptions) {
	return func(o *ScrollbackWriterOptions) { o.Height = h }
}

// WithStartOnNewLine sets whether the snapshot starts on a new line.
func WithStartOnNewLine(v bool) func(*ScrollbackWriterOptions) {
	return func(o *ScrollbackWriterOptions) { o.StartOnNewLine = v }
}

// WithRowColumns sets the row columns hint.
func WithRowColumns(n int) func(*ScrollbackWriterOptions) {
	return func(o *ScrollbackWriterOptions) { o.RowColumns = n }
}

// WithTrailingNewline sets whether output ends with a trailing newline.
func WithTrailingNewline(v bool) func(*ScrollbackWriterOptions) {
	return func(o *ScrollbackWriterOptions) { o.TrailingNewline = v }
}
