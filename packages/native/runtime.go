package native

import (
	"sync"
)

// Runtime represents the native runtime environment
type Runtime struct {
	mu       sync.RWMutex
	version  string
	platform string
	arch     string
}

// NewRuntime creates a new Runtime instance
func NewRuntime() *Runtime {
	return &Runtime{
		version:  "0.1.0",
		platform: "unknown",
		arch:     "unknown",
	}
}

// Version returns the runtime version
func (r *Runtime) Version() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.version
}

// Platform returns the platform name
func (r *Runtime) Platform() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.platform
}

// Arch returns the architecture
func (r *Runtime) Arch() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.arch
}

// Info returns runtime information
func (r *Runtime) Info() map[string]string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return map[string]string{
		"version":  r.version,
		"platform": r.platform,
		"arch":     r.arch,
	}
}
