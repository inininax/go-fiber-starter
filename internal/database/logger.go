package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Logger는 GORM 로그를 slog로 전달하는 어댑터다.
type Logger struct {
	slog               *slog.Logger
	level              gormlogger.LogLevel
	slowQueryThreshold time.Duration
}

func NewLogger(l *slog.Logger, debug bool, slowQueryThreshold time.Duration) *Logger {
	level := gormlogger.Warn
	if debug {
		level = gormlogger.Info // SQL/파라미터까지 출력
	}
	return &Logger{slog: l, level: level, slowQueryThreshold: slowQueryThreshold}
}

func (l *Logger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	cp := *l
	cp.level = level
	return &cp
}

func (l *Logger) Info(ctx context.Context, msg string, args ...any) {
	l.slog.InfoContext(ctx, fmt.Sprintf(msg, args...))
}

func (l *Logger) Warn(ctx context.Context, msg string, args ...any) {
	l.slog.WarnContext(ctx, fmt.Sprintf(msg, args...))
}

func (l *Logger) Error(ctx context.Context, msg string, args ...any) {
	l.slog.ErrorContext(ctx, fmt.Sprintf(msg, args...))
}

func (l *Logger) Trace(_ context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.level <= gormlogger.Silent {
		return
	}
	elapsed := time.Since(begin)
	sql, rows := fc()

	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		l.slog.Error("gorm query failed", "error", err, "sql", sql, "rows", rows, "elapsed_ms", elapsed.Milliseconds())
	case l.slowQueryThreshold > 0 && elapsed > l.slowQueryThreshold:
		l.slog.Warn("slow query", "sql", sql, "rows", rows, "elapsed_ms", elapsed.Milliseconds())
	case l.level >= gormlogger.Info:
		l.slog.Debug("query", "sql", sql, "rows", rows, "elapsed_ms", elapsed.Milliseconds())
	}
}
