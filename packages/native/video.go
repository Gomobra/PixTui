package native

import (
	"fmt"
	"sync"
	"time"
)

// PlaybackState represents the playback state of a video.
type PlaybackState int

const (
	PlaybackStateStopped   PlaybackState = iota // 0
	PlaybackStatePlaying                        // 1
	PlaybackStatePaused                         // 2
	PlaybackStateEnded                          // 3
	PlaybackStateBuffering                      // 4
	PlaybackStateFailed                         // 5
)

func (s PlaybackState) String() string {
	switch s {
	case PlaybackStateStopped:
		return "stopped"
	case PlaybackStatePlaying:
		return "playing"
	case PlaybackStatePaused:
		return "paused"
	case PlaybackStateEnded:
		return "ended"
	case PlaybackStateBuffering:
		return "buffering"
	case PlaybackStateFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// VideoQualityTier represents adaptive quality tiers for video decoding.
type VideoQualityTier int

const (
	QualityTierFull   VideoQualityTier = 0 // 1.0 scale
	QualityTierHigh   VideoQualityTier = 1 // 0.75 scale
	QualityTierMedium VideoQualityTier = 2 // 0.50 scale
	QualityTierLow    VideoQualityTier = 3 // 0.35 scale
	QualityTierLowest VideoQualityTier = 4 // 0.25 scale
)

var qualityScales = []float64{1.0, 0.75, 0.5, 0.35, 0.25}

// Video represents a terminal-renderable video with FFmpeg backend.
// Mirrors packages/native/src/video.zig.
type Video struct {
	mu       sync.RWMutex
	ID       uint64
	Width    int
	Height   int
	Duration float64
	FPS      float64
	Codec    string
	Format   string
	BitRate  int64

	Frame       *Image
	Playing     bool
	CurrentTime float64
	Loop        bool

	// Playback state
	state PlaybackState

	// AV sync offset (microseconds)
	avSyncOffsetUs int64

	// Adaptive quality
	quality VideoQualityTier

	// Frame pool
	framePool   []*Image
	framePoolMu sync.Mutex

	ffmpeg   *FFmpegWrapper
	path     string
	lastTime float64

	lastError string
}

// NewVideo creates a new Video instance with given dimensions.
func NewVideo(width, height int) *Video {
	return &Video{
		Width:     width,
		Height:    height,
		Frame:     NewImage(width, height),
		ffmpeg:    NewFFmpegWrapper(),
		Loop:      false,
		state:     PlaybackStateStopped,
		quality:   QualityTierFull,
		framePool: make([]*Image, 0, 4),
	}
}

// LoadVideo loads a video file using FFmpeg, probes metadata, and extracts the first frame.
func LoadVideo(path string) (*Video, error) {
	ffmpeg := NewFFmpegWrapper()
	if !ffmpeg.IsAvailable() {
		return nil, fmt.Errorf("ffmpeg not available")
	}

	// Probe video metadata
	info, err := ffmpeg.ProbeVideo(path)
	if err != nil {
		return nil, fmt.Errorf("failed to probe video: %w", err)
	}

	vid := &Video{
		Width:     info.Width,
		Height:    info.Height,
		Duration:  info.Duration,
		FPS:       info.FPS,
		Codec:     info.Codec,
		Format:    info.Format,
		BitRate:   info.BitRate,
		Frame:     NewImage(info.Width, info.Height),
		ffmpeg:    ffmpeg,
		path:      path,
		lastTime:  0,
		state:     PlaybackStateStopped,
		quality:   QualityTierFull,
		framePool: make([]*Image, 0, 4),
	}

	// Extract first frame
	frame, err := ffmpeg.ExtractFrame(path, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to extract first frame: %w", err)
	}
	vid.Frame = frame
	vid.state = PlaybackStatePaused

	return vid, nil
}

// Play starts video playback.
func (v *Video) Play() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.state = PlaybackStatePlaying
	v.Playing = true
	v.lastTime = 0
}

// Pause pauses video playback.
func (v *Video) Pause() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.state = PlaybackStatePaused
	v.Playing = false
}

// Toggle toggles between play and pause.
func (v *Video) Toggle() {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.Playing {
		v.state = PlaybackStatePaused
		v.Playing = false
	} else {
		v.state = PlaybackStatePlaying
		v.Playing = true
		v.lastTime = 0
	}
}

// Seek seeks to a specific time in the video.
func (v *Video) Seek(seconds float64) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.path == "" {
		return fmt.Errorf("no video loaded")
	}
	if v.ffmpeg == nil {
		v.ffmpeg = NewFFmpegWrapper()
	}
	frame, err := v.ffmpeg.ExtractFrame(v.path, seconds)
	if err != nil {
		v.state = PlaybackStateFailed
		v.lastError = err.Error()
		return err
	}
	v.Frame = frame
	v.CurrentTime = seconds
	v.state = PlaybackStatePaused
	v.lastError = ""
	return nil
}

