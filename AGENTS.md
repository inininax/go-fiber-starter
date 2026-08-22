# AGENTS.md

이 저장소에서 작업하는 모든 AI 에이전트(Codex, Claude, Cursor, opencode 등)의 운영 가이드.
사용법은 README.md, 개발 사양은 PROMPT.md를 참고. 이 문서는 에이전트가 실수하기 쉬운 것만 다룬다.

> **규칙 변경은 이 파일 하나만 수정한다.** CLAUDE.md / GEMINI.md / .windsurfrules /
> .github/copilot-instructions.md는 이 파일을 가리키는 정적 포인터라 수정 불요.

## 명령어

```bash
make run                # 실행(sqlite 자동 준비). make만 치면 타깃 목록
make test               # 전체 테스트(-race). 커버리지는 make test-cov
go test ./internal/modules/task/ -run TestHandler_Create -race -count=1   # 단일 테스트 패턴
gofmt -w . && go vet ./... && go build ./...   # 커밋 전 최소 검증 세트
docker compose up -d postgres                 # postgres 기동(mysql은 --profile mysql)
```

- **CI는 `gofmt -l`이 비어있어야 통과한다.** 파일 수정 후 반드시 `gofmt -w .` 포함.
- 새 마이그레이션은 `make migrate-new name=add_users`로 생성(버전 번호 자동 계산). 수동으로 파일 만들면 버전 충돌 위험.
- postgres/mysql 로컬 실행 시 환경변수 예시는 README "Quickstart"와 `.env.example` 주석 참고.

## 상세 규칙 인덱스 (필요 시 읽기)

- [docs/ai-rules/new-module-checklist.md](docs/ai-rules/new-module-checklist.md) — 새 도메인 모듈 추가 시: 복제→마이그레이션 이원화→조립→테스트 체크리스트
- 새 상세 규칙은 `docs/ai-rules/`에 kebab-case.md로 추가하고 위 인덱스에 한 줄 트리거를 등록한다 (절차: docs/ai-rules/README.md). 코어(이 파일)는 작게 유지.

## 아키텍처 불변식 (위반 시 리뷰 거부)

- 의존성 방향: `handler → service → repository → model`. 역방향 참조 금지.
- `fiber.Ctx`는 handler에만, `*gorm.DB`는 repository에만 노출.
- Repository **인터페이스는 service.go(소비자)에 선언**한다. 구현은 repository.go.
- 모듈 조립 위치: `internal/router/wiring.go`(service 생성) + `router.go`(라우트 마운트).
- 새 도메인은 `internal/modules/task/`를 템플릿으로 복제한다(model/dto/repo/service/handler/routes).

## Fiber v3 주의사항 — v2 문법은 컴파일 오류 또는 버그 유발

- 핸들러 시그니처: `func(c fiber.Ctx) error` (포인터 `*fiber.Ctx` 아님).
- 바디 파싱: `c.BodyParser()` 없음 → `c.Bind().Body(&dto)`.
- 요청 ID: `requestid.FromContext(c)` 사용(Locals 문자열 키 아님).
- 서버 시작: `app.Listen(addr, fiber.ListenConfig{GracefulContext: ctx, ShutdownTimeout: ...})`.
- ReadTimeout/WriteTimeout/IdleTimeout/BodyLimit은 **fiber.Config**에 있다(ListenConfig 아님).
- validate 태그는 자동 실행이 아니라 `fiber.Config.StructValidator` 등록 필요 → `internal/validator` 패키지 사용.
- 검증 실패 매핑은 `task/handler.go`의 `bindErrorToAppError`(400 파싱 / 422 필드별 details) 패턴을 따른다.

## 마이그레이션 이원화 정책 (혼동 금지)

- sqlite(dev 전용): GORM AutoMigrate. prod(pg/mysql)+`DB_AUTO_MIGRATE=true`는 시작 시 하드 차단됨(`database.AutoMigrateIfNeeded`).
- postgres/mysql(prod): `db/migrations/{postgres,mysql}/NNNNNN_name.{up,down}.sql` + `cmd/migrate up|down|version|force`.
- 새 모델 추가 시 두 곳 모두 처리: (1) `cmd/api/main.go`의 `AutoMigrateIfNeeded(...)` 목록에 추가(sqlite용),
  (2) 드라이버별 SQL 파일 작성(pg/mysql용). 한쪽만 하면 환경에 따라 스키마 누락.

## 응답/에러 규약

- 모든 응답은 `common.Envelope`({success, data|error, meta}). 핸들러에서 직접 JSON 작성 금지 — `common.OK/Created/NoContent` 사용.
- 에러는 `common.AppError` 코드 카탈로그(errors.go)로 반환하고 `fmt.Errorf("...: %w")` 래핑.
- 전역 ErrorHandler(router.go): 알 수 없는 에러는 로그에 상세 남기고 클라이언트엔 고정 메시지 500. 내부 메시지 노출 금지.

## 설정

- 환경변수 단일 출처(internal/config). `.env.example`과 config.go가 함께 변경되어야 한다.
- `Validate()`는 문제를 전부 수집해 한 번에 보고한다(fail-fast). 새 설정 추가 시 규칙도 반드시 추가.

## 테스트 관례

- service 단위 테스트: `fakeRepo` 주입(외부 I/O 없음). handler 통합 테스트: 임시 디렉터리 sqlite + 실제 GORM + `app.Test()`.
- 테스트 헬퍼 `do()`는 204 등 빈 본문을 건너뛴다 — 빈 본문 디코딩으로 인한 EOF 재발 방지.
- sqlite `:memory:`를 쓰지 않는다(풀 연결마다 별도 DB가 되어 플레이크). 임시 파일 경로를 쓸 것.

## 기타 컨벤션

- 주석은 "왜"만 설명. 자명한 코드 주석 금지.
- context는 `c.Context()`로 꺼내 service/repository까지 전파. 버림 금지.
- 페이지네이션 limit은 `common.MaxLimit`(100)으로 클램프 — count 비용 방어.
