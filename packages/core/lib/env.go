package lib

import (
	"sync"
)

// Environment manages environment variables
type Environment struct {
	mu   sync.RWMutex
	vars map[string]string
}

func NewEnvironment() *Environment {
	return &Environment{vars: make(map[string]string)}
}

func (e *Environment) Get(key string) (string, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	v, ok := e.vars[key]
	return v, ok
}

func (e *Environment) Set(key, value string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.vars[key] = value
}

func (e *Environment) Delete(key string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.vars, key)
}
