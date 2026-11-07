package animation

import (
	"math"
	"reflect"
	"sync"
	"time"
)

// EasingFunc is a function that computes the easing value for a given t (0 to 1).
type EasingFunc func(t float64) float64

// Easing functions.
var (
	// Linear is a constant-speed easing.
	Linear EasingFunc = func(t float64) float64 { return t }

	// EaseIn starts slow and accelerates (quadratic).
	EaseIn EasingFunc = func(t float64) float64 { return t * t }

	// EaseOut starts fast and decelerates (quadratic).
	EaseOut EasingFunc = func(t float64) float64 { return 1 - math.Pow(1-t, 2) }

	// EaseInOut starts slow, accelerates, then decelerates (quadratic).
	EaseInOut EasingFunc = func(t float64) float64 {
		if t < 0.5 {
			return 2 * t * t
		}
		return -1 + (4-2*t)*t
	}

	// EaseInCubic is a cubic ease-in.
	EaseInCubic EasingFunc = func(t float64) float64 { return t * t * t }

	// EaseOutCubic is a cubic ease-out.
	EaseOutCubic EasingFunc = func(t float64) float64 { return 1 - math.Pow(1-t, 3) }

	// EaseInOutCubic is a cubic ease-in-out.
	EaseInOutCubic EasingFunc = func(t float64) float64 {
		if t < 0.5 {
			return 4 * t * t * t
		}
		return 1 - math.Pow(-2*t+2, 3)/2
	}

	// EaseInQuart is a quartic ease-in.
	EaseInQuart EasingFunc = func(t float64) float64 { return t * t * t * t }

	// EaseOutQuart is a quartic ease-out.
	EaseOutQuart EasingFunc = func(t float64) float64 { return 1 - math.Pow(1-t, 4) }

	// EaseInOutQuart is a quartic ease-in-out.
	EaseInOutQuart EasingFunc = func(t float64) float64 {
		if t < 0.5 {
			return 8 * t * t * t * t
		}
		return 1 - math.Pow(-2*t+2, 4)/2
	}

	// EaseInQuint is a quintic ease-in.
	EaseInQuint EasingFunc = func(t float64) float64 { return t * t * t * t * t }

	// EaseOutQuint is a quintic ease-out.
	EaseOutQuint EasingFunc = func(t float64) float64 { return 1 - math.Pow(1-t, 5) }

	// EaseInOutQuint is a quintic ease-in-out.
	EaseInOutQuint EasingFunc = func(t float64) float64 {
		if t < 0.5 {
			return 16 * t * t * t * t * t
		}
		return 1 - math.Pow(-2*t+2, 5)/2
	}

	// EaseInSine is a sine-based ease-in.
	EaseInSine EasingFunc = func(t float64) float64 { return 1 - math.Cos(t*math.Pi/2) }

	// EaseOutSine is a sine-based ease-out.
	EaseOutSine EasingFunc = func(t float64) float64 { return math.Sin(t * math.Pi / 2) }

	// EaseInOutSine is a sine-based ease-in-out.
	EaseInOutSine EasingFunc = func(t float64) float64 { return -(math.Cos(math.Pi*t) - 1) / 2 }

	// EaseInExpo is an exponential ease-in.
	EaseInExpo EasingFunc = func(t float64) float64 {
		if t == 0 {
			return 0
		}
		return math.Pow(2, 10*(t-1))
	}

	// EaseOutExpo is an exponential ease-out.
	EaseOutExpo EasingFunc = func(t float64) float64 {
		if t == 1 {
			return 1
		}
		return 1 - math.Pow(2, -10*t)
	}

	// EaseInOutExpo is an exponential ease-in-out.
	EaseInOutExpo EasingFunc = func(t float64) float64 {
		if t == 0 {
			return 0
		}
		if t == 1 {
			return 1
		}
		if t < 0.5 {
			return math.Pow(2, 20*t-10) / 2
		}
		return (2 - math.Pow(2, -20*t+10)) / 2
	}

	// EaseInCirc is a circular ease-in.
	EaseInCirc EasingFunc = func(t float64) float64 { return 1 - math.Sqrt(1-t*t) }

	// EaseOutCirc is a circular ease-out.
	EaseOutCirc EasingFunc = func(t float64) float64 { return math.Sqrt(1 - math.Pow(t-1, 2)) }

	// EaseInOutCirc is a circular ease-in-out.
	EaseInOutCirc EasingFunc = func(t float64) float64 {
		if t < 0.5 {
			return (1 - math.Sqrt(1-4*t*t)) / 2
		}
		return (math.Sqrt(1-math.Pow(-2*t+2, 2)) + 1) / 2
	}

	// EaseInBounce is a bounce ease-in.
	EaseInBounce EasingFunc = func(t float64) float64 { return 1 - EaseOutBounce(1-t) }

	// EaseOutBounce is a bounce ease-out.
	EaseOutBounce EasingFunc = func(t float64) float64 {
		const n1 = 7.5625
		const d1 = 2.75
		if t < 1/d1 {
			return n1 * t * t
		} else if t < 2/d1 {
			return n1*(t-1.5/d1)*t + 0.75
		} else if t < 2.5/d1 {
			return n1*(t-2.25/d1)*t + 0.9375
		}
		return n1*(t-2.625/d1)*t + 0.984375
	}

	// EaseInOutBounce is a bounce ease-in-out.
	EaseInOutBounce EasingFunc = func(t float64) float64 {
		if t < 0.5 {
			return (1 - EaseOutBounce(1-2*t)) / 2
		}
		return (1 + EaseOutBounce(2*t-1)) / 2
	}

	// EaseInBack is a back ease-in (overshoots).
	EaseInBack EasingFunc = func(t float64) float64 {
		const s = 1.70158
		return t * t * ((s+1)*t - s)
	}

	// EaseOutBack is a back ease-out (overshoots).
	EaseOutBack EasingFunc = func(t float64) float64 {
		const s = 1.70158
		return (t-1)*(t-1)*((s+1)*t+s) + 1
	}

	// EaseInOutBack is a back ease-in-out (overshoots).
	EaseInOutBack EasingFunc = func(t float64) float64 {
		const s = 1.70158 * 1.525
		if t < 0.5 {
			return (t * 2) * (t * 2) * ((s+1)*(t*2) - s) / 2
		}
		return ((t*2-2)*(t*2-2)*((s+1)*(t*2-2)+s) + 2) / 2
	}

	// EaseInElastic is an elastic ease-in.
	EaseInElastic EasingFunc = func(t float64) float64 {
		const c4 = (2 * math.Pi) / 3
		if t == 0 {
			return 0
		}
		if t == 1 {
			return 1
		}
		return -(math.Pow(2, 10*(t-1)) * math.Sin((t*10-0.75)*c4))
	}

	// EaseOutElastic is an elastic ease-out.
	EaseOutElastic EasingFunc = func(t float64) float64 {
		const c4 = (2 * math.Pi) / 3
		if t == 0 {
			return 0
		}
		if t == 1 {
			return 1
		}
		return math.Pow(2, -10*t)*math.Sin((t*10-0.75)*c4) + 1
	}

	// EaseInOutElastic is an elastic ease-in-out.
	EaseInOutElastic EasingFunc = func(t float64) float64 {
		const c4 = (2 * math.Pi) / 3
		if t == 0 {
			return 0
		}
		if t == 1 {
			return 1
		}
		if t < 0.5 {
			return -(math.Pow(2, 20*t-10) * math.Sin((20*t-11.125)*c4)) / 2
		}
		return (math.Pow(2, -20*t+10)*math.Sin((20*t-11.125)*c4))/2 + 1
	}
)

