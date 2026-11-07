package renderables

import (
	"github.com/Gomobra/PixTui/packages/core"
	"github.com/Gomobra/PixTui/packages/native"
)

// BoxOptions contains options for BoxRenderable.
// All color fields use core.RGBA for full RGBA styling parity with TypeScript.
type BoxOptions struct {
	core.RenderableOptions[core.Renderable]
	BackgroundColor    *core.RGBA
	BorderStyle        string
	Border             bool
	BorderColor        *core.RGBA
	ShouldFill         bool
	Title              string
	TitleColor         *core.RGBA
	TitleAlignment     string
	BottomTitle        string
	Focusable          bool
	FocusedBorderColor *core.RGBA
}

// BoxRenderable is a box container with optional border and title.
type BoxRenderable struct {
	*core.Renderable
	options BoxOptions
}

// NewBoxRenderable creates a new BoxRenderable.
func NewBoxRenderable(ctx *core.RenderContext, opts BoxOptions) *BoxRenderable {
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
	r := &BoxRenderable{
		Renderable: base,
		options:    opts,
	}
	if r.options.BackgroundColor == nil {
		transparent := core.RGBA{R: 0, G: 0, B: 0, A: 0}
		r.options.BackgroundColor = &transparent
	}
	if r.options.BorderStyle == "" {
		r.options.BorderStyle = "single"
	}
	if r.options.BorderColor == nil {
		white := core.RGBA{R: 255, G: 255, B: 255, A: 255}
		r.options.BorderColor = &white
	}
	if r.options.TitleAlignment == "" {
		r.options.TitleAlignment = "left"
	}
	if r.options.FocusedBorderColor == nil {
		focused := core.RGBA{R: 0, G: 170, B: 255, A: 255}
		r.options.FocusedBorderColor = &focused
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

// RenderSelf renders the box into the buffer.
func (r *BoxRenderable) RenderSelf(buffer *core.NativeBuffer, deltaTime float64) {
	if r.options.RenderBefore != nil {
		r.options.RenderBefore(r.Renderable, buffer, deltaTime)
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

	// Clamp to buffer bounds
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

	// Resolve colors: per-option RGBA takes precedence over renderable defaults.
	bgColor := r.options.BackgroundColor
	if bgColor == nil {
		transparent := core.RGBA{R: 0, G: 0, B: 0, A: 0}
		bgColor = &transparent
	}

	borderColor := r.options.BorderColor
	if borderColor == nil {
		white := core.RGBA{R: 255, G: 255, B: 255, A: 255}
		borderColor = &white
	}

	fg := r.options.ForegroundColor
	attrs := r.options.Attributes

	// Fill background
	if r.options.ShouldFill {
		fillRect(buffer, x, y, w, h, ' ', nil, bgColor, attrs)
	}

	// Draw border
	if r.options.Border && w >= 2 && h >= 2 {
		borderChars := r.BorderStyleChars()
		topLeft := borderChars[0]
		topRight := borderChars[1]
		bottomLeft := borderChars[2]
		bottomRight := borderChars[3]
		horiz := borderChars[4]
		vert := borderChars[5]

		// Top and bottom borders
		for col := x + 1; col < x+w-1; col++ {
			writeCell(buffer, col, y, horiz, borderColor, nil, attrs)
			writeCell(buffer, col, y+h-1, horiz, borderColor, nil, attrs)
		}
		// Left and right borders
		for row := y + 1; row < y+h-1; row++ {
			writeCell(buffer, x, row, vert, borderColor, nil, attrs)
			writeCell(buffer, x+w-1, row, vert, borderColor, nil, attrs)
		}
		// Corners
		writeCell(buffer, x, y, topLeft, borderColor, nil, attrs)
		writeCell(buffer, x+w-1, y, topRight, borderColor, nil, attrs)
		writeCell(buffer, x, y+h-1, bottomLeft, borderColor, nil, attrs)
		writeCell(buffer, x+w-1, y+h-1, bottomRight, borderColor, nil, attrs)

		// Title
		if r.options.Title != "" && h > 1 {
			titleX := x + 1
			switch r.options.TitleAlignment {
			case "center":
				titleX = x + (w - len(r.options.Title))/2
			case "right":
				titleX = x + w - len(r.options.Title) - 1
			}
			if titleX < x+1 {
				titleX = x + 1
			}
			if titleX+len(r.options.Title) > x+w-1 {
				titleX = x + w - len(r.options.Title) - 1
			}
			titleFg := fg
			if r.options.TitleColor != nil {
				titleFg = r.options.TitleColor
			}
			for i, ch := range r.options.Title {
				if titleX+i >= x+w-1 {
					break
				}
				writeCell(buffer, titleX+i, y, uint32(ch), titleFg, bgColor, attrs)
			}
		}
	}

	if r.options.RenderAfter != nil {
		r.options.RenderAfter(r.Renderable, buffer, deltaTime)
	}
}

// BorderStyleChars returns border characters for the current style.
func (r *BoxRenderable) BorderStyleChars() [6]uint32 {
	chars := [6]uint32{
		'┌', '┐', '└', '┘', '─', '│', // single
	}
	switch r.options.BorderStyle {
	case "double":
		chars = [6]uint32{'╔', '╗', '╚', '╝', '═', '║'}
	case "rounded":
		chars = [6]uint32{'╭', '╮', '╰', '╯', '─', '│'}
	case "heavy":
		chars = [6]uint32{'┏', '┓', '┗', '┛', '━', '┃'}
	}
	return chars
}

// GetOptions returns the box options.
func (r *BoxRenderable) GetOptions() BoxOptions {
	return r.options
}
