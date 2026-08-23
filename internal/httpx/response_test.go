package httpx

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"

	"go-fiber-starter/internal/apperror"
)

func newTestApp() *fiber.App {
	return fiber.New(fiber.Config{ErrorHandler: func(c fiber.Ctx, err error) error { return err }})
}

func decode(t *testing.T, body []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatalf("decode %s: %v", body, err)
	}
	return m
}

func TestOK_EnvelopeShape(t *testing.T) {
	app := newTestApp()
	app.Get("/", func(c fiber.Ctx) error {
		return OK(c, fiber.Map{"id": 1})
	})
	resp, err := app.Test(newGet("/"))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200, got %d", resp.StatusCode)
	}
	m := decode(t, bodyBytes(resp))
	if m["success"] != true {
		t.Fatalf("success flag missing: %v", m)
	}
	data, ok := m["data"].(map[string]any)
	if !ok || data["id"] != float64(1) {
		t.Fatalf("data payload broken: %v", m["data"])
	}
}

func TestCreated_Status201(t *testing.T) {
	app := newTestApp()
	app.Post("/", func(c fiber.Ctx) error { return Created(c, fiber.Map{}) })
	resp, err := app.Test(newPost("/"))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("want 201, got %d", resp.StatusCode)
	}
}

func TestNoContent_EmptyBody(t *testing.T) {
	app := newTestApp()
	app.Delete("/", func(c fiber.Ctx) error { return NoContent(c) })
	resp, err := app.Test(newDelete("/"))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent || len(bodyBytes(resp)) != 0 {
		t.Fatalf("want 204 + empty body, got %d (%q)", resp.StatusCode, bodyBytes(resp))
	}
}

func TestErrorBody_MirrorsAppErrorContract(t *testing.T) {
	appErr := &apperror.AppError{
		Code:    "TASK_NOT_FOUND",
		Status:  http.StatusNotFound,
		Message: "task not found",
		Details: []apperror.Detail{{Field: "Title", Reason: "required"}},
	}
	app := newTestApp()
	app.Get("/", func(c fiber.Ctx) error {
		return c.Status(appErr.Status).JSON(ErrorBody(appErr))
	})
	resp, err := app.Test(newGet("/"))
	if err != nil {
		t.Fatal(err)
	}
	m := decode(t, bodyBytes(resp))
	if m["success"] != false {
		t.Fatalf("success must be false on error envelope: %v", m)
	}
	errObj, _ := m["error"].(map[string]any)
	if errObj == nil || errObj["code"] != "TASK_NOT_FOUND" || errObj["message"] != "task not found" {
		t.Fatalf("error contract broken: %v", m["error"])
	}
	details, _ := errObj["details"].([]any)
	if len(details) != 1 {
		t.Fatalf("details lost: %v", errObj["details"])
	}
}
