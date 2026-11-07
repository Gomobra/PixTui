package core

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
)

// StyleDefinition describes a single resolved syntax style.
type StyleDefinition struct {
	FG       *RGBA
	BG       *RGBA
	Bold     bool
	Italic   bool
	Underline bool
	Dim      bool
}

// StyleDefinitionInput is the input form of a style definition, accepting
// flexible color inputs. Mirrors packages/core/src/syntax-style.ts StyleDefinitionInput.
type StyleDefinitionInput struct {
	FG       ColorInput
	BG       ColorInput
	Bold     bool
	Italic   bool
	Underline bool
	Dim      bool
}

// MergedStyle is the result of merging one or more style definitions.
type MergedStyle struct {
	FG        *RGBA
	BG        *RGBA
	Attributes uint32
}

// ThemeTokenStyle maps a list of theme scopes to a style. Mirrors
// packages/core/src/syntax-style.ts ThemeTokenStyle.
type ThemeTokenStyle struct {
	Scope []string
	Style struct {
		Foreground ColorInput
		Background ColorInput
		Bold       bool
		Italic     bool
		Underline  bool
		Dim        bool
	}
}

// ColorInput is a flexible color value accepted by style definitions.
type ColorInput interface{}

// cloneStyle returns a deep copy of a style definition.
func cloneStyle(style StyleDefinition) StyleDefinition {
	out := style
	if style.FG != nil {
		c := *style.FG
		out.FG = &c
	}
	if style.BG != nil {
		c := *style.BG
		out.BG = &c
	}
	return out
}

// ParseColor converts a flexible color input into an RGBA. Mirrors the color
// parsing behaviour of packages/core/src/lib/RGBA.ts parseColor.
func ParseColor(input ColorInput) *RGBA {
	switch v := input.(type) {
	case nil:
		return nil
	case RGBA:
		c := v
		return &c
	case *RGBA:
		if v == nil {
			return nil
		}
		c := *v
		return &c
	case string:
		return parseHexColor(v)
	case []uint8:
		if len(v) < 3 {
			return nil
		}
		a := uint8(255)
		if len(v) >= 4 {
			a = v[3]
		}
		return &RGBA{R: v[0], G: v[1], B: v[2], A: a}
	case [4]uint8:
		return &RGBA{R: v[0], G: v[1], B: v[2], A: v[3]}
	case [3]uint8:
		return &RGBA{R: v[0], G: v[1], B: v[2], A: 255}
	default:
		return nil
	}
}

// parseHexColor parses a CSS-style hex color string (#RRGGBB, #RRGGBBAA,
// RRGGBB, RRGGBBAA).
func parseHexColor(s string) *RGBA {
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 && len(s) != 8 {
		return nil
	}
	r, err := strconv.ParseUint(s[0:2], 16, 8)
	if err != nil {
		return nil
	}
	g, err := strconv.ParseUint(s[2:4], 16, 8)
	if err != nil {
		return nil
	}
	b, err := strconv.ParseUint(s[4:6], 16, 8)
	if err != nil {
		return nil
	}
	var a uint64 = 255
	if len(s) == 8 {
		a, err = strconv.ParseUint(s[6:8], 16, 8)
		if err != nil {
			return nil
		}
	}
	return &RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)}
}

// convertThemeToStyles flattens a list of theme token styles into a map of
// scope -> StyleDefinition, cloning each definition. Mirrors
// packages/core/src/syntax-style.ts convertThemeToStyles.
func convertThemeToStyles(theme []ThemeTokenStyle) map[string]StyleDefinition {
	flatStyles := make(map[string]StyleDefinition)
	for _, tokenStyle := range theme {
		def := StyleDefinition{}
		if c := ParseColor(tokenStyle.Style.Foreground); c != nil {
			clone := *c
			def.FG = &clone
		}
		if c := ParseColor(tokenStyle.Style.Background); c != nil {
			clone := *c
			def.BG = &clone
		}
		def.Bold = tokenStyle.Style.Bold
		def.Italic = tokenStyle.Style.Italic
		def.Underline = tokenStyle.Style.Underline
		def.Dim = tokenStyle.Style.Dim
		for _, scope := range tokenStyle.Scope {
			flatStyles[scope] = cloneStyle(def)
		}
	}
	return flatStyles
}

// SyntaxStyle manages syntax highlighting style definitions bound to a native
// Context. Mirrors packages/core/src/syntax-style.ts SyntaxStyle.
type SyntaxStyle struct {
	mu           sync.RWMutex
	owner        NativeResourceOwner
	destroyed    bool
	nextStyleID  uint32
	nameCache    map[string]uint32
	styleDefs    map[string]StyleDefinition
	mergedCache  map[string]MergedStyle
}

