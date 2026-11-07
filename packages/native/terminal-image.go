package native

import (
	"compress/flate"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"io"
	"math"
)

// KittyPixelFormat mirrors terminal-image.zig KittyPixelFormat.
type KittyPixelFormat int

const (
	KittyPixelFormatAuto KittyPixelFormat = iota
	KittyPixelFormatRGB
	KittyPixelFormatRGBA
)

// KittyEncoding mirrors terminal-image.zig KittyEncoding.
type KittyEncoding uint32

const (
	KittyEncodingRaw KittyEncoding = iota
	KittyEncodingZlib
	KittyEncodingPNG
	KittyEncodingFile
)

// KITTY_PREPARE_BYTES_MAX bounds optional preparation independently of the
// image decoder's limits. Mirrors terminal-image.zig KITTY_PREPARE_BYTES_MAX.
const KITTY_PREPARE_BYTES_MAX = 64 * 1024 * 1024

// KittyPayloadSize returns the encoded payload size for an image.
// Mirrors terminal-image.zig kittyPayloadSize.
func KittyPayloadSize(img *Image) (int, error) {
	if img.EncodedPNG != nil {
		return len(img.EncodedPNG), nil
	}
	count, err := safeMul(img.Width, img.Height)
	if err != nil {
		return 0, err
	}
	bytesPerPixel := 3
	if img.HasAlpha {
		bytesPerPixel = 4
	}
	return safeMul(count, bytesPerPixel)
}

// WriteKittyPayload writes the kitty payload (PNG or raw RGB/RGBA pixels)
// to the writer. Mirrors terminal-image.zig writeKittyPayload.
func WriteKittyPayload(w io.Writer, img *Image) error {
	if img.EncodedPNG != nil {
		_, err := w.Write(img.EncodedPNG)
		return err
	}
	pixels, err := ensurePixels(img)
	if err != nil {
		return err
	}
	count, err := safeMul(img.Width, img.Height)
	if err != nil {
		return err
	}
	if len(pixels) != count*4 {
		return fmt.Errorf("invalid image data")
	}
	if img.HasAlpha {
		_, err := w.Write(pixels)
		return err
	}
	// Convert RGBA to RGB
	var rgb [3072]byte
	offset := 0
	for offset < count {
		n := min(count-offset, len(rgb)/3)
		for i := 0; i < n; i++ {
			src := (offset + i) * 4
			copy(rgb[i*3:][:3], pixels[src:][:3])
		}
		if _, err := w.Write(rgb[:n*3]); err != nil {
			return err
		}
		offset += n
	}
	return nil
}

// WriteKittyTransmitEncoded writes a kitty transmit escape sequence, optionally
// compressing with zlib. Returns the encoding used.
// Mirrors terminal-image.zig writeKittyTransmitEncoded.
func WriteKittyTransmitEncoded(alloc *Allocator, w io.Writer, img *Image, id uint32, tmux bool, compress bool) (KittyEncoding, error) {
	fallback := KittyEncodingRaw
	if img.EncodedPNG != nil {
		fallback = KittyEncodingPNG
	}
	if compress && fallback == KittyEncodingRaw {
		size, err := KittyPayloadSize(img)
		if err != nil {
			return fallback, err
		}
		if size > 8 && size <= KITTY_PREPARE_BYTES_MAX {
			bytes := alloc.Alloc(size)
			if bytes != nil {
				defer alloc.Free(bytes)
				var compBuf bytesBuffer
				compBuf.buf = bytes
				compressor, err := flate.NewWriter(&compBuf, flate.BestSpeed)
				if err == nil {
					if err := WriteKittyPayload(compressor, img); err == nil {
						if err := compressor.Close(); err == nil {
							payload := compBuf.written()
							base64Size := base64.StdEncoding.EncodedLen(len(payload))
							if base64Size+len(",o=z") < base64.StdEncoding.EncodedLen(size) {
								offset := 0
								for offset < len(payload) {
									end := min(offset+3072, len(payload))
									if tmux {
										if _, err := w.Write([]byte("\x1bPtmux;\x1b\x1b_G")); err != nil {
											return fallback, err
										}
									} else {
										if _, err := w.Write([]byte("\x1b_G")); err != nil {
											return fallback, err
										}
									}
									if offset == 0 {
										if _, err := fmt.Fprintf(w, "a=t,f=%d,s=%d,v=%d,i=%d,m=%d,q=2;",
											boolToInt(!img.HasAlpha)+24, img.Width, img.Height, id,
											boolToInt(end < len(payload))); err != nil {
											return fallback, err
										}
									} else {
										if _, err := fmt.Fprintf(w, "m=%d,q=2;", boolToInt(end < len(payload))); err != nil {
											return fallback, err
										}
									}
									encoded := [4096]byte{}
									encLen := base64.StdEncoding.EncodedLen(end - offset)
									payloadEnc := Base64Encode(encoded[:encLen], payload[offset:end])
									if _, err := w.Write(payloadEnc); err != nil {
										return fallback, err
									}
									if tmux {
										if _, err := w.Write([]byte("\x1b\x1b\\\x1b\\")); err != nil {
											return fallback, err
										}
									} else {
										if _, err := w.Write([]byte("\x1b\\")); err != nil {
											return fallback, err
										}
									}
									offset = end
								}
								return KittyEncodingZlib, nil
							}
						}
					}
				}
			}
		}
	}
	if err := WriteKittyTransmit(w, img, id, tmux); err != nil {
		return fallback, err
	}
	return fallback, nil
}

