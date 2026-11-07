# API Reference

## Core Package

### CliRenderer

The main renderer for terminal UI.

```go
renderer := core.NewCliRenderer(width, height)
```

**Methods:**
- `Start() error` — Start the renderer
- `Render(root, delta) error` — Render a frame
- `WriteFrame() error` — Write the current frame to terminal
- `Destroy()` — Clean up resources
- `Context() *RenderContext` — Get the render context
- `Dimensions() (int, int)` — Get terminal dimensions

### Renderable

Base renderable interface.

```go
type RenderableIface interface {
    RenderSelf(buffer *NativeBuffer, x, y, width, height int)
    AddChild(child *Renderable)
    RemoveChild(child *Renderable)
    SetVisible(visible bool)
    SetZIndex(zIndex int)
}
```

### RenderableOptions

Common options for all renderables.

```go
type RenderableOptions[T any] struct {
    ID      string
    Width   *string
    Height  *string
    Visible *bool
    ZIndex  *int
}
```

### NativeBuffer

Terminal buffer for cell-based rendering.

```go
buffer := core.NewNativeBuffer(width, height)
```

**Methods:**
- `Access() native.BufferAccess` — Get buffer access
- `SetCell(x, y int, char rune, fg, bg uint16, attrs uint32)` — Set a cell
- `Clear()` — Clear the buffer
- `Resize(width, height int)` — Resize the buffer

### RenderContext

Context for creating renderables.

```go
ctx := renderer.Context()
```

**Methods:**
- `Capabilities() *TerminalCapabilities` — Get terminal capabilities
- `SetCapabilities(caps *TerminalCapabilities)` — Set terminal capabilities

## Native Package

### Terminal

Terminal I/O handling.

```go
terminal := native.NewTerminal(stdoutFd)
```

**Methods:**
- `GetDimensions() (int, int)` — Get terminal dimensions
- `SetDimensions(width, height int)` — Set terminal dimensions
- `WatchResize(callback func(ResizeEvent))` — Watch for resize events
- `StopResizeWatch()` — Stop watching for resize
- `Close()` — Close the terminal

### Yoga

Flexbox layout engine.

```go
node := native.NewYogaNode()
```

**Methods:**
- `SetWidth(width float64)` — Set width
- `SetHeight(height float64)` — Set height
- `SetFlexDirection(dir FlexDirection)` — Set flex direction
- `SetJustifyContent(justify JustifyContent)` — Set justify content
- `SetAlignItems(align AlignItems)` — Set align items
- `CalculateLayout(width, height float64)` — Calculate layout

### Buffer

Optimized buffer for cell-based rendering.

```go
buffer := native.NewOptimizedBuffer(width, height)
```

**Methods:**
- `SetCell(x, y int, char rune, fg, bg uint16, attrs uint32)` — Set a cell
- `Clear()` — Clear the buffer
- `FillRect(x, y, width, height int, char rune, fg, bg uint16, attrs uint32)` — Fill a rectangle
- `Access() BufferAccess` — Get buffer access
- `Resize(width, height int)` — Resize the buffer
- `Destroy()` — Destroy the buffer

### Image

Image loading and rendering.

```go
img, err := native.LoadImage(path)
```

**Methods:**
- `Width() int` — Get image width
- `Height() int` — Get image height
- `Data() []uint8` — Get RGBA data
- `Destroy()` — Destroy the image

### Video

Video decoding and rendering.

```go
video := native.NewVideo(width, height)
```

**Methods:**
- `Play()` — Start playback
- `Pause()` — Pause playback
- `Toggle()` — Toggle play/pause
- `Seek(seconds float64)` — Seek to position
- `Update()` — Update frame
- `Destroy()` — Destroy the video

### Audio

Audio engine for sound generation.

```go
audio := native.NewAudioEngine(sampleRate int)
```

**Methods:**
- `AddVoice(voice *AudioVoice)` — Add a voice
- `RemoveVoice(id int)` — Remove a voice
- `Mix()` — Mix all voices
- `WriteMix(buffer []float32)` — Write mixed audio
- `Destroy()` — Destroy the audio engine

### Clipboard

Clipboard operations.

