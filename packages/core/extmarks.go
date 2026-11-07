package core

import (
	"sync"
)

// ExtmarkType defines the kind of extmark.
type ExtmarkType string

const (
	// ExtmarkTypeSelection marks a text selection range.
	ExtmarkTypeSelection ExtmarkType = "selection"
	// ExtmarkTypeHighlight marks a syntax or search highlight range.
	ExtmarkTypeHighlight ExtmarkType = "highlight"
	// ExtmarkTypeError marks an error range (e.g. diagnostic).
	ExtmarkTypeError ExtmarkType = "error"
	// ExtmarkTypeCustom is a user-defined extmark.
	ExtmarkTypeCustom ExtmarkType = "custom"
)

// Extmark represents a range marker with associated metadata.
type Extmark struct {
	ID        string
	Type      ExtmarkType
	StartLine int
	StartCol  int
	EndLine   int
	EndCol    int
	// Style attributes to apply to this range
	Attributes uint32
	// Foreground color override (0 = inherit)
	Foreground RGBA
	// Background color override (0 = inherit)
	Background RGBA
	// Priority determines rendering order (higher = drawn on top)
	Priority int
	// Data is user-defined metadata attached to this extmark
	Data interface{}
}

// ExtmarkManager manages extmarks for a renderable or buffer.
type ExtmarkManager struct {
	mu       sync.RWMutex
	marks    map[string]*Extmark
	nextID   int
	priority map[string]int // markID -> priority
}

// NewExtmarkManager creates a new ExtmarkManager.
func NewExtmarkManager() *ExtmarkManager {
	return &ExtmarkManager{
		marks:    make(map[string]*Extmark),
		nextID:   1,
		priority: make(map[string]int),
	}
}

// Add creates a new extmark and returns its ID.
func (m *ExtmarkManager) Add(extmark *Extmark) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	if extmark.ID == "" {
		extmark.ID = generateExtmarkID(m.nextID)
		m.nextID++
	}

	m.marks[extmark.ID] = extmark
	m.priority[extmark.ID] = extmark.Priority
	return extmark.ID
}

// Get returns an extmark by ID.
func (m *ExtmarkManager) Get(id string) (*Extmark, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	mark, ok := m.marks[id]
	return mark, ok
}

// Remove removes an extmark by ID.
func (m *ExtmarkManager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.marks, id)
	delete(m.priority, id)
}

// Clear removes all extmarks.
func (m *ExtmarkManager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.marks = make(map[string]*Extmark)
	m.priority = make(map[string]int)
	m.nextID = 1
}

// List returns all extmarks sorted by priority (descending).
func (m *ExtmarkManager) List() []*Extmark {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Extmark, 0, len(m.marks))
	for _, mark := range m.marks {
		result = append(result, mark)
	}

	// Simple insertion sort by priority descending
	for i := 1; i < len(result); i++ {
		key := result[i]
		j := i - 1
		for j >= 0 && result[j].Priority < key.Priority {
			result[j+1] = result[j]
			j--
		}
		result[j+1] = key
	}

	return result
}

// ListInRange returns extmarks that intersect with the given range.
func (m *ExtmarkManager) ListInRange(startLine, startCol, endLine, endCol int) []*Extmark {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Extmark
	for _, mark := range m.marks {
		if mark.Overlaps(startLine, startCol, endLine, endCol) {
			result = append(result, mark)
		}
	}
	return result
}

// Count returns the number of extmarks.
func (m *ExtmarkManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.marks)
}

// Overlaps checks if this extmark overlaps with the given range.
func (e *Extmark) Overlaps(startLine, startCol, endLine, endCol int) bool {
	// No overlap if extmark ends before range starts
	if e.EndLine < startLine || (e.EndLine == startLine && e.EndCol <= startCol) {
		return false
	}
	// No overlap if extmark starts after range ends
	if e.StartLine > endLine || (e.StartLine == endLine && e.StartCol >= endCol) {
		return false
	}
	return true
}

// Contains checks if a position is within this extmark.
func (e *Extmark) Contains(line, col int) bool {
	if line < e.StartLine || line > e.EndLine {
		return false
	}
	if line == e.StartLine && col < e.StartCol {
		return false
	}
	if line == e.EndLine && col >= e.EndCol {
		return false
	}
	return true
}

