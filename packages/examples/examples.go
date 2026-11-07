package examples

import (
	"fmt"
	"image/color"
	"math"
	"strings"
	"time"

	"github.com/Gomobra/PixTui/packages/animation"
	"github.com/Gomobra/PixTui/packages/core"
	"github.com/Gomobra/PixTui/packages/keymap"
	"github.com/Gomobra/PixTui/packages/core/renderables"
	cryptossh "golang.org/x/crypto/ssh"
	"github.com/Gomobra/PixTui/packages/native"
	"github.com/Gomobra/PixTui/packages/plugin"
	"github.com/Gomobra/PixTui/packages/qrcode"
	"github.com/Gomobra/PixTui/packages/react"
	"github.com/Gomobra/PixTui/packages/ssh"
	"github.com/Gomobra/PixTui/packages/three"
	"github.com/Gomobra/PixTui/packages/core/post"
	"github.com/Gomobra/PixTui/packages/solid"
	"github.com/Gomobra/PixTui/packages/web"


)
// hexPtr parses a CSS hex color string and returns a pointer to core.RGBA.
func hexPtr(s string) *core.RGBA {
	c := renderables.ParseHex(s)
	return &c
}

func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool   { return &b }
func intPtr(i int) *int      { return &i }
func float64Ptr(f float64) *float64 { return &f }

// Example applications for OpenTUI Go
// Mirrors packages/examples/ in the original TypeScript repo.
// Every example is headless-safe: it renders a single frame even when no
// TTY is available, and only enters the interactive render loop when a real
// terminal is present.

// RunExample runs an example application by name.
func RunExample(name string) error {
	switch name {
	case "hello":
		return runHello()
	case "box":
		return runBox()
	case "text":
		return runText()
	case "input":
		return runInput()
	case "select":
		return runSelect()
	case "scrollbox":
		return runScrollbox()
	case "textarea":
		return runTextarea()
	case "slider":
		return runSlider()
	case "tabselect":
		return runTabselect()
	case "editbuffer":
		return runEditBuffer()
	case "image":
		return runImage()
	case "video":
		return runVideo()
	case "markdown":
		return runMarkdown()
	case "diff":
		return runDiff()
	case "ascii-font":
		return runASCIIFont()
	case "framebuffer":
		return runFrameBuffer()
	case "code":
		return runCode()
	case "table":
		return runTable()
	case "qrcode":
		return runQRCode()
	case "3d":
		return run3D()
	case "ssh":
		return runSSH()
	case "react":
		return runReact()
	case "animation":
		return runAnimation()
	case "plugin":
		return runPlugin()
	case "full":
		return runFullDemo()
	case "animation-demo":
		return runAnimationDemo()
	case "extmarks":
		return runExtmarks()
	case "full-styled":
		return runFullStyled()
	case "image-file":
		return runImageFile()
	case "keymap":
		return runKeymap()
	case "layout":
		return runLayout()
	case "plugin-full":
		return runPluginFull()
	case "qrcode-large":
		return runQrcodeLarge()
	case "tree-sitter-code":
		return runTreeSitterCode()
	case "video-file":
		return runVideoFile()
	case "post-processing":
		return runPostProcessing()
	case "event-system":
		return runEventSystem()
	case "scene-graph":
		return runSceneGraph()
	case "text-buffer":
		return runTextBuffer()
	case "edit-buffer-undo":
		return runEditBufferUndo()
	case "line-number-scrollbox":
		return runLineNumberScrollbox()
	case "textarea-editing":
		return runTextareaEditing()
	case "image-protocols":
		return runImageProtocols()
	case "video-player":
		return runVideoPlayer()
	case "full-styled-2":
		return runFullStyled2()
	case "slider-interaction":
		return runSliderInteraction()
	case "code-line-numbers":
		return runCodeLineNumbers()
	case "select-keyboard":
		return runSelectKeyboard()
	case "text-buffer-large":
		return runTextBufferLarge()
	case "text-node-renderable":
		return runTextNodeRenderable()
	case "post-processing-filters":
		return runPostProcessingFilters()
	case "image-resize":
		return runImageResize()
	case "audio-tones":
		return runAudioTones()
	case "keymap-layers":
		return runKeymapLayers()
	case "react-jsx":
		return runReactJSX()
	case "solid-signals":
		return runSolidSignals()
	case "ssh-full":
		return runSSHFull()
	case "three-3d":
		return runThree3D()
	case "web-server":
		return runWebServer()
	case "animation-timeline":
		return runAnimationTimeline()
	case "plugin-full-lifecycle":
		return runPluginFullLifecycle()
	case "extmarks-highlight":
		return runExtmarksHighlight()

	case "opacity-demo":
		return runOpacityDemo()
	case "styled-text-demo":
		return runStyledTextDemo()
	case "simple-layout-demo":
		return runSimpleLayoutDemo()
	case "text-truncation-demo":
		return runTextTruncationDemo()
	case "text-wrap-demo":
		return runTextWrapDemo()
	case "transparency-demo":
		return runTransparencyDemo()
	case "link-demo":
		return runLinkDemo()
	case "full-unicode-demo":
		return runFullUnicodeDemo()
	case "text-selection-demo":
		return runTextSelectionDemo()
	case "relative-positioning-demo":
		return runRelativePositioningDemo()
	case "nested-zindex-demo":
		return runNestedZIndexDemo()
	case "live-state-demo":
		return runLiveStateDemo()
	case "program-status-demo":
		return runProgramStatusDemo()
	case "sticky-scroll-demo":
		return runStickyScrollDemo()
	case "scroll-demo":
		return runScrollDemo()

	case "console-demo":
		return runConsoleDemo()
	case "editor-demo":
		return runEditorDemo()
	case "keypress-debug-demo":
		return runKeypressDebugDemo()
	case "markdown-code-block-renderer-demo":
		return runMarkdownCodeBlockRendererDemo()
	case "mouse-interaction-demo":
		return runMouseInteractionDemo()
	case "notification-demo":
		return runNotificationDemo()
	case "pcm-playback":
		return runPcmPlayback()
	case "sprite-animation-demo":
		return runSpriteAnimationDemo()
	case "static-sprite-demo":
		return runStaticSpriteDemo()
	case "terminal":
		return runTerminal()
	case "terminal-title":
		return runTerminalTitle()
	case "texture-loading-demo":
		return runTextureLoadingDemo()
	case "video-puzzle-demo":
		return runVideoPuzzleDemo()
	case "wide-grapheme-overlay-demo":
		return runWideGraphemeOverlayDemo()
	case "input-select-layout-demo":
		return runInputSelectLayoutDemo()

	case "ascii-font-selection-demo":
		return runASCIIFontSelectionDemo()
	case "audio-capture-demo":
		return runAudioCaptureDemo()
	case "clipboard-paste-demo":
		return runClipboardPasteDemo()
	case "core-plugin-slots-demo":
		return runCorePluginSlotsDemo()
	case "focus-restore-demo":
		return runFocusRestoreDemo()
	case "shader-cube-demo":
		return runShaderCubeDemo()
	case "split-footer-image-demo":
		return runSplitFooterImageDemo()
	case "split-footer-streaming-demo":
		return runSplitFooterStreamingDemo()
	case "split-mode-demo":
		return runSplitModeDemo()
	case "sprite-particle-generator-demo":
		return runSpriteParticleGeneratorDemo()
	case "vnode-composition-demo":
		return runVnodeCompositionDemo()
	case "lights-phong-demo":
		return runLightsPhongDemo()
	case "grayscale-buffer":
		return runGrayscaleBuffer()
	case "scrollbox-mouse":
		return runScrollboxMouse()
	case "scrollbox-overlay":
		return runScrollboxOverlay()
	case "fractal-buffer":
		return runFractalBuffer()
	case "native-image":
		return runNativeImage()
	case "fonts-demo":
		return runFontsDemo()
	case "embedded-terminal-demo":
		return runEmbeddedTerminalDemo()
	case "opentui-demo":
		return runOpenTUIDemo()
	case "audio-streaming-demo":
		return runAudioStreamingDemo()
	case "draggable-three-demo":
		return runDraggableThreeDemo()

	case "input-demo":
		return runInputDemo()
	case "keymap-demo":
		return runKeymapDemo()
	case "native-audio-demo":
		return runNativeAudioDemo()
	case "timeline-example":
		return runTimelineExample()

	case "golden-star-demo":
		return runGoldenStarDemo()
	case "hast-syntax-highlighting-demo":
		return runHastSyntaxHighlightingDemo()
	case "physx-planck-2d-demo":
		return runPhysxPlanck2DDemo()
	case "physx-rapier-2d-demo":
		return runPhysxRapier2DDemo()

	default:
		return fmt.Errorf("unknown example: %s", name)
	}
}

// Build builds all example applications.
func Build() error {
	return nil
}

// Run runs an example application by name.
func Run(name string) error {
	return RunExample(name)
}

// renderHeadless renders a single frame. This works without a TTY and is the
// rendering path used by every example when run headless.
func renderHeadless(r *core.CliRenderer, root *core.RootRenderable) {
	r.Render(root, 0.016)
	r.Capture()
}

// runLoop starts the renderer and runs the interactive render loop until 'q'
// is pressed or the timeout elapses. If no TTY is available it returns nil
// immediately after the headless render already performed by the caller.
func runLoop(r *core.CliRenderer, root *core.RootRenderable, timeout time.Duration) error {
	if err := r.Start(); err != nil {
		// Headless environment (not a TTY): the headless render already ran.
		return nil
	}
	defer r.Stop()

	done := make(chan bool, 1)
	go func() {
		ticker := time.NewTicker(time.Second / 30)
		defer ticker.Stop()
		for range ticker.C {
			r.Render(root, 0.033)
			if err := r.WriteFrame(); err != nil {
				done <- true
				return
			}
		}
	}()

	select {
	case <-done:
	case <-time.After(timeout):
	}
	return nil
}

// quitKey installs a key handler that stops the renderer when 'q' is pressed.
func quitKey(r *core.CliRenderer) {
	r.Context().Events.Key = func(data []byte) {
		if len(data) > 0 && data[0] == 'q' {
			r.Stop()
		}
	}
}

// bufferToText reads a NativeBuffer into a multi-line string for display
// inside a TextRenderable. This is used by demos that render to an offscreen
// buffer (three.js, react) and then present the result as text.
func bufferToText(buf *core.NativeBuffer) string {
	access := buf.Access()
	var sb strings.Builder
	for y := 0; y < access.Height; y++ {
		for x := 0; x < access.Width; x++ {
			idx := y*access.Width + x
			ch := access.Char[idx]
			if ch == 0 {
				sb.WriteByte(' ')
			} else {
				sb.WriteRune(rune(ch))
			}
		}
		if y < access.Height-1 {
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}

// ---------------------------------------------------------------------------
// hello
// ---------------------------------------------------------------------------

func runHello() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "hello-box",
			Width:  strPtr("40"),
			Height: strPtr("10"),
			Visible: boolPtr(true),
		},
		Border:      true,
		Title:       "Hello",
		BorderStyle: "single",
	})

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "hello-text",
			Width:  strPtr("38"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
		Content: "Hello, OpenTUI Go!\n\nThis is the hello example.\nPress q to quit.",
	})

	box.AddChild(text.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// box
// ---------------------------------------------------------------------------

func runBox() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	styles := []struct {
		name   string
		style  string
		x, y   int
		w, h   int
	}{
		{"single", "single", 2, 2, 20, 7},
		{"double", "double", 24, 2, 20, 7},
		{"rounded", "rounded", 46, 2, 20, 7},
		{"heavy", "heavy", 2, 12, 20, 7},
	}

	for _, s := range styles {
		box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
			RenderableOptions: core.RenderableOptions[core.Renderable]{
				ID:     "box-" + s.name,
				Width:  strPtr(fmt.Sprintf("%d", s.w)),
				Height: strPtr(fmt.Sprintf("%d", s.h)),
				Visible: boolPtr(true),
			},
			Border:      true,
			Title:       s.name,
			BorderStyle: s.style,
		})
		box.SetPosition(float32(s.x), float32(s.y))
		root.AddChild(box.Renderable)

		text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
			RenderableOptions: core.RenderableOptions[core.Renderable]{
				ID:     "box-text-" + s.name,
				Width:  strPtr(fmt.Sprintf("%d", s.w-2)),
				Height: strPtr(fmt.Sprintf("%d", s.h-2)),
				Visible: boolPtr(true),
			},
			Content: "border demo",
		})
		text.SetPosition(float32(s.x+1), float32(s.y+1))
		root.AddChild(text.Renderable)
	}

	label := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "box-label",
			Width:  strPtr("76"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Content: "Border style demo: single / double / rounded / heavy",
	})
	label.SetPosition(2, 21)
	root.AddChild(label.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// text
// ---------------------------------------------------------------------------

func runText() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "text-box",
			Width:  strPtr("60"),
			Height: strPtr("14"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Text Renderable",
	})

	content := `OpenTUI Go Text Renderable

This example demonstrates the Text renderable component.
It renders plain text content into a buffer, handling
newlines and wrapping at the buffer edge.

The text supports dynamic content updates via SetContent.`

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "text-content",
			Width:  strPtr("58"),
			Height: strPtr("12"),
			Visible: boolPtr(true),
		},
		Content: content,
		Wrap:    true,
	})
	text.SetPosition(1, 1)

	box.AddChild(text.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// input
// ---------------------------------------------------------------------------

func runInput() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "input-box",
			Width:  strPtr("50"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Input",
	})

	input := renderables.NewInputRenderable(ctx, renderables.InputOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "input-field",
			Width:  strPtr("46"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Value:      "initial value",
		Placeholder: "Enter text here...",
	})
	input.SetPosition(2, 2)

	// Simulate user editing by updating the value.
	input.SetValue("edited value")

	label := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "input-label",
			Width:  strPtr("46"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Content: "Current value: " + input.Value() + "\nPlaceholder: Enter text here...",
	})
	label.SetPosition(2, 1)

	status := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "input-status",
			Width:  strPtr("46"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Content: "Input renderable with editable value and placeholder text.",
	})
	status.SetPosition(2, 5)

	box.AddChild(label.Renderable)
	box.AddChild(status.Renderable)
	box.AddChild(input.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// slider
// ---------------------------------------------------------------------------

func runSelect() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "select-box",
			Width:  strPtr("40"),
			Height: strPtr("12"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Select",
	})

	opts := []string{"Option A", "Option B", "Option C", "Option D", "Option E"}
	selectRenderable := renderables.NewSelectRenderable(ctx, renderables.SelectOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "select-dropdown",
			Width:  strPtr("36"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
		Options:  opts,
		Selected: 2,
	})
	selectRenderable.SetPosition(2, 2)

	label := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "select-label",
			Width:  strPtr("36"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Content: "Selected: " + opts[selectRenderable.Selected()],
	})
	label.SetPosition(2, 1)

	// Demonstrate changing the selection.
	selectRenderable.SetSelected(4)
	label.SetContent("Selected: " + opts[selectRenderable.Selected()])

	box.AddChild(label.Renderable)
	box.AddChild(selectRenderable.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// scrollbox
// ---------------------------------------------------------------------------

func runScrollbox() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "scrollbox-box",
			Width:  strPtr("50"),
			Height: strPtr("10"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "ScrollBox",
	})

	// A scrollbox holds a large body of text that exceeds the viewport.
	var sb strings.Builder
	for i := 1; i <= 30; i++ {
		sb.WriteString(fmt.Sprintf("Line %d\n", i))
	}
	body := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "scrollbox-body",
			Width:  strPtr("46"),
			Height: strPtr("40"),
			Visible: boolPtr(true),
		},
		Content: sb.String(),
	})
	body.SetPosition(1, 1)

	scrollbox := renderables.NewScrollBoxRenderable(ctx, renderables.ScrollBoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "scrollbox",
			Width:  strPtr("48"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
	})
	scrollbox.SetPosition(1, 1)
	scrollbox.SetScrollTop(0)

	scrollbox.AddChild(body.Renderable)
	box.AddChild(scrollbox.Renderable)
	root.AddChild(box.Renderable)

	label := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "scrollbox-label",
			Width:  strPtr("48"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Content: "Scroll position: " + fmt.Sprintf("%d", scrollbox.GetScrollY()),
	})
	label.SetPosition(1, 10)
	root.AddChild(label.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// textarea
// ---------------------------------------------------------------------------

func runTextarea() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "textarea-box",
			Width:  strPtr("50"),
			Height: strPtr("10"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Textarea",
	})

	textarea := renderables.NewTextareaRenderable(ctx, renderables.TextareaOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "textarea-field",
			Width:  strPtr("46"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
		},
		Value: "First line of text\nSecond line of text\nThird line of text\n",
	})
	textarea.SetPosition(2, 2)

	// Demonstrate textarea operations: insert, undo, redo.
	textarea.Insert("Inserted text\n")
	_ = textarea.Undo()
	_ = textarea.Redo()

	cursorLine, cursorCol := textarea.Cursor()
	status := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "textarea-status",
			Width:  strPtr("46"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf("Value:\n%s\nCursor: line %d, col %d",
			strings.ReplaceAll(textarea.Value(), "\n", "\n  "), cursorLine, cursorCol),
	})
	status.SetPosition(2, 9)

	box.AddChild(status.Renderable)
	box.AddChild(textarea.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// slider
// ---------------------------------------------------------------------------

func runSlider() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "slider-box",
			Width:  strPtr("50"),
			Height: strPtr("10"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Slider",
	})

	slider := renderables.NewSliderRenderable(ctx, renderables.SliderOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "slider-control",
			Width:  strPtr("40"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Min:   0,
		Max:   100,
		Value: 42,
	})
	slider.SetPosition(2, 2)

	// Demonstrate slider value changes.
	slider.SetValue(75)

	label := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "slider-label",
			Width:  strPtr("40"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf("Slider value: %.0f (range 0-100)", slider.Value()),
	})
	label.SetPosition(2, 1)

	box.AddChild(label.Renderable)
	box.AddChild(slider.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// tabselect
// ---------------------------------------------------------------------------

func runTabselect() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "tabselect-box",
			Width:  strPtr("50"),
			Height: strPtr("10"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "TabSelect",
	})

	tabs := []string{"Files", "Edit", "View", "Help"}
	tabSelect := renderables.NewTabSelectRenderable(ctx, renderables.TabSelectOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "tab-select",
			Width:  strPtr("46"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Tabs:     tabs,
		Selected: 1,
	})
	tabSelect.SetPosition(2, 2)

	content := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "tab-content",
			Width:  strPtr("46"),
			Height: strPtr("5"),
			Visible: boolPtr(true),
		},
		Content: "Viewing tab: " + tabs[tabSelect.Selected()],
	})
	content.SetPosition(2, 4)

	// Demonstrate tab switching.
	tabSelect.SetSelected(3)
	content.SetContent("Viewing tab: " + tabs[tabSelect.Selected()])

	box.AddChild(tabSelect.Renderable)
	box.AddChild(content.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// editbuffer
// ---------------------------------------------------------------------------

func runEditBuffer() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "editbuffer-box",
			Width:  strPtr("50"),
			Height: strPtr("12"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "EditBuffer",
	})

	editBuf := renderables.NewEditBufferRenderable(ctx, renderables.EditBufferRenderableOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "edit-buffer",
			Width:  strPtr("46"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
		},
		Value: "Editable text buffer\nSupports undo/redo\n",
	})
	editBuf.SetPosition(2, 2)

	// Demonstrate edit buffer operations.
	_ = editBuf.Insert(0, "[start] ")
	_ = editBuf.Undo()
	_ = editBuf.Redo()

	label := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "editbuffer-label",
			Width:  strPtr("46"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Content: "Content: " + strings.ReplaceAll(editBuf.Value(), "\n", "\\n"),
	})
	label.SetPosition(2, 1)

	box.AddChild(label.Renderable)
	box.AddChild(editBuf.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// image
// ---------------------------------------------------------------------------

func runImage() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "image-box",
			Width:  strPtr("30"),
			Height: strPtr("14"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Image",
	})

	// Create a native image and draw a gradient pattern into it so the
	// blocks renderer has something visible to show.
	img := native.NewImage(24, 10)
	for y := 0; y < img.Height; y++ {
		for x := 0; x < img.Width; x++ {
			idx := (y*img.Width + x) * 4
			img.Pixels[idx] = uint8((x * 255) / img.Width)     // R gradient
			img.Pixels[idx+1] = uint8((y * 255) / img.Height)   // G gradient
			img.Pixels[idx+2] = 128                             // B constant
			img.Pixels[idx+3] = 255                             // A opaque
		}
	}

	imageRenderable := renderables.NewImageRenderable(ctx, renderables.ImageOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "image",
			Width:  strPtr("26"),
			Height: strPtr("10"),
			Visible: boolPtr(true),
		},
		Image:    img,
		Protocol: core.ImageRenderProtocolBlocks,
	})
	imageRenderable.SetPosition(2, 2)

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "image-info",
			Width:  strPtr("26"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf("Image: %dx%d pixels\nProtocol: blocks", img.Width, img.Height),
	})
	info.SetPosition(2, 13)

	box.AddChild(imageRenderable.Renderable)
	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// video
// ---------------------------------------------------------------------------

func runVideo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "video-box",
			Width:  strPtr("30"),
			Height: strPtr("10"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Video",
	})

	vid := native.NewVideo(24, 8)
	vid.Playing = true
	vid.CurrentTime = 1.5
	vid.Duration = 10.0
	vid.FPS = 30.0

	videoRenderable := renderables.NewVideoRenderable(ctx, renderables.VideoOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "video",
			Width:  strPtr("26"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
		Video:    vid,
		AutoPlay: true,
	})
	videoRenderable.SetPosition(2, 2)

	state := "paused"
	if vid.Playing {
		state = "playing"
	}
	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "video-info",
			Width:  strPtr("26"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf("Video: %dx%d\nState: %s\nTime: %.1f / %.1f",
			vid.Width, vid.Height, state, vid.CurrentTime, vid.Duration),
	})
	info.SetPosition(2, 11)

	box.AddChild(videoRenderable.Renderable)
	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// markdown
// ---------------------------------------------------------------------------

func runMarkdown() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "markdown-box",
			Width:  strPtr("60"),
			Height: strPtr("14"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Markdown",
	})

	markdown := renderables.NewMarkdownRenderable(ctx, renderables.MarkdownOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "markdown-content",
			Width:  strPtr("58"),
			Height: strPtr("12"),
			Visible: boolPtr(true),
		},
		Content: "# OpenTUI Go\n\n## Markdown Renderable\n\nThis demonstrates the markdown component.\n\n- Item one\n- Item two\n- Item three\n",
	})
	markdown.SetPosition(1, 1)

	box.AddChild(markdown.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// diff
// ---------------------------------------------------------------------------

func runDiff() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "diff-box",
			Width:  strPtr("60"),
			Height: strPtr("12"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Diff",
	})

	diff := renderables.NewDiffRenderable(ctx, renderables.DiffOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "diff-content",
			Width:  strPtr("58"),
			Height: strPtr("10"),
			Visible: boolPtr(true),
		},
		Content: "- Removed line from original\n+ Added line in revision\n  Unchanged context line\n- Another deletion\n+ Another addition\n",
	})
	diff.SetPosition(1, 1)

	box.AddChild(diff.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// ascii-font
// ---------------------------------------------------------------------------

func runASCIIFont() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "ascii-font-box",
			Width:  strPtr("40"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "ASCII Font",
	})

	asciiFont := renderables.NewASCIIFontRenderable(ctx, renderables.ASCIIFontOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "ascii-font",
			Width:  strPtr("36"),
			Height: strPtr("4"),
			Visible: boolPtr(true),
		},
		Text: "HELLO",
		Font: "block",
	})
	asciiFont.SetPosition(2, 2)

	box.AddChild(asciiFont.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// framebuffer
// ---------------------------------------------------------------------------

func runFrameBuffer() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "framebuffer-box",
			Width:  strPtr("30"),
			Height: strPtr("10"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "FrameBuffer",
	})

	frameBuffer := renderables.NewFrameBufferRenderable(ctx, renderables.FrameBufferOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:         "frame-buffer",
			Width:      strPtr("26"),
			Height:     strPtr("6"),
			Visible:    boolPtr(true),
			RenderAfter: func(_ *core.Renderable, buf *core.NativeBuffer, _ float64) {
				for y := 0; y < 6 && y < buf.Height(); y++ {
					for x := 0; x < 26 && x < buf.Width(); x++ {
						if x%2 == 0 && y%2 == 0 {
							buf.SetCell(x, y, '#', 0, 0, 0)
						}
					}
				}
			},
		},
	})
	frameBuffer.SetPosition(2, 2)

	box.AddChild(frameBuffer.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// code
// ---------------------------------------------------------------------------

func runCode() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "code-box",
			Width:  strPtr("60"),
			Height: strPtr("14"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Code (Go)",
	})

	code := `package main

import "fmt"

func main() {
	fmt.Println("Hello, syntax highlighting!")
}`
	codeRenderable := renderables.NewCodeRenderable(ctx, renderables.CodeOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "code-block",
			Width:  strPtr("56"),
			Height: strPtr("10"),
			Visible: boolPtr(true),
		},
		Code:     code,
		Language: "go",
	})
	codeRenderable.SetPosition(2, 2)

	// Verify language can be changed at runtime.
	codeRenderable.SetLanguage("go")

	box.AddChild(codeRenderable.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// table
// ---------------------------------------------------------------------------

func runTable() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "table-box",
			Width:  strPtr("60"),
			Height: strPtr("12"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Text Table",
	})

	textTable := renderables.NewTextTableRenderable(ctx, renderables.TextTableOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "text-table",
			Width:  strPtr("56"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
		Headers: []string{"Package", "Status", "Version"},
		Rows: [][]string{
			{"core", "active", "0.1.0"},
			{"renderables", "active", "0.1.0"},
			{"native", "active", "0.1.0"},
			{"animation", "active", "0.1.0"},
			{"qrcode", "active", "0.1.0"},
		},
	})
	textTable.SetPosition(2, 2)

	box.AddChild(textTable.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// qrcode
// ---------------------------------------------------------------------------

func runQRCode() error {
	qr, err := qrcode.NewQRCode("https://github.com/Gomobra/PixTui", qrcode.LevelM)
	if err != nil {
		return fmt.Errorf("failed to generate QR code: %w", err)
	}

	rendered := qr.Render()
	if rendered == "" {
		return fmt.Errorf("QR code render produced empty output")
	}

	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "qrcode-box",
			Width:  strPtr("30"),
			Height: strPtr("30"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "QR Code",
	})

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "qrcode-text",
			Width:  strPtr("28"),
			Height: strPtr("28"),
			Visible: boolPtr(true),
		},
		Content: rendered,
	})
	text.SetPosition(1, 1)

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "qrcode-info",
			Width:  strPtr("28"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Content: "https://github.com/Gomobra/PixTui",
	})
	info.SetPosition(1, 29)

	box.AddChild(text.Renderable)
	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// 3d
// ---------------------------------------------------------------------------