```go
clipboard := native.NewClipboard()
```

**Methods:**
- `Copy(text string) error` — Copy text to clipboard
- `Paste() (string, error)` — Paste text from clipboard
- `Destroy()` — Destroy the clipboard

### Input Parser

Parse terminal input events.

```go
parser := native.NewInputParser()
```

**Methods:**
- `Parse(data []byte) []InputEvent` — Parse input data
- `Destroy()` — Destroy the parser

### Capabilities

Terminal capability detection.

```go
caps := native.NewTerminalCapabilities()
```

**Methods:**
- `SupportsColor() bool` — Check color support
- `SupportsMouse() bool` — Check mouse support
- `SupportsUnicode() bool` — Check unicode support
- `GetImageProtocol() ImageRenderProtocol` — Get image protocol

### Syntax Style

Syntax highlighting styles.

```go
style := native.NewSyntaxStyle()
```

**Methods:**
- `RegisterStyle(name string, style StyleDefinition)` — Register a style
- `ResolveStyleId(name string) int` — Resolve style ID
- `GetStyleId(name string) int` — Get style ID
- `MergeStyles(base, overlay int) int` — Merge styles
- `Destroy()` — Destroy the style

### Event Emitter

Event handling system.

```go
emitter := native.NewEventEmitter()
```

**Methods:**
- `On(event string, handler func(...interface{}))` — Register handler
- `Off(event string, handler func(...interface{}))` — Unregister handler
- `Emit(event string, args ...interface{})` — Emit event
- `ListenerCount(event string) int` — Get listener count

### File Logger

File-based logging.

```go
logger := native.NewFileLogger(path string)
```

**Methods:**
- `Log(level LogLevel, message string)` — Log a message
- `Close()` — Close the logger

### Mem Registry

Memory tracking.

```go
registry := native.NewMemRegistry()
```

**Methods:**
- `Register(size int) int` — Register allocation
- `Unregister(id int)` — Unregister allocation
- `TotalAllocated() int` — Get total allocated
- `AllocationCount() int` — Get allocation count

### Native Renderable

Native renderable wrapper.

```go
nr := native.NewNativeRenderable()
```

**Methods:**
- `Paint(buffer *NativeBuffer, x, y, width, height int)` — Paint to buffer
- `SetPaint(fn func(*NativeBuffer, int, int, int, int))` — Set paint function
- `Destroy()` — Destroy the renderable

### Tree Sitter

Tree-sitter syntax highlighting.

```go
lang := native.NewTreeSitterLanguage(language)
```

**Methods:**
- `Language() *C.TSLanguage` — Get the language
- `Destroy()` — Destroy the language

### Embedded Terminal

Embedded terminal emulator.

```go
terminal := native.NewEmbeddedTerminal()
```

**Methods:**
- `StartProcess(cmd string, args []string)` — Start a process
- `WriteStdin(data []byte)` — Write to stdin
- `Close()` — Close the terminal

### Split Scrollback

Split scrollback buffer.

```go
scrollback := native.NewSplitScrollback()
```

**Methods:**
- `AddLine(line string)` — Add a line
- `GetLines() []string` — Get all lines
- `Clear()` — Clear the buffer
- `Destroy()` — Destroy the scrollback

### Stdout Writer

Stdout writer with buffering.

```go
writer := native.NewStdoutWriter()
```

**Methods:**
- `Write(data []byte) (int, error)` — Write data
- `Flush()` — Flush the buffer
- `Destroy()` — Destroy the writer

### Native Span Feed

Native span feed for text rendering.

```go
feed := native.NewNativeSpanFeed()
```

**Methods:**
- `Feed(data []byte)` — Feed data
- `Drain()` — Drain the feed
- `Close()` — Close the feed

### Encoded Unicode

Encoded unicode handling.

```go
encoded := native.NewEncodedUnicode()
```

**Methods:**
- `Encode(text string) []byte` — Encode text
- `Decode(data []byte) string` — Decode data
- `Destroy()` — Destroy the encoder

### Link

Terminal hyperlink support.

```go
link := native.NewLink(url, text string)
```

**Methods:**
- `URL() string` — Get the URL
- `Text() string` — Get the text
- `Destroy()` — Destroy the link

