package router

import (
	"gorm.io/gorm"

	"go-fiber-starter/internal/database"
	"go-fiber-starter/internal/modules/task"
)

// 조립(wiring): 각 모듈의 생성자에 의존성을 주입한다.
// DI 프레임워크 없이 명시적 조립을 유지한다(추적 가능성 우선).

func taskService(db *gorm.DB) *task.Service {
	return task.NewService(task.NewRepository(db))
}

func pingDB(db *gorm.DB) error {
	return database.Ping(db)
}
