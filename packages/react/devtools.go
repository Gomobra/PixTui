package react

import (
	"fmt"
	"sync"
)

// DevTools provides read-only inspection of the React renderer's fiber tree,
// mirroring the react-devtools integration in the TypeScript package. It can
// be attached to a ReactRenderer to expose the current fiber tree, component
// names, hook counts, and host element types for debugging and testing.
type DevTools struct {
	mu       sync.Mutex
	renderer *ReactRenderer
	enabled  bool
}

// NewDevTools creates a DevTools instance attached to the given renderer.
// The tools are disabled by default; call Enable to begin inspecting.
func NewDevTools(renderer *ReactRenderer) *DevTools {
	return &DevTools{
		renderer: renderer,
	}
}

// Enable enables the DevTools inspection.
func (d *DevTools) Enable() {
	d.mu.Lock()
	d.enabled = true
	d.mu.Unlock()
}

// Disable disables the DevTools inspection.
func (d *DevTools) Disable() {
	d.mu.Lock()
	d.enabled = false
	d.mu.Unlock()
}

// Enabled reports whether DevTools is currently enabled.
func (d *DevTools) Enabled() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.enabled
}

// FiberNode is a serializable representation of a fiber used by DevTools.
type FiberNode struct {
	Type        string               `json:"type"`
	Key         string               `json:"key,omitempty"`
	Host        bool                 `json:"host"`
	Component   string               `json:"component,omitempty"`
	PropCount   int                  `json:"propCount"`
	ChildCount  int                  `json:"childCount"`
	HookCount   int                  `json:"hookCount"`
	DomAttached bool                 `json:"domAttached"`
	Children    []*FiberNode         `json:"children,omitempty"`
}

// FiberTree returns a serializable snapshot of the current fiber tree. It
// returns nil when the renderer has no root fiber or DevTools is disabled.
func (d *DevTools) FiberTree() *FiberNode {
	d.mu.Lock()
	if !d.enabled {
		d.mu.Unlock()
		return nil
	}
	r := d.renderer
	d.mu.Unlock()

	r.mu.RLock()
	root := r.root
	r.mu.RUnlock()

	if root == nil {
		return nil
	}

	return d.fiberNode(root)
}

func (d *DevTools) fiberNode(f *fiber) *FiberNode {
	f.mu.Lock()
	node := &FiberNode{
		Key:         f.key,
		Host:        isHostElement(ReactElement{Type: f.typ}),
		PropCount:   len(f.props),
		ChildCount:  len(f.children),
		HookCount:   len(f.state) + len(f.effects) + len(f.memos) + len(f.cbs),
		DomAttached: f.domAttached,
	}
	switch typ := f.typ.(type) {
	case string:
		node.Type = typ
	case ComponentFunc:
		node.Component = componentName(typ)
		node.Type = "component"
	case *contextProviderType:
		node.Type = "context.Provider"
		node.Component = "Context.Provider"
	default:
		node.Type = fmt.Sprintf("%T", typ)
	}
	f.mu.Unlock()

	for _, c := range f.children {
		node.Children = append(node.Children, d.fiberNode(c))
	}
	return node
}

// Inspect returns a human-readable string summary of the fiber tree for
// debugging. It is a no-op (returns empty string) when DevTools is disabled.
func (d *DevTools) Inspect() string {
	tree := d.FiberTree()
	if tree == nil {
		return "<devtools disabled>"
	}
	return d.formatTree(tree, 0)
}

func (d *DevTools) formatTree(n *FiberNode, depth int) string {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}
	label := n.Type
	if n.Component != "" {
		label = n.Component
	}
	if n.Key != "" {
		label += " [key=" + n.Key + "]"
	}
	s := fmt.Sprintf("%s%s (props=%d children=%d hooks=%d attached=%v)\n",
		indent, label, n.PropCount, n.ChildCount, n.HookCount, n.DomAttached)
	for _, c := range n.Children {
		s += d.formatTree(c, depth+1)
	}
	return s
}
