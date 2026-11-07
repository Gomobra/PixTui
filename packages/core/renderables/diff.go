package renderables

import (
	"fmt"
	"regexp"

	"github.com/Gomobra/PixTui/packages/core"
	"github.com/Gomobra/PixTui/packages/native"
)

// DiffOptions contains options for DiffRenderable.
type DiffOptions struct {
	core.RenderableOptions[core.Renderable]
	Content string
	View    string // "unified" or "split"
	SyncScroll bool
	ShowLineNumbers bool
	LineNumberFg   *core.RGBA
	LineNumberBg   *core.RGBA
	AddedBg        *core.RGBA
	RemovedBg      *core.RGBA
	ContextBg      *core.RGBA
	AddedSignColor *core.RGBA
	RemovedSignColor *core.RGBA
	GutterPaddingLeft  int
	GutterPaddingRight int
}

// DiffLineType represents the type of a diff line.
type DiffLineType int

const (
	DiffLineContext DiffLineType = iota
	DiffLineAdd
	DiffLineRemove
	DiffLineEmpty
)

// DiffLine represents a single line in a diff.
type DiffLine struct {
	Content   string
	LineNum   int
	LineType  DiffLineType
	Sign      string
	SignColor *core.RGBA
	BgColor   *core.RGBA
	HideLineNumber bool
}

// DiffHunk represents a parsed hunk.
type DiffHunk struct {
	OldStart int
	NewStart int
	OldCount int
	NewCount int
	Lines    []DiffLine
}

// DiffRenderable renders diff content with +/- coloring.
type DiffRenderable struct {
	*core.Renderable
	options DiffOptions
	content string
	view    string
	parsed  []DiffHunk
	parseErr error
}

// NewDiffRenderable creates a new DiffRenderable.
func NewDiffRenderable(ctx *core.RenderContext, opts DiffOptions) *DiffRenderable {
	base := core.NewRenderable(ctx, opts.ID, core.RenderableOptions[core.Renderable]{
		Width:   opts.Width,
		Height:  opts.Height,
		Visible: opts.Visible,
		ZIndex:  opts.ZIndex,
	})
	view := opts.View
	if view == "" {
		view = "unified"
	}
	r := &DiffRenderable{
		Renderable: base,
		options:    opts,
		content:    opts.Content,
		view:       view,
	}
	// Create native scene node
	if ctx != nil && ctx.NativeScene != nil {
		node, err := ctx.NativeScene.CreateNode(r.Renderable, native.SceneNodeKindText)
		if err == nil {
			r.Renderable.SetNativeNode(node)
		}
	}
	if opts.Content != "" {
		r.parseDiff()
	}
	return r
}

// Content returns the diff source.
func (r *DiffRenderable) Content() string {
	return r.content
}

// SetContent sets the diff source and requests a re-render.
func (r *DiffRenderable) SetContent(content string) {
	r.content = content
	r.parseDiff()
	r.RequestRender()
}

// View returns the current view mode.
func (r *DiffRenderable) View() string {
	return r.view
}

// SetView sets the view mode ("unified" or "split").
func (r *DiffRenderable) SetView(v string) {
	r.view = v
	r.parseDiff()
	r.RequestRender()
}

// ParseError returns any parse error.
func (r *DiffRenderable) ParseError() error {
	return r.parseErr
}

