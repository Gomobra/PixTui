package core

import (
	"image/color"
)

// RGBA represents a 4-component color
type RGBA struct {
	R, G, B, A uint8
}

func (c RGBA) RGBA() (uint32, uint32, uint32, uint32) {
	return uint32(c.R) * 257, uint32(c.G) * 257, uint32(c.B) * 257, uint32(c.A) * 257
}

var _ color.Color = RGBA{}

// TextAttributes constants (bit flags)
const (
	TextAttributeNone          = 0
	TextAttributeBold          = 1 << 0
	TextAttributeDim           = 1 << 1
	TextAttributeItalic        = 1 << 2
	TextAttributeUnderline     = 1 << 3
	TextAttributeBlink         = 1 << 4
	TextAttributeInverse       = 1 << 5
	TextAttributeHidden        = 1 << 6
	TextAttributeStrikethrough = 1 << 7
	TextAttributeReverse       = TextAttributeInverse
)

// ThemeMode is the terminal theme mode
type ThemeMode string

const (
	ThemeModeDark  ThemeMode = "dark"
	ThemeModeLight ThemeMode = "light"
)

// CursorStyle is the cursor style
type CursorStyle string

const (
	CursorStyleBlock     CursorStyle = "block"
	CursorStyleLine      CursorStyle = "line"
	CursorStyleUnderline CursorStyle = "underline"
	CursorStyleDefault   CursorStyle = "default"
)

// SelectionOccupancy defines how selection occupies cells
type SelectionOccupancy string

const (
	SelectionOccupancyCell     SelectionOccupancy = "cell"
	SelectionOccupancyBoundary SelectionOccupancy = "boundary"
)

// SelectionBehavior defines how selection expands
type SelectionBehavior string

const (
	SelectionBehaviorCell SelectionBehavior = "cell"
	SelectionBehaviorWord SelectionBehavior = "word"
	SelectionBehaviorLine SelectionBehavior = "line"
)

// MousePointerStyle is the mouse pointer style
type MousePointerStyle string

const (
	MousePointerAuto          MousePointerStyle = "auto"
	MousePointerDefault       MousePointerStyle = "default"
	MousePointerNone          MousePointerStyle = "none"
	MousePointerContextMenu   MousePointerStyle = "context-menu"
	MousePointerHelp          MousePointerStyle = "help"
	MousePointerPointer       MousePointerStyle = "pointer"
	MousePointerProgress      MousePointerStyle = "progress"
	MousePointerWait          MousePointerStyle = "wait"
	MousePointerCell          MousePointerStyle = "cell"
	MousePointerCrosshair     MousePointerStyle = "crosshair"
	MousePointerText          MousePointerStyle = "text"
	MousePointerVerticalText  MousePointerStyle = "vertical-text"
	MousePointerAlias         MousePointerStyle = "alias"
	MousePointerCopy          MousePointerStyle = "copy"
	MousePointerMove          MousePointerStyle = "move"
	MousePointerNoDrop        MousePointerStyle = "no-drop"
	MousePointerNotAllowed    MousePointerStyle = "not-allowed"
	MousePointerGrab          MousePointerStyle = "grab"
	MousePointerGrabbing      MousePointerStyle = "grabbing"
	MousePointerAllScroll     MousePointerStyle = "all-scroll"
	MousePointerColResize     MousePointerStyle = "col-resize"
	MousePointerRowResize     MousePointerStyle = "row-resize"
	MousePointerNResize       MousePointerStyle = "n-resize"
	MousePointerEResize       MousePointerStyle = "e-resize"
	MousePointerSResize       MousePointerStyle = "s-resize"
	MousePointerWResize       MousePointerStyle = "w-resize"
	MousePointerNeResize      MousePointerStyle = "ne-resize"
	MousePointerNwResize      MousePointerStyle = "nw-resize"
	MousePointerSeResize      MousePointerStyle = "se-resize"
	MousePointerSwResize      MousePointerStyle = "sw-resize"
	MousePointerEwResize      MousePointerStyle = "ew-resize"
	MousePointerNsResize      MousePointerStyle = "ns-resize"
	MousePointerNeswResize    MousePointerStyle = "nesw-resize"
	MousePointerNwseResize    MousePointerStyle = "nwse-resize"
	MousePointerZoomIn        MousePointerStyle = "zoom-in"
	MousePointerZoomOut       MousePointerStyle = "zoom-out"
)

