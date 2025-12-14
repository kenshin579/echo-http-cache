# Unit Test GitHub Actions 구현 문서

## 1. 워크플로우 파일 생성

### 파일 위치
`.github/workflows/test.yml`

### 워크플로우 코드

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
      run_integration_tests: false
      unit_test_args: '-v -short'
      coverage_enabled: true
      post_comment: true
```

---

## 2. 파라미터 설명

| 파라미터 | 값 | 설명 |
|---------|-----|------|
| `go_version` | '1.25' | 프로젝트 Go 버전 |
| `run_unit_tests` | true | 단위 테스트 활성화 |
| `run_integration_tests` | false | 통합 테스트 비활성화 (Docker 불필요) |
| `unit_test_args` | '-v -short' | 상세 출력 + short 모드 |
| `coverage_enabled` | true | 커버리지 수집 |
| `post_comment` | true | PR 코멘트 게시 |

---

## 3. 통합 테스트 활성화 (선택)

통합 테스트가 필요한 경우 아래와 같이 설정 변경:

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

---

## 4. 디렉토리 구조

```
.github/
└── workflows/
    └── test.yml    # 신규 생성
```

---

## 5. 테스트 트리거 조건

| 이벤트 | 대상 브랜치 | 실행 내용 |
|--------|------------|----------|
| PR 생성/업데이트 | main | 단위 테스트 + 커버리지 |
| Push | main | 단위 테스트 + 커버리지 |
