# @pixtui/web
Documentation website for PixTui

## Installation

```sh
go get PixTui/packages/web
```

## Usage

```go
package main

import (
	"PixTui/packages/web"
)

func main() {
	site := web.NewSite("PixTui Docs")
	page := web.NewPage("Getting Started", "# Getting Started\n\nWelcome to PixTui!")
	site.AddPage(page)
}
```
