package native

// Tree-sitter language bindings for syntax highlighting.
// These provide the tree-sitter grammars for supported languages.

import (
	"unsafe"

	tsgo "github.com/tree-sitter/tree-sitter-go/bindings/go"
	tsjs "github.com/tree-sitter/tree-sitter-javascript/bindings/go"
	tspy "github.com/tree-sitter/tree-sitter-python/bindings/go"
)

// SupportedLanguage is the identifier for a supported tree-sitter language.
type SupportedLanguage string

const (
	LanguageGo         SupportedLanguage = "go"
	LanguageJavaScript SupportedLanguage = "javascript"
	LanguagePython     SupportedLanguage = "python"
	LanguageText       SupportedLanguage = "text" // plain text, no highlighting
)

// LanguageRegistry maps language identifiers to tree-sitter languages.
type LanguageRegistry struct {
	languages map[SupportedLanguage]*TreeSitterLanguage
	parsers   map[SupportedLanguage]*TreeSitterParser
}

// NewLanguageRegistry creates a new language registry with all supported languages.
func NewLanguageRegistry() *LanguageRegistry {
	registry := &LanguageRegistry{
		languages: make(map[SupportedLanguage]*TreeSitterLanguage),
		parsers:   make(map[SupportedLanguage]*TreeSitterParser),
	}

	// Register Go language
	registry.RegisterLanguage(LanguageGo, tsgo.Language)
	// Register JavaScript language
	registry.RegisterLanguage(LanguageJavaScript, tsjs.Language)
	// Register Python language
	registry.RegisterLanguage(LanguagePython, tspy.Language)

	return registry
}

// RegisterLanguage registers a tree-sitter language for the given identifier.
func (r *LanguageRegistry) RegisterLanguage(id SupportedLanguage, langFn func() unsafe.Pointer) {
	lang := NewTreeSitterLanguage(langFn)
	parser := NewTreeSitterParser()
	parser.SetLanguage(lang)
	r.languages[id] = lang
	r.parsers[id] = parser
}

// GetLanguage returns the tree-sitter language for the given identifier.
func (r *LanguageRegistry) GetLanguage(id SupportedLanguage) *TreeSitterLanguage {
	return r.languages[id]
}

// GetParser returns the tree-sitter parser for the given identifier.
func (r *LanguageRegistry) GetParser(id SupportedLanguage) *TreeSitterParser {
	return r.parsers[id]
}

// IsSupported returns true if the language is supported.
func (r *LanguageRegistry) IsSupported(id SupportedLanguage) bool {
	_, ok := r.languages[id]
	return ok
}

// Destroy cleans up all parser resources.
func (r *LanguageRegistry) Destroy() {
	for _, parser := range r.parsers {
		parser.Destroy()
	}
}

var _ = unsafe.Pointer(nil)
