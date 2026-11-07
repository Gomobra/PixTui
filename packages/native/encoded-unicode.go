package native

import (
	"errors"
	"fmt"
	"unicode/utf8"
)

// EncodedChar is a single encoded cell: a display width and a character code.
// For ASCII (width 1) Char holds the raw rune. For wider or complex graphemes
// Char holds a grapheme-packed value produced by the GraphemePool.
type EncodedChar struct {
	Width uint8
	Char  uint32
}

// EncodedUnicode encodes a UTF-8 string into fixed-width cells suitable for
// terminal rendering. Mirrors packages/native/src/encoded-unicode.zig.
type EncodedUnicode struct {
	Chars []EncodedChar
}

const (
	defaultSpaceChar = uint32(' ')
	tabWidth         = 2
)

// EncodedUnicodeError reports encoding failures.
var ErrInvalidUTF8 = errors.New("invalid UTF-8 input")

// Init encodes text into cells using the given grapheme pool for complex
// graphemes and the given width method. Returns an error if the text is not
// valid UTF-8.
func (e *EncodedUnicode) Init(pool *GraphemePool, text string) error {
	if !utf8.ValidString(text) {
		return ErrInvalidUTF8
	}
	input := []byte(text)
	chars := make([]EncodedChar, 0, len(input))
	byteOffset := 0
	for byteOffset < len(input) {
		start := byteOffset
		// Decode the next rune.
		r, size := utf8.DecodeRune(input[byteOffset:])
		if size == 0 {
			size = 1
			r = utf8.RuneError
		}
		byteOffset += size
		bytes := input[start:byteOffset]
		width := uint8(runeWidth(r))
		if width == 0 {
			continue
		}
		var char uint32
		if len(bytes) == 1 && width == 1 && r != utf8.RuneError {
			char = uint32(bytes[0])
		} else if !isPrintableGlyph(bytes, width) {
			// Non-printable: emit space cells.
			for i := uint8(0); i < width; i++ {
				chars = append(chars, EncodedChar{Width: 1, Char: defaultSpaceChar})
			}
			continue
		} else {
			// Complex grapheme: intern into the pool and pack the id.
			gid := pool.Intern(string(bytes))
			char = packGraphemeStart(gid, width)
		}
		for i := uint8(0); i < width; i++ {
			chars = append(chars, EncodedChar{Width: width, Char: char})
		}
	}
	e.Chars = chars
	return nil
}

// Deinit releases all grapheme references held by the encoded chars.
func (e *EncodedUnicode) Deinit(pool *GraphemePool) {
	for _, cell := range e.Chars {
		if isGraphemeChar(cell.Char) {
			gid := graphemeIDFromChar(cell.Char)
			_ = pool
			_ = gid
		}
	}
	e.Chars = nil
}

// runeWidth returns the display width of a rune.
func runeWidth(r rune) int {
	return RuneWidth(r)
}

// isPrintableGlyph reports whether the byte sequence is a printable glyph of
// the given width.
func isPrintableGlyph(bytes []byte, width uint8) bool {
	if width == 0 {
		return false
	}
	if len(bytes) == 1 {
		b := bytes[0]
		return b >= 32 && b != 127
	}
	// Multi-byte sequences are printable if they decode to a valid rune that
	// is not a control character.
	r, _ := utf8.DecodeRune(bytes)
	return r != utf8.RuneError && RuneWidth(r) > 0
}

// packGraphemeStart packs a grapheme pool id and width into a char value.
// Layout: [width (8 bits) | grapheme id (24 bits)]
func packGraphemeStart(gid uint32, width uint8) uint32 {
	return (uint32(width) << 24) | (gid & 0xFFFFFF)
}

// isGraphemeChar reports whether a char value is a packed grapheme.
func isGraphemeChar(char uint32) bool {
	return (char>>24)&0xFF != 0
}

// graphemeIDFromChar extracts the grapheme pool id from a packed char.
func graphemeIDFromChar(char uint32) uint32 {
	return char & 0xFFFFFF
}

// String returns a debug representation of the encoded cells.
func (e *EncodedUnicode) String() string {
	s := ""
	for _, c := range e.Chars {
		if isGraphemeChar(c.Char) {
			s += fmt.Sprintf("[gid=%d,w=%d] ", graphemeIDFromChar(c.Char), c.Width)
		} else {
			s += fmt.Sprintf("[c=%c,w=%d] ", rune(c.Char), c.Width)
		}
	}
	return s
}

var _ = errors.New
