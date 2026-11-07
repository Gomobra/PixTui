package core

import (
	"errors"
	"fmt"
	"sync"
	"unicode/utf8"

	"github.com/Gomobra/PixTui/packages/native"
)

// BufferAccess provides read access to buffer cell planes. Mirrors
// packages/core/src/buffer.ts BufferAccess.
type BufferAccess struct {
	Width      int
	Height     int
	Generation uint64
	Char       []uint32
	Fg         []uint16
	Bg         []uint16
	Attributes []uint32
}

// NativeResourceOwner owns checked resources independently of the Sessions and
// scenes that use them. Mirrors packages/core/src/buffer.ts NativeResourceOwner.
type NativeResourceOwner interface {
	ResourceContext() *native.ResourceContext
}

// RenderLib is a minimal abstraction over the native render library. The Go
// port does not yet expose the full Zig binding surface, so this is a
// placeholder satisfied by the core render pipeline.
type RenderLib interface {
	CreateContext(options NativeContextOptions) (uint64, error)
	DestroyContext(context uint64) error

	// Session lifecycle
	SessionAttachRenderer(context uint64, session uint64, options NativeSessionRendererOptions)
	CreateSession(context uint64, options NativeSessionOptions) (uint64, error)
	DestroySession(context, session uint64) error
	SessionClose(context, session uint64)
	SessionCancel(context, session uint64)

	// Frame / render
	SessionFrameCommit(context, session uint64, frame NativeSceneFrameRequest, force bool) NativeSessionRenderStatus
	SessionSetScreen(context, session uint64, alternate bool, width, height int, trailingOutput []byte)
	SessionResizeRenderer(context, session uint64, width, height int)
	SessionControl(context, session uint64, command NativeSessionControl) error
	SessionClipboard(context, session uint64, target int, bytes []byte) bool
	SessionNotification(context, session uint64, message, title string) bool
	SessionCompleteOutput(context, session uint64, ticket *NativeOutputTicket, end bool)
	SessionWrite(context, session uint64, bytes []byte) error
	SessionGetCapabilities(context, session uint64) NativeSessionCapabilities
	SessionGetRendererState(context, session uint64) NativeRendererState

	// Terminal
	SessionSetupTerminal(context, session uint64, options NativeSessionTerminalOptions)
	SessionSuspend(context, session uint64)
	SessionResume(context, session uint64)

	// Output pump
	SessionPump(context, session uint64, now int64, max int) *nativePumpResult

	GetYogaHost() YogaHost
}

// YogaHost exposes mutation scoping used by resources.
type YogaHost interface {
	RunMutation(func() error) error
}

// noopYogaHost is a stub YogaHost.
type noopYogaHost struct{}

func (noopYogaHost) RunMutation(func() error) error { return nil }

// NativeContextOptions describes options for creating a native context.
type NativeContextOptions struct {
	ObjectCapacity int
	RenderCellsMax int
}

// NativeSessionOptions configures session output buffering.
type NativeSessionOptions struct {
	ChunkSize       int
	SpanCapacity    int
	MaxBytes        int64
	ControlCapacity int
}

// NativeSessionControl is a control command sent to the session.
type NativeSessionControl struct{}

// NativeSessionRendererOptions configures an attached renderer.
type NativeSessionRendererOptions struct{}

// NativeSessionTerminalOptions configures terminal setup/suspend/resume.
type NativeSessionTerminalOptions struct{}

// NativeOutputTicket is an opaque ticket for an in-flight output write.
type NativeOutputTicket struct{}

// NativeRendererState is renderer state returned by the native session.
type NativeRendererState struct {
	FramePending         bool
	ImageUploadsPending  int
}

// NativeSessionRenderStatus is the status of a frame render.
type NativeSessionRenderStatus int

const (
	NativeSessionRenderStatusIdle NativeSessionRenderStatus = iota
	NativeSessionRenderStatusPending
	NativeSessionRenderStatusClosed
)

// NativeSessionCapabilities reports session capabilities.
type NativeSessionCapabilities struct{}

// noopRenderLib is a stub RenderLib used when no real native binding is bound.
type noopRenderLib struct{}

