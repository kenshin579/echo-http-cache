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

// Algorithm is the string type for caching algorithms labels.
type Algorithm string

const (
	// LRU is the constant for Least Recently Used.
	LRU Algorithm = "LRU"

	// MRU is the constant for Most Recently Used.
	MRU Algorithm = "MRU"

	// LFU is the constant for Least Frequently Used.
	LFU Algorithm = "LFU"

	// MFU is the constant for Most Frequently Used.
	MFU Algorithm = "MFU"
)

type (
	// CacheMemoryStore is the built-in store implementation for Cache
	CacheMemoryStore struct {
		mutex     sync.RWMutex
		capacity  int
		algorithm Algorithm
		store     map[uint64][]byte
	}
)

func NewCacheMemoryStore() (store *CacheMemoryStore) {
	return NewCacheMemoryStoreWithConfig(CacheMemoryStoreConfig{
		Capacity:  10,
		Algorithm: LFU,
	})
}

func NewCacheMemoryStoreWithConfig(config CacheMemoryStoreConfig) (store *CacheMemoryStore) {
	store = &CacheMemoryStore{}
	store.capacity = config.Capacity
	store.algorithm = config.Algorithm

	if config.Capacity == 0 {
		store.capacity = DefaultCacheMemoryStoreConfig.Capacity
	}
	if config.Algorithm == "" {
		store.algorithm = DefaultCacheMemoryStoreConfig.Algorithm
	}
	store.mutex = sync.RWMutex{}
	store.store = make(map[uint64][]byte, store.capacity)
	return store
}

// CacheMemoryStoreConfig represents configuration for CacheMemoryStoreConfig
type CacheMemoryStoreConfig struct {
	Capacity  int
	Algorithm Algorithm
}

// DefaultCacheMemoryStoreConfig provides default configuration values for CacheMemoryStoreConfig
var DefaultCacheMemoryStoreConfig = CacheMemoryStoreConfig{
	Capacity:  10,
	Algorithm: LRU,
}

// Get implements the cache Adapter interface Get method.
func (store *CacheMemoryStore) Get(key uint64) ([]byte, bool) {
	store.mutex.RLock()
	response, ok := store.store[key]
	store.mutex.RUnlock()

	if ok {
		return response, true
	}
	return nil, false
}

// Set implements the cache Adapter interface Set method.
func (store *CacheMemoryStore) Set(key uint64, response []byte, _ time.Time) {
	store.mutex.RLock()
	length := len(store.store)
	store.mutex.RUnlock()

	if length > 0 && length == store.capacity {
		store.evict()
	}

	store.mutex.Lock()
	store.store[key] = response
	store.mutex.Unlock()
}

// Release implements the Adapter interface Release method.
func (store *CacheMemoryStore) Release(key uint64) {
	store.mutex.RLock()
	_, ok := store.store[key]
	store.mutex.RUnlock()

	if ok {
		store.mutex.Lock()
		delete(store.store, key)
		store.mutex.Unlock()
	}
}

func (store *CacheMemoryStore) evict() {
	selectedKey := uint64(0)
	lastAccess := time.Now()
	frequency := 2147483647

	if store.algorithm == MRU {
		lastAccess = time.Time{}
	} else if store.algorithm == MFU {
		frequency = 0
	}

	for k, v := range store.store {
		r := toCacheResponse(v)
		switch store.algorithm {
		case LRU:
			if r.LastAccess.Before(lastAccess) {
				selectedKey = k
				lastAccess = r.LastAccess
			}
		case MRU:
			if r.LastAccess.After(lastAccess) ||
				r.LastAccess.Equal(lastAccess) {
				selectedKey = k
				lastAccess = r.LastAccess
			}
		case LFU:
			if r.Frequency < frequency {
				selectedKey = k
				frequency = r.Frequency
			}
		case MFU:
			if r.Frequency >= frequency {
				selectedKey = k
				frequency = r.Frequency
			}
		}
	}

	store.Release(selectedKey)
}
