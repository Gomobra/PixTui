package native

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

// ghostty-vt is a port of the ghostty-vt Zig library: a VT terminal emulator
// that parses ANSI escape sequences and maintains a screen buffer.
//
// It provides a Terminal that can be fed raw bytes via a Stream, and can
// produce the rendered plain text of the visible screen.

// VTAttribute bits for character cells.
const (
	VTAttrBold uint32 = 1 << iota
	VTAttrDim
	VTAttrItalic
	VTAttrUnderline
	VTAttrBlink
	VTAttrReverse
	VTAttrInvisible
	VTAttrStrikethrough
)

// VTCell is a single cell in the terminal screen buffer.
type VTCell struct {
	Rune  rune
	Width int // cell width (0 = double-width continuation, 1 = normal, 2 = wide)
	FG    uint32
	BG    uint32
	Attrs uint32
}

// EmptyCell returns a blank cell.
func EmptyCell() VTCell {
	return VTCell{Rune: ' ', Width: 1, FG: 0, BG: 0, Attrs: 0}
}

// VTMode tracks terminal mode flags.
type VTMode uint32

const (
	VTMODE_ALTSCREEN VTMode = 1 << iota
	VTMODE_ORIGIN           // DECOM origin mode
	VTMODE_WRAP             // line wrap enabled
)

// VTTerminal is a VT terminal emulator. It parses an incoming byte stream of
// escape sequences and printable characters, maintaining a grid of cells.
type VTTerminal struct {
	cols   int
	rows   int
	screen []VTCell // main screen buffer
	alt    []VTCell // alternate screen buffer

	cur     *[]VTCell // points to active buffer
	cursorX int
	cursorY int
	savedX  int
	savedY  int

	topMargin    int
	bottomMargin int

	originMode bool
	wrapMode   bool
	altScreen  bool

	// SGR state
	fg        uint32
	bg        uint32
	attrs     uint32
	fgDefault bool
	bgDefault bool

	// Parser state
	inEscape bool
	inCSI    bool
	inOSC    bool
	inDCS    bool
	params   []int
	paramBuf strings.Builder
	oscBuf   strings.Builder

	gen   uint64 // buffer generation, incremented on changes
	dirty bool
}

// NewVTTerminal creates a new VT terminal with the given dimensions.
func NewVTTerminal(cols, rows int) *VTTerminal {
	if cols <= 0 {
		cols = 80
	}
	if rows <= 0 {
		rows = 24
	}
	t := &VTTerminal{
		cols:         cols,
		rows:         rows,
		screen:       make([]VTCell, cols*rows),
		alt:          make([]VTCell, cols*rows),
		cursorX:      0,
		cursorY:      0,
		topMargin:    0,
		bottomMargin: rows - 1,
		originMode:   false,
		wrapMode:     true,
		altScreen:    false,
		fgDefault:    true,
		bgDefault:    true,
	}
	t.cur = &t.screen
	for i := range *t.cur {
		(*t.cur)[i] = EmptyCell()
	}
	for i := range t.alt {
		t.alt[i] = EmptyCell()
	}
	return t
}

// activeBuffer returns a pointer to the active screen buffer slice.
func (t *VTTerminal) activeBuffer() []VTCell {
	if t.altScreen {
		return t.alt
	}
	return t.screen
}

// cellIndex converts x,y to a flat index.
func (t *VTTerminal) cellIndex(x, y int) int {
	return y*t.cols + x
}

