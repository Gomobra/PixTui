package react

import (
	"testing"

	"github.com/Gomobra/PixTui/packages/core"
)

func TestReactRendererCreate(t *testing.T) {
	ctx := &core.RenderContext{
		Width:  80,
		Height: 24,
	}
	renderer := NewReactRenderer(ctx)
	if renderer == nil {
		t.Fatal("expected non-nil renderer")
	}
}

func TestCreateElement(t *testing.T) {
	element := CreateElement("box", map[string]interface{}{
		"border": true,
	})
	if element.Type != "box" {
		t.Errorf("expected type 'box', got '%s'", element.Type)
	}
	if element.Props["border"] != true {
		t.Error("expected border prop to be true")
	}
}

func TestCreateRef(t *testing.T) {
	ref := CreateRef()
	if ref == nil {
		t.Fatal("expected non-nil ref")
	}
}

func TestCreateContext(t *testing.T) {
	ctx := CreateContext("default")
	if ctx.DefaultValue != "default" {
		t.Errorf("expected default value 'default', got '%v'", ctx.DefaultValue)
	}
}

func TestUseState(t *testing.T) {
	value, setValue := UseState(0)
	if value != 0 {
		t.Errorf("expected initial value 0, got %v", value)
	}
	setValue(42)
}

func TestUseEffect(t *testing.T) {
	called := false
	UseEffect(func() func() {
		called = true
		return nil
	}, nil)
	if !called {
		t.Error("expected effect to be called")
	}
}

func TestUseMemo(t *testing.T) {
	result := UseMemo(func() interface{} {
		// memo computation
		return nil
	}, nil)
	if result != nil {
		t.Error("expected nil result from memo with no return")
	}
}

func TestBox(t *testing.T) {
	element := CreateElement("box", map[string]interface{}{
		"border": true,
		"title":  "Test",
	})
	if element.Type != "box" {
		t.Errorf("expected type 'box', got '%s'", element.Type)
	}
}

func TestText(t *testing.T) {
	element := CreateElement("text", map[string]interface{}{
		"content": "Hello",
	})
	if element.Type != "text" {
		t.Errorf("expected type 'text', got '%s'", element.Type)
	}
}

func TestInput(t *testing.T) {
	element := CreateElement("input", map[string]interface{}{
		"value": "test",
	})
	if element.Type != "input" {
		t.Errorf("expected type 'input', got '%s'", element.Type)
	}
}

func TestSelect(t *testing.T) {
	element := CreateElement("select", map[string]interface{}{
		"options": []string{"a", "b"},
	})
	if element.Type != "select" {
		t.Errorf("expected type 'select', got '%s'", element.Type)
	}
}

func TestRenderTest(t *testing.T) {
	element := CreateElement("box", map[string]interface{}{
		"border": true,
	})
	buffer := core.NewNativeBuffer(80, 24)
	renderer := NewReactRenderer(&core.RenderContext{})
	renderer.Render(element, buffer, 0)
}


// lifecycleLog mencatat pemanggilan effect/memo/callback untuk menguji
// semantik React hook (mount sekali, rerun hanya jika deps berubah, cleanup
// sebelum rerun dan saat unmount).
type lifecycleLog struct {
	runs           int
	cleanups       int
	layoutRuns     int
	layoutCleanups int
	memoCalls      int
	cbCalls        int
	cbLast         func()
	cbMarker       interface{}
}

// hookComp mendaftarkan layout effect, passive effect, memo dan callback yang
// semua dikunci oleh nilai dep.
func hookComp(props map[string]interface{}) ReactElement {
	dep := props["dep"]
	log := props["log"].(*lifecycleLog)

	UseLayoutEffect(func() func() {
		log.layoutRuns++
		return func() { log.layoutCleanups++ }
	}, []interface{}{dep})

	UseEffect(func() func() {
		log.runs++
		return func() { log.cleanups++ }
	}, []interface{}{dep})

	_ = UseMemo(func() interface{} {
		log.memoCalls++
		return dep
	}, []interface{}{dep})

	// Tangkap dep menurut nilai pada saat callback dibuat agar identitasnya
	// bisa dibedakan secara perilaku (Go tidak membolehkan perbandingan
	// fungsi-fungsi kecuali ke nil).
	cbDep := dep
	cb := UseCallback(func() {
		log.cbCalls++
		log.cbMarker = cbDep
	}, []interface{}{dep})
	log.cbLast = cb

	return CreateElement("box", nil)
}

