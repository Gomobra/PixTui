package native

import (
	"fmt"
)

// CliRenderer is the central orchestrator of the rendering pipeline.
// Mirrors packages/core/src/renderer.ts (CliRenderer) and packages/native/src/renderer.zig.
type CliRenderer struct {
	context *ResourceContext
	session *Session
	width   int
	height  int

	// Buffer
	buffer *OptimizedBuffer

	// Render state
	dirty     bool
	frameID   uint64
	targetFPS float64
	maxFPS    float64

	// Terminal capabilities
	capabilities map[string]bool

	// Event handlers
	onResize func(width, height int)
	onKey    func(data []byte)
	onMouse  func(x, y int, button uint32, pressed bool)
	onPaste  func(data []byte)
}

// NewCliRenderer creates a new CliRenderer.
func NewCliRenderer(context *ResourceContext, session *Session) *CliRenderer {
	return &CliRenderer{
		context:      context,
		session:      session,
		width:        80,
		height:       24,
		buffer:       NewOptimizedBuffer(80, 24),
		targetFPS:    30,
		maxFPS:       60,
		capabilities: make(map[string]bool),
	}
}

// SetDimensions sets the renderer dimensions.
func (r *CliRenderer) SetDimensions(width, height int) {
	r.width = width
	r.height = height
	r.buffer.Resize(width, height)
	if r.session != nil {
		r.session.SetDimensions(width, height)
	}
	if r.onResize != nil {
		r.onResize(width, height)
	}
}

// Dimensions returns the current dimensions.
func (r *CliRenderer) Dimensions() (int, int) {
	return r.width, r.height
}

// Buffer returns the render buffer.
func (r *CliRenderer) Buffer() *OptimizedBuffer {
	return r.buffer
}

// Render performs a render pass.
func (r *CliRenderer) Render() error {
	if r.session == nil {
		return fmt.Errorf("no session attached")
	}
	r.frameID++
	r.dirty = false

	// Clear buffer
	r.buffer.Clear()

	// Render scene graph
	if r.session.rootNode != nil {
		r.renderNode(r.session.rootNode, 0)
	}

	return nil
}

// renderNode renders a single scene node and its children.
func (r *CliRenderer) renderNode(node *SceneNode, depth int) {
	// Apply paint options
	if node.paint != nil && node.paint.ShouldFill {
		r.buffer.FillRect(
			int(node.x), int(node.y),
			int(node.width), int(node.height),
			' ', 0, 0, 0,
		)
	}

	// Render text nodes
	if node.kind == SceneNodeKindText && node.text != "" {
		r.drawText(node)
	}

	// Render box borders
	if node.kind == SceneNodeKindBox && node.boxDetails != nil {
		r.drawBox(node)
	}

	// Render children
	for _, child := range node.children {
		r.renderNode(child, depth+1)
	}
}

// drawText renders text content of a text node.
func (r *CliRenderer) drawText(node *SceneNode) {
	x := int(node.x)
	y := int(node.y)
	for _, ch := range node.text {
		if ch == '\n' {
			x = int(node.x)
			y++
			continue
		}
		if x >= r.width || y >= r.height {
			break
		}
		fg := uint16(0)
		bg := uint16(0)
		if node.textOptions != nil {
			fg = uint16(node.textOptions.Foreground.R)<<8 | uint16(node.textOptions.Foreground.G)
			bg = uint16(node.textOptions.Background.R)<<8 | uint16(node.textOptions.Background.G)
		}
		r.buffer.SetCell(x, y, uint32(ch), fg, bg, node.textOptions.Attributes)
		x++
	}
}

// drawBox renders a box node with borders.
func (r *CliRenderer) drawBox(node *SceneNode) {
	if node.boxDetails == nil {
		return
	}
	x := int(node.x)
	y := int(node.y)
	w := int(node.width)
	h := int(node.height)
	if w < 2 || h < 2 {
		return
	}

	borderChars := node.boxDetails.CustomBorderChars
	topLeft := uint32('┌')
	topRight := uint32('┐')
	bottomLeft := uint32('└')
	bottomRight := uint32('┘')
	horiz := uint32('─')
	vert := uint32('│')
	if len(borderChars) > 0 && borderChars[0] != 0 {
		topLeft = borderChars[0]
	}
	if len(borderChars) > 1 && borderChars[1] != 0 {
		topRight = borderChars[1]
	}
	if len(borderChars) > 2 && borderChars[2] != 0 {
		bottomLeft = borderChars[2]
	}
	if len(borderChars) > 3 && borderChars[3] != 0 {
		bottomRight = borderChars[3]
	}
	if len(borderChars) > 4 && borderChars[4] != 0 {
		horiz = borderChars[4]
	}
	if len(borderChars) > 5 && borderChars[5] != 0 {
		vert = borderChars[5]
	}

	for col := x + 1; col < x+w-1; col++ {
		r.buffer.SetCell(col, y, horiz, 0, 0, 0)
		r.buffer.SetCell(col, y+h-1, horiz, 0, 0, 0)
	}
	for row := y + 1; row < y+h-1; row++ {
		r.buffer.SetCell(x, row, vert, 0, 0, 0)
		r.buffer.SetCell(x+w-1, row, vert, 0, 0, 0)
	}
	r.buffer.SetCell(x, y, topLeft, 0, 0, 0)
	r.buffer.SetCell(x+w-1, y, topRight, 0, 0, 0)
	r.buffer.SetCell(x, y+h-1, bottomLeft, 0, 0, 0)
	r.buffer.SetCell(x+w-1, y+h-1, bottomRight, 0, 0, 0)
}

// SetResizeHandler sets the resize event handler.
func (r *CliRenderer) SetResizeHandler(handler func(width, height int)) {
	r.onResize = handler
}

// SetKeyHandler sets the key event handler.
func (r *CliRenderer) SetKeyHandler(handler func(data []byte)) {
	r.onKey = handler
}

// SetMouseHandler sets the mouse event handler.
func (r *CliRenderer) SetMouseHandler(handler func(x, y int, button uint32, pressed bool)) {
	r.onMouse = handler
}

// SetPasteHandler sets the paste event handler.
func (r *CliRenderer) SetPasteHandler(handler func(data []byte)) {
	r.onPaste = handler
}

// Capture returns a snapshot of the current frame for testing.
func (r *CliRenderer) Capture() BufferAccess {
	return r.buffer.Access()
}
