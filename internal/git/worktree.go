package git

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Worktree represents a git worktree entry
type Worktree struct {
	Path   string
	HEAD   string
	Branch string
	Bare   bool
}

// ListWorktrees returns all worktrees for a repository
func ListWorktrees(repoDir string) ([]Worktree, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = repoDir

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git worktree list failed: %w", err)
	}

	return parseWorktreeList(out), nil
}

func parseWorktreeList(data []byte) []Worktree {
	var worktrees []Worktree
	var current Worktree

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			if current.Path != "" {
				worktrees = append(worktrees, current)
				current = Worktree{}
			}
			continue
		}

		switch {
		case strings.HasPrefix(line, "worktree "):
			current.Path = strings.TrimPrefix(line, "worktree ")
		case strings.HasPrefix(line, "HEAD "):
			current.HEAD = strings.TrimPrefix(line, "HEAD ")
		case strings.HasPrefix(line, "branch "):
			branch := strings.TrimPrefix(line, "branch ")
			// Remove refs/heads/ prefix
			current.Branch = strings.TrimPrefix(branch, "refs/heads/")
		case line == "bare":
			current.Bare = true
		}
	}

	// Don't forget the last entry
	if current.Path != "" {
		worktrees = append(worktrees, current)
	}

	return worktrees
}

// AddWorktree creates a new worktree
func AddWorktree(repoDir, path, branch string, createBranch bool) error {
	args := []string{"worktree", "add"}
	if createBranch {
		args = append(args, "-b", branch)
	}
	args = append(args, path)
	if !createBranch {
		args = append(args, branch)
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = repoDir

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git worktree add failed: %w\n%s", err, out)
	}

	return nil
}

// RemoveWorktree removes a worktree
func RemoveWorktree(repoDir, path string, force bool) error {
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, path)

	cmd := exec.Command("git", args...)
	cmd.Dir = repoDir

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git worktree remove failed: %w\n%s", err, out)
	}

	return nil
}

// BranchExists checks if a branch exists
func BranchExists(repoDir, branch string) bool {
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	cmd.Dir = repoDir
	return cmd.Run() == nil
}

// DeleteBranch deletes a local branch
func DeleteBranch(repoDir, branch string, force bool) error {
	flag := "-d"
	if force {
		flag = "-D"
	}

	cmd := exec.Command("git", "branch", flag, branch)
	cmd.Dir = repoDir

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git branch delete failed: %w\n%s", err, out)
	}

	return nil
}

// GetRepoRoot returns the root directory of the git repository
func GetRepoRoot(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = dir

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not a git repository: %w", err)
	}

	return strings.TrimSpace(string(out)), nil
}

// GetCurrentBranch returns the current branch name
func GetCurrentBranch(dir string) (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = dir

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	return strings.TrimSpace(string(out)), nil
}
