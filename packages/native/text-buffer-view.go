package native

import (
	"strings"
	"sync"
)

// LineInfo contains line layout information for text rendering.
type LineInfo struct {
	LineStartCols    []int
	LineWidthCols    []int
	LineWidthColsMax int
	LineSources      []int
	LineWraps        []int
}

// WrapMode defines how lines are wrapped.
type WrapMode uint32

const (
	WrapModeNone WrapMode = 0
	WrapModeChar WrapMode = 1
	WrapModeWord WrapMode = 2
)

// TextBufferView provides a view into a TextBuffer with line wrapping,
// virtual lines, and scroll position.
type TextBufferView struct {
	mu          sync.RWMutex
	buffer      *TextBuffer
	wrapWidth   int
	wrapMode    WrapMode
	virtualLine []string
	scrollX     int // horizontal scroll in cells
	scrollY     int // vertical scroll in virtual lines
}

// NewTextBufferView creates a new view for the given buffer.
func NewTextBufferView(buffer *TextBuffer) *TextBufferView {
	return &TextBufferView{
		buffer:    buffer,
		wrapWidth: 80,
		wrapMode:  WrapModeWord,
	}
}

// SetBuffer changes the underlying buffer and rewraps.
func (v *TextBufferView) SetBuffer(buffer *TextBuffer) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.buffer = buffer
	v.scrollX = 0
	v.scrollY = 0
	v.rewrapLocked()
}

// SetWrapWidth sets the wrap width in cells.
func (v *TextBufferView) SetWrapWidth(width int) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.wrapWidth = width
	v.rewrapLocked()
}

// SetWrapMode sets the wrap mode (0=none, 1=char, 2=word).
func (v *TextBufferView) SetWrapMode(mode WrapMode) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.wrapMode = mode
	v.rewrapLocked()
}

// SetScrollX sets the horizontal scroll position in cells.
func (v *TextBufferView) SetScrollX(x int) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if x < 0 {
		x = 0
	}
	v.scrollX = x
}

// SetScrollY sets the vertical scroll position, clamped to [0, maxScrollY].
func (v *TextBufferView) SetScrollY(y int) {
	v.mu.Lock()
	defer v.mu.Unlock()
	maxY := len(v.virtualLine) - 1
	if maxY < 0 {
		maxY = 0
	}
	if y < 0 {
		y = 0
	}
	if y > maxY {
		y = maxY
	}
	v.scrollY = y
}

// ScrollX returns the horizontal scroll position.
func (v *TextBufferView) ScrollX() int {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.scrollX
}

// ScrollY returns the vertical scroll position.
func (v *TextBufferView) ScrollY() int {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.scrollY
}

// ScrollUp scrolls up by n virtual lines.
func (v *TextBufferView) ScrollUp(n int) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.scrollY -= n
	if v.scrollY < 0 {
		v.scrollY = 0
	}
}

// ScrollDown scrolls down by n virtual lines.
func (v *TextBufferView) ScrollDown(n int) {
	v.mu.Lock()
	defer v.mu.Unlock()
	maxY := len(v.virtualLine) - 1
	if maxY < 0 {
		maxY = 0
	}
	v.scrollY += n
	if v.scrollY > maxY {
		v.scrollY = maxY
	}
}

// VirtualLineCount returns the number of virtual lines.
func (v *TextBufferView) VirtualLineCount() int {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return len(v.virtualLine)
}

// GetVirtualLine returns the virtual line at the given index.
func (v *TextBufferView) GetVirtualLine(index int) (string, bool) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	if index < 0 || index >= len(v.virtualLine) {
		return "", false
	}
	return v.virtualLine[index], true
}

// GetVisibleLines returns the visible virtual lines starting at scrollY.
func (v *TextBufferView) GetVisibleLines(count int) []string {
	v.mu.RLock()
	defer v.mu.RUnlock()
	if v.scrollY >= len(v.virtualLine) {
		return nil
	}
	end := v.scrollY + count
	if end > len(v.virtualLine) {
		end = len(v.virtualLine)
	}
	result := make([]string, end-v.scrollY)
	copy(result, v.virtualLine[v.scrollY:end])
	return result
}

// GetLineCount returns the number of logical lines.
func (v *TextBufferView) GetLineCount() int {
	v.mu.RLock()
	defer v.mu.RUnlock()
	if v.buffer == nil {
		return 0
	}
	return v.buffer.LineCount()
}

// GetLineInfo returns line information for rendering.
func (v *TextBufferView) GetLineInfo() LineInfo {
	v.mu.RLock()
	defer v.mu.RUnlock()

	lineStartCols := make([]int, len(v.virtualLine))
	lineWidthCols := make([]int, len(v.virtualLine))
	lineSources := make([]int, len(v.virtualLine))
	lineWraps := make([]int, len(v.virtualLine))
	maxWidth := 0

	for i, vl := range v.virtualLine {
		lineStartCols[i] = 0
		lineWidthCols[i] = StringWidth(vl)
		if lineWidthCols[i] > maxWidth {
			maxWidth = lineWidthCols[i]
		}
		lineSources[i] = i
		lineWraps[i] = 0
	}

	return LineInfo{
		LineStartCols:    lineStartCols,
		LineWidthCols:    lineWidthCols,
		LineWidthColsMax: maxWidth,
		LineSources:      lineSources,
		LineWraps:        lineWraps,
	}
}

// GetContent returns the buffer content.
func (v *TextBufferView) GetContent() string {
	v.mu.RLock()
	defer v.mu.RUnlock()
	if v.buffer == nil {
		return ""
	}
	return v.buffer.GetContent()
}

// rewrap rewraps the buffer content into virtual lines.
// Caller must hold the write lock.
func (v *TextBufferView) rewrapLocked() {
	if v.buffer == nil {
		v.virtualLine = nil
		return
	}
	content := v.buffer.GetContent()
	if v.wrapMode == WrapModeNone || v.wrapWidth <= 0 {
		v.virtualLine = splitLines(content)
		return
	}

	var lines []string
	for _, line := range splitLines(content) {
		if line == "" {
			lines = append(lines, "")
			continue
		}
		if v.wrapMode == WrapModeChar {
			runes := []rune(line)
			for i := 0; i < len(runes); i += v.wrapWidth {
				end := i + v.wrapWidth
				if end > len(runes) {
					end = len(runes)
				}
				lines = append(lines, string(runes[i:end]))
			}
		} else {
			// Word wrap
			words := strings.Fields(line)
			var current string
			for _, word := range words {
				if current == "" {
					current = word
				} else if StringWidth(current+" "+word) <= v.wrapWidth {
					current = current + " " + word
				} else {
					lines = append(lines, current)
					current = word
				}
			}
			if current != "" {
				lines = append(lines, current)
			}
		}
	}
	v.virtualLine = lines
}
