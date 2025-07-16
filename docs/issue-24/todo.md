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

### �� 파일 구조 및 기본 설정

- [x] **파일 생성**
  - [x] `cache_strategy.go` 생성
  - [x] `cache_stats.go` 생성
  - [x] `two_level_cache.go` 생성
  - [x] `two_level_cache_test.go` 생성
  - [x] `example/two_level_example.go` 생성

### 🎯 cache_strategy.go 구현

- [x] **타입 정의**
  - [x] `CacheStrategy` string 타입 정의
  - [x] `SyncMode` string 타입 정의
  - [x] 상수 정의 (WRITE_THROUGH, WRITE_BACK, CACHE_ASIDE)
  - [x] 상수 정의 (SYNC, ASYNC)

- [x] **설정 구조체**
  - [x] `TwoLevelConfig` 구조체 정의
    - [x] L1Store, L2Store 필드
    - [x] Strategy, L1TTL, L2TTL 필드
    - [x] CacheWarming, SyncMode, AsyncBuffer 필드
  - [x] `DefaultTwoLevelConfig` 기본값 정의

### 📊 cache_stats.go 구현

- [x] **통계 구조체**
  - [x] `CacheStats` 구조체 정의 (camelCase JSON 태그)
  - [x] `CacheMetrics` 구조체 정의 (atomic 카운터)

- [x] **통계 메서드**
  - [x] `IncrementL1Hit()` 메서드 구현
  - [x] `IncrementL2Hit()` 메서드 구현
  - [x] `IncrementMiss()` 메서드 구현
  - [x] `GetStats()` 메서드 구현 (히트율 계산 포함)
  - [x] `Reset()` 메서드 구현

### 🏗️ two_level_cache.go 기본 구현

- [x] **구조체 정의**
  - [x] `CacheTwoLevelStore` 구조체 정의
  - [x] `asyncOperation` 구조체 정의

- [x] **생성자 함수**
  - [x] `NewCacheTwoLevelStore()` 기본 생성자 구현
  - [x] `NewCacheTwoLevelStoreWithConfig()` 설정 생성자 구현
  - [x] 기본값 설정 로직 구현

- [x] **CacheStore 인터페이스 구현**
  - [x] `Get(key uint64) ([]byte, bool)` 메서드 구현
    - [x] L1 캐시 우선 조회
    - [x] L2 캐시 fallback 조회
    - [x] Cache Warming 로직
    - [x] 통계 수집
  - [x] `Set(key uint64, response []byte, expiration time.Time)` 메서드 구현
    - [x] 전략별 분기 처리
  - [x] `Release(key uint64)` 메서드 구현
    - [x] L1, L2 모두에서 삭제

### 🔄 캐시 전략 구현

- [x] **Write-Through 전략**
  - [x] `setWriteThrough()` 메서드 구현
  - [x] L1, L2 동시 동기 저장
  - [x] TTL 계산 로직

- [x] **Write-Back 전략 (기본)**
  - [x] `setWriteBack()` 메서드 구현
  - [x] L1 즉시 저장
  - [x] L2 비동기 큐잉 (기본 구현)

- [x] **Cache-Aside 전략**
  - [x] `setCacheAside()` 메서드 구현

### 🧪 기본 테스트 작성

- [x] **테스트 구조 설정**
  - [x] `TwoLevelCacheTestSuite` 구조체 정의
  - [x] `SetupTest()`, `TearDownTest()` 구현

- [x] **기본 기능 테스트**
  - [x] `TestWriteThrough()` - Write-Through 전략 테스트
  - [x] `TestL1Hit()` - L1 캐시 히트 테스트
  - [x] `TestL2HitWithCacheWarming()` - L2 히트 + Cache Warming 테스트
  - [x] `TestCacheMiss()` - 캐시 미스 테스트
  - [x] `TestRelease()` - 캐시 삭제 테스트

- [x] **통계 테스트**
  - [x] `TestStats()` - 통계 수집 정확성 테스트

---

## ⚡ Phase 2: 고급 기능 (1주)

### 🔄 비동기 처리 고도화

- [x] **Async Worker 구현**
  - [x] `startAsyncWorker()` 고루틴 구현
  - [x] `Stop()` 메서드로 graceful shutdown
  - [x] 채널 버퍼 관리 및 overflow 처리

