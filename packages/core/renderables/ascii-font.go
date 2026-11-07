package renderables

import (
	_ "embed"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/Gomobra/PixTui/packages/core"
	"github.com/Gomobra/PixTui/packages/native"
)

//go:embed fonts/block.json
var blockFontRaw []byte

// fontSegment is a run of glyphs within a single font line that share a color
// index. It mirrors the parsed <cN>...</cN> color-tag segments from the cfonts
// font definitions.
type fontSegment struct {
	Text       string `json:"text"`
	ColorIndex int    `json:"colorIndex"`
}

// asciiFontDefinition is the raw block font definition loaded from the
// embedded fonts/block.json (cfonts format).
type asciiFontDefinition struct {
	Name            string              `json:"name"`
	Lines           int                 `json:"lines"`
	LetterspaceSize int                 `json:"letterspace_size"`
	Colors          int                 `json:"colors"`
	Chars           map[string][]string `json:"chars"`
}

// asciiCharDef is a single character definition: one segment list per font
// line.
type asciiCharDef [][]fontSegment

// parsedFont is a block font with the color tags pre-parsed into segments so
// they can be rendered cell by cell.
type parsedFont struct {
	def   *asciiFontDefinition
	chars map[string]asciiCharDef
}

// colorTagRegex matches cfonts color tags such as <c1>...</c1>.
var colorTagRegex = regexp.MustCompile(`<c(\d+)>(.*?)</c\d+>`)

// blockFont is the embedded "block" FIGlet-style font. It covers A-Z, 0-9 and
// common punctuation/space.
var blockFont *parsedFont

func init() {
	if def, err := parseBlockFont(blockFontRaw); err == nil {
		blockFont = def
	}
}

// parseBlockFont decodes a cfonts font definition and pre-parses its color
// tags into segments.
func parseBlockFont(data []byte) (*parsedFont, error) {
	var def asciiFontDefinition
	if err := json.Unmarshal(data, &def); err != nil {
		return nil, err
	}
	if def.Lines < 1 {
		def.Lines = 1
	}
	if def.LetterspaceSize < 0 {
		def.LetterspaceSize = 0
	}
	chars := make(map[string]asciiCharDef, len(def.Chars))
	for ch, lines := range def.Chars {
		cd := make(asciiCharDef, len(lines))
		for i, line := range lines {
			cd[i] = parseFontSegments(line)
		}
		chars[ch] = cd
	}
	return &parsedFont{def: &def, chars: chars}, nil
}

// parseFontSegments splits a font line into color-tagged segments. Text outside
// of tags uses color index 0.
func parseFontSegments(line string) []fontSegment {
	var segments []fontSegment
	lastIndex := 0
	for _, m := range colorTagRegex.FindAllStringSubmatchIndex(line, -1) {
		if m[0] > lastIndex {
			if plain := line[lastIndex:m[0]]; plain != "" {
				segments = append(segments, fontSegment{Text: plain, ColorIndex: 0})
			}
		}
		colorIndex := 0
		if m[2] < m[3] {
			if n, err := strconv.Atoi(line[m[2]:m[3]]); err == nil {
				colorIndex = n - 1
				if colorIndex < 0 {
					colorIndex = 0
				}
			}
		}
		segments = append(segments, fontSegment{Text: line[m[4]:m[5]], ColorIndex: colorIndex})
		lastIndex = m[1]
	}
	if lastIndex < len(line) {
		if remaining := line[lastIndex:]; remaining != "" {
			segments = append(segments, fontSegment{Text: remaining, ColorIndex: 0})
		}
	}
	return segments
}

// measureText returns the width (in cells) and height (in lines) of the given
// text when rendered with this font, matching the cfonts measurement logic.
func (f *parsedFont) measureText(text string) (width, height int) {
	if f == nil {
		return 0, 0
	}
	height = f.def.Lines
	if height < 1 {
		height = 1
	}
	runes := []rune(text)
	currentX := 0
	for i := 0; i < len(runes); i++ {
		currentX += f.charWidth(runes[i])
		if i < len(runes)-1 {
			currentX += f.def.LetterspaceSize
		}
	}
	return currentX, height
}

// charWidth returns the cell width of a single character, falling back to the
// space character width (or 1) when the character is not defined in the font.
func (f *parsedFont) charWidth(ch rune) int {
	charDef := f.chars[strings.ToUpper(string(ch))]
	if charDef != nil && len(charDef) > 0 && len(charDef[0]) > 0 {
		width := 0
		for _, seg := range charDef[0] {
			width += utf8.RuneCountInString(seg.Text)
		}
		return width
	}
	spaceDef := f.chars[" "]
	if spaceDef != nil && len(spaceDef) > 0 && len(spaceDef[0]) > 0 {
		width := 0
		for _, seg := range spaceDef[0] {
			width += utf8.RuneCountInString(seg.Text)
		}
		return width
	}
	return 1
}

