package echo_http_cache

import (
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/kenshin579/echo-http-cache/test"
)

// setupBenchmarkStores creates memory and redis stores for benchmarking
func setupBenchmarkStores(b *testing.B) (CacheStore, CacheStore, func()) {
	memoryStore := NewCacheMemoryStore()
	db, _ := test.NewRedisDB()
	redisStore := NewCacheRedisStoreWithConfig(redis.Options{
		Addr: db.Addr(),
	})
	cleanup := func() {
		db.Close()
	}
	return memoryStore, redisStore, cleanup
}

// BenchmarkTwoLevelCache_L1Hit benchmarks L1 cache hits
func BenchmarkTwoLevelCache_L1Hit(b *testing.B) {
	memoryStore, redisStore, cleanup := setupBenchmarkStores(b)
	defer cleanup()

	config := TwoLevelConfig{
		L1Store:      memoryStore,
		L2Store:      redisStore,
		Strategy:     "write-through",
		L1TTL:        5 * time.Minute,
		L2TTL:        30 * time.Minute,
		SyncMode:     "sync",
		CacheWarming: true,
	}

	store := NewCacheTwoLevelStoreWithConfig(config)
	twoLevel := store.(*CacheTwoLevelStore)
	defer twoLevel.Stop()

	// Pre-populate cache
	key := uint64(12345)
	value := []byte("benchmark-value")
	expiration := time.Now().Add(10 * time.Minute)
	store.Set(key, value, expiration)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = store.Get(key)
		}
	})
}

// BenchmarkTwoLevelCache_L2Hit benchmarks L2 cache hits (L1 miss)
func BenchmarkTwoLevelCache_L2Hit(b *testing.B) {
	memoryStore, redisStore, cleanup := setupBenchmarkStores(b)
	defer cleanup()

	config := TwoLevelConfig{
		L1Store:      memoryStore,
		L2Store:      redisStore,
		Strategy:     "write-through",
		L1TTL:        5 * time.Minute,
		L2TTL:        30 * time.Minute,
		SyncMode:     "sync",
		CacheWarming: false, // Disable warming to test L2 hits
	}

	store := NewCacheTwoLevelStoreWithConfig(config)
	twoLevel := store.(*CacheTwoLevelStore)
	defer twoLevel.Stop()

	// Pre-populate only L2
	key := uint64(12346)
	value := []byte("benchmark-value-l2")
	expiration := time.Now().Add(10 * time.Minute)
	redisStore.Set(key, value, expiration)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = store.Get(key)
		}
	})
}

// BenchmarkTwoLevelCache_Miss benchmarks cache misses
func BenchmarkTwoLevelCache_Miss(b *testing.B) {
	memoryStore, redisStore, cleanup := setupBenchmarkStores(b)
	defer cleanup()

	config := TwoLevelConfig{
		L1Store:      memoryStore,
		L2Store:      redisStore,
		Strategy:     "write-through",
		L1TTL:        5 * time.Minute,
		L2TTL:        30 * time.Minute,
		SyncMode:     "sync",
		CacheWarming: false,
	}

	store := NewCacheTwoLevelStoreWithConfig(config)
	twoLevel := store.(*CacheTwoLevelStore)
	defer twoLevel.Stop()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			// Use different keys to ensure misses
			key := uint64(99999 + i)
			_, _ = store.Get(key)
			i++
		}
	})
}

// BenchmarkTwoLevelCache_Set_WriteThrough benchmarks write-through strategy
func BenchmarkTwoLevelCache_Set_WriteThrough(b *testing.B) {
	memoryStore, redisStore, cleanup := setupBenchmarkStores(b)
	defer cleanup()

	config := TwoLevelConfig{
		L1Store:      memoryStore,
		L2Store:      redisStore,
		Strategy:     "write-through",
		L1TTL:        5 * time.Minute,
		L2TTL:        30 * time.Minute,
		SyncMode:     "sync",
		CacheWarming: true,
	}

	store := NewCacheTwoLevelStoreWithConfig(config)
	twoLevel := store.(*CacheTwoLevelStore)
	defer twoLevel.Stop()

	value := []byte("benchmark-set-value")
	expiration := time.Now().Add(10 * time.Minute)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := uint64(20000 + i)
			store.Set(key, value, expiration)
			i++
		}
	})
}

