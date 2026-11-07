package native

import (
	"image"
	"io"
)

// pixtui is the core module aggregating the main public API of the native
// package. It mirrors the Zig pixtui.zig which aggregates and re-exports
// the primary types and constructors for native integration.
//
// In Go all types live in the same `native` package, so no type aliasing is
// needed. This file provides a compile-time verification that the public
// API surface expected by pixtui.zig is present.

// Compile-time check that all expected constructors and symbols exist.
// These reference the functions defined in their respective files.
var _ = struct {
	NewResourceContext       func() *ResourceContext
	NewTerminal              func() *Terminal
	NewVTTerminal            func(int, int) *VTTerminal
	NewVTStream              func(*VTTerminal) *VTStream
	NewRope                  func() *Rope
	NewRopeWithContent       func(string) *Rope
	NewTextBuffer            func() *TextBuffer
	NewTextBufferWithContent func(string) *TextBuffer
	NewEditBuffer            func() *EditBuffer
	NewEditBufferWithContent func(string) *EditBuffer
	NewEditorView            func(*EditBuffer) *EditorView
	NewEditorViewWithContent func(string) *EditorView
	NewOptimizedBuffer       func(int, int) *OptimizedBuffer
	NewSession               func(uint64, *ResourceContext) *Session
	NewScene                 func(uint64) *Scene
	NewClipboard             func() *Clipboard
	NewGraphemePool          func() *GraphemePool
	NewGraphemeTracker       func(*GraphemePool) *GraphemeTracker
	NewLinkPool              func() *LinkPool
	NewLinkTracker           func(*LinkPool) *LinkTracker
	NewEventEmitter          func() *EventEmitter
	NewRuntime               func() *Runtime
	NewAudioEngine           func(int) *AudioEngine
	NewAudioEngineWithFormat func(AudioFormat) *AudioEngine
	NewAudioVoice            func(uint64, int, float64, float64, Waveform) *AudioVoice
	NewImage                 func(int, int) *Image
	NewImageFromRGBA         func(*image.RGBA) *Image
	NewVideo                 func(int, int) *Video
	NewFFmpegWrapper         func() *FFmpegWrapper
	NewInputParser           func() *InputParser
	NewFileLogger            func(string) (*FileLogger, error)
	NewCliRenderer           func(*ResourceContext, *Session) *CliRenderer
	NewNativeRenderable      func(uint64, string, func(*OptimizedBuffer, float64)) *NativeRenderable
	NewSplitScrollback       func() *SplitScrollback
	NewStdoutWriter          func() *StdoutWriter
	NewStdoutOutput          func(io.Writer) *StdoutOutput
	NewStdoutOutputForFile   func() *StdoutOutput
	NewEmbeddedTerminal      func(int, int) *EmbeddedTerminal
	SetTextWorkEnabled       func(bool)
	SetIO                    func(io.Writer)
	DefaultStdOptions        func() StdOptions
}{
	NewResourceContext:       NewResourceContext,
	NewTerminal:              NewTerminal,
	NewVTTerminal:            NewVTTerminal,
	NewVTStream:              NewVTStream,
	NewRope:                  NewRope,
	NewRopeWithContent:       NewRopeWithContent,
	NewTextBuffer:            NewTextBuffer,
	NewTextBufferWithContent: NewTextBufferWithContent,
	NewEditBuffer:            NewEditBuffer,
	NewEditBufferWithContent: NewEditBufferWithContent,
	NewEditorView:            NewEditorView,
	NewEditorViewWithContent: NewEditorViewWithContent,
	NewOptimizedBuffer:       NewOptimizedBuffer,
	NewSession:               NewSession,
	NewScene:                 NewScene,
	NewClipboard:             NewClipboard,
	NewGraphemePool:          NewGraphemePool,
	NewGraphemeTracker:       NewGraphemeTracker,
	NewLinkPool:              NewLinkPool,
	NewLinkTracker:           NewLinkTracker,
	NewEventEmitter:          NewEventEmitter,
	NewRuntime:               NewRuntime,
	NewAudioEngine:           NewAudioEngine,
	NewAudioEngineWithFormat: NewAudioEngineWithFormat,
	NewAudioVoice:            NewAudioVoice,
	NewImage:                 NewImage,
	NewImageFromRGBA:         NewImageFromRGBA,
	NewVideo:                 NewVideo,
	NewFFmpegWrapper:         NewFFmpegWrapper,
	NewInputParser:           NewInputParser,
	NewFileLogger:            NewFileLogger,
	NewCliRenderer:           NewCliRenderer,
	NewNativeRenderable:      NewNativeRenderable,
	NewSplitScrollback:       NewSplitScrollback,
	NewStdoutWriter:          NewStdoutWriter,
	NewStdoutOutput:          NewStdoutOutput,
	NewStdoutOutputForFile:   NewStdoutOutputForFile,
	NewEmbeddedTerminal:      NewEmbeddedTerminal,
	SetTextWorkEnabled:       SetTextWorkEnabled,
	SetIO:                    SetIO,
	DefaultStdOptions:        DefaultStdOptions,
}
