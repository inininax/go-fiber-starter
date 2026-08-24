package task

import (
	"context"
	"errors"
	"fmt"

	"go-fiber-starter/internal/apperror"
	"go-fiber-starter/internal/pagination"
)

// Repository는 Service가 필요한 저장소 기능만 정의한 인터페이스다(소비자 정의).
// 구현은 repository.go(GORM), 테스트 대역은 service_test.go(fake)에서 제공한다.
type Repository interface {
	Create(ctx context.Context, t *Task) error
	FindByID(ctx context.Context, id uint) (*Task, error)
	List(ctx context.Context, q pagination.PageQuery) ([]Task, int64, error)
	Update(ctx context.Context, id uint, mutate func(*Task)) (*Task, error)
	Delete(ctx context.Context, id uint) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (Response, error) {
	t := Task{Title: req.Title, DueDate: req.DueDate}
	if err := s.repo.Create(ctx, &t); err != nil {
		return Response{}, fmt.Errorf("create task: %w", err)
	}
	return toResponse(t), nil
}

func (s *Service) Get(ctx context.Context, id uint) (Response, error) {
	t, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return Response{}, fmt.Errorf("get task %d: %w", id, err)
	}
	return toResponse(*t), nil
}

func (s *Service) List(ctx context.Context, page, limit int) ([]Response, pagination.PageMeta, error) {
	q := pagination.NewPageQuery(page, limit)
	tasks, total, err := s.repo.List(ctx, q)
	if err != nil {
		return nil, pagination.PageMeta{}, fmt.Errorf("list tasks: %w", err)
	}
	return toResponses(tasks), pagination.NewPageMeta(q, total), nil
}

// Update는 present-field만 반영한다. nil 포인터 = 변경 없음.
func (s *Service) Update(ctx context.Context, id uint, req UpdateRequest) (Response, error) {
	mutate := func(t *Task) {
		if req.Title != nil {
			t.Title = *req.Title
		}
		if req.Done != nil {
			t.Done = *req.Done
		}
		if req.DueDate.Set {
			if req.DueDate.Null {
				t.DueDate = nil // 명시적 null = 해제
			} else {
				v := req.DueDate.Value
				t.DueDate = &v
			}
		}
	}
	t, err := s.repo.Update(ctx, id, mutate)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return Response{}, ErrTaskNotFound
		}
		return Response{}, fmt.Errorf("update task %d: %w", id, err)
	}
	return toResponse(*t), nil
}

func (s *Service) Delete(ctx context.Context, id uint) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return ErrTaskNotFound
		}
		return fmt.Errorf("delete task %d: %w", id, err)
	}
	return nil
}
