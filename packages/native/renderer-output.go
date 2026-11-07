package native

import (
	"errors"
	"io"
	"math"
	"os"
	"runtime"
	"time"
	"unicode/utf8"
)

// WriteStatus classifies the outcome of a frame or control write.
// Mirrors packages/native/src/renderer-output.zig WriteStatus.
type WriteStatus int

const (
	// WriteStatusOK means the write was accepted and published.
	WriteStatusOK WriteStatus = iota
	// WriteStatusSkipped means the write was rejected due to transient output
	// pressure; no bytes were published and the caller should retry after drain.
	WriteStatusSkipped
	// WriteStatusFailed means the write failed permanently; no bytes published.
	WriteStatusFailed
)

// ErrBufferFull is returned when a staged frame write cannot be admitted.
var ErrBufferFull = errors.New("buffer full")

const utf16BufferSize = 4096

// utf8ToUtf16Chunk converts a prefix of input (UTF-8) into output (UTF-16),
// never splitting a surrogate pair. It returns the number of input bytes
// consumed and UTF-16 units written. An invalid or incomplete UTF-8 sequence
// returns an error.
func utf8ToUtf16Chunk(output []uint16, input []byte) (inputLen, outputLen int, err error) {
	if len(output) < 2 {
		return 0, 0, errors.New("utf16 output buffer too small")
	}
	in := 0
	out := 0
	for in < len(input) {
		b := input[in]
		if b < 0x80 {
			if out == len(output) {
				break
			}
			output[out] = uint16(b)
			in++
			out++
			continue
		}
		r, size := utf8.DecodeRune(input[in:])
		if r == utf8.RuneError && size == 0 {
			return 0, 0, errors.New("invalid utf8")
		}
		if len(input)-in < size {
			return 0, 0, errors.New("incomplete utf8")
		}
		needed := 1
		if r > 0xFFFF {
			needed = 2
		}
		if len(output)-out < needed {
			break
		}
		if needed == 1 {
			output[out] = uint16(r)
		} else {
			v := r - 0x10000
			output[out] = uint16(0xD800 + (v >> 10))
			output[out+1] = uint16(0xDC00 + (v & 0x3FF))
		}
		in += size
		out += needed
	}
	return in, out, nil
}

// completeUtf8Prefix returns the longest prefix of data that ends on a rune
// boundary, so a multibyte UTF-8 sequence is never split at the tail.
func completeUtf8Prefix(data []byte) []byte {
	if len(data) == 0 {
		return data
	}
	start := len(data) - 1
	for start > 0 && len(data)-start < 4 && data[start]&0xC0 == 0x80 {
		start--
	}
	_, size := utf8.DecodeRune(data[start:])
	if size == 0 {
		// Invalid leading byte: keep everything (mirrors Zig's catch-return-data).
		return data
	}
	if size > len(data)-start {
		// Incomplete sequence at the tail: cut it off.
		return data[:start]
	}
	return data
}

// StdoutOutput writes drained output to an io.Writer. On Windows consoles it
// transcodes UTF-8 to UTF-16 for WriteConsoleW. On non-Windows platforms it
// writes bytes verbatim.
type StdoutOutput struct {
	writer         io.Writer
	utf16Buffer    [utf16BufferSize]uint16
	windowsConsole bool
}

// NewStdoutOutput creates a StdoutOutput writing to the given writer.
func NewStdoutOutput(w io.Writer) *StdoutOutput {
	return &StdoutOutput{writer: w}
}

// NewStdoutOutputForFile creates a StdoutOutput writing to os.Stdout.
func NewStdoutOutputForFile() *StdoutOutput {
	return NewStdoutOutput(os.Stdout)
}

// WriteSome writes one direct byte write, or one bounded UTF-16 console chunk.
// An incomplete UTF-8 suffix stays with the caller.
func (o *StdoutOutput) WriteSome(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}
	if runtime.GOOS == "windows" && o.windowsConsole {
		prefix := completeUtf8Prefix(data[:minInt(len(data), utf16BufferSize)])
		if len(prefix) == 0 {
			return 0, nil
		}
		chunkIn, chunkOut, err := utf8ToUtf16Chunk(o.utf16Buffer[:], prefix)
		if err != nil {
			return 0, err
		}
		if !o.writeWindowsConsoleUtf16(o.utf16Buffer[:chunkOut]) {
			return 0, errors.New("write failed")
		}
		return chunkIn, nil
	}
	n, err := o.writer.Write(data)
	if err != nil {
		return 0, err
	}
	if n == 0 {
		return 0, errors.New("write failed")
	}
	return n, nil
}

// SupportsDirectWrite reports whether the writer can be written to directly.
func (o *StdoutOutput) SupportsDirectWrite() bool {
	if runtime.GOOS != "windows" || o.windowsConsole {
		return true
	}
	return false
}

