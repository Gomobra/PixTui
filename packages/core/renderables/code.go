package renderables

import (
	"fmt"

	"github.com/Gomobra/PixTui/packages/core"
	"github.com/Gomobra/PixTui/packages/native"
)

// CodeOptions contains options for CodeRenderable.
type CodeOptions struct {
	core.RenderableOptions[core.Renderable]
	Code        string
	Language    string
	LineNumbers bool
}

// CodeRenderable renders code with syntax highlighting using tree-sitter.
type CodeRenderable struct {
	*core.Renderable
	options CodeOptions
}

// tree-sitter queries for common languages.
// These queries capture syntax nodes with capture names that map to
// text attributes via the default syntax theme.

var codeQueries = map[native.SupportedLanguage]string{
	native.LanguageGo: `
		[
			(comment) @comment
			(keyword) @keyword
			(function_definition name: (identifier) @function)
			(method_definition name: (field_identifier) @function)
			(string_literal) @string
			(interpreted_string_literal) @string
			(raw_string_literal) @string
			(integer_literal) @number
			(floating_point_literal) @number
			(type_identifier) @type
			(builtin_type_identifier) @type.builtin
			(parameter) @variable
			(field_identifier) @variable
			(constant_identifier) @constant
			(operator) @operator
			(punctuation) @punctuation
		]
	`,
	native.LanguageJavaScript: `
		[
			(comment) @comment
			(keyword) @keyword
			(function_declaration name: (identifier) @function)
			(arrow_function) @function
			(string) @string
			(template_string) @string
			(number) @number
			(type_identifier) @type
			(identifier) @variable
			(constant) @constant
			(operator) @operator
			(punctuation) @punctuation
		]
	`,
	native.LanguagePython: `
		[
			(comment) @comment
			(keyword) @keyword
			(function_definition name: (identifier) @function)
			(string) @string
			(integer) @number
			(float) @number
			(identifier) @type
			(parameter) @variable
			(constant) @constant
			(operator) @operator
			(punctuation) @punctuation
		]
	`,
}

// NewCodeRenderable creates a new CodeRenderable.
func NewCodeRenderable(ctx *core.RenderContext, opts CodeOptions) *CodeRenderable {
	base := core.NewRenderable(ctx, opts.ID, core.RenderableOptions[core.Renderable]{
		Width:   opts.Width,
		Height:  opts.Height,
		Visible: opts.Visible,
		ZIndex:  opts.ZIndex,
	})
	r := &CodeRenderable{
		Renderable: base,
		options:    opts,
	}
	// Create native scene node
	if ctx != nil && ctx.NativeScene != nil {
		node, err := ctx.NativeScene.CreateNode(r.Renderable, native.SceneNodeKindText)
		if err == nil {
			r.Renderable.SetNativeNode(node)
		}
	}
	return r
}

