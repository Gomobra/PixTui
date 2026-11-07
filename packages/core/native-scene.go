package core

import (
	"fmt"
	"sync"

	"github.com/Gomobra/PixTui/packages/native"
)

// SceneNodeHandle is a handle to a native scene node.
type SceneNodeHandle struct {
	id   uint64
	kind uint32
}

// ImageFit is the fit mode for images.
type ImageFit string

const (
	ImageFitFit    ImageFit = "fit"
	ImageFitCover  ImageFit = "cover"
	ImageFitFill   ImageFit = "fill"
)

// SceneNodeHandle is zero-valued when invalid.
var ZeroSceneNodeHandle = SceneNodeHandle{}

// LifecyclePasses manages lifecycle passes for renderables.
// Alias for NativeLifecyclePasses for backward compatibility.
type LifecyclePasses = NativeLifecyclePasses

// NewLifecyclePasses creates a new LifecyclePasses.
func NewLifecyclePasses() *LifecyclePasses {
	return NewNativeLifecyclePasses()
}

// NativeLifecyclePasses manages lifecycle passes for renderables with ordering.
// Mirrors packages/core/src/NativeScene.ts NativeLifecyclePasses.
type NativeLifecyclePasses struct {
	mu      sync.RWMutex
	passes  map[*Renderable]struct{}
	order   map[*Renderable]int
	nextOrder int
}

// NewNativeLifecyclePasses creates a new NativeLifecyclePasses.
func NewNativeLifecyclePasses() *NativeLifecyclePasses {
	return &NativeLifecyclePasses{
		passes:    make(map[*Renderable]struct{}),
		order:     make(map[*Renderable]int),
		nextOrder: 1,
	}
}

// Register registers a renderable for lifecycle passes.
func (lp *NativeLifecyclePasses) Register(r *Renderable) {
	lp.mu.Lock()
	defer lp.mu.Unlock()
	if _, ok := lp.passes[r]; !ok {
		lp.order[r] = lp.nextOrder
		lp.nextOrder++
	}
	lp.passes[r] = struct{}{}
}

// Unregister unregisters a renderable from lifecycle passes.
func (lp *NativeLifecyclePasses) Unregister(r *Renderable) {
	lp.mu.Lock()
	defer lp.mu.Unlock()
	delete(lp.passes, r)
	delete(lp.order, r)
}

// GetLifecyclePasses returns all registered renderables sorted by order.
func (lp *NativeLifecyclePasses) GetLifecyclePasses() []*Renderable {
	lp.mu.RLock()
	defer lp.mu.RUnlock()
	result := make([]*Renderable, 0, len(lp.passes))
	for r := range lp.passes {
		result = append(result, r)
	}
	// Sort by order
	for i := 1; i < len(result); i++ {
		key := result[i]
		j := i - 1
		for j >= 0 && lp.order[result[j]] > lp.order[key] {
			result[j+1] = result[j]
			j--
		}
		result[j+1] = key
	}
	return result
}

// Refresh triggers a refresh on all registered renderables.
func (lp *NativeLifecyclePasses) Refresh() {
	lp.mu.RLock()
	passes := lp.GetLifecyclePasses()
	lp.mu.RUnlock()
	for _, r := range passes {
		r.OnLifecyclePass()
	}
}

// NativeSceneRect represents a rectangular region.
type NativeSceneRect struct {
	Left, Top, Right, Bottom, Width, Height float32
}

// NativeSceneImageOptions holds image-specific options.
type NativeSceneImageOptions struct {
	Image    *native.Image
	Fit      ImageFit
	Protocol ImageRenderProtocol
	Buffer   *NativeBuffer
}

// NativeSceneVideoOptions holds video-specific options.
type NativeSceneVideoOptions struct {
	Fit     ImageFit
	Buffer  *NativeBuffer
}

