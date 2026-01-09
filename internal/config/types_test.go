package config

import "testing"

func TestConfigDefaults(t *testing.T) {
	cfg := NewConfig()
	if cfg.Version != 1 {
		t.Errorf("expected version 1, got %d", cfg.Version)
	}
	if cfg.Worktree.BaseDir != "../wm_{repo}" {
		t.Errorf("expected default base_dir, got %s", cfg.Worktree.BaseDir)
	}
}
