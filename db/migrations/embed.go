// Package migrations는 버전 기반 SQL 마이그레이션 파일을 바이너리에 임베드한다.
// 디렉터리 규약: db/migrations/<driver>/NNNNNN_name.up.sql|.down.sql
package migrations

import "embed"

//go:embed postgres/*.sql mysql/*.sql
var FS embed.FS
