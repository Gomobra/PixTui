package native

import (
	"testing"
)

func TestContextCreate(t *testing.T) {
	ctx := NewResourceContext()
	if ctx == nil {
		t.Fatal("expected non-nil context")
	}
	if ctx.ID() == 0 {
		t.Error("expected non-zero context ID")
	}
}

func TestContextClose(t *testing.T) {
	ctx := NewResourceContext()
	ctx.Close()
	if ctx.Status() != ContextStatusClosed {
		t.Error("expected context to be closed")
	}
}

func TestContextCreateSession(t *testing.T) {
	ctx := NewResourceContext()
	session, err := ctx.CreateSession()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session == nil {
		t.Fatal("expected non-nil session")
	}
	if session.ID() == 0 {
		t.Error("expected non-zero session ID")
	}
}

func TestContextDestroySession(t *testing.T) {
	ctx := NewResourceContext()
	session, _ := ctx.CreateSession()
	err := ctx.DestroySession(session.ID())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestContextCreateBuffer(t *testing.T) {
	ctx := NewResourceContext()
	buf, err := ctx.CreateBuffer(80, 24)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf == nil {
		t.Fatal("expected non-nil buffer")
	}
	if buf.Width() != 80 || buf.Height() != 24 {
		t.Errorf("expected 80x24, got %dx%d", buf.Width(), buf.Height())
	}
}

func TestBufferClear(t *testing.T) {
	buf := NewOptimizedBuffer(10, 5)
	buf.Clear()
	access := buf.Access()
	if access.Char[0] != ' ' {
		t.Error("expected space char after clear")
	}
}

func TestBufferSetGetCell(t *testing.T) {
	buf := NewOptimizedBuffer(10, 5)
	buf.SetCell(0, 0, 'A', 0xFFFF, 0, 0)
	ch, fg, _, attrs, ok := buf.GetCell(0, 0)
	if !ok {
		t.Error("expected cell to be valid")
	}
	if ch != 'A' {
		t.Errorf("expected char 'A', got %c", ch)
	}
	if fg != 0xFFFF {
		t.Errorf("expected fg 0xFFFF, got 0x%X", fg)
	}
	_ = attrs
}

func TestBufferResize(t *testing.T) {
	buf := NewOptimizedBuffer(10, 5)
	buf.Resize(20, 10)
	if buf.Width() != 20 || buf.Height() != 10 {
		t.Errorf("expected 20x10, got %dx%d", buf.Width(), buf.Height())
	}
}

func TestBufferSwap(t *testing.T) {
	buf := NewOptimizedBuffer(10, 5)
	gen1 := buf.Generation()
	buf.Swap()
	gen2 := buf.Generation()
	if gen2 == gen1 {
		t.Error("expected generation to change after swap")
	}
}

func TestYogaCreate(t *testing.T) {
	yoga := NewYoga()
	if yoga == nil {
		t.Fatal("expected non-nil yoga")
	}
	node := yoga.CreateNode()
	if node == nil {
		t.Fatal("expected non-nil node")
	}
}

func TestYogaCalculateLayout(t *testing.T) {
	yoga := NewYoga()
	node := yoga.CreateNode()
	node.SetWidth(100)
	node.SetHeight(50)
	yoga.CalculateLayout(node, 100, 50)
	layout := node.GetComputedLayout()
	if layout.Width != 100 || layout.Height != 50 {
		t.Errorf("expected 100x50, got %vx%v", layout.Width, layout.Height)
	}
}

func TestTerminalCreate(t *testing.T) {
	term := NewTerminal()
	if term == nil {
		t.Fatal("expected non-nil terminal")
	}
	w, h := term.GetDimensions()
	if w != 80 || h != 24 {
		t.Errorf("expected 80x24, got %dx%d", w, h)
	}
}

func TestClipboardCreate(t *testing.T) {
	clip := NewClipboard()
	if clip == nil {
		t.Fatal("expected non-nil clipboard")
	}
}

func TestTextBuffer(t *testing.T) {
	tb := NewTextBuffer()
	tb.SetContent("hello\nworld")
	if tb.Length() != 11 {
		t.Errorf("expected length 11, got %d", tb.Length())
	}
	if tb.LineCount() != 2 {
		t.Errorf("expected 2 lines, got %d", tb.LineCount())
	}
	line, ok := tb.GetLine(0)
	if !ok || line != "hello" {
		t.Errorf("expected line 'hello', got '%s'", line)
	}
}

func TestEditBuffer(t *testing.T) {
	eb := NewEditBuffer()
	eb.SetContent("hello")
	eb.Insert(5, " world")
	if eb.GetContent() != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", eb.GetContent())
	}
	if !eb.Undo() {
		t.Error("expected undo to succeed")
	}
	if eb.GetContent() != "hello" {
		t.Errorf("expected 'hello' after undo, got '%s'", eb.GetContent())
	}
	if !eb.Redo() {
		t.Error("expected redo to succeed")
	}
	if eb.GetContent() != "hello world" {
		t.Errorf("expected 'hello world' after redo, got '%s'", eb.GetContent())
	}
}