// Animation represents a single animation that interpolates a value over time.
type Animation struct {
	ID         string
	Duration   time.Duration
	Easing     EasingFunc
	From       float64
	To         float64
	OnUpdate   func(value float64)
	OnComplete func()

	elapsed time.Duration
	running bool
	loop    bool
}

// NewAnimation creates a new Animation.
func NewAnimation(id string, duration time.Duration, from, to float64, easing EasingFunc) *Animation {
	return &Animation{
		ID:       id,
		Duration: duration,
		Easing:   easing,
		From:     from,
		To:       to,
		elapsed:  0,
		running:  false,
	}
}

// Start begins the animation.
func (a *Animation) Start() {
	a.elapsed = 0
	a.running = true
}

// Stop stops the animation.
func (a *Animation) Stop() {
	a.running = false
}

// IsRunning returns true if the animation is currently running.
func (a *Animation) IsRunning() bool {
	return a.running
}

// Update advances the animation by the given delta time. Returns true if complete.
func (a *Animation) Update(deltaTime time.Duration) bool {
	if !a.running {
		return true
	}

	a.elapsed += deltaTime
	t := float64(a.elapsed) / float64(a.Duration)
	if t > 1.0 {
		t = 1.0
	}

	eased := a.Easing(t)
	value := a.From + (a.To-a.From)*eased

	if a.OnUpdate != nil {
		a.OnUpdate(value)
	}

	if t >= 1.0 {
		if a.loop {
			a.elapsed = 0
			return false
		}
		a.running = false
		if a.OnComplete != nil {
			a.OnComplete()
		}
		return true
	}
	return false
}

// SetLoop sets whether the animation loops.
func (a *Animation) SetLoop(loop bool) {
	a.loop = loop
}

// SetEasing sets the easing function for the animation.
func (a *Animation) SetEasing(easing EasingFunc) {
	a.Easing = easing
}

// SetRange sets the from/to range for the animation.
func (a *Animation) SetRange(from, to float64) {
	a.From = from
	a.To = to
}

// Progress returns the current progress (0 to 1) of the animation.
func (a *Animation) Progress() float64 {
	if a.Duration == 0 {
		return 1
	}
	t := float64(a.elapsed) / float64(a.Duration)
	if t > 1.0 {
		return 1.0
	}
	if t < 0 {
		return 0
	}
	return t
}

// Elapsed returns the elapsed time of the animation.
func (a *Animation) Elapsed() time.Duration {
	return a.elapsed
}

// Animator manages multiple animations and advances them over time.
type Animator struct {
	mu         sync.Mutex
	animations map[string]*Animation
	nextID     int
}

// NewAnimator creates a new Animator.
func NewAnimator() *Animator {
	return &Animator{
		animations: make(map[string]*Animation),
	}
}

