package task

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"

	"go-fiber-starter/internal/apperror"
	"go-fiber-starter/internal/httpx"
	"go-fiber-starter/internal/testutil"
	"go-fiber-starter/internal/validator"
)

// newTestApp은 실제 GORM(sqlite 파일) + 실제 라우트를 갖춘 통합 테스트용 앱을 만든다.
// TB 인터페이스로 테스트(T)와 벤치마크(B) 모두 지원한다.
func newTestApp(tb testing.TB) *fiber.App {
	tb.Helper()

	dsn := filepath.Join(tb.TempDir(), "test.db")
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		tb.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&Task{}); err != nil {
		tb.Fatalf("migrate: %v", err)
	}

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			appErr := apperror.AsAppError(err)
			return c.Status(appErr.Status).JSON(httpx.ErrorBody(appErr))
		},
		StructValidator: validator.NewFiberStructValidator(),
	})
	RegisterRoutes(app.Group("/api/v1"), NewService(NewRepository(db)))
	return app
}

func TestHandler_Create_Returns201AndEnvelope(t *testing.T) {
	app := newTestApp(t)
	resp, env := testutil.Do(t, app, http.MethodPost, "/api/v1/tasks", `{"title":"write tests"}`, "")

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("want 201, got %d (%+v)", resp.StatusCode, env)
	}
	if !env.Success || env.Error != nil {
		t.Fatalf("unexpected envelope: %+v", env)
	}
	var data Response
	if err := json.Unmarshal(env.Data, &data); err != nil {
		t.Fatal(err)
	}
	if data.Title != "write tests" || data.ID == 0 {
		t.Fatalf("unexpected data: %+v", data)
	}
}

func TestHandler_Create_ValidationFails_422WithDetails(t *testing.T) {
	app := newTestApp(t)
	resp, env := testutil.Do(t, app, http.MethodPost, "/api/v1/tasks", `{"title":""}`, "")

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("want 422, got %d", resp.StatusCode)
	}
	if env.Success || env.Error == nil || len(env.Error.Code) == 0 {
		t.Fatalf("expected error envelope: %+v", env)
	}
}

func TestHandler_Create_MalformedBody_400(t *testing.T) {
	app := newTestApp(t)
	resp, env := testutil.Do(t, app, http.MethodPost, "/api/v1/tasks", `{not-json`, "")

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("want 400, got %d (%+v)", resp.StatusCode, env)
	}
}

func TestHandler_List_ReturnsMeta(t *testing.T) {
	app := newTestApp(t)
	for i := range 3 {
		body := fmt.Sprintf(`{"title":"task-%d"}`, i)
		if resp, _ := testutil.Do(t, app, http.MethodPost, "/api/v1/tasks", body, ""); resp.StatusCode != http.StatusCreated {
			t.Fatalf("seed create failed: %d", resp.StatusCode)
		}
	}

	resp, env := testutil.Do(t, app, http.MethodGet, "/api/v1/tasks?page=1&limit=2", "", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200, got %d", resp.StatusCode)
	}
	var meta struct {
		Page       int   `json:"page"`
		Limit      int   `json:"limit"`
		TotalCount int64 `json:"total_count"`
		TotalPages int64 `json:"total_pages"`
	}
	if err := json.Unmarshal(env.Meta, &meta); err != nil {
		t.Fatal(err)
	}
	if meta.TotalCount != 3 || meta.TotalPages != 2 || meta.Limit != 2 {
		t.Fatalf("unexpected meta: %+v", meta)
	}
	var items []Response
	if err := json.Unmarshal(env.Data, &items); err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 { // limit=2 적용 확인
		t.Fatalf("want 2 items, got %d", len(items))
	}
}

func TestHandler_Get_NotFound_EnvelopeCode(t *testing.T) {
	app := newTestApp(t)
	resp, env := testutil.Do(t, app, http.MethodGet, "/api/v1/tasks/999", "", "")

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("want 404, got %d", resp.StatusCode)
	}
	if env.Error == nil || env.Error.Code != CodeTaskNotFound {
		t.Fatalf("expected TASK_NOT_FOUND code, got %+v", env.Error)
	}
}

func TestHandler_Update_PartialPatch(t *testing.T) {
	app := newTestApp(t)
	_, createdEnv := testutil.Do(t, app, http.MethodPost, "/api/v1/tasks", `{"title":"keep me"}`, "")
	var createdData Response
	_ = json.Unmarshal(createdEnv.Data, &createdData)

	url := fmt.Sprintf("/api/v1/tasks/%d", createdData.ID)
	resp, env := testutil.Do(t, app, http.MethodPatch, url, `{"done":true}`, "")

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200, got %d (%+v)", resp.StatusCode, env)
	}
	var data Response
	if err := json.Unmarshal(env.Data, &data); err != nil {
		t.Fatal(err)
	}
	if !data.Done || data.Title != "keep me" {
		t.Fatalf("partial patch broken: %+v", data)
	}
}

func TestHandler_Delete_204Then404(t *testing.T) {
	app := newTestApp(t)
	_, createdEnv := testutil.Do(t, app, http.MethodPost, "/api/v1/tasks", `{"title":"temp"}`, "")
	var data Response
	_ = json.Unmarshal(createdEnv.Data, &data)

	url := fmt.Sprintf("/api/v1/tasks/%d", data.ID)
	resp, _ := testutil.Do(t, app, http.MethodDelete, url, "", "")
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("want 204, got %d", resp.StatusCode)
	}
	resp, env := testutil.Do(t, app, http.MethodGet, url, "", "")
	if resp.StatusCode != http.StatusNotFound || env.Error.Code != CodeTaskNotFound {
		t.Fatalf("want 404 TASK_NOT_FOUND after delete, got %d %+v", resp.StatusCode, env.Error)
	}
}

func TestHandler_BadIDParam_400(t *testing.T) {
	app := newTestApp(t)
	resp, env := testutil.Do(t, app, http.MethodGet, "/api/v1/tasks/abc", "", "")
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("want 400, got %d (%+v)", resp.StatusCode, env)
	}
}

// due_date는 **time.Time으로 null 해제를 지원한다 — 이 계약의 유일한 검증이다.
func TestHandler_Update_DueDateNullClears(t *testing.T) {
	app := newTestApp(t)
	_, createdEnv := testutil.Do(t, app, http.MethodPost, "/api/v1/tasks",
		`{"title":"with due","due_date":"2030-01-02T03:04:05Z"}`, "")
	var created Response
	_ = json.Unmarshal(createdEnv.Data, &created)
	if created.DueDate == nil {
		t.Fatal("seed must carry due_date")
	}

	url := fmt.Sprintf("/api/v1/tasks/%d", created.ID)
	resp, env := testutil.Do(t, app, http.MethodPatch, url, `{"due_date":null}`, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200, got %d (%+v)", resp.StatusCode, env)
	}
	var data Response
	if err := json.Unmarshal(env.Data, &data); err != nil {
		t.Fatal(err)
	}
	if data.DueDate != nil {
		t.Fatalf("due_date must be cleared by explicit null: %+v", data.DueDate)
	}
}
