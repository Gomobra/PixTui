package testing

import "sync"

// SimpleHighlight is a single syntax highlight range produced by a tree-sitter
// query. It mirrors the TypeScript SimpleHighlight type from
// packages/core/src/lib/tree-sitter/types.js.
type SimpleHighlight struct {
	// Start is the inclusive start byte offset of the highlight.
	Start int
	// End is the exclusive end byte offset of the highlight.
	End int
	// StyleID identifies the syntax style applied to the highlighted range.
	StyleID int
}

// HighlightResult is the result of a highlight operation.
type HighlightResult struct {
	Highlights []SimpleHighlight
	Warning    string
	Error      string
}

// timerClock is a clock capable of scheduling one-shot timers. It is satisfied by
// ManualClock and is used to drive the mock's optional auto-resolve timeout.
type timerClock interface {
	SetTimeout(func(), float64) int
	ClearTimeout(int)
}

// MockTreeSitterClientOptions configures a MockTreeSitterClient.
type MockTreeSitterClientOptions struct {
	// AutoResolveTimeout, when non-nil, causes each HighlightOnce call to resolve
	// automatically after the given delay (in milliseconds) unless resolved
	// manually first.
	AutoResolveTimeout *float64
	// Clock drives the auto-resolve timeout. Defaults to a ManualClock.
	Clock timerClock
}

// highlightOnceCall tracks a single in-flight HighlightOnce request.
type highlightOnceCall struct {
	mu      sync.Mutex
	future  *HighlightOnceFuture
	timeout int
	resolved bool
}

// HighlightOnceFuture is the future returned by MockTreeSitterClient.HighlightOnce.
// Get blocks until the result is resolved (manually or by the auto-resolve timeout).
type HighlightOnceFuture struct {
	ch chan HighlightResult
}

// Get returns the highlight result, blocking until it is resolved.
func (f *HighlightOnceFuture) Get() HighlightResult {
	return <-f.ch
}

// MockTreeSitterClient is a test double for a tree-sitter client.
// It captures HighlightOnce invocations and resolves them either manually via
// ResolveHighlightOnce/ResolveAllHighlightOnce or automatically after a
// configured timeout.
type MockTreeSitterClient struct {
	mu                 sync.Mutex
	clock              timerClock
	autoResolveTimeout float64
	mockResult         HighlightResult
	highlightPromises  []*highlightOnceCall
}

// NewMockTreeSitterClient creates a new MockTreeSitterClient.
// With no options, highlights never auto-resolve and must be resolved manually.
func NewMockTreeSitterClient(options ...MockTreeSitterClientOptions) *MockTreeSitterClient {
	var opts MockTreeSitterClientOptions
	if len(options) > 0 {
		opts = options[0]
	}
	clock := opts.Clock
	if clock == nil {
		clock = NewManualClock()
	}
	autoResolveTimeout := 0.0
	if opts.AutoResolveTimeout != nil {
		autoResolveTimeout = *opts.AutoResolveTimeout
	}
	return &MockTreeSitterClient{
		clock:              clock,
		autoResolveTimeout: autoResolveTimeout,
		mockResult:         HighlightResult{Highlights: make([]SimpleHighlight, 0)},
		highlightPromises:  make([]*highlightOnceCall, 0),
	}
}

// HighlightOnce records a highlight request and returns a future that resolves
// with the configured mock result, either manually or after the auto-resolve timeout.
func (m *MockTreeSitterClient) HighlightOnce(_content, _filetype string) *HighlightOnceFuture {
	ch := make(chan HighlightResult, 1)
	future := &HighlightOnceFuture{ch: ch}
	call := &highlightOnceCall{future: future}

	var timeout int
	if m.autoResolveTimeout > 0 {
		timeout = m.clock.SetTimeout(func() {
			m.resolveHighlightCall(call, m.mockResult)
		}, m.autoResolveTimeout)
	}
	call.timeout = timeout

	m.mu.Lock()
	m.highlightPromises = append(m.highlightPromises, call)
	m.mu.Unlock()

	return future
}

func (m *MockTreeSitterClient) resolveHighlightCall(call *highlightOnceCall, result HighlightResult) {
	call.mu.Lock()
	if call.resolved {
		call.mu.Unlock()
		return
	}
	call.resolved = true
	timeout := call.timeout
	call.mu.Unlock()

	if timeout >= 0 {
		m.clock.ClearTimeout(timeout)
	}
	// Buffered channel, so this never blocks.
	call.future.ch <- result

	m.mu.Lock()
	for i, c := range m.highlightPromises {
		if c == call {
			m.highlightPromises = append(m.highlightPromises[:i], m.highlightPromises[i+1:]...)
			break
		}
	}
	m.mu.Unlock()
}

// SetMockResult sets the result returned by every HighlightOnce call.
func (m *MockTreeSitterClient) SetMockResult(result HighlightResult) {
	m.mu.Lock()
	m.mockResult = result
	m.mu.Unlock()
}

// ResolveHighlightOnce manually resolves the HighlightOnce request at index
// (default 0). Calls outside the valid range are ignored.
func (m *MockTreeSitterClient) ResolveHighlightOnce(index int) {
	m.mu.Lock()
	if index < 0 || index >= len(m.highlightPromises) {
		m.mu.Unlock()
		return
	}
	call := m.highlightPromises[index]
	m.mu.Unlock()
	m.resolveHighlightCall(call, m.mockResult)
}

// ResolveAllHighlightOnce manually resolves all in-flight HighlightOnce requests.
func (m *MockTreeSitterClient) ResolveAllHighlightOnce() {
	m.mu.Lock()
	calls := append([]*highlightOnceCall{}, m.highlightPromises...)
	m.highlightPromises = make([]*highlightOnceCall, 0)
	m.mu.Unlock()
	for _, call := range calls {
		m.resolveHighlightCall(call, m.mockResult)
	}
}

// IsHighlighting returns true if there are in-flight HighlightOnce requests.
func (m *MockTreeSitterClient) IsHighlighting() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.highlightPromises) > 0
}

// Destroy resolves all in-flight highlight requests and releases resources.
func (m *MockTreeSitterClient) Destroy() error {
	m.ResolveAllHighlightOnce()
	return nil
}
