// Package clock provides a logical clock for TTLs without syscalls.
package clock

import (
	"sync/atomic"
	"time"
)

type Clock struct {
	tick atomic.Uint64
	stop chan struct{}
}

func NewClock() *Clock {
	c := &Clock{
		stop: make(chan struct{}),
	}
	go c.tickGoroutine()
	return c
}

func (c *Clock) Now() uint64 {
	return c.tick.Load()
}

func (c *Clock) tickGoroutine() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.tick.Add(1)
		case <-c.stop:
			return
		}
	}
}

// Stop gracefully stops the clock's ticker goroutine.
func (c *Clock) Stop() {
	close(c.stop)
}