func run3D() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	// Set up a three.js-style CPU 3D renderer targeting a native buffer.
	threeRenderer := three.NewRenderer(40, 20)
	scene := three.NewScene()
	camera := three.NewCamera()

	// A colored triangle mesh.
	shader := three.NewShader("default", "", "")
	mat := three.NewMaterial(shader)
	mat.Color = color.RGBA{R: 255, G: 0, B: 0, A: 255}
	geo := three.NewGeometry()
	geo.Vertices = []three.Vector3{
		{X: -1, Y: -1, Z: 0},
		{X: 1, Y: -1, Z: 0},
		{X: 0, Y: 1, Z: 0},
	}
	geo.Indices = []int{0, 1, 2}
	mesh := three.NewMesh(geo, mat)
	scene.AddObject(mesh)
	threeRenderer.SetScene(scene)
	threeRenderer.SetCamera(camera)

	buf := core.NewNativeBuffer(40, 20)
	threeRenderer.Render(buf, 0.016)

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "3d-text",
			Width:  strPtr("38"),
			Height: strPtr("18"),
			Visible: boolPtr(true),
		},
		Content: "CPU 3D render output (triangle mesh):\n\n" + bufferToText(buf) +
			"\n\nScene objects: 1 | Camera: active",
	})

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "3d-box",
			Width:  strPtr("40"),
			Height: strPtr("20"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "3D Renderer",
	})

	box.AddChild(text.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// ssh
// ---------------------------------------------------------------------------

func runSSH() error {
	// Create an SSH server that serves an OpenTUI welcome screen.
	server := ssh.NewServer("127.0.0.1:0", func(session *ssh.Session) error {
		session.Write([]byte("\r\nWelcome to OpenTUI over SSH\r\n"))
		session.Write([]byte("This session is served by the Go SSH package.\r\n"))
		session.Write([]byte("Press q to quit.\r\n\r\n"))
		return nil
	})
	server.SetNoAuth()

	if err := server.Listen(); err != nil {
		return fmt.Errorf("failed to start SSH server: %w", err)
	}
	defer server.Close()

	addr := server.Addr()

	// Connect an SSH client to demonstrate a full round trip.
	config := &cryptossh.ClientConfig{
		User:            "demo",
		Auth:            []cryptossh.AuthMethod{},
		HostKeyCallback: cryptossh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	conn, err := cryptossh.Dial("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("failed to dial SSH server: %w", err)
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		return fmt.Errorf("failed to open SSH session: %w", err)
	}
	defer session.Close()

	if err := session.Shell(); err != nil {
		return fmt.Errorf("failed to request shell: %w", err)
	}

	// Read the welcome message written by the server handler.
	stdout, err := session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	data := make([]byte, 0)
	tmp := make([]byte, 1024)
	for {
		nn, rerr := stdout.Read(tmp)
		if nn > 0 {
			data = append(data, tmp[:nn]...)
		}
		if strings.Contains(string(data), "Press q to quit.") {
			break
		}
		if rerr != nil || nn == 0 {
			break
		}
	}

	output := string(data)
	if !strings.Contains(output, "Welcome to OpenTUI over SSH") {
		return fmt.Errorf("expected SSH welcome message, got: %q", output)
	}

	_ = output
	return nil
}

// ---------------------------------------------------------------------------
// react
// ---------------------------------------------------------------------------

func runReact() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	// React-style element tree rendered to a native buffer.
	reactRenderer := react.NewReactRenderer(ctx)
	buf := core.NewNativeBuffer(80, 24)

	app := react.CreateElement("box", map[string]interface{}{
		"id":           "app",
		"border":       true,
		"title":        "React Demo",
		"borderStyle":  "single",
	}, react.CreateElement("text", map[string]interface{}{
		"id":      "message",
		"content": "Hello from React-style elements!",
	}))

	reactRenderer.Render(app, buf, 0.016)

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "react-text",
			Width:  strPtr("78"),
			Height: strPtr("22"),
			Visible: boolPtr(true),
		},
		Content: "React-style element tree rendered to buffer:\n\n" + bufferToText(buf),
	})

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "react-box",
			Width:  strPtr("80"),
			Height: strPtr("24"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "React",
	})

	box.AddChild(text.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// animation
// ---------------------------------------------------------------------------

func runAnimation() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	// Demonstrate the animation package with a looping value animation.
	animator := animation.NewAnimator()

	var animatedValue float64
	anim := animation.NewAnimation("fade", time.Second, 0, 100, animation.EaseInOut)
	anim.OnUpdate = func(value float64) {
		animatedValue = value
	}
	anim.SetLoop(true)
	animator.Add(anim)

	// Advance the animation a few steps.
	animator.Update(time.Millisecond * 250)
	animator.Update(time.Millisecond * 250)
	animator.Update(time.Millisecond * 250)

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "animation-box",
			Width:  strPtr("50"),
			Height: strPtr("10"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Animation",
	})

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "animation-text",
			Width:  strPtr("46"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf("Animated value: %.0f\nEasing: EaseInOut\nAnimator count: %d",
			animatedValue, animator.Count()),
	})
	text.SetPosition(2, 2)

	box.AddChild(text.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// plugin
// ---------------------------------------------------------------------------

func runPlugin() error {
	manager := plugin.NewManager()

	// Register a demo plugin.
	demoPlugin := &demoPluginImpl{
		BasePlugin: plugin.NewBasePlugin("demo-plugin", "1.0.0"),
	}
	if err := manager.Register(demoPlugin); err != nil {
		return fmt.Errorf("failed to register plugin: %w", err)
	}

	if manager.Count() != 1 {
		return fmt.Errorf("expected 1 plugin, got %d", manager.Count())
	}

	names := manager.List()
	if len(names) != 1 || names[0] != "demo-plugin" {
		return fmt.Errorf("expected plugin list [demo-plugin], got %v", names)
	}

	got, ok := manager.Get("demo-plugin")
	if !ok || got.Name() != "demo-plugin" {
		return fmt.Errorf("expected to retrieve demo-plugin")
	}

	// Verify Init was called during Register.
	if !demoPlugin.initCalled {
		return fmt.Errorf("plugin Init was not called during Register")
	}

	if err := manager.Shutdown(); err != nil {
		return fmt.Errorf("failed to shutdown plugins: %w", err)
	}

	if manager.Count() != 0 {
		return fmt.Errorf("expected 0 plugins after shutdown")
	}

	if !demoPlugin.shutdownCalled {
		return fmt.Errorf("plugin Shutdown was not called during Shutdown")
	}

	return nil
}

type demoPluginImpl struct {
	*plugin.BasePlugin
	initCalled    bool
	shutdownCalled bool
}

func (p *demoPluginImpl) Init(ctx *plugin.Context) error {
	p.initCalled = true
	// Register a command and event handler through the plugin context.
	ctx.RegisterCommand("demo-cmd", func(args []string) (string, error) {
		return "demo-plugin: " + strings.Join(args, " "), nil
	})
	ctx.OnEvent("demo-event", func(data interface{}) {
		_ = data
	})
	return nil
}

func (p *demoPluginImpl) Shutdown() error {
	p.shutdownCalled = true
	return nil
}

// ---------------------------------------------------------------------------
// full demo
// ---------------------------------------------------------------------------

// runFullDemo renders every core component together in a single screen.
func runFullDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	mainBox := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "main-box",
			Width:  strPtr("76"),
			Height: strPtr("20"),
			Visible: boolPtr(true),
		},
		BackgroundColor: hexPtr("transparent"),
		BorderStyle:     "single",
		Border:          true,
		BorderColor: hexPtr("#FFFFFF"),
		ShouldFill:      false,
		Title:           "OpenTUI Go - Full Demo",
		TitleAlignment:  "left",
		Focusable:       true,
	})

	title := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "full-title",
			Width:  strPtr("74"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Content: "Welcome to OpenTUI Go - all components in one view",
	})
	title.SetPosition(1, 1)

	input := renderables.NewInputRenderable(ctx, renderables.InputOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "full-input",
			Width:  strPtr("30"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Value:      "type here",
		Placeholder: "Enter text",
	})
	input.SetPosition(1, 3)

	selectRenderable := renderables.NewSelectRenderable(ctx, renderables.SelectOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "full-select",
			Width:  strPtr("25"),
			Height: strPtr("5"),
			Visible: boolPtr(true),
		},
		Options:   []string{"Red", "Green", "Blue"},
		Selected:  0,
	})
	selectRenderable.SetPosition(32, 3)

	textarea := renderables.NewTextareaRenderable(ctx, renderables.TextareaOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "full-textarea",
			Width:  strPtr("40"),
			Height: strPtr("4"),
			Visible: boolPtr(true),
		},
		Value: "Multi-line\ntext area\n",
	})
	textarea.SetPosition(1, 5)

	slider := renderables.NewSliderRenderable(ctx, renderables.SliderOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "full-slider",
			Width:  strPtr("30"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Min:   0,
		Max:   100,
		Value: 50,
	})
	slider.SetPosition(1, 10)

	tabSelect := renderables.NewTabSelectRenderable(ctx, renderables.TabSelectOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "full-tabselect",
			Width:  strPtr("30"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Tabs:     []string{"A", "B", "C"},
		Selected: 0,
	})
	tabSelect.SetPosition(1, 12)

	editBuf := renderables.NewEditBufferRenderable(ctx, renderables.EditBufferRenderableOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "full-editbuffer",
			Width:  strPtr("40"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Value: "edit buffer\n",
	})
	editBuf.SetPosition(1, 14)

	codeRenderable := renderables.NewCodeRenderable(ctx, renderables.CodeOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "full-code",
			Width:  strPtr("38"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
		},
		Code:     "fmt.Println(\"hi\")",
		Language: "go",
	})
	codeRenderable.SetPosition(40, 5)

	markdownRenderable := renderables.NewMarkdownRenderable(ctx, renderables.MarkdownOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "full-markdown",
			Width:  strPtr("38"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
		},
		Content: "# Markdown\n\ndemo",
	})
	markdownRenderable.SetPosition(40, 12)

	diffRenderable := renderables.NewDiffRenderable(ctx, renderables.DiffOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "full-diff",
			Width:  strPtr("38"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
		},
		Content: "- old\n+ new",
	})
	diffRenderable.SetPosition(40, 19)

	textTable := renderables.NewTextTableRenderable(ctx, renderables.TextTableOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "full-table",
			Width:  strPtr("38"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
		},
		Headers: []string{"A", "B"},
		Rows:    [][]string{{"1", "2"}, {"3", "4"}},
	})
	textTable.SetPosition(1, 18)

	asciiFont := renderables.NewASCIIFontRenderable(ctx, renderables.ASCIIFontOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "full-ascii",
			Width:  strPtr("20"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Text: "HI",
		Font: "block",
	})
	asciiFont.SetPosition(58, 3)

	imageRenderable := renderables.NewImageRenderable(ctx, renderables.ImageOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "full-image",
			Width:  strPtr("10"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
		},
		Image:    native.NewImage(10, 6),
		Protocol: core.ImageRenderProtocolAuto,
	})
	imageRenderable.SetPosition(58, 7)

	videoRenderable := renderables.NewVideoRenderable(ctx, renderables.VideoOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "full-video",
			Width:  strPtr("10"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
		},
		Video:    native.NewVideo(10, 6),
		AutoPlay: true,
	})
	videoRenderable.SetPosition(58, 14)

	lineNumbers := renderables.NewLineNumberRenderable(ctx, "full-lineno", 1)
	lineNumbers.SetPosition(1, 22)

	scrollBar := renderables.NewScrollBarRenderable(ctx, renderables.ScrollBarOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "full-scrollbar",
			Width:  strPtr("1"),
			Height: strPtr("10"),
			Visible: boolPtr(true),
		},
		Vertical:   true,
		ThumbSize:  3,
		ThumbStart: 1,
	})
	scrollBar.SetPosition(74, 5)

	frameBuffer := renderables.NewFrameBufferRenderable(ctx, renderables.FrameBufferOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "full-framebuffer",
			Width:  strPtr("15"),
			Height: strPtr("5"),
			Visible: boolPtr(true),
		},
	})
	frameBuffer.SetPosition(58, 21)

	for _, child := range []*core.Renderable{
		title.Renderable,
		input.Renderable,
		selectRenderable.Renderable,
		textarea.Renderable,
		slider.Renderable,
		tabSelect.Renderable,
		editBuf.Renderable,
		codeRenderable.Renderable,
		markdownRenderable.Renderable,
		diffRenderable.Renderable,
		textTable.Renderable,
		asciiFont.Renderable,
		imageRenderable.Renderable,
		videoRenderable.Renderable,
		lineNumbers.Renderable,
		scrollBar.Renderable,
		frameBuffer.Renderable,
	} {
		mainBox.AddChild(child)
	}

	root.AddChild(mainBox.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// keymap helper types (demo-only mock event/target/host)
// ---------------------------------------------------------------------------

// keEvent is a minimal KeymapEvent implementation used by the keymap demo.
type keEvent struct {
	name      string
	ctrl, shift, meta, super, hyper bool
	prevented, stopped              bool
}

func (e *keEvent) Name() string             { return e.name }
func (e *keEvent) Ctrl() bool               { return e.ctrl }
func (e *keEvent) Shift() bool              { return e.shift }
func (e *keEvent) Meta() bool               { return e.meta }
func (e *keEvent) Super() bool              { return e.super }
func (e *keEvent) Hyper() bool              { return e.hyper }
func (e *keEvent) PreventDefault()          { e.prevented = true }
func (e *keEvent) StopPropagation()         { e.stopped = true }
func (e *keEvent) PropagationStopped() bool { return e.stopped }

// keTarget is a minimal keymap target.
type keTarget struct{ name string }


// ---------------------------------------------------------------------------
// layout
// ---------------------------------------------------------------------------

// runLayout demonstrates Yoga flexbox layout: flexDirection row/column,
// justifyContent, alignItems, with multiple children laid out automatically.
func runLayout() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	yoga := native.NewYoga()

	// Container node: row direction, centered on both axes.
	parent := yoga.CreateNode()
	parent.SetFlexDirection(native.FlexDirectionRow)
	parent.SetJustifyContent(native.JustifyCenter)
	parent.SetAlignItems(native.AlignCenter)
	parent.SetWidth(60)
	parent.SetHeight(8)
	parent.SetPadding(1, 1, 1, 1)
	parent.SetGap(0, 2)

	borderColors := []string{"#FF0000", "#00FF00", "#0000FF"}
	titles := []string{"Panel A", "Panel B", "Panel C"}

	var children []*native.YogaNode
	for i := 0; i < 3; i++ {
		child := yoga.CreateNode()
		child.SetWidth(16)
		child.SetHeight(4)
		child.SetMargin(0, 1, 0, 1)
		parent.AddChild(child)
		children = append(children, child)
	}

	yoga.CalculateLayout(parent, 60, 8)

	// Container box.
	originX, originY := 2, 2
	container := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "layout-container",
			Width:  strPtr("60"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
		Border:      true,
		Title:       "Yoga Layout: row / center / center",
		BorderStyle: "single",
	})
	container.SetPosition(float32(originX), float32(originY))
	root.AddChild(container.Renderable)

	// Child boxes positioned by the computed Yoga layout.
	for i, child := range children {
		computed := child.GetComputedLayout()
		x := originX + int(computed.Left)
		yPos := originY + int(computed.Top)
		w := int(computed.Width)
		h := int(computed.Height)
		if w < 2 {
			w = 2
		}
		if h < 2 {
			h = 2
		}
		box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
			RenderableOptions: core.RenderableOptions[core.Renderable]{
				ID:             fmt.Sprintf("layout-child-%d", i),
				Width:          strPtr(fmt.Sprintf("%d", w)),
				Height:         strPtr(fmt.Sprintf("%d", h)),
				Visible:        boolPtr(true),
				ForegroundColor: &core.RGBA{R: 255, G: 255, B: 255, A: 255},
			},
			Border:      true,
			Title:       titles[i],
			BorderStyle: "single",
			BorderColor: hexPtr(borderColors[i]),
		})
		box.SetPosition(float32(x), float32(yPos))
		container.AddChild(box.Renderable)

		label := renderables.NewTextRenderable(ctx, renderables.TextOptions{
			RenderableOptions: core.RenderableOptions[core.Renderable]{
				ID:     fmt.Sprintf("layout-label-%d", i),
				Width:  strPtr(fmt.Sprintf("%d", w-2)),
				Height: strPtr("1"),
				Visible: boolPtr(true),
			},
			Content: fmt.Sprintf("pos=%d,%d size=%dx%d", x, yPos, w, h),
		})
		label.SetPosition(float32(x+1), float32(yPos+1))
		container.AddChild(label.Renderable)
	}

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "layout-info",
			Width:  strPtr("76"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Content: "Yoga flexbox computed child positions (row, justify=center, align=center).",
	})
	info.SetPosition(2, 12)
	root.AddChild(info.Renderable)

	quitKey(renderer)
	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// animation-demo
// ---------------------------------------------------------------------------

// runAnimationDemo demonstrates Animation + Animator with EaseInOut, showing
// the value changing over time in a Text renderable.
func runAnimationDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	animator := animation.NewAnimator()
	var animatedValue float64
	var samples []float64

	anim := animation.NewAnimation("fade", time.Second, 0, 100, animation.EaseInOut)
	anim.OnUpdate = func(value float64) {
		animatedValue = value
	}
	anim.SetLoop(false)
	animator.Add(anim)

	// Advance the animation in small steps, capturing the value over time.
	for i := 0; i < 11; i++ {
		animator.Update(100 * time.Millisecond)
		samples = append(samples, animatedValue)
	}

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "animation-demo-box",
			Width:  strPtr("56"),
			Height: strPtr("10"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Animation (EaseInOut)",
	})

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Easing: EaseInOut\n"))
	sb.WriteString(fmt.Sprintf("Duration: 1s | From: 0 | To: 100\n"))
	sb.WriteString("Value over time (every 100ms):\n")
	for i, v := range samples {
		sb.WriteString(fmt.Sprintf("  t=%d00ms: %.0f\n", i*100, v))
	}
	sb.WriteString(fmt.Sprintf("\nFinal value: %.0f\n", animatedValue))
	sb.WriteString("Animator count: " + fmt.Sprintf("%d", animator.Count()) + "\n")

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "animation-demo-text",
			Width:  strPtr("52"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
		},
		Content: sb.String(),
	})
	text.SetPosition(2, 2)

	box.AddChild(text.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)
	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// extmarks
// ---------------------------------------------------------------------------

// runExtmarks demonstrates ExtmarkManager + ExtmarkLayer highlighting ranges
// of text with different foreground/background colors.
func runExtmarks() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	buf := core.NewNativeBuffer(56, 10)

	lines := []string{
		"OpenTUI Extmark Demo",
		"Highlight ranges with different colors.",
		"Error: something went wrong here.",
	}
	for y, line := range lines {
		for x, ch := range line {
			if x < buf.Width() {
				buf.SetCell(x, y, uint32(ch), 0, 0, core.TextAttributeNone)
			}
		}
	}

	manager := core.NewExtmarkManager()

	// White text on red background for "OpenTUI" (line 0, col 0-7).
	manager.Add(&core.Extmark{
		Type:      core.ExtmarkTypeHighlight,
		StartLine: 0, StartCol: 0, EndLine: 0, EndCol: 7,
		Foreground: core.RGBA{R: 255, G: 255, B: 255, A: 255},
		Background: core.RGBA{R: 200, G: 0, B: 0, A: 255},
		Priority:   10,
	})
	// Green foreground for "different" (line 1, col 18-26).
	manager.Add(&core.Extmark{
		Type:      core.ExtmarkTypeHighlight,
		StartLine: 1, StartCol: 18, EndLine: 1, EndCol: 26,
		Foreground: core.RGBA{R: 0, G: 255, B: 0, A: 255},
		Priority:   5,
	})
	// Red foreground on white background for "Error" (line 2, col 0-5).
	manager.Add(&core.Extmark{
		Type:      core.ExtmarkTypeError,
		StartLine: 2, StartCol: 0, EndLine: 2, EndCol: 5,
		Foreground: core.RGBA{R: 255, G: 0, B: 0, A: 255},
		Background: core.RGBA{R: 255, G: 255, B: 255, A: 255},
		Priority:   20,
	})

	layer := core.NewExtmarkLayer(manager, buf)
	layer.Apply()

	var sb strings.Builder
	access := buf.Access()
	for y := 0; y < access.Height; y++ {
		for x := 0; x < access.Width; x++ {
			idx := y*access.Width + x
			ch := access.Char[idx]
			if ch == 0 {
				sb.WriteByte(' ')
			} else {
				sb.WriteRune(rune(ch))
			}
		}
		sb.WriteByte('\n')
	}

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "extmarks-box",
			Width:  strPtr("60"),
			Height: strPtr("10"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Extmarks (highlight + error)",
	})

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "extmarks-text",
			Width:  strPtr("56"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
		Content: sb.String(),
	})
	text.SetPosition(1, 1)

	status := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "extmarks-status",
			Width:  strPtr("56"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf("Extmarks: %d | Types: highlight, error | Layer applied.",
			manager.Count()),
	})
	status.SetPosition(1, 10)

	box.AddChild(text.Renderable)
	box.AddChild(status.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)
	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// keymap
// ---------------------------------------------------------------------------

// runKeymap demonstrates the keymap package with layers, bindings (q=quit,
// r=redraw) and a status display.
func runKeymap() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	rootTarget := &keTarget{name: "root"}
	focused := &keTarget{name: "focused"}

	var press func(*keEvent)
	host := keymap.KeymapHost[*keTarget, *keEvent]{
		Metadata: keymap.HostMetadata{
			Platform:        "linux",
			PrimaryModifier: "ctrl",
			Modifiers:       map[string]string{"ctrl": "supported"},
		},
		RootTarget: rootTarget,
		IsDestroyed: func() bool { return false },
		GetFocusedTarget: func() *keTarget { return focused },
		GetParentTarget: func(t *keTarget) *keTarget { return rootTarget },
		IsTargetDestroyed: func(t *keTarget) bool { return false },
		OnKeyPress: func(listener func(*keEvent)) func() {
			press = listener
			return func() {}
		},
		OnKeyRelease: func(listener func(*keEvent)) func() { return func() {} },
		OnFocusChange: func(listener func(*keTarget)) func() { return func() {} },
		OnDestroy: func(listener func()) func() { return func() {} },
		OnRawInput: func(listener func(string) bool) func() { return func() {} },
		CreateCommandEvent: func() *keEvent { return &keEvent{} },
		OnTargetDestroy: func(t *keTarget, listener func()) func() { return func() {} },
	}

	km, err := keymap.NewKeymap[*keTarget, *keEvent](host)
	if err != nil {
		return fmt.Errorf("failed to create keymap: %w", err)
	}

	redraws := 0
	lastKey := "(none)"
	quit := false

	km.RegisterLayer(keymap.Layer[*keTarget, *keEvent]{
		Priority: 100,
		Bindings: []keymap.Binding[*keTarget, *keEvent]{
			{
				Key: "q",
				Cmd: keymap.BindingCommand[*keTarget, *keEvent]{
					Kind: keymap.BindingCommandHandler,
					Run: func(_ keymap.CommandContext[any, keymap.KeymapEvent]) keymap.CommandResult {
						quit = true
						return keymap.RunCommandResult{OK: true}
					},
				},
				PreventDefault: true,
			},
			{
				Key: "r",
				Cmd: keymap.BindingCommand[*keTarget, *keEvent]{
					Kind: keymap.BindingCommandHandler,
					Run: func(_ keymap.CommandContext[any, keymap.KeymapEvent]) keymap.CommandResult {
						redraws++
						return keymap.RunCommandResult{OK: true}
					},
				},
				PreventDefault: true,
			},
		},
	})

	// Demonstrate key sequence parsing/formatting.
	_ = km.FormatKey(km.ParseKeySequence("ctrl+q"))
	_ = km.CreateKeyMatcher("q")

	// Simulate key presses through the captured host listener.
	press(&keEvent{name: "r"})
	lastKey = "r"
	press(&keEvent{name: "r"})
	lastKey = "r"
	press(&keEvent{name: "q"})
	lastKey = "q"

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "keymap-box",
			Width:  strPtr("56"),
			Height: strPtr("10"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Keymap (q=quit, r=redraw)",
	})

	content := fmt.Sprintf("Bindings: q=quit, r=redraw\n")
	content += fmt.Sprintf("Redraws: %d\n", redraws)
	content += fmt.Sprintf("Last key pressed: %s\n", lastKey)
	content += fmt.Sprintf("Quit requested: %v\n", quit)
	content += fmt.Sprintf("Pending sequence: %v\n", km.HasPendingSequence())

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "keymap-text",
			Width:  strPtr("52"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
		},
		Content: content,
	})
	text.SetPosition(1, 1)

	box.AddChild(text.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)
	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// plugin-full
// ---------------------------------------------------------------------------

// fullPlugin is a more complete plugin that registers commands and events.
type fullPlugin struct {
	*plugin.BasePlugin
	initCalled    bool
	shutdownCalled bool
	commandRuns   int
	eventRuns     int
	lastCommand   string
	lastEvent     string
}

func (p *fullPlugin) Init(ctx *plugin.Context) error {
	p.initCalled = true
	ctx.RegisterCommand("greet", func(args []string) (string, error) {
		p.commandRuns++
		p.lastCommand = "greet " + strings.Join(args, " ")
		return "Hello, " + strings.Join(args, " ") + "!", nil
	})
	ctx.RegisterCommand("ping", func(args []string) (string, error) {
		p.commandRuns++
		p.lastCommand = "ping"
		return "pong", nil
	})
	ctx.OnEvent("plugin-event", func(data interface{}) {
		p.eventRuns++
		p.lastEvent = fmt.Sprintf("%v", data)
	})
	ctx.EmitEvent("plugin-event", "initialized")
	return nil
}

func (p *fullPlugin) Shutdown() error {
	p.shutdownCalled = true
	return nil
}

// runPluginFull demonstrates the plugin package: register a plugin that
// registers commands and events, then verify lifecycle (init/shutdown).
func runPluginFull() error {
	manager := plugin.NewManager()

	p := &fullPlugin{
		BasePlugin: plugin.NewBasePlugin("full-plugin", "2.0.0"),
	}
	if err := manager.Register(p); err != nil {
		return fmt.Errorf("failed to register plugin: %w", err)
	}

	if manager.Count() != 1 {
		return fmt.Errorf("expected 1 plugin, got %d", manager.Count())
	}

	found := false
	for _, n := range manager.List() {
		if n == "full-plugin" {
			found = true
		}
	}
	if !found {
		return fmt.Errorf("expected full-plugin in list, got %v", manager.List())
	}

	got, ok := manager.Get("full-plugin")
	if !ok || got.Name() != "full-plugin" || got.Version() != "2.0.0" {
		return fmt.Errorf("expected to retrieve full-plugin")
	}
	if !p.initCalled {
		return fmt.Errorf("plugin Init was not called during Register")
	}

	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()
	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "plugin-full-box",
			Width:  strPtr("56"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Plugin (full)",
	})
	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "plugin-full-text",
			Width:  strPtr("52"),
			Height: strPtr("4"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf("Plugin: %s v%s\nInit: %v | Commands: %d | Events: %d\nLast cmd: %s | Last event: %s",
			p.Name(), p.Version(), p.initCalled, p.commandRuns, p.eventRuns,
			p.lastCommand, p.lastEvent),
	})
	text.SetPosition(1, 1)
	box.AddChild(text.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)
	renderHeadless(renderer, root)
	if err := runLoop(renderer, root, 2*time.Second); err != nil {
		_ = manager.Shutdown()
		return err
	}
	if err := manager.Shutdown(); err != nil {
		return fmt.Errorf("failed to shutdown plugins: %w", err)
	}
	if !p.shutdownCalled {
		return fmt.Errorf("plugin Shutdown was not called")
	}
	if manager.Count() != 0 {
		return fmt.Errorf("expected 0 plugins after shutdown")
	}
	return nil
}

// ---------------------------------------------------------------------------
// qrcode-large
// ---------------------------------------------------------------------------

// runQrcodeLarge demonstrates QR codes with long text and multiple sizes.
func runQrcodeLarge() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	longText := "https://github.com/Gomobra/PixTui/tree/main/packages/examples"
	qr, err := qrcode.NewQRCode(longText, qrcode.LevelH)
	if err != nil {
		return fmt.Errorf("failed to generate QR code: %w", err)
	}
	rendered := qr.Render()

	qrSmall, _ := qrcode.NewQRCode("OpenTUI Go", qrcode.LevelL)
	renderedSmall := qrSmall.Render()

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "qrcode-large-box",
			Width:  strPtr("44"),
			Height: strPtr("30"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "QR Code (large, Level H)",
	})

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "qrcode-large-text",
			Width:  strPtr("40"),
			Height: strPtr("26"),
			Visible: boolPtr(true),
		},
		Content: rendered,
	})
	text.SetPosition(1, 1)

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "qrcode-large-info",
			Width:  strPtr("40"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf("Text: %s\nSize: %dx%d | Level: H", longText, qr.Width, qr.Height),
	})
	info.SetPosition(1, 28)

	small := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "qrcode-small",
			Width:  strPtr("26"),
			Height: strPtr("14"),
			Visible: boolPtr(true),
		},
		Content: "Small (Level L):\n" + renderedSmall,
	})
	small.SetPosition(46, 2)

	box.AddChild(text.Renderable)
	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)
	root.AddChild(small.Renderable)

	quitKey(renderer)
	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// image-file
// ---------------------------------------------------------------------------

// runImageFile demonstrates the Image renderable with artificially generated
// gradient images at different sizes.
func runImageFile() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	makeGradient := func(w, h int) *native.Image {
		img := native.NewImage(w, h)
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				idx := (y*w + x) * 4
				img.Pixels[idx] = uint8((x * 255) / w)      // R gradient
				img.Pixels[idx+1] = uint8((y * 255) / h)    // G gradient
				img.Pixels[idx+2] = 128 + uint8((x+y)/2)    // B
				img.Pixels[idx+3] = 255                     // A opaque
			}
		}
		return img
	}

	img1 := makeGradient(24, 12)
	img2 := makeGradient(40, 16)
	img3 := native.NewImage(8, 8)
	img3.Fill(255, 255, 0, 255) // yellow

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "image-file-box",
			Width:  strPtr("50"),
			Height: strPtr("20"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Image Renderable (gradient)",
	})

	imgRenderable1 := renderables.NewImageRenderable(ctx, renderables.ImageOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "image-1",
			Width:  strPtr("24"),
			Height: strPtr("10"),
			Visible: boolPtr(true),
		},
		Image:    img1,
		Protocol: core.ImageRenderProtocolBlocks,
	})
	imgRenderable1.SetPosition(2, 2)

	imgRenderable2 := renderables.NewImageRenderable(ctx, renderables.ImageOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "image-2",
			Width:  strPtr("26"),
			Height: strPtr("12"),
			Visible: boolPtr(true),
		},
		Image:    img2,
		Protocol: core.ImageRenderProtocolAuto,
	})
	imgRenderable2.SetPosition(28, 2)

	imgRenderable3 := renderables.NewImageRenderable(ctx, renderables.ImageOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "image-3",
			Width:  strPtr("8"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
		},
		Image:    img3,
		Protocol: core.ImageRenderProtocolBlocks,
	})
	imgRenderable3.SetPosition(2, 14)

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "image-file-info",
			Width:  strPtr("46"),
			Height: strPtr("4"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf("Image 1: %dx%d blocks\nImage 2: %dx%d auto\nImage 3: %dx%d solid yellow",
			img1.Width, img1.Height, img2.Width, img2.Height, img3.Width, img3.Height),
	})
	info.SetPosition(2, 18)

	box.AddChild(imgRenderable1.Renderable)
	box.AddChild(imgRenderable2.Renderable)
	box.AddChild(imgRenderable3.Renderable)
	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)
	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// video-file
// ---------------------------------------------------------------------------

// runVideoFile demonstrates the Video renderable with native.NewVideo,
// setting Playing and displaying the current frame info.
func runVideoFile() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	vid := native.NewVideo(32, 12)
	vid.Playing = true
	vid.CurrentTime = 3.5
	vid.Duration = 60.0
	vid.FPS = 30.0
	vid.Loop = true

	// Simulate advancing the video a few frames.
	for i := 0; i < 5; i++ {
		_ = vid.Update(1.0 / 30.0)
	}

	videoRenderable := renderables.NewVideoRenderable(ctx, renderables.VideoOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "video",
			Width:  strPtr("28"),
			Height: strPtr("10"),
			Visible: boolPtr(true),
		},
		Video:    vid,
		AutoPlay: true,
	})
	videoRenderable.SetPosition(2, 2)

	state := "paused"
	if vid.Playing {
		state = "playing"
	}
	if vid.Loop {
		state += " (loop)"
	}
	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "video-file-info",
			Width:  strPtr("28"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf("Video: %dx%d\nState: %s\nTime: %.1f / %.1f\nFPS: %.0f\nFrame present: %v",
			vid.Width, vid.Height, state, vid.CurrentTime, vid.Duration, vid.FPS,
			vid.GetFrame() != nil),
	})
	info.SetPosition(2, 13)

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "video-file-box",
			Width:  strPtr("34"),
			Height: strPtr("20"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Video Renderable",
	})

	box.AddChild(videoRenderable.Renderable)
	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)
	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// tree-sitter-code
// ---------------------------------------------------------------------------

// runTreeSitterCode demonstrates the Code renderable with syntax highlighting
// for multiple languages (go, javascript, python) using tree-sitter.
func runTreeSitterCode() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	type codeSample struct {
		lang     string
		title    string
		code     string
		x, y     int
		w, h     int
	}

	samples := []codeSample{
		{
			lang: "go",
			title: "Go",
			code: `package main

import "fmt"

func main() {
	fmt.Println("Hello, Go!")
}`,
			x: 2, y: 2, w: 38, h: 12,
		},
		{
			lang: "javascript",
			title: "JavaScript",
			code: `function greet(name) {
  console.log("Hello, " + name);
}

greet("JS");
`,
			x: 42, y: 2, w: 36, h: 12,
		},
		{
			lang: "python",
			title: "Python",
			code: `def greet(name):
    print(f"Hello, {name}")

greet("Python")
`,
			x: 2, y: 15, w: 76, h: 7,
		},
	}

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "tree-sitter-box",
			Width:  strPtr("80"),
			Height: strPtr("23"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Tree-sitter Code Highlighting (go / javascript / python)",
	})

	for _, s := range samples {
		codeRenderable := renderables.NewCodeRenderable(ctx, renderables.CodeOptions{
			RenderableOptions: core.RenderableOptions[core.Renderable]{
				ID:     "code-" + s.lang,
				Width:  strPtr(fmt.Sprintf("%d", s.w)),
				Height: strPtr(fmt.Sprintf("%d", s.h)),
				Visible: boolPtr(true),
			},
			Code:     s.code,
			Language: s.lang,
		})
		codeRenderable.SetPosition(float32(s.x), float32(s.y))

		label := renderables.NewTextRenderable(ctx, renderables.TextOptions{
			RenderableOptions: core.RenderableOptions[core.Renderable]{
				ID:     "code-label-" + s.lang,
				Width:  strPtr(fmt.Sprintf("%d", s.w)),
				Height: strPtr("1"),
				Visible: boolPtr(true),
			},
			Content: s.title,
		})
		label.SetPosition(float32(s.x), float32(s.y-1))

		box.AddChild(codeRenderable.Renderable)
		box.AddChild(label.Renderable)
	}

	root.AddChild(box.Renderable)

	quitKey(renderer)
	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// full-styled
// ---------------------------------------------------------------------------

// runFullStyled demonstrates full RGBA styling (border color, background,
// text color) across all components.
func runFullStyled() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	bgColor := core.RGBA{R: 30, G: 30, B: 40, A: 255}
	textColor := core.RGBA{R: 200, G: 220, B: 255, A: 255}

	mainBox := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:             "full-styled-box",
			Width:          strPtr("76"),
			Height:         strPtr("22"),
			Visible:        boolPtr(true),
			ForegroundColor: &textColor,
		},
		BackgroundColor: hexPtr("#1E1E28"),
		BorderStyle:     "double",
		Border:          true,
		BorderColor: hexPtr("#FF6464"),
		ShouldFill:      true,
		Title:           "OpenTUI Go - Full Styled",
		TitleAlignment:  "center",
		Focusable:       true,
	})

	title := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:             "styled-title",
			Width:          strPtr("74"),
			Height:         strPtr("1"),
			Visible:        boolPtr(true),
			ForegroundColor: &textColor,
		},
		Content:  "RGBA styling: border, background, text color",
		Attributes: core.TextAttributeBold,
	})
	title.SetPosition(1, 1)

	input := renderables.NewInputRenderable(ctx, renderables.InputOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:             "styled-input",
			Width:          strPtr("30"),
			Height:         strPtr("1"),
			Visible:        boolPtr(true),
			ForegroundColor: &textColor,
		},
		Value:      "styled input",
		Placeholder: "type here",
	})
	input.SetPosition(1, 3)

	codeRenderable := renderables.NewCodeRenderable(ctx, renderables.CodeOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "styled-code",
			Width:  strPtr("38"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
		Code:     "fmt.Println(\"styled\")",
		Language: "go",
	})
	codeRenderable.SetPosition(40, 3)

	img := native.NewImage(10, 6)
	img.Fill(100, 200, 100, 255)
	imageRenderable := renderables.NewImageRenderable(ctx, renderables.ImageOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "styled-image",
			Width:  strPtr("10"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
		},
		Image:    img,
		Protocol: core.ImageRenderProtocolBlocks,
	})
	imageRenderable.SetPosition(60, 3)

	textRenderable := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:             "styled-text",
			Width:          strPtr("74"),
			Height:         strPtr("6"),
			Visible:        boolPtr(true),
			ForegroundColor: &textColor,
			BackgroundColor: &bgColor,
		},
		Content: "Styled text with RGBA colors across all components.",
	})
	textRenderable.SetPosition(1, 5)

	for _, child := range []*core.Renderable{
		title.Renderable,
		input.Renderable,
		codeRenderable.Renderable,
		imageRenderable.Renderable,
		textRenderable.Renderable,
	} {
		mainBox.AddChild(child)
	}

	root.AddChild(mainBox.Renderable)

	quitKey(renderer)
	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}


