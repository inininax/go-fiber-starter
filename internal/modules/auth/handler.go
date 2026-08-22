package auth

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"go-fiber-starter/internal/httpx"
	"go-fiber-starter/internal/validator"
)

// IdentityKey는 c.Locals에 저장되는 Identity의 키다.
const IdentityKey = "auth.identity"

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token     string    `json:"token"`
	TokenType string    `json:"token_type"`
	ExpiresAt time.Time `json:"expires_at"`
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type Handler struct {
	svc *Service
}

// Login은 POST /api/v1/auth/login
func (h *Handler) Login(c fiber.Ctx) error {
	var req LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return validator.BindErrorToAppError(err) // 파싱 400 / 검증 422 공용 계약
	}
	token, expiresAt, err := h.svc.Issue(c.Context(), req.Username, req.Password)
	if err != nil {
		return err // ErrInvalidCredential 그대로 전달 → 401
	}
	return httpx.OK(c, LoginResponse{Token: token, TokenType: "Bearer", ExpiresAt: expiresAt})
}
