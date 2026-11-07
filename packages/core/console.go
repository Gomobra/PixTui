package core

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/Gomobra/PixTui/packages/native"
)

// LogLevel represents the severity of a console log entry.
type LogLevel string

const (
	LogLevelLog    LogLevel = "LOG"
	LogLevelInfo   LogLevel = "INFO"
	LogLevelWarn   LogLevel = "WARN"
	LogLevelError  LogLevel = "ERROR"
	LogLevelDebug  LogLevel = "DEBUG"
)

// ConsolePosition represents the position of the console on screen.
type ConsolePosition string

const (
	ConsolePositionTop    ConsolePosition = "top"
	ConsolePositionBottom ConsolePosition = "bottom"
	ConsolePositionLeft   ConsolePosition = "left"
	ConsolePositionRight  ConsolePosition = "right"
)

// ConsoleAction represents an action that can be triggered from the console.
type ConsoleAction string

const (
	ConsoleActionScrollUp        ConsoleAction = "scroll-up"
	ConsoleActionScrollDown      ConsoleAction = "scroll-down"
	ConsoleActionScrollToTop     ConsoleAction = "scroll-to-top"
	ConsoleActionScrollToBottom  ConsoleAction = "scroll-to-bottom"
	ConsoleActionPositionPrevious ConsoleAction = "position-previous"
	ConsoleActionPositionNext    ConsoleAction = "position-next"
	ConsoleActionSizeIncrease    ConsoleAction = "size-increase"
	ConsoleActionSizeDecrease    ConsoleAction = "size-decrease"
	ConsoleActionSaveLogs        ConsoleAction = "save-logs"
	ConsoleActionCopySelection   ConsoleAction = "copy-selection"
)

// ConsoleKeyBinding binds a key to a console action.
type ConsoleKeyBinding struct {
	Name   string
	Ctrl   bool
	Shift  bool
	Action ConsoleAction
}

// KeyAliasMap maps key alias names to their values.
type KeyAliasMap map[string]string

// CallerInfo contains information about the calling function.
type CallerInfo struct {
	FunctionName string
	FullPath     string
	FileName     string
	LineNumber   int
	ColumnNumber int
}

// LogEntry represents a single console log entry.
type LogEntry struct {
	Date      time.Time
	Level     LogLevel
	Args      []interface{}
	CallerInfo *CallerInfo
}

// ConsoleOptions contains configuration for the Console.
type ConsoleOptions struct {
	Position         ConsolePosition
	SizePercent      int
	ZIndex           int
	ColorInfo        *RGBA
	ColorWarn        *RGBA
	ColorError       *RGBA
	ColorDebug       *RGBA
	ColorDefault     *RGBA
	BackgroundColor  *RGBA
	StartInDebugMode bool
	Title            string
	TitleBarColor    *RGBA
	TitleBarTextColor *RGBA
	CursorColor      *RGBA
	MaxStoredLogs    int
	MaxDisplayLines  int
	OnCopySelection  func(text string)
	KeyBindings      []ConsoleKeyBinding
	KeyAliasMap      KeyAliasMap
	SelectionColor   *RGBA
	CopyButtonColor  *RGBA
	Clock            Clock
}

// defaultConsoleOptions returns the default console options.
func defaultConsoleOptions() ConsoleOptions {
	return ConsoleOptions{
		Position:         ConsolePositionBottom,
		SizePercent:      30,
		ZIndex:           1000,
		ColorInfo:        &RGBA{0, 255, 255, 255},
		ColorWarn:        &RGBA{255, 255, 0, 255},
		ColorError:       &RGBA{255, 0, 0, 255},
		ColorDebug:       &RGBA{128, 128, 128, 255},
		ColorDefault:     &RGBA{255, 255, 255, 255},
		BackgroundColor:  &RGBA{26, 26, 26, 179},
		StartInDebugMode: false,
		Title:            "Console",
		TitleBarColor:    &RGBA{13, 13, 13, 179},
		TitleBarTextColor: &RGBA{255, 255, 255, 255},
		CursorColor:      &RGBA{0, 160, 255, 255},
		MaxStoredLogs:    2000,
		MaxDisplayLines:  3000,
		SelectionColor:   &RGBA{77, 128, 204, 128},
		CopyButtonColor:  &RGBA{0, 160, 255, 255},
	}
}

// displayLine represents a single line of text to display in the console.
type displayLine struct {
	text   string
	level  LogLevel
	indent bool
}

// consoleSelection represents a text selection range.
type consoleSelection struct {
	StartLine int
	StartCol  int
	EndLine   int
	EndCol    int
}

// TerminalConsoleCache caches log entries and manages console capture.
type TerminalConsoleCache struct {
	mu              sync.RWMutex
	cachedLogs      []LogEntry
	maxCacheSize    int
	collectCallerInfo bool
	cachingEnabled  bool
	emitter         *native.EventEmitter
}

// NewTerminalConsoleCache creates a new TerminalConsoleCache.
func NewTerminalConsoleCache() *TerminalConsoleCache {
	return &TerminalConsoleCache{
		cachedLogs:     make([]LogEntry, 0),
		maxCacheSize:   1000,
		cachingEnabled: true,
		emitter:        native.NewEventEmitter(),
	}
}

// CachedLogs returns a copy of the cached logs.
func (c *TerminalConsoleCache) CachedLogs() []LogEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]LogEntry, len(c.cachedLogs))
	copy(result, c.cachedLogs)
	return result
}

// SetCollectCallerInfo enables or disables caller info collection.
func (c *TerminalConsoleCache) SetCollectCallerInfo(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.collectCallerInfo = enabled
}