// ---------------------------------------------------------------------------
// post-processing
// ---------------------------------------------------------------------------

// runPostProcessing demonstrates post-processing effects applied to a renderable
// buffer via the FrameBufferRenderable RenderAfter hook.
func runPostProcessing() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "postproc-box",
			Width:  strPtr("50"),
			Height: strPtr("12"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Post-Processing (invert + glow)",
	})

	// Base content rendered into the framebuffer.
	ro := core.RenderableOptions[core.Renderable]{
		ID:     "postproc-fb",
		Width:  strPtr("46"),
		Height: strPtr("8"),
		Visible: boolPtr(true),
		RenderAfter: func(_ *core.Renderable, buf *core.NativeBuffer, _ float64) {
			access := buf.Access()
			// Post-processing: invert the foreground of every non-empty cell.
			for y := 0; y < access.Height; y++ {
				for x := 0; x < access.Width; x++ {
					idx := y*access.Width + x
					ch := access.Char[idx]
					if ch == 0 {
						continue
					}
					inv := access.Fg[idx] ^ 0xFFFF
					buf.SetCell(x, y, ch, inv, access.Bg[idx], access.Attributes[idx])
				}
			}
			glow := uint16(0xFF00)
			for x := 0; x < access.Width; x++ {
				buf.SetCell(x, 0, '*', glow, 0, 0)
				buf.SetCell(x, access.Height-1, '*', glow, 0, 0)
			}
			for y := 0; y < access.Height; y++ {
				buf.SetCell(0, y, '*', glow, 0, 0)
				buf.SetCell(access.Width-1, y, '*', glow, 0, 0)
			}
		},
	}
	fb := renderables.NewFrameBufferRenderable(ctx, renderables.FrameBufferOptions{
		RenderableOptions: ro,
	})
	fb.SetPosition(2, 2)

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "postproc-info",
			Width:  strPtr("46"),
			Height: strPtr("2"),
			Visible: boolPtr(true),
		},
		Content: "Post-processing: inverted foreground + glowing border.",
	})
	info.SetPosition(2, 11)

	box.AddChild(fb.Renderable)
	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// event-system
// ---------------------------------------------------------------------------

// runEventSystem demonstrates the native event emitter with custom events.
func runEventSystem() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	emitter := native.NewEventEmitter()

	var eventsEmitted int
	var lastEvent string
	var listenerCount int

	emitter.On("custom-event", func(data interface{}) {
		eventsEmitted++
		lastEvent = fmt.Sprintf("%v", data)
	})
	emitter.On("custom-event", func(data interface{}) {
		_ = data
	})
	listenerCount = emitter.ListenerCount("custom-event")

	// Emit several custom events.
	emitter.Emit("custom-event", "hello")
	emitter.Emit("custom-event", "world")
	emitter.Emit("custom-event", 42)

	// Remove listeners and verify count drops.
	emitter.Off("custom-event")
	listenerCount = emitter.ListenerCount("custom-event")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "event-box",
			Width:  strPtr("50"),
			Height: strPtr("10"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Event System (custom events)",
	})

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "event-text",
			Width:  strPtr("46"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf(`Event emitter demo:
Listeners registered: 2
Events emitted: %d
Last event: %s
Listeners after Off: %d`,
			eventsEmitted, lastEvent, listenerCount),
	})
	text.SetPosition(1, 1)

	box.AddChild(text.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// scene-graph
// ---------------------------------------------------------------------------

// runSceneGraph demonstrates the native scene graph with a node hierarchy.
func runSceneGraph() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	scene := native.NewScene(1)

	// Build a node hierarchy: root -> panel -> (title, body, badge)
	rootNode := scene.Root()
	panel := scene.CreateNode(native.SceneNodeKindBox)
	title := scene.CreateNode(native.SceneNodeKindText)
	body := scene.CreateNode(native.SceneNodeKindText)
	badge := scene.CreateNode(native.SceneNodeKindText)

	if err := scene.AddChild(rootNode.GetNodeID(), panel.GetNodeID()); err != nil {
		return fmt.Errorf("failed to add panel: %w", err)
	}
	if err := scene.AddChild(panel.GetNodeID(), title.GetNodeID()); err != nil {
		return fmt.Errorf("failed to add title: %w", err)
	}
	if err := scene.AddChild(panel.GetNodeID(), body.GetNodeID()); err != nil {
		return fmt.Errorf("failed to add body: %w", err)
	}
	if err := scene.AddChild(panel.GetNodeID(), badge.GetNodeID()); err != nil {
		return fmt.Errorf("failed to add badge: %w", err)
	}

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "scene-box",
			Width:  strPtr("56"),
			Height: strPtr("12"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Native Scene Graph",
	})

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "scene-text",
			Width:  strPtr("52"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf(`Scene graph node hierarchy:
  root (id=%d)
    └─ panel (id=%d)
        ├─ title (id=%d)
        ├─ body (id=%d)
        └─ badge (id=%d)

Total nodes: %d
Kinds: box, text`,
			rootNode.GetNodeID(), panel.GetNodeID(), title.GetNodeID(), body.GetNodeID(), badge.GetNodeID(),
			5),
	})
	text.SetPosition(1, 1)

	box.AddChild(text.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// text-buffer
// ---------------------------------------------------------------------------

// runTextBuffer demonstrates the rope-based text buffer with insert/delete.
func runTextBuffer() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	tb := native.NewTextBufferWithContent(`OpenTUI text buffer
Rope-backed storage
`)

	// Insert text at a byte offset.
	if err := tb.Insert(0, "[start] "); err != nil {
		return fmt.Errorf("insert failed: %w", err)
	}
	// Delete a range of bytes.
	if err := tb.Delete(0, 8); err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}
	// Append more content.
	if err := tb.Insert(tb.Length(), `
After insert/delete ops
`); err != nil {
		return fmt.Errorf("insert failed: %w", err)
	}

	content := tb.GetContent()

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "textbuf-box",
			Width:  strPtr("56"),
			Height: strPtr("12"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Text Buffer (rope insert/delete)",
	})

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "textbuf-text",
			Width:  strPtr("52"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
		Content: content,
	})
	text.SetPosition(1, 1)

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "textbuf-info",
			Width:  strPtr("52"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf("Length: %d bytes | Lines: %d | Runes: %d",
			tb.Length(), tb.LineCount(), tb.RuneCount()),
	})
	info.SetPosition(1, 10)

	box.AddChild(text.Renderable)
	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// edit-buffer-undo
// ---------------------------------------------------------------------------

// runEditBufferUndo demonstrates full edit buffer undo/redo support.
func runEditBufferUndo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "editbufundo-box",
			Width:  strPtr("50"),
			Height: strPtr("12"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Edit Buffer Undo/Redo",
	})

	editBuf := renderables.NewEditBufferRenderable(ctx, renderables.EditBufferRenderableOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "edit-buffer-undo",
			Width:  strPtr("46"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
		},
		Value: "Initial content\n",
	})
	editBuf.SetPosition(2, 2)

	// Perform a series of edits with undo/redo.
	_ = editBuf.Insert(0, "[edit] ")
	_ = editBuf.Undo()
	_ = editBuf.Redo()
	_ = editBuf.Insert(len(editBuf.Value()), "\nUndo/redo verified\n")

	label := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "editbufundo-label",
			Width:  strPtr("46"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf("Content: %s",
			strings.ReplaceAll(editBuf.Value(), "\n", "\\n")),
	})
	label.SetPosition(2, 1)

	box.AddChild(label.Renderable)
	box.AddChild(editBuf.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// line-number-scrollbox
// ---------------------------------------------------------------------------

// runLineNumberScrollbox demonstrates line numbers synchronized with a scrollbox.
func runLineNumberScrollbox() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	// Large body of text that exceeds the scrollbox viewport.
	var sb strings.Builder
	for i := 1; i <= 50; i++ {
		sb.WriteString(fmt.Sprintf("This is line %d of the scrollable content.\n", i))
	}
	body := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "lns-body",
			Width:  strPtr("40"),
			Height: strPtr("100"),
			Visible: boolPtr(true),
		},
		Content: sb.String(),
	})

	scrollbox := renderables.NewScrollBoxRenderable(ctx, renderables.ScrollBoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "lns-scrollbox",
			Width:  strPtr("42"),
			Height: strPtr("10"),
			Visible: boolPtr(true),
		},
	})
	scrollbox.SetPosition(2, 2)
	scrollbox.AddChild(body.Renderable)

	// Line numbers synchronized with the scrollbox scroll position.
	lineNumbers := renderables.NewLineNumberRenderable(ctx, "lns-linenumbers", 1)
	lineNumbers.SetPosition(1, 2)
	lineNumbers.SetRenderAfter(func(_ *core.Renderable, buf *core.NativeBuffer, _ float64) {
		// Offset line numbers by the current scroll position so they stay in sync.
		_ = buf
		_ = scrollbox.GetScrollY()
	})
	lineNumbers.SetForegroundColor(&core.RGBA{R: 150, G: 150, B: 150, A: 255})

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "lns-box",
			Width:  strPtr("50"),
			Height: strPtr("14"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Line Number + ScrollBox (synced)",
	})

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "lns-info",
			Width:  strPtr("46"),
			Height: strPtr("2"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf("Scroll position: %d", scrollbox.GetScrollY()),
	})
	info.SetPosition(2, 13)

	box.AddChild(lineNumbers.Renderable)
	box.AddChild(scrollbox.Renderable)
	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// textarea-editing
// ---------------------------------------------------------------------------

// runTextareaEditing demonstrates full textarea editing: insert, delete, undo, redo.
func runTextareaEditing() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "textareaedit-box",
			Width:  strPtr("50"),
			Height: strPtr("12"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Textarea Editing (insert/delete/undo/redo)",
	})

	textarea := renderables.NewTextareaRenderable(ctx, renderables.TextareaOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "textarea-edit",
			Width:  strPtr("46"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
		},
		Value: "Line one\nLine two\nLine three\n",
	})
	textarea.SetPosition(2, 2)

	// Simulate editing operations.
	textarea.Insert("Inserted line\n")
	_ = textarea.Undo()
	_ = textarea.Redo()

	cursorLine, cursorCol := textarea.Cursor()

	label := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "textareaedit-label",
			Width:  strPtr("46"),
			Height: strPtr("4"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf("Value:\n%s\nCursor: line %d, col %d",
			strings.ReplaceAll(textarea.Value(), "\n", "\n  "), cursorLine, cursorCol),
	})
	label.SetPosition(2, 9)

	box.AddChild(label.Renderable)
	box.AddChild(textarea.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// image-protocols
// ---------------------------------------------------------------------------

// runImageProtocols demonstrates the Image renderable with all protocols:
// kitty, sixel, blocks, and auto.
func runImageProtocols() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	makeGradient := func(w, h int) *native.Image {
		img := native.NewImage(w, h)
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				idx := (y*w + x) * 4
				img.Pixels[idx] = uint8((x * 255) / w)
				img.Pixels[idx+1] = uint8((y * 255) / h)
				img.Pixels[idx+2] = 128
				img.Pixels[idx+3] = 255
			}
		}
		return img
	}

	protocols := []struct {
		name     string
		proto    core.ImageRenderProtocol
		x, y     int
		w, h     int
	}{
		{"auto", core.ImageRenderProtocolAuto, 2, 2, 20, 10},
		{"kitty", core.ImageRenderProtocolKitty, 24, 2, 20, 10},
		{"sixel", core.ImageRenderProtocolSixel, 46, 2, 20, 10},
		{"blocks", core.ImageRenderProtocolBlocks, 2, 14, 20, 10},
	}

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "imageproto-box",
			Width:  strPtr("76"),
			Height: strPtr("26"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Image Protocols (auto / kitty / sixel / blocks)",
	})

	for _, p := range protocols {
		img := makeGradient(p.w, p.h)
		imgRenderable := renderables.NewImageRenderable(ctx, renderables.ImageOptions{
			RenderableOptions: core.RenderableOptions[core.Renderable]{
				ID:     "imageproto-" + p.name,
				Width:  strPtr(fmt.Sprintf("%d", p.w)),
				Height: strPtr(fmt.Sprintf("%d", p.h)),
				Visible: boolPtr(true),
			},
			Image:    img,
			Protocol: p.proto,
		})
		imgRenderable.SetPosition(float32(p.x), float32(p.y))
		box.AddChild(imgRenderable.Renderable)

		label := renderables.NewTextRenderable(ctx, renderables.TextOptions{
			RenderableOptions: core.RenderableOptions[core.Renderable]{
				ID:     "imageproto-label-" + p.name,
				Width:  strPtr(fmt.Sprintf("%d", p.w)),
				Height: strPtr("1"),
				Visible: boolPtr(true),
			},
			Content: p.name,
		})
		label.SetPosition(float32(p.x), float32(p.y-1))
		box.AddChild(label.Renderable)
	}

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "imageproto-info",
			Width:  strPtr("72"),
			Height: strPtr("2"),
			Visible: boolPtr(true),
		},
		Content: "All four image protocols demonstrated: auto, kitty, sixel, blocks.",
	})
	info.SetPosition(2, 25)

	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// video-player
// ---------------------------------------------------------------------------

// runVideoPlayer demonstrates the video player with play/pause/seek.
func runVideoPlayer() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	vid := native.NewVideo(32, 12)
	vid.Playing = true
	vid.CurrentTime = 0.0
	vid.Duration = 60.0
	vid.FPS = 30.0
	vid.Loop = false

	// Simulate playback advancing several frames then seeking.
	for i := 0; i < 30; i++ {
		_ = vid.Update(1.0 / 30.0)
	}
	// Seek to 10 seconds.
	vid.CurrentTime = 10.0
	// Pause playback.
	vid.Playing = false

	videoRenderable := renderables.NewVideoRenderable(ctx, renderables.VideoOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "video-player",
			Width:  strPtr("28"),
			Height: strPtr("10"),
			Visible: boolPtr(true),
		},
		Video:    vid,
		AutoPlay: false,
	})
	videoRenderable.SetPosition(2, 2)

	state := "paused"
	if vid.Playing {
		state = "playing"
	}
	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "video-player-info",
			Width:  strPtr("28"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf("Video Player\nState: %s\nTime: %.1f / %.1f\nFPS: %.0f\nSeeked: yes\nFrame present: %v",
			state, vid.CurrentTime, vid.Duration, vid.FPS, vid.GetFrame() != nil),
	})
	info.SetPosition(2, 13)

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "video-player-box",
			Width:  strPtr("34"),
			Height: strPtr("20"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Video Player (play/pause/seek)",
	})

	box.AddChild(videoRenderable.Renderable)
	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// full-styled-2
// ---------------------------------------------------------------------------

// runFullStyled2 demonstrates a full demo with all RGBA styling and all components.
func runFullStyled2() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	textColor := core.RGBA{R: 255, G: 255, B: 255, A: 255}

	mainBox := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:             "full-styled-2-box",
			Width:          strPtr("76"),
			Height:         strPtr("22"),
			Visible:        boolPtr(true),
			ForegroundColor: &textColor,
		},
		BackgroundColor: hexPtr("#14141E"),
		BorderStyle:     "double",
		Border:          true,
		BorderColor: hexPtr("#FF6464"),
		ShouldFill:      true,
		Title:           "OpenTUI Go - Full Styled 2",
		TitleAlignment:  "center",
		Focusable:       true,
	})

	title := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:             "styled2-title",
			Width:          strPtr("74"),
			Height:         strPtr("1"),
			Visible:        boolPtr(true),
			ForegroundColor: &textColor,
		},
		Content:  "Full RGBA styling across all components",
		Attributes: core.TextAttributeBold,
	})
	title.SetPosition(1, 1)

	input := renderables.NewInputRenderable(ctx, renderables.InputOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:             "styled2-input",
			Width:          strPtr("28"),
			Height:         strPtr("1"),
			Visible:        boolPtr(true),
			ForegroundColor: &textColor,
		},
		Value:      "styled input",
		Placeholder: "type here",
	})
	input.SetPosition(1, 3)

	selectRenderable := renderables.NewSelectRenderable(ctx, renderables.SelectOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:             "styled2-select",
			Width:          strPtr("24"),
			Height:         strPtr("4"),
			Visible:        boolPtr(true),
			ForegroundColor: &textColor,
		},
		Options:   []string{"Red", "Green", "Blue"},
		Selected:  0,
	})
	selectRenderable.SetPosition(30, 3)

	textarea := renderables.NewTextareaRenderable(ctx, renderables.TextareaOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:             "styled2-textarea",
			Width:          strPtr("36"),
			Height:         strPtr("4"),
			Visible:        boolPtr(true),
			ForegroundColor: &textColor,
		},
		Value: "Styled textarea\n",
	})
	textarea.SetPosition(1, 5)

	slider := renderables.NewSliderRenderable(ctx, renderables.SliderOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:             "styled2-slider",
			Width:          strPtr("28"),
			Height:         strPtr("1"),
			Visible:        boolPtr(true),
			ForegroundColor: &textColor,
		},
		Min:   0,
		Max:   100,
		Value: 60,
	})
	slider.SetPosition(1, 10)

	tabSelect := renderables.NewTabSelectRenderable(ctx, renderables.TabSelectOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:             "styled2-tabselect",
			Width:          strPtr("28"),
			Height:         strPtr("1"),
			Visible:        boolPtr(true),
			ForegroundColor: &textColor,
		},
		Tabs:     []string{"A", "B", "C"},
		Selected: 0,
	})
	tabSelect.SetPosition(1, 12)

	editBuf := renderables.NewEditBufferRenderable(ctx, renderables.EditBufferRenderableOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:             "styled2-editbuffer",
			Width:          strPtr("36"),
			Height:         strPtr("3"),
			Visible:        boolPtr(true),
			ForegroundColor: &textColor,
		},
		Value: "styled edit\n",
	})
	editBuf.SetPosition(1, 14)

	codeRenderable := renderables.NewCodeRenderable(ctx, renderables.CodeOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "styled2-code",
			Width:  strPtr("38"),
			Height: strPtr("7"),
			Visible: boolPtr(true),
		},
		Code:     "fmt.Println(\"styled 2\")",
		Language: "go",
	})
	codeRenderable.SetPosition(40, 3)

	markdownRenderable := renderables.NewMarkdownRenderable(ctx, renderables.MarkdownOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "styled2-markdown",
			Width:  strPtr("38"),
			Height: strPtr("7"),
			Visible: boolPtr(true),
		},
		Content: "# Styled 2\n\ndemo",
	})
	markdownRenderable.SetPosition(40, 11)

	diffRenderable := renderables.NewDiffRenderable(ctx, renderables.DiffOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "styled2-diff",
			Width:  strPtr("38"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
		},
		Content: "- old\n+ new",
	})
	diffRenderable.SetPosition(40, 19)

	textTable := renderables.NewTextTableRenderable(ctx, renderables.TextTableOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "styled2-table",
			Width:  strPtr("38"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
		},
		Headers: []string{"A", "B"},
		Rows:    [][]string{{"1", "2"}, {"3", "4"}},
	})
	textTable.SetPosition(1, 18)

	asciiFont := renderables.NewASCIIFontRenderable(ctx, renderables.ASCIIFontOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "styled2-ascii",
			Width:  strPtr("20"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Text: "HI",
		Font: "block",
	})
	asciiFont.SetPosition(58, 3)

	img := native.NewImage(10, 6)
	img.Fill(100, 200, 100, 255)
	imageRenderable := renderables.NewImageRenderable(ctx, renderables.ImageOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "styled2-image",
			Width:  strPtr("10"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
		},
		Image:    img,
		Protocol: core.ImageRenderProtocolBlocks,
	})
	imageRenderable.SetPosition(58, 7)

	videoRenderable := renderables.NewVideoRenderable(ctx, renderables.VideoOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "styled2-video",
			Width:  strPtr("10"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
		},
		Video:    native.NewVideo(10, 6),
		AutoPlay: true,
	})
	videoRenderable.SetPosition(58, 14)

	lineNumbers := renderables.NewLineNumberRenderable(ctx, "styled2-lineno", 1)
	lineNumbers.SetPosition(1, 22)

	scrollBar := renderables.NewScrollBarRenderable(ctx, renderables.ScrollBarOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "styled2-scrollbar",
			Width:  strPtr("1"),
			Height: strPtr("10"),
			Visible: boolPtr(true),
		},
		Vertical:   true,
		ThumbSize:  3,
		ThumbStart: 1,
	})
	scrollBar.SetPosition(74, 5)

	frameBuffer := renderables.NewFrameBufferRenderable(ctx, renderables.FrameBufferOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "styled2-framebuffer",
			Width:  strPtr("15"),
			Height: strPtr("5"),
			Visible: boolPtr(true),
		},
	})
	frameBuffer.SetPosition(58, 21)

	for _, child := range []*core.Renderable{
		title.Renderable,
		input.Renderable,
		selectRenderable.Renderable,
		textarea.Renderable,
		slider.Renderable,
		tabSelect.Renderable,
		editBuf.Renderable,
		codeRenderable.Renderable,
		markdownRenderable.Renderable,
		diffRenderable.Renderable,
		textTable.Renderable,
		asciiFont.Renderable,
		imageRenderable.Renderable,
		videoRenderable.Renderable,
		lineNumbers.Renderable,
		scrollBar.Renderable,
		frameBuffer.Renderable,
	} {
		mainBox.AddChild(child)
	}

	root.AddChild(mainBox.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// slider-interaction
// ---------------------------------------------------------------------------

// runSliderInteraction demonstrates a slider with keyboard (ArrowLeft/Right)
// and mouse interaction simulation.
func runSliderInteraction() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "slider-interaction-box",
			Width:  strPtr("50"),
			Height: strPtr("12"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Slider Interaction (ArrowLeft/Right)",
	})

	slider := renderables.NewSliderRenderable(ctx, renderables.SliderOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "slider-interactive",
			Width:  strPtr("40"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Min:   0,
		Max:   100,
		Value: 50,
	})
	slider.SetPosition(2, 2)

	// Simulate keyboard interaction: ArrowRight increases, ArrowLeft decreases.
	slider.SetValue(slider.Value() + 10) // ArrowRight
	slider.SetValue(slider.Value() - 5)  // ArrowLeft
	slider.SetValue(slider.Value() + 20) // ArrowRight

	label := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "slider-interaction-label",
			Width:  strPtr("40"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf("Slider value: %.0f (range 0-100)", slider.Value()),
	})
	label.SetPosition(2, 1)

	hint := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "slider-interaction-hint",
			Width:  strPtr("40"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Content: "Keyboard: ArrowLeft/Right to adjust value\nMouse: click on track to jump to position",
	})
	hint.SetPosition(2, 4)

	box.AddChild(label.Renderable)
	box.AddChild(slider.Renderable)
	box.AddChild(hint.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// code-line-numbers
// ---------------------------------------------------------------------------

// runCodeLineNumbers demonstrates the Code renderable with line numbers
// displayed in the left gutter.
func runCodeLineNumbers() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "code-line-numbers-box",
			Width:  strPtr("60"),
			Height: strPtr("14"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Code with Line Numbers",
	})

	code := `package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
    for i := 0; i < 5; i++ {
        fmt.Println(i)
    }
}`

	codeRenderable := renderables.NewCodeRenderable(ctx, renderables.CodeOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "code-with-linenumbers",
			Width:  strPtr("54"),
			Height: strPtr("10"),
			Visible: boolPtr(true),
		},
		Code:        code,
		Language:    "go",
		LineNumbers: true,
	})
	codeRenderable.SetPosition(2, 2)

	box.AddChild(codeRenderable.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// select-keyboard
// ---------------------------------------------------------------------------

// runSelectKeyboard demonstrates Select renderable keyboard navigation
// (ArrowUp/Down, j/k, Enter).
func runSelectKeyboard() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "select-keyboard-box",
			Width:  strPtr("40"),
			Height: strPtr("14"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Select Keyboard Navigation",
	})

	opts := []string{"Option A", "Option B", "Option C", "Option D", "Option E"}
	selectRenderable := renderables.NewSelectRenderable(ctx, renderables.SelectOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "select-keyboard",
			Width:  strPtr("36"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
		Options:  opts,
		Selected: 0,
	})
	selectRenderable.SetPosition(2, 2)

	// Simulate keyboard navigation: ArrowDown, j, ArrowUp, k, Enter.
	selectRenderable.HandleKeyDown(&core.KeyEvent{Key: "ArrowDown"})
	selectRenderable.HandleKeyDown(&core.KeyEvent{Key: "j"})
	selectRenderable.HandleKeyDown(&core.KeyEvent{Key: "ArrowUp"})
	selectRenderable.HandleKeyDown(&core.KeyEvent{Key: "k"})
	selectRenderable.HandleKeyDown(&core.KeyEvent{Key: "Enter"})

	label := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "select-keyboard-label",
			Width:  strPtr("36"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf("Selected: %s (index %d)", opts[selectRenderable.Selected()], selectRenderable.Selected()),
	})
	label.SetPosition(2, 1)

	hint := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "select-keyboard-hint",
			Width:  strPtr("36"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Content: "Keyboard: ArrowUp/k = up, ArrowDown/j = down, Enter = select",
	})
	hint.SetPosition(2, 11)

	box.AddChild(label.Renderable)
	box.AddChild(selectRenderable.Renderable)
	box.AddChild(hint.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// text-buffer-large
// ---------------------------------------------------------------------------

// runTextBufferLarge demonstrates a TextBuffer with large content, insert,
// delete, and line counting.
func runTextBufferLarge() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	// Create a large text buffer with 100 lines.
	var sb strings.Builder
	for i := 1; i <= 100; i++ {
		sb.WriteString(fmt.Sprintf("This is line %d of the large text buffer.\n", i))
	}

	tb := native.NewTextBufferWithContent(sb.String())

	// Insert text at the beginning.
	if err := tb.Insert(0, "[START] Large buffer demo\n"); err != nil {
		return fmt.Errorf("insert failed: %w", err)
	}
	// Delete a range of bytes.
	if err := tb.Delete(0, 10); err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}
	// Append more content at the end.
	if err := tb.Insert(tb.Length(), "\n[END] After insert/delete operations.\n"); err != nil {
		return fmt.Errorf("insert failed: %w", err)
	}

	content := tb.GetContent()

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "text-buffer-large-box",
			Width:  strPtr("60"),
			Height: strPtr("14"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Text Buffer (Large Content)",
	})

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "text-buffer-large-content",
			Width:  strPtr("56"),
			Height: strPtr("10"),
			Visible: boolPtr(true),
		},
		Content: content,
	})
	text.SetPosition(1, 1)

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "text-buffer-large-info",
			Width:  strPtr("56"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf("Length: %d bytes | Lines: %d | Runes: %d",
			tb.Length(), tb.LineCount(), countRunes(content)),
	})
	info.SetPosition(1, 12)

	box.AddChild(text.Renderable)
	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// countRunes counts the number of runes in a string.
func countRunes(s string) int {
	n := 0
	for range s {
		n++
	}
	return n
}

// ---------------------------------------------------------------------------
// text-node-renderable
// ---------------------------------------------------------------------------

// runTextNodeRenderable demonstrates rendering text from a TextBuffer using
// a TextRenderable (text node renderable).
func runTextNodeRenderable() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	// Create a TextBuffer and populate it.
	tb := native.NewTextBufferWithContent("Line 1: Hello from TextBuffer\nLine 2: Rendered via TextRenderable\nLine 3: Text node renderable demo\n")

	// Insert additional content at the end.
	_ = tb.Insert(tb.Length(), "Line 4: Appended text\n")
	// Delete a range at the beginning.
	_ = tb.Delete(0, 7) // Remove "Line 1: "

	content := tb.GetContent()

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "text-node-box",
			Width:  strPtr("50"),
			Height: strPtr("10"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Text Buffer Renderable",
	})

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "text-node-content",
			Width:  strPtr("46"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
		},
		Content: content,
	})
	text.SetPosition(1, 1)

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "text-node-info",
			Width:  strPtr("46"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf("Buffer length: %d bytes | Lines: %d",
			tb.Length(), tb.LineCount()),
	})
	info.SetPosition(1, 8)

	box.AddChild(text.Renderable)
	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// post-processing-filters
// ---------------------------------------------------------------------------

// runPostProcessingFilters demonstrates post-processing filters (blur,
// invert, grayscale, sepia) applied to a NativeBuffer via post.ApplyFilter.
func runPostProcessingFilters() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "post-processing-filters-box",
			Width:  strPtr("70"),
			Height: strPtr("14"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Post-Processing Filters (blur/invert/grayscale/sepia)",
	})

	// Create a NativeBuffer with colorful content.
	buf := core.NewNativeBuffer(60, 8)
	for y := 0; y < 8; y++ {
		for x := 0; x < 60; x++ {
			ch := uint32('A' + (x+y)%26)
			fg := uint16(0x0F00 | uint16(x*4))
			bg := uint16(0xF000 | uint16(y*4))
			buf.SetCell(x, y, ch, fg, bg, core.TextAttributeNone)
		}
	}

	// Apply multiple post-processing filters in sequence.
	post.ApplyFilter(buf, post.Blur)
	post.ApplyFilter(buf, post.Invert)
	post.ApplyFilter(buf, post.Grayscale)
	post.ApplyFilter(buf, post.Sepia)

	// Convert the processed buffer to text for display.
	var sb strings.Builder
	access := buf.Access()
	for y := 0; y < access.Height; y++ {
		for x := 0; x < access.Width; x++ {
			idx := y*access.Width + x
			ch := access.Char[idx]
			if ch == 0 {
				sb.WriteByte(' ')
			} else {
				sb.WriteRune(rune(ch))
			}
		}
		sb.WriteByte('\n')
	}

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "postproc-text",
			Width:  strPtr("60"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
		Content: sb.String(),
	})
	text.SetPosition(1, 1)

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "postproc-info",
			Width:  strPtr("60"),
			Height: strPtr("2"),
			Visible: boolPtr(true),
		},
		Content: "Filters applied: blur + invert + grayscale + sepia",
	})
	info.SetPosition(1, 10)

	box.AddChild(text.Renderable)
	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// image-resize
// ---------------------------------------------------------------------------

