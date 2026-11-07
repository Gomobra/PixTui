package solid

import (
	"sync"
)

// SlotMode determines how multiple slot entries are rendered.
type SlotMode int

const (
	// SlotModeAppend appends slot entries after the slot's children (default).
	SlotModeAppend SlotMode = iota
	// SlotModeSingleWinner renders only the first entry.
	SlotModeSingleWinner
	// SlotModeReplace renders only slot entries, hiding the slot's children
	// unless no entry produces output.
	SlotModeReplace
)

// String returns the slot mode name.
func (m SlotMode) String() string {
	switch m {
	case SlotModeSingleWinner:
		return "single_winner"
	case SlotModeReplace:
		return "replace"
	default:
		return "append"
	}
}

// SlotRegistry is a simplified registry of slot renderers. Plugins register
// renderers under a slot name; the Slot component resolves and renders them.
type SlotRegistry[T any] struct {
	mu      sync.RWMutex
	entries map[string][]SlotEntry[T]
}

// SlotEntry is a registered renderer for a slot.
type SlotEntry[T any] struct {
	ID      string
	Renderer func(ctx interface{}, props map[string]interface{}) T
}

// NewSlotRegistry creates a new empty slot registry.
func NewSlotRegistry[T any]() *SlotRegistry[T] {
	return &SlotRegistry[T]{
		entries: make(map[string][]SlotEntry[T]),
	}
}

// Register adds a renderer to the given slot name.
func (r *SlotRegistry[T]) Register(slotName string, id string, renderer func(ctx interface{}, props map[string]interface{}) T) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries[slotName] = append(r.entries[slotName], SlotEntry[T]{ID: id, Renderer: renderer})
}

// ResolveEntries returns all renderers registered for the given slot name.
func (r *SlotRegistry[T]) ResolveEntries(slotName string) []SlotEntry[T] {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]SlotEntry[T], 0, len(r.entries[slotName]))
	out = append(out, r.entries[slotName]...)
	return out
}

// SlotProps are the props accepted by the Slot component.
type SlotProps struct {
	Registry              *SlotRegistry[SolidComponent]
	Name                  string
	Mode                  SlotMode
	Children              []SolidComponent
	PluginFailurePlaceholder func(error) SolidComponent
}

// Slot renders registered slot entries for the given name, mirroring the
// Solid slot plugin from @pixtui/solid. Entries are rendered according to
// the slot mode: append (default) shows children followed by entries,
// single_winner shows only the first entry, replace shows only entries.
func Slot(props SlotProps) SolidComponent {
	entries := props.Registry.ResolveEntries(props.Name)

	renderEntry := func(entry SlotEntry[SolidComponent], fallback SolidComponent) SolidComponent {
		if entry.Renderer == nil {
			return fallback
		}
		rendered := entry.Renderer(nil, map[string]interface{}{})
		return rendered
	}

	if len(entries) == 0 {
		if len(props.Children) > 0 {
			return props.Children[0]
		}
		return SolidComponent{}
	}

	switch props.Mode {
	case SlotModeSingleWinner:
		winner := entries[0]
		rendered := renderEntry(winner, SolidComponent{})
		if rendered.Fn == nil && rendered.Type == "" && len(props.Children) > 0 {
			return props.Children[0]
		}
		return rendered
	case SlotModeReplace:
		var rendered []SolidComponent
		hasOutput := false
		for _, entry := range entries {
			r := renderEntry(entry, SolidComponent{})
			if r.Fn != nil || r.Type != "" {
				hasOutput = true
			}
			rendered = append(rendered, r)
		}
		if !hasOutput && len(props.Children) > 0 {
			return props.Children[0]
		}
		// Return a fragment-like wrapper holding all rendered entries.
		if len(rendered) == 1 {
			return rendered[0]
		}
		return SolidComponent{Type: "fragment", Props: map[string]interface{}{"children": rendered}}
	default: // SlotModeAppend
		if len(entries) == 0 && len(props.Children) > 0 {
			return props.Children[0]
		}
		all := make([]SolidComponent, 0, len(props.Children)+len(entries))
		all = append(all, props.Children...)
		for _, entry := range entries {
			all = append(all, renderEntry(entry, SolidComponent{}))
		}
		if len(all) == 1 {
			return all[0]
		}
		return SolidComponent{Type: "fragment", Props: map[string]interface{}{"children": all}}
	}
}

// createSlot returns a BoundSlot function that renders a Slot for the given
// registry and slot name, mirroring createSlot in the TypeScript plugin.
func createSlot(registry *SlotRegistry[SolidComponent], name string, mode SlotMode) func(children ...SolidComponent) SolidComponent {
	return func(children ...SolidComponent) SolidComponent {
		return Slot(SlotProps{
			Registry: registry,
			Name:     name,
			Mode:     mode,
			Children: children,
		})
	}
}

// createSolidSlotRegistry is the Solid-specific slot registry constructor.
// It is a no-op alias for NewSlotRegistry kept for API compatibility with
// the TypeScript package.
func createSolidSlotRegistry() *SlotRegistry[SolidComponent] {
	return NewSlotRegistry[SolidComponent]()
}

var _ = SlotMode(0) // ensure SlotMode is used
