package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/donghun/wm/internal/config"
	"github.com/donghun/wm/internal/detect"
	"github.com/donghun/wm/internal/ui"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize WM configuration",
	Long:  "Create a .wm.yaml configuration file with interactive prompts.",
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Check if already initialized
	configPath := filepath.Join(cwd, config.ConfigFileName)
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("%s already exists. Delete it first to reinitialize", config.ConfigFileName)
	}

	console := ui.NewConsole()
	repoName := filepath.Base(cwd)

	// Detect package manager
	detection := detect.Detect(cwd)

	console.Print("WM Init")
	console.Print("=======")
	console.Print("")

	if detection.PackageManager != "" {
		msg := fmt.Sprintf("Detected: %s", detection.PackageManager)
		if detection.IsMonorepo {
			msg += " (monorepo)"
		}
		console.Print(msg)
		console.Print("")
	}

	// Step 1: Base directory
	baseDir := console.Input(
		"Worktree base directory",
		"../wm_"+repoName,
	)

	// Step 2: Sync files
	syncFiles := console.Input(
		"Files to sync (comma-separated)",
		".env",
	)

	// Step 3: Post-install command (if detected)
	var installCmd string
	if detection.InstallCommand != "" {
		installCmd = console.Input(
			"Post-install command",
			detection.InstallCommand,
		)
	}

	// Build config
	cfg := config.NewConfig()
	cfg.Worktree.BaseDir = baseDir

	if syncFiles != "" {
		parts := strings.Split(syncFiles, ",")
		cfg.Sync = make([]config.SyncItem, len(parts))
		for i, part := range parts {
			path := strings.TrimSpace(part)
			cfg.Sync[i] = config.SyncItem{
				Src:  path,
				Dst:  path,
				Mode: "copy",
				When: "always",
			}
		}
	}

	if installCmd != "" {
		cfg.Tasks.PostInstall = config.PostInstallConfig{
			Mode:     "background",
			Commands: []string{installCmd},
		}
	}

	// Save config
	if err := config.SaveConfig(configPath, cfg); err != nil {
		return err
	}

	console.Print("")
	console.Printf("Created %s\n", config.ConfigFileName)
	console.Print("")
	console.Print("Next steps:")
	console.Print("  wm add <branch>  # Create a worktree")
	console.Print("  wm list          # List worktrees")

	return nil
}
