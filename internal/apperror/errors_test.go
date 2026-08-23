package apperror

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestSentinelCatalog_StatusesAndCodes(t *testing.T) {
	// 카탈로그 불변식: 코드/상태 쌍이 계약이다. 실수로 센티널을 지우거나 바꾸면 여기서 잡힌다.
	cases := []struct {
		err    *AppError
		code   string
		status int
	}{
		{ErrInvalidRequest, CodeInvalidRequest, http.StatusUnprocessableEntity},
		{ErrNotFound, CodeNotFound, http.StatusNotFound},
		{ErrConflict, CodeConflict, http.StatusConflict},
		{ErrInternal, CodeInternal, http.StatusInternalServerError},
		{ErrUnauthorized, CodeUnauthorized, http.StatusUnauthorized},
		{ErrInvalidCredential, CodeInvalidCredential, http.StatusUnauthorized},
	}
	for _, tc := range cases {
		if tc.err.Code != tc.code || tc.err.Status != tc.status {
			t.Errorf("%s: want (%s,%d), got (%s,%d)", tc.code, tc.code, tc.status, tc.err.Code, tc.err.Status)
		}
		if tc.err.Message == "" {
			t.Errorf("%s: message must not be empty", tc.code)
		}
	}
}

func TestAsAppError_Passthrough(t *testing.T) {
	sample := &AppError{Code: "X_NOT_FOUND", Status: http.StatusNotFound, Message: "x not found"}
	if got := AsAppError(sample); got.Code != "X_NOT_FOUND" {
		t.Fatalf("AppError passthrough broken: %+v", got)
	}
}

func TestAsAppError_UnknownError_BecomesInternal(t *testing.T) {
	got := AsAppError(fmt.Errorf("db down"))
	if got.Code != CodeInternal || got.Status != http.StatusInternalServerError {
		t.Fatalf("want INTERNAL_ERROR/500, got %s/%d", got.Code, got.Status)
	}
}

func TestIs_MatchesAcrossCauseCopies(t *testing.T) {
	wrapped := fmt.Errorf("op failed: %w", ErrNotFound.WithCause(errors.New("row gone")))
	if !errors.Is(wrapped, ErrNotFound) {
		t.Fatal("errors.Is must match sentinel through WithCause copies")
	}
	if errors.Is(wrapped, ErrConflict) {
		t.Fatal("different code must not match")
	}
}

func TestError_ContainsCodeAndCauseForLogs(t *testing.T) {
	e := ErrConflict.WithCause(errors.New("unique violation"))
	msg := e.Error()
	if !strings.Contains(msg, CodeConflict) || !strings.Contains(msg, "unique violation") {
		t.Fatalf("Error() should expose code+cause for logs: %q", msg)
	}
}

func TestNewValidation_DetailsPreserved(t *testing.T) {
	details := []Detail{{Field: "Title", Reason: "required"}}
	got := NewValidation(details).WithCause(errors.New("bind"))
	if len(got.Details) != 1 || got.Details[0].Field != "Title" {
		t.Fatalf("details lost: %+v", got.Details)
	}
	if got.Status != http.StatusUnprocessableEntity {
		t.Fatalf("want 422, got %d", got.Status)
	}
}
