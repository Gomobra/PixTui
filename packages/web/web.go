package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// Site represents the documentation website.
type Site struct {
	title      string
	pages      map[string]*Page
	order      []string // insertion order for stable listing
	apiDocs    map[string]*APIDoc
	apiOrder   []string
	releaseNotes map[string]*ReleaseNote
	releaseOrder []string
	docSearch   *DocSearch
	mu          sync.RWMutex
}

// NewSite creates a new documentation site.
func NewSite(title string) *Site {
	return &Site{
		title:        title,
		pages:        make(map[string]*Page),
		order:        make([]string, 0),
		apiDocs:      make(map[string]*APIDoc),
		apiOrder:     make([]string, 0),
		releaseNotes: make(map[string]*ReleaseNote),
		releaseOrder: make([]string, 0),
		docSearch:    NewDocSearch(),
	}
}

// AddPage adds a page to the site. If a page with the same path already
// exists it is replaced. Returns true if the page was newly added.
func (s *Site) AddPage(page *Page) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.pages[page.Path]; exists {
		s.pages[page.Path] = page
		return false
	}
	s.pages[page.Path] = page
	s.order = append(s.order, page.Path)
	// Index the page for search. The searchable body is the page content so
	// that in-page terms (not just the description) are discoverable, and the
	// declared search symbols are indexed separately for symbol lookup.
	s.docSearch.AddEntry(&SearchEntry{
		Chapter:  page.Chapter,
		Title:    page.Title,
		NavTitle: page.NavTitle,
		URL:      page.Path,
		Text:     page.Content,
		Symbols:  page.SearchSymbols,
	})
	return true
}

// RemovePage removes a page from the site by path. Returns true if a page
// was removed.
func (s *Site) RemovePage(path string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.pages[path]; !exists {
		return false
	}
	delete(s.pages, path)
	for i, p := range s.order {
		if p == path {
			s.order = append(s.order[:i], s.order[i+1:]...)
			break
		}
	}
	return true
}

// GetPage returns a page by path.
func (s *Site) GetPage(path string) *Page {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.pages[path]
}

// GetPageByTitle returns the first page matching the given title.
func (s *Site) GetPageByTitle(title string) *Page {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, page := range s.pages {
		if page.Title == title {
			return page
		}
	}
	return nil
}

// ListPages returns all pages in insertion order.
func (s *Site) ListPages() []*Page {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Page, len(s.order))
	for i, path := range s.order {
		result[i] = s.pages[path]
	}
	return result
}

// ListPagesByOrder returns pages sorted by Order field, then by insertion order.
func (s *Site) ListPagesByOrder() []*Page {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Page, len(s.order))
	for i, path := range s.order {
		result[i] = s.pages[path]
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Order != result[j].Order {
			return result[i].Order < result[j].Order
		}
		return result[i].Path < result[j].Path
	})
	return result
}

// ListPagesByChapter returns pages grouped by chapter.
func (s *Site) ListPagesByChapter() map[string][]*Page {
	s.mu.RLock()
	defer s.mu.RUnlock()
	byChapter := make(map[string][]*Page)
	for _, path := range s.order {
		p := s.pages[path]
		chapter := p.Chapter
		if chapter == "" {
			chapter = "_general"
		}
		byChapter[chapter] = append(byChapter[chapter], p)
	}
	return byChapter
}

// AddPageWithOrder adds a page with a specific display order.
func (s *Site) AddPageWithOrder(page *Page, order int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	page.SetOrder(order)
	if _, exists := s.pages[page.Path]; exists {
		s.pages[page.Path] = page
		return false
	}
	s.pages[page.Path] = page
	s.order = append(s.order, page.Path)
	return true
}

// UpsertPage adds or replaces a page dynamically. Returns true if newly added.
func (s *Site) UpsertPage(page *Page) bool {
	return s.AddPage(page)
}

// PageCount returns the number of pages in the site.
func (s *Site) PageCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.pages)
}

// Title returns the site title.
func (s *Site) Title() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.title
}

// AddAPIDoc adds API documentation for a module. If a doc with the same module
// already exists it is replaced. Returns true if the doc was newly added.
func (s *Site) AddAPIDoc(doc *APIDoc) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.apiDocs[doc.Module]; exists {
		s.apiDocs[doc.Module] = doc
		return false
	}
	s.apiDocs[doc.Module] = doc
	s.apiOrder = append(s.apiOrder, doc.Module)
	// Index the doc for search.
	s.docSearch.AddEntry(&SearchEntry{
		Chapter: "API Reference",
		Title:   doc.Title,
		URL:     "/api/" + doc.Slug,
		Text:    doc.Description,
	})
	return true
}

// AddReleaseNote adds release notes for a version. If a note with the same
// version already exists it is replaced. Returns true if the note was newly added.
func (s *Site) AddReleaseNote(note *ReleaseNote) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.releaseNotes[note.Version]; exists {
		s.releaseNotes[note.Version] = note
		return false
	}
	s.releaseNotes[note.Version] = note
	s.releaseOrder = append(s.releaseOrder, note.Version)
	// Index the release for search.
	s.docSearch.AddEntry(&SearchEntry{
		Chapter: "Releases",
		Title:   note.Version,
		URL:     "/releases/" + note.Version,
		Text:    note.Summary + " " + note.Body,
	})
	return true
}

// GetAPIDoc returns API documentation by module name.
func (s *Site) GetAPIDoc(module string) *APIDoc {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.apiDocs[module]
}

// GetReleaseNote returns release notes by version.
func (s *Site) GetReleaseNote(version string) *ReleaseNote {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.releaseNotes[version]
}

// ListAPIDocs returns all API docs in insertion order.
func (s *Site) ListAPIDocs() []*APIDoc {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*APIDoc, len(s.apiOrder))
	for i, module := range s.apiOrder {
		result[i] = s.apiDocs[module]
	}
	return result
}

// APIDocCount returns the number of API docs in the site.
func (s *Site) APIDocCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.apiDocs)
}

// ListReleaseNotes returns all release notes in insertion order.
func (s *Site) ListReleaseNotes() []*ReleaseNote {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*ReleaseNote, len(s.releaseOrder))
	for i, version := range s.releaseOrder {
		result[i] = s.releaseNotes[version]
	}
	return result
}

// ReleaseNoteCount returns the number of release notes in the site.
func (s *Site) ReleaseNoteCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.releaseNotes)
}

// Search searches the documentation index for the given query and returns
// matching results sorted by relevance.
func (s *Site) Search(query string) []SearchResult {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.docSearch.Search(query)
}

