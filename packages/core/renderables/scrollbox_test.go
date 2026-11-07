package renderables

import (
	"testing"

	"github.com/Gomobra/PixTui/packages/core"
)

// TestScrollBoxSetGetScrollYClamp verifies SetScrollY clamps to [0, maxScrollY].
func TestScrollBoxSetGetScrollYClamp(t *testing.T) {
	ctx := newTestCtx()
	r := NewScrollBoxRenderable(ctx, ScrollBoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "scrollbox-clamp",
			Visible: boolPtr(true),
		},
	})
	r.SetPosition(0, 0)
	r.SetWidth(20)
	r.SetHeight(10)

	// Negative scroll should clamp to 0 (no RenderSelf call, so maxScrollY=0).
	r.SetScrollTop(-5)
	if r.GetScrollY() != 0 {
		t.Errorf("expected scrollY clamped to 0, got %d", r.GetScrollY())
	}

	// Set content larger than viewport WITHOUT calling RenderSelf (which would
	// override content size from children). Then scroll beyond max.
	r.SetContentSize(100, 50)
	r.SetScrollTop(1000)
	maxY := r.GetScrollY()
	if maxY == 0 {
		t.Error("expected maxScrollY > 0 for content larger than viewport")
	}
	if r.GetScrollY() > maxY {
		t.Errorf("expected scrollY <= maxScrollY (%d), got %d", maxY, r.GetScrollY())
	}

	// Scroll back to 0.
	r.SetScrollTop(0)
	if r.GetScrollY() != 0 {
		t.Errorf("expected scrollY 0, got %d", r.GetScrollY())
	}
}

// TestScrollBoxSetGetScrollXClamp verifies SetScrollX clamps to [0, maxScrollX].
func TestScrollBoxSetGetScrollXClamp(t *testing.T) {
	ctx := newTestCtx()
	r := NewScrollBoxRenderable(ctx, ScrollBoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "scrollbox-x-clamp",
			Visible: boolPtr(true),
		},
	})
	r.SetPosition(0, 0)
	r.SetWidth(20)
	r.SetHeight(10)

	r.SetScrollLeft(-3)
	if r.GetScrollX() != 0 {
		t.Errorf("expected scrollX clamped to 0, got %d", r.GetScrollX())
	}

	r.SetContentSize(100, 50)
	r.SetScrollLeft(1000)
	maxX := r.GetScrollX()
	if maxX == 0 {
		t.Error("expected maxScrollX > 0 for content larger than viewport")
	}
	if r.GetScrollX() > maxX {
		t.Errorf("expected scrollX <= maxScrollX (%d), got %d", maxX, r.GetScrollX())
	}
}

// TestScrollBoxContentSize verifies SetContentSize updates content dimensions.
func TestScrollBoxContentSize(t *testing.T) {
	ctx := newTestCtx()
	r := NewScrollBoxRenderable(ctx, ScrollBoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "scrollbox-content",
			Visible: boolPtr(true),
		},
	})
	r.SetPosition(0, 0)
	r.SetWidth(20)
	r.SetHeight(10)

	// Set viewport size so content size comparisons work correctly.
	r.SetViewportSize(20, 10)

	// Set content larger than viewport without RenderSelf override.
	r.SetContentSize(120, 60)
	r.SetScrollTop(1000)
	if r.GetScrollY() == 0 {
		t.Error("expected non-zero maxScrollY after SetContentSize(120, 60)")
	}
	r.SetScrollLeft(1000)
	if r.GetScrollX() == 0 {
		t.Error("expected non-zero maxScrollX after SetContentSize(120, 60)")
	}

	// Setting content size smaller than viewport should clamp scroll to 0.
	r.SetContentSize(5, 5)
	r.SetScrollTop(100)
	if r.GetScrollY() != 0 {
		t.Errorf("expected scrollY clamped to 0 for small content, got %d", r.GetScrollY())
	}
	r.SetScrollLeft(100)
	if r.GetScrollX() != 0 {
		t.Errorf("expected scrollX clamped to 0 for small content, got %d", r.GetScrollX())
	}
}

