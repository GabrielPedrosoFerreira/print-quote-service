// Package apptest provides deterministic test doubles for the application ports.
// They are simple fakes with real behavior, not mocks with scripted expectations,
// so tests assert outcomes instead of call sequences.
package apptest

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Clock is a manually controlled clock.
type Clock struct {
	mu  sync.Mutex
	now time.Time
}

func NewClock(start time.Time) *Clock { return &Clock{now: start} }

func (c *Clock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *Clock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

// SequentialIDs returns qt_1, qt_2, ... so assertions can use predictable IDs.
type SequentialIDs struct{ next atomic.Int64 }

func (s *SequentialIDs) NewID() string {
	return fmt.Sprintf("qt_%d", s.next.Add(1))
}
