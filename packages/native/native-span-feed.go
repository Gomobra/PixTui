package native

import (
	"errors"
	"math"
	"sync"
	"unsafe"
)

// GrowthPolicy controls whether the stream grows or blocks when full.
type GrowthPolicy int

const (
	GrowthGrow GrowthPolicy = iota
	GrowthBlock
)

// Options configure a NativeSpanFeed stream.
type Options struct {
	ChunkSize         uint32
	InitialChunks     uint32
	MaxBytes          uint64
	GrowthPolicy      GrowthPolicy
	AutoCommitOnFull  bool
	SpanQueueCapacity uint32
}

// Stats reports stream statistics.
type Stats struct {
	BytesWritten     uint64
	SpansCommitted   uint64
	Chunks           uint32
	PendingSpans     uint32
	OutstandingSpans uint32
	Reserved         uint32
	OutstandingBytes uint64
}

// StreamError classifies stream failures.
type StreamError int

const (
	ErrNoSpace StreamError = iota
	ErrMaxBytes
	ErrInvalid
	ErrOutOfMemory
	ErrBusy
)

func (e StreamError) Error() string {
	switch e {
	case ErrNoSpace:
		return "no space"
	case ErrMaxBytes:
		return "max bytes exceeded"
	case ErrInvalid:
		return "invalid operation"
	case ErrOutOfMemory:
		return "out of memory"
	case ErrBusy:
		return "stream busy"
	default:
		return "unknown stream error"
	}
}

// SpanInfo describes a committed span of bytes in a chunk.
type SpanInfo struct {
	ChunkPtr   uintptr
	Offset     uint32
	Len        uint32
	ChunkIndex uint32
	SlotIndex  uint32
	ReleaseID  uint64
}

// chunk is a single fixed-size buffer in the stream's chunk pool.
type chunk struct {
	ptr      []byte
	len      uint32
	refcount uint8
}

const spanRingNone = uint32(math.MaxUint32)

const (
	spanStateFree     uint8 = 0
	spanStateQueued   uint8 = 1
	spanStateBorrowed uint8 = 2
)

// ControlSequenceReservation holds headroom for one control publication.
type ControlSequenceReservation struct {
	Bytes uint64
	Spans uint64
}

// admissionKind classifies an admission check.
type admissionKind int

const (
	admissionOrdinary admissionKind = iota
	admissionReservedControl
	admissionControlCapacity
)

// CallbackFn is the completion callback signature.
type CallbackFn func(streamPtr uintptr, eventID uint32, arg0 uintptr, arg1 uint64)

const (
	eventChunkAdded uint32 = 2
	eventClosed     uint32 = 5
	eventError      uint32 = 6
	eventDataAvail  uint32 = 7
)

// spanRing is a ring of span slots with a free list and a queue.
type spanRing struct {
	buffer   []spanRingEntry
	capacity uint32
	head     uint32
	tail     uint32
	freeHead uint32
	queued   uint32
	borrowed uint32
	bytes    uint64
	nextID   uint64
}

type spanRingEntry struct {
	span  SpanInfo
	next  uint32
	state uint8
}

func newSpanRing(capacity uint32) spanRing {
	ring := spanRing{
		capacity: capacity,
		head:     spanRingNone,
		tail:     spanRingNone,
		freeHead: spanRingNone,
		nextID:   1,
	}
	if capacity > 0 {
		ring.buffer = make([]spanRingEntry, capacity)
		for i := uint32(0); i < capacity; i++ {
			var next uint32 = spanRingNone
			if i+1 < capacity {
				next = i + 1
			}
			ring.buffer[i] = spanRingEntry{next: next, state: spanStateFree}
		}
		ring.freeHead = 0
	}
	return ring
}

func (r *spanRing) grow(stream *Stream) error {
	if stream.options.GrowthPolicy == GrowthBlock {
		return ErrNoSpace
	}
	oldCap := r.capacity
	newCap := oldCap * 2
	if newCap < oldCap {
		return ErrNoSpace
	}
	newBuf := make([]spanRingEntry, newCap)
	copy(newBuf, r.buffer)
	for i := oldCap; i < newCap; i++ {
		var next uint32 = spanRingNone
		if i+1 < newCap {
			next = i + 1
		}
		newBuf[i] = spanRingEntry{next: next, state: spanStateFree}
	}
	r.buffer = newBuf
	r.capacity = newCap
	r.freeHead = oldCap
	return nil
}