// parsePatch parses unified diff content into hunks.
func (r *DiffRenderable) parseDiff() {
	r.parsed = nil
	r.parseErr = nil

	content := r.content
	if content == "" {
		return
	}

	lines := splitLines(content)
	hunkRe := regexp.MustCompile(`^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@`)

	var currentHunk *DiffHunk
	for _, line := range lines {
		if hunkRe.MatchString(line) {
			matches := hunkRe.FindStringSubmatch(line)
			oldStart := 1
			if matches[1] != "" {
				fmt.Sscanf(matches[1], "%d", &oldStart)
			}
			newStart := 1
			if matches[3] != "" {
				fmt.Sscanf(matches[3], "%d", &newStart)
			}
			currentHunk = &DiffHunk{
				OldStart: oldStart,
				NewStart: newStart,
			}
			r.parsed = append(r.parsed, *currentHunk)
			currentHunk = &r.parsed[len(r.parsed)-1]
			continue
		}

		if currentHunk == nil {
			continue
		}

		if len(line) == 0 {
			continue
		}

		dl := DiffLine{Content: line[1:]}
		switch line[0] {
		case '+':
			dl.LineType = DiffLineAdd
			dl.LineNum = currentHunk.NewStart
			currentHunk.NewStart++
			currentHunk.NewCount++
		case '-':
			dl.LineType = DiffLineRemove
			dl.LineNum = currentHunk.OldStart
			currentHunk.OldStart++
			currentHunk.OldCount++
		case ' ':
			dl.LineType = DiffLineContext
			dl.LineNum = currentHunk.NewStart
			currentHunk.NewStart++
			currentHunk.OldStart++
			currentHunk.NewCount++
			currentHunk.OldCount++
		case '\\':
			dl.LineType = DiffLineEmpty
			continue
		default:
			continue
		}
		currentHunk.Lines = append(currentHunk.Lines, dl)
	}

	if len(r.parsed) == 0 {
		r.parseErr = fmt.Errorf("no valid diff hunks found")
	}
}

