package config

// Config represents the .wm.yaml file structure
type Config struct {
	Version  int            `yaml:"version"`
	Worktree WorktreeConfig `yaml:"worktree"`
	Scan     ScanConfig     `yaml:"scan"`
	Sync     []SyncItem     `yaml:"sync"`
	Tasks    TasksConfig    `yaml:"tasks"`
}

type WorktreeConfig struct {
	BaseDir string `yaml:"base_dir"`
}

type ScanConfig struct {
	IgnoreDirs []string `yaml:"ignore_dirs"`
}

// SyncItem can be a string path or an object with src/dst/mode/when
type SyncItem struct {
	Src  string `yaml:"src"`
	Dst  string `yaml:"dst,omitempty"`
	Mode string `yaml:"mode,omitempty"` // "copy" (default) or "symlink"
	When string `yaml:"when,omitempty"` // "always" (default) or "missing"
}

type TasksConfig struct {
	PostInstall PostInstallConfig `yaml:"post_install"`
}

type PostInstallConfig struct {
	Mode     string   `yaml:"mode"`
	Commands []string `yaml:"commands"`
	Notify   string   `yaml:"notify,omitempty"`
}

// NewConfig returns a Config with default values
func NewConfig() *Config {
	return &Config{
		Version: 1,
		Worktree: WorktreeConfig{
			BaseDir: "../wm_{repo}",
		},
		Scan: ScanConfig{
			IgnoreDirs: []string{".git", "node_modules", "dist", "build", ".next", "target", "vendor"},
		},
		Sync:  []SyncItem{},
		Tasks: TasksConfig{},
	}
}