// Add adds an animation to the animator and returns its ID.
func (an *Animator) Add(animation *Animation) string {
	an.mu.Lock()
	defer an.mu.Unlock()

	if animation.ID == "" {
		an.nextID++
		animation.ID = "anim_" + itoa(an.nextID)
	}

	animation.Start()
	an.animations[animation.ID] = animation
	return animation.ID
}

// Remove removes an animation by ID.
func (an *Animator) Remove(id string) {
	an.mu.Lock()
	defer an.mu.Unlock()
	delete(an.animations, id)
}

// Get returns an animation by ID.
func (an *Animator) Get(id string) (*Animation, bool) {
	an.mu.Lock()
	defer an.mu.Unlock()
	a, ok := an.animations[id]
	return a, ok
}

// Update advances all animations by the given delta time. Returns IDs of completed animations.
func (an *Animator) Update(deltaTime time.Duration) []string {
	an.mu.Lock()
	defer an.mu.Unlock()

	var completed []string
	for id, a := range an.animations {
		if a.Update(deltaTime) {
			completed = append(completed, id)
		}
	}
	return completed
}

// Count returns the number of active animations.
func (an *Animator) Count() int {
	an.mu.Lock()
	defer an.mu.Unlock()
	return len(an.animations)
}

// StopAll stops all running animations.
func (an *Animator) StopAll() {
	an.mu.Lock()
	defer an.mu.Unlock()
	for _, a := range an.animations {
		a.Stop()
	}
}

// PauseAll pauses all running animations without clearing them.
func (an *Animator) PauseAll() {
	an.mu.Lock()
	defer an.mu.Unlock()
	for _, a := range an.animations {
		a.Stop()
	}
}

// ResumeAll resumes all animations that were previously started.
func (an *Animator) ResumeAll() {
	an.mu.Lock()
	defer an.mu.Unlock()
	for _, a := range an.animations {
		a.Start()
	}
}

// TimelineEntry schedules an animation at a specific time offset.
type TimelineEntry struct {
	Animation *Animation
	Offset    time.Duration // delay before this animation starts
	started   bool
}

// Timeline manages a sequence of animations that play over time.
type Timeline struct {
	mu         sync.Mutex
	entries    []*TimelineEntry
	animator   *Animator
	running    bool
	elapsed    time.Duration
	duration   time.Duration
	loop       bool
	onComplete func()
}

// NewTimeline creates a new Timeline.
func NewTimeline() *Timeline {
	return &Timeline{
		animator: NewAnimator(),
	}
}

// SetDuration sets the total duration of the timeline.
func (t *Timeline) SetDuration(d time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.duration = d
}

// SetLoop sets whether the timeline loops when it reaches its duration.
func (t *Timeline) SetLoop(loop bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.loop = loop
}

// SetOnComplete sets a callback invoked when the timeline completes (non-looping).
func (t *Timeline) SetOnComplete(cb func()) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onComplete = cb
}

// Add schedules an animation to play after the given offset.
func (t *Timeline) Add(animation *Animation, offset time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.entries = append(t.entries, &TimelineEntry{
		Animation: animation,
		Offset:    offset,
		started:   false,
	})
}

// Play starts the timeline from the beginning.
func (t *Timeline) Play() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.running = true
	t.elapsed = 0
	t.animator.StopAll()
	for _, entry := range t.entries {
		entry.started = false
		entry.Animation.Stop()
	}
}

// Stop stops the timeline.
func (t *Timeline) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.running = false
	t.animator.StopAll()
}

// Pause pauses the timeline without resetting.
func (t *Timeline) Pause() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.running = false
	t.animator.PauseAll()
}

// Resume resumes a paused timeline.
func (t *Timeline) Resume() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.running = true
	t.animator.ResumeAll()
}

// Restart restarts the timeline from the beginning.
func (t *Timeline) Restart() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.running = true
	t.elapsed = 0
	t.animator.StopAll()
	for _, entry := range t.entries {
		entry.started = false
		entry.Animation.Stop()
		entry.Animation.elapsed = 0
	}
}

// IsRunning returns true if the timeline is running.
func (t *Timeline) IsRunning() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.running
}

// IsComplete returns true if the timeline has completed (non-looping).
func (t *Timeline) IsComplete() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return !t.running && t.elapsed > 0 && t.duration > 0 && t.elapsed >= t.duration
}

// Elapsed returns the elapsed time of the timeline.
func (t *Timeline) Elapsed() time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.elapsed
}

