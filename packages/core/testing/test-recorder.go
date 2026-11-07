package testing

import (
	"fmt"
	"strings"
	"time"

	"github.com/Gomobra/PixTui/packages/core"
	"github.com/Gomobra/PixTui/packages/native"
)

// RecordBuffersOptions configures which cell buffer arrays to record alongside each frame.
type RecordBuffersOptions struct {
	FG         bool
	BG         bool
	Attributes bool
}

// RecordedBuffers holds a snapshot of cell buffer arrays for a recorded frame.
type RecordedBuffers struct {
	FG         []uint16
	BG         []uint16
	Attributes []uint32
}

// RecordedFrame is a single recorded frame with metadata.
type RecordedFrame struct {
	Frame       string
	Timestamp   time.Duration
	FrameNumber int
	Buffers     *RecordedBuffers
}

// TestRecorderOptions configures a TestRecorder.
type TestRecorderOptions struct {
	RecordBuffers RecordBuffersOptions
	Now           func() time.Duration
}

// TestRecorder captures frames rendered by a CliRenderer.
// Each call to RecordFrame snapshots the current buffer into the recorded frames list.
// It can also be started/stopped to automatically capture frames on renderer frame events.
type TestRecorder struct {
	renderer    *core.CliRenderer
	frames      []RecordedFrame
	recording   bool
	frameNumber int
	startTime   time.Duration
	recordBuf   RecordBuffersOptions
	now         func() time.Duration
}

// NewTestRecorder creates a new TestRecorder for the given renderer.
func NewTestRecorder(renderer *core.CliRenderer) *TestRecorder {
	return &TestRecorder{
		renderer: renderer,
		frames:   make([]RecordedFrame, 0),
		now:      func() time.Duration { return time.Since(time.Time{}) },
	}
}

// NewTestRecorderWithOptions creates a new TestRecorder with the given options.
func NewTestRecorderWithOptions(renderer *core.CliRenderer, options TestRecorderOptions) *TestRecorder {
	now := options.Now
	if now == nil {
		now = func() time.Duration { return time.Since(time.Time{}) }
	}
	return &TestRecorder{
		renderer:  renderer,
		frames:    make([]RecordedFrame, 0),
		recordBuf: options.RecordBuffers,
		now:       now,
	}
}

// Rec starts recording frames. If already recording, this is a no-op.
func (r *TestRecorder) Rec() {
	if r.recording {
		return
	}
	r.recording = true
	r.frames = make([]RecordedFrame, 0)
	r.frameNumber = 0
	r.startTime = r.now()
}

// Stop stops recording frames. If not recording, this is a no-op.
func (r *TestRecorder) Stop() {
	if !r.recording {
		return
	}
	r.recording = false
}

// RecordFrame captures the current frame from the renderer and appends it to the recording.
func (r *TestRecorder) RecordFrame() {
	r.captureFrame()
}

// Frames returns all recorded frames as a slice of RecordedFrame snapshots.
func (r *TestRecorder) Frames() []RecordedFrame {
	result := make([]RecordedFrame, len(r.frames))
	copy(result, r.frames)
	return result
}

// FrameCount returns the number of frames currently recorded.
func (r *TestRecorder) FrameCount() int {
	return len(r.frames)
}

// RecordedFrames is an alias for Frames.
func (r *TestRecorder) RecordedFrames() []RecordedFrame {
	return r.Frames()
}

// IsRecording returns true if the recorder is currently recording.
func (r *TestRecorder) IsRecording() bool {
	return r.recording
}

// Clear removes all recorded frames and resets the frame counter.
func (r *TestRecorder) Clear() {
	r.frames = make([]RecordedFrame, 0)
	r.frameNumber = 0
}

// AssertFrameContains checks that the frame at frameIndex contains the given text.
func (r *TestRecorder) AssertFrameContains(frameIndex int, text string) error {
	if frameIndex < 0 || frameIndex >= len(r.frames) {
		return fmt.Errorf("frame index %d out of range (0-%d)", frameIndex, len(r.frames)-1)
	}
	frame := r.frames[frameIndex].Frame
	if !strings.Contains(frame, text) {
		return fmt.Errorf("frame %d does not contain %q", frameIndex, text)
	}
	return nil
}

// AssertFrameEquals checks that the frame at frameIndex exactly equals the expected text.
func (r *TestRecorder) AssertFrameEquals(frameIndex int, expected string) error {
	if frameIndex < 0 || frameIndex >= len(r.frames) {
		return fmt.Errorf("frame index %d out of range (0-%d)", frameIndex, len(r.frames)-1)
	}
	frame := r.frames[frameIndex].Frame
	if frame != expected {
		return fmt.Errorf("frame %d does not match expected\nexpected:\n%q\ngot:\n%q", frameIndex, expected, frame)
	}
	return nil
}

// captureFrame captures the current frame from the renderer's buffer.
func (r *TestRecorder) captureFrame() {
	access := r.renderer.Capture()
	frame := bufferAccessToString(access)

	recordedFrame := RecordedFrame{
		Frame:       frame,
		Timestamp:   r.now() - r.startTime,
		FrameNumber: r.frameNumber,
	}

	if r.recordBuf.FG || r.recordBuf.BG || r.recordBuf.Attributes {
		buffers := &RecordedBuffers{}
		if r.recordBuf.FG {
			buffers.FG = append([]uint16{}, access.Fg...)
		}
		if r.recordBuf.BG {
			buffers.BG = append([]uint16{}, access.Bg...)
		}
		if r.recordBuf.Attributes {
			buffers.Attributes = append([]uint32{}, access.Attributes...)
		}
		recordedFrame.Buffers = buffers
	}

	r.frameNumber++
	r.frames = append(r.frames, recordedFrame)
}

// bufferAccessToString converts a BufferAccess snapshot into a string
// by decoding cell characters row by row.
func bufferAccessToString(access native.BufferAccess) string {
	var sb strings.Builder
	for y := 0; y < access.Height; y++ {
		for x := 0; x < access.Width; x++ {
			idx := y*access.Width + x
			ch := access.Char[idx]
			if ch == 0 {
				sb.WriteRune(' ')
			} else {
				sb.WriteRune(rune(ch))
			}
		}
		sb.WriteRune('\n')
	}
	return sb.String()
}
