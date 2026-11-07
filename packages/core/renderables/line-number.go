package renderables

import (
	"fmt"

	"github.com/Gomobra/PixTui/packages/core"
)

// LineNumberRenderable renders line numbers for text
type LineNumberRenderable struct {
	*core.Renderable
	startLine   int
	renderAfter func(*core.Renderable, *core.NativeBuffer, float64)

	foregroundColor *core.RGBA
	backgroundColor *core.RGBA
	attributes      uint32
}

// NewLineNumberRenderable creates a new LineNumberRenderable
func NewLineNumberRenderable(ctx *core.RenderContext, id string, startLine int) *LineNumberRenderable {
	base := core.NewRenderable(ctx, id, core.RenderableOptions[core.Renderable]{})
	return &LineNumberRenderable{
		Renderable:  base,
		startLine:   startLine,
	}
}

// SetRenderAfter sets the render-after hook
func (r *LineNumberRenderable) SetRenderAfter(fn func(*core.Renderable, *core.NativeBuffer, float64)) {
	r.renderAfter = fn
}

// SetForegroundColor sets the foreground color for line numbers.
func (r *LineNumberRenderable) SetForegroundColor(c *core.RGBA) {
	r.foregroundColor = c
	r.RequestRender()
}

// SetBackgroundColor sets the background color for line numbers.
func (r *LineNumberRenderable) SetBackgroundColor(c *core.RGBA) {
	r.backgroundColor = c
	r.RequestRender()
}

// SetAttributes sets the text attributes for line numbers.
func (r *LineNumberRenderable) SetAttributes(attrs uint32) {
	r.attributes = attrs
	r.RequestRender()
}

// RenderSelf renders line numbers into the buffer
func (r *LineNumberRenderable) RenderSelf(buffer *core.NativeBuffer, deltaTime float64) {
	fg := r.foregroundColor
	bg := r.backgroundColor
	attrs := r.attributes

	for i := 0; i < buffer.Height(); i++ {
		lineNum := r.startLine + i
		str := fmt.Sprintf("%4d  ", lineNum)
		cx := 0
		for _, ch := range str {
			if i >= buffer.Height() {
				break
			}
			writeCell(buffer, cx, i, uint32(ch), fg, bg, attrs)
			cx++
		}
	}

	if r.renderAfter != nil {
		r.renderAfter(r.Renderable, buffer, deltaTime)
	}
}
