package task

import (
	"errors"
	"testing"

	"github.com/gofiber/fiber/v3"

	"go-fiber-starter/internal/apperror"
)

// FuzzParseID는 경로 파라미터 파서가 임의 입력에서 패닉 없이
// "성공 ⟺ 양의 정수 문자열" 계약을 유지하는지 검증한다.
func FuzzParseID(f *testing.F) {
	f.Add("1")
	f.Add("999999999999999")
	f.Add("0")  // 0은 거부 대상
	f.Add("-1") // 부호 거부
	f.Add("+5") // 부호 허용 여부와 무관하게 패닉만 금지
	f.Add("abc")
	f.Add("")
	f.Add("99999999999999999999999") // uint64 오버플로
	f.Add("１")                       // 전각 숫자(비 ASCII)

	f.Fuzz(func(t *testing.T, raw string) {
		id, err := parseID(raw)
		if err != nil {
			var appErr *apperror.AppError
			if !errors.As(err, &appErr) || appErr.Status != fiber.StatusBadRequest {
				t.Fatalf("failure must be AppError/400, got %v", err)
			}
			return
		}
		if id == 0 {
			t.Fatal("zero id must never succeed")
		}
	})
}
