package renderables

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/Gomobra/PixTui/packages/core"
	"github.com/Gomobra/PixTui/packages/native"
)

// ---- StyledText for placeholder ----

// StyledTextChunk is a single styled run of text.
type StyledTextChunk struct {
	Text   string
	FG     *core.RGBA
	BG     *core.RGBA
	Attrs  uint32
}

// StyledText is a list of styled chunks.
type StyledText struct {
	Chunks []StyledTextChunk
}

// ---- Textarea action types ----

// TextareaAction is an action triggered by a key binding.
type TextareaAction string

const (
	ActionMoveLeft          TextareaAction = "move-left"
	ActionMoveRight         TextareaAction = "move-right"
	ActionMoveUp            TextareaAction = "move-up"
	ActionMoveDown          TextareaAction = "move-down"
	ActionSelectLeft        TextareaAction = "select-left"
	ActionSelectRight       TextareaAction = "select-right"
	ActionSelectUp          TextareaAction = "select-up"
	ActionSelectDown        TextareaAction = "select-down"
	ActionLineHome          TextareaAction = "line-home"
	ActionLineEnd           TextareaAction = "line-end"
	ActionSelectLineHome    TextareaAction = "select-line-home"
	ActionSelectLineEnd     TextareaAction = "select-line-end"
	ActionVisualLineHome    TextareaAction = "visual-line-home"
	ActionVisualLineEnd     TextareaAction = "visual-line-end"
	ActionSelectVisualLineHome TextareaAction = "select-visual-line-home"
	ActionSelectVisualLineEnd TextareaAction = "select-visual-line-end"
	ActionBufferHome        TextareaAction = "buffer-home"
	ActionBufferEnd         TextareaAction = "buffer-end"
	ActionSelectBufferHome  TextareaAction = "select-buffer-home"
	ActionSelectBufferEnd   TextareaAction = "select-buffer-end"
	ActionDeleteLine        TextareaAction = "delete-line"
	ActionDeleteToLineEnd   TextareaAction = "delete-to-line-end"
	ActionDeleteToLineStart TextareaAction = "delete-to-line-start"
	ActionBackspace         TextareaAction = "backspace"
	ActionDelete            TextareaAction = "delete"
	ActionNewline           TextareaAction = "newline"
	ActionUndo              TextareaAction = "undo"
	ActionRedo              TextareaAction = "redo"
	ActionWordForward       TextareaAction = "word-forward"
	ActionWordBackward      TextareaAction = "word-backward"
	ActionSelectWordForward TextareaAction = "select-word-forward"
	ActionSelectWordBackward TextareaAction = "select-word-backward"
	ActionDeleteWordForward TextareaAction = "delete-word-forward"
	ActionDeleteWordBackward TextareaAction = "delete-word-backward"
	ActionSelectAll         TextareaAction = "select-all"
	ActionSubmit            TextareaAction = "submit"
)

// KeyBinding maps a key event to an action.
type KeyBinding struct {
	Name      string
	Action    TextareaAction
	Shift     bool
	Ctrl      bool
	Alt       bool
	Super     bool
}

// TextareaWrapMode defines how text is wrapped.
type TextareaWrapMode string

const (
	WrapNone  TextareaWrapMode = "none"
	WrapChar  TextareaWrapMode = "char"
	WrapWord  TextareaWrapMode = "word"
)

// TextareaOptions contains options for TextareaRenderable.
type TextareaOptions struct {
	core.RenderableOptions[core.Renderable]
	Value               string
	Placeholder         string
	PlaceholderStyled   *StyledText
	PlaceholderColor    *core.RGBA
	ReadOnly            bool
	Wrap                bool
	WrapMode            TextareaWrapMode
	ShowLineNumbers     bool
	SyntaxStyle         string
	FocusedBackground   *core.RGBA
	FocusedForeground   *core.RGBA
	UnfocusedBackground *core.RGBA
	UnfocusedForeground *core.RGBA
	KeyBindings         []KeyBinding
	OnSubmit            func()
}

// cursorPos represents a cursor position as (line, byteCol).
type cursorPos struct {
	line    int
	byteCol int
}

// textareaSnapshot stores state for undo/redo.
type textareaSnapshot struct {
	value    string
	cursor   cursorPos
	selStart cursorPos
	selEnd   cursorPos
	scrollX  int
	scrollY  int
}

// TextareaRenderable is a multi-line text area with editing, cursor, and
// selection support.
type TextareaRenderable struct {
	*core.Renderable
	options TextareaOptions
	value   string
	lines   []string
	cursor  cursorPos
	selStart cursorPos
	selEnd   cursorPos
	scrollX int
	scrollY int
	dirty   bool

	history []textareaSnapshot
	future  []textareaSnapshot
	clipboard *native.Clipboard

	editorView *core.EditorView
	extmarks   *core.ExtmarkManager

	focused     bool
	keyBindings map[[4]bool]TextareaAction // modifiers+key -> action
	actionHandlers map[TextareaAction]func() bool
}

// NewTextareaRenderable creates a new TextareaRenderable.
func NewTextareaRenderable(ctx *core.RenderContext, opts TextareaOptions) *TextareaRenderable {
	base := core.NewRenderable(ctx, opts.ID, core.RenderableOptions[core.Renderable]{
		Width:           opts.Width,
		Height:          opts.Height,
		Visible:         opts.Visible,
		ZIndex:          opts.ZIndex,
		ForegroundColor: opts.ForegroundColor,
		BackgroundColor: opts.BackgroundColor,
		RenderBefore:    opts.RenderBefore,
		RenderAfter:     opts.RenderAfter,
		OnMouse:         opts.OnMouse,
		OnMouseDown:     opts.OnMouseDown,
		OnMouseUp:       opts.OnMouseUp,
		OnMouseMove:     opts.OnMouseMove,
		OnMouseDrag:     opts.OnMouseDrag,
		OnMouseDragEnd:  opts.OnMouseDragEnd,
		OnMouseDrop:     opts.OnMouseDrop,
		OnMouseOver:     opts.OnMouseOver,
		OnMouseOut:      opts.OnMouseOut,
		OnMouseScroll:   opts.OnMouseScroll,
		OnPaste:         opts.OnPaste,
		OnKeyDown:       opts.OnKeyDown,
	})
	r := &TextareaRenderable{
		Renderable: base,
		options:    opts,
		value:      opts.Value,
		clipboard:  native.NewClipboard(),
		editorView: core.NewEditorView(core.NewEditBuffer()),
		extmarks:   core.NewExtmarkManager(),
		keyBindings: make(map[[4]bool]TextareaAction),
		actionHandlers: make(map[TextareaAction]func() bool),
	}
	r.SetFocusable(true)
	r.rebuildLines()
	r.buildDefaultKeyBindings()
	r.buildActionHandlers()
	if opts.Value != "" {
		r.editorView.SetContent(opts.Value)
	}
	// Create native scene node
	if ctx != nil && ctx.NativeScene != nil {
		node, err := ctx.NativeScene.CreateNode(r.Renderable, native.SceneNodeKindText)
		if err == nil {
			r.Renderable.SetNativeNode(node)
		}
	}
	return r
}