func TestHookLifecycle(t *testing.T) {
	ctx := &core.RenderContext{Width: 80, Height: 24}
	renderer := NewReactRenderer(ctx)
	buffer := core.NewNativeBuffer(80, 24)

	log := &lifecycleLog{}
	makeEl := func(dep string) ReactElement {
		return CreateComponent(hookComp, map[string]interface{}{
			"dep": dep,
			"log": log,
		})
	}

	// Mount: passive effect berjalan sekali (terunda, setelah commit/paint).
	renderer.Render(makeEl("a"), buffer, 0)
	if log.runs != 1 {
		t.Fatalf("useEffect: expected 1 run after mount, got %d", log.runs)
	}
	if log.cleanups != 0 {
		t.Fatalf("useEffect: expected no cleanup after mount, got %d", log.cleanups)
	}
	if log.layoutRuns != 1 {
		t.Fatalf("useLayoutEffect: expected 1 run after mount, got %d", log.layoutRuns)
	}
	if log.memoCalls != 1 {
		t.Fatalf("useMemo: expected 1 computation on mount, got %d", log.memoCalls)
	}

	// Sama deps: semuanya no-op (tidak dijalankan ulang).
	renderer.Render(makeEl("a"), buffer, 0)
	if log.runs != 1 {
		t.Fatalf("useEffect: should not re-run with same deps, got %d", log.runs)
	}
	if log.cleanups != 0 {
		t.Fatalf("useEffect: expected no cleanup with same deps, got %d", log.cleanups)
	}
	if log.layoutRuns != 1 {
		t.Fatalf("useLayoutEffect: should not re-run with same deps, got %d", log.layoutRuns)
	}
	if log.layoutCleanups != 0 {
		t.Fatalf("useLayoutEffect: expected no cleanup with same deps, got %d", log.layoutCleanups)
	}
	if log.memoCalls != 1 {
		t.Fatalf("useMemo: should be cached with same deps, got %d", log.memoCalls)
	}
	if log.cbLast == nil {
		t.Fatal("useCallback: expected a stored callback")
	}

	// Deps berubah: cleanup lama dijalankan sebelum efek baru; memo dihitung ulang.
	renderer.Render(makeEl("b"), buffer, 0)
	if log.runs != 2 {
		t.Fatalf("useEffect: should re-run on deps change, got %d", log.runs)
	}
	if log.cleanups != 1 {
		t.Fatalf("useEffect: previous cleanup should run before re-run, got %d", log.cleanups)
	}
	if log.layoutRuns != 2 {
		t.Fatalf("useLayoutEffect: should re-run on deps change, got %d", log.layoutRuns)
	}
	if log.layoutCleanups != 1 {
		t.Fatalf("useLayoutEffect: previous cleanup should run before re-run, got %d", log.layoutCleanups)
	}
	if log.memoCalls != 2 {
		t.Fatalf("useMemo: should recompute on deps change, got %d", log.memoCalls)
	}

	// Unmount: ganti root agar fiber komponen dilepas; semua cleanup dijalankan.
	renderer.Render(CreateElement("text", map[string]interface{}{"content": "bye"}), buffer, 0)
	if log.cleanups != 2 {
		t.Fatalf("useEffect: cleanup should run on unmount, got %d", log.cleanups)
	}
	if log.layoutCleanups != 2 {
		t.Fatalf("useLayoutEffect: cleanup should run on unmount, got %d", log.layoutCleanups)
	}
}

func TestHookMemoCallbackIdentity(t *testing.T) {
	ctx := &core.RenderContext{Width: 80, Height: 24}
	renderer := NewReactRenderer(ctx)
	buffer := core.NewNativeBuffer(80, 24)

	log := &lifecycleLog{}
	makeEl := func(dep string) ReactElement {
		return CreateComponent(hookComp, map[string]interface{}{
			"dep": dep,
			"log": log,
		})
	}

	// Mount: callback dibuat untuk dep "a" dan disimpan.
	renderer.Render(makeEl("a"), buffer, 0)
	log.cbLast()
	if log.cbMarker != "a" {
		t.Fatalf("useCallback: expected callback created with dep 'a', got %v", log.cbMarker)
	}
	if log.memoCalls != 1 {
		t.Fatalf("useMemo: expected 1 computation on mount, got %d", log.memoCalls)
	}

	// Sama deps: callback yang sama dipakai ulang (identitas stabil).
	renderer.Render(makeEl("a"), buffer, 0)
	log.cbLast()
	if log.cbMarker != "a" {
		t.Fatalf("useCallback: callback identity should be stable with same deps, got %v", log.cbMarker)
	}
	if log.memoCalls != 1 {
		t.Fatalf("useMemo: should be cached with same deps, got %d", log.memoCalls)
	}

	// Deps berubah: callback baru dibuat ulang.
	renderer.Render(makeEl("b"), buffer, 0)
	log.cbLast()
	if log.cbMarker != "b" {
		t.Fatalf("useCallback: callback should be recreated with changed deps, got %v", log.cbMarker)
	}
	if log.memoCalls != 2 {
		t.Fatalf("useMemo: should recompute on deps change, got %d", log.memoCalls)
	}
}

var capturedRef *Ref

func refCaptureComp(props map[string]interface{}) ReactElement {
	r := UseRef(props["value"])
	capturedRef = r
	r.Current = props["value"]
	return CreateElement("box", nil)
}

func TestUseRefPersist(t *testing.T) {
	ctx := &core.RenderContext{Width: 80, Height: 24}
	renderer := NewReactRenderer(ctx)
	buffer := core.NewNativeBuffer(80, 24)

	renderer.Render(CreateComponent(refCaptureComp, map[string]interface{}{"value": 1}), buffer, 0)
	first := capturedRef
	renderer.Render(CreateComponent(refCaptureComp, map[string]interface{}{"value": 2}), buffer, 0)
	if capturedRef != first {
		t.Fatal("useRef: should return the same ref object across renders")
	}
	if capturedRef.Current != 2 {
		t.Fatalf("useRef: expected Current to be 2, got %v", capturedRef.Current)
	}
}
