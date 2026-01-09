package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func setupTestRepo(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
	}

	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = tmpDir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("setup command %v failed: %v\n%s", args, err, out)
		}
	}

	// Create initial commit
	readme := filepath.Join(tmpDir, "README.md")
	if err := os.WriteFile(readme, []byte("# Test"), 0644); err != nil {
		t.Fatal(err)
	}

	cmds = [][]string{
		{"git", "add", "."},
		{"git", "commit", "-m", "initial"},
	}

	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = tmpDir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("commit command %v failed: %v\n%s", args, err, out)
		}
	}

	return tmpDir
}

func buildWM(t *testing.T) string {
	t.Helper()
	tmpBin := filepath.Join(t.TempDir(), "wm")

	cmd := exec.Command("go", "build", "-o", tmpBin, "..")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}

	return tmpBin
}

func TestE2E_ListEmpty(t *testing.T) {
	wmBin := buildWM(t)
	repoDir := setupTestRepo(t)

	cmd := exec.Command(wmBin, "list")
	cmd.Dir = repoDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("wm list failed: %v\n%s", err, out)
	}

	if !strings.Contains(string(out), "PATH") {
		t.Errorf("expected table header, got: %s", out)
	}
}

func TestE2E_AddAndRemove(t *testing.T) {
	wmBin := buildWM(t)
	repoDir := setupTestRepo(t)

	// Create .wm.yaml
	configContent := `version: 1
worktree:
  base_dir: "../wm_test"
sync: []
`
	if err := os.WriteFile(filepath.Join(repoDir, ".wm.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Add worktree (auto-create branch)
	cmd := exec.Command(wmBin, "add", "feature-test")
	cmd.Dir = repoDir
	cmd.Stdin = strings.NewReader("y\n") // Answer yes to create branch
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("wm add failed: %v\n%s", err, out)
	}

	if !strings.Contains(string(out), "Worktree ready") {
		t.Errorf("expected success message, got: %s", out)
	}

	// Verify worktree exists - need to resolve to absolute path with symlinks resolved
	wtPath := filepath.Join(repoDir, "..", "wm_test", "feature-test")
	wtPath, _ = filepath.Abs(wtPath)
	// Resolve symlinks (macOS /tmp -> /private/tmp)
	if realPath, err := filepath.EvalSymlinks(wtPath); err == nil {
		wtPath = realPath
	}
	if _, err := os.Stat(wtPath); os.IsNotExist(err) {
		t.Error("worktree directory was not created")
	}

	// List should show 2 worktrees
	cmd = exec.Command(wmBin, "list")
	cmd.Dir = repoDir
	out, _ = cmd.CombinedOutput()
	if !strings.Contains(string(out), "feature-test") {
		t.Errorf("expected feature-test in list, got: %s", out)
	}

	// Remove worktree - use absolute path with symlinks resolved
	cmd = exec.Command(wmBin, "remove", "-f", wtPath)
	cmd.Dir = repoDir
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("wm remove failed: %v\n%s", err, out)
	}

	// List should show 1 worktree
	cmd = exec.Command(wmBin, "list")
	cmd.Dir = repoDir
	out, _ = cmd.CombinedOutput()
	if strings.Contains(string(out), "feature-test") {
		t.Errorf("feature-test should be removed, got: %s", out)
	}
}

func TestE2E_SyncFiles(t *testing.T) {
	wmBin := buildWM(t)
	repoDir := setupTestRepo(t)

	// Create .env in repo
	envContent := "SECRET=test123"
	if err := os.WriteFile(filepath.Join(repoDir, ".env"), []byte(envContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create .wm.yaml with sync
	configContent := `version: 1
worktree:
  base_dir: "../wm_sync_test"
sync:
  - ".env"
`
	if err := os.WriteFile(filepath.Join(repoDir, ".wm.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Commit .env so we can create worktree
	cmd := exec.Command("git", "add", ".env", ".wm.yaml")
	cmd.Dir = repoDir
	cmd.Run()
	cmd = exec.Command("git", "commit", "-m", "add env")
	cmd.Dir = repoDir
	cmd.Run()

	// Add worktree
	cmd = exec.Command(wmBin, "add", "sync-test")
	cmd.Dir = repoDir
	cmd.Stdin = strings.NewReader("y\n")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("wm add failed: %v\n%s", err, out)
	}

	// Verify .env was synced
	wtPath := filepath.Join(repoDir, "..", "wm_sync_test", "sync-test")
	syncedEnv := filepath.Join(wtPath, ".env")
	content, err := os.ReadFile(syncedEnv)
	if err != nil {
		t.Fatalf("failed to read synced .env: %v", err)
	}

	if string(content) != envContent {
		t.Errorf("expected '%s', got '%s'", envContent, content)
	}
}

func TestE2E_RemoveWithBranch(t *testing.T) {
	wmBin := buildWM(t)
	repoDir := setupTestRepo(t)

	// Create config
	configContent := `version: 1
worktree:
  base_dir: "../wm_branch_test"
`
	if err := os.WriteFile(filepath.Join(repoDir, ".wm.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Add worktree
	cmd := exec.Command(wmBin, "add", "branch-to-delete")
	cmd.Dir = repoDir
	cmd.Stdin = strings.NewReader("y\n")
	cmd.Run()

	// Resolve to absolute path with symlinks resolved
	wtPath := filepath.Join(repoDir, "..", "wm_branch_test", "branch-to-delete")
	wtPath, _ = filepath.Abs(wtPath)
	// Resolve symlinks (macOS /tmp -> /private/tmp)
	if realPath, err := filepath.EvalSymlinks(wtPath); err == nil {
		wtPath = realPath
	}

	// Remove with -b flag
	cmd = exec.Command(wmBin, "remove", "-f", "-b", wtPath)
	cmd.Dir = repoDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("wm remove -b failed: %v\n%s", err, out)
	}

	// Verify branch is deleted
	cmd = exec.Command("git", "branch", "--list", "branch-to-delete")
	cmd.Dir = repoDir
	out, _ = cmd.CombinedOutput()
	if strings.Contains(string(out), "branch-to-delete") {
		t.Error("branch should have been deleted")
	}
}

func TestE2E_Version(t *testing.T) {
	wmBin := buildWM(t)

	cmd := exec.Command(wmBin, "version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("wm version failed: %v\n%s", err, out)
	}

	if !strings.Contains(string(out), "wm") {
		t.Errorf("expected version info, got: %s", out)
	}
}