// setCell writes a rune at the cursor, advancing the cursor.
func (t *VTTerminal) setCell(r rune) {
	buf := t.activeBuffer()
	w := RuneWidth(r)
	if w == 0 {
		w = 1
	}
	idx := t.cellIndex(t.cursorX, t.cursorY)
	buf[idx] = VTCell{
		Rune:  r,
		Width: w,
		FG:    t.fg,
		BG:    t.bg,
		Attrs: t.attrs,
	}
	// If double-width, fill the next cell as a continuation.
	if w > 1 && t.cursorX+1 < t.cols {
		buf[idx+1] = VTCell{
			Rune:  0,
			Width: 0, // continuation marker
			FG:    t.fg,
			BG:    t.bg,
			Attrs: t.attrs,
		}
	}
	t.cursorX += w
	if t.cursorX >= t.cols {
		if t.wrapMode {
			t.cursorX = 0
			t.cursorY++
			if t.cursorY > t.bottomMargin {
				t.cursorY = t.bottomMargin
			}
			if t.cursorY >= t.rows {
				t.scroll(1)
				t.cursorY = t.rows - 1
			}
		} else {
			t.cursorX = t.cols - 1
		}
	}
	t.gen++
	t.dirty = true
}

// scroll shifts the screen up by n lines.
func (t *VTTerminal) scroll(n int) {
	if n <= 0 || n >= t.rows {
		return
	}
	buf := t.activeBuffer()
	top := t.topMargin
	bottom := t.bottomMargin
	if bottom >= t.rows {
		bottom = t.rows - 1
	}
	lineSize := t.cols
	moveLines := (bottom - top + 1) - n
	if moveLines > 0 {
		moveStart := (top + n) * lineSize
		copy(buf[top*lineSize:], buf[moveStart:moveStart+moveLines*lineSize])
	}
	clearStart := (top + moveLines) * lineSize
	clearEnd := (bottom + 1) * lineSize
	for i := clearStart; i < clearEnd; i++ {
		buf[i] = EmptyCell()
	}
	t.gen++
	t.dirty = true
}

// insertLine inserts a blank line at the cursor row, shifting down.
func (t *VTTerminal) insertLine(n int) {
	buf := t.activeBuffer()
	lineSize := t.cols
	top := t.topMargin
	bottom := t.bottomMargin
	if bottom >= t.rows {
		bottom = t.rows - 1
	}
	if t.cursorY < top || t.cursorY > bottom {
		return
	}
	// Shift lines [cursorY, bottom-n] down by n
	srcStart := t.cursorY * lineSize
	srcEnd := (bottom - n + 1) * lineSize
	dstStart := (t.cursorY + n) * lineSize
	if srcEnd > srcStart {
		copy(buf[dstStart:dstStart+(srcEnd-srcStart)], buf[srcStart:srcEnd])
	}
	clearStart := t.cursorY * lineSize
	clearEnd := (t.cursorY + n) * lineSize
	if clearEnd > bottom*lineSize+lineSize {
		clearEnd = bottom*lineSize + lineSize
	}
	for i := clearStart; i < clearEnd; i++ {
		buf[i] = EmptyCell()
	}
	t.gen++
	t.dirty = true
}

// deleteLine removes n lines at the cursor row.
func (t *VTTerminal) deleteLine(n int) {
	buf := t.activeBuffer()
	lineSize := t.cols
	top := t.topMargin
	bottom := t.bottomMargin
	if bottom >= t.rows {
		bottom = t.rows - 1
	}
	if t.cursorY < top || t.cursorY > bottom {
		return
	}
	srcStart := (t.cursorY + n) * lineSize
	srcEnd := (bottom + 1) * lineSize
	dstStart := t.cursorY * lineSize
	if srcEnd > srcStart {
		copy(buf[dstStart:dstStart+(srcEnd-srcStart)], buf[srcStart:srcEnd])
	}
	clearStart := (bottom - n + 1) * lineSize
	clearEnd := (bottom + 1) * lineSize
	for i := clearStart; i < clearEnd; i++ {
		buf[i] = EmptyCell()
	}
	t.gen++
	t.dirty = true
}

