package validator

import (
	"errors"

	"github.com/go-playground/validator/v10"

	"go-fiber-starter/internal/apperror"
)

// BindErrorToAppError는 Fiber Bind 실패를 통일된 에러로 변환한다.
// 파싱 실패 → 400, validate 태그 위반 → 422 + 필드별 details.
// 모든 모듈 핸들러가 이 함수 하나를 써야 클라이언트 에러 계약이 일관된다.
func BindErrorToAppError(err error) *apperror.AppError {
	var valErrs validator.ValidationErrors
	if errors.As(err, &valErrs) {
		details := make([]apperror.Detail, 0, len(valErrs))
		for _, ve := range valErrs {
			details = append(details, apperror.Detail{Field: ve.Field(), Reason: ve.ActualTag()})
		}
		return apperror.NewValidation(details).WithCause(err)
	}
	return apperror.NewBadRequest("malformed request body").WithCause(err)
}
