// Package workspace provides the core domain logic for worktree management.
// It follows the Repository pattern - one Workspace represents one git repository.
package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Devdha/wm/internal/config"
	"github.com/Devdha/wm/internal/git"
	"github.com/Devdha/wm/internal/runner"
	"github.com/Devdha/wm/internal/sync"
	"github.com/Devdha/wm/internal/ui"
)

// AddResult contains the result of adding a worktree
type AddResult struct {
	Branch        string   // Branch name
	Path          string   // Worktree path
	CreatedBranch bool     // Whether a new branch was created
	SyncedFiles   []string // Files that were synced
	PostInstall   []string // Post-install commands that were run
	IsBackground  bool     // Whether post-install ran in background
}

// RemoveResult contains the result of removing a worktree
type RemoveResult struct {
	Path          string // Worktree path that was removed
	Branch        string // Associated branch name
	BranchDeleted bool   // Whether the branch was also deleted
}

// Workspace represents a git repository with WM configuration
type Workspace struct {
	Root   string         // Repository root path
	Name   string         // Repository name (basename of root)
	Config *config.Config // WM configuration
	UI     ui.Prompter    // User interaction handler
}

// Open creates a Workspace from the current directory
func Open(ui ui.Prompter) (*Workspace, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}
	return OpenAt(cwd, ui)
}

// OpenAt creates a Workspace from a specific directory
func OpenAt(dir string, prompter ui.Prompter) (*Workspace, error) {
	root, err := git.GetRepoRoot(dir)
	if err != nil {
		return nil, err
	}

	cfg := loadConfigOrDefault(root)

	return &Workspace{
		Root:   root,
		Name:   filepath.Base(root),
		Config: cfg,
		UI:     prompter,
	}, nil
}

func loadConfigOrDefault(root string) *config.Config {
	configPath := filepath.Join(root, config.ConfigFileName)
	if cfg, err := config.LoadConfig(configPath); err == nil {
		return cfg
	}
	return config.NewConfig()
}

// ListWorktrees returns all worktrees in this workspace
func (w *Workspace) ListWorktrees() ([]git.Worktree, error) {
	return git.ListWorktrees(w.Root)
}

// AddWorktree creates a new worktree with optional sync and post-install
func (w *Workspace) AddWorktree(branch string, customPath string) (*AddResult, error) {
	wtPath := w.resolveWorktreePath(branch, customPath)
	createBranch := !git.BranchExists(w.Root, branch)

	if createBranch {
		msg := fmt.Sprintf("Branch '%s' does not exist. Create it?", branch)
		if !w.UI.Confirm(msg) {
			return nil, fmt.Errorf("user canceled")
		}
	}

	if err := git.AddWorktree(w.Root, wtPath, branch, createBranch); err != nil {
		return nil, err
	}

	result := &AddResult{
		Branch:        branch,
		Path:          wtPath,
		CreatedBranch: createBranch,
		SyncedFiles:   []string{},
		PostInstall:   []string{},
	}

	if syncedFiles, err := w.syncFiles(wtPath); err != nil {
		return nil, err
	} else {
		result.SyncedFiles = syncedFiles
	}

	if postInstall, isBackground, err := w.runPostInstall(wtPath); err != nil {
		return nil, err
	} else {
		result.PostInstall = postInstall
		result.IsBackground = isBackground
	}

	return result, nil
}

func (w *Workspace) resolveWorktreePath(branch, customPath string) string {
	if customPath != "" {
		if filepath.IsAbs(customPath) {
			return customPath
		}
		cwd, _ := os.Getwd()
		return filepath.Join(cwd, customPath)
	}

	baseDir := strings.ReplaceAll(w.Config.Worktree.BaseDir, "{repo}", w.Name)
	if !filepath.IsAbs(baseDir) {
		baseDir = filepath.Join(w.Root, baseDir)
	}
	return filepath.Join(baseDir, sanitizeBranchName(branch))
}

