## 요약


## 변경 유형

- [ ] feat
- [ ] fix
- [ ] refactor
- [ ] docs
- [ ] chore

## 검증 체크리스트 (커밋 전 필수)

- [ ] `gofmt -l .` 빈 출력
- [ ] `go vet ./...` 통과
- [ ] `go test ./... -race -count=1` 통과
- [ ] 마이그레이션 변경 시 sqlite(AutoMigrate 목록) + pg/mysql SQL 쌍 모두 작성
- [ ] API 계약 변경 시 `api/openapi.yaml` 동기화(드리프트 테스트 통과로 확인됨)
- [ ] 사용자 관점 변경 시 CHANGELOG.md Unreleased에 한 줄 추가
