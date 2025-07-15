# Echo HTTP middleware for caching
This library is Echo HTTP middleware that provides server-side application layer caching for REST APIs. 

> I tried to use this library for my personal project to cache API response using a redis standalone server, but it didn't work well. The library hasn't been updated in a while and it seems like it is not maintained so I forked from https://github.com/SporkHubr/echo-http-cache and modified the code. 

## Features

- Support different stores for caching:
  - Local memory (LRU, LFU, MRU, MFU algorithms)
  - Redis Standalone
  - Redis Cluster (NEW! ✨)
- Enable to set different expiration time for each API
- Enable to exclude any APIs not to cache

## Installation

`echo-http-cache` supports 2 last Go versions and requires a Go version with [modules](https://github.com/golang/go/wiki/Modules) support. So make sure to initialize a Go module as follows.

```bash
go mod init github.com/my/repo
```

And then install `kenshin579/echo-http-cache`

```bash
go get github.com/kenshin579/echo-http-cache
```

## Requirements

- Go 1.24 or higher
- Echo v4
- go-redis v8.11.5
- Redis 6.0+ for Redis Cluster support

## Quickstart

Here are three examples:

- Using local memory
- Using redis standalone server
- Using redis cluster (NEW!)

### Local Memory Example

```golang
package main

import (
	"net/http"
	"time"

	echoCacheMiddleware "github.com/kenshin579/echo-http-cache"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	e.Use(echoCacheMiddleware.CacheWithConfig(echoCacheMiddleware.CacheConfig{
		Store: echoCacheMiddleware.NewCacheMemoryStoreWithConfig(echoCacheMiddleware.CacheMemoryStoreConfig{
			Capacity:  5,
			Algorithm: echoCacheMiddleware.LFU,
		}),
		Expiration: 10 * time.Second,
	}))

	e.GET("/hello", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, Echo!")
	})

	e.Start(":8080")
}
```

### Redis Standalone Example

```golang
package main

import (
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	echoCacheMiddleware "github.com/kenshin579/echo-http-cache"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	e.Use(echoCacheMiddleware.CacheWithConfig(echoCacheMiddleware.CacheConfig{
		Store: echoCacheMiddleware.NewCacheRedisStoreWithConfig(redis.Options{
			Addr:     "localhost:6379",
			Password: "password",
		}),
		Expiration: 5 * time.Minute,
		IncludePathsWithExpiration: map[string]time.Duration{
			"/hello": 1 * time.Minute,
		},
		ExcludePaths: []string{"/ping"},
	}))

	e.GET("/hello", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, Echo!")
	})

	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	e.Start(":8080")
}
```

### Redis Cluster Example

```golang
package main

import (
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	echoCacheMiddleware "github.com/kenshin579/echo-http-cache"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	
	// Configure Redis Cluster store
	store := echoCacheMiddleware.NewCacheRedisClusterStoreWithConfig(redis.ClusterOptions{
		Addrs: []string{
			"localhost:17000",
			"localhost:17001", 
			"localhost:17002",
			"localhost:17003",
			"localhost:17004",
			"localhost:17005",
		},
	})
	
	e.Use(echoCacheMiddleware.CacheWithConfig(echoCacheMiddleware.CacheConfig{
		Store:      store,
		Expiration: 5 * time.Minute,
	}))

	e.GET("/api/data", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"data": "This response is cached in Redis Cluster",
			"time": time.Now().Format(time.RFC3339),
		})
	})

	e.Start(":8080")
}
```

### Redis Cluster Cache

`echo-http-cache` now supports true Redis Cluster for distributed caching across multiple Redis nodes.

**Note**: If you're using Docker for local development, please be aware of [known networking limitations](docs/issue-23/known-issues.md) that may affect Redis Cluster connectivity. These issues do not occur in production environments with properly configured Redis Cluster.

#### Docker Setup for Testing

A `docker-compose.yml` file is provided for local testing with Redis Cluster:

```bash
# Start Redis Cluster
docker-compose up -d

# Stop Redis Cluster
docker-compose down
```

The Docker setup includes:
- 6 Redis nodes (3 masters, 3 replicas)
- Ports mapped from 17000-17005 (to avoid conflicts with local Redis)
- Automatic cluster initialization

**Important**: Due to Docker networking limitations on macOS and Windows, the cluster may not work correctly for caching operations. This is a known issue with Docker, not with the implementation.

## Redis Cluster vs Redis Ring

Previously, this library used Redis Ring (client-side sharding) for cluster support. Now it uses actual Redis Cluster (server-side sharding) which provides:

- ✅ Automatic failover
- ✅ Native cluster support with 16384 hash slots
- ✅ Better high availability
- ✅ Easier node management

### Migration from Redis Ring to Redis Cluster

If you were using the old Redis Ring implementation:

```golang
// Old (Redis Ring) - This was incorrectly named as cluster
store := echoCacheMiddleware.NewCacheRedisClusterWithConfig(redis.RingOptions{
    Addrs: map[string]string{
        "shard1": "localhost:17000",
        "shard2": "localhost:17001",
    },
})

// New (Redis Cluster) - True Redis Cluster support
store := echoCacheMiddleware.NewCacheRedisClusterStoreWithConfig(redis.ClusterOptions{
    Addrs: []string{
        "localhost:17000",
        "localhost:17001",
        "localhost:17002",
        "localhost:17003",
        "localhost:17004",
        "localhost:17005",
    },
})
```

⚠️ **Breaking Change**: The function name has changed from `NewCacheRedisClusterWithConfig` to `NewCacheRedisClusterStoreWithConfig` to maintain consistency with other store implementations.

## Advanced Features

### Clear Cache

For Redis Cluster, you can clear all cached data:

```golang
// Example: Clear cache endpoint
e.DELETE("/api/cache/clear", func(c echo.Context) error {
    if redisStore, ok := store.(*echoCacheMiddleware.CacheRedisClusterStore); ok {
        err := redisStore.Clear()
        if err != nil {
            return c.JSON(http.StatusInternalServerError, map[string]string{
                "error": "Failed to clear cache",
            })
        }
        return c.JSON(http.StatusOK, map[string]string{
            "message": "Cache cleared successfully",
        })
    }
    return c.JSON(http.StatusBadRequest, map[string]string{
        "error": "Store is not Redis Cluster",
    })
})
```

**Note**: The `Clear()` method iterates through all master nodes and executes FLUSHDB on each. This operation:
- Is not atomic across the cluster
- May take time proportional to the number of keys
- Should be used sparingly in production

## License

This code is released under the [MIT License](https://github.com/kenshin579/echo-http-cache/blob/main/LICENSE)
