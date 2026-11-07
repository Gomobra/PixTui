package native

import "sync"

// SplitScrollback tracks how many rows of a scrolled-back region have been
// published to the terminal surface, so subsequent content can be rendered at
// the correct offset without re-publishing already-visible rows.
// Mirrors packages/native/src/split-scrollback.zig.
type SplitScrollback struct {
	mu            sync.RWMutex
	publishedRows uint32
	tailColumn    uint32
}

// NewSplitScrollback creates a SplitScrollback with no published rows.
func NewSplitScrollback() *SplitScrollback {
	return &SplitScrollback{}
}

// Reset seeds the split scrollback with a starting row count.
func (s *SplitScrollback) Reset(seedRows uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.publishedRows = seedRows
	s.tailColumn = 0
}

// RenderOffset returns the number of rows to offset rendering by when the
// surface is at the given offset. Returns 0 when surface_offset is 0.
func (s *SplitScrollback) RenderOffset(surfaceOffset uint32) uint32 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if surfaceOffset == 0 {
		return 0
	}
	if s.publishedRows < surfaceOffset {
		return s.publishedRows
	}
	return surfaceOffset
}

// NoteViewportScroll decrements the published row count by the number of
// lines scrolled, clamping at zero. When all rows are scrolled away, the
// tail column is reset.
func (s *SplitScrollback) NoteViewportScroll(lines uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if lines > s.publishedRows {
		s.publishedRows = 0
	} else {
		s.publishedRows -= lines
	}
	if s.publishedRows == 0 {
		s.tailColumn = 0
	}
}

// NoteNewline records that a newline was published, incrementing the row
// count and resetting the tail column.
func (s *SplitScrollback) NoteNewline() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.publishedRows == 0 {
		s.publishedRows = 1
	}
	s.publishedRows++
	s.tailColumn = 0
}

// PublishSnapshotRows publishes a series of rows, each treated as ending with
// a trailing newline except possibly the last.
func (s *SplitScrollback) PublishSnapshotRows(rowCount, rowColumns, terminalWidth uint32, trailingNewline bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if rowCount == 0 {
		return
	}
	var row uint32
	for row = 0; row < rowCount; row++ {
		isLast := row+1 < rowCount || trailingNewline
		s.publishRow(rowColumns, terminalWidth, isLast)
	}
}

// publishRow publishes a single row of the given column count.
func (s *SplitScrollback) publishRow(columns, width uint32, trailingNewline bool) {
	s.publishColumns(columns, width)
	if trailingNewline {
		s.noteNewlineLocked()
	}
}

// publishColumns accounts for a number of columns wrapping across the
// terminal width, incrementing published rows as needed.
func (s *SplitScrollback) publishColumns(columns, width uint32) {
	if columns == 0 {
		return
	}
	safeWidth := width
	if safeWidth < 1 {
		safeWidth = 1
	}
	remaining := columns
	for remaining > 0 {
		if s.publishedRows == 0 {
			s.publishedRows = 1
		}
		if s.tailColumn >= safeWidth {
			s.publishedRows++
			s.tailColumn = 0
		}
		availableWidth := safeWidth - s.tailColumn
		step := remaining
		if step > availableWidth {
			step = availableWidth
		}
		s.tailColumn += step
		remaining -= step
		if remaining > 0 && s.tailColumn >= safeWidth {
			s.publishedRows++
			s.tailColumn = 0
		}
	}
}

// noteNewlineLocked is the unlocked variant of NoteNewline; caller holds the
// write lock.
func (s *SplitScrollback) noteNewlineLocked() {
	if s.publishedRows == 0 {
		s.publishedRows = 1
	}
	s.publishedRows++
	s.tailColumn = 0
}

// PublishedRows returns the current published row count (for testing).
func (s *SplitScrollback) PublishedRows() uint32 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.publishedRows
}

// TailColumn returns the current tail column (for testing).
func (s *SplitScrollback) TailColumn() uint32 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tailColumn
}
