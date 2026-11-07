package lib

import (
	"fmt"
	"sync"

	"github.com/Gomobra/PixTui/packages/native"
)

// KeyEventType represents the type of a key event.
type KeyEventType string

const (
	KeyEventTypePress   KeyEventType = "press"
	KeyEventTypeRelease KeyEventType = "release"
)

// ParsedKey is the interface that a parsed key must satisfy.
type ParsedKey interface {
	Name() string
	Ctrl() bool
	Meta() bool
	Shift() bool
	Option() bool
	Sequence() string
	Number() bool
	Raw() string
	EventType() KeyEventType
	Source() string
	Code() string
	Super() bool
	Hyper() bool
	CapsLock() bool
	NumLock() bool
	BaseCode() int
	Repeated() bool
}

// KeyEvent represents a keyboard key event.
type KeyEvent struct {
	name       string
	ctrl       bool
	meta       bool
	shift      bool
	option     bool
	sequence   string
	number     bool
	raw        string
	eventType  KeyEventType
	source     string
	code       string
	super      bool
	hyper      bool
	capsLock   bool
	numLock    bool
	baseCode   int
	repeated   bool

	defaultPrevented   bool
	propagationStopped bool
}

// NewKeyEvent creates a new KeyEvent from a ParsedKey.
func NewKeyEvent(key ParsedKey) *KeyEvent {
	return &KeyEvent{
		name:       key.Name(),
		ctrl:       key.Ctrl(),
		meta:       key.Meta(),
		shift:      key.Shift(),
		option:     key.Option(),
		sequence:   key.Sequence(),
		number:     key.Number(),
		raw:        key.Raw(),
		eventType:  key.EventType(),
		source:     key.Source(),
		code:       key.Code(),
		super:      key.Super(),
		hyper:      key.Hyper(),
		capsLock:   key.CapsLock(),
		numLock:    key.NumLock(),
		baseCode:   key.BaseCode(),
		repeated:   key.Repeated(),
	}
}

// Name returns the key name.
func (k *KeyEvent) Name() string { return k.name }

// Ctrl returns whether Ctrl is active.
func (k *KeyEvent) Ctrl() bool { return k.ctrl }

// Meta returns whether Meta is active.
func (k *KeyEvent) Meta() bool { return k.meta }

// Shift returns whether Shift is active.
func (k *KeyEvent) Shift() bool { return k.shift }

// Option returns whether Option is active.
func (k *KeyEvent) Option() bool { return k.option }

// Sequence returns the key sequence.
func (k *KeyEvent) Sequence() string { return k.sequence }

// Number returns whether it's a number key.
func (k *KeyEvent) Number() bool { return k.number }

// Raw returns the raw input string.
func (k *KeyEvent) Raw() string { return k.raw }

// EventType returns the event type.
func (k *KeyEvent) EventType() KeyEventType { return k.eventType }

// Source returns the event source ("raw" or "kitty").
func (k *KeyEvent) Source() string { return k.source }

// Code returns the key code.
func (k *KeyEvent) Code() string { return k.code }

// Super returns whether Super is active.
func (k *KeyEvent) Super() bool { return k.super }

// Hyper returns whether Hyper is active.
func (k *KeyEvent) Hyper() bool { return k.hyper }

// CapsLock returns whether CapsLock is active.
func (k *KeyEvent) CapsLock() bool { return k.capsLock }

// NumLock returns whether NumLock is active.
func (k *KeyEvent) NumLock() bool { return k.numLock }

// BaseCode returns the base code.
func (k *KeyEvent) BaseCode() int { return k.baseCode }

// Repeated returns whether the key is repeated.
func (k *KeyEvent) Repeated() bool { return k.repeated }

// DefaultPrevented returns whether default was prevented.
func (k *KeyEvent) DefaultPrevented() bool { return k.defaultPrevented }

// PropagationStopped returns whether propagation was stopped.
func (k *KeyEvent) PropagationStopped() bool { return k.propagationStopped }

// PreventDefault marks the event as default-prevented.
func (k *KeyEvent) PreventDefault() { k.defaultPrevented = true }

// StopPropagation stops event propagation.
func (k *KeyEvent) StopPropagation() { k.propagationStopped = true }

// PasteEvent represents a paste event.
type PasteEvent struct {
	Type      string
	Bytes     []byte
	Metadata  interface{}

	defaultPrevented   bool
	propagationStopped bool
}

// NewPasteEvent creates a new PasteEvent.
func NewPasteEvent(bytes []byte, metadata interface{}) *PasteEvent {
	return &PasteEvent{
		Type:     "paste",
		Bytes:    bytes,
		Metadata: metadata,
	}
}

// DefaultPrevented returns whether default was prevented.
func (p *PasteEvent) DefaultPrevented() bool { return p.defaultPrevented }

// PropagationStopped returns whether propagation was stopped.
func (p *PasteEvent) PropagationStopped() bool { return p.propagationStopped }

// PreventDefault marks the event as default-prevented.
func (p *PasteEvent) PreventDefault() { p.defaultPrevented = true }

// StopPropagation stops event propagation.
func (p *PasteEvent) StopPropagation() { p.propagationStopped = true }

// KeyHandlerEventMap maps event names to their argument types.
// In Go we use interface{} for the event data.
const (
	KeyHandlerEventKeyPress   = "keypress"
	KeyHandlerEventKeyRelease = "keyrelease"
	KeyHandlerEventPaste      = "paste"
)

// KeyHandler handles keyboard input events and emits them to listeners.
type KeyHandler struct {
	emitter *native.EventEmitter
}

// NewKeyHandler creates a new KeyHandler.
func NewKeyHandler() *KeyHandler {
	return &KeyHandler{
		emitter: native.NewEventEmitter(),
	}
}

