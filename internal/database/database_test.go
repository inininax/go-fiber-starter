package database

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"go-fiber-starter/internal/config"
)

func TestSQLite_CreatesPrivateDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission bits are not applicable on Windows")
	}

	root := t.TempDir()
	cfg := &config.Config{
		DBDriver: config.DBDriverSQLite,
		DBDSN:    filepath.Join(root, "nested", "app.db"),
	}
	db, err := New(context.Background(), cfg, NewLogger(slog.Default(), false, time.Minute))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})

	info, err := os.Stat(filepath.Join(root, "nested"))
	if err != nil {
		t.Fatalf("stat created directory: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("expected a directory")
	}
	if got := info.Mode().Perm(); got != 0o700 {
		t.Fatalf("want dir perm 0700, got %o", got)
	}
}
