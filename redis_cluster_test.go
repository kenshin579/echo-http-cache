package echo_http_cache

import (
	"testing"
	"time"
)

func TestCacheRedisClusterStore_Get(t *testing.T) {
	// 실제 Redis Cluster 없이 기본 테스트만 수행
	// Mock 테스트는 추후 redismock 의존성 추가 후 구현 예정
	store := NewCacheRedisClusterStore()

	// 인터페이스 구현 확인
	var _ CacheStore = store
}

func TestCacheRedisClusterStore_Set(t *testing.T) {
	store := NewCacheRedisClusterStore()

	// 인터페이스 구현 확인
	key := uint64(12345)
	value := []byte("test-value")
	expiration := time.Now().Add(5 * time.Minute)

	// Set 메서드가 존재하는지 확인
	store.Set(key, value, expiration)
}

func TestCacheRedisClusterStore_Release(t *testing.T) {
	store := NewCacheRedisClusterStore()

	// Release 메서드가 존재하는지 확인
	key := uint64(12345)
	store.Release(key)
}

func TestCacheRedisClusterStore_Clear(t *testing.T) {
	store := NewCacheRedisClusterStore()
	redisStore, ok := store.(*CacheRedisClusterStore)
	if !ok {
		t.Fatal("Failed to cast to CacheRedisClusterStore")
	}

	// Clear 메서드가 존재하는지 확인
	_ = redisStore.Clear()
}
