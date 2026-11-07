package native

// Tree-sitter syntax highlighting integration for PixTui.
// Provides tree-sitter parser/wrapper for Code renderable syntax highlighting.
// Mirrors packages/native/src/tree-sitter.zig from the original repo.

// #cgo CFLAGS: -I/data/data/com.termux/files/usr/include
// #cgo LDFLAGS: -L/data/data/com.termux/files/usr/lib -ltree-sitter
// #include <tree_sitter/api.h>
import "C"
import (
	"fmt"
	"unsafe"
)

// TreeSitterLanguage represents a tree-sitter language grammar.
type TreeSitterLanguage struct {
	language *C.TSLanguage
}

// NewTreeSitterLanguage creates a new language wrapper from a TSLanguage pointer.
// The language function is provided by tree-sitter grammar bindings (e.g. tree_sitter_go()).
func NewTreeSitterLanguage(langFn func() unsafe.Pointer) *TreeSitterLanguage {
	return &TreeSitterLanguage{
		language: (*C.TSLanguage)(langFn()),
	}
}

// Pointer returns the underlying C TSLanguage pointer.
func (l *TreeSitterLanguage) Pointer() *C.TSLanguage {
	return l.language
}

// TreeSitterParser is a tree-sitter parser wrapper.
type TreeSitterParser struct {
	parser *C.TSParser
}

// NewTreeSitterParser creates a new tree-sitter parser.
func NewTreeSitterParser() *TreeSitterParser {
	p := C.ts_parser_new()
	return &TreeSitterParser{parser: p}
}

// Destroy frees the parser resources.
func (p *TreeSitterParser) Destroy() {
	if p.parser != nil {
		C.ts_parser_delete(p.parser)
		p.parser = nil
	}
}

// SetLanguage sets the language for the parser.
func (p *TreeSitterParser) SetLanguage(lang *TreeSitterLanguage) bool {
	if p.parser == nil {
		return false
	}
	return bool(C.ts_parser_set_language(p.parser, lang.language))
}

// Parse parses source code and returns a syntax tree.
// oldTree can be nil for a fresh parse, or a previous tree for incremental parsing.
func (p *TreeSitterParser) Parse(source string, oldTree *TreeSitterTree) *TreeSitterTree {
	if p.parser == nil {
		return nil
	}

	cSource := C.CString(source)
	defer C.free(unsafe.Pointer(cSource))

	cSourceLen := C.uint(len(source))

	var oldCTree *C.TSTree
	if oldTree != nil {
		oldCTree = oldTree.tree
	}

	ctree := C.ts_parser_parse_string(p.parser, oldCTree, cSource, cSourceLen)
	if ctree == nil {
		return nil
	}

	return &TreeSitterTree{tree: ctree}
}

// TreeSitterTree represents a parsed syntax tree.
type TreeSitterTree struct {
	tree *C.TSTree
}

// RootNode returns the root node of the tree.
func (t *TreeSitterTree) RootNode() TreeSitterNode {
	if t.tree == nil {
		return TreeSitterNode{}
	}
	node := C.ts_tree_root_node(t.tree)
	return TreeSitterNode{node: node}
}

// Delete frees the tree resources.
func (t *TreeSitterTree) Delete() {
	if t.tree != nil {
		C.ts_tree_delete(t.tree)
		t.tree = nil
	}
}

// TreeSitterNode represents a node in the syntax tree.
type TreeSitterNode struct {
	node C.TSNode
}

// IsValid returns true if the node is valid (non-zero).
func (n TreeSitterNode) IsValid() bool {
	return bool(C.ts_node_is_null(n.node))
}

// Type returns the node type string (e.g. "function_definition", "string_literal").
func (n TreeSitterNode) Type() string {
	cStr := C.ts_node_type(n.node)
	return C.GoString(cStr)
}

// StartByte returns the start byte offset of the node.
func (n TreeSitterNode) StartByte() uint32 {
	return uint32(C.ts_node_start_byte(n.node))
}

// EndByte returns the end byte offset of the node.
func (n TreeSitterNode) EndByte() uint32 {
	return uint32(C.ts_node_end_byte(n.node))
}

// StartPosition returns the start (row, column) of the node.
func (n TreeSitterNode) StartPosition() (uint32, uint32) {
	pt := C.ts_node_start_point(n.node)
	return uint32(pt.row), uint32(pt.column)
}