// SetCachingEnabled enables or disables caching.
func (c *TerminalConsoleCache) SetCachingEnabled(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cachingEnabled = enabled
}

// ClearConsole clears all cached logs.
func (c *TerminalConsoleCache) ClearConsole() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cachedLogs = make([]LogEntry, 0)
}

// AddLogEntry adds a log entry to the cache and emits an "entry" event.
func (c *TerminalConsoleCache) AddLogEntry(level LogLevel, args []interface{}) LogEntry {
	c.mu.Lock()
	defer c.mu.Unlock()

	var callerInfo *CallerInfo
	if c.collectCallerInfo {
		callerInfo = getCallerInfo()
	}

	entry := LogEntry{
		Date:       time.Now(),
		Level:      level,
		Args:       args,
		CallerInfo: callerInfo,
	}

	if c.cachingEnabled {
		if len(c.cachedLogs) >= c.maxCacheSize {
			c.cachedLogs = c.cachedLogs[1:]
		}
		c.cachedLogs = append(c.cachedLogs, entry)
	}

	c.emitter.Emit("entry", entry)
	return entry
}

// On registers an event listener.
func (c *TerminalConsoleCache) On(event string, listener func(interface{})) {
	c.emitter.On(event, listener)
}

// Off removes all listeners for an event.
func (c *TerminalConsoleCache) Off(event string) {
	c.emitter.Off(event)
}

// globalConsoleCache is the singleton terminal console cache.
var globalConsoleCache = NewTerminalConsoleCache()

// getCallerInfo attempts to extract caller information from the call stack.
func getCallerInfo() *CallerInfo {
	pc := make([]uintptr, 10)
	n := runtime.Callers(2, pc)
	if n == 0 {
		return nil
	}
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	file := frame.File
	function := frame.Function
	line := frame.Line

	var fileName string
	parts := strings.Split(file, "/")
	if len(parts) > 0 {
		fileName = parts[len(parts)-1]
	} else {
		fileName = file
	}

	return &CallerInfo{
		FunctionName: function,
		FullPath:     file,
		FileName:     fileName,
		LineNumber:   line,
		ColumnNumber: 0,
	}
}

// formatArguments formats log arguments into a string.
func formatArguments(args []interface{}) string {
	parts := make([]string, len(args))
	for i, arg := range args {
		if err, ok := arg.(error); ok {
			parts[i] = fmt.Sprintf("Error: %s", err.Error())
			continue
		}
		parts[i] = fmt.Sprintf("%v", arg)
	}
	result := strings.Join(parts, " ")
	result = strings.ReplaceAll(result, "\r\n", "\n")
	var sb strings.Builder
	for _, r := range result {
		if r >= 32 || r == '\n' || r == '\t' {
			sb.WriteRune(r)
		} else {
			sb.WriteString(fmt.Sprintf("\\x%02x", r))
		}
	}
	return sb.String()
}

// formatTimestamp formats a time as HH:MM:SS.
func formatTimestamp(t time.Time) string {
	return t.Format("15:04:05")
}

// Console is a debug console overlay that displays log entries.
// Mirrors packages/core/src/console.ts TerminalConsole.
type Console struct {
	mu              sync.RWMutex
	renderer        *CliRenderer
	options         ConsoleOptions

	isVisible       bool
	isFocused       bool
	debugModeEnabled bool

	frameBuffer      *NativeBuffer
	consoleX         int
	consoleY         int
	consoleWidth     int
	consoleHeight    int
	scrollTopIndex   int
	isScrolledToBottom bool
	currentLineIndex int
	displayLines     []displayLine
	allLogEntries    []LogEntry
	needsUpdate      bool

	selectionStart  *selectionPos
	selectionEnd    *selectionPos
	isDragging      bool
	copyButtonBounds copyButtonBounds

	keyBindingsMap     map[string]ConsoleAction
	keyAliasMap        KeyAliasMap
	keyBindings        []ConsoleKeyBinding
	mergedKeyBindings  []ConsoleKeyBinding
	actionHandlers     map[ConsoleAction]func() bool

	entryListener      func(entry LogEntry)
}

// selectionPos represents a selection position.
type selectionPos struct {
	Line int
	Col  int
}

// copyButtonBounds represents the copy button bounds.
type copyButtonBounds struct {
	X      int
	Y      int
	Width  int
	Height int
}