// CursorStyleOptions configures cursor appearance
type CursorStyleOptions struct {
	Style    CursorStyle
	Blinking bool
	Color    *RGBA
	Cursor   MousePointerStyle
}

// DebugOverlayCorner is the corner for debug overlay
type DebugOverlayCorner int

const (
	DebugOverlayCornerTopLeft DebugOverlayCorner = iota
	DebugOverlayCornerTopRight
	DebugOverlayCornerBottomLeft
	DebugOverlayCornerBottomRight
)

// TargetChannel is the color target channel
type TargetChannel int

const (
	TargetChannelFG  TargetChannel = 1
	TargetChannelBG  TargetChannel = 2
	TargetChannelBoth TargetChannel = 3
)

// WidthMethod is the method for calculating character cell width
type WidthMethod string

const (
	WidthMethodWCWidth     WidthMethod = "wcwidth"
	WidthMethodUnicode     WidthMethod = "unicode"
	WidthMethodUnicodeWide WidthMethod = "unicode-wide"
)

// TerminalMultiplexer is the detected terminal multiplexer
type TerminalMultiplexer string

const (
	TerminalMultiplexerNone   TerminalMultiplexer = "none"
	TerminalMultiplexerTMux   TerminalMultiplexer = "tmux"
	TerminalMultiplexerZellij TerminalMultiplexer = "zellij"
	TerminalMultiplexerScreen TerminalMultiplexer = "screen"
	TerminalMultiplexerUnknown TerminalMultiplexer = "unknown"
)

// TerminalCapabilityState is the state of a terminal capability
type TerminalCapabilityState string

const (
	TerminalCapabilityUnknown   TerminalCapabilityState = "unknown"
	TerminalCapabilitySupported TerminalCapabilityState = "supported"
	TerminalCapabilityUnsupported TerminalCapabilityState = "unsupported"
)

// ImageRenderProtocol is the protocol for rendering images
type ImageRenderProtocol string

const (
	ImageRenderProtocolAuto  ImageRenderProtocol = "auto"
	ImageRenderProtocolKitty ImageRenderProtocol = "kitty"
	ImageRenderProtocolSixel ImageRenderProtocol = "sixel"
	ImageRenderProtocolBlocks ImageRenderProtocol = "blocks"
)

// TerminalInfo describes the terminal
type TerminalInfo struct {
	Name          string
	Version       string
	FromXTVersion bool
}

// TerminalCapabilities describes terminal capabilities
type TerminalCapabilities struct {
	KittyKeyboard          bool
	KittyGraphics          bool
	RGB                    bool
	ANSI256                bool
	Unicode                WidthMethod
	SGRPixels              bool
	ColorSchemeUpdates     bool
	ExplicitWidth          bool
	ScaledText             bool
	Sixel                  bool
	FocusTracking          bool
	Sync                   bool
	BracketedPaste         bool
	Hyperlinks             bool
	OSC52                  bool
	OSC52Support           TerminalCapabilityState
	Notifications          bool
	ExplicitCursorPositioning bool
	Remote                 bool
	ProgramStatus          bool
	Multiplexer            TerminalMultiplexer
	ImageProtocol          ImageRenderProtocol
	Terminal               TerminalInfo
}

// ProgramStatusKind is the kind of program status
type ProgramStatusKind string

const (
	ProgramStatusKindPermission ProgramStatusKind = "permission"
	ProgramStatusKindQuestion   ProgramStatusKind = "question"
	ProgramStatusKindAuth       ProgramStatusKind = "auth"
)

// ProgramStatusRecord is a program status record
type ProgramStatusRecord struct {
	ID      string
	App     string
	Title   string
	Message string
}

// ProgramStatus is a program status report
type ProgramStatus interface {
	isProgramStatus()
}

type ProgramStatusIdle struct {
	ProgramStatusRecord
	State string
}

