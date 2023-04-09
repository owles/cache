package cache

import (
	"time"
)

type Cache interface {
	Name() string
	Size() int64
	Set(key string, val any, t time.Duration) error
	Get(key string, def ...any) any
	Remember(key string, callback func() any, t time.Duration) (any, error)
	RememberForever(key string, callback func() any) (any, error)
	Forget(key string) bool
	Flush() bool
}

// ToDo - [-] Add cast to value ...
// ToDo - [-] Add memory cache ...
