package task

import (
	"errors"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"

	"go-fiber-starter/internal/common"
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
		return bindErrorToAppError(err)
	}
	res, err := h.svc.Create(c.Context(), req)
	if err != nil {
		return err
	}
	return common.Created(c, res)
}

// List는 GET /api/v1/tasks?page=&limit=
func (h *Handler) List(c fiber.Ctx) error {
	page := atoiDefault(c.Query("page"), common.DefaultPage)
	limit := atoiDefault(c.Query("limit"), common.DefaultLimit)

	items, meta, err := h.svc.List(c.Context(), page, limit)
	if err != nil {
		return err
	}
	return common.OKWithMeta(c, items, meta)
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
	return common.OK(c, res)
}

// Update는 PATCH /api/v1/tasks/:id (부분 수정)
func (h *Handler) Update(c fiber.Ctx) error {
	id, err := parseID(c.Params("id"))
	if err != nil {
		return err
	}
	var req UpdateRequest
	if err := c.Bind().Body(&req); err != nil {
		return bindErrorToAppError(err)
	}
	res, err := h.svc.Update(c.Context(), id, req)
	if err != nil {
		return err
	}
	return common.OK(c, res)
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
	return common.NoContent(c)
}

func parseID(raw string) (uint, error) {
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		return 0, common.NewBadRequest("path parameter 'id' must be a positive integer")
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

// bindErrorToAppError는 Fiber Bind 실패를 통일된 에러로 변환한다.
// 파싱 실패 → 400, validator 위반 → 422 + 필드별 상세.
func bindErrorToAppError(err error) *common.AppError {
	var valErrs validator.ValidationErrors
	if errors.As(err, &valErrs) {
		details := make([]common.Detail, 0, len(valErrs))
		for _, ve := range valErrs {
			details = append(details, common.Detail{Field: ve.Field(), Reason: ve.ActualTag()})
		}
		return common.NewValidation(details).WithCause(err)
	}
	return common.NewBadRequest("malformed request body").WithCause(err)
}
