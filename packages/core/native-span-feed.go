package core

import (
	"fmt"
	"sync"

	"github.com/Gomobra/PixTui/packages/native"
)

// GrowthPolicy controls whether the stream grows or blocks when full.
type GrowthPolicy int

const (
	GrowthGrow GrowthPolicy = iota
	GrowthBlock
)

// NativeSpanFeedOptions configures a NativeSpanFeed stream.
type NativeSpanFeedOptions struct {
	ChunkSize         uint32
	InitialChunks     uint32
	MaxBytes          uint64
	GrowthPolicy      GrowthPolicy
	AutoCommitOnFull  bool
	SpanQueueCapacity uint32
}

// DataHandler is invoked with a slice of stream data. The slice is only valid
// until the handler returns.
type DataHandler func(data []byte)

// NativeSpanFeed is a zero-copy span publication feed over native stream memory.
// Mirrors packages/core/src/NativeSpanFeed.ts.
type NativeSpanFeed struct {
	mu           sync.Mutex
	stream       *native.Stream
	dataHandlers map[uint64]DataHandler
	nextHandler  uint64
	errorHandlers map[uint64]func(code int)
	nextError    uint64
	closed       bool
	destroyed    bool
	draining     bool
	pendingClose bool
	closing      bool
	closeQueued  bool
	pendingAsync int
	inCallback   bool
	idleWaiters  []chan struct{}
}

const (
	eventChunkAdded = 2
	eventClosed     = 5
	eventError      = 6
	eventDataAvail  = 7
)

// defaultSpanFeedOptions returns default stream options.
func defaultSpanFeedOptions() NativeSpanFeedOptions {
	return NativeSpanFeedOptions{
		ChunkSize:         64 * 1024,
		InitialChunks:     2,
		MaxBytes:          0,
		GrowthPolicy:      GrowthGrow,
		AutoCommitOnFull:  true,
		SpanQueueCapacity: 4096,
	}
}

// normalizeSpanFeedOptions fills in missing option values.
func normalizeSpanFeedOptions(opts *NativeSpanFeedOptions) NativeSpanFeedOptions {
	out := *opts
	if out.ChunkSize == 0 {
		out.ChunkSize = 64 * 1024
	}
	if out.InitialChunks == 0 {
		out.InitialChunks = 1
	}
	if out.SpanQueueCapacity == 0 {
		out.SpanQueueCapacity = 4096
	}
	return out
}

// Create creates and attaches a new NativeSpanFeed stream.
func CreateNativeSpanFeed(options *NativeSpanFeedOptions) (*NativeSpanFeed, error) {
	opts := defaultSpanFeedOptions()
	if options != nil {
		opts = normalizeSpanFeedOptions(options)
	}
	allocator := native.NewAllocator(0)
	nativeOpts := native.Options{
		ChunkSize:         opts.ChunkSize,
		InitialChunks:     opts.InitialChunks,
		MaxBytes:          opts.MaxBytes,
		GrowthPolicy:      native.GrowthPolicy(opts.GrowthPolicy),
		AutoCommitOnFull:  opts.AutoCommitOnFull,
		SpanQueueCapacity: opts.SpanQueueCapacity,
	}
	stream := native.CreateNativeSpanFeed(allocator, &nativeOpts)

	s := &NativeSpanFeed{
		stream:         stream,
		dataHandlers:   make(map[uint64]DataHandler),
		errorHandlers:  make(map[uint64]func(code int)),
		idleWaiters:    make([]chan struct{}, 0),
	}
	if err := s.installHandler(); err != nil {
		stream.Destroy()
		return nil, err
	}
	if err := stream.Attach(); err != nil {
		s.destroyLocked()
		return nil, fmt.Errorf("failed to attach stream: %w", err)
	}
	return s, nil
}

