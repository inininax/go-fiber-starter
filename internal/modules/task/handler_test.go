package task

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"

	"go-fiber-starter/internal/common"
	"go-fiber-starter/internal/validator"
)

// newTestApp은 실제 GORM(sqlite 파일) + 실제 라우트를 갖춘 통합 테스트용 앱을 만든다.
func newTestApp(t *testing.T) *fiber.App {
	t.Helper()

	dsn := filepath.Join(t.TempDir(), "test.db")
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&Task{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			appErr := common.AsAppError(err)
			return c.Status(appErr.Status).JSON(common.ErrorBody(appErr))
		},
		StructValidator: validator.NewFiberStructValidator(),
	})
	RegisterRoutes(app.Group("/api/v1"), NewService(NewRepository(db)))
	return app
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type apiEnvelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Meta    json.RawMessage `json:"meta"`
	Error   *apiError       `json:"error"`
}

func do(t *testing.T, app *fiber.App, method, target, body string) (*http.Response, apiEnvelope) {
	t.Helper()
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, target, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, target, nil)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, target, err)
	}
	var env apiEnvelope
	// 204 등 본문 없는 응답을 허용한다.
	if resp.StatusCode != http.StatusNoContent && resp.ContentLength != 0 {
		if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
			t.Fatalf("decode response: %v", err)
		}
	}
	return resp, env
}

func TestHandler_Create_Returns201AndEnvelope(t *testing.T) {
	app := newTestApp(t)
	resp, env := do(t, app, http.MethodPost, "/api/v1/tasks", `{"title":"write tests"}`)

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
	resp, env := do(t, app, http.MethodPost, "/api/v1/tasks", `{"title":""}`)

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("want 422, got %d", resp.StatusCode)
	}
	if env.Success || env.Error == nil || len(env.Error.Code) == 0 {
		t.Fatalf("expected error envelope: %+v", env)
	}
}

func TestHandler_Create_MalformedBody_400(t *testing.T) {
	app := newTestApp(t)
	resp, env := do(t, app, http.MethodPost, "/api/v1/tasks", `{not-json`)

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("want 400, got %d (%+v)", resp.StatusCode, env)
	}
}

func TestHandler_List_ReturnsMeta(t *testing.T) {
	app := newTestApp(t)
	for i := range 3 {
		body := fmt.Sprintf(`{"title":"task-%d"}`, i)
		if resp, _ := do(t, app, http.MethodPost, "/api/v1/tasks", body); resp.StatusCode != http.StatusCreated {
			t.Fatalf("seed create failed: %d", resp.StatusCode)
		}
	}

	resp, env := do(t, app, http.MethodGet, "/api/v1/tasks?page=1&limit=2", "")
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
	resp, env := do(t, app, http.MethodGet, "/api/v1/tasks/999", "")

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("want 404, got %d", resp.StatusCode)
	}
	if env.Error == nil || env.Error.Code != common.CodeTaskNotFound {
		t.Fatalf("expected TASK_NOT_FOUND code, got %+v", env.Error)
	}
}

func TestHandler_Update_PartialPatch(t *testing.T) {
	app := newTestApp(t)
	_, createdEnv := do(t, app, http.MethodPost, "/api/v1/tasks", `{"title":"keep me"}`)
	var createdData Response
	_ = json.Unmarshal(createdEnv.Data, &createdData)

	url := fmt.Sprintf("/api/v1/tasks/%d", createdData.ID)
	resp, env := do(t, app, http.MethodPatch, url, `{"done":true}`)

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
	_, createdEnv := do(t, app, http.MethodPost, "/api/v1/tasks", `{"title":"temp"}`)
	var data Response
	_ = json.Unmarshal(createdEnv.Data, &data)

	url := fmt.Sprintf("/api/v1/tasks/%d", data.ID)
	resp, _ := do(t, app, http.MethodDelete, url, "")
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("want 204, got %d", resp.StatusCode)
	}
	resp, env := do(t, app, http.MethodGet, url, "")
	if resp.StatusCode != http.StatusNotFound || env.Error.Code != common.CodeTaskNotFound {
		t.Fatalf("want 404 TASK_NOT_FOUND after delete, got %d %+v", resp.StatusCode, env.Error)
	}
}

func TestHandler_BadIDParam_400(t *testing.T) {
	app := newTestApp(t)
	resp, env := do(t, app, http.MethodGet, "/api/v1/tasks/abc", "")
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("want 400, got %d (%+v)", resp.StatusCode, env)
	}
}
