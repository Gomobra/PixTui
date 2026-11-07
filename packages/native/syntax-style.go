package native

import (
	"sync"
)

// SyntaxStyle represents a syntax highlighting style
type SyntaxStyle struct {
	mu         sync.RWMutex
	name       string
	foreground RGBA
	background RGBA
	attributes uint32
}

// NewSyntaxStyle creates a new syntax style
func NewSyntaxStyle(name string, fg, bg RGBA, attributes uint32) *SyntaxStyle {
	return &SyntaxStyle{
		name:       name,
		foreground: fg,
		background: bg,
		attributes: attributes,
	}
}

// Name returns the style name
func (s *SyntaxStyle) Name() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.name
}

// Foreground returns the foreground color
func (s *SyntaxStyle) Foreground() RGBA {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.foreground
}

// Background returns the background color
func (s *SyntaxStyle) Background() RGBA {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.background
}

// Attributes returns the text attributes
func (s *SyntaxStyle) Attributes() uint32 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.attributes
}

// SyntaxStyleManager manages syntax styles
type SyntaxStyleManager struct {
	mu     sync.RWMutex
	styles map[string]*SyntaxStyle
}

// NewSyntaxStyleManager creates a new syntax style manager
func NewSyntaxStyleManager() *SyntaxStyleManager {
	return &SyntaxStyleManager{
		styles: make(map[string]*SyntaxStyle),
	}
}

// AddStyle adds a syntax style
func (m *SyntaxStyleManager) AddStyle(style *SyntaxStyle) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.styles[style.Name()] = style
}

// GetStyle returns a syntax style by name
func (m *SyntaxStyleManager) GetStyle(name string) (*SyntaxStyle, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	style, ok := m.styles[name]
	return style, ok
}

// RemoveStyle removes a syntax style
func (m *SyntaxStyleManager) RemoveStyle(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.styles, name)
}

// StyleCount returns the number of styles
func (m *SyntaxStyleManager) StyleCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.styles)
}
