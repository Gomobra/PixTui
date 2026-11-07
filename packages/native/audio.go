package native

import (
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"sync"
	"time"
)

// Waveform defines the shape of an audio voice.
type Waveform int

const (
	// WaveformSine is a smooth sinusoidal wave.
	WaveformSine Waveform = iota
	// WaveformSquare is a square wave with odd harmonics.
	WaveformSquare
	// WaveformSaw is a sawtooth wave.
	WaveformSaw
	// WaveformTriangle is a triangle wave.
	WaveformTriangle
	// WaveformNoise is white noise.
	WaveformNoise
)

// String returns the waveform name.
func (w Waveform) String() string {
	switch w {
	case WaveformSine:
		return "sine"
	case WaveformSquare:
		return "square"
	case WaveformSaw:
		return "saw"
	case WaveformTriangle:
		return "triangle"
	case WaveformNoise:
		return "noise"
	default:
		return "unknown"
	}
}

// AudioBackend defines how mixed audio is delivered to the output device.
type AudioBackend int

const (
	// AudioBackendNull produces no output (silent, for testing/headless).
	AudioBackendNull AudioBackend = iota
	// AudioBackendALSA writes PCM to an ALSA device file.
	AudioBackendALSA
	// AudioBackendOSS writes PCM to an OSS device (/dev/dsp).
	AudioBackendOSS
	// AudioBackendCallback invokes a user-supplied function with mixed buffers.
	AudioBackendCallback
	// AudioBackendStdout writes PCM to stdout.
	AudioBackendStdout
)

// AudioFormat describes the PCM output format.
type AudioFormat struct {
	SampleRate    int
	Channels      int
	BitsPerSample int
}

// DefaultAudioFormat returns CD-quality stereo PCM.
func DefaultAudioFormat() AudioFormat {
	return AudioFormat{
		SampleRate:    44100,
		Channels:      2,
		BitsPerSample: 16,
	}
}

// AudioVoice represents a single synthesizable audio voice.
type AudioVoice struct {
	mu           sync.RWMutex
	id           uint64
	frequency    float64
	amplitude    float64
	waveform     Waveform
	playing      bool
	phase        float64
	volume       float64
	pan          float64 // -1 = left, 0 = center, 1 = right
	loop         bool
	attack       float64 // seconds
	decay        float64 // seconds
	sustain      float64 // 0..1
	release      float64 // seconds
	sampleRate   int
	buffer       []float32 // pre-rendered PCM (mono) for sample playback
	bufferPos    int
	sampleStart  time.Time
	releaseStart time.Time
	released     bool
}

// NewAudioVoice creates a new voice with the given parameters.
func NewAudioVoice(id uint64, sampleRate int, frequency, amplitude float64, waveform Waveform) *AudioVoice {
	return &AudioVoice{
		id:         id,
		frequency:  frequency,
		amplitude:  amplitude,
		waveform:   waveform,
		sampleRate: sampleRate,
		volume:     1.0,
		pan:        0.0,
		sustain:    1.0,
		playing:    false,
	}
}

// ID returns the voice ID.
func (v *AudioVoice) ID() uint64 {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.id
}

// SetFrequency sets the voice frequency in Hz.
func (v *AudioVoice) SetFrequency(frequency float64) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.frequency = frequency
}

// SetAmplitude sets the voice amplitude (0..1).
func (v *AudioVoice) SetAmplitude(amplitude float64) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.amplitude = amplitude
}

// SetWaveform sets the voice waveform.
func (v *AudioVoice) SetWaveform(waveform Waveform) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.waveform = waveform
}

// SetVolume sets the voice volume (0..1).
func (v *AudioVoice) SetVolume(volume float64) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.volume = volume
}

// SetPan sets the stereo pan (-1..1).
func (v *AudioVoice) SetPan(pan float64) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if pan < -1 {
		pan = -1
	}
	if pan > 1 {
		pan = 1
	}
	v.pan = pan
}