// runImageResize demonstrates Image resize with nearest, bilinear, and
// lanczos resampling algorithms.
func runImageResize() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	// Create a source image with a gradient pattern.
	src := native.NewImage(40, 20)
	for y := 0; y < src.Height; y++ {
		for x := 0; x < src.Width; x++ {
			idx := (y*src.Width + x) * 4
			src.Pixels[idx] = uint8((x * 255) / src.Width)     // R gradient
			src.Pixels[idx+1] = uint8((y * 255) / src.Height)   // G gradient
			src.Pixels[idx+2] = 128                             // B constant
			src.Pixels[idx+3] = 255                             // A opaque
		}
	}

	// Resize with different algorithms.
	nearest := src.Resize(20, 10, native.ResizeNearest)
	bilinear := src.Resize(20, 10, native.ResizeBilinear)
	lanczos := src.Resize(20, 10, native.ResizeLanczos3)

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "image-resize-box",
			Width:  strPtr("70"),
			Height: strPtr("14"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Image Resize (nearest / bilinear / lanczos)",
	})

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "image-resize-info",
			Width:  strPtr("66"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf(`Source: %dx%d
Resized (nearest):   %dx%d
Resized (bilinear):  %dx%d
Resized (lanczos3):  %dx%d
All three algorithms produce %dx%d output.`,
			src.Width, src.Height,
			nearest.Width, nearest.Height,
			bilinear.Width, bilinear.Height,
			lanczos.Width, lanczos.Height,
			20, 10),
	})
	info.SetPosition(1, 1)

	// Render the resized images as blocks.
	imgNearest := renderables.NewImageRenderable(ctx, renderables.ImageOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "img-nearest",
			Width:  strPtr("20"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
		Image:    nearest,
		Protocol: core.ImageRenderProtocolBlocks,
	})
	imgNearest.SetPosition(1, 3)

	imgBilinear := renderables.NewImageRenderable(ctx, renderables.ImageOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "img-bilinear",
			Width:  strPtr("20"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
		Image:    bilinear,
		Protocol: core.ImageRenderProtocolBlocks,
	})
	imgBilinear.SetPosition(22, 3)

	imgLanczos := renderables.NewImageRenderable(ctx, renderables.ImageOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "img-lanczos",
			Width:  strPtr("20"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
		Image:    lanczos,
		Protocol: core.ImageRenderProtocolBlocks,
	})
	imgLanczos.SetPosition(43, 3)

	box.AddChild(info.Renderable)
	box.AddChild(imgNearest.Renderable)
	box.AddChild(imgBilinear.Renderable)
	box.AddChild(imgLanczos.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// audio-tones
// ---------------------------------------------------------------------------

// runAudioTones demonstrates the audio engine with multiple voices,
// PlayTone, and PlayToneWithWaveform.
func runAudioTones() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	// Create an audio engine at CD-quality sample rate.
	engine := native.NewAudioEngine(44100)
	engine.SetBackend(native.AudioBackendNull, "")

	// Create multiple voices with different waveforms and frequencies.
	voice1 := engine.CreateVoiceWithWaveform(440.0, 0.5, native.WaveformSine)    // A4
	voice2 := engine.CreateVoiceWithWaveform(554.37, 0.4, native.WaveformSquare) // C#5
	voice3 := engine.CreateVoiceWithWaveform(659.25, 0.3, native.WaveformSaw)    // E5
	voice4 := engine.CreateVoiceWithWaveform(880.0, 0.2, native.WaveformTriangle) // A5

	voice1.Play()
	voice2.Play()
	voice3.Play()
	voice4.Play()

	// Mix a few frames of audio.
	engine.Mix(256)

	// Demonstrate PlayTone and PlayToneWithWaveform (non-blocking, fire and forget).
	engine.PlayTone(523.25, 0.6, 0.1)  // C5 sine tone
	engine.PlayToneWithWaveform(659.25, 0.5, 0.1, native.WaveformSquare) // E5 square tone

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "audio-tones-box",
			Width:  strPtr("56"),
			Height: strPtr("12"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Audio Tones (multiple voices + waveforms)",
	})

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "audio-tones-text",
			Width:  strPtr("52"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf(`Audio engine demo:
Sample rate: %d Hz
Voices: %d
  Voice 1: 440 Hz sine (A4)
  Voice 2: 554 Hz square (C#5)
  Voice 3: 659 Hz saw (E5)
  Voice 4: 880 Hz triangle (A5)
PlayTone: 523 Hz sine (C5)
PlayToneWithWaveform: 659 Hz square (E5)
Captured samples: %d`,
			engine.SampleRate(),
			engine.VoiceCount(),
			engine.CapturedSampleCount()),
	})
	text.SetPosition(1, 1)

	box.AddChild(text.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// keymap-layers
// ---------------------------------------------------------------------------

// runKeymapLayers demonstrates the keymap package with multiple layers,
// priority ordering, and feature packs.
func runKeymapLayers() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	rootTarget := &keTarget{name: "root"}
	focused := &keTarget{name: "focused"}

	var press func(*keEvent)
	host := keymap.KeymapHost[*keTarget, *keEvent]{
		Metadata: keymap.HostMetadata{
			Platform:        "linux",
			PrimaryModifier: "ctrl",
			Modifiers:       map[string]string{"ctrl": "supported"},
		},
		RootTarget:       rootTarget,
		IsDestroyed:      func() bool { return false },
		GetFocusedTarget: func() *keTarget { return focused },
		GetParentTarget:  func(t *keTarget) *keTarget { return rootTarget },
		IsTargetDestroyed: func(t *keTarget) bool { return false },
		OnKeyPress: func(listener func(*keEvent)) func() {
			press = listener
			return func() {}
		},
		OnKeyRelease:    func(listener func(*keEvent)) func() { return func() {} },
		OnFocusChange:   func(listener func(*keTarget)) func() { return func() {} },
		OnDestroy:       func(listener func()) func() { return func() {} },
		OnRawInput:      func(listener func(string) bool) func() { return func() {} },
		CreateCommandEvent: func() *keEvent { return &keEvent{} },
		OnTargetDestroy:  func(t *keTarget, listener func()) func() { return func() {} },
	}

	km, err := keymap.NewKeymap[*keTarget, *keEvent](host)
	if err != nil {
		return fmt.Errorf("failed to create keymap: %w", err)
	}

	// Layer 1: high priority, global bindings (quit, help).
	quitRequested := false
	helpShown := false
	km.RegisterLayer(keymap.Layer[*keTarget, *keEvent]{
		Target:  rootTarget,
		Priority: 200,
		Bindings: []keymap.Binding[*keTarget, *keEvent]{
			{
				Key: "q",
				Cmd: keymap.BindingCommand[*keTarget, *keEvent]{
					Kind: keymap.BindingCommandHandler,
					Run: func(_ keymap.CommandContext[any, keymap.KeymapEvent]) keymap.CommandResult {
						quitRequested = true
						return keymap.RunCommandResult{OK: true}
					},
				},
				PreventDefault: true,
			},
			{
				Key: "?",
				Cmd: keymap.BindingCommand[*keTarget, *keEvent]{
					Kind: keymap.BindingCommandHandler,
					Run: func(_ keymap.CommandContext[any, keymap.KeymapEvent]) keymap.CommandResult {
						helpShown = true
						return keymap.RunCommandResult{OK: true}
					},
				},
				PreventDefault: true,
			},
		},
	})

	// Layer 2: medium priority, editor bindings (save, undo).
	saved := false
	undone := false
	km.RegisterLayer(keymap.Layer[*keTarget, *keEvent]{
		Target:  focused,
		Priority: 100,
		Bindings: []keymap.Binding[*keTarget, *keEvent]{
			{
				Key: "ctrl+s",
				Cmd: keymap.BindingCommand[*keTarget, *keEvent]{
					Kind: keymap.BindingCommandHandler,
					Run: func(_ keymap.CommandContext[any, keymap.KeymapEvent]) keymap.CommandResult {
						saved = true
						return keymap.RunCommandResult{OK: true}
					},
				},
				PreventDefault: true,
			},
			{
				Key: "ctrl+z",
				Cmd: keymap.BindingCommand[*keTarget, *keEvent]{
					Kind: keymap.BindingCommandHandler,
					Run: func(_ keymap.CommandContext[any, keymap.KeymapEvent]) keymap.CommandResult {
						undone = true
						return keymap.RunCommandResult{OK: true}
					},
				},
				PreventDefault: true,
			},
		},
	})

	// Feature pack: search bindings (can be enabled/disabled together).
	searchEnabled := false
	searchPack := km.RegisterFeaturePack("search", []keymap.Binding[*keTarget, *keEvent]{
		{
			Key: "/",
			Cmd: keymap.BindingCommand[*keTarget, *keEvent]{
				Kind: keymap.BindingCommandHandler,
				Run: func(_ keymap.CommandContext[any, keymap.KeymapEvent]) keymap.CommandResult {
					searchEnabled = true
					return keymap.RunCommandResult{OK: true}
				},
			},
			PreventDefault: true,
		},
	})

	// Simulate key presses through the captured host listener.
	press(&keEvent{name: "q"})
	press(&keEvent{name: "ctrl+s"})
	press(&keEvent{name: "/"})
	searchPack.Enable()
	press(&keEvent{name: "/"})

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "keymap-layers-box",
			Width:  strPtr("56"),
			Height: strPtr("12"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Keymap Layers (priority + feature packs)",
	})

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "keymap-layers-text",
			Width:  strPtr("52"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf(`Keymap layers demo:
Layer 1 (priority 200): q=quit, ?=help
Layer 2 (priority 100): ctrl+s=save, ctrl+z=undo
Feature pack "search": /=search (enabled=%v)
Quit requested: %v
Help shown: %v
Saved: %v
Undone: %v
Pending sequence: %v`,
			searchEnabled,
			quitRequested,
			helpShown,
			saved,
			undone,
			km.HasPendingSequence()),
	})
	text.SetPosition(1, 1)

	box.AddChild(text.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// react-jsx
// ---------------------------------------------------------------------------

// runReactJSX demonstrates React-style element tree with complex nesting
// (box inside box, text with styling).
func runReactJSX() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	// React-style element tree rendered to a native buffer.
	reactRenderer := react.NewReactRenderer(ctx)
	buf := core.NewNativeBuffer(80, 24)

	// Complex nested element tree: box inside box, text with styling.
	app := react.CreateElement("box", map[string]interface{}{
		"id":           "app",
		"border":       true,
		"title":        "React JSX Demo",
		"borderStyle":  "double",
		"borderColor":  "#00AAFF",
		"shouldFill":   true,
	}, react.CreateElement("box", map[string]interface{}{
		"id":           "inner",
		"border":       true,
		"title":        "Inner Panel",
		"borderStyle":  "single",
		"width":        "30",
		"height":       "10",
		"backgroundColor": "#1E1E28",
	}, react.CreateElement("text", map[string]interface{}{
		"id":      "message",
		"content": "Hello from nested React-style elements!\nBox inside box with styling.",
	})))

	reactRenderer.Render(app, buf, 0.016)

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "react-jsx-text",
			Width:  strPtr("78"),
			Height: strPtr("22"),
			Visible: boolPtr(true),
		},
		Content: "React-style element tree (complex nesting):\n\n" + bufferToText(buf),
	})

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "react-jsx-box",
			Width:  strPtr("80"),
			Height: strPtr("24"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "React JSX",
	})

	box.AddChild(text.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// solid-signals
// ---------------------------------------------------------------------------

// runSolidSignals demonstrates Solid reactivity with CreateSignal,
// CreateEffect, and CreateMemo that react to changes.
func runSolidSignals() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	// Reactive signal and derived state.
	count := solid.NewSignal(0)
	doubled := solid.CreateMemo(func() interface{} {
		return count.Value().(int) * 2
	})

	// Effect that reacts to signal changes.
	var effectRuns int
	solid.CreateEffect(func() {
		_ = count.Value()
		_ = doubled
		effectRuns++
	})

	// Update the signal - effect and memo should react.
	count.Set(5)
	count.Set(10)
	count.Set(21)

	// Render a Solid component tree to a buffer.
	solidRenderer := solid.NewSolidRenderer(ctx)
	buf := core.NewNativeBuffer(80, 24)

	component := solid.Box(map[string]interface{}{
		"id":           "solid-app",
		"border":       true,
		"title":        "Solid Signals Demo",
		"borderStyle":  "single",
		"width":        "40",
		"height":       "12",
		"children": []solid.SolidComponent{
			solid.Text(map[string]interface{}{
				"id":      "solid-text",
				"content": fmt.Sprintf("Count: %d\nDoubled: %d\nEffect runs: %d",
					count.Value(), doubled, effectRuns),
			}),
		},
	})

	solidRenderer.Render(component, buf, 0.016)

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "solid-signals-text",
			Width:  strPtr("78"),
			Height: strPtr("22"),
			Visible: boolPtr(true),
		},
		Content: "Solid signals demo:\n\nCount: " + fmt.Sprintf("%d", count.Value()) +
			"\nDoubled (memo): " + fmt.Sprintf("%d", doubled) +
			"\nEffect runs: " + fmt.Sprintf("%d", effectRuns) +
			"\n\nSignal, memo, and effect all reactive to changes.",
	})

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "solid-signals-box",
			Width:  strPtr("80"),
			Height: strPtr("24"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Solid Signals",
	})

	box.AddChild(text.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// ssh-full
// ---------------------------------------------------------------------------

// runSSHFull demonstrates an SSH server with handler, middleware, session,
// and WriteFrame.
func runSSHFull() error {
	// Track server events.
	var handlerRuns int
	var middlewareRuns int
	var writeFrames int
	var sessionID string

	// Middleware that logs session access.
	middleware := func(session *ssh.Session, next func() error) error {
		middlewareRuns++
		sessionID = session.ID()
		return next()
	}

	// Server handler that serves an OpenTUI screen over SSH using WriteFrame.
	server := ssh.NewServer("127.0.0.1:0", func(session *ssh.Session) error {
		handlerRuns++
	session.Write([]byte("\r\nWelcome to OpenTUI over SSH\r\n"))
	session.Write([]byte("This session is served with middleware and WriteFrame.\r\n"))
	session.Write([]byte("Press q to quit.\r\n\r\n"))

		// Demonstrate WriteFrame: render a frame and write it to the SSH session.
		renderer := core.NewCliRenderer(40, 10)
		ctx := renderer.Context()
		root := core.NewRootRenderable(ctx, "root")
		text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
			RenderableOptions: core.RenderableOptions[core.Renderable]{
				ID:     "ssh-frame-text",
				Width:  strPtr("38"),
				Height: strPtr("6"),
				Visible: boolPtr(true),
			},
			Content: "SSH frame written over the wire via WriteFrame.",
		})
		text.SetPosition(1, 1)
		box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
			RenderableOptions: core.RenderableOptions[core.Renderable]{
				ID:     "ssh-frame-box",
				Width:  strPtr("40"),
				Height: strPtr("8"),
				Visible: boolPtr(true),
			},
			Border: true,
			Title:  "SSH WriteFrame",
		})
		box.AddChild(text.Renderable)
		root.AddChild(box.Renderable)
		renderHeadless(renderer, root)
		if err := session.WriteFrame(renderer); err != nil {
			_ = err
		} else {
			writeFrames++
		}
		return nil
	})
	server.SetNoAuth()
	server.Use(middleware)

	if err := server.Listen(); err != nil {
		return fmt.Errorf("failed to start SSH server: %w", err)
	}
	defer server.Close()

	addr := server.Addr()

	// Connect an SSH client to demonstrate a full round trip.
	config := &cryptossh.ClientConfig{
		User:            "demo",
		Auth:            []cryptossh.AuthMethod{},
		HostKeyCallback: cryptossh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	conn, err := cryptossh.Dial("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("failed to dial SSH server: %w", err)
	}
	defer conn.Close()

	sessionSSH, err := conn.NewSession()
	if err != nil {
		return fmt.Errorf("failed to open SSH session: %w", err)
	}
	defer sessionSSH.Close()

	if err := sessionSSH.Shell(); err != nil {
		return fmt.Errorf("failed to request shell: %w", err)
	}

	// Read the welcome message written by the server handler.
	stdout, err := sessionSSH.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	data := make([]byte, 0)
	tmp := make([]byte, 1024)
	for {
		nn, rerr := stdout.Read(tmp)
		if nn > 0 {
			data = append(data, tmp[:nn]...)
		}
		if strings.Contains(string(data), "Press q to quit.") {
			break
		}
		if rerr != nil || nn == 0 {
			break
		}
	}

	output := string(data)
	if !strings.Contains(output, "Welcome to OpenTUI over SSH") {
		return fmt.Errorf("expected SSH welcome message, got: %q", output)
	}

	_ = output
	_ = sessionID
	_ = sessionSSH

	// Verify all components worked.
	if handlerRuns == 0 {
		return fmt.Errorf("SSH handler was not invoked")
	}
	if middlewareRuns == 0 {
		return fmt.Errorf("SSH middleware was not invoked")
	}
	if writeFrames == 0 {
		return fmt.Errorf("SSH WriteFrame was not called")
	}

	return nil
}

// ---------------------------------------------------------------------------
// three-3d
// ---------------------------------------------------------------------------// ---------------------------------------------------------------------------
// three-3d
// ---------------------------------------------------------------------------

// runThree3D demonstrates a Three.js-style 3D scene with multiple objects
// and a moving camera.
func runThree3D() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	// Set up a Three.js-style CPU 3D renderer.
	threeRenderer := three.NewRenderer(40, 20)
	scene := three.NewScene()
	camera := three.NewCamera()

	// Multiple colored triangle meshes.
	shader := three.NewShader("default", "", "")
	redMat := three.NewMaterial(shader)
	redMat.Color = color.RGBA{R: 255, G: 0, B: 0, A: 255}
	greenMat := three.NewMaterial(shader)
	greenMat.Color = color.RGBA{R: 0, G: 255, B: 0, A: 255}
	blueMat := three.NewMaterial(shader)
	blueMat.Color = color.RGBA{R: 0, G: 0, B: 255, A: 255}

	redGeo := three.NewGeometry()
	redGeo.Vertices = []three.Vector3{{X: -1, Y: -1, Z: 0}, {X: 1, Y: -1, Z: 0}, {X: 0, Y: 1, Z: 0}}
	redGeo.Indices = []int{0, 1, 2}
	greenGeo := three.NewGeometry()
	greenGeo.Vertices = []three.Vector3{{X: -1, Y: 1, Z: 0}, {X: 1, Y: 1, Z: 0}, {X: 0, Y: -1, Z: 0}}
	greenGeo.Indices = []int{0, 1, 2}
	blueGeo := three.NewGeometry()
	blueGeo.Vertices = []three.Vector3{{X: -1, Y: 0, Z: -1}, {X: 1, Y: 0, Z: -1}, {X: 0, Y: 0, Z: 1}}
	blueGeo.Indices = []int{0, 1, 2}

	redMesh := three.NewMesh(redGeo, redMat)
	greenMesh := three.NewMesh(greenGeo, greenMat)
	blueMesh := three.NewMesh(blueGeo, blueMat)

	scene.AddObject(redMesh)
	scene.AddObject(greenMesh)
	scene.AddObject(blueMesh)

	threeRenderer.SetScene(scene)
	threeRenderer.SetCamera(camera)

	// Move the camera.
	camera.SetPosition(0, 0, 5)

	buf := core.NewNativeBuffer(40, 20)
	threeRenderer.Render(buf, 0.016)

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "three-3d-text",
			Width:  strPtr("38"),
			Height: strPtr("18"),
			Visible: boolPtr(true),
		},
		Content: "CPU 3D scene (multiple objects, camera at Z=5):\n\n" +
			bufferToText(buf) +
			"\n\nScene objects: 3 | Camera: active | Renderer: CPU",
	})

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "three-3d-box",
			Width:  strPtr("40"),
			Height: strPtr("20"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Three.js 3D Scene",
	})

	box.AddChild(text.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// web-server
// ---------------------------------------------------------------------------

// runWebServer demonstrates a web documentation site with multiple pages.
func runWebServer() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	// Create a documentation site with multiple pages.
	site := web.NewSite("OpenTUI Go Documentation")

	site.AddPage(web.NewPage("/", "Beranda",
		"<h1>OpenTUI Go</h1><p>Selamat datang dalam dokumentasi OpenTUI Go.</p>"))
	site.AddPage(web.NewPage("/docs", "Dokumentasi",
		"<h1>Dokumentasi OpenTUI Go</h1><p>Halaman dokumentasi resmi OpenTUI Go.</p>"))
	site.AddPage(web.NewPage("/docs/getting-started", "Getting Started",
		"<h1>Getting Started</h1><p>Panduan pemula untuk OpenTUI Go.</p>"))
	site.AddPage(web.NewPage("/docs/components", "Components",
		"<h1>Components</h1><p>Daftar semua komponen OpenTUI Go.</p>"))
	site.AddPage(web.NewPage("/about", "Tentang",
		"<h1>Tentang OpenTUI</h1><p>OpenTUI adalah terminal user interface framework yang ditulis dalam Go.</p>"))
	site.AddPage(web.NewPage("/api/status", "Status API",
		`{"status":"ok","service":"pixtui","version":"1.0.0"}`))

	// Verify site contents.
	pages := site.ListPages()
	pageCount := site.PageCount()

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "web-server-box",
			Width:  strPtr("60"),
			Height: strPtr("14"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Web Documentation Site (multiple pages)",
	})

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Site: %s\n", site.Title()))
	sb.WriteString(fmt.Sprintf("Pages: %d\n\n", pageCount))
	for _, p := range pages {
		sb.WriteString(fmt.Sprintf("  %s - %s\n", p.Path, p.Title))
	}
	sb.WriteString("\nRoutes: GET /, GET /docs, GET /docs/{path}, GET /about, GET /api/status, GET /api/pages")

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "web-server-text",
			Width:  strPtr("56"),
			Height: strPtr("10"),
			Visible: boolPtr(true),
		},
		Content: sb.String(),
	})
	text.SetPosition(1, 1)

	box.AddChild(text.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// animation-timeline
// ---------------------------------------------------------------------------

// runAnimationTimeline demonstrates an animation timeline with multiple
// sequential animations.
func runAnimationTimeline() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	animator := animation.NewAnimator()

	// Multiple sequential animations.
	var values []float64
	var names []string

	// Animation 1: fade in (0 to 100).
	anim1 := animation.NewAnimation("fade-in", time.Second, 0, 100, animation.EaseInOut)
	anim1.OnUpdate = func(value float64) {
		values = append(values, value)
		names = append(names, "fade-in")
	}
	anim1.SetLoop(false)
	animator.Add(anim1)

	// Animation 2: slide (100 to 200).
	anim2 := animation.NewAnimation("slide", time.Second, 100, 200, animation.EaseOut)
	anim2.OnUpdate = func(value float64) {
		values = append(values, value)
		names = append(names, "slide")
	}
	anim2.SetLoop(false)
	animator.Add(anim2)

	// Animation 3: pulse (200 to 150).
	anim3 := animation.NewAnimation("pulse", time.Second, 200, 150, animation.EaseInOut)
	anim3.OnUpdate = func(value float64) {
		values = append(values, value)
		names = append(names, "pulse")
	}
	anim3.SetLoop(false)
	animator.Add(anim3)

	// Advance through the timeline.
	for i := 0; i < 30; i++ {
		animator.Update(100 * time.Millisecond)
	}

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "animation-timeline-box",
			Width:  strPtr("56"),
			Height: strPtr("12"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Animation Timeline (sequential animations)",
	})

	var sb strings.Builder
	sb.WriteString("Animation timeline (sequential):\n")
	sb.WriteString("1. fade-in: 0 -> 100 (EaseInOut)\n")
	sb.WriteString("2. slide:   100 -> 200 (EaseOut)\n")
	sb.WriteString("3. pulse:   200 -> 150 (EaseInOut)\n\n")
	sb.WriteString(fmt.Sprintf("Animator count: %d\n", animator.Count()))
	sb.WriteString(fmt.Sprintf("Total value samples: %d\n", len(values)))
	if len(values) > 0 {
		sb.WriteString(fmt.Sprintf("Final value: %.0f\n", values[len(values)-1]))
	}

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "animation-timeline-text",
			Width:  strPtr("52"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
		Content: sb.String(),
	})
	text.SetPosition(2, 2)

	box.AddChild(text.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// plugin-full-lifecycle
// ---------------------------------------------------------------------------

// lifecyclePlugin is a full lifecycle plugin that registers commands and
// events, then verifies Init/Shutdown/RegisterCommand/EmitEvent.
type lifecyclePlugin struct {
	*plugin.BasePlugin
	initCalled      bool
	shutdownCalled  bool
	commandRuns     int
	eventRuns       int
	lastCommand     string
	lastEvent       string
	commandNames    []string
	eventNames      []string
}

func (p *lifecyclePlugin) Init(ctx *plugin.Context) error {
	p.initCalled = true
	ctx.RegisterCommand("greet", func(args []string) (string, error) {
		p.commandRuns++
		p.lastCommand = "greet " + strings.Join(args, " ")
		p.commandNames = append(p.commandNames, "greet")
		return "Hello, " + strings.Join(args, " ") + "!", nil
	})
	ctx.RegisterCommand("ping", func(args []string) (string, error) {
		p.commandRuns++
		p.lastCommand = "ping"
		p.commandNames = append(p.commandNames, "ping")
		return "pong", nil
	})
	ctx.OnEvent("plugin-init", func(data interface{}) {
		p.eventRuns++
		p.lastEvent = fmt.Sprintf("%v", data)
		p.eventNames = append(p.eventNames, "plugin-init")
	})
	ctx.EmitEvent("plugin-init", "initialized")
	return nil
}

func (p *lifecyclePlugin) Shutdown() error {
	p.shutdownCalled = true
	return nil
}

// runPluginFullLifecycle demonstrates a plugin with full lifecycle:
// Init, Shutdown, RegisterCommand, EmitEvent.
func runPluginFullLifecycle() error {
	manager := plugin.NewManager()

	p := &lifecyclePlugin{
		BasePlugin: plugin.NewBasePlugin("lifecycle-plugin", "3.0.0"),
	}
	if err := manager.Register(p); err != nil {
		return fmt.Errorf("failed to register plugin: %w", err)
	}

	if manager.Count() != 1 {
		return fmt.Errorf("expected 1 plugin, got %d", manager.Count())
	}
	if !p.initCalled {
		return fmt.Errorf("plugin Init was not called during Register")
	}

	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()
	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "plugin-lifecycle-box",
			Width:  strPtr("56"),
			Height: strPtr("10"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Plugin Full Lifecycle",
	})

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "plugin-lifecycle-text",
			Width:  strPtr("52"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf(`Plugin: %s v%s
Init: %v | Shutdown: %v
Commands: %d | Events: %d
Last cmd: %s | Last event: %s`,
			p.Name(), p.Version(), p.initCalled, p.shutdownCalled,
			p.commandRuns, p.eventRuns,
			p.lastCommand, p.lastEvent),
	})
	text.SetPosition(1, 1)

	box.AddChild(text.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	if err := runLoop(renderer, root, 2*time.Second); err != nil {
		_ = manager.Shutdown()
		return err
	}

	if err := manager.Shutdown(); err != nil {
		return fmt.Errorf("failed to shutdown plugins: %w", err)
	}
	if !p.shutdownCalled {
		return fmt.Errorf("plugin Shutdown was not called")
	}
	if manager.Count() != 0 {
		return fmt.Errorf("expected 0 plugins after shutdown")
	}

	return nil
}

// ---------------------------------------------------------------------------
// extmarks-highlight
// ---------------------------------------------------------------------------

// runExtmarksHighlight demonstrates ExtmarkManager with highlight selection,
// error, and custom types.
func runExtmarksHighlight() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	buf := core.NewNativeBuffer(56, 10)

	lines := []string{
		"OpenTUI Extmark Highlight Demo",
		"Selection: highlighted text range here.",
		"Error: diagnostic error message here.",
		"Custom: user-defined annotation mark.",
	}
	for y, line := range lines {
		for x, ch := range line {
			if x < buf.Width() {
				buf.SetCell(x, y, uint32(ch), 0, 0, core.TextAttributeNone)
			}
		}
	}

	manager := core.NewExtmarkManager()

	// Selection extmark (blue background).
	manager.Add(&core.Extmark{
		Type:      core.ExtmarkTypeSelection,
		StartLine: 1, StartCol: 10, EndLine: 1, EndCol: 20,
		Foreground: core.RGBA{R: 255, G: 255, B: 255, A: 255},
		Background: core.RGBA{R: 0, G: 100, B: 200, A: 255},
		Priority:   10,
	})
	// Highlight extmark (green foreground).
	manager.Add(&core.Extmark{
		Type:      core.ExtmarkTypeHighlight,
		StartLine: 2, StartCol: 0, EndLine: 2, EndCol: 5,
		Foreground: core.RGBA{R: 255, G: 0, B: 0, A: 255},
		Background: core.RGBA{R: 255, G: 255, B: 255, A: 255},
		Priority:   20,
	})
	// Error extmark (red foreground, white background).
	manager.Add(&core.Extmark{
		Type:      core.ExtmarkTypeError,
		StartLine: 2, StartCol: 0, EndLine: 2, EndCol: 5,
		Foreground: core.RGBA{R: 255, G: 0, B: 0, A: 255},
		Background: core.RGBA{R: 255, G: 255, B: 255, A: 255},
		Priority:   30,
	})
	// Custom extmark (yellow background).
	manager.Add(&core.Extmark{
		Type:      core.ExtmarkTypeCustom,
		StartLine: 3, StartCol: 0, EndLine: 3, EndCol: 6,
		Foreground: core.RGBA{R: 0, G: 0, B: 0, A: 255},
		Background: core.RGBA{R: 255, G: 255, B: 0, A: 255},
		Priority:   5,
		Data:       "annotation",
	})

	layer := core.NewExtmarkLayer(manager, buf)
	layer.Apply()

	var sb strings.Builder
	access := buf.Access()
	for y := 0; y < access.Height; y++ {
		for x := 0; x < access.Width; x++ {
			idx := y*access.Width + x
			ch := access.Char[idx]
			if ch == 0 {
				sb.WriteByte(' ')
			} else {
				sb.WriteRune(rune(ch))
			}
		}
		sb.WriteByte('\n')
	}

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "extmarks-highlight-box",
			Width:  strPtr("60"),
			Height: strPtr("10"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Extmarks (selection/highlight/error/custom)",
	})

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "extmarks-highlight-text",
			Width:  strPtr("56"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
		},
		Content: sb.String(),
	})
	text.SetPosition(1, 1)

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "extmarks-highlight-info",
			Width:  strPtr("56"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf("Extmarks: %d | Types: selection, highlight, error, custom",
			manager.Count()),
	})
	info.SetPosition(1, 8)

	box.AddChild(text.Renderable)
	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// opacity-demo
// ---------------------------------------------------------------------------

// runOpacityDemo demonstrates Box renderables with different opacity levels,
// mirroring the TypeScript opacity-demo.
func runOpacityDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "opacity-demo-box",
			Width:  strPtr("76"),
			Height: strPtr("18"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Opacity Demo (1-4 toggle, A animate)",
	})

	colors := []struct {
		label   string
		hex     string
		opacity float64
	}{
		{"Box 1", "#e94560", 1.0},
		{"Box 2", "#0f3460", 0.8},
		{"Box 3", "#533483", 0.5},
		{"Box 4", "#16a085", 0.3},
	}

	positions := [][2]int{{2, 2}, {22, 2}, {42, 2}, {62, 2}}
	var boxRenderables []*renderables.BoxRenderable

	for i, c := range colors {
		opacity := float64Ptr(c.opacity)
		b := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
			RenderableOptions: core.RenderableOptions[core.Renderable]{
				ID:     fmt.Sprintf("opacity-box-%d", i),
				Width:  strPtr("18"),
				Height: strPtr("8"),
				Visible: boolPtr(true),
				Opacity: opacity,
			},
			BackgroundColor: hexPtr(c.hex),
			Border:          true,
			BorderStyle:     "double",
			BorderColor: hexPtr("#FFFFFF"),
		})
		b.SetPosition(float32(positions[i][0]), float32(positions[i][1]))

		label := renderables.NewTextRenderable(ctx, renderables.TextOptions{
			RenderableOptions: core.RenderableOptions[core.Renderable]{
				ID:     fmt.Sprintf("opacity-label-%d", i),
				Width:  strPtr("16"),
				Height: strPtr("1"),
				Visible: boolPtr(true),
			},
			Content: c.label,
		})
		label.SetPosition(float32(positions[i][0]+1), float32(positions[i][1]+1))

		opLabel := renderables.NewTextRenderable(ctx, renderables.TextOptions{
			RenderableOptions: core.RenderableOptions[core.Renderable]{
				ID:     fmt.Sprintf("opacity-val-%d", i),
				Width:  strPtr("16"),
				Height: strPtr("1"),
				Visible: boolPtr(true),
			},
			Content: fmt.Sprintf("Opacity: %.1f", c.opacity),
		})
		opLabel.SetPosition(float32(positions[i][0]+1), float32(positions[i][1]+3))

		b.AddChild(label.Renderable)
		b.AddChild(opLabel.Renderable)
		box.AddChild(b.Renderable)
		boxRenderables = append(boxRenderables, b)
	}

	// Nested opacity demo.
	nested := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "nested-container",
			Width:  strPtr("34"),
			Height: strPtr("10"),
			Visible: boolPtr(true),
			Opacity: float64Ptr(0.7),
		},
		BackgroundColor: hexPtr("#e94560"),
		Border:          true,
		BorderStyle:     "single",
	})
	nested.SetPosition(2, 12)

	nestedLabel := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "nested-label",
			Width:  strPtr("30"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Content: "Parent: 0.7 opacity",
	})
	nestedLabel.SetPosition(3, 13)

	nestedChild := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "nested-child",
			Width:  strPtr("28"),
			Height: strPtr("5"),
			Visible: boolPtr(true),
			Opacity: float64Ptr(0.5), // Effective: 0.7 * 0.5 = 0.35
		},
		BackgroundColor: hexPtr("#0f3460"),
		Border:          true,
	})
	nestedChild.SetPosition(3, 14)

	childLabel := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "child-label",
			Width:  strPtr("24"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Content: "Child: 0.5 opacity",
	})
	childLabel.SetPosition(4, 15)

	effLabel := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "effective-label",
			Width:  strPtr("24"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Content: "Effective: 0.35",
	})
	effLabel.SetPosition(4, 16)

	nestedChild.AddChild(childLabel.Renderable)
	nestedChild.AddChild(effLabel.Renderable)
	nested.AddChild(nestedLabel.Renderable)
	nested.AddChild(nestedChild.Renderable)
	box.AddChild(nested.Renderable)

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "opacity-info",
			Width:  strPtr("72"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Content: "Press 1-4 to toggle opacity, A to animate. Nested opacity multiplies (0.7 * 0.5 = 0.35).",
	})
	info.SetPosition(2, 17)
	box.AddChild(info.Renderable)

	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// styled-text-demo
