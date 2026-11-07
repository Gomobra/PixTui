# Components

PixTui provides 22 renderable components for building terminal user interfaces.

## Core Components

### Box

Container component with optional border and title.

```go
box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
    RenderableOptions: core.RenderableOptions[core.Renderable]{
        ID:      "my-box",
        Width:   stringPtr("60"),
        Height:  stringPtr("7"),
        Visible: boolPtr(true),
    },
    Border:      true,
    BorderStyle: "rounded",
    Title:       "My Box",
    Content:     "Hello!",
})
```

### Text

Text content with wrapping, alignment, and truncation options.

```go
text := renderables.NewTextRenderable(ctx, renderables.TextOptions{
    RenderableOptions: core.RenderableOptions[core.Renderable]{
        ID:      "my-text",
        Width:   stringPtr("40"),
        Height:  stringPtr("3"),
        Visible: boolPtr(true),
    },
    Content: "Hello, World!",
    Wrap:    true,
})
```

### Input

Single-line text input with placeholder and validation.

```go
input := renderables.NewInputRenderable(ctx, renderables.InputOptions{
    RenderableOptions: core.RenderableOptions[core.Renderable]{
        ID:      "my-input",
        Width:   stringPtr("30"),
        Height:  stringPtr("1"),
        Visible: boolPtr(true),
    },
    Value:       "",
    Placeholder: "Enter text...",
})
```

### Select

Selectable options dropdown.

```go
select := renderables.NewSelectRenderable(ctx, renderables.SelectOptions{
    RenderableOptions: core.RenderableOptions[core.Renderable]{
        ID:      "my-select",
        Width:   stringPtr("20"),
        Height:  stringPtr("5"),
        Visible: boolPtr(true),
    },
    Options:  []string{"Option 1", "Option 2", "Option 3"},
    Selected: 0,
})
```

### ScrollBox

Scrollable content container with optional scrollbar.

```go
scrollbox := renderables.NewScrollBoxRenderable(ctx, renderables.ScrollBoxOptions{
    RenderableOptions: core.RenderableOptions[core.Renderable]{
        ID:      "my-scrollbox",
        Width:   stringPtr("40"),
        Height:  stringPtr("10"),
        Visible: boolPtr(true),
    },
})
```

### Textarea

Multi-line text editing area.

```go
textarea := renderables.NewTextareaRenderable(ctx, renderables.TextareaOptions{
    RenderableOptions: core.RenderableOptions[core.Renderable]{
        ID:      "my-textarea",
        Width:   stringPtr("40"),
        Height:  stringPtr("5"),
        Visible: boolPtr(true),
    },
    Value: "Initial text...",
})
```

### Slider

Numeric value selection with min/max range.

```go
slider := renderables.NewSliderRenderable(ctx, renderables.SliderOptions{
    RenderableOptions: core.RenderableOptions[core.Renderable]{
        ID:      "my-slider",
        Width:   stringPtr("20"),
        Height:  stringPtr("1"),
        Visible: boolPtr(true),
    },
    Min:   0,
    Max:   100,
    Value: 50,
})
```

### TabSelect

Tab-style selection component.

```go
tabselect := renderables.NewTabSelectRenderable(ctx, renderables.TabSelectOptions{
    RenderableOptions: core.RenderableOptions[core.Renderable]{
        ID:      "my-tabs",
        Width:   stringPtr("40"),
        Height:  stringPtr("3"),
        Visible: boolPtr(true),
    },
    Tabs:     []string{"Tab 1", "Tab 2", "Tab 3"},
    Selected: 0,
})
```

### Code

Syntax-highlighted code display.

```go
code := renderables.NewCodeRenderable(ctx, renderables.CodeOptions{
    RenderableOptions: core.RenderableOptions[core.Renderable]{
        ID:      "my-code",
        Width:   stringPtr("60"),
        Height:  stringPtr("10"),
        Visible: boolPtr(true),
    },
    Code:     "package main\n\nfunc main() {\n\tprintln(\"Hello\")\n}",
    Language: "go",
})
```

### Markdown

Markdown content rendering.

```go
markdown := renderables.NewMarkdownRenderable(ctx, renderables.MarkdownOptions{
    RenderableOptions: core.RenderableOptions[core.Renderable]{
        ID:      "my-markdown",
        Width:   stringPtr("60"),
        Height:  stringPtr("10"),
        Visible: boolPtr(true),
    },
    Content: "# Hello World\n\nThis is **bold** text.",
})
```

### Diff

Diff-style text display.

```go
diff := renderables.NewDiffRenderable(ctx, renderables.DiffOptions{
    RenderableOptions: core.RenderableOptions[core.Renderable]{
        ID:      "my-diff",
        Width:   stringPtr("60"),
        Height:  stringPtr("10"),
        Visible: boolPtr(true),
    },
    Content: "--- a/file.go\n+++ b/file.go\n@@ -1,3 +1,3 @@\n-old line\n+new line",
})
```

### Image

Image rendering with protocol support.

```go
image := renderables.NewImageRenderable(ctx, renderables.ImageOptions{
    RenderableOptions: core.RenderableOptions[core.Renderable]{
        ID:      "my-image",
        Width:   stringPtr("40"),
        Height:  stringPtr("20"),
        Visible: boolPtr(true),
    },
    Path:     "/path/to/image.png",
    Protocol: core.ImageRenderProtocolAuto,
})
```

