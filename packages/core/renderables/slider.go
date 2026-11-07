package renderables

import (
	"fmt"

	"github.com/Gomobra/PixTui/packages/core"
	"github.com/Gomobra/PixTui/packages/native"
)

// SliderOptions contains options for SliderRenderable.
type SliderOptions struct {
	core.RenderableOptions[core.Renderable]
	Min, Max, Value float64
	// TrackColor is the color of the slider track (unfilled portion).
	TrackColor *core.RGBA
	// FillColor is the color of the filled portion of the track.
	FillColor *core.RGBA
	// ThumbColor is the color of the slider thumb.
	ThumbColor *core.RGBA
}

// SliderRenderable is a slider control.
type SliderRenderable struct {
	*core.Renderable
	options SliderOptions

	// Drag state for mouse interaction.
	dragging      bool
	dragStartX    int
	dragStartValue float64
}

// NewSliderRenderable creates a new SliderRenderable.
func NewSliderRenderable(ctx *core.RenderContext, opts SliderOptions) *SliderRenderable {
	base := core.NewRenderable(ctx, opts.ID, core.RenderableOptions[core.Renderable]{
		Width:   opts.Width,
		Height:  opts.Height,
		Visible: opts.Visible,
		ZIndex:  opts.ZIndex,
	})
	r := &SliderRenderable{
		Renderable: base,
		options:    opts,
	}
	if r.options.Max == 0 {
		r.options.Max = 100
	}
	// Create native scene node
	if ctx != nil && ctx.NativeScene != nil {
		node, err := ctx.NativeScene.CreateNode(r.Renderable, native.SceneNodeKindSlider)
		if err == nil {
			r.Renderable.SetNativeNode(node)
		}
	}
	return r
}

// Value returns the slider value.
func (r *SliderRenderable) Value() float64 {
	return r.options.Value
}

// SetValue sets the slider value, clamped to [Min, Max], and requests a re-render.
func (r *SliderRenderable) SetValue(v float64) {
	min := r.options.Min
	max := r.options.Max
	if max <= min {
		max = min + 1
	}
	if v < min {
		v = min
	}
	if v > max {
		v = max
	}
	r.options.Value = v
	r.RequestRender()
}

// HandleKeyDown handles keyboard input for the slider.
// ArrowLeft/Right move by a small step, Home/End jump to min/max,
// PageUp/PageDown move by a larger step.
func (r *SliderRenderable) HandleKeyDown(key *core.KeyEvent) bool {
	min := r.options.Min
	max := r.options.Max
	if max <= min {
		max = min + 1
	}
	range_ := max - min
	smallStep := range_ / 100.0
	if smallStep == 0 {
		smallStep = 1.0
	}
	largeStep := range_ / 10.0
	if largeStep == 0 {
		largeStep = 1.0
	}

	shift := key.Modifiers&native.ModShift != 0
	ctrl := key.Modifiers&native.ModCtrl != 0

	step := smallStep
	if shift {
		step = largeStep
	}
	if ctrl {
		step = range_
	}

	newValue := r.options.Value
	switch key.Key {
	case "ArrowLeft":
		newValue -= step
	case "ArrowRight":
		newValue += step
	case "Home":
		newValue = min
	case "End":
		newValue = max
	case "PageUp":
		newValue += largeStep
	case "PageDown":
		newValue -= largeStep
	default:
		return false
	}

	r.SetValue(newValue)
	return true
}

// HandleMouse handles mouse interaction for the slider thumb drag.
func (r *SliderRenderable) HandleMouse(event *core.MouseEvent) bool {
	x := int(r.GetX())
	y := int(r.GetY())
	w := int(r.GetWidth())
	h := int(r.GetHeight())

	if w < 1 || h < 1 {
		return false
	}

	// Compute thumb position for hit testing.
	min := r.options.Min
	max := r.options.Max
	val := r.options.Value
	if max <= min {
		max = min + 1
	}
	ratio := (val - min) / (max - min)
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	thumbPos := x + int(ratio*float64(w-1))

	thumbX := thumbPos
	thumbY := y
	thumbW := 1
	thumbH := h

	inThumb := event.X >= thumbX && event.X < thumbX+thumbW && event.Y >= thumbY && event.Y < thumbY+thumbH

	switch {
	case event.Button == 1 && event.Pressed:
		// Left mouse button pressed.
		if inThumb {
			r.dragging = true
			r.dragStartX = event.X
			r.dragStartValue = val
		} else {
			// Click on track: jump value to click position.
			relX := event.X - x
			if relX < 0 {
				relX = 0
			}
			if relX >= w {
				relX = w - 1
			}
			clickRatio := float64(relX) / float64(w-1)
			if clickRatio < 0 {
				clickRatio = 0
			}
			if clickRatio > 1 {
				clickRatio = 1
			}
			newValue := min + clickRatio*(max-min)
			r.SetValue(newValue)
			r.dragging = true
			r.dragStartX = event.X
			r.dragStartValue = newValue
		}
		return true

	case event.Button == 1 && !event.Pressed:
		// Left mouse button released.
		if r.dragging {
			r.dragging = false
			return true
		}
		return false

	case event.Button == 0 && event.Pressed && r.dragging:
		// Mouse move while dragging (Button 0, Pressed=true indicates motion in some backends).
		// Also handle drag via MouseMove-like events: update value from current X.
		relX := event.X - x
		if relX < 0 {
			relX = 0
		}
		if relX >= w {
			relX = w - 1
		}
		dragRatio := float64(relX) / float64(w-1)
		if dragRatio < 0 {
			dragRatio = 0
		}
		if dragRatio > 1 {
			dragRatio = 1
		}
		newValue := min + dragRatio*(max-min)
		r.SetValue(newValue)
		return true

	default:
		return false
	}
}

// RenderSelf renders the slider into the buffer.
func (r *SliderRenderable) RenderSelf(buffer *core.NativeBuffer, deltaTime float64) {
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
	fg := r.options.ForegroundColor
	bg := r.options.BackgroundColor
	attrs := r.options.Attributes

	trackColor := r.options.TrackColor
	if trackColor == nil {
		trackColor = fg
	}
	fillColor := r.options.FillColor
	if fillColor == nil {
		fillColor = fg
	}
	thumbColor := r.options.ThumbColor
	if thumbColor == nil {
		thumbColor = fg
	}

	// Calculate slider position
	min := r.options.Min
	max := r.options.Max
	val := r.options.Value
	if max <= min {
		max = min + 1
	}
	ratio := (val - min) / (max - min)
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	thumbPos := x + int(ratio*float64(w-1))

	// Draw track
	for col := x; col < x+w; col++ {
		writeCell(buffer, col, y, '─', trackColor, bg, attrs)
	}

	// Draw filled portion
	for col := x; col < thumbPos && col < x+w; col++ {
		writeCell(buffer, col, y, '━', fillColor, bg, attrs)
	}

	// Draw thumb with visual feedback when dragging.
	thumbChar := uint32('█')
	thumbAttrs := attrs
	if r.dragging {
		thumbChar = uint32('▓')
		thumbAttrs |= core.TextAttributeBold
	}
	if thumbPos >= x && thumbPos < x+w {
		writeCell(buffer, thumbPos, y, thumbChar, thumbColor, bg, thumbAttrs)
	}

	if r.options.RenderAfter != nil {
		r.options.RenderAfter(r.Renderable, buffer, deltaTime)
	}
}

var _ = fmt.Sprintf
