package task

import (
	"context"
	"errors"
	"sync"
	"testing"

	"go-fiber-starter/internal/common"
)

// fakeRepo는 Repository 인터페이스의 인메모리 대역이다(외부 I/O 없음).
type fakeRepo struct {
	mu     sync.Mutex
	tasks  map[uint]*Task
	nextID uint
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{tasks: map[uint]*Task{}, nextID: 1}
}

func (f *fakeRepo) Create(_ context.Context, t *Task) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	t.ID = f.nextID
	f.nextID++
	cp := *t
	f.tasks[t.ID] = &cp
	return nil
}

func (f *fakeRepo) FindByID(_ context.Context, id uint) (*Task, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	t, ok := f.tasks[id]
	if !ok {
		return nil, common.ErrTaskNotFound
	}
	cp := *t
	return &cp, nil
}

func (f *fakeRepo) List(_ context.Context, q common.PageQuery) ([]Task, int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]Task, 0, len(f.tasks))
	for _, t := range f.tasks {
		out = append(out, *t)
	}
	return out, int64(len(out)), nil
}

func (f *fakeRepo) Update(_ context.Context, id uint, mutate func(*Task)) (*Task, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	t, ok := f.tasks[id]
	if !ok {
		return nil, common.ErrNotFound
	}
	mutate(t)
	return t, nil
}

func (f *fakeRepo) Delete(_ context.Context, id uint) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.tasks[id]; !ok {
		return common.ErrTaskNotFound
	}
	delete(f.tasks, id)
	return nil
}

func newTestService() *Service {
	return NewService(newFakeRepo())
}

func TestService_Create(t *testing.T) {
	svc := newTestService()
	res, err := svc.Create(context.Background(), CreateRequest{Title: "buy milk"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if res.ID != 1 || res.Title != "buy milk" || res.Done {
		t.Fatalf("unexpected response: %+v", res)
	}
}

func TestService_Get_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.Get(context.Background(), 999)
	if !errors.Is(err, common.ErrTaskNotFound) {
		t.Fatalf("want ErrTaskNotFound, got %v", err)
	}
}

func TestService_Update_PartialSemantics(t *testing.T) {
	svc := newTestService()
	created, _ := svc.Create(context.Background(), CreateRequest{Title: "original"})

	newTitle := "updated"
	res, err := svc.Update(context.Background(), created.ID, UpdateRequest{Title: &newTitle})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if res.Title != "updated" {
		t.Fatalf("title not applied: %+v", res)
	}
	if res.Done { // 미전달 필드는 변경되면 안 된다
		t.Fatalf("done should remain false: %+v", res)
	}
}

func TestService_Update_NotFound(t *testing.T) {
	svc := newTestService()
	title := "x"
	_, err := svc.Update(context.Background(), 42, UpdateRequest{Title: &title})
	if !errors.Is(err, common.ErrTaskNotFound) {
		t.Fatalf("want ErrTaskNotFound, got %v", err)
	}
}

func TestService_Delete_NotFound(t *testing.T) {
	svc := newTestService()
	if err := svc.Delete(context.Background(), 7); !errors.Is(err, common.ErrTaskNotFound) {
		t.Fatalf("want ErrTaskNotFound, got %v", err)
	}
}

func TestService_List_ClampsLimit(t *testing.T) {
	q := common.NewPageQuery(1, 1000)
	if q.Limit != common.MaxLimit {
		t.Fatalf("limit must be clamped to %d, got %d", common.MaxLimit, q.Limit)
	}
}
