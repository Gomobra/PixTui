package native

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"sync"
	"unicode/utf8"
)

// EmbeddedTerminal is an in-process terminal emulator that drives a session's
// output buffer directly, without requiring an external TTY. It mirrors the
// embedded-terminal compositor in packages/native/src/embedded-terminal.
type EmbeddedTerminal struct {
	mu         sync.Mutex
	term       *Terminal      // terminal I/O (raw mode, dimensions, input/output)
	process    *os.Process    // attached child process (if any)
	stdinPipe  io.WriteCloser // write end of the child's stdin pipe
	stdoutPipe io.ReadCloser  // read end of the child's stdout pipe
	width      int
	height     int
	buffer     *OptimizedBuffer
	cursorX    int
	cursorY    int
	altScreen  bool
	dirty      bool
	closed     bool
	wg         sync.WaitGroup // tracks the stdout reader goroutine
}

// NewEmbeddedTerminal creates a new embedded terminal with the given dimensions.
func NewEmbeddedTerminal(width, height int) *EmbeddedTerminal {
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}
	return &EmbeddedTerminal{
		term:    NewTerminal(),
		width:   width,
		height:  height,
		buffer:  NewOptimizedBuffer(width, height),
		cursorX: 0,
		cursorY: 0,
	}
}

// Process returns the attached child process, if any.
func (t *EmbeddedTerminal) Process() *os.Process {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.process
}

// Width returns the terminal width in cells.
func (t *EmbeddedTerminal) Width() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.width
}

// Height returns the terminal height in cells.
func (t *EmbeddedTerminal) Height() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.height
}

// Buffer returns the underlying output buffer.
func (t *EmbeddedTerminal) Buffer() *OptimizedBuffer {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.buffer
}

// Resize changes the terminal dimensions and resets the buffer.
func (t *EmbeddedTerminal) Resize(width, height int) error {
	if width <= 0 || height <= 0 {
		return errors.New("invalid dimensions")
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.width = width
	t.height = height
	t.buffer.Resize(width, height)
	t.cursorX = 0
	t.cursorY = 0
	t.dirty = true
	return nil
}

// Write feeds text into the terminal emulator, processing line wrapping and
// newlines.
func (t *EmbeddedTerminal) Write(data []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return 0, errors.New("terminal closed")
	}
	n := 0
	for n < len(data) {
		r, size := utf8.DecodeRune(data[n:])
		if size == 0 {
			size = 1
			r = utf8.RuneError
		}
		n += size
		if r == '\n' {
			t.cursorY++
			t.cursorX = 0
			if t.cursorY >= t.height {
				t.scroll(1)
				t.cursorY = t.height - 1
			}
			continue
		}
		if r == '\r' {
			t.cursorX = 0
			continue
		}
		w := RuneWidth(r)
		if w == 0 {
			continue
		}
		if t.cursorX+w > t.width {
			t.cursorY++
			t.cursorX = 0
			if t.cursorY >= t.height {
				t.scroll(1)
				t.cursorY = t.height - 1
			}
		}
		if t.cursorY < t.height && t.cursorX < t.width {
			t.buffer.SetCell(t.cursorX, t.cursorY, uint32(r), 0xFFFF, 0, 0)
		}
		t.cursorX += w
		t.dirty = true
	}
	return n, nil
}

// scroll shifts the buffer up by lines rows.
func (t *EmbeddedTerminal) scroll(lines int) {
	if lines <= 0 || lines >= t.height {
		return
	}
	size := t.width * t.height
	moveCells := (t.height - lines) * t.width
	copy(t.buffer.char, t.buffer.char[lines*t.width:])
	copy(t.buffer.fg, t.buffer.fg[lines*t.width:])
	copy(t.buffer.bg, t.buffer.bg[lines*t.width:])
	copy(t.buffer.attrs, t.buffer.attrs[lines*t.width:])
	for i := moveCells; i < size; i++ {
		t.buffer.char[i] = ' '
		t.buffer.fg[i] = 0
		t.buffer.bg[i] = 0
		t.buffer.attrs[i] = 0
	}
	t.buffer.gen++
	t.dirty = true
}