// DocSearch returns the site's search index.
func (s *Site) DocSearch() *DocSearch {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.docSearch
}

// Page represents a documentation page with content and metadata.
type Page struct {
	mu          sync.RWMutex
	Path        string            `json:"path"`
	Title       string            `json:"title"`
	NavTitle    string            `json:"nav_title,omitempty"`
	Chapter     string            `json:"chapter,omitempty"`
	Content     string            `json:"content"`
	Description string            `json:"description,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	SearchSymbols []string         `json:"search_symbols,omitempty"`
	Order       int               `json:"order,omitempty"`
	Weight      int               `json:"weight,omitempty"`
	CreatedAt   time.Time         `json:"created_at,omitempty"`
	UpdatedAt   time.Time         `json:"updated_at,omitempty"`
}

// NewPage creates a new documentation page with the given path, title, and
// content. CreatedAt and UpdatedAt are set to the current time.
func NewPage(path, title, content string) *Page {
	now := time.Now()
	return &Page{
		Path:         path,
		Title:        title,
		NavTitle:     title,
		Chapter:      "",
		Content:      content,
		Metadata:     make(map[string]string),
		SearchSymbols: make([]string, 0),
		Order:        0,
		Weight:       0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// NewPageWithMetadata creates a new page with additional metadata.
func NewPageWithMetadata(path, title, content string, metadata map[string]string) *Page {
	p := NewPage(path, title, content)
	p.Metadata = metadata
	return p
}

// NewPageWithOptions creates a new page with full metadata control.
func NewPageWithOptions(path, title, navTitle, chapter, content, description string, order, weight int, searchSymbols []string, metadata map[string]string) *Page {
	now := time.Now()
	p := &Page{
		Path:           path,
		Title:          title,
		NavTitle:       navTitle,
		Chapter:        chapter,
		Content:        content,
		Description:    description,
		Metadata:       metadata,
		SearchSymbols:  searchSymbols,
		Order:          order,
		Weight:         weight,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if p.Metadata == nil {
		p.Metadata = make(map[string]string)
	}
	if p.SearchSymbols == nil {
		p.SearchSymbols = make([]string, 0)
	}
	if p.NavTitle == "" {
		p.NavTitle = title
	}
	return p
}

// UpdateContent updates the page content and refreshes UpdatedAt.
func (p *Page) UpdateContent(content string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Content = content
	p.UpdatedAt = time.Now()
}

// SetMetadata sets a single metadata key/value on the page.
func (p *Page) SetMetadata(key, value string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.Metadata == nil {
		p.Metadata = make(map[string]string)
	}
	p.Metadata[key] = value
	p.UpdatedAt = time.Now()
}

// SetNavTitle sets the navigation title for the page.
func (p *Page) SetNavTitle(navTitle string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.NavTitle = navTitle
	p.UpdatedAt = time.Now()
}

// SetChapter sets the chapter/group the page belongs to.
func (p *Page) SetChapter(chapter string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Chapter = chapter
	p.UpdatedAt = time.Now()
}

// AddSearchSymbol adds a symbol to the page's search index.
func (p *Page) AddSearchSymbol(symbol string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.SearchSymbols == nil {
		p.SearchSymbols = make([]string, 0)
	}
	for _, s := range p.SearchSymbols {
		if s == symbol {
			return
		}
	}
	p.SearchSymbols = append(p.SearchSymbols, symbol)
	p.UpdatedAt = time.Now()
}

// SetOrder sets the display order of the page.
func (p *Page) SetOrder(order int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Order = order
	p.UpdatedAt = time.Now()
}

// SetWeight sets the weight/priority of the page for search ranking.
func (p *Page) SetWeight(weight int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Weight = weight
	p.UpdatedAt = time.Now()
}

// SearchableText returns a concatenated string of all searchable text on the page.
func (p *Page) SearchableText() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	var parts []string
	parts = append(parts, p.Title, p.NavTitle, p.Chapter, p.Content, p.Description)
	for _, s := range p.SearchSymbols {
		parts = append(parts, s)
	}
	for k, v := range p.Metadata {
		parts = append(parts, k, v)
	}
	return strings.Join(parts, " ")
}

// Logo represents the OpenTUI logo.
var Logo = []byte{}

// Build builds the documentation site.
func Build() error {
	// Build logic (placeholder for static site generation).
	return nil
}

// ---------------------------------------------------------------------------
// API documentation types
// ---------------------------------------------------------------------------

// APIDoc represents documentation for a single API module, including its
// exports grouped by kind (functions, classes, interfaces, structs, types,
// enums, variables, namespaces).
type APIDoc struct {
	Module      string        `json:"module"`
	Slug        string        `json:"slug"`
	Title       string        `json:"title"`
	Description string        `json:"description,omitempty"`
	Exports     []APIDocExport `json:"exports,omitempty"`
	Groups      []APIDocGroup  `json:"groups,omitempty"`
}

// APIDocExport represents a single exported symbol within a module.
type APIDocExport struct {
	Name         string          `json:"name"`
	Kind         string          `json:"kind"` // function, class, interface, struct, type, enum, variable, namespace
	Declarations []string        `json:"declarations,omitempty"`
	Members      []APIDocMember   `json:"members,omitempty"`
	Label        string          `json:"label,omitempty"`
	Deprecated   bool            `json:"deprecated,omitempty"`
	Guide        string          `json:"guide,omitempty"`
}

// APIDocMember represents a member of an export (e.g. a struct field or method).
type APIDocMember struct {
	Feature string `json:"feature"`
	Label   string `json:"label,omitempty"`
}

// APIDocGroup groups exports by kind.
type APIDocGroup struct {
	ID      string         `json:"id"`
	Title   string         `json:"title"`
	Exports []APIDocExport `json:"exports"`
}

// APIIndex is the index of all API modules.
type APIIndex struct {
	Modules []*APIDoc `json:"modules"`
	Groups  []APIDocGroup `json:"groups,omitempty"`
}

// Known export kind IDs (mirrors the TypeScript EXPORT_GROUPS).
const (
	ExportKindFunctions  = "functions"
	ExportKindClasses    = "classes"
	ExportKindInterfaces = "interfaces"
	ExportKindStructs    = "structs"
	ExportKindTypes      = "types"
	ExportKindEnums      = "enums"
	ExportKindVariables  = "variables"
	ExportKindNamespaces = "namespaces"
)

// ExportGroupTitles maps kind IDs to human-readable titles.
var ExportGroupTitles = map[string]string{
	ExportKindFunctions:  "Functions",
	ExportKindClasses:    "Classes",
	ExportKindInterfaces: "Interfaces",
	ExportKindStructs:    "Structs",
	ExportKindTypes:      "Types",
	ExportKindEnums:      "Enums",
	ExportKindVariables:  "Constants and variables",
	ExportKindNamespaces: "Namespaces and re-exports",
}

// ExportGroupOrder defines the display order of export groups.
var ExportGroupOrder = []string{
	ExportKindFunctions,
	ExportKindClasses,
	ExportKindInterfaces,
	ExportKindStructs,
	ExportKindTypes,
	ExportKindEnums,
	ExportKindVariables,
	ExportKindNamespaces,
}

// groupExports groups APIDocExport entries by their kind.
func groupExports(exports []APIDocExport) []APIDocGroup {
	byKind := make(map[string][]APIDocExport)
	for _, exp := range exports {
		byKind[exp.Kind] = append(byKind[exp.Kind], exp)
	}
	var groups []APIDocGroup
	for _, kind := range ExportGroupOrder {
		exps, ok := byKind[kind]
		if !ok || len(exps) == 0 {
			continue
		}
		groups = append(groups, APIDocGroup{
			ID:      kind,
			Title:   ExportGroupTitles[kind],
			Exports: exps,
		})
	}
	return groups
}

// ---------------------------------------------------------------------------
// Release notes types
// ---------------------------------------------------------------------------

// ReleaseNote represents release notes for a single version.
type ReleaseNote struct {
	Version  string         `json:"version"`
	Date     string         `json:"date"`
	Summary  string         `json:"summary"`
	Body     string         `json:"body,omitempty"`
	Changes  []ReleaseChange `json:"changes,omitempty"`
}

// ReleaseChange represents a single change entry in release notes.
type ReleaseChange struct {
	Section  string   `json:"section"` // Breaking changes, Added, Changed, Deprecated, Removed, Fixed, Security
	Markdown string   `json:"markdown"`
	Docs     []string `json:"docs,omitempty"`
}

// Known release note sections (mirrors the TypeScript RELEASE_SECTIONS).
var ReleaseSections = []string{
	"Breaking changes",
	"Added",
	"Changed",
	"Deprecated",
	"Removed",
	"Fixed",
	"Security",
}

// ---------------------------------------------------------------------------
// Documentation search types
// ---------------------------------------------------------------------------

// SearchEntry is a single searchable unit in the documentation index.
type SearchEntry struct {
	Chapter  string   `json:"chapter"`
	Title    string   `json:"title"`
	NavTitle string   `json:"nav_title,omitempty"`
	URL      string   `json:"url"`
	Text     string   `json:"text"`
	Symbols  []string `json:"symbols,omitempty"`
}

// SearchResult is a single search result with a relevance score.
type SearchResult struct {
	Entry     *SearchEntry `json:"entry"`
	Score     int          `json:"score"`
	Highlight string       `json:"highlight,omitempty"`
}

// DocSearch holds the documentation search index and performs queries.
type DocSearch struct {
	mu     sync.RWMutex
	index  []*SearchEntry
	query  string
	results []SearchResult
}

// NewDocSearch creates a new empty documentation search index.
func NewDocSearch() *DocSearch {
	return &DocSearch{
		index:  make([]*SearchEntry, 0),
		results: make([]SearchResult, 0),
	}
}

// AddEntry adds a search entry to the index.
func (ds *DocSearch) AddEntry(entry *SearchEntry) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.index = append(ds.index, entry)
}

// Search searches the index for the given query and returns matching results
// sorted by relevance (descending score).
func (ds *DocSearch) Search(query string) []SearchResult {
	ds.mu.Lock()
	ds.query = query
	ds.results = ds.results[:0]
	ds.mu.Unlock()

	tokens := tokenize(query)
	if len(tokens) == 0 {
		return nil
	}

	ds.mu.RLock()
	entries := make([]*SearchEntry, len(ds.index))
	copy(entries, ds.index)
	ds.mu.RUnlock()

	var matches []SearchResult
	for _, entry := range entries {
		score := scoreEntry(entry, tokens)
		if score > 0 {
			matches = append(matches, SearchResult{
				Entry:     entry,
				Score:     score,
				Highlight: makeHighlight(entry.Text, tokens),
			})
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Score != matches[j].Score {
			return matches[i].Score > matches[j].Score
		}
		return matches[i].Entry.URL < matches[j].Entry.URL
	})

	ds.mu.Lock()
	ds.results = matches
	ds.mu.Unlock()
	return matches
}

// Query returns the last search query.
func (ds *DocSearch) Query() string {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.query
}

// Results returns the last search results.
func (ds *DocSearch) Results() []SearchResult {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	result := make([]SearchResult, len(ds.results))
	copy(result, ds.results)
	return result
}

// ResultCount returns the number of results from the last search.
func (ds *DocSearch) ResultCount() int {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return len(ds.results)
}

// IndexCount returns the number of indexed entries.
func (ds *DocSearch) IndexCount() int {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return len(ds.index)
}

// tokenize splits a query into lowercase tokens on non-alphanumeric boundaries.
func tokenize(query string) []string {
	var tokens []string
	current := strings.Builder{}
	for _, r := range query {
		if isAlphaNum(r) {
			current.WriteRune(toLower(r))
		} else if current.Len() > 0 {
			tokens = append(tokens, current.String())
			current.Reset()
		}
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}
	return tokens
}

func isAlphaNum(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}

func toLower(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r - 'A' + 'a'
	}
	return r
}

// scoreEntry scores a search entry against the given tokens. A token must be
// found in at least one of the title, navigation title, body text, search
// symbols, or URL for the entry to match at all, mirroring the TypeScript
// search index coverage rules.
func scoreEntry(entry *SearchEntry, tokens []string) int {
	titleLower := strings.ToLower(entry.Title)
	navTitleLower := strings.ToLower(entry.NavTitle)
	textLower := strings.ToLower(entry.Text)
	urlLower := strings.ToLower(entry.URL)
	symbolsLower := make([]string, len(entry.Symbols))
	for i, symbol := range entry.Symbols {
		symbolsLower[i] = strings.ToLower(symbol)
	}

	score := 0
	allMatch := true
	for _, token := range tokens {
		matched := false
		if strings.Contains(titleLower, token) {
			score += 100
			matched = true
		}
		if strings.Contains(navTitleLower, token) {
			score += 90
			matched = true
		}
		if strings.Contains(textLower, token) {
			score += 10
			matched = true
		}
		for _, symbol := range symbolsLower {
			if strings.Contains(symbol, token) {
				score += 50
				matched = true
				break
			}
		}
		if strings.Contains(urlLower, token) {
			score += 5
			matched = true
		}
		if !matched {
			allMatch = false
		}
	}
	if !allMatch {
		return 0
	}
	// Bonus for exact title match.
	if titleLower == strings.Join(tokens, " ") {
		score += 1000
	}
	return score
}

// makeHighlight returns a snippet of text centered on the first matching token.
func makeHighlight(text string, tokens []string) string {
	textLower := strings.ToLower(text)
	bestIdx := -1
	bestLen := 0
	for _, token := range tokens {
		idx := strings.Index(textLower, token)
		if idx != -1 && (bestIdx == -1 || idx < bestIdx) {
			bestIdx = idx
			bestLen = len(token)
		}
	}
	if bestIdx == -1 {
		return text
	}
	start := bestIdx - 80
	if start < 0 {
		start = 0
	}
	end := bestIdx + bestLen + 80
	if end > len(text) {
		end = len(text)
	}
	snippet := text[start:end]
	if start > 0 {
		snippet = "…" + snippet
	}
	if end < len(text) {
		snippet = snippet + "…"
	}
	return snippet
}

// ---------------------------------------------------------------------------
// Serve
// ---------------------------------------------------------------------------

// Serve serves the documentation site via HTTP with multiple routes.
//
// Routes:
//   GET /                    - home page
//   GET /docs                - documentation index
//   GET /docs/{slug}         - specific documentation page
//   GET /docs/releases       - release notes list (HTML)
//   GET /api                 - API index
//   GET /api/{module}        - specific API module
//   GET /api/status          - API status
//   GET /api/pages           - list of pages
//   GET /api/pages/{path}    - specific page JSON
//   GET /api/health          - detailed health check
//   GET /releases            - list of releases (JSON)
//   GET /releases/{version}  - specific release notes (JSON)
//   GET /search?q=...        - documentation search (HTML or JSON)
//   GET /about               - about page
//   GET /health              - health check JSON
//   GET /sitemap.xml         - sitemap
//   GET /robots.txt          - robots.txt
// renderNav renders the site navigation bar HTML.
func renderNav(site *Site) string {
	var sb strings.Builder
	sb.WriteString("<nav><strong>OpenTUI Go Docs</strong> ")
	sb.WriteString("<a href=\"/\">Home</a> ")
	sb.WriteString("<a href=\"/docs\">Docs</a> ")
	sb.WriteString("<a href=\"/api\">API</a> ")
	sb.WriteString("<a href=\"/releases\">Releases</a> ")
	sb.WriteString("<a href=\"/about\">About</a> ")
	sb.WriteString("</nav>")
	return sb.String()
}

// htmlEscape escapes HTML special characters in s.
func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

// ---------------------------------------------------------------------------
// Router
// ---------------------------------------------------------------------------

// Middleware wraps an http.Handler with pre/post-processing logic.
type Middleware func(http.Handler) http.Handler

// routeEntry is a single registered route.
type routeEntry struct {
	pattern  string
	handler  http.Handler
	methods  map[string]bool // nil means all methods
	params   []string        // parameter names in order
	regexp   *regexp.Regexp
}

// Router dispatches HTTP requests to handlers based on method and path pattern.
// It supports route parameters in {name} syntax and method-based routing.
type Router struct {
	mu        sync.RWMutex
	routes    []routeEntry
	middleware []Middleware
}

// NewRouter creates a new Router.
func NewRouter() *Router {
	return &Router{
		routes:    make([]routeEntry, 0),
		middleware: make([]Middleware, 0),
	}
}

// USE registers middleware that wraps every route handler, in the order added.
func (r *Router) USE(mw Middleware) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.middleware = append(r.middleware, mw)
}

// HandleFunc registers a handler for a path pattern with optional HTTP methods.
// When no methods are given, the handler accepts all methods.
func (r *Router) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request), methods ...string) {
	r.Handle(pattern, http.HandlerFunc(handler), methods...)
}

// Handle registers a handler for a path pattern with optional HTTP methods.
func (r *Router) Handle(pattern string, handler http.Handler, methods ...string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	methodSet := make(map[string]bool)
	if len(methods) > 0 {
		for _, m := range methods {
			methodSet[strings.ToUpper(m)] = true
		}
	}

	reg, paramNames := compilePattern(pattern)
	r.routes = append(r.routes, routeEntry{
		pattern: pattern,
		handler: handler,
		methods: methodSet,
		params:  paramNames,
		regexp:  reg,
	})
}

// GET registers a GET-only route.
func (r *Router) GET(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	r.HandleFunc(pattern, handler, "GET")
}

// POST registers a POST-only route.
func (r *Router) POST(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	r.HandleFunc(pattern, handler, "POST")
}

// PUT registers a PUT-only route.
func (r *Router) PUT(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	r.HandleFunc(pattern, handler, "PUT")
}

// DELETE registers a DELETE-only route.
func (r *Router) DELETE(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	r.HandleFunc(pattern, handler, "DELETE")
}

// ServeHTTP dispatches the request to the matching route handler.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	path := req.URL.Path
	method := req.Method

	// HEAD requests are matched against GET routes; the response body is
	// discarded so only headers and status code are returned.
	effectiveMethod := method
	if method == http.MethodHead {
		effectiveMethod = http.MethodGet
	}

	for _, entry := range r.routes {
		matches := entry.regexp.FindStringSubmatch(path)
		if matches == nil {
			continue
		}
		if len(entry.methods) > 0 && !entry.methods[effectiveMethod] {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			r.mu.RUnlock()
			return
		}

		// Extract route parameters into the request context.
		if len(entry.params) > 0 {
			values := make(map[string]string, len(entry.params))
			for i, name := range entry.params {
				if i+1 < len(matches) {
					values[name] = matches[i+1]
				}
			}
			req = req.WithContext(context.WithValue(req.Context(), paramKey{}, values))
		}

		r.mu.RUnlock()

		// Apply middleware (outermost = last registered).
		var h http.Handler = entry.handler
		for i := len(r.middleware) - 1; i >= 0; i-- {
			h = r.middleware[i](h)
		}
		if method == http.MethodHead {
			h = &headHandler{Handler: h}
		}
		h.ServeHTTP(w, req)
		return
	}

	r.mu.RUnlock()
	http.NotFound(w, req)
}

// headHandler wraps an http.Handler so that HEAD responses discard the body
// while still sending the headers and status code a GET would produce.
type headHandler struct {
	http.Handler
}

func (h *headHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Handler.ServeHTTP(&headResponseWriter{ResponseWriter: w}, r)
}

// headResponseWriter discards the response body so only headers and status
// code are sent, as required for HEAD responses.
type headResponseWriter struct {
	http.ResponseWriter
	wroteHeader bool
	code        int
}

func (hw *headResponseWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func (hw *headResponseWriter) WriteHeader(code int) {
	if hw.wroteHeader {
		return
	}
	hw.wroteHeader = true
	hw.code = code
	hw.ResponseWriter.WriteHeader(code)
}

// paramKey is the context key for route parameters.
type paramKey struct{}

// RouteParam returns the value of a route parameter from the request context.
func RouteParam(req *http.Request, name string) string {
	if values, ok := req.Context().Value(paramKey{}).(map[string]string); ok {
		return values[name]
	}
	return ""
}

// Routes returns a copy of all registered route patterns, useful for
// documentation and server introspection.
func (r *Router) Routes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	patterns := make([]string, len(r.routes))
	for i, entry := range r.routes {
		patterns[i] = entry.pattern
	}
	return patterns
}

// compilePattern converts a path pattern with {name} parameters to a regexp.
// For example, "/docs/{slug}" becomes "^/docs/(?P<slug>[^/]+)$".
func compilePattern(pattern string) (*regexp.Regexp, []string) {
	var b strings.Builder
	b.WriteString("^")
	var params []string
	i := 0
	for i < len(pattern) {
		if pattern[i] == '{' {
			j := strings.IndexByte(pattern[i:], '}')
			if j == -1 {
				b.WriteString(regexp.QuoteMeta(string(pattern[i])))
				i++
				continue
			}
			name := pattern[i+1 : i+j]
			params = append(params, name)
			fmt.Fprintf(&b, "(?P<%s>[^/]+)", name)
			i += j + 1
		} else {
			b.WriteString(regexp.QuoteMeta(string(pattern[i])))
			i++
		}
	}
	b.WriteString("$")
	reg := regexp.MustCompile(b.String())
	return reg, params
}

// ---------------------------------------------------------------------------
// CORS middleware
// ---------------------------------------------------------------------------

// CORS returns middleware that sets Cross-Origin Resource Sharing headers.
func CORS(allowOrigin string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding")
			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Max-Age", "86400")
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ---------------------------------------------------------------------------
// OpenAPI documentation
// ---------------------------------------------------------------------------

// OpenAPIInfo is the OpenAPI root info object.
type OpenAPIInfo struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version"`
}

