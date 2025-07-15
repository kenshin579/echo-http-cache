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

	// Redis Cluster 캐시 스토어 생성
	store := echo_http_cache.NewCacheRedisClusterStoreWithConfig(redis.ClusterOptions{
		Addrs: []string{
			"localhost:17000",
			"localhost:17001",
			"localhost:17002",
			"localhost:17003",
			"localhost:17004",
			"localhost:17005",
		},
	})

	// 캐시 미들웨어 설정
	e.Use(echo_http_cache.CacheWithConfig(echo_http_cache.CacheConfig{
		Store:      store,
		Expiration: 5 * time.Minute,
	}))

	// API 엔드포인트
	e.GET("/api/data", func(c echo.Context) error {
		// 이 응답은 5분간 캐시됩니다
		log.Println("Generating new response...")
		return c.JSON(http.StatusOK, map[string]interface{}{
			"data":   "This response is cached in Redis Cluster",
			"time":   time.Now().Format(time.RFC3339),
			"cached": false,
		})
	})

	// 캐시 삭제 엔드포인트
	e.DELETE("/api/cache/clear", func(c echo.Context) error {
		if redisStore, ok := store.(*echo_http_cache.CacheRedisClusterStore); ok {
			if err := redisStore.Clear(); err != nil {
				log.Printf("Cache clear error: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error":   "Failed to clear cache",
					"details": err.Error(),
				})
			}
		}
		return c.JSON(http.StatusOK, map[string]string{
			"message": "Cache cleared successfully",
		})
	})

	log.Println("Server starting on :8080...")
	log.Fatal(e.Start(":8080"))
}
