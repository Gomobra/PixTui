package renderables

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/Gomobra/PixTui/packages/core"
	"github.com/Gomobra/PixTui/packages/native"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
)

// ---- Markdown table options ----

// MarkdownTableStyle is the visual style for markdown tables.
type MarkdownTableStyle string

const (
	MarkdownTableStyleGrid     MarkdownTableStyle = "grid"
	MarkdownTableStyleColumns  MarkdownTableStyle = "columns"
)

// TextTableColumnWidthMode is the column sizing strategy.
type TextTableColumnWidthMode string

const (
	TextTableColumnWidthContent TextTableColumnWidthMode = "content"
	TextTableColumnWidthFull    TextTableColumnWidthMode = "full"
)

// TextTableColumnFitter is the column fitting method.
type TextTableColumnFitter string

const (
	TextTableColumnFitterProportional TextTableColumnFitter = "proportional"
)

// MarkdownTableOptions contains options for rendering markdown tables.
type MarkdownTableOptions struct {
	Style        MarkdownTableStyle
	WidthMode    TextTableColumnWidthMode
	ColumnFitter TextTableColumnFitter
	WrapMode     string // "none" | "char" | "word"
	CellPadding  int
	CellPaddingX int
	CellPaddingY int
	Borders      *bool
	OuterBorder  *bool
	BorderStyle  string
	BorderColor  *core.RGBA
	Selectable   *bool
}

// ResolvedTableOptions holds resolved table rendering options.
type ResolvedTableOptions struct {
	ColumnWidthMode TextTableColumnWidthMode
	ColumnFitter    TextTableColumnFitter
	WrapMode        string
	CellPadding     int
	CellPaddingX    int
	CellPaddingY    int
	ColumnGap       int
	Border          bool
	OuterBorder     bool
	ShowBorders     bool
	BorderStyle     string
	BorderColor     *core.RGBA
	Selectable      bool
}

// ---- Markdown render node ----

// RenderNodeContext is the context passed to a custom renderNode.
type RenderNodeContext struct {
	SyntaxStyle  interface{}
	Conceal      bool
	ConcealCode  bool
	TreeSitter   interface{}
	DefaultRender func() *MarkdownRenderable
}

// MarkdownRenderNode is a custom node renderer.
type MarkdownRenderNode func(token ast.Node, context RenderNodeContext) *MarkdownRenderable

// ---- Markdown colors ----

var (
	mdColorWhite    = &core.RGBA{R: 255, G: 255, B: 255, A: 255}
	mdColorGray     = &core.RGBA{R: 160, G: 160, B: 160, A: 255}
	mdColorDarkGray = &core.RGBA{R: 90, G: 90, B: 90, A: 255}
	mdColorCyan     = &core.RGBA{R: 0, G: 170, B: 255, A: 255}
	mdColorBlue     = &core.RGBA{R: 0, G: 110, B: 255, A: 255}
)

// ---- Markdown options ----

// MarkdownOptions contains options for MarkdownRenderable.
type MarkdownOptions struct {
	core.RenderableOptions[core.Renderable]
	Content        string
	SyntaxStyle    interface{}
	Conceal        bool
	ConcealCode    bool
	Streaming      bool
	TableOptions   *MarkdownTableOptions
	RenderNode     MarkdownRenderNode
	InternalBlockMode string // "coalesced" | "top-level"
}

// BlockState represents the state of a rendered markdown block.
type BlockState struct {
	Token          ast.Node
	TokenRaw       string
	MarginTop      int
	Renderable     *MarkdownRenderable
	TableContent   interface{}
	TracksMargin   bool
	CanUpdate      bool
}

// MarkdownRenderable renders markdown content using gomarkdown/markdown.
type MarkdownRenderable struct {
	*core.Renderable
	options MarkdownOptions

	syntaxStyle interface{}
	fg          *core.RGBA
	bg          *core.RGBA
	conceal     bool
	concealCode bool
	content     string
	streaming   bool
	tableOpts   *MarkdownTableOptions
	renderNode  MarkdownRenderNode
	internalMode string

	parseState    interface{}
	blockStates   []BlockState
	stableCount   int
	styleDirty    bool
	capMu         sync.RWMutex
	capSubscribed bool
}

