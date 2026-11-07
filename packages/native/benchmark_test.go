package native

import (
	"testing"
)

func BenchmarkBufferSetCell(b *testing.B) {
	buf := NewOptimizedBuffer(80, 24)

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(80 * 24 * 12))
	for i := 0; i < b.N; i++ {
		x := i % 80
		y := (i / 80) % 24
		buf.SetCell(x, y, 'A', 7, 0, 0)
	}
}

func BenchmarkBufferFillRect(b *testing.B) {
	buf := NewOptimizedBuffer(80, 24)

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(80 * 24 * 12))
	for i := 0; i < b.N; i++ {
		buf.FillRect(0, 0, 80, 24, ' ', 0, 0, 0)
	}
}

func BenchmarkBufferClear(b *testing.B) {
	buf := NewOptimizedBuffer(80, 24)

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(80 * 24 * 12))
	for i := 0; i < b.N; i++ {
		buf.Clear()
	}
}

func BenchmarkBufferAccess(b *testing.B) {
	buf := NewOptimizedBuffer(80, 24)
	// Seed the buffer with data so the snapshot is non-trivial.
	for y := 0; y < 24; y++ {
		for x := 0; x < 80; x++ {
			buf.SetCell(x, y, 'A', 7, 0, 0)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(80 * 24 * 12))
	for i := 0; i < b.N; i++ {
		_ = buf.Access()
	}
}

func BenchmarkYogaCalculateLayout(b *testing.B) {
	yoga := NewYoga()
	root := yoga.CreateNode()
	root.SetWidth(80)
	root.SetHeight(24)

	// Build a nested tree: 10 children, each with 5 grandchildren.
	for i := 0; i < 10; i++ {
		child := yoga.CreateNode()
		child.SetWidth(10)
		child.SetHeight(5)
		root.AddChild(child)
		for j := 0; j < 5; j++ {
			grandchild := yoga.CreateNode()
			grandchild.SetWidth(4)
			grandchild.SetHeight(2)
			child.AddChild(grandchild)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(80 * 24 * 4))
	for i := 0; i < b.N; i++ {
		yoga.CalculateLayout(root, 80, 24)
	}
}

func BenchmarkTextBufferInsert(b *testing.B) {
	tb := NewTextBuffer()
	text := "Hello, World! "

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(len(text)))
	for i := 0; i < b.N; i++ {
		offset := i % (tb.Length() + 1)
		if err := tb.Insert(offset, text); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTextBufferDelete(b *testing.B) {
	content := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. "
	tb := NewTextBufferWithContent(content)

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(len(content)))
	for i := 0; i < b.N; i++ {
		length := tb.Length()
		if length < 10 {
			tb.SetContent(content)
			continue
		}
		// Always delete 5 bytes from the start (position 0), valid when length >= 5.
		if err := tb.Delete(0, 5); err != nil {
			b.Fatal(err)
		}
	}
}

// ============================================================================
// Comprehensive benchmarks (added)
// ============================================================================

// BenchmarkNativeSceneCreateDestroy repeatedly creates and destroys scene
// nodes in a native Scene, exercising scene graph allocation and cleanup.
func BenchmarkNativeSceneCreateDestroy(b *testing.B) {
	scene := NewScene(0)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Create a tree of scene nodes: 1 parent with 8 children.
		parent := scene.CreateNode(SceneNodeKindBox)
		var childIDs [8]uint64
		for j := 0; j < 8; j++ {
			child := scene.CreateNode(SceneNodeKindText)
			childIDs[j] = child.id
			if err := scene.AddChild(parent.id, child.id); err != nil {
				b.Fatal(err)
			}
		}
		// Destroy all children first, then the parent.
		for j := 0; j < 8; j++ {
			if err := scene.DestroyNode(childIDs[j]); err != nil {
				b.Fatal(err)
			}
		}
		scene.DestroyNode(parent.id)
	}
}

// BenchmarkYogaLayoutNested performs Yoga layout on a deeply nested tree
// of layout nodes, measuring layout recalculation cost for deep hierarchies.
func BenchmarkYogaLayoutNested(b *testing.B) {
	yoga := NewYoga()
	root := yoga.CreateNode()
	root.SetWidth(120)
	root.SetHeight(40)

	// Build a deep nested tree: depth 10, each node has 3 children.
	const depth = 10
	const fanout = 3
	var build func(node *YogaNode, level int)
	build = func(node *YogaNode, level int) {
		if level >= depth {
			return
		}
		for i := 0; i < fanout; i++ {
			child := yoga.CreateNode()
			child.SetWidth(30)
			child.SetHeight(8)
			node.AddChild(child)
			build(child, level+1)
		}
	}
	build(root, 0)

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(120 * 40 * 4))
	for i := 0; i < b.N; i++ {
		yoga.CalculateLayout(root, 120, 40)
	}
}

// BenchmarkYogaLayoutWide performs Yoga layout on a very wide shallow tree
// (many siblings), measuring layout cost for wide hierarchies.
func BenchmarkYogaLayoutWide(b *testing.B) {
	yoga := NewYoga()
	root := yoga.CreateNode()
	root.SetWidth(200)
	root.SetHeight(24)

	const childCount = 200
	for i := 0; i < childCount; i++ {
		child := yoga.CreateNode()
		child.SetWidth(8)
		child.SetHeight(1)
		root.AddChild(child)
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(200 * 24 * 4))
	for i := 0; i < b.N; i++ {
		yoga.CalculateLayout(root, 200, 24)
	}
}

// BenchmarkTextBufferLargeOps performs many insert/delete operations on a
// large TextBuffer, simulating heavy editing workloads.
func BenchmarkTextBufferLargeOps(b *testing.B) {
	// Build a large document (~200KB)
	var content []byte
	for i := 0; i < 5000; i++ {
		content = append(content, []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.\n")...)
	}
	tb := NewTextBufferWithContent(string(content))

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(len(content)))
	for i := 0; i < b.N; i++ {
		length := tb.Length()
		if length == 0 {
			length = 1
		}
		offset := i % length
		insertText := " [edit-" + itoa(i) + "] "
		if err := tb.Insert(offset, insertText); err != nil {
			b.Fatal(err)
		}
		// Delete inserted text to keep size stable
		if err := tb.Delete(offset, offset+len(insertText)); err != nil {
			b.Fatal(err)
		}
	}
}

// itoa is a small integer-to-string helper for benchmarks.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var buf [16]byte
	pos := len(buf)
	neg := false
	if i < 0 {
		neg = true
		i = -i
	}
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
