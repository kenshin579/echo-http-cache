# Two-Level Caching 구현 가이드

## 1. 구현 개요

이 문서는 echo-http-cache 라이브러리에 Two-Level Caching 기능을 추가하기 위한 상세한 구현 가이드입니다.

## 2. 파일 구조

```
echo-http-cache/
├── two_level_cache.go           # Two-Level Store 구현
├── two_level_cache_test.go      # 테스트 코드
├── cache_strategy.go            # 캐시 전략 정의
├── cache_stats.go               # 통계 및 모니터링
└── example/
    └── two_level_example.go     # 사용 예제
```

## 3. 핵심 로직 설계

### 3.1 캐시 조회 로직
```go
func (store *CacheTwoLevelStore) Get(key uint64) ([]byte, bool) {
    // 1. L1 캐시 조회 (메모리)
    if data, found := store.l1Store.Get(key); found {
        return data, true // L1 Hit
    }
    
    // 2. L2 캐시 조회 (Redis)
    if data, found := store.l2Store.Get(key); found {
        // Cache Warming: L2에서 찾은 데이터를 L1에 승격
        if store.config.CacheWarming {
            store.l1Store.Set(key, data, time.Now().Add(store.config.L1TTL))
        }
        return data, true // L2 Hit
    }
    
    return nil, false // Cache Miss
}
```

### 3.2 캐시 저장 로직
```go
func (store *CacheTwoLevelStore) Set(key uint64, response []byte, expiration time.Time) {
    switch store.config.Strategy {
    case WriteThrough:
        // L1, L2 동시 저장
        store.l1Store.Set(key, response, expiration)
        store.l2Store.Set(key, response, expiration)
        
    case WriteBack:
        // L1에 먼저 저장, L2는 비동기 처리
        store.l1Store.Set(key, response, expiration)
        go store.asyncUpdateL2(key, response, expiration)
        
    case CacheAside:
        // 애플리케이션 레벨에서 관리
        store.l1Store.Set(key, response, expiration)
        store.l2Store.Set(key, response, expiration)
    }
}
```

### 3.3 설정 사용법

#### 기본 설정 지원
```go
// 간단한 설정
store := echo_http_cache.NewCacheTwoLevelStore(
    echo_http_cache.NewCacheMemoryStore(),
    echo_http_cache.NewCacheRedisStoreWithConfig(redisOpts),
)
```

#### 고급 설정 지원
```go
// 상세 설정
store := echo_http_cache.NewCacheTwoLevelStoreWithConfig(
    echo_http_cache.TwoLevelStoreConfig{
        L1Store: echo_http_cache.NewCacheMemoryStoreWithConfig(memoryConfig),
        L2Store: echo_http_cache.NewCacheRedisStoreWithConfig(redisConfig),
        Strategy: echo_http_cache.WriteThrough,
        L1TTL: 5 * time.Minute,
        L2TTL: 30 * time.Minute,
        CacheWarming: true,
        SyncMode: echo_http_cache.AsyncMode,
    },
)
```

#### 설정 호환성
```go
// 기존 방식 (단일 스토어)
e.Use(echo_http_cache.Cache(memoryStore))

// 새로운 방식 (투레벨 스토어)
e.Use(echo_http_cache.Cache(twoLevelStore))
```

## 4. 구현 코드

### 4.1 cache_strategy.go - 캐시 전략 정의

```go
package echo_http_cache

import "time"

// CacheStrategy defines the caching strategy for two-level cache
type CacheStrategy string

const (
    // WriteThrough writes to both L1 and L2 synchronously
    WriteThrough CacheStrategy = "WRITE_THROUGH"
    
    // WriteBack writes to L1 first, then L2 asynchronously
    WriteBack CacheStrategy = "WRITE_BACK"
    
    // CacheAside application manages cache explicitly
    CacheAside CacheStrategy = "CACHE_ASIDE"
)

// SyncMode defines the synchronization mode
type SyncMode string

const (
    // SyncMode - synchronous operations
    SyncMode SyncMode = "SYNC"
    
    // AsyncMode - asynchronous operations where possible
    AsyncMode SyncMode = "ASYNC"
)

// TwoLevelConfig represents configuration for TwoLevelStore
type TwoLevelConfig struct {
    L1Store      CacheStore
    L2Store      CacheStore
    Strategy     CacheStrategy
    L1TTL        time.Duration
    L2TTL        time.Duration
    CacheWarming bool
    SyncMode     SyncMode
    AsyncBuffer  int // Buffer size for async operations
}

// DefaultTwoLevelConfig provides default configuration
var DefaultTwoLevelConfig = TwoLevelConfig{
    Strategy:     WriteThrough,
    L1TTL:        5 * time.Minute,
    L2TTL:        30 * time.Minute,
    CacheWarming: true,
    SyncMode:     AsyncMode,
    AsyncBuffer:  1000,
}
```