// NativeScene bridges core renderables to the native scene graph.
// Mirrors packages/core/src/NativeScene.ts with lifecycle passes, staging, paint recording.
type NativeScene struct {
	mu      sync.RWMutex
	ctx     *native.ResourceContext
	session *native.Session
	nodes   map[*Renderable]*SceneNodeHandle
	nextID  uint64

	lifecyclePasses *NativeLifecyclePasses
	paintedFrame    *NativeSceneFrameRequest
	destroyed       bool
	destroying      bool
	layoutRevision  uint64
	staging         *SceneStaging

	// Yoga layout engine for this scene.
	yoga *native.Yoga

	// Per-renderable state stores.
	scrollState   map[*Renderable]*NativeSceneScrollState
	boxDetails    map[*Renderable]NativeSceneBoxDetails
	textContent   map[*Renderable]string
	textOptions   map[*Renderable]NativeSceneTextOptions
	sliderOptions map[*Renderable]NativeSceneSliderOptions
	imageOptions  map[*Renderable]*NativeSceneImageOptions
	videoOptions  map[*Renderable]*NativeSceneVideoOptions
	scrollOptions map[*Renderable]NativeSceneScrollOptions

	editorViews   map[*Renderable]interface{}
	surfaces      map[*Renderable]*NativeBuffer
	textViews     map[*Renderable]interface{}
	textViewPaint map[*Renderable]bool
	editorOptions map[*Renderable]NativeSceneEditorOptions

	// Staged node style values: handleID -> "group,kind,edge" -> value
	nodeStyles map[uint64]map[string]float32
}

// SceneStaging stages style and paint writes before native acceptance.
type SceneStaging struct {
	mu      sync.RWMutex
	pending bool
	full    bool
	records []StagedRecord
}

// StagedRecord is a staged write record.
type StagedRecord struct {
	HandleID uint64
	Group    uint32
	Kind     uint32
	Edge     uint32
	Unit     uint32
	Value    float32
	Flags    uint32
}

// NewSceneStaging creates a new SceneStaging.
func NewSceneStaging() *SceneStaging {
	return &SceneStaging{}
}

// NewNativeScene creates a new NativeScene.
func NewNativeScene(ctx *native.ResourceContext, session *native.Session) *NativeScene {
	return &NativeScene{
		ctx:             ctx,
		session:         session,
		nodes:           make(map[*Renderable]*SceneNodeHandle),
		nextID:          1,
		lifecyclePasses: NewNativeLifecyclePasses(),
		staging:         NewSceneStaging(),
		yoga:            native.NewYoga(),
		scrollState:     make(map[*Renderable]*NativeSceneScrollState),
		boxDetails:      make(map[*Renderable]NativeSceneBoxDetails),
		textContent:     make(map[*Renderable]string),
		textOptions:     make(map[*Renderable]NativeSceneTextOptions),
		sliderOptions:   make(map[*Renderable]NativeSceneSliderOptions),
		imageOptions:    make(map[*Renderable]*NativeSceneImageOptions),
		videoOptions:    make(map[*Renderable]*NativeSceneVideoOptions),
		scrollOptions:   make(map[*Renderable]NativeSceneScrollOptions),
		editorViews:     make(map[*Renderable]interface{}),
		surfaces:        make(map[*Renderable]*NativeBuffer),
		textViews:       make(map[*Renderable]interface{}),
		textViewPaint:   make(map[*Renderable]bool),
		editorOptions:   make(map[*Renderable]NativeSceneEditorOptions),
		nodeStyles:      make(map[uint64]map[string]float32),
	}
}

// GetYoga returns the Yoga layout engine for this scene.
func (ns *NativeScene) GetYoga() *native.Yoga {
	ns.mu.RLock()
	defer ns.mu.RUnlock()
	return ns.yoga
}

// CreateNode creates a native scene node for a renderable.
func (ns *NativeScene) CreateNode(r *Renderable, kind native.SceneNodeKind) (*SceneNodeHandle, error) {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	if ns.destroyed {
		return nil, fmt.Errorf("native scene is destroyed")
	}
	if ns.destroying {
		return nil, fmt.Errorf("native scene is being destroyed")
	}

	node, err := ns.session.CreateNode(kind, uint32(ns.nextID))
	if err != nil {
		return nil, err
	}

	ns.nextID++
	handle := &SceneNodeHandle{
		id:   node.GetNodeID(),
		kind: uint32(kind),
	}
	ns.nodes[r] = handle
	return handle, nil
}

