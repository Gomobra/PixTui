package animation

import (
	"testing"
	"time"
)

func TestAnimationCreate(t *testing.T) {
	anim := NewAnimation("test-anim", time.Second, 0, 100, Linear)
	if anim.ID != "test-anim" {
		t.Errorf("expected ID 'test-anim', got '%s'", anim.ID)
	}
	if anim.From != 0 || anim.To != 100 {
		t.Error("expected from=0, to=100")
	}
	if anim.IsRunning() {
		t.Error("expected not running initially")
	}
}

func TestAnimationStartStop(t *testing.T) {
	anim := NewAnimation("anim", time.Second, 0, 100, Linear)
	anim.Start()
	if !anim.IsRunning() {
		t.Error("expected running after Start")
	}
	anim.Stop()
	if anim.IsRunning() {
		t.Error("expected not running after Stop")
	}
}

func TestAnimationUpdate(t *testing.T) {
	var lastValue float64
	anim := NewAnimation("anim", time.Second, 0, 100, Linear)
	anim.OnUpdate = func(v float64) {
		lastValue = v
	}
	anim.Start()

	// Update halfway
	complete := anim.Update(time.Second / 2)
	if complete {
		t.Error("expected not complete at halfway")
	}
	if lastValue < 49 || lastValue > 51 {
		t.Errorf("expected value around 50, got %f", lastValue)
	}

	// Update to complete
	complete = anim.Update(time.Second / 2)
	if !complete {
		t.Error("expected complete after full duration")
	}
	if lastValue != 100 {
		t.Errorf("expected final value 100, got %f", lastValue)
	}
}

func TestAnimationLoop(t *testing.T) {
	var callCount int
	anim := NewAnimation("loop-anim", time.Millisecond*100, 0, 1, Linear)
	anim.OnUpdate = func(v float64) {
		callCount++
	}
	anim.SetLoop(true)
	anim.Start()

	// Update multiple times - should loop
	anim.Update(time.Millisecond * 50)
	anim.Update(time.Millisecond * 50)
	anim.Update(time.Millisecond * 50)
	anim.Update(time.Millisecond * 50)

	if callCount < 4 {
		t.Errorf("expected at least 4 update calls for looping anim, got %d", callCount)
	}
}

func TestAnimationEasing(t *testing.T) {
	// Linear: t=0.5 -> 0.5
	if Linear(0.5) != 0.5 {
		t.Error("expected linear easing to return t")
	}

	// EaseIn: t=0.5 -> 0.25
	if EaseIn(0.5) != 0.25 {
		t.Errorf("expected ease-in(0.5)=0.25, got %f", EaseIn(0.5))
	}

	// EaseOut: t=0.5 -> 0.75
	if EaseOut(0.5) != 0.75 {
		t.Errorf("expected ease-out(0.5)=0.75, got %f", EaseOut(0.5))
	}

	// EaseInOut: t=0.5 -> 0.5
	if EaseInOut(0.5) != 0.5 {
		t.Errorf("expected ease-in-out(0.5)=0.5, got %f", EaseInOut(0.5))
	}

	// Boundary values
	if EaseIn(0) != 0 || EaseIn(1) != 1 {
		t.Error("expected ease-in boundaries to be 0 and 1")
	}
	if EaseOut(0) != 0 || EaseOut(1) != 1 {
		t.Error("expected ease-out boundaries to be 0 and 1")
	}
}

func TestAnimatorAddRemove(t *testing.T) {
	an := NewAnimator()
	anim := NewAnimation("a1", time.Second, 0, 10, Linear)
	id := an.Add(anim)
	if id != "a1" {
		t.Errorf("expected ID 'a1', got '%s'", id)
	}
	if an.Count() != 1 {
		t.Errorf("expected 1 animation, got %d", an.Count())
	}

	an.Remove(id)
	if an.Count() != 0 {
		t.Errorf("expected 0 animations after remove, got %d", an.Count())
	}
}

func TestAnimatorAutoID(t *testing.T) {
	an := NewAnimator()
	anim := NewAnimation("", time.Second, 0, 10, Linear)
	id := an.Add(anim)
	if id == "" {
		t.Error("expected auto-generated ID")
	}
	if anim.ID != id {
		t.Error("expected animation ID to be set to auto ID")
	}
}

func TestAnimatorUpdateComplete(t *testing.T) {
	an := NewAnimator()
	anim := NewAnimation("quick", time.Millisecond*10, 0, 1, Linear)
	var completed bool
	anim.OnComplete = func() {
		completed = true
	}
	an.Add(anim)

	results := an.Update(time.Millisecond * 20)
	if len(results) == 0 {
		t.Error("expected animation to complete")
	}
	if !completed {
		t.Error("expected OnComplete to be called")
	}
}

func TestAnimatorStopAll(t *testing.T) {
	an := NewAnimator()
	anim := NewAnimation("long", time.Second, 0, 10, Linear)
	anim.Start()
	an.Add(anim)
	an.StopAll()
	if anim.IsRunning() {
		t.Error("expected animation to be stopped by StopAll")
	}
}