func (r *spanRing) push(stream *Stream, span SpanInfo) error {
	if r.queued >= r.capacity {
		if err := r.ensureAdditionalCapacity(stream, 1); err != nil {
			return err
		}
	}
	index := r.freeHead
	if index >= r.capacity {
		return ErrNoSpace
	}
	entry := &r.buffer[index]
	if entry.state != spanStateFree {
		return ErrNoSpace
	}
	r.freeHead = entry.next
	entry.span = span
	entry.state = spanStateQueued
	entry.span.SlotIndex = index
	entry.span.ReleaseID = r.nextID
	r.nextID++
	if r.tail == spanRingNone {
		r.head = index
	} else {
		r.buffer[r.tail].next = index
	}
	r.tail = index
	r.queued++
	r.bytes += uint64(span.Len)
	return nil
}

func (r *spanRing) ensureAdditionalCapacity(stream *Stream, additional uint32) error {
	unpublished := uint64(additional) + uint64(stream.controlChunks) - uint64(stream.controlSpans)
	if math.MaxUint64-r.nextID < unpublished {
		return ErrNoSpace
	}
	required := uint64(r.queued) + uint64(r.borrowed) + unpublished
	for required > uint64(r.capacity) {
		if err := r.grow(stream); err != nil {
			return err
		}
	}
	return nil
}

func (r *spanRing) popMany(out []SpanInfo) uint32 {
	toRead := r.queued
	if uint32(len(out)) < toRead {
		toRead = uint32(len(out))
	}
	for i := uint32(0); i < toRead; i++ {
		entry := &r.buffer[r.head]
		if entry.state != spanStateQueued {
			break
		}
		out[i] = entry.span
		entry.state = spanStateBorrowed
		r.head = entry.next
	}
	r.queued -= toRead
	r.borrowed += toRead
	if r.queued == 0 {
		r.tail = spanRingNone
	}
	return toRead
}

func (r *spanRing) release(slotIndex uint32, releaseID uint64) (SpanInfo, error) {
	if slotIndex >= r.capacity {
		return SpanInfo{}, ErrInvalid
	}
	entry := &r.buffer[slotIndex]
	if entry.state != spanStateBorrowed || entry.span.ReleaseID != releaseID {
		return SpanInfo{}, ErrInvalid
	}
	span := entry.span
	entry.state = spanStateFree
	entry.next = r.freeHead
	r.freeHead = slotIndex
	r.borrowed--
	r.bytes -= uint64(span.Len)
	return span, nil
}

// defaultOptions returns the default stream options.
func defaultOptions() Options {
	return Options{
		ChunkSize:         64 * 1024,
		InitialChunks:     2,
		MaxBytes:          0,
		GrowthPolicy:      GrowthGrow,
		AutoCommitOnFull:  true,
		SpanQueueCapacity: 4096,
	}
}

func normalizeOptions(opts *Options) Options {
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

// Stream is a chunked byte stream with a span ring for zero-copy publication.
// Mirrors packages/native/src/native-span-feed.zig.
type Stream struct {
	mu              sync.Mutex
	allocator       *Allocator
	options         Options
	chunks          []*chunk
	currentIdx      uint32
	writeOffset     uint32
	pendingIdx      uint32
	pendingOffset   uint32
	pendingLen      uint32
	reservedActive  bool
	reservedIdx     uint32
	reservedOffset  uint32
	reservedLen     uint32
	attached        bool
	callback        CallbackFn
	inCallback      bool
	producing       bool
	closed          bool
	spanRing        spanRing
	stagedBytes     uint64
	controlChunks   uint32
	controlSpans    uint32
	controlBytes    uint64
	controlSequence ControlSequenceReservation
	stats           Stats
}

// CreateNativeSpanFeed creates a new stream with the given allocator and options.
func CreateNativeSpanFeed(allocator *Allocator, options *Options) *Stream {
	opts := defaultOptions()
	if options != nil {
		opts = normalizeOptions(options)
	}
	s := &Stream{
		allocator: allocator,
		options:   opts,
		spanRing:  newSpanRing(opts.SpanQueueCapacity),
		stats:     Stats{},
	}
	for i := uint32(0); i < opts.InitialChunks; i++ {
		if err := s.addChunkLocked(); err != nil {
			break
		}
	}
	s.stats.Chunks = uint32(len(s.chunks))
	return s
}

// Attach marks the stream as attached and notifies the callback of existing
// chunks and queued spans.
func (s *Stream) Attach() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.producing {
		return ErrBusy
	}
	if s.closed {
		return ErrInvalid
	}
	notify := false
	var queued uint32
	defer func() { s.finish(notify, queued) }()
	s.attached = true
	if s.callback == nil {
		return nil
	}
	for _, ch := range s.chunks {
		s.emitChunkAdded(ch)
	}
	queued = s.spanRing.queued
	if queued > 0 {
		notify = true
	}
	return nil
}