// renderToBuffer draws the text into buffer starting at (x, y), one glyph cell
// at a time, using the font glyphs. Only non-space glyph cells are written so
// that the background shows through in the gaps.
func (f *parsedFont) renderToBuffer(buffer *core.NativeBuffer, x, y int, text string, color []core.RGBA, backgroundColor core.RGBA) {
	if f == nil || buffer == nil {
		return
	}
	if len(color) == 0 {
		color = []core.RGBA{core.RGBA(native.White)}
	}
	bw := buffer.Width()
	bh := buffer.Height()
	runes := []rune(text)
	currentX := x
	for i := 0; i < len(runes); i++ {
		ch := runes[i]
		charDef := f.chars[strings.ToUpper(string(ch))]
		if charDef == nil {
			currentX += f.charWidth(ch)
			if i < len(runes)-1 {
				currentX += f.def.LetterspaceSize
			}
			continue
		}
		charWidth := 0
		if len(charDef) > 0 && len(charDef[0]) > 0 {
			for _, seg := range charDef[0] {
				charWidth += utf8.RuneCountInString(seg.Text)
			}
		}
		if currentX >= bw {
			break
		}
		if currentX+charWidth < 0 {
			currentX += charWidth + f.def.LetterspaceSize
			continue
		}
		for lineIdx := 0; lineIdx < f.def.Lines && lineIdx < len(charDef); lineIdx++ {
			renderY := y + lineIdx
			if renderY < 0 || renderY >= bh {
				continue
			}
			segmentX := currentX
			for _, seg := range charDef[lineIdx] {
				segColor := color[0]
				if seg.ColorIndex >= 0 && seg.ColorIndex < len(color) {
					segColor = color[seg.ColorIndex]
				}
				segRunes := []rune(seg.Text)
				for ci := 0; ci < len(segRunes); ci++ {
					renderX := segmentX + ci
					if renderX < 0 || renderX >= bw {
						continue
					}
					g := segRunes[ci]
					if g != ' ' {
						writeCell(buffer, renderX, renderY, uint32(g), &segColor, &backgroundColor, 0)
					}
				}
				segmentX += len(segRunes)
			}
		}
		currentX += charWidth
		if i < len(runes)-1 {
			currentX += f.def.LetterspaceSize
		}
	}
}

// getFont returns the parsed font for the given name. Currently only "block"
// is embedded; unknown names fall back to "block" so that something still
// renders.
func getFont(name string) *parsedFont {
	if strings.EqualFold(name, "block") || name == "" {
		return blockFont
	}
	return blockFont
}

// ASCIIFontOptions contains options for ASCIIFontRenderable.
type ASCIIFontOptions struct {
	core.RenderableOptions[core.Renderable]
	// Text is the text to render in the ASCII font.
	Text string
	// Font is the font name to use (currently "block" is supported).
	Font string
	// Color is the foreground color(s) of the glyphs. A single color or a list
	// of colors matched against the font's per-segment color indices.
	Color []core.RGBA
	// BackgroundColor is the background color of the font buffer. When
	// transparent the parent buffer shows through the empty cells.
	BackgroundColor *core.RGBA
}

// ASCIIFontRenderable renders text using a FIGlet-style ASCII font. It renders
// each character cell by cell from the font glyphs into an offscreen buffer
// which is then composited into the parent buffer.
type ASCIIFontRenderable struct {
	*core.Renderable
	options         ASCIIFontOptions
	frameBuffer     *core.NativeBuffer
	text            string
	font            string
	color           []core.RGBA
	backgroundColor core.RGBA
}

// NewASCIIFontRenderable creates a new ASCIIFontRenderable. The renderable
// dimensions are derived from the measured text so the buffer exactly fits the
// glyphs.
func NewASCIIFontRenderable(ctx *core.RenderContext, opts ASCIIFontOptions) *ASCIIFontRenderable {
	text := opts.Text
	font := opts.Font
	if font == "" {
		font = "block"
	}
	fontDef := getFont(font)
	if fontDef == nil {
		fontDef = blockFont
	}
	width, height := fontDef.measureText(text)
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}

	base := core.NewRenderable(ctx, opts.ID, core.RenderableOptions[core.Renderable]{
		Width:   opts.Width,
		Height:  opts.Height,
		Visible: opts.Visible,
		ZIndex:  opts.ZIndex,
	})
	r := &ASCIIFontRenderable{
		Renderable:      base,
		options:         opts,
		text:            text,
		font:            font,
		frameBuffer:     core.NewNativeBuffer(width, height),
		color:           opts.Color,
		backgroundColor: core.RGBA(native.Transparent),
	}
	if len(r.color) == 0 {
		r.color = []core.RGBA{core.RGBA(native.White)}
	}
	if opts.BackgroundColor != nil {
		r.backgroundColor = *opts.BackgroundColor
	}
	r.SetWidth(float32(width))
	r.SetHeight(float32(height))

	if ctx != nil && ctx.NativeScene != nil {
		if node, err := ctx.NativeScene.CreateNode(r.Renderable, native.SceneNodeKindText); err == nil {
			r.Renderable.SetNativeNode(node)
		}
	}
	r.renderFontToBuffer()
	return r
}