// Update advances the timeline by the given delta time. Returns true if complete.
func (t *Timeline) Update(deltaTime time.Duration) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.running {
		return true
	}

	t.elapsed += deltaTime

	// Handle looping
	if t.duration > 0 && t.elapsed >= t.duration {
		if t.loop {
			overshoot := t.elapsed % t.duration
			t.elapsed = 0
			t.animator.StopAll()
			for _, entry := range t.entries {
				entry.started = false
				entry.Animation.Stop()
				entry.Animation.elapsed = 0
			}
			if overshoot > 0 {
				// Re-apply overshoot by processing entries
				t.running = true
				t.mu.Unlock()
				result := t.Update(overshoot)
				t.mu.Lock()
				return result
			}
			return false
		}
		t.running = false
		if t.onComplete != nil {
			t.onComplete()
		}
		return true
	}

	// Add animations whose offset has been reached
	for _, entry := range t.entries {
		if !entry.started && t.elapsed >= entry.Offset {
			entry.Animation.Start()
			t.animator.Add(entry.Animation)
			entry.started = true
		}
	}

	// Update animator
	t.animator.Update(deltaTime)

	// Check if all entries have been added and all animations are complete
	allAdded := true
	allComplete := true
	for _, entry := range t.entries {
		if !entry.started {
			allAdded = false
		}
		if entry.Animation.IsRunning() {
			allComplete = false
		}
	}

	if allAdded && allComplete {
		t.running = false
		if t.onComplete != nil {
			t.onComplete()
		}
		return true
	}
	return false
}

// Count returns the number of timeline entries.
func (t *Timeline) Count() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.entries)
}

// ============================================================
// Enhanced Timeline: property-based animation (ported from TypeScript core)
// ============================================================

// EasingName is a named easing function.
type EasingName string

const (
	EaseLinear           EasingName = "linear"
	EaseInQuad           EasingName = "inQuad"
	EaseOutQuad          EasingName = "outQuad"
	EaseInOutQuad        EasingName = "inOutQuad"
	EaseInExpoName       EasingName = "inExpo"
	EaseOutExpoName      EasingName = "outExpo"
	EaseInOutSineName    EasingName = "inOutSine"
	EaseOutBounceName    EasingName = "outBounce"
	EaseInBounceName     EasingName = "inBounce"
	EaseInCircName       EasingName = "inCirc"
	EaseOutCircName      EasingName = "outCirc"
	EaseInOutCircName    EasingName = "inOutCirc"
	EaseInBackName       EasingName = "inBack"
	EaseOutBackName      EasingName = "outBack"
	EaseInOutBackName    EasingName = "inOutBack"
	EaseInElasticName    EasingName = "inElastic"
	EaseOutElasticName   EasingName = "outElastic"
	EaseInOutElasticName EasingName = "inOutElastic"
)

// easingByName maps easing names to easing functions.
var easingByName = map[EasingName]EasingFunc{
	EaseLinear:           Linear,
	EaseInQuad:           EaseIn,
	EaseOutQuad:          EaseOut,
	EaseInOutQuad:        EaseInOut,
	EaseInExpoName:       EaseInExpo,
	EaseOutExpoName:      EaseOutExpo,
	EaseInOutSineName:    EaseInOutSine,
	EaseOutBounceName:    EaseOutBounce,
	EaseInBounceName:     EaseInBounce,
	EaseInCircName:       EaseInCirc,
	EaseOutCircName:      EaseOutCirc,
	EaseInOutCircName:    EaseInOutCirc,
	EaseInBackName:       EaseInBack,
	EaseOutBackName:      EaseOutBack,
	EaseInOutBackName:    EaseInOutBack,
	EaseInElasticName:    EaseInElastic,
	EaseOutElasticName:   EaseOutElastic,
	EaseInOutElasticName: EaseInOutElastic,
}

// EasingForName returns the EasingFunc for a named easing, falling back to Linear.
func EasingForName(name EasingName) EasingFunc {
	if f, ok := easingByName[name]; ok {
		return f
	}
	return Linear
}

// EnhancedAnimationItem represents a single property-based animation entry in an EnhancedTimeline.
type EnhancedAnimationItem struct {
	Target     any
	Targets    []any              // multiple targets (TypeScript-compatible); if set, Target is ignored
	Properties map[string]float64 // end values
	Duration   time.Duration
	Easing     EasingFunc
	StartTime  time.Duration
	Loop       int // 1 = once, >1 = finite loops, 0 = infinite
	LoopDelay  time.Duration
	Alternate  bool
	Once       bool

	OnStart    func()
	OnUpdate   func(value float64, target any, progress float64)
	OnComplete func()
	OnLoop     func()

	onceCb func()

	// Internal state
	initialValues []map[string]float64 // one per target
	started       bool
	completed     bool
	currentLoop   int
}

// EnhancedTimeline manages a sequence of property-based animations with easing,
// looping, alternating, callbacks, and sub-timeline syncing.
type EnhancedTimeline struct {
	mu           sync.Mutex
	items        []*EnhancedAnimationItem
	subTimelines []*EnhancedTimelineItem
	currentTime  time.Duration
	isPlaying    bool
	isComplete   bool
	duration     time.Duration
	loop         bool
	frameRate    time.Duration
	autoplay     bool
	onComplete   func()
	onPause      func()

	nextListenerID int
	listeners      map[int]func(*EnhancedTimeline)
}

// EnhancedTimelineItem wraps a sub-timeline within a parent timeline.
type EnhancedTimelineItem struct {
	Timeline  *EnhancedTimeline
	StartTime time.Duration
	started   bool
}

