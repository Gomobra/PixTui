package testing

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Gomobra/PixTui/packages/core"
	"github.com/Gomobra/PixTui/packages/native"
)

// TestRendererSetup holds the components of a test renderer environment.
type TestRendererSetup struct {
	Renderer  *core.CliRenderer
	MockInput *MockKeys
	MockMouse *MockMouse
	Context   *core.RenderContext
	Buffer    *core.NativeBuffer
}

// ExternalOutputCommit represents a single external output commit.
type ExternalOutputCommit struct {
	Text           string
	Rows           []string
	Width          int
	Height         int
	RowColumns     int
	StartOnNewLine bool
	TrailingNewline bool
}

// ExternalOutputRecorder records external output from the renderer.
type ExternalOutputRecorder struct {
	mu      sync.Mutex
	commits []ExternalOutputCommit
}

func newExternalOutputRecorder() *ExternalOutputRecorder {
	return &ExternalOutputRecorder{
		commits: make([]ExternalOutputCommit, 0),
	}
}

func (r *ExternalOutputRecorder) record(text string, rows []string, width, height, rowColumns int, startOnNewLine, trailingNewline bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.commits = append(r.commits, ExternalOutputCommit{
		Text:           text,
		Rows:           rows,
		Width:          width,
		Height:         height,
		RowColumns:     rowColumns,
		StartOnNewLine: startOnNewLine,
		TrailingNewline: trailingNewline,
	})
}

// Take returns and clears all recorded commits.
func (r *ExternalOutputRecorder) Take() []ExternalOutputCommit {
	r.mu.Lock()
	defer r.mu.Unlock()
	commits := r.commits
	r.commits = make([]ExternalOutputCommit, 0)
	return commits
}

// TakeText returns all recorded output as a single string and clears the buffer.
func (r *ExternalOutputRecorder) TakeText() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var sb strings.Builder
	for i, c := range r.commits {
		if i > 0 {
			sb.WriteString("\n")
		}
		for j, row := range c.Rows {
			if j > 0 {
				sb.WriteString("\n")
			}
			sb.WriteString(row)
		}
	}
	r.commits = make([]ExternalOutputCommit, 0)
	return sb.String()
}

// Clear removes all recorded commits.
func (r *ExternalOutputRecorder) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.commits = make([]ExternalOutputCommit, 0)
}

// NativeRenderStats represents statistics from the native renderer.
type NativeRenderStats struct {
	FrameCount   uint64
	CellsUpdated int64
}

// TestRenderer provides a test environment backed by a CliRenderer with mock input and mouse.
type TestRenderer struct {
	mu   sync.Mutex
	setup *TestRendererSetup

	mockKeys       *MockKeys
	mockMouse      *MockMouse
	externalOutput *ExternalOutputRecorder
	keyHandler     func([]byte)
	disposed       bool
	width          int
	height         int
	frameID        uint64
}

// NewTestRenderer creates a new TestRenderer with the given dimensions.
func NewTestRenderer(width, height int) *TestRenderer {
	renderer := core.NewCliRenderer(width, height)
	context := renderer.Context()
	buffer := renderer.Buffer()
	externalOutput := newExternalOutputRecorder()

	tr := &TestRenderer{
		setup: &TestRendererSetup{
			Renderer: renderer,
			Context:  context,
			Buffer:   buffer,
		},
		externalOutput: externalOutput,
		width:          width,
		height:         height,
	}

	tr.mockKeys = NewMockKeys(tr)
	tr.mockMouse = NewMockMouse(tr)

	tr.setup.MockInput = tr.mockKeys
	tr.setup.MockMouse = tr.mockMouse

	// Set up key handler to route mock input bytes.
	renderer.SetKeyHandler(func(data []byte) {
		tr.mu.Lock()
		handler := tr.keyHandler
		tr.mu.Unlock()
		if handler != nil {
			handler(data)
		}
	})

	return tr
}

// Setup returns the TestRendererSetup.
func (tr *TestRenderer) Setup() *TestRendererSetup {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	return tr.setup
}

// Renderer returns the underlying CliRenderer.
func (tr *TestRenderer) Renderer() *core.CliRenderer {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	return tr.setup.Renderer
}

// Context returns the RenderContext.
func (tr *TestRenderer) Context() *core.RenderContext {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	return tr.setup.Context
}

// Buffer returns the NativeBuffer.
func (tr *TestRenderer) Buffer() *core.NativeBuffer {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	return tr.setup.Buffer
}

// MockKeys returns the mock keyboard.
func (tr *TestRenderer) MockKeys() *MockKeys {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	return tr.mockKeys
}

// MockMouse returns the mock mouse.
func (tr *TestRenderer) MockMouse() *MockMouse {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	return tr.mockMouse
}

// ExternalOutput returns the external output recorder.
func (tr *TestRenderer) ExternalOutput() *ExternalOutputRecorder {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	return tr.externalOutput
}

// SetKeyHandler sets the key input handler on the underlying renderer and stores it for mock input feeding.
func (tr *TestRenderer) SetKeyHandler(handler func([]byte)) {
	tr.mu.Lock()
	tr.keyHandler = handler
	tr.mu.Unlock()
	tr.setup.Renderer.SetKeyHandler(handler)
}