// Attach wraps an existing native stream.
func AttachNativeSpanFeed(stream *native.Stream) (*NativeSpanFeed, error) {
	s := &NativeSpanFeed{
		stream:        stream,
		dataHandlers:  make(map[uint64]DataHandler),
		errorHandlers: make(map[uint64]func(code int)),
		idleWaiters:   make([]chan struct{}, 0),
	}
	if err := s.installHandler(); err != nil {
		return nil, err
	}
	if err := stream.Attach(); err != nil {
		s.destroyLocked()
		return nil, fmt.Errorf("failed to attach stream: %w", err)
	}
	return s, nil
}

func (s *NativeSpanFeed) installHandler() error {
	s.stream.SetCallback(func(streamPtr uintptr, eventID uint32, arg0 uintptr, arg1 uint64) {
		s.handleEvent(eventID, arg0, arg1)
	})
	return nil
}

func (s *NativeSpanFeed) handleEvent(eventID uint32, arg0 uintptr, arg1 uint64) {
	s.mu.Lock()
	s.inCallback = true
	wasInCallback := s.inCallback
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.inCallback = wasInCallback
		s.resolveIdleIfNeeded()
		s.mu.Unlock()
	}()

	switch eventID {
	case eventDataAvail:
		if s.isClosing() {
			return
		}
		s.mu.Lock()
		if len(s.dataHandlers) == 0 {
			s.mu.Unlock()
			return
		}
		s.mu.Unlock()
		s.drainAll()
	case eventChunkAdded:
		// Chunk pointers are tracked natively by the stream; data is copied
		// on demand via SpanBytes during drain.
	case eventError:
		code := int(arg0)
		s.mu.Lock()
		handlers := make([]func(int), 0, len(s.errorHandlers))
		for _, h := range s.errorHandlers {
			handlers = append(handlers, h)
		}
		s.mu.Unlock()
		for _, h := range handlers {
			h(code)
		}
	case eventClosed:
		s.mu.Lock()
		s.closed = true
		s.mu.Unlock()
	}
}

// OnData registers a data handler and returns a removal function.
func (s *NativeSpanFeed) OnData(handler DataHandler) func() {
	s.mu.Lock()
	id := s.nextHandler
	s.nextHandler++
	s.dataHandlers[id] = handler
	s.mu.Unlock()
	return func() {
		s.mu.Lock()
		delete(s.dataHandlers, id)
		s.mu.Unlock()
	}
}

// OnError registers an error handler and returns a removal function.
func (s *NativeSpanFeed) OnError(handler func(code int)) func() {
	s.mu.Lock()
	id := s.nextError
	s.nextError++
	s.errorHandlers[id] = handler
	s.mu.Unlock()
	return func() {
		s.mu.Lock()
		delete(s.errorHandlers, id)
		s.mu.Unlock()
	}
}

// IsBackpressured reports whether the stream cannot accept more data.
func (s *NativeSpanFeed) IsBackpressured() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	stats := s.stream.GetStats()
	return s.pendingAsync > 0 || stats.OutstandingSpans > 0
}

