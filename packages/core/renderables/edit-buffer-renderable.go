package renderables

import (
	"github.com/Gomobra/PixTui/packages/core"
	"github.com/Gomobra/PixTui/packages/native"
)

// EditBufferRenderableOptions contains options for EditBufferRenderable
type EditBufferRenderableOptions struct {
	core.RenderableOptions[core.Renderable]
	Value      string
	Placeholder string
	ReadOnly   bool
}

// EditBufferRenderable is an editable text area with undo/redo
type EditBufferRenderable struct {
	*core.Renderable
	options EditBufferRenderableOptions
	editBuf *core.EditBuffer
}

// NewEditBufferRenderable creates a new EditBufferRenderable
func NewEditBufferRenderable(ctx *core.RenderContext, opts EditBufferRenderableOptions) *EditBufferRenderable {
	base := core.NewRenderable(ctx, opts.ID, core.RenderableOptions[core.Renderable]{
		Width:   opts.Width,
		Height:  opts.Height,
		Visible: opts.Visible,
		ZIndex:  opts.ZIndex,
	})
	editBuf := core.NewEditBuffer()
	if opts.Value != "" {
		editBuf.SetContent(opts.Value)
	}
	r := &EditBufferRenderable{
		Renderable: base,
		options:    opts,
		editBuf:    editBuf,
	}
	// Create native scene node
	if ctx != nil && ctx.NativeScene != nil {
		node, err := ctx.NativeScene.CreateNode(r.Renderable, native.SceneNodeKindEditor)
		if err == nil {
			r.Renderable.SetNativeNode(node)
		}
	}
	return r
}

// Value returns the current buffer content
func (r *EditBufferRenderable) Value() string {
	return r.editBuf.GetContent()
}

// SetValue sets the buffer content and requests a re-render
func (r *EditBufferRenderable) SetValue(value string) {
	r.editBuf.SetContent(value)
	r.RequestRender()
}

// Insert inserts text at the given offset
func (r *EditBufferRenderable) Insert(offset int, text string) error {
	return r.editBuf.Insert(offset, text)
}

// Delete deletes text from start to end
func (r *EditBufferRenderable) Delete(start, end int) error {
	return r.editBuf.Delete(start, end)
}

// Undo reverts the last change
func (r *EditBufferRenderable) Undo() bool {
	return r.editBuf.Undo()
}

// Redo reapplies the last undone change
func (r *EditBufferRenderable) Redo() bool {
	return r.editBuf.Redo()
}

// RenderSelf renders the edit buffer into the buffer
func (r *EditBufferRenderable) RenderSelf(buffer *core.NativeBuffer, deltaTime float64) {
	fg := r.options.ForegroundColor
	bg := r.options.BackgroundColor
	attrs := r.options.Attributes

	content := r.editBuf.GetContent()
	y := 0
	for _, line := range splitLines(content) {
		if y >= buffer.Height() {
			break
		}
		x := 0
		for _, ch := range line {
			if x >= buffer.Width() {
				break
			}
			writeCell(buffer, x, y, uint32(ch), fg, bg, attrs)
			x++
		}
		y++
	}

	if r.options.RenderAfter != nil {
		r.options.RenderAfter(r.Renderable, buffer, deltaTime)
	}
}
