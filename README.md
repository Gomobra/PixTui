# PixTui

[![Go Version](https://img.shields.io/badge/go-1.26+-blue.svg)](https://golang.org/)
[![MIT License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)]()
[![Tests](https://img.shields.io/badge/tests-passing-brightgreen.svg)]()

A Go library for building beautiful terminal user interfaces (TUI) with rich layouts, interactive components, and reactive programming.

---

## Table of Contents

- [About PixTui](#about-pixtui)
- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Examples](#examples)
- [Packages](#packages)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [License](#license)
- [Contact](#contact)
- [Acknowledgments](#acknowledgments)

---

## About PixTui

PixTui is a Go library for creating terminal applications with rich layouts, interactive components, and reactive programming. Inspired by modern UI frameworks, PixTui brings a component-based approach to terminal UI development.

Whether you're building a CLI dashboard, a text editor, or a full-screen terminal application, PixTui provides the building blocks you need.

---

## Features

- **22 renderable components** — Box, Text, Input, Select, ScrollBox, Textarea, Slider, Markdown, Code, Video, Image, and more
- **Yoga flexbox layout** — Responsive terminal layouts with flexbox
- **React & Solid bindings** — Reactive UI with signals, effects, and context
- **SSH server** — Password, public key, and keyboard-interactive authentication
- **Three.js renderer** — CPU-based 3D rendering in terminal
- **QR code encoder** — Full QR code generation with Reed-Solomon error correction
- **Keymap system** — Multi-layer key bindings with Emacs/Vim support
- **Animation & Timeline** — Tweening and animation engine
- **Plugin system** — Extensible plugin architecture
- **Web server** — Built-in HTTP documentation server
- **122 examples** — Ready-to-run example applications

---

## Installation

### Prerequisites

- Go 1.26 or higher

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

---

## Quick Start

Create a simple terminal application with a box:

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

Run it:

```bash
go run main.go
```

---

## Examples

PixTui includes 122 example applications covering all major features:

```bash
# Basic components
go run ./cmd/pixtui-demo/ hello
go run ./cmd/pixtui-demo/ box
go run ./cmd/pixtui-demo/ text
go run ./cmd/pixtui-demo/ input

go run ./cmd/pixtui-demo/ scrollbox
go run ./cmd/pixtui-demo/ textarea
go run ./cmd/pixtui-demo/ slider

# Advanced
go run ./cmd/pixtui-demo/ video
go run ./cmd/pixtui-demo/ markdown
go run ./cmd/pixtui-demo/ code
go run ./cmd/pixtui-demo/ 3d
go run ./cmd/pixtui-demo/ ssh
```

---

## Packages

| Package | Description |
|---------|-------------|
| `native` | Terminal, buffer, image, video, audio, tree-sitter, yoga layout |
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

---

## Roadmap

- [ ] Add more renderable components
- [ ] WebGPU shader support
- [ ] Mobile terminal support
- [ ] Plugin marketplace
- [ ] VS Code extension

---

## Contributing

Contributions are what make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

If you have a suggestion that would make this better, please fork the repo and create a pull request. You can also simply open an issue with the tag "enhancement". Don't forget to give the project a star! Thanks again!

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

---

## License

Distributed under the MIT License. See `LICENSE` for more information.

---

## Contact

Gomobra — [@Gomobra](https://github.com/Gomobra)

Project Link: [https://github.com/Gomobra/PixTui](https://github.com/Gomobra/PixTui)

---

## Acknowledgments

- [Choose an Open Source License](https://choosealicense.com)
- [GitHub Emoji Cheat Sheet](https://www.webpagefx.com/tools/emoji-cheat-sheet)
- [Img Shields](https://shields.io)
- [GitHub Pages](https://pages.github.com)
- [Font Awesome](https://fontawesome.com)