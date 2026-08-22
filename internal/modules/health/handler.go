package health

import (
	"github.com/gofiber/fiber/v3"
)

// Handler는 쿠버네티스 프로브용 엔드포인트를 제공한다.
// livez(프로세스 생존)와 readyz(의존성 준비)를 분리한다:
//   - livez 실패 → 재시작. DB 장애와 무관해야 하므로 어떤 검사도 하지 않는다.
//   - readyz 실패 → 트래픽 제외. DB ping 등 의존성 상태를 확인한다.
type Handler struct {
	pingDB func() error
}

func NewHandler(pingDB func() error) *Handler {
	return &Handler{pingDB: pingDB}
}

// Livez는 GET /livez
func (h *Handler) Livez(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}

// Readyz는 GET /readyz
func (h *Handler) Readyz(c fiber.Ctx) error {
	if err := h.pingDB(); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status": "unavailable",
			"checks": fiber.Map{"database": "down"},
		})
	}
	return c.JSON(fiber.Map{"status": "ready", "checks": fiber.Map{"database": "up"}})
}