// eraseScreen erases part or all of the screen.
func (t *VTTerminal) eraseScreen(mode int) {
	buf := t.activeBuffer()
	lineSize := t.cols
	_ = lineSize
	switch mode {
	case 0: // from cursor to end
		start := t.cellIndex(t.cursorX, t.cursorY)
		for i := start; i < len(buf); i++ {
			buf[i] = EmptyCell()
		}
	case 1: // from start to cursor
		end := t.cellIndex(t.cursorX, t.cursorY) + 1
		for i := 0; i < end; i++ {
			buf[i] = EmptyCell()
		}
	case 2: // entire screen
		for i := range buf {
			buf[i] = EmptyCell()
		}
	case 3: // saved lines (not implemented)
	}
	t.gen++
	t.dirty = true
}

// eraseLine erases part or all of the current line.
func (t *VTTerminal) eraseLine(mode int) {
	buf := t.activeBuffer()
	lineSize := t.cols
	rowStart := t.cursorY * lineSize
	switch mode {
	case 0: // from cursor to end of line
		start := rowStart + t.cursorX
		for i := start; i < rowStart+lineSize; i++ {
			buf[i] = EmptyCell()
		}
	case 1: // from start of line to cursor
		end := rowStart + t.cursorX + 1
		for i := rowStart; i < end; i++ {
			buf[i] = EmptyCell()
		}
	case 2: // entire line
		for i := rowStart; i < rowStart+lineSize; i++ {
			buf[i] = EmptyCell()
		}
	}
	t.gen++
	t.dirty = true
}

// insertChars inserts n blank cells at the cursor.
func (t *VTTerminal) insertChars(n int) {
	buf := t.activeBuffer()
	lineSize := t.cols
	rowStart := t.cursorY * lineSize
	// Shift cells [cursorX, cols-n] right by n
	srcStart := rowStart + t.cursorX
	srcEnd := rowStart + t.cols - n
	dstStart := rowStart + t.cursorX + n
	if srcEnd > srcStart {
		copy(buf[dstStart:dstStart+(srcEnd-srcStart)], buf[srcStart:srcEnd])
	}
	for i := srcStart; i < dstStart; i++ {
		buf[i] = EmptyCell()
	}
	t.gen++
	t.dirty = true
}

// deleteChars removes n cells at the cursor.
func (t *VTTerminal) deleteChars(n int) {
	buf := t.activeBuffer()
	lineSize := t.cols
	rowStart := t.cursorY * lineSize
	srcStart := rowStart + t.cursorX + n
	srcEnd := rowStart + t.cols
	dstStart := rowStart + t.cursorX
	if srcEnd > srcStart {
		copy(buf[dstStart:dstStart+(srcEnd-srcStart)], buf[srcStart:srcEnd])
	}
	for i := srcEnd - n; i < srcEnd; i++ {
		buf[i] = EmptyCell()
	}
	t.gen++
	t.dirty = true
}

