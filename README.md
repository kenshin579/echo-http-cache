# Echo HTTP middleware for caching
This library is Echo HTTP middleware that provides server-side application layer caching for REST APIs. 

> I tried to use this library for my personal project to cache API response using a redis standalone server, but it didn't work well. The library hasn't been updated in a while and it seems like it is not maintained so I forked from https://github.com/SporkHubr/echo-http-cache and modifeid the code. 

## Feature

- Support different store (local memory, redis Standalone, redis cluster) for caching
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



## Quickstart

Here are two examples:

- Using local memory
- Using redis standalone server



```golang
package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	echoCacheMiddleware "github.com/kenshin579/echo-http-cache"
	"github.com/labstack/echo/v4"
)

func Test_LocalMemory(t *testing.T) {
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

func Test_Redis_Standalone(t *testing.T) {
	e := echo.New()
	e.Use(echoCacheMiddleware.CacheWithConfig(echoCacheMiddleware.CacheConfig{
		Store: echoCacheMiddleware.NewCacheRedisStoreWithConfig(redis.Options{
			Addr:     "localhost",
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



## License

This code is released under the [MIT License](https://github.com/kenshin579/echo-http-cache/blob/main/LICENSE)
