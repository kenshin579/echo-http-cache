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
