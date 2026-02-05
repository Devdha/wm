package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Devdha/wm/internal/config"
	"github.com/Devdha/wm/internal/detect"
	"github.com/Devdha/wm/internal/ui"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize WM configuration",
	Long: `Create a .wm.yaml configuration file with interactive prompts.

Examples:
  wm init`,
	RunE: runInit,
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

	// Beautiful header
	fmt.Println()
	ui.DrawBox("WM - Git Worktree Manager", 35)
	fmt.Println()

	if detection.PackageManager != "" {
		msg := fmt.Sprintf("%s Detected: %s", ui.IconPackage, detection.PackageManager)
		if detection.IsMonorepo {
			msg += " (monorepo)"
		}
		ui.Info.Println(msg)
		fmt.Println()
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

	fmt.Println()
	ui.PrintSuccess("Created " + config.ConfigFileName)
	fmt.Println()
	ui.Bold.Println("Next steps:")
	ui.Muted.Println("  wm add <branch>  Create a worktree")
	ui.Muted.Println("  wm list          List worktrees")
	fmt.Println()

	return nil
}
