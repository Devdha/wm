package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const ConfigFileName = ".wm.yaml"

// rawConfig is used for initial parsing to handle mixed sync types
type rawConfig struct {
	Version  int             `yaml:"version"`
	Worktree WorktreeConfig  `yaml:"worktree"`
	Scan     ScanConfig      `yaml:"scan"`
	Sync     []yaml.Node     `yaml:"sync"`
	Tasks    TasksConfig     `yaml:"tasks"`
}

// LoadConfig reads and parses a .wm.yaml file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	// Parse into raw config first to handle mixed sync types
	var raw rawConfig
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Build the final config
	cfg := NewConfig()
	cfg.Version = raw.Version
	cfg.Worktree = raw.Worktree
	cfg.Scan = raw.Scan
	cfg.Tasks = raw.Tasks

	// Handle mixed string/object sync items
	cfg.Sync = make([]SyncItem, len(raw.Sync))
	for i, node := range raw.Sync {
		if node.Kind == yaml.ScalarNode {
			// String value - just a path
			cfg.Sync[i] = SyncItem{
				Src:  node.Value,
				Mode: "copy",
				When: "always",
			}
		} else if node.Kind == yaml.MappingNode {
			// Object value
			var item SyncItem
			if err := node.Decode(&item); err != nil {
				return nil, fmt.Errorf("failed to parse sync item %d: %w", i, err)
			}
			// Set defaults
			if item.Mode == "" {
				item.Mode = "copy"
			}
			if item.When == "" {
				item.When = "always"
			}
			if item.Dst == "" {
				item.Dst = item.Src
			}
			cfg.Sync[i] = item
		}
	}

	return cfg, nil
}

// FindConfig searches for .wm.yaml starting from dir and walking up
func FindConfig(dir string) (string, error) {
	// Make sure we have an absolute path
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}
	dir = absDir

	for {
		path := filepath.Join(dir, ConfigFileName)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("no %s found", ConfigFileName)
}