// Cursor returns the cursor position.
func (t *EmbeddedTerminal) Cursor() (int, int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.cursorX, t.cursorY
}

// SetCursor sets the cursor position, clamped to the buffer.
func (t *EmbeddedTerminal) SetCursor(x, y int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if x < 0 {
		x = 0
	}
	if x > t.width {
		x = t.width
	}
	if y < 0 {
		y = 0
	}
	if y >= t.height {
		y = t.height - 1
	}
	t.cursorX = x
	t.cursorY = y
}

// Clear clears the entire buffer and resets the cursor.
func (t *EmbeddedTerminal) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.buffer.Clear()
	t.cursorX = 0
	t.cursorY = 0
	t.dirty = true
}

// HideCursor hides the cursor (no-op for embedded terminal).
func (t *EmbeddedTerminal) HideCursor() error { return nil }

// ShowCursor shows the cursor (no-op for embedded terminal).
func (t *EmbeddedTerminal) ShowCursor() error { return nil }

// EnableAlternateScreen switches to the alternate screen.
func (t *EmbeddedTerminal) EnableAlternateScreen() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.altScreen = true
	return nil
}

// DisableAlternateScreen switches back to the main screen.
func (t *EmbeddedTerminal) DisableAlternateScreen() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.altScreen = false
	return nil
}

// StartProcess starts a child process with piped stdin and stdout. The child's
// stdout is fed into the terminal emulator's buffer via Write, and WriteStdin
// can be used to send data to the child's stdin. Only one process may be
// attached at a time.
func (t *EmbeddedTerminal) StartProcess(cmd string, args ...string) error {
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return errors.New("terminal closed")
	}
	if t.process != nil {
		t.mu.Unlock()
		return errors.New("process already started")
	}
	t.mu.Unlock()

	c := exec.Command(cmd, args...)

	stdinPipe, err := c.StdinPipe()
	if err != nil {
		return err
	}
	stdoutPipe, err := c.StdoutPipe()
	if err != nil {
		stdinPipe.Close()
		return err
	}

	if err := c.Start(); err != nil {
		stdinPipe.Close()
		stdoutPipe.Close()
		return err
	}

	t.mu.Lock()
	t.process = c.Process
	t.stdinPipe = stdinPipe
	t.stdoutPipe = stdoutPipe
	t.mu.Unlock()

	t.wg.Add(1)
	go t.readStdoutLoop()

	return nil
}

// readStdoutLoop reads from the child's stdout and feeds it into the terminal
// emulator's Write method so the child's output appears in the buffer.
func (t *EmbeddedTerminal) readStdoutLoop() {
	defer t.wg.Done()
	buf := make([]byte, 4096)
	for {
		n, err := t.stdoutPipe.Read(buf)
		if n > 0 {
			t.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}
}

// WriteStdin sends data to the child process's stdin. Returns an error if no
// process is attached or the terminal is closed.
func (t *EmbeddedTerminal) WriteStdin(data []byte) (int, error) {
	t.mu.Lock()
	if t.closed || t.stdinPipe == nil {
		t.mu.Unlock()
		return 0, errors.New("no process attached")
	}
	pipe := t.stdinPipe
	t.mu.Unlock()

	return pipe.Write(data)
}

// Read reads data from the terminal (delegates to the underlying terminal input).
func (t *EmbeddedTerminal) Read(data []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return 0, errors.New("terminal closed")
	}
	return t.term.ReadInput(data)
}

// Close releases the embedded terminal resources, stopping any attached child
// process and waiting for the stdout reader goroutine to exit.
func (t *EmbeddedTerminal) Close() error {
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return nil
	}
	t.closed = true

	if t.stdinPipe != nil {
		t.stdinPipe.Close()
	}
	if t.process != nil {
		_ = t.process.Kill()
	}
	t.mu.Unlock()

	t.wg.Wait()

	t.mu.Lock()
	if t.stdoutPipe != nil {
		t.stdoutPipe.Close()
	}
	t.mu.Unlock()

	return nil
}

// IsDirty returns true if the buffer has changed since the last check.
func (t *EmbeddedTerminal) IsDirty() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.dirty
}

// MarkClean marks the buffer as clean.
func (t *EmbeddedTerminal) MarkClean() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.dirty = false
}
