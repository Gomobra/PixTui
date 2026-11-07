package core

import (
	"testing"
)

func TestExtmarkManagerAddGet(t *testing.T) {
	m := NewExtmarkManager()
	id := m.Add(&Extmark{
		Type:      ExtmarkTypeSelection,
		StartLine: 0,
		StartCol:  0,
		EndLine:   1,
		EndCol:    5,
		Priority:  10,
	})
	if id == "" {
		t.Fatal("expected non-empty extmark ID")
	}

	mark, ok := m.Get(id)
	if !ok {
		t.Fatal("expected to find extmark")
	}
	if mark.Type != ExtmarkTypeSelection {
		t.Errorf("expected selection type, got %s", mark.Type)
	}
}

func TestExtmarkManagerRemove(t *testing.T) {
	m := NewExtmarkManager()
	id := m.Add(&Extmark{Type: ExtmarkTypeHighlight})
	m.Remove(id)
	if _, ok := m.Get(id); ok {
		t.Error("expected extmark to be removed")
	}
}

func TestExtmarkManagerClear(t *testing.T) {
	m := NewExtmarkManager()
	m.Add(&Extmark{Type: ExtmarkTypeSelection})
	m.Add(&Extmark{Type: ExtmarkTypeHighlight})
	m.Clear()
	if m.Count() != 0 {
		t.Error("expected 0 extmarks after clear")
	}
}

func TestExtmarkManagerCount(t *testing.T) {
	m := NewExtmarkManager()
	if m.Count() != 0 {
		t.Error("expected 0 extmarks initially")
	}
	m.Add(&Extmark{Type: ExtmarkTypeSelection})
	if m.Count() != 1 {
		t.Error("expected 1 extmark after add")
	}
	m.Add(&Extmark{Type: ExtmarkTypeHighlight})
	if m.Count() != 2 {
		t.Error("expected 2 extmarks after second add")
	}
}

func TestExtmarkOverlaps(t *testing.T) {
	mark := &Extmark{
		StartLine: 0,
		StartCol:  0,
		EndLine:   0,
		EndCol:    10,
	}

	if !mark.Overlaps(0, 5, 0, 7) {
		t.Error("expected overlap for contained range")
	}
	if !mark.Overlaps(0, 0, 0, 5) {
		t.Error("expected overlap for same start")
	}
	if mark.Overlaps(0, 10, 0, 20) {
		t.Error("expected no overlap for adjacent range")
	}
	if mark.Overlaps(1, 0, 1, 5) {
		t.Error("expected no overlap for different line")
	}
}

func TestExtmarkContains(t *testing.T) {
	mark := &Extmark{
		StartLine: 0,
		StartCol:  2,
		EndLine:   0,
		EndCol:    8,
	}

	if !mark.Contains(0, 5) {
		t.Error("expected position to be contained")
	}
	if mark.Contains(0, 1) {
		t.Error("expected position before start to not be contained")
	}
	if mark.Contains(0, 8) {
		t.Error("expected position at end to not be contained")
	}
	if mark.Contains(1, 0) {
		t.Error("expected position on different line to not be contained")
	}
}

func TestExtmarkSetAttributes(t *testing.T) {
	m := NewExtmarkManager()
	id := m.Add(&Extmark{Type: ExtmarkTypeHighlight})
	if !m.SetAttributes(id, TextAttributeBold) {
		t.Error("expected SetAttributes to succeed")
	}
	mark, _ := m.Get(id)
	if mark.Attributes != TextAttributeBold {
		t.Error("expected bold attribute to be set")
	}
	if m.SetAttributes("nonexistent", TextAttributeBold) {
		t.Error("expected SetAttributes to fail for nonexistent id")
	}
}

func TestExtmarkSetStyle(t *testing.T) {
	m := NewExtmarkManager()
	id := m.Add(&Extmark{Type: ExtmarkTypeSelection})
	fg := RGBA{R: 255, G: 0, B: 0, A: 255}
	bg := RGBA{R: 0, G: 0, B: 255, A: 255}
	if !m.SetStyle(id, fg, bg) {
		t.Error("expected SetStyle to succeed")
	}
	mark, _ := m.Get(id)
	if mark.Foreground != fg {
		t.Error("expected foreground to be set")
	}
	if mark.Background != bg {
		t.Error("expected background to be set")
	}
}

func TestExtmarkSetRange(t *testing.T) {
	m := NewExtmarkManager()
	id := m.Add(&Extmark{Type: ExtmarkTypeHighlight})
	if !m.SetRange(id, 5, 10, 5, 20) {
		t.Error("expected SetRange to succeed")
	}
	mark, _ := m.Get(id)
	if mark.StartLine != 5 || mark.StartCol != 10 || mark.EndLine != 5 || mark.EndCol != 20 {
		t.Error("expected range to be updated")
	}
}

func TestExtmarkListSortByPriority(t *testing.T) {
	m := NewExtmarkManager()
	m.Add(&Extmark{Type: ExtmarkTypeHighlight, Priority: 1})
	m.Add(&Extmark{Type: ExtmarkTypeSelection, Priority: 10})
	m.Add(&Extmark{Type: ExtmarkTypeError, Priority: 5})

	marks := m.List()
	if len(marks) != 3 {
		t.Fatalf("expected 3 marks, got %d", len(marks))
	}
	if marks[0].Priority != 10 {
		t.Error("expected highest priority first")
	}
	if marks[2].Priority != 1 {
		t.Error("expected lowest priority last")
	}
}

func TestExtmarkLayer(t *testing.T) {
	buffer := NewNativeBuffer(10, 5)
	buffer.SetCell(0, 0, 'A', 0, 0, 0)
	buffer.SetCell(1, 0, 'B', 0, 0, 0)
	buffer.SetCell(2, 0, 'C', 0, 0, 0)

	m := NewExtmarkManager()
	m.Add(&Extmark{
		Type:      ExtmarkTypeHighlight,
		StartLine: 0,
		StartCol:  1,
		EndLine:   0,
		EndCol:    2,
		Attributes: TextAttributeBold,
	})

	layer := NewExtmarkLayer(m, buffer)
	layer.Apply()

	// Cell at (1,0) should have bold attribute
	_, _, _, attrs, _ := buffer.GetCell(1, 0)
	if attrs&TextAttributeBold == 0 {
		t.Error("expected bold attribute on highlighted cell")
	}

	// Cell at (0,0) should not have bold attribute
	_, _, _, attrs2, _ := buffer.GetCell(0, 0)
	if attrs2&TextAttributeBold != 0 {
		t.Error("expected no bold attribute on non-highlighted cell")
	}
}
