// migrate는 버전 기반 SQL 마이그레이션 CLI다 (postgres/mysql 전용).
// sqlite 개발 환경은 AutoMigrate를 사용하므로 이 도구의 대상이 아니다.
//
// 사용법:
//
//	go run ./cmd/migrate up            # 미적용 마이그레이션 전부 적용
//	go run ./cmd/migrate down          # 직전 버전으로 롤백(1스텝, 실수 방지 기본값)
//	go run ./cmd/migrate version       # 현재 버전 조회
//	go run ./cmd/migrate force <ver>   # 더티 상태 수동 복구
package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"

	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	migrations "go-fiber-starter/db/migrations"
	"go-fiber-starter/internal/config"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		slog.Error("migrate failed", "error", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if cfg.DBDriver == config.DBDriverSQLite {
		return fmt.Errorf("sqlite does not use SQL migrations; DB_AUTO_MIGRATE=true handles schema in dev")
	}

	sub, err := migrations.DriverFS(string(cfg.DBDriver))
	if err != nil {
		return fmt.Errorf("locate migrations for %s: %w", cfg.DBDriver, err)
	}
	src, err := iofs.New(sub, ".")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	// golang-migrate는 databaseURL의 스킴(postgres://|mysql://)으로 드라이버를 식별한다.
	murl, err := migrateURL(cfg)
	if err != nil {
		return err
	}
	m, err := migrate.NewWithSourceInstance("iofs", src, murl)
	if err != nil {
		return fmt.Errorf("init migrator: %w", err)
	}
	defer func() {
		srcErr, dbErr := m.Close() // Close는 소스/DB 각각의 에러를 반환한다
		if srcErr != nil || dbErr != nil {
			fmt.Fprintf(os.Stderr, "close migrator: source=%v database=%v\n", srcErr, dbErr)
		}
	}()

	if len(args) == 0 {
		return fmt.Errorf("usage: migrate up|down|version|force <n>")
	}

	switch args[0] {
	case "up":
		err = m.Up()
	case "down":
		err = m.Steps(-1) // 1스텝만 롤백
	case "version":
		v, dirty, verr := m.Version()
		if verr != nil {
			return verr
		}
		fmt.Printf("version=%d dirty=%v\n", v, dirty)
		return nil
	case "force":
		if len(args) < 2 {
			return fmt.Errorf("usage: migrate force <version>")
		}
		v, perr := strconv.Atoi(args[1])
		if perr != nil {
			return fmt.Errorf("invalid version %q: %w", args[1], perr)
		}
		err = m.Force(v)
	default:
		return fmt.Errorf("unknown command %q; usage: migrate up|down|version|force <n>", args[0])
	}

	if err != nil && !isNoChange(err) {
		return err
	}
	fmt.Println("done:", args[0])
	return nil
}

// migrateURL은 설정된 DSN을 golang-migrate URL 규약으로 정규화한다.
//   - postgres: postgres://user:pass@host:5432/db
//   - mysql:    mysql://user:pass@tcp(host:3306)/db  (go-sql-driver 형식 유지)
//
// DSN에 스킴이 이미 있으면 드라이버명으로 치환하고, 없으면 접두어로 붙인다.
// key-value DSN(host=... user=...)은 URL 파싱이 불가하므로 명확한 에러로 안내한다.
func migrateURL(cfg *config.Config) (string, error) {
	dsn := cfg.DBDSN
	if strings.Contains(dsn, "=") && !strings.Contains(dsn, "://") && cfg.DBDriver == config.DBDriverPostgres {
		return "", fmt.Errorf(
			"DB_DSN is in key-value form which golang-migrate cannot parse; use URL form: postgres://user:pass@host:5432/db")
	}
	return string(cfg.DBDriver) + "://" + stripScheme(dsn), nil
}

func isNoChange(err error) bool {
	return err == migrate.ErrNoChange
}

func stripScheme(dsn string) string {
	for _, prefix := range []string{"postgres://", "postgresql://", "mysql://"} {
		if len(dsn) > len(prefix) && dsn[:len(prefix)] == prefix {
			return dsn[len(prefix):]
		}
	}
	return dsn
}