// SetCallback installs a completion callback.
func (s *Stream) SetCallback(cb CallbackFn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.callback = cb
	if cb == nil || !s.attached {
		return
	}
	for _, ch := range s.chunks {
		s.emitChunkAdded(ch)
	}
	queued := s.spanRing.queued
	if queued > 0 {
		s.emitDataAvailable(queued)
	}
}

// ReserveControlCapacity installs control-output capacity headroom.
func (s *Stream) ReserveControlCapacity(bytes uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.producing {
		return ErrBusy
	}
	if s.closed {
		return ErrInvalid
	}
	if s.controlChunks != 0 || s.pendingLen != 0 || s.reservedActive || s.stagedBytes != 0 ||
		s.spanRing.queued != 0 || s.spanRing.borrowed != 0 {
		return ErrBusy
	}
	if bytes == 0 {
		return ErrInvalid
	}
	chunkSize := uint64(s.options.ChunkSize)
	count := (bytes + chunkSize - 1) / chunkSize
	if count > math.MaxUint32 {
		return ErrNoSpace
	}
	required := uint32(count)
	if uint32(len(s.chunks)) < required {
		return ErrNoSpace
	}
	if required > s.spanRing.capacity {
		return ErrNoSpace
	}
	s.controlChunks = required
	s.writeOffset = s.options.ChunkSize
	s.pendingOffset = s.writeOffset
	return nil
}

// SetControlSequenceReservation sets byte-counter and release-ID headroom.
func (s *Stream) SetControlSequenceReservation(res ControlSequenceReservation) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.producing {
		return ErrBusy
	}
	if s.closed {
		return ErrInvalid
	}
	if s.pendingLen != 0 || s.reservedActive || s.stagedBytes != 0 || s.controlSpans != 0 {
		return ErrBusy
	}
	if (res.Bytes == 0) != (res.Spans == 0) {
		return ErrInvalid
	}
	if res.Bytes != 0 && s.controlChunks == 0 {
		return ErrInvalid
	}
	if math.MaxUint64-s.stats.BytesWritten < res.Bytes {
		return ErrNoSpace
	}
	if math.MaxUint64-s.spanRing.nextID < res.Spans {
		return ErrNoSpace
	}
	s.controlSequence = res
	return nil
}

// Write copies data into the stream's chunk pool.
func (s *Stream) Write(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.writeLocked(data)
}

