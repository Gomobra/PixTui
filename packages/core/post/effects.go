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
	"math/rand"

	"github.com/Gomobra/PixTui/packages/native"
)

// PostEffectBuffer is the interface satisfied by *core.NativeBuffer. It lets
// the post package read and write buffer cells without importing the core
// package, which would create an import cycle.
type PostEffectBuffer interface {
	Width() int
	Height() int
	SetCell(x, y int, ch uint32, fg, bg uint16, attrs uint32)
	GetCell(x, y int) (uint32, uint16, uint16, uint32, bool)
	Access() native.BufferAccess
}

// PostEffect is a post-processing effect that runs on the render buffer after
// the renderable tree has been painted. deltaTime is the elapsed time since
// the previous frame, in seconds.
type PostEffect interface {
	Apply(buffer PostEffectBuffer, deltaTime float64)
}

// PostProcessor manages an ordered list of PostEffect instances and applies
// them in sequence to the render buffer.
type PostProcessor struct {
	effects []PostEffect
}

// NewPostProcessor creates an empty PostProcessor.
func NewPostProcessor() *PostProcessor {
	return &PostProcessor{}
}

// AddEffect appends an effect to the processor's list.
func (p *PostProcessor) AddEffect(e PostEffect) {
	p.effects = append(p.effects, e)
}

// RemoveEffect removes the first occurrence of the given effect.
func (p *PostProcessor) RemoveEffect(e PostEffect) {
	for i, eff := range p.effects {
		if eff == e {
			p.effects = append(p.effects[:i], p.effects[i+1:]...)
			return
		}
	}
}

// ClearEffects removes all effects from the processor.
func (p *PostProcessor) ClearEffects() {
	p.effects = nil
}

// Run applies all effects to the buffer in order, passing deltaTime to each.
func (p *PostProcessor) Run(buffer PostEffectBuffer, deltaTime float64) {
	for _, eff := range p.effects {
		eff.Apply(buffer, deltaTime)
	}
}

// ----- Built-in effects -----

// DistortionEffect applies animated glitch/distortion effects (shift, flip,
// color glitches) to random lines of the buffer.
type DistortionEffect struct {
	GlitchChancePerSecond float64
	MaxGlitchLines        int
	MinGlitchDuration     float64
	MaxGlitchDuration     float64
	MaxShiftAmount        int
	ShiftFlipRatio        float64
	ColorGlitchChance     float64

	lastGlitchTime   float64
	glitchDuration   float64
	activeGlitches   []activeGlitch
	rng              *rand.Rand
}

type activeGlitch struct {
	y      int
	typ    string
	amount int
}

func NewDistortionEffect() *DistortionEffect {
	return &DistortionEffect{
		GlitchChancePerSecond: 0.5,
		MaxGlitchLines:        3,
		MinGlitchDuration:     0.05,
		MaxGlitchDuration:     0.2,
		MaxShiftAmount:        10,
		ShiftFlipRatio:        0.6,
		ColorGlitchChance:     0.2,
		rng:                   rand.New(rand.NewSource(rand.Int63())),
	}
}

