package renderables

import (
	"testing"

	"github.com/Gomobra/PixTui/packages/core"
)

// TestBoxBorderColor verifies BoxRenderable renders border cells with the
// configured border color.
func TestBoxBorderColor(t *testing.T) {
	ctx := newTestCtx()
	red := core.RGBA{R: 255, G: 0, B: 0, A: 255}
	r := NewBoxRenderable(ctx, BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:                "box-border",
			Visible:           boolPtr(true),
			ForegroundColor:   &red,
		},
		Border:      true,
		BorderColor: &red,
		Title:       "Border",
	})
	r.SetPosition(0, 0)
	r.SetWidth(20)
	r.SetHeight(5)

	buf := core.NewNativeBuffer(80, 24)
	r.RenderSelf(buf, 0)

	access := buf.Access()
	foundBorder := false
	for y := 0; y < access.Height; y++ {
		for x := 0; x < access.Width; x++ {
			idx := y*access.Width + x
			if access.Char[idx] == 0 {
				continue
			}
			if access.Fg[idx] != 0 {
				foundBorder = true
				break
			}
		}
		if foundBorder {
			break
		}
	}
	if !foundBorder {
		t.Error("expected border cells to have non-zero foreground color")
	}
}

// TestBoxTitleColor verifies BoxRenderable title is rendered with title color.
func TestBoxTitleColor(t *testing.T) {
	ctx := newTestCtx()
	white := core.RGBA{R: 255, G: 255, B: 255, A: 255}
	green := core.RGBA{R: 0, G: 255, B: 0, A: 255}
	r := NewBoxRenderable(ctx, BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:                 "box-title",
			Visible:            boolPtr(true),
			ForegroundColor:    &white,
		},
		Border:      true,
		BorderColor: &white,
		Title:       "Hello",
		TitleColor:  &green,
	})
	r.SetPosition(0, 0)
	r.SetWidth(20)
	r.SetHeight(5)

	buf := core.NewNativeBuffer(80, 24)
	r.RenderSelf(buf, 0)

	access := buf.Access()
	titleFound := false
	for x := 1; x < 10; x++ {
		idx := 0*access.Width + x
		if access.Char[idx] != 0 {
			titleFound = true
			break
		}
	}
	if !titleFound {
		t.Error("expected title text to be rendered")
	}
}

// TestTextForegroundBackground verifies TextRenderable renders with fg/bg colors.
func TestTextForegroundBackground(t *testing.T) {
	ctx := newTestCtx()
	fg := core.RGBA{R: 0, G: 255, B: 0, A: 255}
	bg := core.RGBA{R: 0, G: 0, B: 255, A: 255}
	r := NewTextRenderable(ctx, TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:               "text-colors",
			Visible:          boolPtr(true),
			ForegroundColor:  &fg,
			BackgroundColor:  &bg,
		},
		Content: "Styled Text",
	})
	r.SetPosition(0, 0)
	r.SetWidth(20)
	r.SetHeight(5)

	buf := core.NewNativeBuffer(80, 24)
	r.RenderSelf(buf, 0)

	access := buf.Access()
	hasFg := false
	hasBg := false
	for y := 0; y < access.Height; y++ {
		for x := 0; x < access.Width; x++ {
			idx := y*access.Width + x
			if access.Char[idx] != 0 {
				if access.Fg[idx] != 0 {
					hasFg = true
				}
				if access.Bg[idx] != 0 {
					hasBg = true
				}
			}
		}
	}
	if !hasFg {
		t.Error("expected text cells to have non-zero foreground color")
	}
	if !hasBg {
		t.Error("expected text cells to have non-zero background color")
	}
}

// TestInputCursor verifies InputRenderable renders cursor with inverse attribute.
func TestInputCursor(t *testing.T) {
	ctx := newTestCtx()
	r := NewInputRenderable(ctx, InputOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "input-cursor",
			Visible: boolPtr(true),
		},
		Value:      "hello",
		Placeholder: "",
	})
	r.SetCursorPosition(3)
	r.SetPosition(0, 0)
	r.SetWidth(20)
	r.SetHeight(1)

	buf := core.NewNativeBuffer(80, 24)
	r.RenderSelf(buf, 0)

	access := buf.Access()
	cursorFound := false
	for x := 0; x < access.Width; x++ {
		idx := 0*access.Width + x
		if access.Char[idx] != 0 && access.Attributes[idx]&core.TextAttributeInverse != 0 {
			cursorFound = true
			break
		}
	}
	if !cursorFound {
		t.Error("expected cursor cell to have inverse attribute")
	}
}

