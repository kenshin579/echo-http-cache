# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an Echo HTTP middleware library for server-side caching of REST API responses. It's a fork of echo-http-cache that adds Redis Cluster support and two-level caching capabilities.

## Development Commands

### Testing

```bash
# Run all tests
go test -v ./...

# Run specific test
go test -v -run TestCache

# Run integration tests (requires Redis Cluster)
docker-compose up -d
go test -v -run TestCacheRedisClusterStore
docker-compose down

# Run benchmarks
go test -bench=. -benchmem

# Run two-level cache benchmarks
go test -bench=BenchmarkTwoLevel -benchmem
```

### Building

```bash
# Build the library
go build

# Run example applications
cd example
go run redis_cluster_example.go

cd example/two_level
go run main.go
```

### Docker Environment

```bash
# Start Redis Cluster (6 nodes: 3 masters, 3 replicas)
docker-compose up -d

# Stop Redis Cluster
docker-compose down

# Cluster ports: 17000-17005 (mapped from internal 7000-7005)
```

## Architecture

### Core Cache System

The middleware intercepts GET requests and caches responses based on URL and query parameters:

1. **cache.go** - Core middleware implementation
   - `CacheWithConfig()` creates the Echo middleware
   - URL params are sorted for consistent cache keys
   - Cache keys generated via FNV-64a hash of `{method}:{url}`
   - Only successful responses (< 400 status) are cached
   - Empty response bodies are not cached

2. **CacheStore Interface** - All store implementations follow this:
   ```go
   type CacheStore interface {
       Get(key uint64) ([]byte, bool)
       Set(key uint64, response []byte, expiration time.Time)
       Release(key uint64)
   }
   ```

### Store Implementations

1. **memory.go** - In-memory cache with eviction algorithms
   - Supports LRU, LFU, MRU, MFU algorithms
   - Thread-safe with sync.Mutex
   - Fixed capacity with eviction on overflow

2. **redis.go** - Redis Standalone
   - Uses `go-redis/cache/v8` library
   - Simple key-value caching

3. **redis_cluster.go** - Redis Cluster (NEW)
   - True Redis Cluster support (not Ring)
   - Uses `redis.ClusterClient` with 16384 hash slots
   - `Clear()` method flushes all master nodes (use sparingly)

### Two-Level Caching (NEW)

**two_level_cache.go** - L1 (memory) + L2 (Redis) with multiple strategies:

- **WriteThrough** - Writes to both L1 and L2 synchronously
- **WriteBack** - Writes to L1 immediately, L2 asynchronously via channel
- **CacheAside** - Application-managed caching

Features:
- **Cache Warming** - L2 hits are promoted to L1 (async by default)
- **Async Operations** - Buffered channel for non-blocking L2 writes
- **Smart TTLs** - Different expiration times for L1 (5m) and L2 (30m)
- **Statistics** - Tracks L1 hits, L2 hits, misses via `cache_stats.go`

The async worker goroutine processes operations from `asyncChan`:
- Must call `Stop()` on shutdown to gracefully drain the channel
- Falls back to sync writes if channel is full

### Configuration

**CacheConfig** fields:
- `Store` - Which CacheStore implementation to use
- `Expiration` - Default TTL for all paths
- `IncludePaths` - Whitelist paths to cache
- `IncludePathsWithExpiration` - Per-path TTLs (higher priority)
- `ExcludePaths` - Blacklist paths from caching

Path matching uses `strings.Contains()`, so partial matches work.

## Testing Strategy

- Unit tests use `miniredis` for mocking Redis
- Integration tests require actual Redis Cluster (via docker-compose)
- Tests with `_integration_test.go` suffix need Docker
- Example tests in `example/example_test.go` and `example/two_level/e2e_test.go`

## Known Issues

Redis Cluster has Docker networking limitations on macOS/Windows that may affect local testing. This is a Docker issue, not an implementation issue. See docs for details. Production environments are unaffected.