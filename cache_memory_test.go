package cache

import (
	"context"
	"testing"
	"time"
)

var cache = NewMemoryCache(context.Background(), 0)

func TestMemoryGet(t *testing.T) {
	cache.Set("foo", "bar", 0)

	if cache.Get("foo") != "bar" {
		t.Error("Cache get failed: foo -> bar")
	}
}

func TestMemoryGetDefault(t *testing.T) {
	if cache.Get("foo.bar", "baz") != "baz" {
		t.Error("Cache get default failed: foo.bar -> baz")
	}
}

func TestMemoryGetNil(t *testing.T) {
	if cache.Get("foo.bar.baz") != nil {
		t.Error("Cache get default failed: foo.bar.baz -> nil")
	}
}

func TestMemoryTime(t *testing.T) {
	cache.Set("foo.bar", "baz", time.Second)
	time.Sleep(time.Second * 2)

	if cache.Get("foo.bar") != nil {
		t.Error("Cache get life time failed: foo.bar -> nil")
	}
}

func TestMemoryTimeUnlimited(t *testing.T) {
	cache.Set("foo.bar", "baz", 0)

	if cache.Get("foo.bar") != "baz" {
		t.Error("Cache get life time failed: foo.bar -> baz")
	}
}