### 4.2 cache_stats.go - 통계 및 모니터링

```go
package echo_http_cache

import (
    "sync/atomic"
    "time"
)

// CacheStats represents cache statistics
type CacheStats struct {
    L1Hits       int64   `json:"l1Hits"`
    L2Hits       int64   `json:"l2Hits"`
    TotalMiss    int64   `json:"totalMiss"`
    TotalRequest int64   `json:"totalRequest"`
    HitRate      float64 `json:"hitRate"`
    L1HitRate    float64 `json:"l1HitRate"`
    L2HitRate    float64 `json:"l2HitRate"`
    L1Size       int     `json:"l1Size"`
    L2Size       int     `json:"l2Size"`
    LastUpdate   time.Time `json:"lastUpdate"`
}

// CacheMetrics holds atomic counters for thread-safe statistics
type CacheMetrics struct {
    l1Hits       int64
    l2Hits       int64
    totalMiss    int64
    totalRequest int64
}

// IncrementL1Hit atomically increments L1 hit counter
func (m *CacheMetrics) IncrementL1Hit() {
    atomic.AddInt64(&m.l1Hits, 1)
    atomic.AddInt64(&m.totalRequest, 1)
}

// IncrementL2Hit atomically increments L2 hit counter
func (m *CacheMetrics) IncrementL2Hit() {
    atomic.AddInt64(&m.l2Hits, 1)
    atomic.AddInt64(&m.totalRequest, 1)
}

// IncrementMiss atomically increments miss counter
func (m *CacheMetrics) IncrementMiss() {
    atomic.AddInt64(&m.totalMiss, 1)
    atomic.AddInt64(&m.totalRequest, 1)
}

// GetStats returns current statistics
func (m *CacheMetrics) GetStats() CacheStats {
    l1Hits := atomic.LoadInt64(&m.l1Hits)
    l2Hits := atomic.LoadInt64(&m.l2Hits)
    totalMiss := atomic.LoadInt64(&m.totalMiss)
    totalRequest := atomic.LoadInt64(&m.totalRequest)
    
    var hitRate, l1HitRate, l2HitRate float64
    if totalRequest > 0 {
        hitRate = float64(l1Hits+l2Hits) / float64(totalRequest) * 100
        l1HitRate = float64(l1Hits) / float64(totalRequest) * 100
        l2HitRate = float64(l2Hits) / float64(totalRequest) * 100
    }
    
    return CacheStats{
        L1Hits:       l1Hits,
        L2Hits:       l2Hits,
        TotalMiss:    totalMiss,
        TotalRequest: totalRequest,
        HitRate:      hitRate,
        L1HitRate:    l1HitRate,
        L2HitRate:    l2HitRate,
        LastUpdate:   time.Now(),
    }
}

// Reset resets all counters
func (m *CacheMetrics) Reset() {
    atomic.StoreInt64(&m.l1Hits, 0)
    atomic.StoreInt64(&m.l2Hits, 0)
    atomic.StoreInt64(&m.totalMiss, 0)
    atomic.StoreInt64(&m.totalRequest, 0)
}
```

### 4.3 two_level_cache.go - 메인 구현