// Write feeds bytes into the terminal emulator, parsing escape sequences.
func (t *VTTerminal) Write(data []byte) (int, error) {
	n := 0
	for n < len(data) {
		err := t.feedByte(data[n])
		if err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

// feedByte processes a single byte through the parser state machine.
func (t *VTTerminal) feedByte(b byte) error {
	if t.inOSC {
		if b == 0x07 || b == 0x1b { // BEL or ESC terminates OSC
			t.handleOSC(t.oscBuf.String())
			t.oscBuf.Reset()
			t.inOSC = false
			if b == 0x1b {
				t.inEscape = true
			}
		} else {
			t.oscBuf.WriteByte(b)
		}
		return nil
	}
	if t.inDCS {
		if b == 0x1b || b == 0x5c { // ESC or ST terminates DCS
			t.inDCS = false
			if b == 0x1b {
				t.inEscape = true
			}
		}
		return nil
	}
	if t.inEscape {
		t.inEscape = false
		switch b {
		case '[':
			t.inCSI = true
			t.params = nil
			t.paramBuf.Reset()
		case ']':
			t.inOSC = true
			t.oscBuf.Reset()
		case 'P':
			t.inDCS = true
		case '7': // DECSC save cursor
			t.savedX = t.cursorX
			t.savedY = t.cursorY
		case '8': // DECRC restore cursor
			t.cursorX = t.savedX
			t.cursorY = t.savedY
		case 'c': // RIS reset
			t.reset()
		case 'M': // RI reverse index
			if t.cursorY > t.topMargin {
				t.cursorY--
			} else {
				t.scroll(1)
			}
		case '=': // DECKPAM keypad mode (ignore)
		case '>': // DECKPNM (ignore)
		default:
			// ignore unknown escape
		}
		return nil
	}
	if t.inCSI {
		switch {
		case b >= '0' && b <= '9':
			t.paramBuf.WriteByte(b)
		case b == ';':
			if p, err := strconv.Atoi(t.paramBuf.String()); err == nil {
				t.params = append(t.params, p)
			}
			t.paramBuf.Reset()
		case b == ':':
			// subparameter (ignore)
		default:
			// final character
			if p, err := strconv.Atoi(t.paramBuf.String()); err == nil {
				t.params = append(t.params, p)
			}
			t.paramBuf.Reset()
			t.inCSI = false
			t.handleCSI(b, t.params)
		}
		return nil
	}

	// Normal character processing
	if b < 0x20 {
		switch b {
		case 0x07: // BEL
			// bell (ignore)
		case 0x08: // BS backspace
			if t.cursorX > 0 {
				t.cursorX--
			}
		case 0x09: // TAB
			t.cursorX = ((t.cursorX / 8) + 1) * 8
			if t.cursorX >= t.cols {
				t.cursorX = t.cols - 1
			}
		case 0x0a: // LF newline
			t.cursorX = 0
			t.cursorY++
			if t.cursorY > t.bottomMargin {
				t.cursorY = t.bottomMargin
			}
			if t.cursorY >= t.rows {
				t.scroll(1)
				t.cursorY = t.rows - 1
			}
		case 0x0d: // CR carriage return
			t.cursorX = 0
		case 0x1b: // ESC
			t.inEscape = true
		}
		return nil
	}

	// Decode a rune
	r, size := utf8.DecodeRune([]byte{b})
	if size == 1 && r == utf8.RuneError {
		// Try reading more bytes would need lookahead; treat as replacement
		r = utf8.RuneError
		size = 1
	}
	_ = size
	t.setCell(r)
	return nil
}

// handleCSI processes a CSI sequence with the given final byte and parameters.
func (t *VTTerminal) handleCSI(final byte, params []int) {
	if len(params) == 0 {
		params = append(params, 0)
	}
	p1 := params[0]
	if p1 < 0 {
		p1 = 0
	}
	switch final {
	case 'A': // CUU cursor up
		t.cursorY -= p1
		if t.cursorY < t.topMargin {
			t.cursorY = t.topMargin
		}
	case 'B': // CUD cursor down
		t.cursorY += p1
		if t.cursorY > t.bottomMargin {
			t.cursorY = t.bottomMargin
		}
		if t.cursorY >= t.rows {
			t.cursorY = t.rows - 1
		}
	case 'C': // CUF cursor forward
		t.cursorX += p1
		if t.cursorX >= t.cols {
			t.cursorX = t.cols - 1
		}
	case 'D': // CUB cursor backward
		t.cursorX -= p1
		if t.cursorX < 0 {
			t.cursorX = 0
		}
	case 'E': // CNL cursor next line
		t.cursorX = 0
		t.cursorY += p1
		if t.cursorY > t.bottomMargin {
			t.cursorY = t.bottomMargin
		}
		if t.cursorY >= t.rows {
			t.cursorY = t.rows - 1
		}
	case 'F': // CPL cursor previous line
		t.cursorX = 0
		t.cursorY -= p1
		if t.cursorY < t.topMargin {
			t.cursorY = t.topMargin
		}
	case 'G': // CHA cursor horizontal absolute
		t.cursorX = p1 - 1
		if t.cursorX < 0 {
			t.cursorX = 0
		}
		if t.cursorX >= t.cols {
			t.cursorX = t.cols - 1
		}
	case 'H': // CUP cursor position
		if len(params) >= 2 {
			t.cursorY = params[1] - 1
			t.cursorX = params[0] - 1
		} else {
			t.cursorY = p1 - 1
			t.cursorX = 0
		}
		if t.cursorY < 0 {
			t.cursorY = 0
		}
		if t.cursorY >= t.rows {
			t.cursorY = t.rows - 1
		}
		if t.cursorX < 0 {
			t.cursorX = 0
		}
		if t.cursorX >= t.cols {
			t.cursorX = t.cols - 1
		}
	case 'f': // HVP horizontal and vertical position
		t.handleCSI('H', params)
	case 'J': // ED erase screen
		t.eraseScreen(p1)
	case 'K': // EL erase line
		t.eraseLine(p1)
	case 'L': // IL insert lines
		t.insertLine(p1)
	case 'M': // DL delete lines
		t.deleteLine(p1)
	case 'P': // ICH insert characters
		t.insertChars(p1)
	case 'X': // DCH delete characters
		t.deleteChars(p1)
	case 'm': // SGR select graphic rendition
		t.handleSGR(params)
	case 'h': // SM set mode
		t.handleSetMode(params, true)
	case 'l': // RM reset mode
		t.handleSetMode(params, false)
	case 'r': // DECSTBM set top/bottom margins
		if len(params) >= 2 {
			t.topMargin = params[0] - 1
			t.bottomMargin = params[1] - 1
			if t.topMargin < 0 {
				t.topMargin = 0
			}
			if t.bottomMargin >= t.rows {
				t.bottomMargin = t.rows - 1
			}
			if t.topMargin > t.bottomMargin {
				t.bottomMargin = t.topMargin
			}
			t.cursorX = 0
			t.cursorY = 0
		}
	case 's': // save cursor
		t.savedX = t.cursorX
		t.savedY = t.cursorY
	case 'u': // restore cursor
		t.cursorX = t.savedX
		t.cursorY = t.savedY
	}
}

// handleSGR processes SGR parameters.
func (t *VTTerminal) handleSGR(params []int) {
	for _, p := range params {
		switch p {
		case 0: // reset
			t.fg = 0
			t.bg = 0
			t.attrs = 0
			t.fgDefault = true
			t.bgDefault = true
		case 1: // bold
			t.attrs |= VTAttrBold
		case 2: // dim
			t.attrs |= VTAttrDim
		case 3: // italic
			t.attrs |= VTAttrItalic
		case 4: // underline
			t.attrs |= VTAttrUnderline
		case 5: // blink
			t.attrs |= VTAttrBlink
		case 7: // reverse
			t.attrs |= VTAttrReverse
		case 8: // invisible
			t.attrs |= VTAttrInvisible
		case 9: // strikethrough
			t.attrs |= VTAttrStrikethrough
		case 21: // bold off
			t.attrs &^= VTAttrBold
		case 22: // normal intensity
			t.attrs &^= (VTAttrBold | VTAttrDim)
		case 23: // italic off
			t.attrs &^= VTAttrItalic
		case 24: // underline off
			t.attrs &^= VTAttrUnderline
		case 25: // blink off
			t.attrs &^= VTAttrBlink
		case 27: // reverse off
			t.attrs &^= VTAttrReverse
		case 28: // invisible off
			t.attrs &^= VTAttrInvisible
		case 29: // strikethrough off
			t.attrs &^= VTAttrStrikethrough
		case 30, 31, 32, 33, 34, 35, 36, 37: // foreground color
			t.fg = uint32(p - 30)
			t.fgDefault = false
		case 38: // extended foreground
			if len(params) > 1 {
				t.handleExtendedColor(true, params[1:])
			}
		case 39: // default foreground
			t.fg = 0
			t.fgDefault = true
		case 40, 41, 42, 43, 44, 45, 46, 47: // background color
			t.bg = uint32(p - 40)
			t.bgDefault = false
		case 48: // extended background
			if len(params) > 1 {
				t.handleExtendedColor(false, params[1:])
			}
		case 49: // default background
			t.bg = 0
			t.bgDefault = true
		case 90, 91, 92, 93, 94, 95, 96, 97: // bright foreground
			t.fg = uint32(p - 90 + 8)
			t.fgDefault = false
		case 100, 101, 102, 103, 104, 105, 106, 107: // bright background
			t.bg = uint32(p - 100 + 8)
			t.bgDefault = false
		}
	}
}

// handleExtendedColor handles 38/48 extended color sequences.
func (t *VTTerminal) handleExtendedColor(fg bool, rest []int) {
	if len(rest) == 0 {
		return
	}
	if rest[0] == 5 && len(rest) >= 2 { // 256-color
		color := uint32(rest[1])
		if fg {
			t.fg = color
			t.fgDefault = false
		} else {
			t.bg = color
			t.bgDefault = false
		}
	} else if rest[0] == 2 && len(rest) >= 4 { // truecolor RGB
		color := uint32(rest[1]<<16 | rest[2]<<8 | rest[3])
		if fg {
			t.fg = color | 0x1000000 // mark as truecolor
			t.fgDefault = false
		} else {
			t.bg = color | 0x1000000
			t.bgDefault = false
		}
	}
}

// handleSetMode processes SM/RM mode sequences.
func (t *VTTerminal) handleSetMode(params []int, set bool) {
	for _, p := range params {
		switch p {
		case 4: // IRM insert mode (ignore)
		case 20: // LNM automatic newline (ignore)
		case 25: // DECTCEM cursor show/hide
			// cursor visibility (ignore)
		case 1049: // alt screen
			if set {
				t.altScreen = true
				t.cur = &t.alt
			} else {
				t.altScreen = false
				t.cur = &t.screen
			}
		case 1000, 1002, 1003, 1006: // mouse modes (ignore)
		case 1048: // save/restore cursor (ignore)
		case 2004: // bracketed paste (ignore)
		}
	}
}

// handleOSC processes an OSC sequence string.
func (t *VTTerminal) handleOSC(data string) {
	// OSC sequences are ignored for now (title, colors, etc.)
	_ = data
}

// reset performs a terminal reset (RIS).
func (t *VTTerminal) reset() {
	buf := t.activeBuffer()
	for i := range buf {
		buf[i] = EmptyCell()
	}
	t.cursorX = 0
	t.cursorY = 0
	t.topMargin = 0
	t.bottomMargin = t.rows - 1
	t.originMode = false
	t.wrapMode = true
	t.fg = 0
	t.bg = 0
	t.attrs = 0
	t.fgDefault = true
	t.bgDefault = true
	t.gen++
	t.dirty = true
}

// Cols returns the terminal width.
func (t *VTTerminal) Cols() int {
	return t.cols
}

// Rows returns the terminal height.
func (t *VTTerminal) Rows() int {
	return t.rows
}

// Cursor returns the cursor position.
func (t *VTTerminal) Cursor() (int, int) {
	return t.cursorX, t.cursorY
}

// Generation returns the buffer generation counter.
func (t *VTTerminal) Generation() uint64 {
	return t.gen
}

// IsDirty returns true if the buffer has changed since the last MarkClean.
func (t *VTTerminal) IsDirty() bool {
	return t.dirty
}

// MarkClean marks the buffer as clean.
func (t *VTTerminal) MarkClean() {
	t.dirty = false
}

// Resize changes the terminal dimensions.
func (t *VTTerminal) Resize(cols, rows int) error {
	if cols <= 0 || rows <= 0 {
		return errors.New("invalid dimensions")
	}
	if cols == t.cols && rows == t.rows {
		return nil
	}
	// Build new buffers preserving as much content as possible
	newScreen := make([]VTCell, cols*rows)
	for i := range newScreen {
		newScreen[i] = EmptyCell()
	}
	minCols := t.cols
	if cols < minCols {
		minCols = cols
	}
	minRows := t.rows
	if rows < minRows {
		minRows = rows
	}
	for y := 0; y < minRows; y++ {
		for x := 0; x < minCols; x++ {
			newScreen[y*cols+x] = t.screen[y*t.cols+x]
		}
	}
	newAlt := make([]VTCell, cols*rows)
	for i := range newAlt {
		newAlt[i] = EmptyCell()
	}
	for y := 0; y < minRows; y++ {
		for x := 0; x < minCols; x++ {
			newAlt[y*cols+x] = t.alt[y*t.cols+x]
		}
	}
	t.cols = cols
	t.rows = rows
	t.screen = newScreen
	t.alt = newAlt
	if t.cursorX >= cols {
		t.cursorX = cols - 1
	}
	if t.cursorY >= rows {
		t.cursorY = rows - 1
	}
	t.bottomMargin = rows - 1
	t.gen++
	t.dirty = true
	return nil
}

// Clear clears the screen and resets the cursor.
func (t *VTTerminal) Clear() {
	buf := t.activeBuffer()
	for i := range buf {
		buf[i] = EmptyCell()
	}
	t.cursorX = 0
	t.cursorY = 0
	t.gen++
	t.dirty = true
}

// Cell returns the cell at (x, y), or an empty cell if out of bounds.
func (t *VTTerminal) Cell(x, y int) VTCell {
	buf := t.activeBuffer()
	if x < 0 || x >= t.cols || y < 0 || y >= t.rows {
		return EmptyCell()
	}
	return buf[y*t.cols+x]
}

// PlainString returns the visible text of the screen, with newlines between rows.
// Trailing spaces on each line are trimmed. Double-width continuation cells are skipped.
func (t *VTTerminal) PlainString() string {
	buf := t.activeBuffer()
	var sb strings.Builder
	lineSize := t.cols
	for y := 0; y < t.rows; y++ {
		rowStart := y * lineSize
		rowEnd := rowStart + lineSize
		// Find last non-empty cell in the row
		last := rowStart
		for i := rowStart; i < rowEnd; i++ {
			c := buf[i]
			if c.Rune != ' ' || c.Width != 1 || c.Attrs != 0 {
				last = i + 1
			}
		}
		lineCells := buf[rowStart:last]
		for i := 0; i < len(lineCells); i++ {
			c := lineCells[i]
			if c.Width == 0 {
				continue // continuation cell
			}
			sb.WriteRune(c.Rune)
		}
		if y < t.rows-1 {
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}

// String returns a debug representation of the terminal screen.
func (t *VTTerminal) String() string {
	return t.PlainString()
}

// VTStream feeds data into a terminal. This mirrors the ghostty-vt Stream API.
type VTStream struct {
	terminal *VTTerminal
	buf      bytes.Buffer
}

// NewVTStream creates a new stream writing to the given terminal.
func NewVTStream(term *VTTerminal) *VTStream {
	return &VTStream{terminal: term}
}

// NextSlice feeds a slice of bytes into the terminal.
func (s *VTStream) NextSlice(data []byte) (int, error) {
	return s.terminal.Write(data)
}

// NextString feeds a string into the terminal.
func (s *VTStream) NextString(str string) (int, error) {
	return s.terminal.Write([]byte(str))
}

// Flush flushes any buffered data. In this implementation data is written
// directly, so this is a no-op.
func (s *VTStream) Flush() error {
	return nil
}

// deinit releases resources (no-op in Go due to GC).
func (s *VTStream) deinit() {
	s.buf.Reset()
}

// PlainString returns the plain text of the terminal screen.
func (t *VTTerminal) PlainStringAlloc() (string, error) {
	return t.PlainString(), nil
}

// Ensure fmt.Stringer interface.
var _ fmt.Stringer = (*VTTerminal)(nil)