func (s *Stream) writeLocked(data []byte) error {
	if s.producing {
		return ErrBusy
	}
	if s.closed {
		return ErrInvalid
	}
	if len(data) == 0 {
		return nil
	}
	if s.reservedActive {
		return ErrBusy
	}
	notify := false
	defer func() { s.finish(notify, 0) }()
	s.producing = true
	defer func() { s.producing = false }()

	remaining := len(data)
	srcIndex := 0
	autoCommit := s.options.AutoCommitOnFull
	chunkLen := uint64(s.options.ChunkSize)

	for remaining > 0 {
		if s.pendingLen == 0 {
			if err := s.checkAdmission(0, 1, admissionOrdinary); err != nil {
				return err
			}
		}
		available := chunkLen - uint64(s.writeOffset)
		if available == 0 {
			if s.pendingLen > 0 {
				if err := s.commitLocked(&notify); err != nil {
					return err
				}
			}
			if err := s.ensureWritableChunkLocked(); err != nil {
				return err
			}
			available = chunkLen
		}
		if uint64(remaining) > available && !autoCommit {
			return ErrNoSpace
		}
		toWrite := uint64(remaining)
		if toWrite > available {
			toWrite = available
		}
		if err := s.checkAdmission(toWrite, 0, admissionOrdinary); err != nil {
			return err
		}
		if s.pendingLen == 0 {
			s.pendingIdx = s.currentIdx
			s.pendingOffset = s.writeOffset
		}
		ch := s.chunks[s.currentIdx]
		copy(ch.ptr[s.writeOffset:s.writeOffset+uint32(toWrite)], data[srcIndex:srcIndex+int(toWrite)])
		s.writeOffset += uint32(toWrite)
		s.pendingLen += uint32(toWrite)
		s.stats.BytesWritten += uint64(toWrite)
		srcIndex += int(toWrite)
		remaining -= int(toWrite)
		if uint64(s.writeOffset) == chunkLen && autoCommit {
			if err := s.commitLocked(&notify); err != nil {
				return err
			}
			if remaining > 0 {
				if err := s.ensureWritableChunkLocked(); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// WriteAtomic publishes one complete byte sequence atomically.
func (s *Stream) WriteAtomic(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	notify, err := s.publishAtomic(data, admissionOrdinary)
	if err != nil {
		return err
	}
	s.finish(notify, 0)
	return nil
}

// WriteReservedControlAtomic publishes one packet using control storage.
func (s *Stream) WriteReservedControlAtomic(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	notify, err := s.publishAtomic(data, admissionReservedControl)
	if err != nil {
		return err
	}
	s.finish(notify, 0)
	return nil
}

func (s *Stream) publishAtomic(data []byte, kind admissionKind) (bool, error) {
	if kind == admissionControlCapacity {
		return false, ErrInvalid
	}
	control := kind != admissionOrdinary
	if s.producing {
		return false, ErrBusy
	}
	if s.closed {
		return false, ErrInvalid
	}
	if len(data) == 0 {
		return false, nil
	}
	if s.reservedActive || s.pendingLen != 0 {
		return false, ErrBusy
	}
	if control {
		if s.controlChunks == 0 {
			return false, ErrInvalid
		}
		if s.stagedBytes != 0 || s.controlSpans != 0 {
			return false, ErrBusy
		}
	}
	s.producing = true
	defer func() { s.producing = false }()

	chunkSize := uint64(s.options.ChunkSize)
	requiredChunks := (uint64(len(data)) + chunkSize - 1) / chunkSize
	if requiredChunks > math.MaxUint32 {
		return false, ErrNoSpace
	}
	if kind == admissionReservedControl {
		if uint64(len(data)) > s.controlSequence.Bytes || requiredChunks > s.controlSequence.Spans {
			return false, ErrNoSpace
		}
	}
	if control {
		if requiredChunks > uint64(s.controlChunks) {
			return false, ErrNoSpace
		}
		s.controlSpans = uint32(requiredChunks)
		s.controlBytes = uint64(len(data))
	}
	defer func() {
		if control {
			s.controlSpans = 0
			s.controlBytes = 0
		}
	}()
	if err := s.checkAdmission(uint64(len(data)), uint32(requiredChunks), kind); err != nil {
		return false, err
	}
	if err := s.spanRing.ensureAdditionalCapacity(s, uint32(requiredChunks)); err != nil {
		return false, err
	}

	if !control {
		free := uint64(0)
		for i := uint32(0); i < uint32(len(s.chunks)); i++ {
			if s.isChunkFree(i) {
				free++
			}
		}
		if free < requiredChunks {
			if s.options.GrowthPolicy == GrowthBlock {
				return false, ErrNoSpace
			}
			missing := requiredChunks - free
			for missing > 0 {
				if err := s.addChunkLocked(); err != nil {
					return false, err
				}
				missing--
			}
		}
	}

	notify := false
	sourceOffset := uint64(0)
	lastChunkIndex := s.currentIdx
	chunkCount := uint32(len(s.chunks))
	if control {
		chunkCount = s.controlChunks
	}
	for i := uint32(0); i < chunkCount; i++ {
		if control {
			ch := s.chunks[i]
			if ch.refcount != 0 {
				continue
			}
		} else {
			if !s.isChunkFree(i) {
				continue
			}
		}
		ch := s.chunks[i]
		avail := uint64(ch.len)
		if avail > uint64(len(data))-sourceOffset {
			avail = uint64(len(data)) - sourceOffset
		}
		if avail == 0 {
			continue
		}
		copy(ch.ptr[0:avail], data[sourceOffset:sourceOffset+avail])
		span := SpanInfo{
			ChunkPtr:   uintptr(slicePtr(ch.ptr)),
			Offset:     0,
			Len:        uint32(avail),
			ChunkIndex: i,
		}
		if err := s.spanRing.push(s, span); err != nil {
			return false, err
		}
		s.markSpanPending(i)
		s.stats.SpansCommitted++
		sourceOffset += avail
		lastChunkIndex = i
		if sourceOffset == uint64(len(data)) {
			break
		}
	}
	if sourceOffset != uint64(len(data)) {
		return false, ErrNoSpace
	}
	if !control {
		s.currentIdx = lastChunkIndex
		s.writeOffset = s.chunks[lastChunkIndex].len
		s.pendingIdx = lastChunkIndex
		s.pendingOffset = s.writeOffset
		s.pendingLen = 0
	}
	s.stats.BytesWritten += uint64(len(data))
	if kind == admissionReservedControl {
		if s.controlSequence.Bytes >= uint64(len(data)) {
			s.controlSequence.Bytes -= uint64(len(data))
		}
		if s.controlSequence.Spans >= requiredChunks {
			s.controlSequence.Spans -= requiredChunks
		}
	}
	return notify, nil
}

// Reserve returns a zero-copy reservation into the current chunk.
func (s *Stream) Reserve(minLen uint32) (ReserveInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.producing {
		return ReserveInfo{}, ErrBusy
	}
	if s.closed {
		return ReserveInfo{}, ErrInvalid
	}
	s.producing = true
	defer func() { s.producing = false }()
	return s.reserveLocked(minLen)
}

func (s *Stream) reserveLocked(minLen uint32) (ReserveInfo, error) {
	if s.reservedActive {
		return ReserveInfo{}, ErrBusy
	}
	if s.pendingLen != 0 {
		return ReserveInfo{}, ErrBusy
	}
	if err := s.checkAdmission(uint64(minLen), 1, admissionOrdinary); err != nil {
		return ReserveInfo{}, err
	}
	if err := s.ensureWritableChunkLocked(); err != nil {
		return ReserveInfo{}, err
	}
	ch := s.chunks[s.currentIdx]
	available := uint64(ch.len) - uint64(s.writeOffset)
	counterBytes := math.MaxUint64 - s.stats.BytesWritten - s.controlSequence.Bytes - uint64(s.stagedBytes)
	if available > counterBytes {
		available = counterBytes
	}
	maxBytes := s.byteLimit()
	if maxBytes != 0 {
		used := uint64(s.spanRing.bytes) + uint64(s.stagedBytes) + s.controlHeldBytes()
		if maxBytes > used {
			avail := maxBytes - used
			if available > avail {
				available = avail
			}
		}
	}
	if uint64(minLen) > available {
		return ReserveInfo{}, ErrNoSpace
	}
	s.reservedActive = true
	s.reservedIdx = s.currentIdx
	s.reservedOffset = s.writeOffset
	s.reservedLen = uint32(available)
	return ReserveInfo{
		Ptr:      uintptr(slicePtr(ch.ptr[s.writeOffset:])),
		Len:      uint32(available),
		Reserved: 0,
	}, nil
}

// CommitReserved commits len bytes of a previously reserved region.
func (s *Stream) CommitReserved(length uint32) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.producing {
		return ErrBusy
	}
	if s.closed {
		return ErrInvalid
	}
	notify := false
	defer func() { s.finish(notify, 0) }()
	return s.commitReservedLocked(length, &notify)
}

// Commit commits any pending span.
func (s *Stream) Commit() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.producing {
		return ErrBusy
	}
	if s.closed {
		return ErrInvalid
	}
	notify := false
	defer func() { s.finish(notify, 0) }()
	if s.reservedActive {
		return ErrBusy
	}
	return s.commitLocked(&notify)
}

// DrainSpans drains queued spans into out, returning the number drained.
func (s *Stream) DrainSpans(out []SpanInfo) uint32 {
	return s.drainSpans(out)
}

// SpanBytes returns a copy of the span's data, located by its chunk pointer.
func (s *Stream) SpanBytes(chunkPtr uintptr, offset, length uint32) ([]byte, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, ch := range s.chunks {
		if uintptr(slicePtr(ch.ptr)) == chunkPtr {
			if int(offset)+int(length) > int(ch.len) {
				return nil, false
			}
			out := make([]byte, length)
			copy(out, ch.ptr[offset:offset+length])
			return out, true
		}
	}
	return nil, false
}

// GetStats returns current stream statistics.
func (s *Stream) GetStats() Stats {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := s.stats
	out.PendingSpans = s.spanRing.queued
	out.OutstandingSpans = s.spanRing.queued + s.spanRing.borrowed
	out.OutstandingBytes = s.spanRing.bytes
	return out
}

func (s *Stream) bounded() bool {
	return s.options.MaxBytes != 0 || s.options.GrowthPolicy == GrowthBlock
}

func (s *Stream) byteLimit() uint64 {
	if s.options.GrowthPolicy == GrowthBlock {
		poolBytes := uint64(s.options.ChunkSize) * uint64(len(s.chunks))
		if s.options.MaxBytes == 0 {
			return poolBytes
		}
		if poolBytes < s.options.MaxBytes {
			return poolBytes
		}
		return s.options.MaxBytes
	}
	return s.options.MaxBytes
}

// atomicByteLimit returns the byte headroom usable for a single atomic frame
// publication. When the stream is bounded this is the byte limit; otherwise
// there is no effective bound.
func (s *Stream) atomicByteLimit() uint64 {
	if s.bounded() {
		return s.byteLimit()
	}
	return ^uint64(0)
}

func (s *Stream) stagedSpans() uint64 {
	chunkSize := uint64(s.options.ChunkSize)
	if chunkSize == 0 {
		return 0
	}
	return (uint64(s.stagedBytes) + chunkSize - 1) / chunkSize
}

func (s *Stream) controlHeldBytes() uint64 {
	return uint64(s.controlChunks)*uint64(s.options.ChunkSize) - s.controlBytes
}

func (s *Stream) checkAdmission(bytes uint64, spans uint32, kind admissionKind) error {
	producerSpan := uint64(0)
	if s.pendingLen != 0 || s.reservedActive {
		producerSpan = 1
	}
	controlSlots := uint64(s.controlChunks) - uint64(s.controlSpans)
	var heldIDs uint64
	if kind == admissionReservedControl {
		heldIDs = s.controlSequence.Spans - uint64(spans)
	} else {
		if uint64(s.controlSpans) > controlSlots {
			heldIDs = uint64(s.controlSpans)
		} else {
			heldIDs = controlSlots
		}
	}
	remainingIDs := math.MaxUint64 - s.spanRing.nextID
	for _, used := range []uint64{producerSpan, s.stagedSpans(), uint64(spans), heldIDs} {
		if used > remainingIDs {
			return ErrNoSpace
		}
		remainingIDs -= used
	}
	requiredSpans := uint64(s.spanRing.queued) + uint64(s.spanRing.borrowed) + producerSpan + uint64(spans) + controlSlots + s.stagedSpans()
	if s.bounded() && requiredSpans > uint64(s.options.SpanQueueCapacity) {
		return ErrNoSpace
	}
	if kind != admissionControlCapacity {
		var heldBytes uint64
		if kind == admissionReservedControl {
			if s.controlSequence.Bytes > bytes {
				heldBytes = s.controlSequence.Bytes - bytes
			}
		} else {
			heldBytes = s.controlSequence.Bytes
		}
		remainingBytes := math.MaxUint64 - s.stats.BytesWritten
		for _, used := range []uint64{heldBytes, uint64(s.stagedBytes), uint64(s.reservedLen), bytes} {
			if used > remainingBytes {
				return ErrNoSpace
			}
			remainingBytes -= used
		}
	}
	maxBytes := s.byteLimit()
	if maxBytes == 0 {
		return nil
	}
	remaining := maxBytes
	for _, used := range []uint64{s.controlHeldBytes(), uint64(s.spanRing.bytes), uint64(s.pendingLen), uint64(s.reservedLen), uint64(s.stagedBytes), bytes} {
		if used > remaining {
			if s.options.GrowthPolicy == GrowthBlock {
				return ErrNoSpace
			}
			return ErrMaxBytes
		}
		remaining -= used
	}
	return nil
}

// HasAtomicCapacity reports whether at least one atomic byte can be written.
func (s *Stream) HasAtomicCapacity() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.checkAdmission(1, 1, admissionOrdinary); err != nil {
		return false
	}
	for i := uint32(0); i < uint32(len(s.chunks)); i++ {
		if s.isChunkFree(i) {
			return true
		}
	}
	if s.options.GrowthPolicy == GrowthBlock {
		return false
	}
	if uint32(len(s.chunks)) == math.MaxUint32 {
		return false
	}
	grownBytes := (uint64(len(s.chunks)) + 1) * uint64(s.options.ChunkSize)
	return s.options.MaxBytes == 0 || grownBytes <= s.options.MaxBytes
}

// SetStagedBytes sets the staged byte count.
func (s *Stream) SetStagedBytes(bytes uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.producing {
		return ErrBusy
	}
	if bytes <= uint64(s.stagedBytes) {
		s.stagedBytes = bytes
		return nil
	}
	if s.closed {
		return ErrInvalid
	}
	chunkSize := uint64(s.options.ChunkSize)
	spans := (bytes + chunkSize - 1) / chunkSize
	added := spans - s.stagedSpans()
	if added > math.MaxUint32 {
		return ErrNoSpace
	}
	if err := s.checkAdmission(bytes-uint64(s.stagedBytes), uint32(added), admissionOrdinary); err != nil {
		return err
	}
	s.stagedBytes = bytes
	return nil
}

func (s *Stream) hasPendingBytes() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pendingLen > 0
}