// NewEnhancedTimeline creates a new EnhancedTimeline.
func NewEnhancedTimeline(opts ...EnhancedTimelineOption) *EnhancedTimeline {
	t := &EnhancedTimeline{
		duration:     time.Second,
		loop:         false,
		frameRate:    time.Second / 60,
		autoplay:     true,
		items:        make([]*EnhancedAnimationItem, 0),
		subTimelines: make([]*EnhancedTimelineItem, 0),
		listeners:    make(map[int]func(*EnhancedTimeline)),
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

// EnhancedTimelineOption configures an EnhancedTimeline.
type EnhancedTimelineOption func(*EnhancedTimeline)

// EnhancedTimelineDuration sets the timeline duration.
func EnhancedTimelineDuration(d time.Duration) EnhancedTimelineOption {
	return func(t *EnhancedTimeline) { t.duration = d }
}

// EnhancedTimelineLoop sets whether the timeline loops.
func EnhancedTimelineLoop(loop bool) EnhancedTimelineOption {
	return func(t *EnhancedTimeline) { t.loop = loop }
}

// EnhancedTimelineFrameRate sets the frame rate.
func EnhancedTimelineFrameRate(fps int) EnhancedTimelineOption {
	return func(t *EnhancedTimeline) { t.frameRate = time.Second / time.Duration(fps) }
}

// EnhancedTimelineAutoplay sets whether the timeline autoplay.
func EnhancedTimelineAutoplay(autoplay bool) EnhancedTimelineOption {
	return func(t *EnhancedTimeline) { t.autoplay = autoplay }
}

// EnhancedTimelineOnComplete sets the completion callback.
func EnhancedTimelineOnComplete(cb func()) EnhancedTimelineOption {
	return func(t *EnhancedTimeline) { t.onComplete = cb }
}

// EnhancedTimelineOnPause sets the pause callback.
func EnhancedTimelineOnPause(cb func()) EnhancedTimelineOption {
	return func(t *EnhancedTimeline) { t.onPause = cb }
}

// AddProperty animates properties on a target over time.
func (t *EnhancedTimeline) AddProperty(target any, endValues map[string]float64, duration time.Duration, startTime time.Duration, opts ...EnhancedAnimationOption) *EnhancedTimeline {
	item := &EnhancedAnimationItem{
		Target:      target,
		Targets:     []any{target},
		Properties:  endValues,
		Duration:    duration,
		Easing:      Linear,
		StartTime:   startTime,
		Loop:        1,
		Once:        false,
		started:     false,
		completed:   false,
		currentLoop: 0,
	}
	for _, opt := range opts {
		opt(item)
	}
	t.mu.Lock()
	t.items = append(t.items, item)
	t.mu.Unlock()
	return t
}

// Once schedules a property animation to run once (removed after completion),
// starting at the timeline's current time.
func (t *EnhancedTimeline) Once(target any, endValues map[string]float64, duration time.Duration) *EnhancedTimeline {
	t.mu.Lock()
	item := &EnhancedAnimationItem{
		Target:     target,
		Targets:    []any{target},
		Properties: endValues,
		Duration:   duration,
		Easing:     Linear,
		StartTime:  t.currentTime,
		Loop:       1,
		Once:       true,
		started:    false,
		completed:  false,
	}
	t.items = append(t.items, item)
	t.mu.Unlock()
	return t
}

// EnhancedAnimationOption configures an EnhancedAnimationItem.
type EnhancedAnimationOption func(*EnhancedAnimationItem)

func EnhancedAnimationEasing(e EasingFunc) EnhancedAnimationOption {
	return func(a *EnhancedAnimationItem) { a.Easing = e }
}

// EnhancedAnimationLoop sets loop count (1=once, >1 finite, 0 infinite).
func EnhancedAnimationLoop(n int) EnhancedAnimationOption {
	return func(a *EnhancedAnimationItem) { a.Loop = n }
}

// EnhancedAnimationLoopDelay sets delay between loops.
func EnhancedAnimationLoopDelay(d time.Duration) EnhancedAnimationOption {
	return func(a *EnhancedAnimationItem) { a.LoopDelay = d }
}

// EnhancedAnimationAlternate enables ping-pong looping.
func EnhancedAnimationAlternate(alt bool) EnhancedAnimationOption {
	return func(a *EnhancedAnimationItem) { a.Alternate = alt }
}

// EnhancedAnimationOnStart sets the start callback.
func EnhancedAnimationOnStart(cb func()) EnhancedAnimationOption {
	return func(a *EnhancedAnimationItem) { a.OnStart = cb }
}

// EnhancedAnimationOnUpdate sets the per-frame update callback.
func EnhancedAnimationOnUpdate(cb func(value float64, target any, progress float64)) EnhancedAnimationOption {
	return func(a *EnhancedAnimationItem) { a.OnUpdate = cb }
}

// EnhancedAnimationOnComplete sets the completion callback.
func EnhancedAnimationOnComplete(cb func()) EnhancedAnimationOption {
	return func(a *EnhancedAnimationItem) { a.OnComplete = cb }
}

// EnhancedAnimationOnLoop sets the loop callback.
func EnhancedAnimationOnLoop(cb func()) EnhancedAnimationOption {
	return func(a *EnhancedAnimationItem) { a.OnLoop = cb }
}

// EnhancedAnimationOnce marks the animation to run once and be removed.
func EnhancedAnimationOnce() EnhancedAnimationOption {
	return func(a *EnhancedAnimationItem) { a.Once = true }
}

// Call schedules a callback at a given time.
func (t *EnhancedTimeline) Call(cb func(), startTime time.Duration) *EnhancedTimeline {
	t.mu.Lock()
	t.items = append(t.items, &EnhancedAnimationItem{
		StartTime: startTime,
		Duration:  0,
		onceCb:    cb,
		started:   false,
		completed: false,
	})
	t.mu.Unlock()
	return t
}

// Sync adds a sub-timeline that plays in sync with this timeline.
func (t *EnhancedTimeline) Sync(sub *EnhancedTimeline, startTime time.Duration) *EnhancedTimeline {
	t.mu.Lock()
	t.subTimelines = append(t.subTimelines, &EnhancedTimelineItem{
		Timeline:  sub,
		StartTime: startTime,
		started:   false,
	})
	t.mu.Unlock()
	return t
}

// Play starts the timeline.
func (t *EnhancedTimeline) Play() *EnhancedTimeline {
	t.mu.Lock()
	if t.isComplete {
		t.Restart()
		t.mu.Unlock()
		return t
	}
	for _, st := range t.subTimelines {
		if st.started {
			st.Timeline.Play()
		}
	}
	t.isPlaying = true
	t.notifyStateChange()
	t.mu.Unlock()
	return t
}

// Pause pauses the timeline.
func (t *EnhancedTimeline) Pause() *EnhancedTimeline {
	t.mu.Lock()
	for _, st := range t.subTimelines {
		st.Timeline.Pause()
	}
	t.isPlaying = false
	if t.onPause != nil {
		t.onPause()
	}
	t.notifyStateChange()
	t.mu.Unlock()
	return t
}

// Restart restarts the timeline from the beginning.
func (t *EnhancedTimeline) Restart() *EnhancedTimeline {
	t.mu.Lock()
	t.isComplete = false
	t.currentTime = 0
	t.isPlaying = true
	t.resetItems()
	t.notifyStateChange()
	t.mu.Unlock()
	return t
}

// ResetItems resets all animation items to their initial state.
func (t *EnhancedTimeline) ResetItems() {
	t.mu.Lock()
	t.resetItems()
	t.mu.Unlock()
}

func (t *EnhancedTimeline) resetItems() {
	for _, item := range t.items {
		if item.Duration == 0 && item.onceCb != nil {
			// callback item
			continue
		}
		item.completed = false
		item.started = false
		item.currentLoop = 0
		item.initialValues = nil
	}
	for _, st := range t.subTimelines {
		st.started = false
		if st.Timeline != nil {
			st.Timeline.Restart()
			st.Timeline.Pause()
		}
	}
}

// AddStateChangeListener registers a listener for playback state changes.
// Returns an ID that can be used with RemoveStateChangeListener.
func (t *EnhancedTimeline) AddStateChangeListener(listener func(*EnhancedTimeline)) int {
	t.mu.Lock()
	t.nextListenerID++
	id := t.nextListenerID
	t.listeners[id] = listener
	t.mu.Unlock()
	return id
}

// RemoveStateChangeListener removes a previously registered listener by ID.
func (t *EnhancedTimeline) RemoveStateChangeListener(id int) {
	t.mu.Lock()
	delete(t.listeners, id)
	t.mu.Unlock()
}

func (t *EnhancedTimeline) notifyStateChange() {
	// Caller must hold t.mu.
	listeners := make([]func(*EnhancedTimeline), 0, len(t.listeners))
	for _, l := range t.listeners {
		listeners = append(listeners, l)
	}
	for _, l := range listeners {
		l(t)
	}
}

// Update advances the timeline by deltaTime.
func (t *EnhancedTimeline) Update(deltaTime time.Duration) {
	t.mu.Lock()
	for _, st := range t.subTimelines {
		evaluateSubTimeline(st, t.currentTime+deltaTime, deltaTime)
	}

	if !t.isPlaying {
		t.mu.Unlock()
		return
	}

	t.currentTime += deltaTime

	for _, item := range t.items {
		evaluateEnhancedItem(item, t.currentTime, deltaTime)
	}

	// Remove completed "once" items
	for i := len(t.items) - 1; i >= 0; i-- {
		if t.items[i].Once && t.items[i].completed {
			t.items = append(t.items[:i], t.items[i+1:]...)
		}
	}

	if t.loop && t.currentTime >= t.duration {
		overshoot := t.currentTime % t.duration
		t.resetItems()
		t.currentTime = 0
		if overshoot > 0 {
			t.mu.Unlock()
			t.Update(overshoot)
			t.mu.Lock()
			return
		}
	} else if !t.loop && t.currentTime >= t.duration {
		t.currentTime = t.duration
		t.isPlaying = false
		t.isComplete = true
		if t.onComplete != nil {
			t.onComplete()
		}
		t.notifyStateChange()
	}
	t.mu.Unlock()
}

func evaluateSubTimeline(st *EnhancedTimelineItem, timelineTime time.Duration, deltaTime time.Duration) {
	if st.Timeline == nil {
		return
	}
	if timelineTime < st.StartTime {
		return
	}
	if !st.started {
		st.started = true
		st.Timeline.Play()
		overshoot := timelineTime - st.StartTime
		st.Timeline.Update(overshoot)
		return
	}
	st.Timeline.Update(deltaTime)
}

func evaluateEnhancedItem(item *EnhancedAnimationItem, timelineTime time.Duration, deltaTime time.Duration) {
	if item.Duration == 0 && item.onceCb != nil {
		if !item.started && timelineTime >= item.StartTime {
			item.started = true
			item.onceCb()
			item.completed = true
		}
		return
	}

	if timelineTime < item.StartTime {
		return
	}

	animationTime := timelineTime - item.StartTime
	duration := item.Duration

	if timelineTime >= item.StartTime && !item.started {
		captureInitialValues(item)
		if item.OnStart != nil {
			item.OnStart()
		}
		item.started = true
	}

	if duration == 0 {
		if !item.completed {
			applyEnhancedProgress(item, 1.0, false, timelineTime, deltaTime)
			if item.OnComplete != nil {
				item.OnComplete()
			}
			item.completed = true
		}
		return
	}

	maxLoops := 1
	if item.Loop == 0 {
		maxLoops = -1 // infinite
	} else if item.Loop > 1 {
		maxLoops = item.Loop
	}
	loopDelay := item.LoopDelay
	cycleTime := duration + loopDelay

	var currentCycle int
	var timeInCycle time.Duration
	if maxLoops == -1 {
		currentCycle = int(animationTime / cycleTime)
		timeInCycle = animationTime % cycleTime
	} else {
		currentCycle = int(animationTime / cycleTime)
		timeInCycle = animationTime % cycleTime
	}

	if item.OnLoop != nil && item.currentLoop != 0 && currentCycle > item.currentLoop && (maxLoops == -1 || currentCycle < maxLoops) {
		item.OnLoop()
	}
	item.currentLoop = currentCycle

	if item.OnComplete != nil && !item.completed && (maxLoops == -1 || currentCycle == maxLoops-1) && timeInCycle >= duration {
		finalReversed := item.Alternate && currentCycle%2 == 1
		applyEnhancedProgress(item, 1.0, finalReversed, timelineTime, deltaTime)
		item.OnComplete()
		item.completed = true
		return
	}

	if maxLoops != -1 && currentCycle >= maxLoops {
		if !item.completed {
			finalReversed := item.Alternate && (maxLoops-1)%2 == 1
			applyEnhancedProgress(item, 1.0, finalReversed, timelineTime, deltaTime)
			if item.OnComplete != nil {
				item.OnComplete()
			}
			item.completed = true
		}
		return
	}

	if timeInCycle == 0 && animationTime > 0 && (maxLoops == -1 || currentCycle < maxLoops) {
		currentCycle--
		timeInCycle = cycleTime
	}

	if timeInCycle >= duration {
		isReversed := item.Alternate && currentCycle%2 == 1
		applyEnhancedProgress(item, 1.0, isReversed, timelineTime, deltaTime)
		// Mark completed if this is the final cycle (non-looping or last finite loop).
		if maxLoops != -1 && currentCycle == maxLoops-1 {
			item.completed = true
		}
		return
	}

	progress := float64(timeInCycle) / float64(duration)
	isReversed := item.Alternate && currentCycle%2 == 1
	applyEnhancedProgress(item, progress, isReversed, timelineTime, deltaTime)
}

// targets returns the list of animation targets (honoring Targets if set, else Target).
func (item *EnhancedAnimationItem) targets() []any {
	if len(item.Targets) > 0 {
		return item.Targets
	}
	if item.Target != nil {
		return []any{item.Target}
	}
	return nil
}

func captureInitialValues(item *EnhancedAnimationItem) {
	if len(item.Properties) == 0 {
		return
	}
	if item.initialValues != nil {
		return
	}
	targets := item.targets()
	item.initialValues = make([]map[string]float64, len(targets))
	for i, target := range targets {
		item.initialValues[i] = make(map[string]float64)
		for key := range item.Properties {
			item.initialValues[i][key] = readFloatProperty(target, key)
		}
	}
}

func readFloatProperty(target any, key string) float64 {
	switch v := target.(type) {
	case map[string]float64:
		return v[key]
	case map[string]any:
		if f, ok := v[key].(float64); ok {
			return f
		}
	case *float64:
		if key == "" {
			return *v
		}
	case struct {
		X float64
	}:
		if key == "x" || key == "X" {
			return v.X
		}
	case struct {
		Y float64
	}:
		if key == "y" || key == "Y" {
			return v.Y
		}
	case struct {
		X, Y float64
	}:
		if key == "x" || key == "X" {
			return v.X
		}
		if key == "y" || key == "Y" {
			return v.Y
		}
	case struct {
		X, Y, Z float64
	}:
		switch key {
		case "x", "X":
			return v.X
		case "y", "Y":
			return v.Y
		case "z", "Z":
			return v.Z
		}
	}
	// Fallback: reflect over exported float64 fields matching the key (case-insensitive).
	rv := reflect.ValueOf(target)
	if rv.Kind() == reflect.Ptr && !rv.IsNil() {
		rv = rv.Elem()
	}
	if rv.Kind() == reflect.Struct {
		rt := rv.Type()
		for i := 0; i < rt.NumField(); i++ {
			field := rt.Field(i)
			if field.Type.Kind() != reflect.Float64 {
				continue
			}
			name := field.Name
			if stringsEqualFold(name, key) {
				return rv.Field(i).Float()
			}
		}
	}
	return 0
}

func stringsEqualFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca := a[i]
		cb := b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 32
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 32
		}
		if ca != cb {
			return false
		}
	}
	return true
}

