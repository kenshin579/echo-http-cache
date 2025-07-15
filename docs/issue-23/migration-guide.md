# Redis Ring to Redis Cluster Migration Guide

## 개요

이 가이드는 기존 Redis Ring (client-side sharding) 구현에서 실제 Redis Cluster (server-side sharding) 구현으로 마이그레이션하는 방법을 설명합니다.

## 주요 차이점

### Redis Ring (기존)
- Client-side sharding
- 수동 노드 관리
- 노드 장애 시 수동 대응 필요
- `redis.RingOptions` 사용

### Redis Cluster (신규)
- Server-side sharding (16384 슬롯)
- 자동 페일오버
- 노드 추가/제거 시 자동 리밸런싱
- `redis.ClusterOptions` 사용

## 마이그레이션 단계

### 1. 코드 변경

#### 기존 코드 (Redis Ring)
```go
import (
    "github.com/go-redis/redis/v8"
    echocache "github.com/kenshin579/echo-http-cache"
)

// Redis Ring 설정
store := echocache.NewCacheRedisClusterWithConfig(redis.RingOptions{
    Addrs: map[string]string{
        "shard1": "localhost:7000",
        "shard2": "localhost:7001",
        "shard3": "localhost:7002",
    },
    Password: "password",
    DB:       0,
})
```

#### 새로운 코드 (Redis Cluster)
```go
import (
    "github.com/go-redis/redis/v8"
    echocache "github.com/kenshin579/echo-http-cache"
)

// Redis Cluster 설정
store := echocache.NewCacheRedisClusterStoreWithConfig(redis.ClusterOptions{
    Addrs: []string{
        "localhost:7000",
        "localhost:7001",
        "localhost:7002",
        "localhost:7003",
        "localhost:7004",
        "localhost:7005",
    },
    Password: "password", // 모든 노드에 동일한 비밀번호 사용
})
```

### 2. 함수명 변경 사항

| 기존 (Redis Ring) | 신규 (Redis Cluster) |
|-------------------|---------------------|
| `NewCacheRedisClusterWithConfig` | `NewCacheRedisClusterStoreWithConfig` |

### 3. 옵션 매핑

#### 기본 옵션
- `Addrs`: map[string]string → []string 형식으로 변경
- `Password`: 동일하게 사용
- `PoolSize`: 동일하게 사용

#### 제거된 옵션
- `DB`: Redis Cluster는 database 선택을 지원하지 않음 (항상 DB 0 사용)
- `HashReplicas`: Server-side sharding이므로 불필요

#### 새로운 옵션 (선택사항)
- `MaxRedirects`: 클러스터 리다이렉트 최대 횟수 (기본값: 8)
- `ReadOnly`: 읽기 전용 복제본 사용 여부
- `RouteByLatency`: 레이턴시 기반 라우팅
- `RouteRandomly`: 랜덤 라우팅

### 4. Clear() 메서드 사용

Redis Cluster에서 캐시를 전체 삭제하려면:

```go
if redisStore, ok := store.(*echocache.CacheRedisClusterStore); ok {
    err := redisStore.Clear()
    if err != nil {
        log.Printf("Failed to clear cache: %v", err)
    }
}
```

⚠️ **주의**: Clear()는 모든 마스터 노드에서 FLUSHDB를 실행합니다. 프로덕션에서는 신중히 사용하세요.

## 마이그레이션 체크리스트

- [ ] Redis Cluster 환경 준비 (최소 3개 마스터 노드)
- [ ] 코드에서 함수명 변경
- [ ] RingOptions → ClusterOptions 변경
- [ ] Addrs 형식 변경 (map → slice)
- [ ] DB 옵션 제거 (있는 경우)
- [ ] 테스트 환경에서 동작 확인
- [ ] 모니터링 설정 업데이트
- [ ] 프로덕션 배포

## 트러블슈팅

### 1. "MOVED" 에러 발생
- 원인: 클라이언트가 잘못된 노드에 접근
- 해결: ClusterOptions의 모든 노드 주소가 올바른지 확인

### 2. "CROSSSLOT" 에러 발생
- 원인: 하나의 명령에서 여러 슬롯의 키 접근
- 해결: 캐시 키 설계 검토

### 3. 연결 실패
- 원인: 방화벽 또는 네트워크 설정
- 해결: 모든 클러스터 노드 포트(기본 7000-7005) 접근 가능 확인

## 성능 고려사항

1. **연결 풀 크기**: ClusterOptions.PoolSize 적절히 설정
2. **읽기 부하 분산**: ReadOnly 옵션 활용 고려
3. **네트워크 레이턴시**: 같은 가용 영역에 클러스터 구성

## 참고 자료

- [Redis Cluster Tutorial](https://redis.io/topics/cluster-tutorial)
- [go-redis Cluster Documentation](https://pkg.go.dev/github.com/go-redis/redis/v8#ClusterClient)
- [Echo HTTP Cache Documentation](https://github.com/kenshin579/echo-http-cache) 