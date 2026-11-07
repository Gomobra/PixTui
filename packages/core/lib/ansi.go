package lib

import (
	"github.com/Gomobra/PixTui/packages/native"
)

// ANSI wraps native ANSI functions
type ANSI struct{}

func NewANSI() *ANSI {
	return &ANSI{}
}

// SGR returns the ANSI SGR sequence for colors and attributes
func (a *ANSI) SGR(fg, bg native.RGBA, attributes uint32) string {
	return native.SGR(fg, bg, attributes)
}

// Reset returns the ANSI reset sequence
func (a *ANSI) Reset() string {
	return native.Reset()
}

// RGBA wraps native RGBA
type RGBA = native.RGBA

func NewRGBA(r, g, b, a uint8) RGBA {
	return native.RGBA{R: r, G: g, B: b, A: a}
}

func RGB(r, g, b uint8) RGBA {
	return native.RGBA{R: r, G: g, B: b, A: 255}
}

var (
	Transparent = native.Transparent
	White       = native.White
	Black       = native.Black
	Red         = native.Red
	Green       = native.Green
	Blue        = native.Blue
	Yellow      = native.Yellow
	Magenta     = native.Magenta
	Cyan        = native.Cyan
)

// TextAttributes constants
const (
	TextAttributeNone          = native.TextAttributeNone
	TextAttributeBold          = native.TextAttributeBold
	TextAttributeDim           = native.TextAttributeDim
	TextAttributeItalic        = native.TextAttributeItalic
	TextAttributeUnderline     = native.TextAttributeUnderline
	TextAttributeBlink         = native.TextAttributeBlink
	TextAttributeInverse       = native.TextAttributeInverse
	TextAttributeHidden        = native.TextAttributeHidden
	TextAttributeStrikethrough = native.TextAttributeStrikethrough
)
