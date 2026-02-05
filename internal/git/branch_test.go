package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestListRemoteBranches(t *testing.T) {
	// Create a temporary git repository
	tmpDir, err := os.MkdirTemp("", "wm-branch-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize git repo with main branch
	runGit(t, tmpDir, "init", "-b", "main")
	runGit(t, tmpDir, "config", "user.name", "Test User")
	runGit(t, tmpDir, "config", "user.email", "test@example.com")

	// Create initial commit
	testFile := filepath.Join(tmpDir, "README.md")
	if err := os.WriteFile(testFile, []byte("# Test"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	runGit(t, tmpDir, "add", "README.md")
	runGit(t, tmpDir, "commit", "-m", "Initial commit")

	// Create a bare remote repository
	remoteDir, err := os.MkdirTemp("", "wm-branch-test-remote-*")
	if err != nil {
		t.Fatalf("Failed to create remote dir: %v", err)
	}
	defer os.RemoveAll(remoteDir)

	runGit(t, remoteDir, "init", "--bare")

	// Add remote and push
	runGit(t, tmpDir, "remote", "add", "origin", remoteDir)
	runGit(t, tmpDir, "push", "-u", "origin", "main")

	// Create and push a feature branch
	runGit(t, tmpDir, "checkout", "-b", "feature/test")
	testFile2 := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile2, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	runGit(t, tmpDir, "add", "test.txt")
	runGit(t, tmpDir, "commit", "-m", "Add test file")
	runGit(t, tmpDir, "push", "-u", "origin", "feature/test")

	// Go back to main
	runGit(t, tmpDir, "checkout", "main")

	// Test ListRemoteBranches
	branches, err := ListRemoteBranches(tmpDir)
	if err != nil {
		t.Fatalf("ListRemoteBranches failed: %v", err)
	}

	if len(branches) != 2 {
		t.Errorf("Expected 2 remote branches, got %d: %v", len(branches), branches)
	}

	expectedBranches := map[string]bool{
		"origin/main":         false,
		"origin/feature/test": false,
	}

	for _, branch := range branches {
		if _, ok := expectedBranches[branch]; ok {
			expectedBranches[branch] = true
		}
	}

	for branch, found := range expectedBranches {
		if !found {
			t.Errorf("Expected branch %s not found in results", branch)
		}
	}
}

func TestRemoteBranchExists(t *testing.T) {
	// Create a temporary git repository
	tmpDir, err := os.MkdirTemp("", "wm-branch-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize git repo with main branch
	runGit(t, tmpDir, "init", "-b", "main")
	runGit(t, tmpDir, "config", "user.name", "Test User")
	runGit(t, tmpDir, "config", "user.email", "test@example.com")

	// Create initial commit
	testFile := filepath.Join(tmpDir, "README.md")
	if err := os.WriteFile(testFile, []byte("# Test"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	runGit(t, tmpDir, "add", "README.md")
	runGit(t, tmpDir, "commit", "-m", "Initial commit")

	// Create a bare remote repository
	remoteDir, err := os.MkdirTemp("", "wm-branch-test-remote-*")
	if err != nil {
		t.Fatalf("Failed to create remote dir: %v", err)
	}
	defer os.RemoveAll(remoteDir)

	runGit(t, remoteDir, "init", "--bare")

	// Add remote and push
	runGit(t, tmpDir, "remote", "add", "origin", remoteDir)
	runGit(t, tmpDir, "push", "-u", "origin", "main")

	// Test existing branch
	if !RemoteBranchExists(tmpDir, "main") {
		t.Error("RemoteBranchExists should return true for existing branch 'main'")
	}

	// Test non-existing branch
	if RemoteBranchExists(tmpDir, "nonexistent") {
		t.Error("RemoteBranchExists should return false for non-existing branch")
	}
}

func TestFetchBranch(t *testing.T) {
	// Create a temporary git repository
	tmpDir, err := os.MkdirTemp("", "wm-branch-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize git repo with main branch
	runGit(t, tmpDir, "init", "-b", "main")
	runGit(t, tmpDir, "config", "user.name", "Test User")
	runGit(t, tmpDir, "config", "user.email", "test@example.com")

	// Create initial commit
	testFile := filepath.Join(tmpDir, "README.md")
	if err := os.WriteFile(testFile, []byte("# Test"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	runGit(t, tmpDir, "add", "README.md")
	runGit(t, tmpDir, "commit", "-m", "Initial commit")

	// Create a bare remote repository
	remoteDir, err := os.MkdirTemp("", "wm-branch-test-remote-*")
	if err != nil {
		t.Fatalf("Failed to create remote dir: %v", err)
	}
	defer os.RemoveAll(remoteDir)

	runGit(t, remoteDir, "init", "--bare")

	// Add remote and push
	runGit(t, tmpDir, "remote", "add", "origin", remoteDir)
	runGit(t, tmpDir, "push", "-u", "origin", "main")

	// Create a feature branch on remote only
	runGit(t, tmpDir, "checkout", "-b", "feature/remote")
	testFile2 := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile2, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	runGit(t, tmpDir, "add", "test.txt")
	runGit(t, tmpDir, "commit", "-m", "Add test file")
	runGit(t, tmpDir, "push", "-u", "origin", "feature/remote")

	// Go back to main and delete local feature branch
	runGit(t, tmpDir, "checkout", "main")
	runGit(t, tmpDir, "branch", "-D", "feature/remote")

	// Test FetchBranch
	if err := FetchBranch(tmpDir, "feature/remote"); err != nil {
		t.Fatalf("FetchBranch failed: %v", err)
	}

	// Verify the branch ref exists
	cmd := exec.Command("git", "show-ref", "refs/remotes/origin/feature/remote")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Error("Branch should exist after fetch")
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
}
