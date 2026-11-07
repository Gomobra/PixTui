# PixTui Solid

[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)](https://go.dev/)

Solid-style bindings for PixTui.

## About

PixTui Solid provides Solid-style bindings for building terminal user interfaces. It includes a reactive system with signals, effects, memos, and context, along with a component reconciler.

## Features

- **Reactive primitives** — `CreateSignal`, `CreateEffect`, `CreateMemo`
- **Component model** — `CreateComponent` for declarative UI
- **Context** — `CreateContext`, `Provider`, `useContext` for dependency injection
- **Portal** — Render children to a different location in the tree
- **Dynamic** — Render component based on runtime value
- **Plugin slot** — Extensible slot system for plugins
- **Scrollback writer** — Write snapshots to scrollback buffer
- **Test utilities** — `TestRender`, `CaptureCharFrame`, `AssertNoPanic`

## Installation

```bash
go get github.com/Gomobra/PixTui/packages/solid
```

## Quick Start

```go
package main

import (
	"github.com/Gomobra/PixTui/packages/core"
	"github.com/Gomobra/PixTui/packages/solid"
)

func main() {
	renderer := core.NewCliRenderer(80, 24)
	solidRenderer := solid.NewSolidRenderer(renderer.Context())

	component := solid.Box(map[string]interface{}{
		"border": true,
		"title":  "Hello",
	})

	solidRenderer.Render(component, renderer.Context())
}
```

## Components

| Component | Purpose |
|---|---|
| `Box` | Container with optional border and title |
| `Text` | Text content with wrapping and alignment |
| `Input` | Single-line text input |
| `Select` | Selectable options |
| `ScrollBox` | Scrollable content container |
| `Textarea` | Multi-line text editing |
| `Slider` | Numeric value selection |
| `TabSelect` | Tab-style selection |
| `EditBuffer` | Editable text buffer view |
| `TextTable` | Tabular text display |
| `Code` | Code display |
| `Markdown` | Markdown content display |
| `Diff` | Diff-style text display |
| `Image` / `Video` | Media renderables |
| `ASCIIFont` | Text rendered with ASCII-style fonts |
| `FrameBuffer` | Framebuffer-backed content |
| `LineNumber` | Line-number display |
| `ScrollBar` | Scroll position indicator |
| `TextNode` | Text node |

## Reactive Primitives

| Primitive | Description |
|---|---|
| `CreateSignal` | Create reactive signal |
| `CreateEffect` | Run effect on signal change |
| `CreateMemo` | Memoized computed value |
| `CreateComponent` | Create component from render function |
| `CreateContext` | Create context for dependency injection |

## License

Distributed under the MIT License. See `LICENSE` for more information.

## Contact

Gomobra — [@Gomobra](https://github.com/Gomobra)

Project Link: [https://github.com/Gomobra/PixTui](https://github.com/Gomobra/PixTui)