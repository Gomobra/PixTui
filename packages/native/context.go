package native

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// Status codes mirroring opentui.h
const (
	OT_OK                            = 0
	OT_INVALID_ARGUMENT              = -1
	OT_UNSUPPORTED_VERSION           = -2
	OT_OUT_OF_MEMORY                 = -3
	OT_WRONG_THREAD                  = -4
	OT_INTERNAL_ERROR                = -5
	OT_CONTEXT_BUSY                  = -6
	OT_WRONG_CONTEXT                 = -7
	OT_WRONG_KIND                    = -8
	OT_STALE_HANDLE                  = -9
	OT_WRONG_SESSION                 = -10
	OT_OUTPUT_BACKPRESSURE           = -11
	OT_SESSION_CLOSED                = -12
	OT_OUTPUT_BUSY                   = -13
	OT_STALE_OUTPUT                  = -14
	OT_OUTPUT_FAILED                 = -15
	OT_OBJECT_LIMIT                  = -16
	OT_RENDERER_ATTACHED             = -17
	OT_RENDERER_NOT_ATTACHED         = -18
	OT_INVALID_PHASE                 = -19
	OT_CONTROL_PACKET_LIMIT          = -20
	OT_LEASE_LIMIT                   = -21
	OT_LEASE_BYTES_LIMIT             = -22
	OT_STALE_LEASE                   = -23
	OT_UNSUPPORTED_RESOURCE          = -24
	OT_STALE_FRAME                   = -25
	OT_LAYOUT_LIMIT                  = -26
	OT_FRAME_BUSY                    = -27
	OT_FRAME_REQUEST_LIMIT           = -28
	OT_IMAGE_UNSUPPORTED_FORMAT      = -40
	OT_IMAGE_UNSUPPORTED_COLOR_SPACE = -41
	OT_IMAGE_MALFORMED_DATA          = -42
	OT_IMAGE_DIMENSION_LIMIT         = -43
	OT_IMAGE_MEMORY_LIMIT            = -44
	OT_IMAGE_OUTPUT_TOO_SMALL        = -45
	OT_IMAGE_UNSUPPORTED_FEATURE     = -46
	OT_IMAGE_BUSY                    = -47
	OT_VIDEO_FAILED                  = -60
	OT_AUDIO_DECODE_FAILED           = -61
	OT_AUDIO_DEVICE_FAILED           = -62
)

// Scene node kinds
const (
	OT_SCENE_ROOT      = 0
	OT_SCENE_BOX       = 1
	OT_SCENE_TEXT      = 2
	OT_SCENE_SLIDER    = 3
	OT_SCENE_ARROW     = 4
	OT_SCENE_EDITOR    = 5
	OT_SCENE_CUSTOM    = 6
	OT_SCENE_TEXT_VIEW = 7
	OT_SCENE_IMAGE     = 8
	OT_SCENE_VIDEO     = 9
)

// Scroll options
const (
	OT_SCENE_SCROLL_X             = 1
	OT_SCENE_SCROLL_Y             = 2
	OT_SCENE_SCROLL_FOLLOW_NONE   = 0
	OT_SCENE_SCROLL_FOLLOW_TOP    = 1
	OT_SCENE_SCROLL_FOLLOW_BOTTOM = 2
	OT_SCENE_SCROLL_FOLLOW_LEFT   = 3
	OT_SCENE_SCROLL_FOLLOW_RIGHT  = 4
)

// Border masks
const (
	OT_BORDER_NONE   = 0
	OT_BORDER_LEFT   = 1
	OT_BORDER_BOTTOM = 2
	OT_BORDER_RIGHT  = 4
	OT_BORDER_TOP    = 8
	OT_BORDER_ALL    = 15
)

// Border styles
const (
	OT_BORDER_SINGLE  = 0
	OT_BORDER_DOUBLE  = 1
	OT_BORDER_ROUNDED = 2
	OT_BORDER_HEAVY   = 3
)

// Color intents
const (
	OT_COLOR_RGB     = 0
	OT_COLOR_INDEXED = 1
	OT_COLOR_DEFAULT = 2
)

