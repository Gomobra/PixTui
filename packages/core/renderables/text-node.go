package renderables

import (
	"github.com/Gomobra/PixTui/packages/core"
	"github.com/Gomobra/PixTui/packages/native"
)

// TextNodeOptions contains options for TextNodeRenderable.
type TextNodeOptions struct {
	core.RenderableOptions[core.Renderable]
	// Content is the text content to render.
	Content string
	// Style is the text style applied to the content.
	Style *core.TextStyle
}

// TextNodeRenderable renders a text node with styling.
type TextNodeRenderable struct {
	*core.Renderable
	options TextNodeOptions
	content string
	style   *core.TextStyle
}

// NewTextNodeRenderable creates a new TextNodeRenderable.
func NewTextNodeRenderable(ctx *core.RenderContext, opts TextNodeOptions) *TextNodeRenderable {
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
	style := opts.Style
	if style == nil {
		style = &core.TextStyle{}
	}
	r := &TextNodeRenderable{
		Renderable: base,
		options:    opts,
		content:    opts.Content,
		style:      style,
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

// GetContent returns the text content.
func (r *TextNodeRenderable) GetContent() string {
	return r.content
}

// SetContent sets the text content and requests a re-render.
func (r *TextNodeRenderable) SetContent(content string) {
	r.content = content
	r.options.Content = content
	r.RequestRender()
}

// GetStyle returns the current text style.
func (r *TextNodeRenderable) GetStyle() *core.TextStyle {
	return r.style
}

// SetStyle sets the text style and requests a re-render.
func (r *TextNodeRenderable) SetStyle(style *core.TextStyle) {
	if style == nil {
		style = &core.TextStyle{}
	}
	r.style = style
	r.options.Style = style
	r.RequestRender()
}

// RenderSelf renders the text node into the buffer with styling.
func (r *TextNodeRenderable) RenderSelf(buffer *core.NativeBuffer, deltaTime float64) {
	if r.options.RenderBefore != nil {
		r.options.RenderBefore(r.Renderable, buffer, deltaTime)
	}

	content := r.content
	if content == "" {
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

	style := r.style
	if style == nil {
		style = &core.TextStyle{}
	}

	fg := style.FG
	if fg == nil {
		fg = r.options.ForegroundColor
	}
	bg := style.BG
	if bg == nil {
		bg = r.options.BackgroundColor
	}
	attrs := style.Attributes
	if attrs == 0 {
		attrs = r.options.Attributes
	}

	lines := splitLines(content)
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
