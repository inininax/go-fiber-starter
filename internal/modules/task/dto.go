package task

import "time"

// CreateRequest는 POST /tasks 요청 본문이다. validate 태그는 fiber StructValidator가 실행한다.
type CreateRequest struct {
	Title   string     `json:"title" validate:"required,min=1,max=200"`
	DueDate *time.Time `json:"due_date"`
}

// UpdateRequest는 PATCH /tasks/:id 요청 본문이다.
// 포인터 필드로 present-field만 반영하는 부분 수정을 지원한다 (nil = 변경 없음).
type UpdateRequest struct {
	Title   *string     `json:"title" validate:"omitempty,min=1,max=200"`
	Done    *bool       `json:"done"`
	DueDate **time.Time `json:"due_date"` // null로 명시적 해제, 값으로 변경, 미전달 = 무시
}

// Response는 API 응답 DTO다. 도메인 모델을 그대로 노출하지 않는다 (스키마 진화 자유도 확보).
type Response struct {
	ID        uint       `json:"id"`
	Title     string     `json:"title"`
	Done      bool       `json:"done"`
	DueDate   *time.Time `json:"due_date,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func toResponse(t Task) Response {
	return Response{
		ID:        t.ID,
		Title:     t.Title,
		Done:      t.Done,
		DueDate:   t.DueDate,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

func toResponses(tasks []Task) []Response {
	out := make([]Response, 0, len(tasks))
	for _, t := range tasks {
		out = append(out, toResponse(t))
	}
	return out
}
