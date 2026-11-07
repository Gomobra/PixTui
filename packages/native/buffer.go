package native

import (
	"sync"
)

// BufferAccess provides read access to buffer cell arrays
type BufferAccess struct {
	Width      int
	Height     int
	Generation uint64
	Char       []uint32
	Fg         []uint16
	Bg         []uint16
	Attributes []uint32
}

// OptimizedBuffer is a high-performance double-buffering memory structure
// used to store terminal cell state. Mirrors packages/native/src/buffer.zig.
type OptimizedBuffer struct {
	mu           sync.RWMutex
	width        int
	height       int
	gen          uint64
	char         []uint32
	fg           []uint16
	bg           []uint16
	attrs        []uint32
	nextChar     []uint32
	nextFg       []uint16
	nextBg       []uint16
	nextAttrs    []uint32
	scissorStack []ScissorRect
	opacityStack []float32
	alphaStack   []bool
}

// ScissorRect represents a clipping rectangle
type ScissorRect struct {
	X, Y, Width, Height int
}

// NewOptimizedBuffer creates a new buffer with the given dimensions
func NewOptimizedBuffer(width, height int) *OptimizedBuffer {
	size := width * height
	b := &OptimizedBuffer{
		width:     width,
		height:    height,
		char:      make([]uint32, size),
		fg:        make([]uint16, size),
		bg:        make([]uint16, size),
		attrs:     make([]uint32, size),
		nextChar:  make([]uint32, size),
		nextFg:    make([]uint16, size),
		nextBg:    make([]uint16, size),
		nextAttrs: make([]uint32, size),
	}
	for i := 0; i < size; i++ {
		b.char[i] = ' '
		b.nextChar[i] = ' '
	}
	return b
}

// Width returns the buffer width in cells
func (b *OptimizedBuffer) Width() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.width
}

// Height returns the buffer height in cells
func (b *OptimizedBuffer) Height() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.height
}

// Generation returns the current generation counter
func (b *OptimizedBuffer) Generation() uint64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.gen
}

// Access returns a snapshot of the current buffer state
func (b *OptimizedBuffer) Access() BufferAccess {
	b.mu.RLock()
	defer b.mu.RUnlock()
	size := b.width * b.height
	access := BufferAccess{
		Width:      b.width,
		Height:     b.height,
		Generation: b.gen,
		Char:       make([]uint32, size),
		Fg:         make([]uint16, size),
		Bg:         make([]uint16, size),
		Attributes: make([]uint32, size),
	}
	copy(access.Char, b.char)
	copy(access.Fg, b.fg)
	copy(access.Bg, b.bg)
	copy(access.Attributes, b.attrs)
	return access
}

// Clear fills the buffer with spaces and default colors
func (b *OptimizedBuffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	size := b.width * b.height
	for i := 0; i < size; i++ {
		b.char[i] = ' '
		b.fg[i] = 0
		b.bg[i] = 0
		b.attrs[i] = 0
	}
	b.gen++
}

// SetCell sets a single cell at (x, y) with scissor and opacity applied
func (b *OptimizedBuffer) SetCell(x, y int, ch uint32, fg, bg uint16, attrs uint32) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if x < 0 || x >= b.width || y < 0 || y >= b.height {
		return
	}
	// Apply scissor stack
	for _, rect := range b.scissorStack {
		if x < rect.X || x >= rect.X+rect.Width || y < rect.Y || y >= rect.Y+rect.Height {
			return
		}
	}
	idx := y*b.width + x
	b.char[idx] = ch
	b.fg[idx] = fg
	b.bg[idx] = bg
	b.attrs[idx] = attrs
}