// RenderSelf renders the diff into the buffer with +/- coloring.
func (r *DiffRenderable) RenderSelf(buffer *core.NativeBuffer, deltaTime float64) {
	if r.options.RenderBefore != nil {
		r.options.RenderBefore(r.Renderable, buffer, deltaTime)
	}

	if r.parseErr != nil {
		r.renderErrorView(buffer)
		if r.options.RenderAfter != nil {
			r.options.RenderAfter(r.Renderable, buffer, deltaTime)
		}
		return
	}

	if len(r.parsed) == 0 {
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

	// Colors
	green := &core.RGBA{G: 120, A: 255}
	red := &core.RGBA{R: 220, A: 255}
	lnFg := r.options.LineNumberFg
	if lnFg == nil {
		lnFg = &core.RGBA{R: 136, G: 136, B: 136, A: 255}
	}
	addedBg := r.options.AddedBg
	removedBg := r.options.RemovedBg
	contextBg := r.options.ContextBg
	addedSign := r.options.AddedSignColor
	if addedSign == nil {
		addedSign = &core.RGBA{G: 200, A: 255}
	}
	removedSign := r.options.RemovedSignColor
	if removedSign == nil {
		removedSign = &core.RGBA{R: 200, A: 255}
	}

	gutterPL := r.options.GutterPaddingLeft
	if gutterPL < 0 { gutterPL = 1 }
	gutterPR := r.options.GutterPaddingRight
	if gutterPR < 0 { gutterPR = 1 }

	// Build flat line list based on view mode
	var flatLines []DiffLine
	if r.view == "split" {
		flatLines = r.buildSplitLines()
	} else {
		flatLines = r.buildUnifiedLines()
	}

	row := y
	lineIdx := 0
	for _, dl := range flatLines {
		if row >= y+h {
			break
		}

		// Determine colors
		var fg *core.RGBA
		var bg *core.RGBA
		var sign string
		var signColor *core.RGBA

		switch dl.LineType {
		case DiffLineAdd:
			fg = green
			bg = addedBg
			sign = " +"
			signColor = addedSign
		case DiffLineRemove:
			fg = red
			bg = removedBg
			sign = " -"
			signColor = removedSign
		case DiffLineContext:
			fg = nil
			bg = contextBg
			sign = "  "
			signColor = lnFg
		default:
			fg = nil
			bg = nil
			sign = "  "
			signColor = lnFg
		}

		// Gutter area
		gutterX := x
		for gx := 0; gx < gutterPL; gx++ {
			if gutterX+gx >= x+w { break }
			writeCell(buffer, gutterX+gx, row, ' ', signColor, bg, core.TextAttributeNone)
		}

		// Line number
		if r.options.ShowLineNumbers && !dl.HideLineNumber {
			lnStr := ""
			if dl.LineNum > 0 {
				lnStr = fmt.Sprintf("%d", dl.LineNum)
			} else {
				lnStr = " "
			}
			lnX := gutterX + gutterPL + gutterPR
			if lnX+len(lnStr) <= x+w {
				writeText(buffer, lnX, row, lnStr, lnFg, bg, core.TextAttributeNone)
			}
		}

		// Sign column
		signX := gutterX + gutterPL + gutterPR + 7
		if signX+len(sign) <= x+w {
			writeText(buffer, signX, row, sign, signColor, bg, core.TextAttributeNone)
		}

		// Content
		contentX := signX + len(sign) + 1
		contentStr := dl.Content
		if contentStr == "" {
			contentStr = " "
		}
		maxCol := x + w
		col := contentX
		for _, ch := range contentStr {
			if col >= maxCol { break }
			writeCell(buffer, col, row, uint32(ch), fg, bg, core.TextAttributeNone)
			col++
		}

		row++
		lineIdx++
	}

	if r.options.RenderAfter != nil {
		r.options.RenderAfter(r.Renderable, buffer, deltaTime)
	}
}

func (r *DiffRenderable) renderErrorView(buffer *core.NativeBuffer) {
	x := int(r.GetX())
	y := int(r.GetY())
	w := int(r.GetWidth())
	h := int(r.GetHeight())
	if w < 1 || h < 1 { return }
	if x < 0 { w += x; x = 0 }
	if y < 0 { h += y; y = 0 }
	if x+w > buffer.Width() { w = buffer.Width() - x }
	if y+h > buffer.Height() { h = buffer.Height() - y }
	if w <= 0 || h <= 0 { return }

	fillRect(buffer, x, y, w, h, ' ', nil, nil, 0)
	errMsg := "Error parsing diff: "
	if r.parseErr != nil {
		errMsg += r.parseErr.Error()
	} else {
		errMsg += "Unknown error"
	}
	writeText(buffer, x, y, errMsg, &core.RGBA{R: 220, A: 255}, nil, 0)
	writeText(buffer, x, y+1, r.content, &core.RGBA{R: 200, G: 200, B: 200, A: 255}, nil, core.TextAttributeDim)
}

func (r *DiffRenderable) buildUnifiedLines() []DiffLine {
	var result []DiffLine
	for _, hunk := range r.parsed {
		for _, dl := range hunk.Lines {
			result = append(result, dl)
		}
	}
	return result
}

func (r *DiffRenderable) buildSplitLines() []DiffLine {
	var result []DiffLine
	for _, hunk := range r.parsed {
		// Split removes and adds side by side
		var removes []DiffLine
		var adds []DiffLine
		for _, dl := range hunk.Lines {
			if dl.LineType == DiffLineRemove {
				removes = append(removes, dl)
			} else if dl.LineType == DiffLineAdd {
				adds = append(adds, dl)
			}
		}
		maxLen := len(removes)
		if len(adds) > maxLen { maxLen = len(adds) }
		for i := 0; i < maxLen; i++ {
			if i < len(removes) {
				dl := removes[i]
				dl.LineType = DiffLineRemove
				result = append(result, dl)
			} else {
				result = append(result, DiffLine{Content: "", LineType: DiffLineEmpty})
			}
			if i < len(adds) {
				dl := adds[i]
				dl.LineType = DiffLineAdd
				result = append(result, dl)
			} else {
				result = append(result, DiffLine{Content: "", LineType: DiffLineEmpty})
			}
		}
	}
	return result
}

var _ = native.SceneNodeKindText
var _ = regexp.MustCompile