func (e *DistortionEffect) Apply(buffer PostEffectBuffer, deltaTime float64) {
	width := buffer.Width()
	height := buffer.Height()
	if width == 0 || height == 0 {
		return
	}

	e.lastGlitchTime += deltaTime

	if len(e.activeGlitches) > 0 && e.lastGlitchTime >= e.glitchDuration {
		e.activeGlitches = nil
		e.glitchDuration = 0
	}

	if len(e.activeGlitches) == 0 && e.rng.Float64() < e.GlitchChancePerSecond*deltaTime {
		e.lastGlitchTime = 0
		e.glitchDuration = e.MinGlitchDuration + e.rng.Float64()*(e.MaxGlitchDuration-e.MinGlitchDuration)
		numGlitches := 1 + e.rng.Intn(e.MaxGlitchLines)
		seen := make(map[int]bool)
		for i := 0; i < numGlitches; i++ {
			y := e.rng.Intn(height)
			if seen[y] {
				continue
			}
			seen[y] = true
			var typ string
			var amount int
			typeRoll := e.rng.Float64()
			if typeRoll < e.ColorGlitchChance {
				typ = "color"
			} else {
				shiftRoll := (typeRoll - e.ColorGlitchChance) / (1 - e.ColorGlitchChance)
				if shiftRoll < e.ShiftFlipRatio {
					typ = "shift"
					amount = e.rng.Intn(2*e.MaxShiftAmount+1) - e.MaxShiftAmount
				} else {
					typ = "flip"
				}
			}
			e.activeGlitches = append(e.activeGlitches, activeGlitch{y: y, typ: typ, amount: amount})
		}
	}

	if len(e.activeGlitches) == 0 {
		return
	}

	access := buffer.Access()
	chars := access.Char
	fgs := access.Fg
	bgs := access.Bg
	attrs := access.Attributes

	for _, g := range e.activeGlitches {
		y := g.y
		if y < 0 || y >= height {
			continue
		}
		baseIndex := y * width

		if g.typ == "shift" || g.typ == "flip" {
			tempChar := make([]uint32, width)
			tempFg := make([]uint16, width)
			tempBg := make([]uint16, width)
			tempAttr := make([]uint32, width)

			copy(tempChar, chars[baseIndex:baseIndex+width])
			copy(tempFg, fgs[baseIndex:baseIndex+width])
			copy(tempBg, bgs[baseIndex:baseIndex+width])
			copy(tempAttr, attrs[baseIndex:baseIndex+width])

			if g.typ == "shift" {
				for x := 0; x < width; x++ {
					srcX := (x - g.amount + width) % width
					destIndex := baseIndex + x
					chars[destIndex] = tempChar[srcX]
					attrs[destIndex] = tempAttr[srcX]
					fgs[destIndex] = tempFg[srcX]
					bgs[destIndex] = tempBg[srcX]
				}
			} else { // flip
				for x := 0; x < width; x++ {
					srcX := width - 1 - x
					destIndex := baseIndex + x
					chars[destIndex] = tempChar[srcX]
					attrs[destIndex] = tempAttr[srcX]
					fgs[destIndex] = tempFg[srcX]
					bgs[destIndex] = tempBg[srcX]
				}
			}
		} else if g.typ == "color" {
			glitchStart := e.rng.Intn(width)
			maxPossibleLength := width - glitchStart
			glitchLength := e.rng.Intn(maxPossibleLength) + 1
			if e.rng.Float64() < 0.2 {
				glitchLength = e.rng.Intn(width/4) + 1
			}
			if glitchLength > maxPossibleLength {
				glitchLength = maxPossibleLength
			}

			for x := glitchStart; x < glitchStart+glitchLength; x++ {
				if x >= width {
					break
				}
				destIndex := baseIndex + x

				colorMode := e.rng.Float64()
				var rFg, gFg, bFg, rBg, gBg, bBg float64
				if colorMode < 0.33 {
					rFg = e.rng.Float64()
					gFg = e.rng.Float64()
					bFg = e.rng.Float64()
					rBg = e.rng.Float64()
					gBg = e.rng.Float64()
					bBg = e.rng.Float64()
				} else if colorMode < 0.66 {
					emphasis := e.rng.Float64()
					switch {
					case emphasis < 0.25:
						rFg = e.rng.Float64()
					case emphasis < 0.5:
						gFg = e.rng.Float64()
					case emphasis < 0.75:
						bFg = e.rng.Float64()
					default:
						glitchColorRoll := e.rng.Float64()
						switch {
						case glitchColorRoll < 0.33:
							rFg, bFg = 1, 1
						case glitchColorRoll < 0.66:
							gFg, bFg = 1, 1
						default:
							rFg, gFg = 1, 1
						}
					}
					if e.rng.Float64() < 0.5 {
						rBg, gBg, bBg = 1-rFg, 1-gFg, 1-bFg
					} else {
						rBg = rFg * (e.rng.Float64()*0.5 + 0.2)
						gBg = gFg * (e.rng.Float64()*0.5 + 0.2)
						bBg = bFg * (e.rng.Float64()*0.5 + 0.2)
					}
				} else {
					if e.rng.Float64() > 0.5 {
						rFg = 1
					}
					if e.rng.Float64() > 0.5 {
						gFg = 1
					}
					if e.rng.Float64() > 0.5 {
						bFg = 1
					}
					rBg, gBg, bBg = 1-rFg, 1-gFg, 1-bFg
				}

				fgs[destIndex] = rgbToColorIndex(rFg, gFg, bFg)
				bgs[destIndex] = rgbToColorIndex(rBg, gBg, bBg)
			}
		}
	}

	// Write back modified data.
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x
			buffer.SetCell(x, y, chars[idx], fgs[idx], bgs[idx], attrs[idx])
		}
	}
}

