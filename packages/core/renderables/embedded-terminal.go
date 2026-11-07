package renderables

import (
	"github.com/Gomobra/PixTui/packages/core"
	"github.com/Gomobra/PixTui/packages/native"
)

// EmbeddedTerminalOptions contains options for EmbeddedTerminalRenderable.
type EmbeddedTerminalOptions struct {
	core.RenderableOptions[core.Renderable]
	// Width is the embedded terminal width in cells.
	Width int
	// Height is the embedded terminal height in cells.
	Height int
	// Scrollback is the number of scrollback lines (reserved for future use).
	Scrollback int
}

// EmbeddedTerminalRenderable renders an embedded terminal's output buffer
// into the parent buffer. It wraps a native.EmbeddedTerminal and mirrors the
// TypeScript EmbeddedTerminal renderable from pixtui-repo.
type EmbeddedTerminalRenderable struct {
	*core.Renderable
	terminal *native.EmbeddedTerminal
	options  EmbeddedTerminalOptions
}

// NewEmbeddedTerminalRenderable creates a new EmbeddedTerminalRenderable.
func NewEmbeddedTerminalRenderable(ctx *core.RenderContext, opts EmbeddedTerminalOptions) *EmbeddedTerminalRenderable {
	width := opts.Width
	if width <= 0 {
		width = 80
	}
	height := opts.Height
	if height <= 0 {
		height = 24
	}

	base := core.NewRenderable(ctx, opts.ID, core.RenderableOptions[core.Renderable]{
		ID:             opts.ID,
		Visible:        opts.Visible,
		ZIndex:         opts.ZIndex,
		Buffered:       opts.Buffered,
		Live:           opts.Live,
		Opacity:        opts.Opacity,
		AnimationFPS:   opts.AnimationFPS,
		ForegroundColor: opts.ForegroundColor,
		BackgroundColor: opts.BackgroundColor,
		Attributes:      opts.Attributes,
		RenderBefore:    opts.RenderBefore,
		RenderAfter:     opts.RenderAfter,
		OnMouse:         opts.OnMouse,
		OnMouseDown:     opts.OnMouseDown,
		OnMouseUp:       opts.OnMouseUp,
		OnMouseMove:     opts.OnMouseMove,
		OnMouseDrag:     opts.OnMouseDrag,
		OnMouseDragEnd:  opts.OnMouseDragEnd,
		OnMouseDrop:     opts.OnMouseDrop,
		OnMouseOver:     opts.OnMouseOver,
		OnMouseOut:      opts.OnMouseOut,
		OnMouseScroll:   opts.OnMouseScroll,
		OnPaste:         opts.OnPaste,
		OnKeyDown:       opts.OnKeyDown,
	})
	r := &EmbeddedTerminalRenderable{
		Renderable: base,
		terminal:   native.NewEmbeddedTerminal(width, height),
		options:    opts,
	}
	// Create native scene node
	if ctx != nil && ctx.NativeScene != nil {
		node, err := ctx.NativeScene.CreateNode(r.Renderable, native.SceneNodeKindCustom)
		if err == nil {
			r.Renderable.SetNativeNode(node)
		}
	}
	return r
}

// Terminal returns the underlying embedded terminal.
func (r *EmbeddedTerminalRenderable) Terminal() *native.EmbeddedTerminal {
	return r.terminal
}

// Width returns the embedded terminal width in cells.
func (r *EmbeddedTerminalRenderable) Width() int {
	return r.terminal.Width()
}

// Height returns the embedded terminal height in cells.
func (r *EmbeddedTerminalRenderable) Height() int {
	return r.terminal.Height()
}

// Write feeds data into the embedded terminal emulator.
func (r *EmbeddedTerminalRenderable) Write(data []byte) (int, error) {
	return r.terminal.Write(data)
}

// Read reads data from the embedded terminal input.
func (r *EmbeddedTerminalRenderable) Read(data []byte) (int, error) {
	return r.terminal.Read(data)
}

// Resize resizes the embedded terminal to the given dimensions.
func (r *EmbeddedTerminalRenderable) Resize(width, height int) error {
	return r.terminal.Resize(width, height)
}

// Close releases the embedded terminal resources.
func (r *EmbeddedTerminalRenderable) Close() error {
	return r.terminal.Close()
}

// RenderSelf renders the embedded terminal output into the buffer.
func (r *EmbeddedTerminalRenderable) RenderSelf(buffer *core.NativeBuffer, deltaTime float64) {
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

	// Get a snapshot of the embedded terminal's output buffer.
	access := r.terminal.Buffer().Access()

	tw := access.Width
	th := access.Height

	// Number of rows/columns we can actually copy from the terminal buffer.
	rows := h
	if rows > th {
		rows = th
	}
	cols := w
	if cols > tw {
		cols = tw
	}

	// Copy terminal cells into the render buffer, clipped to the renderable
	// region. Cells outside the terminal buffer remain as cleared (spaces).
	for row := 0; row < rows; row++ {
		ty := row
		ry := y + row
		baseIdx := ty * tw
		for col := 0; col < cols; col++ {
			tx := col
			rx := x + col
			idx := baseIdx + tx
			if idx >= len(access.Char) {
				continue
			}
			ch := access.Char[idx]
			if ch == 0 {
				ch = ' '
			}
			buffer.SetCell(rx, ry, ch, access.Fg[idx], access.Bg[idx], access.Attributes[idx])
		}
	}

	if r.options.RenderAfter != nil {
		r.options.RenderAfter(r.Renderable, buffer, deltaTime)
	}
}
