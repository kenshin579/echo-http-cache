package echo_http_cache

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/kenshin579/echo-http-cache/test"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
)

type cacheRedisStoreTestSuite struct {
	suite.Suite
	ctx       context.Context
	miniredis *miniredis.Miniredis

	cacheStore CacheStore
	echo       *echo.Echo
}

func TestCacheRedisStoreSuite(t *testing.T) {
	suite.Run(t, new(cacheRedisStoreTestSuite))
}
func (suite *cacheRedisStoreTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	db, _ := test.NewRedisDB()
	store := NewCacheRedisStoreWithConfig(redis.Options{
		Addr: db.Addr(),
	})
	suite.miniredis = db
	suite.cacheStore = store

	suite.echo = echo.New()
	suite.echo.Use(CacheWithConfig(CacheConfig{
		Store:        store,
		Expiration:   5 * time.Second,
		IncludePaths: []string{"/test", "/empty"},
		IncludePathsWithExpiration: map[string]time.Duration{
			"/expired": 1 * time.Minute,
		},
	}))
}

func (suite *cacheRedisStoreTestSuite) TearDownSuite() {
	suite.miniredis.Close()
}

func (suite *cacheRedisStoreTestSuite) Test_Redis_CacheStore() {
	key := generateKey("GET", "1")

	suite.Run("Set, Get", func() {
		suite.cacheStore.Set(key, []byte("test"), time.Now().Add(1*time.Minute))

		data, ok := suite.cacheStore.Get(key)

		suite.True(ok)
		suite.Equal("test", string(data))
	})

	suite.Run("Delete", func() {
		suite.cacheStore.Release(key)
		_, ok := suite.cacheStore.Get(key)
		suite.False(ok)

	})
}

func (suite *cacheRedisStoreTestSuite) Test_Echo_CacheWithConfig() {
	suite.echo.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	suite.echo.GET("/empty/string", func(c echo.Context) error {
		return c.String(http.StatusOK, "")
	})

	suite.echo.GET("/empty/json", func(c echo.Context) error {
		return c.String(http.StatusOK, `{"symbolId":"","type":"","price":0.0}`)
	})

	suite.echo.GET("/expired", func(c echo.Context) error {
		return c.String(http.StatusOK, "expired")
	})

	suite.Run("GET /expired - first call to store response in the cache", func() {
		req := httptest.NewRequest(http.MethodGet, "/expired", nil)
		rec := httptest.NewRecorder()

		suite.echo.ServeHTTP(rec, req)

		suite.Equal(http.StatusOK, rec.Code)
		suite.Equal(`expired`, rec.Body.String())

		key := generateKey(http.MethodGet, "/expired")
		_, ok := suite.cacheStore.Get(key)
		suite.True(ok)
	})

	suite.Run("GET /test - return actual response and store in the cache", func() {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		suite.echo.ServeHTTP(rec, req)

		suite.Equal(http.StatusOK, rec.Code)
		suite.Equal("test", rec.Body.String())

		key := generateKey(http.MethodGet, "/test")
		data, ok := suite.cacheStore.Get(key)
		suite.True(ok)

		var cacheResponse CacheResponse
		err := json.Unmarshal(data, &cacheResponse)
		suite.NoError(err)
		suite.Equal("test", string(cacheResponse.Body))
		suite.Equal(1, cacheResponse.Frequency)
	})

	suite.Run("GET /test - not expired. return response from the cache", func() {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		suite.echo.ServeHTTP(rec, req)

		suite.Equal(http.StatusOK, rec.Code)
		suite.Equal("test", rec.Body.String())

		key := generateKey(http.MethodGet, "/test")
		data, ok := suite.cacheStore.Get(key)
		suite.True(ok)

		var cacheResponse CacheResponse
		err := json.Unmarshal(data, &cacheResponse)
		suite.NoError(err)
		suite.Equal("test", string(cacheResponse.Body))
		suite.Equal(2, cacheResponse.Frequency)
	})

	suite.Run("GET /test - expired. return actual response", func() {
		time.Sleep(5 * time.Second)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		suite.echo.ServeHTTP(rec, req)

		suite.Equal(http.StatusOK, rec.Code)
		suite.Equal("test", rec.Body.String())

		key := generateKey(http.MethodGet, "/test")
		data, ok := suite.cacheStore.Get(key)
		suite.True(ok)

		var cacheResponse CacheResponse
		err := json.Unmarshal(data, &cacheResponse)
		suite.NoError(err)
		suite.Equal("test", string(cacheResponse.Body))
		suite.Equal(1, cacheResponse.Frequency)
	})

	suite.Run("GET /empty/string", func() {
		req := httptest.NewRequest(http.MethodGet, "/empty/string", nil)
		rec := httptest.NewRecorder()

		suite.echo.ServeHTTP(rec, req)

		suite.Equal(http.StatusOK, rec.Code)
		suite.Equal("", rec.Body.String())

		key := generateKey(http.MethodGet, "/empty/string")
		_, ok := suite.cacheStore.Get(key)
		suite.False(ok)
	})

	suite.Run("GET /empty/json", func() {
		req := httptest.NewRequest(http.MethodGet, "/empty/json", nil)
		rec := httptest.NewRecorder()

		suite.echo.ServeHTTP(rec, req)

		suite.Equal(http.StatusOK, rec.Code)
		suite.Equal(`{"symbolId":"","type":"","price":0.0}`, rec.Body.String())

		key := generateKey(http.MethodGet, "/empty/json")
		_, ok := suite.cacheStore.Get(key)
		suite.False(ok)
	})

	suite.Run("GET /expired - second call, still get the response from the cache", func() {
		req := httptest.NewRequest(http.MethodGet, "/expired", nil)
		rec := httptest.NewRecorder()

		suite.echo.ServeHTTP(rec, req)

		suite.Equal(http.StatusOK, rec.Code)
		suite.Equal(`expired`, rec.Body.String())

		key := generateKey(http.MethodGet, "/expired")
		data, ok := suite.cacheStore.Get(key)
		suite.True(ok)

		var cacheResponse CacheResponse
		err := json.Unmarshal(data, &cacheResponse)
		suite.NoError(err)
		suite.Equal(2, cacheResponse.Frequency)
	})
}
