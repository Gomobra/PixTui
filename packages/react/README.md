# @opentui/react
React bindings for OpenTUI Go

## Installation

```sh
go get opentui-go/packages/react
```

## Usage

```go
package main

import (
	"opentui-go/packages/react"
	"opentui-go/packages/react/src/components"
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
