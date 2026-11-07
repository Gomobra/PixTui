package testing

import (
	"io"
	"testing"
	"time"

	"github.com/Gomobra/PixTui/packages/core"
)

func TestMockTreeSitterClientManualResolve(t *testing.T) {
	client := NewMockTreeSitterClient()
	if client.IsHighlighting() {
		t.Fatal("expected not highlighting before any request")
	}

	future := client.HighlightOnce("code", "go")
	if !client.IsHighlighting() {
		t.Fatal("expected highlighting after HighlightOnce")
	}

	client.SetMockResult(HighlightResult{
		Highlights: []SimpleHighlight{{Start: 0, End: 4, StyleID: 1}},
		Warning:    "w",
	})
	client.ResolveHighlightOnce(0)

	if client.IsHighlighting() {
		t.Fatal("expected not highlighting after resolve")
	}

	result := future.Get()
	if len(result.Highlights) != 1 || result.Highlights[0].Start != 0 || result.Warning != "w" {
		t.Fatalf("unexpected result: %+v", result)
	}

	if err := client.Destroy(); err != nil {
		t.Fatalf("destroy failed: %v", err)
	}
}

func TestMockTreeSitterClientResolveAll(t *testing.T) {
	client := NewMockTreeSitterClient()
	a := client.HighlightOnce("a", "go")
	b := client.HighlightOnce("b", "go")

	client.ResolveAllHighlightOnce()
	if client.IsHighlighting() {
		t.Fatal("expected not highlighting after ResolveAllHighlightOnce")
	}

	if got := a.Get(); len(got.Highlights) != 0 {
		t.Fatalf("expected empty highlights, got %+v", got)
	}
	if got := b.Get(); len(got.Highlights) != 0 {
		t.Fatalf("expected empty highlights, got %+v", got)
	}
}

func TestMockTreeSitterClientAutoResolve(t *testing.T) {
	clock := NewManualClock()
	client := NewMockTreeSitterClient(
		MockTreeSitterClientOptions{
			AutoResolveTimeout: ptrFloat(100),
			Clock:              clock,
		},
	)
	client.SetMockResult(HighlightResult{Highlights: []SimpleHighlight{{Start: 1, End: 2, StyleID: 3}}})

	future := client.HighlightOnce("code", "go")
	if !client.IsHighlighting() {
		t.Fatal("expected highlighting before auto-resolve")
	}

	clock.Advance(100)
	if client.IsHighlighting() {
		t.Fatal("expected not highlighting after auto-resolve timeout")
	}

	if got := future.Get(); len(got.Highlights) != 1 || got.Highlights[0].StyleID != 3 {
		t.Fatalf("unexpected auto-resolved result: %+v", got)
	}
}

func TestMockTreeSitterClientResolveIndexOutOfRange(t *testing.T) {
	client := NewMockTreeSitterClient()
	client.HighlightOnce("code", "go")
	// Resolving out-of-range indices must not panic.
	client.ResolveHighlightOnce(5)
	client.ResolveHighlightOnce(-1)
}

func ptrFloat(v float64) *float64 {
	return &v
}

func TestCreateTerminalCapabilitiesDefaults(t *testing.T) {
	caps := CreateTerminalCapabilities(TerminalCapabilitiesOverrides{})
	if caps.KittyKeyboard || caps.RGB || caps.ANSI256 {
		t.Fatal("expected default capabilities to be disabled")
	}
	if caps.Unicode != core.WidthMethodUnicode {
		t.Fatalf("expected default unicode %q, got %q", core.WidthMethodUnicode, caps.Unicode)
	}
	if caps.Multiplexer != core.TerminalMultiplexerNone {
		t.Fatalf("expected default multiplexer %q, got %q", core.TerminalMultiplexerNone, caps.Multiplexer)
	}
	if caps.Terminal.Name != "" || caps.Terminal.Version != "" || caps.Terminal.FromXTVersion {
		t.Fatal("expected empty default terminal info")
	}
}