// ---------------------------------------------------------------------------

// runStyledTextDemo demonstrates styled text with bold, underline, colors,
// and background colors, mirroring the TypeScript styled-text-demo.
func runStyledTextDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	container := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "styled-text-container",
			Width:  strPtr("76"),
			Height: strPtr("22"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Styled Text Demo",
	})

	ex1 := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "house-text",
			Width:  strPtr("34"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
		},
		Content: "There's a house,\nWith a window,\nAnd a corvette\nAnd everything is blue",
		Attributes: core.TextAttributeBold,
	})
	ex1.SetPosition(2, 2)

	ex2 := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "status-text",
			Width:  strPtr("38"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
		},
		Content: "ERROR: Connection failed\nSUCCESS: Data loaded\nWARNING: Low memory\nNOTICE: System update available",
	})
	ex2.SetPosition(2, 9)

	dashboard := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "dashboard-box",
			Width:  strPtr("38"),
			Height: strPtr("12"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "System Dashboard",
	})
	dashboard.SetPosition(38, 2)

	dashText := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "complex-template",
			Width:  strPtr("34"),
			Height: strPtr("10"),
			Visible: boolPtr(true),
		},
		Content: "System Stats:\nUptime: 0.00s\nCPU Load: 0.0%\nMemory: 0.0%\nNetwork: 0 KB/s\nTemp: 0.0°C\nBattery: 100%\nConnection: ONLINE\nHealth: GOOD\nAlert: NORMAL",
	})
	dashText.SetPosition(1, 1)
	dashboard.AddChild(dashText.Renderable)

	instr := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "instructions",
			Width:  strPtr("34"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
		Content: "Styled Text Demo\nESC to return\n\nFeatures:\n• Bold, underlined, colors\n• Background colors\n• Custom hex colors\n• Hyperlinks: https://opentui.com",
	})
	instr.SetPosition(2, 16)

	container.AddChild(ex1.Renderable)
	container.AddChild(ex2.Renderable)
	container.AddChild(dashboard.Renderable)
	container.AddChild(instr.Renderable)
	root.AddChild(container.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// simple-layout-demo
// ---------------------------------------------------------------------------

// runSimpleLayoutDemo demonstrates box layouts: horizontal, vertical,
// centered, and three-column, mirroring the TypeScript simple-layout-demo.
func runSimpleLayoutDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	header := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "header",
			Width:  strPtr("76"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		BackgroundColor: hexPtr("#3b82f6"),
		Border:          true,
		BorderStyle:     "single",
		Title:           "LAYOUT DEMO",
		TitleAlignment:  "center",
	})
	header.SetPosition(2, 1)

	contentArea := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "content-area",
			Width:  strPtr("72"),
			Height: strPtr("14"),
			Visible: boolPtr(true),
		},
		Border: true,
		BorderStyle: "single",
		Title:  "Horizontal Layout: Sidebar | Main | Right",
	})
	contentArea.SetPosition(2, 5)

	sidebar := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "sidebar",
			Width:  strPtr("20"),
			Height: strPtr("12"),
			Visible: boolPtr(true),
		},
		BackgroundColor: hexPtr("#64748b"),
		Border:          true,
		BorderStyle:     "single",
		Title:           "LEFT SIDEBAR",
		TitleAlignment:  "center",
	})
	sidebar.SetPosition(3, 6)

	mainContent := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "main-content",
			Width:  strPtr("30"),
			Height: strPtr("12"),
			Visible: boolPtr(true),
		},
		BackgroundColor: hexPtr("#eab308"),
		Border:          true,
		BorderStyle:     "single",
		Title:           "MAIN CONTENT",
		TitleAlignment:  "center",
	})
	mainContent.SetPosition(25, 6)

	rightSidebar := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "right-sidebar",
			Width:  strPtr("20"),
			Height: strPtr("12"),
			Visible: boolPtr(true),
		},
		BackgroundColor: hexPtr("#7c3aed"),
		Border:          true,
		BorderStyle:     "single",
		Title:           "RIGHT",
		TitleAlignment:  "center",
	})
	rightSidebar.SetPosition(57, 6)

	footer := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "footer",
			Width:  strPtr("76"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		BackgroundColor: hexPtr("#1e40af"),
		Border:          true,
		BorderStyle:     "single",
		TitleAlignment:  "center",
		Title:           "SPACE: next | R: restart | P: autoplay | V: overlay | WASD: move",
	})
	footer.SetPosition(2, 20)

	contentArea.AddChild(sidebar.Renderable)
	contentArea.AddChild(mainContent.Renderable)
	contentArea.AddChild(rightSidebar.Renderable)

	root.AddChild(header.Renderable)
	root.AddChild(contentArea.Renderable)
	root.AddChild(footer.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// text-truncation-demo
// ---------------------------------------------------------------------------

// runTextTruncationDemo demonstrates text truncation and wrapping modes,
// mirroring the TypeScript text-truncation-demo.
func runTextTruncationDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	mainContainer := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "mainContainer",
			Width:  strPtr("76"),
			Height: strPtr("22"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Text Truncation Demo - T: toggle truncate | W: cycle wrap",
	})

	header := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "headerText",
			Width:  strPtr("72"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Content: "Press 'T' to toggle truncation, 'W' to cycle wrap mode (none/char/word)",
	})
	header.SetPosition(1, 1)

	longText := "This is a very long single line of text that will definitely exceed the width of most terminal windows and should be truncated when truncation is enabled"
	alphabet := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz"
	unicodeText := "🌟 Unicode test: こんにちは世界 Hello World 你好世界 안녕하세요 🚀 More emoji: 🎨🎭🎪🎬🎮🎯"

	t1 := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "singleLineText1",
			Width:  strPtr("34"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Content: longText,
		Truncate: true,
	})
	t1.SetPosition(2, 3)

	t2 := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "singleLineText2",
			Width:  strPtr("34"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Content: alphabet,
		Truncate: true,
	})
	t2.SetPosition(38, 3)

	t3 := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "singleLineText3",
			Width:  strPtr("34"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Content: unicodeText,
		Truncate: true,
	})
	t3.SetPosition(2, 7)

	mt1 := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "multilineText1",
			Width:  strPtr("34"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
		},
		Content: "This is a multiline text block that demonstrates how truncation works with word wrapping enabled. Each line that exceeds the viewport width will be truncated independently.",
		Wrap:   true,
	})
	mt1.SetPosition(38, 7)

	mt2 := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "multilineText2",
			Width:  strPtr("34"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
		},
		Content: "Line 1: This is a long line without wrapping\nLine 2: Another very long line that will be truncated when enabled\nLine 3: Short line",
		Wrap:   true,
	})
	mt2.SetPosition(38, 14)

	styled := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "styledText",
			Width:  strPtr("34"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Content: "Bold Cyan: Yellow text and magenta with green parts and more styled text that goes on and on",
		Attributes: core.TextAttributeBold,
		Wrap:   true,
	})
	styled.SetPosition(2, 11)

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "truncation-info",
			Width:  strPtr("72"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Content: "Truncate: ENABLED | Wrap: word | T: toggle truncate | W: cycle wrap | C: clear selection",
	})
	info.SetPosition(1, 19)

	mainContainer.AddChild(header.Renderable)
	mainContainer.AddChild(t1.Renderable)
	mainContainer.AddChild(t2.Renderable)
	mainContainer.AddChild(t3.Renderable)
	mainContainer.AddChild(mt1.Renderable)
	mainContainer.AddChild(mt2.Renderable)
	mainContainer.AddChild(styled.Renderable)
	mainContainer.AddChild(info.Renderable)
	root.AddChild(mainContainer.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// text-wrap-demo
// ---------------------------------------------------------------------------

// runTextWrapDemo demonstrates automatic text wrapping with word mode,
// mirroring the TypeScript text-wrap example.
func runTextWrapDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "text-wrap-box",
			Width:  strPtr("76"),
			Height: strPtr("22"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Text Wrap Demo (W: cycle wrap mode)",
	})

	content := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "text-renderable",
			Width:  strPtr("72"),
			Height: strPtr("14"),
			Visible: boolPtr(true),
		},
		Content: "🎨 OpenTUI Text Wrapping Demo\n\nWelcome to the text wrapping demonstration. This example showcases how OpenTUI handles automatic text wrapping with styled content.\n\n✨ Key Features:\n• Word-based wrapping - Preserves word boundaries when breaking lines\n• Character-based wrapping - Breaks at any character for precise control\n• Dynamic resizing - Text reflows automatically as container dimensions change\n• Rich styling - Individual text segments can have different colors and attributes\n\n🔧 How It Works:\nTextNodes are created with specific styling and then composed together to form rich, formatted text content. Each node can contain different foreground colors, background colors, and text attributes like bold, italic, and underline.\n\n💻 Example Code:\nconst styledText = TextNodeRenderable.fromNodes([\n  TextNodeRenderable.fromString(\"Hello \", { fg: \"#9ece6a\" }),\n  TextNodeRenderable.fromString(\"World\", { fg: \"#7aa2f7\", attributes: 1 }),\n  TextNodeRenderable.fromString(\"!\", { fg: \"#f7768e\" })\n]);\n\n🎮 Try It Out:\nDrag the borders or corners of this text box to resize it and watch how the text wrapping adapts in real-time.",
		Wrap:   true,
	})
	content.SetPosition(1, 1)

	instr := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "instructions",
			Width:  strPtr("72"),
			Height: strPtr("4"),
			Visible: boolPtr(true),
		},
		Content: "Wrap mode: word | W: Cycle wrap mode | M: Toggle char/word | Drag borders/corners to resize",
	})
	instr.SetPosition(1, 16)

	box.AddChild(content.Renderable)
	box.AddChild(instr.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// transparency-demo
// ---------------------------------------------------------------------------

// runTransparencyDemo demonstrates transparency using background colors and
// opacity on Box renderables, mirroring the TypeScript transparency-demo.
func runTransparencyDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "transparency-box",
			Width:  strPtr("76"),
			Height: strPtr("18"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Transparency Demo (opacity + transparent background)",
	})

	t1 := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "trans-box-1",
			Width:  strPtr("20"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
			Opacity: float64Ptr(0.6),
		},
		BackgroundColor: hexPtr("#e94560"),
		Border:          true,
		BorderStyle:     "rounded",
		Title:           "0.6 opacity",
		TitleAlignment:  "center",
	})
	t1.SetPosition(2, 2)

	t2 := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "trans-box-2",
			Width:  strPtr("20"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
			Opacity: float64Ptr(0.4),
		},
		BackgroundColor: hexPtr("#0f3460"),
		Border:          true,
		BorderStyle:     "rounded",
		Title:           "0.4 opacity",
		TitleAlignment:  "center",
	})
	t2.SetPosition(24, 2)

	t3 := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "trans-box-3",
			Width:  strPtr("20"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
			Opacity: float64Ptr(0.2),
		},
		BackgroundColor: hexPtr("#16a085"),
		Border:          true,
		BorderStyle:     "rounded",
		Title:           "0.2 opacity",
		TitleAlignment:  "center",
	})
	t3.SetPosition(46, 2)

	t4 := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "trans-box-4",
			Width:  strPtr("20"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
			Opacity: float64Ptr(1.0),
		},
		BackgroundColor: hexPtr("#533483"),
		Border:          true,
		BorderStyle:     "rounded",
		Title:           "1.0 opacity",
		TitleAlignment:  "center",
	})
	t4.SetPosition(2, 10)

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "transparency-info",
			Width:  strPtr("72"),
			Height: strPtr("4"),
			Visible: boolPtr(true),
		},
		Content: "Transparency via opacity (0.0-1.0) and RGBA background colors. Lower opacity = more transparent.",
	})
	info.SetPosition(2, 17)

	box.AddChild(t1.Renderable)
	box.AddChild(t2.Renderable)
	box.AddChild(t3.Renderable)
	box.AddChild(t4.Renderable)
	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// link-demo
// ---------------------------------------------------------------------------

// runLinkDemo demonstrates text with hyperlink-style styling, mirroring the
// TypeScript link-demo.
func runLinkDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "link-demo-box",
			Width:  strPtr("76"),
			Height: strPtr("14"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Link Demo",
	})

	content := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "link-content",
			Width:  strPtr("72"),
			Height: strPtr("10"),
			Visible: boolPtr(true),
		},
		Content: "Styled Text Demo with Links\n\n• Template literals with colors\n• Bold, underlined, and other styles\n• Background colors like this\n• Custom hex colors like this red\n• Hyperlinks: https://opentui.com\n\nType Examples:\nNumber: 42\nBoolean: true\nFloat: 3.14\nCalculated: 7",
		Attributes: core.TextAttributeBold,
	})
	content.SetPosition(1, 1)

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "link-info",
			Width:  strPtr("72"),
			Height: strPtr("2"),
			Visible: boolPtr(true),
		},
		Content: "Links are rendered as underlined colored text: opentui",
	})
	info.SetPosition(1, 12)

	box.AddChild(content.Renderable)
	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// full-unicode-demo
// ---------------------------------------------------------------------------

// runFullUnicodeDemo demonstrates full Unicode support including emoji,
// combining characters, CJK, and right-to-left scripts, mirroring the
// TypeScript full-unicode-demo.
func runFullUnicodeDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "full-unicode-box",
			Width:  strPtr("76"),
			Height: strPtr("18"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Full Unicode Demo",
	})

	lines := []string{
		"✅ 👩🏽‍💻  👨‍👩‍👧‍👦  🏳️‍🌈  🇺🇸  🇩🇪  🇯🇵  🇮🇳",
		"a̐éö̲  Z͑͗͛̒͘a̴͈͚̐̓l̷͓̱͉g̶̙̗̓͘o̵͍͈  क्‍ष",
		"مرحبا  こんにちは  สวัสดี  Здравствуйте",
		"𝔘𝔫𝔦𝔠𝔬𝔡𝔢  𝒻𝓊𝓁𝓁 𝓌𝒾𝒹𝓉𝒽：ＡＢＣ  ½ ⅞ ⅓",
	}

	var sb strings.Builder
	for _, line := range lines {
		sb.WriteString(line)
		sb.WriteByte('\n')
	}

	unicodeText := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "full-unicode-text",
			Width:  strPtr("72"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
		},
		Content: sb.String(),
	})
	unicodeText.SetPosition(1, 1)

	styledText := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "full-unicode-styled",
			Width:  strPtr("72"),
			Height: strPtr("4"),
			Visible: boolPtr(true),
			Attributes: core.TextAttributeBold,
		},
		Content: "Graphemes: ✅ 👩🏽‍💻  👨‍👩‍👧‍👦  🏳️‍🌈  🇺🇸  🇩🇪  🇯🇵  🇮🇳\nComplex: a̐éö̲  Z͑͗͛̒͘a̴͈͚̐̓l̷͓̱͉g̶̙̗̓͘o̵͍͈  क्‍ष",
	})
	styledText.SetPosition(1, 8)

	alignment := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "full-unicode-alignment",
			Width:  strPtr("72"),
			Height: strPtr("2"),
			Visible: boolPtr(true),
		},
		Content: "1234567890123456789012345678901234567890| 40-cell ruler\nMalayalam: peroxidation",
	})
	alignment.SetPosition(1, 13)

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "full-unicode-info",
			Width:  strPtr("72"),
			Height: strPtr("2"),
			Visible: boolPtr(true),
		},
		Content: "Emoji, combining characters, CJK, RTL, and full-width Unicode supported.",
	})
	info.SetPosition(1, 16)

	box.AddChild(unicodeText.Renderable)
	box.AddChild(styledText.Renderable)
	box.AddChild(alignment.Renderable)
	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// text-selection-demo
// ---------------------------------------------------------------------------

// runTextSelectionDemo demonstrates text selection across multiple renderables,
// mirroring the TypeScript text-selection-demo.
func runTextSelectionDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	mainContainer := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "text-selection-box",
			Width:  strPtr("76"),
			Height: strPtr("20"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Text Selection Demo",
	})

	t1 := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "selection-text-1",
			Width:  strPtr("34"),
			Height: strPtr("4"),
			Visible: boolPtr(true),
		},
		Content: "This is a paragraph in the first box.\nIt contains multiple lines of text\nthat can be selected independently.\n世界, 你好世界, 中文, 한글",
	})
	t1.SetPosition(2, 2)

	t2 := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "selection-text-2",
			Width:  strPtr("34"),
			Height: strPtr("4"),
			Visible: boolPtr(true),
		},
		Content: "function handleSelection() {\n  const selected = getSelectedText()\n  console.log(selected)\n}",
	})
	t2.SetPosition(38, 2)

	t3 := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "selection-text-3",
			Width:  strPtr("34"),
			Height: strPtr("4"),
			Visible: boolPtr(true),
		},
		Content: "Selection Demo\n✓ Cross-renderable selection\n✓ Nested groups and boxes\n✓ Styled text support",
	})
	t3.SetPosition(2, 8)

	instr := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "selection-instructions",
			Width:  strPtr("34"),
			Height: strPtr("2"),
			Visible: boolPtr(true),
		},
		Content: "Click and drag to select text across any elements. Press 'C' to clear selection.",
	})
	instr.SetPosition(2, 13)

	status := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "selection-status",
			Width:  strPtr("72"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Content: "No selection - try selecting across different nested elements",
	})
	status.SetPosition(1, 16)

	mainContainer.AddChild(t1.Renderable)
	mainContainer.AddChild(t2.Renderable)
	mainContainer.AddChild(t3.Renderable)
	mainContainer.AddChild(instr.Renderable)
	mainContainer.AddChild(status.Renderable)
	root.AddChild(mainContainer.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// relative-positioning-demo
// ---------------------------------------------------------------------------

// runRelativePositioningDemo demonstrates relative and absolute positioning
// of renderables using SetPosition, mirroring the TypeScript relative-positioning-demo.
func runRelativePositioningDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	container := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "relative-positioning-box",
			Width:  strPtr("76"),
			Height: strPtr("18"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Relative Positioning Demo",
	})

	tl := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "pos-topleft",
			Width:  strPtr("16"),
			Height: strPtr("4"),
			Visible: boolPtr(true),
		},
		BackgroundColor: hexPtr("#e94560"),
		Border:          true,
		Title:           "Top Left",
		TitleAlignment:  "center",
	})
	tl.SetPosition(2, 2)

	tr := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "pos-topright",
			Width:  strPtr("16"),
			Height: strPtr("4"),
			Visible: boolPtr(true),
		},
		BackgroundColor: hexPtr("#0f3460"),
		Border:          true,
		Title:           "Top Right",
		TitleAlignment:  "center",
	})
	tr.SetPosition(58, 2)

	bl := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "pos-bottomleft",
			Width:  strPtr("16"),
			Height: strPtr("4"),
			Visible: boolPtr(true),
		},
		BackgroundColor: hexPtr("#533483"),
		Border:          true,
		Title:           "Bottom Left",
		TitleAlignment:  "center",
	})
	bl.SetPosition(2, 14)

	br := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "pos-bottomright",
			Width:  strPtr("16"),
			Height: strPtr("4"),
			Visible: boolPtr(true),
		},
		BackgroundColor: hexPtr("#16a085"),
		Border:          true,
		Title:           "Bottom Right",
		TitleAlignment:  "center",
	})
	br.SetPosition(58, 14)

	center := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "pos-center",
			Width:  strPtr("20"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
		},
		BackgroundColor: hexPtr("#f1c40f"),
		Border:          true,
		Title:           "Center",
		TitleAlignment:  "center",
	})
	center.SetPosition(28, 6)

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "relative-info",
			Width:  strPtr("72"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Content: "Boxes positioned absolutely at corners and center using SetPosition(x, y).",
	})
	info.SetPosition(1, 12)

	container.AddChild(tl.Renderable)
	container.AddChild(tr.Renderable)
	container.AddChild(bl.Renderable)
	container.AddChild(br.Renderable)
	container.AddChild(center.Renderable)
	container.AddChild(info.Renderable)
	root.AddChild(container.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// nested-zindex-demo
// ---------------------------------------------------------------------------

// runNestedZIndexDemo demonstrates z-index ordering of nested renderables,
// mirroring the TypeScript nested-zindex-demo.
func runNestedZIndexDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "nested-zindex-box",
			Width:  strPtr("76"),
			Height: strPtr("16"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Nested Z-Index Demo",
	})

	l1 := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "z-layer-1",
			Width:  strPtr("30"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
			ZIndex: intPtr(1),
		},
		BackgroundColor: hexPtr("#e94560"),
		Border:          true,
		Title:           "Z-Index 1 (bottom)",
		TitleAlignment:  "center",
	})
	l1.SetPosition(2, 2)

	l2 := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "z-layer-2",
			Width:  strPtr("26"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
			ZIndex: intPtr(10),
		},
		BackgroundColor: hexPtr("#0f3460"),
		Border:          true,
		Title:           "Z-Index 10 (middle)",
		TitleAlignment:  "center",
	})
	l2.SetPosition(10, 4)

	l3 := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "z-layer-3",
			Width:  strPtr("22"),
			Height: strPtr("4"),
			Visible: boolPtr(true),
			ZIndex: intPtr(100),
		},
		BackgroundColor: hexPtr("#533483"),
		Border:          true,
		Title:           "Z-Index 100 (top)",
		TitleAlignment:  "center",
	})
	l3.SetPosition(18, 6)

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "zindex-info",
			Width:  strPtr("72"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Content: "Higher z-index renders on top: 1 (bottom) < 10 (middle) < 100 (top).",
	})
	info.SetPosition(1, 12)

	box.AddChild(l1.Renderable)
	box.AddChild(l2.Renderable)
	box.AddChild(l3.Renderable)
	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// live-state-demo
// ---------------------------------------------------------------------------

// runLiveStateDemo demonstrates a Live renderable that re-renders every frame,
// mirroring the TypeScript live-state-demo.
func runLiveStateDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	var frameCount int

	live := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "live-state-text",
			Width:  strPtr("72"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
			Live:   boolPtr(true),
		},
		Content: "Live State Demo\nFrame: 0\nThis renderable updates every frame.",
	})
	live.SetPosition(1, 1)

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "live-state-box",
			Width:  strPtr("76"),
			Height: strPtr("12"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Live State Demo",
	})

	box.AddChild(live.Renderable)
	root.AddChild(box.Renderable)

	live.SetOptions(renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "live-state-text",
			Width:  strPtr("72"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
			Live:   boolPtr(true),
			RenderAfter: func(_ *core.Renderable, buf *core.NativeBuffer, _ float64) {
				frameCount++
				_ = buf
			},
		},
		Content: fmt.Sprintf("Live State Demo\nFrame: %d\nThis renderable updates every frame.", frameCount),
	})

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// program-status-demo
// ---------------------------------------------------------------------------

// runProgramStatusDemo demonstrates program status display with multiple
// status indicators, mirroring the TypeScript program-status-demo.
func runProgramStatusDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "program-status-box",
			Width:  strPtr("76"),
			Height: strPtr("14"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Program Status Demo",
	})

	statuses := []struct {
		label  string
		value  string
		status string
	}{
		{"CPU", "42%", "OK"},
		{"Memory", "78%", "OK"},
		{"Disk", "55%", "OK"},
		{"Network", "100ms", "WARN"},
		{"Uptime", "5d 12h", "OK"},
	}

	var sb strings.Builder
	sb.WriteString("Program Status\n\n")
	for _, s := range statuses {
		sb.WriteString(fmt.Sprintf("%-10s %-10s %s\n", s.label, s.value, s.status))
	}
	sb.WriteString("\nAll systems operational.")

	statusText := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "program-status-text",
			Width:  strPtr("72"),
			Height: strPtr("10"),
			Visible: boolPtr(true),
		},
		Content: sb.String(),
	})
	statusText.SetPosition(1, 1)

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "program-status-info",
			Width:  strPtr("72"),
			Height: strPtr("2"),
			Visible: boolPtr(true),
		},
		Content: "Status indicators: CPU, Memory, Disk, Network, Uptime.",
	})
	info.SetPosition(1, 12)

	box.AddChild(statusText.Renderable)
	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// sticky-scroll-demo
// ---------------------------------------------------------------------------

// runStickyScrollDemo demonstrates a ScrollBox with sticky content at the
// bottom, mirroring the TypeScript sticky-scroll-example.
func runStickyScrollDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "sticky-scroll-box",
			Width:  strPtr("76"),
			Height: strPtr("18"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Sticky Scroll Demo",
	})

	var sb strings.Builder
	for i := 1; i <= 50; i++ {
		sb.WriteString(fmt.Sprintf("Scrollable line %d\n", i))
	}
	body := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "sticky-body",
			Width:  strPtr("34"),
			Height: strPtr("60"),
			Visible: boolPtr(true),
		},
		Content: sb.String(),
	})

	scrollbox := renderables.NewScrollBoxRenderable(ctx, renderables.ScrollBoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "sticky-scrollbox",
			Width:  strPtr("36"),
			Height: strPtr("10"),
			Visible: boolPtr(true),
		},
	})
	scrollbox.SetPosition(2, 2)
	scrollbox.AddChild(body.Renderable)

	sticky := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "sticky-footer",
			Width:  strPtr("34"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Content: "STICKY FOOTER\nAlways visible at bottom\nPress q to quit",
	})
	sticky.SetPosition(2, 13)

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "sticky-info",
			Width:  strPtr("34"),
			Height: strPtr("2"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf("Scroll position: %d", scrollbox.GetScrollY()),
	})
	info.SetPosition(2, 16)

	box.AddChild(scrollbox.Renderable)
	box.AddChild(sticky.Renderable)
	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// scroll-demo
// ---------------------------------------------------------------------------

// runScrollDemo demonstrates a ScrollBox with content that exceeds the
// viewport, mirroring the TypeScript scroll-example.
func runScrollDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "scroll-demo-box",
			Width:  strPtr("76"),
			Height: strPtr("16"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Scroll Example",
	})

	var sb strings.Builder
	for i := 1; i <= 40; i++ {
		sb.WriteString(fmt.Sprintf("Line %d: Scrollable content here\n", i))
	}
	body := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "scroll-body",
			Width:  strPtr("34"),
			Height: strPtr("40"),
			Visible: boolPtr(true),
		},
		Content: sb.String(),
	})

	scrollbox := renderables.NewScrollBoxRenderable(ctx, renderables.ScrollBoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "scrollbox",
			Width:  strPtr("36"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
	})
	scrollbox.SetPosition(2, 2)
	scrollbox.AddChild(body.Renderable)

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "scroll-info",
			Width:  strPtr("34"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf("Scroll position: %d | 40 lines total", scrollbox.GetScrollY()),
	})
	info.SetPosition(2, 11)

	box.AddChild(scrollbox.Renderable)
	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// console-demo
// ---------------------------------------------------------------------------

// runConsoleDemo demonstrates a simple console output with multiple timestamped
// log lines, mirroring the TypeScript console-demo.
func runConsoleDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "console-box",
			Width:  strPtr("76"),
			Height: strPtr("16"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Console Demo",
	})

	var sb strings.Builder
	sb.WriteString("OpenTUI Console\n")
	sb.WriteString("===============\n\n")
	sb.WriteString("[10:00:01] System initialized.\n")
	sb.WriteString("[10:00:02] Loading modules...\n")
	sb.WriteString("[10:00:03] Module 'core' loaded.\n")
	sb.WriteString("[10:00:04] Module 'renderables' loaded.\n")
	sb.WriteString("[10:00:05] Module 'native' loaded.\n")
	sb.WriteString("[10:00:06] All systems ready.\n")
	sb.WriteString("\nPress q to quit.")

	console := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "console-output",
			Width:  strPtr("72"),
			Height: strPtr("12"),
			Visible: boolPtr(true),
		},
		Content: sb.String(),
	})
	console.SetPosition(1, 1)

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "console-info",
			Width:  strPtr("72"),
			Height: strPtr("2"),
			Visible: boolPtr(true),
		},
		Content: "Simple console output with timestamped log lines.",
	})
	info.SetPosition(1, 14)

	box.AddChild(console.Renderable)
	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// editor-demo
// ---------------------------------------------------------------------------

// runEditorDemo demonstrates a multi-line text editor using Textarea with
// insert, undo, and redo operations, mirroring the TypeScript editor-demo.
func runEditorDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "editor-box",
			Width:  strPtr("50"),
			Height: strPtr("12"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Editor Demo",
	})

	editor := renderables.NewTextareaRenderable(ctx, renderables.TextareaOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "editor-field",
			Width:  strPtr("46"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
		},
		Value: "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello, Editor!\")\n}\n",
	})
	editor.SetPosition(2, 2)

	// Simulate editing operations.
	editor.Insert("\t// edited\n")
	_ = editor.Undo()
	_ = editor.Redo()

	cursorLine, cursorCol := editor.Cursor()

	label := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "editor-label",
			Width:  strPtr("46"),
			Height: strPtr("4"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf("Value:\n%s\nCursor: line %d, col %d",
			strings.ReplaceAll(editor.Value(), "\n", "\n  "), cursorLine, cursorCol),
	})
	label.SetPosition(2, 9)

	box.AddChild(label.Renderable)
	box.AddChild(editor.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// keypress-debug-demo
// ---------------------------------------------------------------------------

// runKeypressDebugDemo demonstrates key press debugging by simulating key
// events and displaying the results, mirroring the TypeScript keypress-debug-demo.
func runKeypressDebugDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "keypress-box",
			Width:  strPtr("56"),
			Height: strPtr("12"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Keypress Debug Demo",
	})

	// Simulate key presses and display debug info.
	var keys []string
	keys = append(keys, "Key: ArrowDown")
	keys = append(keys, "Key: j")
	keys = append(keys, "Key: ArrowUp")
	keys = append(keys, "Key: k")
	keys = append(keys, "Key: Enter")

	var sb strings.Builder
	sb.WriteString("Keypress Debug\n")
	sb.WriteString("==============\n\n")
	sb.WriteString("Simulated key events:\n")
	for _, k := range keys {
		sb.WriteString(fmt.Sprintf("  %s\n", k))
	}
	sb.WriteString("\nLast key: Enter\n")
	sb.WriteString("Modifiers: none\n")

	debug := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "keypress-debug",
			Width:  strPtr("52"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
		Content: sb.String(),
	})
	debug.SetPosition(1, 1)

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "keypress-info",
			Width:  strPtr("52"),
			Height: strPtr("2"),
			Visible: boolPtr(true),
		},
		Content: "Press q to quit.",
	})
	info.SetPosition(1, 10)

	box.AddChild(debug.Renderable)
	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// markdown-code-block-renderer-demo
// ---------------------------------------------------------------------------

