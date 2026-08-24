package task

import (
	"encoding/json"
	"time"
)

// CreateRequest는 POST /tasks 요청 본문이다. validate 태그는 fiber StructValidator가 실행한다.
type CreateRequest struct {
	Title   string     `json:"title" validate:"required,min=1,max=200"`
	DueDate *time.Time `json:"due_date"`
}

// NullableTime은 PATCH 부분 수정에서 세 가지 상태를 구분한다:
// 미전달(Set=false), 명시적 null 해제(Null=true), 값 변경(Null=false, Set=true).
// 표준 unmarshal로는 **time.Time의 null과 미전달이 구분되지 않아 필요하다.
type NullableTime struct {
	Set   bool
	Null  bool
	Value time.Time
}

func (n *NullableTime) UnmarshalJSON(data []byte) error {
	n.Set = true
	if string(data) == "null" {
		n.Null = true
		return nil
	}
	return json.Unmarshal(data, &n.Value)
}

// UpdateRequest는 PATCH /tasks/:id 요청 본문이다.
// 포인터 필드로 present-field만 반영하는 부분 수정을 지원한다 (nil = 변경 없음).
type UpdateRequest struct {
	Title   *string      `json:"title" validate:"omitempty,min=1,max=200"`
	Done    *bool        `json:"done"`
	DueDate NullableTime `json:"due_date"`
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
