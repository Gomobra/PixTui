package testing

import (
	"fmt"
	"strings"

	"github.com/Gomobra/PixTui/packages/keymap"
)

// Modifier bit flags for MockKeyEvent. Matches the kitty keyboard modifier encoding
// used by the TypeScript reference (shift=1, meta=2, ctrl=4, super=8, hyper=16).
const (
	MockKeyModifierShift = 1 << iota // 1
	MockKeyModifierMeta              // 2
	MockKeyModifierCtrl              // 4
	MockKeyModifierSuper             // 8
	MockKeyModifierHyper             // 16
)

// MockKeyEvent is a mock implementation of keymap.KeymapEvent for testing key handling.
type MockKeyEvent struct {
	name               string
	ctrl, shift, meta, super, hyper bool
	preventedDefault   bool
	propagationStopped bool
}

// NewMockKeyEvent creates a new MockKeyEvent with the given name and optional modifier bit flags.
// Modifiers are combined using the MockKeyModifier* constants (e.g. MockKeyModifierCtrl|MockKeyModifierShift).
func NewMockKeyEvent(name string, modifiers ...uint32) *MockKeyEvent {
	k := &MockKeyEvent{name: name}
	for _, m := range modifiers {
		if m&MockKeyModifierShift != 0 {
			k.shift = true
		}
		if m&MockKeyModifierMeta != 0 {
			k.meta = true
		}
		if m&MockKeyModifierCtrl != 0 {
			k.ctrl = true
		}
		if m&MockKeyModifierSuper != 0 {
			k.super = true
		}
		if m&MockKeyModifierHyper != 0 {
			k.hyper = true
		}
	}
	return k
}

// Name returns the key name.
func (k *MockKeyEvent) Name() string {
	return k.name
}

// Ctrl returns whether the Ctrl modifier is active.
func (k *MockKeyEvent) Ctrl() bool {
	return k.ctrl
}

// Shift returns whether the Shift modifier is active.
func (k *MockKeyEvent) Shift() bool {
	return k.shift
}

// Meta returns whether the Meta modifier is active.
func (k *MockKeyEvent) Meta() bool {
	return k.meta
}

// Super returns whether the Super modifier is active.
func (k *MockKeyEvent) Super() bool {
	return k.super
}

// Hyper returns whether the Hyper modifier is active.
func (k *MockKeyEvent) Hyper() bool {
	return k.hyper
}

// PreventDefault marks the event as default-prevented.
func (k *MockKeyEvent) PreventDefault() {
	k.preventedDefault = true
}

// StopPropagation stops event propagation.
func (k *MockKeyEvent) StopPropagation() {
	k.propagationStopped = true
}

// PropagationStopped returns whether propagation has been stopped.
func (k *MockKeyEvent) PropagationStopped() bool {
	return k.propagationStopped
}

// Compile-time interface check.
var _ keymap.KeymapEvent = (*MockKeyEvent)(nil)

// Kitty keyboard protocol key mappings.
var kittyKeyCodeMap = map[string]int{
	"escape":   27,
	"tab":      9,
	"return":   13,
	"backspace": 127,
	"insert":   57348,
	"delete":   57349,
	"left":     57350,
	"right":    57351,
	"up":       57352,
	"down":     57353,
	"pageup":   57354,
	"pagedown": 57355,
	"home":     57356,
	"end":      57357,
	"f1":       57364,
	"f2":       57365,
	"f3":       57366,
	"f4":       57367,
	"f5":       57368,
	"f6":       57369,
	"f7":       57370,
	"f8":       57371,
	"f9":       57372,
	"f10":      57373,
	"f11":      57374,
	"f12":      57375,
}

