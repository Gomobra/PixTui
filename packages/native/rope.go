package native

import (
	"errors"
	"strings"
	"unicode/utf8"
)

// Rope is a persistent rope data structure for efficient text editing.
//
// It is a port of the Zig Rope in packages/native/src/rope.zig. Text is stored
// as a binary tree of leaf nodes holding string slices, which makes
// insertions and deletions at arbitrary positions cheaper than rewriting a
// flat string. The structure is persistent in spirit: operations rebuild only
// the nodes along the affected path, sharing the rest of the tree.
//
// All positions are byte offsets into the logical content.
type Rope struct {
	root      *ropeNode
	weight    int // total byte length of the logical content
	undoStack []ropeState
	redoStack []ropeState
}

// ropeState is a snapshot of the rope used for undo/redo.
type ropeState struct {
	root   *ropeNode
	weight int
}

// ropeNode is a single node in the rope tree.
type ropeNode struct {
	left  *ropeNode
	right *ropeNode
	value string // leaf value; empty for branch nodes
	// weight is the total byte length of the left subtree for a branch, or
	// len(value) for a leaf.
	weight int
	isLeaf bool
}

// NewRope creates a new empty Rope.
func NewRope() *Rope {
	return &Rope{}
}

// NewRopeWithContent creates a new Rope initialized with the given content.
func NewRopeWithContent(content string) *Rope {
	r := &Rope{}
	if len(content) > 0 {
		r.root = &ropeNode{isLeaf: true, value: content, weight: len(content)}
		r.weight = len(content)
	}
	return r
}

// Insert inserts text at the given byte position.
func (r *Rope) Insert(pos int, text string) error {
	if pos < 0 || pos > r.weight {
		return errors.New("position out of range")
	}
	if !utf8.ValidString(text) {
		return errors.New("invalid UTF-8")
	}
	if len(text) == 0 {
		return nil
	}
	r.root = insertNode(r.root, pos, text)
	r.weight += len(text)
	return nil
}

// insertNode inserts text at position pos within the subtree rooted at node.
// pos is measured in bytes relative to the start of node's content.
func insertNode(node *ropeNode, pos int, text string) *ropeNode {
	textLen := len(text)

	if node == nil {
		return &ropeNode{isLeaf: true, value: text, weight: textLen}
	}

	if node.isLeaf {
		// Split this leaf at pos and place the new text between the halves.
		before := node.value[:pos]
		after := node.value[pos:]

		left := &ropeNode{isLeaf: true, value: before, weight: len(before)}
		mid := &ropeNode{isLeaf: true, value: text, weight: textLen}
		right := &ropeNode{isLeaf: true, value: after, weight: len(after)}

		midRight := &ropeNode{
			left:   mid,
			right:  right,
			weight: textLen,
			isLeaf: false,
		}
		return &ropeNode{
			left:   left,
			right:  midRight,
			weight: len(before),
			isLeaf: false,
		}
	}

	if pos <= node.weight {
		node.left = insertNode(node.left, pos, text)
		node.weight += textLen
	} else {
		node.right = insertNode(node.right, pos-node.weight, text)
	}
	return node
}

// Delete removes the bytes in [start, end).
func (r *Rope) Delete(start, end int) error {
	if start < 0 || end > r.weight || start > end {
		return errors.New("invalid range")
	}
	if start == end {
		return nil
	}
	r.root = deleteRange(r.root, start, end)
	r.weight = totalBytes(r.root)
	return nil
}

// deleteRange removes [start, end) from the subtree rooted at node.
func deleteRange(node *ropeNode, start, end int) *ropeNode {
	if node == nil {
		return nil
	}

	if node.isLeaf {
		n := len(node.value)
		switch {
		case start == 0 && end >= n:
			return nil // delete the whole leaf
		case start == 0:
			return &ropeNode{isLeaf: true, value: node.value[end:], weight: n - end}
		case end >= n:
			return &ropeNode{isLeaf: true, value: node.value[:start], weight: start}
		default:
			return &ropeNode{
				isLeaf: true,
				value:  node.value[:start] + node.value[end:],
				weight: start,
			}
		}
	}

	if end <= node.weight {
		node.left = deleteRange(node.left, start, end)
		if node.left == nil {
			return node.right
		}
	} else if start >= node.weight {
		node.right = deleteRange(node.right, start-node.weight, end-node.weight)
		if node.right == nil {
			return node.left
		}
	} else {
		// The range straddles the boundary between left and right.
		node.left = deleteRange(node.left, start, node.weight)
		node.right = deleteRange(node.right, 0, end-node.weight)
		if node.left == nil {
			return node.right
		}
		if node.right == nil {
			return node.left
		}
	}
	return node
}

// Replace replaces the single byte at position pos with text.
// It is equivalent to Delete(pos, pos+1) followed by Insert(pos, text).
func (r *Rope) Replace(pos int, text string) error {
	if pos < 0 || pos >= r.weight {
		return errors.New("position out of range")
	}
	if !utf8.ValidString(text) {
		return errors.New("invalid UTF-8")
	}
	if err := r.Delete(pos, pos+1); err != nil {
		return err
	}
	return r.Insert(pos, text)
}

