package web

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSiteConfigDefaults(t *testing.T) {
	cfg := defaultSiteConfig()
	if cfg.Title != "OpenTUI Go Documentation" {
		t.Errorf("expected default title, got %q", cfg.Title)
	}
	if cfg.OutputDir != "_site" {
		t.Errorf("expected default output dir _site, got %q", cfg.OutputDir)
	}
	if !cfg.TrailingSlash {
		t.Error("expected trailing slash by default")
	}
}

func TestNewSiteWithConfig(t *testing.T) {
	cfg := SiteConfig{Title: "My Site", OutputDir: "_out"}
	site := NewSiteWithConfig(cfg)
	if site.Title() != "My Site" {
		t.Errorf("expected title 'My Site', got %q", site.Title())
	}
}

func TestContentDirAddGetCollection(t *testing.T) {
	cd := NewContentDir("src/content")
	schema := CollectionSchema{Fields: map[string]string{"title": "string", "draft": "boolean"}}
	c := cd.AddCollection("docs", schema)
	if c.Name != "docs" {
		t.Errorf("expected collection name 'docs', got %q", c.Name)
	}
	if c.Schema.Fields["title"] != "string" {
		t.Error("expected schema field title=string")
	}
	got := cd.GetCollection("docs")
	if got != c {
		t.Error("expected GetCollection to return the same collection")
	}
	if cd.GetCollection("missing") != nil {
		t.Error("expected nil for missing collection")
	}
}

func TestContentDirAddEntry(t *testing.T) {
	cd := NewContentDir("src/content")
	cd.AddEntry("docs", map[string]any{"title": "Getting Started", "draft": false})
	cd.AddEntry("docs", map[string]any{"title": "Advanced", "draft": true})

	c := cd.GetCollection("docs")
	if c == nil {
		t.Fatal("expected docs collection to exist")
	}
	if len(c.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(c.Entries))
	}
	if c.Entries[0]["title"] != "Getting Started" {
		t.Errorf("expected first entry title 'Getting Started', got %v", c.Entries[0]["title"])
	}
}

func TestFileRouterRegisterAndMatch(t *testing.T) {
	fr := NewFileRouter()
	var called bool
	fr.Register("/docs/*", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/docs/getting-started", nil)
	w := httptest.NewRecorder()
	fr.ServeHTTP(w, req)

	if !called {
		t.Error("expected wildcard handler to be called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestFileRouterExactMatch(t *testing.T) {
	fr := NewFileRouter()
	var called bool
	fr.Register("/about", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	// Exact match
	req := httptest.NewRequest(http.MethodGet, "/about", nil)
	w := httptest.NewRecorder()
	fr.ServeHTTP(w, req)
	if !called || w.Code != http.StatusOK {
		t.Error("expected exact match to work")
	}

	// Non-match
	called = false
	req = httptest.NewRequest(http.MethodGet, "/about/", nil)
	w = httptest.NewRecorder()
	fr.ServeHTTP(w, req)
	if called {
		t.Error("expected no match for /about/")
	}
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for non-match, got %d", w.Code)
	}
}

func TestBuildToDir(t *testing.T) {
	dir := t.TempDir()
	site := NewSite("Test Site")
	site.AddPage(NewPage("/", "Home", "<p>home</p>"))
	site.AddPage(NewPage("/docs/intro", "Intro", "<p>intro content</p>"))
	site.AddPage(NewPage("/docs/guide", "Guide", "<p>guide content</p>"))

	if err := site.BuildToDir(dir); err != nil {
		t.Fatalf("BuildToDir: %v", err)
	}

	t.Logf("temp dir: %s", dir)
	// List files for debugging
	filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		t.Logf("file: %s", p)
		return nil
	})

	// Check index.html (root page renders to index/index.html)
	index, err := os.ReadFile(filepath.Join(dir, "index", "index.html"))
	if err != nil {
		t.Fatalf("expected index/index.html: %v", err)
	}
	if !strings.Contains(string(index), "<p>home</p>") {
		t.Error("expected index/index.html to contain home content")
	}

	// Check docs/intro/index.html
	intro, err := os.ReadFile(filepath.Join(dir, "docs", "intro", "index.html"))
	if err != nil {
		t.Fatalf("expected docs/intro/index.html: %v", err)
	}
	if !strings.Contains(string(intro), "<p>intro content</p>") {
		t.Error("expected intro index.html to contain intro content")
	}

	// Check sitemap.xml
	sitemap, err := os.ReadFile(filepath.Join(dir, "sitemap.xml"))
	if err != nil {
		t.Fatalf("expected sitemap.xml: %v", err)
	}
	if !strings.Contains(string(sitemap), "<url>") {
		t.Error("expected sitemap.xml to contain url entries")
	}

	// Check robots.txt
	robots, err := os.ReadFile(filepath.Join(dir, "robots.txt"))
	if err != nil {
		t.Fatalf("expected robots.txt: %v", err)
	}
	if !strings.Contains(string(robots), "User-agent: *") {
		t.Error("expected robots.txt to contain User-agent")
	}

	// Ensure rendered output is valid HTML (starts with doctype)
	if !strings.HasPrefix(string(index), "<!DOCTYPE html>") {
		t.Error("expected index.html to start with <!DOCTYPE html>")
	}
}
