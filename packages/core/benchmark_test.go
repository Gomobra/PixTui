package core_test

import (
	"strings"
	"testing"

	"github.com/Gomobra/PixTui/packages/core"
	"github.com/Gomobra/PixTui/packages/core/renderables"
	"github.com/Gomobra/PixTui/packages/native"
)

func benchBoolPtr(b bool) *bool { return &b }
func benchIntPtr(i int) *int    { return &i }

func benchNewCtx() *core.RenderContext {
	return &core.RenderContext{
		Width:  80,
		Height: 24,
	}
}

// BenchmarkRenderableRender renders Box + Text + Input renderables
// repeatedly into a shared NativeBuffer, mirroring the original TypeScript
// benchmark suite.
func BenchmarkRenderableRender(b *testing.B) {
	ctx := benchNewCtx()
	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "bench-box",
			Visible: benchBoolPtr(true),
		},
		Border: true,
		Title:  "Bench",
	})
	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "bench-text",
			Visible: benchBoolPtr(true),
		},
		Content: "Hello World",
	})
	input := renderables.NewInputRenderable(ctx, renderables.InputOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{ID: "bench-input"},
		Value:             "benchmark input",
	})

	buf := core.NewNativeBuffer(80, 24)

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(80 * 24 * 12)) // 12 bytes per cell (char+fg+bg+attrs)
	for i := 0; i < b.N; i++ {
		buf.Clear()
		box.RenderSelf(buf, 0)
		text.RenderSelf(buf, 0)
		input.RenderSelf(buf, 0)
	}
}

// BenchmarkRendererRender performs a full render pass through CliRenderer
// over a tree containing Box, Text, and Input renderables.
func BenchmarkRendererRender(b *testing.B) {
	renderer := core.NewCliRenderer(80, 24)
	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "root")

	box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "bench-box",
			Visible: benchBoolPtr(true),
		},
		Border: true,
		Title:  "Bench",
	})
	text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "bench-text",
			Visible: benchBoolPtr(true),
		},
		Content: "Hello World",
	})
	input := renderables.NewInputRenderable(ctx, renderables.InputOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{ID: "bench-input"},
		Value:             "benchmark input",
	})

	root.AddChild(box.Renderable)
	root.AddChild(text.Renderable)
	root.AddChild(input.Renderable)

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(80 * 24 * 12))
	for i := 0; i < b.N; i++ {
		if err := renderer.Render(root, 0.016); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkTextRender renders a large multi-line text block into a buffer.
func BenchmarkTextRender(b *testing.B) {
	ctx := benchNewCtx()
	var content string
	for i := 0; i < 200; i++ {
		content += "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.\n"
	}
	r := renderables.NewTextRenderable(ctx, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "bench-text-large",
			Visible: benchBoolPtr(true),
		},
		Content: content,
		Wrap:    true,
	})
	buf := core.NewNativeBuffer(80, 24)

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(len(content)))
	for i := 0; i < b.N; i++ {
		buf.Clear()
		r.RenderSelf(buf, 0)
	}
}

// BenchmarkBoxRender renders a Box with border and title into a buffer.
func BenchmarkBoxRender(b *testing.B) {
	ctx := benchNewCtx()
	r := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:     "bench-box",
			Visible: benchBoolPtr(true),
		},
		Border: true,
		Title:  "Benchmark Box",
	})
	buf := core.NewNativeBuffer(80, 24)

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(80 * 24 * 12))
	for i := 0; i < b.N; i++ {
		buf.Clear()
		r.RenderSelf(buf, 0)
	}
}

// ============================================================================
// Comprehensive benchmarks (added)
// ============================================================================