// BenchmarkTwoLevelCache_Set_WriteBack benchmarks write-back strategy
func BenchmarkTwoLevelCache_Set_WriteBack(b *testing.B) {
	memoryStore, redisStore, cleanup := setupBenchmarkStores(b)
	defer cleanup()

	config := TwoLevelConfig{
		L1Store:      memoryStore,
		L2Store:      redisStore,
		Strategy:     "write-back",
		L1TTL:        5 * time.Minute,
		L2TTL:        30 * time.Minute,
		SyncMode:     "async",
		CacheWarming: true,
	}

	store := NewCacheTwoLevelStoreWithConfig(config)
	twoLevel := store.(*CacheTwoLevelStore)
	defer twoLevel.Stop()

	value := []byte("benchmark-writeback-value")
	expiration := time.Now().Add(10 * time.Minute)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := uint64(30000 + i)
			store.Set(key, value, expiration)
			i++
		}
	})
}

// BenchmarkTwoLevelCache_Set_CacheAside benchmarks cache-aside strategy
func BenchmarkTwoLevelCache_Set_CacheAside(b *testing.B) {
	memoryStore, redisStore, cleanup := setupBenchmarkStores(b)
	defer cleanup()

	config := TwoLevelConfig{
		L1Store:      memoryStore,
		L2Store:      redisStore,
		Strategy:     "cache-aside",
		L1TTL:        5 * time.Minute,
		L2TTL:        30 * time.Minute,
		SyncMode:     "sync",
		CacheWarming: true,
	}

	store := NewCacheTwoLevelStoreWithConfig(config)
	twoLevel := store.(*CacheTwoLevelStore)
	defer twoLevel.Stop()

	value := []byte("benchmark-cacheaside-value")
	expiration := time.Now().Add(10 * time.Minute)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := uint64(40000 + i)
			store.Set(key, value, expiration)
			i++
		}
	})
}

// BenchmarkTwoLevelCache_Mixed benchmarks mixed read/write operations
func BenchmarkTwoLevelCache_Mixed(b *testing.B) {
	memoryStore, redisStore, cleanup := setupBenchmarkStores(b)
	defer cleanup()

	config := TwoLevelConfig{
		L1Store:      memoryStore,
		L2Store:      redisStore,
		Strategy:     "write-through",
		L1TTL:        5 * time.Minute,
		L2TTL:        30 * time.Minute,
		SyncMode:     "sync",
		CacheWarming: true,
	}

	store := NewCacheTwoLevelStoreWithConfig(config)
	twoLevel := store.(*CacheTwoLevelStore)
	defer twoLevel.Stop()

	// Pre-populate some data
	for i := 0; i < 100; i++ {
		key := uint64(50000 + i)
		value := []byte("mixed-benchmark-value")
		expiration := time.Now().Add(10 * time.Minute)
		store.Set(key, value, expiration)
	}

	value := []byte("mixed-new-value")
	expiration := time.Now().Add(10 * time.Minute)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%4 == 0 {
				// 25% writes
				key := uint64(50000 + (i % 1000))
				store.Set(key, value, expiration)
			} else {
				// 75% reads
				key := uint64(50000 + (i % 100))
				_, _ = store.Get(key)
			}
			i++
		}
	})
}

// BenchmarkComparison_SingleLevel_Memory benchmarks single level memory cache
func BenchmarkComparison_SingleLevel_Memory(b *testing.B) {
	store := NewCacheMemoryStore()

	// Pre-populate cache
	key := uint64(60000)
	value := []byte("single-memory-value")
	expiration := time.Now().Add(10 * time.Minute)
	store.Set(key, value, expiration)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = store.Get(key)
		}
	})
}

// BenchmarkComparison_SingleLevel_Redis benchmarks single level Redis cache
func BenchmarkComparison_SingleLevel_Redis(b *testing.B) {
	db, _ := test.NewRedisDB()
	defer db.Close()
	store := NewCacheRedisStoreWithConfig(redis.Options{
		Addr: db.Addr(),
	})

	// Pre-populate cache
	key := uint64(61000)
	value := []byte("single-redis-value")
	expiration := time.Now().Add(10 * time.Minute)
	store.Set(key, value, expiration)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = store.Get(key)
		}
	})
}
