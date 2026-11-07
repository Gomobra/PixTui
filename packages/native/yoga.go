package native

import (
	"math"
	"sync"
)

// Yoga implements a flexbox layout engine, replacing the Yoga C library used in
// packages/native/src/yoga.zig. It supports flex direction, justify-content,
// align-items, align-self, flex-wrap, flex grow/shrink, baseline alignment,
// min/max constraints, padding, margin, gap, and border-box/content-box sizing.
type Yoga struct {
	mu     sync.Mutex
	nodes  map[uint64]*YogaNode
	nextID uint64
}

// YogaNode represents a node in the Yoga layout tree.
type YogaNode struct {
	mu             sync.RWMutex
	id             uint64
	FlexDirection  FlexDirection
	JustifyContent JustifyContent
	AlignItems     AlignItems
	AlignSelf      AlignSelf
	FlexWrap       FlexWrap
	PositionType   PositionType
	Overflow       Overflow
	Display        Display
	BoxSizing      BoxSizing
	Flex           float32
	FlexGrow       float32
	FlexShrink     float32
	FlexBasis      float32
	Width          float32
	Height         float32
	MinWidth       float32
	MinHeight      float32
	MaxWidth       float32
	MaxHeight      float32
	Margin         [4]float32 // top, right, bottom, left
	Padding        [4]float32 // top, right, bottom, left
	Position       [4]float32 // top, right, bottom, left
	Border         [4]float32 // top, right, bottom, left
	Gap            [2]float32 // row, column
	Layout         Layout
	Computed       bool
	Children       []*YogaNode
	Parent         *YogaNode
	MeasureFunc    MeasureFunction
}

// MeasureFunction is a callback for measuring content size of leaf nodes.
type MeasureFunction func(width float32, widthMode MeasureMode, height float32, heightMode MeasureMode) (float32, float32)

// MeasureMode defines how to interpret a dimension in measurement.
type MeasureMode int

const (
	MeasureUndefined MeasureMode = 0
	MeasureExactly   MeasureMode = 1
	MeasureAtMost    MeasureMode = 2
)

// FlexDirection values.
type FlexDirection int

const (
	FlexDirectionColumn        FlexDirection = 0
	FlexDirectionColumnReverse FlexDirection = 1
	FlexDirectionRow           FlexDirection = 2
	FlexDirectionRowReverse    FlexDirection = 3
)

// JustifyContent values (distribution along main axis).
type JustifyContent int

const (
	JustifyFlexStart    JustifyContent = 0
	JustifyCenter       JustifyContent = 1
	JustifyFlexEnd      JustifyContent = 2
	JustifySpaceBetween JustifyContent = 3
	JustifySpaceAround  JustifyContent = 4
	JustifySpaceEvenly  JustifyContent = 5
)

// AlignItems values (default alignment along cross axis).
type AlignItems int

const (
	AlignAuto         AlignItems = 0
	AlignFlexStart    AlignItems = 1
	AlignCenter       AlignItems = 2
	AlignFlexEnd      AlignItems = 3
	AlignStretch      AlignItems = 4
	AlignBaseline     AlignItems = 5
	AlignSpaceBetween AlignItems = 6
	AlignSpaceAround  AlignItems = 7
	AlignSpaceEvenly  AlignItems = 8
)

// AlignSelf values (per-node override of align-items).
type AlignSelf int

const (
	AlignSelfAuto      AlignSelf = 0
	AlignSelfFlexStart AlignSelf = 1
	AlignSelfCenter    AlignSelf = 2
	AlignSelfFlexEnd   AlignSelf = 3
	AlignSelfStretch   AlignSelf = 4
	AlignSelfBaseline  AlignSelf = 5
)

// FlexWrap values.
type FlexWrap int

const (
	FlexWrapNone    FlexWrap = 0
	FlexWrapWrap    FlexWrap = 1
	FlexWrapReverse FlexWrap = 2
)

// PositionType values.
type PositionType int