func (r *TextareaRenderable) buildDefaultKeyBindings() {
	defaults := []KeyBinding{
		{Name: "left", Action: ActionMoveLeft},
		{Name: "right", Action: ActionMoveRight},
		{Name: "up", Action: ActionMoveUp},
		{Name: "down", Action: ActionMoveDown},
		{Name: "left", Shift: true, Action: ActionSelectLeft},
		{Name: "right", Shift: true, Action: ActionSelectRight},
		{Name: "up", Shift: true, Action: ActionSelectUp},
		{Name: "down", Shift: true, Action: ActionSelectDown},
		{Name: "home", Action: ActionBufferHome},
		{Name: "end", Action: ActionBufferEnd},
		{Name: "home", Shift: true, Action: ActionSelectBufferHome},
		{Name: "end", Shift: true, Action: ActionSelectBufferEnd},
		{Name: "a", Ctrl: true, Action: ActionLineHome},
		{Name: "e", Ctrl: true, Action: ActionLineEnd},
		{Name: "a", Ctrl: true, Shift: true, Action: ActionSelectLineHome},
		{Name: "e", Ctrl: true, Shift: true, Action: ActionSelectLineEnd},
		{Name: "f", Ctrl: true, Action: ActionMoveRight},
		{Name: "b", Ctrl: true, Action: ActionMoveLeft},
		{Name: "w", Ctrl: true, Action: ActionDeleteWordBackward},
		{Name: "backspace", Ctrl: true, Action: ActionDeleteWordBackward},
		{Name: "d", Super: true, Action: ActionDeleteWordForward},
		{Name: "delete", Super: true, Action: ActionDeleteWordForward},
		{Name: "delete", Ctrl: true, Action: ActionDeleteWordForward},
		{Name: "d", Ctrl: true, Shift: true, Action: ActionDeleteLine},
		{Name: "k", Ctrl: true, Action: ActionDeleteToLineEnd},
		{Name: "u", Ctrl: true, Action: ActionDeleteToLineStart},
		{Name: "backspace", Action: ActionBackspace},
		{Name: "backspace", Shift: true, Action: ActionBackspace},
		{Name: "d", Ctrl: true, Action: ActionDelete},
		{Name: "delete", Action: ActionDelete},
		{Name: "delete", Shift: true, Action: ActionDelete},
		{Name: "return", Action: ActionNewline},
		{Name: "kpenter", Action: ActionNewline},
		{Name: "linefeed", Action: ActionNewline},
		{Name: "return", Super: true, Action: ActionSubmit},
		{Name: "kpenter", Super: true, Action: ActionSubmit},
		{Name: "z", Ctrl: true, Action: ActionUndo},
		{Name: "z", Super: true, Action: ActionUndo},
		{Name: "z", Super: true, Shift: true, Action: ActionRedo},
		{Name: ".", Ctrl: true, Action: ActionRedo},
		{Name: "f", Super: true, Action: ActionWordForward},
		{Name: "b", Super: true, Action: ActionWordBackward},
		{Name: "right", Super: true, Action: ActionWordForward},
		{Name: "left", Super: true, Action: ActionWordBackward},
		{Name: "right", Ctrl: true, Action: ActionWordForward},
		{Name: "left", Ctrl: true, Action: ActionWordBackward},
		{Name: "f", Super: true, Shift: true, Action: ActionSelectWordForward},
		{Name: "b", Super: true, Shift: true, Action: ActionSelectWordBackward},
		{Name: "right", Super: true, Shift: true, Action: ActionSelectWordForward},
		{Name: "left", Super: true, Shift: true, Action: ActionSelectWordBackward},
		{Name: "backspace", Super: true, Action: ActionDeleteWordBackward},
		{Name: "left", Super: true, Action: ActionVisualLineHome},
		{Name: "right", Super: true, Action: ActionVisualLineEnd},
		{Name: "up", Super: true, Action: ActionBufferHome},
		{Name: "down", Super: true, Action: ActionBufferEnd},
		{Name: "left", Super: true, Shift: true, Action: ActionSelectVisualLineHome},
		{Name: "right", Super: true, Shift: true, Action: ActionSelectVisualLineEnd},
		{Name: "up", Super: true, Shift: true, Action: ActionSelectBufferHome},
		{Name: "down", Super: true, Shift: true, Action: ActionSelectBufferEnd},
		{Name: "a", Super: true, Action: ActionSelectAll},
	}
	for _, kb := range defaults {
		key := [4]bool{kb.Shift, kb.Ctrl, kb.Alt, kb.Super}
		r.keyBindings[key] = kb.Action
		// Also store by name
		_ = kb.Name
	}
}

func (r *TextareaRenderable) buildActionHandlers() {
	r.actionHandlers[ActionMoveLeft] = func() bool { r.moveLeft(false); return true }
	r.actionHandlers[ActionMoveRight] = func() bool { r.moveRight(false); return true }
	r.actionHandlers[ActionMoveUp] = func() bool { r.moveUp(false); return true }
	r.actionHandlers[ActionMoveDown] = func() bool { r.moveDown(false); return true }
	r.actionHandlers[ActionSelectLeft] = func() bool { r.moveLeft(true); return true }
	r.actionHandlers[ActionSelectRight] = func() bool { r.moveRight(true); return true }
	r.actionHandlers[ActionSelectUp] = func() bool { r.moveUp(true); return true }
	r.actionHandlers[ActionSelectDown] = func() bool { r.moveDown(true); return true }
	r.actionHandlers[ActionLineHome] = func() bool { r.gotoLineHome(false); return true }
	r.actionHandlers[ActionLineEnd] = func() bool { r.gotoLineEnd(false); return true }
	r.actionHandlers[ActionSelectLineHome] = func() bool { r.gotoLineHome(true); return true }
	r.actionHandlers[ActionSelectLineEnd] = func() bool { r.gotoLineEnd(true); return true }
	r.actionHandlers[ActionVisualLineHome] = func() bool { r.gotoVisualLineHome(false); return true }
	r.actionHandlers[ActionVisualLineEnd] = func() bool { r.gotoVisualLineEnd(false); return true }
	r.actionHandlers[ActionSelectVisualLineHome] = func() bool { r.gotoVisualLineHome(true); return true }
	r.actionHandlers[ActionSelectVisualLineEnd] = func() bool { r.gotoVisualLineEnd(true); return true }
	r.actionHandlers[ActionBufferHome] = func() bool { r.gotoBufferHome(false); return true }
	r.actionHandlers[ActionBufferEnd] = func() bool { r.gotoBufferEnd(false); return true }
	r.actionHandlers[ActionSelectBufferHome] = func() bool { r.gotoBufferHome(true); return true }
	r.actionHandlers[ActionSelectBufferEnd] = func() bool { r.gotoBufferEnd(true); return true }
	r.actionHandlers[ActionDeleteLine] = func() bool { r.DeleteLine(r.cursor.line); return true }
	r.actionHandlers[ActionDeleteToLineEnd] = func() bool { r.deleteToLineEnd(); return true }
	r.actionHandlers[ActionDeleteToLineStart] = func() bool { r.deleteToLineStart(); return true }
	r.actionHandlers[ActionBackspace] = func() bool { r.backspace(); return true }
	r.actionHandlers[ActionDelete] = func() bool { r.deleteForward(); return true }
	r.actionHandlers[ActionNewline] = func() bool { r.insertNewline(); return true }
	r.actionHandlers[ActionUndo] = func() bool { return r.Undo() }
	r.actionHandlers[ActionRedo] = func() bool { return r.Redo() }
	r.actionHandlers[ActionWordForward] = func() bool { r.moveWordForward(false); return true }
	r.actionHandlers[ActionWordBackward] = func() bool { r.moveWordBackward(false); return true }
	r.actionHandlers[ActionSelectWordForward] = func() bool { r.moveWordForward(true); return true }
	r.actionHandlers[ActionSelectWordBackward] = func() bool { r.moveWordBackward(true); return true }
	r.actionHandlers[ActionDeleteWordForward] = func() bool { r.deleteWordForward(); return true }
	r.actionHandlers[ActionDeleteWordBackward] = func() bool { r.deleteWordBackward(); return true }
	r.actionHandlers[ActionSelectAll] = func() bool { r.selectAll(); return true }
	r.actionHandlers[ActionSubmit] = func() bool { r.submit(); return true }
}

