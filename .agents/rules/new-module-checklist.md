# 새 모듈 추가 체크리스트

트리거: `internal/modules/` 아래에 새 도메인을 추가할 때.
`task` 모듈이 유일한 공식 템플릿이다. 이 체크리스트를 순서대로 수행한다.

## 1. 패키지 생성 (task 복제)

- [ ] `internal/modules/<name>/` 에 model.go / dto.go / repository.go / service.go / handler.go / routes.go 생성
- [ ] Repository **인터페이스는 service.go(소비자)에 선언**, GORM 구현은 repository.go
- [ ] DTO ↔ 모델 매핑은 명시적 함수로 (toResponse 등). 모델 직접 노출 금지
- [ ] fiber.Ctx는 handler에만, *gorm.DB는 repository에만

## 2. 검증/에러 연결

- [ ] 요청 DTO에 `validate:"..."` 태그 (StructValidator가 자동 실행)
- [ ] 도메인 에러는 common.errors.go 카탈로그에 코드+상태 등록 후 service에서 반환
- [ ] Bind 실패 매핑은 task/handler.go의 bindErrorToAppError 패턴 재사용

## 3. 마이그레이션 이원화 (놓치기 쉬움 — 둘 다!)

- [ ] sqlite(dev): `cmd/api/main.go`의 `AutoMigrateIfNeeded(...)` 목록에 모델 추가
- [ ] postgres/mysql(prod): `make migrate-new name=<desc>`로 SQL 파일 생성 후 양쪽 드라이버 작성(up/down)
- [ ] 한쪽만 하면 환경에 따라 스키마 누락 → 시작 시 readyz 실패로 발견됨

## 4. 조립

- [ ] `internal/router/wiring.go`: `<name>NewService(<name>NewRepository(db))`
- [ ] `internal/router/router.go`: `<name>.RegisterRoutes(v1, svc)`
- [ ] 인증 보호가 필요하면 guard 전달 방법은 auth/routes.go와 router.go의 AUTH_ENABLED 분기 참고

## 5. 테스트 (커밋 전 필수)

- [ ] service 단위 테스트: fake repository 주입, 외부 I/O 없음
- [ ] handler 통합 테스트: 임시 디렉터리 sqlite + 실제 GORM + app.Test()
  - 성공/검증실패(422)/미존재(404) 시나리오 최소 3개
  - 응답 디코딩은 빈 본문(204 등) 건너뛰는 do() 헬퍼 패턴 사용
- [ ] `gofmt -w . && go vet ./... && go test ./... -race` 통과

## 6. 문서

- [ ] README의 API 예제 섹션에 curl 추가