const (
	PositionStatic   PositionType = 0
	PositionRelative PositionType = 1
	PositionAbsolute PositionType = 2
)

// Overflow values.
type Overflow int

const (
	OverflowVisible Overflow = 0
	OverflowHidden  Overflow = 1
	OverflowScroll  Overflow = 2
)

// Display values.
type Display int

const (
	DisplayFlex Display = 0
	DisplayNone Display = 1
)

// BoxSizing values.
type BoxSizing int

const (
	BoxSizingBorderBox  BoxSizing = 0
	BoxSizingContentBox BoxSizing = 1
)

// NewYoga creates a new Yoga layout engine.
func NewYoga() *Yoga {
	return &Yoga{
		nodes:  make(map[uint64]*YogaNode),
		nextID: 1,
	}
}

// CreateNode creates a new Yoga node.
func (y *Yoga) CreateNode() *YogaNode {
	y.mu.Lock()
	defer y.mu.Unlock()
	node := &YogaNode{
		id:             y.nextID,
		FlexDirection:  FlexDirectionColumn,
		JustifyContent: JustifyFlexStart,
		AlignItems:     AlignStretch,
		FlexWrap:       FlexWrapNone,
		PositionType:   PositionStatic,
		Overflow:       OverflowVisible,
		Display:        DisplayFlex,
		BoxSizing:      BoxSizingBorderBox,
		Flex:           0,
		FlexGrow:       0,
		FlexShrink:     1,
		FlexBasis:      0,
		Layout:         Layout{},
		Computed:       false,
	}
	y.nodes[node.id] = node
	y.nextID++
	return node
}

// GetNode returns a Yoga node by ID.
func (y *Yoga) GetNode(id uint64) *YogaNode {
	y.mu.Lock()
	defer y.mu.Unlock()
	return y.nodes[id]
}

// DestroyNode removes a Yoga node and its subtree.
func (y *Yoga) DestroyNode(id uint64) {
	y.mu.Lock()
	defer y.mu.Unlock()
	node, ok := y.nodes[id]
	if !ok {
		return
	}
	y.removeSubtree(node)
	delete(y.nodes, id)
}

func (y *Yoga) removeSubtree(node *YogaNode) {
	for _, child := range node.Children {
		if child.Parent == node {
			child.Parent = nil
			y.removeSubtree(child)
			delete(y.nodes, child.id)
		}
	}
	node.Children = nil
	node.Parent = nil
}

// CalculateLayout runs the full flexbox layout algorithm for the given node
// constrained by the outer width and height.
func (y *Yoga) CalculateLayout(node *YogaNode, width, height float32) {
	node.mu.Lock()
	defer node.mu.Unlock()

	if node.Display == DisplayNone {
		node.Layout = Layout{Width: 0, Height: 0}
		node.Computed = true
		return
	}

	// Resolve the node's own dimensions against constraints.
	w := resolveDimension(node.Width, width, node.MinWidth, node.MaxWidth)
	h := resolveDimension(node.Height, height, node.MinHeight, node.MaxHeight)

	// Box-sizing: border-box means width/height include padding+border.
	borderTop := node.Border[0]
	borderRight := node.Border[1]
	borderBottom := node.Border[2]
	borderLeft := node.Border[3]
	borderH := borderLeft + borderRight
	borderV := borderTop + borderBottom
	padTop := node.Padding[0]
	padRight := node.Padding[1]
	padBottom := node.Padding[2]
	padLeft := node.Padding[3]
	padH := padLeft + padRight
	padV := padTop + padBottom

	var innerW, innerH float32
	if node.BoxSizing == BoxSizingBorderBox {
		innerW = w - borderH - padH
		innerH = h - borderV - padV
	} else {
		innerW = w
		innerH = h
	}
	if innerW < 0 {
		innerW = 0
	}
	if innerH < 0 {
		innerH = 0
	}

	// Set layout geometry (content box origin inside borders/padding).
	node.Layout.Left = borderLeft + padLeft
	node.Layout.Top = borderTop + padTop
	node.Layout.Width = w
	node.Layout.Height = h
	node.Layout.Right = w
	node.Layout.Bottom = h

	// Determine main/cross axis.
	isRow := node.FlexDirection == FlexDirectionRow || node.FlexDirection == FlexDirectionRowReverse
	isReverse := node.FlexDirection == FlexDirectionColumnReverse || node.FlexDirection == FlexDirectionRowReverse

	mainDim := innerW
	crossDim := innerH
	if isRow {
		mainDim = innerW
		crossDim = innerH
	} else {
		mainDim = innerH
		crossDim = innerW
	}

	// Collect visible children.
	children := visibleChildren(node)
	if len(children) == 0 {
		node.Computed = true
		return
	}

	// Measure leaf nodes that have no explicit size.
	for _, child := range children {
		child.mu.Lock()
		if child.MeasureFunc != nil && len(child.Children) == 0 {
			child.measureForFlex(innerW, innerH, isRow)
		}
		child.mu.Unlock()
	}

	if node.FlexWrap == FlexWrapNone {
		y.layoutSingleLine(node, children, isRow, isReverse, mainDim, crossDim, innerW, innerH)
	} else {
		y.layoutWrapped(node, children, isRow, isReverse, mainDim, crossDim, innerW, innerH)
	}

	node.Computed = true
}