func (r *TextareaRenderable) submit() {
	if r.options.OnSubmit != nil {
		r.options.OnSubmit()
	}
}

// Value returns the textarea content.
func (r *TextareaRenderable) Value() string {
	return r.value
}

// SetValue sets the textarea content and requests a re-render.
func (r *TextareaRenderable) SetValue(value string) {
	r.pushHistory()
	r.value = value
	r.options.Value = value
	r.cursor = cursorPos{0, 0}
	r.selStart = cursorPos{0, 0}
	r.selEnd = cursorPos{0, 0}
	r.scrollX = 0
	r.scrollY = 0
	r.rebuildLines()
	r.RequestRender()
}

// Options returns the textarea options.
func (r *TextareaRenderable) Options() TextareaOptions {
	return r.options
}

// SetOptions updates the textarea options and requests a re-render.
func (r *TextareaRenderable) SetOptions(opts TextareaOptions) {
	r.options = opts
	r.value = opts.Value
	r.rebuildLines()
	r.RequestRender()
}

// Cursor returns the cursor position (line, byteCol).
func (r *TextareaRenderable) Cursor() (int, int) {
	return r.cursor.line, r.cursor.byteCol
}

// SetCursor sets the cursor position and clamps it to the document.
func (r *TextareaRenderable) SetCursor(line, byteCol int) {
	r.cursor = r.clampCursor(cursorPos{line, byteCol})
	r.selStart = r.cursor
	r.selEnd = r.cursor
	r.scrollToCursor()
	r.RequestRender()
}

// Selection returns the current selection start and end (line, byteCol).
func (r *TextareaRenderable) Selection() (int, int, int, int) {
	return r.selStart.line, r.selStart.byteCol, r.selEnd.line, r.selEnd.byteCol
}

// SetSelection sets the selection range.
func (r *TextareaRenderable) SetSelection(sline, scol, eline, ecol int) {
	r.selStart = r.clampCursor(cursorPos{sline, scol})
	r.selEnd = r.clampCursor(cursorPos{eline, ecol})
	r.cursor = r.selEnd
	r.RequestRender()
}

// ClearSelection clears the current selection.
func (r *TextareaRenderable) ClearSelection() {
	r.selStart = r.cursor
	r.selEnd = r.cursor
	r.RequestRender()
}

// Insert inserts text at the cursor position.
func (r *TextareaRenderable) Insert(text string) {
	if r.options.ReadOnly {
		return
	}
	if r.hasSelection() {
		r.deleteSelection()
	}
	r.pushHistory()
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	lines := splitLines(text)
	if len(lines) == 1 {
		line := r.cursor.line
		col := r.cursor.byteCol
		cur := r.lines[line]
		if col > len(cur) {
			col = len(cur)
		}
		before := cur[:col]
		after := cur[col:]
		r.lines[line] = before + lines[0] + after
		r.cursor.byteCol = col + len(lines[0])
	} else {
		line := r.cursor.line
		col := r.cursor.byteCol
		cur := r.lines[line]
		if col > len(cur) {
			col = len(cur)
		}
		before := cur[:col]
		after := cur[col:]
		r.lines[line] = before + lines[0]
		tail := make([]string, len(lines)-1)
		copy(tail, lines[1:])
		tail[len(tail)-1] = tail[len(tail)-1] + after
		r.lines = append(r.lines[:line+1], append(tail, r.lines[line+1:]...)...)
		r.cursor.line = line + len(lines) - 1
		r.cursor.byteCol = len(lines[len(lines)-1])
	}
	r.cursor = r.clampCursor(r.cursor)
	r.selStart = r.cursor
	r.selEnd = r.cursor
	r.syncValue()
	r.scrollToCursor()
	r.RequestRender()
}

// Delete removes the byte range [start, end) within the current line.
func (r *TextareaRenderable) Delete(start, end int) error {
	if r.options.ReadOnly {
		return nil
	}
	line := r.cursor.line
	if line < 0 || line >= len(r.lines) {
		return nil
	}
	cur := r.lines[line]
	if start < 0 || end > len(cur) || start > end {
		return nil
	}
	r.pushHistory()
	r.lines[line] = cur[:start] + cur[end:]
	if r.cursor.byteCol > end {
		r.cursor.byteCol = end
	}
	if r.cursor.byteCol > len(r.lines[line]) {
		r.cursor.byteCol = len(r.lines[line])
	}
	r.selStart = r.cursor
	r.selEnd = r.cursor
	r.syncValue()
	r.RequestRender()
	return nil
}

// InsertAtLine inserts text at the given line/byteCol.
func (r *TextareaRenderable) InsertAtLine(line, byteCol int, text string) {
	if r.options.ReadOnly {
		return
	}
	if line < 0 || line >= len(r.lines) {
		return
	}
	r.pushHistory()
	byteCol = clampInt(byteCol, 0, len(r.lines[line]))
	cur := r.lines[line]
	r.lines[line] = cur[:byteCol] + text + cur[byteCol:]
	r.syncValue()
	r.RequestRender()
}

// DeleteLine removes the given line index.
func (r *TextareaRenderable) DeleteLine(line int) error {
	if r.options.ReadOnly {
		return nil
	}
	if line < 0 || line >= len(r.lines) {
		return nil
	}
	r.pushHistory()
	r.lines = append(r.lines[:line], r.lines[line+1:]...)
	if len(r.lines) == 0 {
		r.lines = []string{""}
	}
	if r.cursor.line >= len(r.lines) {
		r.cursor.line = len(r.lines) - 1
	}
	r.cursor.byteCol = clampInt(r.cursor.byteCol, 0, len(r.lines[r.cursor.line]))
	r.selStart = r.cursor
	r.selEnd = r.cursor
	r.syncValue()
	r.RequestRender()
	return nil
}

