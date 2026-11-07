# @opentui/web
Documentation website for OpenTUI Go

## Installation

```sh
go get opentui-go/packages/web
```

## Usage

```go
package main

import (
	"opentui-go/packages/web"
)

func main() {
	site := web.NewSite("OpenTUI Docs")
	page := web.NewPage("Getting Started", "# Getting Started\n\nWelcome to OpenTUI Go!")
	site.AddPage(page)
}
```