func TestTimelinePlay(t *testing.T) {
	tl := NewTimeline()
	anim := NewAnimation("tl-anim", time.Millisecond*50, 0, 10, Linear)
	tl.Add(anim, 0)
	tl.Play()

	if !tl.IsRunning() {
		t.Error("expected timeline to be running after Play")
	}

	// Update enough to complete
	tl.Update(time.Millisecond * 100)
	if tl.IsRunning() {
		t.Error("expected timeline to complete after duration")
	}
}

func TestTimelineOffset(t *testing.T) {
	tl := NewTimeline()
	anim := NewAnimation("offset-anim", time.Second, 0, 10, Linear)
	tl.Add(anim, time.Millisecond*50)
	tl.Play()

	// Before offset - animation should not be running
	tl.Update(time.Millisecond * 10)
	if anim.IsRunning() {
		t.Error("expected animation not running before offset")
	}

	// After offset - animation should be running
	tl.Update(time.Millisecond * 50)
	if !anim.IsRunning() {
		t.Error("expected animation running after offset")
	}
}

func TestTimelineStop(t *testing.T) {
	tl := NewTimeline()
	anim := NewAnimation("stop-anim", time.Second, 0, 10, Linear)
	tl.Add(anim, 0)
	tl.Play()
	tl.Stop()
	if tl.IsRunning() {
		t.Error("expected timeline not running after Stop")
	}
	if anim.IsRunning() {
		t.Error("expected animation stopped when timeline stopped")
	}
}

func TestEasingElastic(t *testing.T) {
	// Boundaries
	if EaseInElastic(0) != 0 || EaseInElastic(1) != 1 {
		t.Error("expected ease-in-elastic boundaries to be 0 and 1")
	}
	if EaseOutElastic(0) != 0 || EaseOutElastic(1) != 1 {
		t.Error("expected ease-out-elastic boundaries to be 0 and 1")
	}
	if EaseInOutElastic(0) != 0 || EaseInOutElastic(1) != 1 {
		t.Error("expected ease-in-out-elastic boundaries to be 0 and 1")
	}
	// Values should be within [0, 1] mostly (elastic can overshoot slightly, but endpoints exact)
	if EaseOutElastic(0.5) < 0 || EaseOutElastic(0.5) > 2 {
		t.Errorf("expected ease-out-elastic(0.5) in reasonable range, got %f", EaseOutElastic(0.5))
	}
	// EasingForName should resolve elastic names
	if EasingForName("outElastic")(0.5) != EaseOutElastic(0.5) {
		t.Error("expected EasingForName('outElastic') to match EaseOutElastic")
	}
	if EasingForName("inOutElastic")(0.5) != EaseInOutElastic(0.5) {
		t.Error("expected EasingForName('inOutElastic') to match EaseInOutElastic")
	}
}

func TestEnhancedTimelineOnce(t *testing.T) {
	tl := NewEnhancedTimeline()
	target := map[string]float64{"x": 0}
	tl.Once(target, map[string]float64{"x": 100}, time.Second)
	if tl.Count() != 1 {
		t.Errorf("expected 1 item, got %d", tl.Count())
	}
	tl.Play()
	tl.Update(time.Second)
	// After completion, once item should be removed
	if tl.Count() != 0 {
		t.Errorf("expected once item to be removed after completion, got %d", tl.Count())
	}
	if target["x"] != 100 {
		t.Errorf("expected target x to be 100, got %f", target["x"])
	}
}

func TestEnhancedTimelineMultipleTargets(t *testing.T) {
	tl := NewEnhancedTimeline()
	a := map[string]float64{"x": 0}
	b := map[string]float64{"x": 10}
	item := &EnhancedAnimationItem{
		Targets:   []any{a, b},
		Properties: map[string]float64{"x": 50},
		Duration:  time.Second,
		Easing:    Linear,
		StartTime: 0,
		Loop:      1,
	}
	tl.mu.Lock()
	tl.items = append(tl.items, item)
	tl.mu.Unlock()
	tl.Play()
	tl.Update(time.Second / 2)
	if a["x"] != 25 || b["x"] != 30 {
		t.Errorf("expected both targets to animate toward 50, got a=%f b=%f", a["x"], b["x"])
	}
}

func TestTimelineEngine(t *testing.T) {
	engine := NewTimelineEngine()
	if engine.Count() != 0 {
		t.Error("expected empty engine")
	}
	tl := NewEnhancedTimeline()
	engine.Register(tl)
	if engine.Count() != 1 {
		t.Errorf("expected 1 registered timeline, got %d", engine.Count())
	}
	engine.Update(time.Second)
	engine.Unregister(tl)
	if engine.Count() != 0 {
		t.Errorf("expected 0 after unregister, got %d", engine.Count())
	}
}