// NewMarkdownRenderable creates a new MarkdownRenderable.
func NewMarkdownRenderable(ctx *core.RenderContext, opts MarkdownOptions) *MarkdownRenderable {
	base := core.NewRenderable(ctx, opts.ID, core.RenderableOptions[core.Renderable]{
		Width:   opts.Width,
		Height:  opts.Height,
		Visible: opts.Visible,
		ZIndex:  opts.ZIndex,
	})
	internalMode := opts.InternalBlockMode
	if internalMode == "" {
		internalMode = "coalesced"
	}
	conceal := opts.Conceal
	if !conceal {
		conceal = true
	}
	r := &MarkdownRenderable{
		Renderable:    base,
		options:       opts,
		syntaxStyle:   opts.SyntaxStyle,
		conceal:       conceal,
		concealCode:   opts.ConcealCode,
		content:       opts.Content,
		streaming:     opts.Streaming,
		tableOpts:     opts.TableOptions,
		renderNode:    opts.RenderNode,
		internalMode:  internalMode,
	}
	// Create native scene node
	if ctx != nil && ctx.NativeScene != nil {
		node, err := ctx.NativeScene.CreateNode(r.Renderable, native.SceneNodeKindText)
		if err == nil {
			r.Renderable.SetNativeNode(node)
		}
	}
	if opts.Content != "" {
		r.updateBlocks()
	}
	r.subscribeCapabilities(ctx)
	return r
}

func (r *MarkdownRenderable) subscribeCapabilities(ctx *core.RenderContext) {
	r.capMu.Lock()
	defer r.capMu.Unlock()
	if r.capSubscribed { return }
	r.capSubscribed = true
	_ = ctx
	// Capability subscription handled via RenderContext.Capabilities
}

// Content returns the markdown source.
func (r *MarkdownRenderable) Content() string {
	return r.content
}

// SetContent sets the markdown source and requests a re-render.
func (r *MarkdownRenderable) SetContent(content string) {
	if r.content == content { return }
	r.content = content
	r.updateBlocks()
	r.RequestRender()
}

func (r *MarkdownRenderable) updateBlocks() {
	r.blockStates = nil
	r.stableCount = 0
	r.styleDirty = true
}

// baseStyle returns the default style for ordinary markdown text.
func (r *MarkdownRenderable) baseStyle() *core.RGBA {
	if r.options.ForegroundColor != nil {
		return r.options.ForegroundColor
	}
	return mdColorWhite
}

// styleFor resolves a named markdown scope into a style.
func (r *MarkdownRenderable) styleFor(scope string) mdStyle {
	base := mdStyle{fg: r.baseStyle(), bg: r.options.BackgroundColor}
	switch scope {
	case "default":
		return base
	case "markup.raw":
		return mdStyle{fg: mdColorCyan, bg: nil, attrs: base.attrs}
	case "markup.strong":
		return mdStyle{fg: nil, bg: nil, attrs: base.attrs | core.TextAttributeBold}
	case "markup.italic":
		return mdStyle{fg: nil, bg: nil, attrs: base.attrs | core.TextAttributeItalic}
	case "markup.strikethrough":
		return mdStyle{fg: nil, bg: nil, attrs: base.attrs | core.TextAttributeStrikethrough}
	case "markup.link":
		return mdStyle{fg: mdColorBlue, bg: nil, attrs: base.attrs}
	case "markup.link.label":
		return base
	case "markup.link.url":
		return mdStyle{fg: mdColorCyan, bg: nil, attrs: base.attrs | core.TextAttributeDim}
	case "markup.heading":
		return mdStyle{fg: nil, bg: nil, attrs: base.attrs | core.TextAttributeBold}
	case "markup.list":
		return mdStyle{fg: mdColorGray, bg: nil, attrs: base.attrs}
	case "markup.quote":
		return mdStyle{fg: mdColorGray, bg: nil, attrs: base.attrs}
	case "markup.codeblock":
		if r.options.ConcealCode {
			return base
		}
		return mdStyle{fg: mdColorGray, bg: nil, attrs: base.attrs}
	case "markup.tablehead":
		return mdStyle{fg: nil, bg: nil, attrs: base.attrs | core.TextAttributeBold}
	case "conceal":
		return mdStyle{fg: mdColorDarkGray, bg: nil, attrs: base.attrs}
	case "hr":
		return mdStyle{fg: mdColorGray, bg: nil, attrs: base.attrs}
	}
	return base
}

type mdStyle struct {
	fg    *core.RGBA
	bg    *core.RGBA
	attrs uint32
}

type mdRun struct {
	text  string
	style mdStyle
}

