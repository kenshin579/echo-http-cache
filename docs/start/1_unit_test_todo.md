# Unit Test GitHub Actions TODO

## Phase 1: 워크플로우 생성

- [x] `.github/workflows/` 디렉토리 생성 (없는 경우)
- [x] `.github/workflows/unit-test.yml` 파일 생성
- [x] actions 워크플로우 호출 설정 추가
- [x] Go 버전 1.25로 업데이트
- [x] push 트리거 추가 (main 브랜치)

## Phase 2: 검증

- [ ] PR push하여 워크플로우 실행 확인
- [ ] 워크플로우 실행 성공 확인
- [ ] PR 코멘트 정상 표시 확인
- [ ] 커버리지 리포트 확인

## Phase 3: 통합 테스트 (선택)

- [x] `run_integration_tests: true` 설정
- [x] `integration_test_args` 추가
- [ ] Docker Compose 동작 확인
- [ ] Redis Cluster 연결 테스트
