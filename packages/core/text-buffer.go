package core

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"unicode/utf8"
)

// TextBuffer stores text with rope-based storage, mirrors
// packages/native/src/text-buffer.zig.
type TextBuffer struct {
	mu        sync.RWMutex
	id        string
	rope      *Rope // backing store
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
func (tb *TextBuffer) ID() string {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	return tb.id
}

// SetContent replaces the entire buffer content. Inserts all content into
// the rope at position 0 and rebuilds the line cache.
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

// GetContent returns the buffer content as a string.
func (tb *TextBuffer) GetContent() string {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	return tb.rope.GetString()
}

// Insert inserts text at the given byte offset using the Rope.
// Invalidates the line cache.
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

// Delete deletes text from start to end byte offset using the Rope.
// Invalidates the line cache.
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

// TextBufferView provides a view into a TextBuffer with wrapping and virtual lines.
type TextBufferView struct {
	mu          sync.RWMutex
	buffer      *TextBuffer
	wrapWidth   int
	wrapMode    uint32 // 0=none, 1=char, 2=word
	virtualLine []string
	scrollY     int // current vertical scroll position in virtual lines
}

// NewTextBufferView creates a new view for the given buffer.
func NewTextBufferView(buffer *TextBuffer) *TextBufferView {
	return &TextBufferView{
		buffer:    buffer,
		wrapWidth: 80,
		wrapMode:  2,
	}
}

// SetWrapWidth sets the wrap width in cells.
func (v *TextBufferView) SetWrapWidth(width int) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.wrapWidth = width
	v.rewrap()
}

// SetWrapMode sets the wrap mode (0=none, 1=char, 2=word).
func (v *TextBufferView) SetWrapMode(mode uint32) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.wrapMode = mode
	v.rewrap()
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
		LineStartCols:     lineStartCols,
		LineWidthCols:     lineWidthCols,
		LineWidthColsMax:  maxWidth,
		LineSources:       lineSources,
		LineWraps:         lineWraps,
	}
}

// GetLineCount returns the number of logical lines.
func (v *TextBufferView) GetLineCount() int {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.buffer.LineCount()
}

// GetVirtualLineCount returns the number of virtual lines.
func (v *TextBufferView) GetVirtualLineCount() int {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return len(v.virtualLine)
}