// BenchmarkRenderFullTree renders a full tree with many renderables
// (Box containers, Text leaves, Input fields) through CliRenderer, mirroring
// the TypeScript render-tree benchmark scenarios.
func BenchmarkRenderFullTree(b *testing.B) {
	renderer := core.NewCliRenderer(120, 40)
	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "bench-root")

	const colCount = 8
	const rowPerCol = 6
	for c := 0; c < colCount; c++ {
		col := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
			RenderableOptions: core.RenderableOptions[core.Renderable]{
				ID:     "bench-col-" + itoa(c),
				Visible: benchBoolPtr(true),
			},
			Border:     true,
			Title:      "Column " + itoa(c),
			ShouldFill: true,
		})
		for r := 0; r < rowPerCol; r++ {
			row := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
				RenderableOptions: core.RenderableOptions[core.Renderable]{
					ID: "bench-row-" + itoa(c) + "-" + itoa(r),
				},
				Border: false,
			})
			text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
				RenderableOptions: core.RenderableOptions[core.Renderable]{
					ID: "bench-text-" + itoa(c) + "-" + itoa(r),
				},
				Content: "Row content line " + itoa(r) + " in column " + itoa(c),
			})
			inp := renderables.NewInputRenderable(ctx, renderables.InputOptions{
				RenderableOptions: core.RenderableOptions[core.Renderable]{
					ID: "bench-input-" + itoa(c) + "-" + itoa(r),
				},
				Value: "input value " + itoa(c) + "-" + itoa(r),
			})
			row.AddChild(text.Renderable)
			row.AddChild(inp.Renderable)
			col.AddChild(row.Renderable)
		}
		root.AddChild(col.Renderable)
	}

	buf := core.NewNativeBuffer(120, 40)

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(120 * 40 * 12))
	for i := 0; i < b.N; i++ {
		buf.Clear()
		if err := renderer.Render(root, 0.016); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRenderableCreateDestroy repeatedly creates and destroys renderables,
// exercising the Renderable lifecycle including native node and Yoga node cleanup.
func BenchmarkRenderableCreateDestroy(b *testing.B) {
	ctx := benchNewCtx()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
			RenderableOptions: core.RenderableOptions[core.Renderable]{
				ID: "bench-temp-" + itoa(i),
			},
			Border: true,
			Title:  "Temp",
		})
		for j := 0; j < 3; j++ {
			child := renderables.NewTextRenderable(ctx, renderables.TextOptions{
				RenderableOptions: core.RenderableOptions[core.Renderable]{
					ID: "bench-temp-child-" + itoa(i) + "-" + itoa(j),
				},
				Content: "child " + itoa(j),
			})
			r.AddChild(child.Renderable)
		}
		r.Destroy()
	}
}



// BenchmarkYogaLayoutNested performs Yoga layout on a deeply nested tree
// of layout nodes, measuring layout recalculation cost for deep hierarchies.
func BenchmarkYogaLayoutNested(b *testing.B) {
	yoga := native.NewYoga()
	root := yoga.CreateNode()
	root.SetWidth(120)
	root.SetHeight(40)

	// Build a deep nested tree: depth 10, each node has 3 children.
	const depth = 10
	const fanout = 3
	var build func(node *native.YogaNode, level int)
	build = func(node *native.YogaNode, level int) {
		if level >= depth {
			return
		}
		for i := 0; i < fanout; i++ {
			child := yoga.CreateNode()
			child.SetWidth(30)
			child.SetHeight(8)
			node.AddChild(child)
			build(child, level+1)
		}
	}
	build(root, 0)

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(120 * 40 * 4))
	for i := 0; i < b.N; i++ {
		yoga.CalculateLayout(root, 120, 40)
	}
}

// BenchmarkYogaLayoutWide performs Yoga layout on a very wide shallow tree
// (many siblings), measuring layout cost for wide hierarchies.
func BenchmarkYogaLayoutWide(b *testing.B) {
	yoga := native.NewYoga()
	root := yoga.CreateNode()
	root.SetWidth(200)
	root.SetHeight(24)

	const childCount = 200
	for i := 0; i < childCount; i++ {
		child := yoga.CreateNode()
		child.SetWidth(8)
		child.SetHeight(1)
		root.AddChild(child)
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(200 * 24 * 4))
	for i := 0; i < b.N; i++ {
		yoga.CalculateLayout(root, 200, 24)
	}
}

// BenchmarkTextBufferLargeOps performs insert/delete operations on a large
// native TextBuffer, simulating editing a large document.
func BenchmarkTextBufferLargeOps(b *testing.B) {
	// Build a large document (~100KB)
	var sb strings.Builder
	for i := 0; i < 2000; i++ {
		sb.WriteString("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.\n")
	}
	content := sb.String()

	tb := native.NewTextBufferWithContent(content)

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(len(content)))
	for i := 0; i < b.N; i++ {
		length := tb.Length()
		if length == 0 { length = 1 }
		offset := i % length
		insertText := " [edit-" + itoa(i) + "] "
		if err := tb.Insert(offset, insertText); err != nil {
			b.Fatal(err)
		}
		// Delete the inserted text to keep buffer size stable
		if err := tb.Delete(offset, offset+len(insertText)); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEditBufferUndoRedo repeatedly performs insert/undo/redo cycles
// on a core EditBuffer, testing the undo stack.
func BenchmarkEditBufferUndoRedo(b *testing.B) {
	eb := core.NewEditBuffer()
	eb.SetContent("Initial content for undo/redo benchmarking. ")

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(eb.Length()))
	for i := 0; i < b.N; i++ {
		text := "undo-redo-step-" + itoa(i) + " "
		if err := eb.Insert(0, text); err != nil {
			b.Fatal(err)
		}
		if !eb.Undo() {
			b.Fatal("undo failed")
		}
		if !eb.Redo() {
			b.Fatal("redo failed")
		}
		// Periodically reset to avoid unbounded growth
		if i%100 == 99 {
			eb.SetContent("Initial content for undo/redo benchmarking. ")
		}
	}
}

