package testing

import (
	"github.com/Gomobra/PixTui/packages/core"
)

// TerminalCapabilitiesOverrides holds optional overrides for terminal
// capabilities used in tests. Only non-nil pointer fields are applied over the
// defaults, mirroring the TypeScript Partial<TerminalCapabilities> semantics.
type TerminalCapabilitiesOverrides struct {
	Terminal *core.TerminalInfo

	KittyKeyboard             *bool
	KittyGraphics             *bool
	RGB                       *bool
	ANSI256                   *bool
	Unicode                   *core.WidthMethod
	SGRPixels                 *bool
	ColorSchemeUpdates        *bool
	ExplicitWidth             *bool
	ScaledText                *bool
	Sixel                     *bool
	FocusTracking             *bool
	Sync                      *bool
	BracketedPaste            *bool
	Hyperlinks                *bool
	OSC52                     *bool
	OSC52Support              *core.TerminalCapabilityState
	Notifications             *bool
	ExplicitCursorPositioning *bool
	Remote                    *bool
	ProgramStatus             *bool
	Multiplexer               *core.TerminalMultiplexer
	ImageProtocol             *core.ImageRenderProtocol
}

// setIfSet copies *src into *dst when src is non-nil. It models the TypeScript
// spread of Partial fields, where only provided overrides take effect.
func setIfSet[T any](dst *T, src *T) {
	if src != nil {
		*dst = *src
	}
}

// CreateTerminalCapabilities returns a TerminalCapabilities populated with test
// defaults (all capabilities disabled, empty terminal info) with the given
// overrides applied. Mirrors the TypeScript createTerminalCapabilities helper.
func CreateTerminalCapabilities(overrides TerminalCapabilitiesOverrides) *core.TerminalCapabilities {
	caps := &core.TerminalCapabilities{
		KittyKeyboard:           false,
		KittyGraphics:           false,
		RGB:                     false,
		ANSI256:                 false,
		Unicode:                 core.WidthMethodUnicode,
		SGRPixels:               false,
		ColorSchemeUpdates:      false,
		ExplicitWidth:           false,
		ScaledText:              false,
		Sixel:                   false,
		FocusTracking:           false,
		Sync:                    false,
		BracketedPaste:          false,
		Hyperlinks:              false,
		OSC52:                   false,
		OSC52Support:            core.TerminalCapabilityUnknown,
		Notifications:           false,
		ExplicitCursorPositioning: false,
		Remote:                  false,
		ProgramStatus:           false,
		Multiplexer:             core.TerminalMultiplexerNone,
		ImageProtocol:           core.ImageRenderProtocolAuto,
		Terminal: core.TerminalInfo{
			Name:          "",
			Version:       "",
			FromXTVersion: false,
		},
	}
	if overrides.Terminal != nil {
		caps.Terminal = *overrides.Terminal
	}

	setIfSet(&caps.KittyKeyboard, overrides.KittyKeyboard)
	setIfSet(&caps.KittyGraphics, overrides.KittyGraphics)
	setIfSet(&caps.RGB, overrides.RGB)
	setIfSet(&caps.ANSI256, overrides.ANSI256)
	setIfSet(&caps.Unicode, overrides.Unicode)
	setIfSet(&caps.SGRPixels, overrides.SGRPixels)
	setIfSet(&caps.ColorSchemeUpdates, overrides.ColorSchemeUpdates)
	setIfSet(&caps.ExplicitWidth, overrides.ExplicitWidth)
	setIfSet(&caps.ScaledText, overrides.ScaledText)
	setIfSet(&caps.Sixel, overrides.Sixel)
	setIfSet(&caps.FocusTracking, overrides.FocusTracking)
	setIfSet(&caps.Sync, overrides.Sync)
	setIfSet(&caps.BracketedPaste, overrides.BracketedPaste)
	setIfSet(&caps.Hyperlinks, overrides.Hyperlinks)
	setIfSet(&caps.OSC52, overrides.OSC52)
	setIfSet(&caps.OSC52Support, overrides.OSC52Support)
	setIfSet(&caps.Notifications, overrides.Notifications)
	setIfSet(&caps.ExplicitCursorPositioning, overrides.ExplicitCursorPositioning)
	setIfSet(&caps.Remote, overrides.Remote)
	setIfSet(&caps.ProgramStatus, overrides.ProgramStatus)
	setIfSet(&caps.Multiplexer, overrides.Multiplexer)
	setIfSet(&caps.ImageProtocol, overrides.ImageProtocol)

	return caps
}

// SetRendererCapabilities installs the given terminal capabilities on the
// renderer's render context and returns them. Mirrors the TypeScript
// setRendererCapabilities helper.
func SetRendererCapabilities(renderer *core.CliRenderer, overrides TerminalCapabilitiesOverrides) *core.TerminalCapabilities {
	caps := CreateTerminalCapabilities(overrides)
	renderer.Context().Capabilities = caps
	return caps
}
