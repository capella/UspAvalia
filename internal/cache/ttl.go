// Package cache provides small in-memory caches for expensive computations.
package cache

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// TTL memoizes a single value of type T for a fixed duration. The zero
// value is ready to use and safe for concurrent access.
type TTL[T any] struct {
	mu         sync.Mutex
	value      T
	hasValue   bool
	expiration time.Time
	// loading is non-nil while a load is in flight and is closed when it ends.
	loading chan struct{}
	// generation changes on Set and Invalidate, so a load that started
	// before either of them does not overwrite the newer state.
	generation uint64
}

// Get returns the cached value and true while it has not expired.
func (c *TTL[T]) Get() (T, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.hasValue && time.Now().Before(c.expiration) {
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
	c.hasValue = true
	c.expiration = time.Now().Add(ttl)
	c.generation++
}

// Invalidate drops the cached value so the next Get misses.
func (c *TTL[T]) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	var zero T
	c.value = zero
	c.hasValue = false
	c.expiration = time.Time{}
	c.generation++
}

// GetOrLoad returns the cached value, loading it with load when needed. At
// most one load runs at a time:
//
//   - While the value is fresh it is returned as is.
//   - Once it has expired, callers keep getting the stale value and a single
//     background load replaces it, so nobody waits on a refresh.
//   - While there is no value at all (first use, or after Invalidate), one
//     caller loads and the concurrent ones wait for its result.
func (c *TTL[T]) GetOrLoad(ttl time.Duration, load func() T) T {
	for {
		c.mu.Lock()
		if c.hasValue {
			value := c.value
			if !time.Now().Before(c.expiration) && c.loading == nil {
				c.loading = make(chan struct{})
				go c.refresh(ttl, load, c.generation)
			}
			c.mu.Unlock()
			return value
		}
		if inFlight := c.loading; inFlight != nil {
			// Wait for the load in flight, then look again: it may have
			// panicked or been invalidated, leaving nothing to return.
			c.mu.Unlock()
			<-inFlight
			continue
		}
		c.loading = make(chan struct{})
		generation := c.generation
		c.mu.Unlock()
		return c.load(ttl, load, generation)
	}
}

// load runs load and stores its result unless the cache was Set or
// Invalidated in the meantime. The caller must have set c.loading.
func (c *TTL[T]) load(ttl time.Duration, load func() T, generation uint64) T {
	defer func() {
		c.mu.Lock()
		close(c.loading)
		c.loading = nil
		c.mu.Unlock()
	}()

	value := load()

	c.mu.Lock()
	if c.generation == generation {
		c.value = value
		c.hasValue = true
		c.expiration = time.Now().Add(ttl)
	}
	c.mu.Unlock()
	return value
}

// refresh is load for the background case: there is no request to recover
// a panic, so it is logged and the stale value stays until the next try.
func (c *TTL[T]) refresh(ttl time.Duration, load func() T, generation uint64) {
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("cache: background refresh panicked: %v", r)
		}
	}()
	c.load(ttl, load, generation)
}
