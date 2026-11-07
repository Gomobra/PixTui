package react

import (
	"testing"

	"github.com/Gomobra/PixTui/packages/core"
	coretesting "github.com/Gomobra/PixTui/packages/core/testing"
)

func TestTestRender(t *testing.T) {
	tr := TestRender(CreateElement("box", map[string]interface{}{
		"border": true,
		"title":  "Test",
	}), TestRendererOptions{Width: 80, Height: 24})
	if tr == nil {
		t.Fatal("expected non-nil test renderer")
	}
	if tr.React == nil {
		t.Error("expected non-nil React renderer")
	}
	_ = tr.Capture()
	_ = tr.Dispose()
	var _ *coretesting.TestRenderer = tr.TestRenderer
}

func TestTestRenderDefaultDimensions(t *testing.T) {
	tr := TestRender(CreateElement("text", map[string]interface{}{
		"content": "hello",
	}), TestRendererOptions{})
	if tr == nil {
		t.Fatal("expected non-nil test renderer")
	}
	_ = tr.Capture()
	_ = tr.Dispose()
}

func TestTestRenderRendersElement(t *testing.T) {
	tr := TestRender(CreateElement("box", map[string]interface{}{
		"id":   "rendered-box",
		"border": true,
	}), TestRendererOptions{Width: 80, Height: 24})
	_ = tr
	// Should not panic
}

func TestAssertNoPanic(t *testing.T) {
	AssertNoPanic(t, "no panic", func() {
		// no-op
	})
}



var _ = core.RenderContext{}
var _ = coretesting.NewTestRenderer
