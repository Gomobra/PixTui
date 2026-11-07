package react

import (
	"testing"
)

func TestJSXCreateElement(t *testing.T) {
	el := jsx("box", JSXProps{"border": true})
	if el.Type != "box" {
		t.Errorf("expected type 'box', got %v", el.Type)
	}
	if el.Props["border"] != true {
		t.Error("expected border prop to be true")
	}
}

func TestJSXComponent(t *testing.T) {
	fn := func(props map[string]interface{}) ReactElement {
		return CreateElement("text", map[string]interface{}{"content": "hi"})
	}
	el := jsxComponent(fn, JSXProps{"x": 1})
	if _, ok := el.Type.(ComponentFunc); !ok {
		t.Error("expected component function as type")
	}
	if el.Props["x"] != 1 {
		t.Error("expected x prop to be 1")
	}
}

func TestJSXFragment(t *testing.T) {
	el := jsxFragment(jsx("text", JSXProps{"content": "a"}), jsx("text", JSXProps{"content": "b"}))
	if el.Type != Fragment {
		t.Errorf("expected fragment type, got %v", el.Type)
	}
	if len(el.Children) != 2 {
		t.Errorf("expected 2 children, got %d", len(el.Children))
	}
}

func TestJSXsAlias(t *testing.T) {
	a := jsx("box", JSXProps{"id": "a"})
	b := jsxs("box", JSXProps{"id": "b"})
	if a.Type != b.Type {
		t.Error("expected jsxs to produce same type as jsx")
	}
}

func TestJSXDEV(t *testing.T) {
	el := jsxDEV("box", JSXProps{"id": "dev"})
	if el.Type != "box" {
		t.Errorf("expected type 'box', got %v", el.Type)
	}
	if el.Props["id"] != "dev" {
		t.Error("expected id prop 'dev'")
	}
}

func TestFragmentJSXConstant(t *testing.T) {
	if FragmentJSX != Fragment {
		t.Error("expected FragmentJSX to equal Fragment")
	}
}

func TestJSXToReactElement(t *testing.T) {
	el := jsx("box", JSXProps{"title": "T"}, jsx("text", JSXProps{"content": "hello"}))
	re := el.ToReactElement()
	if re.Type != "box" {
		t.Errorf("expected type 'box', got %v", re.Type)
	}
	if len(re.Children) != 1 {
		t.Errorf("expected 1 child, got %d", len(re.Children))
	}
	if re.Children[0].Type != "text" {
		t.Errorf("expected child type 'text', got %v", re.Children[0].Type)
	}
}