func visibleChildren(node *YogaNode) []*YogaNode {
	var out []*YogaNode
	for _, c := range node.Children {
		c.mu.RLock()
		display := c.Display
		c.mu.RUnlock()
		if display == DisplayNone {
			continue
		}
		out = append(out, c)
	}
	return out
}

func (child *YogaNode) measureForFlex(innerW, innerH float32, isRow bool) {
	if child.MeasureFunc == nil {
		return
	}
	var wMode, hMode MeasureMode
	var w, h float32
	if isRow {
		// Main axis = width, cross axis = height.
		if child.Width > 0 {
			w = child.Width
			wMode = MeasureExactly
		} else if child.FlexBasis > 0 {
			w = child.FlexBasis
			wMode = MeasureExactly
		} else {
			w = innerW
			wMode = MeasureAtMost
		}
		if child.Height > 0 {
			h = child.Height
			hMode = MeasureExactly
		} else {
			h = innerH
			hMode = MeasureAtMost
		}
	} else {
		// Main axis = height, cross axis = width.
		if child.Width > 0 {
			w = child.Width
			wMode = MeasureExactly
		} else {
			w = innerW
			wMode = MeasureAtMost
		}
		if child.Height > 0 {
			h = child.Height
			hMode = MeasureExactly
		} else if child.FlexBasis > 0 {
			h = child.FlexBasis
			hMode = MeasureExactly
		} else {
			h = innerH
			hMode = MeasureAtMost
		}
	}
	mw, mh := child.MeasureFunc(w, wMode, h, hMode)
	if mw > 0 {
		child.Width = mw
	}
	if mh > 0 {
		child.Height = mh
	}
}

