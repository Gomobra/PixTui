package renderables

import (
	"github.com/Gomobra/PixTui/packages/core"
	"github.com/Gomobra/PixTui/packages/native"
)

// ScrollBarOptions contains options for ScrollBarRenderable
type ScrollBarOptions struct {
	core.RenderableOptions[core.Renderable]
	Vertical   bool
	ThumbSize  float64
	ThumbStart float64
}

// ScrollBarRenderable renders a scroll bar
type ScrollBarRenderable struct {
	*core.Renderable
	options ScrollBarOptions
}

// NewScrollBarRenderable creates a new ScrollBarRenderable
func NewScrollBarRenderable(ctx *core.RenderContext, opts ScrollBarOptions) *ScrollBarRenderable {
	base := core.NewRenderable(ctx, opts.ID, core.RenderableOptions[core.Renderable]{
		Width:   opts.Width,
		Height:  opts.Height,
		Visible: opts.Visible,
		ZIndex:  opts.ZIndex,
	})
	r := &ScrollBarRenderable{
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

// RenderSelf renders the scroll bar into the buffer
func (r *ScrollBarRenderable) RenderSelf(buffer *core.NativeBuffer, deltaTime float64) {
	fg := r.options.ForegroundColor
	bg := r.options.BackgroundColor
	attrs := r.options.Attributes

	if r.options.Vertical {
		// Vertical scroll bar on the right edge
		x := buffer.Width() - 1
		if x < 0 {
			return
		}
		for y := 0; y < buffer.Height(); y++ {
			writeCell(buffer, x, y, '│', fg, bg, attrs)
		}
		// Draw thumb
		thumbStart := int(r.options.ThumbStart)
		thumbSize := int(r.options.ThumbSize)
		for y := thumbStart; y < thumbStart+thumbSize && y < buffer.Height(); y++ {
			writeCell(buffer, x, y, '█', fg, bg, attrs)
		}
	} else {
		// Horizontal scroll bar on the bottom edge
		y := buffer.Height() - 1
		if y < 0 {
			return
		}
		for x := 0; x < buffer.Width(); x++ {
			writeCell(buffer, x, y, '─', fg, bg, attrs)
		}
		// Draw thumb
		thumbStart := int(r.options.ThumbStart)
		thumbSize := int(r.options.ThumbSize)
		for x := thumbStart; x < thumbStart+thumbSize && x < buffer.Width(); x++ {
			writeCell(buffer, x, y, '█', fg, bg, attrs)
		}
	}

	if r.options.RenderAfter != nil {
		r.options.RenderAfter(r.Renderable, buffer, deltaTime)
	}
}