func setFloatProperty(target any, key string, value float64) bool {
	switch v := target.(type) {
	case map[string]float64:
		v[key] = value
		return true
	case map[string]any:
		v[key] = value
		return true
	case *struct {
		X float64
	}:
		if key == "x" || key == "X" {
			v.X = value
			return true
		}
	case *struct {
		Y float64
	}:
		if key == "y" || key == "Y" {
			v.Y = value
			return true
		}
	case *struct {
		X, Y float64
	}:
		if key == "x" || key == "X" {
			v.X = value
			return true
		}
		if key == "y" || key == "Y" {
			v.Y = value
			return true
		}
	case *struct {
		X, Y, Z float64
	}:
		switch key {
		case "x", "X":
			v.X = value
			return true
		case "y", "Y":
			v.Y = value
			return true
		case "z", "Z":
			v.Z = value
			return true
		}
	}
	// Fallback: reflect over exported float64 fields matching the key (case-insensitive).
	rv := reflect.ValueOf(target)
	if rv.Kind() == reflect.Ptr && !rv.IsNil() {
		rv = rv.Elem()
	}
	if rv.Kind() == reflect.Struct && rv.CanAddr() {
		rt := rv.Type()
		for i := 0; i < rt.NumField(); i++ {
			field := rt.Field(i)
			if field.Type.Kind() != reflect.Float64 {
				continue
			}
			name := field.Name
			if stringsEqualFold(name, key) {
				rv.Field(i).SetFloat(value)
				return true
			}
		}
	}
	return false
}

