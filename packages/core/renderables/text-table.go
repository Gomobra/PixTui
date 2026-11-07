package renderables

import (
	"github.com/Gomobra/PixTui/packages/core"
	"github.com/Gomobra/PixTui/packages/native"
)

// TextTableOptions contains options for TextTableRenderable
type TextTableOptions struct {
	core.RenderableOptions[core.Renderable]
	Headers []string
	Rows    [][]string
	// HeaderColor is the color for header cells.
	HeaderColor *core.RGBA
	// RowColor is the color for body row cells.
	RowColor *core.RGBA
	// BorderColor is the color for table borders/separators.
	BorderColor *core.RGBA
}

// TextTableRenderable renders a text table
type TextTableRenderable struct {
	*core.Renderable
	options TextTableOptions
}

// NewTextTableRenderable creates a new TextTableRenderable
func NewTextTableRenderable(ctx *core.RenderContext, opts TextTableOptions) *TextTableRenderable {
	base := core.NewRenderable(ctx, opts.ID, core.RenderableOptions[core.Renderable]{
		Width:   opts.Width,
		Height:  opts.Height,
		Visible: opts.Visible,
		ZIndex:  opts.ZIndex,
	})
	r := &TextTableRenderable{
		Renderable: base,
		options:    opts,
	}
	// Create native scene node
	if ctx != nil && ctx.NativeScene != nil {
		node, err := ctx.NativeScene.CreateNode(r.Renderable, native.SceneNodeKindBox)
		if err == nil {
			r.Renderable.SetNativeNode(node)
		}
	}
	return r
}

// RenderSelf renders the table into the buffer
func (r *TextTableRenderable) RenderSelf(buffer *core.NativeBuffer, deltaTime float64) {
	// Resolve colors: per-option RGBA takes precedence over renderable defaults.
	fg := r.options.ForegroundColor
	bg := r.options.BackgroundColor
	attrs := r.options.Attributes

	headerColor := r.options.HeaderColor
	if headerColor == nil {
		headerColor = fg
	}
	rowColor := r.options.RowColor
	if rowColor == nil {
		rowColor = fg
	}
	borderColor := r.options.BorderColor
	if borderColor == nil {
		borderColor = fg
	}

	y := 0
	// Render headers
	if len(r.options.Headers) > 0 {
		x := 0
		for _, header := range r.options.Headers {
			headerAttrs := attrs | core.TextAttributeBold
			for _, ch := range header {
				if x >= buffer.Width() || y >= buffer.Height() {
					break
				}
				writeCell(buffer, x, y, uint32(ch), headerColor, bg, headerAttrs)
				x++
			}
			if x < buffer.Width() {
				writeCell(buffer, x, y, ' ', borderColor, bg, attrs)
				x++
			}
		}
		y++
	}

	// Render rows
	for _, row := range r.options.Rows {
		if y >= buffer.Height() {
			break
		}
		x := 0
		for _, cell := range row {
			for _, ch := range cell {
				if x >= buffer.Width() || y >= buffer.Height() {
					break
				}
				writeCell(buffer, x, y, uint32(ch), rowColor, bg, attrs)
				x++
			}
			if x < buffer.Width() {
				writeCell(buffer, x, y, ' ', borderColor, bg, attrs)
				x++
			}
		}
		y++
	}

	if r.options.RenderAfter != nil {
		r.options.RenderAfter(r.Renderable, buffer, deltaTime)
	}
}