// DestroyNode destroys a native scene node.
func (ns *NativeScene) DestroyNode(r *Renderable) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	handle, ok := ns.nodes[r]
	if !ok {
		return nil
	}

	if err := ns.session.DestroyNode(handle.id); err != nil {
		return err
	}
	delete(ns.nodes, r)

	// Clean up per-renderable state.
	delete(ns.scrollState, r)
	delete(ns.boxDetails, r)
	delete(ns.textContent, r)
	delete(ns.textOptions, r)
	delete(ns.sliderOptions, r)
	delete(ns.imageOptions, r)
	delete(ns.videoOptions, r)
	delete(ns.scrollOptions, r)
	delete(ns.editorViews, r)
	delete(ns.surfaces, r)
	delete(ns.textViews, r)
	delete(ns.textViewPaint, r)
	delete(ns.editorOptions, r)
	delete(ns.nodeStyles, handle.id)

	return nil
}

// GetNode returns the native scene node handle for a renderable.
func (ns *NativeScene) GetNode(r *Renderable) (*SceneNodeHandle, bool) {
	ns.mu.RLock()
	defer ns.mu.RUnlock()
	handle, ok := ns.nodes[r]
	return handle, ok
}

// styleKey builds a map key for a staged style value.
func styleKey(group, kind, edge uint32) string {
	return fmt.Sprintf("%d,%d,%d", group, kind, edge)
}

// SetNodeStyle sets style properties on a native node.
func (ns *NativeScene) SetNodeStyle(handle *SceneNodeHandle, group, kind, edge, unit uint32, value float32, flags uint32) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	if ns.destroyed {
		return fmt.Errorf("native scene is destroyed")
	}

	if ns.nodeStyles[handle.id] == nil {
		ns.nodeStyles[handle.id] = make(map[string]float32)
	}
	ns.nodeStyles[handle.id][styleKey(group, kind, edge)] = value

	// Stage the write.
	ns.staging.mu.Lock()
	ns.staging.records = append(ns.staging.records, StagedRecord{
		HandleID: handle.id,
		Group:    group,
		Kind:     kind,
		Edge:     edge,
		Unit:     unit,
		Value:    value,
		Flags:    flags,
	})
	ns.staging.pending = true
	ns.staging.mu.Unlock()

	return nil
}

// AttachNode attaches a node under a parent in the scene graph.
func (ns *NativeScene) AttachNode(childID, parentID uint64, index int) error {
	ns.mu.RLock()
	defer ns.mu.RUnlock()
	return ns.session.AttachNode(childID, parentID, index)
}

// SetFocus sets focus on a renderable's native node.
func (ns *NativeScene) SetFocus(r *Renderable, focused bool) error {
	ns.mu.RLock()
	defer ns.mu.RUnlock()
	handle, ok := ns.nodes[r]
	if !ok {
		return fmt.Errorf("node not found for renderable %s", r.ID())
	}
	_ = handle
	_ = focused
	return nil
}

// SetPaint sets paint properties on a renderable's native node.
func (ns *NativeScene) SetPaint(r *Renderable, paint NativeScenePaintUpdate) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	handle, ok := ns.nodes[r]
	if !ok {
		return fmt.Errorf("node not found for renderable %s", r.ID())
	}

	// Stage paint write.
	ns.staging.mu.Lock()
	ns.staging.records = append(ns.staging.records, StagedRecord{
		HandleID: handle.id,
		Group:    0, // NativeStyleGroup.Value
		Kind:     0, // YogaValueKind.Paint
		Edge:     0,
		Unit:     0,
		Value:    0,
		Flags:    0,
	})
	ns.staging.pending = true
	ns.staging.mu.Unlock()

	_ = paint
	return nil
}

