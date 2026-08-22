package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"

	"go-fiber-starter/internal/apperror"
	"go-fiber-starter/internal/httpx"
	"go-fiber-starter/internal/testutil"
	"go-fiber-starter/internal/validator"
)

func newTestApp(t *testing.T, withGuard bool) *fiber.App {
	t.Helper()
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			appErr := apperror.AsAppError(err)
			return c.Status(appErr.Status).JSON(httpx.ErrorBody(appErr))
		},
		StructValidator: validator.NewFiberStructValidator(),
	})
	svc := newTestService(testTTL)
	RegisterRoutes(app.Group("/api/v1"), svc)

	dummy := func(c fiber.Ctx) error { return httpx.OK(c, fiber.Map{"ok": true}) }
	if withGuard {
		app.Get("/protected", RequireAuth(svc), dummy)
	} else {
		app.Get("/protected", dummy)
	}
	return app
}

func TestLogin_Success_ReturnsToken(t *testing.T) {
	app := newTestApp(t, false)
	resp, env := testutil.Do(t, app, http.MethodPost, "/api/v1/auth/login",
		`{"username":"admin","password":"admin123"}`, "")

	if resp.StatusCode != http.StatusOK || !env.Success {
		t.Fatalf("want 200 success, got %d (%+v)", resp.StatusCode, env)
	}
	var data LoginResponse
	if err := json.Unmarshal(env.Data, &data); err != nil {
		t.Fatal(err)
	}
	if data.Token == "" || data.TokenType != "Bearer" {
		t.Fatalf("unexpected login data: %+v", data)
	}

	// 발급 토큰이 실제로 Verify를 통과하는지
	svc := newTestService(testTTL)
	id, err := svc.Verify(data.Token)
	if err != nil || id.Username != "admin" {
		t.Fatalf("issued token failed verification: %v", err)
	}
}

func TestLogin_WrongPassword_401InvalidCredentials(t *testing.T) {
	app := newTestApp(t, false)
	resp, env := testutil.Do(t, app, http.MethodPost, "/api/v1/auth/login",
		`{"username":"admin","password":"nope"}`, "")

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", resp.StatusCode)
	}
	if env.Error == nil || env.Error.Code != apperror.CodeInvalidCredential {
		t.Fatalf("expected INVALID_CREDENTIALS, got %+v", env.Error)
	}
}

func TestLogin_MalformedBody_400(t *testing.T) {
	app := newTestApp(t, false)
	resp, _ := testutil.Do(t, app, http.MethodPost, "/api/v1/auth/login", `{bad`, "")
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", resp.StatusCode)
	}
}

func TestRequireAuth_MissingToken_401(t *testing.T) {
	app := newTestApp(t, true)
	resp, env := testutil.Do(t, app, http.MethodGet, "/protected", "", "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", resp.StatusCode)
	}
	if env.Error == nil || env.Error.Code != apperror.CodeUnauthorized {
		t.Fatalf("expected UNAUTHORIZED, got %+v", env.Error)
	}
}

func TestRequireAuth_ValidToken_Passes(t *testing.T) {
	app := newTestApp(t, true)
	svc := newTestService(testTTL)
	token, _, err := svc.Issue(context.Background(), "admin", "admin123")
	if err != nil {
		t.Fatal(err)
	}
	resp, _ := testutil.Do(t, app, http.MethodGet, "/protected", "", token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("want 200 with valid bearer, got %d", resp.StatusCode)
	}
}
