package git

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func runGit(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git %s failed: %w\n%s", args[0], err, out)
	}
	return out, nil
}

func runGitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s failed: %w", args[0], err)
	}
	return strings.TrimSpace(string(out)), nil
}

func runGitSilent(dir string, args ...string) bool {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	return cmd.Run() == nil
}

// Worktree represents a git worktree entry
type Worktree struct {
	Path   string
	HEAD   string
	Branch string
	Bare   bool
}

// ListWorktrees returns all worktrees for a repository
func ListWorktrees(repoDir string) ([]Worktree, error) {
	out, err := runGit(repoDir, "worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
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
			current.Branch = strings.TrimPrefix(branch, "refs/heads/")
		case line == "bare":
			current.Bare = true
		}
	}

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
	_, err := runGit(repoDir, args...)
	return err
}

// RemoveWorktree removes a worktree
func RemoveWorktree(repoDir, path string, force bool) error {
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, path)
	_, err := runGit(repoDir, args...)
	return err
}

// BranchExists checks if a branch exists
func BranchExists(repoDir, branch string) bool {
	return runGitSilent(repoDir, "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
}

// DeleteBranch deletes a local branch
func DeleteBranch(repoDir, branch string, force bool) error {
	flag := "-d"
	if force {
		flag = "-D"
	}
	_, err := runGit(repoDir, "branch", flag, branch)
	return err
}

// GetRepoRoot returns the root directory of the git repository
func GetRepoRoot(dir string) (string, error) {
	out, err := runGitOutput(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("not a git repository: %w", err)
	}
	return out, nil
}

// GetCurrentBranch returns the current branch name
func GetCurrentBranch(dir string) (string, error) {
	return runGitOutput(dir, "branch", "--show-current")
}

// WorktreeStatus returns the status summary for a worktree path
func WorktreeStatus(dir string) (modified int, untracked int, err error) {
	out, err := runGitOutput(dir, "status", "--porcelain")
	if err != nil {
		return 0, 0, err
	}

	if out == "" {
		return 0, 0, nil
	}

	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if len(line) < 2 {
			continue
		}
		xy := line[:2]
		if xy[0] == '?' && xy[1] == '?' {
			untracked++
		} else {
			modified++
		}
	}

	return modified, untracked, nil
}

// WorktreeAheadBehind returns ahead/behind counts relative to remote
func WorktreeAheadBehind(dir string) (ahead int, behind int, err error) {
	out, err := runGitOutput(dir, "rev-list", "--left-right", "--count", "HEAD...@{upstream}")
	if err != nil {
		return 0, 0, nil
	}

	var a, b int
	_, err = fmt.Sscanf(out, "%d\t%d", &a, &b)
	if err != nil {
		return 0, 0, nil
	}

	return a, b, nil
}
