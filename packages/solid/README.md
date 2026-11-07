# @opentui/solid
Solid bindings for OpenTUI Go

## Installation

```sh
go get PixTui/packages/solid
```

## Usage

```go
package main

import (
	"PixTui/packages/solid"
	"PixTui/packages/solid/src/components"
)

func main() {
	renderer := solid.NewSolidRenderer(ctx)
	component := components.Box(map[string]interface{}{
		"border": true,
		"title":  "Hello",
	})
	renderer.Render(component, buffer, 0)
}
```