// Create creates a new SyntaxStyle owned by the given resource owner.
func CreateSyntaxStyle(owner NativeResourceOwner) (*SyntaxStyle, error) {
	if owner == nil {
		return nil, fmt.Errorf("SyntaxStyle requires an explicit resource owner")
	}
	ctx := owner.ResourceContext()
	if ctx == nil {
		return nil, fmt.Errorf("SyntaxStyle requires an explicit resource owner")
	}
	if err := ctx.AssertAlive(); err != nil {
		return nil, err
	}
	return &SyntaxStyle{
		owner:       owner,
		nameCache:   make(map[string]uint32),
		styleDefs:   make(map[string]StyleDefinition),
		mergedCache: make(map[string]MergedStyle),
	}, nil
}

// FromTheme builds a SyntaxStyle from a list of theme token styles.
func SyntaxStyleFromTheme(theme []ThemeTokenStyle, owner NativeResourceOwner) (*SyntaxStyle, error) {
	style, err := CreateSyntaxStyle(owner)
	if err != nil {
		return nil, err
	}
	defs := convertThemeToStyles(theme)
	for name, def := range defs {
		if _, err := style.registerDefinition(name, def); err != nil {
			style.Destroy()
			return nil, err
		}
	}
	return style, nil
}

// registerDefinition stores a pre-resolved StyleDefinition directly.
func (s *SyntaxStyle) registerDefinition(name string, def StyleDefinition) (uint32, error) {
	if err := s.guard(); err != nil {
		return 0, err
	}
	s.mu.Lock()
	id := s.nextStyleID
	s.nextStyleID++
	s.nameCache[name] = id
	s.styleDefs[name] = def
	s.mergedCache = make(map[string]MergedStyle)
	s.mu.Unlock()
	return id, nil
}

// FromStyles builds a SyntaxStyle from a map of name -> StyleDefinitionInput.
func CreateSyntaxStyleFromStyles(styles map[string]StyleDefinitionInput, owner NativeResourceOwner) (*SyntaxStyle, error) {
	style, err := CreateSyntaxStyle(owner)
	if err != nil {
		return nil, err
	}
	for name, styleDef := range styles {
		if id, err := style.RegisterStyle(name, styleDef); err != nil || id == 0 {
			_ = id
			style.Destroy()
			return nil, err
		}
	}
	return style, nil
}

// guard verifies the style is alive and its owner resource context is alive.
func (s *SyntaxStyle) guard() error {
	s.mu.RLock()
	destroyed := s.destroyed
	owner := s.owner
	s.mu.RUnlock()
	if destroyed {
		return fmt.Errorf("SyntaxStyle is destroyed")
	}
	if owner == nil {
		return fmt.Errorf("SyntaxStyle owner is nil")
	}
	if err := owner.ResourceContext().AssertAlive(); err != nil {
		return err
	}
	return nil
}

// RegisterStyle registers a named style definition and returns its id.
func (s *SyntaxStyle) RegisterStyle(name string, style StyleDefinitionInput) (uint32, error) {
	if err := s.guard(); err != nil {
		return 0, err
	}

	definition := StyleDefinition{
		Bold:     style.Bold,
		Italic:   style.Italic,
		Underline: style.Underline,
		Dim:      style.Dim,
	}
	if c := ParseColor(style.FG); c != nil {
		clone := *c
		definition.FG = &clone
	}
	if c := ParseColor(style.BG); c != nil {
		clone := *c
		definition.BG = &clone
	}

	attributes := CreateTextAttributes(TextAttributesOptions{
		Bold:         definition.Bold,
		Italic:       definition.Italic,
		Underline:    definition.Underline,
		Dim:          definition.Dim,
	})

	s.mu.Lock()
	id := s.nextStyleID
	s.nextStyleID++
	s.nameCache[name] = id
	s.styleDefs[name] = definition
	s.mergedCache = make(map[string]MergedStyle)
	s.mu.Unlock()

	_ = attributes
	return id, nil
}

// ResolveStyleId returns the registered id for a name, or false if unknown.
func (s *SyntaxStyle) ResolveStyleId(name string) (uint32, bool) {
	if err := s.guard(); err != nil {
		return 0, false
	}
	s.mu.RLock()
	id, ok := s.nameCache[name]
	s.mu.RUnlock()
	if ok {
		return id, true
	}
	return 0, false
}

// GetStyleId returns the id for a name, falling back to the base name for
// scoped styles (e.g. "keyword.control" -> "keyword").
func (s *SyntaxStyle) GetStyleId(name string) (uint32, bool) {
	if id, ok := s.ResolveStyleId(name); ok {
		return id, true
	}
	if idx := strings.IndexByte(name, '.'); idx >= 0 {
		baseName := name[:idx]
		return s.ResolveStyleId(baseName)
	}
	return 0, false
}

