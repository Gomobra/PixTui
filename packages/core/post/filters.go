// Package post provides post-processing pipeline (filters, matrices, effects)
// for the OpenTUI Go port. It mirrors the TypeScript implementation found in
// packages/core/src/post/ of the original repository.
//
// To avoid an import cycle (core -> post -> core), the post package does not
// import core. Instead it operates on the PostEffectBuffer interface, which
// is satisfied by *core.NativeBuffer.
package post

import (
	"math"

)

// Filter identifies a post-processing filter.
type Filter string

const (
	// Blur applies a box-blur convolution over the buffer.
	Blur Filter = "blur"
	// Sharpen applies a sharpening convolution over the buffer.
	Sharpen Filter = "sharpen"
	// EdgeDetect applies an edge-detection convolution over the buffer.
	EdgeDetect Filter = "edge-detect"
	// Invert inverts all color channels (photographic negative).
	Invert Filter = "invert"
	// Grayscale converts colors to monochrome gray using luminance weights.
	Grayscale Filter = "grayscale"
	// Sepia applies a sepia tone transformation.
	Sepia Filter = "sepia"
	// Brightness adds an additive offset to all RGB channels.
	Brightness Filter = "brightness"
	// Contrast scales color channels around the midpoint.
	Contrast Filter = "contrast"
)

// toU8 clamps a normalized [0,1] value to a byte [0,255].
func toU8(value float64) uint8 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 255
	}
	return uint8(math.Round(value * 255))
}

// colorIndexToRGB approximates the reverse of rgbaTo256 (core/renderables/helpers.go),
// converting a 256-color ANSI index to normalized RGB [0,1].
func colorIndexToRGB(idx uint16) (r, g, b float64) {
	switch {
	case idx == 0:
		return 0, 0, 0
	case idx == 231:
		return 1, 1, 1
	case idx >= 232 && idx <= 255:
		// Grayscale ramp 232-255.
		gray := float64(idx-232) * 255.0 / 23.0
		v := gray / 255.0
		return v, v, v
	case idx >= 16 && idx <= 231:
		// 6x6x6 color cube.
		n := idx - 16
		rIdx := n / 36
		rem := n % 36
		gIdx := rem / 6
		bIdx := rem % 6
		return float64(rIdx) * 255.0 / 5.0 / 255.0,
			float64(gIdx) * 255.0 / 5.0 / 255.0,
			float64(bIdx) * 255.0 / 5.0 / 255.0
	default:
		// Standard ANSI 0-15 approximate.
		switch idx {
		case 0:
			return 0, 0, 0
		case 1:
			return 0.5, 0, 0
		case 2:
			return 0, 0.5, 0
		case 3:
			return 0.5, 0.5, 0
		case 4:
			return 0, 0, 0.5
		case 5:
			return 0.5, 0, 0.5
		case 6:
			return 0, 0.5, 0.5
		case 7:
			return 0.75, 0.75, 0.75
		case 8:
			return 0.25, 0.25, 0.25
		case 9:
			return 1, 0, 0
		case 10:
			return 0, 1, 0
		case 11:
			return 1, 1, 0
		case 12:
			return 0, 0, 1
		case 13:
			return 1, 0, 1
		case 14:
			return 0, 1, 1
		case 15:
			return 1, 1, 1
		}
		return 0, 0, 0
	}
}

// rgbToColorIndex approximates rgbaTo256, converting normalized RGB [0,1] to
// a 256-color ANSI index.
func rgbToColorIndex(r, g, b float64) uint16 {
	r8 := toU8(r)
	g8 := toU8(g)
	b8 := toU8(b)

	// Grayscale ramp.
	if r8 == g8 && g8 == b8 {
		if r8 < 8 {
			return 0
		}
		if r8 > 247 {
			return 231
		}
		return 232 + uint16((int(r8)*24)/255)
	}
	// 6x6x6 color cube.
	rIdx := uint16(int(r8) * 5 / 255)
	gIdx := uint16(int(g8) * 5 / 255)
	bIdx := uint16(int(b8) * 5 / 255)
	return 16 + 36*rIdx + 6*gIdx + bIdx
}