// OpenAPI is a minimal OpenAPI document describing the web API.
type OpenAPI struct {
	OpenAPI string        `json:"openapi"`
	Info    OpenAPIInfo   `json:"info"`
	Paths   map[string]*OpenAPIPathItem `json:"paths"`
}

// OpenAPIPathItem describes operations for a single path.
type OpenAPIPathItem struct {
	Get    *OpenAPIOperation `json:"get,omitempty"`
	Post   *OpenAPIOperation `json:"post,omitempty"`
	Put    *OpenAPIOperation `json:"put,omitempty"`
	Delete *OpenAPIOperation `json:"delete,omitempty"`
}

// OpenAPIOperation describes a single API operation.
type OpenAPIOperation struct {
	Summary     string                   `json:"summary,omitempty"`
	Description string                   `json:"description,omitempty"`
	Parameters  []OpenAPIParameter       `json:"parameters,omitempty"`
	Responses   map[string]*OpenAPIResponse `json:"responses,omitempty"`
}

// OpenAPIParameter describes a single parameter.
type OpenAPIParameter struct {
	Name        string `json:"name"`
	In          string `json:"in"` // path, query, header
	Required    bool   `json:"required"`
	Schema      *OpenAPISchema `json:"schema,omitempty"`
}

// OpenAPISchema describes a JSON schema.
type OpenAPISchema struct {
	Type string `json:"type"`
}

