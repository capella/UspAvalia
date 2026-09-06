// Package cache provides small in-memory caches for expensive computations.
package cache

import (
	"sync"
	"time"
)

// TTL memoizes a single value of type T for a fixed duration. The zero
// value is ready to use and safe for concurrent access.
type TTL[T any] struct {
	mu         sync.RWMutex
	value      T
	expiration time.Time
}

// Get returns the cached value and true while it has not expired.
func (c *TTL[T]) Get() (T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if time.Now().Before(c.expiration) {
		return c.value, true
	}
	var zero T
	return zero, false
}

// Set stores value for ttl.
func (c *TTL[T]) Set(value T, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value = value
	c.expiration = time.Now().Add(ttl)
}

// Invalidate drops the cached value so the next Get misses.
func (c *TTL[T]) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	var zero T
	c.value = zero
	c.expiration = time.Time{}
}

// GetOrLoad returns the cached value, or calls load, stores its result for
// ttl and returns it. load runs outside the lock, so concurrent misses may
// load more than once; the last result wins.
func (c *TTL[T]) GetOrLoad(ttl time.Duration, load func() T) T {
	if value, ok := c.Get(); ok {
		return value
	}
	value := load()
	c.Set(value, ttl)
	return value
}
