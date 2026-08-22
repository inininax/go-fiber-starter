package router

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	fibermw "github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/gorm"

	"go-fiber-starter/internal/common"
	"go-fiber-starter/internal/config"
	appmw "go-fiber-starter/internal/middleware"
	"go-fiber-starter/internal/modules/auth"
	"go-fiber-starter/internal/modules/health"
	"go-fiber-starter/internal/modules/task"
)

const bodyLimit = 1 << 20 // 1 MiB

// New는 fiber 앱을 생성하고 미들웨어 체인/라우트/에러 처리를 조립한다.
func New(cfg *config.Config, db *gorm.DB, log *slog.Logger, reg *prometheus.Registry) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:         cfg.AppName,
		BodyLimit:       bodyLimit,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		IdleTimeout:     cfg.IdleTimeout,
		StructValidator: newStructValidator(),
		ErrorHandler:    errorHandler(log),
		// 역프록시 뒤에서 실제 클라이언트 IP를 신뢰할 수 있게 한다(allowlist 필수는 config.Validate가 보장).
		TrustProxy:       cfg.TrustProxy,
		TrustProxyConfig: fiber.TrustProxyConfig{Proxies: cfg.TrustProxyProxies},
		ProxyHeader:      cfg.TrustProxyHeader,
	})

	// 순서가 계약이다: requestid(상관관계) → recover → 보안헤더 → cors → rate limit
	// → 메트릭 수집 → 요청 로그 (로그는 최후순위라 latency 측정이 가장 정확)
	app.Use(requestid.New())
	app.Use(fibermw.New())
	app.Use(helmet.New())
	app.Use(cors.New(cors.Config{AllowOrigins: cfg.CORSAllowedOrigins}))
	app.Use(limiter.New(limiter.Config{
		Max:               cfg.RateLimitPerMinute,
		Expiration:        time.Minute,
		LimiterMiddleware: limiter.SlidingWindow{},
	}))
	prom := appmw.NewPrometheus(reg)
	app.Use(prom.Handler())
	app.Use(appmw.RequestLogger(log))

	h := health.NewHandler(func() error { return pingDB(db) })
	app.Get("/livez", h.Livez)
	app.Get("/readyz", h.Readyz)
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.HandlerFor(reg, promhttp.HandlerOpts{})))

	v1 := app.Group("/api/v1")

	var taskGuard []fiber.Handler
	if cfg.AuthEnabled {
		authSvc := auth.NewService(auth.NewDemoAuthenticator(cfg), cfg.AuthJWTSecret, cfg.AuthTokenTTL)
		auth.RegisterRoutes(v1, authSvc) // 로그인은 가드 밖
		taskGuard = []fiber.Handler{auth.RequireAuth(authSvc)}
		log.Info("auth enabled", "protected", "/api/v1/tasks")
	}
	task.RegisterRoutes(v1, taskService(db), taskGuard...)

	return app
}

// errorHandler는 모든 에러를 통일 엔벨로프로 변환한다.
// 미들웨어 바깥에서 실행되므로 요청 로그 미기록 에러도 여기서 남긴다. 500은 내부 상세를 노출하지 않는다.
func errorHandler(log *slog.Logger) fiber.ErrorHandler {
	return func(c fiber.Ctx, err error) error {
		var fe *fiber.Error
		if errors.As(err, &fe) {
			appErr := mapFiberError(fe)
			return c.Status(fe.Code).JSON(common.ErrorBody(appErr))
		}

		appErr := common.AsAppError(err)
		fields := []any{
			"error", err.Error(),
			"method", c.Method(),
			"path", c.Path(),
			"request_id", requestid.FromContext(c),
		}
		if appErr.Status >= http.StatusInternalServerError {
			log.Error("request failed", fields...)
			// 클라이언트에는 고정 메시지만 노출한다.
			safe := *common.ErrInternal
			return c.Status(safe.Status).JSON(common.ErrorBody(&safe))
		}
		log.Warn("request rejected", fields...)
		return c.Status(appErr.Status).JSON(common.ErrorBody(appErr))
	}
}

func mapFiberError(fe *fiber.Error) *common.AppError {
	switch fe.Code {
	case fiber.StatusNotFound:
		return common.ErrNotFound.WithCause(fe)
	case fiber.StatusTooManyRequests:
		return &common.AppError{Code: "RATE_LIMITED", Status: fe.Code, Message: "too many requests"}
	case fiber.StatusRequestEntityTooLarge:
		return &common.AppError{Code: "PAYLOAD_TOO_LARGE", Status: fe.Code, Message: "request body too large"}
	default:
		return &common.AppError{Code: fmt.Sprintf("HTTP_%d", fe.Code), Status: fe.Code, Message: fe.Message}
	}
}