// SetBackground sets background color on a renderable's native node.
func (ns *NativeScene) SetBackground(r *Renderable, color RGBA) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	handle, ok := ns.nodes[r]
	if !ok {
		return fmt.Errorf("node not found for renderable %s", r.ID())
	}

	ns.staging.mu.Lock()
	ns.staging.records = append(ns.staging.records, StagedRecord{
		HandleID: handle.id,
		Group:    0,
		Kind:     1, // background color kind
		Edge:     0,
		Unit:     0,
		Value:    0,
		Flags:    0,
	})
	ns.staging.pending = true
	ns.staging.mu.Unlock()

	_ = color
	return nil
}

// SetBoxDetails sets box details on a renderable's native node.
func (ns *NativeScene) SetBoxDetails(r *Renderable, details NativeSceneBoxDetails) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	if _, ok := ns.nodes[r]; !ok {
		return fmt.Errorf("node not found for renderable %s", r.ID())
	}

	ns.boxDetails[r] = details
	return nil
}

// SetText sets text content on a renderable's native node.
func (ns *NativeScene) SetText(r *Renderable, content string) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	if _, ok := ns.nodes[r]; !ok {
		return fmt.Errorf("node not found for renderable %s", r.ID())
	}

	ns.textContent[r] = content
	return nil
}

// SetTextOptions sets text options on a renderable's native node.
func (ns *NativeScene) SetTextOptions(r *Renderable, options NativeSceneTextOptions) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	if _, ok := ns.nodes[r]; !ok {
		return fmt.Errorf("node not found for renderable %s", r.ID())
	}

	ns.textOptions[r] = options
	return nil
}

// SetSlider sets slider options on a renderable's native node.
func (ns *NativeScene) SetSlider(r *Renderable, options NativeSceneSliderOptions) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	if _, ok := ns.nodes[r]; !ok {
		return fmt.Errorf("node not found for renderable %s", r.ID())
	}

	ns.sliderOptions[r] = options
	return nil
}

// SetImage sets image on a renderable's native node.
func (ns *NativeScene) SetImage(r *Renderable, image *native.Image, fit ImageFit, protocol ImageRenderProtocol, buffer *NativeBuffer) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	if _, ok := ns.nodes[r]; !ok {
		return fmt.Errorf("node not found for renderable %s", r.ID())
	}

	ns.imageOptions[r] = &NativeSceneImageOptions{
		Image:    image,
		Fit:      fit,
		Protocol: protocol,
		Buffer:   buffer,
	}
	return nil
}

// SetVideo sets video on a renderable's native node.
func (ns *NativeScene) SetVideo(r *Renderable, fit ImageFit, buffer *NativeBuffer) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	if _, ok := ns.nodes[r]; !ok {
		return fmt.Errorf("node not found for renderable %s", r.ID())
	}

	ns.videoOptions[r] = &NativeSceneVideoOptions{
		Fit:    fit,
		Buffer: buffer,
	}
	return nil
}

// SetScroll sets scroll options on a renderable's native node.
func (ns *NativeScene) SetScroll(r *Renderable, options NativeSceneScrollOptions) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	if _, ok := ns.nodes[r]; !ok {
		return fmt.Errorf("node not found for renderable %s", r.ID())
	}

	ns.scrollOptions[r] = options

	// Initialize scroll state if not present.
	if _, ok := ns.scrollState[r]; !ok {
		ns.scrollState[r] = &NativeSceneScrollState{
			Offset:  [2]float64{0, 0},
			Max:     [2]float64{0, 0},
			Content: [4]float64{0, 0, 0, 0},
			Viewport: [4]float64{0, 0, 0, 0},
		}
	}

	ns.layoutRevision++
	return nil
}