// Update updates the video state based on elapsed time. Should be called each frame.
func (v *Video) Update(deltaTime float64) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if !v.Playing || v.path == "" {
		return nil
	}

	v.state = PlaybackStatePlaying
	v.lastTime += deltaTime
	v.CurrentTime += deltaTime

	// If we've reached the end
	if v.CurrentTime >= v.Duration {
		if v.Loop {
			v.CurrentTime = 0
			v.lastTime = 0
		} else {
			v.Playing = false
			v.state = PlaybackStateEnded
			return nil
		}
	}

	// Extract frame at current time (throttle by FPS)
	frameInterval := 1.0 / v.FPS
	if v.FPS <= 0 {
		frameInterval = 1.0 / 30.0
	}

	if v.lastTime >= frameInterval {
		frame, err := v.ffmpeg.ExtractFrame(v.path, v.CurrentTime)
		if err != nil {
			v.state = PlaybackStateFailed
			v.lastError = err.Error()
			return err
		}
		v.Frame = frame
		v.lastTime = 0
		v.lastError = ""
	}

	return nil
}

// GetFrame returns the current video frame.
func (v *Video) GetFrame() *Image {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.Frame
}

// IsPlaying returns whether the video is playing.
func (v *Video) IsPlaying() bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.Playing
}

// GetDuration returns the video duration in seconds.
func (v *Video) GetDuration() float64 {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.Duration
}

// GetFPS returns the video FPS.
func (v *Video) GetFPS() float64 {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.FPS
}

// GetState returns the current playback state.
func (v *Video) GetState() PlaybackState {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.state
}

// GetAVSyncOffset returns the audio-video sync offset in microseconds.
func (v *Video) GetAVSyncOffset() int64 {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.avSyncOffsetUs
}

// SetAVSyncOffset sets the audio-video sync offset in microseconds.
func (v *Video) SetAVSyncOffset(offsetUs int64) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	const maxOffset = 60_000_000 // 60 seconds in microseconds
	if offsetUs < -maxOffset || offsetUs > maxOffset {
		return fmt.Errorf("av sync offset out of range: %d", offsetUs)
	}
	v.avSyncOffsetUs = offsetUs
	return nil
}

// GetQualityTier returns the current adaptive quality tier.
func (v *Video) GetQualityTier() VideoQualityTier {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.quality
}

// SetQualityTier sets the adaptive quality tier.
func (v *Video) SetQualityTier(tier VideoQualityTier) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if tier < 0 {
		tier = 0
	}
	if tier > QualityTierLowest {
		tier = QualityTierLowest
	}
	v.quality = tier
}

// GetQualityScale returns the current quality scale factor.
func (v *Video) GetQualityScale() float64 {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return qualityScales[v.quality]
}

// GetLastError returns the last error message, or empty string if none.
func (v *Video) GetLastError() string {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.lastError
}

// GetFramePool returns the frame pool for video frame reuse.
func (v *Video) GetFramePool() []*Image {
	v.framePoolMu.Lock()
	defer v.framePoolMu.Unlock()
	return v.framePool
}

// AcquireFrameFromPool gets a frame from the pool or creates a new one.
func (v *Video) AcquireFrameFromPool() *Image {
	v.framePoolMu.Lock()
	defer v.framePoolMu.Unlock()
	if len(v.framePool) > 0 {
		frame := v.framePool[len(v.framePool)-1]
		v.framePool = v.framePool[:len(v.framePool)-1]
		return frame
	}
	return NewImage(v.Width, v.Height)
}

// ReleaseFrameToPool returns a frame to the pool for reuse.
func (v *Video) ReleaseFrameToPool(frame *Image) {
	v.framePoolMu.Lock()
	defer v.framePoolMu.Unlock()
	if len(v.framePool) < 4 {
		v.framePool = append(v.framePool, frame)
	}
}

// LoadAudio loads an audio file using FFmpeg and returns PCM samples.
func LoadAudio(path string) ([]float32, int, error) {
	ffmpeg := NewFFmpegWrapper()
	if !ffmpeg.IsAvailable() {
		return nil, 0, fmt.Errorf("ffmpeg not available")
	}
	return ffmpeg.ExtractAudio(path)
}

// WaitForFrame waits until the next frame interval and returns the frame.
func (v *Video) WaitForFrame(deltaTime float64) (*Image, error) {
	v.mu.RLock()
	playing := v.Playing
	path := v.path
	currentTime := v.CurrentTime
	duration := v.Duration
	loop := v.Loop
	frame := v.Frame
	v.mu.RUnlock()

	if !playing || path == "" {
		return frame, nil
	}

	// Simulate waiting for next frame
	time.Sleep(time.Duration(deltaTime * 1e9))

	newTime := currentTime + deltaTime
	if newTime >= duration {
		if loop {
			newTime = 0
		} else {
			return frame, nil
		}
	}

	// Extract frame at new time
	v.mu.Lock()
	v.CurrentTime = newTime
	v.lastTime = 0
	extracted, err := v.ffmpeg.ExtractFrame(path, newTime)
	if err != nil {
		v.state = PlaybackStateFailed
		v.lastError = err.Error()
		v.mu.Unlock()
		return frame, err
	}
	v.Frame = extracted
	frame = extracted
	v.mu.Unlock()

	return frame, nil
}
