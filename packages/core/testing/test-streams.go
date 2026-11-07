package testing

import (
	"io"
	"sync"
)

// TestWriteStream is a test sink for writes. It implements io.Writer, discards
// all written data, and reports itself as a TTY with the given dimensions. It
// mirrors the TypeScript TestWriteStream.
type TestWriteStream struct {
	IsTTY   bool
	Columns int
	Rows    int
}

// NewTestWriteStream creates a TestWriteStream with the given columns and rows
// (default 80x24 when not supplied by the caller).
func NewTestWriteStream(columns, rows int) *TestWriteStream {
	return &TestWriteStream{IsTTY: true, Columns: columns, Rows: rows}
}

// Write discards the data and returns len(p), nil.
func (s *TestWriteStream) Write(p []byte) (int, error) {
	return len(p), nil
}

// GetColorDepth returns the color depth of the test terminal (24-bit truecolor).
func (s *TestWriteStream) GetColorDepth() int {
	return 24
}

// RecordingWriteStream is a TestWriteStream that records every write and supports
// holding writes (parking their acknowledgement) until Release() is called. This
// models a slow terminal or socket: the Session keeps a frame pending and later
// output is queued. It mirrors the TypeScript RecordingWriteStream.
type RecordingWriteStream struct {
	TestWriteStream
	mu        sync.Mutex
	writes    []byte
	held      bool
	releaseCh chan struct{}
	parked    int
}

// NewRecordingWriteStream creates a RecordingWriteStream with the given
// columns and rows (default 80x24 when not supplied by the caller).
func NewRecordingWriteStream(columns, rows int) *RecordingWriteStream {
	return &RecordingWriteStream{
		TestWriteStream: TestWriteStream{IsTTY: true, Columns: columns, Rows: rows},
		releaseCh:       make(chan struct{}),
	}
}

// Write records the chunk. If a hold is active, the write's acknowledgement is
// parked until Release() is called, at which point the call returns.
func (s *RecordingWriteStream) Write(p []byte) (int, error) {
	s.mu.Lock()
	s.writes = append(s.writes, append([]byte(nil), p...)...)
	if s.held {
		s.parked++
		ch := s.releaseCh
		s.mu.Unlock()
		<-ch
		s.mu.Lock()
		s.parked--
		s.mu.Unlock()
		return len(p), nil
	}
	s.mu.Unlock()
	return len(p), nil
}

// PendingWrite returns true while a write is waiting for Release().
func (s *RecordingWriteStream) PendingWrite() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.parked > 0
}

// Hold parks subsequent write acknowledgements until Release() is called.
func (s *RecordingWriteStream) Hold() {
	s.mu.Lock()
	s.held = true
	s.releaseCh = make(chan struct{})
	s.mu.Unlock()
}

// Release stops holding and completes any parked write.
func (s *RecordingWriteStream) Release() {
	s.mu.Lock()
	s.held = false
	ch := s.releaseCh
	s.releaseCh = nil
	s.mu.Unlock()
	if ch != nil {
		close(ch)
	}
}

// Bytes returns a copy of all recorded bytes.
func (s *RecordingWriteStream) Bytes() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]byte, len(s.writes))
	copy(out, s.writes)
	return out
}

// Text returns all recorded bytes as a string.
func (s *RecordingWriteStream) Text() string {
	return string(s.Bytes())
}

// Clear removes all recorded bytes.
func (s *RecordingWriteStream) Clear() {
	s.mu.Lock()
	s.writes = nil
	s.mu.Unlock()
}

// TestStdout is an io.Writer backed by a TestWriteStream, carrying TTY
// dimensions. It is the type returned by CreateTestStdout.
type TestStdout = *TestWriteStream

type testStdin struct{}

// Read yields no data, immediately returning EOF. It mirrors the TypeScript
// createTestStdin readable that never produces data.
func (testStdin) Read([]byte) (int, error) { return 0, io.EOF }

// CreateTestStdin returns an io.Reader that yields no data (immediately EOF).
func CreateTestStdin() io.Reader {
	return testStdin{}
}

// CreateTestStdout returns a TestWriteStream (io.Writer) with the given columns
// and rows (default 80x24).
func CreateTestStdout(columns, rows int) *TestWriteStream {
	return NewTestWriteStream(columns, rows)
}