### Kitty Transport

Kitty graphics protocol transport.

```go
transport := native.NewKittyTransport()
```

**Methods:**
- `Send(data []byte)` — Send data
- `Receive() []byte` — Receive data
- `Destroy()` — Destroy the transport

### Editor View

Editor view with syntax highlighting.

```go
editor := native.NewEditorView()
```

**Methods:**
- `SetContent(content string)` — Set content
- `GetContent() string` — Get content
- `SetSyntaxStyle(style *SyntaxStyle)` — Set syntax style
- `Destroy()` — Destroy the editor view

### Text Buffer

Text buffer with rope data structure.

```go
buffer := native.NewTextBuffer()
```

**Methods:**
- `Insert(pos int, text string)` — Insert text
- `Delete(start, end int)` — Delete text
- `Replace(start, end int, text string)` — Replace text
- `GetString() string` — Get content
- `Length() int` — Get length
- `Destroy()` — Destroy the buffer

### Text Buffer View

Text buffer view with scrolling.

```go
view := native.NewTextBufferView(buffer)
```

**Methods:**
- `SetScrollY(y int)` — Set scroll position
- `ScrollUp(lines int)` — Scroll up
- `ScrollDown(lines int)` — Scroll down
- `Destroy()` — Destroy the view

### Edit Buffer

Edit buffer with undo/redo.

```go
editBuffer := native.NewEditBuffer()
```

**Methods:**
- `Insert(pos int, text string)` — Insert text
- `Delete(start, end int)` — Delete text
- `Undo()` — Undo last edit
- `Redo()` — Redo last undone edit
- `CanUndo() bool` — Check if undo available
- `CanRedo() bool` — Check if redo available
- `GetString() string` — Get content
- `Destroy()` — Destroy the buffer

### Rope

Rope data structure for text editing.

```go
rope := native.NewRope()
```

**Methods:**
- `Insert(pos int, text string)` — Insert text
- `Delete(start, end int)` — Delete text
- `Replace(start, end int, text string)` — Replace text
- `Append(text string)` — Append text
- `Prepend(text string)` — Prepend text
- `Length() int` — Get length
- `GetString() string` — Get content
- `StoreUndo(meta interface{})` — Store undo state
- `Undo(meta interface{}) bool` — Undo
- `Redo() bool` — Redo
- `CanUndo() bool` — Check if undo available
- `CanRedo() bool` — Check if redo available
- `ClearHistory()` — Clear undo/redo history
- `Destroy()` — Destroy the rope

### Ghostty VT

VT terminal emulator.

```go
vt := native.NewVTTerminal(width, height)
```

**Methods:**
- `Feed(data []byte)` — Feed data to terminal
- `FeedString(s string)` — Feed string to terminal
- `PlainString() string` — Get plain text
- `Cell(x, y int) *VTCell` — Get cell at position
- `Resize(width, height int)` — Resize terminal
- `Clear()` — Clear terminal
- `Destroy()` — Destroy the terminal

### Lib

Entry point module.

```go
native.SetIO(writer)
```

**Functions:**
- `SetIO(w io.Writer)` — Set process-wide I/O writer
- `DefaultStdOptions()` — Get default std options

### Opentui

Core module aggregating public API.

```go
// Compile-time verification of all public API
```

## React Package

### ReactRenderer

React-style renderer for PixTui.

```go
renderer := react.NewReactRenderer(ctx)
```

**Methods:**
- `Render(element ReactElement, ctx *RenderContext)` — Render element tree
- `Destroy()` — Destroy the renderer

### Hooks

| Hook | Description |
|---|---|
| `UseState` | State management |
| `UseEffect` | Side effects |
| `UseLayoutEffect` | Synchronous side effects |
| `UseMemo` | Memoized values |
| `UseCallback` | Memoized callbacks |
| `UseRef` | Mutable ref |
| `UseContext` | Context consumption |

### JSX

JSX runtime support.

```go
element := react.JSX("box", props, children...)
```

### DevTools

Development tools for debugging.

```go
devtools := react.NewDevTools(renderer)
devtools.Enable()
tree := devtools.FiberTree()
```

### Test Utilities

