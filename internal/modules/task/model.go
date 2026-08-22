package task

import (
	"time"

	"gorm.io/gorm"
)

// Task는 예제 도메인 모델이다. 새 모듈 추가 시 이 파일 구조를 복제한다.
type Task struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Title     string         `gorm:"size:200;not null" json:"title"`
	Done      bool           `gorm:"not null;default:false;index" json:"done"`
	DueDate   *time.Time     `json:"due_date,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Task) TableName() string { return "tasks" }
