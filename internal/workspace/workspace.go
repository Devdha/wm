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
func (w *Workspace) AddWorktree(branch string, customPath string) error {
	wtPath := w.resolveWorktreePath(branch, customPath)
	createBranch := !git.BranchExists(w.Root, branch)

	if createBranch {
		msg := fmt.Sprintf("Branch '%s' does not exist. Create it?", branch)
		if !w.UI.Confirm(msg) {
			ui.PrintWarning("Aborted.")
			return nil
		}
	}

	fmt.Println()
	ui.PrintStep(ui.IconBolt, "Creating worktree...")
	ui.PrintSubStep(fmt.Sprintf("Branch: %s", branch))
	ui.PrintSubStepEnd(fmt.Sprintf("Path: %s", wtPath))
	fmt.Println()

	if err := git.AddWorktree(w.Root, wtPath, branch, createBranch); err != nil {
		return err
	}

	if err := w.syncFiles(wtPath); err != nil {
		return err
	}

	if err := w.runPostInstall(wtPath); err != nil {
		return err
	}

	fmt.Println()
	ui.Success.Print(ui.IconCheck + " ")
	ui.Bold.Print("Worktree ready: ")
	fmt.Println(wtPath)
	fmt.Println()
	ui.Muted.Printf("  cd %s\n", wtPath)
	fmt.Println()
	return nil
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

func (w *Workspace) syncFiles(wtPath string) error {
	if len(w.Config.Sync) == 0 {
		return nil
	}

	ui.PrintStep(ui.IconPackage, "Syncing files...")
	if err := sync.SyncAll(w.Root, wtPath, w.Config.Sync); err != nil {
		return fmt.Errorf("failed to sync files: %w", err)
	}

	for _, item := range w.Config.Sync {
		ui.PrintSubStep(fmt.Sprintf("%s %s", item.Src, ui.Success.Sprint(ui.IconCheck)))
	}
	fmt.Println()
	return nil
}

func (w *Workspace) runPostInstall(wtPath string) error {
	cmds := w.Config.Tasks.PostInstall.Commands
	if len(cmds) == 0 {
		return nil
	}

	ui.PrintStep(ui.IconRocket, "Running post-install...")
	isBackground := w.Config.Tasks.PostInstall.Mode == "background"

	if err := runner.RunCommands(wtPath, cmds, isBackground); err != nil {
		return fmt.Errorf("post-install failed: %w", err)
	}

	for _, cmd := range cmds {
		suffix := ""
		if isBackground {
			suffix = " " + ui.Muted.Sprint("(background)")
		}
		ui.PrintSubStep(cmd + suffix)
	}
	fmt.Println()
	return nil
}

// RemoveWorktree removes a worktree and optionally its branch
func (w *Workspace) RemoveWorktree(path string, deleteBranch, force bool) error {
	worktrees, err := w.ListWorktrees()
	if err != nil {
		return err
	}

	target := w.findWorktree(worktrees, path)
	if target == nil {
		return fmt.Errorf("worktree '%s' not found", path)
	}

	if target.Path == w.Root {
		return fmt.Errorf("cannot remove the main worktree")
	}

	if !force {
		msg := fmt.Sprintf("Remove worktree at '%s'", target.Path)
		if deleteBranch && target.Branch != "" {
			msg += fmt.Sprintf(" and branch '%s'", target.Branch)
		}
		msg += "?"
		if !w.UI.Confirm(msg) {
			ui.PrintWarning("Aborted.")
			return nil
		}
	}

	if deleteBranch && target.Branch != "" {
		if err := w.checkBranchNotUsedElsewhere(worktrees, target); err != nil {
			return err
		}
	}

	fmt.Println()
	spinner := ui.NewSpinner("Removing worktree...")
	spinner.Start()

	if err := git.RemoveWorktree(w.Root, target.Path, force); err != nil {
		spinner.Fail("Failed to remove worktree")
		return err
	}
	spinner.Success(fmt.Sprintf("Removed worktree: %s", target.Path))

	if deleteBranch && target.Branch != "" {
		w.deleteBranch(target.Branch)
	}

	fmt.Println()
	return nil
}

func (w *Workspace) findWorktree(worktrees []git.Worktree, path string) *git.Worktree {
	absPath := resolvePath(path)
	sanitizedPath := sanitizeBranchName(path)

	for i, wt := range worktrees {
		wtResolved := resolvePath(wt.Path)
		if wtResolved == absPath || wt.Path == path ||
			strings.HasSuffix(wt.Path, "/"+path) ||
			strings.HasSuffix(wt.Path, "/"+sanitizedPath) {
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

func (w *Workspace) deleteBranch(branch string) {
	spinner := ui.NewSpinner(fmt.Sprintf("Deleting branch '%s'...", branch))
	spinner.Start()

	if err := git.DeleteBranch(w.Root, branch, false); err != nil {
		spinner.Fail(fmt.Sprintf("Failed to delete branch '%s': %v", branch, err))
		ui.Muted.Println("  Tip: Use 'git branch -D' to force delete.")
	} else {
		spinner.Success(fmt.Sprintf("Deleted branch: %s", branch))
	}
}