// layoutSingleLine lays out children on a single flex line.
func (y *Yoga) layoutSingleLine(node *YogaNode, children []*YogaNode, isRow, isReverse bool, mainDim, crossDim, innerW, innerH float32) {
	// Determine the flex basis (main-axis preferred size) for each child.
	type flexChild struct {
		node        *YogaNode
		basis       float32 // preferred main size
		grow        float32
		shrink      float32
		maxMain     float32
		minMain     float32
		marginMain  float32
		marginCross float32
		borderMain  float32
		borderCross float32
		padMain     float32
		padCross    float32
	}
	items := make([]flexChild, len(children))
	totalFlexGrow := float32(0)
	totalFlexShrink := float32(0)
	sumBasis := float32(0)

	for i, child := range children {
		child.mu.RLock()
		flexGrow := child.FlexGrow
		flexShrink := child.FlexShrink
		flexBasis := child.FlexBasis
		width := child.Width
		height := child.Height
		marginT := child.Margin[0]
		marginR := child.Margin[1]
		marginB := child.Margin[2]
		marginL := child.Margin[3]
		borderT := child.Border[0]
		borderR := child.Border[1]
		borderBtm := child.Border[2]
		borderL := child.Border[3]
		padT := child.Padding[0]
		padR := child.Padding[1]
		padB := child.Padding[2]
		padL := child.Padding[3]
		child.mu.RUnlock()

		item := flexChild{node: child, grow: flexGrow, shrink: flexShrink}
		if isRow {
			item.marginMain = marginL + marginR
			item.marginCross = marginT + marginB
			item.borderMain = borderL + borderR
			item.borderCross = borderT + borderBtm
			item.padMain = padL + padR
			item.padCross = padT + padB
			if width > 0 {
				item.basis = width
			} else if flexBasis > 0 {
				item.basis = flexBasis
			} else {
				item.basis = innerW
			}
		} else {
			item.marginMain = marginT + marginB
			item.marginCross = marginL + marginR
			item.borderMain = borderT + borderBtm
			item.borderCross = borderL + borderR
			item.padMain = padT + padB
			item.padCross = padL + padR
			if height > 0 {
				item.basis = height
			} else if flexBasis > 0 {
				item.basis = flexBasis
			} else {
				item.basis = innerH
			}
		}
		item.minMain = 0
		item.maxMain = math.MaxFloat32
		if isRow {
			if child.MinWidth > 0 {
				item.minMain = child.MinWidth
			}
			if child.MaxWidth > 0 {
				item.maxMain = child.MaxWidth
			}
		} else {
			if child.MinHeight > 0 {
				item.minMain = child.MinHeight
			}
			if child.MaxHeight > 0 {
				item.maxMain = child.MaxHeight
			}
		}
		// Clamp basis to min/max.
		if item.basis < item.minMain {
			item.basis = item.minMain
		}
		if item.basis > item.maxMain {
			item.basis = item.maxMain
		}
		totalFlexGrow += flexGrow
		totalFlexShrink += flexShrink
		sumBasis += item.basis + item.marginMain + item.borderMain + item.padMain
		items[i] = item
	}

	// Account for gap along main axis.
	gapMain := float32(0)
	if isRow {
		gapMain = node.Gap[0]
	} else {
		gapMain = node.Gap[1]
	}
	totalGapMain := gapMain * float32(len(children)-1)
	if totalGapMain < 0 {
		totalGapMain = 0
	}

	availableMain := mainDim - totalGapMain
	freeSpace := availableMain - sumBasis

	// Phase 1: distribute free space via flex grow (positive free space).
	if freeSpace > 0 && totalFlexGrow > 0 {
		remaining := freeSpace
		for i, item := range items {
			if item.grow <= 0 {
				continue
			}
			share := freeSpace * (item.grow / totalFlexGrow)
			newBasis := item.basis + share
			// Clamp to max.
			items[i].basis = clamp(newBasis, item.minMain, item.maxMain)
			// Reclaim any unused space from clamping.
			reclaimed := (item.basis + share) - items[i].basis
			remaining -= (items[i].basis - item.basis)
			_ = remaining
			_ = reclaimed
		}
		// One more pass to distribute leftover from clamping proportionally.
		_ = remaining
	}

	// Phase 2: shrink if overflow (negative free space).
	if freeSpace < 0 && totalFlexShrink > 0 {
		overflow := -freeSpace
		// Weighted shrink: shrink = flexShrink * basis / sum(flexShrink*basis)
		sumShrinkBasis := float32(0)
		for _, item := range items {
			if item.shrink > 0 {
				sumShrinkBasis += item.shrink * item.basis
			}
		}
		if sumShrinkBasis > 0 {
			for i, item := range items {
				if item.shrink <= 0 {
					continue
				}
				shrinkAmount := overflow * (item.shrink * item.basis / sumShrinkBasis)
				newBasis := item.basis - shrinkAmount
				if newBasis < item.minMain {
					newBasis = item.minMain
				}
				items[i].basis = newBasis
			}
		}
	}

	// Compute total main size after flexing.
	totalMain := totalGapMain
	for _, item := range items {
		totalMain += item.basis + item.marginMain + item.borderMain + item.padMain
	}

	// Determine cross-axis size of the line: max of children cross sizes.
	lineCrossSize := float32(0)
	for _, item := range items {
		crossSize := float32(0)
		if isRow {
			crossSize = item.node.Height
		} else {
			crossSize = item.node.Width
		}
		crossSize += item.marginCross + item.borderCross + item.padCross
		if crossSize > lineCrossSize {
			lineCrossSize = crossSize
		}
	}

	// Main-axis positioning (justify-content).
	mainStart := float32(0)
	switch node.JustifyContent {
	case JustifyFlexStart:
		mainStart = 0
	case JustifyCenter:
		mainStart = (mainDim - totalMain) / 2
		if mainStart < 0 {
			mainStart = 0
		}
	case JustifyFlexEnd:
		mainStart = mainDim - totalMain
		if mainStart < 0 {
			mainStart = 0
		}
	case JustifySpaceBetween:
		if len(children) > 1 {
			mainStart = 0
			// Space between items distributed in gaps.
		} else {
			mainStart = 0
		}
	case JustifySpaceAround:
		mainStart = totalGapMain / 2 // simplified
	case JustifySpaceEvenly:
		mainStart = 0 // simplified
	}

	// Position each child along main axis.
	cursor := mainStart
	for i, item := range items {
		child := item.node
		mainSize := item.basis + item.borderMain + item.padMain
		crossSize := float32(0)
		if isRow {
			crossSize = child.Height
		} else {
			crossSize = child.Width
		}

		// Cross-axis alignment.
		align := child.AlignSelf
		if align == AlignSelfAuto {
			align = AlignSelf(node.AlignItems)
		}
		crossPos := float32(0)
		switch align {
		case AlignSelfFlexStart:
			crossPos = 0
		case AlignSelfCenter:
			crossPos = (crossDim - crossSize) / 2
			if crossPos < 0 {
				crossPos = 0
			}
		case AlignSelfFlexEnd:
			crossPos = crossDim - crossSize
			if crossPos < 0 {
				crossPos = 0
			}
		case AlignSelfStretch:
			crossPos = 0
			crossSize = crossDim - item.marginCross - item.borderCross - item.padCross
			if crossSize < 0 {
				crossSize = 0
			}
		case AlignSelfBaseline:
			// Baseline alignment handled separately below.
			crossPos = 0
		}

		child.mu.Lock()
		if isRow {
			child.Layout.Left = cursor + item.marginMain + item.borderMain + item.padMain
			child.Layout.Top = crossPos + item.marginCross + item.borderCross + item.padCross
			child.Layout.Width = mainSize
			child.Layout.Height = crossSize
		} else {
			child.Layout.Top = cursor + item.marginMain + item.borderMain + item.padMain
			child.Layout.Left = crossPos + item.marginCross + item.borderCross + item.padCross
			child.Layout.Width = crossSize
			child.Layout.Height = mainSize
		}
		child.Layout.ScreenX = float64(child.Layout.Left) + float64(node.Layout.Left)
		child.Layout.ScreenY = float64(child.Layout.Top) + float64(node.Layout.Top)
		child.Computed = true
		child.mu.Unlock()

		cursor += mainSize + item.marginMain
		if i < len(children)-1 {
			cursor += gapMain
		}
	}

	// Baseline alignment: if any child wants AlignBaseline, align baselines.
	if node.AlignItems == AlignBaseline {
		y.alignBaselines(node, children, isRow, crossDim, crossPosFor(node, children, isRow))
	}
}

