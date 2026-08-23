package main

import (
	"strings"
	"testing"

	"go-fiber-starter/internal/config"
)

func TestMigrateURL_PostgresURLForm(t *testing.T) {
	cfg := &config.Config{DBDriver: config.DBDriverPostgres, DBDSN: "postgres://u:p@h/db?sslmode=disable"}
	got, err := migrateURL(cfg)
	if err != nil {
		t.Fatal(err)
	}
	want := "postgres://u:p@h/db?sslmode=disable"
	if got != want {
		t.Fatalf("want %q, got %q", want, got)
	}
}

func TestMigrateURL_MysqlTCPForm(t *testing.T) {
	cfg := &config.Config{DBDriver: config.DBDriverMySQL, DBDSN: "u:p@tcp(h:3306)/db?parseTime=true"}
	got, err := migrateURL(cfg)
	if err != nil {
		t.Fatal(err)
	}
	want := "mysql://u:p@tcp(h:3306)/db?parseTime=true"
	if got != want {
		t.Fatalf("want %q, got %q", want, got)
	}
}

func TestMigrateURL_KeyValueDSN_Rejected(t *testing.T) {
	cfg := &config.Config{DBDriver: config.DBDriverPostgres, DBDSN: "host=localhost user=u dbname=d"}
	if _, err := migrateURL(cfg); err == nil || !strings.Contains(err.Error(), "URL form") {
		t.Fatalf("key-value DSN must be rejected with guidance, got %v", err)
	}
}

// FuzzStripScheme은 임의 DSN 입력에 대해 스킴 제거가 패닉 없이 다음 불변식을
// 유지하는지 검증한다: 알려진 접두어로 시작하면 정확히 한 번 제거, 아니면 원본 유지.
func FuzzStripScheme(f *testing.F) {
	f.Add("postgres://u:p@h/db")
	f.Add("postgresql://u:p@h/db")
	f.Add("mysql://u@tcp(h)/db")
	f.Add("mysql://mysql://double")
	f.Add("plain-dsn")
	f.Add("")
	f.Add("postgres:/single-slash")

	f.Fuzz(func(t *testing.T, dsn string) {
		out := stripScheme(dsn)

		for _, prefix := range []string{"postgres://", "postgresql://", "mysql://"} {
			if !strings.HasPrefix(dsn, prefix) {
				continue
			}
			// 접두어가 입력 전체와 동일한 경우(빈 DSN)는 길이 가드로 미제거가 계약이다.
			if len(dsn) > len(prefix) && out != dsn[len(prefix):] {
				t.Fatalf("prefix %q not stripped exactly once: in=%q out=%q", prefix, dsn, out)
			}
			if len(dsn) == len(prefix) && out != dsn {
				t.Fatalf("prefix-only input must stay unchanged: in=%q out=%q", dsn, out)
			}
			return
		}
		if out != dsn {
			t.Fatalf("no known prefix but output changed: in=%q out=%q", dsn, out)
		}
	})
}
