# Unit Test GitHub Actions 구현 문서

## 1. 워크플로우 파일

### 파일 위치
`.github/workflows/unit-test.yml`

### 워크플로우 코드

```yaml
name: Tests

on:
  pull_request:
    branches: [main, master]
    types: [opened, synchronize, reopened]
  push:
    branches: [main]

jobs:
  test:
    uses: kenshin579/actions/.github/workflows/unit-test-go.yml@main
    with:
      go_version: '1.25'
      run_unit_tests: true
      run_integration_tests: true
      unit_test_args: '-v -short'
      integration_test_args: '-v -run TestCacheRedisClusterStore'
      coverage_enabled: true
      docker_compose_file: 'docker-compose.yml'
      post_comment: true
```

---

## 2. 파라미터 설명

| 파라미터 | 값 | 설명 |
|---------|-----|------|
| `go_version` | '1.25' | 프로젝트 Go 버전 |
| `run_unit_tests` | true | 단위 테스트 활성화 |
| `run_integration_tests` | true | 통합 테스트 활성화 (Docker 사용) |
| `unit_test_args` | '-v -short' | 상세 출력 + short 모드 |
| `integration_test_args` | '-v -run TestCacheRedisClusterStore' | Redis Cluster 테스트 |
| `coverage_enabled` | true | 커버리지 수집 |
| `docker_compose_file` | 'docker-compose.yml' | Docker Compose 파일 경로 |
| `post_comment` | true | PR 코멘트 게시 |

---

## 3. 테스트 트리거 조건

| 이벤트 | 대상 브랜치 | 실행 내용 |
|--------|------------|----------|
| PR 생성/업데이트 | main, master | 단위 + 통합 테스트 |
| Push | main | 단위 + 통합 테스트 |
