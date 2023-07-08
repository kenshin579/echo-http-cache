/*
MIT License

Copyright (c) 2023 Frank Oh

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package echo_http_cache

import (
	"context"
	"time"

	redisCache "github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
)

type (
	// CacheRedisClusterStore is the redis cluster store
	CacheRedisClusterStore struct {
		store *redisCache.Cache
	}
)

// NewCacheRedisClusterWithConfig initializes Redis adapter.
func NewCacheRedisClusterWithConfig(opt redis.RingOptions) CacheStore {
	return &CacheRedisClusterStore{
		redisCache.New(&redisCache.Options{
			Redis: redis.NewRing(&opt),
		}),
	}
}

// Get implements the cache CacheRedisClusterStore interface Get method.
func (store *CacheRedisClusterStore) Get(key uint64) ([]byte, bool) {
	var c []byte
	if err := store.store.Get(context.Background(), keyAsString(key), &c); err == nil {
		return c, true
	}

	return nil, false
}

// Set implements the cache CacheRedisClusterStore interface Set method.
func (store *CacheRedisClusterStore) Set(key uint64, response []byte, expiration time.Time) {
	store.store.Set(&redisCache.Item{
		Key:   keyAsString(key),
		Value: response,
		TTL:   expiration.Sub(time.Now()),
	})
}

// Release implements the cache CacheRedisClusterStore interface Release method.
func (store *CacheRedisClusterStore) Release(key uint64) {
	store.store.Delete(context.Background(), keyAsString(key))
}
