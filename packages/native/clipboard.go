package native

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

// Clipboard provides cross-platform clipboard access
type Clipboard struct {
	mu       sync.Mutex
	services map[uint64]*ClipboardService
	nextID   uint64
}

// NewClipboard creates a new Clipboard instance
func NewClipboard() *Clipboard {
	return &Clipboard{
		services: make(map[uint64]*ClipboardService),
	}
}

// Copy copies text to the clipboard using platform-specific commands
func (c *Clipboard) Copy(text string) (ClipboardStatus, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch runtime.GOOS {
	case "linux":
		return c.copyLinux(text)
	case "darwin":
		return c.copyDarwin(text)
	case "windows":
		return c.copyWindows(text)
	default:
		return ClipboardStatusUnsupported, fmt.Errorf("clipboard copy not implemented on %s", runtime.GOOS)
	}
}

// Paste pastes text from the clipboard using platform-specific commands
func (c *Clipboard) Paste() (string, ClipboardStatus, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch runtime.GOOS {
	case "linux":
		return c.pasteLinux()
	case "darwin":
		return c.pasteDarwin()
	case "windows":
		return c.pasteWindows()
	default:
		return "", ClipboardStatusUnsupported, fmt.Errorf("clipboard paste not implemented on %s", runtime.GOOS)
	}
}

// copyLinux copies text to clipboard using wl-copy or xclip
func (c *Clipboard) copyLinux(text string) (ClipboardStatus, error) {
	// Try wl-copy first (Wayland)
	if cmd := exec.Command("wl-copy"); cmd != nil {
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err == nil {
			return ClipboardStatusOK, nil
		}
	}
	// Try xclip (X11)
	if cmd := exec.Command("xclip", "-selection", "clipboard"); cmd != nil {
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err == nil {
			return ClipboardStatusOK, nil
		}
	}
	// Try xsel
	if cmd := exec.Command("xsel", "--clipboard", "--input"); cmd != nil {
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err == nil {
			return ClipboardStatusOK, nil
		}
	}
	return ClipboardStatusUnsupported, errors.New("no clipboard tool found on Linux")
}

// pasteLinux pastes from clipboard using wl-clipboard or xclip
func (c *Clipboard) pasteLinux() (string, ClipboardStatus, error) {
	var output []byte
	var err error

	// Try wl-clipboard first
	if cmd := exec.Command("wl-paste"); cmd != nil {
		output, err = cmd.Output()
		if err == nil {
			return string(output), ClipboardStatusOK, nil
		}
	}
	// Try xclip
	if cmd := exec.Command("xclip", "-selection", "clipboard", "-o"); cmd != nil {
		output, err = cmd.Output()
		if err == nil {
			return string(output), ClipboardStatusOK, nil
		}
	}
	// Try xsel
	if cmd := exec.Command("xsel", "--clipboard", "--output"); cmd != nil {
		output, err = cmd.Output()
		if err == nil {
			return string(output), ClipboardStatusOK, nil
		}
	}
	return "", ClipboardStatusUnsupported, err
}

// copyDarwin copies text to clipboard using pbcopy
func (c *Clipboard) copyDarwin(text string) (ClipboardStatus, error) {
	if cmd := exec.Command("pbcopy"); cmd != nil {
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err == nil {
			return ClipboardStatusOK, nil
		}
	}
	return ClipboardStatusUnsupported, errors.New("pbcopy not available")
}

// pasteDarwin pastes from clipboard using pbpaste
func (c *Clipboard) pasteDarwin() (string, ClipboardStatus, error) {
	if cmd := exec.Command("pbpaste"); cmd != nil {
		output, err := cmd.Output()
		if err == nil {
			return string(output), ClipboardStatusOK, nil
		}
		return "", ClipboardStatusUnsupported, err
	}
	return "", ClipboardStatusUnsupported, errors.New("pbpaste not available")
}

// copyWindows copies text to clipboard using clip
func (c *Clipboard) copyWindows(text string) (ClipboardStatus, error) {
	if cmd := exec.Command("clip"); cmd != nil {
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err == nil {
			return ClipboardStatusOK, nil
		}
	}
	return ClipboardStatusUnsupported, errors.New("clip not available")
}

// pasteWindows pastes from clipboard using powershell
func (c *Clipboard) pasteWindows() (string, ClipboardStatus, error) {
	if cmd := exec.Command("powershell", "-command", "Get-Clipboard"); cmd != nil {
		output, err := cmd.Output()
		if err == nil {
			return string(output), ClipboardStatusOK, nil
		}
		return "", ClipboardStatusUnsupported, err
	}
	return "", ClipboardStatusUnsupported, errors.New("powershell not available")
}

// ClipboardServiceHandle is a handle to a clipboard service
type ClipboardServiceHandle uint64

// ClipboardOperationHandle is a handle to a clipboard operation
type ClipboardOperationHandle uint64

// ClipboardStatus represents clipboard operation status
type ClipboardStatus int

const (
	ClipboardStatusOK          ClipboardStatus = iota
	ClipboardStatusUnsupported ClipboardStatus = iota
	ClipboardStatusBusy        ClipboardStatus = iota
	ClipboardStatusError       ClipboardStatus = iota
)

// ClipboardService represents a platform clipboard service
type ClipboardService struct {
	id   uint64
	kind string
}

// String returns the status as a string
func (s ClipboardStatus) String() string {
	switch s {
	case ClipboardStatusOK:
		return "ok"
	case ClipboardStatusUnsupported:
		return "unsupported"
	case ClipboardStatusBusy:
		return "busy"
	case ClipboardStatusError:
		return "error"
	default:
		return "unknown"
	}
}

var _ = fmt.Sprintf
