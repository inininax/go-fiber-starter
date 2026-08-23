package task

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

// BenchmarkTasksList는 대표 읽기 경로의 핸들러 레벨 비용을 추적한다.
// sqlite 파일 DB + 실제 GORM이라 수치는 절대값이 아닌 회귀 감지용 상대 지표다.
// b.Loop()는 타이머 관리를 내장하므로 ResetTimer를 호출하지 않는다(Go 1.24+).
func BenchmarkTasksList(b *testing.B) {
	app := newTestApp(b)
	req := func() *http.Request {
		return httptest.NewRequest(http.MethodGet, "/api/v1/tasks?page=1&limit=20", nil)
	}

	b.ReportAllocs()
	for b.Loop() {
		resp, err := app.Test(req())
		if err != nil || resp.StatusCode != http.StatusOK {
			b.Fatalf("list failed: %v %d", err, resp.StatusCode)
		}
	}
}

// BenchmarkTasksCreate는 쓰기 경로(파싱+검증+insert)를 추적한다.
func BenchmarkTasksCreate(b *testing.B) {
	app := newTestApp(b)

	b.ReportAllocs()
	i := 0
	for b.Loop() {
		body := `{"title":"bench-` + strconv.Itoa(i) + `"}`
		r := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(r)
		if err != nil || resp.StatusCode != http.StatusCreated {
			b.Fatalf("create failed: %v %d", err, resp.StatusCode)
		}
		i++
	}
}