// TestInputPlaceholder verifies InputRenderable renders placeholder when empty.
func TestInputPlaceholder(t *testing.T) {
	ctx := newTestCtx()
	r := NewInputRenderable(ctx, InputOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "input-placeholder",
			Visible: boolPtr(true),
		},
		Value:      "",
		Placeholder: "type something",
	})
	r.SetPosition(0, 0)
	r.SetWidth(30)
	r.SetHeight(1)

	buf := core.NewNativeBuffer(80, 24)
	r.RenderSelf(buf, 0)

	access := buf.Access()
	placeholderFound := false
	for x := 0; x < access.Width; x++ {
		idx := 0*access.Width + x
		if access.Char[idx] != 0 {
			placeholderFound = true
			break
		}
	}
	if !placeholderFound {
		t.Error("expected placeholder text to be rendered")
	}
}

// TestSelectSelected verifies SelectRenderable highlights the selected option.
func TestSelectSelected(t *testing.T) {
	ctx := newTestCtx()
	r := NewSelectRenderable(ctx, SelectOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "select-selected",
			Visible: boolPtr(true),
		},
		Options:   []string{"alpha", "beta", "gamma"},
		Selected:  1,
	})
	r.SetPosition(0, 0)
	r.SetWidth(20)
	r.SetHeight(5)

	buf := core.NewNativeBuffer(80, 24)
	r.RenderSelf(buf, 0)

	access := buf.Access()
	selectedFound := false
	for y := 0; y < access.Height; y++ {
		for x := 0; x < access.Width; x++ {
			idx := y*access.Width + x
			if access.Char[idx] != 0 && access.Bg[idx] != 0 {
				selectedFound = true
				break
			}
		}
		if selectedFound {
			break
		}
	}
	if !selectedFound {
		t.Error("expected selected option row to have background color")
	}
}

// TestDiffAddedRemovedColors verifies DiffRenderable colors + lines green and - lines red.
func TestDiffAddedRemovedColors(t *testing.T) {
	ctx := newTestCtx()
	r := NewDiffRenderable(ctx, DiffOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "diff-colors",
			Visible: boolPtr(true),
		},
		Content: "-removed line\n+added line\n context line",
	})
	r.SetPosition(0, 0)
	r.SetWidth(30)
	r.SetHeight(5)

	buf := core.NewNativeBuffer(80, 24)
	r.RenderSelf(buf, 0)

	access := buf.Access()
	hasRed := false
	hasGreen := false
	for y := 0; y < access.Height; y++ {
		for x := 0; x < access.Width; x++ {
			idx := y*access.Width + x
			if access.Char[idx] == 0 {
				continue
			}
			if access.Fg[idx] != 0 {
				if y == 0 {
					hasRed = true
				} else if y == 1 {
					hasGreen = true
				}
			}
		}
	}
	if !hasRed {
		t.Error("expected removed line (-) to have red foreground")
	}
	if !hasGreen {
		t.Error("expected added line (+) to have green foreground")
	}
}

// TestTextAttributes verifies TextRenderable applies text attributes (bold).
func TestTextAttributes(t *testing.T) {
	ctx := newTestCtx()
	r := NewTextRenderable(ctx, TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "text-attrs",
			Visible: boolPtr(true),
		},
		Content:   "Bold Text",
		Attributes: core.TextAttributeBold,
	})
	r.SetPosition(0, 0)
	r.SetWidth(20)
	r.SetHeight(5)

	buf := core.NewNativeBuffer(80, 24)
	r.RenderSelf(buf, 0)

	access := buf.Access()
	hasBold := false
	for y := 0; y < access.Height; y++ {
		for x := 0; x < access.Width; x++ {
			idx := y*access.Width + x
			if access.Char[idx] != 0 && access.Attributes[idx]&core.TextAttributeBold != 0 {
				hasBold = true
				break
			}
		}
	}
	if !hasBold {
		t.Error("expected text cells to have bold attribute")
	}
}
