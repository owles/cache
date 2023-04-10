package cache

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"time"
)

type Redis struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisCache(ctx context.Context, addr, user, pass string, db int) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Username: user,
		Password: pass,
		DB:       db,
	})

	if _, err := client.Ping(context.Background()).Result(); err != nil {
		return nil, errors.New("redis connection error")
	}

	return &Redis{
		client: client,
		ctx:    ctx,
	}, nil
}

func (c *Redis) Name() string {
	return "redis"
}

func (c *Redis) Size() int64 {
	return c.client.DBSize(c.ctx).Val()
}

func (c *Redis) Set(key string, val any, t time.Duration) error {
	ok, err := c.client.SetNX(c.ctx, key, val, t).Result()
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("can't set cache value")
	}

	return nil
}

func (c *Redis) Get(key string, def ...any) any {
	val, err := c.client.Get(c.ctx, key).Result()
	if err != nil {
		if len(def) == 0 {
			return nil
		}

		return def[0]
	}

	return val
}

func (c *Redis) Remember(key string, callback func() any, t time.Duration) (any, error) {
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

func (c *Redis) RememberForever(key string, callback func() any) (any, error) {
	return c.Remember(key, callback, 0)
}

func (c *Redis) Flush() bool {
	ok, err := c.client.FlushAll(c.ctx).Result()
	if err != nil || ok != "OK" {
		return false
	}

	return true
}

func (c *Redis) Forget(key string) bool {
	_, err := c.client.Del(c.ctx, key).Result()
	return err == nil
}
