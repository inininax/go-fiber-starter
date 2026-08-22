package task

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"go-fiber-starter/internal/common"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := filepath.Join(t.TempDir(), "test.db")
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&Task{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

// dryRunFirst는 DryRun 세션으로 First 쿼리의 SQL을 생성해 실행 없이 검사한다.
// Update는 트랜잭션 안에서만 의미가 있어 SQL 생성 단계로 잠금 절을 검증한다.
func dryRunFirst(t *testing.T, db *gorm.DB) string {
	t.Helper()
	var tsk Task
	stmt := applyRowLock(db.Session(&gorm.Session{DryRun: true})).First(&tsk, 1)
	if stmt.Error != nil {
		t.Fatalf("build query: %v", stmt.Error)
	}
	return stmt.Statement.SQL.String()
}

func TestRowLock_SQLite_OmitsRowLock(t *testing.T) {
	db := newTestDB(t)

	sql := dryRunFirst(t, db)
	if strings.Contains(strings.ToUpper(sql), "FOR UPDATE") {
		t.Fatalf("sqlite must not generate FOR UPDATE: %s", sql)
	}
}

func TestRowLock_Postgres_GeneratesRowLock(t *testing.T) {
	db, err := gorm.Open(gormpostgres.New(gormpostgres.Config{
		DSN: "postgres://ci:ci@localhost:5432/ci_migrate?sslmode=disable",
	}), &gorm.Config{DisableAutomaticPing: true})
	if err != nil {
		t.Fatalf("open postgres dialector: %v", err)
	}

	sql := dryRunFirst(t, db)
	if !strings.Contains(strings.ToUpper(sql), "FOR UPDATE") {
		t.Fatalf("postgres must generate FOR UPDATE: %s", sql)
	}
}

func TestUpdate_NotFound_ReturnsErrNotFound(t *testing.T) {
	repo := NewRepository(newTestDB(t))

	_, err := repo.Update(context.Background(), 999, func(*Task) {})
	if !errors.Is(err, common.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}