// KeyCodes for special keys (ANSI escape sequences).
var keyCodes = map[string]string{
	"RETURN":    "\r",
	"LINEFEED":  "\n",
	"TAB":       "\t",
	"BACKSPACE": "\b",
	"DELETE":    "\x1b[3~",
	"HOME":      "\x1b[H",
	"END":       "\x1b[F",
	"ESCAPE":    "\x1b",
	"ARROW_UP":  "\x1b[A",
	"ARROW_DOWN": "\x1b[B",
	"ARROW_RIGHT": "\x1b[C",
	"ARROW_LEFT":  "\x1b[D",
	"F1": "\x1bOP",
	"F2": "\x1bOQ",
	"F3": "\x1bOR",
	"F4": "\x1bOS",
	"F5": "\x1b[15~",
	"F6": "\x1b[17~",
	"F7": "\x1b[18~",
	"F8": "\x1b[19~",
	"F9": "\x1b[20~",
	"F10": "\x1b[21~",
	"F11": "\x1b[23~",
	"F12": "\x1b[24~",
}

// MockKeysOptions configures the mock keyboard.
type MockKeysOptions struct {
	KittyKeyboard      bool
	OtherModifiersMode bool
}

// MockKeys simulates keyboard input for testing.
type MockKeys struct {
	renderer          *TestRenderer
	useKittyKeyboard  bool
	useOtherModifiers bool
}

// NewMockKeys creates a new MockKeys for the given TestRenderer.
func NewMockKeys(renderer *TestRenderer) *MockKeys {
	return &MockKeys{
		renderer: renderer,
	}
}

// resolveKeyInput resolves a key string to its byte value and name.
func resolveKeyInput(key string) (string, string) {
	if v, ok := keyCodes[key]; ok {
		return v, strings.ToLower(key)
	}
	return key, ""
}

// encodeKittySequence encodes a key using the Kitty keyboard protocol.
// CSI <codepoint>u or CSI <codepoint>;<modMask+1>u
// Modifier encoding: shift=1, meta=2, ctrl=4, super=8, hyper=16.
func encodeKittySequence(codepoint int, shift, meta, ctrl, super, hyper bool) string {
	modMask := 0
	if shift {
		modMask |= 1
	}
	if meta {
		modMask |= 2
	}
	if ctrl {
		modMask |= 4
	}
	if super {
		modMask |= 8
	}
	if hyper {
		modMask |= 16
	}

	if modMask == 0 {
		return fmt.Sprintf("\x1b[%du", codepoint)
	}
	return fmt.Sprintf("\x1b[%d;%du", codepoint, modMask+1)
}

// encodeModifyOtherKeysSequence encodes a key using the modifyOtherKeys protocol.
// CSI 27;<modMask+1>;<charCode>~
func encodeModifyOtherKeysSequence(charCode int, shift, meta, ctrl, super, hyper bool) string {
	modMask := 0
	if shift {
		modMask |= 1
	}
	if meta {
		modMask |= 2
	}
	if ctrl {
		modMask |= 4
	}
	if super {
		modMask |= 8
	}
	if hyper {
		modMask |= 16
	}

	if modMask == 0 {
		return string(rune(charCode))
	}
	return fmt.Sprintf("\x1b[27;%d;%d~", modMask+1, charCode)
}

// valueToKeyNameMap maps control characters/escape sequences to kitty key names.
var valueToKeyNameMap = map[string]string{
	"\b":   "backspace",
	"\r":   "return",
	"\n":   "return",
	"\t":   "tab",
	"\x1b": "escape",
	"\x1b[A": "up",
	"\x1b[B": "down",
	"\x1b[C": "right",
	"\x1b[D": "left",
	"\x1b[H": "home",
	"\x1b[F": "end",
	"\x1b[3~": "delete",
}

// PressKeys presses multiple keys in sequence.
func (m *MockKeys) PressKeys(keys ...string) {
	for _, key := range keys {
		m.PressKey(key)
	}
}

