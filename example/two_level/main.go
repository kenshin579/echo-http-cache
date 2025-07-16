package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	echo_http_cache "github.com/kenshin579/echo-http-cache"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Create L1 cache (Memory)
	memoryStore := echo_http_cache.NewCacheMemoryStoreWithConfig(
		echo_http_cache.CacheMemoryStoreConfig{
			Capacity:  100,
			Algorithm: echo_http_cache.LRU,
		},
	)

	// Create L2 cache (Redis)
	redisStore := echo_http_cache.NewCacheRedisStoreWithConfig(redis.Options{
		Addr:     "localhost:6379",
		Password: "", // No password
		DB:       0,  // Default DB
	})

	// Create Two-Level Cache Store
	twoLevelStore := echo_http_cache.NewCacheTwoLevelStoreWithConfig(
		echo_http_cache.TwoLevelConfig{
			L1Store:      memoryStore,
			L2Store:      redisStore,
			Strategy:     echo_http_cache.WriteThrough,
			L1TTL:        5 * time.Minute,  // Memory cache for 5 minutes
			L2TTL:        30 * time.Minute, // Redis cache for 30 minutes
			CacheWarming: true,
		},
	)

	// Setup cache middleware
	e.Use(echo_http_cache.CacheWithConfig(echo_http_cache.CacheConfig{
		Store:        twoLevelStore,
		Expiration:   10 * time.Minute,
		IncludePaths: []string{"/api/"},
	}))

	// API endpoints
	e.GET("/api/data", func(c echo.Context) error {
		log.Println("Generating new response...")
		return c.JSON(http.StatusOK, map[string]interface{}{
			"data":      "This response is cached in both memory and Redis",
			"timestamp": time.Now().Format(time.RFC3339),
			"cached":    false,
		})
	})

	e.GET("/api/user/:id", func(c echo.Context) error {
		userID := c.Param("id")
		log.Printf("Fetching user %s...", userID)
		return c.JSON(http.StatusOK, map[string]interface{}{
			"user_id":   userID,
			"name":      "User " + userID,
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// Cache management endpoints
	if twoLevel, ok := twoLevelStore.(*echo_http_cache.CacheTwoLevelStore); ok {
		// Get cache statistics
		e.GET("/cache/stats", func(c echo.Context) error {
			stats := twoLevel.GetStats()
			return c.JSON(http.StatusOK, stats)
		})

		// Clear L1 cache only
		e.DELETE("/cache/l1", func(c echo.Context) error {
			if err := twoLevel.ClearL1(); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": err.Error(),
				})
			}
			return c.JSON(http.StatusOK, map[string]string{
				"message": "L1 cache cleared successfully",
			})
		})

		// Clear L2 cache only
		e.DELETE("/cache/l2", func(c echo.Context) error {
			if err := twoLevel.ClearL2(); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": err.Error(),
				})
			}
			return c.JSON(http.StatusOK, map[string]string{
				"message": "L2 cache cleared successfully",
			})
		})

		// Clear all caches
		e.DELETE("/cache/all", func(c echo.Context) error {
			if err := twoLevel.ClearAll(); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": err.Error(),
				})
			}
			return c.JSON(http.StatusOK, map[string]string{
				"message": "All caches cleared successfully",
			})
		})

		// Reset statistics
		e.POST("/cache/stats/reset", func(c echo.Context) error {
			twoLevel.ResetStats()
			return c.JSON(http.StatusOK, map[string]string{
				"message": "Cache statistics reset successfully",
			})
		})
	}

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	log.Println("Server starting on :8080...")
	log.Println("Available endpoints:")
	log.Println("  GET  /api/data        - Cached API endpoint")
	log.Println("  GET  /api/user/:id    - User data endpoint")
	log.Println("  GET  /cache/stats     - Cache statistics")
	log.Println("  DELETE /cache/l1      - Clear L1 cache")
	log.Println("  DELETE /cache/l2      - Clear L2 cache")
	log.Println("  DELETE /cache/all     - Clear all caches")
	log.Println("  POST /cache/stats/reset - Reset statistics")
	log.Println("  GET  /health          - Health check")

	log.Fatal(e.Start(":8080"))
}