// NewConsole creates a new Console attached to the given renderer.
func NewConsole(renderer *CliRenderer, options ConsoleOptions) *Console {
	opts := defaultConsoleOptions()
	if options.Position != "" {
		opts.Position = options.Position
	}
	if options.SizePercent > 0 {
		opts.SizePercent = options.SizePercent
	}
	if options.ZIndex != 0 {
		opts.ZIndex = options.ZIndex
	}
	if options.ColorInfo != nil {
		opts.ColorInfo = options.ColorInfo
	}
	if options.ColorWarn != nil {
		opts.ColorWarn = options.ColorWarn
	}
	if options.ColorError != nil {
		opts.ColorError = options.ColorError
	}
	if options.ColorDebug != nil {
		opts.ColorDebug = options.ColorDebug
	}
	if options.ColorDefault != nil {
		opts.ColorDefault = options.ColorDefault
	}
	if options.BackgroundColor != nil {
		opts.BackgroundColor = options.BackgroundColor
	}
	opts.StartInDebugMode = options.StartInDebugMode
	if options.Title != "" {
		opts.Title = options.Title
	}
	if options.TitleBarColor != nil {
		opts.TitleBarColor = options.TitleBarColor
	}
	if options.TitleBarTextColor != nil {
		opts.TitleBarTextColor = options.TitleBarTextColor
	}
	if options.CursorColor != nil {
		opts.CursorColor = options.CursorColor
	}
	if options.MaxStoredLogs > 0 {
		opts.MaxStoredLogs = options.MaxStoredLogs
	}
	if options.MaxDisplayLines > 0 {
		opts.MaxDisplayLines = options.MaxDisplayLines
	}
	if options.SelectionColor != nil {
		opts.SelectionColor = options.SelectionColor
	}
	if options.CopyButtonColor != nil {
		opts.CopyButtonColor = options.CopyButtonColor
	}
	if options.Clock != nil {
		opts.Clock = options.Clock
	} else {
		opts.Clock = SystemClock{}
	}
	opts.KeyBindings = options.KeyBindings
	opts.KeyAliasMap = options.KeyAliasMap
	if options.OnCopySelection != nil {
		opts.OnCopySelection = options.OnCopySelection
	}

	c := &Console{
		renderer:         renderer,
		options:          opts,
		debugModeEnabled: opts.StartInDebugMode,
		displayLines:     make([]displayLine, 0),
		allLogEntries:    make([]LogEntry, 0),
		keyBindingsMap:   make(map[string]ConsoleAction),
		keyAliasMap:      make(KeyAliasMap),
		keyBindings:      make([]ConsoleKeyBinding, 0),
		mergedKeyBindings: make([]ConsoleKeyBinding, 0),
		actionHandlers:   make(map[ConsoleAction]func() bool),
		selectionStart:   nil,
		selectionEnd:     nil,
		copyButtonBounds: copyButtonBounds{X: -1, Y: -1},
	}

	c.updateConsoleDimensions()
	c.scrollToBottom()

	c.entryListener = func(entry LogEntry) {
		c.handleNewLog(entry)
	}
	globalConsoleCache.On("entry", func(entry interface{}) {
		le, ok := entry.(LogEntry)
		if ok {
			c.entryListener(le)
		}
	})

	if opts.StartInDebugMode {
		globalConsoleCache.SetCollectCallerInfo(true)
	}

	c.buildActionHandlers()

	return c
}

// updateConsoleDimensions updates the console dimensions based on renderer size.
func (c *Console) updateConsoleDimensions() {
	width, height := c.renderer.Dimensions()
	sizePercent := float64(c.options.SizePercent) / 100.0

	switch c.options.Position {
	case ConsolePositionTop:
		c.consoleX = 0
		c.consoleY = 0
		c.consoleWidth = width
		c.consoleHeight = maxInt(1, int(float64(height)*sizePercent))
	case ConsolePositionBottom:
		c.consoleHeight = maxInt(1, int(float64(height)*sizePercent))
		c.consoleWidth = width
		c.consoleX = 0
		c.consoleY = height - c.consoleHeight
	case ConsolePositionLeft:
		c.consoleWidth = maxInt(1, int(float64(width)*sizePercent))
		c.consoleHeight = height
		c.consoleX = 0
		c.consoleY = 0
	case ConsolePositionRight:
		c.consoleWidth = maxInt(1, int(float64(width)*sizePercent))
		c.consoleHeight = height
		c.consoleY = 0
		c.consoleX = width - c.consoleWidth
	}

	c.currentLineIndex = maxInt(0, minInt(c.currentLineIndex, c.consoleHeight-1))
}

// handleNewLog handles a new log entry while the console is visible.
func (c *Console) handleNewLog(entry LogEntry) {
	if !c.isVisible {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.allLogEntries = append(c.allLogEntries, entry)

	if len(c.allLogEntries) > c.options.MaxStoredLogs {
		c.allLogEntries = c.allLogEntries[len(c.allLogEntries)-c.options.MaxStoredLogs:]
	}

	newLines := c.processLogEntry(entry)
	c.displayLines = append(c.displayLines, newLines...)

	if len(c.displayLines) > c.options.MaxDisplayLines {
		removed := len(c.displayLines) - c.options.MaxDisplayLines
		c.displayLines = c.displayLines[removed:]
		c.scrollTopIndex = maxInt(0, c.scrollTopIndex-removed)
	}

	if c.isScrolledToBottom {
		c.scrollToBottom()
	}
	c.needsUpdate = true
	c.renderer.RequestRender()
}

// processLogEntry processes a log entry into display lines.
func (c *Console) processLogEntry(entry LogEntry) []displayLine {
	timestamp := formatTimestamp(entry.Date)
	callerSource := "unknown"
	if entry.CallerInfo != nil {
		callerSource = fmt.Sprintf("%s:%d", entry.CallerInfo.FileName, entry.CallerInfo.LineNumber)
	}
	prefix := fmt.Sprintf("[%s] [%s]", timestamp, entry.Level)
	if c.debugModeEnabled {
		prefix += fmt.Sprintf(" [%s]", callerSource)
	}
	prefix += " "

	formattedArgs := formatArguments(entry.Args)
	initialLines := strings.Split(formattedArgs, "\n")

	cells := maxInt(1, c.consoleWidth-1)
	indentedCells := maxInt(1, c.consoleWidth-1-indentWidth)

	var result []displayLine
	for i, line := range initialLines {
		isFirstLine := i == 0
		var segments []string
		if isFirstLine {
			segments = wrapDisplayCells(prefix+line, cells, indentedCells)
		} else {
			segments = wrapDisplayCells(line, indentedCells, indentedCells)
		}
		for segIdx, seg := range segments {
			result = append(result, displayLine{
				text:   seg,
				level:  entry.Level,
				indent: !isFirstLine || segIdx > 0,
			})
		}
	}

	return result
}

// Activate activates console capture.
func (c *Console) Activate() {
	globalConsoleCache.SetCollectCallerInfo(c.debugModeEnabled)
}

// Deactivate deactivates console capture.
func (c *Console) Deactivate() {
	globalConsoleCache.SetCachingEnabled(true)
}

// handleKeyPress handles a key press event.
func (c *Console) handleKeyPress(key string, ctrl, shift bool) {
	if key == "escape" {
		c.blur()
		return
	}

	action := c.getKeyBindingAction(key, ctrl, shift)
	if action != "" {
		if handler, ok := c.actionHandlers[action]; ok {
			handler()
		}
	}
}

// getKeyBindingAction returns the action bound to the given key.
func (c *Console) getKeyBindingAction(key string, ctrl, shift bool) ConsoleAction {
	for _, binding := range c.mergedKeyBindings {
		if binding.Name == key && binding.Ctrl == ctrl && binding.Shift == shift {
			return binding.Action
		}
	}
	return ""
}

// scrollUp scrolls the console up.
func (c *Console) scrollUp() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.currentLineIndex > 0 {
		c.currentLineIndex--
		c.needsUpdate = true
		c.renderer.RequestRender()
	} else if c.scrollTopIndex > 0 {
		c.scrollTopIndex--
		c.isScrolledToBottom = false
		c.needsUpdate = true
		c.renderer.RequestRender()
	}
	return true
}

