package native

import (
	"sync"
)

// Scene represents a scene graph for rendering
type Scene struct {
	mu     sync.RWMutex
	id     uint64
	nodes  map[uint64]*SceneNode
	root   *SceneNode
	nextID uint64
}

// NewScene creates a new Scene
func NewScene(id uint64) *Scene {
	return &Scene{
		id:    id,
		nodes: make(map[uint64]*SceneNode),
	}
}

// ID returns the scene ID
func (s *Scene) ID() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.id
}

// CreateNode creates a new scene node
func (s *Scene) CreateNode(kind SceneNodeKind) *SceneNode {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	node := &SceneNode{
		id:   s.nextID,
		kind: kind,
	}
	s.nodes[node.id] = node
	if s.root == nil {
		s.root = node
	}
	return node
}

// DestroyNode destroys a scene node
func (s *Scene) DestroyNode(id uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	node, ok := s.nodes[id]
	if !ok {
		return nil
	}
	if node == s.root {
		s.root = nil
	}
	delete(s.nodes, id)
	return nil
}

// GetNode returns a scene node by ID
func (s *Scene) GetNode(id uint64) *SceneNode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.nodes[id]
}

// Root returns the root scene node
func (s *Scene) Root() *SceneNode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.root
}

// AddChild adds a child node to a parent
func (s *Scene) AddChild(parentID, childID uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	parent, ok := s.nodes[parentID]
	if !ok {
		return nil
	}
	child, ok := s.nodes[childID]
	if !ok {
		return nil
	}
	child.parent = parent
	parent.children = append(parent.children, child)
	return nil
}
