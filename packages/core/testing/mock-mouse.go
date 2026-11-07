package testing

import (
	"fmt"

	"github.com/Gomobra/PixTui/packages/core"
)

// MouseButton represents a mouse button code, matching the TypeScript MouseButtons constants.
type MouseButton uint32

const (
	MouseButtonLeft     MouseButton = 0
	MouseButtonMiddle   MouseButton = 1
	MouseButtonRight    MouseButton = 2
	MouseButtonWheelUp      MouseButton = 64 // scroll flag + 0
	MouseButtonWheelDown    MouseButton = 65 // scroll flag + 1
	MouseButtonWheelLeft    MouseButton = 66 // scroll flag + 2
	MouseButtonWheelRight   MouseButton = 67 // scroll flag + 3
)

// MockMouseEvent represents a mock mouse event for testing mouse handling.
type MockMouseEvent struct {
	X, Y      int
	Button    uint32
	Pressed   bool
	Modifiers uint32
}

// NewMockMouseEvent creates a new MockMouseEvent at the given position with the given button and pressed state.
func NewMockMouseEvent(x, y int, button uint32, pressed bool) *MockMouseEvent {
	return &MockMouseEvent{
		X:      x,
		Y:      y,
		Button: button,
		Pressed: pressed,
	}
}

// Position returns the (x, y) coordinates of the mouse event.
func (m *MockMouseEvent) Position() (int, int) {
	return m.X, m.Y
}

// IsPressed returns whether the button is in the pressed state.
func (m *MockMouseEvent) IsPressed() bool {
	return m.Pressed
}

// IsLeftButton returns true if the event involves the left mouse button.
func (m *MockMouseEvent) IsLeftButton() bool {
	return m.Button == uint32(MouseButtonLeft)
}

// IsRightButton returns true if the event involves the right mouse button.
func (m *MockMouseEvent) IsRightButton() bool {
	return m.Button == uint32(MouseButtonRight)
}

// IsMiddleButton returns true if the event involves the middle mouse button.
func (m *MockMouseEvent) IsMiddleButton() bool {
	return m.Button == uint32(MouseButtonMiddle)
}

// IsWheel returns true if the event is a wheel (scroll) event.
func (m *MockMouseEvent) IsWheel() bool {
	return m.Button >= 64
}

// MockMouse simulates mouse input for testing.
type MockMouse struct {
	renderer  *TestRenderer
	x, y      int
	buttons   map[uint32]bool
}

// NewMockMouse creates a new MockMouse for the given TestRenderer.
func NewMockMouse(renderer *TestRenderer) *MockMouse {
	return &MockMouse{
		renderer: renderer,
		buttons:  make(map[uint32]bool),
		x:        0,
		y:        0,
	}
}

// Position returns the current mouse position.
func (m *MockMouse) Position() (int, int) {
	return m.x, m.y
}

// PressedButtons returns the currently pressed button codes.
func (m *MockMouse) PressedButtons() []uint32 {
	var result []uint32
	for b := range m.buttons {
		result = append(result, b)
	}
	return result
}

// IsButtonPressed returns whether the given button is currently pressed.
func (m *MockMouse) IsButtonPressed(button uint32) bool {
	return m.buttons[button]
}

// generateMouseEvent generates an SGR mouse event sequence.
// SGR format: \x1b[<b;x;yM or \x1b[<b;x;ym
func generateMouseEvent(eventType string, x, y int, button uint32, modifiers uint32, buttonsPressed map[uint32]bool) string {
	buttonCode := button

	// Add modifier flags (SGR: shift=4, alt=8, ctrl=16)
	if modifiers&MockKeyModifierShift != 0 {
		buttonCode |= 4
	}
	if modifiers&MockKeyModifierMeta != 0 {
		buttonCode |= 8
	}
	if modifiers&MockKeyModifierCtrl != 0 {
		buttonCode |= 16
	}

	switch eventType {
	case "move":
		buttonCode = 32 | 3 // motion flag (32) + button 3 for motion without button press
		if modifiers&MockKeyModifierShift != 0 {
			buttonCode |= 4
		}
		if modifiers&MockKeyModifierMeta != 0 {
			buttonCode |= 8
		}
		if modifiers&MockKeyModifierCtrl != 0 {
			buttonCode |= 16
		}
	case "drag":
		firstButton := uint32(0)
		for b := range buttonsPressed {
			firstButton = b
			break
		}
		if firstButton == 0 {
			firstButton = button
		}
		buttonCode = firstButton | 32
		if modifiers&MockKeyModifierShift != 0 {
			buttonCode |= 4
		}
		if modifiers&MockKeyModifierMeta != 0 {
			buttonCode |= 8
		}
		if modifiers&MockKeyModifierCtrl != 0 {
			buttonCode |= 16
		}
	}

	ansiX := x + 1
	ansiY := y + 1

	pressRelease := "M"
	if eventType == "up" || eventType == "move" || eventType == "drag" {
		pressRelease = "m"
	}

	return fmt.Sprintf("\x1b[<%d;%d;%d%s", buttonCode, ansiX, ansiY, pressRelease)
}

