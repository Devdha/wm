package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".wm.yaml")

	// Write test config
	content := []byte(`version: 1
worktree:
  base_dir: "../custom_wm"
sync:
  - ".env"
  - src: ".env.example"
    dst: ".env"
    mode: copy
    when: missing
tasks:
  post_install:
    mode: background
    commands:
      - "pnpm install"
`)
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Worktree.BaseDir != "../custom_wm" {
		t.Errorf("expected base_dir '../custom_wm', got %s", cfg.Worktree.BaseDir)
	}

	if len(cfg.Sync) != 2 {
		t.Errorf("expected 2 sync items, got %d", len(cfg.Sync))
	}

	if cfg.Sync[0].Src != ".env" {
		t.Errorf("expected first sync item '.env', got %s", cfg.Sync[0].Src)
	}

	if cfg.Sync[1].Mode != "copy" {
		t.Errorf("expected mode 'copy', got %s", cfg.Sync[1].Mode)
	}
}

func TestLoadConfigNotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/.wm.yaml")
	if err == nil {
		t.Error("expected error for nonexistent config")
	}
}

func TestFindConfig(t *testing.T) {
	// Create a temp directory structure
	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "a", "b", "c")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create .wm.yaml in root
	configPath := filepath.Join(tmpDir, ".wm.yaml")
	if err := os.WriteFile(configPath, []byte("version: 1"), 0644); err != nil {
		t.Fatal(err)
	}

	// Search from nested directory
	found, err := FindConfig(nestedDir)
	if err != nil {
		t.Fatalf("FindConfig failed: %v", err)
	}

	if found != configPath {
		t.Errorf("expected %s, got %s", configPath, found)
	}
}

func TestFindConfigNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	_, err := FindConfig(tmpDir)
	if err == nil {
		t.Error("expected error when no config found")
	}
}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".wm.yaml")

	cfg := NewConfig()
	cfg.Worktree.BaseDir = "../test_wm"
	cfg.Sync = []SyncItem{
		{Src: ".env", Mode: "copy", When: "always"},
	}

	if err := SaveConfig(configPath, cfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Verify by loading
	loaded, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if loaded.Worktree.BaseDir != "../test_wm" {
		t.Errorf("expected base_dir '../test_wm', got %s", loaded.Worktree.BaseDir)
	}
}