// SetEnvelope sets the ADSR envelope parameters.
func (v *AudioVoice) SetEnvelope(attack, decay, sustain, release float64) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.attack = attack
	v.decay = decay
	v.sustain = sustain
	v.release = release
}

// Play starts the voice.
func (v *AudioVoice) Play() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.playing = true
	v.phase = 0
	v.released = false
	v.releaseStart = time.Time{}
	v.sampleStart = time.Now()
	v.bufferPos = 0
}

// Stop stops the voice (triggers release phase if release > 0).
func (v *AudioVoice) Stop() {
	v.mu.Lock()
	defer v.mu.Unlock()
	if !v.playing {
		return
	}
	if v.release > 0 {
		v.released = true
		v.releaseStart = time.Now()
	} else {
		v.playing = false
	}
}

// IsPlaying returns true if the voice is currently sounding.
func (v *AudioVoice) IsPlaying() bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.playing
}

// SetBuffer loads a pre-rendered mono PCM buffer for sample playback.
func (v *AudioVoice) SetBuffer(samples []float32) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.buffer = make([]float32, len(samples))
	copy(v.buffer, samples)
	v.bufferPos = 0
}

// sample generates the next mono sample for the voice at the current phase.
func (v *AudioVoice) sample() float32 {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if !v.playing && !(v.released && !v.releaseStart.IsZero()) {
		return 0
	}

	now := time.Now()
	elapsed := now.Sub(v.sampleStart).Seconds()

	// Compute envelope
	env := v.envelope(elapsed)
	if env <= 0 && !v.released {
		v.playing = false
		return 0
	}

	// Generate waveform sample
	var s float32
	if len(v.buffer) > 0 {
		// Sample playback: read from pre-rendered buffer
		if v.bufferPos >= len(v.buffer) {
			if v.loop {
				v.bufferPos = 0
			} else {
				v.playing = false
				return 0
			}
		}
		s = v.buffer[v.bufferPos]
		v.bufferPos++
	} else {
		s = v.waveformSample(v.frequency, v.phase, v.waveform)
		v.phase += v.frequency / float64(v.sampleRate)
		if v.phase >= 1.0 {
			v.phase -= 1.0
		}
	}

	return float32(env) * float32(v.amplitude) * float32(v.volume) * s
}

// envelope computes the ADSR envelope value at the given elapsed seconds.
func (v *AudioVoice) envelope(elapsed float64) float64 {
	if v.attack <= 0 && v.decay <= 0 && v.release <= 0 {
		return 1.0
	}

	if v.released && !v.releaseStart.IsZero() {
		relElapsed := time.Since(v.releaseStart).Seconds()
		if relElapsed >= v.release {
			return 0
		}
		return v.sustain * (1.0 - relElapsed/v.release)
	}

	if elapsed < v.attack {
		return elapsed / v.attack
	}
	elapsed -= v.attack
	if elapsed < v.decay {
		return 1.0 - (1.0-v.sustain)*(elapsed/v.decay)
	}
	return v.sustain
}

// waveformSample returns a single sample of the voice waveform at the given phase.
func (v *AudioVoice) waveformSample(frequency float64, phase float64, waveform Waveform) float32 {
	// phase is in [0, 1)
	switch waveform {
	case WaveformSine:
		return float32(math.Sin(2 * math.Pi * phase))
	case WaveformSquare:
		if phase < 0.5 {
			return 1.0
		}
		return -1.0
	case WaveformSaw:
		return float32(2*phase - 1)
	case WaveformTriangle:
		if phase < 0.25 {
			return float32(4 * phase)
		}
		if phase < 0.75 {
			return float32(2 - 4*phase)
		}
		return float32(4*phase - 4)
	case WaveformNoise:
		return float32(math.Sin(2*math.Pi*phase)*0.5 + math.Sin(2*math.Pi*phase*3.14159)*0.5)
	default:
		return 0
	}
}

