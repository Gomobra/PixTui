package native

import (
	"fmt"
	"sync"
)

// Session represents a rendering session tied to a Context.
// Mirrors packages/native/src/session.zig.
type Session struct {
	mu               sync.RWMutex
	id               uint64
	context          *ResourceContext
	status           ContextStatus
	rootNode         *SceneNode
	nodes            map[uint64]*SceneNode
	renderer         *CliRenderer
	output           *OptimizedBuffer
	width            int
	height           int
	nextNodeID       uint64
	renderStatusFunc renderStatusFunc
	idleFunc         idleFunc
}

// NewSession creates a new Session
func NewSession(id uint64, context *ResourceContext) *Session {
	return &Session{
		id:      id,
		context: context,
		status:  ContextStatusActive,
		nodes:   make(map[uint64]*SceneNode),
		width:   80,
		height:  24,
	}
}

// ID returns the session ID
func (s *Session) ID() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.id
}

// Status returns the session status
func (s *Session) Status() ContextStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status
}

// Close closes the session and all associated nodes
func (s *Session) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.status == ContextStatusClosed {
		return nil
	}
	s.status = ContextStatusClosed
	for _, n := range s.nodes {
		n.parent = nil
		n.children = nil
	}
	s.nodes = make(map[uint64]*SceneNode)
	s.rootNode = nil
	return nil
}

// SetDimensions sets the terminal dimensions for this session
func (s *Session) SetDimensions(width, height int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.width = width
	s.height = height
	if s.output == nil || s.output.Width() != width || s.output.Height() != height {
		s.output = NewOptimizedBuffer(width, height)
	}
}

// Dimensions returns the current dimensions
func (s *Session) Dimensions() (int, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.width, s.height
}

// Output returns the session's output buffer
func (s *Session) Output() *OptimizedBuffer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.output
}

// CreateNode creates a new scene node (public wrapper).
func (s *Session) CreateNode(kind SceneNodeKind, num uint32) (*SceneNode, error) {
	return s.createNode(kind, num)
}

// createNode creates a new scene node
func (s *Session) createNode(kind SceneNodeKind, num uint32) (*SceneNode, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.status == ContextStatusClosed {
		return nil, fmt.Errorf("session %d is closed", s.id)
	}
	s.nextNodeID++
	node := &SceneNode{
		id:   s.nextNodeID,
		kind: kind,
		num:  num,
		paint: &PaintOptions{
			Opacity: 1.0,
		},
		style: &StyleValue{},
	}
	s.nodes[node.id] = node
	if kind == SceneNodeKindRoot {
		s.rootNode = node
	}
	return node, nil
}

// GetNode returns a node by ID (public wrapper).
func (s *Session) GetNode(id uint64) *SceneNode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.nodes[id]
}

// GetNodeID returns the node's ID.
func (n *SceneNode) GetNodeID() uint64 {
	return n.id
}

// GetNodeKind returns the node's kind.
func (n *SceneNode) GetNodeKind() SceneNodeKind {
	return n.kind
}

// DestroyNode destroys a node and detaches its children (public wrapper).
func (s *Session) DestroyNode(id uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	node, ok := s.nodes[id]
	if !ok {
		return fmt.Errorf("node %d not found", id)
	}
	delete(s.nodes, id)
	if node == s.rootNode {
		s.rootNode = nil
	}
	for _, child := range node.children {
		child.parent = nil
	}
	node.children = nil
	return nil
}

// AttachNode moves a node under a parent (public wrapper).
func (s *Session) AttachNode(nodeID, parentID uint64, index int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	node, ok := s.nodes[nodeID]
	if !ok {
		return fmt.Errorf("node %d not found", nodeID)
	}
	if node.parent != nil {
		node.parent.detachChild(nodeID)
	}
	if parentID == 0 {
		node.parent = nil
		return nil
	}
	parent, ok := s.nodes[parentID]
	if !ok {
		return fmt.Errorf("parent node %d not found", parentID)
	}
	node.parent = parent
	parent.attachChild(node, index)
	return nil
}

