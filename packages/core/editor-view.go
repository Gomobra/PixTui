package core

import (
	"errors"
	"fmt"
	"sync"
	"unicode/utf8"
)

// EditorView provides a view into an EditBuffer with cursor, selection, and
// working undo/redo wired to the underlying EditBuffer.
type EditorView struct {
	mu        sync.RWMutex
	id        string
	buffer    *EditBuffer
	cursorX   int // byte offset of cursor
	cursorY   int // logical line (visual row) of cursor
	selection *Selection
	dirty     bool
}


// NewEditorView creates a new EditorView for the given EditBuffer.
func NewEditorView(buffer *EditBuffer) *EditorView {
	return &EditorView{
		buffer: buffer,
	}
}

// NewEditorViewWithContent creates a new EditorView with an EditBuffer containing content.
func NewEditorViewWithContent(content string) *EditorView {
	return &EditorView{
		buffer: NewEditBufferWithContent(content),
	}
}

// ID returns the view ID.
func (ev *EditorView) ID() string {
	ev.mu.RLock()
	defer ev.mu.RUnlock()
	return ev.id
}

// SetID sets the view ID (used by ResourceContext).
func (ev *EditorView) SetID(id string) {
	ev.mu.Lock()
	defer ev.mu.Unlock()
	ev.id = id
}

// SetCursor sets the cursor position.
func (ev *EditorView) SetCursor(x, y int) {
	ev.mu.Lock()
	defer ev.mu.Unlock()
	ev.cursorX = x
	ev.cursorY = y
}

// Cursor returns the cursor position (byte offset, logical line).
func (ev *EditorView) Cursor() (int, int) {
	ev.mu.RLock()
	defer ev.mu.RUnlock()
	return ev.cursorX, ev.cursorY
}

// CursorByteOffset returns the cursor as a byte offset into the buffer content.
func (ev *EditorView) CursorByteOffset() int {
	ev.mu.RLock()
	defer ev.mu.RUnlock()
	return ev.cursorX
}

// GetContent returns the buffer content.
func (ev *EditorView) GetContent() string {
	ev.mu.RLock()
	defer ev.mu.RUnlock()
	return ev.buffer.GetContent()
}

// SetContent replaces the buffer content and resets the cursor/selection.
func (ev *EditorView) SetContent(content string) {
	ev.mu.Lock()
	defer ev.mu.Unlock()
	ev.buffer.SetContent(content)
	ev.cursorX = 0
	ev.cursorY = 0
	ev.selection = nil
	ev.dirty = true
}

// Insert inserts text at the current cursor position, replacing any selection.
func (ev *EditorView) Insert(text string) error {
	ev.mu.Lock()
	defer ev.mu.Unlock()
	if !utf8.ValidString(text) {
		return errors.New("invalid UTF-8")
	}
	if ev.selection != nil && ev.selection.Start != ev.selection.End {
		// Replace selection
		if err := ev.buffer.Delete(ev.selection.Start, ev.selection.End); err != nil {
			return err
		}
		ev.cursorX = ev.selection.Start
		ev.selection = nil
	}
	if err := ev.buffer.Insert(ev.cursorX, text); err != nil {
		return err
	}
	ev.cursorX += len(text)
	content := ev.buffer.GetContent()
	if ev.cursorX > len(content) {
		ev.cursorX = len(content)
	}
	ev.cursorY = ev.lineForOffset(content, ev.cursorX)
	ev.dirty = true
	return nil
}

// Delete deletes text from start to end byte offset.
func (ev *EditorView) Delete(start, end int) error {
	ev.mu.Lock()
	defer ev.mu.Unlock()
	if err := ev.buffer.Delete(start, end); err != nil {
		return err
	}
	if ev.cursorX > end {
		ev.cursorX -= (end - start)
	} else if ev.cursorX > start {
		ev.cursorX = start
	}
	if ev.cursorX < 0 {
		ev.cursorX = 0
	}
	content := ev.buffer.GetContent()
	if ev.cursorX > len(content) {
		ev.cursorX = len(content)
	}
	ev.cursorY = ev.lineForOffset(content, ev.cursorX)
	ev.dirty = true
	return nil
}