// OpenAPIResponse describes a single response.
type OpenAPIResponse struct {
	Description string `json:"description"`
}

// openAPIDoc returns the OpenAPI document for the site's API.
func (s *Site) openAPIDoc() *OpenAPI {

	return &OpenAPI{
		OpenAPI: "3.0.1",
		Info: OpenAPIInfo{
			Title:       s.title,
			Description: "OpenTUI Go documentation and API server.",
			Version:     "1.0.0",
		},
		Paths: map[string]*OpenAPIPathItem{
			"/health": {
				Get: &OpenAPIOperation{
					Summary: "Health check",
					Responses: map[string]*OpenAPIResponse{
						"200": {Description: "OK"},
					},
				},
			},
			"/api": {
				Get: &OpenAPIOperation{
					Summary: "List API documentation modules",
					Responses: map[string]*OpenAPIResponse{
						"200": {Description: "OK"},
					},
				},
			},
			"/api/{module}": {
				Get: &OpenAPIOperation{
					Summary: "Get a single API documentation module",
					Parameters: []OpenAPIParameter{
						{Name: "module", In: "path", Required: true, Schema: &OpenAPISchema{Type: "string"}},
					},
					Responses: map[string]*OpenAPIResponse{
						"200": {Description: "OK"},
						"404": {Description: "Not found"},
					},
				},
			},
			"/api/pages": {
				Get: &OpenAPIOperation{
					Summary: "List all documentation pages",
					Responses: map[string]*OpenAPIResponse{
						"200": {Description: "OK"},
					},
				},
			},
			"/api/health": {
				Get: &OpenAPIOperation{
					Summary: "Detailed API health check",
					Responses: map[string]*OpenAPIResponse{
						"200": {Description: "OK"},
					},
				},
			},
			"/releases": {
				Get: &OpenAPIOperation{
					Summary: "List all release notes",
					Responses: map[string]*OpenAPIResponse{
						"200": {Description: "OK"},
					},
				},
			},
			"/releases/{version}": {
				Get: &OpenAPIOperation{
					Summary: "Get release notes for a version",
					Parameters: []OpenAPIParameter{
						{Name: "version", In: "path", Required: true, Schema: &OpenAPISchema{Type: "string"}},
					},
					Responses: map[string]*OpenAPIResponse{
						"200": {Description: "OK"},
						"404": {Description: "Not found"},
					},
				},
			},
			"/search": {
				Get: &OpenAPIOperation{
					Summary: "Search documentation",
					Parameters: []OpenAPIParameter{
						{Name: "q", In: "query", Required: true, Schema: &OpenAPISchema{Type: "string"}},
					},
					Responses: map[string]*OpenAPIResponse{
						"200": {Description: "OK"},
					},
				},
			},
			"/docs/{slug}": {
				Get: &OpenAPIOperation{
					Summary: "Get a documentation page",
					Parameters: []OpenAPIParameter{
						{Name: "slug", In: "path", Required: true, Schema: &OpenAPISchema{Type: "string"}},
					},
					Responses: map[string]*OpenAPIResponse{
						"200": {Description: "OK"},
						"404": {Description: "Not found"},
					},
				},
			},
			"/about": {
				Get: &OpenAPIOperation{
					Summary: "Get the about page",
					Responses: map[string]*OpenAPIResponse{
						"200": {Description: "OK"},
					},
				},
			},
			"/api/pages/{path}": {
				Get: &OpenAPIOperation{
					Summary: "Get a single documentation page as JSON",
					Parameters: []OpenAPIParameter{
						{Name: "path", In: "path", Required: true, Schema: &OpenAPISchema{Type: "string"}},
					},
					Responses: map[string]*OpenAPIResponse{
						"200": {Description: "OK"},
						"404": {Description: "Not found"},
					},
				},
			},
			"/sitemap.xml": {
				Get: &OpenAPIOperation{
					Summary: "Get the sitemap",
					Responses: map[string]*OpenAPIResponse{
						"200": {Description: "OK"},
					},
				},
			},
			"/robots.txt": {
				Get: &OpenAPIOperation{
					Summary: "Get the robots exclusion policy",
					Responses: map[string]*OpenAPIResponse{
						"200": {Description: "OK"},
					},
				},
			},
			"/openapi.json": {
				Get: &OpenAPIOperation{
					Summary: "Get the OpenAPI document",
					Responses: map[string]*OpenAPIResponse{
						"200": {Description: "OK"},
					},
				},
			},
		},
	}
}

