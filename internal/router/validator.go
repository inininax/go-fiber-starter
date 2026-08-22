package router

import (
	"github.com/gofiber/fiber/v3"

	"go-fiber-starter/internal/validator"
)

// newStructValidator는 fiber.Config에 등록할 검증기를 반환한다.
func newStructValidator() fiber.StructValidator {
	return validator.NewFiberStructValidator()
}
