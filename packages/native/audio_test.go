package native

import (
	"testing"
	"time"
)

func TestAudioEngineCreate(t *testing.T) {
	engine := NewAudioEngine(44100)
	if engine == nil {
		t.Fatal("expected non-nil engine")
	}
	if engine.SampleRate() != 44100 {
		t.Errorf("expected sample rate 44100, got %d", engine.SampleRate())
	}
	if engine.VoiceCount() != 0 {
		t.Errorf("expected 0 voices, got %d", engine.VoiceCount())
	}
}

func TestAudioEngineFormat(t *testing.T) {
	engine := NewAudioEngine(44100)
	format := engine.Format()
	if format.SampleRate != 44100 {
		t.Errorf("expected sample rate 44100, got %d", format.SampleRate)
	}
	if format.Channels != 2 {
		t.Errorf("expected 2 channels, got %d", format.Channels)
	}
	if format.BitsPerSample != 16 {
		t.Errorf("expected 16 bits, got %d", format.BitsPerSample)
	}
}

func TestAudioEngineSetFormat(t *testing.T) {
	engine := NewAudioEngine(44100)
	err := engine.SetFormat(AudioFormat{SampleRate: 48000, Channels: 2, BitsPerSample: 16})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if engine.SampleRate() != 48000 {
		t.Errorf("expected sample rate 48000, got %d", engine.SampleRate())
	}
}

func TestAudioEngineCreateVoice(t *testing.T) {
	engine := NewAudioEngine(44100)
	voice := engine.CreateVoice(440.0, 0.5)
	if voice == nil {
		t.Fatal("expected non-nil voice")
	}
	if engine.VoiceCount() != 1 {
		t.Errorf("expected 1 voice, got %d", engine.VoiceCount())
	}
	if voice.ID() == 0 {
		t.Error("expected non-zero voice ID")
	}
}

func TestAudioEngineDestroyVoice(t *testing.T) {
	engine := NewAudioEngine(44100)
	voice := engine.CreateVoice(440.0, 0.5)
	err := engine.DestroyVoice(voice.ID())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if engine.VoiceCount() != 0 {
		t.Errorf("expected 0 voices after destroy, got %d", engine.VoiceCount())
	}
	err = engine.DestroyVoice(999)
	if err == nil {
		t.Error("expected error for non-existent voice")
	}
}

func TestAudioVoicePlayStop(t *testing.T) {
	engine := NewAudioEngine(44100)
	voice := engine.CreateVoice(440.0, 0.5)
	voice.Play()
	if !voice.IsPlaying() {
		t.Error("expected voice playing")
	}
	voice.Stop()
	if voice.IsPlaying() {
		t.Error("expected voice not playing")
	}
}

func TestAudioVoiceSetters(t *testing.T) {
	engine := NewAudioEngine(44100)
	voice := engine.CreateVoice(440.0, 0.5)
	voice.SetFrequency(880.0)
	voice.SetAmplitude(0.8)
	voice.SetVolume(0.5)
	voice.SetPan(0.5)
	voice.SetWaveform(WaveformSquare)
	voice.SetEnvelope(0.1, 0.1, 0.5, 0.2)
}

func TestAudioEnginePlayTone(t *testing.T) {
	engine := NewAudioEngine(44100)
	engine.PlayTone(440.0, 0.5, 0.01)
	time.Sleep(50 * time.Millisecond)
	if engine.VoiceCount() != 0 {
		t.Errorf("expected 0 voices after tone completes, got %d", engine.VoiceCount())
	}
}

func TestAudioEnginePlayToneWithWaveform(t *testing.T) {
	engine := NewAudioEngine(44100)
	engine.PlayToneWithWaveform(440.0, 0.5, 0.01, WaveformSquare)
	time.Sleep(50 * time.Millisecond)
	if engine.VoiceCount() != 0 {
		t.Errorf("expected 0 voices after tone completes, got %d", engine.VoiceCount())
	}
}

func TestAudioEngineMix(t *testing.T) {
	engine := NewAudioEngine(44100)
	voice := engine.CreateVoice(440.0, 0.5)
	voice.Play()
	engine.Mix(1024)
}

