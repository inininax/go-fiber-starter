package task

import (
	"net/http"

	"go-fiber-starter/internal/apperror"
)

// CodeTaskNotFound는 task 도메인 전용 에러 코드다. 범용 코드는 apperror 카탈로그에 있다.
const CodeTaskNotFound = "TASK_NOT_FOUND"

// ErrTaskNotFound는 미존재 task 조회의 센티널이다.
var ErrTaskNotFound = &apperror.AppError{
	Code: CodeTaskNotFound, Status: http.StatusNotFound, Message: "task not found",
}
