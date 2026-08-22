package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Env string

const (
	EnvLocal Env = "local"
	EnvDev   Env = "dev"
	EnvProd  Env = "prod"
)

type DBDriver string

const (
	DBDriverPostgres DBDriver = "postgres"
	DBDriverMySQL    DBDriver = "mysql"
	DBDriverSQLite   DBDriver = "sqlite"
)

type Config struct {
	AppName  string `env:"APP_NAME" envDefault:"go-fiber-starter"`
	Env      Env    `env:"APP_ENV" envDefault:"local"`
	Port     int    `env:"APP_PORT" envDefault:"8080"`
	LogLevel string `env:"LOG_LEVEL" envDefault:"info"`

	DBDriver             DBDriver      `env:"DB_DRIVER" envDefault:"sqlite"`
	DBDSN                string        `env:"DB_DSN" envDefault:"data/app.db"`
	DBMaxOpenConns       int           `env:"DB_MAX_OPEN_CONNS" envDefault:"25"`
	DBMaxIdleConns       int           `env:"DB_MAX_IDLE_CONNS" envDefault:"5"`
	DBConnMaxLifetime    time.Duration `env:"DB_CONN_MAX_LIFETIME" envDefault:"30m"`
	DBAutoMigrate        bool          `env:"DB_AUTO_MIGRATE" envDefault:"true"`
	DBSlowQueryThreshold time.Duration `env:"DB_SLOW_QUERY_THRESHOLD" envDefault:"500ms"`

	CORSAllowedOrigins []string `env:"CORS_ALLOWED_ORIGINS" envSeparator:"," envDefault:"*"`
	RateLimitPerMinute int      `env:"RATE_LIMIT_PER_MINUTE" envDefault:"120"`

	ReadTimeout  time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"15s"`
	WriteTimeout time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"15s"`
	IdleTimeout  time.Duration `env:"HTTP_IDLE_TIMEOUT" envDefault:"60s"`

	ShutdownGracePeriod time.Duration `env:"SHUTDOWN_GRACE_PERIOD" envDefault:"10s"`

	AuthEnabled      bool          `env:"AUTH_ENABLED" envDefault:"false"`
	AuthJWTSecret    string        `env:"AUTH_JWT_SECRET"`
	AuthTokenTTL     time.Duration `env:"AUTH_TOKEN_TTL" envDefault:"1h"`
	AuthDemoUsername string        `env:"AUTH_DEMO_USERNAME" envDefault:"admin"`
	AuthDemoPassword string        `env:"AUTH_DEMO_PASSWORD" envDefault:"admin123"`
}

// Load는 .env(선택)와 환경변수를 읽어 Config를 구성하고, 규칙 위반 시 모든 오류를 모아 반환한다.
func Load() (*Config, error) {
	_ = godotenv.Load() // .env 없어도 무시 (환경변수 주입 환경 지원)

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parse env: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration:\n%w", err)
	}
	return cfg, nil
}

// Validate는 설정 불변식을 검사한다. 실패한 항목을 전부 수집해 한 번에 보고한다.
func (c *Config) Validate() error {
	var problems []string
	addf := func(format string, args ...any) {
		problems = append(problems, "  - "+fmt.Sprintf(format, args...))
	}

	if c.AppName == "" {
		addf("APP_NAME must not be empty")
	}
	if !isValidPort(c.Port) {
		addf("APP_PORT %d is out of range [1024, 65535]", c.Port)
	}
	switch c.Env {
	case EnvLocal, EnvDev, EnvProd:
	default:
		addf("APP_ENV %q must be one of local|dev|prod", c.Env)
	}
	switch strings.ToLower(c.LogLevel) {
	case "debug", "info", "warn", "error":
	default:
		addf("LOG_LEVEL %q must be one of debug|info|warn|error", c.LogLevel)
	}

	switch c.DBDriver {
	case DBDriverPostgres, DBDriverMySQL, DBDriverSQLite:
	default:
		addf("DB_DRIVER %q must be one of postgres|mysql|sqlite", c.DBDriver)
	}
	if c.DBDSN == "" {
		addf("DB_DSN must not be empty")
	}
	if c.DBDriver == DBDriverSQLite && c.Env == EnvProd {
		addf("DB_DRIVER=sqlite is not allowed with APP_ENV=prod (use postgres or mysql)")
	}
	if c.DBMaxOpenConns < 1 {
		addf("DB_MAX_OPEN_CONNS must be >= 1, got %d", c.DBMaxOpenConns)
	}
	if c.DBMaxIdleConns < 0 || c.DBMaxIdleConns > c.DBMaxOpenConns {
		addf("DB_MAX_IDLE_CONNS must be in [0, DB_MAX_OPEN_CONNS], got %d", c.DBMaxIdleConns)
	}
	if c.RateLimitPerMinute < 1 {
		addf("RATE_LIMIT_PER_MINUTE must be >= 1, got %d", c.RateLimitPerMinute)
	}
	if c.AuthEnabled {
		if len(c.AuthJWTSecret) < 32 {
			addf("AUTH_JWT_SECRET must be >= 32 bytes when AUTH_ENABLED=true (got %d bytes)", len(c.AuthJWTSecret))
		}
		if c.AuthTokenTTL <= 0 {
			addf("AUTH_TOKEN_TTL must be positive, got %s", c.AuthTokenTTL)
		}
	}
	for _, d := range []struct {
		name string
		val  time.Duration
	}{
		{"HTTP_READ_TIMEOUT", c.ReadTimeout},
		{"HTTP_WRITE_TIMEOUT", c.WriteTimeout},
		{"HTTP_IDLE_TIMEOUT", c.IdleTimeout},
		{"SHUTDOWN_GRACE_PERIOD", c.ShutdownGracePeriod},
	} {
		if d.val <= 0 {
			addf("%s must be positive, got %s", d.name, d.val)
		}
	}

	if len(problems) > 0 {
		return fmt.Errorf("%s", strings.Join(problems, "\n"))
	}
	return nil
}

// IsProd는 prod 환경 여부를 반환한다.
func (c *Config) IsProd() bool { return c.Env == EnvProd }

func isValidPort(p int) bool {
	return p >= 1024 && p <= 65535
}
