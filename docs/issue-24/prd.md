# Two-Level Caching 지원 PRD

## 1. 개요

### 1.1 배경
현재 echo-http-cache 라이브러리는 단일 캐시 스토어만 지원하여, 같은 API path에 대해 local memory와 redis를 동시에 사용할 수 없습니다. 이로 인해 다음과 같은 제약사항이 있습니다:

- 빠른 응답을 위한 메모리 캐시와 영속성을 위한 Redis 캐시를 동시에 활용할 수 없음
- 캐시 히트율 최적화를 위한 Multi-level caching 전략 구현 불가
- 네트워크 지연을 최소화하면서도 데이터 영속성을 보장하는 하이브리드 캐시 구조 부재

### 1.2 현재 구조의 한계점
```go
// 현재 CacheConfig 구조
type CacheConfig struct {
    Store CacheStore  // 단일 스토어만 지원
    // ...
}
```

- `CacheConfig`가 단일 `Store` 필드만 보유
- 하나의 캐시 미들웨어 인스턴스당 하나의 스토어만 설정 가능
- Multi-level caching을 위한 설계 부재

### 1.3 목적
Local Memory (L1)와 Redis (L2)를 계층적으로 사용하는 Two-Level Caching 시스템을 구현하여:
- **성능 향상**: 메모리 캐시의 빠른 응답 속도 활용
- **영속성 보장**: Redis를 통한 데이터 영속성 및 분산 캐시 지원
- **캐시 효율성**: 계층적 캐시 구조로 전체 캐시 히트율 향상

## 2. 요구사항

### 2.1 기능 요구사항

#### 2.1.1 Two-Level Store 구현
```go
type CacheTwoLevelStore struct {
    l1Store CacheStore // Memory cache (L1)
    l2Store CacheStore // Redis cache (L2)
    config  TwoLevelConfig
}

type TwoLevelConfig struct {
    Strategy        CacheStrategy
    L1TTL          time.Duration
    L2TTL          time.Duration
    CacheWarming   bool
    SyncMode       SyncMode
}
```

#### 2.1.2 캐시 전략 지원
- **Cache-Aside**: 애플리케이션이 캐시 관리
- **Write-Through**: 쓰기 시 L1, L2 동시 업데이트
- **Write-Back**: L1에 먼저 쓰고 비동기적으로 L2 업데이트
- **Cache Warming**: L2에서 L1으로 데이터 승격



### 2.2 설정 요구사항

#### 2.2.1 기본 설정 지원
- 간단한 생성자 함수 제공
- 기본값으로 즉시 사용 가능한 설정

#### 2.2.2 고급 설정 지원
- 사용자 정의 설정으로 세밀한 제어 가능
- 캐시 전략, TTL, 버퍼 크기 등 조정 가능

### 2.3 호환성 요구사항

#### 2.3.1 기존 인터페이스 호환성
- 기존 `CacheStore` 인터페이스 구현 유지
- 기존 미들웨어 API와 완전 호환
- 기존 단일 스토어 방식도 계속 지원

#### 2.3.2 설정 호환성
- 기존 단일 스토어 방식과 동일한 API 사용
- 기존 코드 변경 없이 투레벨 스토어로 교체 가능

### 2.4 성능 요구사항

#### 2.4.1 응답 시간
- L1 캐시 히트: < 1ms
- L2 캐시 히트: < 10ms
- 전체 캐시 미스율: < 5% (기존 대비)

#### 2.4.2 메모리 사용량
- L1 캐시 용량 제한 설정 가능
- 메모리 사용량 모니터링 기능
- LRU/LFU 등 기존 알고리즘 지원

### 2.5 운영 요구사항

#### 2.5.1 모니터링
```go
type CacheStats struct {
    L1Hits     int64
    L2Hits     int64
    TotalMiss  int64
    HitRate    float64
    L1Size     int
    L2Size     int
}

func (store *CacheTwoLevelStore) GetStats() CacheStats
```

#### 2.5.2 캐시 관리
```go
// 개별 레벨 제어
func (store *CacheTwoLevelStore) ClearL1() error
func (store *CacheTwoLevelStore) ClearL2() error
func (store *CacheTwoLevelStore) ClearAll() error

// 캐시 동기화
func (store *CacheTwoLevelStore) SyncL1ToL2() error
func (store *CacheTwoLevelStore) SyncL2ToL1() error
```

## 3. 구현 계획

### 3.1 Phase 1: 기본 구현 (2주)
- `CacheTwoLevelStore` 구조체 구현
- 기본 Get/Set/Release 메서드 구현
- Write-Through 전략 구현
- 기본 테스트 작성

### 3.2 Phase 2: 고급 기능 (1주)
- Cache Warming 기능 구현
- Write-Back 전략 구현
- 비동기 처리 로직 구현
- 성능 테스트 및 최적화

### 3.3 Phase 3: 모니터링 및 관리 (1주)
- 통계 수집 기능 구현
- 캐시 관리 API 구현
- 문서화 및 예제 작성

## 4. 기술적 고려사항

### 4.1 동시성 처리
- L1, L2 간 데이터 일관성 보장
- 고루틴 안전성 확보
- 데드락 방지

### 4.2 에러 처리
- L1 실패 시 L2 fallback
- L2 실패 시 L1 지속 사용
- 부분적 장애 대응

### 4.3 메모리 관리
- L1 캐시 용량 제한
- 메모리 누수 방지
- GC 최적화

## 5. 성공 지표

### 5.1 성능 지표
- 캐시 히트율: 95% 이상
- P99 응답시간: 10ms 이하
- 메모리 사용량: 기존 대비 150% 이하

### 5.2 기능 지표
- 기존 API 100% 호환성 유지
- 새로운 기능 100% 테스트 커버리지
- 제로 다운타임 배포 지원

### 5.3 운영 지표
- 장애 복구 시간: 30초 이하
- 모니터링 대시보드 제공
- 상세한 문서 및 예제 제공
