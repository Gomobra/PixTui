package native

import (
	"unicode/utf8"
)

// RuneCount returns the number of runes in a UTF-8 string
func RuneCount(s string) int {
	return utf8.RuneCountInString(s)
}

// ByteCount returns the number of bytes in a UTF-8 string
func ByteCount(s string) int {
	return len(s)
}

// DecodeRune decodes the first rune from a UTF-8 byte slice
func DecodeRune(b []byte) (rune, int) {
	return utf8.DecodeRune(b)
}

// EncodeRune encodes a rune into a UTF-8 byte slice
func EncodeRune(b []byte, r rune) int {
	return utf8.EncodeRune(b, r)
}

// Valid reports whether s consists entirely of valid UTF-8
func Valid(s string) bool {
	return utf8.ValidString(s)
}

// RuneStart reports whether b is a continuation byte
func RuneStart(b byte) bool {
	return utf8.RuneStart(b)
}

// RuneLen returns the number of bytes needed to encode r
func RuneLen(r rune) int {
	return utf8.RuneLen(r)
}

// IsPrintable reports whether the rune is a printable character
func IsPrintable(r rune) bool {
	return r >= 32 && r != 127
}

// IsControl reports whether the rune is a control character
func IsControl(r rune) bool {
	return r < 32 || r == 127
}

// StringWidth returns the display width of a UTF-8 string
func StringWidth(s string) int {
	width := 0
	for _, r := range s {
		width += RuneWidth(r)
	}
	return width
}
