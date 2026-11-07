package renderables

import (
	"github.com/Gomobra/PixTui/packages/core"
	"github.com/Gomobra/PixTui/packages/native"
)

// TextBufferRenderableOptions contains options for TextBufferRenderable.
type TextBufferRenderableOptions struct {
	core.RenderableOptions[core.Renderable]
	// Buffer is the TextBuffer to render.
	Buffer *core.TextBuffer
	// Wrap enables line wrapping.
	Wrap bool
	// WrapWidth is the width to wrap at. When 0 the renderable width is used.
	WrapWidth int
}

// TextBufferRenderable renders a TextBuffer with optional line wrapping.
type TextBufferRenderable struct {
	*core.Renderable
	options TextBufferRenderableOptions
}

// NewTextBufferRenderable creates a new TextBufferRenderable.
func NewTextBufferRenderable(ctx *core.RenderContext, opts TextBufferRenderableOptions) *TextBufferRenderable {
	base := core.NewRenderable(ctx, opts.ID, core.RenderableOptions[core.Renderable]{
		Width:            opts.Width,
		Height:           opts.Height,
		Visible:          opts.Visible,
		ZIndex:           opts.ZIndex,
		ForegroundColor:  opts.ForegroundColor,
		BackgroundColor:  opts.BackgroundColor,
		RenderBefore:     opts.RenderBefore,
		RenderAfter:      opts.RenderAfter,
		OnMouse:          opts.OnMouse,
		OnMouseDown:      opts.OnMouseDown,
		OnMouseUp:        opts.OnMouseUp,
		OnMouseMove:      opts.OnMouseMove,
		OnMouseDrag:      opts.OnMouseDrag,
		OnMouseDragEnd:   opts.OnMouseDragEnd,
		OnMouseDrop:      opts.OnMouseDrop,
		OnMouseOver:      opts.OnMouseOver,
		OnMouseOut:       opts.OnMouseOut,
		OnMouseScroll:    opts.OnMouseScroll,
		OnPaste:          opts.OnPaste,
		OnKeyDown:        opts.OnKeyDown,
	})
	r := &TextBufferRenderable{
		Renderable: base,
		options:    opts,
	}
	// Create native scene node
	if ctx != nil && ctx.NativeScene != nil {
		node, err := ctx.NativeScene.CreateNode(r.Renderable, native.SceneNodeKindText)
		if err == nil {
			r.Renderable.SetNativeNode(node)
		}
	}
	return r
}

// GetBuffer returns the current text buffer.
func (r *TextBufferRenderable) GetBuffer() *core.TextBuffer {
	return r.options.Buffer
}

// SetBuffer sets the text buffer and requests a re-render.
func (r *TextBufferRenderable) SetBuffer(buf *core.TextBuffer) {
	r.options.Buffer = buf
	r.RequestRender()
}

// GetWrap returns whether wrapping is enabled.
func (r *TextBufferRenderable) GetWrap() bool {
	return r.options.Wrap
}

// SetWrap enables or disables wrapping and requests a re-render.
func (r *TextBufferRenderable) SetWrap(wrap bool) {
	r.options.Wrap = wrap
	r.RequestRender()
}

// RenderSelf renders the text buffer into the buffer with line wrapping.
func (r *TextBufferRenderable) RenderSelf(buffer *core.NativeBuffer, deltaTime float64) {
	if r.options.RenderBefore != nil {
		r.options.RenderBefore(r.Renderable, buffer, deltaTime)
	}

	buf := r.options.Buffer
	if buf == nil {
		if r.options.RenderAfter != nil {
			r.options.RenderAfter(r.Renderable, buffer, deltaTime)
		}
		return
	}

	x := int(r.GetX())
	y := int(r.GetY())
	w := int(r.GetWidth())
	h := int(r.GetHeight())

	if w < 1 || h < 1 {
		if r.options.RenderAfter != nil {
			r.options.RenderAfter(r.Renderable, buffer, deltaTime)
		}
		return
	}

	// Clamp to buffer bounds.
	if x < 0 {
		w += x
		x = 0
	}
	if y < 0 {
		h += y
		y = 0
	}
	if x+w > buffer.Width() {
		w = buffer.Width() - x
	}
	if y+h > buffer.Height() {
		h = buffer.Height() - y
	}
	if w <= 0 || h <= 0 {
		if r.options.RenderAfter != nil {
			r.options.RenderAfter(r.Renderable, buffer, deltaTime)
		}
		return
	}

	wrapWidth := r.options.WrapWidth
	if wrapWidth <= 0 {
		wrapWidth = w
	}

	var lines []string
	if r.options.Wrap {
		lines = wrapText(buf.GetContent(), wrapWidth)
	} else {
		lines = splitLines(buf.GetContent())
	}

	fg := r.options.ForegroundColor
	bg := r.options.BackgroundColor
	attrs := r.options.Attributes

	row := y
	for _, line := range lines {
		if row >= y+h {
			break
		}
		writeText(buffer, x, row, line, fg, bg, attrs)
		row++
	}

	if r.options.RenderAfter != nil {
		r.options.RenderAfter(r.Renderable, buffer, deltaTime)
	}
}
