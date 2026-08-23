// Package api는 저장소에 버전 관리되는 OpenAPI 스펙을 바이너리에 포함해 제공한다.
// GET /openapi.yaml 라우트(router.go)가 이 FS에서 openapi.yaml을 응답한다.
package api

import "embed"

//go:embed openapi.yaml
var FS embed.FS

// openapi.yaml이 없으면 위 지시문 때문에 컴파일 자체가 실패한다 —
// 스펙 파일과 코드의 동기화는 이 특성으로 보장된다.