// RenderSelf renders the code with syntax highlighting into the buffer.
func (r *CodeRenderable) RenderSelf(buffer *core.NativeBuffer, deltaTime float64) {
	if r.options.Code == "" {
		return
	}

	width := int(r.GetWidth())
	height := int(r.GetHeight())

	lineNumWidth := 0
	if r.options.LineNumbers {
		lines := splitLines(r.options.Code)
		lineCount := len(lines)
		lineNumWidth = 1
		for lineCount >= 10 {
			lineNumWidth++
			lineCount /= 10
		}
		lineNumWidth++ // space after number
		if lineNumWidth >= width {
			lineNumWidth = 0
		}
	}

	// Resolve renderable colors.
	fg := r.options.ForegroundColor
	bg := r.options.BackgroundColor

	// Fill background.
	x := int(r.GetX())
	y := int(r.GetY())
	w := int(r.GetWidth())
	h := int(r.GetHeight())
	if w > 0 && h > 0 && bg != nil {
		fillRect(buffer, x, y, w, h, ' ', nil, bg, 0)
	}

	// Render line numbers in the left gutter.
	if lineNumWidth > 0 {
		lineNumFg := r.options.ForegroundColor
		if lineNumFg == nil {
			lineNumFg = &core.RGBA{R: 150, G: 150, B: 150, A: 255}
		}
		lines := splitLines(r.options.Code)
		for i := 0; i < len(lines) && i < height; i++ {
			numStr := fmt.Sprintf("%d", i+1)
			// Right-align the line number within the gutter.
			numX := lineNumWidth - len(numStr) - 1
			if numX < 0 {
				numX = 0
			}
			for j, ch := range numStr {
				cx := numX + j
				if cx >= lineNumWidth || cx >= width {
					break
				}
				writeCell(buffer, cx, i, uint32(ch), lineNumFg, bg, core.TextAttributeNone)
			}
		}
	}

	lang := native.SupportedLanguage(r.options.Language)
	registry := native.NewLanguageRegistry()
	defer registry.Destroy()

	if !registry.IsSupported(lang) {
		// Fall back to plain text rendering for unsupported languages
		r.renderPlain(buffer, lineNumWidth)
		return
	}

	parser := registry.GetParser(lang)
	langObj := registry.GetLanguage(lang)

	queryStr, ok := codeQueries[lang]
	if !ok {
		r.renderPlain(buffer, lineNumWidth)
		return
	}

	query, err := native.NewTreeSitterQuery(langObj, queryStr)
	if err != nil {
		r.renderPlain(buffer, lineNumWidth)
		return
	}
	defer query.Delete()

	highlights := native.HighlightCode(parser, langObj, query, r.options.Code)
	theme := native.DefaultSyntaxTheme()

	codeLines := splitLines(r.options.Code)

	// Track which cells have been highlighted
	type highlightedCell struct {
		x, y int
	}
	highlighted := make(map[highlightedCell]bool)

	// Apply highlights
	for _, hl := range highlights {
		if hl.StartLine >= len(codeLines) {
			continue
		}
		style := theme.StyleFor(hl.CaptureName)

		// Determine the end line/col
		endLine := hl.EndLine
		if endLine >= len(codeLines) {
			endLine = len(codeLines) - 1
		}

		for y := hl.StartLine; y <= endLine && y < height; y++ {
			line := codeLines[y]
			lineLen := len(line)
			startCol := int(hl.StartCol)
			endCol := int(hl.EndCol)

			if y == hl.StartLine && y == hl.EndLine {
				// Single line highlight
				startCol = int(hl.StartCol)
				endCol = int(hl.EndCol)
			} else if y == hl.StartLine {
				startCol = int(hl.StartCol)
				endCol = lineLen
			} else if y == hl.EndLine {
				startCol = 0
				endCol = int(hl.EndCol)
			} else {
				startCol = 0
				endCol = lineLen
			}

			// Clamp to code width
			if startCol < 0 {
				startCol = 0
			}
			if endCol > lineLen {
				endCol = lineLen
			}

			for x := startCol; x < endCol; x++ {
				if x >= lineLen {
					break
				}
				bx := lineNumWidth + x
				if bx >= width {
					break
				}
				cellKey := highlightedCell{bx, y}
				if highlighted[cellKey] {
					continue
				}
				highlighted[cellKey] = true
				ch := line[x]
				writeCell(buffer, bx, y, uint32(ch), fg, bg, style)
			}
		}
	}

	// Render any characters that weren't highlighted with default style.
	for y, line := range codeLines {
		if y >= height {
			break
		}
		for x := 0; x < len(line); x++ {
			bx := lineNumWidth + x
			if bx >= width {
				break
			}
			cellKey := highlightedCell{bx, y}
			if highlighted[cellKey] {
				continue
			}
			writeCell(buffer, bx, y, uint32(line[x]), fg, bg, core.TextAttributeNone)
		}
	}

	if r.options.RenderAfter != nil {
		r.options.RenderAfter(r.Renderable, buffer, deltaTime)
	}
}

// renderPlain renders the code without syntax highlighting.
func (r *CodeRenderable) renderPlain(buffer *core.NativeBuffer, lineNumWidth int) {
	fg := r.options.ForegroundColor
	bg := r.options.BackgroundColor
	attrs := r.options.Attributes

	y := 0
	width := int(r.GetWidth())
	height := int(r.GetHeight())
	codeStart := lineNumWidth
	if codeStart >= width {
		codeStart = 0
	}

	// Fill background.
	x0 := int(r.GetX())
	y0 := int(r.GetY())
	w := int(r.GetWidth())
	h := int(r.GetHeight())
	if w > 0 && h > 0 && bg != nil {
		fillRect(buffer, x0, y0, w, h, ' ', nil, bg, 0)
	}

	// Render line numbers if enabled.
	if lineNumWidth > 0 {
		lineNumFg := fg
		if lineNumFg == nil {
			lineNumFg = &core.RGBA{R: 150, G: 150, B: 150, A: 255}
		}
		lines := splitLines(r.options.Code)
		for i := 0; i < len(lines) && i < height; i++ {
			numStr := fmt.Sprintf("%d", i+1)
			numX := lineNumWidth - len(numStr) - 1
			if numX < 0 {
				numX = 0
			}
			for j, ch := range numStr {
				cx := numX + j
				if cx >= lineNumWidth || cx >= width {
					break
				}
				writeCell(buffer, cx, i, uint32(ch), lineNumFg, bg, core.TextAttributeNone)
			}
		}
	}

	for _, line := range splitLines(r.options.Code) {
		if y >= height {
			break
		}
		x := 0
		for _, ch := range line {
			bx := codeStart + x
			if bx >= width {
				break
			}
			writeCell(buffer, bx, y, uint32(ch), fg, bg, attrs)
			x++
		}
		y++
	}
}

// SetCode sets the code content and requests a re-render.
func (r *CodeRenderable) SetCode(code string) {
	r.options.Code = code
	r.RequestRender()
}

// SetLanguage sets the programming language and requests a re-render.
func (r *CodeRenderable) SetLanguage(language string) {
	r.options.Language = language
	r.RequestRender()
}

// GetCode returns the current code content.
func (r *CodeRenderable) GetCode() string {
	return r.options.Code
}

// GetLanguage returns the current language.
func (r *CodeRenderable) GetLanguage() string {
	return r.options.Language
}

var _ = fmt.Sprintf
