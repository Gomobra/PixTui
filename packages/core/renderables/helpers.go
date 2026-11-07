package renderables

import (
	"errors"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/Gomobra/PixTui/packages/core"
)

// errInvalidHex is returned when a CSS-style hex color cannot be parsed.
var errInvalidHex = errors.New("invalid hex color")

// clampInt clamps v to the inclusive range [lo, hi].
func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// clampFloat64 clamps v to the inclusive range [lo, hi].
func clampFloat64(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// measureCellWidth returns the approximate number of terminal columns s occupies.
// Tabs are treated as 4 spaces; newlines do not advance the column.
func measureCellWidth(s string) int {
	width := 0
	for _, r := range s {
		switch r {
		case '\t':
			width += 4
		case '\n', '\r':
			// do not advance column width
		default:
			if r == utf8.RuneError {
				width++
				continue
			}
			width++
		}
	}
	return width
}

// parseHex parses a CSS-style hex color "#RRGGBB" or "#RRGGBBAA" into a
// core.RGBA. It returns transparent black when the value cannot be parsed.
func parseHex(s string) core.RGBA {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "#") {
		s = s[1:]
	}
	switch len(s) {
	case 6: // #RRGGBB
		r, err1 := hexByte(s[0:2])
		g, err2 := hexByte(s[2:4])
		b, err3 := hexByte(s[4:6])
		if err1 == nil && err2 == nil && err3 == nil {
			return core.RGBA{R: r, G: g, B: b, A: 255}
		}
	case 8: // #RRGGBBAA
		r, err1 := hexByte(s[0:2])
		g, err2 := hexByte(s[2:4])
		b, err3 := hexByte(s[4:6])
		a, err4 := hexByte(s[6:8])
		if err1 == nil && err2 == nil && err3 == nil && err4 == nil {
			return core.RGBA{R: r, G: g, B: b, A: a}
		}
	}
	return core.RGBA{R: 0, G: 0, B: 0, A: 0}
}

// ParseHex is the exported version of parseHex. It parses a CSS-style hex color
// "#RRGGBB" or "#RRGGBBAA" into a core.RGBA. It returns transparent black when
// the value cannot be parsed.
func ParseHex(s string) core.RGBA {
	return parseHex(s)
}

// hexByte parses exactly two hexadecimal digits into a uint8.
func hexByte(s string) (uint8, error) {
	if len(s) != 2 {
		return 0, errInvalidHex
	}
	v, err := strconv.ParseUint(s, 16, 8)
	return uint8(v), err
}

// rgbaTo256 approximates an RGBA color as a 256-color ANSI index suitable for
// storage in a NativeBuffer cell (fg/bg are uint16 color indices).
func rgbaTo256(c core.RGBA) uint16 {
	if c.A == 0 {
		return 0
	}
	// Grayscale ramp (232-255).
	if c.R == c.G && c.G == c.B {
		if c.R < 8 {
			return 0 // black
		}
		if c.R > 247 {
			return 231 // white
		}
		return 232 + uint16((int(c.R)*24)/255)
	}
	// 6x6x6 color cube (16-231).
	r := uint16(int(c.R) * 5 / 255)
	g := uint16(int(c.G) * 5 / 255)
	b := uint16(int(c.B) * 5 / 255)
	return 16 + 36*r + 6*g + b
}

// writeCell writes a single styled cell into the buffer. Nil colors leave the
// existing/default color unchanged (0).
func writeCell(buffer *core.NativeBuffer, x, y int, ch uint32, fg, bg *core.RGBA, attrs uint32) {
	var f16, b16 uint16
	if fg != nil {
		f16 = rgbaTo256(*fg)
	}
	if bg != nil {
		b16 = rgbaTo256(*bg)
	}
	buffer.SetCell(x, y, ch, f16, b16, attrs)
}

// fillRect fills a rectangular region with a styled cell.
func fillRect(buffer *core.NativeBuffer, x, y, w, h int, ch uint32, fg, bg *core.RGBA, attrs uint32) {
	var f16, b16 uint16
	if fg != nil {
		f16 = rgbaTo256(*fg)
	}
	if bg != nil {
		b16 = rgbaTo256(*bg)
	}
	buffer.FillRect(x, y, w, h, ch, f16, b16, attrs)
}

// writeText writes a string into the buffer on a single line and returns the
// next x position. Newlines are skipped (caller is responsible for line layout).
func writeText(buffer *core.NativeBuffer, x, y int, s string, fg, bg *core.RGBA, attrs uint32) int {
	for _, ch := range s {
		if ch == '\n' || ch == '\r' {
			continue
		}
		writeCell(buffer, x, y, uint32(ch), fg, bg, attrs)
		x++
	}
	return x
}

// wrapText wraps s so that no line exceeds width columns. Existing newlines
// are preserved as hard breaks. Long words are split at the width boundary.
func wrapText(s string, width int) []string {
	if width <= 0 {
		width = 1
	}
	var result []string
	for _, line := range splitLines(s) {
		if measureCellWidth(line) <= width {
			result = append(result, line)
			continue
		}
		words := strings.FieldsFunc(line, func(r rune) bool { return r == ' ' || r == '\t' })
		if len(words) == 0 {
			result = append(result, "")
			continue
		}
		cur := ""
		for _, w := range words {
			add := cur
			if add != "" {
				add += " "
			}
			add += w
			if measureCellWidth(add) <= width {
				cur = add
				continue
			}
			if cur != "" {
				result = append(result, cur)
			}
			if measureCellWidth(w) > width {
				for i := 0; i < len(w); i += width {
					end := i + width
					if end > len(w) {
						end = len(w)
					}
					result = append(result, w[i:end])
				}
				cur = ""
			} else {
				cur = w
			}
		}
		if cur != "" {
			result = append(result, cur)
		}
	}
	if len(result) == 0 {
		result = append(result, "")
	}
	return result
}

// splitLines splits content into lines.
func splitLines(content string) []string {
	if content == "" {
		return []string{""}
	}
	var lines []string
	start := 0
	for i := 0; i < len(content); i++ {
		if content[i] == '\n' {
			lines = append(lines, content[start:i])
			start = i + 1
		}
	}
	lines = append(lines, content[start:])
	return lines
}

// isFocused reports whether this renderable is the current focus target.
func isFocused(ctx *core.RenderContext, self *core.Renderable) bool {
	return ctx != nil && ctx.CurrentFocusedRenderable == self
}

// clampScroll clamps scroll to [0, maxScroll] and returns the clamped value.
func clampScroll(scroll, maxScroll int) int {
	if maxScroll < 0 {
		maxScroll = 0
	}
	return clampInt(scroll, 0, maxScroll)
}
