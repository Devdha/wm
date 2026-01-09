package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/donghun/wm/internal/config"
	"github.com/donghun/wm/internal/git"
	"github.com/donghun/wm/internal/runner"
	"github.com/donghun/wm/internal/sync"
	"github.com/spf13/cobra"
)

var addPath string

var addCmd = &cobra.Command{
	Use:   "add <branch>",
	Short: "Create a new worktree",
	Long:  "Create a new git worktree with file sync and optional background tasks.",
	Args:  cobra.ExactArgs(1),
	RunE:  runAdd,
}

func init() {
	addCmd.Flags().StringVarP(&addPath, "path", "p", "", "Custom path for the worktree")
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	branch := args[0]

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	repoRoot, err := git.GetRepoRoot(cwd)
	if err != nil {
		return err
	}

	repoName := filepath.Base(repoRoot)

	// Load config if exists
	var cfg *config.Config
	configPath := filepath.Join(repoRoot, config.ConfigFileName)
	if _, err := os.Stat(configPath); err == nil {
		cfg, err = config.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
	} else {
		cfg = config.NewConfig()
	}

	// Determine worktree path
	var wtPath string
	if addPath != "" {
		wtPath = addPath
		if !filepath.IsAbs(wtPath) {
			wtPath = filepath.Join(cwd, wtPath)
		}
	} else {
		baseDir := cfg.Worktree.BaseDir
		baseDir = strings.ReplaceAll(baseDir, "{repo}", repoName)
		if !filepath.IsAbs(baseDir) {
			baseDir = filepath.Join(repoRoot, baseDir)
		}
		wtPath = filepath.Join(baseDir, branch)
	}

	// Check if branch exists
	branchExists := git.BranchExists(repoRoot, branch)

	if !branchExists {
		// Prompt user
		fmt.Printf("Branch '%s' does not exist. Create it? [y/N]: ", branch)
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))

		if answer != "y" && answer != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	// Create worktree
	fmt.Printf("Creating worktree at %s...\n", wtPath)
	if err := git.AddWorktree(repoRoot, wtPath, branch, !branchExists); err != nil {
		return err
	}
	fmt.Println("Worktree created.")

	// Sync files
	if len(cfg.Sync) > 0 {
		fmt.Println("Syncing files...")
		if err := sync.SyncAll(repoRoot, wtPath, cfg.Sync); err != nil {
			return fmt.Errorf("failed to sync files: %w", err)
		}
		fmt.Printf("Synced %d file(s).\n", len(cfg.Sync))
	}

	// Run post-install tasks
	if len(cfg.Tasks.PostInstall.Commands) > 0 {
		fmt.Println("Running post-install tasks...")
		isBackground := cfg.Tasks.PostInstall.Mode == "background"
		if err := runner.RunCommands(wtPath, cfg.Tasks.PostInstall.Commands, isBackground); err != nil {
			return fmt.Errorf("post-install failed: %w", err)
		}
		if isBackground {
			fmt.Println("Background tasks started.")
		} else {
			fmt.Println("Post-install completed.")
		}
	}

	fmt.Printf("\nWorktree ready: %s\n", wtPath)
	fmt.Printf("  cd %s\n", wtPath)

	return nil
}