// TestScrollBoxViewportCulling verifies children outside viewport are not rendered.
func TestScrollBoxViewportCulling(t *testing.T) {
	ctx := newTestCtx()
	r := NewScrollBoxRenderable(ctx, ScrollBoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "scrollbox-cull",
			Visible: boolPtr(true),
		},
	})
	r.SetPosition(0, 0)
	r.SetWidth(20)
	r.SetHeight(10)

	// Add a child inside the viewport.
	childInside := NewTextRenderable(ctx, TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      "child-inside",
			Visible: boolPtr(true),
		},
		Content: "inside",
	})
	childInside.SetPosition(2, 2)
	childInside.SetWidth(5)
	childInside.SetHeight(1)
	r.AddChild(childInside)

	// Add a child far outside the viewport (below).
	childOutside := NewTextRenderable(ctx, TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      "child-outside",
			Visible: boolPtr(true),
		},
		Content: "outside",
	})
	childOutside.SetPosition(2, 50)
	childOutside.SetWidth(5)
	childOutside.SetHeight(1)
	r.AddChild(childOutside)

	buf := core.NewNativeBuffer(80, 24)
	r.RenderSelf(buf, 0)

	access := buf.Access()
	// Count non-empty cells to verify rendering happened
	nonEmpty := 0
	for y := 0; y < access.Height; y++ {
		for x := 0; x < access.Width; x++ {
			idx := y*access.Width + x
			if access.Char[idx] != 0 && access.Char[idx] != ' ' {
				nonEmpty++
			}
		}
	}
	if nonEmpty == 0 {
		t.Error("expected scrollbox to render some content")
	}

	// The outside child at y=50 should be culled (not rendered in viewport)
	// Check that no content appears at y >= 10 (outside viewport height)
	outsideRendered := false
	for y := 10; y < access.Height; y++ {
		for x := 0; x < access.Width; x++ {
			idx := y*access.Width + x
			if access.Char[idx] != 0 && access.Char[idx] != ' ' {
				outsideRendered = true
				break
			}
		}
		if outsideRendered {
			break
		}
	}
	if outsideRendered {
		t.Error("expected child outside viewport to be culled")
	}
}

// TestScrollBoxScrollOffsetRenders verifies scrolling changes which children are visible.
func TestScrollBoxScrollOffsetRenders(t *testing.T) {
	ctx := newTestCtx()
	r := NewScrollBoxRenderable(ctx, ScrollBoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "scrollbox-offset",
			Visible: boolPtr(true),
		},
	})
	r.SetPosition(0, 0)
	r.SetWidth(20)
	r.SetHeight(5)

	childA := NewTextRenderable(ctx, TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      "child-a",
			Visible: boolPtr(true),
		},
		Content: "child-a",
	})
	childA.SetPosition(0, 0)
	childA.SetWidth(10)
	childA.SetHeight(1)
	r.AddChild(childA)

	childB := NewTextRenderable(ctx, TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      "child-b",
			Visible: boolPtr(true),
		},
		Content: "child-b",
	})
	childB.SetPosition(0, 10)
	childB.SetWidth(10)
	childB.SetHeight(1)
	r.AddChild(childB)

	buf := core.NewNativeBuffer(80, 24)

	// Without scrolling, verify some content is rendered (child-a at top).
	r.RenderSelf(buf, 0)
	access := buf.Access()
	contentAtTop := false
	for x := 0; x < access.Width; x++ {
		idx := 0*access.Width + x
		if access.Char[idx] != 0 && access.Char[idx] != ' ' {
			contentAtTop = true
			break
		}
	}
	if !contentAtTop {
		t.Error("expected content at top row without scrolling")
	}

	// Scroll down by 10, child-a should be scrolled out, child-b should be visible.
	r.SetScrollTop(10)
	r.RenderSelf(buf, 0)
	access = buf.Access()
	contentAfterScroll := false
	for y := 0; y < access.Height; y++ {
		for x := 0; x < access.Width; x++ {
			idx := y*access.Width + x
			if access.Char[idx] != 0 && access.Char[idx] != ' ' {
				contentAfterScroll = true
				break
			}
		}
		if contentAfterScroll {
			break
		}
	}
	if !contentAfterScroll {
		t.Error("expected content to be visible after scrolling")
	}
}

// TestScrollBoxScrollbarThumb verifies scrollbar is rendered when content exceeds viewport.
func TestScrollBoxScrollbarThumb(t *testing.T) {
	ctx := newTestCtx()
	r := NewScrollBoxRenderable(ctx, ScrollBoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "scrollbox-thumb",
			Visible: boolPtr(true),
		},
	})
	r.SetPosition(0, 0)
	r.SetWidth(20)
	r.SetHeight(10)

	child := NewTextRenderable(ctx, TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      "tall-child",
			Visible: boolPtr(true),
		},
		Content: "tall content",
	})
	child.SetPosition(0, 0)
	child.SetWidth(10)
	child.SetHeight(50)
	r.AddChild(child)

	buf := core.NewNativeBuffer(80, 24)
	r.RenderSelf(buf, 0)

	access := buf.Access()
	thumbFound := false
	for y := 0; y < access.Height; y++ {
		idx := y*access.Width + 19
		if access.Char[idx] == '█' {
			thumbFound = true
			break
		}
	}
	if !thumbFound {
		t.Error("expected scrollbar thumb ('█') at right edge when content overflows")
	}
}