// WriteKittyTransmit writes a kitty transmit escape sequence for an image.
// Mirrors terminal-image.zig writeKittyTransmit.
func WriteKittyTransmit(w io.Writer, img *Image, id uint32, tmux bool) error {
	return WriteKittyTransmitFormat(w, img, id, tmux, KittyPixelFormatAuto)
}

// WriteKittyTransmitFormat writes a kitty transmit escape sequence with a
// requested pixel format. Mirrors terminal-image.zig writeKittyTransmitFormat.
func WriteKittyTransmitFormat(w io.Writer, img *Image, id uint32, tmux bool, requestedFormat KittyPixelFormat) error {
	const rawChunk = 3072
	pixelCount, err := safeMul(img.Width, img.Height)
	if err != nil {
		return err
	}
	rgbaLen, err := safeMul(pixelCount, 4)
	if err != nil {
		return err
	}

	if requestedFormat == KittyPixelFormatAuto {
		if img.EncodedPNG != nil {
			png := img.EncodedPNG
			offset := 0
			first := true
			for offset < len(png) {
				end := min(offset+rawChunk, len(png))
				more := end < len(png)
				if tmux {
					if _, err := w.Write([]byte("\x1bPtmux;\x1b\x1b_G")); err != nil {
						return err
					}
				} else {
					if _, err := w.Write([]byte("\x1b_G")); err != nil {
						return err
					}
				}
				if first {
					if _, err := fmt.Fprintf(w, "a=t,f=100,i=%d,m=%d,q=2;", id, boolToInt(more)); err != nil {
						return err
					}
				} else {
					if _, err := fmt.Fprintf(w, "m=%d,q=2;", boolToInt(more)); err != nil {
						return err
					}
				}
				encoded := [4096]byte{}
				encLen := base64.StdEncoding.EncodedLen(end - offset)
				payload := Base64Encode(encoded[:encLen], png[offset:end])
				if _, err := w.Write(payload); err != nil {
					return err
				}
				if tmux {
					if _, err := w.Write([]byte("\x1b\x1b\\\x1b\\")); err != nil {
						return err
					}
				} else {
					if _, err := w.Write([]byte("\x1b\\")); err != nil {
						return err
					}
				}
				offset = end
				first = false
			}
			return nil
		}
	}

	pixels, err := ensurePixels(img)
	if err != nil {
		return err
	}
	if len(pixels) < rgbaLen {
		return fmt.Errorf("invalid image data")
	}

	format := requestedFormat
	if format == KittyPixelFormatAuto {
		if img.HasAlpha {
			format = KittyPixelFormatRGBA
		} else {
			format = KittyPixelFormatRGB
		}
	}

	if format == KittyPixelFormatRGB {
		rgbLen, err := safeMul(pixelCount, 3)
		if err != nil {
			return err
		}
		var raw [rawChunk]byte
		rgbOffset := 0
		first := true
		for rgbOffset < rgbLen {
			chunkLen := min(rawChunk, rgbLen-rgbOffset)
			firstPixel := rgbOffset / 3
			chunkPixels := chunkLen / 3
			for index := 0; index < chunkPixels; index++ {
				source := (firstPixel + index) * 4
				copy(raw[index*3:][:3], pixels[source:][:3])
			}
			more := rgbOffset+chunkLen < rgbLen
			if tmux {
				if _, err := w.Write([]byte("\x1bPtmux;\x1b\x1b_G")); err != nil {
					return err
				}
			} else {
				if _, err := w.Write([]byte("\x1b_G")); err != nil {
					return err
				}
			}
			if first {
				if _, err := fmt.Fprintf(w, "a=t,f=24,s=%d,v=%d,i=%d,m=%d,q=2;",
					img.Width, img.Height, id, boolToInt(more)); err != nil {
					return err
				}
			} else {
				if _, err := fmt.Fprintf(w, "m=%d,q=2;", boolToInt(more)); err != nil {
					return err
				}
			}
			encoded := [4096]byte{}
			encLen := base64.StdEncoding.EncodedLen(chunkLen)
			payload := Base64Encode(encoded[:encLen], raw[:chunkLen])
			if _, err := w.Write(payload); err != nil {
				return err
			}
			if tmux {
				if _, err := w.Write([]byte("\x1b\x1b\\\x1b\\")); err != nil {
					return err
				}
			} else {
				if _, err := w.Write([]byte("\x1b\\")); err != nil {
					return err
				}
			}
			rgbOffset += chunkLen
			first = false
		}
		return nil
	}

	// RGBA
	offset := 0
	first := true
	for offset < rgbaLen {
		end := min(offset+rawChunk, rgbaLen)
		more := end < rgbaLen
		if tmux {
			if _, err := w.Write([]byte("\x1bPtmux;\x1b\x1b_G")); err != nil {
				return err
			}
		} else {
			if _, err := w.Write([]byte("\x1b_G")); err != nil {
				return err
			}
		}
		if first {
			if _, err := fmt.Fprintf(w, "a=t,f=32,s=%d,v=%d,i=%d,m=%d,q=2;",
				img.Width, img.Height, id, boolToInt(more)); err != nil {
				return err
			}
		} else {
			if _, err := fmt.Fprintf(w, "m=%d,q=2;", boolToInt(more)); err != nil {
				return err
			}
		}
		encoded := [4096]byte{}
		encLen := base64.StdEncoding.EncodedLen(end - offset)
		payload := Base64Encode(encoded[:encLen], pixels[offset:end])
		if _, err := w.Write(payload); err != nil {
			return err
		}
		if tmux {
			if _, err := w.Write([]byte("\x1b\x1b\\\x1b\\")); err != nil {
				return err
			}
		} else {
			if _, err := w.Write([]byte("\x1b\\")); err != nil {
				return err
			}
		}
		offset = end
		first = false
	}
	return nil
}

