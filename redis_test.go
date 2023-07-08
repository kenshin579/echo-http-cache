package echo_http_cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/kenshin579/echo-http-cache/test"
	"github.com/stretchr/testify/suite"
)

type redisStoreTestSuite struct {
	suite.Suite
	ctx       context.Context
	miniredis *miniredis.Miniredis

	cacheStore CacheStore
}

func TestRedisStoreSuite(t *testing.T) {
	suite.Run(t, new(redisStoreTestSuite))
}
func (suite *redisStoreTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	db, _ := test.NewRedisDB()
	store := NewRedis(redis.Options{
		Addr: db.Addr(),
	})
	suite.miniredis = db
	suite.cacheStore = store
}

func (suite *redisStoreTestSuite) TearDownSuite() {
	suite.miniredis.Close()
}

func (suite *redisStoreTestSuite) Test() {
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
