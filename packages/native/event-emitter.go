package native

import (
	"sync"
)

// EventEmitter is a simple event emitter
type EventEmitter struct {
	mu        sync.RWMutex
	listeners map[string][]func(interface{})
}

// NewEventEmitter creates a new EventEmitter
func NewEventEmitter() *EventEmitter {
	return &EventEmitter{
		listeners: make(map[string][]func(interface{})),
	}
}

// On registers an event listener
func (e *EventEmitter) On(event string, listener func(interface{})) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.listeners[event] = append(e.listeners[event], listener)
}

// Off removes all listeners for an event
func (e *EventEmitter) Off(event string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.listeners, event)
}

// Emit emits an event to all listeners
func (e *EventEmitter) Emit(event string, data interface{}) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	for _, listener := range e.listeners[event] {
		listener(data)
	}
}

// ListenerCount returns the number of listeners for an event
func (e *EventEmitter) ListenerCount(event string) int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.listeners[event])
}