func (noopRenderLib) CreateContext(options NativeContextOptions) (uint64, error) { return 1, nil }
func (noopRenderLib) DestroyContext(context uint64) error                         { return nil }
func (noopRenderLib) CreateSession(context uint64, options NativeSessionOptions) (uint64, error) {
	return 1, nil
}
func (noopRenderLib) SessionAttachRenderer(context uint64, session uint64, options NativeSessionRendererOptions) {}
func (noopRenderLib) DestroySession(context, session uint64) error { return nil }
func (noopRenderLib) SessionClose(context, session uint64)         {}
func (noopRenderLib) SessionCancel(context, session uint64)        {}
func (noopRenderLib) SessionFrameCommit(context, session uint64, frame NativeSceneFrameRequest, force bool) NativeSessionRenderStatus {
	return NativeSessionRenderStatusIdle
}
func (noopRenderLib) SessionSetScreen(context, session uint64, alternate bool, width, height int, trailingOutput []byte) {}
func (noopRenderLib) SessionResizeRenderer(context, session uint64, width, height int)                              {}
func (noopRenderLib) SessionControl(context, session uint64, command NativeSessionControl) error                    { return nil }
func (noopRenderLib) SessionClipboard(context, session uint64, target int, bytes []byte) bool                       { return false }
func (noopRenderLib) SessionNotification(context, session uint64, message, title string) bool                       { return false }
func (noopRenderLib) SessionCompleteOutput(context, session uint64, ticket *NativeOutputTicket, end bool)           {}
func (noopRenderLib) SessionWrite(context, session uint64, bytes []byte) error                                     { return nil }
func (noopRenderLib) SessionGetCapabilities(context, session uint64) NativeSessionCapabilities                      { return NativeSessionCapabilities{} }
func (noopRenderLib) SessionGetRendererState(context, session uint64) NativeRendererState                          { return NativeRendererState{} }
func (noopRenderLib) SessionSetupTerminal(context, session uint64, options NativeSessionTerminalOptions)            {}
func (noopRenderLib) SessionSuspend(context, session uint64)                                                      {}
func (noopRenderLib) SessionResume(context, session uint64)                                                       {}
func (noopRenderLib) SessionPump(context, session uint64, now int64, max int) *nativePumpResult                   { return nil }
func (noopRenderLib) GetYogaHost() YogaHost                                                                       { return noopYogaHost{} }

// ResourceContext owns checked resources independently of the Sessions and
// scenes that use them. Mirrors packages/core/src/buffer.ts ResourceContext,
// wrapping the native.ResourceContext.
type ResourceContext struct {
	mu        sync.RWMutex
	inner     *native.ResourceContext
	renderLib RenderLib
	context   uint64
	disposed  bool
}

// NewResourceContext creates a new resource-owning context.
func NewResourceContext() *ResourceContext {
	return &ResourceContext{
		inner:     native.NewResourceContext(),
		renderLib: noopRenderLib{},
		context:   1,
	}
}

// ResourceContext returns the underlying native resource context, satisfying
// NativeResourceOwner.
func (rc *ResourceContext) ResourceContext() *native.ResourceContext {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return rc.inner
}

// Context returns the opaque native context handle.
func (rc *ResourceContext) Context() uint64 {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return rc.context
}

// RenderLib returns the render library bound to this context.
func (rc *ResourceContext) RenderLib() RenderLib {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return rc.renderLib
}

// Disposed reports whether the context has been destroyed.
func (rc *ResourceContext) Disposed() bool {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return rc.disposed
}

// assertAlive returns an error if the context has been disposed.
func (rc *ResourceContext) assertAlive() error {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	if rc.disposed {
		return errors.New("Resource Context is destroyed")
	}
	if rc.inner == nil {
		return errors.New("Resource Context is nil")
	}
	return nil
}

// AssertAlive is the public form of assertAlive.
func (rc *ResourceContext) AssertAlive() error {
	return rc.assertAlive()
}

// Destroy destroys the context and releases owned resources.
func (rc *ResourceContext) Destroy() error {
	rc.mu.Lock()
	if rc.disposed {
		rc.mu.Unlock()
		return nil
	}
	rc.disposed = true
	inner := rc.inner
	rc.mu.Unlock()

	if inner != nil {
		if err := inner.Close(); err != nil {
			return err
		}
	}
	return nil
}

// OptimizedBuffer is a framebuffer storing terminal cell state, bound to a
// resource owner's context. Mirrors packages/core/src/buffer.ts OptimizedBuffer.
type OptimizedBuffer struct {
	mu          sync.RWMutex
	id          string
	lib         RenderLib
	owner       NativeResourceOwner
	inner       *native.OptimizedBuffer
	width       int
	height      int
	widthMethod WidthMethod
	respectAlpha bool
	destroyed   bool
}

// bufferIdCounter generates unique buffer ids.
var bufferIdCounter uint64

// nextBufferId returns the next unique buffer id.
func nextBufferId() uint64 {
	bufferIdCounter++
	return bufferIdCounter
}

// Create creates a new context-owned buffer of the given dimensions.
func CreateOptimizedBuffer(width, height int, widthMethod WidthMethod, opts *OptimizedBufferOptions) (*OptimizedBuffer, error) {
	owner := NativeResourceOwner(nil)
	respectAlpha := false
	id := ""
	if opts != nil {
		owner = opts.Owner
		respectAlpha = opts.RespectAlpha
		id = opts.ID
	}
	if owner == nil {
		return nil, fmt.Errorf("OptimizedBuffer requires an explicit resource owner")
	}
	ctx := owner.ResourceContext()
	if ctx == nil {
		return nil, fmt.Errorf("OptimizedBuffer requires an explicit resource owner")
	}
	if err := ctx.AssertAlive(); err != nil {
		return nil, err
	}

	inner := native.NewOptimizedBuffer(width, height)
	if id == "" {
		id = fmt.Sprintf("fb_%d", nextBufferId())
	}
	b := &OptimizedBuffer{
		id:           id,
		lib:          noopRenderLib{},
		owner:        owner,
		inner:        inner,
		width:        width,
		height:       height,
		widthMethod:  widthMethod,
		respectAlpha: respectAlpha,
	}
	return b, nil
}

