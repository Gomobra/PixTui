package native

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"unicode/utf8"
)

// InputParser parses terminal input sequences (escape sequences, kitty protocol, mouse)
// Mirrors packages/native/src/input-parser.zig with real implementation.
type InputParser struct {
	mu             sync.Mutex
	buffer         []byte
	kitty          bool
	mouse          bool
	bracketedPaste bool
}

// NewInputParser creates a new InputParser
func NewInputParser() *InputParser {
	return &InputParser{
		buffer: make([]byte, 0, 1024),
	}
}

// Parse parses raw input bytes into events
func (p *InputParser) Parse(data []byte) ([]InputEvent, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.buffer = append(p.buffer, data...)

	var events []InputEvent

	for len(p.buffer) > 0 {
		event, consumed, err := p.parseOne()
		if err != nil {
			return events, err
		}
		if consumed == 0 {
			break
		}
		p.buffer = p.buffer[consumed:]
		if event.Type != InputEventNone {
			events = append(events, event)
		}
	}

	return events, nil
}

// parseOne parses a single input event from the buffer
func (p *InputParser) parseOne() (InputEvent, int, error) {
	if len(p.buffer) == 0 {
		return InputEvent{Type: InputEventNone}, 0, nil
	}

	b := p.buffer[0]

	// Regular printable character
	if b >= 32 && b < 127 {
		return InputEvent{
			Type: InputEventChar,
			Char: rune(b),
			Raw:  []byte{b},
		}, 1, nil
	}

	// Escape sequence starts with ESC (0x1b)
	if b == 27 {
		if len(p.buffer) == 1 {
			return InputEvent{Type: InputEventEscape}, 1, nil
		}

		// CSI sequence: ESC [ ...
		if p.buffer[1] == '[' {
			return p.parseCSI()
		}

		// Escape key itself or ESC + char (Alt+char)
		if p.buffer[1] >= 32 && p.buffer[1] < 127 {
			return InputEvent{
				Type: InputEventChar,
				Char: rune(p.buffer[1]),
				Raw:  p.buffer[:2],
				Mod:  ModAlt,
			}, 2, nil
		}

		return InputEvent{Type: InputEventEscape}, 1, nil
	}

	// Control characters (C0 controls)
	if b < 32 {
		return InputEvent{
			Type: InputEventControl,
			Char: rune(b),
			Raw:  []byte{b},
		}, 1, nil
	}

	// UTF-8 multi-byte sequence
	if b&0xE0 == 0xC0 && len(p.buffer) >= 2 {
		r, size := decodeUTF8(p.buffer)
		if size > 0 {
			return InputEvent{
				Type: InputEventChar,
				Char: r,
				Raw:  p.buffer[:size],
			}, size, nil
		}
	}

	return InputEvent{Type: InputEventUnknown}, 1, nil
}

