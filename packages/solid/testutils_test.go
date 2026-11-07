package solid

import (
	"testing"
)

func TestSolidTestRender(t *testing.T) {
	tr := TestRender(Box(map[string]interface{}{
		"id":     "test-root",
		"border": true,
		"title":  "Root",
	}), TestRenderOptions{Width: 80, Height: 24})
	if tr == nil {
		t.Fatal("expected non-nil test renderer")
	}
	_ = CaptureCharFrame(tr)
}

func TestSolidTestRenderDefaultDimensions(t *testing.T) {
	tr := TestRender(Text(map[string]interface{}{
		"content": "hello solid",
	}), TestRenderOptions{})
	if tr == nil {
		t.Fatal("expected non-nil test renderer")
	}
	_ = CaptureCharFrame(tr)
}

func TestSolidTestRenderNested(t *testing.T) {
	tr := TestRender(Box(map[string]interface{}{
		"id":      "outer",
		"border":  true,
		"children": []SolidComponent{
			Text(map[string]interface{}{"content": "nested"}),
		},
	}), TestRenderOptions{Width: 80, Height: 24})
	if tr == nil {
		t.Fatal("expected non-nil test renderer")
	}
	_ = CaptureCharFrame(tr)
}

func TestSolidAssertNoPanic(t *testing.T) {
	AssertNoPanic(t, "no panic", func() {
		// no-op
	})
}
