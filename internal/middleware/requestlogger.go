package middleware

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
)

const RequestIDKey = "request_id"

// RequestLogger는 요청 단위 구조화 로그를 남긴다.
// requestid 미들웨어가 선행 실행되어야 상관관계 ID가 포함된다.
func RequestLogger(log *slog.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		reqID, _ := c.Locals(RequestIDKey).(string)
		status := c.Response().StatusCode()
		fields := []any{
			"method", c.Method(),
			"path", c.Path(),
			"status", status,
			"latency_ms", float64(time.Since(start).Microseconds()) / 1000,
			"ip", c.IP(),
		}
		if reqID != "" {
			fields = append(fields, "request_id", reqID)
		}

		switch {
		case status >= 500:
			log.Error("http_request", fields...)
		case status >= 400:
			log.Warn("http_request", fields...)
		default:
			log.Info("http_request", fields...)
		}
		return err
	}
}