// ContextStatus represents lifecycle state
type ContextStatus int

const (
	ContextStatusActive  ContextStatus = iota
	ContextStatusClosing ContextStatus = iota
	ContextStatusClosed  ContextStatus = iota
)

// Handle mirrors ot_handle from opentui.h
type Handle struct {
	ContextID  uint64
	Slot       uint32
	Generation uint32
}

// IsValid returns true if the handle is not all-zero
func (h Handle) IsValid() bool {
	return h.ContextID != 0 || h.Slot != 0 || h.Generation != 0
}

// ZeroHandle is an invalid handle
var ZeroHandle = Handle{}

// ResourceContext owns checked resources independently of Sessions and scenes
type ResourceContext struct {
	mu          sync.RWMutex
	id          uint64
	status      ContextStatus
	allocator   *Allocator
	sessions    map[uint64]*Session
	buffers     map[uint64]*OptimizedBuffer
	textBuffers map[uint64]*TextBuffer
	editBuffers map[uint64]*EditBuffer
	editorViews map[uint64]*EditorView
	images      map[uint64]*Image
	videos      map[uint64]*Video
	nextSlot    uint64
}

// Allocator manages native memory allocations
type Allocator struct {
	mu        sync.Mutex
	allocated int64
	limit     int64
}

// NewAllocator creates a new allocator with optional byte limit
func NewAllocator(limit int64) *Allocator {
	return &Allocator{limit: limit}
}

// Alloc allocates bytes (tracked)
func (a *Allocator) Alloc(size int) []byte {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.limit > 0 && a.allocated+int64(size) > a.limit {
		return nil
	}
	a.allocated += int64(size)
	return make([]byte, size)
}

// Free releases tracked bytes
func (a *Allocator) Free(ptr []byte) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.allocated -= int64(cap(ptr))
	_ = ptr
}

// Stats returns allocation statistics
func (a *Allocator) Stats() (allocated int64, limit int64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.allocated, a.limit
}

// NewResourceContext creates a new resource-owning context
func NewResourceContext() *ResourceContext {
	return &ResourceContext{
		id:          1,
		status:      ContextStatusActive,
		allocator:   NewAllocator(0),
		sessions:    make(map[uint64]*Session),
		buffers:     make(map[uint64]*OptimizedBuffer),
		textBuffers: make(map[uint64]*TextBuffer),
		editBuffers: make(map[uint64]*EditBuffer),
		editorViews: make(map[uint64]*EditorView),
		images:      make(map[uint64]*Image),
		videos:      make(map[uint64]*Video),
		nextSlot:    1,
	}
}

// ID returns the context ID
func (c *ResourceContext) ID() uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.id
}

// Status returns the context status
func (c *ResourceContext) Status() ContextStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.status
}

// assertAlive returns error if context is not active
func (c *ResourceContext) assertAlive() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.status != ContextStatusActive {
		return fmt.Errorf("context %d is closed", c.id)
	}
	return nil
}

// AssertAlive returns an error if the context is not active.
func (c *ResourceContext) AssertAlive() error {
	return c.assertAlive()
}

// Close destroys the context and all associated resources
func (c *ResourceContext) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.status == ContextStatusClosed {
		return nil
	}
	c.status = ContextStatusClosing

	for _, s := range c.sessions {
		s.Close()
	}
	c.sessions = make(map[uint64]*Session)
	c.buffers = make(map[uint64]*OptimizedBuffer)
	c.textBuffers = make(map[uint64]*TextBuffer)
	c.editBuffers = make(map[uint64]*EditBuffer)
	c.editorViews = make(map[uint64]*EditorView)
	c.images = make(map[uint64]*Image)
	c.videos = make(map[uint64]*Video)

	c.status = ContextStatusClosed
	return nil
}

// nextSlot generates the next resource slot ID
func (c *ResourceContext) nextSlotID() uint64 {
	slot := atomic.AddUint64(&c.nextSlot, 1)
	return slot
}