// Undo reverts the last change.
func (r *TextareaRenderable) Undo() bool {
	if len(r.history) == 0 {
		return false
	}
	r.future = append(r.future, r.captureSnapshot())
	snap := r.history[len(r.history)-1]
	r.history = r.history[:len(r.history)-1]
	r.restoreSnapshot(snap)
	return true
}

// Redo reapplies the last undone change.
func (r *TextareaRenderable) Redo() bool {
	if len(r.future) == 0 {
		return false
	}
	r.history = append(r.history, r.captureSnapshot())
	snap := r.future[len(r.future)-1]
	r.future = r.future[:len(r.future)-1]
	r.restoreSnapshot(snap)
	return true
}

// hasSelection reports whether a non-empty selection is active.
func (r *TextareaRenderable) hasSelection() bool {
	return r.selStart.line != r.selEnd.line || r.selStart.byteCol != r.selEnd.byteCol
}

// selectionNormalized returns the selection ordered from start to end.
func (r *TextareaRenderable) selectionNormalized() (cursorPos, cursorPos) {
	start, end := r.selStart, r.selEnd
	if end.line < start.line || (end.line == start.line && end.byteCol < start.byteCol) {
		start, end = end, start
	}
	return start, end
}

// deleteSelection deletes the currently selected text.
func (r *TextareaRenderable) deleteSelection() {
	if !r.hasSelection() {
		return
	}
	r.pushHistory()
	start, end := r.selectionNormalized()
	if start.line == end.line {
		cur := r.lines[start.line]
		if start.byteCol > len(cur) {
			start.byteCol = len(cur)
		}
		if end.byteCol > len(cur) {
			end.byteCol = len(cur)
		}
		r.lines[start.line] = cur[:start.byteCol] + cur[end.byteCol:]
		r.cursor.byteCol = start.byteCol
		r.cursor.line = start.line
	} else {
		first := r.lines[start.line]
		last := r.lines[end.line]
		if start.byteCol > len(first) {
			start.byteCol = len(first)
		}
		if end.byteCol > len(last) {
			end.byteCol = len(last)
		}
		merged := first[:start.byteCol] + last[end.byteCol:]
		r.lines = append(r.lines[:start.line], append([]string{merged}, r.lines[end.line+1:]...)...)
		r.cursor.line = start.line
		r.cursor.byteCol = start.byteCol
	}
	r.cursor = r.clampCursor(r.cursor)
	r.selStart = r.cursor
	r.selEnd = r.cursor
	r.syncValue()
	r.RequestRender()
}

// syncValue rebuilds r.value from r.lines.
func (r *TextareaRenderable) syncValue() {
	var sb strings.Builder
	for i, line := range r.lines {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(line)
	}
	r.value = sb.String()
	r.options.Value = r.value
	r.dirty = true
}

// rebuildLines splits r.value into lines.
func (r *TextareaRenderable) rebuildLines() {
	r.lines = splitLines(r.value)
	if len(r.lines) == 0 {
		r.lines = []string{""}
	}
	r.dirty = true
}

// clampCursor clamps a cursor position to valid document bounds.
func (r *TextareaRenderable) clampCursor(pos cursorPos) cursorPos {
	if pos.line < 0 {
		pos.line = 0
	}
	if pos.line >= len(r.lines) {
		pos.line = len(r.lines) - 1
	}
	if pos.byteCol < 0 {
		pos.byteCol = 0
	}
	if pos.byteCol > len(r.lines[pos.line]) {
		pos.byteCol = len(r.lines[pos.line])
	}
	return pos
}

// byteColToX converts a byte column to a display column (accounting for tabs).
func (r *TextareaRenderable) byteColToX(line, byteCol int) int {
	if line < 0 || line >= len(r.lines) {
		return 0
	}
	x := 0
	for i, ch := range r.lines[line] {
		if i >= byteCol {
			break
		}
		if ch == '\t' {
			x += 4 - (x % 4)
		} else {
			x++
		}
	}
	return x
}

// handleKeyDown processes a key event and returns true if handled.
func (r *TextareaRenderable) handleKeyDown(key *core.KeyEvent) bool {
	if key == nil {
		return false
	}
	shift := key.Modifiers&native.ModShift != 0
	ctrl := key.Modifiers&native.ModCtrl != 0
	alt := key.Modifiers&native.ModAlt != 0
	super := key.Modifiers&native.ModSuper != 0

	mods := [4]bool{shift, ctrl, alt, super}
	if action, ok := r.keyBindings[mods]; ok {
		if handler, ok := r.actionHandlers[action]; ok {
			return handler()
		}
	}

	switch key.Key {
	case "Enter", "\n", "\r", "Return", "KPEnter", "Linefeed":
		r.insertNewline()
		return true
	case "Backspace":
		if ctrl || alt {
			r.deleteWordBackward()
			return true
		}
		r.backspace()
		return true
	case "Delete":
		if ctrl || alt {
			r.deleteWordForward()
			return true
		}
		r.deleteForward()
		return true
	case "ArrowLeft", "Left":
		r.moveLeft(shift)
		return true
	case "ArrowRight", "Right":
		r.moveRight(shift)
		return true
	case "ArrowUp", "Up":
		r.moveUp(shift)
		return true
	case "ArrowDown", "Down":
		r.moveDown(shift)
		return true
	case "Home":
		r.gotoLineHome(shift)
		return true
	case "End":
		r.gotoLineEnd(shift)
		return true
	case "PageUp":
		r.ScrollUp(r.viewportHeight())
		return true
	case "PageDown":
		r.ScrollDown(r.viewportHeight())
		return true
	case "a", "A":
		if ctrl && !shift {
			r.gotoLineHome(false)
			return true
		}
		if ctrl && shift {
			r.gotoLineHome(true)
			return true
		}
		if super {
			r.selectAll()
			return true
		}
	case "e", "E":
		if ctrl && !shift {
			r.gotoLineEnd(false)
			return true
		}
		if ctrl && shift {
			r.gotoLineEnd(true)
			return true
		}
	case "k", "K":
		if ctrl {
			r.deleteToLineEnd()
			return true
		}
	case "u", "U":
		if ctrl {
			r.deleteToLineStart()
			return true
		}
	case "w", "W":
		if ctrl {
			r.deleteWordBackward()
			return true
		}
	case "d", "D":
		if ctrl && shift {
			r.DeleteLine(r.cursor.line)
			return true
		}
	case "v", "V":
		if ctrl {
			text, _, err := r.clipboard.Paste()
			if err == nil {
				r.Insert(text)
			}
			return true
		}
	case "Insert":
		if shift {
			text, _, err := r.clipboard.Paste()
			if err == nil {
				r.Insert(text)
			}
			return true
		}
	}
	if !ctrl && !alt && !super {
		if key.Key == "Space" || key.Key == "space" {
			r.Insert(" ")
			return true
		}
		if len(key.Key) == 1 {
			r.Insert(key.Key)
			return true
		}
	}
	return false
}

// HandleKeyDown is the exported entry point for key handling.
func (r *TextareaRenderable) HandleKeyDown(key *core.KeyEvent) bool {
	return r.handleKeyDown(key)
}

