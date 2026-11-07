# Getting Started with PixTui

## Installation

### Prerequisites

- Go 1.26 or higher
- A terminal environment (Linux, macOS, or Windows)

### Install

```bash
go get github.com/Gomobra/PixTui
```

### From Source

```bash
git clone https://github.com/Gomobra/PixTui.git
cd PixTui
go build ./...
```

## Your First Application

Create a simple terminal application with a box:

```go
package main

import (
	"log"

	"github.com/Gomobra/PixTui/packages/core"
	"github.com/Gomobra/PixTui/packages/core/renderables"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	renderContext := renderer.Context()
	root := core.NewRootRenderable(renderContext, "root")

	box := renderables.NewBoxRenderable(renderContext, renderables.BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      "welcome-box",
			Width:   stringPtr("60"),
			Height:  stringPtr("7"),
			Visible: boolPtr(true),
		},
		Border:      true,
		BorderStyle: "rounded",
		Title:       "PixTui",
	})

	message := renderables.NewTextRenderable(renderContext, renderables.TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      "welcome-message",
			Width:   stringPtr("56"),
			Height:  stringPtr("3"),
			Visible: boolPtr(true),
		},
		Content: "Hello from PixTui!",
		Wrap:    true,
	})

	box.AddChild(message.Renderable)
	root.AddChild(box.Renderable)

	if err := renderer.Start(); err != nil {
		return err
	}
	if err := renderer.Render(root, 1.0/60.0); err != nil {
		return err
	}

	return renderer.WriteFrame()
}

func stringPtr(value string) *string { return &value }
func boolPtr(value bool) *bool       { return &value }
```

## Running Examples

PixTui includes 122 example applications. Run them with:

```bash
go run ./cmd/pixtui-demo/ hello
go run ./cmd/pixtui-demo/ box
go run ./cmd/pixtui-demo/ text
go run ./cmd/pixtui-demo/ 3d
```

## Building

```bash
go build ./...
go test ./...
go vet ./...
```

## Next Steps

- [Components](components.md) — Learn about all available components
- [API Reference](api-reference.md) — Detailed API documentation
- [Examples](examples.md) — Browse example applications
- [Contributing](contributing.md) — How to contribute