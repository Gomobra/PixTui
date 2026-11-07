package web

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// ---------------------------------------------------------------------------
// Astro-style site configuration
// ---------------------------------------------------------------------------

// SiteConfig holds the configuration for a documentation site, mirroring the
// Astro configuration concept (site title, base URL, output directory, etc.).
type SiteConfig struct {
	// Title is the site title shown in HTML <title> tags and the OpenAPI info.
	Title string
	// BaseURL is the public base URL of the site (e.g. "https://opentui.com").
	BaseURL string
	// OutputDir is the directory static site generation writes HTML files to.
	OutputDir string
	// TrailingSlash appends a trailing slash to directory-style URLs.
	TrailingSlash bool
	// AssetsDir is the subdirectory under OutputDir where static assets live.
	AssetsDir string
}

// defaultSiteConfig returns a sensible default configuration.
func defaultSiteConfig() SiteConfig {
	return SiteConfig{
		Title:        "OpenTUI Go Documentation",
		BaseURL:      "",
		OutputDir:    "_site",
		TrailingSlash: true,
		AssetsDir:    "_astro",
	}
}

// NewSiteWithConfig creates a new Site pre-populated with default pages and
// configured according to cfg.
func NewSiteWithConfig(cfg SiteConfig) *Site {
	title := cfg.Title
	if title == "" {
		title = defaultSiteConfig().Title
	}
	s := NewSite(title)
	_ = cfg.BaseURL
	_ = cfg.TrailingSlash
	_ = cfg.AssetsDir
	return s
}

// ---------------------------------------------------------------------------
// Content directory and collections
// ---------------------------------------------------------------------------

// CollectionSchema describes the shape of entries in a content collection.
// It mirrors the Astro content schema concept (Zod-like field definitions).
type CollectionSchema struct {
	// Fields maps field names to a description of the expected type/constraints.
	Fields map[string]string `json:"fields,omitempty"`
}

// Collection is a named group of content entries with an associated schema.
type Collection struct {
	Name   string          `json:"name"`
	Schema CollectionSchema `json:"schema,omitempty"`
	// Entries holds the parsed entries in this collection.
	Entries []map[string]any `json:"entries,omitempty"`
}

// ContentDir represents a content directory (like Astro's src/content) that
// holds named collections of entries.
type ContentDir struct {
	Path        string                `json:"path"`
	Collections map[string]*Collection `json:"collections,omitempty"`
}

// NewContentDir creates a ContentDir for the given filesystem path.
func NewContentDir(path string) *ContentDir {
	return &ContentDir{
		Path:        path,
		Collections: make(map[string]*Collection),
	}
}

// AddCollection registers a named collection with a schema.
func (cd *ContentDir) AddCollection(name string, schema CollectionSchema) *Collection {
	c := &Collection{Name: name, Schema: schema}
	cd.Collections[name] = c
	return c
}

// GetCollection returns a collection by name, or nil if not found.
func (cd *ContentDir) GetCollection(name string) *Collection {
	return cd.Collections[name]
}

// AddEntry appends an entry to a named collection, creating the collection
// with an empty schema if it does not exist.
func (cd *ContentDir) AddEntry(name string, entry map[string]any) {
	c := cd.Collections[name]
	if c == nil {
		c = &Collection{Name: name}
		cd.Collections[name] = c
	}
	c.Entries = append(c.Entries, entry)
}

// ---------------------------------------------------------------------------
// File-based routing
// ---------------------------------------------------------------------------

// FileRoute maps a filesystem path pattern to an HTTP handler.
type FileRoute struct {
	Pattern string        `json:"pattern"`
	Handler http.HandlerFunc `json:"-"`
}

// FileRouter routes requests based on file paths, mirroring file-based routing
// where each file under a content directory becomes a route.
type FileRouter struct {
	routes []FileRoute
}

// NewFileRouter creates a FileRouter.
func NewFileRouter() *FileRouter {
	return &FileRouter{routes: make([]FileRoute, 0)}
}

// Register maps a file path pattern to a handler.
func (fr *FileRouter) Register(pattern string, handler http.HandlerFunc) {
	fr.routes = append(fr.routes, FileRoute{Pattern: pattern, Handler: handler})
}