// parseCSI parses a CSI escape sequence (ESC [ ... final_byte)
func (p *InputParser) parseCSI() (InputEvent, int, error) {
	// Find the end of the CSI sequence (final byte in range 0x40-0x7E)
	end := 2
	for end < len(p.buffer) {
		c := p.buffer[end]
		if c >= 0x40 && c <= 0x7E {
			end++
			break
		}
		if c >= 0x20 && c <= 0x3F {
			// Parameter bytes, continue
			end++
			continue
		}
		// Invalid byte in CSI, abort
		return InputEvent{Type: InputEventUnknown}, 1, nil
	}

	if end == 2 {
		return InputEvent{Type: InputEventNone}, 0, nil
	}

	seq := string(p.buffer[2 : end-1])
	final := p.buffer[end-1]

	// Mouse events (SGR/URXVT mode: ESC [ < cb ; cx ; cy M/m)
	if p.mouse && final == 'M' || final == 'm' {
		mouseEvent := p.parseMouseSeq(seq, final)
		if mouseEvent.Type != InputEventNone {
			return mouseEvent, end, nil
		}
	}

	// Bracketed paste: ESC [ 200 ~ ... ESC [ 201 ~
	if seq == "200" && final == '~' {
		p.bracketedPaste = true
		return InputEvent{Type: InputEventPasteStart}, end, nil
	}
	if seq == "201" && final == '~' {
		p.bracketedPaste = false
		return InputEvent{Type: InputEventPasteEnd}, end, nil
	}

	// Kitty keyboard protocol
	if p.kitty {
		kittyEvent := p.parseKittySeq(seq, final)
		if kittyEvent.Type != InputEventNone {
			return kittyEvent, end, nil
		}
	}

	// Arrow keys and function keys
	switch seq {
	case "A":
		return InputEvent{Type: InputEventKey, Key: KeyUp}, end, nil
	case "B":
		return InputEvent{Type: InputEventKey, Key: KeyDown}, end, nil
	case "C":
		return InputEvent{Type: InputEventKey, Key: KeyRight}, end, nil
	case "D":
		return InputEvent{Type: InputEventKey, Key: KeyLeft}, end, nil
	case "H":
		return InputEvent{Type: InputEventKey, Key: KeyHome}, end, nil
	case "F":
		return InputEvent{Type: InputEventKey, Key: KeyEnd}, end, nil
	case "Z":
		return InputEvent{Type: InputEventKey, Key: KeyBackTab}, end, nil
	}

	// Function keys F1-F12 (simple)
	if len(seq) == 1 {
		switch seq[0] {
		case '1':
			return InputEvent{Type: InputEventKey, Key: KeyF1}, end, nil
		case '2':
			return InputEvent{Type: InputEventKey, Key: KeyF2}, end, nil
		case '3':
			return InputEvent{Type: InputEventKey, Key: KeyF3}, end, nil
		case '4':
			return InputEvent{Type: InputEventKey, Key: KeyF4}, end, nil
		case '5':
			return InputEvent{Type: InputEventKey, Key: KeyF5}, end, nil
		case '6':
			return InputEvent{Type: InputEventKey, Key: KeyF6}, end, nil
		case '7':
			return InputEvent{Type: InputEventKey, Key: KeyF7}, end, nil
		case '8':
			return InputEvent{Type: InputEventKey, Key: KeyF8}, end, nil
		case '9':
			return InputEvent{Type: InputEventKey, Key: KeyF9}, end, nil
		case '0':
			return InputEvent{Type: InputEventKey, Key: KeyF10}, end, nil
		}
	}

	// Modified keys with modifiers (ESC [ modifier ; key ~ or ESC [ modifier key)
	// e.g., ESC [ 1 ; 5 A for Ctrl+Up
	if strings.Contains(seq, ";") {
		return p.parseModifiedKey(seq, final, end)
	}

	return InputEvent{Type: InputEventCSI, Raw: p.buffer[:end]}, end, nil
}

// parseMouseSeq parses mouse escape sequences
func (p *InputParser) parseMouseSeq(seq string, final byte) InputEvent {
	// Format: < cb ; cx ; cy M or < cb ; cx ; cy m
	parts := strings.Split(seq, ";")
	if len(parts) < 3 {
		return InputEvent{Type: InputEventNone}
	}

	// Remove leading '<' from first part
	cbStr := strings.TrimPrefix(parts[0], "<")
	var cb, cx, cy int
	fmt.Sscanf(cbStr, "%d", &cb)
	fmt.Sscanf(parts[1], "%d", &cx)
	fmt.Sscanf(parts[2], "%d", &cy)

	button := uint32(cb & 0x03)
	pressed := final == 'M' // 'M' = pressed, 'm' = released
	modifiers := uint32((cb >> 2) & 0x03)

	return InputEvent{
		Type:      InputEventMouse,
		Button:    button,
		Pressed:   pressed,
		Modifiers: modifiers,
		X:         cx,
		Y:         cy,
	}
}

// parseModifiedKey parses keys with modifiers
func (p *InputParser) parseModifiedKey(seq string, final byte, end int) (InputEvent, int, error) {
	parts := strings.Split(seq, ";")
	if len(parts) < 2 {
		return InputEvent{Type: InputEventNone}, 0, nil
	}

	var modifier int
	fmt.Sscanf(parts[0], "%d", &modifier)
	keyPart := parts[len(parts)-1]

	var key string
	switch keyPart {
	case "A":
		key = KeyUp
	case "B":
		key = KeyDown
	case "C":
		key = KeyRight
	case "D":
		key = KeyLeft
	case "H":
		key = KeyHome
	case "F":
		key = KeyEnd
	default:
		if len(keyPart) == 1 {
			switch keyPart[0] {
			case '1':
				key = KeyF1
			case '2':
				key = KeyF2
			case '3':
				key = KeyF3
			case '4':
				key = KeyF4
			case '5':
				key = KeyF5
			case '6':
				key = KeyF6
			case '7':
				key = KeyF7
			case '8':
				key = KeyF8
			case '9':
				key = KeyF9
			case '0':
				key = KeyF10
			}
		}
	}

	var mods uint32
	if modifier&1 != 0 {
		mods |= ModShift
	}
	if modifier&2 != 0 {
		mods |= ModAlt
	}
	if modifier&4 != 0 {
		mods |= ModCtrl
	}
	if modifier&8 != 0 {
		mods |= ModSuper
	}

	return InputEvent{
		Type:      InputEventKey,
		Key:       key,
		Modifiers: mods,
	}, end, nil
}