// scrollDown scrolls the console down.
func (c *Console) scrollDown() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	displayLineCount := len(c.displayLines)
	logAreaHeight := maxInt(1, c.consoleHeight-1)
	maxScrollTop := maxInt(0, displayLineCount-logAreaHeight)
	canCursorMoveDown := c.currentLineIndex < logAreaHeight-1 && c.scrollTopIndex+c.currentLineIndex < displayLineCount-1

	if canCursorMoveDown {
		c.currentLineIndex++
		c.needsUpdate = true
		c.renderer.RequestRender()
	} else if c.scrollTopIndex < maxScrollTop {
		c.scrollTopIndex++
		c.isScrolledToBottom = c.scrollTopIndex == maxScrollTop
		c.needsUpdate = true
		c.renderer.RequestRender()
	}
	return true
}

// scrollToTop scrolls to the top.
func (c *Console) scrollToTop() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.scrollTopIndex > 0 || c.currentLineIndex > 0 {
		c.scrollTopIndex = 0
		c.currentLineIndex = 0
		c.isScrolledToBottom = len(c.displayLines) <= maxInt(1, c.consoleHeight-1)
		c.needsUpdate = true
		c.renderer.RequestRender()
	}
	return true
}

// scrollToBottomAction scrolls to the bottom.
func (c *Console) scrollToBottomAction() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	logAreaHeight := maxInt(1, c.consoleHeight-1)
	maxScrollPossible := maxInt(0, len(c.displayLines)-logAreaHeight)
	if c.scrollTopIndex < maxScrollPossible || !c.isScrolledToBottom {
		c.scrollToBottom()
		c.needsUpdate = true
		c.renderer.RequestRender()
	}
	return true
}

// scrollToBottom scrolls to the bottom (must hold lock).
func (c *Console) scrollToBottom() {
	displayLineCount := len(c.displayLines)
	logAreaHeight := maxInt(1, c.consoleHeight-1)
	maxScrollTop := maxInt(0, displayLineCount-logAreaHeight)
	c.scrollTopIndex = maxScrollTop
	c.isScrolledToBottom = true

	visibleLineCount := minInt(logAreaHeight, displayLineCount-c.scrollTopIndex)
	if c.currentLineIndex >= visibleLineCount {
		c.currentLineIndex = maxInt(0, visibleLineCount-1)
	}
}

// positionPrevious moves to the previous console position.
func (c *Console) positionPrevious() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	positions := []ConsolePosition{ConsolePositionTop, ConsolePositionRight, ConsolePositionBottom, ConsolePositionLeft}
	currentIndex := -1
	for i, p := range positions {
		if p == c.options.Position {
			currentIndex = i
			break
		}
	}
	if currentIndex < 0 {
		currentIndex = 0
	}
	prevIndex := (currentIndex - 1 + len(positions)) % len(positions)
	c.options.Position = positions[prevIndex]
	width, height := c.renderer.Dimensions()
	c.resize(width, height)
	return true
}

// positionNext moves to the next console position.
func (c *Console) positionNext() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	positions := []ConsolePosition{ConsolePositionTop, ConsolePositionRight, ConsolePositionBottom, ConsolePositionLeft}
	currentIndex := -1
	for i, p := range positions {
		if p == c.options.Position {
			currentIndex = i
			break
		}
	}
	if currentIndex < 0 {
		currentIndex = 0
	}
	nextIndex := (currentIndex + 1) % len(positions)
	c.options.Position = positions[nextIndex]
	width, height := c.renderer.Dimensions()
	c.resize(width, height)
	return true
}

// sizeIncrease increases the console size.
func (c *Console) sizeIncrease() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.options.SizePercent = minInt(100, c.options.SizePercent+5)
	width, height := c.renderer.Dimensions()
	c.resize(width, height)
	return true
}

// sizeDecrease decreases the console size.
func (c *Console) sizeDecrease() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.options.SizePercent = maxInt(10, c.options.SizePercent-5)
	width, height := c.renderer.Dimensions()
	c.resize(width, height)
	return true
}

// saveLogsAction saves logs to a file.
func (c *Console) saveLogsAction() bool {
	c.saveLogsToFile()
	return true
}

// triggerCopyAction triggers the copy action.
func (c *Console) triggerCopyAction() bool {
	c.triggerCopy()
	return true
}

