# PixTui Examples

[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)](https://go.dev/)

Example applications for PixTui.

## About

PixTui Examples provides 122 ready-to-run example applications covering all major features of PixTui. Each example demonstrates a specific component, feature, or use case.

## Installation

```bash
go get github.com/Gomobra/PixTui/packages/examples
```

## Usage

```go
package main

import (
	"github.com/Gomobra/PixTui/packages/examples"
)

func main() {
	examples.RunExample("hello")
}
```

## Available Examples

### Basic Components

| Example | Description |
|---|---|
| `hello` | Hello world |
| `box` | Box component |
| `text` | Text rendering |
| `input` | Input component |
| `select` | Select component |
| `scrollbox` | Scrollable box |
| `textarea` | Textarea component |
| `slider` | Slider component |
| `tabselect` | Tab select component |

### Advanced

| Example | Description |
|---|---|
| `video` | Video rendering |
| `markdown` | Markdown rendering |
| `code` | Code display |
| `3d` | 3D rendering |
| `ssh` | SSH server |
| `react` | React bindings |
| `solid` | Solid bindings |
| `animation` | Animation system |
| `plugin` | Plugin system |
| `qrcode` | QR code generation |

### Run All Examples

```bash
go run ./cmd/pixtui-demo/ hello
go run ./cmd/pixtui-demo/ box
go run ./cmd/pixtui-demo/ video
go run ./cmd/pixtui-demo/ 3d
go run ./cmd/pixtui-demo/ ssh
```

## License

Distributed under the MIT License. See `LICENSE` for more information.

## Contact

Gomobra — [@Gomobra](https://github.com/Gomobra)

Project Link: [https://github.com/Gomobra/PixTui](https://github.com/Gomobra/PixTui)