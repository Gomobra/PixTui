package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Gomobra/PixTui/packages/core"
	"github.com/Gomobra/PixTui/packages/core/renderables"
	"github.com/Gomobra/PixTui/packages/keymap"
	"github.com/Gomobra/PixTui/packages/native"
	"github.com/Gomobra/PixTui/packages/qrcode"
)

// demoEvent implements keymap.KeymapEvent for the demo.
type demoEvent struct {
	name      string
	ctrl      bool
	shift     bool
	meta      bool
	super     bool
	hyper     bool
	prevented bool
	stopped   bool
}

func (e *demoEvent) Name() string             { return e.name }
func (e *demoEvent) Ctrl() bool               { return e.ctrl }
func (e *demoEvent) Shift() bool              { return e.shift }
func (e *demoEvent) Meta() bool               { return e.meta }
func (e *demoEvent) Super() bool              { return e.super }
func (e *demoEvent) Hyper() bool              { return e.hyper }
func (e *demoEvent) PreventDefault()          { e.prevented = true }
func (e *demoEvent) StopPropagation()         { e.stopped = true }
func (e *demoEvent) PropagationStopped() bool { return e.stopped }

type demoTarget struct{ name string }

func main() {
	fmt.Fprintf(os.Stderr, "PixTui Demo - Starting...\n")

	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()

	// Set up keymap
	focused := &demoTarget{name: "focused"}
	host := keymap.KeymapHost[*demoTarget, *demoEvent]{
		Metadata: keymap.HostMetadata{
			Platform:        "linux",
			PrimaryModifier: "ctrl",
			Modifiers:       map[string]string{"ctrl": "supported"},
		},
		RootTarget:         &demoTarget{name: "root"},
		GetFocusedTarget:   func() *demoTarget { return focused },
		CreateCommandEvent: func() *demoEvent { return &demoEvent{} },
	}
	km, err := keymap.NewKeymap(host)
	if err != nil {
		fmt.Fprintf(os.Stderr, "keymap: %v\n", err)
		os.Exit(1)
	}

	km.RegisterLayer(keymap.Layer[*demoTarget, *demoEvent]{
		Priority: 0,
		Bindings: []keymap.Binding[*demoTarget, *demoEvent]{
			{Key: "q", Cmd: keymap.BindingCommand[*demoTarget, *demoEvent]{Kind: keymap.BindingCommandName, Name: "quit"}, PreventDefault: true},
			{Key: "r", Cmd: keymap.BindingCommand[*demoTarget, *demoEvent]{Kind: keymap.BindingCommandName, Name: "redraw"}, PreventDefault: true},
		},
		Commands: []keymap.Command{
			{Name: "quit", Run: func(ctx keymap.CommandContext[any, keymap.KeymapEvent]) keymap.CommandResult {
				renderer.Stop()
				os.Exit(0)
				return keymap.RunCommandResult{OK: true}
			}},
			{Name: "redraw", Run: func(ctx keymap.CommandContext[any, keymap.KeymapEvent]) keymap.CommandResult {
				renderer.RequestRender()
				return keymap.RunCommandResult{OK: true}
			}},
		},
	})

	ctx.Events.Key = func(data []byte) {
		if len(data) > 0 {
			key := string(data)
			km.DispatchKey(&demoEvent{name: key}, false)
		}
	}

	ctx.Events.Resize = func(w, h int) {
		fmt.Fprintf(os.Stderr, "Resized to %dx%d\n", w, h)
	}

	// Create root
	root := core.NewRootRenderable(ctx, "root")

	// Create main container box
	mainBox := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      "main-box",
			Width:   strPtr("78"),
			Height:  strPtr("22"),
			Visible: boolPtr(true),
		},
		BackgroundColor: func() *core.RGBA { c := renderables.ParseHex("transparent"); return &c }(),
		BorderStyle:     "single",
		Border:          true,
		BorderColor:     func() *core.RGBA { c := renderables.ParseHex("#FFFFFF"); return &c }(),
		ShouldFill:      false,
		Title:           "PixTui - Complete TUI Library",
		TitleAlignment:  "left",
		Focusable:       true,
	})

	// Create title text
	titleText := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      "title-text",
			Width:   strPtr("76"),
			Height:  strPtr("1"),
			Visible: boolPtr(true),
		},
		Content: "Welcome to PixTui - A complete Go port of PixTui",
	})

	// Create info text
	infoText := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      "info-text",
			Width:   strPtr("76"),
			Height:  strPtr("4"),
			Visible: boolPtr(true),
		},
		Content: "Available components:\n  Box • Text • Input • Select • ScrollBox • Textarea • Slider\n  TabSelect • EditBuffer • Image • Video • LineNumber • ScrollBar\n  TextTable • ASCIIFont • FrameBuffer • Code • Markdown • Diff\n\nPress 'q' to quit, 'r' to redraw",
	})

	// Create input field
	input := renderables.NewInputRenderable(ctx, renderables.InputOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      "input-field",
			Width:   strPtr("40"),
			Height:  strPtr("1"),
			Visible: boolPtr(true),
		},
		Value:       "Type something here...",
		Placeholder: "Enter text",
	})

	// Create select dropdown
	selectRenderable := renderables.NewSelectRenderable(ctx, renderables.SelectOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      "select-dropdown",
			Width:   strPtr("30"),
			Height:  strPtr("6"),
			Visible: boolPtr(true),
		},
		Options:  []string{"Option 1", "Option 2", "Option 3", "Option 4", "Option 5"},
		Selected: 0,
	})

	// Create textarea
	textarea := renderables.NewTextareaRenderable(ctx, renderables.TextareaOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      "textarea",
			Width:   strPtr("50"),
			Height:  strPtr("5"),
			Visible: boolPtr(true),
		},
		Value: "Multi-line text area\nLine 2\nLine 3\n",
	})

	// Create slider
	slider := renderables.NewSliderRenderable(ctx, renderables.SliderOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      "slider",
			Width:   strPtr("30"),
			Height:  strPtr("1"),
			Visible: boolPtr(true),
		},
		Min:   0,
		Max:   100,
		Value: 50,
	})

	// Create tab select
	tabSelect := renderables.NewTabSelectRenderable(ctx, renderables.TabSelectOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      "tab-select",
			Width:   strPtr("40"),
			Height:  strPtr("1"),
			Visible: boolPtr(true),
		},
		Tabs:     []string{"Tab 1", "Tab 2", "Tab 3"},
		Selected: 0,
	})

	// Create edit buffer renderable
	editBufRenderable := renderables.NewEditBufferRenderable(ctx, renderables.EditBufferRenderableOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      "edit-buffer",
			Width:   strPtr("40"),
			Height:  strPtr("4"),
			Visible: boolPtr(true),
		},
		Value: "Editable text buffer\nWith undo/redo support\n",
	})

	// Create line number renderable
	lineNumbers := renderables.NewLineNumberRenderable(ctx, "line-numbers", 1)

	// Create scroll bar
	scrollBar := renderables.NewScrollBarRenderable(ctx, renderables.ScrollBarOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      "scroll-bar",
			Width:   strPtr("1"),
			Height:  strPtr("10"),
			Visible: boolPtr(true),
		},
		Vertical:   true,
		ThumbSize:  3,
		ThumbStart: 2,
	})

	// Create text table
	textTable := renderables.NewTextTableRenderable(ctx, renderables.TextTableOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      "text-table",
			Width:   strPtr("50"),
			Height:  strPtr("6"),
			Visible: boolPtr(true),
		},
		Headers: []string{"Name", "Status", "Value"},
		Rows: [][]string{
			{"Alpha", "Active", "100"},
			{"Beta", "Inactive", "200"},
			{"Gamma", "Active", "300"},
			{"Delta", "Pending", "400"},
		},
	})

	// Create ASCII font renderable
	asciiFont := renderables.NewASCIIFontRenderable(ctx, renderables.ASCIIFontOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      "ascii-font",
			Width:   strPtr("30"),
			Height:  strPtr("3"),
			Visible: boolPtr(true),
		},
		Text: "HELLO",
		Font: "block",
	})

	// Create frame buffer renderable
	frameBuffer := renderables.NewFrameBufferRenderable(ctx, renderables.FrameBufferOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      "frame-buffer",
			Width:   strPtr("20"),
			Height:  strPtr("5"),
			Visible: boolPtr(true),
		},
	})

	// Create code renderable
	codeRenderable := renderables.NewCodeRenderable(ctx, renderables.CodeOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      "code-block",
			Width:   strPtr("40"),
			Height:  strPtr("5"),
			Visible: boolPtr(true),
		},
		Code:     "package main\n\nfunc main() {\n    println(\"Hello, Go!\")\n}",
		Language: "go",
	})

	// Create markdown renderable
	markdownRenderable := renderables.NewMarkdownRenderable(ctx, renderables.MarkdownOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      "markdown",
			Width:   strPtr("40"),
			Height:  strPtr("5"),
			Visible: boolPtr(true),
		},
		Content: "# Heading\n\nSome **bold** text and *italic* text.",
	})

	// Create diff renderable
	diffRenderable := renderables.NewDiffRenderable(ctx, renderables.DiffOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      "diff",
			Width:   strPtr("40"),
			Height:  strPtr("5"),
			Visible: boolPtr(true),
		},
		Content: "- Removed line\n+ Added line\n  Unchanged line",
	})

	// Create image renderable (placeholder)
	img := native.NewImage(20, 10)
	imageRenderable := renderables.NewImageRenderable(ctx, renderables.ImageOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      "image",
			Width:   strPtr("20"),
			Height:  strPtr("10"),
			Visible: boolPtr(true),
		},
		Image:    img,
		Protocol: core.ImageRenderProtocolAuto,
	})

	// Create video renderable (placeholder)
	vid := native.NewVideo(20, 10)
	videoRenderable := renderables.NewVideoRenderable(ctx, renderables.VideoOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      "video",
			Width:   strPtr("20"),
			Height:  strPtr("10"),
			Visible: boolPtr(true),
		},
		Video:    vid,
		AutoPlay: true,
	})

	// Build hierarchy
	mainBox.AddChild(titleText.Renderable)
	mainBox.AddChild(infoText.Renderable)
	mainBox.AddChild(input.Renderable)
	mainBox.AddChild(selectRenderable.Renderable)
	mainBox.AddChild(textarea.Renderable)
	mainBox.AddChild(slider.Renderable)
	mainBox.AddChild(tabSelect.Renderable)
	mainBox.AddChild(editBufRenderable.Renderable)
	mainBox.AddChild(textTable.Renderable)
	mainBox.AddChild(codeRenderable.Renderable)
	mainBox.AddChild(markdownRenderable.Renderable)
	mainBox.AddChild(diffRenderable.Renderable)
	mainBox.AddChild(asciiFont.Renderable)
	mainBox.AddChild(imageRenderable.Renderable)
	mainBox.AddChild(videoRenderable.Renderable)
	mainBox.AddChild(frameBuffer.Renderable)
	mainBox.AddChild(lineNumbers.Renderable)
	mainBox.AddChild(scrollBar.Renderable)

	root.AddChild(mainBox.Renderable)

	// Create QR code demo
	qr, err := qrcode.NewQRCode("https://github.com/Gomobra/PixTui", qrcode.LevelM)
	if err == nil {
		fmt.Fprintf(os.Stderr, "QR Code generated (%dx%d):\n%s\n", qr.Width, qr.Height, qr.Render())
	}

	// Start renderer
	if err := renderer.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Note: Terminal not available (not a TTY). Demo will run in headless mode.\n")
		// In headless mode, just render once and exit
		renderer.Render(root, 0.016)
		access := renderer.Capture()
		fmt.Fprintf(os.Stderr, "Rendered frame: %dx%d cells, generation %d\n", access.Width, access.Height, access.Generation)
		return
	}

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Render loop
	go func() {
		ticker := time.NewTicker(time.Second / 30)
		defer ticker.Stop()
		for range ticker.C {
			renderer.Render(root, 0.033)
			if err := renderer.WriteFrame(); err != nil {
				return
			}
		}
	}()

	<-sigChan
	renderer.Destroy()
}

func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }
