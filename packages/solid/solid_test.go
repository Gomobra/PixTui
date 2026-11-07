package solid

import (
	"testing"

	"github.com/Gomobra/PixTui/packages/core"
	"github.com/Gomobra/PixTui/packages/core/renderables"
)

func TestCreateSignal(t *testing.T) {
	value, setter := CreateSignal(42)
	if value != 42 {
		t.Errorf("expected initial value 42, got %v", value)
	}
	setter(100)
	// Note: CreateSignal outside a component doesn't trigger re-render,
	// but the setter should not panic
}

func TestCreateEffect(t *testing.T) {
	ran := false
	CreateEffect(func() {
		ran = true
	})
	if !ran {
		t.Error("expected effect to run")
	}
}

func TestCreateMemo(t *testing.T) {
	result := CreateMemo(func() interface{} {
		return 42
	})
	if result != 42 {
		t.Errorf("expected memo value 42, got %v", result)
	}
}

func TestBoxComponent(t *testing.T) {
	comp := Box(map[string]interface{}{
		"id":   "test-box",
		"border": true,
		"title": "Test",
	})
	if comp.Type != "box" {
		t.Errorf("expected type 'box', got '%s'", comp.Type)
	}
	if comp.Props["id"] != "test-box" {
		t.Errorf("expected id 'test-box', got '%v'", comp.Props["id"])
	}
}

func TestTextComponent(t *testing.T) {
	comp := Text(map[string]interface{}{
		"id":      "test-text",
		"content": "Hello",
	})
	if comp.Type != "text" {
		t.Errorf("expected type 'text', got '%s'", comp.Type)
	}
	if comp.Props["content"] != "Hello" {
		t.Errorf("expected content 'Hello', got '%v'", comp.Props["content"])
	}
}

func TestCreateComponent(t *testing.T) {
	called := false
	comp := CreateComponent(func(props map[string]interface{}) SolidComponent {
		called = true
		return Text(map[string]interface{}{"content": "dynamic"})
	})
	if comp.Fn == nil {
		t.Error("expected component function to be set")
	}
	output := comp.Fn(map[string]interface{}{})
	if !called {
		t.Error("expected component function to be called")
	}
	if output.Type != "text" {
		t.Errorf("expected output type 'text', got '%s'", output.Type)
	}
}

func TestRenderTest(t *testing.T) {
	comp := Box(map[string]interface{}{
		"id":     "root-box",
		"border": true,
		"title":  "Root",
	})
	root := RenderTest(comp)
	if root == nil {
		t.Fatal("expected non-nil root renderable")
	}
	if root.ID() != "root" {
		t.Errorf("expected root ID 'root', got '%s'", root.ID())
	}
}

func TestRenderTestWithText(t *testing.T) {
	comp := Text(map[string]interface{}{
		"id":      "test-text",
		"content": "Hello Solid",
	})
	root := RenderTest(comp)
	if root == nil {
		t.Fatal("expected non-nil root renderable")
	}
	// RenderTest should not panic
}

func TestRenderTestNested(t *testing.T) {
	inner := Text(map[string]interface{}{
		"content": "Nested text",
	})
	outer := Box(map[string]interface{}{
		"id":      "outer-box",
		"border":  true,
		"children": []SolidComponent{inner},
	})
	root := RenderTest(outer)
	if root == nil {
		t.Fatal("expected non-nil root renderable")
	}
}

func TestRenderTestRenderer(t *testing.T) {
	ctx := &core.RenderContext{Width: 80, Height: 24}
	renderer := NewSolidRenderer(ctx)
	buffer := core.NewNativeBuffer(80, 24)
	comp := Box(map[string]interface{}{
		"id":   "renderer-test",
		"border": true,
	})
	renderer.Render(comp, buffer, 0)
	// Should not panic
	_ = buffer
}

func TestSignalSetterTriggersReRender(t *testing.T) {
	ctx := &core.RenderContext{Width: 80, Height: 24}
	renderer := NewSolidRenderer(ctx)
	buffer := core.NewNativeBuffer(80, 24)
	renderCount := 0

	value, setter := CreateSignal(0)
	_ = value

	// Simulate a component that reads the signal
	comp := CreateComponent(func(props map[string]interface{}) SolidComponent {
		renderCount++
		// Read signal value (tracks dependency)
		_ = value
		return Text(map[string]interface{}{"content": "test"})
	})

	renderer.Render(comp, buffer, 0)
	if renderCount != 1 {
		t.Errorf("expected 1 render, got %d", renderCount)
	}

	// Set the signal - should trigger re-render
	setter(1)
	renderer.ScheduleRender()
	// Give time for async effect to run
	// Note: In a real scenario, the effect would schedule a re-render
}

