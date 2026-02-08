package git

import (
	"bytes"
	"strings"
)

// ListRemoteBranches returns all branches on origin
func ListRemoteBranches(repoDir string) ([]string, error) {
	out, err := runGitOutput(repoDir, "branch", "-r", "--format=%(refname:short)")
	if err != nil {
		return nil, err
	}

	var branches []string
	lines := bytes.Split([]byte(out), []byte("\n"))

	for _, line := range lines {
		lineStr := strings.TrimSpace(string(line))
		if lineStr != "" && !strings.Contains(lineStr, "HEAD") {
			branches = append(branches, lineStr)
		}
	}

	return branches, nil
}

// RemoteBranchExists checks if branch exists on origin
func RemoteBranchExists(repoDir, branch string) bool {
	out, err := runGitOutput(repoDir, "ls-remote", "--heads", "origin", branch)
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(out)) > 0
}

// FetchBranch fetches a specific branch from origin
func FetchBranch(repoDir, branch string) error {
	_, err := runGit(repoDir, "fetch", "origin", branch)
	return err
}

// CheckoutRemoteBranch creates local branch tracking remote
func CheckoutRemoteBranch(repoDir, branch string) error {
	if err := FetchBranch(repoDir, branch); err != nil {
		return err
	}
	_, err := runGit(repoDir, "checkout", "-b", branch, "origin/"+branch)
	return err
}
