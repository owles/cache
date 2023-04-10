package cache

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"time"
	"unsafe"
)

type memNode struct {
	value    any
	expireAt int64
	size     int64
}
type memTable map[string]*memNode

type Memory struct {
	mu       sync.RWMutex
	ctx      context.Context
	table    memTable
	memAlloc int64
	memAllow int64
}

const memNodeSize = int64(unsafe.Sizeof(memNode{}))

func sizeOf(v reflect.Value) int64 {
	size := int64(0)

	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			size += sizeOf(v.Index(i))
		}
		size += int64(v.Cap()-v.Len()) * int64(v.Type().Elem().Size())

	case reflect.Map:
		for _, key := range v.MapKeys() {
			size += sizeOf(key) + sizeOf(v.MapIndex(key))
		}
		size += int64(v.Type().Size())

	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			size += sizeOf(v.Field(i))
		}
		size += int64(v.Type().Size())

	case reflect.String:
		size += int64(v.Len() + int(v.Type().Size()))

	case reflect.Ptr:
		if v.IsNil() {
			size += int64(reflect.New(v.Type()).Type().Size())
			break
		}
		if sizeInd := sizeOf(reflect.Indirect(v)); sizeInd != 0 {
			size += sizeInd
			break
		}
		size += int64(v.Type().Size())

	case reflect.Interface:
		size += sizeOf(v.Elem()) + int64(v.Type().Size())

	default:
		size += int64(v.Type().Size())
	}

	return size
}

func (n memNode) IsExpired() bool {
	return n.expireAt != 0 && time.Now().UnixMicro() > n.expireAt
}

func (n memNode) Size() int64 {
	return n.size
}

func NewMemoryCache(ctx context.Context, memAllow int64) *Memory {
	cache := &Memory{
		ctx:      ctx,
		table:    make(memTable, 0),
		memAlloc: 0,
		memAllow: memAllow,
	}
	cache.garbageThread()

	return cache
}

func (c *Memory) Name() string {
	return "memory"
}

func (c *Memory) Size() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.memAlloc
}

func (c *Memory) remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.memAlloc -= c.table[key].Size()
	delete(c.table, key)
}

func (c *Memory) garbageCollect(flush bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, node := range c.table {
		if node.IsExpired() || flush {
			c.memAlloc -= c.table[key].Size()
			delete(c.table, key)
		}
	}
}

func (c *Memory) garbageThread() {
	go func() {
		for {
			select {
			case <-time.After(time.Minute):
				c.garbageCollect(false)
			case <-c.ctx.Done():
				break
			}
		}
	}()
}

func (c *Memory) Set(key string, val any, t time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	expireAt := int64(0)
	if t > 0 {
		expireAt = time.Now().Add(t).UnixMicro()
	}

	var size int64 = 0
	if c.memAllow > 0 {
		size = sizeOf(reflect.ValueOf(key)) + sizeOf(reflect.ValueOf(val)) + memNodeSize
		if c.memAlloc+size > c.memAllow {
			return errors.New("memory limit")
		}

		if _, ok := c.table[key]; ok {
			c.memAlloc -= c.table[key].Size()
		}
	}

	c.table[key] = &memNode{
		value:    val,
		expireAt: expireAt,
		size:     size,
	}

	c.memAlloc += size

	return nil
}

func (c *Memory) Get(key string, def ...any) any {
	c.mu.RLock()
	val, ok := c.table[key]

	defer func() {
		c.mu.RUnlock()
		if val != nil && val.IsExpired() {
			c.remove(key)
		}
	}()

	if !ok || val == nil || val.IsExpired() {
		if len(def) > 0 {
			return def[0]
		}
		return nil
	}

	return val.value
}

func (c *Memory) Remember(key string, callback func() any, t time.Duration) (any, error) {
	val := c.Get(key, nil)
	if val != nil {
		return val, nil
	}

	val = callback()

	err := c.Set(key, val, t)
	if err != nil {
		return nil, err
	}

	return val, nil
}

func (c *Memory) RememberForever(key string, callback func() any) (any, error) {
	return c.Remember(key, callback, 0)
}

func (c *Memory) Flush() bool {
	c.garbageCollect(true)
	return true
}

func (c *Memory) Forget(key string) bool {
	c.remove(key)
	return true
}