// Undo reverts the last change via the EditBuffer. Returns false if nothing to undo.
func (ev *EditorView) Undo() bool {
	ev.mu.Lock()
	defer ev.mu.Unlock()
	if !ev.buffer.Undo() {
		return false
	}
	content := ev.buffer.GetContent()
	if ev.cursorX > len(content) {
		ev.cursorX = len(content)
	}
	ev.cursorY = ev.lineForOffset(content, ev.cursorX)
	ev.dirty = true
	return true
}

// Redo reapplies the last undone change via the EditBuffer. Returns false if nothing to redo.
func (ev *EditorView) Redo() bool {
	ev.mu.Lock()
	defer ev.mu.Unlock()
	if !ev.buffer.Redo() {
		return false
	}
	content := ev.buffer.GetContent()
	if ev.cursorX > len(content) {
		ev.cursorX = len(content)
	}
	ev.cursorY = ev.lineForOffset(content, ev.cursorX)
	ev.dirty = true
	return true
}

// CanUndo returns true if an undo is available.
func (ev *EditorView) CanUndo() bool {
	ev.mu.RLock()
	defer ev.mu.RUnlock()
	return ev.buffer.CanUndo()
}

// CanRedo returns true if a redo is available.
func (ev *EditorView) CanRedo() bool {
	ev.mu.RLock()
	defer ev.mu.RUnlock()
	return ev.buffer.CanRedo()
}

// Length returns the byte length of the buffer content.
func (ev *EditorView) Length() int {
	ev.mu.RLock()
	defer ev.mu.RUnlock()
	return ev.buffer.Length()
}

// IsDirty returns true if the buffer has unsaved changes.
func (ev *EditorView) IsDirty() bool {
	ev.mu.RLock()
	defer ev.mu.RUnlock()
	return ev.dirty
}

// MarkClean marks the buffer as clean (no unsaved changes).
func (ev *EditorView) MarkClean() {
	ev.mu.Lock()
	defer ev.mu.Unlock()
	ev.dirty = false
}

// GetRuneAtCursor returns the rune at the cursor position and its byte width.
func (ev *EditorView) GetRuneAtCursor() (rune, int, bool) {
	ev.mu.RLock()
	defer ev.mu.RUnlock()
	return ev.buffer.GetRuneAt(ev.cursorX)
}

// SetSelection sets the selection range by byte offsets.
func (ev *EditorView) SetSelection(start, end int) {
	ev.mu.Lock()
	defer ev.mu.Unlock()
	content := ev.buffer.GetContent()
	if start < 0 {
		start = 0
	}
	if end > len(content) {
		end = len(content)
	}
	if start > end {
		start, end = end, start
	}
	ev.selection = &Selection{Start: start, End: end}
}

// GetSelection returns the current selection.
func (ev *EditorView) GetSelection() *Selection {
	ev.mu.RLock()
	defer ev.mu.RUnlock()
	if ev.selection == nil {
		return nil
	}
	return &Selection{Start: ev.selection.Start, End: ev.selection.End}
}

// lineForOffset returns the logical line number for a byte offset.
func (ev *EditorView) lineForOffset(content string, offset int) int {
	if offset < 0 {
		offset = 0
	}
	if offset > len(content) {
		offset = len(content)
	}
	lines := 0
	for i := 0; i < offset; i++ {
		if content[i] == '\n' {
			lines++
		}
	}
	return lines
}

// String returns a debug string.
func (ev *EditorView) String() string {
	ev.mu.RLock()
	defer ev.mu.RUnlock()
	return fmt.Sprintf("EditorView{id=%s, cursor=(%d,%d), dirty=%v, len=%d}",
		ev.id, ev.cursorX, ev.cursorY, ev.dirty, ev.buffer.Length())
}

var _ = errors.New