// GetCell returns the cell at (x, y)
func (b *OptimizedBuffer) GetCell(x, y int) (uint32, uint16, uint16, uint32, bool) {
	if x < 0 || x >= b.width || y < 0 || y >= b.height {
		return 0, 0, 0, 0, false
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	idx := y*b.width + x
	return b.char[idx], b.fg[idx], b.bg[idx], b.attrs[idx], true
}

// FillRect fills a rectangular region
func (b *OptimizedBuffer) FillRect(x, y, w, h int, ch uint32, fg, bg uint16, attrs uint32) {
	if x < 0 {
		w += x
		x = 0
	}
	if y < 0 {
		h += y
		y = 0
	}
	if x+w > b.width {
		w = b.width - x
	}
	if y+h > b.height {
		h = b.height - y
	}
	if w <= 0 || h <= 0 {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	for row := 0; row < h; row++ {
		idx := (y+row)*b.width + x
		for col := 0; col < w; col++ {
			b.char[idx+col] = ch
			b.fg[idx+col] = fg
			b.bg[idx+col] = bg
			b.attrs[idx+col] = attrs
		}
	}
	b.gen++
}

// Resize changes the buffer dimensions
func (b *OptimizedBuffer) Resize(width, height int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if width == b.width && height == b.height {
		return
	}
	size := width * height
	newChar := make([]uint32, size)
	newFg := make([]uint16, size)
	newBg := make([]uint16, size)
	newAttrs := make([]uint32, size)
	nextChar := make([]uint32, size)
	nextFg := make([]uint16, size)
	nextBg := make([]uint16, size)
	nextAttrs := make([]uint32, size)

	for i := 0; i < size; i++ {
		newChar[i] = ' '
		nextChar[i] = ' '
	}

	minW := b.width
	if width < minW {
		minW = width
	}
	minH := b.height
	if height < minH {
		minH = height
	}
	for row := 0; row < minH; row++ {
		srcIdx := row * b.width
		dstIdx := row * width
		copy(newChar[dstIdx:dstIdx+minW], b.char[srcIdx:srcIdx+minW])
		copy(newFg[dstIdx:dstIdx+minW], b.fg[srcIdx:srcIdx+minW])
		copy(newBg[dstIdx:dstIdx+minW], b.bg[srcIdx:srcIdx+minW])
		copy(newAttrs[dstIdx:dstIdx+minW], b.attrs[srcIdx:srcIdx+minW])
	}

	b.width = width
	b.height = height
	b.char = newChar
	b.fg = newFg
	b.bg = newBg
	b.attrs = newAttrs
	b.nextChar = nextChar
	b.nextFg = nextFg
	b.nextBg = nextBg
	b.nextAttrs = nextAttrs
	b.gen++
}

// Swap swaps current and next buffers (double-buffering)
func (b *OptimizedBuffer) Swap() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.char, b.nextChar = b.nextChar, b.char
	b.fg, b.nextFg = b.nextFg, b.fg
	b.bg, b.nextBg = b.nextBg, b.bg
	b.attrs, b.nextAttrs = b.nextAttrs, b.attrs
	b.gen++
}

// NextAccess returns access to the next (pending) buffer for paint recording
func (b *OptimizedBuffer) NextAccess() BufferAccess {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return BufferAccess{
		Width:      b.width,
		Height:     b.height,
		Generation: b.gen,
		Char:       b.nextChar,
		Fg:         b.nextFg,
		Bg:         b.nextBg,
		Attributes: b.nextAttrs,
	}
}

// CommitNext makes the next buffer the current buffer
func (b *OptimizedBuffer) CommitNext() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.char, b.nextChar = b.nextChar, b.char
	b.fg, b.nextFg = b.nextFg, b.fg
	b.bg, b.nextBg = b.nextBg, b.bg
	b.attrs, b.nextAttrs = b.nextAttrs, b.attrs
	b.gen++
}

// PushScissor pushes a scissor rectangle onto the stack
func (b *OptimizedBuffer) PushScissor(x, y, width, height int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.scissorStack = append(b.scissorStack, ScissorRect{X: x, Y: y, Width: width, Height: height})
}

// PopScissor pops the last scissor rectangle from the stack
func (b *OptimizedBuffer) PopScissor() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.scissorStack) > 0 {
		b.scissorStack = b.scissorStack[:len(b.scissorStack)-1]
	}
}

// PushOpacity pushes an opacity multiplier onto the stack
func (b *OptimizedBuffer) PushOpacity(opacity float32) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.opacityStack = append(b.opacityStack, opacity)
}

// PopOpacity pops the last opacity from the stack
func (b *OptimizedBuffer) PopOpacity() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.opacityStack) > 0 {
		b.opacityStack = b.opacityStack[:len(b.opacityStack)-1]
	}
}

// CombinedOpacity returns the combined opacity from the stack
func (b *OptimizedBuffer) CombinedOpacity() float32 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	result := float32(1.0)
	for _, o := range b.opacityStack {
		result *= o
	}
	return result
}

// PushAlpha pushes an alpha blending flag onto the stack
func (b *OptimizedBuffer) PushAlpha(enabled bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.alphaStack = append(b.alphaStack, enabled)
}

// PopAlpha pops the last alpha flag from the stack
func (b *OptimizedBuffer) PopAlpha() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.alphaStack) > 0 {
		b.alphaStack = b.alphaStack[:len(b.alphaStack)-1]
	}
}

// AlphaEnabled returns whether alpha blending is enabled
func (b *OptimizedBuffer) AlphaEnabled() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if len(b.alphaStack) == 0 {
		return false
	}
	return b.alphaStack[len(b.alphaStack)-1]
}
