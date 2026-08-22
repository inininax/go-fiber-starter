package common

import (
	"errors"
	"fmt"
	"net/http"
)

// 에러 코드 카탈로그: 클라이언트 노출용 안정 식별자. 코드 추가 시 여기만 수정.
const (
	CodeInvalidRequest    = "INVALID_REQUEST"
	CodeNotFound          = "NOT_FOUND"
	CodeConflict          = "CONFLICT"
	CodeInternal          = "INTERNAL_ERROR"
	CodeUnauthorized      = "UNAUTHORIZED"
	CodeInvalidCredential = "INVALID_CREDENTIALS"

	CodeTaskNotFound = "TASK_NOT_FOUND"
)

// AppError는 도메인/전송 계층에서 사용하는 비즈니스 에러다.
// service/repository는 이 타입(또는 센티널)을 반환하고, router의 전역 ErrorHandler가
// HTTP 응답 엔벨로프로 변환한다.
type AppError struct {
	Code    string   `json:"code"`
	Status  int      `json:"-"`
	Message string   `json:"message"`
	Details []Detail `json:"details,omitempty"`
	err     error    // 원인(내부 로깅용, 응답에는 미노출)
}

type Detail struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

func (e *AppError) Error() string {
	if e.err != nil {
		return fmt.Sprintf("%s (%s): %v", e.Message, e.Code, e.err)
	}
	return fmt.Sprintf("%s (%s)", e.Message, e.Code)
}

func (e *AppError) Unwrap() error { return e.err }

// Is는 코드 기반 동등성을 제공한다. WithCause로 복사된 인스턴스도
// errors.Is(err, common.ErrXxx) 매칭이 되도록 한다.
func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// WithCause는 원인 에러를 붙여 새 AppError를 반환한다 (체이닝용).
func (e *AppError) WithCause(err error) *AppError {
	cp := *e
	cp.err = err
	return &cp
}

var (
	ErrInvalidRequest = &AppError{Code: CodeInvalidRequest, Status: http.StatusUnprocessableEntity, Message: "invalid request"}
	ErrNotFound       = &AppError{Code: CodeNotFound, Status: http.StatusNotFound, Message: "resource not found"}
	ErrConflict       = &AppError{Code: CodeConflict, Status: http.StatusConflict, Message: "resource conflict"}
	ErrInternal       = &AppError{Code: CodeInternal, Status: http.StatusInternalServerError, Message: "internal server error"}

	ErrUnauthorized      = &AppError{Code: CodeUnauthorized, Status: http.StatusUnauthorized, Message: "authentication required"}
	ErrInvalidCredential = &AppError{Code: CodeInvalidCredential, Status: http.StatusUnauthorized, Message: "invalid username or password"}

	ErrTaskNotFound = (&AppError{Code: CodeTaskNotFound, Status: http.StatusNotFound, Message: "task not found"})
)

func NewBadRequest(msg string, details ...Detail) *AppError {
	return &AppError{Code: CodeInvalidRequest, Status: http.StatusBadRequest, Message: msg, Details: details}
}

func NewValidation(details []Detail) *AppError {
	return &AppError{Code: CodeInvalidRequest, Status: http.StatusUnprocessableEntity, Message: "validation failed", Details: details}
}

// AsAppError는 err를 *AppError로 변환한다. 아니면 INTERNAL_ERROR로 래핑한다.
func AsAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return ErrInternal.WithCause(err)
}
