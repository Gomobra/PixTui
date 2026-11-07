# @opentui/three
Three.js WebGPU renderer for OpenTUI Go

## Installation

```sh
go get PixTui/packages/three
```

## Usage

```go
package main

import (
	"PixTui/packages/three"
)

func main() {
	renderer := three.NewRenderer(80, 24)
	scene := three.NewScene()
	renderer.Render(scene, buffer, 0)
}
```
