package solid

import (
	"testing"

	coretesting "github.com/Gomobra/PixTui/packages/core/testing"
)

// TestRenderOptions configures a test renderer created by TestRender.
type TestRenderOptions struct {
	Width  int
	Height int
}

// TestRender mounts a Solid component tree into a test renderer and returns
// the underlying test renderer for inspection. It mirrors the TypeScript
// `testRender` exported from @pixtui/solid.
func TestRender(component SolidComponent, opts TestRenderOptions) *coretesting.TestRenderer {
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
	renderer := NewSolidRenderer(ctx)

	buffer := tr.Buffer()
	renderer.Render(component, buffer, 0)

	return tr
}

// CaptureCharFrame returns the current buffer contents as a string of
// characters with newlines between rows. Provided for convenience on the
// test renderer returned by TestRender.
func CaptureCharFrame(tr *coretesting.TestRenderer) string {
	return tr.CaptureCharFrame()
}

// AssertNoPanic runs a function and fails the test if it panics.
func AssertNoPanic(t *testing.T, name string, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("%s panicked: %v", name, r)
		}
	}()
	fn()
}


