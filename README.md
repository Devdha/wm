# WM - Git Worktree Manager

A CLI tool that makes git worktree easier to use with file sync and background tasks.

![CleanShot 2026-02-05 at 16 37 12](https://github.com/user-attachments/assets/51f9519d-01ea-478e-aefd-adbe3deb134a)

## Features

- **Interactive UI** - Arrow key selection for add/remove operations
- **Origin branch detection** - Automatically fetch from remote when branch exists on origin
- **File sync** - Copy or symlink files (`.env`, configs) to worktrees
- **Post-install tasks** - Run commands after worktree creation (background supported)
- **Pretty output** - Colored output with spinners and styled tables

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

# Create a worktree (interactive mode)
wm add

# Or specify branch directly
wm add feature-login

# List all worktrees
wm list

# Remove a worktree (interactive mode)
wm remove

# Or specify path/branch
wm remove feature-login

# Remove worktree and delete branch
wm remove -b feature-login
```

## Interactive Mode

### `wm add` (no arguments)

Select from origin branches or type a new branch name:

```
? Enter branch name or select from origin:
  [                    ]  ← Type here or press ↓
  ─────────────────────
  ❯ origin/feature-auth
    origin/bugfix-123
    origin/develop
```

- **Tab**: Switch between input and selection
- **↑↓**: Navigate options
- **Enter**: Confirm

### `wm remove` (no arguments)

Select a worktree to remove:

```
? Select worktree to remove:
  ❯ ../wm_repo/feature-auth (feature-auth)
    ../wm_repo/bugfix-123   (bugfix-123)
    ../wm_repo/main         (main) [main]
```

### Origin Branch Detection

When you run `wm add feature-auth` and the branch exists on origin:

```
⚡ Creating worktree for 'feature-auth'...

  Branch 'feature-auth' exists on origin but not locally.
? Checkout from origin? (Y/n)

📦 Fetching origin/feature-auth...
✓ Worktree ready: ../wm_repo/feature-auth
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

### `wm add [branch]`

Create a new worktree.

- Without arguments: Interactive mode (select from origin or type new name)
- With branch: Create worktree for that branch
- If branch exists on origin but not locally: Prompts to fetch

Options:
- `--path, -p`: Custom worktree path

**Note**: Branch names with slashes (e.g., `feature/auth`) create flat folders (`feature-auth`), not nested directories.

### `wm list`

List all worktrees in a styled table.

```
┌────────────────────────────┬────────────────┬─────────┐
│ PATH                       │ BRANCH         │ HEAD    │
├────────────────────────────┼────────────────┼─────────┤
│ ../wm_repo/feature-auth    │ feature-auth   │ a1b2c3d │
│ ../wm_repo/main            │ main           │ d4e5f6g │
└────────────────────────────┴────────────────┴─────────┘
```

### `wm remove [path]`

Remove a worktree.

- Without arguments: Interactive mode (select from list)
- With path or branch name: Remove that worktree

Options:
- `-f, --force`: Skip confirmation
- `-b, --branch`: Also delete the branch

## Claude Code Skill

A skill for Claude Code users is included:

```bash
# Copy skill (Claude Code users)
cp -r skills/wm ~/.claude/skills/
```

Or copy to your project's `.claude/skills/` for project-specific use.

## License

MIT
