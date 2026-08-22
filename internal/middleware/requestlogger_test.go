package middleware

import (
	"bytes"
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

func newLogApp(t *testing.T, withRequestID bool) (*fiber.App, *bytes.Buffer) {
	t.Helper()
	var buf bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&buf, nil))

	app := fiber.New()
	if withRequestID {
		app.Use(requestid.New())
	}
	app.Use(RequestLogger(log))
	app.Get("/ok", func(c fiber.Ctx) error { return c.SendStatus(200) })
	return app, &buf
}

func TestRequestLogger_IncludesRequestID(t *testing.T) {
	app, buf := newLogApp(t, true)

	req := httptest.NewRequest("GET", "/ok", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("want 200, got %d", resp.StatusCode)
	}

	m := jsonDecodeLogLine(t, buf.String())
	rid, ok := m["request_id"]
	if !ok || rid == "" {
		t.Fatalf("request_id must be present and non-empty, got: %v", m)
	}
}

func TestRequestLogger_NoRequestIDMiddleware_OmitsField(t *testing.T) {
	app, buf := newLogApp(t, false)

	req := httptest.NewRequest("GET", "/ok", nil)
	if _, err := app.Test(req); err != nil {
		t.Fatal(err)
	}

	m := jsonDecodeLogLine(t, buf.String())
	if _, ok := m["request_id"]; ok {
		t.Fatalf("request_id must be omitted without requestid middleware, got: %v", m)
	}
}