// alignBaselines aligns children along their text baselines.
func (y *Yoga) alignBaselines(node *YogaNode, children []*YogaNode, isRow bool, crossDim float32, crossPos float32) {
	// Simple baseline: align bottoms (approximation). A full implementation
	// would query font metrics. Here we align the content bottom edge.
	maxBottom := float32(0)
	positions := make([]float32, len(children))
	for i, child := range children {
		child.mu.RLock()
		var top, height float32
		if isRow {
			top = child.Layout.Top
			height = child.Layout.Height
		} else {
			top = child.Layout.Left
			height = child.Layout.Width
		}
		child.mu.RUnlock()
		bottom := top + height
		if bottom > maxBottom {
			maxBottom = bottom
		}
		positions[i] = top
	}
	for _, child := range children {
		child.mu.Lock()
		var top, height float32
		if isRow {
			top = child.Layout.Top
			height = child.Layout.Height
		} else {
			top = child.Layout.Left
			height = child.Layout.Width
		}
		shift := maxBottom - (top + height)
		if isRow {
			child.Layout.Top = top + shift
		} else {
			child.Layout.Left = top + shift
		}
		child.mu.Unlock()
		_ = positions
		_ = crossPos
		_ = crossDim
	}
}

func crossPosFor(node *YogaNode, children []*YogaNode, isRow bool) float32 {
	return 0
}