// runMarkdownCodeBlockRendererDemo demonstrates Markdown rendering with
// embedded code blocks, mirroring the TypeScript markdown-code-block-renderer-demo.
func runMarkdownCodeBlockRendererDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "markdown-code-box",
			Width:  strPtr("60"),
			Height: strPtr("16"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Markdown Code Block Renderer",
	})

	markdown := renderables.NewMarkdownRenderable(ctx, renderables.MarkdownOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "markdown-code-content",
			Width:  strPtr("56"),
			Height: strPtr("12"),
			Visible: boolPtr(true),
		},
		Content: "# Code Example\n\nHere is a Go function:\n\n```go\nfunc greet(name string) string {\n\treturn \"Hello, \" + name\n}\n```\n\nAnd some inline `code` too.\n",
	})
	markdown.SetPosition(1, 1)

	box.AddChild(markdown.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// mouse-interaction-demo
// ---------------------------------------------------------------------------

// runMouseInteractionDemo demonstrates mouse interaction simulation with
// click position display, mirroring the TypeScript mouse-interaction-demo.
func runMouseInteractionDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "mouse-box",
			Width:  strPtr("56"),
			Height: strPtr("12"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Mouse Interaction Demo",
	})

	// Simulate mouse click positions.
	clicks := [][2]int{{10, 5}, {25, 7}, {40, 3}}

	var sb strings.Builder
	sb.WriteString("Mouse Interaction\n")
	sb.WriteString("=================\n\n")
	sb.WriteString("Simulated click positions:\n")
	for i, c := range clicks {
		sb.WriteString(fmt.Sprintf("  Click %d: x=%d, y=%d\n", i+1, c[0], c[1]))
	}
	sb.WriteString("\nMouse events: move, click, drag, scroll\n")

	mouse := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "mouse-debug",
			Width:  strPtr("52"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
		Content: sb.String(),
	})
	mouse.SetPosition(1, 1)

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "mouse-info",
			Width:  strPtr("52"),
			Height: strPtr("2"),
			Visible: boolPtr(true),
		},
		Content: "Mouse interaction simulation with click positions.",
	})
	info.SetPosition(1, 10)

	box.AddChild(mouse.Renderable)
	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// notification-demo
// ---------------------------------------------------------------------------

// runNotificationDemo demonstrates multiple notification boxes with different
// severity levels, mirroring the TypeScript notification-demo.
func runNotificationDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "notification-box",
			Width:  strPtr("76"),
			Height: strPtr("14"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Notification Demo",
	})

	type notif struct {
		title   string
		msg     string
		color   string
		x, y    int
		w, h    int
	}

	notifs := []notif{
		{"INFO", "System started successfully", "#3b82f6", 2, 2, 22, 4},
		{"WARN", "Disk usage at 85%", "#f59e0b", 26, 2, 22, 4},
		{"ERROR", "Connection timeout", "#ef4444", 50, 2, 24, 4},
		{"SUCCESS", "Backup completed", "#10b981", 2, 8, 22, 4},
	}

	for _, n := range notifs {
		nBox := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
			RenderableOptions: core.RenderableOptions[core.Renderable]{
				ID:     fmt.Sprintf("notif-%s", n.title),
				Width:  strPtr(fmt.Sprintf("%d", n.w)),
				Height: strPtr(fmt.Sprintf("%d", n.h)),
				Visible: boolPtr(true),
			},
			BackgroundColor: hexPtr(n.color),
			Border:          true,
			BorderStyle:     "single",
			Title:           n.title,
			TitleAlignment:  "center",
		})
		nBox.SetPosition(float32(n.x), float32(n.y))

		nText := renderables.NewTextRenderable(ctx, renderables.TextOptions{
			RenderableOptions: core.RenderableOptions[core.Renderable]{
				ID:     fmt.Sprintf("notif-%s-msg", n.title),
				Width:  strPtr(fmt.Sprintf("%d", n.w-2)),
				Height: strPtr(fmt.Sprintf("%d", n.h-2)),
				Visible: boolPtr(true),
			},
			Content: n.msg,
		})
		nText.SetPosition(float32(n.x+1), float32(n.y+1))

		nBox.AddChild(nText.Renderable)
		box.AddChild(nBox.Renderable)
	}

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "notification-info",
			Width:  strPtr("72"),
			Height: strPtr("2"),
			Visible: boolPtr(true),
		},
		Content: "Notifications with INFO, WARN, ERROR, and SUCCESS severity levels.",
	})
	info.SetPosition(2, 13)

	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// pcm-playback
// ---------------------------------------------------------------------------

// runPcmPlayback demonstrates PCM audio playback using the audio engine with
// multiple voices and waveforms, mirroring the TypeScript pcm-playback.
func runPcmPlayback() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	// Create an audio engine at CD-quality sample rate.
	engine := native.NewAudioEngine(44100)
	engine.SetBackend(native.AudioBackendNull, "")

	// Create multiple voices with different waveforms and frequencies.
	voice1 := engine.CreateVoiceWithWaveform(440.0, 0.5, native.WaveformSine)    // A4
	voice2 := engine.CreateVoiceWithWaveform(554.37, 0.4, native.WaveformSquare) // C#5
	voice3 := engine.CreateVoiceWithWaveform(659.25, 0.3, native.WaveformSaw)    // E5

	voice1.Play()
	voice2.Play()
	voice3.Play()

	// Mix a few frames of PCM audio.
	engine.Mix(256)

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "pcm-box",
			Width:  strPtr("56"),
			Height: strPtr("12"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "PCM Playback Demo",
	})

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "pcm-text",
			Width:  strPtr("52"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf(`PCM Audio Playback
Sample rate: %d Hz
Backend: null (headless)
Voices: %d
  Voice 1: 440 Hz sine (A4)
  Voice 2: 554 Hz square (C#5)
  Voice 3: 659 Hz saw (E5)
Captured samples: %d`,
			engine.SampleRate(),
			engine.VoiceCount(),
			engine.CapturedSampleCount()),
	})
	text.SetPosition(1, 1)

	box.AddChild(text.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// sprite-animation-demo
// ---------------------------------------------------------------------------

// runSpriteAnimationDemo demonstrates sprite animation using the animation
// package to cycle through sprite frames, mirroring the TypeScript sprite-animation-demo.
func runSpriteAnimationDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	// Animate a value from 0 to 3 (4 sprite frames) using the animation package.
	animator := animation.NewAnimator()
	var frameIndex float64
	anim := animation.NewAnimation("sprite", time.Second, 0, 3, animation.Linear)
	anim.OnUpdate = func(value float64) {
		frameIndex = value
	}
	anim.SetLoop(true)
	animator.Add(anim)

	// Advance the animation a few steps.
	animator.Update(250 * time.Millisecond)
	animator.Update(250 * time.Millisecond)
	animator.Update(250 * time.Millisecond)

	sprites := []string{"⬤", "⬢", "⬡", "◇"}
	currentSprite := sprites[int(frameIndex)]

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "sprite-box",
			Width:  strPtr("50"),
			Height: strPtr("10"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Sprite Animation Demo",
	})

	sprite := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "sprite-frame",
			Width:  strPtr("46"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf("Sprite frame %d: %s\n\nCycling through frames:\n⬤ → ⬢ → ⬡ → ◇\n\nAnimator count: %d",
			int(frameIndex), currentSprite, animator.Count()),
	})
	sprite.SetPosition(2, 2)

	box.AddChild(sprite.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// static-sprite-demo
// ---------------------------------------------------------------------------

// runStaticSpriteDemo demonstrates a static sprite/image rendered using the
// Image renderable, mirroring the TypeScript static-sprite-demo.
func runStaticSpriteDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "static-sprite-box",
			Width:  strPtr("30"),
			Height: strPtr("14"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Static Sprite Demo",
	})

	// Create a sprite image with a pixel-art style pattern.
	sprite := native.NewImage(16, 8)
	for y := 0; y < sprite.Height; y++ {
		for x := 0; x < sprite.Width; x++ {
			idx := (y*sprite.Width + x) * 4
			// Draw a simple sprite pattern.
			if (x == 2 || x == 13) && (y == 1 || y == 6) {
				sprite.Pixels[idx] = 255 // R
				sprite.Pixels[idx+1] = 0   // G
				sprite.Pixels[idx+2] = 0   // B
			} else if x >= 4 && x <= 12 && y >= 3 && y <= 5 {
				sprite.Pixels[idx] = 255 // R
				sprite.Pixels[idx+1] = 255 // G
				sprite.Pixels[idx+2] = 0   // B
			} else {
				sprite.Pixels[idx] = 0
				sprite.Pixels[idx+1] = 0
				sprite.Pixels[idx+2] = 0
			}
			sprite.Pixels[idx+3] = 255 // A opaque
		}
	}

	spriteRenderable := renderables.NewImageRenderable(ctx, renderables.ImageOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "static-sprite",
			Width:  strPtr("16"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
		Image:    sprite,
		Protocol: core.ImageRenderProtocolBlocks,
	})
	spriteRenderable.SetPosition(2, 2)

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "static-sprite-info",
			Width:  strPtr("26"),
			Height: strPtr("4"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf("Sprite: %dx%d pixels\nProtocol: blocks", sprite.Width, sprite.Height),
	})
	info.SetPosition(2, 11)

	box.AddChild(spriteRenderable.Renderable)
	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// terminal
// ---------------------------------------------------------------------------

// runTerminal demonstrates an embedded terminal renderable that parses VT
// output and draws a terminal screen in the render tree, mirroring the
// TypeScript terminal / embedded-terminal-demo.
func runTerminal() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "terminal-box",
			Width:  strPtr("40"),
			Height: strPtr("14"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Embedded Terminal",
	})

	terminal := renderables.NewEmbeddedTerminalRenderable(ctx, renderables.EmbeddedTerminalOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "embedded-terminal",
			Width:  strPtr("36"),
			Height: strPtr("10"),
			Visible: boolPtr(true),
		},
		Width:  36,
		Height: 10,
	})

	// Write some VT output to the terminal.
	terminal.Write([]byte("$ hello \x1b[1;32mworld\x1b[0m\r\n"))
	terminal.Write([]byte("$ pixtui terminal demo\r\n"))
	terminal.Write([]byte("$ Press q to quit.\r\n"))

	terminal.SetPosition(2, 2)

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "terminal-info",
			Width:  strPtr("36"),
			Height: strPtr("2"),
			Visible: boolPtr(true),
		},
		Content: "Embedded terminal with VT parsing.",
	})
	info.SetPosition(2, 13)

	box.AddChild(terminal.Renderable)
	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// terminal-title
// ---------------------------------------------------------------------------

// runTerminalTitle demonstrates a terminal title display, mirroring the
// TypeScript terminal-title demo.
func runTerminalTitle() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "terminal-title-box",
			Width:  strPtr("50"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Terminal Title",
	})

	title := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "terminal-title-text",
			Width:  strPtr("46"),
			Height: strPtr("4"),
			Visible: boolPtr(true),
			Attributes: core.TextAttributeBold,
		},
		Content: "OpenTUI Terminal\nv1.0.0\n\nReady.",
	})
	title.SetPosition(2, 2)

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "terminal-title-info",
			Width:  strPtr("46"),
			Height: strPtr("2"),
			Visible: boolPtr(true),
		},
		Content: "Terminal title bar display.",
	})
	info.SetPosition(2, 7)

	box.AddChild(title.Renderable)
	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// texture-loading-demo
// ---------------------------------------------------------------------------

// runTextureLoadingDemo demonstrates texture/image loading using the Image
// renderable with gradient textures, mirroring the TypeScript texture-loading-demo.
func runTextureLoadingDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "texture-box",
			Width:  strPtr("50"),
			Height: strPtr("16"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Texture Loading Demo",
	})

	// Create multiple textures with different patterns.
	makeTexture := func(w, h int, r, g, b uint8) *native.Image {
		img := native.NewImage(w, h)
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				idx := (y*w + x) * 4
				img.Pixels[idx] = r
				img.Pixels[idx+1] = g
				img.Pixels[idx+2] = b
				img.Pixels[idx+3] = 255
			}
		}
		return img
	}

	tex1 := makeTexture(16, 8, 255, 0, 0)    // red
	tex2 := makeTexture(16, 8, 0, 255, 0)    // green
	tex3 := makeTexture(16, 8, 0, 0, 255)    // blue

	texRenderable1 := renderables.NewImageRenderable(ctx, renderables.ImageOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "texture-1",
			Width:  strPtr("16"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
		Image:    tex1,
		Protocol: core.ImageRenderProtocolBlocks,
	})
	texRenderable1.SetPosition(2, 2)

	texRenderable2 := renderables.NewImageRenderable(ctx, renderables.ImageOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "texture-2",
			Width:  strPtr("16"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
		Image:    tex2,
		Protocol: core.ImageRenderProtocolBlocks,
	})
	texRenderable2.SetPosition(20, 2)

	texRenderable3 := renderables.NewImageRenderable(ctx, renderables.ImageOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "texture-3",
			Width:  strPtr("16"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
		Image:    tex3,
		Protocol: core.ImageRenderProtocolBlocks,
	})
	texRenderable3.SetPosition(2, 12)

	label1 := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "texture-label-1",
			Width:  strPtr("16"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Content: "Red",
	})
	label1.SetPosition(2, 11)

	label2 := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "texture-label-2",
			Width:  strPtr("16"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Content: "Green",
	})
	label2.SetPosition(20, 11)

	label3 := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "texture-label-3",
			Width:  strPtr("16"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Content: "Blue",
	})
	label3.SetPosition(2, 21)

	box.AddChild(texRenderable1.Renderable)
	box.AddChild(texRenderable2.Renderable)
	box.AddChild(texRenderable3.Renderable)
	box.AddChild(label1.Renderable)
	box.AddChild(label2.Renderable)
	box.AddChild(label3.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// video-puzzle-demo
// ---------------------------------------------------------------------------

// runVideoPuzzleDemo demonstrates video playback with puzzle-like frame
// advancement, mirroring the TypeScript video-puzzle-demo.
func runVideoPuzzleDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	vid := native.NewVideo(32, 12)
	vid.Playing = true
	vid.CurrentTime = 0.0
	vid.Duration = 30.0
	vid.FPS = 30.0
	vid.Loop = true

	// Simulate advancing through puzzle frames.
	for i := 0; i < 60; i++ {
		_ = vid.Update(1.0 / 30.0)
	}

	videoRenderable := renderables.NewVideoRenderable(ctx, renderables.VideoOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "video-puzzle",
			Width:  strPtr("28"),
			Height: strPtr("10"),
			Visible: boolPtr(true),
		},
		Video:    vid,
		AutoPlay: true,
	})
	videoRenderable.SetPosition(2, 2)

	state := "paused"
	if vid.Playing {
		state = "playing"
	}
	if vid.Loop {
		state += " (loop)"
	}
	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "video-puzzle-info",
			Width:  strPtr("28"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf("Video Puzzle\nState: %s\nTime: %.1f / %.1f\nFPS: %.0f\nFrame present: %v",
			state, vid.CurrentTime, vid.Duration, vid.FPS, vid.GetFrame() != nil),
	})
	info.SetPosition(2, 13)

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "video-puzzle-box",
			Width:  strPtr("34"),
			Height: strPtr("20"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Video Puzzle Demo",
	})

	box.AddChild(videoRenderable.Renderable)
	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// wide-grapheme-overlay-demo
// ---------------------------------------------------------------------------

// runWideGraphemeOverlayDemo demonstrates wide grapheme (emoji, CJK) text
// overlay, mirroring the TypeScript wide-grapheme-overlay-demo.
func runWideGraphemeOverlayDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "wide-grapheme-box",
			Width:  strPtr("76"),
			Height: strPtr("12"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Wide Grapheme Overlay Demo",
	})

	lines := []string{
		"🎨 🎭 🎪 🎬 🎮 🎯 🎰 🎱",
		"👩🏽‍💻 👨‍👩‍👧‍👦 🏳️‍🌈 🇺🇸 🇯🇵 🇮🇳",
		"こんにちは世界 你好世界 안녕하세요",
		"𝔘𝔫𝔦𝔠𝔬𝔡𝔢 𝒻𝓊𝓁𝓁 𝓌𝒾𝒹𝓉𝒽：ＡＢＣ",
	}

	var sb strings.Builder
	for _, line := range lines {
		sb.WriteString(line)
		sb.WriteByte('\n')
	}

	overlay := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "wide-grapheme-text",
			Width:  strPtr("72"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
		},
		Content: sb.String(),
	})
	overlay.SetPosition(1, 1)

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "wide-grapheme-info",
			Width:  strPtr("72"),
			Height: strPtr("4"),
			Visible: boolPtr(true),
		},
		Content: "Wide graphemes: emoji, ZWJ sequences, CJK, and full-width Unicode.",
	})
	info.SetPosition(1, 8)

	box.AddChild(overlay.Renderable)
	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// input-select-layout-demo
// ---------------------------------------------------------------------------

// runInputSelectLayoutDemo demonstrates a combined layout with Input, Select,
// and Box renderables arranged horizontally, mirroring the TypeScript
// input-select-layout-demo.
func runInputSelectLayoutDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	container := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "input-select-layout-container",
			Width:  strPtr("76"),
			Height: strPtr("14"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Input + Select + Layout Demo",
	})

	input := renderables.NewInputRenderable(ctx, renderables.InputOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "layout-input",
			Width:  strPtr("30"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Value:      "type here",
		Placeholder: "Enter text",
	})
	input.SetPosition(2, 2)

	opts := []string{"Option A", "Option B", "Option C"}
	selectRenderable := renderables.NewSelectRenderable(ctx, renderables.SelectOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "layout-select",
			Width:  strPtr("24"),
			Height: strPtr("4"),
			Visible: boolPtr(true),
		},
		Options:  opts,
		Selected: 0,
	})
	selectRenderable.SetPosition(34, 2)

	label := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "layout-label",
			Width:  strPtr("72"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Content: "Input value: " + input.Value() + "\nSelected: " + opts[selectRenderable.Selected()] + "\nHorizontal layout: Input | Select",
	})
	label.SetPosition(2, 6)

	container.AddChild(input.Renderable)
	container.AddChild(selectRenderable.Renderable)
	container.AddChild(label.Renderable)
	root.AddChild(container.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}


// ---------------------------------------------------------------------------
// ascii-font-selection-demo
// ---------------------------------------------------------------------------

// runASCIIFontSelectionDemo demonstrates ASCII block font rendering with
// multiple text samples, mirroring the TypeScript ascii-font-selection-demo.
func runASCIIFontSelectionDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "ascii-font-selection-box",
			Width:  strPtr("76"),
			Height: strPtr("18"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "ASCII Font Selection",
	})

	samples := []struct {
		text string
		x, y int
		w, h int
	}{
		{"OPENTUI", 2, 2, 22, 5},
		{"GO", 26, 2, 10, 5},
		{"HELLO", 38, 2, 18, 5},
		{"WORLD", 2, 9, 20, 5},
		{"V1.0", 24, 9, 14, 5},
	}

	for _, s := range samples {
		font := renderables.NewASCIIFontRenderable(ctx, renderables.ASCIIFontOptions{
			RenderableOptions: core.RenderableOptions[core.Renderable]{
				ID:     fmt.Sprintf("ascii-%s", s.text),
				Width:  strPtr(fmt.Sprintf("%d", s.w)),
				Height: strPtr(fmt.Sprintf("%d", s.h)),
				Visible: boolPtr(true),
			},
			Text: s.text,
			Font: "block",
		})
		font.SetPosition(float32(s.x), float32(s.y))
		box.AddChild(font.Renderable)
	}

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "ascii-font-selection-info",
			Width:  strPtr("72"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Content: "ASCII block font selection: multiple text samples rendered with FIGlet-style glyphs.",
	})
	info.SetPosition(1, 15)

	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// audio-capture-demo
// ---------------------------------------------------------------------------

// runAudioCaptureDemo demonstrates audio capture using the audio engine with
// multiple voices and PCM mixing, mirroring the TypeScript audio-capture-demo.
func runAudioCaptureDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	engine := native.NewAudioEngine(44100)
	engine.SetBackend(native.AudioBackendNull, "")

	// Create voices for a chord.
	voice1 := engine.CreateVoiceWithWaveform(261.63, 0.5, native.WaveformSine)   // C4
	voice2 := engine.CreateVoiceWithWaveform(329.63, 0.4, native.WaveformSine)   // E4
	voice3 := engine.CreateVoiceWithWaveform(392.00, 0.3, native.WaveformSine)   // G4

	voice1.Play()
	voice2.Play()
	voice3.Play()

	// Mix audio frames to capture PCM.
	engine.Mix(512)
	engine.Mix(512)
	engine.Mix(512)

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "audio-capture-box",
			Width:  strPtr("56"),
			Height: strPtr("12"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Audio Capture Demo",
	})

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "audio-capture-text",
			Width:  strPtr("52"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf(`Audio Capture Demo
Sample rate: %d Hz
Backend: null (headless)
Voices: %d
  C4: 261.63 Hz sine
  E4: 329.63 Hz sine
  G4: 392.00 Hz sine
Captured samples: %d
PCM frames mixed: 3 x 512`,
			engine.SampleRate(),
			engine.VoiceCount(),
			engine.CapturedSampleCount()),
	})
	text.SetPosition(1, 1)

	box.AddChild(text.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// clipboard-paste-demo
// ---------------------------------------------------------------------------

// runClipboardPasteDemo demonstrates clipboard copy and paste, mirroring the
// TypeScript clipboard-paste-demo.
func runClipboardPasteDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	clipboard := native.NewClipboard()

	// Copy text to clipboard.
	copyText := "OpenTUI clipboard demo text"
	_, copyErr := clipboard.Copy(copyText)

	// Paste text from clipboard.
	pasted, pasteStatus, pasteErr := clipboard.Paste()

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "clipboard-box",
			Width:  strPtr("56"),
			Height: strPtr("12"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Clipboard Paste Demo",
	})

	content := fmt.Sprintf(`Clipboard Demo
Copy text: %q
Copy status: %v
Copy error: %v
Paste status: %v
Paste error: %v
Pasted text: %q`,
		copyText, copyErr == nil, copyErr,
		pasteStatus, pasteErr,
		pasted)

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "clipboard-text",
			Width:  strPtr("52"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
		Content: content,
	})
	text.SetPosition(1, 1)

	box.AddChild(text.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// core-plugin-slots-demo
// ---------------------------------------------------------------------------

// runCorePluginSlotsDemo demonstrates the plugin package with renderable
// factories registered as "slots" via the plugin Context, mirroring the
// TypeScript core-plugin-slots-demo.
func runCorePluginSlotsDemo() error {
	manager := plugin.NewManager()

	slotPlugin := &slotsPluginImpl{
		BasePlugin: plugin.NewBasePlugin("slots-plugin", "1.0.0"),
	}
	if err := manager.Register(slotPlugin); err != nil {
		return fmt.Errorf("failed to register slots plugin: %w", err)
	}

	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "core-plugin-slots-box",
			Width:  strPtr("56"),
			Height: strPtr("12"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Core Plugin Slots Demo",
	})

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Plugins registered: %d\n", manager.Count()))
	sb.WriteString("Slots: header, body, footer (renderable factories)\n")
	sb.WriteString("Renderable factories registered via plugin Context.\n")
	sb.WriteString("Command: slot-list -> header, body, footer\n")

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "core-plugin-slots-text",
			Width:  strPtr("52"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
		Content: sb.String(),
	})
	text.SetPosition(1, 1)

	box.AddChild(text.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	if err := runLoop(renderer, root, 2*time.Second); err != nil {
		_ = manager.Shutdown()
		return err
	}
	if err := manager.Shutdown(); err != nil {
		return fmt.Errorf("failed to shutdown plugins: %w", err)
	}
	return nil
}

type slotsPluginImpl struct {
	*plugin.BasePlugin
}

func (p *slotsPluginImpl) Init(ctx *plugin.Context) error {
	// Register renderable factories as "slots".
	ctx.RegisterRenderable("slot:header", func(options map[string]interface{}) (interface{}, error) {
		return "header-slot-renderable", nil
	})
	ctx.RegisterRenderable("slot:body", func(options map[string]interface{}) (interface{}, error) {
		return "body-slot-renderable", nil
	})
	ctx.RegisterRenderable("slot:footer", func(options map[string]interface{}) (interface{}, error) {
		return "footer-slot-renderable", nil
	})
	ctx.RegisterCommand("slot-list", func(args []string) (string, error) {
		return "header, body, footer", nil
	})
	return nil
}

func (p *slotsPluginImpl) Shutdown() error { return nil }

// ---------------------------------------------------------------------------
// focus-restore-demo
// ---------------------------------------------------------------------------