// FeedInput feeds raw input bytes to the renderer's key handler.
func (tr *TestRenderer) FeedInput(data []byte) {
	tr.mu.Lock()
	handler := tr.keyHandler
	disposed := tr.disposed
	tr.mu.Unlock()
	if disposed {
		return
	}
	if handler != nil {
		handler(data)
	}
}

// RenderOnce performs a single render pass over the given root renderable.
func (tr *TestRenderer) RenderOnce(root *core.RootRenderable) {
	tr.mu.Lock()
	renderer := tr.setup.Renderer
	tr.mu.Unlock()

	if err := renderer.Render(root, 0.016); err != nil {
		panic(fmt.Sprintf("test renderer render failed: %v", err))
	}

	tr.mu.Lock()
	tr.frameID++
	tr.mu.Unlock()
}

// Flush flushes all pending effects. Rendering is synchronous in Go, so this is a no-op.
func (tr *TestRenderer) Flush() {
	// In Go, rendering is synchronous; there are no pending effects to flush.
}

// WaitFor waits for the condition to become true, polling every 5ms up to the timeout.
func (tr *TestRenderer) WaitFor(condition func() bool, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()

	for {
		if condition() {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for condition after %v", timeout)
		}
		<-ticker.C
	}
}

// WaitForFrame waits for the next frame to be rendered.
func (tr *TestRenderer) WaitForFrame() error {
	tr.mu.Lock()
	lastFrameID := tr.frameID
	tr.mu.Unlock()

	return tr.WaitFor(func() bool {
		tr.mu.Lock()
		current := tr.frameID
		tr.mu.Unlock()
		return current > lastFrameID
	}, 5*time.Second)
}

// WaitForVisualIdle waits until the visual output stops changing (buffer generation stable).
func (tr *TestRenderer) WaitForVisualIdle(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	tr.mu.Lock()
	lastGen := tr.setup.Buffer.Access().Generation
	tr.mu.Unlock()

	quietCount := 0
	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for visual idle after %v", timeout)
		}

		tr.mu.Lock()
		currentGen := tr.setup.Buffer.Access().Generation
		tr.mu.Unlock()

		if currentGen == lastGen {
			quietCount++
			if quietCount >= 2 {
				return nil
			}
		} else {
			quietCount = 0
			lastGen = currentGen
		}

		<-ticker.C
	}
}

// bufferToString converts BufferAccess to a string representation with newlines between rows.
func bufferToString(access native.BufferAccess) string {
	var sb strings.Builder
	for y := 0; y < access.Height; y++ {
		for x := 0; x < access.Width; x++ {
			idx := y*access.Width + x
			if idx < len(access.Char) {
				r := access.Char[idx]
				if r == 0 {
					sb.WriteRune(' ')
				} else {
					sb.WriteRune(rune(r))
				}
			} else {
				sb.WriteRune(' ')
			}
		}
		if y < access.Height-1 {
			sb.WriteRune('\n')
		}
	}
	return sb.String()
}

// CaptureCharFrame captures the current frame as a string of characters.
func (tr *TestRenderer) CaptureCharFrame() string {
	tr.mu.Lock()
	access := tr.setup.Buffer.Access()
	tr.mu.Unlock()

	return bufferToString(access)
}

// CaptureSpans captures spans from the current frame buffer.
func (tr *TestRenderer) CaptureSpans() *core.CapturedFrame {
	tr.mu.Lock()
	access := tr.setup.Buffer.Access()
	w, h := tr.setup.Renderer.Dimensions()
	tr.mu.Unlock()

	frame := &core.CapturedFrame{
		Cols:   w,
		Rows:   h,
		Cursor: [2]int{0, 0},
		Lines:  make([]core.CapturedLine, h),
	}

	for y := 0; y < h; y++ {
		line := core.CapturedLine{
			Spans: make([]core.CapturedSpan, w),
		}
		for x := 0; x < w; x++ {
			idx := y*w + x
			ch := access.Char[idx]
			if ch == 0 {
				ch = ' '
			}
			line.Spans[x] = core.CapturedSpan{
				Text:      string(rune(ch)),
				FG:        core.RGBA{},
				BG:        core.RGBA{},
				Attributes: access.Attributes[idx],
				Width:     1,
			}
		}
		frame.Lines[y] = line
	}

	return frame
}

// Resize resizes the renderer to the given dimensions.
func (tr *TestRenderer) Resize(width, height int) {
	tr.mu.Lock()
	tr.setup.Renderer.SetDimensions(width, height)
	tr.width = width
	tr.height = height
	tr.mu.Unlock()
}

// GetNativeStats returns statistics from the native renderer.
func (tr *TestRenderer) GetNativeStats() NativeRenderStats {
	tr.mu.Lock()
	stats := NativeRenderStats{
		FrameCount:   tr.frameID,
		CellsUpdated: 0,
	}
	tr.mu.Unlock()
	return stats
}

// Dispose releases all resources held by the test renderer.
func (tr *TestRenderer) Dispose() error {
	tr.mu.Lock()
	if tr.disposed {
		tr.mu.Unlock()
		return nil
	}
	tr.disposed = true
	renderer := tr.setup.Renderer
	tr.mu.Unlock()

	return renderer.Destroy()
}

// Close is an alias for Dispose.
func (tr *TestRenderer) Close() error {
	return tr.Dispose()
}
