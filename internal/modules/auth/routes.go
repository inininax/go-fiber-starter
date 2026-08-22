package auth

import "github.com/gofiber/fiber/v3"

// RegisterRoutes는 /api/v1/auth/login 을 마운트한다. 로그인 자체는 인증 대상이 아니다.
func RegisterRoutes(v1 fiber.Router, svc *Service) {
	h := NewHandler(svc)
	v1.Post("/auth/login", h.Login)
}
