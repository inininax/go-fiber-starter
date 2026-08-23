package auth

import "github.com/gofiber/fiber/v3"

// RegisterRoutes는 /api/v1/auth/login 을 마운트한다. 로그인 자체는 인증 대상이 아니다.
// guard는 선택 미들웨어(예: 로그인 전용 rate limiter)로, 순서대로 핸들러 앞에 붙는다.
func RegisterRoutes(v1 fiber.Router, svc *Service, guard ...fiber.Handler) {
	h := NewHandler(svc)
	hs := make([]any, len(guard)+1)
	for i, g := range guard {
		hs[i] = g
	}
	hs[len(guard)] = h.Login
	// Post(path, handler, handlers...)라 첫 요소는 별도 인자로 넘긴다. hs는 항상 최소 1개(Login)다.
	v1.Post("/auth/login", hs[0], hs[1:]...)
}
