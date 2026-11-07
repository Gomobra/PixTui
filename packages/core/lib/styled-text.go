package lib

import (
	"github.com/Gomobra/PixTui/packages/core"
)

// StyledText is a piece of text with associated styling
type StyledText struct {
	Chunks []StyledTextChunk
}

// StyledTextChunk is a chunk of styled text
type StyledTextChunk struct {
	Text      string
	Foreground *core.RGBA
	Background *core.RGBA
	Attributes uint32
}

// NewStyledText creates a StyledText from a plain string
func NewStyledText(text string) *StyledText {
	return &StyledText{
		Chunks: []StyledTextChunk{
			{Text: text},
		},
	}
}

// String returns the plain text content
func (s *StyledText) String() string {
	var result string
	for _, chunk := range s.Chunks {
		result += chunk.Text
	}
	return result
}
