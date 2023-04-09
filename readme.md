# Go cache

Helpers for redis and memory (not memcache) cache.

## Install and usage

```
go get github.com/owles/cache
```

```go 
func main() {
    redisCache, err := cache.NewRedisCache(
        context.Background(),
        "localhost:6379",
        "",
        0,
    )

    if err != nil {
        log.Fatal(err)
    }

    var err error
    var val any

    // Store value with unlimited time
    if err = redisCache.Set("no.expire.key", "value", 0); err != nil {
        log.fatal(err)		
    }
    val = redisCache.Get("no.expire.key", "value", "default")
	fmt.Println(val)

    // Store value with expire time
    if err = redisCache.Set("1sec.expire.key", "value", time.Second); err != nil {
        log.fatal(err)		
    }
    val = redisCache.Get("no.expire.key", "value", "default")
    fmt.Println(val)

    // Store value by callback 
    val, _ = redisCache.RememberForever("func.no.expire", func() any {
        return "value"
    })
    fmt.Println(val)

    // Forget key
    redisCache.Forget("no.expire.key")

    // Clear cache
    redisCache.Flush()
}
```