// PressKey presses a single key with optional modifier bit flags.
// Modifiers are combined using MockKeyModifier* constants.
func (m *MockKeys) PressKey(key string, modifiers ...uint32) {
	shift := false
	meta := false
	ctrl := false
	super := false
	hyper := false

	for _, mod := range modifiers {
		if mod&MockKeyModifierShift != 0 {
			shift = true
		}
		if mod&MockKeyModifierMeta != 0 {
			meta = true
		}
		if mod&MockKeyModifierCtrl != 0 {
			ctrl = true
		}
		if mod&MockKeyModifierSuper != 0 {
			super = true
		}
		if mod&MockKeyModifierHyper != 0 {
			hyper = true
		}
	}

	keyValue, keyName := resolveKeyInput(key)

	// Kitty keyboard protocol mode takes precedence.
	if m.useKittyKeyboard {
		if mapped, ok := valueToKeyNameMap[keyValue]; ok {
			keyName = mapped
		}
		if strings.HasPrefix(keyName, "arrow_") {
			keyName = keyName[6:]
		}
		if code, ok := kittyKeyCodeMap[keyName]; ok {
			seq := encodeKittySequence(code, shift, meta, ctrl, super, hyper)
			m.renderer.FeedInput([]byte(seq))
			return
		}
		if len(keyValue) == 1 && !strings.HasPrefix(keyValue, "\x1b") {
			codepoint := int(rune(keyValue[0]))
			seq := encodeKittySequence(codepoint, shift, meta, ctrl, super, hyper)
			m.renderer.FeedInput([]byte(seq))
			return
		}
	}

	// ModifyOtherKeys mode (CSI u protocol variant).
	if m.useOtherModifiers && !m.useKittyKeyboard && len(modifiers) > 0 {
		valueToCharCodeMap := map[string]int{
			"\b":   127,
			"\r":   13,
			"\n":   13,
			"\t":   9,
			"\x1b": 27,
			" ":    32,
		}

		var charCode int
		if v, ok := valueToCharCodeMap[keyValue]; ok {
			charCode = v
		} else if len(keyValue) == 1 && !strings.HasPrefix(keyValue, "\x1b") {
			charCode = int(rune(keyValue[0]))
		}

		if charCode != 0 {
			seq := encodeModifyOtherKeysSequence(charCode, shift, meta, ctrl, super, hyper)
			m.renderer.FeedInput([]byte(seq))
			return
		}
	}

	// Regular (non-Kitty, non-modifyOtherKeys) mode.
	keyCode := keyValue

	if len(modifiers) > 0 {
		if strings.HasPrefix(keyCode, "\x1b[") && len(keyCode) > 2 {
			modifier := 1
			if shift {
				modifier += 1
			}
			if meta {
				modifier += 2
			}
			if ctrl {
				modifier += 4
			}
			if super {
				modifier += 8
			}
			if hyper {
				modifier += 16
			}

			if modifier > 1 {
				// Check for CSI sequence like \x1b[3~ (delete, insert, etc.)
				if len(keyCode) >= 4 && keyCode[len(keyCode)-1] == '~' {
					n := keyCode[2 : len(keyCode)-1]
					keyCode = fmt.Sprintf("\x1b[%s;%d~", n, modifier)
				} else {
					// Arrow keys and other single-letter endings
					ending := keyCode[len(keyCode)-1:]
					keyCode = fmt.Sprintf("\x1b[1;%d%s", modifier, ending)
				}
			}
		} else if len(keyCode) == 1 {
			char := keyCode
			if char == "\t" && shift {
				if meta {
					keyCode = "\x1b\x1b[Z"
				} else {
					keyCode = "\x1b[Z"
				}
				m.renderer.FeedInput([]byte(keyCode))
				return
			}
			if char == "\b" && (ctrl || super || hyper) {
				modifier := 1
				if shift {
					modifier += 1
				}
				if meta {
					modifier += 2
				}
				if ctrl {
					modifier += 4
				}
				if super {
					modifier += 8
				}
				if hyper {
					modifier += 16
				}
				keyCode = fmt.Sprintf("\x1b[27;%d;127~", modifier)
			} else if ctrl {
				if char >= "a" && char <= "z" {
					keyCode = string(rune(char[0] - 96))
				} else if char >= "A" && char <= "Z" {
					keyCode = string(rune(char[0] - 64))
				} else {
					specialCtrlMap := map[string]string{
						"[": "\x1b",
						"\\": "\x1c",
						"]": "\x1d",
						"^": "\x1e",
						"_": "\x1f",
						"?": "\x7f",
						"/": "\x1f",
						"-": "\x1f",
						".": "\x1e",
						",": "\x1c",
						"@": "\x00",
						" ": "\x00",
					}
					if v, ok := specialCtrlMap[char]; ok {
						keyCode = v
					}
				}
				if meta {
					keyCode = "\x1b" + keyCode
				}
			} else {
				if shift && char >= "a" && char <= "z" {
					char = strings.ToUpper(char)
				}
				if meta {
					keyCode = "\x1b" + char
				} else {
					keyCode = char
				}
			}
		} else if meta && !strings.HasPrefix(keyCode, "\x1b") {
			keyCode = "\x1b" + keyCode
		}
	}

	m.renderer.FeedInput([]byte(keyCode))
}

