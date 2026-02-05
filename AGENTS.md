# WM - Git Worktree Manager

이 프로젝트는 git worktree를 쉽게 관리하는 CLI 도구입니다.

## 개발 명령어

```bash
# 빌드
go build ./...

# 테스트
go test ./...

# 로컬 실행
go run . <command>
```

## 릴리스

```bash
# 태그 생성 및 푸시 (GitHub Actions가 GoReleaser 실행)
git tag v0.x.x
git push origin v0.x.x

# npm 배포 (2FA 계정은 로컬에서 실행)
./scripts/publish-npm-local.sh 0.x.x
```

## 프로젝트 구조

- `cmd/` - CLI 명령어 (cobra)
- `internal/workspace/` - 핵심 worktree 관리 로직
- `internal/git/` - git 명령어 래퍼
- `internal/config/` - .wm.yaml 설정 파싱
- `internal/sync/` - 파일 동기화
- `internal/detect/` - 패키지 매니저 감지 (npm, pnpm, poetry 등)

## 주요 파일

- `internal/workspace/workspace.go` - worktree 생성/삭제 핵심 로직
- `internal/workspace/workspace.go:113` - `sanitizeBranchName()` 슬래시→하이픈 변환