// clamp01 clamps v to [0,1].
func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// ApplyFilter applies the named filter to the buffer. params are optional
// filter-specific arguments (e.g. strength for brightness/contrast).
func ApplyFilter(buffer PostEffectBuffer, filter Filter, params ...float64) {
	width := buffer.Width()
	height := buffer.Height()
	if width == 0 || height == 0 {
		return
	}

	access := buffer.Access()
	chars := access.Char
	fgs := access.Fg
	bgs := access.Bg
	attrs := access.Attributes

	switch filter {
	case Blur:
		applyConvolution(buffer, width, height, chars, fgs, bgs, attrs,
			[3][3]float64{
				{1 / 9.0, 1 / 9.0, 1 / 9.0},
				{1 / 9.0, 1 / 9.0, 1 / 9.0},
				{1 / 9.0, 1 / 9.0, 1 / 9.0},
			})
	case Sharpen:
		applyConvolution(buffer, width, height, chars, fgs, bgs, attrs,
			[3][3]float64{
				{0, -1, 0},
				{-1, 5, -1},
				{0, -1, 0},
			})
	case EdgeDetect:
		applyConvolution(buffer, width, height, chars, fgs, bgs, attrs,
			[3][3]float64{
				{-1, -1, -1},
				{-1, 8, -1},
				{-1, -1, -1},
			})
	case Invert:
		strength := 1.0
		if len(params) > 0 {
			strength = params[0]
		}
		applyInvert(buffer, width, height, chars, fgs, bgs, attrs, strength)
	case Grayscale:
		applyGrayscale(buffer, width, height, chars, fgs, bgs, attrs)
	case Sepia:
		applySepia(buffer, width, height, chars, fgs, bgs, attrs)
	case Brightness:
		brightness := 0.0
		if len(params) > 0 {
			brightness = params[0]
		}
		applyBrightness(buffer, width, height, chars, fgs, bgs, attrs, brightness)
	case Contrast:
		contrast := 1.0
		if len(params) > 0 {
			contrast = params[0]
		}
		applyContrast(buffer, width, height, chars, fgs, bgs, attrs, contrast)
	}
}

// applyConvolution applies a 3x3 kernel to fg and bg of every cell, using
// luminance for the convolution and writing back the resulting grayscale.
func applyConvolution(buffer PostEffectBuffer, width, height int, chars []uint32, fgs, bgs []uint16, attrs []uint32, kernel [3][3]float64) {
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x
			// Sample from the ORIGINAL snapshots (chars/fgs/bgs), but write
			// into separate accumulators so we don't pollute subsequent reads.
			r := convolveChannel(fgs, bgs, width, height, x, y, 1, kernel)
			g := convolveChannel(fgs, bgs, width, height, x, y, 1, kernel)
			b := convolveChannel(fgs, bgs, width, height, x, y, 1, kernel)
			fgs[idx] = rgbToColorIndex(r, g, b)
			r = convolveChannel(fgs, bgs, width, height, x, y, 2, kernel)
			g = convolveChannel(fgs, bgs, width, height, x, y, 2, kernel)
			b = convolveChannel(fgs, bgs, width, height, x, y, 2, kernel)
			bgs[idx] = rgbToColorIndex(r, g, b)
			buffer.SetCell(x, y, chars[idx], fgs[idx], bgs[idx], attrs[idx])
		}
	}
}

// convolveChannel samples the 3x3 neighborhood around (x, y) using the given
// kernel and returns the resulting normalized channel value [0,1].
func convolveChannel(fgs, bgs []uint16, width, height, x, y int, target uint16, kernel [3][3]float64) float64 {
	var sum float64
	for ky := -1; ky <= 1; ky++ {
		for kx := -1; kx <= 1; kx++ {
			nx := x + kx
			ny := y + ky
			if nx < 0 || nx >= width || ny < 0 || ny >= height {
				continue
			}
			idx := ny*width + nx
			var val uint16
			if target == 1 {
				val = fgs[idx]
			} else {
				val = bgs[idx]
			}
			r, g, b := colorIndexToRGB(val)
			lum := 0.299*r + 0.587*g + 0.114*b
			sum += lum * kernel[ky+1][kx+1]
		}
	}
	return clamp01(sum)
}