func (s *Stream) hasPendingSpans() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.spanRing.queued > 0
}

// ReleaseSpan releases a borrowed span back to the ring.
func (s *Stream) ReleaseSpan(slotIndex uint32, releaseID uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	span, err := s.spanRing.release(slotIndex, releaseID)
	if err != nil {
		return err
	}
	if int(span.ChunkIndex) >= len(s.chunks) {
		return ErrInvalid
	}
	ch := s.chunks[span.ChunkIndex]
	if ch.refcount == 0 {
		return ErrInvalid
	}
	ch.refcount--
	if span.ChunkIndex < s.controlChunks {
		s.controlSpans--
		s.controlBytes -= uint64(span.Len)
	}
	return nil
}

// MarkSpanConsumed releases a span, ignoring errors.
func (s *Stream) MarkSpanConsumed(span SpanInfo) {
	_ = s.ReleaseSpan(span.SlotIndex, span.ReleaseID)
}

func (s *Stream) finish(notify bool, queuedOverride uint32) {
	if notify && s.callback != nil {
		queued := queuedOverride
		if queued == 0 {
			queued = s.spanRing.queued
		}
		if queued > 0 {
			s.emitDataAvailable(queued)
		}
	}
}

func (s *Stream) isChunkFree(index uint32) bool {
	if index < s.controlChunks {
		return false
	}
	if int(index) >= len(s.chunks) {
		return false
	}
	return s.chunks[index].refcount == 0
}