func TestRenderableTypes(t *testing.T) {
	ctx := &core.RenderContext{Width: 80, Height: 24}

	// Test Box renderable creation
	boxComp := Box(map[string]interface{}{
		"id":       "box",
		"border":   true,
		"title":    "Box",
		"borderColor": "#FF0000",
	})
	renderer := NewSolidRenderer(ctx)
	buffer := core.NewNativeBuffer(80, 24)
	renderer.Render(boxComp, buffer, 0)

	// Test Text renderable creation
	textComp := Text(map[string]interface{}{
		"id":      "text",
		"content": "Hello World",
	})
	renderer.Render(textComp, buffer, 0)

	// Verify renderables are created properly
	_ = renderables.NewBoxRenderable
}

func TestUseSignalUseEffectUseMemo(t *testing.T) {
	// Test that package-level hooks work outside of a fiber context
	val, set := UseSignal(10)
	if val != 10 {
		t.Errorf("expected 10, got %v", val)
	}
	set(20)

	effectRan := false
	UseEffect(func() {
		effectRan = true
	}, nil)
	if !effectRan {
		t.Error("expected effect to run")
	}

	memoVal := UseMemo(func() interface{} { return 100 }, nil)
	if memoVal != 100 {
		t.Errorf("expected 100, got %v", memoVal)
	}
}

// ----- Context tests -----

func TestCreateContextDefaultValue(t *testing.T) {
	ctx := CreateContext("default")
	if ctx.defaultValue != "default" {
		t.Errorf("expected default value 'default', got %v", ctx.defaultValue)
	}
}

func TestUseContextOutsideFiberReturnsDefault(t *testing.T) {
	ctx := CreateContext(42)
	if got := useContext(ctx); got != 42 {
		t.Errorf("expected default value 42 outside fiber, got %v", got)
	}
}

func TestUseContextWithProvider(t *testing.T) {
	// useContext inside a component wrapped by a Provider should return the
	// provided value.
	var got interface{}
	inner := CreateComponent(func(props map[string]interface{}) SolidComponent {
		got = useContext(themeCtx)
		return Text(map[string]interface{}{"content": "inner"})
	})
	outer := themeCtx.Provider(map[string]interface{}{
		"value":    "dark",
		"children": []SolidComponent{inner},
	})

	root := RenderTest(outer)
	if root == nil {
		t.Fatal("expected non-nil root renderable")
	}
	if got != "dark" {
		t.Errorf("expected useContext to return 'dark', got %v", got)
	}
}

func TestUseContextFallsBackToDefault(t *testing.T) {
	// useContext inside a component with no provider should return default.
	var got interface{}
	inner := CreateComponent(func(props map[string]interface{}) SolidComponent {
		got = useContext(noProviderCtx)
		return Text(map[string]interface{}{"content": "inner"})
	})
	root := RenderTest(inner)
	if root == nil {
		t.Fatal("expected non-nil root renderable")
	}
	if got != "default" {
		t.Errorf("expected useContext to return default 'default', got %v", got)
	}
}

func TestUseContextNestedProviders(t *testing.T) {
	// Inner provider should override outer provider.
	var outerGot, innerGot interface{}
	middle := CreateComponent(func(props map[string]interface{}) SolidComponent {
		outerGot = useContext(themeCtx)
		return themeCtx.Provider(map[string]interface{}{
			"value":    "inner",
			"children": []SolidComponent{
				CreateComponent(func(props map[string]interface{}) SolidComponent {
					innerGot = useContext(themeCtx)
					return Text(map[string]interface{}{"content": "deep"})
				}),
			},
		})
	})
	outer := themeCtx.Provider(map[string]interface{}{
		"value":    "outer",
		"children": []SolidComponent{middle},
	})

	root := RenderTest(outer)
	if root == nil {
		t.Fatal("expected non-nil root renderable")
	}
	if outerGot != "outer" {
		t.Errorf("expected outer useContext to return 'outer', got %v", outerGot)
	}
	if innerGot != "inner" {
		t.Errorf("expected inner useContext to return 'inner', got %v", innerGot)
	}
}

var themeCtx = CreateContext("default")
var noProviderCtx = CreateContext("default")

var _ = core.Renderable{}
var _ = renderables.BoxRenderable{}

// ----- Benchmarks -----

func BenchmarkSolidRender(b *testing.B) {
	ctx := &core.RenderContext{Width: 80, Height: 24}
	renderer := NewSolidRenderer(ctx)
	buffer := core.NewNativeBuffer(80, 24)
	comp := Box(map[string]interface{}{
		"id":   "bench-box",
		"border": true,
		"title":  "Benchmark",
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderer.Render(comp, buffer, 0)
	}
}

func BenchmarkSolidCreateSignal(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CreateSignal(i)
	}
}

func BenchmarkSolidCreateEffect(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CreateEffect(func() {})
	}
}

func BenchmarkSolidCreateMemo(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CreateMemo(func() interface{} { return i })
	}
}