// RenderSelf renders the parsed markdown into the buffer.
func (r *MarkdownRenderable) RenderSelf(buffer *core.NativeBuffer, deltaTime float64) {
	if r.options.RenderBefore != nil {
		r.options.RenderBefore(r.Renderable, buffer, deltaTime)
	}

	content := r.content
	if content == "" {
		if r.options.RenderAfter != nil {
			r.options.RenderAfter(r.Renderable, buffer, deltaTime)
		}
		return
	}

	x := int(r.GetX())
	y := int(r.GetY())
	w := int(r.GetWidth())
	h := int(r.GetHeight())

	if w < 1 || h < 1 {
		if r.options.RenderAfter != nil {
			r.options.RenderAfter(r.Renderable, buffer, deltaTime)
		}
		return
	}
	if x < 0 {
		w += x
		x = 0
	}
	if y < 0 {
		h += y
		y = 0
	}
	if x+w > buffer.Width() {
		w = buffer.Width() - x
	}
	if y+h > buffer.Height() {
		h = buffer.Height() - y
	}
	if w <= 0 || h <= 0 {
		if r.options.RenderAfter != nil {
			r.options.RenderAfter(r.Renderable, buffer, deltaTime)
		}
		return
	}

	// Fill background with the renderable's background color.
	bg := r.options.BackgroundColor
	if bg != nil {
		fillRect(buffer, x, y, w, h, ' ', nil, bg, 0)
	}

	extensions := parser.CommonExtensions
	doc := markdown.Parse([]byte(content), parser.NewWithExtensions(extensions))

	ren := &mdRenderer{
		buffer:  buffer,
		ox:      x,
		oy:      y,
		w:       w,
		h:       h,
		base:    r.styleFor("default"),
		conceal: r.options.Conceal,
	}
	ren.renderDocument(doc.(*ast.Document))
	ren.writeLines()

	if r.options.RenderAfter != nil {
		r.options.RenderAfter(r.Renderable, buffer, deltaTime)
	}
}

// mdRenderer walks the gomarkdown AST and accumulates styled lines.
type mdRenderer struct {
	buffer  *core.NativeBuffer
	ox, oy  int
	w, h    int
	base    mdStyle
	conceal bool
	lines   [][]mdRun
}

func (r *mdRenderer) styleFor(scope string) mdStyle {
	base := r.base
	switch scope {
	case "default":
		return base
	case "markup.raw":
		return mdStyle{fg: mdColorCyan, bg: nil, attrs: base.attrs}
	case "markup.strong":
		return mdStyle{fg: nil, bg: nil, attrs: base.attrs | core.TextAttributeBold}
	case "markup.italic":
		return mdStyle{fg: nil, bg: nil, attrs: base.attrs | core.TextAttributeItalic}
	case "markup.strikethrough":
		return mdStyle{fg: nil, bg: nil, attrs: base.attrs | core.TextAttributeStrikethrough}
	case "markup.link":
		return mdStyle{fg: mdColorBlue, bg: nil, attrs: base.attrs}
	case "markup.link.label":
		return base
	case "markup.link.url":
		return mdStyle{fg: mdColorCyan, bg: nil, attrs: base.attrs | core.TextAttributeDim}
	case "markup.heading":
		return mdStyle{fg: nil, bg: nil, attrs: base.attrs | core.TextAttributeBold}
	case "markup.list":
		return mdStyle{fg: mdColorGray, bg: nil, attrs: base.attrs}
	case "markup.quote":
		return mdStyle{fg: mdColorGray, bg: nil, attrs: base.attrs}
	case "markup.codeblock":
		return mdStyle{fg: mdColorGray, bg: nil, attrs: base.attrs}
	case "markup.tablehead":
		return mdStyle{fg: nil, bg: nil, attrs: base.attrs | core.TextAttributeBold}
	case "conceal":
		return mdStyle{fg: mdColorDarkGray, bg: nil, attrs: base.attrs}
	case "hr":
		return mdStyle{fg: mdColorGray, bg: nil, attrs: base.attrs}
	}
	return base
}

func mergeStyles(base, overlay mdStyle) mdStyle {
	fg := base.fg
	if overlay.fg != nil { fg = overlay.fg }
	bg := base.bg
	if overlay.bg != nil { bg = overlay.bg }
	return mdStyle{fg: fg, bg: bg, attrs: base.attrs | overlay.attrs}
}

func (r *mdRenderer) renderDocument(n *ast.Document) {
	for _, c := range n.GetChildren() {
		r.renderBlock(c, nil)
	}
}

