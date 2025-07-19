// e2e_test.go
// 간단한 E2E 테스트: 서버를 띄우고 주요 엔드포인트를 호출하여 정상 동작을 확인합니다.
package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	echo_http_cache "github.com/kenshin579/echo-http-cache"
	testpkg "github.com/kenshin579/echo-http-cache/test"
)

func TestE2E(t *testing.T) {
	// miniredis 인스턴스 직접 생성
	mredis, _ := testpkg.NewRedisDB()
	defer mredis.Close()

	// main.go 서버는 수동으로 띄운다고 가정하고, 여기서는 띄우지 않음

	// 서버가 뜰 때까지 health 체크로 대기 (최대 5초)
	ready := false
	for i := 0; i < 50; i++ {
		resp, err := http.Get("http://localhost:8080/health")
		if err == nil && resp.StatusCode == 200 {
			ready = true
			resp.Body.Close()
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !ready {
		t.Fatalf("서버가 5초 내에 준비되지 않았습니다.")
	}

	// 일반 엔드포인트 테스트
	endpoints := []string{
		"/api/data",
		"/api/user/123",
		"/health",
	}

	for _, ep := range endpoints {
		resp, err := http.Get("http://localhost:8080" + ep)
		if err != nil {
			t.Errorf("%s 호출 실패: %v", ep, err)
			continue
		}
		body, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != 200 {
			t.Errorf("%s 응답 코드 오류: %d", ep, resp.StatusCode)
		} else if len(body) == 0 {
			t.Errorf("%s 응답 바디 없음", ep)
		} else {
			t.Logf("%s 성공! 응답: %s", ep, string(body))
		}
	}

	// 캐시 통계 검증 (첫 번째 호출 - 캐시 미스 확인)
	t.Log("=== 첫 번째 캐시 통계 확인 ===")
	initialStats := getCacheStats(t)
	if initialStats == nil {
		return
	}

	// 캐시된 엔드포인트 다시 호출 (캐시 히트 유도)
	t.Log("=== 캐시된 엔드포인트 재호출 ===")
	for _, ep := range []string{"/api/data", "/api/user/123"} {
		resp, err := http.Get("http://localhost:8080" + ep)
		if err != nil {
			t.Errorf("%s 재호출 실패: %v", ep, err)
			continue
		}
		resp.Body.Close()
		t.Logf("%s 재호출 완료", ep)
	}

	// 캐시 통계 재확인 (캐시 히트 확인)
	t.Log("=== 두 번째 캐시 통계 확인 ===")
	finalStats := getCacheStats(t)
	if finalStats == nil {
		return
	}

	// 캐시 동작 검증
	validateCacheStats(t, initialStats, finalStats)
}

func getCacheStats(t *testing.T) *echo_http_cache.CacheStats {
	resp, err := http.Get("http://localhost:8080/cache/stats")
	if err != nil {
		t.Errorf("/cache/stats 호출 실패: %v", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("/cache/stats 응답 코드 오류: %d", resp.StatusCode)
		return nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("/cache/stats 응답 읽기 실패: %v", err)
		return nil
	}

	var stats echo_http_cache.CacheStats
	if err := json.Unmarshal(body, &stats); err != nil {
		t.Errorf("/cache/stats JSON 파싱 실패: %v, 응답: %s", err, string(body))
		return nil
	}

	t.Logf("캐시 통계: L1Hits=%d, L2Hits=%d, TotalMiss=%d, TotalRequest=%d, HitRate=%.2f%%, L1Size=%d, L2Size=%d",
		stats.L1Hits, stats.L2Hits, stats.TotalMiss, stats.TotalRequest, stats.HitRate, stats.L1Size, stats.L2Size)

	return &stats
}

func validateCacheStats(t *testing.T, initial, final *echo_http_cache.CacheStats) {
	// 총 요청 수가 증가했는지 확인
	if final.TotalRequest <= initial.TotalRequest {
		t.Errorf("총 요청 수가 증가하지 않음: initial=%d, final=%d", initial.TotalRequest, final.TotalRequest)
	}

	// 캐시 히트가 발생했는지 확인 (L1 또는 L2)
	totalHitsInitial := initial.L1Hits + initial.L2Hits
	totalHitsFinal := final.L1Hits + final.L2Hits

	if totalHitsFinal <= totalHitsInitial {
		t.Errorf("캐시 히트가 증가하지 않음: initial=%d, final=%d", totalHitsInitial, totalHitsFinal)
	} else {
		t.Logf("캐시 히트 증가 확인: %d -> %d", totalHitsInitial, totalHitsFinal)
	}

	// 히트율이 합리적인 범위에 있는지 확인 (0-100% 사이)
	if final.HitRate < 0 || final.HitRate > 100 {
		t.Errorf("비정상적인 히트율: %.2f%%", final.HitRate)
	} else {
		t.Logf("히트율 정상: %.2f%%", final.HitRate)
	}

	// 캐시 크기가 음수가 아닌지 확인
	if final.L1Size < 0 || final.L2Size < 0 {
		t.Errorf("비정상적인 캐시 크기: L1=%d, L2=%d", final.L1Size, final.L2Size)
	}

	// 히트율 계산이 맞는지 검증
	expectedHitRate := float64(totalHitsFinal) / float64(final.TotalRequest) * 100
	if final.TotalRequest > 0 && final.HitRate != expectedHitRate {
		t.Errorf("히트율 계산 오류: 예상=%.2f%%, 실제=%.2f%%", expectedHitRate, final.HitRate)
	}

	t.Logf("캐시 동작 검증 완료!")
}
