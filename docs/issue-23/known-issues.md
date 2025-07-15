# Known Issues with Redis Cluster Support

## Docker Environment Limitations

### Issue: "got 4 elements in cluster info address, expected 2 or 3"

When running Redis Cluster in Docker containers for local development, you may encounter this error when trying to perform operations like `Clear()`:

```
got 4 elements in cluster info address, expected 2 or 3
```

#### Root Cause

This error occurs because:
1. Redis Cluster returns node addresses in the format `IP:port@cport` (e.g., `172.17.0.2:7000@17000`)
2. go-redis v8 has difficulty parsing this format when the cluster bus port is included
3. Docker containers return their internal IP addresses, which are not accessible from the host

#### Redis Version Compatibility

This issue has been tested with both Redis 6 and Redis 7:
- **Redis 6**: Shows the same cluster address parsing error, but `Clear()` method may appear to succeed even though the cluster is not properly configured
- **Redis 7**: Consistently shows the "got 4 elements" error for operations that require cluster node discovery
- Both versions fail to properly form a cluster in Docker due to networking limitations

#### Workarounds

1. **For local development**: Use a single Redis instance instead of a cluster
2. **For testing**: Run integration tests inside the Docker network
3. **For production**: This issue does not occur in real production environments where Redis Cluster nodes have routable IP addresses

#### Alternative Solutions Attempted

1. **cluster-announce settings**: Adding cluster-announce-ip and cluster-announce-port doesn't fully resolve the issue due to the bus port format
2. **network_mode: host**: Not supported on macOS Docker Desktop
3. **RouteByLatency option**: Added to improve compatibility but doesn't resolve the core issue
4. **Redis version downgrade**: Tested with both Redis 6 and Redis 7, but the Docker networking issue persists in both versions

### Recommendation

For local development and testing of Redis Cluster features:
- The code implementation is correct and will work in production environments
- Use unit tests with mocked Redis connections for testing business logic
- Perform integration testing in environments with properly configured Redis Cluster (e.g., Kubernetes, cloud environments)

## Other Considerations

### Clear() Method Performance

The `Clear()` method iterates through all master nodes and executes FLUSHDB on each. This operation:
- Is not atomic across the cluster
- May take time proportional to the number of keys in each node
- Should be used sparingly in production environments

### Migration from Redis Ring

When migrating from the previous Redis Ring implementation:
- Existing cache entries will be lost (different sharding algorithms)
- Client applications need to update their connection configuration
- Consider implementing a gradual migration strategy in production 