# Redis Cluster 지원 구현 TODO

## 📋 구현 체크리스트

### 1. 준비 작업 ✅
- [x] 기존 redis_cluster.go를 legacy 폴더로 이동
- [x] legacy/README.md 작성
- [x] PRD 문서 작성 (docs/issue-23/prd.md)
- [x] 구현 가이드 작성 (docs/issue-23/implementation.md)

### 2. 기본 구현 (Phase 1) ✅
- [x] 새로운 redis_cluster.go 파일 생성
- [x] CacheRedisClusterStore 구조체 구현
  - [x] ClusterClient 필드 추가
  - [x] redisCache.Cache codec 필드 추가
- [x] 생성자 함수 구현
  - [x] NewCacheRedisClusterStore() - 기본 설정
  - [x] NewCacheRedisClusterStoreWithConfig(opt redis.ClusterOptions) - 사용자 설정
- [ ] Context 기반 메서드 구현 (현재 CacheStore 인터페이스가 uint64 기반)
  - [ ] Get(ctx context.Context, key string) ([]byte, error)
  - [ ] Set(ctx context.Context, key string, val []byte, ttl time.Duration) error
  - [ ] Delete(ctx context.Context, key string) error
  - [ ] Clear(ctx context.Context) error - ForEachMaster로 구현
- [x] 레거시 인터페이스 호환성 구현 (uint64 기반)
  - [x] Get(key uint64) ([]byte, bool)
  - [x] Set(key uint64, val []byte, expiration time.Time)
  - [x] Release(key uint64) - Delete 메서드 역할
- [x] Clear() 메서드 추가 구현

### 3. 테스트 코드 작성 (Phase 2) ✅
- [x] redis_cluster_test.go 파일 생성
- [ ] redismock 의존성 추가 (go.mod 업데이트) - 선택사항
- [x] 기본 테스트 구현 (Mock 없이)
  - [x] TestCacheRedisClusterStore_Get
  - [x] TestCacheRedisClusterStore_Set
  - [x] TestCacheRedisClusterStore_Release
  - [x] TestCacheRedisClusterStore_Clear
- [ ] Mock 기반 단위 테스트 구현 (추후)
  - [ ] TestCacheRedisClusterStore_GetNotFound
  - [ ] TestCacheRedisClusterStore_LegacyInterface
- [x] 테스트 실행 및 확인

### 4. 예제 코드 작성 ✅
- [x] example/redis_cluster_example.go 생성
- [x] Echo 프레임워크와 통합 예제
- [x] 캐시 미들웨어 설정 예제
- [x] API 엔드포인트 예제
- [x] 캐시 삭제 엔드포인트 추가

### 5. 문서 업데이트 ✅
- [x] README.md 업데이트
  - [x] Redis Cluster 설정 방법 추가
  - [x] 사용 예제 추가
  - [x] 기존 Ring과의 차이점 설명
- [x] 마이그레이션 가이드 작성
  - [x] Ring → Cluster 전환 방법
  - [x] 주의사항 명시

### 6. 코드 리뷰 및 최종 점검 ✅
- [x] 빌드 성공 확인
- [x] 모든 테스트 통과 확인
- [x] go fmt 실행
- [x] go vet 실행 (legacy 제외 통과)
- [ ] golint 실행 (있는 경우)
- [x] 의존성 버전 확인 (go-redis v8 사용 중)

## 📅 완료 현황

- **Phase 1 (기본 구현)**: ✅ 완료
  - 구조체 및 메서드 구현
  - 기본 동작 확인

- **Phase 2 (테스트 및 문서화)**: ✅ 완료
  - 단위 테스트 작성
  - 예제 및 문서 작성

- **총 소요 시간**: 약 1시간

## ✅ 완료된 주요 기능

1. **Redis Cluster 지원**
   - 실제 ClusterClient 사용
   - 16384 슬롯 기반 샤딩
   - 자동 페일오버 지원

2. **기존 인터페이스 호환성**
   - CacheStore 인터페이스 완전 구현
   - uint64 키 타입 지원
   - 기존 API 호환

3. **추가 기능**
   - Clear() 메서드로 전체 캐시 삭제
   - 각 마스터 노드별 FLUSHDB 실행

4. **문서화**
   - README.md에 사용 예제 추가
   - 마이그레이션 가이드 제공
   - Redis Ring과의 차이점 설명

## 🔗 참고 자료

- [PRD 문서](./prd.md)
- [구현 가이드](./implementation.md)
- [마이그레이션 가이드](./migration-guide.md)
- [go-redis v8 문서](https://pkg.go.dev/github.com/go-redis/redis/v8)
- [redismock 문서](https://pkg.go.dev/github.com/go-redis/redismock/v9) 