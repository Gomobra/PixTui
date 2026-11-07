# @pixtui/react
React bindings for PixTui

## Installation

```sh
go get PixTui/packages/react
```

## Usage

```go
package main

import (
	"PixTui/packages/react"
	"PixTui/packages/react/src/components"
)

func main() {
	renderer := react.NewReactRenderer(ctx)
	element := components.Box(map[string]interface{}{
		"border": true,
		"title":  "Hello",
	})
	renderer.Render(element, buffer, 0)
}
```
