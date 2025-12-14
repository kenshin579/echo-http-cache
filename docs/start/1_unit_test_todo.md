# Unit Test GitHub Actions TODO

## Phase 1: 워크플로우 생성

- [ ] `.github/workflows/` 디렉토리 생성 (없는 경우)
- [ ] `.github/workflows/test.yml` 파일 생성
- [ ] my-actions 워크플로우 호출 설정 추가

## Phase 2: 검증

- [ ] feature 브랜치 생성
- [ ] 테스트 PR 생성
- [ ] 워크플로우 실행 확인
- [ ] PR 코멘트 정상 표시 확인
- [ ] 커버리지 리포트 확인

## Phase 3: 통합 테스트 (선택)

- [ ] `run_integration_tests: true` 설정
- [ ] `integration_test_args` 추가
- [ ] Docker Compose 동작 확인
- [ ] Redis Cluster 연결 테스트