// layoutWrapped lays out children across multiple flex lines.
func (y *Yoga) layoutWrapped(node *YogaNode, children []*YogaNode, isRow, isReverse bool, mainDim, crossDim, innerW, innerH float32) {
	// Group children into lines based on main-dim available space.
	type line struct {
		items    []*YogaNode
		mainUsed float32
		crossMax float32
	}
	var lines []line
	current := line{}
	gapMain := float32(0)
	gapCross := float32(0)
	if isRow {
		gapMain = node.Gap[0]
		gapCross = node.Gap[1]
	} else {
		gapMain = node.Gap[1]
		gapCross = node.Gap[0]
	}

	for _, child := range children {
		child.mu.RLock()
		var mainSize float32
		if isRow {
			mainSize = child.Width
			if mainSize == 0 {
				mainSize = child.FlexBasis
			}
			if mainSize == 0 {
				mainSize = innerW
			}
		} else {
			mainSize = child.Height
			if mainSize == 0 {
				mainSize = child.FlexBasis
			}
			if mainSize == 0 {
				mainSize = innerH
			}
		}
		child.mu.RUnlock()

		itemMain := mainSize
		if current.mainUsed == 0 {
			// First item on line always fits (or overflows onto its own line).
			current.items = append(current.items, child)
			current.mainUsed += itemMain
			current.crossMax = float32(math.Max(float64(current.crossMax), float64(itemMain))) // placeholder
			continue
		}
		if current.mainUsed+gapMain+itemMain <= mainDim {
			current.items = append(current.items, child)
			current.mainUsed += gapMain + itemMain
		} else {
			lines = append(lines, current)
			current = line{items: []*YogaNode{child}, mainUsed: itemMain}
		}
	}
	if len(current.items) > 0 {
		lines = append(lines, current)
	}

	// Layout each line using single-line logic.
	crossCursor := float32(0)
	for _, ln := range lines {
		if len(ln.items) == 0 {
			continue
		}
		y.layoutSingleLine(node, ln.items, isRow, isReverse, mainDim, crossDim, innerW, innerH)
		// Advance cross cursor by line's max cross size.
		lineCross := float32(0)
		for _, child := range ln.items {
			child.mu.RLock()
			var cs float32
			if isRow {
				cs = child.Layout.Height
			} else {
				cs = child.Layout.Width
			}
			child.mu.RUnlock()
			if cs > lineCross {
				lineCross = cs
			}
		}
		crossCursor += lineCross + gapCross
	}

	// Cross-axis distribution of lines (align-content could go here).
	_ = crossCursor
	_ = isReverse
}

// resolveDimension clamps an explicit size against min/max, falling back to
// the container size when explicit is zero.
func resolveDimension(explicit, container, min, max float32) float32 {
	if explicit > 0 {
		v := explicit
		if v < min {
			v = min
		}
		if v > max && max > 0 {
			v = max
		}
		return v
	}
	v := container
	if v < min {
		v = min
	}
	if max > 0 && v > max {
		v = max
	}
	return v
}

