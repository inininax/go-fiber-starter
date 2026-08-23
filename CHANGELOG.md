# Changelog

모든 주목할 만한 변경은 이 파일에 기록된다.
형식은 [Keep a Changelog](https://keepachangelog.com/ko/1.1.0/)를 따르고, 버전은 [Semantic Versioning](https://semver.org/lang/ko/)을 준수한다.

## [Unreleased]

### Changed

- GitHub Actions 의존성 갱신: checkout@v7, setup-go@v7, golangci-lint-action@v9 (dependabot)

## [0.3.0] - 2026-08-23

### Added

- `.github/dependabot.yml` — go modules + github-actions 주간 갱신(minor/patch 그룹핑)
- OpenAPI 3.0.3 스펙(`api/openapi.yaml`) + `GET /openapi.yaml` 서빙 + 라우트 드리프트 검증 테스트
- livez 응답에 빌드 commit 노출(`Config.BuildCommit`, `dev` 폴백)
- CI 커버리지 게이트(`-coverpkg=./...` 계량, COVERAGE_MIN=65 fail-under)
- 전역 rate limiter의 probe/metrics/spec 경로 skip(exact 매칭)

## [0.2.0] - 2026-08-23

### Added

- JWT 인증 스캐폴드: `POST /api/v1/auth/login`, HS256 발급/검증, RequireAuth 가드
- 관찰 가능성: 에러 실횅 상태 메트릭화(EffectiveStatus), request_id 상관관계, Prometheus 수집
- 보안: prod 데모 크리덴셜 차단 + constant-time 비교, prod CORS `*` 거부, TRUST_PROXY,
  로그인 전용 rate limiter, task UPDATE 행잠금(sqlite 예외), SQLite 디렉터리 0o700
- CI: postgres/mysql 마이그레이션 검증 잡, golangci-lint v2 연동
- AI 규칙 구조: AGENTS.md 단일 출처 + 심링크 포인터 + `.agents/rules/`(hanppyeom-ttang 패턴)

### Fixed

- 에러 응답이 메트릭·액세스 로그에 200으로 기록되던 결함
- 액세스 로그 request_id 누락(fiber v3 requestid API 전환)
- 마이그레이션 임베드 경로 버그(CI migrations 잡 실패)

## [0.1.0] - 2026-08-23

### Added

- 초기 스캐폴드: Fiber v3 + GORM 레이어드 아키텍처(handler→service→repository)
- task 예제 CRUD 모듈(모듈 템플릿) + health(livez/readyz)
- 통일 응답 엔벨로프/에러 카탈로그, fail-fast 설정 검증, graceful shutdown
- 마이그레이션 이원화: sqlite AutoMigrate(dev) / postgres·mysql 버전 SQL(`cmd/migrate`)
- 멀티 에이전트 협업 문서 체계(AGENTS.md 단일 출처)