// ProcessParsedKey processes a parsed key and emits keypress/keyrelease events.
func (h *KeyHandler) ProcessParsedKey(parsedKey ParsedKey) bool {
	event := NewKeyEvent(parsedKey)

	switch parsedKey.EventType() {
	case KeyEventTypePress:
		h.emitter.Emit(KeyHandlerEventKeyPress, event)
	case KeyEventTypeRelease:
		h.emitter.Emit(KeyHandlerEventKeyRelease, event)
	default:
		h.emitter.Emit(KeyHandlerEventKeyPress, event)
	}

	return true
}

// ProcessPaste processes a paste event and emits it to listeners.
func (h *KeyHandler) ProcessPaste(bytes []byte, metadata interface{}) {
	event := NewPasteEvent(bytes, metadata)
	h.emitter.Emit(KeyHandlerEventPaste, event)
}

// On registers a listener for an event.
func (h *KeyHandler) On(event string, listener func(interface{})) {
	h.emitter.On(event, listener)
}

// Off removes all listeners for an event.
func (h *KeyHandler) Off(event string) {
	h.emitter.Off(event)
}

// Emit emits an event to all listeners.
func (h *KeyHandler) Emit(event string, data interface{}) {
	h.emitter.Emit(event, data)
}

// ListenerCount returns the number of listeners for an event.
func (h *KeyHandler) ListenerCount(event string) int {
	return h.emitter.ListenerCount(event)
}

// InternalKeyHandler is used internally by the renderer to ensure global handlers
// can preventDefault before renderable handlers process events.
type InternalKeyHandler struct {
	*KeyHandler

	renderableHandlers map[string][]func(interface{})
	mu                 sync.RWMutex
}

// NewInternalKeyHandler creates a new InternalKeyHandler.
func NewInternalKeyHandler() *InternalKeyHandler {
	return &InternalKeyHandler{
		KeyHandler:         NewKeyHandler(),
		renderableHandlers: make(map[string][]func(interface{})),
	}
}

// Emit emits an event with priority: global handlers first, then renderable handlers.
func (h *InternalKeyHandler) Emit(event string, args ...interface{}) bool {
	return h.emitWithPriority(event, args...)
}

func (h *InternalKeyHandler) emitWithPriority(event string, args ...interface{}) bool {
	hasGlobalListeners := false

	// Emit to global handlers (via the embedded KeyHandler's emitter)
	globalListeners := h.emitter.ListenerCount(event)
	if globalListeners > 0 {
		hasGlobalListeners = true

		// We need to manually iterate to check for stopPropagation between handlers.
		// Since native.EventEmitter doesn't expose the listener list, we emit directly.
		if len(args) > 0 {
			h.emitter.Emit(event, args[0])
		} else {
			h.emitter.Emit(event, nil)
		}

		// Check if propagation was stopped after this handler
		if event == KeyHandlerEventKeyPress || event == KeyHandlerEventKeyRelease || event == KeyHandlerEventPaste {
			if len(args) > 0 {
				if keyEvent, ok := args[0].(*KeyEvent); ok && keyEvent.PropagationStopped() {
					return hasGlobalListeners
				}
				if pasteEvent, ok := args[0].(*PasteEvent); ok && pasteEvent.PropagationStopped() {
					return hasGlobalListeners
				}
			}
		}
	}

	// Emit to renderable handlers
	h.mu.RLock()
	renderableSet := h.renderableHandlers[event]
	renderableHandlers := make([]func(interface{}), len(renderableSet))
	copy(renderableHandlers, renderableSet)
	h.mu.RUnlock()

	hasRenderableListeners := false

	if len(renderableHandlers) > 0 {
		hasRenderableListeners = true

		if event == KeyHandlerEventKeyPress || event == KeyHandlerEventKeyRelease || event == KeyHandlerEventPaste {
			if len(args) > 0 {
				if keyEvent, ok := args[0].(*KeyEvent); ok {
					if keyEvent.DefaultPrevented() || keyEvent.PropagationStopped() {
						return hasGlobalListeners || hasRenderableListeners
					}
				}
				if pasteEvent, ok := args[0].(*PasteEvent); ok {
					if pasteEvent.DefaultPrevented() || pasteEvent.PropagationStopped() {
						return hasGlobalListeners || hasRenderableListeners
					}
				}
			}
		}

		for _, handler := range renderableHandlers {
			if len(args) > 0 {
				handler(args[0])
			} else {
				handler(nil)
			}

			if event == KeyHandlerEventKeyPress || event == KeyHandlerEventKeyRelease || event == KeyHandlerEventPaste {
				if len(args) > 0 {
					if keyEvent, ok := args[0].(*KeyEvent); ok {
						if keyEvent.PropagationStopped() {
							return hasGlobalListeners || hasRenderableListeners
						}
					}
					if pasteEvent, ok := args[0].(*PasteEvent); ok {
						if pasteEvent.PropagationStopped() {
							return hasGlobalListeners || hasRenderableListeners
						}
					}
				}
			}
		}
	}

	return hasGlobalListeners || hasRenderableListeners
}

// OnInternal registers a renderable handler for an event.
func (h *InternalKeyHandler) OnInternal(event string, handler func(interface{})) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.renderableHandlers[event] = append(h.renderableHandlers[event], handler)
}

// OffInternal removes a renderable handler for an event.
func (h *InternalKeyHandler) OffInternal(event string, handler func(interface{})) {
	h.mu.Lock()
	defer h.mu.Unlock()
	handlers := h.renderableHandlers[event]
	for i, h2 := range handlers {
		if fmt.Sprintf("%p", h2) == fmt.Sprintf("%p", handler) {
			h.renderableHandlers[event] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}
}

var _ = fmt.Sprintf