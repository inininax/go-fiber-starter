package testutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
)

// APIError는 응답 엔벨로프 error 블록의 테스트용 디코딩 타깃이다.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// APIEnvelope는 응답 엔벨로프의 테스트용 디코딩 타깃이다.
// Data/Meta는 RawMessage로 받아 호출자가 필요한 타입으로 언마셜한다.
type APIEnvelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Meta    json.RawMessage `json:"meta"`
	Error   *APIError       `json:"error"`
}

// Do는 테스트 요청을 전송하고 응답과 엔벨로프를 반환한다.
// 204 등 본문 없는 응답은 디코딩을 건너뛴다(빈 본문 디코딩 EOF 재발 방지).
func Do(t *testing.T, app *fiber.App, method, target, body, bearer string) (*http.Response, APIEnvelope) {
	t.Helper()
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, target, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, target, nil)
	}
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, target, err)
	}
	var env APIEnvelope
	if resp.StatusCode != http.StatusNoContent && resp.ContentLength != 0 {
		if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
			t.Fatalf("decode response: %v", err)
		}
	}
	return resp, env
}
