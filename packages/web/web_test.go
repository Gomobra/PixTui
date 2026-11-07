package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRouterRegisterAndMatch(t *testing.T) {
	r := NewRouter()
	called := false
	r.GET("/hello", func(w http.ResponseWriter, req *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if !called {
		t.Error("expected handler to be called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestRouterRouteParameter(t *testing.T) {
	r := NewRouter()
	var captured string
	r.GET("/docs/{slug}", func(w http.ResponseWriter, req *http.Request) {
		captured = RouteParam(req, "slug")
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/docs/getting-started", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if captured != "getting-started" {
		t.Errorf("expected slug 'getting-started', got %q", captured)
	}
}

func TestRouterMethodNotAllowed(t *testing.T) {
	r := NewRouter()
	r.GET("/only-get", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/only-get", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405 for POST on GET-only route, got %d", w.Code)
	}
}

func TestRouterNotFound(t *testing.T) {
	r := NewRouter()
	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestRouterMiddleware(t *testing.T) {
	r := NewRouter()
	var middlewareCalled bool
	r.USE(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			middlewareCalled = true
			next.ServeHTTP(w, req)
		})
	})
	r.GET("/mw", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/mw", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if !middlewareCalled {
		t.Error("expected middleware to be called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestCORSMiddleware(t *testing.T) {
	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})
	handler := CORS("*")(inner)

	// Normal request
	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if !called {
		t.Error("expected inner handler to be called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("expected Access-Control-Allow-Origin: *")
	}

	// OPTIONS preflight
	called = false
	req = httptest.NewRequest(http.MethodOptions, "/api", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if called {
		t.Error("expected inner handler NOT to be called for OPTIONS")
	}
	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204 for OPTIONS, got %d", w.Code)
	}
}

func TestOpenAPIDoc(t *testing.T) {
	site := NewSite("Test Site")
	doc := site.openAPIDoc()

	if doc.OpenAPI != "3.0.1" {
		t.Errorf("expected openapi 3.0.1, got %s", doc.OpenAPI)
	}
	if doc.Info.Title != "Test Site" {
		t.Errorf("expected title 'Test Site', got %s", doc.Info.Title)
	}
	if doc.Info.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %s", doc.Info.Version)
	}
	if _, ok := doc.Paths["/health"]; !ok {
		t.Error("expected /health path in OpenAPI doc")
	}
	if _, ok := doc.Paths["/openapi.json"]; !ok {
		t.Error("expected /openapi.json path in OpenAPI doc")
	}
	if _, ok := doc.Paths["/api/{module}"]; !ok {
		t.Error("expected /api/{module} path in OpenAPI doc")
	}
}

func TestSitePageLifecycle(t *testing.T) {
	site := NewSite("Test Site")

	page := NewPage("/test", "Test Page", "<p>content</p>")
	if !site.AddPage(page) {
		t.Error("expected AddPage to return true for new page")
	}
	if site.AddPage(page) {
		t.Error("expected AddPage to return false for duplicate page")
	}

	if site.PageCount() != 1 {
		t.Errorf("expected 1 page, got %d", site.PageCount())
	}

	got := site.GetPage("/test")
	if got == nil {
		t.Fatal("expected to find page /test")
	}
	if got.Title != "Test Page" {
		t.Errorf("expected title 'Test Page', got %s", got.Title)
	}

	pages := site.ListPages()
	if len(pages) != 1 {
		t.Errorf("expected 1 page from ListPages, got %d", len(pages))
	}

	if !site.RemovePage("/test") {
		t.Error("expected RemovePage to return true")
	}
	if site.RemovePage("/test") {
		t.Error("expected RemovePage to return false for missing page")
	}
	if site.PageCount() != 0 {
		t.Errorf("expected 0 pages after removal, got %d", site.PageCount())
	}
}

func TestSiteAPIDocLifecycle(t *testing.T) {
	site := NewSite("Test Site")

	doc := &APIDoc{
		Module:     "test/mod",
		Slug:       "mod",
		Title:      "Test Module",
		Description: "A test module",
	}
	if !site.AddAPIDoc(doc) {
		t.Error("expected AddAPIDoc to return true for new doc")
	}
	if site.AddAPIDoc(doc) {
		t.Error("expected AddAPIDoc to return false for duplicate doc")
	}

	if site.APIDocCount() != 1 {
		t.Errorf("expected 1 API doc, got %d", site.APIDocCount())
	}

	got := site.GetAPIDoc("test/mod")
	if got == nil {
		t.Fatal("expected to find API doc test/mod")
	}
	if got.Title != "Test Module" {
		t.Errorf("expected title 'Test Module', got %s", got.Title)
	}

	docs := site.ListAPIDocs()
	if len(docs) != 1 {
		t.Errorf("expected 1 API doc from ListAPIDocs, got %d", len(docs))
	}
}

func TestSiteReleaseNoteLifecycle(t *testing.T) {
	site := NewSite("Test Site")

	note := &ReleaseNote{
		Version:  "1.0.0",
		Date:     "2026-01-01",
		Summary:  "Test release",
		Body:     "Test body",
		Changes: []ReleaseChange{
			{Section: "Added", Markdown: "- new feature"},
		},
	}
	if !site.AddReleaseNote(note) {
		t.Error("expected AddReleaseNote to return true for new note")
	}
	if site.AddReleaseNote(note) {
		t.Error("expected AddReleaseNote to return false for duplicate note")
	}

	if site.ReleaseNoteCount() != 1 {
		t.Errorf("expected 1 release note, got %d", site.ReleaseNoteCount())
	}

	got := site.GetReleaseNote("1.0.0")
	if got == nil {
		t.Fatal("expected to find release note 1.0.0")
	}
	if got.Summary != "Test release" {
		t.Errorf("expected summary 'Test release', got %s", got.Summary)
	}

	notes := site.ListReleaseNotes()
	if len(notes) != 1 {
		t.Errorf("expected 1 release note from ListReleaseNotes, got %d", len(notes))
	}
}

func TestSiteSearch(t *testing.T) {
	site := NewSite("Test Site")

	page := NewPage("/search-test", "Searchable Page", "<p>unique searchable term here</p>")
	page.AddSearchSymbol("unique-term")
	site.AddPage(page)

	results := site.Search("unique")
	if len(results) == 0 {
		t.Error("expected search results for 'unique'")
	}

	results = site.Search("nonexistentxyz")
	if len(results) != 0 {
		t.Errorf("expected no search results for 'nonexistentxyz', got %d", len(results))
	}

	if site.DocSearch().IndexCount() == 0 {
		t.Error("expected search index to have entries")
	}
}

func TestSiteHealthEndpoint(t *testing.T) {
	site := NewSite("Test Site")
	site.AddPage(NewPage("/", "Home", "<p>home</p>"))

	// We can't easily test the handler since it's a closure in Serve.
	// Instead test the underlying data it uses.
	if site.PageCount() != 1 {
		t.Errorf("expected 1 page, got %d", site.PageCount())
	}
}

func TestRouterHeadMethod(t *testing.T) {
	r := NewRouter()
	r.GET("/head", func(w http.ResponseWriter, req *http.Request) {
		// A HEAD handler must not write a body; the router discards it.
		w.Write([]byte("body should be discarded"))
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodHead, "/head", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200 for HEAD, got %d", w.Code)
	}
	if w.Body.Len() != 0 {
		t.Errorf("expected empty HEAD body, got %q", w.Body.String())
	}
}

func TestRouterMultipleMethods(t *testing.T) {
	r := NewRouter()
	var getCalled, postCalled bool

	r.HandleFunc("/multi", func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodGet {
			getCalled = true
		} else if req.Method == http.MethodPost {
			postCalled = true
		}
		w.WriteHeader(http.StatusOK)
	}, "GET", "POST")

	// GET
	req := httptest.NewRequest(http.MethodGet, "/multi", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK || !getCalled {
		t.Error("expected GET handler to be called")
	}

	// POST
	postCalled = false
	req = httptest.NewRequest(http.MethodPost, "/multi", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK || !postCalled {
		t.Error("expected POST handler to be called")
	}

	// PUT should be method not allowed
	req = httptest.NewRequest(http.MethodPut, "/multi", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405 for PUT, got %d", w.Code)
	}
}

func TestRouterConvenienceMethods(t *testing.T) {
	r := NewRouter()
	r.GET("/get-only", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.POST("/post-only", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	r.PUT("/put-only", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.DELETE("/delete-only", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	for _, tc := range []struct {
		method   string
		path     string
		expected int
	}{
		{http.MethodGet, "/get-only", http.StatusOK},
		{http.MethodPost, "/post-only", http.StatusCreated},
		{http.MethodPut, "/put-only", http.StatusOK},
		{http.MethodDelete, "/delete-only", http.StatusNoContent},
	} {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != tc.expected {
			t.Errorf("%s %s: expected status %d, got %d", tc.method, tc.path, tc.expected, w.Code)
		}
	}
}

func TestSitePageMethods(t *testing.T) {
	page := NewPage("/page", "Page Title", "<p>content</p>")

	if page.Path != "/page" {
		t.Errorf("expected path '/page', got %s", page.Path)
	}
	if page.Title != "Page Title" {
		t.Errorf("expected title 'Page Title', got %s", page.Title)
	}
	if page.NavTitle != "Page Title" {
		t.Errorf("expected navTitle 'Page Title', got %s", page.NavTitle)
	}

	page.UpdateContent("<p>new content</p>")
	if page.Content != "<p>new content</p>" {
		t.Error("expected content to be updated")
	}

	page.SetMetadata("key", "value")
	if page.Metadata["key"] != "value" {
		t.Error("expected metadata key/value")
	}

	page.SetNavTitle("Nav Title")
	if page.NavTitle != "Nav Title" {
		t.Error("expected navTitle to be updated")
	}

	page.SetChapter("chapter1")
	if page.Chapter != "chapter1" {
		t.Error("expected chapter to be updated")
	}

	page.AddSearchSymbol("sym1")
	page.AddSearchSymbol("sym1") // duplicate, should be ignored
	if len(page.SearchSymbols) != 1 {
		t.Errorf("expected 1 search symbol, got %d", len(page.SearchSymbols))
	}

	page.SetOrder(5)
	if page.Order != 5 {
		t.Errorf("expected order 5, got %d", page.Order)
	}

	page.SetWeight(10)
	if page.Weight != 10 {
		t.Errorf("expected weight 10, got %d", page.Weight)
	}

	text := page.SearchableText()
	if !strings.Contains(text, "Page Title") {
		t.Error("expected searchable text to contain title")
	}
	if !strings.Contains(text, "new content") {
		t.Error("expected searchable text to contain content")
	}
}

func TestNewPageWithOptions(t *testing.T) {
	page := NewPageWithOptions(
		"/opts", "Opts", "NavOpts", "chap",
		"<p>content</p>", "desc", 3, 7,
		[]string{"sym1", "sym2"},
		map[string]string{"k": "v"},
	)
	if page.Path != "/opts" {
		t.Errorf("expected path '/opts', got %s", page.Path)
	}
	if page.Chapter != "chap" {
		t.Errorf("expected chapter 'chap', got %s", page.Chapter)
	}
	if page.Order != 3 {
		t.Errorf("expected order 3, got %d", page.Order)
	}
	if page.Weight != 7 {
		t.Errorf("expected weight 7, got %d", page.Weight)
	}
	if len(page.SearchSymbols) != 2 {
		t.Errorf("expected 2 search symbols, got %d", len(page.SearchSymbols))
	}
	if page.Metadata["k"] != "v" {
		t.Error("expected metadata k=v")
	}
}

func TestSiteListPagesByOrder(t *testing.T) {
	site := NewSite("Test Site")

	p1 := NewPageWithOptions("/a", "A", "", "", "<p>a</p>", "", 2, 0, nil, nil)
	p2 := NewPageWithOptions("/b", "B", "", "", "<p>b</p>", "", 1, 0, nil, nil)
	p3 := NewPageWithOptions("/c", "C", "", "", "<p>c</p>", "", 0, 0, nil, nil)
	site.AddPage(p1)
	site.AddPage(p2)
	site.AddPage(p3)

	pages := site.ListPagesByOrder()
	if len(pages) != 3 {
		t.Fatalf("expected 3 pages, got %d", len(pages))
	}
	if pages[0].Path != "/c" {
		t.Errorf("expected first page /c, got %s", pages[0].Path)
	}
	if pages[1].Path != "/b" {
		t.Errorf("expected second page /b, got %s", pages[1].Path)
	}
	if pages[2].Path != "/a" {
		t.Errorf("expected third page /a, got %s", pages[2].Path)
	}
}

func TestSiteListPagesByChapter(t *testing.T) {
	site := NewSite("Test Site")

	p1 := NewPageWithOptions("/a", "A", "", "chap1", "<p>a</p>", "", 0, 0, nil, nil)
	p2 := NewPageWithOptions("/b", "B", "", "chap1", "<p>b</p>", "", 0, 0, nil, nil)
	p3 := NewPageWithOptions("/c", "C", "", "chap2", "<p>c</p>", "", 0, 0, nil, nil)
	site.AddPage(p1)
	site.AddPage(p2)
	site.AddPage(p3)

	byChapter := site.ListPagesByChapter()
	if len(byChapter["chap1"]) != 2 {
		t.Errorf("expected 2 pages in chap1, got %d", len(byChapter["chap1"]))
	}
	if len(byChapter["chap2"]) != 1 {
		t.Errorf("expected 1 page in chap2, got %d", len(byChapter["chap2"]))
	}
	if len(byChapter["_general"]) != 0 {
		t.Errorf("expected 0 pages in _general, got %d", len(byChapter["_general"]))
	}
}

func TestHtmlEscape(t *testing.T) {
	if htmlEscape(`&<>"`) != `&amp;&lt;&gt;&quot;` {
		t.Error("htmlEscape did not escape correctly")
	}
}