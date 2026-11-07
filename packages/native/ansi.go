package native

import (
	"fmt"
	"strings"
)

// RGBA represents a 4-component color (red, green, blue, alpha)
type RGBA struct {
	R, G, B, A uint8
}

// TextAttributes constants (bit flags)
const (
	TextAttributeNone          = 0
	TextAttributeBold          = 1 << 0
	TextAttributeDim           = 1 << 1
	TextAttributeItalic        = 1 << 2
	TextAttributeUnderline     = 1 << 3
	TextAttributeBlink         = 1 << 4
	TextAttributeInverse       = 1 << 5
	TextAttributeHidden        = 1 << 6
	TextAttributeStrikethrough = 1 << 7
)

// Default colors
var (
	Transparent = RGBA{0, 0, 0, 0}
	White       = RGBA{255, 255, 255, 255}
	Black       = RGBA{0, 0, 0, 255}
	Red         = RGBA{255, 0, 0, 255}
	Green       = RGBA{0, 255, 0, 255}
	Blue        = RGBA{0, 0, 255, 255}
	Yellow      = RGBA{255, 255, 0, 255}
	Magenta     = RGBA{255, 0, 255, 255}
	Cyan        = RGBA{0, 255, 255, 255}
)

// ANSI SGR escape sequences
const (
	ANSIReset         = "\x1b[0m"
	ANSIBold          = "\x1b[1m"
	ANSIDim           = "\x1b[2m"
	ANSIItalic        = "\x1b[3m"
	ANSIUnderline     = "\x1b[4m"
	ANSIBlink         = "\x1b[5m"
	ANSIInverse       = "\x1b[7m"
	ANSIHidden        = "\x1b[8m"
	ANSISTrikethrough = "\x1b[9m"
)

// SGR returns the ANSI SGR sequence for the given attributes and colors
func SGR(fg, bg RGBA, attributes uint32) string {
	var parts []string
	if attributes&TextAttributeBold != 0 {
		parts = append(parts, "1")
	}
	if attributes&TextAttributeDim != 0 {
		parts = append(parts, "2")
	}
	if attributes&TextAttributeItalic != 0 {
		parts = append(parts, "3")
	}
	if attributes&TextAttributeUnderline != 0 {
		parts = append(parts, "4")
	}
	if attributes&TextAttributeBlink != 0 {
		parts = append(parts, "5")
	}
	if attributes&TextAttributeInverse != 0 {
		parts = append(parts, "7")
	}
	if attributes&TextAttributeHidden != 0 {
		parts = append(parts, "8")
	}
	if attributes&TextAttributeStrikethrough != 0 {
		parts = append(parts, "9")
	}
	if fg.A > 0 {
		parts = append(parts, fmt.Sprintf("38;2;%d;%d;%d", fg.R, fg.G, fg.B))
	}
	if bg.A > 0 {
		parts = append(parts, fmt.Sprintf("48;2;%d;%d;%d", bg.R, bg.G, bg.B))
	}
	if len(parts) == 0 {
		return ANSIReset
	}
	return "\x1b[" + strings.Join(parts, ";") + "m"
}

// Reset returns the ANSI reset sequence
func Reset() string {
	return ANSIReset
}