func (r *mdRenderer) renderBlock(n ast.Node, prefix []mdRun) {
	lines := r.collectBlockLines(n, r.w)
	r.emit(lines, prefix)
}

func (r *mdRenderer) emit(lines [][]mdRun, prefix []mdRun) {
	for _, line := range lines {
		if len(prefix) > 0 {
			nl := make([]mdRun, 0, len(prefix)+len(line))
			nl = append(nl, prefix...)
			nl = append(nl, line...)
			line = nl
		}
		r.lines = append(r.lines, line)
	}
	r.lines = append(r.lines, []mdRun{})
}

func (r *mdRenderer) collectBlockLines(n ast.Node, maxWidth int) [][]mdRun {
	if maxWidth < 1 { maxWidth = 1 }
	switch node := n.(type) {
	case *ast.Document:
		var all [][]mdRun
		for _, c := range node.GetChildren() {
			all = append(all, r.collectBlockLines(c, maxWidth)...)
		}
		return all
	case *ast.Paragraph:
		return r.wrapInline(r.inlineRuns(node.GetChildren(), r.base), maxWidth)
	case *ast.Heading:
		style := mergeStyles(r.base, r.styleFor("markup.heading"))
		return r.wrapInline(r.inlineRuns(node.GetChildren(), style), maxWidth)
	case *ast.BlockQuote:
		var all [][]mdRun
		for _, c := range node.GetChildren() {
			all = append(all, r.collectBlockLines(c, maxWidth-3)...)
		}
		return all
	case *ast.List:
		return r.collectListLines(node, maxWidth)
	case *ast.CodeBlock:
		return r.collectCodeBlockLines(node, maxWidth)
	case *ast.HorizontalRule:
		return [][]mdRun{{
			mdRun{text: strings.Repeat("─", maxWidth), style: r.styleFor("hr")},
		}}
	case *ast.Table:
		return r.collectTableLines(node, maxWidth)
	default:
		return r.wrapInline(r.inlineRuns(node.GetChildren(), r.base), maxWidth)
	}
}

func (r *mdRenderer) collectListLines(n *ast.List, maxWidth int) [][]mdRun {
	var items []*ast.ListItem
	for _, c := range n.GetChildren() {
		if li, ok := c.(*ast.ListItem); ok {
			items = append(items, li)
		}
	}
	ordered := false
	// Check list flags for ordered
	start := 1
	bulletChar := byte('-')
	markers := make([]string, len(items))
	maxMarker := 0
	for i := range items {
		var m string
		if ordered {
			m = fmt.Sprintf("%d.", start+i)
		} else {
			m = string(bulletChar)
		}
		markers[i] = m
		if w := measureCellWidth(m); w > maxMarker {
			maxMarker = w
		}
	}
	indent := strings.Repeat(" ", maxMarker+1)
	listStyle := r.styleFor("markup.list")
	contentWidth := maxWidth - maxMarker - 1
	if contentWidth < 1 { contentWidth = 1 }
	var all [][]mdRun
	for i := range items {
		li := items[i]
		markerPad := strings.Repeat(" ", maxMarker-measureCellWidth(markers[i]))
		firstPrefix := []mdRun{{text: markers[i] + markerPad + " ", style: listStyle}}
		contPrefix := []mdRun{{text: indent, style: listStyle}}
		var runs []mdRun
		for _, c := range li.GetChildren() {
			if subList, ok := c.(*ast.List); ok {
				nested := r.collectListLines(subList, maxWidth-maxMarker-2)
				indentPrefix := []mdRun{{text: indent, style: listStyle}}
				for _, line := range nested {
					nl := make([]mdRun, 0, len(indentPrefix)+len(line))
					nl = append(nl, indentPrefix...)
					nl = append(nl, line...)
					all = append(all, nl)
				}
				continue
			}
			runs = append(runs, r.inlineRuns(c.GetChildren(), r.base)...)
			runs = append(runs, mdRun{text: "\n", style: r.base})
		}
		lines := r.wrapInline(runs, contentWidth)
		for liIdx, line := range lines {
			prefix := firstPrefix
			if liIdx > 0 { prefix = contPrefix }
			nl := make([]mdRun, 0, len(prefix)+len(line))
			nl = append(nl, prefix...)
			nl = append(nl, line...)
			all = append(all, nl)
		}
	}
	return all
}