```go
tr := react.TestRender(element, opts)
tr.Capture()
tr.Dispose()
```

## Solid Package

### SolidRenderer

Solid-style renderer for PixTui.

```go
renderer := solid.NewSolidRenderer(ctx)
```

**Methods:**
- `Render(component SolidComponent, ctx *RenderContext)` — Render component tree
- `Destroy()` — Destroy the renderer

### Reactive Primitives

| Primitive | Description |
|---|---|
| `CreateSignal` | Create reactive signal |
| `CreateEffect` | Run effect on signal change |
| `CreateMemo` | Memoized computed value |
| `CreateComponent` | Create component from render function |
| `CreateContext` | Create context for dependency injection |

### Components

| Component | Purpose |
|---|---|
| `Box` | Container with optional border and title |
| `Text` | Text content |
| `Input` | Single-line text input |
| `Select` | Selectable options |
| `ScrollBox` | Scrollable content container |
| `Textarea` | Multi-line text editing |
| `Slider` | Numeric value selection |
| `TabSelect` | Tab-style selection |
| `Image` | Image rendering |
| `Video` | Video rendering |
| `Code` | Code display |
| `Markdown` | Markdown content display |
| `Diff` | Diff-style text display |
| `LineNumber` | Line-number display |
| `ScrollBar` | Scroll position indicator |
| `TextTable` | Tabular text display |
| `ASCIIFont` | ASCII-style font text |
| `FrameBuffer` | Framebuffer-backed content |
| `EditBuffer` | Editable text buffer |

### Extras

| Function | Description |
|---|---|
| `Portal` | Render children to different location |
| `Dynamic` | Render component based on runtime value |
| `createScrollbackWriter` | Write snapshots to scrollback |
| `createSolidSlotRegistry` | Create slot registry for plugins |

### Test Utilities

```go
tr := solid.TestRender(component, opts)
tr.CaptureCharFrame()
```

## SSH Package

### Server

SSH server with authentication.

```go
server := ssh.NewServer("0.0.0.0:2222", handler)
```

**Methods:**
- `SetPasswordAuth(cb func(user, password string) bool)` — Set password auth
- `SetPublicKeyAuth(cb func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error))` — Set public key auth
- `SetKeyboardInteractiveAuth(cb func(conn ssh.ConnMetadata, challenge ssh.KeyboardInteractiveChallenge) (*ssh.Permissions, error))` — Set keyboard-interactive auth
- `SetAuthorizedKeys(keys []ssh.PublicKey)` — Set authorized keys
- `LoadAuthorizedKeys(path string) error` — Load authorized_keys file
- `Listen() error` — Start listening
- `Close() error` — Close the server

### Session

SSH session management.

```go
session := ssh.NewSession(conn)
```

**Methods:**
- `WriteFrame(frame string)` — Write ANSI frame
- `SetDimensions(width, height int)` — Set terminal dimensions
- `Close()` — Close the session

## Three Package

### Renderer

3D renderer with CPU rasterization.

```go
renderer := three.NewRenderer(width, height)
```

**Methods:**
- `Render(scene *Scene, camera *Camera)` — Render scene
- `SetClearColor(color color.RGBA)` — Set clear color
- `GetClearColor() color.RGBA` — Get clear color
- `SetActiveCamera(camera *Camera)` — Set active camera
- `GetActiveCamera() *Camera` — Get active camera
- `ToggleDebugStats()` — Toggle debug stats
- `Destroy()` — Destroy the renderer

### Scene

3D scene management.

```go
scene := three.NewScene()
```

**Methods:**
- `Add(obj *Object)` — Add object
- `Remove(obj *Object)` — Remove object
- `GetObjects() []*Object` — Get all objects
- `SetBackgroundColor(color color.RGBA)` — Set background color
- `SetFog(fog *Fog)` — Set fog

### Camera

3D camera.

```go
camera := three.NewCamera()
```

**Methods:**
- `SetPosition(pos Vector3)` — Set position
- `SetRotation(rot Vector3)` — Set rotation
- `SetFOV(fov float64)` — Set field of view
- `UpdateProjectionMatrix()` — Update projection matrix

### Mesh

3D mesh.