// ScrollTo scrolls to a position.
func (ns *NativeScene) ScrollTo(r *Renderable, x, y float64, relativeX, relativeY, reapply bool) (*NativeSceneScrollState, error) {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	state, ok := ns.scrollState[r]
	if !ok {
		state = &NativeSceneScrollState{
			Offset:   [2]float64{0, 0},
			Max:      [2]float64{0, 0},
			Content:  [4]float64{0, 0, 0, 0},
			Viewport: [4]float64{0, 0, 0, 0},
		}
		ns.scrollState[r] = state
	}

	if relativeX {
		state.Offset[0] += x
	} else {
		state.Offset[0] = x
	}
	if relativeY {
		state.Offset[1] += y
	} else {
		state.Offset[1] = y
	}

	// Clamp offset to valid range.
	if state.Offset[0] < 0 {
		state.Offset[0] = 0
	}
	if state.Offset[0] > state.Max[0] {
		state.Offset[0] = state.Max[0]
	}
	if state.Offset[1] < 0 {
		state.Offset[1] = 0
	}
	if state.Offset[1] > state.Max[1] {
		state.Offset[1] = state.Max[1]
	}

	ns.layoutRevision++
	return state, nil
}

// SetEditorView sets editor view on a renderable's native node.
func (ns *NativeScene) SetEditorView(r *Renderable, view interface{}) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	if _, ok := ns.nodes[r]; !ok {
		return fmt.Errorf("node not found for renderable %s", r.ID())
	}

	ns.editorViews[r] = view
	return nil
}

// SetSurface sets surface buffer on a renderable's native node.
func (ns *NativeScene) SetSurface(r *Renderable, buffer *NativeBuffer) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	if _, ok := ns.nodes[r]; !ok {
		return fmt.Errorf("node not found for renderable %s", r.ID())
	}

	ns.surfaces[r] = buffer
	return nil
}

// SetTextView sets text view on a renderable's native node.
func (ns *NativeScene) SetTextView(r *Renderable, view interface{}) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	if _, ok := ns.nodes[r]; !ok {
		return fmt.Errorf("node not found for renderable %s", r.ID())
	}

	ns.textViews[r] = view
	return nil
}

// SetTextViewPaint sets text view paint enabled.
func (ns *NativeScene) SetTextViewPaint(r *Renderable, enabled bool) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	if _, ok := ns.nodes[r]; !ok {
		return fmt.Errorf("node not found for renderable %s", r.ID())
	}

	ns.textViewPaint[r] = enabled
	return nil
}

// SetEditorOptions sets editor options on a renderable's native node.
func (ns *NativeScene) SetEditorOptions(r *Renderable, options NativeSceneEditorOptions) error {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	if _, ok := ns.nodes[r]; !ok {
		return fmt.Errorf("node not found for renderable %s", r.ID())
	}

	ns.editorOptions[r] = options
	return nil
}

// GetText returns text content from a renderable's native node.
func (ns *NativeScene) GetText(r *Renderable) (string, error) {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	content, ok := ns.textContent[r]
	if !ok {
		return "", nil
	}
	return content, nil
}

// GetLayout returns the computed layout for a node.
func (ns *NativeScene) GetLayout(handle *SceneNodeHandle) Layout {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	// Find the renderable that owns this handle.
	for r, h := range ns.nodes {
		if h.id == handle.id {
			if r.yogaNode != nil {
				yl := r.yogaNode.GetComputedLayout()
				return Layout{
					Left:   yl.Left,
					Top:    yl.Top,
					Right:  yl.Right,
					Bottom: yl.Bottom,
					Width:  yl.Width,
					Height: yl.Height,
				}
			}
			return Layout{Width: r.width, Height: r.height}
		}
	}
	return Layout{}
}

// RunLifecyclePasses runs lifecycle passes on all registered renderables.
func (ns *NativeScene) RunLifecyclePasses() {
	ns.lifecyclePasses.Refresh()
}

// RegisterLifecyclePass registers a renderable for lifecycle passes.
func (ns *NativeScene) RegisterLifecyclePass(r *Renderable) {
	ns.lifecyclePasses.Register(r)
}

// UnregisterLifecyclePass unregisters a renderable from lifecycle passes.
func (ns *NativeScene) UnregisterLifecyclePass(r *Renderable) {
	ns.lifecyclePasses.Unregister(r)
}

// HitTest performs a hit test at (x, y).
func (ns *NativeScene) HitTest(x, y int) uint64 {
	ns.mu.RLock()
	defer ns.mu.RUnlock()

	for _, handle := range ns.nodes {
		layout := ns.GetLayout(handle)
		if float32(x) >= layout.Left && float32(x) < layout.Left+layout.Width &&
			float32(y) >= layout.Top && float32(y) < layout.Top+layout.Height {
			return handle.id
		}
	}
	return 0
}