func (s *Stream) commitLocked(notify *bool) error {
	if s.pendingLen == 0 {
		return nil
	}
	ch := s.chunks[s.pendingIdx]
	span := SpanInfo{
		ChunkPtr:   uintptr(slicePtr(ch.ptr)),
		Offset:     s.pendingOffset,
		Len:        s.pendingLen,
		ChunkIndex: s.pendingIdx,
	}
	if err := s.spanRing.push(s, span); err != nil {
		return err
	}
	s.markSpanPending(s.pendingIdx)
	s.stats.SpansCommitted++
	s.pendingLen = 0
	s.pendingOffset = s.writeOffset
	s.pendingIdx = s.currentIdx
	return nil
}

func (s *Stream) markSpanPending(chunkIndex uint32) {
	if int(chunkIndex) >= len(s.chunks) {
		return
	}
	ch := s.chunks[chunkIndex]
	if ch.refcount == 255 {
		s.writeOffset = s.options.ChunkSize
		return
	}
	ch.refcount++
}

func (s *Stream) commitReservedLocked(length uint32, notify *bool) error {
	if !s.reservedActive {
		return ErrInvalid
	}
	if length > s.reservedLen {
		return ErrNoSpace
	}
	s.pendingIdx = s.reservedIdx
	s.pendingOffset = s.reservedOffset
	s.pendingLen = length
	s.writeOffset = s.reservedOffset + length
	s.reservedActive = false
	s.reservedLen = 0
	s.stats.BytesWritten += uint64(length)
	return s.commitLocked(notify)
}