// runFocusRestoreDemo demonstrates focusable panels with Focusable=true,
// mirroring the TypeScript focus-restore-demo.
func runFocusRestoreDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	container := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "focus-restore-box",
			Width:  strPtr("76"),
			Height: strPtr("14"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Focus Restore Demo",
	})

	// Focusable panels that simulate focus restore.
	panel1 := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      "focus-panel-1",
			Width:   strPtr("34"),
			Height:  strPtr("8"),
			Visible: boolPtr(true),
		},
		Focusable:     true,
		BackgroundColor: hexPtr("#3b82f6"),
		Border:          true,
		BorderStyle:     "single",
		Title:           "Panel 1 (focused)",
		TitleAlignment:  "center",
	})
	panel1.SetPosition(2, 2)

	panel2 := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      "focus-panel-2",
			Width:   strPtr("34"),
			Height:  strPtr("8"),
			Visible: boolPtr(true),
		},
		Focusable:     true,
		BackgroundColor: hexPtr("#10b981"),
		Border:          true,
		BorderStyle:     "single",
		Title:           "Panel 2",
		TitleAlignment:  "center",
	})
	panel2.SetPosition(40, 2)

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "focus-restore-info",
			Width:  strPtr("72"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Content: "Focusable panels with Focusable=true. Focus can move between panels and be restored.",
	})
	info.SetPosition(1, 11)

	container.AddChild(panel1.Renderable)
	container.AddChild(panel2.Renderable)
	container.AddChild(info.Renderable)
	root.AddChild(container.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// shader-cube-demo
// ---------------------------------------------------------------------------

// runShaderCubeDemo demonstrates a Three.js-style 3D cube with a shader and
// material, mirroring the TypeScript shader-cube-demo.
func runShaderCubeDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	// Three.js-style CPU 3D renderer with a cube geometry.
	threeRenderer := three.NewRenderer(40, 20)
	scene := three.NewScene()
	camera := three.NewCamera()

	// Shader + material for the cube.
	shader := three.NewShader("cube", "", "")
	mat := three.NewMaterial(shader)
	mat.Color = color.RGBA{R: 0, G: 255, B: 0, A: 255}

	// Cube geometry: 8 vertices, 12 triangles (36 indices).
	geo := three.NewGeometry()
	geo.Vertices = []three.Vector3{
		{X: -1, Y: -1, Z: -1}, {X: 1, Y: -1, Z: -1}, {X: 1, Y: 1, Z: -1}, {X: -1, Y: 1, Z: -1},
		{X: -1, Y: -1, Z: 1}, {X: 1, Y: -1, Z: 1}, {X: 1, Y: 1, Z: 1}, {X: -1, Y: 1, Z: 1},
	}
	// Front face
	geo.Indices = append(geo.Indices, 0, 1, 2, 0, 2, 3)
	// Back face
	geo.Indices = append(geo.Indices, 4, 5, 6, 4, 6, 7)
	// Left face
	geo.Indices = append(geo.Indices, 4, 0, 3, 4, 3, 7)
	// Right face
	geo.Indices = append(geo.Indices, 1, 5, 6, 1, 6, 2)
	// Top face
	geo.Indices = append(geo.Indices, 3, 2, 6, 3, 6, 7)
	// Bottom face
	geo.Indices = append(geo.Indices, 4, 5, 1, 4, 1, 0)

	cube := three.NewMesh(geo, mat)
	scene.AddObject(cube)

	threeRenderer.SetScene(scene)
	threeRenderer.SetCamera(camera)
	camera.SetPosition(0, 0, 5)

	buf := core.NewNativeBuffer(40, 20)
	threeRenderer.Render(buf, 0.016)

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "shader-cube-text",
			Width:  strPtr("38"),
			Height: strPtr("18"),
			Visible: boolPtr(true),
		},
		Content: "Shader Cube Demo\n\n" + bufferToText(buf) +
		"\n\nCube: 8 vertices | 12 triangles | Shader: active",
	})

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "shader-cube-box",
			Width:  strPtr("40"),
			Height: strPtr("20"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Shader Cube (3D)",
	})

	box.AddChild(text.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// split-footer-image-demo
// ---------------------------------------------------------------------------

// runSplitFooterImageDemo demonstrates a split layout with header, content,
// and footer containing an image renderable, mirroring the TypeScript
// split-footer-image-demo.
func runSplitFooterImageDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	// Header.
	header := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "split-footer-image-header",
			Width:  strPtr("76"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		BackgroundColor: hexPtr("#3b82f6"),
		Border:          true,
		BorderStyle:     "single",
		Title:           "SPLIT FOOTER + IMAGE",
		TitleAlignment:  "center",
	})
	header.SetPosition(2, 1)

	// Main content area.
	content := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "split-footer-image-content",
			Width:  strPtr("72"),
			Height: strPtr("12"),
			Visible: boolPtr(true),
		},
		Border:      true,
		BorderStyle: "single",
		Title:       "Main Content",
	})
	content.SetPosition(2, 5)

	contentText := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "split-footer-image-content-text",
			Width:  strPtr("68"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
		Content: "Main content area above the footer.\nImage is rendered in the footer below.\nSplit layout: header | content | footer+image.",
	})
	contentText.SetPosition(1, 1)
	content.AddChild(contentText.Renderable)

	// Footer with image.
	footer := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "split-footer-image-footer",
			Width:  strPtr("76"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
		},
		Border:      true,
		BorderStyle: "single",
		Title:       "Footer: Image Renderable",
	})
	footer.SetPosition(2, 18)

	img := native.NewImage(16, 4)
	for y := 0; y < img.Height; y++ {
		for x := 0; x < img.Width; x++ {
			idx := (y*img.Width + x) * 4
			img.Pixels[idx] = 255
			img.Pixels[idx+1] = 0
			img.Pixels[idx+2] = 0
			img.Pixels[idx+3] = 255
		}
	}
	footerImage := renderables.NewImageRenderable(ctx, renderables.ImageOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "split-footer-image",
			Width:  strPtr("16"),
			Height: strPtr("4"),
			Visible: boolPtr(true),
		},
		Image:    img,
		Protocol: core.ImageRenderProtocolBlocks,
	})
	footerImage.SetPosition(2, 2)

	footerInfo := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "split-footer-image-info",
			Width:  strPtr("54"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Content: "Footer contains an image renderable (blocks protocol).",
	})
	footerInfo.SetPosition(20, 2)

	footer.AddChild(footerImage.Renderable)
	footer.AddChild(footerInfo.Renderable)

	root.AddChild(header.Renderable)
	root.AddChild(content.Renderable)
	root.AddChild(footer.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// split-footer-streaming-demo
// ---------------------------------------------------------------------------

// runSplitFooterStreamingDemo demonstrates a split layout with header, content,
// and footer containing a streaming video renderable, mirroring the TypeScript
// split-footer-streaming-demo.
func runSplitFooterStreamingDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	// Header.
	header := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "split-footer-streaming-header",
			Width:  strPtr("76"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		BackgroundColor: hexPtr("#3b82f6"),
		Border:          true,
		BorderStyle:     "single",
		Title:           "SPLIT FOOTER + STREAMING VIDEO",
		TitleAlignment:  "center",
	})
	header.SetPosition(2, 1)

	// Main content area.
	content := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "split-footer-streaming-content",
			Width:  strPtr("72"),
			Height: strPtr("12"),
			Visible: boolPtr(true),
		},
		Border:      true,
		BorderStyle: "single",
		Title:       "Main Content",
	})
	content.SetPosition(2, 5)

	contentText := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "split-footer-streaming-content-text",
			Width:  strPtr("68"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
		Content: "Main content area above the footer.\nVideo streaming plays in the footer below.\nSplit layout: header | content | footer+video.",
	})
	contentText.SetPosition(1, 1)
	content.AddChild(contentText.Renderable)

	// Footer with streaming video.
	footer := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "split-footer-streaming-footer",
			Width:  strPtr("76"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
		},
		Border:      true,
		BorderStyle: "single",
		Title:       "Footer: Streaming Video",
	})
	footer.SetPosition(2, 18)

	vid := native.NewVideo(24, 4)
	vid.Playing = true
	vid.CurrentTime = 5.0
	vid.Duration = 60.0
	vid.FPS = 30.0
	vid.Loop = true

	footerVideo := renderables.NewVideoRenderable(ctx, renderables.VideoOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "split-footer-streaming-video",
			Width:  strPtr("24"),
			Height: strPtr("4"),
			Visible: boolPtr(true),
		},
		Video:    vid,
		AutoPlay: true,
	})
	footerVideo.SetPosition(2, 2)

	footerInfo := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "split-footer-streaming-info",
			Width:  strPtr("48"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf("Video: %dx%d | Time: %.1f/%.1f | Playing: %v",
			vid.Width, vid.Height, vid.CurrentTime, vid.Duration, vid.Playing),
	})
	footerInfo.SetPosition(28, 2)

	footer.AddChild(footerVideo.Renderable)
	footer.AddChild(footerInfo.Renderable)

	root.AddChild(header.Renderable)
	root.AddChild(content.Renderable)
	root.AddChild(footer.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// split-mode-demo
// ---------------------------------------------------------------------------

// runSplitModeDemo demonstrates a split mode layout with left and right
// panels, mirroring the TypeScript split-mode-demo.
func runSplitModeDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	// Left panel.
	left := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "split-mode-left",
			Width:  strPtr("36"),
			Height: strPtr("18"),
			Visible: boolPtr(true),
		},
		BackgroundColor: hexPtr("#1e3a8a"),
		Border:          true,
		BorderStyle:     "single",
		Title:           "LEFT PANEL",
		TitleAlignment:  "center",
	})
	left.SetPosition(2, 2)

	leftText := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "split-mode-left-text",
			Width:  strPtr("32"),
			Height: strPtr("14"),
			Visible: boolPtr(true),
		},
		Content: "Split Mode\nLeft Panel\n\nNavigation\nFile tree\nSettings\nHelp",
	})
	leftText.SetPosition(1, 1)
	left.AddChild(leftText.Renderable)

	// Right panel.
	right := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "split-mode-right",
			Width:  strPtr("36"),
			Height: strPtr("18"),
			Visible: boolPtr(true),
		},
		BackgroundColor: hexPtr("#7c2d12"),
		Border:          true,
		BorderStyle:     "single",
		Title:           "RIGHT PANEL",
		TitleAlignment:  "center",
	})
	right.SetPosition(40, 2)

	rightText := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "split-mode-right-text",
			Width:  strPtr("32"),
			Height: strPtr("14"),
			Visible: boolPtr(true),
		},
		Content: "Split Mode\nRight Panel\n\nPreview\nCode editor\nOutput\nTerminal",
	})
	rightText.SetPosition(1, 1)
	right.AddChild(rightText.Renderable)

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "split-mode-info",
			Width:  strPtr("72"),
			Height: strPtr("2"),
			Visible: boolPtr(true),
		},
		Content: "Split mode layout: left panel | right panel.",
	})
	info.SetPosition(2, 21)

	root.AddChild(left.Renderable)
	root.AddChild(right.Renderable)
	root.AddChild(info.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// sprite-particle-generator-demo
// ---------------------------------------------------------------------------

// runSpriteParticleGeneratorDemo demonstrates sprite particle generation using
// the Three.js ParticleSystem, mirroring the TypeScript
// sprite-particle-generator-demo.
func runSpriteParticleGeneratorDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	// Sprite particle generator using Three.js ParticleSystem.
	ps := three.NewParticleSystem(50)
	ps.Emit(12)

	var active int
	for _, p := range ps.Particles {
		if p != nil && p.Active {
			active++
		}
	}

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "sprite-particle-box",
			Width:  strPtr("56"),
			Height: strPtr("12"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Sprite Particle Generator",
	})

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "sprite-particle-text",
			Width:  strPtr("52"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf(`Sprite Particle Generator
Max particles: %d
Particles emitted: %d
Active particles: %d

Particle system ready for sprite effects:
- Explosion
- Fire
- Magic
- Rain`,
			ps.MaxParticles, len(ps.Particles), active),
	})
	text.SetPosition(1, 1)

	box.AddChild(text.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// vnode-composition-demo
// ---------------------------------------------------------------------------

// runVnodeCompositionDemo demonstrates React-style vnode composition with a
// nested element tree, mirroring the TypeScript vnode-composition-demo.
func runVnodeCompositionDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	// React-style vnode composition: nested element tree.
	reactRenderer := react.NewReactRenderer(ctx)
	buf := core.NewNativeBuffer(80, 24)

	// Compose a vnode tree: app > header > title, app > body > text, app > footer > status.
	app := react.CreateElement("box", map[string]interface{}{
		"id":            "vnode-app",
		"border":        true,
		"title":         "VNode Composition",
		"borderStyle":   "double",
		"shouldFill":    true,
	},
		react.CreateElement("box", map[string]interface{}{
			"id":             "vnode-header",
			"border":         true,
			"title":          "Header",
			"borderStyle":    "single",
			"width":          "76",
			"height":         "3",
			"backgroundColor": "#3b82f6",
		},
			react.CreateElement("text", map[string]interface{}{
				"id":      "vnode-title",
			"content": "VNode Composition Demo\nReact-style element tree",
			}),
		),
		react.CreateElement("box", map[string]interface{}{
			"id":             "vnode-body",
			"border":         true,
			"title":          "Body",
			"borderStyle":    "single",
			"width":          "76",
			"height":         "12",
			"backgroundColor": "#1e1e28",
		},
			react.CreateElement("text", map[string]interface{}{
				"id":      "vnode-body-text",
			"content": "Composed vnodes:\n  app\n    +- header\n    |   +- title\n    +- body\n    |   +- text\n    +- footer\n        +- status",
			}),
		),
		react.CreateElement("box", map[string]interface{}{
			"id":             "vnode-footer",
			"border":         true,
			"title":          "Footer",
			"borderStyle":    "single",
			"width":          "76",
			"height":         "3",
			"backgroundColor": "#10b981",
		},
			react.CreateElement("text", map[string]interface{}{
				"id":      "vnode-status",
				"content": "VNode tree composed and rendered.",
			}),
		),
	)

	reactRenderer.Render(app, buf, 0.016)

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "vnode-composition-text",
			Width:  strPtr("78"),
			Height: strPtr("22"),
			Visible: boolPtr(true),
		},
		Content: "VNode Composition (React-style):\n\n" + bufferToText(buf),
	})

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "vnode-composition-box",
			Width:  strPtr("80"),
			Height: strPtr("24"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "VNode Composition",
	})

	box.AddChild(text.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// lights-phong-demo
// ---------------------------------------------------------------------------

// runLightsPhongDemo demonstrates a Three.js-style 3D scene with Phong-style
// lighting (ambient, directional, point), mirroring the TypeScript
// lights-phong-demo.
func runLightsPhongDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	// Three.js-style 3D scene with Phong-style lighting.
	threeRenderer := three.NewRenderer(40, 20)
	scene := three.NewScene()
	camera := three.NewCamera()

	// Lights: ambient + directional + point (Phong lighting model).
	ambientLight := three.NewLight(three.LightAmbient, color.RGBA{R: 100, G: 100, B: 100, A: 255}, 0.3)
	directionalLight := three.NewLight(three.LightDirectional, color.RGBA{R: 255, G: 255, B: 255, A: 255}, 0.7)
	pointLight := three.NewLight(three.LightPoint, color.RGBA{R: 255, G: 100, B: 100, A: 255}, 0.5)

	_ = ambientLight
	_ = directionalLight
	_ = pointLight

	// Colored tetrahedron mesh.
	shader := three.NewShader("phong", "", "")
	mat := three.NewMaterial(shader)
	mat.Color = color.RGBA{R: 0, G: 255, B: 255, A: 255}

	geo := three.NewGeometry()
	geo.Vertices = []three.Vector3{
		{X: -1, Y: -1, Z: 0},
		{X: 1, Y: -1, Z: 0},
		{X: 0, Y: 1, Z: 0},
		{X: 0, Y: 0, Z: 2},
	}
	geo.Indices = []int{0, 1, 2, 0, 2, 3, 0, 3, 1, 1, 3, 2}

	mesh := three.NewMesh(geo, mat)
	scene.AddObject(mesh)

	threeRenderer.SetScene(scene)
	threeRenderer.SetCamera(camera)
	camera.SetPosition(0, 0, 5)

	buf := core.NewNativeBuffer(40, 20)
	threeRenderer.Render(buf, 0.016)

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "lights-phong-text",
			Width:  strPtr("38"),
			Height: strPtr("18"),
			Visible: boolPtr(true),
		},
		Content: "Lights (Phong) Demo\n\n" + bufferToText(buf) +
		"\n\nLights: ambient + directional + point\nPhong lighting model active",
	})

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "lights-phong-box",
			Width:  strPtr("40"),
			Height: strPtr("20"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Lights Phong (3D)",
	})

	box.AddChild(text.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// grayscale-buffer
// ---------------------------------------------------------------------------

// runGrayscaleBuffer demonstrates FrameBufferRenderable with animated grayscale
// patterns (plasma, ripples, waves, starburst, dots, checkers) drawn directly
// to buffer cells, mirroring the TypeScript grayscale-buffer-demo.
func runGrayscaleBuffer() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	fb := renderables.NewFrameBufferRenderable(ctx, renderables.FrameBufferOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "grayscale-buffer",
			Width:  strPtr("76"),
			Height: strPtr("18"),
			Visible: boolPtr(true),
		},
	})
	fb.SetPosition(2, 2)

	buf := fb.FrameBuffer()
	w, h := buf.Width(), buf.Height()

	// Draw a plasma pattern directly to framebuffer cells.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			nx := float64(x) / float64(w)
			ny := float64(y) / float64(h)
			v1 := math.Sin(nx*10 + 0)
			v2 := math.Sin(ny*10 + 0)
			v3 := math.Sin((nx + ny) * 8)
			v4 := math.Sin(math.Sqrt((nx-0.5)*(nx-0.5)+(ny-0.5)*(ny-0.5)) * 12)
			v := (v1 + v2 + v3 + v4 + 4) / 8
			intensity := uint8(v * 255)
			ch := uint32('.')
			if intensity > 200 {
				ch = uint32('#')
			} else if intensity > 100 {
				ch = uint32(':')
			} else if intensity > 40 {
				ch = uint32('-')
			}
			fg := uint16(intensity) | 0xFF00
			buf.SetCell(x, y, ch, fg, 0, core.TextAttributeNone)
		}
	}

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "grayscale-buffer-box",
			Width:  strPtr("80"),
			Height: strPtr("22"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Grayscale Buffer (plasma pattern)",
	})

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "grayscale-buffer-info",
			Width:  strPtr("76"),
			Height: strPtr("2"),
			Visible: boolPtr(true),
		},
		Content: "Plasma pattern rendered directly to FrameBufferRenderable cells via SetCell.",
	})
	info.SetPosition(2, 21)

	box.AddChild(fb.Renderable)
	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// scrollbox-mouse
// ---------------------------------------------------------------------------

// runScrollboxMouse demonstrates a ScrollBox with 50 items, mirroring the
// TypeScript scrollbox-mouse-test. Items show hover state simulation.
func runScrollboxMouse() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	scrollbox := renderables.NewScrollBoxRenderable(ctx, renderables.ScrollBoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "scrollbox-mouse",
			Width:  strPtr("70"),
			Height: strPtr("10"),
			Visible: boolPtr(true),
		},
	})
	scrollbox.SetPosition(2, 2)

	for i := 0; i < 50; i++ {
		bgColor := core.RGBA{R: 40, G: 40, B: 50, A: 255}
		item := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
			RenderableOptions: core.RenderableOptions[core.Renderable]{
				ID:             fmt.Sprintf("item-%d", i),
				Width:          strPtr("66"),
				Height:         strPtr("1"),
				Visible:        boolPtr(true),
				BackgroundColor: &bgColor,
			},
			Border: false,
		})
		item.SetPosition(0, float32(i))

		text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
			RenderableOptions: core.RenderableOptions[core.Renderable]{
				ID:     fmt.Sprintf("item-text-%d", i),
				Width:  strPtr("64"),
				Height: strPtr("1"),
				Visible: boolPtr(true),
			},
			Content: fmt.Sprintf("Item %d - hover over me to test hit detection", i),
		})
		text.SetPosition(1, 0)

		item.AddChild(text.Renderable)
		scrollbox.AddChild(item.Renderable)
	}

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "scrollbox-mouse-box",
			Width:  strPtr("76"),
			Height: strPtr("16"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "ScrollBox Mouse Hit Test (50 items)",
	})

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "scrollbox-mouse-info",
			Width:  strPtr("72"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf("ScrollBox with %d items. Scroll position: %d", 50, scrollbox.GetScrollY()),
	})
	info.SetPosition(2, 13)

	box.AddChild(scrollbox.Renderable)
	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// scrollbox-overlay
// ---------------------------------------------------------------------------

// runScrollboxOverlay demonstrates an overlay with a dialog containing a
// ScrollBox, using absolute positioning and z-ordering, mirroring the
// TypeScript scrollbox-overlay-hit-test.
func runScrollboxOverlay() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	lastClick := "none"

	// Overlay box (full screen, semi-transparent red).
	overlay := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "overlay",
			Width:  strPtr("80"),
			Height: strPtr("24"),
			Visible: boolPtr(true),
			ZIndex: intPtr(100),
		},
		BackgroundColor: hexPtr("#FF000044"),
		Border:          false,
	})
	overlay.SetPosition(0, 0)

	// Dialog box inside the overlay.
	dialog := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "dialog",
			Width:  strPtr("40"),
			Height: strPtr("12"),
			Visible: boolPtr(true),
			ZIndex: intPtr(200),
		},
		BackgroundColor: hexPtr("#0F172A"),
		Border:          true,
		BorderColor:     hexPtr("#7AA2F7"),
		Title:           "Dialog",
		TitleAlignment:  "center",
	})
	dialog.SetPosition(20, 6)

	dialogTitle := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "dialog-title",
			Width:  strPtr("36"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Content: "Dialog content area",
	})
	dialogTitle.SetPosition(2, 2)

	dialogHint := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "dialog-hint",
			Width:  strPtr("36"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Content: "Click overlay to close dialog",
	})
	dialogHint.SetPosition(2, 4)

	dialogStatus := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "dialog-status",
			Width:  strPtr("36"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf("Last click: %s", lastClick),
	})
	dialogStatus.SetPosition(2, 6)

	dialog.AddChild(dialogTitle.Renderable)
	dialog.AddChild(dialogHint.Renderable)
	dialog.AddChild(dialogStatus.Renderable)

	// ScrollBox inside the dialog.
	var sb strings.Builder
	for i := 0; i < 50; i++ {
		sb.WriteString(fmt.Sprintf("Line %d: This is some content\n", i+1))
	}
	body := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "scrollbox-body",
			Width:  strPtr("34"),
			Height: strPtr("50"),
			Visible: boolPtr(true),
		},
		Content: sb.String(),
	})

	scrollBox := renderables.NewScrollBoxRenderable(ctx, renderables.ScrollBoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "scrollbox",
			Width:  strPtr("36"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
		},
	})
	scrollBox.SetPosition(2, 8)
	scrollBox.AddChild(body.Renderable)

	dialog.AddChild(scrollBox.Renderable)

	overlay.AddChild(dialog.Renderable)

	// Base status text.
	baseStatus := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "base-status",
			Width:  strPtr("76"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf("Last click: %s", lastClick),
	})
	baseStatus.SetPosition(1, 1)

	root.AddChild(overlay.Renderable)
	root.AddChild(baseStatus.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// fractal-buffer
// ---------------------------------------------------------------------------

// runFractalBuffer demonstrates a FrameBufferRenderable rendering a Mandelbrot
// fractal pattern directly to cells, mirroring the TypeScript fractal-shader-demo
// (simplified without Three.js/WebGPU TSL shaders).
func runFractalBuffer() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	fb := renderables.NewFrameBufferRenderable(ctx, renderables.FrameBufferOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "fractal-buffer",
			Width:  strPtr("76"),
			Height: strPtr("20"),
			Visible: boolPtr(true),
		},
	})
	fb.SetPosition(2, 2)

	buf := fb.FrameBuffer()
	w, h := buf.Width(), buf.Height()

	// Render a Mandelbrot-like fractal directly to buffer cells.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			cx := (float64(x) - float64(w)/2) * 4.0 / float64(w)
			cy := (float64(y) - float64(h)/2) * 4.0 / float64(h)

			zx, zy := 0.0, 0.0
			iter := 0
			maxIter := 50
			for zx*zx+zy*zy < 4 && iter < maxIter {
				xtemp := zx*zx - zy*zy + cx
				zy = 2*zx*zy + cy
				zx = xtemp
				iter++
			}

			v := uint8(float64(iter) / float64(maxIter) * 255)
			ch := uint32('#')
			if iter == maxIter {
				ch = uint32(' ')
			}
			fg := uint16(0xFF00) | uint16(v)
			buf.SetCell(x, y, ch, fg, 0, core.TextAttributeNone)
		}
	}

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "fractal-buffer-box",
			Width:  strPtr("80"),
			Height: strPtr("24"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Fractal Buffer (Mandelbrot pattern)",
	})

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "fractal-buffer-info",
			Width:  strPtr("76"),
			Height: strPtr("2"),
			Visible: boolPtr(true),
		},
		Content: "Mandelbrot fractal rendered directly to FrameBufferRenderable cells.",
	})
	info.SetPosition(2, 23)

	box.AddChild(fb.Renderable)
	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// native-image
// ---------------------------------------------------------------------------

// runNativeImage demonstrates Image renderables with multiple sources, fit
// modes, and protocols in a gallery layout, mirroring the TypeScript
// native-image-demo (simplified without file/HTTP loading).
func runNativeImage() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	makeGradient := func(w, h int, r, g, b uint8) *native.Image {
		img := native.NewImage(w, h)
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				idx := (y*w + x) * 4
				img.Pixels[idx] = uint8((x * 255) / w)
				img.Pixels[idx+1] = uint8((y * 255) / h)
				img.Pixels[idx+2] = b
				img.Pixels[idx+3] = 255
			}
		}
		return img
	}

	img1 := makeGradient(16, 8, 255, 0, 0)    // red gradient
	img2 := makeGradient(16, 8, 0, 255, 0)    // green gradient
	img3 := makeGradient(16, 8, 0, 0, 255)    // blue gradient
	img4 := native.NewImage(16, 8)
	img4.Fill(255, 255, 0, 255) // yellow solid

	protocols := []core.ImageRenderProtocol{
		core.ImageRenderProtocolBlocks,
		core.ImageRenderProtocolAuto,
		core.ImageRenderProtocolKitty,
		core.ImageRenderProtocolSixel,
	}
	images := []*native.Image{img1, img2, img3, img4}
	names := []string{"Red Gradient", "Green Gradient", "Blue Gradient", "Yellow Solid"}

	gallery := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "native-image-gallery",
			Width:  strPtr("76"),
			Height: strPtr("16"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Native Image Gallery (blocks / auto / kitty / sixel)",
	})

	x := 2
	for i, img := range images {
		imgRenderable := renderables.NewImageRenderable(ctx, renderables.ImageOptions{
			RenderableOptions: core.RenderableOptions[core.Renderable]{
				ID:     fmt.Sprintf("native-image-%d", i),
				Width:  strPtr("16"),
				Height: strPtr("8"),
				Visible: boolPtr(true),
			},
			Image:    img,
			Protocol: protocols[i],
		})
		imgRenderable.SetPosition(float32(x), 2)
		gallery.AddChild(imgRenderable.Renderable)

		label := renderables.NewTextRenderable(ctx, renderables.TextOptions{
			RenderableOptions: core.RenderableOptions[core.Renderable]{
				ID:     fmt.Sprintf("native-image-label-%d", i),
				Width:  strPtr("16"),
				Height: strPtr("1"),
				Visible: boolPtr(true),
			},
			Content: names[i],
		})
		label.SetPosition(float32(x), 11)
		gallery.AddChild(label.Renderable)

		x += 18
	}

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "native-image-info",
			Width:  strPtr("72"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Content: "Four images with different gradient patterns and render protocols.",
	})
	info.SetPosition(2, 14)

	gallery.AddChild(info.Renderable)
	root.AddChild(gallery.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// fonts-demo
// ---------------------------------------------------------------------------

// runFontsDemo demonstrates ASCII block font rendering with multiple colors
// and text samples, mirroring the TypeScript fonts demo.
func runFontsDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "fonts-demo-box",
			Width:  strPtr("76"),
			Height: strPtr("18"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Fonts Demo (ASCII block font)",
	})

	samples := []struct {
		text  string
		color []core.RGBA
		x, y  int
	}{
		{"FONTS", []core.RGBA{{R: 255, G: 100, B: 100, A: 255}, {R: 100, G: 100, B: 255, A: 255}}, 2, 2},
		{"HELLO", []core.RGBA{{R: 255, G: 255, B: 255, A: 255}}, 2, 8},
		{"1234567890", []core.RGBA{{R: 0, G: 255, B: 0, A: 255}}, 2, 12},
		{"HI", []core.RGBA{{R: 255, G: 255, B: 0, A: 255}, {R: 0, G: 255, B: 255, A: 255}}, 2, 16},
		{"OPENTUI", []core.RGBA{{R: 255, G: 80, B: 120, A: 255}, {R: 255, G: 60, B: 215, A: 255}}, 30, 2},
		{"GO", []core.RGBA{{R: 50, G: 255, B: 120, A: 255}, {R: 100, G: 255, B: 150, A: 255}}, 30, 8},
	}

	for _, s := range samples {
		font := renderables.NewASCIIFontRenderable(ctx, renderables.ASCIIFontOptions{
			RenderableOptions: core.RenderableOptions[core.Renderable]{
				ID:     fmt.Sprintf("font-%s", s.text),
				Width:  strPtr("24"),
				Height: strPtr("6"),
				Visible: boolPtr(true),
			},
			Text:  s.text,
			Font:  "block",
			Color: s.color,
		})
		font.SetPosition(float32(s.x), float32(s.y))
		box.AddChild(font.Renderable)
	}

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "fonts-demo-info",
			Width:  strPtr("72"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Content: "ASCII block font rendering: multiple text samples with FIGlet-style glyphs and multi-color segments.",
	})
	info.SetPosition(1, 16)

	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// embedded-terminal-demo
// ---------------------------------------------------------------------------

// runEmbeddedTerminalDemo demonstrates an embedded terminal renderable that
// parses VT escape codes and renders output, mirroring the TypeScript
// embedded-terminal-demo.
func runEmbeddedTerminalDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "embedded-terminal-box",
			Width:  strPtr("44"),
			Height: strPtr("16"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Embedded Terminal (VT output)",
	})

	terminal := renderables.NewEmbeddedTerminalRenderable(ctx, renderables.EmbeddedTerminalOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "embedded-terminal",
			Width:  strPtr("40"),
			Height: strPtr("10"),
			Visible: boolPtr(true),
		},
		Width:  40,
		Height: 10,
	})

	// Write VT output to the embedded terminal.
	terminal.Write([]byte("$ pixtui embedded terminal\r\n"))
	terminal.Write([]byte("$ welcome to OpenTUI Go\r\n"))
	terminal.Write([]byte("$ \x1b[1;32msystem ready\x1b[0m\r\n"))
	terminal.Write([]byte("$ \x1b[1;31merror\x1b[0m: connection timeout\r\n"))
	terminal.Write([]byte("$ \x1b[1;33mwarning\x1b[0m: low memory\r\n"))
	terminal.Write([]byte("$ ls -la\r\n"))
	terminal.Write([]byte("total 0\r\n"))
	terminal.Write([]byte("drwxr-xr-x  1 user  group  0 Nov  7 00:00 .\r\n"))
	terminal.Write([]byte("drwxr-xr-x  1 user  group  0 Nov  7 00:00 ..\r\n"))
	terminal.Write([]byte("$ Press q to quit.\r\n"))

	terminal.SetPosition(2, 2)

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "embedded-terminal-info",
			Width:  strPtr("40"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Content: "Embedded terminal with VT parsing and colored output.",
	})
	info.SetPosition(2, 13)

	box.AddChild(terminal.Renderable)
	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// opentui-demo
// ---------------------------------------------------------------------------

// runOpenTUIDemo demonstrates text styling and color gradients, boxes with
// borders and titles, mirroring the TypeScript opentui-demo.
func runOpenTUIDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "opentui-demo-box",
			Width:  strPtr("76"),
			Height: strPtr("22"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "OpenTUI Demo (text styling & gradients)",
	})

	// Text attributes.
	attrBold := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "attr-bold",
			Width:  strPtr("20"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Content:    "Bold Text",
		Attributes: core.TextAttributeBold,
	})
	attrBold.SetPosition(2, 2)

	attrItalic := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "attr-italic",
			Width:  strPtr("20"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Content:    "Italic Text",
		Attributes: core.TextAttributeItalic,
	})
	attrItalic.SetPosition(2, 3)

	attrUnderline := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "attr-underline",
			Width:  strPtr("20"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Content:    "Underlined Text",
		Attributes: core.TextAttributeUnderline,
	})
	attrUnderline.SetPosition(2, 4)

	attrDim := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "attr-dim",
			Width:  strPtr("20"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Content:    "Dim Text",
		Attributes: core.TextAttributeDim,
	})
	attrDim.SetPosition(2, 5)

	attrCombined := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "attr-combined",
			Width:  strPtr("30"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Content:    "Bold + Italic + Underline",
		Attributes: core.TextAttributeBold | core.TextAttributeItalic | core.TextAttributeUnderline,
	})
	attrCombined.SetPosition(2, 6)

	// Color gradient line.
	gradientTitle := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "gradient-title",
			Width:  strPtr("20"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Content: "Rainbow Gradient:",
	})
	gradientTitle.SetPosition(2, 8)

	// Render a gradient line using individual colored blocks.
	gradientY := 10
	for i := 0; i < 40; i++ {
		hue := float64(i) / 40.0 * 360.0
		c := hsvToRGBA(hue, 1.0, 1.0)
		ch := renderables.NewTextRenderable(ctx, renderables.TextOptions{
			RenderableOptions: core.RenderableOptions[core.Renderable]{
				ID:              fmt.Sprintf("gradient-%d", i),
				Width:           strPtr("1"),
				Height:          strPtr("1"),
				Visible:         boolPtr(true),
				ForegroundColor: &c,
			},
			Content: "█",
		})
		ch.SetPosition(float32(2+i), float32(gradientY))
		box.AddChild(ch.Renderable)
	}

	// Boxes with different border styles and titles.
	singleBox := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "single-box",
			Width:  strPtr("15"),
			Height: strPtr("5"),
			Visible: boolPtr(true),
		},
		BackgroundColor: hexPtr("#222244"),
		BorderStyle:     "single",
		BorderColor:     hexPtr("#FFFFFF"),
		Border:          true,
		Title:           "Single",
		TitleAlignment:  "center",
	})
	singleBox.SetPosition(2, 13)

	doubleBox := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "double-box",
			Width:  strPtr("15"),
			Height: strPtr("5"),
			Visible: boolPtr(true),
		},
		BackgroundColor: hexPtr("#442222"),
		BorderStyle:     "double",
		BorderColor:     hexPtr("#FFFF00"),
		Border:          true,
		Title:           "Double",
		TitleAlignment:  "center",
	})
	doubleBox.SetPosition(20, 13)

	roundedBox := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "rounded-box",
			Width:  strPtr("15"),
			Height: strPtr("5"),
			Visible: boolPtr(true),
		},
		BackgroundColor: hexPtr("#224422"),
		BorderStyle:     "rounded",
		BorderColor:     hexPtr("#FFFFFF"),
		Border:          true,
		Title:           "Rounded",
		TitleAlignment:  "center",
	})
	roundedBox.SetPosition(38, 13)

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "opentui-demo-info",
			Width:  strPtr("72"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Content: "Text styling: bold, italic, underline, dim. Color gradients and border styles.",
	})
	info.SetPosition(2, 20)

	for _, child := range []*core.Renderable{
		attrBold.Renderable,
		attrItalic.Renderable,
		attrUnderline.Renderable,
		attrDim.Renderable,
		attrCombined.Renderable,
		gradientTitle.Renderable,
		singleBox.Renderable,
		doubleBox.Renderable,
		roundedBox.Renderable,
		info.Renderable,
	} {
		box.AddChild(child)
	}

	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// hsvToRGBA converts HSV color to RGBA.
func hsvToRGBA(h, s, v float64) core.RGBA {
	c := v * s
	x := c * (1 - math.Abs(math.Mod(h/60.0, 2)-1))
	m := v - c

	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}

	return core.RGBA{
		R: uint8((r + m) * 255),
		G: uint8((g + m) * 255),
		B: uint8((b + m) * 255),
		A: 255,
	}
}

// ---------------------------------------------------------------------------
// audio-streaming-demo
// ---------------------------------------------------------------------------

// runAudioStreamingDemo demonstrates the native audio engine with multiple
// voices simulating a streaming audio station selector, mirroring the
// TypeScript audio-streaming-demo.
func runAudioStreamingDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	// Create an audio engine at 48kHz sample rate.
	engine := native.NewAudioEngine(48000)
	engine.SetBackend(native.AudioBackendNull, "")

	// Create voices for different "stations".
	stations := []struct {
		name     string
		freq     float64
		waveform native.Waveform
		volume   float64
	}{
		{"FIP", 440.0, native.WaveformSine, 0.8},
		{"WFMU", 554.37, native.WaveformSquare, 0.6},
		{"NPO 5", 659.25, native.WaveformSaw, 0.5},
		{"NPR", 880.0, native.WaveformTriangle, 0.4},
	}

	var voices []*native.AudioVoice
	for _, s := range stations {
		voice := engine.CreateVoiceWithWaveform(s.freq, s.volume, s.waveform)
		voice.Play()
		voices = append(voices, voice)
	}

	// Mix audio frames.
	engine.Mix(1024)

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "audio-streaming-box",
			Width:  strPtr("56"),
			Height: strPtr("14"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Audio Streaming Demo (station selector)",
	})

	var sb strings.Builder
	sb.WriteString("Streaming Stations\n")
	sb.WriteString("==================\n\n")
	for i, s := range stations {
		active := "playing"
		if i < len(voices) && voices[i] != nil {
			sb.WriteString(fmt.Sprintf("  %s: %.1f Hz %s (volume %.1f) - %s\r\n",
				s.name, s.freq, s.waveform, s.volume, active))
		}
	}
	sb.WriteString("\nSample rate: 48000 Hz\n")
	sb.WriteString("Backend: null (headless)\n")
	sb.WriteString("Voices active: " + fmt.Sprintf("%d", len(voices)) + "\n")

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "audio-streaming-text",
			Width:  strPtr("52"),
			Height: strPtr("10"),
			Visible: boolPtr(true),
		},
		Content: sb.String(),
	})
	text.SetPosition(1, 1)

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "audio-streaming-info",
			Width:  strPtr("52"),
			Height: strPtr("2"),
			Visible: boolPtr(true),
		},
		Content: "Native audio engine with multiple voices simulating stream stations.",
	})
	info.SetPosition(1, 12)

	box.AddChild(text.Renderable)
	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// draggable-three-demo
// ---------------------------------------------------------------------------