// EmitMouseEvent emits a raw mouse event as an SGR sequence.
func (m *MockMouse) EmitMouseEvent(event core.MouseEvent) {
	eventType := "move"
	if event.Pressed {
		eventType = "down"
	}

	seq := generateMouseEvent(eventType, event.X, event.Y, event.Button, event.Modifiers, m.buttons)
	m.renderer.FeedInput([]byte(seq))

	m.x = event.X
	m.y = event.Y

	if eventType == "down" && event.Button < 64 {
		m.buttons[event.Button] = true
	} else if eventType == "up" {
		delete(m.buttons, event.Button)
	}
}

// MoveTo moves the mouse to the given position.
func (m *MockMouse) MoveTo(x, y int) {
	if len(m.buttons) > 0 {
		var firstButton uint32
		for b := range m.buttons {
			firstButton = b
			break
		}
		seq := generateMouseEvent("drag", x, y, firstButton, 0, m.buttons)
		m.renderer.FeedInput([]byte(seq))
	} else {
		seq := generateMouseEvent("move", x, y, uint32(MouseButtonLeft), 0, m.buttons)
		m.renderer.FeedInput([]byte(seq))
	}
	m.x = x
	m.y = y
}

// Click clicks at the given position (press then release).
func (m *MockMouse) Click(x, y int, button uint32) {
	m.PressDown(x, y, button)
	m.Release(x, y, button)
}

// DoubleClick double-clicks at the given position.
func (m *MockMouse) DoubleClick(x, y int, button uint32) {
	m.Click(x, y, button)
	m.Click(x, y, button)
}

// PressDown presses a mouse button at the given position.
func (m *MockMouse) PressDown(x, y int, button uint32) {
	seq := generateMouseEvent("down", x, y, button, 0, m.buttons)
	m.renderer.FeedInput([]byte(seq))
	m.x = x
	m.y = y
	if button < 64 {
		m.buttons[button] = true
	}
}

// Release releases a mouse button at the given position.
func (m *MockMouse) Release(x, y int, button uint32) {
	seq := generateMouseEvent("up", x, y, button, 0, m.buttons)
	m.renderer.FeedInput([]byte(seq))
	m.x = x
	m.y = y
	delete(m.buttons, button)
}

// Drag drags from one position to another, emitting intermediate drag events.
func (m *MockMouse) Drag(fromX, fromY, toX, toY int, button uint32) {
	m.PressDown(fromX, fromY, button)

	steps := 5
	dx := (toX - fromX) / steps
	dy := (toY - fromY) / steps

	for i := 1; i <= steps; i++ {
		curX := fromX + dx*i
		curY := fromY + dy*i
		seq := generateMouseEvent("drag", curX, curY, button, 0, m.buttons)
		m.renderer.FeedInput([]byte(seq))
		m.x = curX
		m.y = curY
	}

	m.Release(toX, toY, button)
}

// Scroll scrolls at the given position based on deltaX/deltaY.
func (m *MockMouse) Scroll(x, y int, deltaX, deltaY int) {
	var button uint32
	switch {
	case deltaY > 0:
		button = uint32(MouseButtonWheelUp)
	case deltaY < 0:
		button = uint32(MouseButtonWheelDown)
	case deltaX > 0:
		button = uint32(MouseButtonWheelRight)
	case deltaX < 0:
		button = uint32(MouseButtonWheelLeft)
	default:
		return
	}

	seq := generateMouseEvent("scroll", x, y, button, 0, m.buttons)
	m.renderer.FeedInput([]byte(seq))
	m.x = x
	m.y = y
}
