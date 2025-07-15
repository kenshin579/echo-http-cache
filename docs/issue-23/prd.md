# Redis Cluster 지원 개선 PRD

## 1. 개요

### 1.1 배경
현재 echo-http-cache 라이브러리는 Redis Cluster를 지원한다고 명시되어 있으나, 실제로는 Redis Ring (client-side sharding)을 사용하고 있습니다. 이는 실제 Redis Cluster가 아닙니다.

### 1.2 목적
go-redis의 ClusterClient를 사용하여 실제 Redis Cluster에 캐시 데이터를 저장하고 조회할 수 있도록 구현합니다.

## 2. 현재 상태 분석

### 2.1 기존 구현
```go
// 현재 구현 (redis_cluster.go)
func NewCacheRedisClusterWithConfig(opt redis.RingOptions) CacheStore {
    return &CacheRedisClusterStore{
        redisCache.New(&redisCache.Options{
            Redis: redis.NewRing(&opt),  // Ring 사용 (문제점)
        }),
    }
}
```

### 2.2 문제점
- Redis Ring은 Redis Cluster가 아님
- 실제 Redis Cluster 환경에서 사용 불가

## 3. 요구사항

### 3.1 기능 요구사항

#### 3.1.1 Redis Cluster 지원
- `redis.ClusterClient`를 사용한 실제 Redis Cluster 연결
- 기본적인 캐시 저장/조회 기능 구현

#### 3.1.2 기존 인터페이스 호환성
- `CacheStore` 인터페이스 구현 유지
- 기존 API와의 호환성 보장

### 3.2 구현 범위
다음 메서드들을 Redis Cluster에서 동작하도록 구현:
- `Get(key uint64) ([]byte, bool)`
- `Set(key uint64, response []byte, expiration time.Time) error`
- `Delete(key uint64) error`
- `Clear() error` (모든 캐시 삭제)

## 4. 구현 계획

### 4.1 접근 방법
- 새로운 `CacheRedisClusterStoreV2` 구조체 생성
- `redis.ClusterClient` 사용하여 실제 Redis Cluster 연결
- 기존 CacheStore 인터페이스 메서드 구현

### 4.2 구현 단계

1. **기본 구현 (1주)**
   - 새로운 Store 구조체 구현
   - CacheStore 인터페이스 메서드 구현
   - 기본 테스트 작성

2. **테스트 및 문서화 (3일)**
   - 단위 테스트 작성
   - README 업데이트
   - 사용 예제 추가

> 구체적인 구현 코드는 [implementation.md](./implementation.md) 참조

## 5. 테스트 계획

### 5.1 단위 테스트
- Get/Set/Delete/Clear 메서드 테스트
- 키가 존재하지 않을 때의 동작 테스트
- TTL 동작 테스트

### 5.2 통합 테스트
- Docker Compose를 사용한 6노드 Redis Cluster 환경 구축
- 실제 클러스터 환경에서 동작 확인

> 테스트 환경 설정은 [implementation.md](./implementation.md#5-docker-compose-테스트-환경) 참조

## 6. 마이그레이션 가이드

### 6.1 기존 Ring 사용자
기존 Ring 기반 코드에서 Cluster 기반 코드로 전환:
- 기존: `redis.RingOptions` 사용
- 신규: `redis.ClusterOptions` 사용

### 6.2 호환성
- 기존 함수는 deprecated로 표시하되 유지
- 새로운 함수명 사용 권장

## 7. 예상 효과

### 7.1 장점
- 실제 Redis Cluster 사용 가능
- 표준적인 Redis Cluster 운영 환경 지원

### 7.2 제한사항
- 클러스터 구성에 최소 3개의 마스터 노드 필요
- 단일 Redis 인스턴스보다 설정이 복잡

## 8. 타임라인

- 구현: 1주
- 테스트 및 문서화: 3일
- 총 예상 기간: 1.5주

## 9. 참고 자료

- [go-redis ClusterClient Documentation](https://pkg.go.dev/github.com/redis/go-redis/v9#ClusterClient)
- [Redis Cluster Tutorial](https://redis.io/topics/cluster-tutorial)
- [구현 가이드](./implementation.md)