func (ProgramStatusIdle) isProgramStatus() {}

type ProgramStatusWorking struct {
	ProgramStatusRecord
	State    string
	Progress *float64
}

func (ProgramStatusWorking) isProgramStatus() {}

type ProgramStatusBlocked struct {
	ProgramStatusRecord
	State   string
	Kind    ProgramStatusKind
	Progress *float64
}

func (ProgramStatusBlocked) isProgramStatus() {}

type ProgramStatusClear struct {
	ID string
}

func (ProgramStatusClear) isProgramStatus() {}

// CliRendererFrameEvent represents a frame rendered event.
type CliRendererFrameEvent struct {
	FrameID    uint64
	DeltaTime  float64
}

// CliRendererExternalOutputEvent represents an external output event.
type CliRendererExternalOutputEvent struct {
	Text            string
	Rows            []string
	Width           int
	Height          int
	RowColumns      int
	StartOnNewLine  bool
	TrailingNewline bool
}

// RendererEvents defines renderer event handlers
type RendererEvents struct {
	Resize           func(width, height int)
	Key              func(data []byte)
	MemorySnapshot   func(heapUsed, heapTotal, arrayBuffers int64)
	Capabilities     func(caps TerminalCapabilities)
	Selection        func(selection *Selection)
	FocusedRenderable func(current, previous *Renderable)
	FocusedEditor    func(current, previous *EditorView)
	DebugOverlayToggle func(enabled bool)
	ThemeMode        func(mode ThemeMode)
	Frame            func(event CliRendererFrameEvent)
	ExternalOutput   func(event CliRendererExternalOutputEvent)
	Destroy          func()
}

// AnimationTickCallback is called on animation ticks
type AnimationTickCallback func(deltaTime float64, ticks int, time float64)

// AnimationTickOptions configures animation ticks
type AnimationTickOptions struct {
	Tolerance float64
}

// RenderContext is the context passed to renderables during rendering
type RenderContext struct {
	Events                    RendererEvents
	Width                     int
	Height                    int
	TerminalWidth             int
	TerminalHeight            int
	Resolution                *struct{ Width, Height int }
	FrameID                   uint64
	Clock                     Clock
	RequestRender             func()
	RequestAnimationFrame     func(callback func()) int
	CancelAnimationFrame      func(handle int)
	SetCursorPosition         func(x, y int, visible bool)
	SetCursorStyle            func(options CursorStyleOptions)
	SetCursorColor            func(color RGBA)
	SetMousePointer           func(shape MousePointerStyle)
	WidthMethod               WidthMethod
	Capabilities              *TerminalCapabilities
	RequestLive               func()
	DropLive                  func()
	OnAnimationTick           func(fps float64, callback AnimationTickCallback, options *AnimationTickOptions) func()
	HasSelection              bool
	GetSelection              func() *Selection
	RequestSelectionUpdate    func()
	CurrentFocusedRenderable  *Renderable
	CurrentFocusedEditor      *EditorView
	FocusRenderable           func(r *Renderable)
	BlurRenderable            func(r *Renderable)
	ClearSelection            func()
	StartSelection            func(r *Renderable, x, y int, behavior SelectionBehavior)
	UpdateSelection           func(current *Renderable, x, y int, options *SelectionOptions)
	NativeScene               *NativeScene
	LifecyclePasses           *LifecyclePasses
}

// SelectionOptions configures selection updates
type SelectionOptions struct {
	FinishDragging bool
}

// Selection represents a text selection
type Selection struct {
	Start int
	End   int
}

// TextStyle represents text styling applied to a renderable.
type TextStyle struct {
	// FG is the foreground color. When nil, the renderable default is used.
	FG *RGBA
	// BG is the background color. When nil, the renderable default is used.
	BG *RGBA
	// Attributes are the text attributes (bold, dim, italic, ...).
	Attributes uint32
}

// ViewportBounds represents a rectangular viewport
type ViewportBounds struct {
	X, Y, Width, Height int
}

// Highlight represents a text highlight range
type Highlight struct {
	Start    int
	End      int
	StyleID  int
	Priority *int
	HLRef    *int
}