// EndPosition returns the end (row, column) of the node.
func (n TreeSitterNode) EndPosition() (uint32, uint32) {
	pt := C.ts_node_end_point(n.node)
	return uint32(pt.row), uint32(pt.column)
}

// ChildCount returns the number of children.
func (n TreeSitterNode) ChildCount() uint32 {
	return uint32(C.ts_node_child_count(n.node))
}

// Child returns the child at the given index.
func (n TreeSitterNode) Child(index uint32) TreeSitterNode {
	node := C.ts_node_child(n.node, C.uint(index))
	return TreeSitterNode{node: node}
}

// NamedChildCount returns the number of named children.
func (n TreeSitterNode) NamedChildCount() uint32 {
	return uint32(C.ts_node_named_child_count(n.node))
}

// NamedChild returns the named child at the given index.
func (n TreeSitterNode) NamedChild(index uint32) TreeSitterNode {
	node := C.ts_node_named_child(n.node, C.uint(index))
	return TreeSitterNode{node: node}
}

// IsNamed returns true if the node is a named node.
func (n TreeSitterNode) IsNamed() bool {
	return bool(C.ts_node_is_named(n.node))
}

// Text returns the source text covered by this node (requires source string).
func (n TreeSitterNode) Text(source string) string {
	start := n.StartByte()
	end := n.EndByte()
	if start >= uint32(len(source)) {
		return ""
	}
	if end > uint32(len(source)) {
		end = uint32(len(source))
	}
	return source[start:end]
}

// TreeSitterQuery is a compiled tree-sitter query for capturing nodes.
type TreeSitterQuery struct {
	query *C.TSQuery
}

// NewTreeSitterQuery compiles a query string for a given language.
func NewTreeSitterQuery(lang *TreeSitterLanguage, queryString string) (*TreeSitterQuery, error) {
	cQuery := C.CString(queryString)
	defer C.free(unsafe.Pointer(cQuery))

	var errorOffset C.uint32_t
	var errorType C.TSQueryError
	cquery := C.ts_query_new(
		lang.language,
		cQuery,
		C.uint(len(queryString)),
		&errorOffset,
		&errorType,
	)
	if cquery == nil {
		return nil, fmt.Errorf("failed to compile tree-sitter query (error offset %d, type %d)",
			errorOffset, errorType)
	}

	return &TreeSitterQuery{query: cquery}, nil
}

// Cursor returns a new query cursor for iterating matches.
func (q *TreeSitterQuery) Cursor() *TreeSitterQueryCursor {
	return &TreeSitterQueryCursor{cursor: C.ts_query_cursor_new()}
}

// Delete frees the query resources.
func (q *TreeSitterQuery) Delete() {
	if q.query != nil {
		C.ts_query_delete(q.query)
		q.query = nil
	}
}

// CaptureNameForId returns the capture name for a given capture id.
func (q *TreeSitterQuery) CaptureNameForId(id uint32) string {
	var length C.uint32_t
	cStr := C.ts_query_capture_name_for_id(q.query, C.uint(id), &length)
	return C.GoStringN(cStr, C.int(length))
}

// CaptureNames returns all capture names for the query.
func (q *TreeSitterQuery) CaptureNames() []string {
	count := int(C.ts_query_capture_count(q.query))
	names := make([]string, count)
	for i := 0; i < count; i++ {
		var length C.uint32_t
		cStr := C.ts_query_capture_name_for_id(q.query, C.uint(i), &length)
		names[i] = C.GoStringN(cStr, C.int(length))
	}
	return names
}

// TreeSitterQueryCursor is a cursor for executing a query over a tree.
type TreeSitterQueryCursor struct {
	cursor *C.TSQueryCursor
}

// Exec executes the query over the given node and returns matches.
func (c *TreeSitterQueryCursor) Exec(query *TreeSitterQuery, node TreeSitterNode) []TreeSitterMatch {
	C.ts_query_cursor_exec(c.cursor, query.query, node.node)

	var matches []TreeSitterMatch
	var match C.TSQueryMatch
	for bool(C.ts_query_cursor_next_match(c.cursor, &match)) {
		matches = append(matches, TreeSitterMatch{
			query:        query,
			patternIndex: int(match.pattern_index),
			captureCount: int(match.capture_count),
			captures:     match.captures,
		})
	}
	return matches
}