// Text returns the rendered text.
func (r *ASCIIFontRenderable) Text() string {
	return r.text
}

// SetText sets the text, resizes the buffer to the new measurement and
// re-renders the font.
func (r *ASCIIFontRenderable) SetText(text string) {
	r.text = text
	r.options.Text = text
	r.resizeToText()
	r.renderFontToBuffer()
	r.RequestRender()
}

// Font returns the font name.
func (r *ASCIIFontRenderable) Font() string {
	return r.font
}

// SetFont sets the font, resizes the buffer to the new measurement and
// re-renders the font.
func (r *ASCIIFontRenderable) SetFont(font string) {
	if font == "" {
		font = "block"
	}
	r.font = font
	r.options.Font = font
	r.resizeToText()
	r.renderFontToBuffer()
	r.RequestRender()
}

// Color returns a copy of the foreground color(s).
func (r *ASCIIFontRenderable) Color() []core.RGBA {
	out := make([]core.RGBA, len(r.color))
	copy(out, r.color)
	return out
}

// SetColor sets the foreground color(s) and re-renders the font.
func (r *ASCIIFontRenderable) SetColor(color []core.RGBA) {
	if len(color) == 0 {
		color = []core.RGBA{core.RGBA(native.White)}
	}
	r.color = color
	r.options.Color = color
	r.renderFontToBuffer()
	r.RequestRender()
}

// BackgroundColor returns the background color.
func (r *ASCIIFontRenderable) BackgroundColor() core.RGBA {
	return r.backgroundColor
}

// SetBackgroundColor sets the background color and re-renders the font.
func (r *ASCIIFontRenderable) SetBackgroundColor(color core.RGBA) {
	r.backgroundColor = color
	r.renderFontToBuffer()
	r.RequestRender()
}

// Options returns the ASCII font options.
func (r *ASCIIFontRenderable) Options() ASCIIFontOptions {
	return r.options
}

// FrameBuffer returns the internal offscreen buffer.
func (r *ASCIIFontRenderable) FrameBuffer() *core.NativeBuffer {
	return r.frameBuffer
}

// resizeToText resizes the internal buffer to match the current text and font
// measurement.
func (r *ASCIIFontRenderable) resizeToText() {
	fontDef := getFont(r.font)
	if fontDef == nil {
		fontDef = blockFont
	}
	width, height := fontDef.measureText(r.text)
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	r.frameBuffer = core.NewNativeBuffer(width, height)
	r.SetWidth(float32(width))
	r.SetHeight(float32(height))
}

// renderFontToBuffer clears the internal buffer, fills it with the background
// color and draws each character cell by cell from the font glyphs.
func (r *ASCIIFontRenderable) renderFontToBuffer() {
	if r.IsDestroyed() || r.frameBuffer == nil {
		return
	}
	fb := r.frameBuffer
	fb.Clear()
	fillRect(fb, 0, 0, fb.Width(), fb.Height(), ' ', nil, &r.backgroundColor, 0)
	fontDef := getFont(r.font)
	if fontDef == nil {
		fontDef = blockFont
	}
	if fontDef == nil {
		// Graceful fallback: render the text as plain characters.
		x := 0
		for _, g := range []rune(r.text) {
			if x >= fb.Width() {
				break
			}
			writeCell(fb, x, 0, uint32(g), &r.color[0], &r.backgroundColor, 0)
			x++
		}
		return
	}
	fontDef.renderToBuffer(fb, 0, 0, r.text, r.color, r.backgroundColor)
}

// RenderSelf renders the ASCII font into the internal buffer and composites it
// into the parent buffer at the renderable's screen position. Empty/transparent
// cells are skipped so the parent shows through.
func (r *ASCIIFontRenderable) RenderSelf(buffer *core.NativeBuffer, deltaTime float64) {
	if r.options.RenderBefore != nil {
		r.options.RenderBefore(r.Renderable, buffer, deltaTime)
	}
	if r.IsDestroyed() || !r.IsVisible() || r.frameBuffer == nil {
		if r.options.RenderAfter != nil {
			r.options.RenderAfter(r.Renderable, buffer, deltaTime)
		}
		return
	}
	r.renderFontToBuffer()
	compositeBuffer(buffer, r.frameBuffer, int(r.GetX()), int(r.GetY()), true)
	if r.options.RenderAfter != nil {
		r.options.RenderAfter(r.Renderable, buffer, deltaTime)
	}
}