// HandleKeyPress is an alias for HandleKeyDown.
func (r *TextareaRenderable) HandleKeyPress(key *core.KeyEvent) bool {
	return r.handleKeyDown(key)
}

func (r *TextareaRenderable) insertNewline() {
	if r.options.ReadOnly {
		return
	}
	if r.hasSelection() {
		r.deleteSelection()
	}
	r.pushHistory()
	cur := r.lines[r.cursor.line]
	col := r.cursor.byteCol
	if col > len(cur) {
		col = len(cur)
	}
	before := cur[:col]
	after := cur[col:]
	r.lines[r.cursor.line] = before
	r.lines = append(r.lines[:r.cursor.line+1], append([]string{after}, r.lines[r.cursor.line+1:]...)...)
	r.cursor.line++
	r.cursor.byteCol = 0
	r.selStart = r.cursor
	r.selEnd = r.cursor
	r.syncValue()
	r.scrollToCursor()
	r.RequestRender()
}

func (r *TextareaRenderable) backspace() {
	if r.options.ReadOnly {
		return
	}
	if r.hasSelection() {
		r.deleteSelection()
		return
	}
	r.pushHistory()
	if r.cursor.byteCol == 0 && r.cursor.line == 0 {
		return
	}
	if r.cursor.byteCol == 0 {
		prev := r.lines[r.cursor.line-1]
		cur := r.lines[r.cursor.line]
		r.lines[r.cursor.line-1] = prev + cur
		r.lines = append(r.lines[:r.cursor.line], r.lines[r.cursor.line+1:]...)
		r.cursor.line--
		r.cursor.byteCol = len(prev)
	} else {
		cur := r.lines[r.cursor.line]
		col := r.cursor.byteCol
		start := col
		for start > 0 {
			start--
			if cur[start] < 0x80 || cur[start] >= 0xC0 {
				break
			}
		}
		r.lines[r.cursor.line] = cur[:start] + cur[col:]
		r.cursor.byteCol = start
	}
	r.cursor = r.clampCursor(r.cursor)
	r.selStart = r.cursor
	r.selEnd = r.cursor
	r.syncValue()
	r.scrollToCursor()
	r.RequestRender()
}

func (r *TextareaRenderable) deleteForward() {
	if r.options.ReadOnly {
		return
	}
	if r.hasSelection() {
		r.deleteSelection()
		return
	}
	r.pushHistory()
	cur := r.lines[r.cursor.line]
	if r.cursor.byteCol >= len(cur) {
		if r.cursor.line+1 < len(r.lines) {
			next := r.lines[r.cursor.line+1]
			r.lines[r.cursor.line] = cur + next
			r.lines = append(r.lines[:r.cursor.line+1], r.lines[r.cursor.line+2:]...)
		}
	} else {
		col := r.cursor.byteCol
		end := col + 1
		for end < len(cur) && cur[end] >= 0x80 && cur[end] < 0xC0 {
			end++
		}
		r.lines[r.cursor.line] = cur[:col] + cur[end:]
	}
	r.cursor = r.clampCursor(r.cursor)
	r.selStart = r.cursor
	r.selEnd = r.cursor
	r.syncValue()
	r.scrollToCursor()
	r.RequestRender()
}

func (r *TextareaRenderable) moveLeft(sel bool) {
	anchor := r.cursor
	if sel && r.hasSelection() {
		if posEqual(r.cursor, r.selEnd) {
			anchor = r.selStart
		} else {
			anchor = r.selEnd
		}
	}
	if r.cursor.byteCol > 0 {
		cur := r.lines[r.cursor.line]
		col := r.cursor.byteCol
		start := col - 1
		for start > 0 && cur[start] >= 0x80 && cur[start] < 0xC0 {
			start--
		}
		r.cursor.byteCol = start
	} else if r.cursor.line > 0 {
		r.cursor.line--
		r.cursor.byteCol = len(r.lines[r.cursor.line])
	}
	r.cursor = r.clampCursor(r.cursor)
	if sel {
		r.selStart = anchor
		r.selEnd = r.cursor
	} else {
		r.selStart = r.cursor
		r.selEnd = r.cursor
	}
	r.scrollToCursor()
	r.RequestRender()
}

func (r *TextareaRenderable) moveRight(sel bool) {
	anchor := r.cursor
	if sel && r.hasSelection() {
		if posEqual(r.cursor, r.selEnd) {
			anchor = r.selStart
		} else {
			anchor = r.selEnd
		}
	}
	cur := r.lines[r.cursor.line]
	if r.cursor.byteCol < len(cur) {
		col := r.cursor.byteCol
		end := col + 1
		for end < len(cur) && cur[end] >= 0x80 && cur[end] < 0xC0 {
			end++
		}
		r.cursor.byteCol = end
	} else if r.cursor.line+1 < len(r.lines) {
		r.cursor.line++
		r.cursor.byteCol = 0
	}
	r.cursor = r.clampCursor(r.cursor)
	if sel {
		r.selStart = anchor
		r.selEnd = r.cursor
	} else {
		r.selStart = r.cursor
		r.selEnd = r.cursor
	}
	r.scrollToCursor()
	r.RequestRender()
}

func (r *TextareaRenderable) moveUp(sel bool) {
	anchor := r.cursor
	if sel && r.hasSelection() {
		if posEqual(r.cursor, r.selEnd) {
			anchor = r.selStart
		} else {
			anchor = r.selEnd
		}
	}
	if r.cursor.line > 0 {
		r.cursor.line--
		if r.cursor.byteCol > len(r.lines[r.cursor.line]) {
			r.cursor.byteCol = len(r.lines[r.cursor.line])
		}
	}
	r.cursor = r.clampCursor(r.cursor)
	if sel {
		r.selStart = anchor
		r.selEnd = r.cursor
	} else {
		r.selStart = r.cursor
		r.selEnd = r.cursor
	}
	r.scrollToCursor()
	r.RequestRender()
}

func (r *TextareaRenderable) moveDown(sel bool) {
	anchor := r.cursor
	if sel && r.hasSelection() {
		if posEqual(r.cursor, r.selEnd) {
			anchor = r.selStart
		} else {
			anchor = r.selEnd
		}
	}
	if r.cursor.line+1 < len(r.lines) {
		r.cursor.line++
		if r.cursor.byteCol > len(r.lines[r.cursor.line]) {
			r.cursor.byteCol = len(r.lines[r.cursor.line])
		}
	}
	r.cursor = r.clampCursor(r.cursor)
	if sel {
		r.selStart = anchor
		r.selEnd = r.cursor
	} else {
		r.selStart = r.cursor
		r.selEnd = r.cursor
	}
	r.scrollToCursor()
	r.RequestRender()
}

func (r *TextareaRenderable) gotoLineHome(sel bool) {
	anchor := r.cursor
	if sel && r.hasSelection() {
		if posEqual(r.cursor, r.selEnd) {
			anchor = r.selStart
		} else {
			anchor = r.selEnd
		}
	}
	r.cursor.byteCol = 0
	r.cursor = r.clampCursor(r.cursor)
	if sel {
		r.selStart = anchor
		r.selEnd = r.cursor
	} else {
		r.selStart = r.cursor
		r.selEnd = r.cursor
	}
	r.scrollToCursor()
	r.RequestRender()
}

