package config

import (
	"strings"
	"testing"
	"time"

	"go-fiber-starter/internal/common"
)

func validConfig() *Config {
	return &Config{
		AppName:             "test",
		Env:                 EnvLocal,
		Port:                8080,
		LogLevel:            "info",
		DBDriver:            DBDriverSQLite,
		DBDSN:               "data/test.db",
		DBMaxOpenConns:      5,
		DBMaxIdleConns:      2,
		CORSAllowedOrigins:  []string{"*"},
		RateLimitPerMinute:  60,
		ReadTimeout:         15 * time.Second,
		WriteTimeout:        15 * time.Second,
		IdleTimeout:         60 * time.Second,
		ShutdownGracePeriod: 10 * time.Second,
	}
}

func TestValidate_OK(t *testing.T) {
	if err := validConfig().Validate(); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}
}

func TestValidate_CollectsAllProblems(t *testing.T) {
	cfg := validConfig()
	cfg.Port = 80       // 범위 밖
	cfg.Env = "staging" // 허용 외 값
	cfg.DBDriver = "oracle"

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error")
	}
	for _, want := range []string{"APP_PORT", "APP_ENV", "DB_DRIVER"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error should mention %s, got: %v", want, err)
		}
	}
}

func TestValidate_ProdSQLiteRejected(t *testing.T) {
	cfg := validConfig()
	cfg.Env = EnvProd
	cfg.DBDriver = DBDriverSQLite

	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "prod") {
		t.Fatalf("expected prod+sqlite rejection, got %v", err)
	}
}

func TestValidate_IdleConnsBoundedByMaxConns(t *testing.T) {
	cfg := validConfig()
	cfg.DBMaxIdleConns = 10
	cfg.DBMaxOpenConns = 5

	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "DB_MAX_IDLE_CONNS") {
		t.Fatalf("expected idle>open rejection, got %v", err)
	}
}

func TestValidate_AuthRules(t *testing.T) {
	cfg := validConfig()

	// 비활성화면 시크릿 없어도 통과
	if err := cfg.Validate(); err != nil {
		t.Fatalf("auth disabled should not require secret, got %v", err)
	}

	cfg.AuthEnabled = true
	cfg.AuthJWTSecret = "short"
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "AUTH_JWT_SECRET") {
		t.Fatalf("expected short-secret rejection, got %v", err)
	}

	cfg.AuthJWTSecret = strings.Repeat("x", 32)
	cfg.AuthTokenTTL = -1 * time.Second
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "AUTH_TOKEN_TTL") {
		t.Fatalf("expected non-positive TTL rejection, got %v", err)
	}
}

func TestNewPageQuery_ClampsLimit(t *testing.T) {
	q := common.NewPageQuery(0, 500)
	if q.Page != common.DefaultPage || q.Limit != common.MaxLimit {
		t.Fatalf("want page=%d limit=%d, got %+v", common.DefaultPage, common.MaxLimit, q)
	}
}

func TestNewPageMeta_TotalPages(t *testing.T) {
	meta := common.NewPageMeta(common.PageQuery{Page: 2, Limit: 20}, 45)
	if meta.TotalPages != 3 {
		t.Fatalf("want 3 total pages, got %d", meta.TotalPages)
	}
}
