package router

import (
	"context"
	"io"
	"log/slog"
	"net/http"
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

func TestAuthDisabled_LoginRouteAbsent(t *testing.T) {
	cfg := newTestConfig(t) // AUTH_ENABLED 기본 false
	app := newTestApp(t, cfg)

	resp, _ := testutil.Do(t, app, http.MethodPost, "/api/v1/auth/login",
		`{"username":"admin","password":"admin123"}`, "")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("auth disabled: want 404 for login route, got %d", resp.StatusCode)
	}
}
