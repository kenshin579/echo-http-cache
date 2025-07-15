package echo_http_cache

import (
	"context"
	"time"

	redisCache "github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
)

type (
	// CacheRedisStore is the redis standalone store implementation for Cache
	CacheRedisStore struct {
		store *redisCache.Cache
	}
)

func NewCacheRedisStoreWithConfig(opt redis.Options) CacheStore {
	return &CacheRedisStore{
		redisCache.New(&redisCache.Options{
			Redis: redis.NewClient(&opt),
		}),
	}
}

// Get implements the cache CacheRedisStore interface Get method.
func (store *CacheRedisStore) Get(key uint64) ([]byte, bool) {
	var data []byte
	if err := store.store.Get(context.Background(), keyAsString(key), &data); err == nil {
		return data, true
	}

	return nil, false
}

func (store *CacheRedisStore) Set(key uint64, response []byte, expiration time.Time) {
	store.store.Set(&redisCache.Item{
		Key:   keyAsString(key),
		Value: response,
		TTL:   expiration.Sub(time.Now()),
	})
}

func (store *CacheRedisStore) Release(key uint64) {
	store.store.Delete(context.Background(), keyAsString(key))
}
