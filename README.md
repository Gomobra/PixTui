# PixTui

Build beautiful terminal user interfaces in Go.

PixTui is a Go library for creating terminal applications with rich layouts, interactive components, and reactive programming.

## Features

- **22 renderable components** — Box, Text, Input, Select, ScrollBox, Textarea, Slider, Markdown, Code, Video, Image, and more
- **Yoga flexbox layout** — Responsive terminal layouts
- **React & Solid bindings** — Reactive UI with signals, effects, and context
- **SSH server** — Password, public key, and keyboard-interactive auth
- **Three.js renderer** — CPU-based 3D rendering in terminal
- **QR code encoder** — Full QR code generation
- **Keymap system** — Multi-layer key bindings with Emacs/Vim support
- **Animation & Timeline** — Tweening and animation engine
- **Plugin system** — Extensible plugin architecture
- **Web server** — Built-in documentation server

## Installation

```bash
go get github.com/Gomobra/PixTui
```

## Quick Start

```go
package main

import (
    "github.com/Gomobra/PixTui/packages/core"
    "github.com/Gomobra/PixTui/packages/core/renderables"
)

func main() {
    renderer := core.NewCliRenderer(80, 24)
    root := core.NewRootRenderable(renderer.Context())

    box := renderables.NewBoxRenderable(renderer.Context(), renderables.BoxOptions{
        Content: "Hello, PixTui!",
    })

    root.AddChild(box)
    renderer.Render(root)
}
```

## Examples

```bash
go run ./cmd/pixtui-demo/ hello
go run ./cmd/pixtui-demo/ box
go run ./cmd/pixtui-demo/ video
go run ./cmd/pixtui-demo/ 3d
go run ./cmd/pixtui-demo/ ssh
```

## Packages

| Package | Description |
|---------|-------------|
| `native` | Terminal, buffer, image, video, audio, tree-sitter |
| `core` | Renderable system, layout, edit buffer, editor view |
| `react` | React-style hooks and JSX |
| `solid` | Solid-style signals and components |
| `ssh` | SSH server |
| `three` | 3D renderer |
| `web` | HTTP documentation server |
| `keymap` | Key binding engine |
| `animation` | Animation and timeline |
| `plugin` | Plugin system |
| `qrcode` | QR code encoder |
| `examples` | 122 example applications |

## License

MIT