func (o *StdoutOutput) writeWindowsConsoleUtf16(data []uint16) bool {
	// Windows-only path; on this platform WriteConsoleW is unavailable.
	_ = data
	return false
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// UploadState tracks one image upload staged within a frame.
type UploadState struct {
	start       int
	end         int
	waited      bool
	len         int
	cut         bool
	placementID uint32
	state       int
}

const (
	uploadStateKept = iota
	uploadStateDeferred
	uploadStateTooLarge
)

// FeedBackend atomically publishes complete frames to a NativeSpanFeed.Stream.
// The feed owns its chunk memory; the staging buffer keeps failed frames from
// exposing partial ANSI sequences. Mirrors packages/native/src/renderer-output.zig
// FeedBackend.
type FeedBackend struct {
	feed            *Stream
	frameBytes      []byte
	frameActive     bool
	frameStatus     WriteStatus
	oversizeBytes   uint64
	uploads         []UploadState
	upload          *UploadState
	keptUploadBytes int
	uploadsAlone    bool
	lastWriteTimeUs float64
	frameWriteCount int
}

// CreateFeedBackend creates a FeedBackend publishing to the given stream.
func CreateFeedBackend(feed *Stream) *FeedBackend {
	return &FeedBackend{feed: feed}
}

// ShouldSkipFrame reports whether a new frame should be skipped.
func (b *FeedBackend) ShouldSkipFrame() bool {
	stats := b.feed.GetStats()
	cap := b.feed.options.SpanQueueCapacity
	return b.frameActive || (cap > 0 && stats.OutstandingSpans >= cap) || !b.feed.HasAtomicCapacity()
}

// PrepareFrame begins a frame. Returns WriteStatusSkipped when the frame cannot
// be admitted due to transient output pressure.
func (b *FeedBackend) PrepareFrame() WriteStatus {
	if b.frameActive {
		return WriteStatusSkipped
	}
	b.frameStatus = WriteStatusOK

	if b.feed.hasPendingBytes() {
		if err := b.feed.Commit(); err != nil {
			return WriteStatusSkipped
		}
		// Pending bytes belonged to an earlier frame/control write. Queue them
		// first and let the caller retry the new frame after drain.
		return WriteStatusSkipped
	}

	if b.ShouldSkipFrame() {
		return WriteStatusSkipped
	}
	return WriteStatusOK
}

// BeginFrame starts a new frame staging session.
func (b *FeedBackend) BeginFrame() {
	b.frameWriteCount = 0
	b.frameStatus = WriteStatusOK
	b.frameActive = true
	b.oversizeBytes = 0
	b.uploads = b.uploads[:0]
	b.upload = nil
	b.keptUploadBytes = 0
	b.uploadsAlone = false
	_ = b.feed.SetStagedBytes(0)
	b.frameBytes = b.frameBytes[:0]
}

// FailFrame releases the frame's staging. An earlier status wins.
func (b *FeedBackend) FailFrame() {
	if b.frameStatus == WriteStatusOK {
		b.frameStatus = WriteStatusFailed
	}
	b.keptUploadBytes = 0
	_ = b.feed.SetStagedBytes(0)
	b.frameBytes = b.frameBytes[:0]
}

// UploadLimit returns the byte headroom usable for a single upload.
func (b *FeedBackend) UploadLimit() uint64 {
	if b.feed.bounded() {
		return b.feed.atomicByteLimit()
	}
	return math.MaxUint64
}

// BeginUpload opens an upload at the end of the staged frame.
func (b *FeedBackend) BeginUpload(waited bool) {
	b.upload = &UploadState{
		start:       len(b.frameBytes),
		waited:      waited,
		placementID: 0,
	}
}

// EndUpload closes the open upload, classifying it as kept/deferred/too_large.
func (b *FeedBackend) EndUpload(placementID uint32) {
	if b.upload == nil {
		return
	}
	upload := b.upload
	b.upload = nil
	upload.placementID = placementID
	if b.frameStatus != WriteStatusOK {
		return
	}
	state := uploadStateKept
	if uint64(upload.len) > b.UploadLimit() {
		state = uploadStateTooLarge
	} else if upload.cut {
		state = uploadStateDeferred
	}
	b.uploads = append(b.uploads, UploadState{
		start:       upload.start,
		end:         len(b.frameBytes),
		placementID: placementID,
		waited:      upload.waited,
		state:       state,
		len:         upload.len,
	})
	if state == uploadStateKept {
		b.keptUploadBytes += upload.len
	}
}

// DeferUpload records an upload that waits for a later frame.
func (b *FeedBackend) DeferUpload(placementID uint32) {
	if b.frameStatus != WriteStatusOK {
		return
	}
	end := len(b.frameBytes)
	b.uploads = append(b.uploads, UploadState{
		start:       end,
		end:         end,
		placementID: placementID,
		waited:      false,
		state:       uploadStateDeferred,
	})
}

// LastUploadWaits reports whether the last upload did not complete.
func (b *FeedBackend) LastUploadWaits() bool {
	items := b.uploads
	return len(items) != 0 && items[len(items)-1].state != uploadStateKept
}

// trimUploads removes kept uploads, last first, until the frame fits an empty
// queue. An upload that already waited and does not fit is too large.
func (b *FeedBackend) trimUploads() {
	others := len(b.frameBytes) - b.keptUploadBytes
	for len(b.frameBytes) > int(b.UploadLimit()) && len(b.uploads) > 0 {
		index := len(b.uploads) - 1
		upload := &b.uploads[index]
		if upload.state != uploadStateKept {
			b.uploads = b.uploads[:index]
			continue
		}
		tail := b.frameBytes[upload.end:]
		copy(b.frameBytes[upload.start:], tail)
		b.frameBytes = b.frameBytes[:upload.start+len(tail)]
		alone := upload.waited && b.uploadsAlone
		if alone && uint64(others+upload.end-upload.start) > b.UploadLimit() {
			upload.state = uploadStateTooLarge
		} else {
			upload.state = uploadStateDeferred
		}
		b.uploads = b.uploads[:index]
	}
}

// Writer returns the feed backend as an io.Writer that stages bytes into the
// current frame.
func (b *FeedBackend) Writer() io.Writer {
	return b
}

// Write stages data into the current frame. It is safe to call only between
// BeginFrame and EndFrame.
func (b *FeedBackend) Write(data []byte) (int, error) {
	if b.upload != nil {
		u := b.upload
		u.len += len(data)
		if u.cut || b.frameStatus != WriteStatusOK {
			return len(data), nil
		}
		if uint64(b.keptUploadBytes)+uint64(u.len) <= b.UploadLimit() {
			if err := b.stage(data); err != nil {
				return 0, err
			}
		} else {
			b.frameBytes = b.frameBytes[:u.start]
			u.cut = true
		}
		return len(data), nil
	}
	if b.frameStatus != WriteStatusOK {
		if b.oversizeBytes == 0 {
			return 0, ErrBufferFull
		}
		b.oversizeBytes += uint64(len(data))
		return len(data), nil
	}
	required := uint64(len(b.frameBytes)-b.keptUploadBytes) + uint64(len(data))
	if required > math.MaxUint64 {
		b.FailFrame()
		return 0, ErrBufferFull
	}
	if err := b.feed.SetStagedBytes(required); err != nil {
		b.frameStatus = b.admissionStatus(len(b.frameBytes), err)
		b.FailFrame()
		if b.oversizeBytes == 0 {
			return 0, ErrBufferFull
		}
		return len(data), nil
	}
	if b.frameStatus != WriteStatusOK {
		return 0, ErrBufferFull
	}
	if err := b.stage(data); err != nil {
		return 0, err
	}
	return len(data), nil
}

func (b *FeedBackend) stage(data []byte) error {
	required := len(b.frameBytes) + len(data)
	maxBytes := b.feed.byteLimit()
	if maxBytes != 0 && uint64(required) > uint64(cap(b.frameBytes)) {
		doubled := cap(b.frameBytes) * 2
		if doubled < 256 {
			doubled = 256
		}
		if uint64(doubled) > maxBytes {
			doubled = int(maxBytes)
		}
		if uint64(required) > uint64(doubled) {
			doubled = required
		}
		newBuf := make([]byte, len(b.frameBytes), doubled)
		copy(newBuf, b.frameBytes)
		b.frameBytes = newBuf
	}
	if maxBytes != 0 && uint64(len(b.frameBytes)+len(data)) > maxBytes {
		return ErrBufferFull
	}
	b.frameBytes = append(b.frameBytes, data...)
	return nil
}

func (b *FeedBackend) admissionStatus(frameBytes int, err error) WriteStatus {
	if err != ErrNoSpace && err != ErrMaxBytes {
		return WriteStatusFailed
	}
	if !b.feed.bounded() {
		return WriteStatusFailed
	}
	if uint64(frameBytes) <= b.feed.atomicByteLimit() {
		return WriteStatusSkipped
	}
	b.oversizeBytes = uint64(frameBytes)
	return WriteStatusFailed
}

// EndFrame releases the staged frame and publishes it atomically to the feed.
func (b *FeedBackend) EndFrame() WriteStatus {
	writeStart := time.Now()
	status := b.frameStatus
	_ = b.feed.SetStagedBytes(0)

	if status == WriteStatusOK {
		b.trimUploads()
		notify, err := b.feed.publishAtomic(b.frameBytes, admissionOrdinary)
		_ = notify
		if err != nil {
			status = b.admissionStatus(len(b.frameBytes), err)
		}
	}

	b.frameActive = false
	b.frameBytes = b.frameBytes[:0]
	elapsed := time.Since(writeStart)
	b.lastWriteTimeUs = float64(elapsed.Microseconds())
	b.feed.finish(false, 0)
	return status
}

// WriteOut performs a control write between frames. A rejected write publishes
// nothing.
func (b *FeedBackend) WriteOut(data []byte) {
	_ = b.feed.WriteAtomic(data)
}