func (r *mdRenderer) collectCodeBlockLines(n *ast.CodeBlock, maxWidth int) [][]mdRun {
	content := string(n.Literal)
	if n.IsFenced {
		content = strings.TrimRight(content, "\n")
	}
	style := r.styleFor("markup.codeblock")
	lines := [][]mdRun{}
	for _, line := range splitLines(content) {
		if measureCellWidth(line) > maxWidth {
			line = truncateToWidth(line, maxWidth)
		}
		lines = append(lines, []mdRun{{text: line, style: style}})
	}
	return lines
}

func (r *mdRenderer) collectTableLines(n *ast.Table, maxWidth int) [][]mdRun {
	var header []*ast.TableCell
	var rows [][]*ast.TableCell
	for _, c := range n.GetChildren() {
		switch child := c.(type) {
		case *ast.TableHeader:
			for _, hc := range child.GetChildren() {
				if cell, ok := hc.(*ast.TableCell); ok {
					header = append(header, cell)
				}
			}
		case *ast.TableBody:
			for _, tr := range child.GetChildren() {
				if row, ok := tr.(*ast.TableRow); ok {
					var cells []*ast.TableCell
					for _, rc := range row.GetChildren() {
						if cell, ok := rc.(*ast.TableCell); ok {
							cells = append(cells, cell)
						}
					}
					rows = append(rows, cells)
				}
			}
		case *ast.TableRow:
			var cells []*ast.TableCell
			for _, rc := range child.GetChildren() {
				if cell, ok := rc.(*ast.TableCell); ok {
					cells = append(cells, cell)
				}
			}
			rows = append(rows, cells)
		}
	}
	colCount := len(header)
	for _, row := range rows {
		if len(row) > colCount { colCount = len(row) }
	}
	if colCount == 0 { return nil }
	getCell := func(cells []*ast.TableCell, i int) *ast.TableCell {
		if i < len(cells) { return cells[i] }
		return nil
	}
	runsCell := func(cell *ast.TableCell) []mdRun {
		if cell == nil { return []mdRun{{text: " ", style: r.base}} }
		return r.inlineRuns(cell.GetChildren(), r.base)
	}
	widths := make([]int, colCount)
	allCells := [][][]mdRun{}
	addRow := func(cells []*ast.TableCell) {
		rowRuns := make([][]mdRun, colCount)
		for i := 0; i < colCount; i++ {
			cellRuns := runsCell(getCell(cells, i))
			rowRuns[i] = cellRuns
			if w := runsWidth(cellRuns); w > widths[i] { widths[i] = w }
		}
		allCells = append(allCells, rowRuns)
	}
	addRow(header)
	for _, row := range rows { addRow(row) }
	sep := mdRun{text: " | ", style: r.base}
	headStyle := r.styleFor("markup.tablehead")
	lines := [][]mdRun{}
	writeRow := func(rowRuns [][]mdRun, head bool) {
		line := []mdRun{}
		for i := 0; i < colCount; i++ {
			if i > 0 { line = append(line, sep) }
			cell := rowRuns[i]
			if head {
				for i2 := range cell { cell[i2].style = mergeStyles(cell[i2].style, headStyle) }
			}
			line = append(line, cell...)
			if pad := widths[i] - runsWidth(cell); pad > 0 {
				line = append(line, mdRun{text: strings.Repeat(" ", pad), style: r.base})
			}
		}
		lines = append(lines, line)
	}
	writeRow(allCells[0], true)
	sepLine := []mdRun{}
	for i := 0; i < colCount; i++ {
		if i > 0 { sepLine = append(sepLine, sep) }
		sepLine = append(sepLine, mdRun{text: strings.Repeat("─", widths[i]), style: r.styleFor("conceal")})
	}
	lines = append(lines, sepLine)
	for i := 1; i < len(allCells); i++ { writeRow(allCells[i], false) }
	return lines
}

