package renderables

import (
	"strconv"

	"github.com/Gomobra/PixTui/packages/core"
	"github.com/Gomobra/PixTui/packages/native"
)

// FrameBufferOptions contains options for FrameBufferRenderable.
type FrameBufferOptions struct {
	core.RenderableOptions[core.Renderable]
	// RespectAlpha blends the buffer's cells into the parent. When true,
	// empty/transparent cells are skipped so the parent shows through. When
	// false the cells are copied verbatim, overwriting the parent.
	RespectAlpha bool
}

// FrameBufferRenderable renders into its own internal offscreen buffer and
// composites that buffer into the parent buffer during paint.
type FrameBufferRenderable struct {
	*core.Renderable
	options      FrameBufferOptions
	frameBuffer  *core.NativeBuffer
	respectAlpha bool
}

// NewFrameBufferRenderable creates a new FrameBufferRenderable. The internal
// buffer is sized from the Width/Height options (parsed as integers).
func NewFrameBufferRenderable(ctx *core.RenderContext, opts FrameBufferOptions) *FrameBufferRenderable {
	width, height := frameBufferDimensions(opts)
	base := core.NewRenderable(ctx, opts.ID, core.RenderableOptions[core.Renderable]{
		Width:   opts.Width,
		Height:  opts.Height,
		Visible: opts.Visible,
		ZIndex:  opts.ZIndex,
	})
	r := &FrameBufferRenderable{
		Renderable:   base,
		options:      opts,
		frameBuffer:  core.NewNativeBuffer(width, height),
		respectAlpha: opts.RespectAlpha,
	}
	r.SetWidth(float32(width))
	r.SetHeight(float32(height))

	if ctx != nil && ctx.NativeScene != nil {
		if node, err := ctx.NativeScene.CreateNode(r.Renderable, native.SceneNodeKindBox); err == nil {
			r.Renderable.SetNativeNode(node)
		}
	}
	return r
}

// frameBufferDimensions parses the Width/Height option strings into integers,
// defaulting to 1x1 when they are absent or unparseable.
func frameBufferDimensions(opts FrameBufferOptions) (int, int) {
	width := 1
	if opts.Width != nil {
		if n, err := strconv.Atoi(*opts.Width); err == nil && n > 0 {
			width = n
		}
	}
	height := 1
	if opts.Height != nil {
		if n, err := strconv.Atoi(*opts.Height); err == nil && n > 0 {
			height = n
		}
	}
	return width, height
}

// FrameBuffer returns the internal offscreen buffer.
func (r *FrameBufferRenderable) FrameBuffer() *core.NativeBuffer {
	return r.frameBuffer
}

// RespectAlpha returns whether alpha blending is respected when compositing.
func (r *FrameBufferRenderable) RespectAlpha() bool {
	return r.respectAlpha
}

// SetRespectAlpha sets whether alpha blending is respected when compositing.
func (r *FrameBufferRenderable) SetRespectAlpha(v bool) {
	r.respectAlpha = v
	r.options.RespectAlpha = v
	r.RequestRender()
}

// Options returns the frame buffer options.
func (r *FrameBufferRenderable) Options() FrameBufferOptions {
	return r.options
}

// OnResize resizes the internal buffer to the renderable's current reported
// dimensions and requests a re-render.
func (r *FrameBufferRenderable) OnResize() {
	width := int(r.GetWidth())
	height := int(r.GetHeight())
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	r.frameBuffer = core.NewNativeBuffer(width, height)
	r.SetWidth(float32(width))
	r.SetHeight(float32(height))
	r.RequestRender()
}

// compositeBuffer copies src into dst at (dx, dy). When respectAlpha is true,
// empty/transparent cells (a space with default colors and no attributes) are
// skipped so the destination shows through instead of being overwritten.
func compositeBuffer(dst, src *core.NativeBuffer, dx, dy int, respectAlpha bool) {
	if dst == nil || src == nil {
		return
	}
	srcAccess := src.Access()
	dw := dst.Width()
	dh := dst.Height()
	sw := srcAccess.Width
	sh := srcAccess.Height
	for y := 0; y < sh; y++ {
		dstY := dy + y
		if dstY < 0 || dstY >= dh {
			continue
		}
		rowBase := y * sw
		for x := 0; x < sw; x++ {
			dstX := dx + x
			if dstX < 0 || dstX >= dw {
				continue
			}
			idx := rowBase + x
			ch := srcAccess.Char[idx]
			fg := srcAccess.Fg[idx]
			bg := srcAccess.Bg[idx]
			attrs := srcAccess.Attributes[idx]
			if respectAlpha && ch == ' ' && fg == 0 && bg == 0 && attrs == 0 {
				continue
			}
			dst.SetCell(dstX, dstY, ch, fg, bg, attrs)
		}
	}
}

// RenderSelf composites the internal offscreen buffer into the parent buffer at
// the renderable's screen position.
func (r *FrameBufferRenderable) RenderSelf(buffer *core.NativeBuffer, deltaTime float64) {
	if r.options.RenderBefore != nil {
		r.options.RenderBefore(r.Renderable, buffer, deltaTime)
	}
	if r.IsDestroyed() || !r.IsVisible() || r.frameBuffer == nil {
		if r.options.RenderAfter != nil {
			r.options.RenderAfter(r.Renderable, buffer, deltaTime)
		}
		return
	}

	// Fill background with the renderable's background color before compositing.
	x := int(r.GetX())
	y := int(r.GetY())
	w := int(r.GetWidth())
	h := int(r.GetHeight())
	if w > 0 && h > 0 {
		bg := r.options.BackgroundColor
		if bg != nil {
			fillRect(buffer, x, y, w, h, ' ', nil, bg, 0)
		}
	}

	compositeBuffer(buffer, r.frameBuffer, int(r.GetX()), int(r.GetY()), r.respectAlpha)
	if r.options.RenderAfter != nil {
		r.options.RenderAfter(r.Renderable, buffer, deltaTime)
	}
}