// buildActionHandlers builds the action handler map.
func (c *Console) buildActionHandlers() {
	c.actionHandlers = map[ConsoleAction]func() bool{
		ConsoleActionScrollUp:         c.scrollUp,
		ConsoleActionScrollDown:       c.scrollDown,
		ConsoleActionScrollToTop:      c.scrollToTop,
		ConsoleActionScrollToBottom:   c.scrollToBottomAction,
		ConsoleActionPositionPrevious: c.positionPrevious,
		ConsoleActionPositionNext:     c.positionNext,
		ConsoleActionSizeIncrease:     c.sizeIncrease,
		ConsoleActionSizeDecrease:     c.sizeDecrease,
		ConsoleActionSaveLogs:         c.saveLogsAction,
		ConsoleActionCopySelection:    c.triggerCopyAction,
	}
}

// attachStdin attaches the key handler to the renderer.
func (c *Console) attachStdin() {
	if c.isFocused {
		return
	}
	c.renderer.SetKeyHandler(func(data []byte) {
		c.handleKeyInput(data)
	})
	c.isFocused = true
}

// handleKeyInput handles raw key input bytes from the renderer.
func (c *Console) handleKeyInput(data []byte) {
	if len(data) == 0 {
		return
	}
	if data[0] == 27 {
		c.blur()
		return
	}
	if data[0] < 32 {
		key := string(rune(data[0] + 96))
		if key >= "a" && key <= "z" {
			c.handleKeyPress(key, true, false)
		}
		return
	}
	key := string(data[0])
	c.handleKeyPress(key, false, false)
}

// detachStdin detaches the key handler from the renderer.
func (c *Console) detachStdin() {
	if !c.isFocused {
		return
	}
	c.renderer.SetKeyHandler(nil)
	c.isFocused = false
}

// resize resizes the console.
func (c *Console) resize(width, height int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.updateConsoleDimensions()

	if c.frameBuffer != nil {
		c.frameBuffer.inner.Resize(c.consoleWidth, c.consoleHeight)

		displayLineCount := len(c.displayLines)
		logAreaHeight := maxInt(1, c.consoleHeight-1)
		maxScrollTop := maxInt(0, displayLineCount-logAreaHeight)
		c.scrollTopIndex = minInt(c.scrollTopIndex, maxScrollTop)
		c.isScrolledToBottom = c.scrollTopIndex == maxScrollTop
		visibleLineCount := minInt(logAreaHeight, displayLineCount-c.scrollTopIndex)
		c.currentLineIndex = maxInt(0, minInt(c.currentLineIndex, visibleLineCount-1))

		if c.isVisible {
			c.needsUpdate = true
			c.renderer.RequestRender()
		}
	}
}

// clear clears all console logs.
func (c *Console) clear() {
	globalConsoleCache.ClearConsole()
	c.mu.Lock()
	defer c.mu.Unlock()
	c.allLogEntries = make([]LogEntry, 0)
	c.displayLines = make([]displayLine, 0)
	c.needsUpdate = true
	c.renderer.RequestRender()
}

// toggle toggles the console visibility.
func (c *Console) toggle() {
	if c.isVisible {
		if c.isFocused {
			c.hide()
		} else {
			c.focus()
		}
	} else {
		c.show()
	}
	c.renderer.RequestRender()
}

// focus focuses the console (attaches input).
func (c *Console) focus() {
	c.attachStdin()
	c.scrollToBottom()
	c.needsUpdate = true
	c.renderer.RequestRender()
}

// blur blurs the console (detaches input).
func (c *Console) blur() {
	c.detachStdin()
	c.needsUpdate = true
	c.renderer.RequestRender()
}

// show shows the console.
func (c *Console) show() {
	c.mu.Lock()
	if c.renderer.IsDestroyed() {
		c.mu.Unlock()
		return
	}
	if c.isVisible {
		c.mu.Unlock()
		return
	}
	if c.frameBuffer == nil {
		c.frameBuffer = NewNativeBuffer(c.consoleWidth, c.consoleHeight)
	}
	c.isVisible = true
	c.processCachedLogs()
	globalConsoleCache.SetCachingEnabled(false)
	c.currentLineIndex = maxInt(0, minInt(len(c.displayLines)-1, c.consoleHeight-1))
	c.scrollTopIndex = 0
	c.scrollToBottom()

	c.focus()
	c.needsUpdate = true
	c.renderer.RequestRender()
	c.mu.Unlock()
}

// hide hides the console.
func (c *Console) hide() {
	c.mu.Lock()
	if c.isVisible {
		c.stopAutoScroll()
		c.isVisible = false
		c.blur()
		globalConsoleCache.SetCachingEnabled(true)
	}
	c.mu.Unlock()
}

// destroy destroys the console and releases resources.
func (c *Console) destroy() {
	c.stopAutoScroll()
	c.hide()
	c.Deactivate()
	globalConsoleCache.Off("entry")
	c.frameBuffer = nil
}

// stopAutoScroll stops any auto-scrolling.
func (c *Console) stopAutoScroll() {
	// No auto-scroll interval in Go port; kept for API compatibility.
}

// Log logs a message at LOG level.
func (c *Console) Log(args ...interface{}) {
	globalConsoleCache.AddLogEntry(LogLevelLog, args)
}

// Info logs a message at INFO level.
func (c *Console) Info(args ...interface{}) {
	globalConsoleCache.AddLogEntry(LogLevelInfo, args)
}

// Warn logs a message at WARN level.
func (c *Console) Warn(args ...interface{}) {
	globalConsoleCache.AddLogEntry(LogLevelWarn, args)
}