func applyEnhancedProgress(item *EnhancedAnimationItem, progress float64, reversed bool, timelineTime time.Duration, deltaTime time.Duration) {
	if len(item.Properties) == 0 {
		return
	}

	if item.initialValues == nil {
		captureInitialValues(item)
	}

	eased := item.Easing(progress)
	if eased > 1 {
		eased = 1
	}
	if eased < 0 {
		eased = 0
	}
	finalProgress := eased
	if reversed {
		finalProgress = 1 - eased
	}

	targets := item.targets()
	if len(targets) > 0 && item.OnUpdate != nil {
		item.OnUpdate(finalProgress, targets[0], finalProgress)
	}

	for i, target := range targets {
		if i >= len(item.initialValues) {
			continue
		}
		initialValues := item.initialValues[i]
		for key, endValue := range item.Properties {
			startValue := initialValues[key]
			newValue := startValue + (endValue-startValue)*finalProgress
			setFloatProperty(target, key, newValue)
		}
	}
}

// GetCurrentTime returns the current timeline time.
func (t *EnhancedTimeline) GetCurrentTime() time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.currentTime
}

// IsPlaying returns true if the timeline is playing.
func (t *EnhancedTimeline) IsPlaying() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.isPlaying
}

// IsTimelineComplete returns true if the timeline has completed.
func (t *EnhancedTimeline) IsTimelineComplete() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.isComplete
}

