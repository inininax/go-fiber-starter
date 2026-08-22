package task

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"go-fiber-starter/internal/common"
)

// gormRepository는 task.Repository의 GORM 구현체다.
// 인터페이스 정의는 service.go(소비자)에 있다.
type gormRepository struct {
	db *gorm.DB
}

// NewRepository는 GORM 기반 구현체를 반환한다.
func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Create(ctx context.Context, t *Task) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *gormRepository) FindByID(ctx context.Context, id uint) (*Task, error) {
	var t Task
	if err := r.db.WithContext(ctx).First(&t, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrTaskNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *gormRepository) List(ctx context.Context, q common.PageQuery) ([]Task, int64, error) {
	var (
		tasks []Task
		total int64
	)
	if err := r.db.WithContext(ctx).Model(&Task{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := r.db.WithContext(ctx).
		Order("id DESC").
		Limit(q.Limit).
		Offset(q.Offset()).
		Find(&tasks).Error
	return tasks, total, err
}

// Update는 트랜잭션으로 읽기-수정-쓰기를 묶어 부분 갱신의 원자성을 보장한다.
func (r *gormRepository) Update(ctx context.Context, id uint, mutate func(*Task)) (*Task, error) {
	var updated *Task
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var t Task
		if err := tx.First(&t, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return common.ErrNotFound
			}
			return err
		}
		mutate(&t)
		if err := tx.Save(&t).Error; err != nil {
			return err
		}
		updated = &t
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (r *gormRepository) Delete(ctx context.Context, id uint) error {
	res := r.db.WithContext(ctx).Delete(&Task{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrTaskNotFound
	}
	return nil
}