// BenchmarkImageRenderBlocks renders an image using the blocks protocol
// into a NativeBuffer, measuring block-character rendering cost.
func BenchmarkImageRenderBlocks(b *testing.B) {
	ctx := benchNewCtx()
	img := native.NewImage(80, 24)
	r := renderables.NewImageRenderable(ctx, renderables.ImageOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID: "bench-image-blocks",
		},
		Image:     img,
		Protocol:  core.ImageRenderProtocolBlocks,
	})
	buf := core.NewNativeBuffer(80, 24)

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(80 * 24 * 12))
	for i := 0; i < b.N; i++ {
		buf.Clear()
		r.RenderSelf(buf, 0)
	}
}

// BenchmarkMarkdownParse parses a large markdown document using
// MarkdownRenderable, measuring markdown parse + render cost.
func BenchmarkMarkdownParse(b *testing.B) {
	// Build a large markdown document with headings, lists, tables, code blocks
	var sb strings.Builder
	sb.WriteString("# Large Markdown Benchmark Document\n\n")
	for s := 0; s < 50; s++ {
		sb.WriteString("## Section " + itoa(s) + "\n\n")
		for p := 0; p < 5; p++ {
			sb.WriteString("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam.\n\n")
		}
		sb.WriteString("- Item 1\n- Item 2\n- Item 3\n\n")
		sb.WriteString("| Col A | Col B | Col C |\n|-------|-------|-------|\n| val1  | val2  | val3  |\n| val4  | val5  | val6  |\n\n")
		sb.WriteString("```go\npackage main\n\nfunc main() {\n    println(\"hello world\")\n}\n```\n\n")
	}
	mdContent := sb.String()

	ctx := benchNewCtx()
	r := renderables.NewMarkdownRenderable(ctx, renderables.MarkdownOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID: "bench-markdown",
		},
		Content: mdContent,
		Conceal: true,
	})
	buf := core.NewNativeBuffer(80, 24)

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(len(mdContent)))
	for i := 0; i < b.N; i++ {
		buf.Clear()
		r.RenderSelf(buf, 0)
	}
}

// BenchmarkDiffRender renders diff content with +/- coloring using
// DiffRenderable.
func BenchmarkDiffRender(b *testing.B) {
	// Build a large diff with many additions and deletions
	var sb strings.Builder
	for i := 0; i < 500; i++ {
		sb.WriteString("- removed line " + itoa(i) + "\n")
		sb.WriteString("+ added line " + itoa(i) + "\n")
		sb.WriteString("  unchanged line " + itoa(i) + "\n")
	}
	diffContent := sb.String()

	ctx := benchNewCtx()
	r := renderables.NewDiffRenderable(ctx, renderables.DiffOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID: "bench-diff",
		},
		Content: diffContent,
	})
	buf := core.NewNativeBuffer(80, 24)

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(len(diffContent)))
	for i := 0; i < b.N; i++ {
		buf.Clear()
		r.RenderSelf(buf, 0)
	}
}

// BenchmarkScrollBoxRender renders a ScrollBox with many children,
// simulating a scrollable list with many items.
func BenchmarkScrollBoxRender(b *testing.B) {
	renderer := core.NewCliRenderer(80, 40)
	ctx := renderer.Context()
	root := core.NewRootRenderable(ctx, "bench-scroll-root")

	scrollBox := renderables.NewScrollBoxRenderable(ctx, renderables.ScrollBoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID: "bench-scrollbox",
		},
	})

	const itemCount = 200
	for i := 0; i < itemCount; i++ {
		item := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
			RenderableOptions: core.RenderableOptions[core.Renderable]{
				ID: "bench-scroll-item-" + itoa(i),
			},
			Border: false,
		})
		text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
			RenderableOptions: core.RenderableOptions[core.Renderable]{
				ID: "bench-scroll-text-" + itoa(i),
			},
			Content: "Scroll item " + itoa(i) + " with some sample content",
		})
		item.AddChild(text.Renderable)
		scrollBox.AddChild(item.Renderable)
	}
	root.AddChild(scrollBox.Renderable)

	buf := core.NewNativeBuffer(80, 40)

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(80 * 40 * 12))
	for i := 0; i < b.N; i++ {
		buf.Clear()
		if err := renderer.Render(root, 0.016); err != nil {
			b.Fatal(err)
		}
	}
}

// Helper: itoa without fmt import overhead in benchmarks.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var buf [16]byte
	pos := len(buf)
	neg := false
	if i < 0 {
		neg = true
		i = -i
	}
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
