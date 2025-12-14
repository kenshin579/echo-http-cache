# Unit Test GitHub Actions PRD

## 1. 배경 및 목표

### 배경
echo-http-cache 프로젝트는 현재 로컬에서만 테스트를 실행하고 있으며, GitHub Actions를 통한 자동화된 CI/CD 파이프라인이 구성되어 있지 않다. 코드 품질과 안정성을 보장하기 위해 PR 및 push 이벤트에서 자동으로 테스트가 실행되어야 한다.

### 목표
- PR 생성 및 main 브랜치 push 시 자동으로 단위 테스트 실행
- 테스트 결과 및 커버리지를 PR 코멘트로 리포팅
- my-actions 저장소의 재사용 가능한 워크플로우(`unit-test.yml`) 활용

---

## 2. 현재 상태 분석

### 2.1 echo-http-cache 프로젝트

| 항목 | 상태 |
|------|------|
| Go 버전 | 1.25 |
| 테스트 프레임워크 | testify (assert, suite) |
| Mock Redis | miniredis/v2 |
| 통합 테스트 | Docker Compose (Redis Cluster) |
| GitHub Actions | **미구성** |

### 2.2 테스트 파일 구조

```
echo-http-cache/
├── cache_test.go                    # 캐시 미들웨어 테스트
├── memory_test.go                   # 메모리 스토어 테스트
├── redis_test.go                    # Redis 스토어 테스트
├── redis_cluster_test.go            # Redis Cluster 테스트
├── redis_cluster_integration_test.go # 통합 테스트 (build tag: integration)
├── two_level_cache_test.go          # 2단계 캐시 테스트
├── two_level_cache_bench_test.go    # 벤치마크 테스트
├── test/
│   └── redisdb.go                   # 테스트 헬퍼 (miniredis)
└── example/
    ├── example_test.go              # 예제 테스트
    └── two_level/
        └── e2e_test.go              # E2E 테스트
```

### 2.3 테스트 실행 명령어

```bash
# 단위 테스트 (miniredis 사용, Docker 불필요)
go test -v -short ./...

# 통합 테스트 (Docker Compose 필요)
go test -v -tags=integration -run Integration ./...

# 벤치마크
go test -bench=. -benchmem ./...
```

### 2.4 my-actions 워크플로우 (`unit-test.yml`)

**위치:** `kenshin579/my-actions/.github/workflows/unit-test.yml`

**주요 입력 파라미터:**

| 파라미터 | 타입 | 기본값 | 설명 |
|---------|------|--------|------|
| `go_version` | string | '1.25' | Go 버전 |
| `test_path` | string | './...' | 테스트 경로 |
| `run_unit_tests` | boolean | true | 단위 테스트 실행 여부 |
| `run_integration_tests` | boolean | false | 통합 테스트 실행 여부 |
| `unit_test_args` | string | '-v -short' | 단위 테스트 인자 |
| `integration_test_args` | string | '-v -run Integration' | 통합 테스트 인자 |
| `coverage_enabled` | boolean | true | 커버리지 수집 여부 |
| `docker_compose_file` | string | 'docker-compose.yml' | Docker Compose 파일 경로 |
| `post_comment` | boolean | true | PR 코멘트 게시 여부 |

**제공 기능:**
- 단위 테스트 및 통합 테스트 분리 실행
- 커버리지 리포트 생성
- PR 코멘트에 테스트 결과 및 커버리지 표시
- Docker Compose를 통한 서비스 의존성 관리

---

## 3. 요구사항

### 3.1 기능 요구사항

#### FR-1: 단위 테스트 자동 실행
- PR 생성/업데이트 시 단위 테스트 자동 실행
- main 브랜치 push 시 단위 테스트 자동 실행
- miniredis 기반 테스트로 Docker 의존성 없이 실행

#### FR-2: 통합 테스트 자동 실행 (선택적)
- Docker Compose로 Redis Cluster 시작
- `integration` build tag가 있는 테스트 실행
- 통합 테스트는 선택적으로 활성화 가능

#### FR-3: 테스트 결과 리포팅
- PR 코멘트에 테스트 결과 요약 표시
- 커버리지 퍼센트 표시
- 테스트 성공/실패 상태 표시 (이모지)

#### FR-4: 커버리지 수집
- `coverage.out` 파일 생성
- 커버리지 리포트를 아티팩트로 저장

### 3.2 비기능 요구사항

#### NFR-1: 실행 시간
- 단위 테스트: 5분 이내 완료
- 통합 테스트: 10분 이내 완료

#### NFR-2: 재사용성
- my-actions의 공통 워크플로우 활용
- 프로젝트별 커스터마이징 최소화

---

## 4. 기술적 세부사항

### 4.1 워크플로우 구성

**파일 위치:** `.github/workflows/test.yml`

```yaml
name: Go Tests

on:
  pull_request:
    branches: [main]
  push:
    branches: [main]

jobs:
  test:
    uses: kenshin579/my-actions/.github/workflows/unit-test.yml@main
    with:
      go_version: '1.25'
      run_unit_tests: true
      run_integration_tests: false  # 필요시 true로 변경
      unit_test_args: '-v -short'
      coverage_enabled: true
      post_comment: true
```

### 4.2 통합 테스트 활성화 시 설정

```yaml
jobs:
  test:
    uses: kenshin579/my-actions/.github/workflows/unit-test.yml@main
    with:
      go_version: '1.25'
      run_unit_tests: true
      run_integration_tests: true
      unit_test_args: '-v -short'
      integration_test_args: '-v -tags=integration -run Integration'
      docker_compose_file: 'docker-compose.yml'
      coverage_enabled: true
      post_comment: true
```

### 4.3 테스트 분류

| 테스트 유형 | 실행 조건 | 의존성 | 실행 인자 |
|------------|----------|--------|----------|
| 단위 테스트 | 항상 | miniredis (내장) | `-v -short` |
| 통합 테스트 | 선택적 | Docker (Redis Cluster) | `-v -tags=integration` |
| 벤치마크 | 수동 | miniredis | `-bench=. -benchmem` |

---

## 5. 참고 자료

- [my-actions 저장소](https://github.com/kenshin579/my-actions)
- [unit-test.yml 워크플로우](/Users/user/GolandProjects/my-actions/.github/workflows/unit-test.yml)
- [echo-http-cache CLAUDE.md](/Users/user/GolandProjects/echo-http-cache/CLAUDE.md)
