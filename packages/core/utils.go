package core

import (
	"strings"
)

// TextAttributesOptions configures text style flags packed into a single
// attribute bitfield. Mirrors packages/core/src/utils.ts createTextAttributes options.
type TextAttributesOptions struct {
	Bold         bool
	Italic       bool
	Underline    bool
	Dim          bool
	Blink        bool
	Inverse      bool
	Reverse      bool
	Hidden       bool
	Strikethrough bool
}

const (
	attributeBaseMask    = 0xff
	linkIDShift          = 8
	linkIDPayloadMask    = 0xffffff
)

// CreateTextAttributes packs text style flags into a single attribute bitfield.
// Mirrors packages/core/src/utils.ts createTextAttributes.
func CreateTextAttributes(opts TextAttributesOptions) uint32 {
	var attributes uint32 = TextAttributeNone
	if opts.Bold {
		attributes |= TextAttributeBold
	}
	if opts.Italic {
		attributes |= TextAttributeItalic
	}
	if opts.Underline {
		attributes |= TextAttributeUnderline
	}
	if opts.Dim {
		attributes |= TextAttributeDim
	}
	if opts.Blink {
		attributes |= TextAttributeBlink
	}
	if opts.Inverse || opts.Reverse {
		attributes |= TextAttributeInverse
	}
	if opts.Hidden {
		attributes |= TextAttributeHidden
	}
	if opts.Strikethrough {
		attributes |= TextAttributeStrikethrough
	}
	return attributes
}

// AttributesWithLink packs a link id into the high bits (bits 8-31) of an
// attribute bitfield. Mirrors packages/core/src/utils.ts attributesWithLink.
func AttributesWithLink(baseAttributes uint32, linkId uint32) uint32 {
	base := baseAttributes & attributeBaseMask
	linkBits := (linkId & linkIDPayloadMask) << linkIDShift
	return base | linkBits
}

// GetLinkId extracts the link id stored in the high bits of an attribute
// bitfield. Mirrors packages/core/src/utils.ts getLinkId.
func GetLinkId(attributes uint32) uint32 {
	return (attributes >> linkIDShift) & linkIDPayloadMask
}

// buildTreeLines recursively renders a renderable subtree as indented text.
func buildTreeLines(node *Renderable, prefix, parentPrefix string, isLastChild bool, depth, maxDepth int, lines *[]string) {
	if depth >= maxDepth {
		*lines = append(*lines, prefix+node.ID()+" ... (max depth reached)")
		return
	}

	*lines = append(*lines, prefix+node.ID())

	children := node.Children()
	if len(children) > 0 {
		lastChildIndex := len(children) - 1
		for i, child := range children {
			childIsLast := i == lastChildIndex
			connector := "├── "
			if childIsLast {
				connector = "└── "
			}
			childPrefix := parentPrefix
			if isLastChild {
				childPrefix += "    "
			} else {
				childPrefix += "│   "
			}
			if cr, ok := child.(*Renderable); ok {
				buildTreeLines(cr, childPrefix+connector, childPrefix, childIsLast, depth+1, maxDepth, lines)
			} else {
				*lines = append(*lines, childPrefix+connector+"<renderable>")
			}
		}
	}
}

// VisualizeRenderableTree returns a text rendering of a renderable subtree,
// useful for debugging. Mirrors packages/core/src/utils.ts visualizeRenderableTree.
func VisualizeRenderableTree(root *Renderable, maxDepth int) string {
	if maxDepth <= 0 {
		maxDepth = 10
	}
	var lines []string
	buildTreeLines(root, "", "", true, 0, maxDepth, &lines)
	return strings.Join(lines, "\n")
}
