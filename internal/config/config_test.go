package config

import (
	"strings"
	"testing"
	"time"

	"go-fiber-starter/internal/pagination"
)

func validConfig() *Config {
	return &Config{
		AppName:                "test",
		Env:                    EnvLocal,
		Port:                   8080,
		LogLevel:               "info",
		DBDriver:               DBDriverSQLite,
		DBDSN:                  "data/test.db",
		DBMaxOpenConns:         5,
		DBMaxIdleConns:         2,
		CORSAllowedOrigins:     []string{"*"},
		RateLimitPerMinute:     60,
		AuthRateLimitPerMinute: 10,
		ReadTimeout:            15 * time.Second,
		WriteTimeout:           15 * time.Second,
		IdleTimeout:            60 * time.Second,
		ShutdownGracePeriod:    10 * time.Second,
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

func TestValidate_ProdWildcardCORSRejected(t *testing.T) {
	cfg := validConfig()
	cfg.Env = EnvProd
	cfg.DBDriver = DBDriverPostgres

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "CORS_ALLOWED_ORIGINS") {
		t.Fatalf("expected prod wildcard CORS rejection, got %v", err)
	}
}

func TestValidate_ProdExplicitCORSOriginsAccepted(t *testing.T) {
	cfg := validConfig()
	cfg.Env = EnvProd
	cfg.DBDriver = DBDriverPostgres
	cfg.CORSAllowedOrigins = []string{"https://app.example.com"}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected explicit CORS origins accepted in prod, got %v", err)
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

func TestValidate_ProdDefaultDemoCredentialsRejected(t *testing.T) {
	cfg := validConfig()
	cfg.Env = EnvProd
	cfg.AuthEnabled = true
	cfg.AuthJWTSecret = strings.Repeat("x", 32)
	cfg.AuthTokenTTL = time.Hour
	// 데모 기본값이 하나만 남아 있어도 거부한다.
	cfg.AuthDemoUsername = DefaultDemoUsername
	cfg.AuthDemoPassword = "explicit-strong-password"

	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "AUTH_DEMO_USERNAME") {
		t.Fatalf("expected prod+default demo credential rejection, got %v", err)
	}
}

func TestValidate_ProdExplicitDemoCredentialsAccepted(t *testing.T) {
	cfg := validConfig()
	cfg.Env = EnvProd
	cfg.DBDriver = DBDriverPostgres
	cfg.CORSAllowedOrigins = []string{"https://app.example.com"}
	cfg.AuthEnabled = true
	cfg.AuthJWTSecret = strings.Repeat("x", 32)
	cfg.AuthTokenTTL = time.Hour
	cfg.AuthDemoUsername = "ops-user"
	cfg.AuthDemoPassword = "ops-password"

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected explicit credentials accepted in prod, got %v", err)
	}
}

func TestValidate_LocalDefaultDemoCredentialsAccepted(t *testing.T) {
	cfg := validConfig()
	cfg.AuthEnabled = true
	cfg.AuthJWTSecret = strings.Repeat("x", 32)
	cfg.AuthTokenTTL = time.Hour
	cfg.AuthDemoUsername = DefaultDemoUsername
	cfg.AuthDemoPassword = DefaultDemoPassword

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected default demo credentials accepted outside prod, got %v", err)
	}
}

func TestValidate_TrustProxyRules(t *testing.T) {
	// OK: true + 유효한 allowlist
	cfg := validConfig()
	cfg.TrustProxy = true
	cfg.TrustProxyProxies = []string{"10.0.0.5", "172.17.0.0/16"}
	cfg.TrustProxyHeader = "X-Forwarded-For"
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid trust proxy config, got %v", err)
	}

	// true인데 allowlist 없으면 거부
	cfg = validConfig()
	cfg.TrustProxy = true
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "TRUST_PROXY_PROXIES") {
		t.Fatalf("expected missing-proxies rejection, got %v", err)
	}

	// IP/CIDR 형식 오류 항목은 거부
	cfg = validConfig()
	cfg.TrustProxy = true
	cfg.TrustProxyProxies = []string{"10.0.0.5", "not-an-ip"}
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "not-an-ip") {
		t.Fatalf("expected invalid CIDR rejection, got %v", err)
	}

	// true인데 헤더가 빈 문자열이면 거부
	cfg = validConfig()
	cfg.TrustProxy = true
	cfg.TrustProxyProxies = []string{"10.0.0.5"}
	cfg.TrustProxyHeader = ""
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "TRUST_PROXY_HEADER") {
		t.Fatalf("expected empty header rejection, got %v", err)
	}

	// false인데 allowlist를 두면 무효 설정으로 거부
	cfg = validConfig()
	cfg.TrustProxyProxies = []string{"10.0.0.5"}
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "TRUST_PROXY=false") {
		t.Fatalf("expected proxies-without-trust rejection, got %v", err)
	}
}

func TestNewPageQuery_ClampsLimit(t *testing.T) {
	q := pagination.NewPageQuery(0, 500)
	if q.Page != pagination.DefaultPage || q.Limit != pagination.MaxLimit {
		t.Fatalf("want page=%d limit=%d, got %+v", pagination.DefaultPage, pagination.MaxLimit, q)
	}
}

func TestNewPageMeta_TotalPages(t *testing.T) {
	meta := pagination.NewPageMeta(pagination.PageQuery{Page: 2, Limit: 20}, 45)
	if meta.TotalPages != 3 {
		t.Fatalf("want 3 total pages, got %d", meta.TotalPages)
	}
}
