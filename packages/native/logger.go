package native

import (
	"fmt"
	"sync/atomic"
)

// LogLevel mirrors logger.zig LogLevel.
type LogLevel uint8

const (
	LogLevelErr   LogLevel = 0
	LogLevelWarn  LogLevel = 1
	LogLevelInfo  LogLevel = 2
	LogLevelDebug LogLevel = 3
)

// Diagnostic mirrors logger.zig Diagnostic.
// Message is truncated to MaxMessageBytes.
const MaxMessageBytes = 4096

type Diagnostic struct {
	Level      LogLevel
	Truncated  bool
	MessageLen uint32
	Message    [MaxMessageBytes]byte
}

// Diagnostics is a single-owner telemetry storage ring buffer.
// Full queues drop the newest message before formatting it.
// Mirrors logger.zig Diagnostics.
type Diagnostics struct {
	allocator *Allocator
	events    []Diagnostic
	head      uint32
	count     uint32
	dropped   uint64
}

// NewDiagnostics creates a Diagnostics ring buffer with the given capacity.
// Zero capacity disables storage but still counts dropped messages.
// Mirrors logger.zig Diagnostics.init.
func NewDiagnostics(allocator *Allocator, capacity uint32) (*Diagnostics, error) {
	if capacity == 0 {
		return &Diagnostics{}, nil
	}
	events := make([]Diagnostic, capacity)
	return &Diagnostics{
		allocator: allocator,
		events:    events,
	}, nil
}

// Pop removes and returns the oldest record. The record stays readable
// until the next log call. Returns nil if empty.
// Mirrors logger.zig Diagnostics.pop.
func (d *Diagnostics) Pop() *Diagnostic {
	if d.count == 0 {
		return nil
	}
	event := &d.events[d.head]
	d.head = (d.head + 1) % uint32(len(d.events))
	d.count--
	return event
}

func (d *Diagnostics) logMessage(level LogLevel, format string, args ...interface{}) {
	if d.count == uint32(len(d.events)) {
		d.dropped++
		return
	}
	index := (d.head + d.count) % uint32(len(d.events))
	event := &d.events[index]
	event.Level = level
	event.Truncated = false
	msg := fmt.Sprintf(format, args...)
	if len(msg) > MaxMessageBytes {
		msg = msg[:MaxMessageBytes]
		event.Truncated = true
	}
	event.MessageLen = uint32(len(msg))
	copy(event.Message[:], msg)
	d.count++
}

// Logger is a logging sink: callback, diagnostics queue, or process callback.
// Mirrors logger.zig Logger.
type Logger struct {
	kind        loggerKind
	callback    LogCallback
	diagnostics *Diagnostics
}

type loggerKind uint8

const (
	loggerKindCallback loggerKind = iota
	loggerKindDiagnostics
	loggerKindProcess
)

// LogCallback is a process-wide log callback.
// Mirrors logger.zig LogCallback.
type LogCallback func(level uint8, msg string)

// processLogger is the process-wide callback logger.
var processLogger Logger = Logger{kind: loggerKindProcess}

var (
	processCallback atomic.Value // holds *LogCallback
	processThread   atomic.Value // holds uint64 thread ID
)

func init() {
	processCallback.Store((*LogCallback)(nil))
	processThread.Store(uint64(0))
}

// ProcessLogger returns the process-wide logger.
// Mirrors logger.zig processLogger.
func ProcessLogger() *Logger {
	return &processLogger
}

// ProcessLogReady reports whether a process log line from this thread
// reaches a callback now.
// Mirrors logger.zig processLogReady.
func ProcessLogReady() bool {
	cb := processCallback.Load().(*LogCallback)
	if cb == nil {
		return false
	}
	return true
}

// SetLogCallback sets the process-wide log callback.
// Mirrors logger.zig setLogCallback.
func SetLogCallback(callback LogCallback) {
	var cb *LogCallback
	if callback != nil {
		cb = &callback
	}
	processCallback.Store(cb)
	processThread.Store(uint64(1))
}

// LogMessage sends a message through the process logger.
// Mirrors logger.zig logMessage.
func LogMessage(level LogLevel, format string, args ...interface{}) {
	ProcessLogger().LogMessage(level, format, args...)
}

// Err logs an error-level message through the process logger.
// Mirrors logger.zig err.
func Err(format string, args ...interface{}) {
	LogMessage(LogLevelErr, format, args...)
}

// Warn logs a warning-level message through the process logger.
// Mirrors logger.zig warn.
func Warn(format string, args ...interface{}) {
	LogMessage(LogLevelWarn, format, args...)
}

// Info logs an info-level message through the process logger.
// Mirrors logger.zig info.
func Info(format string, args ...interface{}) {
	LogMessage(LogLevelInfo, format, args...)
}

// Debug logs a debug-level message through the process logger.
// Mirrors logger.zig debug.
func Debug(format string, args ...interface{}) {
	LogMessage(LogLevelDebug, format, args...)
}

// LogMessage sends a formatted message through this logger.
// Mirrors logger.zig Logger.logMessage.
func (l *Logger) LogMessage(level LogLevel, format string, args ...interface{}) {
	switch l.kind {
	case loggerKindDiagnostics:
		if l.diagnostics != nil {
			l.diagnostics.logMessage(level, format, args...)
		}
	case loggerKindProcess:
		cb := processCallback.Load().(*LogCallback)
		if cb == nil {
			return
		}
		msg := fmt.Sprintf(format, args...)
		(*cb)(uint8(level), msg)
	case loggerKindCallback:
		if l.callback != nil {
			msg := fmt.Sprintf(format, args...)
			l.callback(uint8(level), msg)
		}
	}
}

// Err logs an error-level message through this logger.
// Mirrors logger.zig Logger.err.
func (l *Logger) Err(format string, args ...interface{}) {
	l.LogMessage(LogLevelErr, format, args...)
}

// Warn logs a warning-level message through this logger.
// Mirrors logger.zig Logger.warn.
func (l *Logger) Warn(format string, args ...interface{}) {
	l.LogMessage(LogLevelWarn, format, args...)
}

// Info logs an info-level message through this logger.
// Mirrors logger.zig Logger.info.
func (l *Logger) Info(format string, args ...interface{}) {
	l.LogMessage(LogLevelInfo, format, args...)
}

// Debug logs a debug-level message through this logger.
// Mirrors logger.zig Logger.debug.
func (l *Logger) Debug(format string, args ...interface{}) {
	l.LogMessage(LogLevelDebug, format, args...)
}

// NewCallbackLogger creates a Logger that invokes the given callback.
func NewCallbackLogger(cb LogCallback) *Logger {
	return &Logger{kind: loggerKindCallback, callback: cb}
}

// NewDiagnosticsLogger creates a Logger that writes to the given diagnostics queue.
func NewDiagnosticsLogger(d *Diagnostics) *Logger {
	return &Logger{kind: loggerKindDiagnostics, diagnostics: d}
}