// CreateSession creates a new Session owned by this context
func (c *ResourceContext) CreateSession() (*Session, error) {
	if err := c.assertAlive(); err != nil {
		return nil, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	id := c.nextSlotID()
	s := NewSession(id, c)
	c.sessions[id] = s
	return s, nil
}

// GetSession returns a session by ID
func (c *ResourceContext) GetSession(id uint64) *Session {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.sessions[id]
}

// DestroySession destroys a session by ID
func (c *ResourceContext) DestroySession(id uint64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	s, ok := c.sessions[id]
	if !ok {
		return fmt.Errorf("session %d not found", id)
	}
	s.Close()
	delete(c.sessions, id)
	return nil
}

// CreateBuffer creates a new OptimizedBuffer owned by this context
func (c *ResourceContext) CreateBuffer(width, height int) (*OptimizedBuffer, error) {
	if err := c.assertAlive(); err != nil {
		return nil, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	id := c.nextSlotID()
	buf := NewOptimizedBuffer(width, height)
	c.buffers[id] = buf
	return buf, nil
}

// GetBuffer returns a buffer by ID
func (c *ResourceContext) GetBuffer(id uint64) *OptimizedBuffer {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.buffers[id]
}

// DestroyBuffer destroys a buffer by ID
func (c *ResourceContext) DestroyBuffer(id uint64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.buffers[id]; !ok {
		return fmt.Errorf("buffer %d not found", id)
	}
	delete(c.buffers, id)
	return nil
}

// CreateTextBuffer creates a new TextBuffer
func (c *ResourceContext) CreateTextBuffer() (*TextBuffer, error) {
	if err := c.assertAlive(); err != nil {
		return nil, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	id := c.nextSlotID()
	tb := NewTextBuffer()
	tb.id = id
	c.textBuffers[id] = tb
	return tb, nil
}

// GetTextBuffer returns a text buffer by ID
func (c *ResourceContext) GetTextBuffer(id uint64) *TextBuffer {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.textBuffers[id]
}

// CreateEditBuffer creates a new EditBuffer
func (c *ResourceContext) CreateEditBuffer() (*EditBuffer, error) {
	if err := c.assertAlive(); err != nil {
		return nil, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	id := c.nextSlotID()
	eb := NewEditBuffer()
	eb.id = id
	c.editBuffers[id] = eb
	return eb, nil
}

// GetEditBuffer returns an edit buffer by ID
func (c *ResourceContext) GetEditBuffer(id uint64) *EditBuffer {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.editBuffers[id]
}

// CreateEditorView creates a new EditorView
func (c *ResourceContext) CreateEditorView(buffer *EditBuffer) (*EditorView, error) {
	if err := c.assertAlive(); err != nil {
		return nil, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	id := c.nextSlotID()
	ev := NewEditorView(buffer)
	ev.id = id
	c.editorViews[id] = ev
	return ev, nil
}

// GetEditorView returns an editor view by ID
func (c *ResourceContext) GetEditorView(id uint64) *EditorView {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.editorViews[id]
}

// CreateImage creates a new Image
func (c *ResourceContext) CreateImage(width, height int) (*Image, error) {
	if err := c.assertAlive(); err != nil {
		return nil, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	id := c.nextSlotID()
	img := NewImage(width, height)
	img.ID = id
	c.images[id] = img
	return img, nil
}

// GetImage returns an image by ID
func (c *ResourceContext) GetImage(id uint64) *Image {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.images[id]
}

// CreateVideo creates a new Video
func (c *ResourceContext) CreateVideo(width, height int) (*Video, error) {
	if err := c.assertAlive(); err != nil {
		return nil, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	id := c.nextSlotID()
	vid := NewVideo(width, height)
	vid.ID = id
	c.videos[id] = vid
	return vid, nil
}

// GetVideo returns a video by ID
func (c *ResourceContext) GetVideo(id uint64) *Video {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.videos[id]
}

// AllocatorStats returns allocator statistics
type AllocatorStats struct {
	Allocated int64
	Limit     int64
}

// AllocatorStats returns the current allocator stats
func (c *ResourceContext) AllocatorStats() AllocatorStats {
	allocated, limit := c.allocator.Stats()
	return AllocatorStats{Allocated: allocated, Limit: limit}
}
