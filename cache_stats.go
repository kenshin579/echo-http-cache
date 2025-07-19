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
	"sync/atomic"
	"time"
)

// CacheStats represents cache statistics
type CacheStats struct {
	L1Hits       int64     `json:"l1Hits"`
	L2Hits       int64     `json:"l2Hits"`
	TotalMiss    int64     `json:"totalMiss"`
	TotalRequest int64     `json:"totalRequest"`
	HitRate      float64   `json:"hitRate"`
	L1HitRate    float64   `json:"l1HitRate"`
	L2HitRate    float64   `json:"l2HitRate"`
	L1Size       int       `json:"l1Size"`
	L2Size       int       `json:"l2Size"`
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