// getNode returns a node by ID (internal).
func (s *Session) getNode(id uint64) *SceneNode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.nodes[id]
}

// destroyNode destroys a node and detaches its children (internal).
func (s *Session) destroyNode(id uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	node, ok := s.nodes[id]
	if !ok {
		return fmt.Errorf("node %d not found", id)
	}
	delete(s.nodes, id)
	if node == s.rootNode {
		s.rootNode = nil
	}
	for _, child := range node.children {
		child.parent = nil
	}
	node.children = nil
	return nil
}

// attachNode moves a node under a parent (internal).
func (s *Session) attachNode(nodeID, parentID uint64, index int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	node, ok := s.nodes[nodeID]
	if !ok {
		return fmt.Errorf("node %d not found", nodeID)
	}
	if node.parent != nil {
		node.parent.detachChild(nodeID)
	}
	if parentID == 0 {
		node.parent = nil
		return nil
	}
	parent, ok := s.nodes[parentID]
	if !ok {
		return fmt.Errorf("parent node %d not found", parentID)
	}
	node.parent = parent
	parent.attachChild(node, index)
	return nil
}

// Renderer returns the session's renderer
func (s *Session) Renderer() *CliRenderer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.renderer
}

// SetRenderer attaches a renderer to the session
func (s *Session) SetRenderer(r *CliRenderer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.renderer = r
}

// SessionRenderStatus represents the status of a native render commit.
// Mirrors NativeSessionRenderStatus from the TypeScript reference.
type SessionRenderStatus int

const (
	// SessionRenderStatusPresented means the frame commits normally to output.
	SessionRenderStatusPresented SessionRenderStatus = iota
	// SessionRenderStatusSkipped means the frame is consumed without output.
	SessionRenderStatusSkipped
	// SessionRenderStatusCancelled means the frame is cancelled.
	SessionRenderStatusCancelled
)

// renderStatusFunc is an optional override that determines the status of each render commit.
type renderStatusFunc func() SessionRenderStatus

// idleFunc is an optional override for the session's idle wait.
type idleFunc func()

// SetRenderStatusFunc sets an override for the render status of each commit.
// When set to a status other than Presented, the frame is consumed without output.
// Returns a restore function that clears the override.
func (s *Session) SetRenderStatusFunc(next func() SessionRenderStatus) func() {
	s.mu.Lock()
	prev := s.renderStatusFunc
	s.renderStatusFunc = next
	s.mu.Unlock()
	return func() {
		s.mu.Lock()
		s.renderStatusFunc = prev
		s.mu.Unlock()
	}
}

// SetIdleFunc sets an override for the session's idle wait.
// When set, the idle call blocks until the override returns.
// Returns a restore function that clears the override.
func (s *Session) SetIdleFunc(idle func()) func() {
	s.mu.Lock()
	prev := s.idleFunc
	s.idleFunc = idle
	s.mu.Unlock()
	return func() {
		s.mu.Lock()
		s.idleFunc = prev
		s.mu.Unlock()
	}
}

// ShouldPresent returns whether the current render commit should be presented.
// If a render status func is set and returns a status other than Presented, the frame is consumed without output.
func (s *Session) ShouldPresent() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.renderStatusFunc == nil {
		return true
	}
	return s.renderStatusFunc() == SessionRenderStatusPresented
}

// IdleFunc returns the current idle override function, or nil if none is set.
func (s *Session) IdleFunc() idleFunc {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.idleFunc
}

// SceneNodeKind represents the kind of a scene node
type SceneNodeKind int

const (
	SceneNodeKindRoot     SceneNodeKind = 0
	SceneNodeKindBox      SceneNodeKind = 1
	SceneNodeKindText     SceneNodeKind = 2
	SceneNodeKindSlider   SceneNodeKind = 3
	SceneNodeKindArrow    SceneNodeKind = 4
	SceneNodeKindEditor   SceneNodeKind = 5
	SceneNodeKindCustom   SceneNodeKind = 6
	SceneNodeKindTextView SceneNodeKind = 7
	SceneNodeKindImage    SceneNodeKind = 8
	SceneNodeKindVideo    SceneNodeKind = 9
)

