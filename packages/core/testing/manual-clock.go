package testing

import (
)

// ManualClock is a manually-controlled clock for testing timer-driven behavior.
// It implements the core.Clock interface.
type ManualClock struct {
	time       float64
	nextID     int
	nextOrder  int
	timers     map[int]*scheduledTimer
}

// NewManualClock creates a new ManualClock starting at time 0.
func NewManualClock() *ManualClock {
	return &ManualClock{
		timers: make(map[int]*scheduledTimer),
	}
}

// Now returns the current mock time in milliseconds.
func (c *ManualClock) Now() float64 {
	return c.time
}

// Since returns the elapsed time since the given start time in milliseconds.
func (c *ManualClock) Since(start float64) float64 {
	return c.time - start
}

// PendingTimerCount returns the number of scheduled timers that have not fired or been cleared.
func (c *ManualClock) PendingTimerCount() int {
	return len(c.timers)
}

// SetTime sets the clock time. If the target is ahead, advances the clock (firing timers).
// If the target is behind, sets the clock back without firing timers.
func (c *ManualClock) SetTime(time float64) {
	target := int(time)
	if target >= int(c.time) {
		c.Advance(time - c.time)
		return
	}
	c.time = time
}

// SetTimeout schedules a one-shot timer and returns a handle.
func (c *ManualClock) SetTimeout(fn func(), delayMs float64) int {
	return c.schedule(fn, delayMs, false)
}

// ClearTimeout clears a timer by handle.
func (c *ManualClock) ClearTimeout(handle int) {
	delete(c.timers, handle)
}

// SetInterval schedules a repeating timer and returns a handle.
func (c *ManualClock) SetInterval(fn func(), delayMs float64) int {
	return c.schedule(fn, delayMs, true)
}

// ClearInterval is an alias for ClearTimeout.
func (c *ManualClock) ClearInterval(handle int) {
	c.ClearTimeout(handle)
}

// Advance advances the clock by delayMs, firing any timers whose fire time is reached.
func (c *ManualClock) Advance(delayMs float64) {
	target := c.time + max(0, delayMs)

	for {
		next := c.peekNextTimer()
		if next == nil || next.fireAt > target {
			break
		}
		delete(c.timers, next.id)
		c.time = next.fireAt
		next.fn()

		if next.repeat {
			if _, ok := c.timers[next.id]; !ok {
				c.timers[next.id] = &scheduledTimer{
					id:      next.id,
					fireAt:  c.time + next.delayMs,
					order:   c.nextOrder,
					delayMs: next.delayMs,
					repeat:  true,
					fn:      next.fn,
				}
				c.nextOrder++
			}
		}
	}

	c.time = target
}

// RunAll runs all pending timers until none remain.
func (c *ManualClock) RunAll() {
	for {
		next := c.peekNextTimer()
		if next == nil {
			return
		}
		c.Advance(next.fireAt - c.time)
	}
}

func (c *ManualClock) schedule(fn func(), delayMs float64, repeat bool) int {
	id := c.nextID
	c.nextID++
	normalized := max(0, delayMs)
	c.timers[id] = &scheduledTimer{
		id:      id,
		fireAt:  c.time + normalized,
		order:   c.nextOrder,
		delayMs: normalized,
		repeat:  repeat,
		fn:      fn,
	}
	c.nextOrder++
	return id
}

func (c *ManualClock) peekNextTimer() *scheduledTimer {
	var next *scheduledTimer
	for _, timer := range c.timers {
		if next == nil || compareTimers(timer, next) < 0 {
			next = timer
		}
	}
	return next
}

func compareTimers(left, right *scheduledTimer) int {
	if left.fireAt != right.fireAt {
		if left.fireAt < right.fireAt {
			return -1
		}
		return 1
	}
	return left.order - right.order
}

type scheduledTimer struct {
	id      int
	fireAt  float64
	order   int
	delayMs float64
	repeat  bool
	fn      func()
}


