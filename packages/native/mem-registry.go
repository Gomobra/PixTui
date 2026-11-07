package native

import (
	"sync"
)

// MemRegistry tracks native memory allocations
type MemRegistry struct {
	mu             sync.RWMutex
	allocations    map[uintptr]int
	nextID         uint64
	totalAllocated int64
}

// NewMemRegistry creates a new MemRegistry
func NewMemRegistry() *MemRegistry {
	return &MemRegistry{
		allocations: make(map[uintptr]int),
	}
}

// Register registers a memory allocation
func (r *MemRegistry) Register(ptr uintptr, size int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.allocations[ptr] = size
	r.totalAllocated += int64(size)
}

// Unregister removes a memory allocation
func (r *MemRegistry) Unregister(ptr uintptr) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if size, ok := r.allocations[ptr]; ok {
		r.totalAllocated -= int64(size)
		delete(r.allocations, ptr)
	}
}

// TotalAllocated returns the total allocated bytes
func (r *MemRegistry) TotalAllocated() int64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.totalAllocated
}

// AllocationCount returns the number of active allocations
func (r *MemRegistry) AllocationCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.allocations)
}