// SceneNode represents a node in the scene graph
type SceneNode struct {
	id            uint64
	kind          SceneNodeKind
	num           uint32
	parent        *SceneNode
	children      []*SceneNode
	x, y          float32
	width, height float32
	paint         *PaintOptions
	boxDetails    *BoxDetails
	text          string
	textOptions   *TextOptions
	style         *StyleValue
	scroll        *ScrollState
	yogaNode      *YogaNode
}

// attachChild adds a child to this node
func (n *SceneNode) attachChild(child *SceneNode, index int) {
	if index < 0 || index > len(n.children) {
		n.children = append(n.children, child)
	} else {
		n.children = append(n.children[:index], append([]*SceneNode{child}, n.children[index:]...)...)
	}
}

// detachChild removes a child by ID
func (n *SceneNode) detachChild(childID uint64) {
	for i, c := range n.children {
		if c.id == childID {
			n.children = append(n.children[:i], n.children[i+1:]...)
			return
		}
	}
}

// PaintOptions holds paint properties for a node
type PaintOptions struct {
	ZIndex                 int32
	Opacity                float32
	TranslateX, TranslateY float64
	BorderSides            uint32
	ShouldFill             bool
	Background             RGBA
	BorderColor            RGBA
	BorderStyle            uint32
	Focusable              bool
	FocusedBorderColor     RGBA
}

// BoxDetails holds box-specific details
type BoxDetails struct {
	Title                string
	BottomTitle          string
	TitleAlignment       uint32
	BottomTitleAlignment uint32
	TitleColor           RGBA
	CustomBorderChars    [11]uint32
}

// TextOptions holds text-specific options
type TextOptions struct {
	Foreground       RGBA
	Background       RGBA
	Attributes       uint32
	WrapMode         uint32
	Truncate         uint32
	FirstLineOffset  uint32
	ScrollX, ScrollY float64
	TabIndicator     uint32
	TabColorSet      uint32
	TabColor         RGBA
	TextAlign        uint32
}

// StyleValue represents a style value
type StyleValue struct {
	Unit  uint32
	Value float32
}

// ScrollState holds scroll state for scroll nodes
type ScrollState struct {
	ViewportNode *SceneNode
	ContentNode  *SceneNode
	Options      *ScrollOptions
	Arrows       [2][2]*ArrowOptions
	Stuck        bool
	Offset       [2]float64
	Max          [2]float64
	Content      [4]float64
	Viewport     [4]float64
	Bars         [2]*ScrollBarState
}

// ScrollOptions holds scroll options
type ScrollOptions struct {
	Axes    uint32
	Follow  uint32
	Culling bool
	Padding [4]float32
	Fills   [3]*RGBA
	Bars    [2]*ScrollBarOptions
}

// ScrollBarOptions holds scroll bar options
type ScrollBarOptions struct {
	Policy     uint32
	Overlay    bool
	Thickness  uint32
	Gap        uint32
	Arrows     bool
	Foreground RGBA
	Background RGBA
}

// ArrowOptions holds arrow options
type ArrowOptions struct {
	Direction  uint32
	Attributes uint32
	Foreground RGBA
	Background RGBA
	Text       string
}

// ScrollBarState holds the current state of a scroll bar
type ScrollBarState struct {
	Visible    bool
	Track      [4]float64
	StartArrow [4]float64
	EndArrow   [4]float64
	ThumbStart float64
	ThumbSize  float64
}

// Layout holds layout geometry
type Layout struct {
	Left, Top, Right, Bottom, Width, Height float32
	ScreenX, ScreenY                        float64
}

// NewSceneNode creates a new scene node
func NewSceneNode(kind SceneNodeKind, num uint32) *SceneNode {
	return &SceneNode{
		kind: kind,
		num:  num,
		paint: &PaintOptions{
			Opacity: 1.0,
		},
		style: &StyleValue{},
	}
}