// Error logs a message at ERROR level.
func (c *Console) Error(args ...interface{}) {
	globalConsoleCache.AddLogEntry(LogLevelError, args)
}

// Debug logs a message at DEBUG level.
func (c *Console) Debug(args ...interface{}) {
	globalConsoleCache.AddLogEntry(LogLevelDebug, args)
}

// processCachedLogs processes all cached logs into the console.
func (c *Console) processCachedLogs() {
	logs := globalConsoleCache.CachedLogs()
	globalConsoleCache.ClearConsole()

	c.allLogEntries = append(c.allLogEntries, logs...)
	if len(c.allLogEntries) > c.options.MaxStoredLogs {
		c.allLogEntries = c.allLogEntries[len(c.allLogEntries)-c.options.MaxStoredLogs:]
	}

	for _, entry := range logs {
		processed := c.processLogEntry(entry)
		c.displayLines = append(c.displayLines, processed...)
	}

	if len(c.displayLines) > c.options.MaxDisplayLines {
		c.displayLines = c.displayLines[len(c.displayLines)-c.options.MaxDisplayLines:]
	}
}

// renderToBuffer renders the console to the given buffer.
func (c *Console) renderToBuffer(buffer *NativeBuffer) {
	c.mu.RLock()
	if !c.isVisible || c.frameBuffer == nil {
		c.mu.RUnlock()
		return
	}
	needsUpdate := c.needsUpdate
	c.mu.RUnlock()

	if needsUpdate {
		c.updateFrameBuffer()
		c.mu.Lock()
		c.needsUpdate = false
		c.mu.Unlock()
	}

	c.copyFrameBufferTo(buffer, c.consoleX, c.consoleY)
}

// copyFrameBufferTo copies the frame buffer into the target buffer at (dx, dy).
func (c *Console) copyFrameBufferTo(dst *NativeBuffer, dx, dy int) {
	c.mu.RLock()
	srcAccess := c.frameBuffer.Access()
	c.mu.RUnlock()

	dw := dst.Width()
	dh := dst.Height()
	sw := srcAccess.Width
	sh := srcAccess.Height

	for y := 0; y < sh; y++ {
		dstY := dy + y
		if dstY < 0 || dstY >= dh {
			continue
		}
		rowBase := y * sw
		for x := 0; x < sw; x++ {
			dstX := dx + x
			if dstX < 0 || dstX >= dw {
				continue
			}
			idx := rowBase + x
			ch := srcAccess.Char[idx]
			fg := srcAccess.Fg[idx]
			bg := srcAccess.Bg[idx]
			attrs := srcAccess.Attributes[idx]
			dst.SetCell(dstX, dstY, ch, fg, bg, attrs)
		}
	}
}

// updateFrameBuffer updates the internal frame buffer.
func (c *Console) updateFrameBuffer() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.frameBuffer == nil {
		return
	}

	c.frameBuffer.Clear()

	bg := *c.options.BackgroundColor
	bg16 := colorToUint16(bg)
	c.frameBuffer.FillRect(0, 0, c.consoleWidth, c.consoleHeight, ' ', bg16, bg16, 0)

	// Draw title bar
	title := c.options.Title
	if c.isFocused {
		title += " (Focused)"
	}
	titleX := maxInt(0, (c.consoleWidth-len(title))/2)
	c.drawTextToBuffer(c.frameBuffer, title, titleX, 0, *c.options.TitleBarTextColor, *c.options.TitleBarColor)

	// Draw copy button
	copyLabel := c.getCopyButtonLabel()
	copyButtonX := c.consoleWidth - len(copyLabel) - 1
	if copyButtonX >= 0 {
		copyEnabled := c.hasSelection()
		copyColor := *c.options.CopyButtonColor
		if !copyEnabled {
			copyColor = RGBA{100, 100, 100, 255}
		}
		c.drawTextToBuffer(c.frameBuffer, copyLabel, copyButtonX, 0, copyColor, *c.options.TitleBarColor)
		c.copyButtonBounds = copyButtonBounds{X: copyButtonX, Y: 0, Width: len(copyLabel), Height: 1}
	} else {
		c.copyButtonBounds = copyButtonBounds{X: -1, Y: -1, Width: 0, Height: 0}
	}

	logAreaHeight := maxInt(1, c.consoleHeight-1)
	startIndex := c.scrollTopIndex
	endIndex := minInt(startIndex+logAreaHeight, len(c.displayLines))
	visibleLines := c.displayLines[startIndex:endIndex]

	lineY := 1
	for i, dl := range visibleLines {
		if lineY >= c.consoleHeight {
			break
		}
		absoluteLineIndex := startIndex + i

		levelColor := *c.options.ColorDefault
		switch dl.level {
		case LogLevelInfo:
			levelColor = *c.options.ColorInfo
		case LogLevelWarn:
			levelColor = *c.options.ColorWarn
		case LogLevelError:
			levelColor = *c.options.ColorError
		case LogLevelDebug:
			levelColor = *c.options.ColorDebug
		}

		showCursor := c.isFocused && lineY-1 == c.currentLineIndex
		if showCursor {
			c.drawTextToBuffer(c.frameBuffer, ">", 0, lineY, *c.options.CursorColor, bg)
		} else {
			c.drawTextToBuffer(c.frameBuffer, " ", 0, lineY, *c.options.ColorDefault, bg)
		}

		fullText := dl.text
		if dl.indent {
			fullText = strings.Repeat(" ", indentWidth) + fullText
		}

		selRange := c.getLineSelectionRange(absoluteLineIndex)

		if selRange != nil {
			adjustedStart := maxInt(0, selRange.Start)
			adjustedEnd := minInt(len(fullText), selRange.End)

			if adjustedStart > 0 {
				c.drawTextToBuffer(c.frameBuffer, fullText[:adjustedStart], 1, lineY, levelColor, bg)
			}
			if adjustedStart < adjustedEnd {
				selText := fullText[adjustedStart:adjustedEnd]
				selBg := *c.options.SelectionColor
				selBg16 := colorToUint16(selBg)
				c.frameBuffer.FillRect(1+adjustedStart, lineY, adjustedEnd-adjustedStart, 1, ' ', selBg16, selBg16, 0)
				c.drawTextToBuffer(c.frameBuffer, selText, 1+adjustedStart, lineY, levelColor, selBg)
			}
			if adjustedEnd < len(fullText) {
				c.drawTextToBuffer(c.frameBuffer, fullText[adjustedEnd:], 1+adjustedEnd, lineY, levelColor, bg)
			}
		} else {
			c.drawTextToBuffer(c.frameBuffer, fullText, 1, lineY, levelColor, bg)
		}

		lineY++
	}
}

