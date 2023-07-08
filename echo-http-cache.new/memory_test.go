package cache

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	store := &CacheMemoryStore{
		sync.RWMutex{},
		2,
		LRU,
		map[uint64][]byte{
			14974843192121052621: CacheResponse{
				Value:      []byte("value 1"),
				Expiration: time.Now(),
				LastAccess: time.Now(),
				Frequency:  1,
			}.bytes(),
		},
	}

	tests := []struct {
		name string
		key  uint64
		want []byte
		ok   bool
	}{
		{
			"returns right response",
			14974843192121052621,
			[]byte("value 1"),
			true,
		},
		{
			"not found",
			123,
			nil,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, ok := store.Get(tt.key)
			assert.Equal(t, tt.ok, ok)

			got := bytesToResponse(b).Value
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSet(t *testing.T) {
	store := &CacheMemoryStore{
		sync.RWMutex{},
		2,
		LRU,
		make(map[uint64][]byte),
	}

	tests := []struct {
		name     string
		key      uint64
		response CacheResponse
	}{
		{
			"sets response cache",
			1,
			CacheResponse{
				Value:      []byte("value 1"),
				Expiration: time.Now().Add(1 * time.Minute),
			},
		},
		{
			"sets response cache",
			2,
			CacheResponse{
				Value:      []byte("value 2"),
				Expiration: time.Now().Add(1 * time.Minute),
			},
		},
		{
			"sets response cache",
			3,
			CacheResponse{
				Value:      []byte("value 3"),
				Expiration: time.Now().Add(1 * time.Minute),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store.Set(tt.key, tt.response.bytes(), tt.response.Expiration)
			if bytesToResponse(store.store[tt.key]).Value == nil {
				t.Errorf(
					"memory.Set() error = store[%v] response is not %s", tt.key, tt.response.Value,
				)
			}
		})
	}
}

func TestRelease(t *testing.T) {
	store := &CacheMemoryStore{
		sync.RWMutex{},
		2,
		LRU,
		map[uint64][]byte{
			14974843192121052621: CacheResponse{
				Expiration: time.Now().Add(1 * time.Minute),
				Value:      []byte("value 1"),
			}.bytes(),
			14974839893586167988: CacheResponse{
				Expiration: time.Now(),
				Value:      []byte("value 2"),
			}.bytes(),
			14974840993097796199: CacheResponse{
				Expiration: time.Now(),
				Value:      []byte("value 3"),
			}.bytes(),
		},
	}

	tests := []struct {
		name        string
		key         uint64
		storeLength int
		wantErr     bool
	}{
		{
			"removes cached response from store",
			14974843192121052621,
			2,
			false,
		},
		{
			"removes cached response from store",
			14974839893586167988,
			1,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store.Release(tt.key)
			if len(store.store) > tt.storeLength {
				t.Errorf("memory.Release() error; store length = %v, want 0", len(store.store))
			}
		})
	}
}

func TestEvict(t *testing.T) {
	tests := []struct {
		name      string
		algorithm Algorithm
	}{
		{
			"lru removes third cached response",
			LRU,
		},
		{
			"mru removes first cached response",
			MRU,
		},
		{
			"lfu removes second cached response",
			LFU,
		},
		{
			"mfu removes third cached response",
			MFU,
		},
	}
	count := 0
	for _, tt := range tests {
		count++

		store := &CacheMemoryStore{
			sync.RWMutex{},
			2,
			tt.algorithm,
			map[uint64][]byte{
				14974843192121052621: CacheResponse{
					Value:      []byte("value 1"),
					Expiration: time.Now().Add(1 * time.Minute),
					LastAccess: time.Now().Add(-1 * time.Minute),
					Frequency:  2,
				}.bytes(),
				14974839893586167988: CacheResponse{
					Value:      []byte("value 2"),
					Expiration: time.Now().Add(1 * time.Minute),
					LastAccess: time.Now().Add(-2 * time.Minute),
					Frequency:  1,
				}.bytes(),
				14974840993097796199: CacheResponse{
					Value:      []byte("value 3"),
					Expiration: time.Now().Add(1 * time.Minute),
					LastAccess: time.Now().Add(-3 * time.Minute),
					Frequency:  3,
				}.bytes(),
			},
		}
		t.Run(tt.name, func(t *testing.T) {
			store.evict()

			if count == 1 {
				if _, ok := store.store[14974840993097796199]; ok {
					t.Errorf("lru is not working properly")
					return
				}
			} else if count == 2 {
				if _, ok := store.store[14974843192121052621]; ok {
					t.Errorf("mru is not working properly")
					return
				}
			} else if count == 3 {
				if _, ok := store.store[14974839893586167988]; ok {
					t.Errorf("lfu is not working properly")
					return
				}
			} else {
				if count == 4 {
					if _, ok := store.store[14974840993097796199]; ok {
						t.Errorf("mfu is not working properly")
					}
				}
			}
		})
	}
}
