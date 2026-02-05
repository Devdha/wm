# WM - Git Worktree Manager

A CLI tool that makes git worktree easier to use with file sync and background tasks.

## Installation

### npm / npx
```bash
# Run directly
npx gitwm --help

# Or install globally
npm install -g gitwm
wm --help
```

### Go
```bash
go install github.com/Devdha/wm@latest
```

### Binary
Download from [GitHub Releases](https://github.com/Devdha/wm/releases)

## Quick Start

```bash
# Initialize in your project
wm init

# Create a worktree for a feature branch
wm add feature-login

# List all worktrees
wm list

# Remove a worktree
wm remove ../wm_myrepo/feature-login

# Remove worktree and delete branch
wm remove -b ../wm_myrepo/feature-login
```

## Configuration

WM uses a `.wm.yaml` file in your project root:

```yaml
version: 1

worktree:
  base_dir: "../wm_{repo}"  # {repo} is replaced with repo name

sync:
  - ".env"                              # Copy .env to worktree
  - "apps/*/.env"                       # Glob patterns supported
  - src: ".env.example"
    dst: ".env"
    mode: copy                          # or "symlink"
    when: missing                       # or "always"

tasks:
  post_install:
    mode: background                    # Run async
    commands:
      - "pnpm install"
```

## Commands

### `wm init`

Interactive setup to create `.wm.yaml`.

### `wm add <branch>`

Create a new worktree. Options:
- `--path, -p`: Custom worktree path

### `wm list`

List all worktrees in table format.

### `wm remove <path>`

Remove a worktree. Options:
- `-f, --force`: Skip confirmation
- `-b, --branch`: Also delete the branch

## Claude Code Skill

Claude Code에서 wm을 사용할 때 도움이 되는 스킬이 포함되어 있습니다.

```bash
# 스킬 복사 (Claude Code 사용자)
cp -r skills/wm ~/.claude/skills/
```

또는 프로젝트의 `.claude/skills/`에 복사하여 프로젝트별로 사용할 수 있습니다.

## License

MIT
