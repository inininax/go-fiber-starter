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
	commit string
}

// NewHandler는 DB ping 검사기와 빌드 commit 식별자를 받는다.
// commit은 -ldflags "-X main.commit=..."로 주입되며 livez 응답으로 배포 버전을 특정한다.
// 테스트 등 조립 경로에서 미주입 시 빈 문자열 대신 "dev"로 폴백한다.
func NewHandler(pingDB func() error, commit string) *Handler {
	if commit == "" {
		commit = "dev"
	}
	return &Handler{pingDB: pingDB, commit: commit}
}

// Livez는 GET /livez
func (h *Handler) Livez(c fiber.Ctx) error {
	// 200과 {"status","commit"} 스키마는 프로브 호환을 위해 고정이다.
	return c.JSON(fiber.Map{"status": "ok", "commit": h.commit})
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
