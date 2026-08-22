# GitHub Copilot 지침

이 저장소의 에이전트 운영 규칙은 루트의 **AGENTS.md**가 단일 출처이다. 반드시 먼저 읽을 것.

핵심 요약 (상세는 AGENTS.md):

- 스택: Go + Fiber v3 + GORM. Fiber v2 문법 사용 금지 (`c.Bind().Body()`, `func(c fiber.Ctx) error`).
- 계층: handler → service → repository → model. `*gorm.DB`는 repository에만.
- 새 도메인은 `internal/modules/task/`를 템플릿으로 복제. 조립은 `internal/router/wiring.go`와 `router.go`.
- 마이그레이션 이원화: sqlite=AutoMigrate(dev), postgres/mysql prod=`cmd/migrate` SQL 버전 관리.
- 응답/에러는 `common.Envelope`/`common.AppError` 카탈로그로 통일.
- 커밋 전 검증: `gofmt -w . && go vet ./... && go test ./... -race`.