func TestCreateTerminalCapabilitiesOverrides(t *testing.T) {
	caps := CreateTerminalCapabilities(TerminalCapabilitiesOverrides{
		RGB:          ptrBool(true),
		Unicode:      ptrWidthMethod(core.WidthMethodWCWidth),
		Multiplexer:  ptrMultiplexer(core.TerminalMultiplexerTMux),
		Terminal: &core.TerminalInfo{Name: "xterm", Version: "388", FromXTVersion: true},
	})
	if !caps.RGB {
		t.Fatal("expected RGB override to apply")
	}
	if caps.Unicode != core.WidthMethodWCWidth {
		t.Fatalf("expected Unicode override %q, got %q", core.WidthMethodWCWidth, caps.Unicode)
	}
	if caps.Multiplexer != core.TerminalMultiplexerTMux {
		t.Fatalf("expected Multiplexer override %q, got %q", core.TerminalMultiplexerTMux, caps.Multiplexer)
	}
	if caps.Terminal.Name != "xterm" || caps.Terminal.Version != "388" || !caps.Terminal.FromXTVersion {
		t.Fatalf("expected terminal override, got %+v", caps.Terminal)
	}
}

func TestSetRendererCapabilities(t *testing.T) {
	renderer := core.NewCliRenderer(80, 24)
	defer renderer.Destroy()

	caps := SetRendererCapabilities(renderer, TerminalCapabilitiesOverrides{
		RGB: ptrBool(true),
	})
	if !caps.RGB {
		t.Fatal("expected SetRendererCapabilities result to have RGB")
	}
	if got := renderer.Context().Capabilities; !got.RGB {
		t.Fatal("expected renderer.Context().Capabilities to have RGB")
	}
}

func ptrBool(v bool) *bool { return &v }
func ptrWidthMethod(v core.WidthMethod) *core.WidthMethod { return &v }
func ptrMultiplexer(v core.TerminalMultiplexer) *core.TerminalMultiplexer { return &v }

func TestTestWriteStream(t *testing.T) {
	s := NewTestWriteStream(80, 24)
	if !s.IsTTY || s.Columns != 80 || s.Rows != 24 {
		t.Fatalf("unexpected stream: %+v", s)
	}
	if s.GetColorDepth() != 24 {
		t.Fatal("expected color depth 24")
	}
	n, err := s.Write([]byte("hello"))
	if err != nil || n != 5 {
		t.Fatalf("expected Write to discard and return (5, nil), got (%d, %v)", n, err)
	}
}

func TestRecordingWriteStream(t *testing.T) {
	s := NewRecordingWriteStream(80, 24)
	if s.PendingWrite() {
		t.Fatal("expected no pending write initially")
	}

	if _, err := s.Write([]byte("first")); err != nil {
		t.Fatal(err)
	}
	if s.PendingWrite() {
		t.Fatal("expected no pending write after unheld write")
	}
	if s.Text() != "first" {
		t.Fatalf("expected text 'first', got %q", s.Text())
	}

	s.Hold()
	done := make(chan error)
	go func() {
		_, err := s.Write([]byte("second"))
		done <- err
	}()
	// Wait for the write to be parked (avoids a race with goroutine startup).
	deadline := time.Now().Add(5 * time.Second)
	parked := false
	for time.Now().Before(deadline) {
		if s.PendingWrite() {
			parked = true
			break
		}
		time.Sleep(time.Millisecond)
	}
	if !parked {
		t.Fatal("expected pending write while held")
	}
	s.Release()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if s.PendingWrite() {
		t.Fatal("expected no pending write after release")
	}
	if s.Text() != "firstsecond" {
		t.Fatalf("expected text 'firstsecond', got %q", s.Text())
	}

	s.Clear()
	if s.Text() != "" {
		t.Fatal("expected empty text after clear")
	}
}

func TestCreateTestStdinAndStdout(t *testing.T) {
	stdin := CreateTestStdin()
	buf := make([]byte, 4)
	if n, err := stdin.Read(buf); n != 0 || err != io.EOF {
		t.Fatalf("expected stdin to EOF immediately, got (%d, %v)", n, err)
	}

	stdout := CreateTestStdout(120, 40)
	if stdout.Columns != 120 || stdout.Rows != 40 || !stdout.IsTTY {
		t.Fatalf("unexpected stdout: %+v", stdout)
	}
}
