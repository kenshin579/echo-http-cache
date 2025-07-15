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
	// CacheRedisClusterStore is the redis cluster store implementation
	CacheRedisClusterStore struct {
		client *redis.ClusterClient
		codec  *redisCache.Cache
	}
)

// NewCacheRedisClusterStore creates a new Redis Cluster cache store with default config
func NewCacheRedisClusterStore() CacheStore {
	return NewCacheRedisClusterStoreWithConfig(redis.ClusterOptions{
		Addrs: []string{"localhost:17000"},
	})
}

// NewCacheRedisClusterStoreWithConfig creates a new Redis Cluster cache store
func NewCacheRedisClusterStoreWithConfig(opt redis.ClusterOptions) CacheStore {
	// Set default options for better compatibility
	if opt.ReadTimeout == 0 {
		opt.ReadTimeout = 3 * time.Second
	}
	if opt.WriteTimeout == 0 {
		opt.WriteTimeout = opt.ReadTimeout
	}
	// Enable route by latency for better performance
	opt.RouteByLatency = true

	client := redis.NewClusterClient(&opt)

	return &CacheRedisClusterStore{
		client: client,
		codec: redisCache.New(&redisCache.Options{
			Redis: client,
		}),
	}
}

// Get implements the cache CacheRedisClusterStore interface Get method.
func (store *CacheRedisClusterStore) Get(key uint64) ([]byte, bool) {
	var data []byte
	err := store.codec.Get(context.Background(), keyAsString(key), &data)
	if err != nil {
		return nil, false
	}
	return data, true
}

// Set implements the cache CacheRedisClusterStore interface Set method.
func (store *CacheRedisClusterStore) Set(key uint64, response []byte, expiration time.Time) {
	store.codec.Set(&redisCache.Item{
		Ctx:   context.Background(),
		Key:   keyAsString(key),
		Value: response,
		TTL:   time.Until(expiration),
	})
}

// Release implements the cache CacheRedisClusterStore interface Release method.
func (store *CacheRedisClusterStore) Release(key uint64) {
	store.codec.Delete(context.Background(), keyAsString(key))
}

// Clear removes all cache entries from all master nodes
func (store *CacheRedisClusterStore) Clear() error {
	ctx := context.Background()
	// Redis Cluster doesn't support FLUSHALL across all nodes
	// We need to iterate through each master node
	return store.client.ForEachMaster(ctx, func(ctx context.Context, shard *redis.Client) error {
		return shard.FlushDB(ctx).Err()
	})
}