```go
package echo_http_cache

import (
    "sync"
    "time"
)

// CacheTwoLevelStore implements two-level caching with L1 (memory) and L2 (Redis)
type CacheTwoLevelStore struct {
    config    TwoLevelConfig
    metrics   *CacheMetrics
    asyncChan chan asyncOperation
    wg        sync.WaitGroup
    stopChan  chan struct{}
}

// asyncOperation represents an async cache operation
type asyncOperation struct {
    operation string
    key       uint64
    data      []byte
    expiration time.Time
}

// NewCacheTwoLevelStore creates a new two-level cache store with default config
func NewCacheTwoLevelStore(l1Store, l2Store CacheStore) CacheStore {
    config := DefaultTwoLevelConfig
    config.L1Store = l1Store
    config.L2Store = l2Store
    return NewCacheTwoLevelStoreWithConfig(config)
}

// NewCacheTwoLevelStoreWithConfig creates a new two-level cache store with custom config
func NewCacheTwoLevelStoreWithConfig(config TwoLevelConfig) CacheStore {
    // Set defaults for missing values
    if config.L1TTL == 0 {
        config.L1TTL = DefaultTwoLevelConfig.L1TTL
    }
    if config.L2TTL == 0 {
        config.L2TTL = DefaultTwoLevelConfig.L2TTL
    }
    if config.AsyncBuffer == 0 {
        config.AsyncBuffer = DefaultTwoLevelConfig.AsyncBuffer
    }
    
    store := &CacheTwoLevelStore{
        config:    config,
        metrics:   &CacheMetrics{},
        asyncChan: make(chan asyncOperation, config.AsyncBuffer),
        stopChan:  make(chan struct{}),
    }
    
    // Start async worker if using WriteBack strategy
    if config.Strategy == WriteBack {
        store.startAsyncWorker()
    }
    
    return store
}

// Get implements CacheStore interface
func (store *CacheTwoLevelStore) Get(key uint64) ([]byte, bool) {
    // 1. Try L1 cache first (memory)
    if data, found := store.config.L1Store.Get(key); found {
        store.metrics.IncrementL1Hit()
        return data, true
    }
    
    // 2. Try L2 cache (Redis)
    if data, found := store.config.L2Store.Get(key); found {
        store.metrics.IncrementL2Hit()
        
        // Cache warming: promote L2 data to L1
        if store.config.CacheWarming {
            l1Expiration := time.Now().Add(store.config.L1TTL)
            store.config.L1Store.Set(key, data, l1Expiration)
        }
        
        return data, true
    }
    
    // Cache miss
    store.metrics.IncrementMiss()
    return nil, false
}

// Set implements CacheStore interface
func (store *CacheTwoLevelStore) Set(key uint64, response []byte, expiration time.Time) {
    switch store.config.Strategy {
    case WriteThrough:
        store.setWriteThrough(key, response, expiration)
    case WriteBack:
        store.setWriteBack(key, response, expiration)
    case CacheAside:
        store.setCacheAside(key, response, expiration)
    }
}

// Release implements CacheStore interface
func (store *CacheTwoLevelStore) Release(key uint64) {
    // Remove from both L1 and L2
    store.config.L1Store.Release(key)
    store.config.L2Store.Release(key)
}

// setWriteThrough implements write-through strategy
func (store *CacheTwoLevelStore) setWriteThrough(key uint64, response []byte, expiration time.Time) {
    // Calculate L1 expiration (shorter TTL)
    l1Expiration := time.Now().Add(store.config.L1TTL)
    if l1Expiration.After(expiration) {
        l1Expiration = expiration
    }
    
    // Calculate L2 expiration (longer TTL)
    l2Expiration := time.Now().Add(store.config.L2TTL)
    if l2Expiration.After(expiration) {
        l2Expiration = expiration
    }
    
    // Write to both caches synchronously
    store.config.L1Store.Set(key, response, l1Expiration)
    store.config.L2Store.Set(key, response, l2Expiration)
}

// setWriteBack implements write-back strategy
func (store *CacheTwoLevelStore) setWriteBack(key uint64, response []byte, expiration time.Time) {
    // Write to L1 immediately
    l1Expiration := time.Now().Add(store.config.L1TTL)
    if l1Expiration.After(expiration) {
        l1Expiration = expiration
    }
    store.config.L1Store.Set(key, response, l1Expiration)
    
    // Queue L2 write for async processing
    l2Expiration := time.Now().Add(store.config.L2TTL)
    if l2Expiration.After(expiration) {
        l2Expiration = expiration
    }
    
    select {
    case store.asyncChan <- asyncOperation{
        operation:  "set",
        key:        key,
        data:       response,
        expiration: l2Expiration,
    }:
    default:
        // Channel is full, fallback to synchronous write
        store.config.L2Store.Set(key, response, l2Expiration)
    }
}

// setCacheAside implements cache-aside strategy
func (store *CacheTwoLevelStore) setCacheAside(key uint64, response []byte, expiration time.Time) {
    // Simple implementation: write to both (similar to write-through)
    store.setWriteThrough(key, response, expiration)
}

// startAsyncWorker starts the async worker goroutine
func (store *CacheTwoLevelStore) startAsyncWorker() {
    store.wg.Add(1)
    go func() {
        defer store.wg.Done()
        for {
            select {
            case op := <-store.asyncChan:
                switch op.operation {
                case "set":
                    store.config.L2Store.Set(op.key, op.data, op.expiration)
                case "release":
                    store.config.L2Store.Release(op.key)
                }
            case <-store.stopChan:
                return
            }
        }
    }()
}

// Stop gracefully stops the two-level cache store
func (store *CacheTwoLevelStore) Stop() {
    if store.config.Strategy == WriteBack {
        close(store.stopChan)
        store.wg.Wait()
        close(store.asyncChan)
    }
}

// GetStats returns cache statistics
func (store *CacheTwoLevelStore) GetStats() CacheStats {
    stats := store.metrics.GetStats()
    
    // Add size information if available
    if memorySizer, ok := store.config.L1Store.(interface{ Size() int }); ok {
        stats.L1Size = memorySizer.Size()
    }
    if redisSizer, ok := store.config.L2Store.(interface{ Size() int }); ok {
        stats.L2Size = redisSizer.Size()
    }
    
    return stats
}

// ClearL1 clears only L1 cache
func (store *CacheTwoLevelStore) ClearL1() error {
    if clearer, ok := store.config.L1Store.(interface{ Clear() error }); ok {
        return clearer.Clear()
    }
    return nil
}

// ClearL2 clears only L2 cache
func (store *CacheTwoLevelStore) ClearL2() error {
    if clearer, ok := store.config.L2Store.(interface{ Clear() error }); ok {
        return clearer.Clear()
    }
    return nil
}

// ClearAll clears both L1 and L2 caches
func (store *CacheTwoLevelStore) ClearAll() error {
    var err1, err2 error
    
    if clearer, ok := store.config.L1Store.(interface{ Clear() error }); ok {
        err1 = clearer.Clear()
    }
    if clearer, ok := store.config.L2Store.(interface{ Clear() error }); ok {
        err2 = clearer.Clear()
    }
    
    if err1 != nil {
        return err1
    }
    return err2
}

// SyncL1ToL2 synchronizes L1 cache to L2
func (store *CacheTwoLevelStore) SyncL1ToL2() error {
    // This would require additional interface methods on CacheStore
    // to iterate over all keys. For now, return not implemented.
    return nil
}

// SyncL2ToL1 synchronizes L2 cache to L1
func (store *CacheTwoLevelStore) SyncL2ToL1() error {
    // This would require additional interface methods on CacheStore
    // to iterate over all keys. For now, return not implemented.
    return nil
}

// ResetStats resets cache statistics
func (store *CacheTwoLevelStore) ResetStats() {
    store.metrics.Reset()
}
```

