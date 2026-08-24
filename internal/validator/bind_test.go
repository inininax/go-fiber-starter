package validator

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/go-playground/validator/v10"

	"go-fiber-starter/internal/apperror"
)

type sampleDTO struct {
	Title string `validate:"required,min=1,max=200"`
}

func validationErrorOf(dto any) error {
	v := validator.New(validator.WithRequiredStructEnabled())
	return v.Struct(dto) // validator.ValidationErrors 반환
}

func TestBindErrorToAppError_ValidationViolation_422WithFieldDetails(t *testing.T) {
	err := fmt.Errorf("bind: %w", validationErrorOf(sampleDTO{Title: ""}))

	got := BindErrorToAppError(err)
	if got.Status != http.StatusUnprocessableEntity {
		t.Fatalf("want 422, got %d", got.Status)
	}
	if got.Code != apperror.CodeInvalidRequest {
		t.Fatalf("want %s, got %s", apperror.CodeInvalidRequest, got.Code)
	}
	if len(got.Details) == 0 || got.Details[0].Field != "Title" {
		t.Fatalf("field detail missing: %+v", got.Details)
	}
}

func TestBindErrorToAppError_ParseFailure_400(t *testing.T) {
	err := errors.New("cannot parse body")

	got := BindErrorToAppError(err)
	if got.Status != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", got.Status)
	}
	if got.Message != "malformed request body" {
		t.Fatalf("unexpected message: %q", got.Message)
	}
	if len(got.Details) != 0 {
		t.Fatalf("parse failure must not carry field details: %+v", got.Details)
	}
}

// 실제 Fiber Bind 경로는 *fiber.BindError로 감싸므로 래핑된 형태도 동일 매핑되어야 한다.
func TestBindErrorToAppError_WrappedValidationErrors_Still422(t *testing.T) {
	inner := validationErrorOf(sampleDTO{Title: ""})
	wrapped := fmt.Errorf("outer: %w", inner)

	got := BindErrorToAppError(wrapped)
	if got.Status != http.StatusUnprocessableEntity {
		t.Fatalf("wrapped validation must stay 422, got %d", got.Status)
	}
}
