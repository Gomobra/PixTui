package renderables

import (
	"fmt"

	"github.com/Gomobra/PixTui/packages/core"
	"github.com/Gomobra/PixTui/packages/native"
)

// VideoOptions contains options for VideoRenderable
type VideoOptions struct {
	core.RenderableOptions[core.Renderable]
	Video   *native.Video
	AutoPlay bool
}

// VideoRenderable renders a video in the terminal
type VideoRenderable struct {
	*core.Renderable
	options VideoOptions
	video   *native.Video
}

// NewVideoRenderable creates a new VideoRenderable
func NewVideoRenderable(ctx *core.RenderContext, opts VideoOptions) *VideoRenderable {
	base := core.NewRenderable(ctx, opts.ID, core.RenderableOptions[core.Renderable]{
		Width:   opts.Width,
		Height:  opts.Height,
		Visible: opts.Visible,
		ZIndex:  opts.ZIndex,
	})
	vid := opts.Video
	if vid == nil {
		vid = native.NewVideo(80, 24)
	}
	r := &VideoRenderable{
		Renderable: base,
		options:    opts,
		video:      vid,
	}
	if r.options.AutoPlay {
		r.video.Playing = true
	}
	// Create native scene node
	if ctx != nil && ctx.NativeScene != nil {
		node, err := ctx.NativeScene.CreateNode(r.Renderable, native.SceneNodeKindVideo)
		if err == nil {
			r.Renderable.SetNativeNode(node)
		}
	}
	return r
}

// Video returns the underlying video
func (r *VideoRenderable) Video() *native.Video {
	return r.video
}

// Play starts video playback
func (r *VideoRenderable) Play() {
	r.video.Playing = true
}

// Pause pauses video playback
func (r *VideoRenderable) Pause() {
	r.video.Playing = false
}

// writeTextAt writes text starting at (x, y), wrapping at the right edge.
func (r *VideoRenderable) writeTextAt(buffer *core.NativeBuffer, x, y int, text string, fg, bg *core.RGBA, attrs uint32) {
	cx := x
	for _, ch := range text {
		if ch == '\n' {
			cx = x
			y++
			continue
		}
		if cx >= buffer.Width() || y >= buffer.Height() {
			return
		}
		writeCell(buffer, cx, y, uint32(ch), fg, bg, attrs)
		cx++
	}
}

// RenderSelf renders the video frame into the buffer
func (r *VideoRenderable) RenderSelf(buffer *core.NativeBuffer, deltaTime float64) {
	if r.options.RenderBefore != nil {
		r.options.RenderBefore(r.Renderable, buffer, deltaTime)
	}

	if r.video == nil {
		if r.options.RenderAfter != nil {
			r.options.RenderAfter(r.Renderable, buffer, deltaTime)
		}
		return
	}

	// Update video frame if playing (Update checks path internally)
	if r.video.Playing {
		_ = r.video.Update(deltaTime)
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

	// Fill background with the renderable's background color.
	bg := r.options.BackgroundColor
	if bg != nil {
		fillRect(buffer, x, y, w, h, ' ', nil, bg, 0)
	}

	// Render frame if available
	if r.video.Frame != nil && r.video.Frame.Width > 0 && r.video.Frame.Height > 0 && len(r.video.Frame.Pixels) > 0 {
		r.renderVideoFrame(buffer, x, y, w, h)
	} else {
		// Placeholder when no frame
		status := "[Video: Paused]"
		if r.video.Playing {
			status = "[Video: Playing]"
		}
		r.writeTextAt(buffer, x, y, status, nil, nil, 0)

		// Show time info if there is room and duration is known
		if r.video.Duration > 0 && y+1 < buffer.Height() {
			timeStr := fmt.Sprintf("%.1fs / %.1fs", r.video.CurrentTime, r.video.Duration)
			r.writeTextAt(buffer, x, y+1, timeStr, nil, nil, 0)
		}
	}

	if r.options.RenderAfter != nil {
		r.options.RenderAfter(r.Renderable, buffer, deltaTime)
	}
}

// renderVideoFrame renders the current video frame using half-block rendering
// (same approach as Image renderBlocks), with time info on the bottom line.
func (r *VideoRenderable) renderVideoFrame(buffer *core.NativeBuffer, x, y, w, h int) {
	img := r.video.Frame
	if img == nil || len(img.Pixels) == 0 {
		return
	}

	imgW := img.Width
	imgH := img.Height
	pixels := img.Pixels

	// Scale factors mapping terminal cells to source pixels.
	sx := float64(imgW) / float64(w)
	sy := float64(imgH) / float64(h*2) // each cell covers 2 pixel rows

	for row := 0; row < h; row++ {
		screenY := y + row
		if screenY >= buffer.Height() {
			break
		}
		for col := 0; col < w; col++ {
			screenX := x + col
			if screenX >= buffer.Width() {
				break
			}
			// Source pixel region covered by this cell.
			px0 := int(float64(col) * sx)
			px1 := int(float64(col+1) * sx)
			py0 := int(float64(row) * sy * 2)
			py1 := int(float64(row+1) * sy * 2)
			if px1 > imgW {
				px1 = imgW
			}
			if py1 > imgH {
				py1 = imgH
			}
			if px0 >= px1 || py0 >= py1 {
				writeCell(buffer, screenX, screenY, ' ', nil, nil, 0)
				continue
			}
			// Average the top half and bottom half pixel rows separately so we
			// can emit ▄ (upper) / ▀ (lower) / █ (full) with color.
			upper := averagePixel(pixels, imgW, px0, px1, py0, (py0+py1)/2)
			lower := averagePixel(pixels, imgW, px0, px1, (py0+py1)/2, py1)
			ch, upperEmpty, lowerEmpty := blockChar(upper, lower)
			var fg, bg *core.RGBA
			if !upperEmpty && !lowerEmpty {
				// Both halves filled: full block, color from upper.
				ch = '█'
				fg = rgbaPtr(upper)
			} else if !upperEmpty && lowerEmpty {
				ch = '▄'
				fg = rgbaPtr(upper)
				bg = rgbaPtr(core.RGBA{R: 0, G: 0, B: 0, A: 0})
			} else if upperEmpty && !lowerEmpty {
				ch = '▀'
				bg = rgbaPtr(lower)
			} else {
				writeCell(buffer, screenX, screenY, ' ', nil, nil, 0)
				continue
			}
			writeCell(buffer, screenX, screenY, ch, fg, bg, 0)
		}
	}

	// Show time info on bottom line if there is room and duration is known
	if r.video.Duration > 0 && h > 1 {
		timeStr := fmt.Sprintf("%.1fs / %.1fs", r.video.CurrentTime, r.video.Duration)
		ty := y + h - 1
		if ty < buffer.Height() {
			r.writeTextAt(buffer, x, ty, timeStr, nil, nil, 0)
		}
	}
}
