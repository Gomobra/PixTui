package renderables

import (
	"github.com/Gomobra/PixTui/packages/core"
	"github.com/Gomobra/PixTui/packages/native"
)

// InputOptions contains options for InputRenderable.
type InputOptions struct {
	core.RenderableOptions[core.Renderable]
	Value      string
	Placeholder string
}

// InputRenderable is a single-line text input.
type InputRenderable struct {
	*core.Renderable
	options InputOptions
	value   string

	cursorPos      int
	selectionStart int
	selectionEnd   int
	clipboard      *native.Clipboard
}

// NewInputRenderable creates a new InputRenderable.
func NewInputRenderable(ctx *core.RenderContext, opts InputOptions) *InputRenderable {
	base := core.NewRenderable(ctx, opts.ID, core.RenderableOptions[core.Renderable]{
		Width:   opts.Width,
		Height:  opts.Height,
		Visible: opts.Visible,
		ZIndex:  opts.ZIndex,
	})
	r := &InputRenderable{
		Renderable: base,
		options:    opts,
		value:      opts.Value,
		cursorPos:  len(opts.Value),
		clipboard:  native.NewClipboard(),
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

// Value returns the current input value.
func (r *InputRenderable) Value() string {
	return r.value
}

// SetValue sets the input value and requests a re-render.
func (r *InputRenderable) SetValue(value string) {
	r.value = value
	r.options.Value = value
	r.cursorPos = len(r.value)
	r.selectionStart = 0
	r.selectionEnd = 0
	r.RequestRender()
}

// SetCursorPosition sets the cursor position, clamped to [0, len(value)].
func (r *InputRenderable) SetCursorPosition(pos int) {
	r.cursorPos = clampInt(pos, 0, len(r.value))
	r.RequestRender()
}

// GetCursorPosition returns the cursor position.
func (r *InputRenderable) GetCursorPosition() int {
	return r.cursorPos
}

// Select sets the selection range [start, end).
func (r *InputRenderable) Select(start, end int) {
	r.selectionStart = clampInt(start, 0, len(r.value))
	r.selectionEnd = clampInt(end, 0, len(r.value))
	if r.selectionStart > r.selectionEnd {
		r.selectionStart, r.selectionEnd = r.selectionEnd, r.selectionStart
	}
	r.RequestRender()
}

// ClearSelection clears the text selection.
func (r *InputRenderable) ClearSelection() {
	r.selectionStart = 0
	r.selectionEnd = 0
	r.RequestRender()
}

// HandleKeyDown handles keyboard input for the input field.
func (r *InputRenderable) HandleKeyDown(key *core.KeyEvent) bool {
	ctrl := key.Modifiers&native.ModCtrl != 0
	shift := key.Modifiers&native.ModShift != 0
	switch key.Key {
	case "Backspace":
		if r.selectionStart != r.selectionEnd {
			r.deleteSelection()
		} else if r.cursorPos > 0 {
			r.value = r.value[:r.cursorPos-1] + r.value[r.cursorPos:]
			r.cursorPos--
		}
		r.options.Value = r.value
		r.RequestRender()
		return true
	case "Delete":
		if r.selectionStart != r.selectionEnd {
			r.deleteSelection()
		} else if r.cursorPos < len(r.value) {
			r.value = r.value[:r.cursorPos] + r.value[r.cursorPos+1:]
		}
		r.options.Value = r.value
		r.RequestRender()
		return true
	case "ArrowLeft":
		if r.cursorPos > 0 {
			r.cursorPos--
		}
		r.RequestRender()
		return true
	case "ArrowRight":
		if r.cursorPos < len(r.value) {
			r.cursorPos++
		}
		r.RequestRender()
		return true
	case "Home":
		r.cursorPos = 0
		r.RequestRender()
		return true
	case "End":
		r.cursorPos = len(r.value)
		r.RequestRender()
		return true
	case "v", "V":
		if ctrl {
			text, _, err := r.clipboard.Paste()
			if err == nil {
				r.Insert(text)
			}
			if r.options.OnPaste != nil {
				r.options.OnPaste(r.Renderable, &core.PasteEvent{Data: []byte(text)})
			}
			return true
		}
	case "Insert":
		if shift {
			text, _, err := r.clipboard.Paste()
			if err == nil {
				r.Insert(text)
			}
			if r.options.OnPaste != nil {
				r.options.OnPaste(r.Renderable, &core.PasteEvent{Data: []byte(text)})
			}
			return true
		}
	default:
		// Printable character
		if key.Key != "" {
			if r.selectionStart != r.selectionEnd {
				r.deleteSelection()
			}
			r.value = r.value[:r.cursorPos] + key.Key + r.value[r.cursorPos:]
			r.cursorPos += len(key.Key)
			r.options.Value = r.value
			r.RequestRender()
			return true
		}
	}
	return false
}

func (r *InputRenderable) deleteSelection() {
	start, end := r.selectionStart, r.selectionEnd
	if start > end {
		start, end = end, start
	}
	r.value = r.value[:start] + r.value[end:]
	r.cursorPos = start
	r.selectionStart = 0
	r.selectionEnd = 0
}

// Insert inserts text at the current cursor position.
func (r *InputRenderable) Insert(text string) {
	if r.selectionStart != r.selectionEnd {
		r.deleteSelection()
	}
	r.value = r.value[:r.cursorPos] + text + r.value[r.cursorPos:]
	r.cursorPos += len(text)
	r.options.Value = r.value
	r.RequestRender()
}

// RenderSelf renders the input into the buffer.
func (r *InputRenderable) RenderSelf(buffer *core.NativeBuffer, deltaTime float64) {
	display := r.value
	hasPlaceholder := false
	if display == "" && r.options.Placeholder != "" {
		display = r.options.Placeholder
		hasPlaceholder = true
	}

	x := 0
	y := 0
	hasSelection := r.selectionStart != r.selectionEnd

	selBg := &core.RGBA{R: 80, G: 80, B: 120, A: 255}
	placeholderFg := &core.RGBA{R: 150, G: 150, B: 150, A: 255}

	fg := r.options.ForegroundColor
	bg := r.options.BackgroundColor
	attrs := r.options.Attributes

	for i, ch := range display {
		if x >= buffer.Width() || y >= buffer.Height() {
			break
		}

		inSelection := hasSelection && !hasPlaceholder && i >= r.selectionStart && i < r.selectionEnd
		isCursor := !hasPlaceholder && i == r.cursorPos

		cellFg := fg
		cellBg := bg
		cellAttrs := attrs

		if hasPlaceholder {
			cellFg = placeholderFg
		} else if inSelection {
			cellBg = selBg
		}

		if isCursor {
			cellAttrs = cellAttrs | core.TextAttributeInverse
		}

		writeCell(buffer, x, y, uint32(ch), cellFg, cellBg, cellAttrs)
		x++
	}

	// Render cursor at end of text (block cursor) when positioned at end.
	if !hasPlaceholder && r.cursorPos >= len(r.value) && r.cursorPos < buffer.Width() {
		writeCell(buffer, r.cursorPos, y, uint32(' '), fg, bg, attrs|core.TextAttributeInverse)
	}

	if r.options.RenderAfter != nil {
		r.options.RenderAfter(r.Renderable, buffer, deltaTime)
	}
}