// GetScrollY returns the current scroll Y.
func (v *TextBufferView) GetScrollY() int {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.scrollY
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

// rewrap rewraps the buffer content into virtual lines.
func (v *TextBufferView) rewrap() {
	content := v.buffer.GetContent()
	if v.wrapMode == 0 || v.wrapWidth <= 0 {
		v.virtualLine = splitLines(content)
		return
	}

	var lines []string
	for _, line := range splitLines(content) {
		if line == "" {
			lines = append(lines, "")
			continue
		}
		if v.wrapMode == 1 {
			runes := []rune(line)
			for i := 0; i < len(runes); i += v.wrapWidth {
				end := i + v.wrapWidth
				if end > len(runes) {
					end = len(runes)
				}
				lines = append(lines, string(runes[i:end]))
			}
		} else {
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

// Rope is a rope data structure for efficient text editing.
type Rope struct {
	mu     sync.RWMutex
	root   *RopeNode
	weight int // total byte length
}

// RopeNode is a node in the rope tree.
type RopeNode struct {
	left  *RopeNode
	right *RopeNode
	value string
	weight int // byte length of left subtree (or value if leaf)
}

// NewRope creates a new empty Rope.
func NewRope() *Rope {
	return &Rope{}
}

// Insert inserts text at the given position.
func (r *Rope) Insert(pos int, text string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if pos < 0 || pos > r.weight {
		return errors.New("position out of range")
	}
	if !utf8.ValidString(text) {
		return errors.New("invalid UTF-8")
	}
	r.root = r.insertNode(r.root, pos, text)
	r.weight += len(text)
	return nil
}

// insertNode recursively inserts text into the rope.
func (r *Rope) insertNode(node *RopeNode, pos int, text string) *RopeNode {
	if node == nil {
		return &RopeNode{value: text, weight: len(text)}
	}
	if pos <= node.weight {
		node.left = r.insertNode(node.left, pos, text)
		node.weight += len(text)
	} else {
		node.right = r.insertNode(node.right, pos-node.weight, text)
	}
	return node
}

// GetString returns the full string content of the rope.
func (r *Rope) GetString() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.toString(r.root)
}

func (r *Rope) toString(node *RopeNode) string {
	if node == nil {
		return ""
	}
	if node.left == nil && node.right == nil {
		return node.value
	}
	return r.toString(node.left) + r.toString(node.right)
}

// Length returns the total length in bytes.
func (r *Rope) Length() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.weight
}

// StringWidth returns the display width of a UTF-8 string.
func StringWidth(s string) int {
	width := 0
	for _, r := range s {
		width += RuneWidth(r)
	}
	return width
}

// RuneWidth returns the display width of a rune (simplified wcwidth).
func RuneWidth(r rune) int {
	if r < 32 || r == 127 {
		return 0
	}
	// CJK and other wide characters
	if (r >= 0x1100 && r <= 0x115f) ||
		(r >= 0x11a3 && r <= 0x11a7) ||
		(r >= 0x11fa && r <= 0x11ff) ||
		(r >= 0x231a && r <= 0x231b) ||
		(r >= 0x2329 && r <= 0x232a) ||
		(r >= 0x2e80 && r <= 0x2e99) ||
		(r >= 0x2e9b && r <= 0x2ef3) ||
		(r >= 0x2f00 && r <= 0x2fd5) ||
		(r >= 0x2ff0 && r <= 0x2ffb) ||
		(r >= 0x3000 && r <= 0x303e) ||
		(r >= 0x3041 && r <= 0x3096) ||
		(r >= 0x309d && r <= 0x309e) ||
		(r >= 0x30a0 && r <= 0x30fa) ||
		(r >= 0x30fc && r <= 0x30ff) ||
		(r >= 0x3105 && r <= 0x312d) ||
		(r >= 0x3131 && r <= 0x318e) ||
		(r >= 0x3190 && r <= 0x31ba) ||
		(r >= 0x31c0 && r <= 0x31e3) ||
		(r >= 0x31f0 && r <= 0x321e) ||
		(r >= 0x3220 && r <= 0x3247) ||
		(r >= 0x3250 && r <= 0x32fe) ||
		(r >= 0x3300 && r <= 0x4dbf) ||
		(r >= 0x4e00 && r <= 0xa48c) ||
		(r >= 0xa490 && r <= 0xa4c6) ||
		(r >= 0xa960 && r <= 0xa97c) ||
		(r >= 0xac00 && r <= 0xd7a3) ||
		(r >= 0xd7b0 && r <= 0xd7c6) ||
		(r >= 0xd7cb && r <= 0xd7fb) ||
		(r >= 0xf900 && r <= 0xfaff) ||
		(r >= 0xfe10 && r <= 0xfe19) ||
		(r >= 0xfe30 && r <= 0xfe6b) ||
		(r >= 0xff01 && r <= 0xff60) ||
		(r >= 0xffe0 && r <= 0xffe6) ||
		(r >= 0xffe8 && r <= 0xffe9) ||
		(r >= 0x1f200 && r <= 0x1f202) ||
		(r >= 0x1f210 && r <= 0x1f23a) ||
		(r >= 0x1f240 && r <= 0x1f248) ||
		(r >= 0x1f250 && r <= 0x1f251) ||
		(r >= 0x1f300 && r <= 0x1f320) ||
		(r >= 0x1f32d && r <= 0x1f335) ||
		(r >= 0x1f337 && r <= 0x1f37c) ||
		(r >= 0x1f37e && r <= 0x1f393) ||
		(r >= 0x1f3a0 && r <= 0x1f3ca) ||
		(r >= 0x1f3cd && r <= 0x1f3cd) ||
		(r >= 0x1f3ce && r <= 0x1f3ce) ||
		(r >= 0x1f3cf && r <= 0x1f3cf) ||
		(r >= 0x1f3d0 && r <= 0x1f3e3) ||
		(r >= 0x1f3e4 && r <= 0x1f3e4) ||
		(r >= 0x1f3e5 && r <= 0x1f3e5) ||
		(r >= 0x1f3e6 && r <= 0x1f3e6) ||
		(r >= 0x1f3e7 && r <= 0x1f3e7) ||
		(r >= 0x1f3e8 && r <= 0x1f3e8) ||
		(r >= 0x1f3e9 && r <= 0x1f3e9) ||
		(r >= 0x1f3ea && r <= 0x1f3ea) ||
		(r >= 0x1f3eb && r <= 0x1f3eb) ||
		(r >= 0x1f3ec && r <= 0x1f3ec) ||
		(r >= 0x1f3ed && r <= 0x1f3ed) ||
		(r >= 0x1f3ee && r <= 0x1f3ee) ||
		(r >= 0x1f3f0 && r <= 0x1f3f0) ||
		(r >= 0x1f3f3 && r <= 0x1f3f3) ||
		(r >= 0x1f3f4 && r <= 0x1f3f4) ||
		(r >= 0x1f3f5 && r <= 0x1f3f5) ||
		(r >= 0x1f3f6 && r <= 0x1f3f6) ||
		(r >= 0x1f3f7 && r <= 0x1f3f7) ||
		(r >= 0x1f3f8 && r <= 0x1f3f8) ||
		(r >= 0x1f3f9 && r <= 0x1f3f9) ||
		(r >= 0x1f3fa && r <= 0x1f3fa) ||
		(r >= 0x1f3fb && r <= 0x1f3fb) ||
		(r >= 0x1f3fc && r <= 0x1f3fc) ||
		(r >= 0x1f3fd && r <= 0x1f3fd) ||
		(r >= 0x1f3fe && r <= 0x1f3fe) ||
		(r >= 0x1f3ff && r <= 0x1f3ff) ||
		(r >= 0x1f400 && r <= 0x1f43e) ||
		(r >= 0x1f440 && r <= 0x1f440) ||
		(r >= 0x1f442 && r <= 0x1f4fc) ||
		(r >= 0x1f4fd && r <= 0x1f4ff) ||
		(r >= 0x1f500 && r <= 0x1f53d) ||
		(r >= 0x1f53e && r <= 0x1f544) ||
		(r >= 0x1f545 && r <= 0x1f545) ||
		(r >= 0x1f546 && r <= 0x1f546) ||
		(r >= 0x1f547 && r <= 0x1f547) ||
		(r >= 0x1f548 && r <= 0x1f548) ||
		(r >= 0x1f549 && r <= 0x1f549) ||
		(r >= 0x1f54a && r <= 0x1f54a) ||
		(r >= 0x1f54b && r <= 0x1f54b) ||
		(r >= 0x1f54c && r <= 0x1f54c) ||
		(r >= 0x1f54d && r <= 0x1f54d) ||
		(r >= 0x1f54e && r <= 0x1f54e) ||
		(r >= 0x1f54f && r <= 0x1f54f) ||
		(r >= 0x1f550 && r <= 0x1f567) ||
		(r >= 0x1f568 && r <= 0x1f568) ||
		(r >= 0x1f569 && r <= 0x1f569) ||
		(r >= 0x1f56a && r <= 0x1f56a) ||
		(r >= 0x1f56b && r <= 0x1f56b) ||
		(r >= 0x1f56c && r <= 0x1f56c) ||
		(r >= 0x1f56d && r <= 0x1f56d) ||
		(r >= 0x1f56e && r <= 0x1f56e) ||
		(r >= 0x1f56f && r <= 0x1f56f) ||
		(r >= 0x1f570 && r <= 0x1f570) ||
		(r >= 0x1f571 && r <= 0x1f571) ||
		(r >= 0x1f572 && r <= 0x1f572) ||
		(r >= 0x1f573 && r <= 0x1f573) ||
		(r >= 0x1f574 && r <= 0x1f574) ||
		(r >= 0x1f575 && r <= 0x1f575) ||
		(r >= 0x1f576 && r <= 0x1f576) ||
		(r >= 0x1f577 && r <= 0x1f577) ||
		(r >= 0x1f578 && r <= 0x1f578) ||
		(r >= 0x1f579 && r <= 0x1f579) ||
		(r >= 0x1f57a && r <= 0x1f57a) ||
		(r >= 0x1f57b && r <= 0x1f57b) ||
		(r >= 0x1f57c && r <= 0x1f57c) ||
		(r >= 0x1f57d && r <= 0x1f57d) ||
		(r >= 0x1f57e && r <= 0x1f57e) ||
		(r >= 0x1f57f && r <= 0x1f57f) ||
		(r >= 0x1f580 && r <= 0x1f580) ||
		(r >= 0x1f581 && r <= 0x1f581) ||
		(r >= 0x1f582 && r <= 0x1f582) ||
		(r >= 0x1f583 && r <= 0x1f583) ||
		(r >= 0x1f584 && r <= 0x1f584) ||
		(r >= 0x1f585 && r <= 0x1f585) ||
		(r >= 0x1f586 && r <= 0x1f586) ||
		(r >= 0x1f587 && r <= 0x1f587) ||
		(r >= 0x1f588 && r <= 0x1f588) ||
		(r >= 0x1f589 && r <= 0x1f589) ||
		(r >= 0x1f58a && r <= 0x1f58a) ||
		(r >= 0x1f58b && r <= 0x1f58b) ||
		(r >= 0x1f58c && r <= 0x1f58c) ||
		(r >= 0x1f58d && r <= 0x1f58d) ||
		(r >= 0x1f58e && r <= 0x1f58e) ||
		(r >= 0x1f58f && r <= 0x1f58f) ||
		(r >= 0x1f590 && r <= 0x1f590) ||
		(r >= 0x1f591 && r <= 0x1f591) ||
		(r >= 0x1f592 && r <= 0x1f592) ||
		(r >= 0x1f593 && r <= 0x1f593) ||
		(r >= 0x1f594 && r <= 0x1f594) ||
		(r >= 0x1f595 && r <= 0x1f595) ||
		(r >= 0x1f596 && r <= 0x1f596) ||
		(r >= 0x1f597 && r <= 0x1f597) ||
		(r >= 0x1f598 && r <= 0x1f598) ||
		(r >= 0x1f599 && r <= 0x1f599) ||
		(r >= 0x1f59a && r <= 0x1f59a) ||
		(r >= 0x1f59b && r <= 0x1f59b) ||
		(r >= 0x1f59c && r <= 0x1f59c) ||
		(r >= 0x1f59d && r <= 0x1f59d) ||
		(r >= 0x1f59e && r <= 0x1f59e) ||
		(r >= 0x1f59f && r <= 0x1f59f) ||
		(r >= 0x1f5a0 && r <= 0x1f5a0) ||
		(r >= 0x1f5a1 && r <= 0x1f5a1) ||
		(r >= 0x1f5a2 && r <= 0x1f5a2) ||
		(r >= 0x1f5a3 && r <= 0x1f5a3) ||
		(r >= 0x1f5a4 && r <= 0x1f5a4) ||
		(r >= 0x1f5a5 && r <= 0x1f5a5) ||
		(r >= 0x1f5a6 && r <= 0x1f5a6) ||
		(r >= 0x1f5a7 && r <= 0x1f5a7) ||
		(r >= 0x1f5a8 && r <= 0x1f5a8) ||
		(r >= 0x1f5a9 && r <= 0x1f5a9) ||
		(r >= 0x1f5aa && r <= 0x1f5aa) ||
		(r >= 0x1f5ab && r <= 0x1f5ab) ||
		(r >= 0x1f5ac && r <= 0x1f5ac) ||
		(r >= 0x1f5ad && r <= 0x1f5ad) ||
		(r >= 0x1f5ae && r <= 0x1f5ae) ||
		(r >= 0x1f5af && r <= 0x1f5af) ||
		(r >= 0x1f5b0 && r <= 0x1f5b0) ||
		(r >= 0x1f5b1 && r <= 0x1f5b1) ||
		(r >= 0x1f5b2 && r <= 0x1f5b2) ||
		(r >= 0x1f5b3 && r <= 0x1f5b3) ||
		(r >= 0x1f5b4 && r <= 0x1f5b4) ||
		(r >= 0x1f5b5 && r <= 0x1f5b5) ||
		(r >= 0x1f5b6 && r <= 0x1f5b6) ||
		(r >= 0x1f5b7 && r <= 0x1f5b7) ||
		(r >= 0x1f5b8 && r <= 0x1f5b8) ||
		(r >= 0x1f5b9 && r <= 0x1f5b9) ||
		(r >= 0x1f5ba && r <= 0x1f5ba) ||
		(r >= 0x1f5bb && r <= 0x1f5bb) ||
		(r >= 0x1f5bc && r <= 0x1f5bc) ||
		(r >= 0x1f5bd && r <= 0x1f5bd) ||
		(r >= 0x1f5be && r <= 0x1f5be) ||
		(r >= 0x1f5bf && r <= 0x1f5bf) ||
		(r >= 0x1f5c0 && r <= 0x1f5c0) ||
		(r >= 0x1f5c1 && r <= 0x1f5c1) ||
		(r >= 0x1f5c2 && r <= 0x1f5c2) ||
		(r >= 0x1f5c3 && r <= 0x1f5c3) ||
		(r >= 0x1f5c4 && r <= 0x1f5c4) {
		return 2
	}
	return 1
}

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
var _ = fmt.Sprintf
