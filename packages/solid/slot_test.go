package solid

import (
	"testing"
)

func TestSlotRegistryRegisterResolve(t *testing.T) {
	reg := NewSlotRegistry[SolidComponent]()
	reg.Register("my-slot", "plugin-a", func(ctx interface{}, props map[string]interface{}) SolidComponent {
		return Box(map[string]interface{}{"id": "plugin-a-box"})
	})
	reg.Register("my-slot", "plugin-b", func(ctx interface{}, props map[string]interface{}) SolidComponent {
		return Text(map[string]interface{}{"content": "plugin-b"})
	})

	entries := reg.ResolveEntries("my-slot")
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].ID != "plugin-a" {
		t.Errorf("expected first entry id 'plugin-a', got %s", entries[0].ID)
	}
	if entries[1].ID != "plugin-b" {
		t.Errorf("expected second entry id 'plugin-b', got %s", entries[1].ID)
	}

	empty := reg.ResolveEntries("nonexistent")
	if len(empty) != 0 {
		t.Errorf("expected 0 entries for nonexistent slot, got %d", len(empty))
	}
}

func TestSlotAppendMode(t *testing.T) {
	reg := NewSlotRegistry[SolidComponent]()
	reg.Register("slot1", "p1", func(ctx interface{}, props map[string]interface{}) SolidComponent {
		return Box(map[string]interface{}{"id": "slot1-p1"})
	})

	comp := Slot(SlotProps{
		Registry: reg,
		Name:     "slot1",
		Mode:     SlotModeAppend,
		Children: []SolidComponent{Text(map[string]interface{}{"content": "child"})},
	})
	if comp.Type != "fragment" {
		t.Errorf("expected fragment in append mode, got %s", comp.Type)
	}
	children, _ := comp.Props["children"].([]SolidComponent)
	if len(children) < 1 {
		t.Errorf("expected at least one child in fragment, got %d", len(children))
	}
}

func TestSlotSingleWinnerMode(t *testing.T) {
	reg := NewSlotRegistry[SolidComponent]()
	reg.Register("slot2", "winner", func(ctx interface{}, props map[string]interface{}) SolidComponent {
		return Box(map[string]interface{}{"id": "slot2-winner"})
	})

	comp := Slot(SlotProps{
		Registry: reg,
		Name:     "slot2",
		Mode:     SlotModeSingleWinner,
	})
	if comp.Type != "box" {
		t.Errorf("expected box in single_winner mode, got %s", comp.Type)
	}
	if comp.Props["id"] != "slot2-winner" {
		t.Errorf("expected id 'slot2-winner', got %v", comp.Props["id"])
	}
}

func TestSlotReplaceMode(t *testing.T) {
	reg := NewSlotRegistry[SolidComponent]()
	reg.Register("slot3", "r1", func(ctx interface{}, props map[string]interface{}) SolidComponent {
		return Text(map[string]interface{}{"content": "replaced"})
	})

	comp := Slot(SlotProps{
		Registry: reg,
		Name:     "slot3",
		Mode:     SlotModeReplace,
		Children: []SolidComponent{Text(map[string]interface{}{"content": "hidden"})},
	})
	if comp.Type != "text" {
		t.Errorf("expected text in replace mode, got %s", comp.Type)
	}
	if comp.Props["content"] != "replaced" {
		t.Errorf("expected content 'replaced', got %v", comp.Props["content"])
	}
}

func TestSlotNoEntriesFallsBackToChildren(t *testing.T) {
	reg := NewSlotRegistry[SolidComponent]()

	comp := Slot(SlotProps{
		Registry: reg,
		Name:     "empty-slot",
		Children: []SolidComponent{Box(map[string]interface{}{"id": "fallback"})},
	})
	if comp.Type != "box" {
		t.Errorf("expected fallback box, got %s", comp.Type)
	}
	if comp.Props["id"] != "fallback" {
		t.Errorf("expected id 'fallback', got %v", comp.Props["id"])
	}
}

func TestCreateSlotBound(t *testing.T) {
	reg := NewSlotRegistry[SolidComponent]()
	reg.Register("bound-slot", "b1", func(ctx interface{}, props map[string]interface{}) SolidComponent {
		return Text(map[string]interface{}{"content": "bound"})
	})

	bound := createSlot(reg, "bound-slot", SlotModeAppend)
	comp := bound(Text(map[string]interface{}{"content": "child"}))
	if comp.Type != "fragment" {
		t.Errorf("expected fragment, got %s", comp.Type)
	}
	children, _ := comp.Props["children"].([]SolidComponent)
	if len(children) < 2 {
		t.Errorf("expected at least 2 children in fragment, got %d", len(children))
	}
}

func TestCreateSolidSlotRegistry(t *testing.T) {
	reg := createSolidSlotRegistry()
	if reg == nil {
		t.Fatal("expected non-nil registry")
	}
	reg.Register("r", "x", func(ctx interface{}, props map[string]interface{}) SolidComponent {
		return SolidComponent{}
	})
	if len(reg.ResolveEntries("r")) != 1 {
		t.Error("expected 1 entry")
	}
}

func TestSlotModeString(t *testing.T) {
	if SlotModeAppend.String() != "append" {
		t.Error("expected 'append'")
	}
	if SlotModeSingleWinner.String() != "single_winner" {
		t.Error("expected 'single_winner'")
	}
	if SlotModeReplace.String() != "replace" {
		t.Error("expected 'replace'")
	}
}
