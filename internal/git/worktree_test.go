package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func setupTestRepo(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	// Resolve symlinks (macOS /var -> /private/var)
	tmpDir, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		t.Fatalf("failed to resolve symlinks: %v", err)
	}

	// Initialize a git repo with initial commit
	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
		{"touch", "README.md"},
		{"git", "add", "."},
		{"git", "commit", "-m", "initial"},
	}

	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = tmpDir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("setup command %v failed: %v\n%s", args, err, out)
		}
	}

	return tmpDir
}

func TestListWorktrees(t *testing.T) {
	repoDir := setupTestRepo(t)

	wts, err := ListWorktrees(repoDir)
	if err != nil {
		t.Fatalf("ListWorktrees failed: %v", err)
	}

	if len(wts) != 1 {
		t.Errorf("expected 1 worktree (main), got %d", len(wts))
	}

	// Resolve symlinks for comparison
	resolvedPath, _ := filepath.EvalSymlinks(wts[0].Path)
	if resolvedPath != repoDir {
		t.Errorf("expected path %s, got %s", repoDir, resolvedPath)
	}
}

func TestAddWorktree(t *testing.T) {
	repoDir := setupTestRepo(t)
	wtPath := filepath.Join(t.TempDir(), "feature-branch")

	err := AddWorktree(repoDir, wtPath, "feature-branch", true)
	if err != nil {
		t.Fatalf("AddWorktree failed: %v", err)
	}

	// Verify worktree exists
	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		t.Error("worktree directory was not created")
	}

	// Verify it shows in list
	wts, _ := ListWorktrees(repoDir)
	if len(wts) != 2 {
		t.Errorf("expected 2 worktrees, got %d", len(wts))
	}
}

func TestRemoveWorktree(t *testing.T) {
	repoDir := setupTestRepo(t)
	wtPath := filepath.Join(t.TempDir(), "to-remove")

	// Add worktree first
	if err := AddWorktree(repoDir, wtPath, "to-remove", true); err != nil {
		t.Fatalf("AddWorktree failed: %v", err)
	}

	// Remove it
	if err := RemoveWorktree(repoDir, wtPath, false); err != nil {
		t.Fatalf("RemoveWorktree failed: %v", err)
	}

	// Verify it's gone from list
	wts, _ := ListWorktrees(repoDir)
	if len(wts) != 1 {
		t.Errorf("expected 1 worktree after removal, got %d", len(wts))
	}
}

func TestBranchExists(t *testing.T) {
	repoDir := setupTestRepo(t)

	// Get the default branch name (could be main or master)
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = repoDir
	out, _ := cmd.Output()
	defaultBranch := string(out)
	if len(defaultBranch) > 0 && defaultBranch[len(defaultBranch)-1] == '\n' {
		defaultBranch = defaultBranch[:len(defaultBranch)-1]
	}

	// Test existing branch
	if !BranchExists(repoDir, defaultBranch) {
		t.Errorf("expected branch '%s' to exist", defaultBranch)
	}

	// Test non-existing branch
	if BranchExists(repoDir, "nonexistent-branch") {
		t.Error("expected nonexistent-branch to not exist")
	}
}

func TestDeleteBranch(t *testing.T) {
	repoDir := setupTestRepo(t)

	// Create a new branch first
	cmd := exec.Command("git", "branch", "test-delete-branch")
	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to create branch: %v\n%s", err, out)
	}

	// Verify it exists
	if !BranchExists(repoDir, "test-delete-branch") {
		t.Fatal("branch should exist before delete")
	}

	// Delete it
	if err := DeleteBranch(repoDir, "test-delete-branch", false); err != nil {
		t.Fatalf("DeleteBranch failed: %v", err)
	}

	// Verify it's gone
	if BranchExists(repoDir, "test-delete-branch") {
		t.Error("branch should not exist after delete")
	}
}

