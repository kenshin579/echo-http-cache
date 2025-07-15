# Redis Cluster Example

이 예제는 echo-http-cache 라이브러리의 Redis Cluster 지원을 보여줍니다.

## 사전 요구사항

- Docker와 Docker Compose 설치
- Go 1.18 이상

## 실행 방법

### 1. Redis Cluster 시작

프로젝트 루트 디렉토리에서:

```bash
# Redis Cluster 시작
docker-compose up -d

# 클러스터 상태 확인 (선택사항)
docker exec -it echo-http-cache_redis-node-1_1 redis-cli -p 7000 cluster info
```

### 2. 예제 애플리케이션 실행

```bash
cd example
go run redis_cluster_example.go
```

### 3. API 테스트

```bash
# 첫 번째 요청 (캐시 미스 - 새로운 응답 생성)
curl http://localhost:8080/api/data

# 두 번째 요청 (캐시 히트 - 캐시된 응답 반환)
curl http://localhost:8080/api/data

# 캐시 삭제
curl -X DELETE http://localhost:8080/api/cache/clear

# 다시 요청 (캐시 미스 - 새로운 응답 생성)
curl http://localhost:8080/api/data
```

### 4. 정리

```bash
# Redis Cluster 중지 및 제거
cd ..
docker-compose down -v
```

## 동작 확인

1. 첫 번째 GET 요청 시 "Generating new response..." 로그가 출력됩니다.
2. 두 번째 GET 요청 시 로그가 출력되지 않고 캐시된 응답이 반환됩니다.
3. DELETE 요청으로 캐시를 삭제한 후 다시 GET 요청하면 새로운 응답이 생성됩니다.

## Redis Cluster 구성

- 6개 노드 (마스터 3개, 슬레이브 3개)
- 포트: 7000-7005
- 자동 페일오버 지원
- 16384 해시 슬롯 사용

## 트러블슈팅

### 클러스터 생성 실패
```bash
# 기존 볼륨 삭제 후 재시작
docker-compose down -v
docker-compose up -d
```

### 연결 오류
- 모든 포트(7000-7005)가 사용 가능한지 확인
- 방화벽 설정 확인 