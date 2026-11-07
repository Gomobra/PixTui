# PixTui React

[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)](https://go.dev/)

React-style bindings for PixTui.

## About

PixTui React provides React-style bindings for building terminal user interfaces. It includes a component reconciler, hooks, context, and JSX runtime support.

## Features

- **Component reconciler** — Fiber-based reconciliation with keyed children support
- **Hooks** — `useState`, `useEffect`, `useLayoutEffect`, `useMemo`, `useCallback`, `useRef`, `useContext`, `useEffectEvent`
- **JSX runtime** — `jsx`, `jsxs`, `jsxDEV` for element creation
- **Context** — `CreateContext`, `Provider`, `useContext` for dependency injection
- **DevTools** — Fiber tree inspection and debugging utilities
- **Test utilities** — `TestRender`, `AssertNoPanic` for testing components

## Installation

```bash
go get github.com/Gomobra/PixTui/packages/react
```

## Quick Start

```go
package main

import (
	"github.com/Gomobra/PixTui/packages/core"
	"github.com/Gomobra/PixTui/packages/react"
)

func main() {
	renderer := core.NewCliRenderer(80, 24)
	reactRenderer := react.NewReactRenderer(renderer.Context())

	element := react.JSX("box", map[string]interface{}{
		"border": true,
		"title":  "Hello",
	})

	reactRenderer.Render(element, renderer.Context())
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

## Hooks

| Hook | Description |
|---|---|
| `UseState` | State management with functional updater |
| `UseEffect` | Side effects with dependency tracking |
| `UseLayoutEffect` | Synchronous side effects |
| `UseMemo` | Memoized values |
| `UseCallback` | Memoized callbacks |
| `UseRef` | Mutable ref object |
| `UseContext` | Context consumption |

## License

Distributed under the MIT License. See `LICENSE` for more information.

## Contact

Gomobra — [@Gomobra](https://github.com/Gomobra)

Project Link: [https://github.com/Gomobra/PixTui](https://github.com/Gomobra/PixTui)