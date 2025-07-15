# Redis Cluster 구현 가이드

## 1. 구현 개요

이 문서는 echo-http-cache 라이브러리에 실제 Redis Cluster 지원을 추가하기 위한 구현 가이드입니다.

## 2. 파일 구조

```
echo-http-cache/
├── redis_cluster.go          # Redis Cluster 구현 (기존 파일 대체)
├── redis_cluster_test.go     # 테스트 코드
└── example/
    └── redis_cluster_example.go  # 사용 예제
```

## 3. 구현 코드

### 3.1 redis_cluster.go

```go
package echocache

import (
    "context"
    "fmt"
    "time"
    
    "github.com/redis/go-redis/v9"
    redisCache "github.com/go-redis/cache/v9"
)

// CacheRedisClusterStore는 실제 Redis Cluster를 사용하는 캐시 스토어입니다
type CacheRedisClusterStore struct {
    client *redis.ClusterClient
    codec  *redisCache.Cache
}

// NewCacheRedisClusterStore creates a new Redis Cluster cache store with default config
func NewCacheRedisClusterStore() CacheStore {
    return NewCacheRedisClusterStoreWithConfig(redis.ClusterOptions{
        Addrs: []string{"localhost:7000"},
    })
}

// NewCacheRedisClusterStoreWithConfig creates a new Redis Cluster cache store
func NewCacheRedisClusterStoreWithConfig(opt redis.ClusterOptions) CacheStore {
    client := redis.NewClusterClient(&opt)
    
    return &CacheRedisClusterStore{
        client: client,
        codec: redisCache.New(&redisCache.Options{
            Redis: client,
        }),
    }
}

// Get implements CacheStore interface
func (store *CacheRedisClusterStore) Get(ctx context.Context, key string) ([]byte, error) {
    var data []byte
    err := store.codec.Get(ctx, key, &data)
    if err != nil {
        return nil, err
    }
    return data, nil
}

// Set implements CacheStore interface
func (store *CacheRedisClusterStore) Set(ctx context.Context, key string, val []byte, ttl time.Duration) error {
    return store.codec.Set(&redisCache.Item{
        Ctx:   ctx,
        Key:   key,
        Value: val,
        TTL:   ttl,
    })
}

// Delete implements CacheStore interface
func (store *CacheRedisClusterStore) Delete(ctx context.Context, key string) error {
    return store.codec.Delete(ctx, key)
}

// Clear implements CacheStore interface
func (store *CacheRedisClusterStore) Clear(ctx context.Context) error {
    // Redis Cluster에서는 FLUSHALL을 지원하지 않으므로
    // 각 노드별로 처리해야 합니다
    return store.client.ForEachMaster(ctx, func(ctx context.Context, shard *redis.Client) error {
        return shard.FlushDB(ctx).Err()
    })
}

// 레거시 인터페이스 지원 (uint64 기반)
func (store *CacheRedisClusterStore) Get(key uint64) ([]byte, error) {
    ctx := context.Background()
    return store.Get(ctx, fmt.Sprintf("%d", key))
}

func (store *CacheRedisClusterStore) Set(key uint64, val []byte, expire time.Duration) error {
    ctx := context.Background()
    return store.Set(ctx, fmt.Sprintf("%d", key), val, expire)
}

func (store *CacheRedisClusterStore) Delete(key uint64) error {
    ctx := context.Background()
    return store.Delete(ctx, fmt.Sprintf("%d", key))
}
```

## 4. 테스트 코드

### 4.1 단위 테스트 (Mock 사용)

Redis Cluster 모킹을 사용한 단위 테스트:

