package auth

import (
	"strings"

	"github.com/gofiber/fiber/v3"

	"go-fiber-starter/internal/apperror"
)

// RequireAuth는 Bearer 토큰을 검증하고 Identity를 Locals에 저장하는 가드다.
//
//	protected := v1.Group("/tasks", auth.RequireAuth(svc))
//
// 핸들러에서 신원 조회: identity, _ := c.Locals(auth.IdentityKey).(auth.Identity)
func RequireAuth(svc *Service) fiber.Handler {
	return func(c fiber.Ctx) error {
		header := c.Get("Authorization")
		token, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || strings.TrimSpace(token) == "" {
			return apperror.ErrUnauthorized
		}
		id, err := svc.Verify(token)
		if err != nil {
			return err
		}
		c.Locals(IdentityKey, id)
		return c.Next()
	}
}