// ServeHTTP dispatches to the first matching route. Patterns support a trailing
// "*" wildcard (e.g. "/docs/*") and exact matches otherwise.
func (fr *FileRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	for _, route := range fr.routes {
		if matchFilePattern(route.Pattern, path) {
			route.Handler(w, r)
			return
		}
	}
	http.NotFound(w, r)
}

// matchFilePattern reports whether path matches pattern. A pattern ending in
// "*" matches any path with the given prefix; otherwise it must be exact.
func matchFilePattern(pattern, path string) bool {
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(path, prefix)
	}
	return pattern == path
}

// ---------------------------------------------------------------------------
// Static site generation
// ---------------------------------------------------------------------------

// renderPageToString renders a page to an HTML string (used for static
// generation). It mirrors renderPage but writes to a strings.Builder.
func (s *Site) renderPageToString(page *Page) string {
	s.mu.RLock()
	nav := renderNav(s)
	title := page.Title
	s.mu.RUnlock()
	var sb strings.Builder
	sb.WriteString("<!DOCTYPE html><html><head><meta charset=\"utf-8\"><title>")
	sb.WriteString(htmlEscape(title))
	sb.WriteString("</title><style>body{font-family:sans-serif;margin:0;padding:1em;background:#f5f5f5;color:#222}nav{background:#1a1a1a;color:#fff;padding:1em}nav a{color:#4fc3f7;margin-right:1em;text-decoration:none}nav a:hover{text-decoration:underline}main{background:#fff;padding:1em;border-radius:4px;margin-top:1em}</style></head><body>")
	sb.WriteString(nav)
	sb.WriteString("<main>")
	sb.WriteString(page.Content)
	sb.WriteString("</main></body></html>")
	return sb.String()
}

// BuildToDir performs static site generation: it renders every page in the
// site to an HTML file under dir, plus sitemap.xml and robots.txt.
func (s *Site) BuildToDir(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	s.mu.RLock()
	paths := make([]string, len(s.order))
	copy(paths, s.order)
	s.mu.RUnlock()

	for _, path := range paths {
		s.mu.RLock()
		page := s.pages[path]
		s.mu.RUnlock()
		if page == nil {
			continue
		}
		html := s.renderPageToString(page)

		outPath := pageToFilePath(dir, page.Path)
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(outPath, []byte(html), 0o644); err != nil {
			return err
		}
	}

	// sitemap.xml
	s.mu.RLock()
	sm := strings.Builder{}
	sm.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	sm.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`)
	for _, p := range paths {
		s.mu.RUnlock()
		s.mu.RLock()
		page := s.pages[p]
		if page != nil {
			fmt.Fprintf(&sm, "\n  <url><loc>%s</loc><lastmod>%s</lastmod></url>", htmlEscape(page.Path), page.UpdatedAt.Format("2006-01-02"))
		}
		s.mu.RUnlock()
		s.mu.RLock()
	}
	sm.WriteString("\n</urlset>")
	sitemap := sm.String()
	s.mu.RUnlock()

	if err := os.WriteFile(filepath.Join(dir, "sitemap.xml"), []byte(sitemap), 0o644); err != nil {
		return err
	}

	// robots.txt
	robots := "User-agent: *\nAllow: /\nSitemap: /sitemap.xml\n"
	if err := os.WriteFile(filepath.Join(dir, "robots.txt"), []byte(robots), 0o644); err != nil {
		return err
	}

	return nil
}

// pageToFilePath converts a page path (e.g. "/docs/getting-started") to a
// filesystem path under dir (e.g. dir/docs/getting-started/index.html).
func pageToFilePath(dir, pagePath string) string {
	p := strings.TrimPrefix(pagePath, "/")
	if p == "" {
		p = "index"
	}
	if !strings.HasSuffix(p, ".html") {
		p = filepath.Join(p, "index.html")
	}
	return filepath.Join(dir, p)
}

// ensureWriteFile writes data to path, creating parent directories as needed.
func ensureWriteFile(path string, data []byte, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, perm)
}