// LineInfo contains line layout information
type LineInfo struct {
	LineStartCols     []int
	LineWidthCols     []int
	LineWidthColsMax  int
	LineSources       []int
	LineWraps         []int
}

// LineInfoProvider provides line info for text rendering
type LineInfoProvider interface {
	GetLineInfo() LineInfo
	GetLineCount() int
	GetVirtualLineCount() int
	GetScrollY() int
}

// CapturedSpan is a captured text span with styling
type CapturedSpan struct {
	Text      string
	FG        RGBA
	BG        RGBA
	Attributes uint32
	Width     int
}

// CapturedLine is a line of captured spans
type CapturedLine struct {
	Spans []CapturedSpan
}

// CapturedFrame is a captured frame for testing
type CapturedFrame struct {
	Cols   int
	Rows   int
	Cursor [2]int
	Lines  []CapturedLine
}

// Clock is a clock interface for timing
type Clock interface {
	Now() float64
	Since(start float64) float64
}

// SystemClock is the system clock implementation
type SystemClock struct{}

func (SystemClock) Now() float64 { return 0 }
func (SystemClock) Since(start float64) float64 { return 0 }

// LayoutEvents are layout change events
type LayoutEvents int

const (
	LayoutChanged LayoutEvents = iota
	Resized
)

// RenderableEvents are renderable lifecycle events
type RenderableEvents int

const (
	RenderableFocused RenderableEvents = iota
	RenderableBlurred
	RenderableDestroyed
)

// Position represents CSS-like positioning
type Position struct {
	Top, Right, Bottom, Left *string
}

// LayoutOptions contains layout configuration
type LayoutOptions struct {
	FlexGrow       *float32
	FlexShrink     *float32
	FlexDirection  *string
	FlexWrap       *string
	AlignItems     *string
	JustifyContent *string
	AlignSelf      *string
	FlexBasis      *string
	Position       *string
	Overflow       *string
	Top, Right, Bottom, Left *string
	MinWidth, MinHeight *string
	MaxWidth, MaxHeight *string
	Margin, MarginX, MarginY *string
	MarginTop, MarginRight, MarginBottom, MarginLeft *string
	Padding, PaddingX, PaddingY *string
	PaddingTop, PaddingRight, PaddingBottom, PaddingLeft *string
	EnableLayout *bool
}

// RenderableOptions contains options for creating a renderable
type RenderableOptions[T any] struct {
	ID             string
	Width, Height  *string
	ZIndex         *int
	Visible        *bool
	Buffered       *bool
	Live           *bool
	AnimationFPS   *float64
	Opacity        *float64
	// ForegroundColor is the default foreground color for the renderable.
	ForegroundColor *RGBA
	Attributes      uint32
	// BackgroundColor is the default background color for the renderable.
	BackgroundColor *RGBA
	RenderBefore   func(r *Renderable, buffer *NativeBuffer, deltaTime float64)
	RenderAfter    func(r *Renderable, buffer *NativeBuffer, deltaTime float64)
	OnMouse        func(r *Renderable, event *MouseEvent)
	OnMouseDown    func(r *Renderable, event *MouseEvent)
	OnMouseUp      func(r *Renderable, event *MouseEvent)
	OnMouseMove    func(r *Renderable, event *MouseEvent)
	OnMouseDrag    func(r *Renderable, event *MouseEvent)
	OnMouseDragEnd func(r *Renderable, event *MouseEvent)
	OnMouseDrop    func(r *Renderable, event *MouseEvent)
	OnMouseOver    func(r *Renderable, event *MouseEvent)
	OnMouseOut     func(r *Renderable, event *MouseEvent)
	OnMouseScroll  func(r *Renderable, event *MouseEvent)
	OnPaste        func(r *Renderable, event *PasteEvent)
	OnKeyDown      func(key *KeyEvent)
}

// MouseEvent represents a mouse event
type MouseEvent struct {
	X, Y      int
	Button    uint32
	Pressed   bool
	Modifiers uint32
}

// PasteEvent represents a paste event
type PasteEvent struct {
	Data []byte
}

// KeyEvent represents a key event
type KeyEvent struct {
	Key       string
	Modifiers uint32
}