// Serve serves the documentation site via HTTP with multiple routes.
//
// Routes:
//   GET /                    - home page
//   GET /docs                - documentation index
//   GET /docs/{slug}         - specific documentation page
//   GET /docs/releases       - release notes list (HTML)
//   GET /api                 - API index
//   GET /api/{module}        - specific API module
//   GET /api/status          - API status
//   GET /api/pages           - list of pages
//   GET /api/pages/{path}    - specific page JSON
//   GET /api/health          - detailed health check
//   GET /releases            - list of releases (JSON)
//   GET /releases/{version}  - specific release notes (JSON)
//   GET /search?q=...        - documentation search (HTML or JSON)
//   GET /about               - about page
//   GET /health              - health check JSON
//   GET /sitemap.xml         - sitemap
//   GET /robots.txt          - robots.txt
//   GET /openapi.json        - OpenAPI document
func Serve(addr string) error {
	site := NewSite("OpenTUI Go Documentation")

	// Register default pages.
	site.AddPageWithOrder(NewPageWithOptions(
		"/", "Home", "Home", "",
		"<h1>OpenTUI Go</h1><p>Welcome to the OpenTUI Go documentation.</p>",
		"OpenTUI Go documentation home page.", 0, 100,
		[]string{"opentui", "home", "welcome"},
		map[string]string{"section": "home", "lang": "en"},
	), 0)
	site.AddPageWithOrder(NewPageWithOptions(
		"/docs", "Documentation", "Documentation", "",
		"<h1>OpenTUI Go Documentation</h1><p>Official OpenTUI Go documentation index.</p>",
		"OpenTUI Go documentation index.", 1, 90,
		[]string{"docs", "documentation", "guide"},
		map[string]string{"section": "docs", "lang": "en"},
	), 1)
	site.AddPageWithOrder(NewPageWithOptions(
		"/about", "About", "About", "",
		"<h1>About OpenTUI</h1><p>OpenTUI is a terminal user interface framework written in Go.</p>",
		"About OpenTUI Go.", 2, 80,
		[]string{"about", "opentui"},
		map[string]string{"section": "about", "lang": "en"},
	), 2)

	// Register default API docs.
	site.AddAPIDoc(&APIDoc{
		Module:     "opentui/core",
		Slug:       "core",
		Title:      "OpenTUI Core",
		Description: "Core rendering and layout API for OpenTUI Go.",
		Exports: []APIDocExport{
			{Name: "Renderer", Kind: ExportKindClasses, Declarations: []string{"type Renderer struct"}, Guide: "/docs/core-concepts/renderer"},
			{Name: "Renderable", Kind: ExportKindInterfaces, Declarations: []string{"type Renderable interface"}, Guide: "/docs/core-concepts/renderables"},
		},
	})
	site.AddAPIDoc(&APIDoc{
		Module:     "opentui/components",
		Slug:       "components",
		Title:      "Components",
		Description: "Built-in UI components for OpenTUI Go.",
		Exports: []APIDocExport{
			{Name: "Box", Kind: ExportKindClasses, Declarations: []string{"type Box struct"}, Guide: "/docs/components/box"},
			{Name: "Text", Kind: ExportKindClasses, Declarations: []string{"type Text struct"}, Guide: "/docs/components/text"},
		},
	})

	// Register default release notes.
	site.AddReleaseNote(&ReleaseNote{
		Version: "0.8.0",
		Date:    "2026-10-29",
		Summary: "A small change repaints only the rows around it, animations share one clock, and ScrollBox scrolls in native code.",
		Body:    "Most of 0.8.0 makes each frame cheaper. In 0.7.0, every frame cleared and repainted the whole screen and compared every row with the terminal, even when only a spinner had moved. Now an unchanged frame paints nothing, and a frame with a small change repaints only the rows around it.",
		Changes: []ReleaseChange{
			{Section: "Added", Markdown: "- renderer.onAnimationTick(fps, callback) runs callbacks on one shared animation clock.", Docs: []string{"/docs/application-apis/animation"}},
			{Section: "Fixed", Markdown: "- ScrollBox now scrolls and follows new content in native code.", Docs: []string{"/docs/components/scrollbox"}},
		},
	})
	site.AddReleaseNote(&ReleaseNote{
		Version: "0.6.0",
		Date:    "2026-10-10",
		Summary: "Layout, hit testing, and painting move into a native scene, and native rendering resources now belong to an owner Context.",
		Body:    "OpenTUI 0.5 kept the renderable tree in JavaScript. On each frame, JavaScript walked the tree and read each node's position back from Yoga. In 0.6.0, native code owns the tree. Each Renderable has a node in a native scene.",
		Changes: []ReleaseChange{
			{Section: "Breaking changes", Markdown: "- Renderable no longer has render(), getLayoutNode(), or a protected yogaNode.", Docs: []string{"/docs/extend/custom-renderables"}},
			{Section: "Added", Markdown: "- Native scene owns the renderable tree.", Docs: []string{"/docs/core-concepts/rendering-pipeline"}},
		},
	})

	router := NewRouter()

	// CORS middleware for API routes.
	apiCORS := CORS("*")

	// Helper: normalize path (strip trailing slash except root)
	normalizePath := func(p string) string {
		if p != "/" && strings.HasSuffix(p, "/") {
			return strings.TrimSuffix(p, "/")
		}
		return p
	}

	// Home route.
	router.GET("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" {
			http.NotFound(w, req)
			return
		}
		page := site.GetPage("/")
		if page == nil {
			http.NotFound(w, req)
			return
		}
		site.renderPage(w, page)
	})

	// Documentation index route.
	router.GET("/docs", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/docs" {
			http.NotFound(w, req)
			return
		}
		page := site.GetPage("/docs")
		if page == nil {
			http.NotFound(w, req)
			return
		}
		site.renderPage(w, page)
	})

	// Documentation sub-page route: /docs/{slug}
	router.GET("/docs/{slug}", func(w http.ResponseWriter, req *http.Request) {
		slug := "/" + RouteParam(req, "slug")
		page := site.GetPage(slug)
		if page == nil {
			if strings.HasSuffix(slug, "/") {
				page = site.GetPage(strings.TrimSuffix(slug, "/"))
			}
			if page == nil {
				http.NotFound(w, req)
				return
			}
		}
		site.renderPage(w, page)
	})

	// Documentation releases page (HTML): /docs/releases
	router.GET("/docs/releases", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/docs/releases" {
			http.NotFound(w, req)
			return
		}
		site.mu.RLock()
		notes := make([]*ReleaseNote, len(site.releaseOrder))
		for i, v := range site.releaseOrder {
			notes[i] = site.releaseNotes[v]
		}
		site.mu.RUnlock()
		sort.Slice(notes, func(i, j int) bool {
			return notes[i].Version > notes[j].Version
		})
		var sb strings.Builder
		sb.WriteString("<h1>Release Notes</h1>")
		for _, n := range notes {
			sb.WriteString("<div style=\"margin-bottom:1.5em;border-bottom:1px solid #ddd;padding-bottom:1em\">")
			sb.WriteString("<h2>" + htmlEscape(n.Version) + " (" + htmlEscape(n.Date) + ")</h2>")
			sb.WriteString("<p>" + htmlEscape(n.Summary) + "</p>")
			if n.Body != "" {
				sb.WriteString("<p>" + htmlEscape(n.Body) + "</p>")
			}
			if len(n.Changes) > 0 {
				sb.WriteString("<ul>")
				for _, c := range n.Changes {
					sb.WriteString("<li><strong>" + htmlEscape(c.Section) + ":</strong> " + htmlEscape(c.Markdown) + "</li>")
				}
				sb.WriteString("</ul>")
			}
			sb.WriteString("</div>")
		}
		sb.WriteString("<p><a href=\"/docs\">Back to docs</a></p>")
		site.renderPage(w, &Page{Title: "Release Notes", Content: sb.String()})
	})

	// About route.
	router.GET("/about", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/about" {
			http.NotFound(w, req)
			return
		}
		page := site.GetPage("/about")
		if page == nil {
			http.NotFound(w, req)
			return
		}
		site.renderPage(w, page)
	})

	// Search page (HTML) and JSON API: /search?q=...
	router.GET("/search", func(w http.ResponseWriter, req *http.Request) {
		query := req.URL.Query().Get("q")
		format := req.URL.Query().Get("format")
		acceptJSON := format == "json" || strings.Contains(req.Header.Get("Accept"), "application/json")
		results := site.Search(query)

		if acceptJSON {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"query":  query,
				"results": results,
				"count":   len(results),
			})
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		nav := renderNav(site)
		fmt.Fprint(w, "<!DOCTYPE html><html><head><meta charset=\"utf-8\"><title>Search - OpenTUI Go</title><style>body{font-family:sans-serif;margin:0;padding:1em;background:#f5f5f5;color:#222}nav{background:#1a1a1a;color:#fff;padding:1em}nav a{color:#4fc3f7;margin-right:1em;text-decoration:none}nav a:hover{text-decoration:underline}main{background:#fff;padding:1em;border-radius:4px;margin-top:1em}form{margin-bottom:1em}input[type=\"text\"]{width:100%%;padding:0.5em;font-size:1em;border:1px solid #ddd;border-radius:4px}ul{list-style:none;padding:0}li{margin-bottom:0.5em;padding:0.5em;border:1px solid #eee;border-radius:4px}li a{color:#1a73e8;text-decoration:none}li a:hover{text-decoration:underline}.badge{background:#e3f2fd;padding:0.1em 0.5em;border-radius:10px;font-size:0.85em;margin-left:0.5em}</style></head><body>")
		fmt.Fprint(w, nav)
		fmt.Fprint(w, "<main>")
		fmt.Fprint(w, "<h1>Search</h1>")
		if query != "" {
			fmt.Fprintf(w, "<p>Query: %s</p>", htmlEscape(query))
		}
		fmt.Fprint(w, "<form action=\"/search\" method=\"get\">")
		fmt.Fprintf(w, "<input type=\"text\" name=\"q\" value=\"%s\" placeholder=\"Search documentation...\">", htmlEscape(query))
		fmt.Fprint(w, "</form>")
		if query != "" {
			if len(results) == 0 {
				fmt.Fprint(w, "<p>No results found.</p>")
			} else {
				fmt.Fprint(w, "<ul>")
				for _, res := range results {
					fmt.Fprintf(w, "<li><a href=\"%s\">%s</a>", htmlEscape(res.Entry.URL), htmlEscape(res.Entry.Title))
					if res.Entry.Chapter != "" {
						fmt.Fprintf(w, " <span class=\"badge\">%s</span>", htmlEscape(res.Entry.Chapter))
					}
					fmt.Fprint(w, "</li>")
				}
				fmt.Fprint(w, "</ul>")
			}
		}
		fmt.Fprint(w, "<p><a href=\"/docs\">Back to docs</a></p>")
		fmt.Fprint(w, "</main></body></html>")
	})

	// Health check route.
	router.GET("/health", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/health" {
			http.NotFound(w, req)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "ok",
			"service": "pixtui-web",
			"pages":   site.PageCount(),
			"apiDocs": site.APIDocCount(),
			"releases": site.ReleaseNoteCount(),
		})
	})

	// API detailed health route: /api/health
	router.GET("/api/health", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/api/health" {
			http.NotFound(w, req)
			return
		}
		pages := site.ListPages()
		apiDocs := site.ListAPIDocs()
		releases := site.ListReleaseNotes()
		type pageSummary struct {
			Path        string            `json:"path"`
			Title       string            `json:"title"`
			Description string            `json:"description,omitempty"`
			UpdatedAt   time.Time         `json:"updated_at,omitempty"`
		}
		summaries := make([]pageSummary, len(pages))
		for i, p := range pages {
			summaries[i] = pageSummary{
				Path:        p.Path,
				Title:       p.Title,
				Description: p.Description,
				UpdatedAt:   p.UpdatedAt,
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":    "ok",
			"service":   "pixtui-web",
			"version":   "1.0.0",
			"pages":     summaries,
			"pageCount": len(summaries),
			"apiDocs":   len(apiDocs),
			"releases":  len(releases),
		})
	})

	// API index route: GET /api
	router.GET("/api", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/api" {
			http.NotFound(w, req)
			return
		}
		docs := site.ListAPIDocs()
		index := APIIndex{Modules: docs, Groups: nil}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(index)
	})

	// API status route.
	router.GET("/api/status", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/api/status" {
			http.NotFound(w, req)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		status := map[string]string{
			"status":  "ok",
			"service": "pixtui",
			"version": "1.0.0",
			"pages":   fmt.Sprintf("%d", site.PageCount()),
		}
		_ = json.NewEncoder(w).Encode(status)
	})

	// API pages list route.
	router.GET("/api/pages", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/api/pages" {
			http.NotFound(w, req)
			return
		}
		pages := site.ListPages()
		type pageSummary struct {
			Path        string            `json:"path"`
			Title       string            `json:"title"`
			Description string            `json:"description,omitempty"`
			UpdatedAt   time.Time         `json:"updated_at,omitempty"`
		}
		summaries := make([]pageSummary, len(pages))
		for i, p := range pages {
			summaries[i] = pageSummary{
				Path:        p.Path,
				Title:       p.Title,
				Description: p.Description,
				UpdatedAt:   p.UpdatedAt,
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"pages": summaries, "count": len(summaries)})
	})

	// API single page route: /api/pages/{path}
	router.GET("/api/pages/{path}", func(w http.ResponseWriter, req *http.Request) {
		pathParam := RouteParam(req, "path")
		path := normalizePath("/" + pathParam)
		if path == "" {
			http.NotFound(w, req)
			return
		}
		page := site.GetPage(path)
		if page == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "page not found"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(page)
	})

	// API module route: /api/{module}
	router.GET("/api/{module}", func(w http.ResponseWriter, req *http.Request) {
		module := RouteParam(req, "module")
		if module == "" {
			http.NotFound(w, req)
			return
		}
		doc := site.GetAPIDoc(module)
		if doc == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "api module not found"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(doc)
	})

	// Wrap API routes with CORS.
	// (In this simple Router, middleware wraps all routes; we apply CORS via USE.)
	router.USE(apiCORS)

	// Releases list route: GET /releases
	router.GET("/releases", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/releases" {
			http.NotFound(w, req)
			return
		}
		site.mu.RLock()
		notes := make([]*ReleaseNote, len(site.releaseOrder))
		for i, v := range site.releaseOrder {
			notes[i] = site.releaseNotes[v]
		}
		site.mu.RUnlock()
		sort.Slice(notes, func(i, j int) bool {
			return notes[i].Version > notes[j].Version
		})
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"releases": notes, "count": len(notes)})
	})

	// Release specific route: /releases/{version}
	router.GET("/releases/{version}", func(w http.ResponseWriter, req *http.Request) {
		version := RouteParam(req, "version")
		if version == "" {
			http.NotFound(w, req)
			return
		}
		note := site.GetReleaseNote(version)
		if note == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "release not found"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(note)
	})

	// OpenAPI document route: GET /openapi.json
	router.GET("/openapi.json", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/openapi.json" {
			http.NotFound(w, req)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(site.openAPIDoc())
	})

	// Sitemap route: GET /sitemap.xml
	router.GET("/sitemap.xml", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/sitemap.xml" {
			http.NotFound(w, req)
			return
		}
		w.Header().Set("Content-Type", "application/xml")
		fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?>`)
		fmt.Fprint(w, `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`)
		site.mu.RLock()
		for _, path := range site.order {
			page := site.pages[path]
			fmt.Fprintf(w, "\n  <url><loc>%s</loc><lastmod>%s</lastmod></url>",
				htmlEscape(page.Path), page.UpdatedAt.Format("2006-01-02"))
		}
		site.mu.RUnlock()
		fmt.Fprint(w, "\n</urlset>")
	})

	// Robots route: GET /robots.txt
	router.GET("/robots.txt", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/robots.txt" {
			http.NotFound(w, req)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, "User-agent: *\nAllow: /\nSitemap: /sitemap.xml\n")
	})

	return http.ListenAndServe(addr, router)
}

// renderPage renders a full HTML page with site navigation.
func (s *Site) renderPage(w http.ResponseWriter, page *Page) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	nav := renderNav(s)
	fmt.Fprint(w, "<!DOCTYPE html><html><head><meta charset=\"utf-8\"><title>" + htmlEscape(page.Title) + "</title><style>body{font-family:sans-serif;margin:0;padding:1em;background:#f5f5f5;color:#222}nav{background:#1a1a1a;color:#fff;padding:1em}nav a{color:#4fc3f7;margin-right:1em;text-decoration:none}nav a:hover{text-decoration:underline}main{background:#fff;padding:1em;border-radius:4px;margin-top:1em}table{border-collapse:collapse;width:100%}th,td{border:1px solid #ddd;padding:0.5em;text-align:left}th{background:#eee}code{background:#f0f0f0;padding:0.2em 0.4em;border-radius:3px}.badge{background:#e3f2fd;padding:0.1em 0.5em;border-radius:10px;font-size:0.85em}</style></head><body>")
	fmt.Fprint(w, nav)
	fmt.Fprint(w, "<main>")
	fmt.Fprint(w, page.Content)
	fmt.Fprint(w, "</main></body></html>")
}