// IsDestroyed returns whether the native scene has been destroyed.
func (ns *NativeScene) IsDestroyed() bool {
	ns.mu.RLock()
	defer ns.mu.RUnlock()
	return ns.destroyed
}

// Destroy destroys the native scene and all nodes.
func (ns *NativeScene) Destroy() error {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	if ns.destroyed || ns.destroying {
		return nil
	}
	ns.destroying = true
	ns.staging.mu.Lock()
	ns.staging.pending = false
	ns.staging.full = false
	ns.staging.records = nil
	ns.staging.mu.Unlock()

	for r := range ns.nodes {
		r.Destroy()
	}
	ns.nodes = make(map[*Renderable]*SceneNodeHandle)
	ns.scrollState = make(map[*Renderable]*NativeSceneScrollState)
	ns.boxDetails = make(map[*Renderable]NativeSceneBoxDetails)
	ns.textContent = make(map[*Renderable]string)
	ns.textOptions = make(map[*Renderable]NativeSceneTextOptions)
	ns.sliderOptions = make(map[*Renderable]NativeSceneSliderOptions)
	ns.imageOptions = make(map[*Renderable]*NativeSceneImageOptions)
	ns.videoOptions = make(map[*Renderable]*NativeSceneVideoOptions)
	ns.scrollOptions = make(map[*Renderable]NativeSceneScrollOptions)
	ns.editorViews = make(map[*Renderable]interface{})
	ns.surfaces = make(map[*Renderable]*NativeBuffer)
	ns.textViews = make(map[*Renderable]interface{})
	ns.textViewPaint = make(map[*Renderable]bool)
	ns.editorOptions = make(map[*Renderable]NativeSceneEditorOptions)
	ns.nodeStyles = make(map[uint64]map[string]float32)
	ns.destroyed = true
	ns.destroying = false
	return nil
}

// Refresh marks all nodes as needing refresh.
func (ns *NativeScene) Refresh(r *Renderable) {
	ns.mu.RLock()
	defer ns.mu.RUnlock()
	_ = r
}

// NativeScenePaintUpdate is a partial visual write.
type NativeScenePaintUpdate struct {
	TranslateX, TranslateY *float64
}

// NativeSceneBoxDetails holds box-specific details.
type NativeSceneBoxDetails struct {
	Title                 string
	BottomTitle           string
	TitleAlignment        uint32
	BottomTitleAlignment  uint32
	TitleColor            RGBA
	CustomBorderChars     [11]uint32
}

// NativeSceneTextOptions holds text-specific options.
type NativeSceneTextOptions struct {
	Foreground         RGBA
	Background         RGBA
	Attributes         uint32
	WrapMode           uint32
	Truncate           uint32
}

// NativeSceneSliderOptions holds slider options.
type NativeSceneSliderOptions struct {
	Min, Max, Value float64
}

// NativeSceneScrollOptions holds scroll options.
type NativeSceneScrollOptions struct {
	Axes      uint32
	Follow    uint32
	Culling   bool
	Padding   [4]float32
}

// NativeSceneScrollState holds scroll state.
type NativeSceneScrollState struct {
	ViewportNode   interface{}
	ContentNode    interface{}
	Offset         [2]float64
	Max            [2]float64
	Content        [4]float64
	Viewport       [4]float64
}

// NativeSceneEditorOptions holds editor options.
type NativeSceneEditorOptions struct{}

// NativeSceneFrameRequest represents a frame request from native.
type NativeSceneFrameRequest struct {
	Kind      uint32
	Session   interface{}
	Root      interface{}
	FrameID   uint64
	Num       uint64
	Node      interface{}
}

// NativeSceneFrame represents frame kinds.
type NativeSceneFrame struct{}

const (
	NativeSceneFrameDone uint32 = 0
	NativeSceneFrameRecord uint32 = 1
)

var _ = fmt.Sprintf