// Count returns the number of animation items in the timeline.
func (t *EnhancedTimeline) Count() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.items)
}

// itoa converts an integer to a string.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var result string
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}

// TimelineEngine manages a set of EnhancedTimelines, advancing them together on a shared clock.
// This is a Go port of the TypeScript TimelineEngine, without the renderer frame-callback integration.
type TimelineEngine struct {
	mu        sync.Mutex
	timelines map[*EnhancedTimeline]struct{}
	tickFps   time.Duration
}

// NewTimelineEngine creates a new TimelineEngine.
func NewTimelineEngine() *TimelineEngine {
	return &TimelineEngine{
		timelines: make(map[*EnhancedTimeline]struct{}),
	}
}

// Register adds a timeline to the engine.
func (e *TimelineEngine) Register(tl *EnhancedTimeline) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.timelines[tl] = struct{}{}
}

// Unregister removes a timeline from the engine.
func (e *TimelineEngine) Unregister(tl *EnhancedTimeline) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.timelines, tl)
}

// Update advances all registered timelines by deltaTime.
func (e *TimelineEngine) Update(deltaTime time.Duration) {
	e.mu.Lock()
	timelines := make([]*EnhancedTimeline, 0, len(e.timelines))
	for tl := range e.timelines {
		timelines = append(timelines, tl)
	}
	e.mu.Unlock()
	for _, tl := range timelines {
		tl.Update(deltaTime)
	}
}

// Clear removes all registered timelines.
func (e *TimelineEngine) Clear() {
	e.mu.Lock()
	e.timelines = make(map[*EnhancedTimeline]struct{})
	e.mu.Unlock()
}

// Count returns the number of registered timelines.
func (e *TimelineEngine) Count() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return len(e.timelines)
}

// Ensure unused imports are referenced.
var _ = math.Pi
var _ = sync.Mutex{}
var _ = time.Second