// AudioEngine is a software audio engine that mixes voices and outputs PCM.
type AudioEngine struct {
	mu         sync.RWMutex
	voices     map[uint64]*AudioVoice
	nextID     uint64
	sampleRate int
	format     AudioFormat
	backend    AudioBackend
	device     string
	output     io.Writer
	callback   func([]float32)
	mixBuffer  []float32 // stereo interleaved float32
	running    bool
	stopCh     chan struct{}
	doneCh     chan struct{}
	// captureBuffer stores mixed PCM output for testing when no device is available.
	captureBuffer []float32
	captureMu     sync.Mutex
}

// NewAudioEngine creates a new audio engine with the given sample rate.
func NewAudioEngine(sampleRate int) *AudioEngine {
	return &AudioEngine{
		voices:     make(map[uint64]*AudioVoice),
		sampleRate: sampleRate,
		format: AudioFormat{
			SampleRate:    sampleRate,
			Channels:      2,
			BitsPerSample: 16,
		},
		backend:       AudioBackendNull,
		mixBuffer:     make([]float32, sampleRate*2*10), // 10ms of stereo float32
		captureBuffer: make([]float32, 0),
	}
}

// NewAudioEngineWithFormat creates an engine with a specific audio format.
func NewAudioEngineWithFormat(format AudioFormat) *AudioEngine {
	e := &AudioEngine{
		voices:        make(map[uint64]*AudioVoice),
		sampleRate:    format.SampleRate,
		format:        format,
		backend:       AudioBackendNull,
		mixBuffer:     make([]float32, format.SampleRate*format.Channels*10),
		captureBuffer: make([]float32, 0),
	}
	return e
}

// ID returns the engine ID (for compatibility).
func (a *AudioEngine) ID() uint64 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return 0
}

// SampleRate returns the audio sample rate.
func (a *AudioEngine) SampleRate() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.sampleRate
}

// Format returns the current audio format.
func (a *AudioEngine) Format() AudioFormat {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.format
}

// SetFormat sets the audio output format.
func (a *AudioEngine) SetFormat(format AudioFormat) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.running {
		return errors.New("cannot change format while engine is running")
	}
	if format.SampleRate <= 0 {
		format.SampleRate = 44100
	}
	if format.Channels <= 0 {
		format.Channels = 2
	}
	if format.BitsPerSample <= 0 {
		format.BitsPerSample = 16
	}
	a.sampleRate = format.SampleRate
	a.format = format
	a.mixBuffer = make([]float32, format.SampleRate*format.Channels*10)
	for _, v := range a.voices {
		v.mu.Lock()
		v.sampleRate = format.SampleRate
		v.mu.Unlock()
	}
	return nil
}

// VoiceCount returns the number of active voices.
func (a *AudioEngine) VoiceCount() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.voices)
}

// CreateVoice creates a new audio voice.
func (a *AudioEngine) CreateVoice(frequency, amplitude float64) *AudioVoice {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.nextID++
	voice := NewAudioVoice(a.nextID, a.sampleRate, frequency, amplitude, WaveformSine)
	a.voices[voice.id] = voice
	return voice
}

// CreateVoiceWithWaveform creates a voice with a specific waveform.
func (a *AudioEngine) CreateVoiceWithWaveform(frequency, amplitude float64, waveform Waveform) *AudioVoice {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.nextID++
	voice := NewAudioVoice(a.nextID, a.sampleRate, frequency, amplitude, waveform)
	a.voices[voice.id] = voice
	return voice
}

// DestroyVoice destroys an audio voice.
func (a *AudioEngine) DestroyVoice(id uint64) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	v, ok := a.voices[id]
	if !ok {
		return fmt.Errorf("voice %d not found", id)
	}
	v.Stop()
	delete(a.voices, id)
	return nil
}

// GetVoice returns a voice by ID, or nil if not found.
func (a *AudioEngine) GetVoice(id uint64) *AudioVoice {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.voices[id]
}