### 4.4 two_level_cache_test.go - 테스트 코드

```go
package echo_http_cache

import (
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
)

type TwoLevelCacheTestSuite struct {
    suite.Suite
    memoryStore  CacheStore
    redisStore   CacheStore
    twoLevelStore CacheStore
}

func TestTwoLevelCacheSuite(t *testing.T) {
    suite.Run(t, new(TwoLevelCacheTestSuite))
}

func (suite *TwoLevelCacheTestSuite) SetupTest() {
    suite.memoryStore = NewCacheMemoryStoreWithConfig(CacheMemoryStoreConfig{
        Capacity:  10,
        Algorithm: LRU,
    })
    
    // For testing, we can use another memory store as mock Redis
    suite.redisStore = NewCacheMemoryStoreWithConfig(CacheMemoryStoreConfig{
        Capacity:  100,
        Algorithm: LRU,
    })
    
    suite.twoLevelStore = NewCacheTwoLevelStoreWithConfig(TwoLevelConfig{
        L1Store:      suite.memoryStore,
        L2Store:      suite.redisStore,
        Strategy:     WriteThrough,
        L1TTL:        5 * time.Minute,
        L2TTL:        30 * time.Minute,
        CacheWarming: true,
    })
}

func (suite *TwoLevelCacheTestSuite) TearDownTest() {
    if twoLevel, ok := suite.twoLevelStore.(*CacheTwoLevelStore); ok {
        twoLevel.Stop()
    }
}

func (suite *TwoLevelCacheTestSuite) TestWriteThrough() {
    key := uint64(12345)
    value := []byte("test-value")
    expiration := time.Now().Add(10 * time.Minute)
    
    // Set using two-level store
    suite.twoLevelStore.Set(key, value, expiration)
    
    // Verify data exists in both L1 and L2
    l1Data, l1Found := suite.memoryStore.Get(key)
    l2Data, l2Found := suite.redisStore.Get(key)
    
    suite.True(l1Found, "Data should exist in L1 cache")
    suite.True(l2Found, "Data should exist in L2 cache")
    suite.Equal(value, l1Data, "L1 data should match")
    suite.Equal(value, l2Data, "L2 data should match")
}

func (suite *TwoLevelCacheTestSuite) TestL1Hit() {
    key := uint64(12345)
    value := []byte("test-value")
    expiration := time.Now().Add(10 * time.Minute)
    
    // Store in both levels
    suite.twoLevelStore.Set(key, value, expiration)
    
    // Get should hit L1 first
    result, found := suite.twoLevelStore.Get(key)
    
    suite.True(found, "Should find data")
    suite.Equal(value, result, "Should return correct data")
    
    // Check statistics
    if twoLevel, ok := suite.twoLevelStore.(*CacheTwoLevelStore); ok {
        stats := twoLevel.GetStats()
        suite.Equal(int64(1), stats.L1Hits, "Should have 1 L1 hit")
        suite.Equal(int64(0), stats.L2Hits, "Should have 0 L2 hits")
    }
}

func (suite *TwoLevelCacheTestSuite) TestL2HitWithCacheWarming() {
    key := uint64(12345)
    value := []byte("test-value")
    expiration := time.Now().Add(10 * time.Minute)
    
    // Store only in L2 (simulating L1 eviction)
    suite.redisStore.Set(key, value, expiration)
    
    // Get should hit L2 and warm L1
    result, found := suite.twoLevelStore.Get(key)
    
    suite.True(found, "Should find data in L2")
    suite.Equal(value, result, "Should return correct data")
    
    // Verify cache warming - data should now be in L1
    l1Data, l1Found := suite.memoryStore.Get(key)
    suite.True(l1Found, "Data should be warmed to L1")
    suite.Equal(value, l1Data, "L1 data should match")
    
    // Check statistics
    if twoLevel, ok := suite.twoLevelStore.(*CacheTwoLevelStore); ok {
        stats := twoLevel.GetStats()
        suite.Equal(int64(0), stats.L1Hits, "Should have 0 L1 hits")
        suite.Equal(int64(1), stats.L2Hits, "Should have 1 L2 hit")
    }
}

func (suite *TwoLevelCacheTestSuite) TestCacheMiss() {
    key := uint64(99999)
    
    // Get non-existent key
    result, found := suite.twoLevelStore.Get(key)
    
    suite.False(found, "Should not find data")
    suite.Nil(result, "Should return nil")
    
    // Check statistics
    if twoLevel, ok := suite.twoLevelStore.(*CacheTwoLevelStore); ok {
        stats := twoLevel.GetStats()
        suite.Equal(int64(1), stats.TotalMiss, "Should have 1 miss")
    }
}

func (suite *TwoLevelCacheTestSuite) TestRelease() {
    key := uint64(12345)
    value := []byte("test-value")
    expiration := time.Now().Add(10 * time.Minute)
    
    // Store data
    suite.twoLevelStore.Set(key, value, expiration)
    
    // Verify data exists
    _, found := suite.twoLevelStore.Get(key)
    suite.True(found, "Data should exist before release")
    
    // Release data
    suite.twoLevelStore.Release(key)
    
    // Verify data is removed from both levels
    _, l1Found := suite.memoryStore.Get(key)
    _, l2Found := suite.redisStore.Get(key)
    
    suite.False(l1Found, "Data should be removed from L1")
    suite.False(l2Found, "Data should be removed from L2")
}

func (suite *TwoLevelCacheTestSuite) TestWriteBackStrategy() {
    // Create write-back store
    writeBackStore := NewCacheTwoLevelStoreWithConfig(TwoLevelConfig{
        L1Store:  suite.memoryStore,
        L2Store:  suite.redisStore,
        Strategy: WriteBack,
        L1TTL:    5 * time.Minute,
        L2TTL:    30 * time.Minute,
    })
    defer func() {
        if twoLevel, ok := writeBackStore.(*CacheTwoLevelStore); ok {
            twoLevel.Stop()
        }
    }()
    
    key := uint64(12345)
    value := []byte("test-value")
    expiration := time.Now().Add(10 * time.Minute)
    
    // Set data
    writeBackStore.Set(key, value, expiration)
    
    // L1 should have data immediately
    l1Data, l1Found := suite.memoryStore.Get(key)
    suite.True(l1Found, "Data should exist in L1 immediately")
    suite.Equal(value, l1Data, "L1 data should match")
    
    // L2 might not have data immediately (async), but should have it soon
    time.Sleep(100 * time.Millisecond) // Give async operation time to complete
    
    l2Data, l2Found := suite.redisStore.Get(key)
    suite.True(l2Found, "Data should eventually exist in L2")
    suite.Equal(value, l2Data, "L2 data should match")
}

func (suite *TwoLevelCacheTestSuite) TestStats() {
    twoLevel, ok := suite.twoLevelStore.(*CacheTwoLevelStore)
    suite.True(ok, "Should be able to cast to TwoLevelStore")
    
    key1 := uint64(1)
    key2 := uint64(2)
    key3 := uint64(3)
    value := []byte("test")
    expiration := time.Now().Add(10 * time.Minute)
    
    // Set some data
    suite.twoLevelStore.Set(key1, value, expiration)
    suite.twoLevelStore.Set(key2, value, expiration)
    
    // Generate some hits and misses
    suite.twoLevelStore.Get(key1) // L1 hit
    suite.twoLevelStore.Get(key2) // L1 hit
    suite.twoLevelStore.Get(key3) // Miss
    
    // Remove key1 from L1 only (to test L2 hit)
    suite.memoryStore.Release(key1)
    suite.twoLevelStore.Get(key1) // L2 hit
    
    stats := twoLevel.GetStats()
    
    suite.Equal(int64(2), stats.L1Hits, "Should have 2 L1 hits")
    suite.Equal(int64(1), stats.L2Hits, "Should have 1 L2 hit")
    suite.Equal(int64(1), stats.TotalMiss, "Should have 1 miss")
    suite.Equal(int64(4), stats.TotalRequest, "Should have 4 total requests")
    suite.Equal(75.0, stats.HitRate, "Hit rate should be 75%")
}
```

