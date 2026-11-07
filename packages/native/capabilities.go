package native

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// TerminalCapabilities detects terminal capabilities
type TerminalCapabilities struct {
	KittyKeyboard             bool
	KittyGraphics             bool
	RGB                       bool
	ANSI256                   bool
	Unicode                   bool
	SGRPixels                 bool
	ColorSchemeUpdates        bool
	ExplicitWidth             bool
	ScaledText                bool
	Sixel                     bool
	FocusTracking             bool
	Sync                      bool
	BracketedPaste            bool
	Hyperlinks                bool
	OSC52                     bool
	OSC52Support              string
	Notifications             bool
	ExplicitCursorPositioning bool
	Remote                    bool
	ProgramStatus             bool
	Multiplexer               string
	ImageProtocol             string
	Terminal                  TerminalInfo
}

// TerminalInfo describes the terminal
type TerminalInfo struct {
	Name          string
	Version       string
	FromXTVersion bool
}

// DetectCapabilities detects terminal capabilities from environment variables
func DetectCapabilities(env map[string]string) *TerminalCapabilities {
	caps := &TerminalCapabilities{
		Terminal: TerminalInfo{
			Name:          env["TERM"],
			FromXTVersion: false,
		},
	}

	term := env["TERM"]

	// Detect color support
	colorterm := env["COLORTERM"]
	if colorterm == "truecolor" || colorterm == "24bit" {
		caps.RGB = true
		caps.ANSI256 = true
	} else if strings.Contains(term, "256color") {
		caps.ANSI256 = true
	}

	// Detect terminal type
	if strings.Contains(term, "kitty") {
		caps.KittyKeyboard = true
		caps.KittyGraphics = true
		caps.ImageProtocol = "kitty"
		caps.Terminal.Name = "kitty"
	}

	if strings.Contains(term, "sixel") {
		caps.Sixel = true
		caps.ImageProtocol = "sixel"
	}

	// Detect terminal program
	termProgram := env["TERM_PROGRAM"]
	switch termProgram {
	case "iTerm.app":
		caps.Terminal.Name = "iTerm2"
		caps.RGB = true
	case "Apple_Terminal":
		caps.Terminal.Name = "Terminal.app"
	case "tmux":
		caps.Multiplexer = "tmux"
	case "screen":
		caps.Multiplexer = "screen"
	case "WezTerm":
		caps.Terminal.Name = "WezTerm"
		caps.RGB = true
	}

	// Detect tmux/zellij
	if env["TMUX"] != "" {
		caps.Multiplexer = "tmux"
	}

	// Detect remote SSH
	if env["SSH_CLIENT"] != "" || env["SSH_TTY"] != "" || env["SSH_CONNECTION"] != "" {
		caps.Remote = true
	}

	// Detect OS
	caps.Terminal.Version = runtime.GOOS

	return caps
}

// DetectFromEnv detects capabilities from the current environment
func DetectFromEnv() *TerminalCapabilities {
	env := make(map[string]string)
	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			env[parts[0]] = parts[1]
		}
	}
	return DetectCapabilities(env)
}

// GetTerminalSize returns the terminal dimensions from environment or ioctl
func GetTerminalSize() (int, int) {
	// Try environment variables first
	if cols := os.Getenv("COLUMNS"); cols != "" {
		if c, err := strconv.Atoi(cols); err == nil && c > 0 {
			if rows := os.Getenv("LINES"); rows != "" {
				if r, err := strconv.Atoi(rows); err == nil && r > 0 {
					return c, r
				}
			}
			return c, 24
		}
	}

	// Try stty
	if cmd := exec.Command("stty", "size"); cmd != nil {
		if output, err := cmd.Output(); err == nil {
			parts := strings.Fields(string(output))
			if len(parts) == 2 {
				if rows, err := strconv.Atoi(parts[0]); err == nil {
					if cols, err := strconv.Atoi(parts[1]); err == nil {
						return cols, rows
					}
				}
			}
		}
	}

	return 80, 24
}

// IsTTY returns true if the given file descriptor is a terminal
func IsTTY(fd uintptr) bool {
	fi, err := os.Stat(fmt.Sprintf("/proc/self/fd/%d", fd))
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