### Video

Video rendering in terminal.

```go
video := renderables.NewVideoRenderable(ctx, renderables.VideoOptions{
    RenderableOptions: core.RenderableOptions[core.Renderable]{
        ID:      "my-video",
        Width:   stringPtr("80"),
        Height:  stringPtr("24"),
        Visible: boolPtr(true),
    },
    AutoPlay: true,
})
```

### TextTable

Tabular text display.

```go
table := renderables.NewTextTableRenderable(ctx, renderables.TextTableOptions{
    RenderableOptions: core.RenderableOptions[core.Renderable]{
        ID:      "my-table",
        Width:   stringPtr("60"),
        Height:  stringPtr("10"),
        Visible: boolPtr(true),
    },
    Headers: []string{"Name", "Age", "City"},
    Rows: [][]string{
        {"Alice", "30", "NYC"},
        {"Bob", "25", "LA"},
    },
})
```

### ASCIIFont

Text rendered with ASCII-style fonts.

```go
asciiFont := renderables.NewASCIIFontRenderable(ctx, renderables.ASCIIFontOptions{
    RenderableOptions: core.RenderableOptions[core.Renderable]{
        ID:      "my-ascii-font",
        Width:   stringPtr("60"),
        Height:  stringPtr("10"),
        Visible: boolPtr(true),
    },
    Text: "HELLO",
    Font: "block",
})
```

### FrameBuffer

Framebuffer-backed content.

```go
framebuffer := renderables.NewFrameBufferRenderable(ctx, renderables.FrameBufferOptions{
    RenderableOptions: core.RenderableOptions[core.Renderable]{
        ID:      "my-framebuffer",
        Width:   stringPtr("40"),
        Height:  stringPtr("20"),
        Visible: boolPtr(true),
    },
})
```

### LineNumber

Line-number display.

```go
lineNumber := renderables.NewLineNumberRenderable(ctx, "my-line-numbers", 1)
```

### ScrollBar

Scroll position indicator.

```go
scrollbar := renderables.NewScrollBarRenderable(ctx, renderables.ScrollBarOptions{
    RenderableOptions: core.RenderableOptions[core.Renderable]{
        ID:      "my-scrollbar",
        Width:   stringPtr("1"),
        Height:  stringPtr("10"),
        Visible: boolPtr(true),
    },
    Vertical:   true,
    ThumbSize:  3,
    ThumbStart: 0,
})
```

### TextNode

Text node renderable.

```go
textNode := renderables.NewTextNodeRenderable(ctx, renderables.TextNodeOptions{
    RenderableOptions: core.RenderableOptions[core.Renderable]{
        ID:      "my-text-node",
        Width:   stringPtr("40"),
        Height:  stringPtr("3"),
        Visible: boolPtr(true),
    },
    Content: "Hello from TextNode!",
})
```

### TextBufferRenderable

Text buffer renderable with scroll support.

```go
textBuffer := renderables.NewTextBufferRenderable(ctx, renderables.TextBufferRenderableOptions{
    RenderableOptions: core.RenderableOptions[core.Renderable]{
        ID:      "my-text-buffer",
        Width:   stringPtr("40"),
        Height:  stringPtr("10"),
        Visible: boolPtr(true),
    },
})
```

### EmbeddedTerminal

Embedded terminal renderable.

```go
embeddedTerminal := renderables.NewEmbeddedTerminalRenderable(ctx, renderables.EmbeddedTerminalOptions{
    RenderableOptions: core.RenderableOptions[core.Renderable]{
        ID:      "my-embedded-terminal",
        Width:   stringPtr("80"),
        Height:  stringPtr("24"),
        Visible: boolPtr(true),
    },
})
```

### EditBufferRenderable

Editable text buffer with undo/redo.

```go
editBuffer := renderables.NewEditBufferRenderable(ctx, renderables.EditBufferRenderableOptions{
    RenderableOptions: core.RenderableOptions[core.Renderable]{
        ID:      "my-edit-buffer",
        Width:   stringPtr("40"),
        Height:  stringPtr("10"),
        Visible: boolPtr(true),
    },
    Value: "Initial text...",
})
```

## Adding Children

All renderables support adding child components:

```go
box.AddChild(text.Renderable)
box.AddChild(input.Renderable)
root.AddChild(box.Renderable)
```

## Styling

Components support various styling options:

```go
box := renderables.NewBoxRenderable(ctx, renderables.BoxOptions{
    RenderableOptions: core.RenderableOptions[core.Renderable]{
        Width:  stringPtr("40"),
        Height: stringPtr("10"),
    },
    Border:           true,
    BorderStyle:      "rounded",
    Title:            "Styled Box",
    BackgroundColor:  &core.RGBA{R: 30, G: 30, B: 40, A: 255},
    BorderColor:      &core.RGBA{R: 100, G: 100, B: 255, A: 255},
})
```

## Layout

PixTui uses Yoga for flexbox layout. Components support:

- Width/Height (fixed, percentage, auto)
- Flex direction (row, column)
- Justify content (start, center, end, space-between, etc.)
- Align items (start, center, end, stretch)
- Padding and margin
- Gap between children