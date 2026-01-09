package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/donghun/wm/internal/config"
	"github.com/donghun/wm/internal/detect"
	"github.com/donghun/wm/internal/tui"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize WM configuration",
	Long:  "Create a .wm.yaml configuration file with interactive TUI.",
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

	repoName := filepath.Base(cwd)

	// Detect package manager
	detection := detect.Detect(cwd)

	// Run TUI
	cfg, err := tui.RunInitTUI(cwd, repoName, detection)
	if err != nil {
		return err
	}

	// Save config
	if err := config.SaveConfig(configPath, cfg); err != nil {
		return err
	}

	fmt.Printf("\nCreated %s\n", config.ConfigFileName)
	fmt.Println("\nNext steps:")
	fmt.Println("  wm add <branch>  # Create a worktree")
	fmt.Println("  wm list          # List worktrees")

	return nil
}