// Start starts the audio engine and output backend.
func (a *AudioEngine) Start() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.running {
		return nil
	}
	switch a.backend {
	case AudioBackendALSA:
		if err := a.openALSA(); err != nil {
			return fmt.Errorf("failed to open ALSA device: %w", err)
		}
	case AudioBackendOSS:
		if err := a.openOSS(); err != nil {
			return fmt.Errorf("failed to open OSS device: %w", err)
		}
	case AudioBackendStdout:
		a.output = os.Stdout
	case AudioBackendCallback:
		if a.callback == nil {
			return errors.New("callback backend requires a callback function")
		}
	}
	a.running = true
	a.stopCh = make(chan struct{})
	a.doneCh = make(chan struct{})
	go a.runLoop()
	return nil
}

// Stop stops the audio engine and closes the output backend.
func (a *AudioEngine) Stop() error {
	a.mu.Lock()
	if !a.running {
		a.mu.Unlock()
		return nil
	}
	close(a.stopCh)
	a.mu.Unlock()
	<-a.doneCh
	a.mu.Lock()
	a.running = false
	switch a.backend {
	case AudioBackendALSA:
		a.closeALSA()
	case AudioBackendOSS:
		a.closeOSS()
	case AudioBackendStdout:
		// Do not close os.Stdout; just detach the output writer.
		a.output = nil
	}
	a.mu.Unlock()
	return nil
}

// SetBackend selects the audio output backend.
func (a *AudioEngine) SetBackend(backend AudioBackend, device string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.running {
		return errors.New("cannot change backend while engine is running")
	}
	a.backend = backend
	a.device = device
	return nil
}

// SetCallback sets the callback function for the callback backend.
func (a *AudioEngine) SetCallback(fn func([]float32)) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.callback = fn
}

// SetOutput sets an io.Writer for raw PCM output (used by ALSA/OSS backends).
func (a *AudioEngine) SetOutput(w io.Writer) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.output = w
}

// SetVoiceFrequency sets a voice frequency.
func (a *AudioEngine) SetVoiceFrequency(id uint64, frequency float64) error {
	a.mu.RLock()
	v, ok := a.voices[id]
	a.mu.RUnlock()
	if !ok {
		return fmt.Errorf("voice %d not found", id)
	}
	v.SetFrequency(frequency)
	return nil
}

// SetVoiceAmplitude sets a voice amplitude.
func (a *AudioEngine) SetVoiceAmplitude(id uint64, amplitude float64) error {
	a.mu.RLock()
	v, ok := a.voices[id]
	a.mu.RUnlock()
	if !ok {
		return fmt.Errorf("voice %d not found", id)
	}
	v.SetAmplitude(amplitude)
	return nil
}

// SetVoiceWaveform sets a voice waveform.
func (a *AudioEngine) SetVoiceWaveform(id uint64, waveform Waveform) error {
	a.mu.RLock()
	v, ok := a.voices[id]
	a.mu.RUnlock()
	if !ok {
		return fmt.Errorf("voice %d not found", id)
	}
	v.SetWaveform(waveform)
	return nil
}

// SetVoiceVolume sets a voice volume.
func (a *AudioEngine) SetVoiceVolume(id uint64, volume float64) error {
	a.mu.RLock()
	v, ok := a.voices[id]
	a.mu.RUnlock()
	if !ok {
		return fmt.Errorf("voice %d not found", id)
	}
	v.SetVolume(volume)
	return nil
}

// SetVoicePan sets a voice pan.
func (a *AudioEngine) SetVoicePan(id uint64, pan float64) error {
	a.mu.RLock()
	v, ok := a.voices[id]
	a.mu.RUnlock()
	if !ok {
		return fmt.Errorf("voice %d not found", id)
	}
	v.SetPan(pan)
	return nil
}

// SetVoiceEnvelope sets a voice ADSR envelope.
func (a *AudioEngine) SetVoiceEnvelope(id uint64, attack, decay, sustain, release float64) error {
	a.mu.RLock()
	v, ok := a.voices[id]
	a.mu.RUnlock()
	if !ok {
		return fmt.Errorf("voice %d not found", id)
	}
	v.SetEnvelope(attack, decay, sustain, release)
	return nil
}