```go
mesh := three.NewMesh(geometry, material)
```

**Methods:**
- `SetPosition(pos Vector3)` — Set position
- `SetRotation(rot Vector3)` — Set rotation
- `SetScale(scale Vector3)` — Set scale
- `SetVisible(visible bool)` — Set visibility

### Geometry

3D geometry.

```go
geometry := three.NewGeometry()
```

**Built-in Geometries:**
- `NewBoxGeometry(width, height, depth float64)` — Box
- `NewPlaneGeometry(width, height float64)` — Plane
- `NewSphereGeometry(radius float64, segments int)` — Sphere
- `NewCylinderGeometry(radiusTop, radiusBottom, height float64, segments int)` — Cylinder
- `NewConeGeometry(radius, height float64, segments int)` — Cone
- `NewTorusGeometry(radius, tube float64, radialSegments, tubularSegments int)` — Torus
- `NewRingGeometry(innerRadius, outerRadius float64, segments int)` — Ring

### Material

3D material.

```go
material := three.NewMaterial(shader)
```

**Methods:**
- `SetColor(color color.RGBA)` — Set color
- `SetTransparent(transparent bool)` — Set transparency
- `SetAlphaTest(alpha float64)` — Set alpha test

### Light

3D light.

```go
light := three.NewLight(LightDirectional)
```

**Types:**
- `LightAmbient` — Ambient light
- `LightDirectional` — Directional light
- `LightPoint` — Point light

### Math

| Type | Description |
|---|---|
| `Vector2` | 2D vector |
| `Vector3` | 3D vector |
| `Quaternion` | Quaternion rotation |
| `Matrix4` | 4x4 matrix |
| `Euler` | Euler angles |
| `Color` | Color with RGBA |

## Web Package

### Site

Documentation site.

```go
site := web.NewSite("My Docs")
```

**Methods:**
- `AddPage(page *Page)` — Add a page
- `GetPage(path string) *Page` — Get a page
- `Serve(addr string)` — Start HTTP server
- `BuildToDir(dir string) error` — Build static site

### Page

Content page.

```go
page := web.NewPage("Getting Started", "# Getting Started\n\nWelcome!")
```

**Methods:**
- `Path() string` — Get path
- `Title() string` — Get title
- `Content() string` — Get content
- `SetMetadata(key, value string)` — Set metadata

### Router

HTTP router.

```go
router := web.NewRouter()
```

**Methods:**
- `Get(path string, handler func(w http.ResponseWriter, r *http.Request))` — Register GET route
- `Post(path string, handler func(w http.ResponseWriter, r *http.Request))` — Register POST route
- `Use(middleware func(http.Handler) http.Handler)` — Add middleware
- `ServeHTTP(w http.ResponseWriter, r *http.Request)` — Serve HTTP

### DocSearch

Full-text search.

```go
search := web.NewDocSearch()
```

**Methods:**
- `Index(page *Page)` — Index a page
- `Search(query string) []SearchResult` — Search
- `Highlight(text, query string) string` — Highlight matches

## Keymap Package

### Keymap

Key binding engine.

```go
keymap := keymap.NewKeymap[TTarget, TEvent]()
```

**Methods:**
- `Bind(key, command string)` — Bind a key
- `Unbind(key string)` — Unbind a key
- `RunCommand(name string)` — Run a command
- `GetCommands() []Command` — Get all commands
- `GetBindings() []Binding` — Get all bindings

### Layer

Keymap layer.

```go
layer := keymap.NewLayer(name string, priority int)
```

**Methods:**
- `AddBinding(binding Binding)` — Add a binding
- `RemoveBinding(key string)` — Remove a binding
- `Enable()` — Enable the layer
- `Disable()` — Disable the layer

## Animation Package

### Animation

Animation with easing.

```go
animation := animation.NewAnimation(duration time.Duration, easing animation.EasingFunc)
```

**Methods:**
- `Start()` — Start the animation
- `Stop()` — Stop the animation
- `Pause()` — Pause the animation
- `Resume()` — Resume the animation
- `OnUpdate(fn func(progress float64))` — Set update callback
- `OnComplete(fn func())` — Set complete callback

### Animator

Animation manager.