### 4.5 example/two_level_example.go - 사용 예제

```go
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
    e.Use(middleware.Recover())
    
    // Create L1 cache (Memory)
    memoryStore := echo_http_cache.NewCacheMemoryStoreWithConfig(
        echo_http_cache.CacheMemoryStoreConfig{
            Capacity:  100,
            Algorithm: echo_http_cache.LRU,
        },
    )
    
    // Create L2 cache (Redis)
    redisStore := echo_http_cache.NewCacheRedisStoreWithConfig(redis.Options{
        Addr:     "localhost:6379",
        Password: "", // No password
        DB:       0,  // Default DB
    })
    
    // Create Two-Level Cache Store
    twoLevelStore := echo_http_cache.NewCacheTwoLevelStoreWithConfig(
        echo_http_cache.TwoLevelStoreConfig{
            L1Store:      memoryStore,
            L2Store:      redisStore,
            Strategy:     echo_http_cache.WriteThrough,
            L1TTL:        5 * time.Minute,  // Memory cache for 5 minutes
            L2TTL:        30 * time.Minute, // Redis cache for 30 minutes
            CacheWarming: true,
        },
    )
    
    // Setup cache middleware
    e.Use(echo_http_cache.CacheWithConfig(echo_http_cache.CacheConfig{
        Store:      twoLevelStore,
        Expiration: 10 * time.Minute,
        IncludePaths: []string{"/api/"},
    }))
    
    // API endpoints
    e.GET("/api/data", func(c echo.Context) error {
        log.Println("Generating new response...")
        return c.JSON(http.StatusOK, map[string]interface{}{
            "data":      "This response is cached in both memory and Redis",
            "timestamp": time.Now().Format(time.RFC3339),
            "cached":    false,
        })
    })
    
    e.GET("/api/user/:id", func(c echo.Context) error {
        userID := c.Param("id")
        log.Printf("Fetching user %s...", userID)
        return c.JSON(http.StatusOK, map[string]interface{}{
            "user_id":   userID,
            "name":      "User " + userID,
            "timestamp": time.Now().Format(time.RFC3339),
        })
    })
    
    // Cache management endpoints
    if twoLevel, ok := twoLevelStore.(*echo_http_cache.CacheTwoLevelStore); ok {
        // Get cache statistics
        e.GET("/cache/stats", func(c echo.Context) error {
            stats := twoLevel.GetStats()
            return c.JSON(http.StatusOK, stats)
        })
        
        // Clear L1 cache only
        e.DELETE("/cache/l1", func(c echo.Context) error {
            if err := twoLevel.ClearL1(); err != nil {
                return c.JSON(http.StatusInternalServerError, map[string]string{
                    "error": err.Error(),
                })
            }
            return c.JSON(http.StatusOK, map[string]string{
                "message": "L1 cache cleared successfully",
            })
        })
        
        // Clear L2 cache only
        e.DELETE("/cache/l2", func(c echo.Context) error {
            if err := twoLevel.ClearL2(); err != nil {
                return c.JSON(http.StatusInternalServerError, map[string]string{
                    "error": err.Error(),
                })
            }
            return c.JSON(http.StatusOK, map[string]string{
                "message": "L2 cache cleared successfully",
            })
        })
        
        // Clear all caches
        e.DELETE("/cache/all", func(c echo.Context) error {
            if err := twoLevel.ClearAll(); err != nil {
                return c.JSON(http.StatusInternalServerError, map[string]string{
                    "error": err.Error(),
                })
            }
            return c.JSON(http.StatusOK, map[string]string{
                "message": "All caches cleared successfully",
            })
        })
        
        // Reset statistics
        e.POST("/cache/stats/reset", func(c echo.Context) error {
            twoLevel.ResetStats()
            return c.JSON(http.StatusOK, map[string]string{
                "message": "Cache statistics reset successfully",
            })
        })
    }
    
    // Health check endpoint
    e.GET("/health", func(c echo.Context) error {
        return c.JSON(http.StatusOK, map[string]string{
            "status": "healthy",
            "time":   time.Now().Format(time.RFC3339),
        })
    })
    
    log.Println("Server starting on :8080...")
    log.Println("Available endpoints:")
    log.Println("  GET  /api/data        - Cached API endpoint")
    log.Println("  GET  /api/user/:id    - User data endpoint")
    log.Println("  GET  /cache/stats     - Cache statistics")
    log.Println("  DELETE /cache/l1      - Clear L1 cache")
    log.Println("  DELETE /cache/l2      - Clear L2 cache")
    log.Println("  DELETE /cache/all     - Clear all caches")
    log.Println("  POST /cache/stats/reset - Reset statistics")
    log.Println("  GET  /health          - Health check")
    
    log.Fatal(e.Start(":8080"))
}
```

