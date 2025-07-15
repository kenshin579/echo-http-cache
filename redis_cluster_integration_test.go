//go:build integration
// +build integration

package echo_http_cache

import (
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

func TestRedisClusterIntegration(t *testing.T) {
	// Redis Cluster 연결 설정
	store := NewCacheRedisClusterStoreWithConfig(redis.ClusterOptions{
		Addrs: []string{
			"localhost:17000",
			"localhost:17001",
			"localhost:17002",
			"localhost:17003",
			"localhost:17004",
			"localhost:17005",
		},
	})

	t.Run("Set and Get", func(t *testing.T) {
		key := uint64(12345)
		value := []byte("test-value-12345")
		expiration := time.Now().Add(5 * time.Minute)

		// Set
		store.Set(key, value, expiration)

		// Get
		result, found := store.Get(key)
		assert.True(t, found)
		assert.Equal(t, value, result)
	})

	t.Run("Get Non-existent Key", func(t *testing.T) {
		key := uint64(99999)

		result, found := store.Get(key)
		assert.False(t, found)
		assert.Nil(t, result)
	})

	t.Run("Set and Release", func(t *testing.T) {
		key := uint64(54321)
		value := []byte("test-value-54321")
		expiration := time.Now().Add(5 * time.Minute)

		// Set
		store.Set(key, value, expiration)

		// Verify it exists
		result, found := store.Get(key)
		assert.True(t, found)
		assert.Equal(t, value, result)

		// Release
		store.Release(key)

		// Verify it's gone
		result, found = store.Get(key)
		assert.False(t, found)
		assert.Nil(t, result)
	})

	t.Run("Clear All", func(t *testing.T) {
		// Set multiple keys
		for i := uint64(1); i <= 5; i++ {
			value := []byte("value-" + string(rune('0'+i)))
			store.Set(i, value, time.Now().Add(5*time.Minute))
		}

		// Verify they exist
		for i := uint64(1); i <= 5; i++ {
			_, found := store.Get(i)
			assert.True(t, found)
		}

		// Clear all
		redisStore := store.(*CacheRedisClusterStore)
		err := redisStore.Clear()
		assert.NoError(t, err)

		// Verify they're all gone
		for i := uint64(1); i <= 5; i++ {
			_, found := store.Get(i)
			assert.False(t, found)
		}
	})

	t.Run("TTL Expiration", func(t *testing.T) {
		key := uint64(77777)
		value := []byte("short-lived-value")
		expiration := time.Now().Add(2 * time.Second)

		// Set with short TTL
		store.Set(key, value, expiration)

		// Verify it exists
		result, found := store.Get(key)
		assert.True(t, found)
		assert.Equal(t, value, result)

		// Wait for expiration
		time.Sleep(3 * time.Second)

		// Verify it's gone
		result, found = store.Get(key)
		assert.False(t, found)
		assert.Nil(t, result)
	})
}