```go
animator := animation.NewAnimator()
```

**Methods:**
- `Add(animation *Animation)` — Add animation
- `Remove(animation *Animation)` — Remove animation
- `Update(delta float64)` — Update all animations
- `StopAll()` — Stop all animations
- `PauseAll()` — Pause all animations
- `ResumeAll()` — Resume all animations

### Timeline

Timeline for sequenced animations.

```go
timeline := animation.NewTimeline()
```

**Methods:**
- `Add(entry TimelineEntry)` — Add entry
- `Play()` — Play the timeline
- `Pause()` — Pause the timeline
- `Stop()` — Stop the timeline
- `Seek(time float64)` — Seek to time

### Easing Functions

| Function | Description |
|---|---|
| `Linear` | Linear easing |
| `EaseInQuad` | Quadratic ease in |
| `EaseOutQuad` | Quadratic ease out |
| `EaseInOutQuad` | Quadratic ease in-out |
| `EaseInCubic` | Cubic ease in |
| `EaseOutCubic` | Cubic ease out |
| `EaseInOutCubic` | Cubic ease in-out |
| `EaseInQuart` | Quartic ease in |
| `EaseOutQuart` | Quartic ease out |
| `EaseInOutQuart` | Quartic ease in-out |
| `EaseInQuint` | Quintic ease in |
| `EaseOutQuint` | Quintic ease out |
| `EaseInOutQuint` | Quintic ease in-out |
| `EaseInSine` | Sine ease in |
| `EaseOutSine` | Sine ease out |
| `EaseInOutSine` | Sine ease in-out |
| `EaseInExpo` | Exponential ease in |
| `EaseOutExpo` | Exponential ease out |
| `EaseInOutExpo` | Exponential ease in-out |
| `EaseInCirc` | Circular ease in |
| `EaseOutCirc` | Circular ease out |
| `EaseInOutCirc` | Circular ease in-out |
| `EaseInBounce` | Bounce ease in |
| `EaseOutBounce` | Bounce ease out |
| `EaseInOutBounce` | Bounce ease in-out |
| `EaseInBack` | Back ease in |
| `EaseOutBack` | Back ease out |
| `EaseInOutBack` | Back ease in-out |
| `EaseInElastic` | Elastic ease in |
| `EaseOutElastic` | Elastic ease out |
| `EaseInOutElastic` | Elastic ease in-out |

## Plugin Package

### Manager

Plugin manager.

```go
manager := plugin.NewManager()
```

**Methods:**
- `Register(plugin Plugin)` — Register a plugin
- `Unregister(name string)` — Unregister a plugin
- `Enable(name string)` — Enable a plugin
- `Disable(name string)` — Disable a plugin
- `Get(name string) Plugin` — Get a plugin
- `List() []Plugin` — List all plugins
- `InitAll()` — Initialize all plugins
- `ShutdownAll()` — Shutdown all plugins

### Plugin Interface

```go
type Plugin interface {
    Name() string
    Version() string
    Init(ctx *Context) error
    Shutdown() error
}
```

### Context

Plugin context.

```go
ctx := plugin.NewContext()
```

**Methods:**
- `RegisterRenderable(name string, factory RenderableFactory)` — Register renderable
- `RegisterCommand(name string, handler CommandHandler)` — Register command
- `OnEvent(event string, handler EventHandler)` — Register event handler
- `RegisterHook(name string, hook HookFunc, priority int)` — Register hook
- `EmitEvent(event string, args ...interface{})` — Emit event

## QRCode Package

### QRCode

QR code generation.

```go
qr, err := qrcode.NewQRCode(text, qrcode.LevelM)
```

**Methods:**
- `Render() string` — Render as text
- `SavePNG(path string) error` — Save as PNG
- `SavePNGModule(path string, border, moduleSize int) error` — Save with options

### Error Correction Levels

| Level | Recovery |
|---|---|
| `LevelL` | ~7% |
| `LevelM` | ~15% |
| `LevelQ` | ~25% |
| `LevelH` | ~30% |

### Encoding

```go
segments := qrcode.EncodeBytes(data)
```

### SVG Rendering

```go
svg := qrcode.RenderSVG(qr)
```

