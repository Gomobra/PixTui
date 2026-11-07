package native

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"

	"golang.org/x/term"
)

// Terminal handles terminal I/O: raw mode, dimensions, and input/output.
// Mirrors packages/native/src/terminal.zig with real terminal support.
type Terminal struct {
	mu           sync.Mutex
	rawMode      bool
	oldState     *term.State
	width        int
	height       int
	mouseEnabled bool
	stdinFd      uintptr
	stdoutFd     uintptr
	isTTY        bool

	// resizeSignal chan receives SIGWINCH signals for terminal resize events.
	resizeSignal chan os.Signal
	// stopResize is closed to request the WatchResize goroutine to stop.
	stopResize chan struct{}
	// resizeDone is closed when the WatchResize goroutine has exited.
	resizeDone chan struct{}
}

// NewTerminal creates a new Terminal instance with real terminal support.
func NewTerminal() *Terminal {
	t := &Terminal{
		width:      80,
		height:     24,
		stdinFd:    os.Stdin.Fd(),
		stdoutFd:   os.Stdout.Fd(),
		stopResize: make(chan struct{}),
		resizeDone: make(chan struct{}),
	}

	// Check if we're running on a terminal
	fi, err := os.Stdin.Stat()
	if err == nil {
		t.isTTY = (fi.Mode() & os.ModeCharDevice) != 0
	}

	// Get terminal dimensions
	if t.isTTY {
		w, h := t.getTermSize()
		if w > 0 && h > 0 {
			t.width = w
			t.height = h
		}
	}

	return t
}

// IsSupported returns true if the terminal is supported.
func (t *Terminal) IsSupported() bool {
	return t.isTTY
}

// EnableRawMode puts the terminal into raw mode.
func (t *Terminal) EnableRawMode() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.rawMode {
		return nil
	}
	if !t.isTTY {
		return fmt.Errorf("terminal not supported: not a TTY")
	}
	oldState, err := term.MakeRaw(int(t.stdinFd))
	if err != nil {
		return fmt.Errorf("failed to enable raw mode: %w", err)
	}
	t.oldState = oldState
	t.rawMode = true
	return nil
}

// DisableRawMode restores the terminal to its previous state.
func (t *Terminal) DisableRawMode() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.rawMode || t.oldState == nil {
		return nil
	}
	err := term.Restore(int(t.stdinFd), t.oldState)
	t.rawMode = false
	t.oldState = nil
	return err
}

// GetDimensions returns the terminal dimensions.
func (t *Terminal) GetDimensions() (int, int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.width, t.height
}

// SetDimensions sets the terminal dimensions.
func (t *Terminal) SetDimensions(width, height int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.width = width
	t.height = height
}

// ReadInput reads input from stdin.
func (t *Terminal) ReadInput(buf []byte) (int, error) {
	return os.Stdin.Read(buf)
}

// WriteOutput writes output to stdout.
func (t *Terminal) WriteOutput(data []byte) (int, error) {
	return os.Stdout.Write(data)
}

// ClearScreen clears the terminal screen.
func (t *Terminal) ClearScreen() error {
	_, err := t.WriteOutput([]byte("\x1b[2J\x1b[H"))
	return err
}

// HideCursor hides the terminal cursor.
func (t *Terminal) HideCursor() error {
	_, err := t.WriteOutput([]byte("\x1b[?25l"))
	return err
}

// ShowCursor shows the terminal cursor.
func (t *Terminal) ShowCursor() error {
	_, err := t.WriteOutput([]byte("\x1b[?25h"))
	return err
}

// EnableAlternateScreen switches to the alternate screen buffer.
func (t *Terminal) EnableAlternateScreen() error {
	_, err := t.WriteOutput([]byte("\x1b[?1049h"))
	return err
}

// DisableAlternateScreen switches back to the main screen buffer.
func (t *Terminal) DisableAlternateScreen() error {
	_, err := t.WriteOutput([]byte("\x1b[?1049l"))
	return err
}

// EnableMouse enables mouse tracking.
func (t *Terminal) EnableMouse() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.mouseEnabled = true
	_, err := t.WriteOutput([]byte("\x1b[?1000h\x1b[?1002h\x1b[?1003h\x1b[?1006h"))
	return err
}

// DisableMouse disables mouse tracking.
func (t *Terminal) DisableMouse() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.mouseEnabled = false
	_, err := t.WriteOutput([]byte("\x1b[?1000l\x1b[?1002l\x1b[?1003l\x1b[?1006l"))
	return err
}

// SetWindowTitle sets the terminal window title.
func (t *Terminal) SetWindowTitle(title string) error {
	_, err := t.WriteOutput([]byte("\x1b]0;" + title + "\x07"))
	return err
}

// Close restores the terminal state.
func (t *Terminal) Close() error {
	t.StopResizeWatch()
	t.DisableRawMode()
	t.DisableMouse()
	t.ShowCursor()
	t.DisableAlternateScreen()
	return nil
}

// ResizeEvent is sent when terminal is resized
type ResizeEvent struct {
	Width  int
	Height int
}


// StopResizeWatch stops the WatchResize goroutine and unregisters the SIGWINCH
// signal handler. It is safe to call multiple times and from Close().
func (t *Terminal) StopResizeWatch() {
	t.mu.Lock()
	stopResize := t.stopResize
	resizeDone := t.resizeDone
	resizeSignal := t.resizeSignal
	t.stopResize = make(chan struct{})
	t.resizeDone = make(chan struct{})
	t.resizeSignal = nil
	t.mu.Unlock()

	if stopResize != nil {
		close(stopResize)
	}
	if resizeDone != nil {
		<-resizeDone
	}
	if resizeSignal != nil && runtime.GOOS != "windows" {
		signal.Stop(resizeSignal)
	}
}