### 4.6 기본 사용법 예제

```go
package main

import (
    "time"
    "github.com/go-redis/redis/v8"
    echo_http_cache "github.com/kenshin579/echo-http-cache"
    "github.com/labstack/echo/v4"
)

func main() {
    e := echo.New()
    
    // Two-Level Cache Store 생성
    memoryStore := echo_http_cache.NewCacheMemoryStoreWithConfig(
        echo_http_cache.CacheMemoryStoreConfig{
            Capacity:  100,
            Algorithm: echo_http_cache.LRU,
        },
    )
    
    redisStore := echo_http_cache.NewCacheRedisStoreWithConfig(redis.Options{
        Addr: "localhost:6379",
    })
    
    twoLevelStore := echo_http_cache.NewCacheTwoLevelStoreWithConfig(
        echo_http_cache.TwoLevelStoreConfig{
            L1Store: memoryStore,
            L2Store: redisStore,
            Strategy: echo_http_cache.WriteThrough,
            L1TTL: 5 * time.Minute,
            L2TTL: 30 * time.Minute,
            CacheWarming: true,
        },
    )
    
    // 캐시 미들웨어 설정
    e.Use(echo_http_cache.CacheWithConfig(echo_http_cache.CacheConfig{
        Store:      twoLevelStore,
        Expiration: 10 * time.Minute,
    }))
    
    e.GET("/api/data", func(c echo.Context) error {
        return c.JSON(200, map[string]interface{}{
            "data": "This is cached in both memory and Redis",
            "time": time.Now(),
        })
    })
    
    e.Start(":8080")
}
```