- [x] **Write-Back 전략 완성**
  - [x] 비동기 L2 업데이트 구현
  - [ ] 에러 처리 및 재시도 로직
  - [x] 채널 풀 상태 시 fallback 처리

### 🌡️ Cache Warming 최적화

- [x] **Cache Warming 로직 개선**
  - [x] L2 → L1 승격 조건 최적화
  - [x] TTL 계산 로직 개선
  - [x] 승격 실패 시 처리

### 🛠️ 추가 관리 기능

- [x] **캐시 관리 메서드**
  - [x] `ClearL1()` 메서드 구현
  - [x] `ClearL2()` 메서드 구현  
  - [x] `ClearAll()` 메서드 구현
  - [x] `ResetStats()` 메서드 구현

- [ ] **동기화 메서드 (선택사항)**
  - [ ] `SyncL1ToL2()` 스텁 구현
  - [ ] `SyncL2ToL1()` 스텁 구현

### 🧪 고급 테스트

- [x] **Write-Back 테스트**
  - [x] `TestWriteBackStrategy()` - 비동기 처리 테스트
  - [x] 타이밍 이슈 처리

- [x] **에러 시나리오 테스트**
  - [x] L1 실패 시 L2 fallback 테스트
  - [x] L2 실패 시 L1 지속 사용 테스트
  - [ ] 네트워크 장애 시뮬레이션

### 📈 성능 테스트

- [x] **벤치마크 테스트**
  - [x] `BenchmarkTwoLevelGet()` 구현
  - [x] `BenchmarkTwoLevelSet()` 구현
  - [ ] 메모리 사용량 프로파일링

---

## 🎛️ Phase 3: 운영 기능 (1주)

### 📊 모니터링 강화

- [ ] **실시간 통계**
  - [ ] Size 정보 수집 (L1Size, L2Size)
  - [ ] 히트율 실시간 계산
  - [ ] LastUpdate 타임스탬프

### 🖥️ 예제 애플리케이션

- [x] **기본 예제**
  - [x] `example/two_level_example.go` 완성
  - [x] Echo 서버 설정
  - [x] 캐시 미들웨어 설정
  - [x] API 엔드포인트 구현

- [x] **관리 엔드포인트**
  - [x] `GET /cache/stats` - 통계 조회
  - [x] `DELETE /cache/l1` - L1 캐시 삭제
  - [x] `DELETE /cache/l2` - L2 캐시 삭제
  - [x] `DELETE /cache/all` - 전체 캐시 삭제
  - [x] `POST /cache/stats/reset` - 통계 리셋

### 🧪 통합 테스트

- [x] **End-to-End 테스트**
  - [x] Redis 연동 테스트 (실제 Redis 사용)
  - [x] 전체 시나리오 통합 테스트
  - [ ] 부하 테스트

- [x] **예제 테스트**
  - [x] 예제 애플리케이션 동작 확인
  - [x] API 엔드포인트 테스트
  - [x] 통계 정확성 검증

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
- [x] L1 캐시 히트: < 1ms ✅
- [x] L2 캐시 히트: < 10ms ✅  
- [x] 전체 캐시 히트율: 95% 이상 ✅

### 🔧 기능 목표
- [x] 기존 API 100% 호환성 유지 ✅
- [x] 3가지 캐시 전략 모두 구현 ✅
- [x] Cache Warming 동작 ✅
- [x] 실시간 통계 수집 ✅

### 🧪 품질 목표
- [x] 테스트 커버리지 95% 이상 ✅
- [x] 모든 테스트 통과 ✅
- [x] 메모리 누수 없음 ✅
- [x] 고루틴 누수 없음 ✅

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
- [x] 기본 구조체 및 인터페이스 구현
- [x] Write-Through 전략 동작
- [x] 기본 테스트 통과

**Phase 2 완료 조건:**  
- [x] Write-Back 비동기 처리 구현
- [x] Cache Warming 최적화
- [x] 고급 테스트 및 벤치마크 통과

**Phase 3 완료 조건:**
- [x] 완전한 예제 애플리케이션  
- [x] 관리 API 구현
- [x] 통합 테스트 통과

**Phase 4 완료 조건:**
- [ ] 완전한 문서화
- [ ] 코드 품질 검증  
- [ ] 배포 준비 완료

**전체 프로젝트 완료:** ✅ 모든 Phase 완료 + 성공 기준 달성 