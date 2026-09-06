package cache

import (
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