### 4.7 캐시 통계 모니터링
```go
e.GET("/cache/stats", func(c echo.Context) error {
    if twoLevelStore, ok := store.(*echo_http_cache.CacheTwoLevelStore); ok {
        stats := twoLevelStore.GetStats()
        return c.JSON(200, stats)
    }
    return c.JSON(400, map[string]string{"error": "Not a two-level store"})
})
```

## 5. 구현 순서

### 5.1 Phase 1: 기본 구현 (1-2주)

1. **캐시 전략 및 구성 구조체 정의**
   ```bash
   touch cache_strategy.go
   # cache_strategy.go 구현
   ```

2. **통계 수집 구조체 구현**
   ```bash
   touch cache_stats.go
   # cache_stats.go 구현
   ```

3. **Two-Level Store 기본 구현**
   ```bash
   touch two_level_cache.go
   # CacheTwoLevelStore 구조체 및 기본 메서드 구현
   ```

4. **기본 테스트 작성**
   ```bash
   touch two_level_cache_test.go
   # 기본 기능 테스트 구현
   ```

### 5.2 Phase 2: 고급 기능 (1주)

1. **Write-Back 전략의 비동기 처리 구현**
2. **Cache Warming 로직 최적화**
3. **에러 처리 및 복구 로직 구현**
4. **성능 테스트 및 벤치마크 작성**