func (r *TextareaRenderable) gotoLineEnd(sel bool) {
	anchor := r.cursor
	if sel && r.hasSelection() {
		if posEqual(r.cursor, r.selEnd) {
			anchor = r.selStart
		} else {
			anchor = r.selEnd
		}
	}
	r.cursor.byteCol = len(r.lines[r.cursor.line])
	r.cursor = r.clampCursor(r.cursor)
	if sel {
		r.selStart = anchor
		r.selEnd = r.cursor
	} else {
		r.selStart = r.cursor
		r.selEnd = r.cursor
	}
	r.scrollToCursor()
	r.RequestRender()
}

func (r *TextareaRenderable) gotoVisualLineHome(sel bool) { r.gotoLineHome(sel) }
func (r *TextareaRenderable) gotoVisualLineEnd(sel bool)  { r.gotoLineEnd(sel) }

func (r *TextareaRenderable) gotoBufferHome(sel bool) {
	anchor := r.cursor
	if sel && r.hasSelection() {
		if posEqual(r.cursor, r.selEnd) {
			anchor = r.selStart
		} else {
			anchor = r.selEnd
		}
	}
	r.cursor = cursorPos{0, 0}
	if sel {
		r.selStart = anchor
		r.selEnd = r.cursor
	} else {
		r.selStart = r.cursor
		r.selEnd = r.cursor
	}
	r.scrollToCursor()
	r.RequestRender()
}

func (r *TextareaRenderable) gotoBufferEnd(sel bool) {
	anchor := r.cursor
	if sel && r.hasSelection() {
		if posEqual(r.cursor, r.selEnd) {
			anchor = r.selStart
		} else {
			anchor = r.selEnd
		}
	}
	last := len(r.lines) - 1
	if last < 0 {
		last = 0
	}
	r.cursor = cursorPos{last, len(r.lines[last])}
	if sel {
		r.selStart = anchor
		r.selEnd = r.cursor
	} else {
		r.selStart = r.cursor
		r.selEnd = r.cursor
	}
	r.scrollToCursor()
	r.RequestRender()
}

func (r *TextareaRenderable) deleteToLineEnd() {
	if r.options.ReadOnly {
		return
	}
	if r.hasSelection() {
		r.deleteSelection()
		return
	}
	r.pushHistory()
	line := r.cursor.line
	cur := r.lines[line]
	col := r.cursor.byteCol
	if col > len(cur) {
		col = len(cur)
	}
	r.lines[line] = cur[:col]
	r.selStart = r.cursor
	r.selEnd = r.cursor
	r.syncValue()
	r.RequestRender()
}

func (r *TextareaRenderable) deleteToLineStart() {
	if r.options.ReadOnly {
		return
	}
	if r.hasSelection() {
		r.deleteSelection()
		return
	}
	r.pushHistory()
	line := r.cursor.line
	cur := r.lines[line]
	col := r.cursor.byteCol
	if col > len(cur) {
		col = len(cur)
	}
	r.lines[line] = cur[col:]
	r.cursor.byteCol = 0
	r.selStart = r.cursor
	r.selEnd = r.cursor
	r.syncValue()
	r.RequestRender()
}

func (r *TextareaRenderable) deleteWordBackward() {
	if r.options.ReadOnly {
		return
	}
	if r.hasSelection() {
		r.deleteSelection()
		return
	}
	r.pushHistory()
	line := r.cursor.line
	col := r.cursor.byteCol
	cur := r.lines[line]
	if col == 0 {
		if line > 0 {
			prev := r.lines[line-1]
			r.lines[line-1] = prev + cur
			r.lines = append(r.lines[:line], r.lines[line+1:]...)
			r.cursor.line--
			r.cursor.byteCol = len(prev)
		}
	} else {
		i := col
		for i > 0 && (cur[i-1] == ' ' || cur[i-1] == '\t') {
			i--
		}
		for i > 0 && !(cur[i-1] == ' ' || cur[i-1] == '\t') {
			i--
		}
		r.lines[line] = cur[:i] + cur[col:]
		r.cursor.byteCol = i
	}
	r.cursor = r.clampCursor(r.cursor)
	r.selStart = r.cursor
	r.selEnd = r.cursor
	r.syncValue()
	r.scrollToCursor()
	r.RequestRender()
}

func (r *TextareaRenderable) deleteWordForward() {
	if r.options.ReadOnly {
		return
	}
	if r.hasSelection() {
		r.deleteSelection()
		return
	}
	r.pushHistory()
	line := r.cursor.line
	cur := r.lines[line]
	col := r.cursor.byteCol
	if col >= len(cur) {
		if line+1 < len(r.lines) {
			next := r.lines[line+1]
			r.lines[line] = cur + next
			r.lines = append(r.lines[:line+1], r.lines[line+2:]...)
		}
	} else {
		i := col
		for i < len(cur) && (cur[i] == ' ' || cur[i] == '\t') {
			i++
		}
		for i < len(cur) && !(cur[i] == ' ' || cur[i] == '\t') {
			i++
		}
		r.lines[line] = cur[:col] + cur[i:]
	}
	r.cursor = r.clampCursor(r.cursor)
	r.selStart = r.cursor
	r.selEnd = r.cursor
	r.syncValue()
	r.scrollToCursor()
	r.RequestRender()
}

func (r *TextareaRenderable) moveWordForward(sel bool) {
	anchor := r.cursor
	if sel && r.hasSelection() {
		if posEqual(r.cursor, r.selEnd) {
			anchor = r.selStart
		} else {
			anchor = r.selEnd
		}
	}
	line := r.cursor.line
	col := r.cursor.byteCol
	cur := r.lines[line]
	if col < len(cur) {
		i := col
		for i < len(cur) && (cur[i] == ' ' || cur[i] == '\t') {
			i++
		}
		for i < len(cur) && !(cur[i] == ' ' || cur[i] == '\t') {
			i++
		}
		r.cursor.byteCol = i
	} else if line+1 < len(r.lines) {
		r.cursor.line++
		r.cursor.byteCol = 0
	}
	r.cursor = r.clampCursor(r.cursor)
	if sel {
		r.selStart = anchor
		r.selEnd = r.cursor
	} else {
		r.selStart = r.cursor
		r.selEnd = r.cursor
	}
	r.scrollToCursor()
	r.RequestRender()
}

func (r *TextareaRenderable) moveWordBackward(sel bool) {
	anchor := r.cursor
	if sel && r.hasSelection() {
		if posEqual(r.cursor, r.selEnd) {
			anchor = r.selStart
		} else {
			anchor = r.selEnd
		}
	}
	line := r.cursor.line
	col := r.cursor.byteCol
	cur := r.lines[line]
	if col > 0 {
		i := col
		for i > 0 && (cur[i-1] == ' ' || cur[i-1] == '\t') {
			i--
		}
		for i > 0 && !(cur[i-1] == ' ' || cur[i-1] == '\t') {
			i--
		}
		r.cursor.byteCol = i
	} else if line > 0 {
		r.cursor.line--
		r.cursor.byteCol = len(r.lines[r.cursor.line])
	}
	r.cursor = r.clampCursor(r.cursor)
	if sel {
		r.selStart = anchor
		r.selEnd = r.cursor
	} else {
		r.selStart = r.cursor
		r.selEnd = r.cursor
	}
	r.scrollToCursor()
	r.RequestRender()
}

