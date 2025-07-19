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
	// Sync - synchronous operations
	Sync SyncMode = "SYNC"

	// Async - asynchronous operations where possible
	Async SyncMode = "ASYNC"
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
	SyncMode:     Async,
	AsyncBuffer:  1000,
}
