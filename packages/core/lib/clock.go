package lib

import (
	"sync"
	"time"
)

// Clock is a simple system clock
type Clock struct {
	mu sync.Mutex
	start time.Time
}

func NewClock() *Clock {
	return &Clock{start: time.Now()}
}

func (c *Clock) Now() float64 {
	return float64(time.Since(c.start).Nanoseconds()) / 1e9
}

func (c *Clock) Since(start float64) float64 {
	return c.Now() - start
}
