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
	operation  string
	key        uint64
	data       []byte
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

// ResetStats resets cache statistics
func (store *CacheTwoLevelStore) ResetStats() {
	store.metrics.Reset()
}