// parseKittySeq parses Kitty keyboard protocol sequences
func (p *InputParser) parseKittySeq(seq string, final byte) InputEvent {
	// Kitty keyboard protocol: ESC [ < codepoint ; modifiers u
	// Simplified implementation
	if final != 'u' {
		return InputEvent{Type: InputEventNone}
	}

	parts := strings.Split(seq, ";")
	if len(parts) < 2 {
		return InputEvent{Type: InputEventNone}
	}

	// Remove leading '<' from first part
	codeStr := strings.TrimPrefix(parts[0], "<")
	var codepoint int
	fmt.Sscanf(codeStr, "%d", &codepoint)

	var mods uint32
	fmt.Sscanf(parts[1], "%d", &mods)

	return InputEvent{
		Type:      InputEventKey,
		Char:      rune(codepoint),
		Modifiers: mods,
	}
}

// decodeUTF8 decodes a UTF-8 sequence from the buffer
func decodeUTF8(buf []byte) (rune, int) {
	if len(buf) < 2 {
		return 0, 0
	}

	b := buf[0]
	var size int
	var prefix byte

	switch {
	case b&0x80 == 0x00:
		size = 1
	case b&0xE0 == 0xC0:
		size = 2
		prefix = 0x1F
	case b&0xF0 == 0xE0:
		size = 3
		prefix = 0x0F
	case b&0xF8 == 0xF0:
		size = 4
		prefix = 0x07
	default:
		return 0, 0
	}

	if len(buf) < size {
		return 0, 0
	}

	// Validate continuation bytes
	for i := 1; i < size; i++ {
		if buf[i]&0xC0 != 0x80 {
			return 0, 0
		}
	}

	r := rune(b & prefix)
	for i := 1; i < size; i++ {
		r = r<<6 | rune(buf[i]&0x3F)
	}

	// Validate the rune
	if !utf8.ValidRune(r) {
		return 0, 0
	}

	return r, size
}

// InputEvent represents a parsed input event
type InputEvent struct {
	Type      InputEventType
	Key       string
	Char      rune
	Mod       uint32
	Modifiers uint32
	Button    uint32
	Pressed   bool
	X, Y      int
	Raw       []byte
}

// InputEventType represents the type of input event
type InputEventType int

const (
	InputEventNone InputEventType = iota
	InputEventChar
	InputEventControl
	InputEventEscape
	InputEventKey
	InputEventMouse
	InputEventCSI
	InputEventPasteStart
	InputEventPasteEnd
	InputEventUnknown
)

// Key constants
const (
	KeyUp      = "Up"
	KeyDown    = "Down"
	KeyLeft    = "Left"
	KeyRight   = "Right"
	KeyHome    = "Home"
	KeyEnd     = "End"
	KeyBackTab = "BackTab"
	KeyF1      = "F1"
	KeyF2      = "F2"
	KeyF3      = "F3"
	KeyF4      = "F4"
	KeyF5      = "F5"
	KeyF6      = "F6"
	KeyF7      = "F7"
	KeyF8      = "F8"
	KeyF9      = "F9"
	KeyF10     = "F10"
	KeyF11     = "F11"
	KeyF12     = "F12"
)

// Modifier constants
const (
	ModShift uint32 = 1 << 0
	ModAlt   uint32 = 1 << 1
	ModCtrl  uint32 = 1 << 2
	ModSuper uint32 = 1 << 3
)

// X, Y accessors for mouse events
func (e *InputEvent) MouseX() int { return e.X }
func (e *InputEvent) MouseY() int { return e.Y }

// EnableKitty enables Kitty keyboard protocol
func (p *InputParser) EnableKitty(enabled bool) {
	p.kitty = enabled
}

// EnableMouse enables mouse tracking
func (p *InputParser) EnableMouse(enabled bool) {
	p.mouse = enabled
}

// EnableBracketedPaste enables bracketed paste mode
func (p *InputParser) EnableBracketedPaste(enabled bool) {
	p.bracketedPaste = enabled
}

// Reset resets the parser state
func (p *InputParser) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.buffer = make([]byte, 0, 1024)
}

var _ = errors.New
