# AI rules 디렉터리 (확장 포인트)

이 디렉터리는 루트 `AGENTS.md` 이외의 주제별 상세 AI 규칙을 관리한다.
OpenCode는 `opencode.json`의 glob(`.agents/rules/**/*.md`)으로 이곳의 모든 `*.md`를
자동 로드한다. 그 외 도구는 `AGENTS.md`(및 그 심링크)를 통해 이 파일들을 읽는다.

## 도구 와이어링 맵

| 도구 | 읽는 것 | 메커니즘 |
|---|---|---|
| Codex CLI / OpenCode / 신버전 Cursor | `AGENTS.md` | 네이티브 자동 로드 |
| OpenCode (이 디렉터리) | `.agents/rules/**/*.md` | `opencode.json`의 `instructions` glob |
| Claude Code | `CLAUDE.md` → `AGENTS.md` | 심링크; AGENTS.md 안의 `@import` 줄 |
| Gemini CLI | `GEMINI.md` → `AGENTS.md` | 심링크 |
| GitHub Copilot | `.github/copilot-instructions.md` → `AGENTS.md` | 심링크 |
| Windsurf | `.windsurfrules` → `AGENTS.md` | 심링크 |
| Cursor | `.cursor/rules/project.mdc` | 요약본(mdc frontmatter 필요, 심링크 불가) |

## 새 규칙 추가 절차

1. 이 디렉터리에 `<topic>.md`를 만든다(예: `jwt-auth.md`, `deployment.md`).
   짧고, 사실적이고, 코드와 대조 검증된 내용만. 자기완결적일 것.
2. 루트 `AGENTS.md` 하단 "상세 규칙" 목록에 한 줄 추가:
   `- @.agents/rules/<topic>.md`
   - Claude Code는 `@import`를 해석하고, 다른 에이전트는 읽기 좋은 포인터로 본다.
3. 빠른 맥락으로 중요한 규칙이라면 `.cursor/rules/project.mdc`에 한 문장 요약을
   미러링한다.

심링크 경유로 파일을 편집하지 말 것. 항상 원본(`AGENTS.md` 또는 이 디렉터리)을 편집.

## 등록된 규칙

- [new-module-checklist.md](new-module-checklist.md) — 새 도메인 모듈 추가 시 수행 체크리스트
