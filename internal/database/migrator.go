package database

import (
	"context"
	"fmt"
	"log/slog"

	"gorm.io/gorm"

	"go-fiber-starter/internal/config"
)

// AutoMigrateIfNeeded는 스키마 동기화 정책을 집행한다.
//
//   - sqlite(dev 전용): GORM AutoMigrate 허용. 개발 사이클 속도 우선.
//   - postgres/mysql(prod): 버전 기반 마이그레이션(cmd/migrate)만 허용.
//     DB_AUTO_MIGRATE=true 조합은 하드 차단한다(컬럼 삭제 불가·락·다운타임 위험).
//   - postgres/mysql(dev): DB_AUTO_MIGRATE=true면 허용하되 경고를 남긴다.
//
// 모델 목록은 조립 계층(cmd/api)에서 주입한다 — database 패키지는 도메인을 모른다.
func AutoMigrateIfNeeded(ctx context.Context, cfg *config.Config, db *gorm.DB, models ...any) error {
	if !cfg.DBAutoMigrate || len(models) == 0 {
		return nil
	}
	if cfg.DBDriver != config.DBDriverSQLite {
		if cfg.IsProd() {
			return fmt.Errorf("DB_AUTO_MIGRATE is not allowed in prod with %s; use `migrate up` (see db/migrations)", cfg.DBDriver)
		}
		slog.WarnContext(ctx, "DB_AUTO_MIGRATE enabled for non-sqlite driver in dev; use versioned migrations for real data",
			"driver", cfg.DBDriver)
	}
	return db.WithContext(ctx).AutoMigrate(models...)
}