// KittyPlacementZ returns the Z placement value for a kitty image placement.
// Mirrors terminal-image.zig kittyPlacementZ.
func KittyPlacementZ(placementID uint32) int32 {
	const belowBg = math.MinInt32 / 2
	const maxZ = belowBg + int32(IMAGE_ID_MASK)
	if maxZ >= 0 || maxZ < belowBg {
		panic("kitty placement Z out of band")
	}
	if placementID > IMAGE_ID_MASK {
		panic("placement ID exceeds mask")
	}
	z := int32(belowBg) + int32(placementID)
	if z >= 0 {
		panic("kitty placement Z not negative")
	}
	return z
}

// WriteKittyPlacement writes a kitty placement escape sequence at a position.
// Mirrors terminal-image.zig writeKittyPlacement.
func WriteKittyPlacement(w io.Writer, id, placementID, x, y, width, height, sourceX, sourceY, sourceWidth, sourceHeight uint32, z int32, tmux bool) error {
	if _, err := fmt.Fprintf(w, "\x1b[%d;%dH", y+1, x+1); err != nil {
		return err
	}
	if tmux {
		if _, err := w.Write([]byte("\x1bPtmux;\x1b\x1b_G")); err != nil {
			return err
		}
	} else {
		if _, err := w.Write([]byte("\x1b_G")); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(w, "a=p,i=%d,p=%d,c=%d,r=%d,x=%d,y=%d,w=%d,h=%d,C=1,z=%d,q=2",
		id, placementID, width, height, sourceX, sourceY, sourceWidth, sourceHeight, z); err != nil {
		return err
	}
	if tmux {
		if _, err := w.Write([]byte("\x1b\x1b\\\x1b\\")); err != nil {
			return err
		}
	} else {
		if _, err := w.Write([]byte("\x1b\\")); err != nil {
			return err
		}
	}
	return nil
}

// WriteKittyPlacementAtCursor writes a kitty placement escape sequence at the
// cursor position. Mirrors terminal-image.zig writeKittyPlacementAtCursor.
func WriteKittyPlacementAtCursor(w io.Writer, id, placementID, width, height uint32, z int32, tmux bool) error {
	if tmux {
		if _, err := w.Write([]byte("\x1bPtmux;\x1b\x1b_G")); err != nil {
			return err
		}
	} else {
		if _, err := w.Write([]byte("\x1b_G")); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(w, "a=p,i=%d,p=%d,c=%d,r=%d,C=1,z=%d,q=2",
		id, placementID, width, height, z); err != nil {
		return err
	}
	if tmux {
		if _, err := w.Write([]byte("\x1b\x1b\\\x1b\\")); err != nil {
			return err
		}
	} else {
		if _, err := w.Write([]byte("\x1b\\")); err != nil {
			return err
		}
	}
	return nil
}

// WriteKittyDelete writes a kitty delete escape sequence.
// Mirrors terminal-image.zig writeKittyDelete.
func WriteKittyDelete(w io.Writer, id uint32, placementID *uint32, freeImage bool, tmux bool) error {
	if tmux {
		if _, err := w.Write([]byte("\x1bPtmux;\x1b\x1b_G")); err != nil {
			return err
		}
	} else {
		if _, err := w.Write([]byte("\x1b_G")); err != nil {
			return err
		}
	}
	freeChar := byte('i')
	if freeImage {
		freeChar = 'I'
	}
	if _, err := fmt.Fprintf(w, "a=d,d=%c,i=%d", freeChar, id); err != nil {
		return err
	}
	if placementID != nil {
		if _, err := fmt.Fprintf(w, ",p=%d", *placementID); err != nil {
			return err
		}
	}
	if _, err := w.Write([]byte(",q=2")); err != nil {
		return err
	}
	if tmux {
		if _, err := w.Write([]byte("\x1b\x1b\\\x1b\\")); err != nil {
			return err
		}
	} else {
		if _, err := w.Write([]byte("\x1b\\")); err != nil {
			return err
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Sixel support
// ---------------------------------------------------------------------------

// SixelColor is an RGB color.
type SixelColor struct {
	R, G, B uint8
}

// SixelMoment accumulates moments for Wu's color quantizer.
type SixelMoment struct {
	Count   uint64
	R, G, B uint64
	Square  uint64
}

const (
	sixelHistogramSide = 33
	sixelHistogramLen  = sixelHistogramSide * sixelHistogramSide * sixelHistogramSide
)

// QuantizedSixel holds the result of sixel color quantization.
type QuantizedSixel struct {
	allocator  *Allocator
	palette    [255][3]uint8
	paletteLen int
	indices    []uint8
}

// Deinit frees the QuantizedSixel resources. Mirrors terminal-image.zig deinit.
func (q *QuantizedSixel) Deinit() {
	if q.allocator != nil && q.indices != nil {
		q.allocator.Free(q.indices)
		q.indices = nil
	}
}

func momentIndex(r, g, b int) int {
	return (r*sixelHistogramSide+g)*sixelHistogramSide + b
}

func addMoment(target *SixelMoment, value SixelMoment) {
	target.Count += value.Count
	target.R += value.R
	target.G += value.G
	target.B += value.B
	target.Square += value.Square
}

type sixelBox struct {
	R0, R1, G0, G1, B0, B1 uint8
}

func volumeStats(moments []SixelMoment, box sixelBox) (count, r, g, b uint64) {
	corners := [8]*SixelMoment{
		&moments[momentIndex(int(box.R1), int(box.G1), int(box.B1))],
		&moments[momentIndex(int(box.R1), int(box.G1), int(box.B0))],
		&moments[momentIndex(int(box.R1), int(box.G0), int(box.B1))],
		&moments[momentIndex(int(box.R1), int(box.G0), int(box.B0))],
		&moments[momentIndex(int(box.R0), int(box.G1), int(box.B1))],
		&moments[momentIndex(int(box.R0), int(box.G1), int(box.B0))],
		&moments[momentIndex(int(box.R0), int(box.G0), int(box.B1))],
		&moments[momentIndex(int(box.R0), int(box.G0), int(box.B0))],
	}
	fields := [4]func(*SixelMoment) uint64{
		func(m *SixelMoment) uint64 { return m.Count },
		func(m *SixelMoment) uint64 { return m.R },
		func(m *SixelMoment) uint64 { return m.G },
		func(m *SixelMoment) uint64 { return m.B },
	}
	var result [4]uint64
	for i, f := range fields {
		var v int64
		v += int64(f(corners[0]))
		v -= int64(f(corners[1]))
		v -= int64(f(corners[2]))
		v += int64(f(corners[3]))
		v -= int64(f(corners[4]))
		v += int64(f(corners[5]))
		v += int64(f(corners[6]))
		v -= int64(f(corners[7]))
		result[i] = uint64(v)
	}
	return result[0], result[1], result[2], result[3]
}

func squareValue(moments []SixelMoment, box sixelBox) uint64 {
	corners := [8]*SixelMoment{
		&moments[momentIndex(int(box.R1), int(box.G1), int(box.B1))],
		&moments[momentIndex(int(box.R1), int(box.G1), int(box.B0))],
		&moments[momentIndex(int(box.R1), int(box.G0), int(box.B1))],
		&moments[momentIndex(int(box.R1), int(box.G0), int(box.B0))],
		&moments[momentIndex(int(box.R0), int(box.G1), int(box.B1))],
		&moments[momentIndex(int(box.R0), int(box.G1), int(box.B0))],
		&moments[momentIndex(int(box.R0), int(box.G0), int(box.B1))],
		&moments[momentIndex(int(box.R0), int(box.G0), int(box.B0))],
	}
	var v int64
	v += int64(corners[0].Square)
	v -= int64(corners[1].Square)
	v -= int64(corners[2].Square)
	v += int64(corners[3].Square)
	v -= int64(corners[4].Square)
	v += int64(corners[5].Square)
	v += int64(corners[6].Square)
	v -= int64(corners[7].Square)
	return uint64(v)
}

func statsEnergy(count, r, g, b uint64) float64 {
	if count == 0 {
		return 0
	}
	rf := float64(r)
	gf := float64(g)
	bf := float64(b)
	cf := float64(count)
	return (2.0*rf*rf + 4.0*gf*gf + 3.0*bf*bf) / (8.0 * cf)
}

func boxVariance(moments []SixelMoment, box sixelBox) float64 {
	count, r, g, b := volumeStats(moments, box)
	square := squareValue(moments, box)
	energy := statsEnergy(count, r, g, b)
	return math.Max(0, float64(square)-energy)
}

type sixelAxis int

const (
	sixelAxisR sixelAxis = iota
	sixelAxisG
	sixelAxisB
)

type sixelCut struct {
	axis  sixelAxis
	at    uint8
	score float64
}

func bestBoxCut(moments []SixelMoment, box sixelBox) *sixelCut {
	var best *sixelCut
	axes := []sixelAxis{sixelAxisR, sixelAxisG, sixelAxisB}
	for _, axis := range axes {
		var low, high uint8
		switch axis {
		case sixelAxisR:
			low, high = box.R0, box.R1
		case sixelAxisG:
			low, high = box.G0, box.G1
		case sixelAxisB:
			low, high = box.B0, box.B1
		}
		at := low + 1
		for at < high {
			first, second := box, box
			switch axis {
			case sixelAxisR:
				first.R1 = at
				second.R0 = at
			case sixelAxisG:
				first.G1 = at
				second.G0 = at
			case sixelAxisB:
				first.B1 = at
				second.B0 = at
			}
			fsCount, fsR, fsG, fsB := volumeStats(moments, first)
			ssCount, ssR, ssG, ssB := volumeStats(moments, second)
			if fsCount == 0 || ssCount == 0 {
				at++
				continue
			}
			score := statsEnergy(fsCount, fsR, fsG, fsB) + statsEnergy(ssCount, ssR, ssG, ssB)
			if best == nil || score > best.score {
				best = &sixelCut{axis: axis, at: at, score: score}
			}
			at++
		}
	}
	return best
}

// QuantizeSixel quantizes an image to a sixel palette. Mirrors terminal-image.zig quantizeSixel.
func QuantizeSixel(alloc *Allocator, img *Image, maxColors int) (*QuantizedSixel, error) {
	if maxColors == 0 || maxColors > 255 {
		return nil, fmt.Errorf("invalid argument")
	}
	pixelCount, err := safeMul(img.Width, img.Height)
	if err != nil {
		return nil, err
	}
	pixels, err := ensurePixels(img)
	if err != nil {
		return nil, err
	}
	if len(pixels) < pixelCount*4 {
		return nil, fmt.Errorf("invalid image data")
	}

	moments := make([]SixelMoment, sixelHistogramLen)
	var activeBins int
	for pixel := 0; pixel < pixelCount; pixel++ {
		offset := pixel * 4
		if pixels[offset+3] < 128 {
			continue
		}
		r := int(pixels[offset])>>3 + 1
		g := int(pixels[offset+1])>>3 + 1
		b := int(pixels[offset+2])>>3 + 1
		bin := &moments[momentIndex(r, g, b)]
		if bin.Count == 0 {
			activeBins++
		}
		bin.Count++
		bin.R += uint64(pixels[offset])
		bin.G += uint64(pixels[offset+1])
		bin.B += uint64(pixels[offset+2])
		bin.Square += (2*uint64(pixels[offset])*uint64(pixels[offset]) +
			4*uint64(pixels[offset+1])*uint64(pixels[offset+1]) +
			3*uint64(pixels[offset+2])*uint64(pixels[offset+2]) + 4) >> 3
	}

	// Build 3D prefix sums (integral histogram)
	for r := 1; r < sixelHistogramSide; r++ {
		var area [sixelHistogramSide]SixelMoment
		for g := 1; g < sixelHistogramSide; g++ {
			var line SixelMoment
			for b := 1; b < sixelHistogramSide; b++ {
				addMoment(&line, moments[momentIndex(r, g, b)])
				addMoment(&area[b], line)
				cumulative := moments[momentIndex(r-1, g, b)]
				addMoment(&cumulative, area[b])
				moments[momentIndex(r, g, b)] = cumulative
			}
		}
	}

	indices := alloc.Alloc(pixelCount)
	if indices == nil {
		return nil, fmt.Errorf("allocation failed")
	}
	result := &QuantizedSixel{
		allocator:  alloc,
		paletteLen: 0,
		indices:    indices,
	}

	if activeBins == 0 {
		for i := range indices {
			indices[i] = 255
		}
		return result, nil
	}

	var boxes [255]sixelBox
	var variances [255]float64
	boxes[0] = sixelBox{}
	variances[0] = boxVariance(moments, boxes[0])
	boxCount := 1

	maxBoxes := min(activeBins, maxColors)
	for boxCount < maxBoxes {
		var splitIndex int
		var highest float64
		for i := 0; i < boxCount; i++ {
			if variances[i] > highest {
				highest = variances[i]
				splitIndex = i
			}
		}
		if highest == 0 {
			break
		}
		cut := bestBoxCut(moments, boxes[splitIndex])
		if cut == nil {
			variances[splitIndex] = 0
			continue
		}
		second := boxes[splitIndex]
		switch cut.axis {
		case sixelAxisR:
			boxes[splitIndex].R1 = cut.at
			second.R0 = cut.at
		case sixelAxisG:
			boxes[splitIndex].G1 = cut.at
			second.G0 = cut.at
		case sixelAxisB:
			boxes[splitIndex].B1 = cut.at
			second.B0 = cut.at
		}
		boxes[boxCount] = second
		variances[splitIndex] = boxVariance(moments, boxes[splitIndex])
		variances[boxCount] = boxVariance(moments, second)
		boxCount++
	}

	result.paletteLen = boxCount

	// Sort boxes by count descending (common colors get shorter palette selectors)
	boxCounts := make([]uint64, boxCount)
	for i := 0; i < boxCount; i++ {
		c, _, _, _ := volumeStats(moments, boxes[i])
		boxCounts[i] = c
	}
	for i := 1; i < boxCount; i++ {
		box := boxes[i]
		count := boxCounts[i]
		destination := i
		for destination > 0 && boxCounts[destination-1] < count {
			boxes[destination] = boxes[destination-1]
			boxCounts[destination] = boxCounts[destination-1]
			destination--
		}
		boxes[destination] = box
		boxCounts[destination] = count
	}

	var tags [32 * 32 * 32]uint8
	for paletteIndex := 0; paletteIndex < boxCount; paletteIndex++ {
		box := boxes[paletteIndex]
		c, r, g, b := volumeStats(moments, box)
		result.palette[paletteIndex] = [3]uint8{
			uint8((r + c/2) / c),
			uint8((g + c/2) / c),
			uint8((b + c/2) / c),
		}
		for rr := box.R0; rr < box.R1; rr++ {
			for gg := box.G0; gg < box.G1; gg++ {
				for bb := box.B0; bb < box.B1; bb++ {
					tags[(int(rr)<<10)|(int(gg)<<5)|int(bb)] = uint8(paletteIndex)
				}
			}
		}
	}

	for y := 0; y < img.Height; y++ {
		for x := 0; x < img.Width; x++ {
			pixel := y*img.Width + x
			offset := pixel * 4
			if pixels[offset+3] < 128 {
				result.indices[pixel] = 255
			} else {
				tagIdx := ((int(pixels[offset]) >> 3) << 10) |
					((int(pixels[offset+1]) >> 3) << 5) |
					(int(pixels[offset+2]) >> 3)
				result.indices[pixel] = tags[tagIdx]
			}
		}
	}

	return result, nil
}

// WriteSixelPayload quantizes and writes a sixel payload to the writer.
// Mirrors terminal-image.zig writeSixelPayload.
func WriteSixelPayload(alloc *Allocator, w io.Writer, img *Image) error {
	quantized, err := QuantizeSixel(alloc, img, 255)
	if err != nil {
		return err
	}
	defer quantized.Deinit()
	return WriteSixelIndexedPayload(alloc, w, quantized.indices,
		quantized.palette[:quantized.paletteLen], uint32(img.Width), uint32(img.Height))
}

// WriteSixelIndexedPayload writes an indexed sixel payload to the writer.
// Mirrors terminal-image.zig writeSixelIndexedPayload.
func WriteSixelIndexedPayload(alloc *Allocator, w io.Writer, indices []uint8, palette [][3]uint8, width, height uint32) error {
	if len(palette) == 0 || len(palette) > 255 {
		return fmt.Errorf("invalid image data")
	}
	pixelCount, err := safeMul(int(width), int(height))
	if err != nil {
		return err
	}
	if len(indices) < pixelCount {
		return fmt.Errorf("invalid image data")
	}
	maskCount, err := safeMul(len(palette), int(width))
	if err != nil {
		return err
	}
	masks := alloc.Alloc(maskCount)
	if masks == nil {
		return fmt.Errorf("allocation failed")
	}
	defer alloc.Free(masks)

	bufWriter := newBufferedWriter(w)

	if _, err := bufWriter.WriteString("0;1;0q\"1;1;"); err != nil {
		return err
	}
	if err := writeUnsigned(bufWriter, int(width)); err != nil {
		return err
	}
	if err := bufWriter.WriteByte(';'); err != nil {
		return err
	}
	if err := writeUnsigned(bufWriter, int(height)); err != nil {
		return err
	}
	for index, color := range palette {
		if err := bufWriter.WriteByte('#'); err != nil {
			return err
		}
		if err := writeUnsigned(bufWriter, index); err != nil {
			return err
		}
		if _, err := bufWriter.WriteString(";2;"); err != nil {
			return err
		}
		for channelIndex, channel := range color {
			if err := writeUnsigned(bufWriter, (int(channel)*100+127)/255); err != nil {
				return err
			}
			if channelIndex < 2 {
				if err := bufWriter.WriteByte(';'); err != nil {
					return err
				}
			}
		}
	}

	outputHeight := int(height)
	bandY := 0
	var generation uint32
	var generations [256]uint32
	var lastNonzero [256]int

	for bandY < outputHeight {
		generation++
		for bit := 0; bit < 6; bit++ {
			y := bandY + bit
			if y >= outputHeight {
				continue
			}
			for x := 0; x < int(width); x++ {
				paletteIndex := int(indices[y*int(width)+x])
				if paletteIndex == 255 {
					continue
				}
				if paletteIndex >= len(palette) {
					return fmt.Errorf("invalid image data")
				}
				if generations[paletteIndex] != generation {
					generations[paletteIndex] = generation
					start := paletteIndex * int(width)
					for i := 0; i < int(width); i++ {
						masks[start+i] = 0
					}
					lastNonzero[paletteIndex] = 0
				}
				masks[paletteIndex*int(width)+x] |= uint8(1 << bit)
				if x+1 > lastNonzero[paletteIndex] {
					lastNonzero[paletteIndex] = x + 1
				}
			}
		}
		firstPlane := true
		for paletteIndex := 0; paletteIndex < len(palette); paletteIndex++ {
			if generations[paletteIndex] != generation {
				continue
			}
			planeLen := lastNonzero[paletteIndex]
			if planeLen == 0 {
				continue
			}
			plane := masks[paletteIndex*int(width) : paletteIndex*int(width)+planeLen]
			if !firstPlane {
				if err := bufWriter.WriteByte('$'); err != nil {
					return err
				}
			}
			firstPlane = false
			if err := bufWriter.WriteByte('#'); err != nil {
				return err
			}
			if err := writeUnsigned(bufWriter, paletteIndex); err != nil {
				return err
			}
			x := 0
			for x < len(plane) {
				mask := plane[x]
				char := '?' + mask
				run := 1
				for x+run < len(plane) && plane[x+run] == mask {
					run++
				}
				if run >= 4 {
					if err := bufWriter.WriteByte('!'); err != nil {
						return err
					}
					if err := writeUnsigned(bufWriter, run); err != nil {
						return err
					}
					if err := bufWriter.WriteByte(char); err != nil {
						return err
					}
				} else {
					for i := 0; i < run; i++ {
						if err := bufWriter.WriteByte(char); err != nil {
							return err
						}
					}
				}
				x += run
			}
		}
		if bandY+6 < outputHeight {
			if err := bufWriter.WriteByte('-'); err != nil {
				return err
			}
		}
		bandY += 6
	}
	return bufWriter.Flush()
}

// WriteSixelFramedPayload writes a sixel payload framed with escape sequences.
// Mirrors terminal-image.zig writeSixelFramedPayload.
func WriteSixelFramedPayload(w io.Writer, payload []byte, tmux bool) error {
	if tmux {
		if _, err := w.Write([]byte("\x1bPtmux;\x1b\x1bP")); err != nil {
			return err
		}
	} else {
		if _, err := w.Write([]byte("\x1bP")); err != nil {
			return err
		}
	}
	if _, err := w.Write(payload); err != nil {
		return err
	}
	if tmux {
		if _, err := w.Write([]byte("\x1b\x1b\\\x1b\\")); err != nil {
			return err
		}
	} else {
		if _, err := w.Write([]byte("\x1b\\")); err != nil {
			return err
		}
	}
	return nil
}

// WriteSixel quantizes and writes a sixel image to the writer.
// Mirrors terminal-image.zig writeSixel.
func WriteSixel(alloc *Allocator, w io.Writer, img *Image, tmux bool) error {
	var payload bytesBuffer
	payload.buf = make([]byte, 0, 4096)
	if err := WriteSixelPayload(alloc, &payload, img); err != nil {
		return err
	}
	return WriteSixelFramedPayload(w, payload.buf, tmux)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func safeMul(a, b int) (int, error) {
	if a == 0 || b == 0 {
		return 0, nil
	}
	result := a * b
	if a != 0 && result/a != b {
		return 0, fmt.Errorf("integer overflow")
	}
	return result, nil
}

func ensurePixels(img *Image) ([]uint8, error) {
	if len(img.Pixels) > 0 {
		return img.Pixels, nil
	}
	if img.EncodedPNG != nil {
		return decodePNGBytes(img.EncodedPNG)
	}
	return nil, fmt.Errorf("no pixel data")
}

func decodePNGBytes(data []byte) ([]uint8, error) {
	img, err := png.Decode(&bytesReader{data: data})
	if err != nil {
		return nil, err
	}
	rgba, ok := img.(*image.RGBA)
	if !ok {
		bounds := img.Bounds()
		rgba = image.NewRGBA(bounds)
		// Draw the decoded image into RGBA
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				rgba.Set(x, y, img.At(x, y))
			}
		}
	}
	return rgba.Pix, nil
}

type bytesBuffer struct {
	buf []byte
}

func (b *bytesBuffer) Write(p []byte) (int, error) {
	b.buf = append(b.buf, p...)
	return len(p), nil
}

func (b *bytesBuffer) written() []byte {
	return b.buf
}

type bytesReader struct {
	data []byte
	pos  int
}

func (r *bytesReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

type bufferedWriter struct {
	w   io.Writer
	buf []byte
}

func newBufferedWriter(w io.Writer) *bufferedWriter {
	return &bufferedWriter{w: w, buf: make([]byte, 0, 8192)}
}

func (b *bufferedWriter) Write(p []byte) (int, error) {
	b.buf = append(b.buf, p...)
	if len(b.buf) >= cap(b.buf) {
		if err := b.Flush(); err != nil {
			return len(p), err
		}
	}
	return len(p), nil
}

func (b *bufferedWriter) WriteByte(c byte) error {
	b.buf = append(b.buf, c)
	if len(b.buf) >= cap(b.buf) {
		return b.Flush()
	}
	return nil
}

func (b *bufferedWriter) WriteString(s string) (int, error) {
	b.buf = append(b.buf, s...)
	if len(b.buf) >= cap(b.buf) {
		if err := b.Flush(); err != nil {
			return len(s), err
		}
	}
	return len(s), nil
}

func (b *bufferedWriter) Flush() error {
	if len(b.buf) == 0 {
		return nil
	}
	_, err := b.w.Write(b.buf)
	b.buf = b.buf[:0]
	return err
}

func writeUnsigned(w io.Writer, value int) error {
	var buffer [20]byte
	index := len(buffer)
	remaining := value
	for {
		index--
		buffer[index] = byte('0' + remaining%10)
		remaining /= 10
		if remaining == 0 {
			break
		}
	}
	_, err := w.Write(buffer[index:])
	return err
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

var _ = image.Image(nil)
