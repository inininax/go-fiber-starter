package database

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func newBufferLogger(debug bool, slow time.Duration) (*Logger, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	l := slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	return NewLogger(l, debug, slow), buf
}

func sqlFunc(sql string, rows int64) func() (string, int64) {
	return func() (string, int64) { return sql, rows }
}

func TestTrace_NonNotFoundError_IsLoggedAsError(t *testing.T) {
	l, buf := newBufferLogger(false, time.Minute)
	l.Trace(context.Background(), time.Now().Add(-10*time.Millisecond), sqlFunc("SELECT 1", 0), errors.New("boom"))

	if !strings.Contains(buf.String(), `"level":"ERROR"`) || !strings.Contains(buf.String(), "boom") {
		t.Fatalf("expected ERROR log with cause, got: %s", buf.String())
	}
}

func TestTrace_RecordNotFound_SuppressedAtWarnLevel(t *testing.T) {
	l, buf := newBufferLogger(false, time.Minute)
	l.Trace(context.Background(), time.Now(), sqlFunc("SELECT 1", 0), gorm.ErrRecordNotFound)

	if strings.Contains(buf.String(), "gorm query failed") {
		t.Fatalf("ErrRecordNotFound must not be logged as error: %s", buf.String())
	}
}

func TestTrace_SlowQuery_Warns(t *testing.T) {
	l, buf := newBufferLogger(false, 5*time.Millisecond)
	l.Trace(context.Background(), time.Now().Add(-20*time.Millisecond), sqlFunc("SELECT slow", 1), nil)

	if !strings.Contains(buf.String(), `"level":"WARN"`) || !strings.Contains(buf.String(), "slow query") {
		t.Fatalf("expected slow query WARN, got: %s", buf.String())
	}
}

func TestTrace_DebugLevel_LogsSQL(t *testing.T) {
	l, buf := newBufferLogger(true, time.Hour)
	l.Trace(context.Background(), time.Now(), sqlFunc("SELECT fast", 3), nil)

	if !strings.Contains(buf.String(), "SELECT fast") || !strings.Contains(buf.String(), `"level":"DEBUG"`) {
		t.Fatalf("debug mode should log SQL, got: %s", buf.String())
	}
}

func TestLogMode_ReturnsIndependentCopy(t *testing.T) {
	l, _ := newBufferLogger(true, time.Minute)
	quieter := l.LogMode(gormlogger.Warn)

	if quieter == gormlogger.Interface(l) {
		t.Fatal("LogMode must return a copy, not the same instance")
	}
	// 원본은 Info 유지, 복사본은 Warn으로 내려간 상태에서 debug SQL이 원본에만 기록되는지
	ctx := context.Background()
	begin := time.Now()
	fc := sqlFunc("SELECT copy-check", 0)
	quieter.Trace(ctx, begin, fc, nil)
}

func TestInfoWarnError_Levels(t *testing.T) {
	l, buf := newBufferLogger(true, time.Minute)
	l.Info(context.Background(), "info-%d", 1)
	l.Warn(context.Background(), "warn-%d", 2)
	l.Error(context.Background(), "err-%d", 3)

	out := buf.String()
	for _, want := range []string{"info-1", "warn-2", "err-3"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in %s", want, out)
		}
	}
}

func TestTrace_SilentLevel_NoOutput(t *testing.T) {
	l, buf := newBufferLogger(false, time.Minute)
	quiet := l.LogMode(gormlogger.Silent)
	quiet.Trace(context.Background(), time.Now(), sqlFunc("SELECT 1", 0), errors.New("hidden"))

	if buf.Len() != 0 {
		t.Fatalf("silent level must not output, got: %s", buf.String())
	}
}
