package native

import (
	"errors"
	"sync"
	"unicode/utf8"
)

// gapBuffer is a gap buffer: a contiguous byte slice with a "gap" (uninitialized
// region) between the left and right segments. Insertions write into the gap and
// shift it to the insertion point; deletions expand the gap. This makes insert
// and delete at the gap position O(1), and O(n) elsewhere (memmove).
type gapBuffer struct {
	mu       sync.RWMutex
	data     []byte // underlying storage
	leftLen  int    // bytes used before the gap
	rightLen int    // bytes used after the gap
	cap      int    // total capacity of data
}

// newGapBuffer creates a gap buffer with initial content.
func newGapBuffer(content string) *gapBuffer {
	contentBytes := []byte(content)
	gapSize := 1024
	total := len(contentBytes) + gapSize
	data := make([]byte, total)
	copy(data, contentBytes)
	copy(data[len(contentBytes):], make([]byte, gapSize))
	return &gapBuffer{
		data:     data,
		leftLen:  len(contentBytes),
		rightLen: 0,
		cap:      total,
	}
}

// length returns the total logical length in bytes.
func (gb *gapBuffer) length() int {
	gb.mu.RLock()
	defer gb.mu.RUnlock()
	return gb.leftLen + gb.rightLen
}

// gapStart returns the byte index where the gap begins.
func (gb *gapBuffer) gapStart() int {
	gb.mu.RLock()
	defer gb.mu.RUnlock()
	return gb.leftLen
}

// gapEnd returns the byte index where the gap ends.
func (gb *gapBuffer) gapEnd() int {
	gb.mu.RLock()
	defer gb.mu.RUnlock()
	return gb.leftLen + (gb.cap - gb.leftLen - gb.rightLen)
}

// bytes returns a copy of the logical content.
func (gb *gapBuffer) bytes() []byte {
	gb.mu.RLock()
	defer gb.mu.RUnlock()
	out := make([]byte, gb.leftLen+gb.rightLen)
	copy(out, gb.data[:gb.leftLen])
	copy(out[gb.leftLen:], gb.data[gb.gapEnd():gb.gapEnd()+gb.rightLen])
	return out
}

// string returns the logical content as a string.
func (gb *gapBuffer) string() string {
	return string(gb.bytes())
}

// ensureGap moves the gap to the given position and ensures it has at least
// `need` free bytes. Must be called with the write lock held.
func (gb *gapBuffer) ensureGapLocked(pos, need int) {
	if pos < 0 {
		pos = 0
	}
	if pos > gb.leftLen+gb.rightLen {
		pos = gb.leftLen + gb.rightLen
	}

	// If gap is already at pos, just ensure size.
	if gb.leftLen == pos {
		gb.growGapLocked(need)
		return
	}

	if pos < gb.leftLen {
		// Move [pos, leftLen) to just before the right segment.
		moveLen := gb.leftLen - pos
		dstStart := gb.cap - gb.rightLen - moveLen
		copy(gb.data[dstStart:dstStart+moveLen], gb.data[pos:gb.leftLen])
		gb.leftLen = pos
	} else {
		// pos > leftLen: move data from right of gap to left of gap.
		// The region [gapEnd, gapEnd + (pos - leftLen)) moves to just before gap.
		rightStart := gb.cap - gb.rightLen
		moveLen := pos - gb.leftLen
		// Shift [rightStart, rightStart+moveLen) left to [leftLen, leftLen+moveLen)
		copy(gb.data[gb.leftLen:gb.leftLen+moveLen], gb.data[rightStart:rightStart+moveLen])
		gb.leftLen = pos
	}
	gb.growGapLocked(need)
}

// growGapLocked ensures the gap is at least `need` bytes, growing storage if needed.
func (gb *gapBuffer) growGapLocked(need int) {
	gapSize := gb.cap - gb.leftLen - gb.rightLen
	if gapSize >= need {
		return
	}
	// Grow: double capacity until gap sufficient.
	newCap := gb.cap
	for {
		newCap *= 2
		if newCap-gb.leftLen-gb.rightLen >= need {
			break
		}
	}
	newData := make([]byte, newCap)
	// Copy left segment.
	copy(newData, gb.data[:gb.leftLen])
	// Copy right segment to the end.
	rightStart := gb.cap - gb.rightLen
	copy(newData[newCap-gb.rightLen:], gb.data[rightStart:])
	gb.data = newData
	gb.cap = newCap
}

// insertLocked inserts text at the current gap position (gap must be at pos).
// Must be called with the write lock held and the gap already positioned.
func (gb *gapBuffer) insertLocked(text []byte) {
	gb.ensureGapLocked(gb.leftLen, len(text))
	copy(gb.data[gb.leftLen:], text)
	gb.leftLen += len(text)
}

// deleteLocked deletes [start, end) bytes. Must be called with write lock held.
func (gb *gapBuffer) deleteLocked(start, end int) {
	if start == end {
		return
	}
	gb.ensureGapLocked(start, 0)
	// Now gap is at `start`. Expand it to cover [start, end).
	gb.rightLen -= (end - start)
}

// EditBuffer is an editable text buffer with a gap buffer backend and undo/redo.
type EditBuffer struct {
	mu         sync.RWMutex
	id         uint64
	buffer     *gapBuffer
	history    []string // snapshots for undo/redo
	pos        int      // current position in history (-1 means before first)
	maxHistory int      // maximum number of history entries
}

// NewEditBuffer creates a new empty EditBuffer.
func NewEditBuffer() *EditBuffer {
	return &EditBuffer{
		buffer:     newGapBuffer(""),
		history:    make([]string, 0),
		pos:        0,
		maxHistory: 1000,
	}
}