func TestEditorView(t *testing.T) {
	eb := NewEditBuffer()
	ev := NewEditorView(eb)
	ev.SetCursor(5, 10)
	x, y := ev.Cursor()
	if x != 5 || y != 10 {
		t.Errorf("expected cursor (5,10), got (%d,%d)", x, y)
	}
}

func TestImageCreate(t *testing.T) {
	img := NewImage(100, 50)
	if img.Width != 100 || img.Height != 50 {
		t.Errorf("expected 100x50, got %dx%d", img.Width, img.Height)
	}
}

func TestVideoCreate(t *testing.T) {
	vid := NewVideo(100, 50)
	if vid.Width != 100 || vid.Height != 50 {
		t.Errorf("expected 100x50, got %dx%d", vid.Width, vid.Height)
	}
}

func TestGraphemePool(t *testing.T) {
	pool := NewGraphemePool()
	id1 := pool.Intern("a")
	id2 := pool.Intern("a")
	id3 := pool.Intern("b")
	if id1 != id2 {
		t.Error("expected same ID for same grapheme")
	}
	if id1 == id3 {
		t.Error("expected different IDs for different graphemes")
	}
	if pool.Lookup(id1) != "a" {
		t.Error("expected lookup to return 'a'")
	}
}

func TestLinkPool(t *testing.T) {
	pool := NewLinkPool()
	id1 := pool.Intern("https://example.com")
	id2 := pool.Intern("https://example.com")
	if id1 != id2 {
		t.Error("expected same ID for same URL")
	}
	if pool.Lookup(id1) != "https://example.com" {
		t.Error("expected lookup to return URL")
	}
}

func TestUTF8Width(t *testing.T) {
	if StringWidth("hello") != 5 {
		t.Error("expected width 5 for 'hello'")
	}
	if RuneWidth('A') != 1 {
		t.Error("expected width 1 for 'A'")
	}
	if RuneWidth('汉') != 2 {
		t.Error("expected width 2 for CJK character")
	}
}

func TestSGR(t *testing.T) {
	seq := SGR(White, Black, TextAttributeBold)
	if seq == "" {
		t.Error("expected non-empty SGR sequence")
	}
	if seq[:2] != "\x1b[" {
		t.Error("expected SGR to start with ESC[")
	}
}

func TestAllocator(t *testing.T) {
	alloc := NewAllocator(1024)
	ptr := alloc.Alloc(100)
	if ptr == nil {
		t.Error("expected non-nil allocation")
	}
	alloc.Free(ptr)
	allocated, limit := alloc.Stats()
	if allocated != 0 {
		t.Errorf("expected 0 allocated after free, got %d", allocated)
	}
	if limit != 1024 {
		t.Errorf("expected limit 1024, got %d", limit)
	}
}