func (r *TextareaRenderable) selectAll() {
	r.selStart = cursorPos{0, 0}
	last := len(r.lines) - 1
	if last < 0 {
		last = 0
	}
	r.selEnd = cursorPos{last, len(r.lines[last])}
	r.cursor = r.selEnd
	r.RequestRender()
}

func (r *TextareaRenderable) scrollToCursor() {
	vh := r.viewportHeight()
	vw := r.viewportWidth()
	cx := r.byteColToX(r.cursor.line, r.cursor.byteCol)
	cy := r.cursor.line - r.scrollY
	if cy < 0 {
		r.scrollY = r.cursor.line
	}
	if cy >= vh {
		r.scrollY = r.cursor.line - vh + 1
	}
	if cx < r.scrollX {
		r.scrollX = cx
	}
	if cx >= r.scrollX+vw {
		r.scrollX = cx - vw + 1
	}
	if r.scrollX < 0 {
		r.scrollX = 0
	}
	if r.scrollY < 0 {
		r.scrollY = 0
	}
}

func (r *TextareaRenderable) viewportWidth() int {
	w := int(r.GetWidth())
	if r.options.ShowLineNumbers {
		w -= 6
	}
	if w < 1 {
		w = 1
	}
	return w
}

func (r *TextareaRenderable) viewportHeight() int {
	h := int(r.GetHeight())
	if h < 1 {
		h = 1
	}
	return h
}

func (r *TextareaRenderable) ScrollUp(n int) {
	r.scrollY -= n
	if r.scrollY < 0 {
		r.scrollY = 0
	}
	r.RequestRender()
}

func (r *TextareaRenderable) ScrollDown(n int) {
	r.scrollY += n
	max := len(r.lines) - r.viewportHeight()
	if max < 0 {
		max = 0
	}
	if r.scrollY > max {
		r.scrollY = max
	}
	r.RequestRender()
}

// handlePaste processes a paste event.
func (r *TextareaRenderable) handlePaste(event *core.PasteEvent) bool {
	if event == nil {
		return false
	}
	if r.options.ReadOnly {
		return false
	}
	text := string(event.Data)
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	r.Insert(text)
	return true
}

// HandlePaste is the exported entry point for paste handling.
func (r *TextareaRenderable) HandlePaste(event *core.PasteEvent) bool {
	return r.handlePaste(event)
}

func (r *TextareaRenderable) captureSnapshot() textareaSnapshot {
	return textareaSnapshot{
		value:    r.value,
		cursor:   r.cursor,
		selStart: r.selStart,
		selEnd:   r.selEnd,
		scrollX:  r.scrollX,
		scrollY:  r.scrollY,
	}
}

func (r *TextareaRenderable) restoreSnapshot(s textareaSnapshot) {
	r.value = s.value
	r.cursor = s.cursor
	r.selStart = s.selStart
	r.selEnd = s.selEnd
	r.scrollX = s.scrollX
	r.scrollY = s.scrollY
	r.rebuildLines()
	r.RequestRender()
}

func (r *TextareaRenderable) pushHistory() {
	r.history = append(r.history, r.captureSnapshot())
	if len(r.history) > 200 {
		r.history = r.history[1:]
	}
	r.future = nil
}

// moveCursor adjusts the cursor by (dx, dy) runes/lines, clearing selection.
func (r *TextareaRenderable) moveCursor(dx, dy int) {
	r.ClearSelection()
	newLine := r.cursor.line + dy
	if newLine < 0 {
		newLine = 0
	}
	if newLine >= len(r.lines) {
		newLine = len(r.lines) - 1
	}
	newCol := r.cursor.byteCol
	if dx > 0 {
		for i := 0; i < dx; i++ {
			cur := r.lines[newLine]
			if newCol < len(cur) {
				end := newCol + 1
				for end < len(cur) && cur[end] >= 0x80 && cur[end] < 0xC0 {
					end++
				}
				newCol = end
			} else if newLine+1 < len(r.lines) {
				newLine++
				newCol = 0
			}
		}
	} else if dx < 0 {
		for i := 0; i < -dx; i++ {
			if newCol > 0 {
				cur := r.lines[newLine]
				start := newCol - 1
				for start > 0 && cur[start] >= 0x80 && cur[start] < 0xC0 {
					start--
				}
				newCol = start
			} else if newLine > 0 {
				newLine--
				newCol = len(r.lines[newLine])
			}
		}
	}
	if newLine < 0 {
		newLine = 0
	}
	if newLine >= len(r.lines) {
		newLine = len(r.lines) - 1
	}
	if newCol < 0 {
		newCol = 0
	}
	if newCol > len(r.lines[newLine]) {
		newCol = len(r.lines[newLine])
	}
	r.cursor = cursorPos{newLine, newCol}
	r.selStart = r.cursor
	r.selEnd = r.cursor
	r.scrollToCursor()
	r.RequestRender()
}

// MoveCursor is the exported version of moveCursor.
func (r *TextareaRenderable) MoveCursor(dx, dy int) {
	r.moveCursor(dx, dy)
}

func posEqual(a, b cursorPos) bool {
	return a.line == b.line && a.byteCol == b.byteCol
}

// Focus sets focus and switches to focused colors.
func (r *TextareaRenderable) Focus() {
	r.Renderable.Focus()
	r.focused = true
	r.applyColors()
	r.RequestRender()
}

// Blur blurs and switches to unfocused colors.
func (r *TextareaRenderable) Blur() {
	r.Renderable.Blur()
	r.focused = false
	r.applyColors()
	r.RequestRender()
}

func (r *TextareaRenderable) applyColors() {
	fg := r.options.ForegroundColor
	bg := r.options.BackgroundColor
	if r.focused {
		if r.options.FocusedForeground != nil { fg = r.options.FocusedForeground }
		if r.options.FocusedBackground != nil { bg = r.options.FocusedBackground }
	} else {
		if r.options.UnfocusedForeground != nil { fg = r.options.UnfocusedForeground }
		if r.options.UnfocusedBackground != nil { bg = r.options.UnfocusedBackground }
	}
	r.options.ForegroundColor = fg
	r.options.BackgroundColor = bg
}

// Placeholder returns the placeholder.
func (r *TextareaRenderable) Placeholder() string {
	return r.options.Placeholder
}

// SetPlaceholder sets the placeholder text.
func (r *TextareaRenderable) SetPlaceholder(text string) {
	r.options.Placeholder = text
	r.RequestRender()
}

// PlaceholderStyled returns the styled placeholder.
func (r *TextareaRenderable) PlaceholderStyled() *StyledText {
	return r.options.PlaceholderStyled
}

// SetPlaceholderStyled sets the styled placeholder.
func (r *TextareaRenderable) SetPlaceholderStyled(st *StyledText) {
	r.options.PlaceholderStyled = st
	r.RequestRender()
}

