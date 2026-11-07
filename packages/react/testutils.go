package react

import (
	"testing"

	coretesting "github.com/Gomobra/PixTui/packages/core/testing"
)

// TestRendererOptions configures a test renderer created by TestRender.
type TestRendererOptions struct {
	Width  int
	Height int
}

// TestRenderer wraps a core testing.TestRenderer together with the React
// renderer used to mount a React element tree. It mirrors the
// @pixtui/react test-utils `testRender` helper.
type TestRenderer struct {
	*coretesting.TestRenderer
	React *ReactRenderer
	Root  ReactElement
}

// TestRender mounts a React element tree into a test renderer and returns a
// TestRenderer that can be used to inspect the rendered output. It is the Go
// equivalent of the TypeScript `testRender` from @pixtui/react/test-utils.
// The React renderer paints directly into the test renderer's buffer, so no
// separate CliRenderer commit is required.
func TestRender(element ReactElement, opts TestRendererOptions) *TestRenderer {
	width := opts.Width
	if width == 0 {
		width = 80
	}
	height := opts.Height
	if height == 0 {
		height = 24
	}

	tr := coretesting.NewTestRenderer(width, height)
	ctx := tr.Context()
	renderer := NewReactRenderer(ctx)

	buffer := tr.Buffer()
	renderer.Render(element, buffer, 0)

	return &TestRenderer{
		TestRenderer: tr,
		React:        renderer,
		Root:         element,
	}
}

// Capture returns the current buffer contents as a string of characters with
// newlines between rows.
func (tr *TestRenderer) Capture() string {
	return tr.TestRenderer.CaptureCharFrame()
}

// Dispose releases all resources held by the test renderer.
func (tr *TestRenderer) Dispose() error {
	return tr.TestRenderer.Dispose()
}

// Compile-time assertion that TestRenderer implements disposal.
var _ interface {
	Dispose() error
} = (*TestRenderer)(nil)

// AssertNoPanic is a small helper that runs a function and fails the test if
// it panics, mirroring the error-boundary behavior expected from test utils.
func AssertNoPanic(t *testing.T, name string, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("%s panicked: %v", name, r)
		}
	}()
	fn()
}