// NewEditBufferWithContent creates a new EditBuffer with initial content.
func NewEditBufferWithContent(content string) *EditBuffer {
	eb := NewEditBuffer()
	eb.SetContent(content)
	return eb
}

// ID returns the buffer ID.
func (eb *EditBuffer) ID() uint64 {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return eb.id
}

// SetID sets the buffer ID (used by ResourceContext).
func (eb *EditBuffer) SetID(id uint64) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.id = id
}

// SetContent replaces the buffer content and resets history atomically.
func (eb *EditBuffer) SetContent(content string) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	if !utf8.ValidString(content) {
		content = sanitizeUTF8(content)
	}
	eb.buffer = newGapBuffer(content)
	eb.history = make([]string, 0)
	eb.history = append(eb.history, content)
	eb.pos = 0
}

// GetContent returns the buffer content.
func (eb *EditBuffer) GetContent() string {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return eb.buffer.string()
}

// Length returns the byte length of the content.
func (eb *EditBuffer) Length() int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return eb.buffer.length()
}

// RuneCount returns the number of runes in the content.
func (eb *EditBuffer) RuneCount() int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return utf8.RuneCountInString(eb.buffer.string())
}

// Insert inserts text at the given byte offset atomically and records history.
func (eb *EditBuffer) Insert(offset int, text string) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	if offset < 0 || offset > eb.buffer.length() {
		return errors.New("offset out of range")
	}
	if !utf8.ValidString(text) {
		return errors.New("invalid UTF-8")
	}
	if len(text) == 0 {
		return nil
	}
	// Move gap to offset and insert.
	gb := eb.buffer
	gb.mu.Lock()
	gb.ensureGapLocked(offset, len(text))
	gb.insertLocked([]byte(text))
	gb.mu.Unlock()

	// Record snapshot after the change.
	content := gb.string()
	eb.pushHistory(content)
	return nil
}

// InsertRune inserts a single rune at the given byte offset.
func (eb *EditBuffer) InsertRune(offset int, r rune) error {
	return eb.Insert(offset, string(r))
}

// Delete deletes text from start to end byte offset atomically and records history.
func (eb *EditBuffer) Delete(start, end int) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	if start < 0 || end > eb.buffer.length() || start > end {
		return errors.New("invalid range")
	}
	if start == end {
		return nil
	}
	gb := eb.buffer
	gb.mu.Lock()
	gb.deleteLocked(start, end)
	gb.mu.Unlock()

	content := gb.string()
	eb.pushHistory(content)
	return nil
}

// DeleteRune deletes the rune at the given byte offset.
func (eb *EditBuffer) DeleteRune(offset int) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	if offset < 0 || offset >= eb.buffer.length() {
		return errors.New("offset out of range")
	}
	_, size := utf8.DecodeRuneInString(eb.buffer.string()[offset:])
	if size == 0 {
		size = 1
	}
	return eb.Delete(offset, offset+size)
}

// pushHistory appends a snapshot, truncating any redo history ahead of pos.
func (eb *EditBuffer) pushHistory(content string) {
	// Drop any redo history ahead of the current position.
	if eb.pos >= 0 && eb.pos < len(eb.history)-1 {
		eb.history = eb.history[:eb.pos+1]
	}
	// Avoid duplicate consecutive snapshots.
	if len(eb.history) > 0 && eb.history[len(eb.history)-1] == content {
		return
	}
	eb.history = append(eb.history, content)
	eb.pos++
	if eb.maxHistory > 0 && len(eb.history) > eb.maxHistory {
		// Drop oldest entry.
		eb.history = eb.history[1:]
		eb.pos--
	}
}

// Undo reverts the last change. Returns false if nothing to undo.
func (eb *EditBuffer) Undo() bool {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	if eb.pos <= 0 {
		return false
	}
	eb.pos--
	content := eb.history[eb.pos]
	gb := newGapBuffer(content)
	eb.buffer = gb
	return true
}

// Redo reapplies the last undone change. Returns false if nothing to redo.
func (eb *EditBuffer) Redo() bool {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	if eb.pos < 0 || eb.pos >= len(eb.history)-1 {
		return false
	}
	eb.pos++
	content := eb.history[eb.pos]
	gb := newGapBuffer(content)
	eb.buffer = gb
	return true
}

// CanUndo returns true if an undo is available.
func (eb *EditBuffer) CanUndo() bool {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return eb.pos > 0
}

// CanRedo returns true if a redo is available.
func (eb *EditBuffer) CanRedo() bool {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return eb.pos >= 0 && eb.pos < len(eb.history)-1
}

// HistoryPosition returns the current undo position.
func (eb *EditBuffer) HistoryPosition() int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return eb.pos
}

// HistoryLength returns the total number of history entries.
func (eb *EditBuffer) HistoryLength() int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return len(eb.history)
}

// SetMaxHistory sets the maximum number of undo history entries.
func (eb *EditBuffer) SetMaxHistory(max int) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	if max < 0 {
		max = 0
	}
	eb.maxHistory = max
}

// GetByte returns the byte at the given offset.
func (eb *EditBuffer) GetByte(offset int) (byte, bool) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	content := eb.buffer.string()
	if offset < 0 || offset >= len(content) {
		return 0, false
	}
	return content[offset], true
}

// GetRuneAt returns the rune and byte width at the given byte offset.
func (eb *EditBuffer) GetRuneAt(offset int) (rune, int, bool) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	content := eb.buffer.string()
	if offset < 0 || offset >= len(content) {
		return 0, 0, false
	}
	r, size := utf8.DecodeRuneInString(content[offset:])
	if r == utf8.RuneError && size == 0 {
		return 0, 0, false
	}
	return r, size, true
}

var _ = errors.New
