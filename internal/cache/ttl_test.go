package cache

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestTTLGetOrLoad(t *testing.T) {
	var c TTL[int]
	loads := 0
	load := func() int { loads++; return loads }

	if _, ok := c.Get(); ok {
		t.Fatal("zero value must start empty")
	}
	if got := c.GetOrLoad(time.Minute, load); got != 1 {
		t.Fatalf("first GetOrLoad = %d, want 1", got)
	}
	if got := c.GetOrLoad(time.Minute, load); got != 1 || loads != 1 {
		t.Fatalf("second GetOrLoad = %d (loads %d), want cached 1", got, loads)
	}

	c.Invalidate()
	if got := c.GetOrLoad(time.Minute, load); got != 2 {
		t.Fatalf("GetOrLoad after Invalidate = %d, want 2", got)
	}
}

func TestTTLExpires(t *testing.T) {
	var c TTL[string]
	c.Set("v", -time.Second) // already expired
	if _, ok := c.Get(); ok {
		t.Fatal("expired value must not be returned")
	}
	c.Set("v", time.Minute)
	if got, ok := c.Get(); !ok || got != "v" {
		t.Fatalf("Get = %q, %v; want \"v\", true", got, ok)
	}
}

// Concurrent misses on an empty cache must share one load.
func TestTTLGetOrLoadLoadsOnceWhenEmpty(t *testing.T) {
	var c TTL[int]
	var loads atomic.Int32
	release := make(chan struct{})
	load := func() int {
		loads.Add(1)
		<-release
		return 7
	}

	const callers = 20
	results := make(chan int, callers)
	for i := 0; i < callers; i++ {
		go func() { results <- c.GetOrLoad(time.Minute, load) }()
	}
	// Give every caller time to reach the cache before the load finishes.
	time.Sleep(50 * time.Millisecond)
	close(release)

	for i := 0; i < callers; i++ {
		if got := <-results; got != 7 {
			t.Fatalf("GetOrLoad = %d, want 7", got)
		}
	}
	if n := loads.Load(); n != 1 {
		t.Fatalf("load ran %d times, want 1", n)
	}
}

// An expired value keeps being served while a single refresh runs.
func TestTTLGetOrLoadServesStaleWhileRefreshing(t *testing.T) {
	var c TTL[string]
	c.Set("old", -time.Second) // already expired

	var loads atomic.Int32
	started := make(chan struct{})
	release := make(chan struct{})
	load := func() string {
		if loads.Add(1) == 1 {
			close(started)
		}
		<-release
		return "new"
	}

	for i := 0; i < 10; i++ {
		if got := c.GetOrLoad(time.Minute, load); got != "old" {
			t.Fatalf("GetOrLoad during refresh = %q, want the stale \"old\"", got)
		}
	}
	<-started
	close(release)

	deadline := time.Now().Add(2 * time.Second)
	for {
		if got, ok := c.Get(); ok && got == "new" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("refreshed value never became visible")
		}
		time.Sleep(time.Millisecond)
	}
	if n := loads.Load(); n != 1 {
		t.Fatalf("load ran %d times, want 1", n)
	}
}

// A load that started before Invalidate must not bring its result back.
func TestTTLInvalidateDuringLoad(t *testing.T) {
	var c TTL[int]
	started := make(chan struct{})
	release := make(chan struct{})
	done := make(chan int)
	go func() {
		done <- c.GetOrLoad(time.Minute, func() int {
			close(started)
			<-release
			return 1
		})
	}()

	<-started
	c.Invalidate()
	close(release)
	if got := <-done; got != 1 {
		t.Fatalf("the loading caller got %d, want its own result 1", got)
	}
	if got := c.GetOrLoad(time.Minute, func() int { return 2 }); got != 2 {
		t.Fatalf("GetOrLoad after Invalidate = %d, want a fresh 2", got)
	}
}

// A panicking background refresh keeps the stale value and can be retried.
func TestTTLRefreshPanicKeepsStaleValue(t *testing.T) {
	var c TTL[string]
	c.Set("old", -time.Second)

	panicked := make(chan struct{})
	if got := c.GetOrLoad(time.Minute, func() string {
		defer close(panicked)
		panic("boom")
	}); got != "old" {
		t.Fatalf("GetOrLoad = %q, want the stale \"old\"", got)
	}
	<-panicked

	deadline := time.Now().Add(2 * time.Second)
	for {
		c.GetOrLoad(time.Minute, func() string { return "new" })
		if got, ok := c.Get(); ok && got == "new" {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("cache never recovered after a panicking refresh")
		}
		time.Sleep(time.Millisecond)
	}
}
