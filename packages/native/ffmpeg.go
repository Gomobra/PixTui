package native

import (
	"bytes"
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

// FFmpegWrapper wraps FFmpeg CLI for video/audio processing.
type FFmpegWrapper struct {
	mu      sync.Mutex
	binary  string
	version string
}

// NewFFmpegWrapper creates a new FFmpeg wrapper.
func NewFFmpegWrapper() *FFmpegWrapper {
	return &FFmpegWrapper{
		binary: "ffmpeg",
	}
}

// IsAvailable checks if FFmpeg is available.
func (f *FFmpegWrapper) IsAvailable() bool {
	cmd := exec.Command(f.binary, "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "ffmpeg version")
}

// GetVersion returns the FFmpeg version string.
func (f *FFmpegWrapper) GetVersion() string {
	if f.version != "" {
		return f.version
	}
	cmd := exec.Command(f.binary, "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		f.version = strings.TrimSpace(lines[0])
	}
	return f.version
}

// ExtractFrame extracts a single frame from a video file at the given timestamp.
// Returns the frame as an *Image.
func (f *FFmpegWrapper) ExtractFrame(videoPath string, timestamp float64) (*Image, error) {
	// Use ffmpeg to extract a single frame as rawvideo RGBA
	cmd := exec.Command(
		f.binary,
		"-ss", strconv.FormatFloat(timestamp, 'f', 6, 64),
		"-i", videoPath,
		"-frames:v", "1",
		"-f", "rawvideo",
		"-pix_fmt", "rgba",
		"-",
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg extract frame failed: %w\nstderr: %s", err, stderr.String())
	}

	data := stdout.Bytes()
	if len(data) == 0 {
		return nil, fmt.Errorf("no frame data returned")
	}

	// We need to know the dimensions. First, probe the video to get dimensions.
	width, height, err := f.probeDimensions(videoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to probe dimensions: %w", err)
	}

	expectedSize := width * height * 4 // RGBA = 4 bytes per pixel
	if len(data) < expectedSize {
		return nil, fmt.Errorf("insufficient frame data: got %d bytes, expected %d", len(data), expectedSize)
	}

	// Take only the first frame
	frameData := data[:expectedSize]

	img := NewImage(width, height)
	copy(img.Pixels, frameData)
	return img, nil
}

// ExtractFrames extracts multiple frames from a video at given timestamps.
func (f *FFmpegWrapper) ExtractFrames(videoPath string, timestamps []float64) ([]*Image, error) {
	frames := make([]*Image, 0, len(timestamps))
	for _, ts := range timestamps {
		frame, err := f.ExtractFrame(videoPath, ts)
		if err != nil {
			return nil, fmt.Errorf("failed to extract frame at %v: %w", ts, err)
		}
		frames = append(frames, frame)
	}
	return frames, nil
}

// probeDimensions probes video dimensions using ffprobe.
func (f *FFmpegWrapper) probeDimensions(videoPath string) (int, int, error) {
	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-of", "csv=p=0",
		videoPath,
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return 0, 0, fmt.Errorf("ffprobe failed: %w\nstderr: %s", err, stderr.String())
	}

	output := strings.TrimSpace(stdout.String())
	parts := strings.Split(output, ",")
	if len(parts) < 2 {
		return 0, 0, fmt.Errorf("unexpected ffprobe output: %s", output)
	}

	width, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid width: %s", parts[0])
	}
	height, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid height: %s", parts[1])
	}

	return width, height, nil
}

// ExtractAudio extracts audio from a video file as raw PCM samples.
// Returns float32 stereo samples.
func (f *FFmpegWrapper) ExtractAudio(videoPath string) ([]float32, int, error) {
	if !f.IsAvailable() {
		return nil, 0, fmt.Errorf("ffmpeg not available")
	}

	cmd := exec.Command(
		f.binary,
		"-i", videoPath,
		"-vn",
		"-f", "f32le",
		"-acodec", "pcm_f32le",
		"-ar", "48000",
		"-ac", "2",
		"-",
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, 0, fmt.Errorf("ffmpeg extract audio failed: %w\nstderr: %s", err, stderr.String())
	}

	data := stdout.Bytes()
	sampleCount := len(data) / 4 // 4 bytes per float32
	samples := make([]float32, sampleCount)

	for i := 0; i < sampleCount; i++ {
		// Read little-endian float32
		idx := i * 4
		bits := uint32(data[idx]) | uint32(data[idx+1])<<8 | uint32(data[idx+2])<<16 | uint32(data[idx+3])<<24
		samples[i] = uint32ToFloat32(bits)
	}

	return samples, 48000, nil
}

// uint32ToFloat32 converts uint32 bits to float32.
func uint32ToFloat32(b uint32) float32 {
	return math.Float32frombits(b)
}

// VideoInfo represents video metadata.
type VideoInfo struct {
	Width    int
	Height   int
	Duration float64
	FPS      float64
	Codec    string
	Format   string
	BitRate  int64
}

// ProbeVideo probes video metadata using ffprobe.
func (f *FFmpegWrapper) ProbeVideo(videoPath string) (*VideoInfo, error) {
	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-show_entries", "stream=codec_name,width,height,r_frame_rate,bit_rate",
		"-show_entries", "format=duration,format_name",
		"-of", "csv=p=0",
		videoPath,
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffprobe failed: %w\nstderr: %s", err, stderr.String())
	}

	output := strings.TrimSpace(stdout.String())
	lines := strings.Split(output, "\n")

	info := &VideoInfo{}

	for _, line := range lines {
		parts := strings.Split(line, ",")
		if len(parts) < 1 {
			continue
		}

		// First line is typically format info: format_name,duration
		if len(parts) >= 2 && info.Format == "" {
			info.Format = strings.TrimSpace(parts[0])
			if d, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64); err == nil {
				info.Duration = d
			}
		}

		// Stream lines: codec_name,width,height,r_frame_rate,bit_rate
		if len(parts) >= 5 {
			info.Codec = strings.TrimSpace(parts[0])
			if w, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
				info.Width = w
			}
			if h, err := strconv.Atoi(strings.TrimSpace(parts[2])); err == nil {
				info.Height = h
			}
			// Parse frame rate (e.g., "30/1" or "30000/1001")
			if fpsStr := strings.TrimSpace(parts[3]); fpsStr != "" {
				if fps, err := parseFrameRate(fpsStr); err == nil {
					info.FPS = fps
				}
			}
			if br, err := strconv.ParseInt(strings.TrimSpace(parts[4]), 10, 64); err == nil {
				info.BitRate = br
			}
		}
	}

	return info, nil
}

// parseFrameRate parses a frame rate string like "30/1" or "30000/1001".
func parseFrameRate(fpsStr string) (float64, error) {
	parts := strings.Split(fpsStr, "/")
	if len(parts) == 1 {
		return strconv.ParseFloat(parts[0], 64)
	}
	if len(parts) == 2 {
		num, err1 := strconv.ParseFloat(parts[0], 64)
		den, err2 := strconv.ParseFloat(parts[1], 64)
		if err1 != nil || err2 != nil || den == 0 {
			return 0, fmt.Errorf("invalid frame rate: %s", fpsStr)
		}
		return num / den, nil
	}
	return 0, fmt.Errorf("invalid frame rate: %s", fpsStr)
}
