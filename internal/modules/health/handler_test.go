package health

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func newApp(ping func() error, commit string) *fiber.App {
	app := fiber.New()
	h := NewHandler(ping, commit)
	app.Get("/livez", h.Livez)
	app.Get("/readyz", h.Readyz)
	return app
}

func TestLivez_ReportsStatusAndCommit(t *testing.T) {
	resp, err := newApp(func() error { return nil }, "abc1234").Test(
		request(http.MethodGet, "/livez"))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200, got %d", resp.StatusCode)
	}
	body := string(readAll(resp))
	if !strings.Contains(body, `"commit":"abc1234"`) || !strings.Contains(body, `"status":"ok"`) {
		t.Fatalf("livez payload broken: %s", body)
	}
}

// readyz 실패는 503 + 프로브 스키마를 반환한다(엔벨로프 미준수는 의도된 예외).
func TestReadyz_DatabaseDown_503(t *testing.T) {
	resp, err := newApp(func() error { return errors.New("connection refused") }, "dev").Test(
		request(http.MethodGet, "/readyz"))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("want 503, got %d", resp.StatusCode)
	}
	body := string(readAll(resp))
	if !strings.Contains(body, `"database":"down"`) || !strings.Contains(body, `unavailable`) {
		t.Fatalf("readyz down payload broken: %s", body)
	}
}

func TestReadyz_Up(t *testing.T) {
	resp, err := newApp(func() error { return nil }, "dev").Test(request(http.MethodGet, "/readyz"))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK || !strings.Contains(string(readAll(resp)), `"database":"up"`) {
		t.Fatalf("readyz up broken: %d %s", resp.StatusCode, readAll(resp))
	}
}

// commit 미주입 조립 경로의 폴백 계약.
func TestNewHandler_EmptyCommit_FallsBackToDev(t *testing.T) {
	resp, err := newApp(func() error { return nil }, "").Test(request(http.MethodGet, "/livez"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(readAll(resp)), `"commit":"dev"`) {
		t.Fatalf("empty commit must fall back to dev: %s", readAll(resp))
	}
}