## Native Package

### Terminal

Terminal I/O handling.

```go
terminal := native.NewTerminal(stdoutFd)
```

**Methods:**
- `GetDimensions() (int, int)` — Get terminal dimensions
- `SetDimensions(width, height int)` — Set terminal dimensions
- `WatchResize(callback func(ResizeEvent))` — Watch for resize events
- `StopResizeWatch()` — Stop watching for resize
- `Close()` — Close the terminal

### Yoga

Flexbox layout engine.

```go
node := native.NewYogaNode()
```

**Methods:**
- `SetWidth(width float64)` — Set width
- `SetHeight(height float64)` — Set height
- `SetFlexDirection(dir FlexDirection)` — Set flex direction
- `SetJustifyContent(justify JustifyContent)` — Set justify content
- `SetAlignItems(align AlignItems)` — Set align items
- `CalculateLayout(width, height float64)` — Calculate layout

### Buffer

Optimized buffer for cell-based rendering.

```go
buffer := native.NewOptimizedBuffer(width, height)
```

**Methods:**
- `SetCell(x, y int, char rune, fg, bg uint16, attrs uint32)` — Set a cell
- `Clear()` — Clear the buffer
- `FillRect(x, y, width, height int, char rune, fg, bg uint16, attrs uint32)` — Fill a rectangle
- `Access() BufferAccess` — Get buffer access
- `Resize(width, height int)` — Resize the buffer
- `Destroy()` — Destroy the buffer

### Image

Image loading and rendering.

```go
img, err := native.LoadImage(path)
```

### Video

Video decoding and rendering.

```go
video := native.NewVideo(width, height)
```

### Audio

Audio engine for sound generation.

```go
audio := native.NewAudioEngine(sampleRate int)
```

### Clipboard

Clipboard operations.

```go
clipboard := native.NewClipboard()
```

### Input Parser

Parse terminal input events.

```go
parser := native.NewInputParser()
```

### Capabilities

Terminal capability detection.

```go
caps := native.NewTerminalCapabilities()
```

### Syntax Style

Syntax highlighting styles.

```go
style := native.NewSyntaxStyle()
```

### Event Emitter

Event handling system.

```go
emitter := native.NewEventEmitter()
```

### File Logger

File-based logging.

```go
logger := native.NewFileLogger(path string)
```

### Mem Registry

Memory tracking.

```go
registry := native.NewMemRegistry()
```

### Native Renderable

Native renderable wrapper.

```go
nr := native.NewNativeRenderable()
```

### Tree Sitter

Tree-sitter syntax highlighting.

```go
lang := native.NewTreeSitterLanguage(language)
```

### Embedded Terminal

Embedded terminal emulator.

```go
terminal := native.NewEmbeddedTerminal()
```

### Split Scrollback

Split scrollback buffer.

```go
scrollback := native.NewSplitScrollback()
```

### Stdout Writer

Stdout writer with buffering.

```go
writer := native.NewStdoutWriter()
```

### Native Span Feed

Native span feed for text rendering.

```go
feed := native.NewNativeSpanFeed()
```

### Encoded Unicode

Encoded unicode handling.

```go
encoded := native.NewEncodedUnicode()
```

### Link

Terminal hyperlink support.

```go
link := native.NewLink(url, text string)
```

### Kitty Transport

Kitty graphics protocol transport.

```go
transport := native.NewKittyTransport()
```

### Editor View

Editor view with syntax highlighting.

```go
editor := native.NewEditorView()
```

### Text Buffer

Text buffer with rope data structure.

```go
buffer := native.NewTextBuffer()
```

### Text Buffer View

Text buffer view with scrolling.

```go
view := native.NewTextBufferView(buffer)
```

### Edit Buffer

Edit buffer with undo/redo.

```go
editBuffer := native.NewEditBuffer()
```

### Rope

Rope data structure for text editing.

```go
rope := native.NewRope()
```

### Ghostty VT

VT terminal emulator.

```go
vt := native.NewVTTerminal(width, height)
```

### Lib

Entry point module.

```go
native.SetIO(writer)
```

### Opentui

Core module aggregating public API.

```go
// Compile-time verification of all public API
```