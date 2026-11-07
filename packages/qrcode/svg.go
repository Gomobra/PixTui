package qrcode

import (
	"fmt"
	"strings"
)

// ---------------------------------------------------------------------------
// SVG and terminal rendering (mirrors @pixtui/qrcode/src/lib/qrcode.render.ts)
// ---------------------------------------------------------------------------

// TerminalRenderOptions controls terminal rendering of a QR code matrix.
type TerminalRenderOptions struct {
	// Border is the quiet zone in modules. ISO requires at least 4 for QR Code.
	Border int
	// ANSI renders using ANSI background colors for scanner-facing output.
	ANSI bool
	// Invert swaps light and dark output.
	Invert bool
}

// NormalizedTerminalRenderOptions is TerminalRenderOptions with defaults applied.
type NormalizedTerminalRenderOptions struct {
	Border int
	ANSI   bool
	Invert bool
}

// normalizeTerminalOptions normalizes terminal render options with defaults.
func normalizeTerminalOptions(options interface{}, defaultBorder, minimumBorder int) (NormalizedTerminalRenderOptions, error) {
	switch o := options.(type) {
	case int:
		if o < minimumBorder {
			return NormalizedTerminalRenderOptions{}, &BorderTooSmallError{Minimum: minimumBorder}
		}
		return NormalizedTerminalRenderOptions{Border: o, ANSI: false, Invert: false}, nil
	case TerminalRenderOptions:
		border := o.Border
		if border == 0 {
			border = defaultBorder
		}
		if border < minimumBorder {
			return NormalizedTerminalRenderOptions{}, &BorderTooSmallError{Minimum: minimumBorder}
		}
		return NormalizedTerminalRenderOptions{Border: border, ANSI: o.ANSI, Invert: o.Invert}, nil
	case nil:
		return NormalizedTerminalRenderOptions{Border: defaultBorder, ANSI: false, Invert: false}, nil
	default:
		return NormalizedTerminalRenderOptions{}, fmt.Errorf("qrcode: invalid terminal render options type %T", options)
	}
}

// BorderTooSmallError is returned when a border is below the minimum.
type BorderTooSmallError struct {
	Minimum int
}

func (e *BorderTooSmallError) Error() string {
	return "qrcode: terminal border/quiet zone must be at least " + string(rune('0'+e.Minimum)) + " modules"
}

// matrixToSvg renders a boolean matrix as an SVG string.
func matrixToSvg(matrix [][]bool, border, minimumBorder, moduleSize int, lightColor, darkColor string) (string, error) {
	if border < minimumBorder {
		return "", &BorderTooSmallError{Minimum: minimumBorder}
	}
	if moduleSize <= 0 {
		return "", fmt.Errorf("qrcode: moduleSize must be positive")
	}
	size := len(matrix)
	dimension := (size + border*2) * moduleSize

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<svg xmlns=\"http://www.w3.org/2000/svg\" viewBox=\"0 0 %d %d\" width=\"%d\" height=\"%d\" shape-rendering=\"crispEdges\">",
		dimension, dimension, dimension, dimension))
	sb.WriteString(fmt.Sprintf("<rect width=\"100%%\" height=\"100%%\" fill=\"%s\"/>", escapeXml(lightColor)))

	var path strings.Builder
	for y := 0; y < size; y++ {
		x := 0
		for x < size {
			if !matrix[y][x] {
				x++
				continue
			}
			startX := x
			for x < size && matrix[y][x] {
				x++
			}
			rectX := (startX + border) * moduleSize
			rectY := (y + border) * moduleSize
			rectW := (x - startX) * moduleSize
			path.WriteString(fmt.Sprintf("M%d,%dh%dv%dh-%dz", rectX, rectY, rectW, moduleSize, rectW))
		}
	}

	if path.Len() > 0 {
		sb.WriteString(fmt.Sprintf("<path d=\"%s\" fill=\"%s\"/>", escapeXml(path.String()), escapeXml(darkColor)))
	}
	sb.WriteString("</svg>")
	return sb.String(), nil
}

// ToSVGString renders the QR code as an SVG string.
func (q *QRCode) ToSVGString(options ...interface{}) (string, error) {
	border := 4
	moduleSize := 1
	lightColor := "#FFFFFF"
	darkColor := "#000000"

	if len(options) > 0 {
		if opt, ok := options[0].(SVGRenderOptions); ok {
			if opt.Border != 0 {
				border = opt.Border
			}
			if opt.ModuleSize != 0 {
				moduleSize = opt.ModuleSize
			}
			if opt.LightColor != "" {
				lightColor = opt.LightColor
			}
			if opt.DarkColor != "" {
				darkColor = opt.DarkColor
			}
		}
	}

	return matrixToSvg(q.modules, border, 4, moduleSize, lightColor, darkColor)
}

// SVGRenderOptions controls SVG rendering of a QR code.
type SVGRenderOptions struct {
	Border     int
	ModuleSize int
	LightColor string
	DarkColor  string
}

// matrixToTerminal renders a boolean matrix as terminal text.
func matrixToTerminal(matrix [][]bool, options NormalizedTerminalRenderOptions) string {
	size := len(matrix)
	darkText := "  "
	lightText := "  "
	if options.Invert {
		darkText, lightText = lightText, darkText
	}
	darkAnsi := "\x1b[40m  "
	lightAnsi := "\x1b[47m  "
	if options.Invert {
		darkAnsi, lightAnsi = lightAnsi, darkAnsi
	}

	var lines []string
	for y := -options.Border; y < size+options.Border; y++ {
		var line strings.Builder
		for x := -options.Border; x < size+options.Border; x++ {
			dark := x >= 0 && y >= 0 && x < size && y < size && matrix[y][x]
			if options.ANSI {
				if dark {
					line.WriteString(darkAnsi)
				} else {
					line.WriteString(lightAnsi)
				}
			} else {
				if dark {
					line.WriteString(darkText)
				} else {
					line.WriteString(lightText)
				}
			}
		}
		if options.ANSI {
			line.WriteString("\x1b[0m")
		}
		lines = append(lines, line.String())
	}
	return strings.Join(lines, "\n")
}

// ToTerminalString renders the QR code as terminal text.
func (q *QRCode) ToTerminalString(options ...interface{}) (string, error) {
	border := 4
	ansi := false
	invert := false

	if len(options) > 0 {
		switch opt := options[0].(type) {
		case int:
			border = opt
		case TerminalRenderOptions:
			border = opt.Border
			if border == 0 {
				border = 4
			}
			ansi = opt.ANSI
			invert = opt.Invert
		}
	}

	normalized, err := normalizeTerminalOptions(border, 4, 4)
	if err != nil {
		return "", err
	}
	normalized.ANSI = ansi
	normalized.Invert = invert

	return matrixToTerminal(q.modules, normalized), nil
}

// escapeXml escapes XML special characters.
func escapeXml(value string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"'", "&apos;",
		"\"", "&quot;",
	)
	return replacer.Replace(value)
}