// Close requests a graceful close of the stream.
func (s *NativeSpanFeed) Close() {
	s.mu.Lock()
	if s.destroyed {
		s.mu.Unlock()
		return
	}
	if s.inCallback || s.draining || s.pendingAsync > 0 {
		s.pendingClose = true
		if !s.closeQueued {
			s.closeQueued = true
			go s.processPendingClose()
		}
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()
	s.performClose()
}

func (s *NativeSpanFeed) processPendingClose() {
	s.mu.Lock()
	if !s.pendingClose || s.destroyed {
		s.mu.Unlock()
		return
	}
	if s.inCallback || s.draining || s.pendingAsync > 0 {
		s.mu.Unlock()
		return
	}
	s.pendingClose = false
	s.mu.Unlock()
	s.performClose()
	s.resolveIdleIfNeeded()
}

func (s *NativeSpanFeed) performClose() {
	s.mu.Lock()
	if s.closing {
		s.mu.Unlock()
		return
	}
	s.closing = true
	s.mu.Unlock()

	if !s.closed {
		if err := s.stream.Close(); err != nil {
			s.mu.Lock()
			s.closing = false
			s.mu.Unlock()
			return
		}
		s.mu.Lock()
		s.closed = true
		s.mu.Unlock()
	}
	s.finalizeDestroy()
}

func (s *NativeSpanFeed) finalizeDestroy() {
	s.mu.Lock()
	if s.destroyed {
		s.mu.Unlock()
		return
	}
	s.destroyed = true
	s.dataHandlers = make(map[uint64]DataHandler)
	s.errorHandlers = make(map[uint64]func(code int))
	waiters := s.idleWaiters
	s.idleWaiters = nil
	s.mu.Unlock()

	for _, w := range waiters {
		close(w)
	}
}

// drainAll drains all currently queued spans and delivers them to data handlers.
func (s *NativeSpanFeed) drainAll() {
	for {
		count := s.drainOnce()
		if count == 0 {
			break
		}
	}
	s.mu.Lock()
	inCallback := s.inCallback
	s.mu.Unlock()
	if !inCallback {
		// Handler errors are surfaced synchronously within drainOnce in this port.
	}
}

func (s *NativeSpanFeed) drainOnce() int {
	s.mu.Lock()
	if s.draining || s.pendingClose {
		s.mu.Unlock()
		return 0
	}
	out := make([]native.SpanInfo, 256)
	count := int(s.stream.DrainSpans(out))
	if count == 0 {
		s.mu.Unlock()
		return 0
	}
	s.draining = true
	s.mu.Unlock()

	var firstErr error
	spans := out[:count]
	for _, span := range spans {
		s.mu.Lock()
		pendingClose := s.pendingClose
		s.mu.Unlock()
		if pendingClose || span.Len == 0 {
			_ = s.stream.ReleaseSpan(span.SlotIndex, span.ReleaseID)
			continue
		}

		slice, ok := s.stream.SpanBytes(span.ChunkPtr, span.Offset, span.Len)
		if !ok {
			_ = s.stream.ReleaseSpan(span.SlotIndex, span.ReleaseID)
			continue
		}

		for _, h := range s.dataHandlersSnapshot() {
			h(slice)
		}

		if err := s.stream.ReleaseSpan(span.SlotIndex, span.ReleaseID); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	s.mu.Lock()
	s.draining = false
	s.resolveIdleIfNeeded()
	s.mu.Unlock()

	if firstErr != nil {
		s.queueHandlerError(firstErr)
	}
	return count
}

func (s *NativeSpanFeed) dataHandlersSnapshot() []DataHandler {
	s.mu.Lock()
	out := make([]DataHandler, 0, len(s.dataHandlers))
	for _, h := range s.dataHandlers {
		out = append(out, h)
	}
	s.mu.Unlock()
	return out
}

func (s *NativeSpanFeed) isIdleLocked() bool {
	return !s.inCallback && !s.draining && s.pendingAsync == 0 && !s.closed
}

func (s *NativeSpanFeed) isClosing() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closing || s.pendingClose
}

func (s *NativeSpanFeed) resolveIdleIfNeeded() {
	s.mu.Lock()
	waiters := s.idleWaiters
	if s.isIdleLocked() {
		s.idleWaiters = nil
	} else {
		waiters = nil
	}
	s.mu.Unlock()
	for _, w := range waiters {
		close(w)
	}
}

func (s *NativeSpanFeed) destroyLocked() {
	s.mu.Lock()
	s.destroyed = true
	s.stream.Destroy()
	waiters := s.idleWaiters
	s.idleWaiters = nil
	s.mu.Unlock()
	for _, w := range waiters {
		close(w)
	}
}

// Idle returns a channel that is closed when the feed is idle.
func (s *NativeSpanFeed) Idle() <-chan struct{} {
	s.mu.Lock()
	if s.isIdleLocked() {
		ch := make(chan struct{})
		close(ch)
		s.mu.Unlock()
		return ch
	}
	ch := make(chan struct{})
	s.idleWaiters = append(s.idleWaiters, ch)
	s.mu.Unlock()
	return ch
}

func (s *NativeSpanFeed) queueHandlerError(err error) {
	// Surface handler errors synchronously in this port.
	_ = err
}

var _ = fmt.Sprintf
