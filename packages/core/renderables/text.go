package renderables

import (
	"github.com/Gomobra/PixTui/packages/core"
	"github.com/Gomobra/PixTui/packages/native"
)

// TextOptions contains options for TextRenderable.
type TextOptions struct {
	core.RenderableOptions[core.Renderable]
	// Content is the text to render.
	Content string
	// Wrap enables wrapping text to the available width.
	Wrap bool
	// WrapWidth overrides the width used for wrapping. When 0 the renderable
	// width (or buffer width) is used.
	WrapWidth int
	// Align is one of: left, center, right.
	Align string
	// Truncate truncates lines that exceed the available width with an ellipsis.
	Truncate bool
	// Attributes are the text attributes (bold, dim, italic, ...) applied to
	// the whole text.
	Attributes uint32
}

// TextRenderable renders text content.
type TextRenderable struct {
	*core.Renderable
	options TextOptions
	content string
}

// NewTextRenderable creates a new TextRenderable.
func NewTextRenderable(ctx *core.RenderContext, opts TextOptions) *TextRenderable {
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
	r := &TextRenderable{
		Renderable: base,
		options:    opts,
		content:    opts.Content,
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

// Content returns the text content.
func (r *TextRenderable) Content() string {
	return r.content
}

// SetContent sets the text content and requests a re-render.
func (r *TextRenderable) SetContent(content string) {
	r.content = content
	r.options.Content = content
	r.RequestRender()
}

// Options returns the text options.
func (r *TextRenderable) Options() TextOptions {
	return r.options
}

// SetOptions updates the text options and requests a re-render.
func (r *TextRenderable) SetOptions(opts TextOptions) {
	r.options = opts
	r.content = opts.Content
	r.RequestRender()
}

// RenderSelf renders the text into the buffer.
func (r *TextRenderable) RenderSelf(buffer *core.NativeBuffer, deltaTime float64) {
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

	wrapWidth := r.options.WrapWidth
	if wrapWidth <= 0 {
		wrapWidth = w
	}

	var lines []string
	if r.options.Wrap {
		lines = wrapText(content, wrapWidth)
	} else {
		lines = splitLines(content)
	}

	fg := r.options.ForegroundColor
	bg := r.options.BackgroundColor
	attrs := r.options.Attributes

	row := y
	for _, line := range lines {
		if row >= y+h {
			break
		}
		display := line
		lineW := measureCellWidth(line)
		if r.options.Truncate && lineW > w {
			// Truncate with an ellipsis.
			if w > 1 {
				display = truncateToWidth(line, w-1) + "…"
			} else {
				display = ""
			}
		}

		tx := x
		switch r.options.Align {
		case "center":
			tx = x + (w-measureCellWidth(display))/2
		case "right":
			tx = x + w - measureCellWidth(display)
		}
		tx = clampInt(tx, x, x+w-measureCellWidth(display))
		if tx < x {
			tx = x
		}

		writeText(buffer, tx, row, display, fg, bg, attrs)
		row++
	}

	if r.options.RenderAfter != nil {
		r.options.RenderAfter(r.Renderable, buffer, deltaTime)
	}
}

// truncateToWidth returns the longest prefix of s that fits within width
// columns without splitting a multi-byte rune.
func truncateToWidth(s string, width int) string {
	if width <= 0 {
		return ""
	}
	w := 0
	for i, r := range s {
		if r == '\n' || r == '\r' {
			continue
		}
		if w+1 > width {
			return s[:i]
		}
		w++
	}
	return s
}
