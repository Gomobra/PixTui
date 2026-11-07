package renderables

import (
	"github.com/Gomobra/PixTui/packages/core"
	"github.com/Gomobra/PixTui/packages/native"
)

// SelectOptions contains options for SelectRenderable.
type SelectOptions struct {
	core.RenderableOptions[core.Renderable]
	Options   []string
	Selected  int
}

// SelectRenderable is a dropdown/select list.
type SelectRenderable struct {
	*core.Renderable
	options     SelectOptions
	scrollOffset int
}

// NewSelectRenderable creates a new SelectRenderable.
func NewSelectRenderable(ctx *core.RenderContext, opts SelectOptions) *SelectRenderable {
	base := core.NewRenderable(ctx, opts.ID, core.RenderableOptions[core.Renderable]{
		Width:   opts.Width,
		Height:  opts.Height,
		Visible: opts.Visible,
		ZIndex:  opts.ZIndex,
	})
	r := &SelectRenderable{
		Renderable: base,
		options:    opts,
	}
	// Clamp selected index
	if len(r.options.Options) > 0 {
		if r.options.Selected < 0 {
			r.options.Selected = 0
		}
		if r.options.Selected >= len(r.options.Options) {
			r.options.Selected = len(r.options.Options) - 1
		}
	} else {
		r.options.Selected = 0
	}
	// Create native scene node
	if ctx != nil && ctx.NativeScene != nil {
		node, err := ctx.NativeScene.CreateNode(r.Renderable, native.SceneNodeKindText)
		if err == nil {
			r.Renderable.SetNativeNode(node)
		}
	}
	return r
}

// Selected returns the currently selected index.
func (r *SelectRenderable) Selected() int {
	return r.options.Selected
}

// SetSelected sets the selected index with clamping and auto-scroll.
func (r *SelectRenderable) SetSelected(idx int) {
	if len(r.options.Options) == 0 {
		r.options.Selected = 0
		r.RequestRender()
		return
	}
	r.options.Selected = clampInt(idx, 0, len(r.options.Options)-1)
	// Auto-scroll to keep selected item visible.
	height := int(r.GetHeight())
	if height < 1 {
		height = 1
	}
	if r.options.Selected < r.scrollOffset {
		r.scrollOffset = r.options.Selected
	} else if r.options.Selected >= r.scrollOffset+height {
		r.scrollOffset = r.options.Selected - height + 1
	}
	r.RequestRender()
}

// ScrollUp scrolls the list up by n items.
func (r *SelectRenderable) ScrollUp(n int) {
	r.scrollOffset = clampInt(r.scrollOffset-n, 0, len(r.options.Options)-1)
	r.RequestRender()
}

// ScrollDown scrolls the list down by n items.
func (r *SelectRenderable) ScrollDown(n int) {
	maxScroll := len(r.options.Options) - 1
	if maxScroll < 0 {
		maxScroll = 0
	}
	r.scrollOffset = clampInt(r.scrollOffset+n, 0, maxScroll)
	r.RequestRender()
}

// HandleKeyDown handles keyboard navigation for the select.
func (r *SelectRenderable) HandleKeyDown(key *core.KeyEvent) bool {
	if len(r.options.Options) == 0 {
		return true
	}
	switch key.Key {
	case "ArrowUp", "k":
		r.SetSelected(r.options.Selected - 1)
		return true
	case "ArrowDown", "j":
		r.SetSelected(r.options.Selected + 1)
		return true
	case "Enter", "Return", "Linefeed":
		return true
	}
	return false
}

// RenderSelf renders the select into the buffer, showing only visible options.
func (r *SelectRenderable) RenderSelf(buffer *core.NativeBuffer, deltaTime float64) {
	height := int(r.GetHeight())
	width := int(r.GetWidth())

	if height < 1 {
		height = 1
	}

	// Clamp scrollOffset to valid range.
	maxScroll := len(r.options.Options) - height
	if maxScroll < 0 {
		maxScroll = 0
	}
	r.scrollOffset = clampInt(r.scrollOffset, 0, maxScroll)

	selBg := &core.RGBA{R: 60, G: 60, B: 100, A: 255}
	selFg := &core.RGBA{R: 255, G: 255, B: 255, A: 255}

	fg := r.options.ForegroundColor
	bg := r.options.BackgroundColor
	attrs := r.options.Attributes

	for i := 0; i < height; i++ {
		optIdx := r.scrollOffset + i
		if optIdx >= len(r.options.Options) {
			break
		}
		opt := r.options.Options[optIdx]

		isSelected := optIdx == r.options.Selected
		prefix := "  "
		if isSelected {
			prefix = "> "
		}

		x := 0
		cellFg := fg
		cellBg := bg
		cellAttrs := attrs
		if isSelected {
			cellBg = selBg
			cellFg = selFg
		}

		for _, ch := range prefix + opt {
			if x >= width {
				break
			}
			writeCell(buffer, x, i, uint32(ch), cellFg, cellBg, cellAttrs)
			x++
		}
	}

	if r.options.RenderAfter != nil {
		r.options.RenderAfter(r.Renderable, buffer, deltaTime)
	}
}