func (r *mdRenderer) inlineRuns(children []ast.Node, style mdStyle) []mdRun {
	var runs []mdRun
	var add func(node ast.Node, st mdStyle)
	add = func(node ast.Node, st mdStyle) {
		switch n := node.(type) {
		case *ast.Text:
			runs = append(runs, mdRun{text: expandTabs(string(n.Literal)), style: st})
		case *ast.Code:
			text := expandTabs(string(n.Literal))
			if r.conceal {
				runs = append(runs, mdRun{text: text, style: mergeStyles(st, r.styleFor("markup.raw"))})
			} else {
				runs = append(runs, mdRun{text: "`" + text + "`", style: mergeStyles(st, r.styleFor("markup.raw"))})
			}
		case *ast.Emph:
			s := mergeStyles(st, r.styleFor("markup.italic"))
			for _, c := range n.GetChildren() { add(c, s) }
		case *ast.Strong:
			s := mergeStyles(st, r.styleFor("markup.strong"))
			for _, c := range n.GetChildren() { add(c, s) }
		case *ast.Del:
			s := mergeStyles(st, r.styleFor("markup.strikethrough"))
			for _, c := range n.GetChildren() { add(c, s) }
		case *ast.Link:
			if !r.conceal {
				runs = append(runs, mdRun{text: "[", style: mergeStyles(st, r.styleFor("markup.link"))})
			}
			for _, c := range n.GetChildren() {
				add(c, mergeStyles(st, r.styleFor("markup.link.label")))
			}
			if !r.conceal {
				runs = append(runs, mdRun{text: "](", style: mergeStyles(st, r.styleFor("markup.link"))})
				if dest := string(n.Destination); dest != "" {
					runs = append(runs, mdRun{text: dest, style: mergeStyles(st, r.styleFor("markup.link.url"))})
				}
				runs = append(runs, mdRun{text: ")", style: mergeStyles(st, r.styleFor("markup.link"))})
			}
		case *ast.Image:
			if !r.conceal {
				runs = append(runs, mdRun{text: "![", style: mergeStyles(st, r.styleFor("markup.link"))})
			}
			for _, c := range n.GetChildren() {
				add(c, mergeStyles(st, r.styleFor("markup.link.label")))
			}
			if !r.conceal {
				runs = append(runs, mdRun{text: "](", style: mergeStyles(st, r.styleFor("markup.link"))})
				if dest := string(n.Destination); dest != "" {
					runs = append(runs, mdRun{text: dest, style: mergeStyles(st, r.styleFor("markup.link.url"))})
				}
				runs = append(runs, mdRun{text: ")", style: mergeStyles(st, r.styleFor("markup.link"))})
			}
		case *ast.Hardbreak:
			runs = append(runs, mdRun{text: "\n", style: st})
		default:
			for _, c := range n.GetChildren() { add(c, st) }
		}
	}
	for _, c := range children { add(c, style) }
	return runs
}

func (r *mdRenderer) wrapInline(runs []mdRun, maxWidth int) [][]mdRun {
	if maxWidth < 1 { maxWidth = 1 }
	lines := [][]mdRun{{}}
	cur := &lines[0]
	curW := 0
	flush := func() {
		lines = append(lines, []mdRun{})
		cur = &lines[len(lines)-1]
		curW = 0
	}
	for _, run := range runs {
		if run.text == "\n" {
			flush()
			continue
		}
		text := run.text
		for len(text) > 0 {
			avail := maxWidth - curW
			if avail <= 0 {
				flush()
				avail = maxWidth
			}
			chunk, rest := splitByWidth(text, avail)
			if chunk == "" {
				flush()
				avail = maxWidth
				if avail <= 0 { avail = 1 }
				chunk, rest = splitByWidth(text, avail)
				if chunk == "" {
					chunk = text[:1]
					rest = text[1:]
				}
			}
			*cur = append(*cur, mdRun{text: chunk, style: run.style})
			curW += measureCellWidth(chunk)
			text = rest
			if text != "" {
				flush()
				avail = maxWidth
			}
		}
	}
	return lines
}

func (r *mdRenderer) writeLines() {
	cx, cy := r.ox, r.oy
	maxX := r.ox + r.w
	maxY := r.oy + r.h
	for _, line := range r.lines {
		if cy >= maxY { break }
		x := cx
		for _, run := range line {
			if x >= maxX { break }
			style := run.style
			fg := style.fg
			if fg == nil { fg = mdColorWhite }
			attrs := style.attrs
			for _, ch := range run.text {
				if x >= maxX { break }
				if ch == '\n' || ch == '\r' { continue }
				writeCell(r.buffer, x, cy, uint32(ch), fg, style.bg, attrs)
				x++
			}
		}
		cy++
	}
}

func expandTabs(s string) string {
	return strings.ReplaceAll(s, "\t", "    ")
}

func splitByWidth(text string, width int) (string, string) {
	if width <= 0 { return "", text }
	w := 0
	for i, r := range text {
		if r == '\t' { w += 4 } else { w++ }
		if w > width {
			return text[:i], text[i:]
		}
	}
	return text, ""
}

func runsWidth(runs []mdRun) int {
	w := 0
	for _, run := range runs { w += measureCellWidth(run.text) }
	return w
}


var _ = sort.Slice
var _ = fmt.Sprintf