### 5.3 Phase 3: 운영 기능 (3-5일)

1. **캐시 관리 API 구현**
2. **모니터링 및 통계 기능 완성**
3. **사용 예제 및 문서 작성**
4. **통합 테스트 실행**

## 6. 테스트 방법

### 6.1 단위 테스트 실행
```bash
go test -v ./... -run TestTwoLevelCache
```

### 6.2 성능 테스트 실행
```bash
go test -v -bench=BenchmarkTwoLevelCache -benchmem
```

### 6.3 예제 애플리케이션 테스트
```bash
# 1. Redis 서버 시작
redis-server

# 2. 예제 실행
cd example
go run two_level_example.go

# 3. API 테스트
curl http://localhost:8080/api/data
curl http://localhost:8080/cache/stats
```

## 7. 주의사항

### 7.1 동시성 처리
- L1, L2 캐시 간 데이터 일관성 보장
- 고루틴 안전성 확보
- 비동기 작업의 적절한 에러 처리

### 7.2 메모리 관리
- L1 캐시 용량 제한 준수
- 메모리 누수 방지
- 적절한 TTL 설정

### 7.3 성능 최적화
- 불필요한 복사 작업 최소화
- 적절한 버퍼 크기 설정
- 통계 수집 오버헤드 최소화

이 구현 가이드를 따르면 효율적이고 안정적인 Two-Level Caching 시스템을 구축할 수 있습니다. 