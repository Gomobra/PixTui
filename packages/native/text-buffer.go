package native

import (
	"errors"
	"sync"
	"unicode/utf8"
)

// TextBuffer stores text using a Rope as the backing store for efficient
// insertions and deletions at arbitrary positions.
type TextBuffer struct {
	mu        sync.RWMutex
	id        uint64
	rope      *Rope
	lines     []string // cached line split
	lineValid bool
}

// NewTextBuffer creates a new empty TextBuffer backed by a Rope.
func NewTextBuffer() *TextBuffer {
	return &TextBuffer{
		rope:      NewRope(),
		lines:     []string{""},
		lineValid: true,
	}
}

// NewTextBufferWithContent creates a new TextBuffer with initial content.
func NewTextBufferWithContent(content string) *TextBuffer {
	tb := NewTextBuffer()
	tb.SetContent(content)
	return tb
}

// ID returns the buffer ID.
func (tb *TextBuffer) ID() uint64 {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	return tb.id
}

// SetID sets the buffer ID (used by ResourceContext).
func (tb *TextBuffer) SetID(id uint64) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.id = id
}

// SetContent replaces the entire buffer content and rebuilds the Rope.
func (tb *TextBuffer) SetContent(content string) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	if !utf8.ValidString(content) {
		content = sanitizeUTF8(content)
	}
	tb.rope = NewRope()
	if len(content) > 0 {
		if err := tb.rope.Insert(0, content); err != nil {
			return
		}
	}
	tb.lines = splitLines(content)
	tb.lineValid = true
}

// GetContent returns the full buffer content as a string.
func (tb *TextBuffer) GetContent() string {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	return tb.rope.GetString()
}

// Length returns the byte length of the content.
func (tb *TextBuffer) Length() int {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	return tb.rope.Length()
}

// LineCount returns the number of logical lines.
func (tb *TextBuffer) LineCount() int {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	tb.ensureLinesRLocked()
	return len(tb.lines)
}

// GetLine returns the line at the given index.
func (tb *TextBuffer) GetLine(index int) (string, bool) {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	tb.ensureLinesRLocked()
	if index < 0 || index >= len(tb.lines) {
		return "", false
	}
	return tb.lines[index], true
}

// ensureLinesRLocked rebuilds the cached line split if invalidated.
// Caller must hold at least a read lock.
func (tb *TextBuffer) ensureLinesRLocked() {
	if tb.lineValid {
		return
	}
	// We need a write lock to mutate; release read and re-acquire.
	tb.mu.RUnlock()
	tb.mu.Lock()
	tb.lines = splitLines(tb.rope.GetString())
	tb.lineValid = true
	tb.mu.Unlock()
	tb.mu.RLock()
}

// Insert inserts text at the given byte offset using the Rope.
func (tb *TextBuffer) Insert(offset int, text string) error {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	if offset < 0 || offset > tb.rope.Length() {
		return errors.New("offset out of range")
	}
	if !utf8.ValidString(text) {
		return errors.New("invalid UTF-8")
	}
	if err := tb.rope.Insert(offset, text); err != nil {
		return err
	}
	tb.lineValid = false
	return nil
}

// InsertRune inserts a single rune at the given byte offset.
func (tb *TextBuffer) InsertRune(offset int, r rune) error {
	return tb.Insert(offset, string(r))
}

// Delete deletes text from start to end byte offset using the Rope.
func (tb *TextBuffer) Delete(start, end int) error {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	if start < 0 || end > tb.rope.Length() || start > end {
		return errors.New("invalid range")
	}
	if start == end {
		return nil
	}
	content := tb.rope.GetString()
	before := content[:start]
	after := content[end:]
	tb.rope = NewRope()
	if len(before) > 0 {
		if err := tb.rope.Insert(0, before); err != nil {
			return err
		}
	}
	if len(after) > 0 {
		if err := tb.rope.Insert(len(before), after); err != nil {
			return err
		}
	}
	tb.lineValid = false
	return nil
}

// GetByte returns the byte at the given offset, or false if out of range.
func (tb *TextBuffer) GetByte(offset int) (byte, bool) {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	content := tb.rope.GetString()
	if offset < 0 || offset >= len(content) {
		return 0, false
	}
	return content[offset], true
}

// GetRuneAt returns the rune and its byte width at the given byte offset.
func (tb *TextBuffer) GetRuneAt(offset int) (rune, int, bool) {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	content := tb.rope.GetString()
	if offset < 0 || offset >= len(content) {
		return 0, 0, false
	}
	r, size := utf8.DecodeRuneInString(content[offset:])
	if r == utf8.RuneError && size == 0 {
		return 0, 0, false
	}
	return r, size, true
}

// RuneOffsetToByteOffset converts a rune index to a byte offset.
func (tb *TextBuffer) RuneOffsetToByteOffset(runeOffset int) int {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	content := tb.rope.GetString()
	bytes := 0
	runes := 0
	for i := 0; i < len(content); i++ {
		if runes == runeOffset {
			return bytes
		}
		_, size := utf8.DecodeRuneInString(content[i:])
		if size == 0 {
			size = 1
		}
		bytes += size
		runes++
		i += size - 1
	}
	return bytes
}

// ByteOffsetToRuneOffset converts a byte offset to a rune index.
func (tb *TextBuffer) ByteOffsetToRuneOffset(byteOffset int) int {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	content := tb.rope.GetString()
	if byteOffset >= len(content) {
		byteOffset = len(content)
	}
	if byteOffset < 0 {
		return 0
	}
	runes := 0
	for i := 0; i < byteOffset; {
		_, size := utf8.DecodeRuneInString(content[i:])
		if size == 0 {
			size = 1
		}
		runes++
		i += size
	}
	return runes
}

// RuneCount returns the number of runes in the buffer.
func (tb *TextBuffer) RuneCount() int {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	content := tb.rope.GetString()
	return utf8.RuneCountInString(content)
}

// InvalidateLines marks the cached line split as stale.
func (tb *TextBuffer) InvalidateLines() {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.lineValid = false
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

// sanitizeUTF8 replaces invalid UTF-8 bytes with the replacement character.
func sanitizeUTF8(s string) string {
	return stringsToValidUTF8(s)
}

func stringsToValidUTF8(s string) string {
	var buf bytesBuilder
	for i := 0; i < len(s); i++ {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size == 0 {
			buf.WriteRune(utf8.RuneError)
		} else {
			buf.WriteRune(r)
			i += size - 1
		}
	}
	return buf.String()
}

// bytesBuilder is a minimal strings.Builder-like accumulator.
type bytesBuilder struct {
	buf []byte
}

func (b *bytesBuilder) WriteRune(r rune) {
	b.buf = append(b.buf, string(r)...)
}

func (b *bytesBuilder) String() string {
	return string(b.buf)
}

var _ = errors.New