func sanitizeBranchName(branch string) string {
	return strings.ReplaceAll(branch, "/", "-")
}

func (w *Workspace) syncFiles(wtPath string) ([]string, error) {
	if len(w.Config.Sync) == 0 {
		return []string{}, nil
	}

	if err := sync.SyncAll(w.Root, wtPath, w.Config.Sync); err != nil {
		return nil, fmt.Errorf("failed to sync files: %w", err)
	}

	syncedFiles := make([]string, 0, len(w.Config.Sync))
	for _, item := range w.Config.Sync {
		syncedFiles = append(syncedFiles, item.Src)
	}

	return syncedFiles, nil
}

func (w *Workspace) runPostInstall(wtPath string) ([]string, bool, error) {
	cmds := w.Config.Tasks.PostInstall.Commands
	if len(cmds) == 0 {
		return []string{}, false, nil
	}

	isBackground := w.Config.Tasks.PostInstall.Mode == "background"

	if err := runner.RunCommands(wtPath, cmds, isBackground); err != nil {
		return nil, false, fmt.Errorf("post-install failed: %w", err)
	}

	return cmds, isBackground, nil
}

// RemoveWorktree removes a worktree and optionally its branch
func (w *Workspace) RemoveWorktree(path string, deleteBranch, force bool) (*RemoveResult, error) {
	worktrees, err := w.ListWorktrees()
	if err != nil {
		return nil, err
	}

	target := w.findWorktree(worktrees, path)
	if target == nil {
		return nil, fmt.Errorf("worktree '%s' not found", path)
	}

	if target.Path == w.Root {
		return nil, fmt.Errorf("cannot remove the main worktree")
	}

	if !force {
		msg := fmt.Sprintf("Remove worktree at '%s'", target.Path)
		if deleteBranch && target.Branch != "" {
			msg += fmt.Sprintf(" and branch '%s'", target.Branch)
		}
		msg += "?"
		if !w.UI.Confirm(msg) {
			return nil, fmt.Errorf("user canceled")
		}
	}

	if deleteBranch && target.Branch != "" {
		if err := w.checkBranchNotUsedElsewhere(worktrees, target); err != nil {
			return nil, err
		}
	}

	if err := git.RemoveWorktree(w.Root, target.Path, force); err != nil {
		return nil, err
	}

	result := &RemoveResult{
		Path:          target.Path,
		Branch:        target.Branch,
		BranchDeleted: false,
	}

	if deleteBranch && target.Branch != "" {
		if err := w.deleteBranch(target.Branch); err == nil {
			result.BranchDeleted = true
		}
	}

	return result, nil
}

func (w *Workspace) findWorktree(worktrees []git.Worktree, path string) *git.Worktree {
	absPath := resolvePath(path)
	sanitizedPath := sanitizeBranchName(path)

	for i, wt := range worktrees {
		wtResolved := resolvePath(wt.Path)
		sep := string(filepath.Separator)
		if wtResolved == absPath || wt.Path == path ||
			strings.HasSuffix(wt.Path, sep+path) ||
			strings.HasSuffix(wt.Path, sep+sanitizedPath) {
			return &worktrees[i]
		}
	}
	return nil
}

func resolvePath(path string) string {
	if !filepath.IsAbs(path) {
		path, _ = filepath.Abs(path)
	}
	resolved, _ := filepath.EvalSymlinks(path)
	if resolved != "" {
		return resolved
	}
	return path
}

func (w *Workspace) checkBranchNotUsedElsewhere(worktrees []git.Worktree, target *git.Worktree) error {
	for _, wt := range worktrees {
		if wt.Path != target.Path && wt.Branch == target.Branch {
			return fmt.Errorf("cannot delete branch '%s': used by worktree at '%s'",
				target.Branch, wt.Path)
		}
	}
	return nil
}

func (w *Workspace) deleteBranch(branch string) error {
	return git.DeleteBranch(w.Root, branch, false)
}