// SetAttributes updates the styling attributes of an extmark.
func (m *ExtmarkManager) SetAttributes(id string, attributes uint32) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	mark, ok := m.marks[id]
	if !ok {
		return false
	}
	mark.Attributes = attributes
	return true
}

// SetStyle updates the foreground and background colors of an extmark.
func (m *ExtmarkManager) SetStyle(id string, fg, bg RGBA) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	mark, ok := m.marks[id]
	if !ok {
		return false
	}
	mark.Foreground = fg
	mark.Background = bg
	return true
}

// SetRange updates the range of an extmark.
func (m *ExtmarkManager) SetRange(id string, startLine, startCol, endLine, endCol int) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	mark, ok := m.marks[id]
	if !ok {
		return false
	}
	mark.StartLine = startLine
	mark.StartCol = startCol
	mark.EndLine = endLine
	mark.EndCol = endCol
	return true
}

// generateExtmarkID generates a unique extmark ID.
func generateExtmarkID(seq int) string {
	return "extmark_" + string(rune('a'+seq%26)) + itoa(seq)
}

// itoa converts an integer to a string (simple implementation).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var result string
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}

// ExtmarkLayer applies extmarks to a buffer during rendering.
type ExtmarkLayer struct {
	manager *ExtmarkManager
	buffer  *NativeBuffer
}

// NewExtmarkLayer creates a new ExtmarkLayer for the given buffer.
func NewExtmarkLayer(manager *ExtmarkManager, buffer *NativeBuffer) *ExtmarkLayer {
	return &ExtmarkLayer{
		manager: manager,
		buffer:  buffer,
	}
}

// Apply applies all extmarks to the buffer, overwriting cell attributes.
func (l *ExtmarkLayer) Apply() {
	if l.manager == nil || l.buffer == nil {
		return
	}

	marks := l.manager.List()
	width := l.buffer.Width()
	height := l.buffer.Height()

	for _, mark := range marks {
		l.applyMark(mark, width, height)
	}
}

// applyMark applies a single extmark to the buffer.
func (l *ExtmarkLayer) applyMark(mark *Extmark, width, height int) {
	if mark.StartLine >= height {
		return
	}

	endLine := mark.EndLine
	if endLine >= height {
		endLine = height - 1
	}

	for y := mark.StartLine; y <= endLine; y++ {
		startCol := 0
		endCol := width

		if y == mark.StartLine && y == mark.EndLine {
			startCol = mark.StartCol
			endCol = mark.EndCol
		} else if y == mark.StartLine {
			startCol = mark.StartCol
		} else if y == mark.EndLine {
			endCol = mark.EndCol
		}

		if startCol < 0 {
			startCol = 0
		}
		if endCol > width {
			endCol = width
		}

		for x := startCol; x < endCol; x++ {
			// Apply attributes
			if mark.Attributes != 0 {
				ch, fg, bg, attrs, _ := l.buffer.GetCell(x, y)
				l.buffer.SetCell(x, y, ch, fg, bg, attrs|mark.Attributes)
			}
			// Apply foreground color override
			if mark.Foreground.A > 0 {
				ch, _, bg, attrs, _ := l.buffer.GetCell(x, y)
				l.buffer.SetCell(x, y, ch, rgbaToUint16(mark.Foreground), bg, attrs)
			}
			// Apply background color override
			if mark.Background.A > 0 {
				ch, fg, _, attrs := l.getCell(x, y)
				l.buffer.SetCell(x, y, ch, fg, rgbaToUint16(mark.Background), attrs)
			}
		}
	}
}

// getCell is a helper to get cell values (wrapper around GetCell).
func (l *ExtmarkLayer) getCell(x, y int) (uint32, uint16, uint16, uint32) {
	ch, fg, bg, attrs, _ := l.buffer.GetCell(x, y)
	return ch, fg, bg, attrs
}

// rgbaToUint16 converts an RGBA to a 16-bit color value for buffer storage.
func rgbaToUint16(c RGBA) uint16 {
	return uint16(c.R)<<8 | uint16(c.G)
}
