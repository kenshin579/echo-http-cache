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
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type TwoLevelCacheTestSuite struct {
	suite.Suite
	memoryStore   CacheStore
	redisStore    CacheStore
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

func (suite *TwoLevelCacheTestSuite) TestClearMethods() {
	twoLevel, ok := suite.twoLevelStore.(*CacheTwoLevelStore)
	suite.True(ok, "Should be able to cast to TwoLevelStore")

	key1 := uint64(12345)
	key2 := uint64(67890)
	value := []byte("test-value")
	expiration := time.Now().Add(10 * time.Minute)

	// Test ClearL1
	suite.twoLevelStore.Set(key1, value, expiration)
	_, found := suite.twoLevelStore.Get(key1)
	suite.True(found, "Data should exist before ClearL1")

	err := twoLevel.ClearL1()
	suite.NoError(err, "ClearL1 should not return error")

	// Data should still be available from L2 (Write-Through strategy)
	_, found = suite.twoLevelStore.Get(key1)
	suite.True(found, "Data should still exist in L2 after ClearL1")

	// Test ClearL2
	suite.twoLevelStore.Set(key2, value, expiration)
	err = twoLevel.ClearL2()
	suite.NoError(err, "ClearL2 should not return error")

	// Data should still be available from L1
	_, found = suite.twoLevelStore.Get(key2)
	suite.True(found, "Data should still exist in L1 after ClearL2")

	// Test ClearAll
	err = twoLevel.ClearAll()
	suite.NoError(err, "ClearAll should not return error")

	_, found = suite.twoLevelStore.Get(key1)
	suite.False(found, "Data should not exist after ClearAll")

	_, found = suite.twoLevelStore.Get(key2)
	suite.False(found, "Data should not exist after ClearAll")

	// Test ResetStats
	twoLevel.ResetStats()
	stats := twoLevel.GetStats()
	suite.Equal(int64(0), stats.TotalRequest, "Stats should be reset")
}