```go
package echocache

import (
    "context"
    "testing"
    "time"
    "errors"
    
    "github.com/go-redis/redismock/v9"
    "github.com/stretchr/testify/assert"
)

func TestCacheRedisClusterStore_Get(t *testing.T) {
    // Redis Cluster Mock 생성
    clusterClient, mock := redismock.NewClusterMock()
    codec := redisCache.New(&redisCache.Options{
        Redis: clusterClient,
    })
    
    store := &CacheRedisClusterStore{
        client: clusterClient,
        codec:  codec,
    }
    
    ctx := context.Background()
    key := "test:key"
    value := []byte("test-value")
    
    // Mock 설정
    mock.ExpectGet(key).SetVal(string(value))
    
    // 테스트 실행
    result, err := store.Get(ctx, key)
    
    // 검증
    assert.NoError(t, err)
    assert.Equal(t, value, result)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCacheRedisClusterStore_Set(t *testing.T) {
    clusterClient, mock := redismock.NewClusterMock()
    codec := redisCache.New(&redisCache.Options{
        Redis: clusterClient,
    })
    
    store := &CacheRedisClusterStore{
        client: clusterClient,
        codec:  codec,
    }
    
    ctx := context.Background()
    key := "test:key"
    value := []byte("test-value")
    ttl := 5 * time.Minute
    
    // Mock 설정
    mock.ExpectSet(key, value, ttl).SetVal("OK")
    
    // 테스트 실행
    err := store.Set(ctx, key, value, ttl)
    
    // 검증
    assert.NoError(t, err)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCacheRedisClusterStore_Delete(t *testing.T) {
    clusterClient, mock := redismock.NewClusterMock()
    codec := redisCache.New(&redisCache.Options{
        Redis: clusterClient,
    })
    
    store := &CacheRedisClusterStore{
        client: clusterClient,
        codec:  codec,
    }
    
    ctx := context.Background()
    key := "test:key"
    
    // Mock 설정
    mock.ExpectDel(key).SetVal(1)
    
    // 테스트 실행
    err := store.Delete(ctx, key)
    
    // 검증
    assert.NoError(t, err)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCacheRedisClusterStore_Clear(t *testing.T) {
    clusterClient, mock := redismock.NewClusterMock()
    
    store := &CacheRedisClusterStore{
        client: clusterClient,
    }
    
    ctx := context.Background()
    
    // 각 마스터 노드에 대해 FLUSHDB 기대
    mock.ExpectFlushDB().SetVal("OK")
    
    // 테스트 실행
    err := store.Clear(ctx)
    
    // 검증
    assert.NoError(t, err)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCacheRedisClusterStore_GetNotFound(t *testing.T) {
    clusterClient, mock := redismock.NewClusterMock()
    codec := redisCache.New(&redisCache.Options{
        Redis: clusterClient,
    })
    
    store := &CacheRedisClusterStore{
        client: clusterClient,
        codec:  codec,
    }
    
    ctx := context.Background()
    key := "nonexistent:key"
    
    // Mock 설정 - 키가 없는 경우
    mock.ExpectGet(key).RedisNil()
    
    // 테스트 실행
    result, err := store.Get(ctx, key)
    
    // 검증
    assert.Error(t, err)
    assert.Nil(t, result)
    assert.NoError(t, mock.ExpectationsWereMet())
}

// 레거시 인터페이스 테스트
func TestCacheRedisClusterStore_LegacyInterface(t *testing.T) {
    clusterClient, mock := redismock.NewClusterMock()
    codec := redisCache.New(&redisCache.Options{
        Redis: clusterClient,
    })
    
    store := &CacheRedisClusterStore{
        client: clusterClient,
        codec:  codec,
    }
    
    key := uint64(12345)
    value := []byte("test-value")
    ttl := 5 * time.Minute
    
    // Set 테스트
    mock.ExpectSet("12345", value, ttl).SetVal("OK")
    err := store.Set(key, value, ttl)
    assert.NoError(t, err)
    
    // Get 테스트
    mock.ExpectGet("12345").SetVal(string(value))
    result, err := store.Get(key)
    assert.NoError(t, err)
    assert.Equal(t, value, result)
    
    // Delete 테스트
    mock.ExpectDel("12345").SetVal(1)
    err = store.Delete(key)
    assert.NoError(t, err)
    
    assert.NoError(t, mock.ExpectationsWereMet())
}
```

## 5. 사용 예제

### 5.1 example/redis_cluster_example.go

```go
package main

import (
    "log"
    "net/http"
    "time"
    
    "github.com/labstack/echo/v4"
    "github.com/redis/go-redis/v9"
    echocache "github.com/sihirpro/echo-http-cache"
)

func main() {
    e := echo.New()
    
    // Redis Cluster 캐시 스토어 생성
    store := echocache.NewCacheRedisClusterStoreWithConfig(redis.ClusterOptions{
        Addrs: []string{
            "localhost:7000",
            "localhost:7001",
            "localhost:7002",
            "localhost:7003",
            "localhost:7004",
            "localhost:7005",
        },
    })
    
    // 캐시 미들웨어 설정
    e.Use(echocache.NewWithConfig(echocache.Config{
        Store:      store,
        Expiration: 5 * time.Minute,
    }))
    
    // API 엔드포인트
    e.GET("/api/data", func(c echo.Context) error {
        // 이 응답은 5분간 캐시됩니다
        return c.JSON(http.StatusOK, map[string]interface{}{
            "data": "This response is cached in Redis Cluster",
            "time": time.Now().Format(time.RFC3339),
        })
    })
    
    log.Fatal(e.Start(":8080"))
}
```

## 6. 구현 순서

1. **기존 코드 백업**
   ```bash
   cp redis_cluster.go redis_cluster_old.go
   ```

2. **새 구현 작성**
   - 위의 `redis_cluster.go` 코드로 기존 파일 대체

3. **테스트 작성**
   - Mock 기반 단위 테스트 작성
   - 모든 메서드에 대한 테스트 커버리지 확보

4. **레거시 테스트 확인**
   - 기존 테스트가 있다면 새 구현과 호환되는지 확인
   - 필요시 테스트 수정

5. **문서 업데이트**
   - README.md에 Redis Cluster 사용법 추가
   - 마이그레이션 가이드 작성

## 7. 주의사항

### 7.1 Clear() 메서드
- Redis Cluster는 FLUSHALL을 지원하지 않음
- `ForEachMaster`를 사용하여 각 마스터 노드별로 FLUSHDB 실행
- 프로덕션에서는 주의해서 사용

### 7.2 키 네임스페이스
- 캐시 키에 적절한 prefix 사용 권장
- 예: `cache:user:123`, `cache:product:456`

### 7.3 Mock 테스트의 장점
- 실제 Redis Cluster 없이 테스트 가능
- CI/CD 환경에서 안정적
- 다양한 에러 시나리오 테스트 가능
- 빠른 실행 속도

## 8. 마이그레이션 체크리스트

- [ ] 기존 `redis_cluster.go` 백업
- [ ] 새 구현으로 파일 교체
- [ ] Mock 기반 단위 테스트 실행
- [ ] 기존 통합 테스트 확인 (있는 경우)
- [ ] 문서 업데이트
- [ ] 예제 코드 테스트 