// Extmarks returns the extmark manager.
func (r *TextareaRenderable) Extmarks() *core.ExtmarkManager {
	return r.extmarks
}

// EditorView returns the editor view.
func (r *TextareaRenderable) EditorView() *core.EditorView {
	return r.editorView
}

// RenderSelf renders the textarea into the buffer.
func (r *TextareaRenderable) RenderSelf(buffer *core.NativeBuffer, deltaTime float64) {
	if r.options.RenderBefore != nil {
		r.options.RenderBefore(r.Renderable, buffer, deltaTime)
	}

	x := int(r.GetX())
	y := int(r.GetY())
	w := int(r.GetWidth())
	h := int(r.GetHeight())

	if w < 1 || h < 1 {
		if r.options.RenderAfter != nil {
			r.options.RenderAfter(r.Renderable, buffer, deltaTime)
		}
		return
	}
	if x < 0 {
		w += x
		x = 0
	}
	if y < 0 {
		h += y
		y = 0
	}
	if x+w > buffer.Width() {
		w = buffer.Width() - x
	}
	if y+h > buffer.Height() {
		h = buffer.Height() - y
	}
	if w <= 0 || h <= 0 {
		if r.options.RenderAfter != nil {
			r.options.RenderAfter(r.Renderable, buffer, deltaTime)
		}
		return
	}

	fg := r.options.ForegroundColor
	bg := r.options.BackgroundColor
	attrs := r.options.Attributes
	selFG := core.RGBA{R: 255, G: 255, B: 255, A: 255}
	selBG := core.RGBA{R: 0, G: 170, B: 255, A: 120}

	gutterW := 0
	if r.options.ShowLineNumbers {
		gutterW = 6
	}
	contentX := x + gutterW
	contentW := w - gutterW
	if contentW < 1 {
		contentW = 1
	}

	// Clear the viewport.
	fillRect(buffer, x, y, w, h, ' ', fg, bg, attrs)

	// Render gutter line numbers.
	if r.options.ShowLineNumbers {
		for row := 0; row < h; row++ {
			lineIdx := r.scrollY + row
			if lineIdx >= len(r.lines) {
				break
			}
			label := strconv.Itoa(lineIdx + 1)
			gx := x + gutterW - len(label) - 1
			if gx < x {
				gx = x
			}
			writeText(buffer, gx, y+row, label, fg, bg, core.TextAttributeDim)
		}
	}

	r.scrollToCursor()

	startLine := r.scrollY
	endLine := startLine + h
	if endLine > len(r.lines) {
		endLine = len(r.lines)
	}

	selStart, selEnd := r.selectionNormalized()

	row := 0
	for lineIdx := startLine; lineIdx < endLine; lineIdx++ {
		if y+row >= y+h {
			break
		}
		line := r.lines[lineIdx]
		display := line
		if r.scrollX > 0 {
			ridx := 0
			cx := 0
			for i, ch := range line {
				if cx >= r.scrollX {
					ridx = i
					break
				}
				if ch == '\t' {
					cx += 4 - (cx % 4)
				} else {
					cx++
				}
			}
			if cx >= r.scrollX {
				display = line[ridx:]
			} else {
				display = ""
			}
		}

		if r.value == "" && (r.options.Placeholder != "" || r.options.PlaceholderStyled != nil) && lineIdx == 0 && r.scrollX == 0 {
			if r.options.PlaceholderStyled != nil {
				px := contentX
				for _, chunk := range r.options.PlaceholderStyled.Chunks {
					chunkFG := chunk.FG
					if chunkFG == nil { chunkFG = fg }
					chunkBG := chunk.BG
					if chunkBG == nil { chunkBG = bg }
					for _, ch := range chunk.Text {
						if px >= x+w { break }
						writeCell(buffer, px, y+row, uint32(ch), chunkFG, chunkBG, chunk.Attrs)
						px++
					}
				}
			} else {
				ph := r.options.Placeholder
				px := contentX
				phFG := r.options.PlaceholderColor
				if phFG == nil { phFG = fg }
				for _, ch := range ph {
					if px >= x+w { break }
					writeCell(buffer, px, y+row, uint32(ch), phFG, bg, core.TextAttributeDim)
					px++
				}
			}
			row++
			continue
		}

		cx := contentX
		maxX := x + w
		dispCol := 0
		lineInSel := (lineIdx > selStart.line || (lineIdx == selStart.line)) &&
			(lineIdx < selEnd.line || (lineIdx == selEnd.line))
		if !r.hasSelection() {
			lineInSel = false
		}

		for _, ch := range display {
			if cx >= maxX {
				break
			}
			colStart := dispCol
			colEnd := dispCol + 1
			if ch == '\t' {
				colEnd = dispCol + (4 - (dispCol % 4))
				if colEnd <= dispCol {
					colEnd = dispCol + 1
				}
			}
			inSel := false
			if lineInSel {
				sStart := 0
				sEnd := 0
				if lineIdx == selStart.line && lineIdx == selEnd.line {
					sStart = selStart.byteCol
					sEnd = selEnd.byteCol
				} else if lineIdx == selStart.line {
					sStart = selStart.byteCol
					sEnd = len(line)
				} else if lineIdx == selEnd.line {
					sStart = 0
					sEnd = selEnd.byteCol
				} else {
					sStart = 0
					sEnd = len(line)
				}
				if colEnd > sStart && colStart < sEnd {
					inSel = true
				}
			}
			if inSel {
				writeCell(buffer, cx, y+row, uint32(ch), &selFG, &selBG, core.TextAttributeInverse)
			} else {
				writeCell(buffer, cx, y+row, uint32(ch), fg, bg, attrs)
			}
			if ch == '\t' {
				cx += 4 - (cx-contentX)%4
				if cx <= contentX {
					cx = contentX + 1
				}
			} else {
				cx++
			}
			dispCol = colEnd
		}
		row++
	}

	// Render cursor.
	if r.cursor.line >= startLine && r.cursor.line < endLine {
		curY := y + (r.cursor.line - startLine)
		curX := contentX + r.byteColToX(r.cursor.line, r.cursor.byteCol) - r.scrollX
		if curX >= contentX && curX < x+w && curY >= y && curY < y+h {
			line := r.lines[r.cursor.line]
			var ch uint32 = ' '
			if r.cursor.byteCol < len(line) {
				b := line[r.cursor.byteCol]
				if b < 0x80 {
					ch = uint32(b)
				} else {
					for i, ru := range line {
						if i == r.cursor.byteCol {
							ch = uint32(ru)
							break
						}
					}
				}
			}
			writeCell(buffer, curX, curY, ch, &selFG, &selBG, core.TextAttributeInverse)
		}
		if r.Context() != nil && r.Context().SetCursorPosition != nil {
			curX2 := contentX + r.byteColToX(r.cursor.line, r.cursor.byteCol) - r.scrollX
			curY2 := y + (r.cursor.line - startLine)
			r.Context().SetCursorPosition(curX2, curY2, true)
		}
	}

	if r.options.RenderAfter != nil {
		r.options.RenderAfter(r.Renderable, buffer, deltaTime)
	}
}

var _ = utf8.RuneError
