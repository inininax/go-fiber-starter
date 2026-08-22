package task

import "github.com/gofiber/fiber/v3"

// RegisterRoutes는 /api/v1 하위에 이 모듈의 라우트를 마운트한다.
//
// 새 모듈 추가 패턴 (README 참고):
//  1. internal/modules/<name> 패키지를 task와 동일 구조로 생성
//  2. NewService(NewRepository(db))로 조립
//  3. RegisterRoutes(v1 fiber.Router) 형태의 함수 제공
//  4. router.New에서 한 줄 추가
func RegisterRoutes(v1 fiber.Router, svc *Service) {
	h := NewHandler(svc)
	tasks := v1.Group("/tasks")

	tasks.Post("/", h.Create)
	tasks.Get("/", h.List)
	tasks.Get("/:id", h.Get)
	tasks.Patch("/:id", h.Update)
	tasks.Delete("/:id", h.Delete)
}
