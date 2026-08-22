package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"

	"go-fiber-starter/internal/common"
)

// localErrorHandler는 router.go 전역 ErrorHandler와 동일한 상태 매핑을 재현한다.
// middleware 패키지에서 router를 임포트하면 의존성 사이클 위험이 있어 로컬 복제로 일치를 검증한다.
func localErrorHandler(c fiber.Ctx, err error) error {
	var fe *fiber.Error
	if errors.As(err, &fe) {
		return c.Status(fe.Code).JSON(common.ErrorBody(common.ErrNotFound.WithCause(fe)))
	}

	appErr := common.AsAppError(err)
	if appErr.Status >= 500 {
		// 클라이언트에는 고정 메시지 500만 노출한다(router.go와 동일).
		safe := *common.ErrInternal
		return c.Status(safe.Status).JSON(common.ErrorBody(&safe))
	}
	return c.Status(appErr.Status).JSON(common.ErrorBody(appErr))
}

// statusVia는 핸들러가 반환한 에러를 언와인드 시점에 EffectiveStatus로 관측한다.
func statusVia(h fiber.Handler) int {
	var got int
	app := fiber.New()
	app.Get("/x", func(c fiber.Ctx) error {
		err := h(c)
		got = EffectiveStatus(c, err)
		return err
	})
	req := httptest.NewRequest("GET", "/x", nil)
	_, _ = app.Test(req)
	return got
}

func TestEffectiveStatus_NilError_ReturnsResponseStatus(t *testing.T) {
	got := statusVia(func(c fiber.Ctx) error {
		return c.SendStatus(503)
	})
	if got != 503 {
		t.Fatalf("want 503, got %d", got)
	}
}

func TestEffectiveStatus_FiberError_UsesCode(t *testing.T) {
	got := statusVia(func(c fiber.Ctx) error {
		return fiber.NewError(fiber.StatusTeapot, "teapot")
	})
	if got != fiber.StatusTeapot {
		t.Fatalf("want %d, got %d", fiber.StatusTeapot, got)
	}
}

func TestEffectiveStatus_AppError_Wrapped_UsesStatus(t *testing.T) {
	err := fmt.Errorf("wrap: %w", common.ErrTaskNotFound)
	got := statusVia(func(c fiber.Ctx) error { return err })
	if got != http.StatusNotFound {
		t.Fatalf("want 404, got %d", got)
	}
}

func TestEffectiveStatus_UnknownError_Is500(t *testing.T) {
	got := statusVia(func(c fiber.Ctx) error { return errors.New("boom") })
	if got != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d", got)
	}
}

// errorHandler는 >=500을 고정 500으로 응답하므로 관측값도 클램프돼야 한다.
func TestEffectiveStatus_AppErrorAbove500_ClampsTo500(t *testing.T) {
	err := fmt.Errorf("wrap: %w", &common.AppError{Code: "SERVICE_UNAVAILABLE", Status: 503, Message: "down"})
	got := statusVia(func(c fiber.Ctx) error { return err })
	if got != http.StatusInternalServerError {
		t.Fatalf("want clamped 500, got %d", got)
	}
}

// TestEffectiveStatus_MatchesGlobalErrorHandler는 실제 응답 코드와 EffectiveStatus가
// 모든 에러 유형에서 일치함을 검증한다(관측값 ≠ 응답 코드 재발 방지).
func TestEffectiveStatus_MatchesGlobalErrorHandler(t *testing.T) {
	cases := []struct {
		name   string
		inject func(c fiber.Ctx) error
	}{
		{"NilErrorWithSendStatus", func(c fiber.Ctx) error { return c.SendStatus(http.StatusServiceUnavailable) }},
		{"FiberError", func(c fiber.Ctx) error { return fiber.NewError(fiber.StatusTeapot, "teapot") }},
		{"WrappedAppError", func(c fiber.Ctx) error {
			return fmt.Errorf("wrap: %w", common.ErrTaskNotFound)
		}},
		{"UnknownError", func(c fiber.Ctx) error { return errors.New("boom") }},
		{"AppErrorAbove500", func(c fiber.Ctx) error {
			return &common.AppError{Code: "SERVICE_UNAVAILABLE", Status: 503, Message: "down"}
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var observed int
			app := fiber.New(fiber.Config{ErrorHandler: localErrorHandler})
			// 관측 지점은 체인 최외곽(미들웨어 언와인드 시점)이므로 라우트보다 먼저 등록한다.
			app.Use(func(c fiber.Ctx) error {
				err := c.Next()
				observed = EffectiveStatus(c, err)
				return err
			})
			app.Get("/boom", tc.inject)

			req := httptest.NewRequest("GET", "/boom", nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			if observed != resp.StatusCode {
				t.Fatalf("observed %d != actual %d", observed, resp.StatusCode)
			}
		})
	}
}

// jsonDecodeLogLine은 slog JSON 핸들러 버퍼의 첫 줄을 파싱한다.
func jsonDecodeLogLine(t *testing.T, buf string) map[string]any {
	t.Helper()
	line := strings.TrimSpace(buf)
	if line == "" {
		t.Fatal("no log output")
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(line), &m); err != nil {
		t.Fatalf("decode log %q: %v", line, err)
	}
	return m
}