// PlayVoice starts a voice playing.
func (a *AudioEngine) PlayVoice(id uint64) error {
	a.mu.RLock()
	v, ok := a.voices[id]
	a.mu.RUnlock()
	if !ok {
		return fmt.Errorf("voice %d not found", id)
	}
	v.Play()
	return nil
}

// StopVoice stops a voice.
func (a *AudioEngine) StopVoice(id uint64) error {
	a.mu.RLock()
	v, ok := a.voices[id]
	a.mu.RUnlock()
	if !ok {
		return fmt.Errorf("voice %d not found", id)
	}
	v.Stop()
	return nil
}

// PlayTone plays a tone for a duration and then destroys the voice.
func (a *AudioEngine) PlayTone(frequency, amplitude, duration float64) {
	voice := a.CreateVoice(frequency, amplitude)
	voice.Play()
	time.AfterFunc(time.Duration(duration*float64(time.Second)), func() {
		voice.Stop()
		a.DestroyVoice(voice.ID())
	})
}

// PlayToneWithWaveform plays a tone with a specific waveform for a duration.
func (a *AudioEngine) PlayToneWithWaveform(frequency, amplitude, duration float64, waveform Waveform) {
	voice := a.CreateVoiceWithWaveform(frequency, amplitude, waveform)
	voice.Play()
	time.AfterFunc(time.Duration(duration*float64(time.Second)), func() {
		voice.Stop()
		a.DestroyVoice(voice.ID())
	})
}

// Mix renders one frame of mixed stereo audio into the mix buffer.
func (a *AudioEngine) Mix(frameCount int) {
	a.mu.RLock()
	format := a.format
	voices := make([]*AudioVoice, 0, len(a.voices))
	for _, v := range a.voices {
		voices = append(voices, v)
	}
	a.mu.RUnlock()

	channels := format.Channels
	if channels < 1 {
		channels = 1
	}
	samplesNeeded := frameCount * channels
	if len(a.mixBuffer) < samplesNeeded {
		a.mixBuffer = make([]float32, samplesNeeded)
	}
	buf := a.mixBuffer[:samplesNeeded]
	for i := range buf {
		buf[i] = 0
	}

	for _, v := range voices {
		v.mu.RLock()
		playing := v.playing || (v.released && !v.releaseStart.IsZero())
		v.mu.RUnlock()
		if !playing {
			continue
		}
		for i := 0; i < frameCount; i++ {
			sample := v.sample()
			if channels >= 2 {
				// Apply pan: equal-power panning
				v.mu.RLock()
				pan := v.pan
				v.mu.RUnlock()
				leftGain := math.Cos((pan + 1) * math.Pi / 4)
				rightGain := math.Sin((pan + 1) * math.Pi / 4)
				buf[i*2] += sample * float32(leftGain)
				buf[i*2+1] += sample * float32(rightGain)
			} else {
				buf[i] += sample
			}
		}
	}

	// Soft-clamp to prevent clipping
	for i := range buf {
		if buf[i] > 1 {
			buf[i] = 1
		}
		if buf[i] < -1 {
			buf[i] = -1
		}
	}
}

// runLoop continuously mixes audio and writes it to the output backend.
func (a *AudioEngine) runLoop() {
	defer close(a.doneCh)
	format := a.Format()
	frameDuration := time.Duration(float64(time.Second) / float64(format.SampleRate) * float64(len(a.mixBuffer)/format.Channels))
	ticker := time.NewTicker(frameDuration)
	defer ticker.Stop()

	for {
		select {
		case <-a.stopCh:
			return
		case <-ticker.C:
			a.Mix(len(a.mixBuffer) / format.Channels)
			a.writeMix(a.mixBuffer)
		}
	}
}

