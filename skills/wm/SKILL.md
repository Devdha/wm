---
name: wm-git-worktree-manager
description: Use when working with git worktrees - creating isolated workspaces for branches, managing multiple features in parallel, or cleaning up worktrees
---

# WM - Git Worktree Manager

git worktree를 쉽게 관리하는 CLI 도구.

## 설치

```bash
# npm
npm install -g gitwm

# Go
go install github.com/Devdha/wm@latest
```

## Quick Reference

| 명령어 | 설명 |
|--------|------|
| `wm init` | 프로젝트 설정 (.wm.yaml 생성) |
| `wm add <branch>` | worktree 생성 |
| `wm list` / `wm ls` | 모든 worktree 조회 |
| `wm remove <path>` | worktree 삭제 |
| `wm remove -b <path>` | worktree + 브랜치 삭제 |

## 핵심 패턴

### Worktree 생성

```bash
# 기본: ../wm_{repo}/{branch} 에 생성
wm add feature-login

# 슬래시 브랜치: feature/auth → feature-auth 폴더로 생성
wm add feature/auth
# 결과: ../wm_myrepo/feature-auth/

# 커스텀 경로
wm add feature-login -p ./workspaces/login
```

### Worktree 삭제

```bash
# 경로로 삭제
wm remove ../wm_myrepo/feature-auth

# 브랜치 이름으로 삭제 (둘 다 동작)
wm remove feature/auth
wm remove feature-auth

# 브랜치도 함께 삭제
wm remove -b feature/auth

# 강제 삭제 (확인 스킵)
wm remove -f feature/auth
```

## 설정 (.wm.yaml)

```yaml
version: 1

worktree:
  base_dir: "../wm_{repo}"  # {repo}는 레포 이름으로 치환

sync:
  - ".env"                  # worktree에 복사
  - "apps/*/.env"           # glob 지원
  - src: ".env.example"
    dst: ".env"
    mode: copy              # 또는 "symlink"
    when: missing           # 또는 "always"

tasks:
  post_install:
    mode: background        # 비동기 실행
    commands:
      - "pnpm install"
```

## When to Use

- 여러 기능을 병렬로 개발할 때
- PR 리뷰하면서 다른 작업할 때
- 긴 빌드/테스트 중에 다른 브랜치 작업할 때
- 브랜치 간 빠른 전환이 필요할 때

## Common Mistakes

| 실수 | 해결 |
|------|------|
| `feature/auth` 브랜치가 중첩 폴더 생성 | v0.1.1+에서 자동으로 `feature-auth` 폴더 생성 |
| main worktree 삭제 시도 | 자동 방지됨 |
| 다른 worktree에서 사용 중인 브랜치 삭제 | 경고 후 확인 |
