# PixTui Web

[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)](https://go.dev/)

HTTP documentation server for PixTui.

## About

PixTui Web provides an HTTP server for serving documentation websites. It includes routing, search, static site generation, and OpenAPI support.

## Features

- **Router** — HTTP router with path parameters, middleware, and CORS
- **Site** — Documentation site with pages, API docs, and release notes
- **Page** — Content pages with metadata and search symbols
- **DocSearch** — Full-text search with scoring and highlighting
- **Static site generation** — Build to directory with sitemap and robots.txt
- **OpenAPI** — Auto-generated OpenAPI documentation
- **Content collections** — File-based content management

## Installation

```bash
go get github.com/Gomobra/PixTui/packages/web
```

## Quick Start

```go
package main

import (
	"github.com/Gomobra/PixTui/packages/web"
)

func main() {
	site := web.NewSite("PixTui Docs")
	page := web.NewPage("Getting Started", "# Getting Started\n\nWelcome to PixTui!")
	site.AddPage(page)

	site.Serve(":8080")
}
```

## Components

| Component | Description |
|---|---|
| `Router` | HTTP router with path parameters and middleware |
| `Site` | Documentation site container |
| `Page` | Content page with metadata |
| `DocSearch` | Full-text search engine |
| `SiteConfig` | Site configuration |
| `FileRouter` | File-based routing |

## License

Distributed under the MIT License. See `LICENSE` for more information.

## Contact

Gomobra — [@Gomobra](https://github.com/Gomobra)

Project Link: [https://github.com/Gomobra/PixTui](https://github.com/Gomobra/PixTui)