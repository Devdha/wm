package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// ListRemoteBranches returns all branches on origin
func ListRemoteBranches(repoDir string) ([]string, error) {
	cmd := exec.Command("git", "branch", "-r", "--format=%(refname:short)")
	cmd.Dir = repoDir

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git branch -r failed: %w", err)
	}

	var branches []string
	lines := bytes.Split(out, []byte("\n"))

	for _, line := range lines {
		lineStr := strings.TrimSpace(string(line))
		// Filter out HEAD reference
		if lineStr != "" && !strings.Contains(lineStr, "HEAD") {
			branches = append(branches, lineStr)
		}
	}

	return branches, nil
}

// RemoteBranchExists checks if branch exists on origin
func RemoteBranchExists(repoDir, branch string) bool {
	cmd := exec.Command("git", "ls-remote", "--heads", "origin", branch)
	cmd.Dir = repoDir

	out, err := cmd.Output()
	if err != nil {
		return false
	}

	return len(strings.TrimSpace(string(out))) > 0
}

// FetchBranch fetches a specific branch from origin
func FetchBranch(repoDir, branch string) error {
	cmd := exec.Command("git", "fetch", "origin", branch)
	cmd.Dir = repoDir

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git fetch failed: %w\n%s", err, out)
	}

	return nil
}

// CheckoutRemoteBranch creates local branch tracking remote
func CheckoutRemoteBranch(repoDir, branch string) error {
	// First fetch the branch
	if err := FetchBranch(repoDir, branch); err != nil {
		return err
	}

	// Create local branch tracking remote
	cmd := exec.Command("git", "checkout", "-b", branch, "origin/"+branch)
	cmd.Dir = repoDir

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout failed: %w\n%s", err, out)
	}

	return nil
}
