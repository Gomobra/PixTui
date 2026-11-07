package native

import (
	"sync"
)

// NativeRenderable represents a renderable that can be painted natively
type NativeRenderable struct {
	mu    sync.RWMutex
	id    uint64
	kind  string
	paint func(*OptimizedBuffer, float64)
}

// NewNativeRenderable creates a new NativeRenderable
func NewNativeRenderable(id uint64, kind string, paint func(*OptimizedBuffer, float64)) *NativeRenderable {
	return &NativeRenderable{
		id:    id,
		kind:  kind,
		paint: paint,
	}
}

// ID returns the renderable ID
func (r *NativeRenderable) ID() uint64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.id
}

// Kind returns the renderable kind
func (r *NativeRenderable) Kind() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.kind
}

// Paint renders the renderable into the buffer
func (r *NativeRenderable) Paint(buffer *OptimizedBuffer, deltaTime float64) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.paint != nil {
		r.paint(buffer, deltaTime)
	}
}

// SetPaint sets the paint function
func (r *NativeRenderable) SetPaint(paint func(*OptimizedBuffer, float64)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.paint = paint
}
