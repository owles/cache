package cache

import (
	"time"
)

const (
	SizeNone = iota
	SizeKB   = 1 << (10 * iota)
	SizeMB
	SizeGB
	SizeTB
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
