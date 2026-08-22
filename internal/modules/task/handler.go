package task

import (
	"strconv"

	"github.com/gofiber/fiber/v3"

	"go-fiber-starter/internal/apperror"
	"go-fiber-starter/internal/httpx"
	"go-fiber-starter/internal/pagination"
	"go-fiber-starter/internal/validator"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Create는 POST /api/v1/tasks
func (h *Handler) Create(c fiber.Ctx) error {
	var req CreateRequest
	if err := c.Bind().Body(&req); err != nil {
		return validator.BindErrorToAppError(err)
	}
	res, err := h.svc.Create(c.Context(), req)
	if err != nil {
		return err
	}
	return httpx.Created(c, res)
}

// List는 GET /api/v1/tasks?page=&limit=
func (h *Handler) List(c fiber.Ctx) error {
	page := atoiDefault(c.Query("page"), pagination.DefaultPage)
	limit := atoiDefault(c.Query("limit"), pagination.DefaultLimit)

	items, meta, err := h.svc.List(c.Context(), page, limit)
	if err != nil {
		return err
	}
	return httpx.OKWithMeta(c, items, meta)
}

// Get은 GET /api/v1/tasks/:id
func (h *Handler) Get(c fiber.Ctx) error {
	id, err := parseID(c.Params("id"))
	if err != nil {
		return err
	}
	res, err := h.svc.Get(c.Context(), id)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// Update는 PATCH /api/v1/tasks/:id (부분 수정)
func (h *Handler) Update(c fiber.Ctx) error {
	id, err := parseID(c.Params("id"))
	if err != nil {
		return err
	}
	var req UpdateRequest
	if err := c.Bind().Body(&req); err != nil {
		return validator.BindErrorToAppError(err)
	}
	res, err := h.svc.Update(c.Context(), id, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, res)
}

// Delete는 DELETE /api/v1/tasks/:id
func (h *Handler) Delete(c fiber.Ctx) error {
	id, err := parseID(c.Params("id"))
	if err != nil {
		return err
	}
	if err := h.svc.Delete(c.Context(), id); err != nil {
		return err
	}
	return httpx.NoContent(c)
}

func parseID(raw string) (uint, error) {
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		return 0, apperror.NewBadRequest("path parameter 'id' must be a positive integer")
	}
	return uint(id), nil
}

func atoiDefault(raw string, def int) int {
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 {
		return def
	}
	return n
}
