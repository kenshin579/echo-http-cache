# Two-Level Caching 구현 TODO

## 📋 프로젝트 개요

**목표**: echo-http-cache 라이브러리에 Local Memory (L1) + Redis (L2) Two-Level Caching 기능 구현

**예상 기간**: 4-5주
- Phase 1: 기본 구현 (2주)
- Phase 2: 고급 기능 (1주)  
- Phase 3: 운영 기능 (1주)
- Phase 4: 문서화 및 마무리 (1주)

---

## 🚀 Phase 1: 기본 구현 (2주)

### 📁 파일 구조 및 기본 설정

- [ ] **파일 생성**
  - [ ] `cache_strategy.go` 생성
  - [ ] `cache_stats.go` 생성
  - [ ] `two_level_cache.go` 생성
  - [ ] `two_level_cache_test.go` 생성
  - [ ] `example/two_level_example.go` 생성

### 🎯 cache_strategy.go 구현

- [ ] **타입 정의**
  - [ ] `CacheStrategy` string 타입 정의
  - [ ] `SyncMode` string 타입 정의
  - [ ] 상수 정의 (WRITE_THROUGH, WRITE_BACK, CACHE_ASIDE)
  - [ ] 상수 정의 (SYNC, ASYNC)

- [ ] **설정 구조체**
  - [ ] `TwoLevelConfig` 구조체 정의
    - [ ] L1Store, L2Store 필드
    - [ ] Strategy, L1TTL, L2TTL 필드
    - [ ] CacheWarming, SyncMode, AsyncBuffer 필드
  - [ ] `DefaultTwoLevelConfig` 기본값 정의

### 📊 cache_stats.go 구현

- [ ] **통계 구조체**
  - [ ] `CacheStats` 구조체 정의 (camelCase JSON 태그)
  - [ ] `CacheMetrics` 구조체 정의 (atomic 카운터)

- [ ] **통계 메서드**
  - [ ] `IncrementL1Hit()` 메서드 구현
  - [ ] `IncrementL2Hit()` 메서드 구현
  - [ ] `IncrementMiss()` 메서드 구현
  - [ ] `GetStats()` 메서드 구현 (히트율 계산 포함)
  - [ ] `Reset()` 메서드 구현

### 🏗️ two_level_cache.go 기본 구현

- [ ] **구조체 정의**
  - [ ] `CacheTwoLevelStore` 구조체 정의
  - [ ] `asyncOperation` 구조체 정의

- [ ] **생성자 함수**
  - [ ] `NewCacheTwoLevelStore()` 기본 생성자 구현
  - [ ] `NewCacheTwoLevelStoreWithConfig()` 설정 생성자 구현
  - [ ] 기본값 설정 로직 구현

- [ ] **CacheStore 인터페이스 구현**
  - [ ] `Get(key uint64) ([]byte, bool)` 메서드 구현
    - [ ] L1 캐시 우선 조회
    - [ ] L2 캐시 fallback 조회
    - [ ] Cache Warming 로직
    - [ ] 통계 수집
  - [ ] `Set(key uint64, response []byte, expiration time.Time)` 메서드 구현
    - [ ] 전략별 분기 처리
  - [ ] `Release(key uint64)` 메서드 구현
    - [ ] L1, L2 모두에서 삭제

### 🔄 캐시 전략 구현

- [ ] **Write-Through 전략**
  - [ ] `setWriteThrough()` 메서드 구현
  - [ ] L1, L2 동시 동기 저장
  - [ ] TTL 계산 로직

- [ ] **Write-Back 전략 (기본)**
  - [ ] `setWriteBack()` 메서드 구현
  - [ ] L1 즉시 저장
  - [ ] L2 비동기 큐잉 (기본 구현)

- [ ] **Cache-Aside 전략**
  - [ ] `setCacheAside()` 메서드 구현

### 🧪 기본 테스트 작성

- [ ] **테스트 구조 설정**
  - [ ] `TwoLevelCacheTestSuite` 구조체 정의
  - [ ] `SetupTest()`, `TearDownTest()` 구현

- [ ] **기본 기능 테스트**
  - [ ] `TestWriteThrough()` - Write-Through 전략 테스트
  - [ ] `TestL1Hit()` - L1 캐시 히트 테스트
  - [ ] `TestL2HitWithCacheWarming()` - L2 히트 + Cache Warming 테스트
  - [ ] `TestCacheMiss()` - 캐시 미스 테스트
  - [ ] `TestRelease()` - 캐시 삭제 테스트

- [ ] **통계 테스트**
  - [ ] `TestStats()` - 통계 수집 정확성 테스트

---

## ⚡ Phase 2: 고급 기능 (1주)

### 🔄 비동기 처리 고도화

- [ ] **Async Worker 구현**
  - [ ] `startAsyncWorker()` 고루틴 구현
  - [ ] `Stop()` 메서드로 graceful shutdown
  - [ ] 채널 버퍼 관리 및 overflow 처리

- [ ] **Write-Back 전략 완성**
  - [ ] 비동기 L2 업데이트 구현
  - [ ] 에러 처리 및 재시도 로직
  - [ ] 채널 풀 상태 시 fallback 처리

### 🌡️ Cache Warming 최적화

- [ ] **Cache Warming 로직 개선**
  - [ ] L2 → L1 승격 조건 최적화
  - [ ] TTL 계산 로직 개선
  - [ ] 승격 실패 시 처리

### 🛠️ 추가 관리 기능

- [ ] **캐시 관리 메서드**
  - [ ] `ClearL1()` 메서드 구현
  - [ ] `ClearL2()` 메서드 구현  
  - [ ] `ClearAll()` 메서드 구현
  - [ ] `ResetStats()` 메서드 구현

