# WM - Git Worktree Manager

A CLI tool that makes git worktree easier to use with file sync and background tasks.

## Installation

```bash
go install github.com/donghun/wm@latest
```

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

## License

MIT