// TypeText types text character by character.
func (m *MockKeys) TypeText(text string) {
	for _, ch := range text {
		m.PressKey(string(ch))
	}
}

// PressEnter presses the Enter key.
func (m *MockKeys) PressEnter(modifiers ...uint32) {
	m.PressKey("\r", modifiers...)
}

// PressEscape presses the Escape key.
func (m *MockKeys) PressEscape(modifiers ...uint32) {
	m.PressKey("\x1b", modifiers...)
}

// PressTab presses the Tab key.
func (m *MockKeys) PressTab(modifiers ...uint32) {
	m.PressKey("\t", modifiers...)
}

// PressBackspace presses the Backspace key.
func (m *MockKeys) PressBackspace(modifiers ...uint32) {
	m.PressKey("\b", modifiers...)
}

// PressArrow presses an arrow key (up, down, left, right).
func (m *MockKeys) PressArrow(direction string, modifiers ...uint32) {
	keyMap := map[string]string{
		"up":    "\x1b[A",
		"down":  "\x1b[B",
		"right": "\x1b[C",
		"left":  "\x1b[D",
	}
	key, ok := keyMap[direction]
	if !ok {
		panic(fmt.Sprintf("unknown arrow direction: %q", direction))
	}
	m.PressKey(key, modifiers...)
}

// PressCtrlC sends Ctrl+C.
func (m *MockKeys) PressCtrlC() {
	m.PressKey("c", MockKeyModifierCtrl)
}

// PasteBracketedText pastes text using bracketed paste mode.
func (m *MockKeys) PasteBracketedText(text string) {
	m.renderer.FeedInput([]byte("\x1b[200~"))
	m.renderer.FeedInput([]byte(text))
	m.renderer.FeedInput([]byte("\x1b[201~"))
}

// EnableKitty enables or disables Kitty keyboard protocol mode.
// Kitty takes precedence over OtherModifiersMode.
func (m *MockKeys) EnableKitty(enabled bool) {
	m.useKittyKeyboard = enabled
	if enabled {
		m.useOtherModifiers = false
	}
}

// EnableOtherModifiers enables or disables modifyOtherKeys mode.
func (m *MockKeys) EnableOtherModifiers(enabled bool) {
	m.useOtherModifiers = enabled
	if enabled {
		m.useKittyKeyboard = false
	}
}

// IsKittyEnabled returns whether Kitty keyboard protocol mode is active.
func (m *MockKeys) IsKittyEnabled() bool {
	return m.useKittyKeyboard
}

// IsOtherModifiersEnabled returns whether modifyOtherKeys mode is active.
func (m *MockKeys) IsOtherModifiersEnabled() bool {
	return m.useOtherModifiers
}