// GetStyleCount returns the number of registered styles.
func (s *SyntaxStyle) GetStyleCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.styleDefs)
}

// ClearNameCache clears the name -> id cache.
func (s *SyntaxStyle) ClearNameCache() {
	if err := s.guard(); err != nil {
		return
	}
	s.mu.Lock()
	s.nameCache = make(map[string]uint32)
	s.mu.Unlock()
}

// GetStyle returns a copy of the named style definition, falling back to the
// base name for scoped styles.
func (s *SyntaxStyle) GetStyle(name string) (StyleDefinition, bool) {
	if err := s.guard(); err != nil {
		return StyleDefinition{}, false
	}
	s.mu.RLock()
	style, ok := s.styleDefs[name]
	if !ok {
		if idx := strings.IndexByte(name, '.'); idx >= 0 {
			baseName := name[:idx]
			style, ok = s.styleDefs[baseName]
		}
	}
	s.mu.RUnlock()
	if !ok {
		return StyleDefinition{}, false
	}
	return cloneStyle(style), true
}

// MergeStyles merges the named styles into a single MergedStyle. Later styles
// override earlier ones. Results are cached.
func (s *SyntaxStyle) MergeStyles(styleNames ...string) (MergedStyle, bool) {
	if err := s.guard(); err != nil {
		return MergedStyle{}, false
	}
	cacheKey := strings.Join(styleNames, ":")
	s.mu.RLock()
	cached, ok := s.mergedCache[cacheKey]
	s.mu.RUnlock()
	if ok {
		return cloneMerged(cached), true
	}

	def := StyleDefinition{}
	for _, name := range styleNames {
		style, ok := s.GetStyle(name)
		if !ok {
			continue
		}
		if style.FG != nil {
			def.FG = style.FG
		}
		if style.BG != nil {
			def.BG = style.BG
		}
		if style.Bold {
			def.Bold = true
		}
		if style.Italic {
			def.Italic = true
		}
		if style.Underline {
			def.Underline = true
		}
		if style.Dim {
			def.Dim = true
		}
	}

	attributes := CreateTextAttributes(TextAttributesOptions{
		Bold:      def.Bold,
		Italic:    def.Italic,
		Underline: def.Underline,
		Dim:       def.Dim,
	})

	merged := MergedStyle{
		FG:         def.FG,
		BG:         def.BG,
		Attributes: attributes,
	}

	s.mu.Lock()
	s.mergedCache[cacheKey] = merged
	s.mu.Unlock()

	return cloneMerged(merged), true
}

func cloneMerged(m MergedStyle) MergedStyle {
	out := m
	if m.FG != nil {
		c := *m.FG
		out.FG = &c
	}
	if m.BG != nil {
		c := *m.BG
		out.BG = &c
	}
	return out
}

// ClearCache clears the merged style cache.
func (s *SyntaxStyle) ClearCache() {
	if err := s.guard(); err != nil {
		return
	}
	s.mu.Lock()
	s.mergedCache = make(map[string]MergedStyle)
	s.mu.Unlock()
}

// GetCacheSize returns the number of cached merged styles.
func (s *SyntaxStyle) GetCacheSize() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.mergedCache)
}

// GetAllStyles returns copies of all registered style definitions.
func (s *SyntaxStyle) GetAllStyles() map[string]StyleDefinition {
	if err := s.guard(); err != nil {
		return nil
	}
	s.mu.RLock()
	out := make(map[string]StyleDefinition, len(s.styleDefs))
	for name, style := range s.styleDefs {
		out[name] = cloneStyle(style)
	}
	s.mu.RUnlock()
	return out
}

// GetRegisteredNames returns the names of all registered styles.
func (s *SyntaxStyle) GetRegisteredNames() []string {
	if err := s.guard(); err != nil {
		return nil
	}
	s.mu.RLock()
	names := make([]string, 0, len(s.styleDefs))
	for name := range s.styleDefs {
		names = append(names, name)
	}
	s.mu.RUnlock()
	return names
}

// Destroy releases the syntax style and its caches.
func (s *SyntaxStyle) Destroy() {
	s.mu.Lock()
	if s.destroyed {
		s.mu.Unlock()
		return
	}
	s.destroyed = true
	s.nameCache = make(map[string]uint32)
	s.styleDefs = make(map[string]StyleDefinition)
	s.mergedCache = make(map[string]MergedStyle)
	owner := s.owner
	ctx := owner.ResourceContext()
	s.mu.Unlock()
	if ctx != nil {
		_ = ctx.AssertAlive()
	}
}