func TestAudioEngineSilenceAll(t *testing.T) {
	engine := NewAudioEngine(44100)
	v1 := engine.CreateVoice(440.0, 0.5)
	v2 := engine.CreateVoice(880.0, 0.5)
	v1.Play()
	v2.Play()
	engine.SilenceAll()
	if v1.IsPlaying() {
		t.Error("expected v1 not playing after silence")
	}
	if v2.IsPlaying() {
		t.Error("expected v2 not playing after silence")
	}
}

func TestAudioEngineCapturedSamples(t *testing.T) {
	engine := NewAudioEngine(44100)
	engine.ClearCapturedSamples()

	voice := engine.CreateVoice(440.0, 0.5)
	voice.Play()

	// Mix and manually write to trigger capture (simulates runLoop behavior)
	for i := 0; i < 10; i++ {
		engine.Mix(1024)
		engine.writeMix(engine.mixBuffer)
	}

	count := engine.CapturedSampleCount()
	if count == 0 {
		t.Error("expected captured samples > 0")
	}

	samples := engine.CapturedSamples()
	if len(samples) != count {
		t.Errorf("expected %d samples, got %d", count, len(samples))
	}

	engine.ClearCapturedSamples()
	if engine.CapturedSampleCount() != 0 {
		t.Error("expected 0 captured samples after clear")
	}
}

func TestAudioWaveformString(t *testing.T) {
	if WaveformSine.String() != "sine" {
		t.Error("expected 'sine'")
	}
	if WaveformSquare.String() != "square" {
		t.Error("expected 'square'")
	}
	if WaveformSaw.String() != "saw" {
		t.Error("expected 'saw'")
	}
	if WaveformTriangle.String() != "triangle" {
		t.Error("expected 'triangle'")
	}
	if WaveformNoise.String() != "noise" {
		t.Error("expected 'noise'")
	}
}

func TestAudioEngineStartStop(t *testing.T) {
	engine := NewAudioEngine(44100)
	err := engine.Start()
	if err != nil {
		t.Fatalf("unexpected error starting engine: %v", err)
	}
	if !engine.running {
		t.Error("expected engine running")
	}
	time.Sleep(50 * time.Millisecond)
	err = engine.Stop()
	if err != nil {
		t.Fatalf("unexpected error stopping engine: %v", err)
	}
	if engine.running {
		t.Error("expected engine stopped")
	}
}

func TestAudioEngineSetBackendCallback(t *testing.T) {
	engine := NewAudioEngine(44100)
	err := engine.SetBackend(AudioBackendCallback, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	engine.SetCallback(func(buf []float32) {})
	err = engine.Start()
	if err != nil {
		t.Fatalf("unexpected error starting engine: %v", err)
	}
	time.Sleep(50 * time.Millisecond)
	engine.Stop()
}

func TestFloat32ToInt16(t *testing.T) {
	samples := []float32{0, 0.5, -0.5, 1.0, -1.0}
	pcm := float32ToInt16(samples)
	if len(pcm) != len(samples)*2 {
		t.Errorf("expected %d bytes, got %d", len(samples)*2, len(pcm))
	}
}

func TestAudioEngineWritePCMNoOutput(t *testing.T) {
	engine := NewAudioEngine(44100)
	err := engine.WritePCM([]byte{0, 0})
	if err == nil {
		t.Error("expected error when no output configured")
	}
}

// ----- Benchmarks -----

func BenchmarkAudioEngineMix(b *testing.B) {
	engine := NewAudioEngine(44100)
	for i := 0; i < 10; i++ {
		engine.CreateVoice(440.0, 0.5).Play()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Mix(1024)
	}
}

func BenchmarkAudioEngineWriteMix(b *testing.B) {
	engine := NewAudioEngine(44100)
	for i := 0; i < 10; i++ {
		engine.CreateVoice(440.0, 0.5).Play()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Mix(1024)
		engine.writeMix(engine.mixBuffer)
	}
}

func BenchmarkAudioVoiceSample(b *testing.B) {
	engine := NewAudioEngine(44100)
	voice := engine.CreateVoice(440.0, 0.5)
	voice.Play()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		voice.sample()
	}
}