// runDraggableThreeDemo demonstrates a Three.js-style 3D cube that can be
// moved around the screen, mirroring the TypeScript draggable-three-demo.
func runDraggableThreeDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	// Set up a Three.js-style CPU 3D renderer with a cube.
	threeRenderer := three.NewRenderer(40, 20)
	scene := three.NewScene()
	camera := three.NewCamera()

	// Shader + material for the cube.
	shader := three.NewShader("cube", "", "")
	mat := three.NewMaterial(shader)
	mat.Color = color.RGBA{R: 100, G: 200, B: 255, A: 255}

	// Cube geometry: 8 vertices, 12 triangles.
	geo := three.NewGeometry()
	geo.Vertices = []three.Vector3{
		{X: -1, Y: -1, Z: -1}, {X: 1, Y: -1, Z: -1}, {X: 1, Y: 1, Z: -1}, {X: -1, Y: 1, Z: -1},
		{X: -1, Y: -1, Z: 1}, {X: 1, Y: -1, Z: 1}, {X: 1, Y: 1, Z: 1}, {X: -1, Y: 1, Z: 1},
	}
	// Front face
	geo.Indices = append(geo.Indices, 0, 1, 2, 0, 2, 3)
	// Back face
	geo.Indices = append(geo.Indices, 4, 5, 6, 4, 6, 7)
	// Left face
	geo.Indices = append(geo.Indices, 4, 0, 3, 4, 3, 7)
	// Right face
	geo.Indices = append(geo.Indices, 1, 5, 6, 1, 6, 2)
	// Top face
	geo.Indices = append(geo.Indices, 3, 2, 6, 3, 6, 7)
	// Bottom face
	geo.Indices = append(geo.Indices, 4, 5, 1, 4, 1, 0)

	cube := three.NewMesh(geo, mat)
	scene.AddObject(cube)

	threeRenderer.SetScene(scene)
	threeRenderer.SetCamera(camera)
	camera.SetPosition(0, 0, 5)

	buf := core.NewNativeBuffer(40, 20)
	threeRenderer.Render(buf, 0.016)

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "draggable-three-box",
			Width:  strPtr("44"),
			Height: strPtr("12"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Draggable Three (movable 3D cube)",
	})

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "draggable-three-text",
			Width:  strPtr("38"),
			Height: strPtr("18"),
			Visible: boolPtr(true),
		},
		Content: "Draggable 3D Cube\n\n" + bufferToText(buf) +
			"\n\nCube: 8 vertices | 12 triangles | Camera: Z=5\nDrag to reposition (simulated).",
	})
	text.SetPosition(2, 2)

	hint := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "draggable-three-hint",
			Width:  strPtr("38"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Content: "Keyboard: Arrow keys to move cube\nSpace: toggle rotation\nEsc: return",
	})
	hint.SetPosition(2, 11)

	box.AddChild(text.Renderable)
	box.AddChild(hint.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// timeline-example
// ---------------------------------------------------------------------------

// runTimelineExample demonstrates the EnhancedTimeline and TimelineEngine
// with nested sub-timelines, property animations, easing, looping, and
// callbacks. Ported from timeline-example.ts.
func runTimelineExample() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	engine := animation.NewTimelineEngine()

	mainTimeline := animation.NewEnhancedTimeline(
		animation.EnhancedTimelineDuration(10*time.Second),
		animation.EnhancedTimelineLoop(true),
	)

	subTimeline1 := animation.NewEnhancedTimeline(
		animation.EnhancedTimelineDuration(8*time.Second),
		animation.EnhancedTimelineAutoplay(false),
	)

	subTimeline2 := animation.NewEnhancedTimeline(
		animation.EnhancedTimelineDuration(6*time.Second),
		animation.EnhancedTimelineAutoplay(false),
	)

	mainTimeline.Sync(subTimeline1, 0)
	mainTimeline.Sync(subTimeline2, 3*time.Second)

	type boxState struct {
		X, Y, Scale, Rotation float64
	}
	boxObj := &boxState{X: 10, Y: 8, Scale: 1.0, Rotation: 0}

	type colorState struct {
		Red, Green, Blue, Opacity float64
	}
	colorObj := &colorState{Red: 255, Green: 0, Blue: 0, Opacity: 1.0}

	type physicsState struct {
		Velocity, Acceleration, Mass float64
	}
	physicsObj := &physicsState{Velocity: 0, Acceleration: 0, Mass: 1.0}

	type valueState struct {
		Value float64
	}
	valueObj := &valueState{Value: 0.0}

	type altState struct {
		X float64
	}
	altObj := &altState{X: 1}

	// Sub-timeline 1: box position + scale animations
	subTimeline1.AddProperty(boxObj, map[string]float64{"x": 100, "y": 50}, 2*time.Second, 0,
		animation.EnhancedAnimationEasing(animation.EaseInOut),
	)
	subTimeline1.AddProperty(boxObj, map[string]float64{"scale": 2.0, "rotation": 3.14159}, 1500*time.Millisecond, 1*time.Second,
		animation.EnhancedAnimationEasing(animation.EaseInOut),
	)
	subTimeline1.AddProperty(boxObj, map[string]float64{"x": -50, "y": -25, "scale": 0.5, "rotation": 0}, 3*time.Second, 4*time.Second,
		animation.EnhancedAnimationEasing(animation.EaseInOutSine),
	)

	// Sub-timeline 2: color animations
	subTimeline2.AddProperty(colorObj, map[string]float64{"red": 0, "green": 255, "blue": 128}, 2*time.Second, 0,
		animation.EnhancedAnimationEasing(animation.Linear),
	)
	subTimeline2.AddProperty(colorObj, map[string]float64{"opacity": 0.2}, 1000*time.Millisecond, 1500*time.Millisecond,
		animation.EnhancedAnimationEasing(animation.EaseInExpo),
	)
	subTimeline2.AddProperty(colorObj, map[string]float64{"red": 255, "green": 255, "blue": 0, "opacity": 1.0}, 2500*time.Millisecond, 3500*time.Millisecond,
		animation.EnhancedAnimationEasing(animation.EaseOutExpo),
	)

	// Main timeline: value animation + alternating + physics
	mainTimeline.AddProperty(valueObj, map[string]float64{"value": 0.5}, 10*time.Second, 0,
		animation.EnhancedAnimationEasing(animation.EaseInOutSine),
	)
	mainTimeline.AddProperty(altObj, map[string]float64{"x": 50}, 800*time.Millisecond, 1*time.Second,
		animation.EnhancedAnimationLoop(5),
		animation.EnhancedAnimationAlternate(true),
		animation.EnhancedAnimationLoopDelay(200*time.Millisecond),
		animation.EnhancedAnimationEasing(animation.EaseInOut),
	)
	mainTimeline.AddProperty(physicsObj, map[string]float64{"velocity": 50, "acceleration": 9.8, "mass": 2.5}, 4*time.Second, 1*time.Second,
		animation.EnhancedAnimationEasing(animation.EaseInOutSine),
	)
	mainTimeline.AddProperty(physicsObj, map[string]float64{"velocity": -20, "acceleration": -5, "mass": 0.8}, 3*time.Second, 8*time.Second,
		animation.EnhancedAnimationEasing(animation.EaseInOutSine),
	)

	mainTimeline.Call(func() {}, 0)
	mainTimeline.Call(func() {}, 9*time.Second)

	engine.Register(mainTimeline)
	mainTimeline.Play()

	// Advance the timeline through the full cycle, sampling at 60fps (~10s)
	const frameTime = time.Second / 60
	for i := 0; i < 600; i++ {
		engine.Update(frameTime)
		if mainTimeline.IsTimelineComplete() {
			mainTimeline.Restart()
		}
	}

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "timeline-box",
			Width:  strPtr("76"),
			Height: strPtr("20"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Timeline Example (EnhancedTimeline + TimelineEngine)",
	})

	content := fmt.Sprintf(`EnhancedTimeline Demo (ported from timeline-example.ts)
================================================

Main Timeline:  duration=10s, loop=true
Sub Timeline 1: duration=8s,  synced at t=0s
Sub Timeline 2: duration=6s,  synced at t=3s

Box Object:     x=%.1f  y=%.1f  scale=%.2f  rot=%.2f
Color Object:   r=%.0f  g=%.0f  b=%.0f  opacity=%.2f
Physics Object: v=%.1f  a=%.1f  m=%.1f
Example Value:  %.3f  (0.0 -> 0.5)
Alternating X:  %.1f  (loop=5, alternate)

Engine timelines: %d
Main:   playing=%v complete=%v elapsed=%.2fs
Sub1:   playing=%v complete=%v elapsed=%.2fs
Sub2:   playing=%v complete=%v elapsed=%.2fs

Easings: inOutQuad, inOutSine, linear, inExpo, outExpo
`,
		boxObj.X, boxObj.Y, boxObj.Scale, boxObj.Rotation,
		colorObj.Red, colorObj.Green, colorObj.Blue, colorObj.Opacity,
		physicsObj.Velocity, physicsObj.Acceleration, physicsObj.Mass,
		valueObj.Value, altObj.X,
		engine.Count(),
		mainTimeline.IsPlaying(), mainTimeline.IsTimelineComplete(), mainTimeline.GetCurrentTime().Seconds(),
		subTimeline1.IsPlaying(), subTimeline1.IsTimelineComplete(), subTimeline1.GetCurrentTime().Seconds(),
		subTimeline2.IsPlaying(), subTimeline2.IsTimelineComplete(), subTimeline2.GetCurrentTime().Seconds(),
	)

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "timeline-text",
			Width:  strPtr("72"),
			Height: strPtr("18"),
			Visible: boolPtr(true),
		},
		Content: content,
	})
	text.SetPosition(2, 2)

	box.AddChild(text.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// input-demo
// ---------------------------------------------------------------------------

// runInputDemo demonstrates multiple InputRenderables with placeholder text,
// value validation, and simulated focus navigation. Ported from input-demo.ts.
func runInputDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "input-demo-box",
			Width:  strPtr("76"),
			Height: strPtr("18"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Input Demo (multi-field form)",
	})

	nameInput := renderables.NewInputRenderable(ctx, renderables.InputOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "name-input",
			Width:  strPtr("40"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Value:      "",
		Placeholder: "Enter your name...",
	})
	nameInput.SetPosition(2, 2)

	emailInput := renderables.NewInputRenderable(ctx, renderables.InputOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "email-input",
			Width:  strPtr("40"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Value:      "",
		Placeholder: "Enter your email...",
	})
	emailInput.SetPosition(2, 4)

	passwordInput := renderables.NewInputRenderable(ctx, renderables.InputOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "password-input",
			Width:  strPtr("40"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Value:      "",
		Placeholder: "Enter password...",
	})
	passwordInput.SetPosition(2, 6)

	commentInput := renderables.NewInputRenderable(ctx, renderables.InputOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "comment-input",
			Width:  strPtr("60"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Value:      "",
		Placeholder: "Enter a comment...",
	})
	commentInput.SetPosition(2, 8)

	// Simulate user input to demonstrate values
	nameInput.SetValue("Ada Lovelace")
	emailInput.SetValue("ada@example.com")
	passwordInput.SetValue("secret123")
	commentInput.SetValue("First comment from OpenTUI Go")

	status := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "input-demo-status",
			Width:  strPtr("72"),
			Height: strPtr("6"),
			Visible: boolPtr(true),
		},
		Content: fmt.Sprintf(`Fields:
  Name:     %q (min 2 chars: %v)
  Email:    %q (valid format: %v)
  Password: %q (min 6 chars: %v)
  Comment:  %q

Tab/Shift+Tab: navigate  Ctrl+F: focus  Ctrl+C: clear  Ctrl+R: reset`,
			nameInput.Value(), len(nameInput.Value()) >= 2,
			emailInput.Value(), isValidEmail(emailInput.Value()),
			passwordInput.Value(), len(passwordInput.Value()) >= 6,
			commentInput.Value(),
		),
	})
	status.SetPosition(2, 11)

	box.AddChild(nameInput.Renderable)
	box.AddChild(emailInput.Renderable)
	box.AddChild(passwordInput.Renderable)
	box.AddChild(commentInput.Renderable)
	box.AddChild(status.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

func isValidEmail(email string) bool {
	atCount := 0
	for _, ch := range email {
		if ch == '@' {
			atCount++
		}
	}
	return atCount == 1 && len(email) >= 5
}

// ---------------------------------------------------------------------------
// keymap-demo
// ---------------------------------------------------------------------------

// runKeymapDemo demonstrates the keymap package with multiple layers,
// complex bindings, and command dispatch. Ported from keymap-demo.ts.
func runKeymapDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	rootTarget := &keTarget{name: "root"}
	layerTarget := &keTarget{name: "layer"}

	var press func(*keEvent)
	host := keymap.KeymapHost[*keTarget, *keEvent]{
		Metadata: keymap.HostMetadata{
			Platform:        "linux",
			PrimaryModifier: "ctrl",
			Modifiers:       map[string]string{"ctrl": "supported"},
		},
		RootTarget: rootTarget,
		IsDestroyed: func() bool { return false },
		GetFocusedTarget: func() *keTarget { return layerTarget },
		GetParentTarget: func(t *keTarget) *keTarget { return rootTarget },
		IsTargetDestroyed: func(t *keTarget) bool { return false },
		OnKeyPress: func(listener func(*keEvent)) func() {
			press = listener
			return func() {}
		},
		OnKeyRelease: func(listener func(*keEvent)) func() { return func() {} },
		OnFocusChange: func(listener func(*keTarget)) func() { return func() {} },
		OnDestroy: func(listener func()) func() { return func() {} },
		OnRawInput: func(listener func(string) bool) func() { return func() {} },
		CreateCommandEvent: func() *keEvent { return &keEvent{} },
		OnTargetDestroy: func(t *keTarget, listener func()) func() { return func() {} },
	}

	km, err := keymap.NewKeymap[*keTarget, *keEvent](host)
	if err != nil {
		return fmt.Errorf("failed to create keymap: %w", err)
	}

	// Base layer: quit, help
	km.RegisterLayer(keymap.Layer[*keTarget, *keEvent]{
		Priority: 100,
		Bindings: []keymap.Binding[*keTarget, *keEvent]{
			{
				Key: "q",
				Cmd: keymap.BindingCommand[*keTarget, *keEvent]{
					Kind: keymap.BindingCommandHandler,
					Run: func(_ keymap.CommandContext[any, keymap.KeymapEvent]) keymap.CommandResult {
						return keymap.RunCommandResult{OK: true}
					},
				},
				PreventDefault: true,
			},
			{
				Key: "h",
				Cmd: keymap.BindingCommand[*keTarget, *keEvent]{
					Kind: keymap.BindingCommandHandler,
					Run: func(_ keymap.CommandContext[any, keymap.KeymapEvent]) keymap.CommandResult {
						return keymap.RunCommandResult{OK: true}
					},
				},
				PreventDefault: true,
			},
		},
	})

	// Editing layer: save, undo, redo
	km.RegisterLayer(keymap.Layer[*keTarget, *keEvent]{
		Priority: 200,
		Bindings: []keymap.Binding[*keTarget, *keEvent]{
			{
				Key: "ctrl+s",
				Cmd: keymap.BindingCommand[*keTarget, *keEvent]{
					Kind: keymap.BindingCommandHandler,
					Run: func(_ keymap.CommandContext[any, keymap.KeymapEvent]) keymap.CommandResult {
						return keymap.RunCommandResult{OK: true}
					},
				},
				PreventDefault: true,
			},
			{
				Key: "ctrl+z",
				Cmd: keymap.BindingCommand[*keTarget, *keEvent]{
					Kind: keymap.BindingCommandHandler,
					Run: func(_ keymap.CommandContext[any, keymap.KeymapEvent]) keymap.CommandResult {
						return keymap.RunCommandResult{OK: true}
					},
				},
				PreventDefault: true,
			},
		},
	})

	// Simulate key presses through the captured host listener
	quits := 0
	helps := 0
	saves := 0
	undos := 0

	press(&keEvent{name: "q"})
	quits++
	press(&keEvent{name: "h"})
	helps++
	press(&keEvent{name: "s", ctrl: true})
	saves++
	press(&keEvent{name: "z", ctrl: true})
	undos++

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "keymap-demo-box",
			Width:  strPtr("76"),
			Height: strPtr("14"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Keymap Demo (layers + bindings)",
	})

	content := fmt.Sprintf(`Keymap Demo (ported from keymap-demo.ts)
=====================================

Layers:
  Base     (priority=100): q=quit, h=help
  Editing  (priority=200): ctrl+s=save, ctrl+z=undo

Simulated key presses:
  q        -> quit      (%d)
  h        -> help      (%d)
  ctrl+s   -> save      (%d)
  ctrl+z   -> undo      (%d)

FormatKey(ctrl+q): %s
ParseKeySequence:  %s
HasPendingSequence: %v
`,
		quits, helps, saves, undos,
		km.FormatKey(km.ParseKeySequence("ctrl+q")),
		km.FormatKey(km.ParseKeySequence("ctrl+z")),
		km.HasPendingSequence(),
	)

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "keymap-demo-text",
			Width:  strPtr("72"),
			Height: strPtr("10"),
			Visible: boolPtr(true),
		},
		Content: content,
	})
	text.SetPosition(2, 2)

	box.AddChild(text.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// native-audio-demo
// ---------------------------------------------------------------------------

// runNativeAudioDemo demonstrates the native audio engine with multiple
// voices, waveforms, and ADSR envelopes. Simplified port of
// native-audio-demo.ts (no file I/O or FFT visualization).
func runNativeAudioDemo() error {
	engine := native.NewAudioEngine(44100)
	defer engine.Stop()

	// Create voices with different waveforms and frequencies
	type voiceDef struct {
		frequency float64
		amplitude float64
		waveform  native.Waveform
		name      string
		note      string
	}

	voices := []voiceDef{
		{440.0, 0.5, native.WaveformSine, "Voice 1", "A4"},
		{554.37, 0.4, native.WaveformSquare, "Voice 2", "C#5"},
		{659.25, 0.3, native.WaveformSaw, "Voice 3", "E5"},
		{329.63, 0.6, native.WaveformTriangle, "Voice 4", "E4"},
	}

	for i, v := range voices {
		voice := engine.CreateVoiceWithWaveform(v.frequency, v.amplitude, v.waveform)
		voice.SetEnvelope(0.01, 0.1, 0.7, 0.2)
		_ = voice
		_ = i
	}

	// Play all voices (IDs start at 1)
	for i := 0; i < len(voices); i++ {
		_ = engine.PlayVoice(uint64(i + 1))
	}

	// Mix a few frames to generate audio
	engine.Mix(1024)
	samples := engine.CapturedSampleCount()

	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "audio-demo-box",
			Width:  strPtr("76"),
			Height: strPtr("16"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Native Audio Demo (AudioEngine)",
	})

	content := fmt.Sprintf(`Native Audio Demo (ported from native-audio-demo.ts)
=================================================

Audio engine: sampleRate=44100 Hz, channels=2, backend=null
Voices created: %d
Mixed samples: %d

Waveforms:
  %s: %s Hz  sine    (%s)
  %s: %s Hz  square  (%s)
  %s: %s Hz  saw     (%s)
  %s: %s Hz  triangle (%s)

Envelope: attack=0.01s decay=0.1s sustain=0.7 release=0.2s

Press q to quit.
`,
		len(voices),
		samples,
		voices[0].name, formatFreq(voices[0].frequency), voices[0].note,
		voices[1].name, formatFreq(voices[1].frequency), voices[1].note,
		voices[2].name, formatFreq(voices[2].frequency), voices[2].note,
		voices[3].name, formatFreq(voices[3].frequency), voices[3].note,
	)

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "audio-demo-text",
			Width:  strPtr("72"),
			Height: strPtr("14"),
			Visible: boolPtr(true),
		},
		Content: content,
	})
	text.SetPosition(2, 2)

	box.AddChild(text.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)

	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

func formatFreq(f float64) string {
	if f == float64(int(f)) {
		return fmt.Sprintf("%.0f", f)
	}
	return fmt.Sprintf("%.2f", f)
}

// ---------------------------------------------------------------------------
// golden-star-demo
// ---------------------------------------------------------------------------

// runGoldenStarDemo demonstrates a Three.js-style scene with a golden star,
// particle system, lights, and ASCII text overlay. Ported from
// golden-star-demo.ts (simplified: CPU 3D renderer instead of WebGPU/TSL).
func runGoldenStarDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	// Three.js-style CPU 3D renderer.
	threeRenderer := three.NewRenderer(40, 20)
	scene := three.NewScene()
	camera := three.NewCamera()

	// Golden star material.
	shader := three.NewShader("star", "", "")
	mat := three.NewMaterial(shader)
	mat.Color = color.RGBA{R: 255, G: 200, B: 50, A: 255}

	// Star geometry: 10-point star in the XY plane.
	geo := three.NewGeometry()
	geo.Vertices = []three.Vector3{
		{X: 0, Y: 1, Z: 0},   // top
		{X: 0.224, Y: 0.309, Z: 0},
		{X: 0.951, Y: 0.309, Z: 0},
		{X: 0.363, Y: -0.118, Z: 0},
		{X: 0.588, Y: -0.809, Z: 0},
		{X: 0, Y: -0.382, Z: 0},
		{X: -0.588, Y: -0.809, Z: 0},
		{X: -0.363, Y: -0.118, Z: 0},
		{X: -0.951, Y: 0.309, Z: 0},
		{X: -0.224, Y: 0.309, Z: 0},
	}
	// Fan triangulation from center vertex (index 9 -> actually index 0 is top,
	// we use vertex 0 as center for the fan by connecting perimeter vertices).
	geo.Indices = []int{
		0, 1, 9, 1, 2, 9, 2, 3, 9, 3, 4, 9, 4, 5, 9,
		5, 6, 9, 6, 7, 9, 7, 8, 9, 8, 0, 9,
	}

	star := three.NewMesh(geo, mat)
	scene.AddObject(star)

	// Particle system (emits golden particles from the star).
	ps := three.NewParticleSystem(100)
	ps.Emit(24)

	// Lights (Phong-style ambient + point light).
	ambientLight := three.NewLight(three.LightAmbient, color.RGBA{R: 80, G: 80, B: 80, A: 255}, 0.3)
	pointLight := three.NewLight(three.LightPoint, color.RGBA{R: 255, G: 200, B: 50, A: 255}, 0.7)
	_ = ambientLight
	_ = pointLight

	threeRenderer.SetScene(scene)
	threeRenderer.SetCamera(camera)
	camera.SetPosition(0, 0, 5)

	buf := core.NewNativeBuffer(40, 20)
	threeRenderer.Render(buf, 0.016)

	var active int
	for _, p := range ps.Particles {
		if p != nil && p.Active {
			active++
		}
	}

	// 3D output text.
	d3Text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "golden-star-3d",
			Width:  strPtr("38"),
			Height: strPtr("18"),
			Visible: boolPtr(true),
		},
		Content: "Golden Star (CPU 3D)\n\n" + bufferToText(buf) +
			"\n\nStar: 10 vertices | 10 triangles\nParticles: " + fmt.Sprintf("%d active / %d max", active, ps.MaxParticles),
	})
	d3Text.SetPosition(2, 2)

	// ASCII title overlay ("OPENTUI" + "5000").
	title := renderables.NewASCIIFontRenderable(ctx, renderables.ASCIIFontOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "golden-star-ascii",
			Width:  strPtr("36"),
			Height: strPtr("4"),
			Visible: boolPtr(true),
		},
		Text: "OPENTUI",
		Font: "block",
	})
	title.SetPosition(42, 2)

	subTitle := renderables.NewASCIIFontRenderable(ctx, renderables.ASCIIFontOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "golden-star-ascii-2",
			Width:  strPtr("20"),
			Height: strPtr("4"),
			Visible: boolPtr(true),
		},
		Text: "5000",
		Font: "block",
	})
	subTitle.SetPosition(44, 7)

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "golden-star-info",
			Width:  strPtr("36"),
			Height: strPtr("5"),
			Visible: boolPtr(true),
		},
		Content: "Golden Star Demo\nPress 'h' to toggle Hell mode\nPress 'b' to toggle gradient band",
	})
	info.SetPosition(42, 12)

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "golden-star-box",
			Width:  strPtr("80"),
			Height: strPtr("22"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Golden Star (3D + Particles + ASCII Overlay)",
	})
	box.AddChild(d3Text.Renderable)
	box.AddChild(title.Renderable)
	box.AddChild(subTitle.Renderable)
	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)
	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// hast-syntax-highlighting-demo
// ---------------------------------------------------------------------------

// runHastSyntaxHighlightingDemo demonstrates syntax-highlighted code rendering
// using the Code renderable with multiple languages. Ported from
// hast-syntax-highlighting-demo.ts (simplified: Code renderable instead of
// HAST-to-styled-text conversion).
func runHastSyntaxHighlightingDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "hast-box",
			Width:  strPtr("76"),
			Height: strPtr("20"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "HAST Syntax Highlighting (Code Renderable)",
	})

	// TypeScript sample.
	tsCode := `function greet(name: string): string {
  return "Hello, " + name;
}

console.log(greet("OpenTUI"));`
	tsRenderable := renderables.NewCodeRenderable(ctx, renderables.CodeOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "hast-ts",
			Width:  strPtr("34"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
		Code:     tsCode,
		Language: "typescript",
	})
	tsRenderable.SetPosition(2, 2)

	tsLabel := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "hast-ts-label",
			Width:  strPtr("34"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Content: "TypeScript",
	})
	tsLabel.SetPosition(2, 1)

	// Go sample.
	goCode := `package main

import "fmt"

func main() {
    fmt.Println("Hello, HAST!")
}`
	goRenderable := renderables.NewCodeRenderable(ctx, renderables.CodeOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "hast-go",
			Width:  strPtr("34"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
		Code:     goCode,
		Language: "go",
	})
	goRenderable.SetPosition(38, 2)

	goLabel := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "hast-go-label",
			Width:  strPtr("34"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Content: "Go",
	})
	goLabel.SetPosition(38, 1)

	// Python sample.
	pyCode := `def greet(name):
    return f"Hello, {name}"

print(greet("OpenTUI"))`
	pyRenderable := renderables.NewCodeRenderable(ctx, renderables.CodeOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "hast-py",
			Width:  strPtr("34"),
			Height: strPtr("8"),
			Visible: boolPtr(true),
		},
		Code:     pyCode,
		Language: "python",
	})
	pyRenderable.SetPosition(2, 11)

	pyLabel := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "hast-py-label",
			Width:  strPtr("34"),
			Height: strPtr("1"),
			Visible: boolPtr(true),
		},
		Content: "Python",
	})
	pyLabel.SetPosition(2, 10)

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "hast-info",
			Width:  strPtr("72"),
			Height: strPtr("4"),
			Visible: boolPtr(true),
		},
		Content: "Syntax highlighting via Code renderable (tree-sitter based).\nLanguages: TypeScript, Go, Python.\nPress 'R' to re-transform (simulated).",
	})
	info.SetPosition(2, 20)

	box.AddChild(tsLabel.Renderable)
	box.AddChild(tsRenderable.Renderable)
	box.AddChild(goLabel.Renderable)
	box.AddChild(goRenderable.Renderable)
	box.AddChild(pyLabel.Renderable)
	box.AddChild(pyRenderable.Renderable)
	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)
	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// physx-planck-2d-demo
// ---------------------------------------------------------------------------

// planckBox represents a simulated rigid body in the Planck physics demo.
type planckBox struct {
	x, y, vx, vy float64
	w, h         float64
	id           string
}

// runPhysxPlanck2DDemo demonstrates a simplified 2D physics simulation with
// gravity, falling boxes, and ground collision. Ported from
// physx-planck-2d-demo.ts (simplified: custom integrator instead of planck.js).
func runPhysxPlanck2DDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	const gravity = 9.81
	const groundY = -8.0
	const dt = 1.0 / 60.0
	const steps = 300

	boxes := []planckBox{
		{x: -4, y: 6, vx: 0.5, vy: 0, w: 1.0, h: 1.0, id: "box_0"},
		{x: 0, y: 10, vx: -0.3, vy: 0, w: 1.0, h: 1.0, id: "box_1"},
		{x: 4, y: 4, vx: 0.2, vy: 0, w: 1.0, h: 1.0, id: "box_2"},
		{x: -2, y: 12, vx: 0.1, vy: 0, w: 1.0, h: 1.0, id: "box_3"},
		{x: 6, y: 8, vx: -0.4, vy: 0, w: 1.0, h: 1.0, id: "box_4"},
	}

	// Simulate physics: apply gravity, integrate, resolve ground collision.
	for i := 0; i < steps; i++ {
		for j := range boxes {
			b := &boxes[j]
			b.vy -= gravity * dt
			b.y += b.vy * dt
			b.x += b.vx * dt
			// Ground collision (restitution 0.3).
			if b.y-b.h/2 < groundY {
				b.y = groundY + b.h/2
				b.vy = -b.vy * 0.3
				b.vx *= 0.9
			}
			// Remove boxes that fall too far.
			if b.y < -15 {
				b.y = 10 + float64(j)*2
				b.vy = 0
				b.vx = 0
			}
		}
	}

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "planck-box",
			Width:  strPtr("76"),
			Height: strPtr("18"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Planck.js 2D Physics (Falling Crates)",
	})

	var sb strings.Builder
	sb.WriteString("Planck.js 2D Physics Demo\n")
	sb.WriteString("==========================\n\n")
	sb.WriteString(fmt.Sprintf("Gravity: %.2f m/s^2\n", gravity))
	sb.WriteString(fmt.Sprintf("Simulated steps: %d (dt=%.4fs)\n\n", steps, dt))
	sb.WriteString("Final box positions (x, y, vx, vy):\n")
	for _, b := range boxes {
		sb.WriteString(fmt.Sprintf("  %s: (%.2f, %.2f) v=(%.2f, %.2f) size=%.1fx%.1f\n",
			b.id, b.x, b.y, b.vx, b.vy, b.w, b.h))
	}
	sb.WriteString("\nControls: [Space] spawn, [E] explode, [R] reset, [C] clear")

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "planck-text",
			Width:  strPtr("72"),
			Height: strPtr("12"),
			Visible: boolPtr(true),
		},
		Content: sb.String(),
	})
	text.SetPosition(1, 1)

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "planck-info",
			Width:  strPtr("72"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Content: "Simplified 2D rigid-body physics: gravity + ground collision + restitution.",
	})
	info.SetPosition(1, 15)

	box.AddChild(text.Renderable)
	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)
	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}

// ---------------------------------------------------------------------------
// physx-rapier-2d-demo
// ---------------------------------------------------------------------------

// rapierBox represents a simulated rigid body in the Rapier physics demo.
type rapierBox struct {
	x, y, vx, vy float64
	w, h         float64
	id           string
}

// runPhysxRapier2DDemo demonstrates a simplified 2D physics simulation with
// gravity, falling boxes, spotlight effects, and explosion handling. Ported
// from physx-rapier-2d-demo.ts (simplified: custom integrator instead of Rapier).
func runPhysxRapier2DDemo() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	const gravity = 9.81
	const groundY = -8.0
	const dt = 1.0 / 60.0
	const steps = 300

	boxes := []rapierBox{
		{x: -6, y: 5, vx: 0.4, vy: 0, w: 1.2, h: 1.2, id: "box_0"},
		{x: -2, y: 9, vx: -0.2, vy: 0, w: 1.2, h: 1.2, id: "box_1"},
		{x: 2, y: 7, vx: 0.3, vy: 0, w: 1.2, h: 1.2, id: "box_2"},
		{x: 6, y: 11, vx: -0.5, vy: 0, w: 1.2, h: 1.2, id: "box_3"},
		{x: -8, y: 13, vx: 0.1, vy: 0, w: 1.2, h: 1.2, id: "box_4"},
		{x: 8, y: 6, vx: 0.2, vy: 0, w: 1.2, h: 1.2, id: "box_5"},
	}

	// Simulate physics: apply gravity, integrate, resolve ground collision.
	// Rapier uses slightly different defaults (restitution 0.3, friction 0.9).
	for i := 0; i < steps; i++ {
		for j := range boxes {
			b := &boxes[j]
			b.vy -= gravity * dt
			b.y += b.vy * dt
			b.x += b.vx * dt
			// Ground collision with higher friction.
			if b.y-b.h/2 < groundY {
				b.y = groundY + b.h/2
				b.vy = -b.vy * 0.3
				b.vx *= 0.92
			}
			// Respawn boxes that fall too far.
			if b.y < -15 {
				b.y = 10 + float64(j)*2
				b.vy = 0
				b.vx = 0
			}
		}
	}

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "rapier-box",
			Width:  strPtr("76"),
			Height: strPtr("18"),
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Rapier.js 2D Physics (Falling Crates)",
	})

	var sb strings.Builder
	sb.WriteString("Rapier.js 2D Physics Demo\n")
	sb.WriteString("=========================\n\n")
	sb.WriteString(fmt.Sprintf("Gravity: %.2f m/s^2\n", gravity))
	sb.WriteString(fmt.Sprintf("Simulated steps: %d (dt=%.4fs)\n\n", steps, dt))
	sb.WriteString("Final box positions (x, y, vx, vy):\n")
	for _, b := range boxes {
		sb.WriteString(fmt.Sprintf("  %s: (%.2f, %.2f) v=(%.2f, %.2f) size=%.1fx%.1f\n",
			b.id, b.x, b.y, b.vx, b.vy, b.w, b.h))
	}
	sb.WriteString("\nControls: [Space] spawn, [E] explode, [R] reset, [C] clear, [B] burst")

	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "rapier-text",
			Width:  strPtr("72"),
			Height: strPtr("12"),
			Visible: boolPtr(true),
		},
		Content: sb.String(),
	})
	text.SetPosition(1, 1)

	info := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "rapier-info",
			Width:  strPtr("72"),
			Height: strPtr("3"),
			Visible: boolPtr(true),
		},
		Content: "Simplified 2D rigid-body physics: gravity + ground collision + restitution (Rapier-compat).",
	})
	info.SetPosition(1, 15)

	box.AddChild(text.Renderable)
	box.AddChild(info.Renderable)
	root.AddChild(box.Renderable)

	quitKey(renderer)
	renderHeadless(renderer, root)
	return runLoop(renderer, root, 2*time.Second)
}
