package solid

import (
	"testing"
)

func TestPortalComponent(t *testing.T) {
	comp := Portal(map[string]interface{}{
		"children": []SolidComponent{Text(map[string]interface{}{"content": "portal"})},
	})
	if comp.Fn == nil {
		t.Error("expected Portal to create a component function")
	}
	out := comp.Fn(nil)
	if out.Type != "text" {
		t.Errorf("expected output type 'text', got %s", out.Type)
	}
}

func TestPortalWithRef(t *testing.T) {
	ref := &RenderableRef{}
	comp := Portal(map[string]interface{}{
		"mount":   ref,
		"children": []SolidComponent{Box(map[string]interface{}{"id": "portal-box"})},
	})
	out := comp.Fn(nil)
	if ref.Value == nil {
		t.Error("expected ref to receive the component output")
	}
	if out.Type != "box" {
		t.Errorf("expected output type 'box', got %s", out.Type)
	}
}

func TestDynamicStringComponent(t *testing.T) {
	comp := Dynamic(map[string]interface{}{
		"component": "text",
		"content":   "dynamic text",
	})
	if comp.Type != "text" {
		t.Errorf("expected type 'text', got %s", comp.Type)
	}
	if comp.Props["content"] != "dynamic text" {
		t.Errorf("expected content 'dynamic text', got %v", comp.Props["content"])
	}
}

func TestDynamicFuncComponent(t *testing.T) {
	fn := func(props map[string]interface{}) SolidComponent {
		return Box(map[string]interface{}{"id": "dynamic-box"})
	}
	comp := Dynamic(map[string]interface{}{
		"component": fn,
		"extra":     "prop",
	})
	if comp.Type != "box" {
		t.Errorf("expected type 'box', got %s", comp.Type)
	}
	if comp.Props["id"] != "dynamic-box" {
		t.Errorf("expected id 'dynamic-box', got %v", comp.Props["id"])
	}
}

func TestCreateDynamic(t *testing.T) {
	comp := createDynamic("input", map[string]interface{}{
		"value":    "test",
		"component": "input",
	})
	if comp.Type != "input" {
		t.Errorf("expected type 'input', got %s", comp.Type)
	}
	if comp.Props["value"] != "test" {
		t.Errorf("expected value 'test', got %v", comp.Props["value"])
	}
}

func TestCreateDynamicFunc(t *testing.T) {
	fn := func(props map[string]interface{}) SolidComponent {
		return Text(map[string]interface{}{"content": "from-func"})
	}
	comp := createDynamic(fn, nil)
	if comp.Type != "text" {
		t.Errorf("expected type 'text', got %s", comp.Type)
	}
	if comp.Props["content"] != "from-func" {
		t.Errorf("expected content 'from-func', got %v", comp.Props["content"])
	}
}

func TestDynamicUnknownType(t *testing.T) {
	comp := Dynamic(map[string]interface{}{
		"component": 42,
	})
	if comp.Type != "int" {
		t.Errorf("expected type 'int' for unknown component, got %s", comp.Type)
	}
}
