package router

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/prometheus/client_golang/prometheus"

	"go-fiber-starter/internal/config"
	"go-fiber-starter/internal/database"
	"go-fiber-starter/internal/modules/task"
	"go-fiber-starter/internal/testutil"
)

// newTestConfig는 Load() 없이 필요한 최소 필드를 채운 Config를 만든다.
func newTestConfig(t *testing.T) *config.Config {
	t.Helper()
	return &config.Config{
		AppName:                "router-test",
		Env:                    config.EnvLocal,
		Port:                   8080,
		LogLevel:               "info",
		DBDriver:               config.DBDriverSQLite,
		DBDSN:                  filepath.Join(t.TempDir(), "test.db"),
		DBMaxOpenConns:         5,
		DBMaxIdleConns:         2,
		CORSAllowedOrigins:     []string{"*"},
		RateLimitPerMinute:     120,
		AuthRateLimitPerMinute: 10,
		ReadTimeout:            time.Second,
		WriteTimeout:           time.Second,
		IdleTimeout:            time.Second,
		ShutdownGracePeriod:    time.Second,
	}
}

// newTestApp은 임시 디렉터리 sqlite + 실제 GORM으로 전체 라우터를 조립한다.
func newTestApp(t *testing.T, cfg *config.Config) *fiber.App {
	t.Helper()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	db, err := database.New(context.Background(), cfg, database.NewLogger(log, false, time.Minute))
	if err != nil {
		t.Fatalf("database.New: %v", err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
	if err := db.AutoMigrate(&task.Task{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	return New(cfg, db, log, prometheus.NewRegistry())
}

func TestAuthLimiter_BlocksExcessLoginAttempts(t *testing.T) {
	cfg := newTestConfig(t)
	cfg.AuthEnabled = true
	cfg.AuthJWTSecret = strings.Repeat("s", 32)
	cfg.AuthDemoUsername = "admin"
	cfg.AuthDemoPassword = "admin123"
	cfg.RateLimitPerMinute = 120
	cfg.AuthRateLimitPerMinute = 2
	app := newTestApp(t, cfg)

	body := `{"username":"admin","password":"nope"}`
	for range cfg.AuthRateLimitPerMinute {
		resp, env := testutil.Do(t, app, http.MethodPost, "/api/v1/auth/login", body, "")
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("attempt within budget: want 401, got %d (%+v)", resp.StatusCode, env)
		}
	}

	resp, env := testutil.Do(t, app, http.MethodPost, "/api/v1/auth/login", body, "")
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("excess attempt: want 429, got %d (%+v)", resp.StatusCode, env)
	}
	if env.Success || env.Error == nil || env.Error.Code != "RATE_LIMITED" {
		t.Fatalf("expected RATE_LIMITED error envelope, got %+v", env)
	}
}

func TestAuthLimiter_DoesNotThrottleOtherRoutes(t *testing.T) {
	cfg := newTestConfig(t)
	cfg.AuthEnabled = true
	cfg.AuthJWTSecret = strings.Repeat("s", 32)
	cfg.RateLimitPerMinute = 120
	cfg.AuthRateLimitPerMinute = 2
	app := newTestApp(t, cfg)

	body := `{"username":"admin","password":"nope"}`
	for range cfg.AuthRateLimitPerMinute {
		testutil.Do(t, app, http.MethodPost, "/api/v1/auth/login", body, "")
	}

	resp, _ := testutil.Do(t, app, http.MethodGet, "/livez", "", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login budget exhausted but livez should pass, got %d", resp.StatusCode)
	}
}

func TestGlobalLimiter_SkipsProbeEndpoints(t *testing.T) {
	cfg := newTestConfig(t)
	cfg.RateLimitPerMinute = 1 // 예산 1로 소진을 쉽게 만든다
	app := newTestApp(t, cfg)

	// 첫 요청은 예산 내 성공
	resp, _ := testutil.Do(t, app, http.MethodGet, "/api/v1/tasks", "", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("first request within budget: want 200, got %d", resp.StatusCode)
	}

	// 두 번째 API 요청은 전역 예산 소진으로 429
	resp, env := testutil.Do(t, app, http.MethodPost, "/api/v1/tasks", `{"title":"x"}`, "")
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("budget exhausted: want 429, got %d (%+v)", resp.StatusCode, env)
	}
	if env.Success || env.Error == nil || env.Error.Code != "RATE_LIMITED" {
		t.Fatalf("expected RATE_LIMITED error envelope, got %+v", env)
	}

	// 프로브 경로는 스킵되어 429가 아니다
	for _, path := range []string{"/livez", "/readyz"} {
		resp, env := testutil.Do(t, app, http.MethodGet, path, "", "")
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("%s should skip global limiter: want 200, got %d (%+v)", path, resp.StatusCode, env)
		}
	}

	// /metrics는 비-엔벨로프 응답이라 testutil 확장 없이 raw로 직접 단언한다.
	mreq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	mresp, err := app.Test(mreq)
	if err != nil {
		t.Fatalf("/metrics: %v", err)
	}
	if mresp.StatusCode != http.StatusOK {
		t.Fatalf("/metrics should skip global limiter: want 200, got %d", mresp.StatusCode)
	}
	if ct := mresp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/plain") {
		t.Fatalf("/metrics content-type: want text/plain, got %q", ct)
	}
}

func TestLivez_ReportsBuildCommit(t *testing.T) {
	cfg := newTestConfig(t)
	cfg.BuildCommit = "abc1234"
	app := newTestApp(t, cfg)

	// livez는 엔벨로프가 아닌 프로브 전용 JSON이라 testutil.Do 대신 raw로 단언한다.
	req := httptest.NewRequest(http.MethodGet, "/livez", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("livez: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("livez: want 200, got %d", resp.StatusCode)
	}
	var body struct {
		Status string `json:"status"`
		Commit string `json:"commit"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode livez: %v", err)
	}
	if body.Status != "ok" || body.Commit != cfg.BuildCommit {
		t.Fatalf("livez: want status=ok commit=%q, got %+v", cfg.BuildCommit, body)
	}
}

func TestAuthDisabled_LoginRouteAbsent(t *testing.T) {
	cfg := newTestConfig(t) // AUTH_ENABLED 기본 false
	app := newTestApp(t, cfg)

	resp, _ := testutil.Do(t, app, http.MethodPost, "/api/v1/auth/login",
		`{"username":"admin","password":"admin123"}`, "")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("auth disabled: want 404 for login route, got %d", resp.StatusCode)
	}
}