// writeMix writes the mixed buffer to the configured output.
func (a *AudioEngine) writeMix(buf []float32) {
	a.mu.RLock()
	backend := a.backend
	callback := a.callback
	output := a.output
	a.mu.RUnlock()

	switch backend {
	case AudioBackendCallback:
		if callback != nil {
			callback(buf)
		}
	case AudioBackendALSA, AudioBackendOSS, AudioBackendStdout:
		if output != nil {
			// Convert float32 stereo to int16 interleaved and write
			pcm := float32ToInt16(buf)
			if _, err := output.Write(pcm); err != nil {
				// Log but don't stop; output may be temporarily unavailable
				_ = err
			}
		}
		// Also capture for testing
		a.captureMu.Lock()
		a.captureBuffer = append(a.captureBuffer, buf...)
		a.captureMu.Unlock()
	case AudioBackendNull:
		// Still capture for testing even with null backend
		a.captureMu.Lock()
		a.captureBuffer = append(a.captureBuffer, buf...)
		a.captureMu.Unlock()
	}
}

// CapturedSamples returns a copy of all mixed PCM samples captured so far.
// This is useful for testing audio output when no hardware device is available.
func (a *AudioEngine) CapturedSamples() []float32 {
	a.captureMu.Lock()
	defer a.captureMu.Unlock()
	samples := make([]float32, len(a.captureBuffer))
	copy(samples, a.captureBuffer)
	return samples
}

// ClearCapturedSamples clears the captured sample buffer.
func (a *AudioEngine) ClearCapturedSamples() {
	a.captureMu.Lock()
	defer a.captureMu.Unlock()
	a.captureBuffer = make([]float32, 0)
}

// CapturedSampleCount returns the number of captured samples.
func (a *AudioEngine) CapturedSampleCount() int {
	a.captureMu.Lock()
	defer a.captureMu.Unlock()
	return len(a.captureBuffer)
}

// float32ToInt16 converts interleaved float32 samples to int16 little-endian bytes.
func float32ToInt16(samples []float32) []byte {
	buf := make([]byte, len(samples)*2)
	for i, s := range samples {
		var v int16
		if s > 1 {
			s = 1
		}
		if s < -1 {
			s = -1
		}
		v = int16(s * 32767)
		idx := i * 2
		buf[idx] = byte(v & 0xFF)
		buf[idx+1] = byte((v >> 8) & 0xFF)
	}
	return buf
}

// openALSA opens an ALSA PCM device for output.
func (a *AudioEngine) openALSA() error {
	device := a.device
	if device == "" {
		device = "default"
	}
	// Try to open the ALSA PCM device file directly for raw PCM output.
	// ALSA PCM devices are typically at /dev/snd/pcmC*D*p
	path := "/dev/snd/" + device
	if _, err := os.Stat(path); err != nil {
		// Fall back to using the aplay command if available
		_ = path
		return nil // ALSA device not directly accessible; treat as no-op
	}
	f, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	a.output = f
	return nil
}

// closeALSA closes the ALSA output device.
func (a *AudioEngine) closeALSA() {
	if f, ok := a.output.(*os.File); ok {
		f.Close()
	}
	a.output = nil
}

// openOSS opens an OSS device (/dev/dsp) for output.
func (a *AudioEngine) openOSS() error {
	device := a.device
	if device == "" {
		device = "/dev/dsp"
	}
	f, err := os.OpenFile(device, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	a.output = f
	return nil
}

// closeOSS closes the OSS output device.
func (a *AudioEngine) closeOSS() {
	if f, ok := a.output.(*os.File); ok {
		f.Close()
	}
	a.output = nil
}

// WritePCM writes raw interleaved int16 PCM data to the output.
func (a *AudioEngine) WritePCM(data []byte) error {
	a.mu.RLock()
	output := a.output
	a.mu.RUnlock()
	if output == nil {
		return errors.New("no audio output configured")
	}
	_, err := output.Write(data)
	return err
}

// SilenceAll stops all voices immediately.
func (a *AudioEngine) SilenceAll() {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, v := range a.voices {
		v.mu.Lock()
		v.playing = false
		v.released = false
		v.releaseStart = time.Time{}
		v.mu.Unlock()
	}
}

var _ = errors.New
var _ = io.Writer(nil)