// drawTextToBuffer draws a string to the buffer at (x, y).
func (c *Console) drawTextToBuffer(buf *NativeBuffer, text string, x, y int, fg, bg RGBA) {
	if y < 0 || y >= c.consoleHeight {
		return
	}
	fg16 := colorToUint16(fg)
	bg16 := colorToUint16(bg)
	for i, ch := range text {
		cellX := x + i
		if cellX >= c.consoleWidth {
			break
		}
		buf.SetCell(cellX, y, uint32(ch), fg16, bg16, 0)
	}
}

// getCopyButtonLabel returns the label for the copy button.
func (c *Console) getCopyButtonLabel() string {
	var lastBinding *ConsoleKeyBinding
	for i := range c.mergedKeyBindings {
		if c.mergedKeyBindings[i].Action == ConsoleActionCopySelection {
			lastBinding = &c.mergedKeyBindings[i]
		}
	}
	if lastBinding != nil {
		return fmt.Sprintf("[Copy (%s)]", lastBinding.Name)
	}
	return "[Copy]"
}

// setDebugMode sets whether debug mode is enabled.
func (c *Console) setDebugMode(enabled bool) {
	c.mu.Lock()
	c.debugModeEnabled = enabled
	c.mu.Unlock()
	globalConsoleCache.SetCollectCallerInfo(enabled)
	if c.isVisible {
		c.needsUpdate = true
		c.renderer.RequestRender()
	}
}

// toggleDebugMode toggles debug mode.
func (c *Console) toggleDebugMode() {
	c.setDebugMode(!c.debugModeEnabled)
}

// hasSelection returns whether there is an active selection.
func (c *Console) hasSelection() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.selectionStart == nil || c.selectionEnd == nil {
		return false
	}
	return c.selectionStart.Line != c.selectionEnd.Line || c.selectionStart.Col != c.selectionEnd.Col
}

// normalizeSelection normalizes the selection so start <= end.
func (c *Console) normalizeSelection() *consoleSelection {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.selectionStart == nil || c.selectionEnd == nil {
		return nil
	}

	start := c.selectionStart
	end := c.selectionEnd

	startBeforeEnd := start.Line < end.Line || (start.Line == end.Line && start.Col <= end.Col)

	if startBeforeEnd {
		return &consoleSelection{
			StartLine: start.Line,
			StartCol:  start.Col,
			EndLine:   end.Line,
			EndCol:    end.Col,
		}
	}
	return &consoleSelection{
		StartLine: end.Line,
		StartCol:  end.Col,
		EndLine:   start.Line,
		EndCol:    start.Col,
	}
}

// getSelectedText returns the currently selected text.
func (c *Console) getSelectedText() string {
	sel := c.normalizeSelection()
	if sel == nil {
		return ""
	}

	var lines []string
	for i := sel.StartLine; i <= sel.EndLine; i++ {
		if i < 0 || i >= len(c.displayLines) {
			continue
		}
		dl := c.displayLines[i]
		fullText := dl.text
		if dl.indent {
			fullText = strings.Repeat(" ", indentWidth) + fullText
		}

		var text string
		if i == sel.StartLine && i == sel.EndLine {
			text = fullText[sel.StartCol:sel.EndCol]
		} else if i == sel.StartLine {
			text = fullText[sel.StartCol:]
		} else if i == sel.EndLine {
			text = fullText[:sel.EndCol]
		} else {
			text = fullText
		}
		lines = append(lines, text)
	}

	return strings.Join(lines, "\n")
}

// clearSelection clears the current selection.
func (c *Console) clearSelection() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.selectionStart = nil
	c.selectionEnd = nil
	c.isDragging = false
	c.stopAutoScroll()
	c.needsUpdate = true
	c.renderer.RequestRender()
}

// triggerCopy copies the selected text to the clipboard via callback.
func (c *Console) triggerCopy() {
	c.mu.Lock()
	if !c.hasSelection() {
		c.mu.Unlock()
		return
	}
	text := c.getSelectedText()
	callback := c.options.OnCopySelection
	c.mu.Unlock()

	if text != "" && callback != nil {
		callback(text)
	}

	c.clearSelection()
}