- [ ] **동기화 메서드 (선택사항)**
  - [ ] `SyncL1ToL2()` 스텁 구현
  - [ ] `SyncL2ToL1()` 스텁 구현

### 🧪 고급 테스트

- [ ] **Write-Back 테스트**
  - [ ] `TestWriteBackStrategy()` - 비동기 처리 테스트
  - [ ] 타이밍 이슈 처리

- [ ] **에러 시나리오 테스트**
  - [ ] L1 실패 시 L2 fallback 테스트
  - [ ] L2 실패 시 L1 지속 사용 테스트
  - [ ] 네트워크 장애 시뮬레이션

### 📈 성능 테스트

- [ ] **벤치마크 테스트**
  - [ ] `BenchmarkTwoLevelGet()` 구현
  - [ ] `BenchmarkTwoLevelSet()` 구현
  - [ ] 메모리 사용량 프로파일링

---

## 🎛️ Phase 3: 운영 기능 (1주)

### 📊 모니터링 강화

- [ ] **실시간 통계**
  - [ ] Size 정보 수집 (L1Size, L2Size)
  - [ ] 히트율 실시간 계산
  - [ ] LastUpdate 타임스탬프

### 🖥️ 예제 애플리케이션

- [ ] **기본 예제**
  - [ ] `example/two_level_example.go` 완성
  - [ ] Echo 서버 설정
  - [ ] 캐시 미들웨어 설정
  - [ ] API 엔드포인트 구현

- [ ] **관리 엔드포인트**
  - [ ] `GET /cache/stats` - 통계 조회
  - [ ] `DELETE /cache/l1` - L1 캐시 삭제
  - [ ] `DELETE /cache/l2` - L2 캐시 삭제
  - [ ] `DELETE /cache/all` - 전체 캐시 삭제
  - [ ] `POST /cache/stats/reset` - 통계 리셋

### 🧪 통합 테스트

- [ ] **End-to-End 테스트**
  - [ ] Redis 연동 테스트 (실제 Redis 사용)
  - [ ] 전체 시나리오 통합 테스트
  - [ ] 부하 테스트

- [ ] **예제 테스트**
  - [ ] 예제 애플리케이션 동작 확인
  - [ ] API 엔드포인트 테스트
  - [ ] 통계 정확성 검증

---

## 📚 Phase 4: 문서화 및 마무리 (1주)

### 📖 문서 완성

- [ ] **README 업데이트**
  - [ ] Two-Level Caching 섹션 추가
  - [ ] 사용법 예제 추가
  - [ ] 설정 옵션 문서화

- [ ] **API 문서**
  - [ ] GoDoc 주석 완성
  - [ ] 메서드별 상세 설명
  - [ ] 사용 예제 추가

### 🧹 코드 품질

- [ ] **코드 리뷰**
  - [ ] Go 관례 준수 확인
  - [ ] 에러 처리 검토
  - [ ] 성능 최적화 검토

- [ ] **테스트 커버리지**
  - [ ] 테스트 커버리지 95% 이상 달성
  - [ ] Edge case 테스트 추가
  - [ ] 문서화된 모든 기능 테스트

### 🚀 배포 준비

- [ ] **호환성 확인**
  - [ ] 기존 API 100% 호환성 확인
  - [ ] Breaking change 없음 검증
  - [ ] 버전 호환성 테스트

- [ ] **최종 검증**
  - [ ] 모든 테스트 통과 확인
  - [ ] 예제 애플리케이션 동작 확인
  - [ ] 성능 목표 달성 확인

---

## 🎯 성공 기준

### 📊 성능 목표
- [ ] L1 캐시 히트: < 1ms ✅
- [ ] L2 캐시 히트: < 10ms ✅  
- [ ] 전체 캐시 히트율: 95% 이상 ✅

### 🔧 기능 목표
- [ ] 기존 API 100% 호환성 유지 ✅
- [ ] 3가지 캐시 전략 모두 구현 ✅
- [ ] Cache Warming 동작 ✅
- [ ] 실시간 통계 수집 ✅

### 🧪 품질 목표
- [ ] 테스트 커버리지 95% 이상 ✅
- [ ] 모든 테스트 통과 ✅
- [ ] 메모리 누수 없음 ✅
- [ ] 고루틴 누수 없음 ✅

---

## 📝 참고 사항

### 🔧 개발 환경
- Go 1.18+
- Redis 6.0+ (테스트용)
- testify 라이브러리

### 🛠️ 도구
```bash
# 테스트 실행
go test -v ./... -run TestTwoLevelCache

# 벤치마크 실행  
go test -v -bench=BenchmarkTwoLevelCache -benchmem

# 커버리지 확인
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 📚 참고 문서
- [PRD 문서](./prd.md)
- [Implementation 가이드](./implementation.md)
- [echo-http-cache 기존 구현](../../)

---

## 🏁 체크리스트 요약

**Phase 1 완료 조건:**
- [ ] 기본 구조체 및 인터페이스 구현
- [ ] Write-Through 전략 동작
- [ ] 기본 테스트 통과

**Phase 2 완료 조건:**  
- [ ] Write-Back 비동기 처리 구현
- [ ] Cache Warming 최적화
- [ ] 고급 테스트 및 벤치마크 통과

**Phase 3 완료 조건:**
- [ ] 완전한 예제 애플리케이션  
- [ ] 관리 API 구현
- [ ] 통합 테스트 통과

**Phase 4 완료 조건:**
- [ ] 완전한 문서화
- [ ] 코드 품질 검증  
- [ ] 배포 준비 완료

**전체 프로젝트 완료:** ✅ 모든 Phase 완료 + 성공 기준 달성 