func (s *Stream) addChunkLocked() error {
	chunkSize := uint64(s.options.ChunkSize)
	maxBytes := s.options.MaxBytes
	allocated := uint64(len(s.chunks)) * chunkSize
	if maxBytes != 0 && allocated+chunkSize > maxBytes {
		return ErrMaxBytes
	}
	if uint32(len(s.chunks)) == math.MaxUint32 {
		return ErrNoSpace
	}
	mem := s.allocator.Alloc(int(chunkSize))
	if mem == nil {
		return ErrOutOfMemory
	}
	ch := &chunk{ptr: mem, len: uint32(chunkSize)}
	s.chunks = append(s.chunks, ch)
	s.stats.Chunks = uint32(len(s.chunks))
	if s.attached && s.callback != nil {
		s.emitChunkAdded(ch)
	}
	return nil
}

func (s *Stream) ensureWritableChunkLocked() error {
	total := uint32(len(s.chunks))
	if total == 0 {
		return ErrInvalid
	}
	attempts := uint32(0)
	index := s.currentIdx % total
	for attempts < total {
		if s.isChunkFree(index) {
			s.currentIdx = index
			s.writeOffset = 0
			s.pendingIdx = index
			s.pendingOffset = 0
			s.pendingLen = 0
			return nil
		}
		index = (index + 1) % total
		attempts++
	}
	if s.options.GrowthPolicy == GrowthBlock {
		return ErrNoSpace
	}
	if err := s.addChunkLocked(); err != nil {
		return err
	}
	newTotal := uint32(len(s.chunks))
	if newTotal == 0 {
		return ErrInvalid
	}
	s.currentIdx = newTotal - 1
	s.writeOffset = 0
	s.pendingIdx = s.currentIdx
	s.pendingOffset = 0
	s.pendingLen = 0
	return nil
}