// Append appends text to the end of the rope.
func (r *Rope) Append(text string) error {
	return r.Insert(r.weight, text)
}

// Prepend prepends text to the start of the rope.
func (r *Rope) Prepend(text string) error {
	return r.Insert(0, text)
}

// Length returns the total byte length of the logical content.
func (r *Rope) Length() int {
	return r.weight
}

// GetString returns the full logical content as a string.
func (r *Rope) GetString() string {
	var sb strings.Builder
	writeRopeString(r.root, &sb)
	return sb.String()
}

func writeRopeString(node *ropeNode, sb *strings.Builder) {
	if node == nil {
		return
	}
	if node.isLeaf {
		sb.WriteString(node.value)
		return
	}
	writeRopeString(node.left, sb)
	writeRopeString(node.right, sb)
}

// GetByte returns the byte at the given offset, or false if out of range.
func (r *Rope) GetByte(offset int) (byte, bool) {
	if offset < 0 || offset >= r.weight {
		return 0, false
	}
	b, _ := getByte(r.root, offset)
	return b, true
}

func getByte(node *ropeNode, offset int) (byte, bool) {
	if node.isLeaf {
		return node.value[offset], true
	}
	if offset < node.weight {
		return getByte(node.left, offset)
	}
	return getByte(node.right, offset-node.weight)
}

// GetRuneAt returns the rune and its byte width at the given byte offset.
func (r *Rope) GetRuneAt(offset int) (rune, int, bool) {
	if offset < 0 || offset >= r.weight {
		return 0, 0, false
	}
	content := r.GetString()
	runeVal, size := utf8.DecodeRuneInString(content[offset:])
	if runeVal == utf8.RuneError && size == 0 {
		return 0, 0, false
	}
	return runeVal, size, true
}

// Slice returns the logical content in [start, end).
func (r *Rope) Slice(start, end int) string {
	if start < 0 {
		start = 0
	}
	if end > r.weight {
		end = r.weight
	}
	if start >= end {
		return ""
	}
	content := r.GetString()
	return content[start:end]
}

// Walk walks every leaf in order, calling f with the leaf value and its
// logical byte index. Returning false from f stops the walk.
func (r *Rope) Walk(ctx interface{}, f func(interface{}, string, int) bool) error {
	index := 0
	walkNode(r.root, ctx, f, &index)
	return nil
}

func walkNode(node *ropeNode, ctx interface{}, f func(interface{}, string, int) bool, index *int) bool {
	if node == nil {
		return true
	}
	if node.isLeaf {
		return f(ctx, node.value, *index)
	}
	if !walkNode(node.left, ctx, f, index) {
		return false
	}
	return walkNode(node.right, ctx, f, index)
}

// totalBytes returns the total byte length of the subtree rooted at node.
func totalBytes(node *ropeNode) int {
	if node == nil {
		return 0
	}
	if node.isLeaf {
		return len(node.value)
	}
	return node.weight + totalBytes(node.right)
}

// StoreUndo saves the current rope state to the undo stack with the given
// metadata. Mirrors rope.zig's store_undo.
func (r *Rope) StoreUndo(meta string) {
	r.undoStack = append(r.undoStack, ropeState{root: r.root, weight: r.weight})
	r.redoStack = nil // clear redo stack on new change
}

// CanUndo returns true if there is an undo state available.
func (r *Rope) CanUndo() bool {
	return len(r.undoStack) > 0
}

// CanRedo returns true if there is a redo state available.
func (r *Rope) CanRedo() bool {
	return len(r.redoStack) > 0
}

// Undo restores the previous rope state from the undo stack and returns its
// metadata. Mirrors rope.zig's undo.
func (r *Rope) Undo(meta string) (string, error) {
	if !r.CanUndo() {
		return "", errors.New("nothing to undo")
	}
	// Save current state to redo stack before restoring
	r.redoStack = append(r.redoStack, ropeState{root: r.root, weight: r.weight})
	// Pop the last undo state
	last := r.undoStack[len(r.undoStack)-1]
	r.undoStack = r.undoStack[:len(r.undoStack)-1]
	r.root = last.root
	r.weight = last.weight
	return meta, nil
}

// Redo restores the next rope state from the redo stack. Mirrors rope.zig's redo.
func (r *Rope) Redo() (string, error) {
	if !r.CanRedo() {
		return "", errors.New("nothing to redo")
	}
	last := r.redoStack[len(r.redoStack)-1]
	r.redoStack = r.redoStack[:len(r.redoStack)-1]
	r.undoStack = append(r.undoStack, ropeState{root: r.root, weight: r.weight})
	r.root = last.root
	r.weight = last.weight
	return "", nil
}

// ClearHistory clears the undo and redo stacks.
func (r *Rope) ClearHistory() {
	r.undoStack = nil
	r.redoStack = nil
}

var _ = errors.New