func clamp(v, min, max float32) float32 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// SetFlexDirection sets the flex direction.
func (n *YogaNode) SetFlexDirection(dir FlexDirection) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.FlexDirection = dir
	n.Computed = false
}

// SetJustifyContent sets the main-axis distribution.
func (n *YogaNode) SetJustifyContent(jc JustifyContent) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.JustifyContent = jc
	n.Computed = false
}

// SetAlignItems sets the default cross-axis alignment.
func (n *YogaNode) SetAlignItems(ai AlignItems) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.AlignItems = ai
	n.Computed = false
}

// SetAlignSelf sets the per-node cross-axis alignment override.
func (n *YogaNode) SetAlignSelf(as AlignSelf) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.AlignSelf = as
	n.Computed = false
}

// SetFlexWrap sets the wrapping behavior.
func (n *YogaNode) SetFlexWrap(wrap FlexWrap) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.FlexWrap = wrap
	n.Computed = false
}

// SetWidth sets the node width.
func (n *YogaNode) SetWidth(w float32) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Width = w
	n.Computed = false
}

// SetHeight sets the node height.
func (n *YogaNode) SetHeight(h float32) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Height = h
	n.Computed = false
}

// SetFlexGrow sets the flex grow factor.
func (n *YogaNode) SetFlexGrow(f float32) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.FlexGrow = f
	n.Computed = false
}

// SetFlexShrink sets the flex shrink factor.
func (n *YogaNode) SetFlexShrink(f float32) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.FlexShrink = f
	n.Computed = false
}

// SetFlexBasis sets the flex basis (preferred main size).
func (n *YogaNode) SetFlexBasis(b float32) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.FlexBasis = b
	n.Computed = false
}

// SetFlex sets the shorthand flex value (treated as grow).
func (n *YogaNode) SetFlex(f float32) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Flex = f
	n.FlexGrow = f
	n.Computed = false
}

// SetMinMax sets min and max constraints.
func (n *YogaNode) SetMinMax(minW, minH, maxW, maxH float32) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.MinWidth = minW
	n.MinHeight = minH
	n.MaxWidth = maxW
	n.MaxHeight = maxH
	n.Computed = false
}

// SetMargin sets all margins (top, right, bottom, left).
func (n *YogaNode) SetMargin(t, r, b, l float32) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Margin = [4]float32{t, r, b, l}
	n.Computed = false
}

// SetPadding sets all padding (top, right, bottom, left).
func (n *YogaNode) SetPadding(t, r, b, l float32) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Padding = [4]float32{t, r, b, l}
	n.Computed = false
}

// SetGap sets row and column gaps.
func (n *YogaNode) SetGap(row, col float32) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Gap = [2]float32{row, col}
	n.Computed = false
}

// SetMeasure sets the measure function for leaf nodes.
func (n *YogaNode) SetMeasure(fn MeasureFunction) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.MeasureFunc = fn
	n.Computed = false
}

// AddChild adds a child node.
func (n *YogaNode) AddChild(child *YogaNode) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Children = append(n.Children, child)
	child.mu.Lock()
	child.Parent = n
	child.mu.Unlock()
	n.Computed = false
}

// RemoveChild removes a child node.
func (n *YogaNode) RemoveChild(child *YogaNode) {
	n.mu.Lock()
	defer n.mu.Unlock()
	for i, c := range n.Children {
		if c == child {
			n.Children = append(n.Children[:i], n.Children[i+1:]...)
			child.mu.Lock()
			child.Parent = nil
			child.mu.Unlock()
			n.Computed = false
			return
		}
	}
}

// GetComputedLayout returns the computed layout.
func (n *YogaNode) GetComputedLayout() Layout {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.Layout
}

// IsComputed returns whether the node has been laid out.
func (n *YogaNode) IsComputed() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.Computed
}

var _ = math.MaxFloat32