func (s *Stream) emitChunkAdded(ch *chunk) {
	wasProducing := s.producing
	s.producing = true
	defer func() { s.producing = wasProducing }()
	s.emitEvent(eventChunkAdded, uintptr(slicePtr(ch.ptr)), uint64(ch.len))
}

func (s *Stream) emitDataAvailable(count uint32) {
	s.emitEvent(eventDataAvail, uintptr(count), 0)
}

func (s *Stream) emitClosed() {
	s.emitEvent(eventClosed, 0, 0)
}

func (s *Stream) emitEvent(event uint32, arg0 uintptr, arg1 uint64) {
	cb := s.callback
	if cb == nil {
		return
	}
	wasInCallback := s.inCallback
	s.inCallback = true
	defer func() { s.inCallback = wasInCallback }()
	cb(uintptr(unsafe.Pointer(s)), event, arg0, arg1)
}

func (s *Stream) copyQueuedPrefix(out []byte) uint32 {
	s.mu.Lock()
	defer s.mu.Unlock()
	index := s.spanRing.head
	copied := uint32(0)
	for index != spanRingNone && uint32(len(out)) > copied {
		entry := &s.spanRing.buffer[index]
		if entry.state != spanStateQueued || entry.span.Len == 0 {
			break
		}
		slice := entry.span.spanBytes(s.chunks)
		take := uint32(len(slice))
		if uint32(len(out))-copied < take {
			take = uint32(len(out)) - copied
		}
		copy(out[copied:], slice[:take])
		copied += take
		index = entry.next
	}
	return copied
}

func (s *Stream) drainSpans(out []SpanInfo) uint32 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.spanRing.popMany(out)
}

// Close closes the stream.
func (s *Stream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.producing {
		return ErrBusy
	}
	if s.closed {
		return nil
	}
	if s.reservedActive || s.stagedBytes != 0 {
		return ErrBusy
	}
	notify := false
	if s.pendingLen > 0 {
		if err := s.commitLocked(&notify); err != nil {
			return err
		}
	}
	s.closed = true
	s.attached = false
	s.finish(notify, 0)
	s.emitClosed()
	return nil
}

// Destroy destroys the stream, panicking if borrows remain.
func (s *Stream) Destroy() {
	if err := s.tryDestroy(); err != nil {
		panic("cannot destroy a stream with outstanding borrows: " + err.Error())
	}
}

func (s *Stream) tryDestroy() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.inCallback || s.producing || s.spanRing.borrowed != 0 || s.reservedActive || s.stagedBytes != 0 {
		return ErrBusy
	}
	if !s.closed {
		if err := s.Close(); err != nil {
			return err
		}
	}
	if s.spanRing.borrowed != 0 {
		return ErrBusy
	}
	for _, ch := range s.chunks {
		if ch != nil {
			s.allocator.Free(ch.ptr)
		}
	}
	s.chunks = nil
	if s.spanRing.capacity > 0 {
		s.spanRing.buffer = nil
	}
	return nil
}

// ReserveInfo describes a zero-copy reservation.
type ReserveInfo struct {
	Ptr      uintptr
	Len      uint32
	Reserved uint32
}

// slicePtr returns an unsafe.Pointer to the first byte of a slice.
func slicePtr(b []byte) unsafe.Pointer {
	if len(b) == 0 {
		return unsafe.Pointer(nil)
	}
	return unsafe.Pointer(&b[0])
}

// spanBytes returns the span's bytes from the chunk pool.
func (s SpanInfo) spanBytes(chunks []*chunk) []byte {
	if int(s.ChunkIndex) >= len(chunks) {
		return nil
	}
	ch := chunks[s.ChunkIndex]
	if ch == nil || int(s.Offset)+int(s.Len) > int(ch.len) {
		return nil
	}
	return ch.ptr[s.Offset : s.Offset+s.Len]
}

var _ = errors.New
