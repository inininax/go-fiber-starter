# AI Rules 확장 가이드

이 디렉터리는 주제별 상세 AI 규칙을 관리한다.

## 구조 원리

```
AGENTS.md            ← 모든 에이전트(Codex, opencode, Cursor…)가 항상 로드하는 코어 규칙.
                       여기엔 "언제 어떤 상세 규칙을 읽는지" 한 줄 트리거만 둔다.
docs/ai-rules/*.md   ← 주제별 상세 규칙. 해당 작업을 할 때만 읽으면 된다.
CLAUDE.md / GEMINI.md / .windsurfrules / .github/copilot-instructions.md
                     ← AGENTS.md를 가리키는 정적 포인터. 수정 불요.
```

네이티브 리더(Codex/opencode/Cursor)는 `@import`를 처리하지 못하고 AGENTS.md 본문만 읽는다.
따라서 **항상 필요한 규칙만 코어에 두고**, 나머지는 인덱스 트리거로 연결해야
토큰 효율과 규칙 누락 방지를 동시에 달성한다.

## 새 규칙 추가 절차

1. 이 디렉터리에 `kebab-case.md` 파일을 만든다 (예: `jwt-auth.md`, `deployment.md`).
2. 파일은 자기완결적이어야 한다 — 다른 규칙 문서를 읽지 않고도 수행 가능하게. 내용 중복 금지.
3. `AGENTS.md`의 "상세 규칙 인덱스"에 한 줄을 추가한다:
   `- [파일명](docs/ai-rules/파일명.md) — <트리거 조건>: <무엇을 위한 규칙인가>`
4. 규칙이 3줄 이내로 끝난다면 파일로 만들지 말고 AGENTS.md 해당 섹션에 직접 추가한다.
5. 검증: 규칙이 실제 코드/설정과 일치하는지 확인 후 커밋. 실행 소스(go.mod, Makefile, CI)와 충돌하면 실행 소스가 옳다.

## 현재 등록된 규칙

- `new-module-checklist.md` — 새 도메인 모듈 추가 시
