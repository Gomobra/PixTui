package renderables

import (
	"github.com/Gomobra/PixTui/packages/core"
	"github.com/Gomobra/PixTui/packages/native"
)

// TabSelectOptions contains options for TabSelectRenderable
type TabSelectOptions struct {
	core.RenderableOptions[core.Renderable]
	Tabs     []string
	Selected int
	// TabColor is the default tab text color.
	TabColor *core.RGBA
	// TabBg is the default tab background color.
	TabBg *core.RGBA
	// SelectedTabColor is the selected tab text color.
	SelectedTabColor *core.RGBA
	// SelectedTabBg is the selected tab background color.
	SelectedTabBg *core.RGBA
}

// TabSelectRenderable is a tab selection control
type TabSelectRenderable struct {
	*core.Renderable
	options TabSelectOptions
}

// NewTabSelectRenderable creates a new TabSelectRenderable
func NewTabSelectRenderable(ctx *core.RenderContext, opts TabSelectOptions) *TabSelectRenderable {
	base := core.NewRenderable(ctx, opts.ID, core.RenderableOptions[core.Renderable]{
		Width:   opts.Width,
		Height:  opts.Height,
		Visible: opts.Visible,
		ZIndex:  opts.ZIndex,
	})
	r := &TabSelectRenderable{
		Renderable: base,
		options:    opts,
	}
	// Create native scene node
	if ctx != nil && ctx.NativeScene != nil {
		node, err := ctx.NativeScene.CreateNode(r.Renderable, native.SceneNodeKindBox)
		if err == nil {
			r.Renderable.SetNativeNode(node)
		}
	}
	return r
}

// Selected returns the currently selected tab index
func (r *TabSelectRenderable) Selected() int {
	return r.options.Selected
}

// SetSelected sets the selected tab index and requests a re-render
func (r *TabSelectRenderable) SetSelected(idx int) {
	if idx >= 0 && idx < len(r.options.Tabs) {
		r.options.Selected = idx
		r.RequestRender()
	}
}

// RenderSelf renders the tab select into the buffer
func (r *TabSelectRenderable) RenderSelf(buffer *core.NativeBuffer, deltaTime float64) {
	// Resolve colors: per-option RGBA takes precedence over renderable defaults.
	tabColor := r.options.TabColor
	if tabColor == nil {
		tabColor = r.options.ForegroundColor
	}
	tabBg := r.options.TabBg
	if tabBg == nil {
		tabBg = r.options.BackgroundColor
	}
	selTabColor := r.options.SelectedTabColor
	if selTabColor == nil {
		selTabColor = r.options.ForegroundColor
	}
	selTabBg := r.options.SelectedTabBg
	if selTabBg == nil {
		selTabBg = r.options.BackgroundColor
	}
	attrs := r.options.Attributes

	x := 0
	for i, tab := range r.options.Tabs {
		if x >= buffer.Width() {
			break
		}
		prefix := " "
		if i == r.options.Selected {
			prefix = "*"
		}
		label := prefix + tab + " "
		cellFg := tabColor
		cellBg := tabBg
		cellAttrs := attrs
		if i == r.options.Selected {
			cellFg = selTabColor
			cellBg = selTabBg
			cellAttrs = cellAttrs | core.TextAttributeBold
		}
		for _, ch := range label {
			if x >= buffer.Width() {
				break
			}
			writeCell(buffer, x, 0, uint32(ch), cellFg, cellBg, cellAttrs)
			x++
		}
		// Add separator
		if x < buffer.Width() {
			writeCell(buffer, x, 0, '|', cellFg, cellBg, cellAttrs)
			x++
		}
	}

	if r.options.RenderAfter != nil {
		r.options.RenderAfter(r.Renderable, buffer, deltaTime)
	}
}