// applyInvert inverts RGB channels: out = strength * (1 - in) + (1 - strength) * in.
func applyInvert(buffer PostEffectBuffer, width, height int, chars []uint32, fgs, bgs []uint16, attrs []uint32, strength float64) {
	strength = clamp01(strength)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x
			for _, arr := range []*[]uint16{&fgs, &bgs} {
				r, g, b := colorIndexToRGB((*arr)[idx])
				nr := 1.0 - r
				ng := 1.0 - g
				nb := 1.0 - b
				r = r + strength*(nr-r)
				g = g + strength*(ng-g)
				b = b + strength*(nb-b)
				(*arr)[idx] = rgbToColorIndex(r, g, b)
			}
			buffer.SetCell(x, y, chars[idx], fgs[idx], bgs[idx], attrs[idx])
		}
	}
}

// applyGrayscale converts to luminance-based gray.
func applyGrayscale(buffer PostEffectBuffer, width, height int, chars []uint32, fgs, bgs []uint16, attrs []uint32) {
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x
			for _, arr := range []*[]uint16{&fgs, &bgs} {
				r, g, b := colorIndexToRGB((*arr)[idx])
				lum := 0.299*r + 0.587*g + 0.114*b
				(*arr)[idx] = rgbToColorIndex(lum, lum, lum)
			}
			buffer.SetCell(x, y, chars[idx], fgs[idx], bgs[idx], attrs[idx])
		}
	}
}

// applySepia applies the standard sepia matrix.
func applySepia(buffer PostEffectBuffer, width, height int, chars []uint32, fgs, bgs []uint16, attrs []uint32) {
	m := Matrix{
		{0.393, 0.769, 0.189, 0},
		{0.349, 0.686, 0.168, 0},
		{0.272, 0.534, 0.131, 0},
		{0, 0, 0, 1},
	}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x
			for _, arr := range []*[]uint16{&fgs, &bgs} {
				r, g, b := colorIndexToRGB((*arr)[idx])
				nr := m[0][0]*r + m[0][1]*g + m[0][2]*b
				ng := m[1][0]*r + m[1][1]*g + m[1][2]*b
				nb := m[2][0]*r + m[2][1]*g + m[2][2]*b
				(*arr)[idx] = rgbToColorIndex(clamp01(nr), clamp01(ng), clamp01(nb))
			}
			buffer.SetCell(x, y, chars[idx], fgs[idx], bgs[idx], attrs[idx])
		}
	}
}

// applyBrightness adds an additive brightness offset to all channels.
func applyBrightness(buffer PostEffectBuffer, width, height int, chars []uint32, fgs, bgs []uint16, attrs []uint32, brightness float64) {
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x
			for _, arr := range []*[]uint16{&fgs, &bgs} {
				r, g, b := colorIndexToRGB((*arr)[idx])
				(*arr)[idx] = rgbToColorIndex(clamp01(r+brightness), clamp01(g+brightness), clamp01(b+brightness))
			}
			buffer.SetCell(x, y, chars[idx], fgs[idx], bgs[idx], attrs[idx])
		}
	}
}

// applyContrast scales channels around the midpoint 0.5.
func applyContrast(buffer PostEffectBuffer, width, height int, chars []uint32, fgs, bgs []uint16, attrs []uint32, contrast float64) {
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x
			for _, arr := range []*[]uint16{&fgs, &bgs} {
				r, g, b := colorIndexToRGB((*arr)[idx])
				r = 0.5 + contrast*(r-0.5)
				g = 0.5 + contrast*(g-0.5)
				b = 0.5 + contrast*(b-0.5)
				(*arr)[idx] = rgbToColorIndex(clamp01(r), clamp01(g), clamp01(b))
			}
			buffer.SetCell(x, y, chars[idx], fgs[idx], bgs[idx], attrs[idx])
		}
	}
}