// OptimizedBufferOptions are options for creating an OptimizedBuffer.
type OptimizedBufferOptions struct {
	Owner       NativeResourceOwner
	RespectAlpha bool
	ID          string
	WidthMethod WidthMethod
}

// ID returns the buffer id.
func (b *OptimizedBuffer) ID() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.id
}

// Width returns the buffer width in cells.
func (b *OptimizedBuffer) Width() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.owner != nil {
		// Session-backed buffers report live dimensions.
	}
	return b.width
}

// Height returns the buffer height in cells.
func (b *OptimizedBuffer) Height() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.height
}

// WidthMethod returns the width method in use.
func (b *OptimizedBuffer) WidthMethod() WidthMethod {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.widthMethod
}

// RespectAlpha reports whether source alpha is respected when composing.
func (b *OptimizedBuffer) RespectAlpha() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.respectAlpha
}

// Destroyed reports whether the buffer has been destroyed.
func (b *OptimizedBuffer) Destroyed() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.destroyed
}

// guard fails loudly if the buffer is destroyed or its owner is gone.
func (b *OptimizedBuffer) guard() error {
	b.mu.RLock()
	destroyed := b.destroyed
	owner := b.owner
	b.mu.RUnlock()
	if destroyed {
		return fmt.Errorf("Buffer %s is destroyed", b.id)
	}
	if owner == nil {
		return fmt.Errorf("Buffer %s has no owner", b.id)
	}
	if err := owner.ResourceContext().AssertAlive(); err != nil {
		return err
	}
	return nil
}

// setRespectAlpha sets whether source alpha is respected when composing.
func (b *OptimizedBuffer) SetRespectAlpha(respectAlpha bool) error {
	if err := b.guard(); err != nil {
		return err
	}
	b.mu.Lock()
	b.respectAlpha = respectAlpha
	b.mu.Unlock()
	return nil
}

// Clear fills the buffer with the given background color.
func (b *OptimizedBuffer) Clear(bg RGBA) error {
	if err := b.guard(); err != nil {
		return err
	}
	b.inner.Clear()
	return nil
}

// SetCell sets a single cell.
func (b *OptimizedBuffer) SetCell(x, y int, char string, fg, bg RGBA, attributes uint32) error {
	if err := b.guard(); err != nil {
		return err
	}
	var ch uint32 = ' '
	if r, ok := charDecoration(char); ok {
		ch = r
	}
	b.inner.SetCell(x, y, ch, packRGBA16(fg), packRGBA16(bg), attributes)
	return nil
}

// charDecoration returns the code point of the first char of a string.
func charDecoration(s string) (uint32, bool) {
	if r, _ := utf8.DecodeRuneInString(s); r != utf8.RuneError {
		return uint32(r), true
	}
	return 0, false
}

// packRGBA16 packs an RGBA into a 16-bit value (5-6-5 style approximation).
func packRGBA16(c RGBA) uint16 {
	return uint16(c.R>>3)<<11 | uint16(c.G>>2)<<5 | uint16(c.B>>3)
}

// FillRect fills a rectangular region with a background color.
func (b *OptimizedBuffer) FillRect(x, y, width, height int, bg RGBA) error {
	if err := b.guard(); err != nil {
		return err
	}
	if width <= 0 || height <= 0 {
		return nil
	}
	b.inner.FillRect(x, y, width, height, ' ', packRGBA16(bg), packRGBA16(bg), 0)
	return nil
}

// Access returns a snapshot of the buffer cell planes.
func (b *OptimizedBuffer) Access() (BufferAccess, error) {
	if err := b.guard(); err != nil {
		return BufferAccess{}, err
	}
	access := b.inner.Access()
	return BufferAccess{
		Width:      access.Width,
		Height:     access.Height,
		Generation: access.Generation,
		Char:       access.Char,
		Fg:         access.Fg,
		Bg:         access.Bg,
		Attributes: access.Attributes,
	}, nil
}

// Resize resizes a context-owned buffer.
func (b *OptimizedBuffer) Resize(width, height int) error {
	if err := b.guard(); err != nil {
		return err
	}
	if b.width == width && b.height == height {
		return nil
	}
	b.inner.Resize(width, height)
	b.mu.Lock()
	b.width = width
	b.height = height
	b.mu.Unlock()
	return nil
}

// Destroy releases the buffer's native storage after the current paint.
func (b *OptimizedBuffer) Destroy() error {
	b.mu.Lock()
	if b.destroyed {
		b.mu.Unlock()
		return nil
	}
	b.destroyed = true
	inner := b.inner
	b.mu.Unlock()
	if inner != nil {
		inner.Clear()
	}
	return nil
}
