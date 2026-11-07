package renderables

import (
	"strings"
	"testing"

	"github.com/Gomobra/PixTui/packages/core"
	"github.com/Gomobra/PixTui/packages/native"
)

// bufferToString renders a BufferAccess snapshot as a multi-line string.
// Each non-empty cell is represented by its rune; trailing spaces per line
// are trimmed and trailing empty lines are removed.
func bufferToString(access native.BufferAccess) string {
	var lines []string
	for y := 0; y < access.Height; y++ {
		var sb strings.Builder
		for x := 0; x < access.Width; x++ {
			idx := y*access.Width + x
			ch := access.Char[idx]
			if ch == 0 || ch == ' ' {
				sb.WriteRune(' ')
			} else {
				sb.WriteRune(rune(ch))
			}
		}
		line := strings.TrimRight(sb.String(), " ")
		lines = append(lines, line)
	}
	// Trim trailing empty lines
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return strings.Join(lines, "\n")
}

// TestRenderableSnapshot captures and compares output of a TextBufferRenderable.
func TestRenderableSnapshot(t *testing.T) {
	ctx := newTestCtx()
	buf := core.NewTextBufferWithContent("Hello\nWorld")
	r := NewTextBufferRenderable(ctx, TextBufferRenderableOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      "test-textbuffer",
			Visible: boolPtr(true),
		},
		Buffer: buf,
	})
	r.SetWidth(80)
	r.SetHeight(24)
	renderBuf := core.NewNativeBuffer(80, 24)
	r.RenderSelf(renderBuf, 0)
	access := renderBuf.Access()
	got := bufferToString(access)
	want := "Hello\nWorld"
	if got != want {
		t.Errorf("TextBufferRenderable snapshot mismatch:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

// TestBoxSnapshot captures and compares output of a Box renderable.
func TestBoxSnapshot(t *testing.T) {
	ctx := newTestCtx()
	r := NewBoxRenderable(ctx, BoxOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      "test-box",
			Visible: boolPtr(true),
		},
		Border: true,
		Title:  "Test",
	})
	r.SetWidth(20)
	r.SetHeight(5)
	buf := core.NewNativeBuffer(20, 5)
	r.RenderSelf(buf, 0)
	access := buf.Access()
	got := bufferToString(access)
	// Width 20: corner(1) + title "Test"(4) + horiz fill(14) + corner(1) = 20
	// Inner rows: corner(1) + spaces(18) + corner(1) = 20
	want := "┌Test" + strings.Repeat("─", 14) + "┐\n" +
		"│" + strings.Repeat(" ", 18) + "│\n" +
		"│" + strings.Repeat(" ", 18) + "│\n" +
		"│" + strings.Repeat(" ", 18) + "│\n" +
		"└" + strings.Repeat("─", 18) + "┘"
	if got != want {
		t.Errorf("Box snapshot mismatch:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

// TestTextSnapshot captures and compares output of a Text renderable.
func TestTextSnapshot(t *testing.T) {
	ctx := newTestCtx()
	r := NewTextRenderable(ctx, TextOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      "test-text",
			Visible: boolPtr(true),
		},
		Content: "Hello World",
	})
	r.SetWidth(80)
	r.SetHeight(24)
	buf := core.NewNativeBuffer(80, 24)
	r.RenderSelf(buf, 0)
	access := buf.Access()
	got := bufferToString(access)
	want := "Hello World"
	if got != want {
		t.Errorf("Text snapshot mismatch:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

// TestInputSnapshot captures and compares output of an Input renderable.
func TestInputSnapshot(t *testing.T) {
	ctx := newTestCtx()
	r := NewInputRenderable(ctx, InputOptions{
		RenderableOptions: core.RenderableOptions[core.Renderable]{
			ID:      "test-input",
			Visible: boolPtr(true),
		},
		Value: "type here",
	})
	r.SetWidth(80)
	r.SetHeight(24)
	buf := core.NewNativeBuffer(80, 24)
	r.RenderSelf(buf, 0)
	access := buf.Access()
	got := bufferToString(access)
	want := "type here"
	if got != want {
		t.Errorf("Input snapshot mismatch:\ngot:\n%s\nwant:\n%s", got, want)
	}
}