// VignetteEffect darkens the corners of the buffer based on distance from center.
type VignetteEffect struct {
	Strength float64
}

func NewVignetteEffect(strength float64) *VignetteEffect {
	return &VignetteEffect{Strength: strength}
}

func (e *VignetteEffect) Apply(buffer PostEffectBuffer, _ float64) {
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

	centerX := float64(width) / 2
	centerY := float64(height) / 2
	maxDistSq := centerX*centerX + centerY*centerY
	if maxDistSq == 0 {
		maxDistSq = 1
	}

	for y := 0; y < height; y++ {
		dy := float64(y) - centerY
		dySq := dy * dy
		for x := 0; x < width; x++ {
			dx := float64(x) - centerX
			distSq := dx*dx + dySq
			baseAttenuation := math.Min(1, distSq/maxDistSq)
			attenuation := baseAttenuation * e.Strength
			inv := 1.0 - attenuation

			idx := y*width + x
			r, g, b := colorIndexToRGB(fgs[idx])
			fgs[idx] = rgbToColorIndex(r*inv, g*inv, b*inv)
			r, g, b = colorIndexToRGB(bgs[idx])
			bgs[idx] = rgbToColorIndex(r*inv, g*inv, b*inv)

			buffer.SetCell(x, y, chars[idx], fgs[idx], bgs[idx], attrs[idx])
		}
	}
}

// RainbowTextEffect applies animated rainbow colors to white-ish foreground cells.
type RainbowTextEffect struct {
	Speed      float64
	Saturation float64
	Value      float64
	Repeats    float64
	time       float64
}

func NewRainbowTextEffect() *RainbowTextEffect {
	return &RainbowTextEffect{
		Speed:      0.01,
		Saturation: 1.0,
		Value:      1.0,
		Repeats:    3.0,
	}
}

func (e *RainbowTextEffect) Apply(buffer PostEffectBuffer, deltaTime float64) {
	width := buffer.Width()
	height := buffer.Height()
	if width == 0 || height == 0 {
		return
	}

	e.time += deltaTime * e.Speed

	access := buffer.Access()
	chars := access.Char
	fgs := access.Fg
	bgs := access.Bg
	attrs := access.Attributes

	angleRad := 25.0 * math.Pi / 180.0
	cosAngle := math.Cos(angleRad)
	sinAngle := math.Sin(angleRad)
	maxProjection := float64(width)*cosAngle + float64(height)*sinAngle
	whiteThreshold := 0.9

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x
			r, g, b := colorIndexToRGB(fgs[idx])

			if r >= whiteThreshold && g >= whiteThreshold && b >= whiteThreshold {
				projection := float64(x)*cosAngle + float64(y)*sinAngle
				hue := math.Mod((projection/maxProjection)*e.Repeats+e.time*0.1, 1.0)
				nr, ng, nb := hsvToRgb(hue, e.Saturation, e.Value)
				fgs[idx] = rgbToColorIndex(nr, ng, nb)
			}
		}
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := y*width + x
			buffer.SetCell(x, y, chars[idx], fgs[idx], bgs[idx], attrs[idx])
		}
	}
}

// hsvToRgb converts HSV (h in [0,1], s/v in [0,1]) to RGB [0,1].
func hsvToRgb(h, s, v float64) (float64, float64, float64) {
	var r, g, b float64
	i := int(math.Floor(h * 6))
	f := h*6 - float64(i)
	p := v * (1 - s)
	q := v * (1 - f*s)
	t := v * (1 - (1-f)*s)

	switch i % 6 {
	case 0:
		r, g, b = v, t, p
	case 1:
		r, g, b = q, v, p
	case 2:
		r, g, b = p, v, t
	case 3:
		r, g, b = p, q, v
	case 4:
		r, g, b = t, p, v
	case 5:
		r, g, b = v, p, q
	}
	return r, g, b
}