// getLineSelectionRange returns the selection range for a given line.
func (c *Console) getLineSelectionRange(lineIndex int) *selectionRange {
	sel := c.normalizeSelection()
	if sel == nil {
		return nil
	}

	if lineIndex < sel.StartLine || lineIndex > sel.EndLine {
		return nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()
	if lineIndex >= len(c.displayLines) {
		return nil
	}
	dl := c.displayLines[lineIndex]
	fullTextLen := len(dl.text)
	if dl.indent {
		fullTextLen += indentWidth
	}

	start := 0
	end := fullTextLen

	if lineIndex == sel.StartLine {
		start = maxInt(0, sel.StartCol)
	}
	if lineIndex == sel.EndLine {
		end = minInt(fullTextLen, sel.EndCol)
	}

	if start >= end {
		return nil
	}
	return &selectionRange{Start: start, End: end}
}

// selectionRange represents a selection range on a line.
type selectionRange struct {
	Start int
	End   int
}

// handleMouse handles a mouse event.
func (c *Console) handleMouse(x, y int, button uint32, pressed bool, scrollDirection string) bool {
	c.mu.RLock()
	if !c.isVisible {
		c.mu.RUnlock()
		return false
	}
	localX := x - c.consoleX
	localY := y - c.consoleY
	copyBounds := c.copyButtonBounds
	c.mu.RUnlock()

	if localX < 0 || localX >= c.consoleWidth || localY < 0 || localY >= c.consoleHeight {
		return false
	}

	if scrollDirection != "" {
		if scrollDirection == "up" {
			c.scrollUp()
		} else if scrollDirection == "down" {
			c.scrollDown()
		}
		return true
	}

	if localY == 0 {
		if pressed && button == 0 &&
			localX >= copyBounds.X && localX < copyBounds.X+copyBounds.Width {
			c.triggerCopy()
			return true
		}
		return true
	}

	lineIndex := c.scrollTopIndex + (localY - 1)
	colIndex := maxInt(0, localX-1)

	if pressed && button == 0 {
		c.clearSelection()
		c.mu.Lock()
		c.selectionStart = &selectionPos{Line: lineIndex, Col: colIndex}
		c.selectionEnd = &selectionPos{Line: lineIndex, Col: colIndex}
		c.isDragging = true
		c.mu.Unlock()
		c.needsUpdate = true
		c.renderer.RequestRender()
		return true
	}

	if c.isDragging {
		c.mu.Lock()
		c.selectionEnd = &selectionPos{Line: lineIndex, Col: colIndex}
		c.mu.Unlock()

		c.needsUpdate = true
		c.renderer.RequestRender()
		return true
	}

	if !pressed {
		if c.isDragging {
			c.mu.Lock()
			c.selectionEnd = &selectionPos{Line: lineIndex, Col: colIndex}
			c.isDragging = false
			c.mu.Unlock()
			c.stopAutoScroll()
			c.needsUpdate = true
			c.renderer.RequestRender()
		}
		return true
	}

	return true
}

// visible returns whether the console is visible.
func (c *Console) visible() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isVisible
}

// bounds returns the console bounds.
func (c *Console) bounds() (x, y, width, height int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.consoleX, c.consoleY, c.consoleWidth, c.consoleHeight
}

// saveLogsToFile saves all log entries to a file.
func (c *Console) saveLogsToFile() {
	c.mu.RLock()
	entries := make([]LogEntry, len(c.allLogEntries))
	copy(entries, c.allLogEntries)
	cachedLogs := globalConsoleCache.CachedLogs()
	debugMode := c.debugModeEnabled
	c.mu.RUnlock()

	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("_console_%d.log", timestamp)

	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}
	filepath := cwd + "/" + filename

	var logLines []string
	allEntries := append(entries, cachedLogs...)
	for _, entry := range allEntries {
		timestampStr := formatTimestamp(entry.Date)
		callerSource := "unknown"
		if entry.CallerInfo != nil {
			callerSource = fmt.Sprintf("%s:%d", entry.CallerInfo.FileName, entry.CallerInfo.LineNumber)
		}
		prefix := fmt.Sprintf("[%s] [%s]", timestampStr, entry.Level)
		if debugMode {
			prefix += fmt.Sprintf(" [%s]", callerSource)
		}
		prefix += " "
		formattedArgs := formatArguments(entry.Args)
		logLines = append(logLines, prefix+formattedArgs)
	}

	content := strings.Join(logLines, "\n")
	err = os.WriteFile(filepath, []byte(content), 0644)
	if err != nil {
		c.Error("Failed to save console logs:", err)
		return
	}
	c.Info(fmt.Sprintf("Console logs saved to: %s", filename))
}

// indentWidth is the number of spaces used for indentation of wrapped lines.
const indentWidth = 2

// maxInt returns the maximum of two integers.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// minInt returns the minimum of two integers.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// colorToUint16 converts an RGBA color to a uint16 for the buffer.
func colorToUint16(c RGBA) uint16 {
	r := uint16((c.R >> 3) & 0x1F)
	g := uint16((c.G >> 2) & 0x3F)
	b := uint16((c.B >> 3) & 0x1F)
	return uint16(r<<11 | g<<5 | b)
}

// wrapDisplayCells wraps text into display cell segments.
func wrapDisplayCells(text string, firstCells, restCells int) []string {
	if firstCells < 1 {
		firstCells = 1
	}
	if restCells < 1 {
		restCells = 1
	}

	var segments []string
	var segment strings.Builder
	cells := 0

	for _, r := range text {
		width := runeWidth(r)
		budget := firstCells
		if segment.Len() > 0 {
			budget = restCells
		}
		if segment.Len() > 0 && cells+width > budget {
			segments = append(segments, segment.String())
			segment.Reset()
			cells = 0
			budget = restCells
		}
		segment.WriteRune(r)
		cells += width
	}
	if segment.Len() > 0 {
		segments = append(segments, segment.String())
	}
	return segments
}

// runeWidth returns the display width of a rune.
func runeWidth(r rune) int {
	if r == '\t' {
		return 4
	}
	if r < 32 || r == 127 {
		return 0
	}
	return 1
}
