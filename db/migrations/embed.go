// Package migrations는 버전 기반 SQL 마이그레이션 파일을 바이너리에 임베드한다.
// 디렉터리 규약: db/migrations/<driver>/NNNNNN_name.up.sql|.down.sql
package migrations

import (
	"io/fs"

	"embed"
)

//go:embed postgres/*.sql mysql/*.sql
var FS embed.FS

// DriverFS는 드라이버명(postgres|mysql)에 해당하는 마이그레이션 서브트리를 반환한다.
//
// 주의: //go:embed는 패키지 디렉터리 기준 상대경로로 임베드하므로 FS 내부 경로에는
// "db/migrations/" 접두어가 없다. 상위 접두어로 fs.Sub를 자르면 빈 트리가 되어
// "open .: file does not exist" 오류가 발생한다(CI migrations 잡에서 실측된 버그).
// fs.Sub는 지연 바인딩이라 미존재 디렉터리도 통과하므로 드라이버를 먼저 검증한다.
func DriverFS(driver string) (fs.FS, error) {
	switch driver {
	case "postgres", "mysql":
	default:
		return nil, fs.ErrNotExist
	}
	return fs.Sub(FS, driver)
}
