package migrations

import (
	"io/fs"
	"testing"
)

// TestDriverFS_ListsSQLPairs는 CI migrations 잡에서 발생한 임베드 경로 버그의
// 회귀 방지 테스트다. 드라이버 서브트리에 up/down 쌍이 모두 보여야 한다.
func TestDriverFS_ListsSQLPairs(t *testing.T) {
	for _, driver := range []string{"postgres", "mysql"} {
		t.Run(driver, func(t *testing.T) {
			sub, err := DriverFS(driver)
			if err != nil {
				t.Fatalf("sub fs: %v", err)
			}
			matches, err := fs.Glob(sub, "*.up.sql")
			if err != nil {
				t.Fatal(err)
			}
			if len(matches) == 0 {
				t.Fatalf("no *.up.sql under %s — embed path prefix drift", driver)
			}
			for _, up := range matches {
				down := trimSuffix(up, ".up.sql") + ".down.sql"
				if _, err := fs.Stat(sub, down); err != nil {
					t.Errorf("missing down migration for %s: %v", up, err)
				}
			}
		})
	}
}

func TestDriverFS_UnknownDriver_Errors(t *testing.T) {
	if _, err := DriverFS("oracle"); err == nil {
		t.Fatal("unknown driver should not resolve")
	}
}

func trimSuffix(s, suffix string) string {
	if len(s) > len(suffix) && s[len(s)-len(suffix):] == suffix {
		return s[:len(s)-len(suffix)]
	}
	return s
}