func TestGetRepoRoot(t *testing.T) {
	repoDir := setupTestRepo(t)

	// Create a subdirectory
	subDir := filepath.Join(repoDir, "sub", "dir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	// GetRepoRoot from subdirectory should return repo root
	root, err := GetRepoRoot(subDir)
	if err != nil {
		t.Fatalf("GetRepoRoot failed: %v", err)
	}

	// Resolve symlinks for comparison (repoDir is already resolved in setupTestRepo)
	resolvedRoot, _ := filepath.EvalSymlinks(root)
	if resolvedRoot != repoDir {
		t.Errorf("expected root %s, got %s", repoDir, resolvedRoot)
	}
}

func TestGetRepoRootNotARepo(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := GetRepoRoot(tmpDir)
	if err == nil {
		t.Error("expected error for non-git directory")
	}
}

func TestGetCurrentBranch(t *testing.T) {
	repoDir := setupTestRepo(t)

	branch, err := GetCurrentBranch(repoDir)
	if err != nil {
		t.Fatalf("GetCurrentBranch failed: %v", err)
	}

	// The branch could be "main" or "master" depending on git config
	if branch == "" {
		t.Error("expected non-empty branch name")
	}
}

func TestAddWorktreeExistingBranch(t *testing.T) {
	repoDir := setupTestRepo(t)
	wtPath := filepath.Join(t.TempDir(), "existing-branch-wt")

	// Create a branch first
	cmd := exec.Command("git", "branch", "existing-branch")
	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to create branch: %v\n%s", err, out)
	}

	// Add worktree for existing branch (createBranch=false)
	err := AddWorktree(repoDir, wtPath, "existing-branch", false)
	if err != nil {
		t.Fatalf("AddWorktree failed: %v", err)
	}

	// Verify worktree exists
	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		t.Error("worktree directory was not created")
	}
}

func TestRemoveWorktreeForce(t *testing.T) {
	repoDir := setupTestRepo(t)
	wtPath := filepath.Join(t.TempDir(), "force-remove")

	// Add worktree first
	if err := AddWorktree(repoDir, wtPath, "force-remove", true); err != nil {
		t.Fatalf("AddWorktree failed: %v", err)
	}

	// Create an untracked file to make it "dirty"
	if err := os.WriteFile(filepath.Join(wtPath, "untracked.txt"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Force remove should still work
	if err := RemoveWorktree(repoDir, wtPath, true); err != nil {
		t.Fatalf("RemoveWorktree with force failed: %v", err)
	}

	// Verify it's gone from list
	wts, _ := ListWorktrees(repoDir)
	if len(wts) != 1 {
		t.Errorf("expected 1 worktree after removal, got %d", len(wts))
	}
}

func TestParseWorktreeList(t *testing.T) {
	// Test parsing of git worktree list --porcelain output
	data := []byte(`worktree /path/to/main
HEAD abc1234567890
branch refs/heads/main

worktree /path/to/feature
HEAD def0987654321
branch refs/heads/feature-branch

`)

	worktrees := parseWorktreeList(data)

	if len(worktrees) != 2 {
		t.Fatalf("expected 2 worktrees, got %d", len(worktrees))
	}

	// Check first worktree
	if worktrees[0].Path != "/path/to/main" {
		t.Errorf("expected path /path/to/main, got %s", worktrees[0].Path)
	}
	if worktrees[0].HEAD != "abc1234567890" {
		t.Errorf("expected HEAD abc1234567890, got %s", worktrees[0].HEAD)
	}
	if worktrees[0].Branch != "main" {
		t.Errorf("expected branch main, got %s", worktrees[0].Branch)
	}

	// Check second worktree
	if worktrees[1].Path != "/path/to/feature" {
		t.Errorf("expected path /path/to/feature, got %s", worktrees[1].Path)
	}
	if worktrees[1].Branch != "feature-branch" {
		t.Errorf("expected branch feature-branch, got %s", worktrees[1].Branch)
	}
}

func TestParseWorktreeListBare(t *testing.T) {
	// Test parsing with bare repo
	data := []byte(`worktree /path/to/bare.git
bare

`)

	worktrees := parseWorktreeList(data)

	if len(worktrees) != 1 {
		t.Fatalf("expected 1 worktree, got %d", len(worktrees))
	}

	if !worktrees[0].Bare {
		t.Error("expected worktree to be marked as bare")
	}
}

func TestParseWorktreeListDetached(t *testing.T) {
	// Test parsing with detached HEAD
	data := []byte(`worktree /path/to/detached
HEAD abc1234567890
detached

`)

	worktrees := parseWorktreeList(data)

	if len(worktrees) != 1 {
		t.Fatalf("expected 1 worktree, got %d", len(worktrees))
	}

	if worktrees[0].Branch != "" {
		t.Errorf("expected empty branch for detached HEAD, got %s", worktrees[0].Branch)
	}
}