// Close closes the cursor.
func (c *TreeSitterQueryCursor) Close() {
	if c.cursor != nil {
		C.ts_query_cursor_delete(c.cursor)
		c.cursor = nil
	}
}

// TreeSitterMatch represents a single query match with its captures.
type TreeSitterMatch struct {
	query        *TreeSitterQuery
	patternIndex int
	captureCount int
	captures     *C.TSQueryCapture
}

// TreeSitterCapture represents a single capture from a query match.
type TreeSitterCapture struct {
	Index uint32
	Node  TreeSitterNode
	Name  string
}

// extractCaptures extracts capture info from a C TSQueryMatch.
func (m *TreeSitterMatch) extractCaptures() []TreeSitterCapture {
	captures := make([]TreeSitterCapture, m.captureCount)
	// captures is a C array pointer; iterate over it
	ptr := unsafe.Pointer(m.captures)
	for i := 0; i < m.captureCount; i++ {
		cap := (*C.TSQueryCapture)(unsafe.Add(ptr, i*C.sizeof_TSQueryCapture))
		node := TreeSitterNode{node: cap.node}
		name := m.query.CaptureNameForId(uint32(cap.index))
		captures[i] = TreeSitterCapture{
			Index: uint32(cap.index),
			Node:  node,
			Name:  name,
		}
	}
	return captures
}

// SyntaxHighlight represents a syntax highlight range.
type SyntaxHighlight struct {
	StartLine   int
	StartCol    int
	EndLine     int
	EndCol      int
	CaptureName string
}

// HighlightCode parses source code with the given language and query, returning
// a list of syntax highlight ranges suitable for rendering in the Code renderable.
func HighlightCode(parser *TreeSitterParser, lang *TreeSitterLanguage, query *TreeSitterQuery, source string) []SyntaxHighlight {
	tree := parser.Parse(source, nil)
	if tree == nil {
		return nil
	}
	defer tree.Delete()

	root := tree.RootNode()
	cursor := query.Cursor()
	defer cursor.Close()

	matches := cursor.Exec(query, root)

	var highlights []SyntaxHighlight
	lines := splitLinesNative(source)

	for _, match := range matches {
		for _, cap := range match.extractCaptures() {
			if !cap.Node.IsValid() {
				continue
			}
			startRow, startCol := cap.Node.StartPosition()
			endRow, endCol := cap.Node.EndPosition()

			// Clamp to source bounds
			if int(startRow) >= len(lines) {
				continue
			}

			highlights = append(highlights, SyntaxHighlight{
				StartLine:   int(startRow),
				StartCol:    int(startCol),
				EndLine:     int(endRow),
				EndCol:      int(endCol),
				CaptureName: cap.Name,
			})
		}
	}

	return highlights
}

// SyntaxTheme maps capture names to text attributes/colors.
type SyntaxTheme struct {
	Styles map[string]uint32 // capture name -> TextAttribute bitmask
}

// DefaultSyntaxTheme returns a reasonable default theme for common tree-sitter captures.
func DefaultSyntaxTheme() *SyntaxTheme {
	return &SyntaxTheme{
		Styles: map[string]uint32{
			"comment":             TextAttributeItalic,
			"keyword":             TextAttributeBold,
			"function":            TextAttributeBold,
			"function.builtin":    TextAttributeBold,
			"string":              0,
			"string.special":      0,
			"number":              0,
			"type":                TextAttributeBold,
			"type.builtin":        TextAttributeBold,
			"variable":            0,
			"variable.builtin":    0,
			"constant":            TextAttributeBold,
			"constant.builtin":    TextAttributeBold,
			"operator":            0,
			"punctuation":         0,
			"punctuation.bracket": 0,
			"tag":                 TextAttributeBold,
			"attribute":           0,
			"property":            0,
		},
	}
}

// StyleFor returns the text attribute for a capture name.
func (t *SyntaxTheme) StyleFor(captureName string) uint32 {
	if style, ok := t.Styles[captureName]; ok {
		return style
	}
	return TextAttributeNone
}

// splitLinesNative splits content into lines.
func splitLinesNative(s string) []string {
	if s == "" {
		return []string{""}
	}
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	lines = append(lines, s[start:])
	return lines
}
