# 아키텍처

이 문서는 요청 생명주기, 계층 경계, 스키마 정책의 "왜"를 한 장으로 설명한다.
상세 규칙은 [AGENTS.md](../AGENTS.md)와 `.agents/rules/`를 참고.

## 요청 생명주기

```
클라이언트
  │
  ▼
[requestid]      상관관계 ID 발급(클라이언트 X-Request-ID는 sanitize 후 수용)
[recover]        패닉을 에러로 변환해 ErrorHandler로 전달
[helmet]         보안 헤더
[cors]           설정 기반 오리진 허용(prod에서 '*' 거부)
[limiter 전역]   IP당 분당 예산 — /livez /readyz /metrics /openapi.yaml은 skip
[prometheus]     method×route패턴×status 카운터/히스토그램(EffectiveStatus 기준)
[requestlogger]  실횅 상태 기반 구조화 액세스 로그(request_id 포함)
  │
  ▼
router (/api/v1 그룹)
  ├─ auth: POST /auth/login (로그인 전용 엄격 limiter)
  └─ task: CRUD (AUTH_ENABLED=true면 RequireAuth 가드)
       │ fiber.Ctx는 여기서 끝난다
       ▼
   handler   Bind().Body() → StructValidator(validate 태그) → DTO
       │        BindErrorToAppError: 파싱=400 / 검증 위반=422+details
       ▼
   service   트랜잭션 경계. Repository 인터페이스 선언(소비자 정의).
       │        에러는 apperror 센티널 + %w 래핑
       ▼
 repository  *gorm.DB 유일 소유 계층. WithContext(ctx) 필수.
             gorm.ErrRecordNotFound → apperror.ErrNotFound 변환
```

### 관찰 정합성 (핵심 설계)

Fiber v3는 미들웨어 언와인드 이후에 ErrorHandler를 실행하므로, 언와인드 시점의
`c.Response().StatusCode()`는 핸들러가 에러를 반환한 경우 200이다.
그래서 두 관측 미들웨어 모두 `EffectiveStatus(c, err)`로 **실횅 상태**를 산정한다:

- `*fiber.Error` → fe.Code
- `apperror.AppError` → Status(단, ≥500은 errorHandler가 고정 500으로 응답하므로 동일하게 클램프)
- 그 외 → 500

규칙이 router.errorHandler와 어긋나면 지표·로그·응답 3자가 불일치한다 — 양쪽은 함께 수정한다.

### 429 흐름

limiter의 `LimitReached`는 `ErrRateLimited`(AppError)를 반환한다 → ErrorHandler가
엔벨로프(`RATE_LIMITED`)로 응답하고 warn 로그를 남긴다. `EffectiveStatus`도 같은 429를
보고해 3자 일치가 유지된다.

## 계층 경계 (위반 금지)

| 계층 | 가질 수 있는 것 | 절대 금지 |
|---|---|---|
| handler | fiber.Ctx, DTO 조립 | 비즈니스 로직, gorm |
| service | Repository 인터페이스 선언, 트랜잭션 경계 | fiber.Ctx |
| repository | *gorm.DB, SQL | HTTP 개념 |
| model | GORM 태그 | JSON 직렬화 책임(DTO가 담당) |

에러 방향은 하위→상위로 `AppError` 센티널만 올려보내고, HTTP 변환은 ErrorHandler 단 한 곳에서 한다.

## 스키마 정책 (이원화)

| 드라이버 | 방식 | 이유 |
|---|---|---|
| sqlite(dev/test) | GORM AutoMigrate | 사이클 속도. `cmd/api/main.go` 모델 목록에 등록 |
| postgres/mysql(prod) | 버전 SQL(`db/migrations/<driver>/`) + `cmd/migrate` | 컬럼 삭제/락 제어. AutoMigrate는 prod에서 하드 차단 |

양쪽을 모두 갱신하지 않으면 환경에 따라 스키마 누락 → readyz나 스모크에서 발각된다.

## 인증

HS256 JWT(만료 필수, 알고리즘 강제). 자격증명 검증은 `Authenticator` 인터페이스 뒤에
숨겨져 있어 env 데모 구현 ↔ DB 구현 교체가 자유롭다. 교체 스케치는
[docs/authenticator-sketch.md](authenticator-sketch.md).

## 의도적 보류 (과잉설계 방지)

compress/ETag, OTel, cursor pagination, 분산 rate limiter — 트리거 조건은 README Roadmap 참고.
