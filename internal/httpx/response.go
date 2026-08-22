package httpx

import (
	"github.com/gofiber/fiber/v3"

	"go-fiber-starter/internal/apperror"
)

// Envelope는 모든 API 응답의 공통 래퍼다.
type Envelope struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   *Error `json:"error,omitempty"`
	Meta    any    `json:"meta,omitempty"`
}

type Error struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details []apperror.Detail `json:"details,omitempty"`
}

func OK(c fiber.Ctx, data any) error {
	return c.JSON(Envelope{Success: true, Data: data})
}

func Created(c fiber.Ctx, data any) error {
	return c.Status(fiber.StatusCreated).JSON(Envelope{Success: true, Data: data})
}

func OKWithMeta(c fiber.Ctx, data, meta any) error {
	return c.JSON(Envelope{Success: true, Data: data, Meta: meta})
}

func NoContent(c fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// ErrorBody는 전역 ErrorHandler와 핸들러에서 동일한 실패 포맷을 쓰게 한다.
func ErrorBody(appErr *apperror.AppError) Envelope {
	return Envelope{
		Success: false,
		Error: &Error{
			Code:    appErr.Code,
			Message: appErr.Message,
			Details: appErr.Details,
		},
	}
}
