package cmd

import (
	"fmt"
	"strings"

	"github.com/Devdha/wm/internal/git"
	"github.com/Devdha/wm/internal/ui"
	"github.com/Devdha/wm/internal/workspace"
	"github.com/spf13/cobra"
)

var addPath string

var addCmd = &cobra.Command{
	Use:   "add [branch]",
	Short: "Create a new worktree",
	Long:  "Create a new git worktree with file sync and optional background tasks. If no branch is provided, an interactive selection menu will appear.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runAdd,
}

func init() {
	addCmd.Flags().StringVarP(&addPath, "path", "p", "", "Custom path for the worktree")
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	ws, err := workspace.Open(ui.NewConsole())
	if err != nil {
		return err
	}

	var branch string
	var customPath string

	// Interactive mode if no branch provided
	if len(args) == 0 {
		branch, customPath, err = runAddInteractive(ws)
		if err != nil {
			return err
		}
	} else {
		branch = args[0]
		customPath = addPath

		// Check if branch exists on origin but not locally
		if !git.BranchExists(ws.Root, branch) && git.RemoteBranchExists(ws.Root, branch) {
			if err := handleRemoteBranch(ws, branch); err != nil {
				return err
			}
		}
	}

	return ws.AddWorktree(branch, customPath)
}

func runAddInteractive(ws *workspace.Workspace) (string, string, error) {
	// Get remote branches
	remoteBranches, err := git.ListRemoteBranches(ws.Root)
	if err != nil {
		remoteBranches = []string{} // Continue even if remote fetch fails
	}

	// Build options from remote branches
	options := make([]ui.SelectOption, 0, len(remoteBranches))
	for _, remoteBranch := range remoteBranches {
		// Strip "origin/" prefix for display
		displayName := strings.TrimPrefix(remoteBranch, "origin/")
		options = append(options, ui.SelectOption{
			Label: remoteBranch,
			Value: displayName,
		})
	}

	// Get branch name (input or select)
	var branch string
	if len(options) > 0 {
		branch, err = ui.InputWithOptions(
			"Enter branch name or select from origin:",
			"branch-name",
			options,
		)
	} else {
		branch, err = ui.Input(
			"Enter branch name:",
			"branch-name",
		)
	}

	if err != nil {
		return "", "", err
	}

	if branch == "" {
		return "", "", fmt.Errorf("branch name cannot be empty")
	}

	// Ask for custom path
	customPath, err := ui.Input(
		"Custom path? (leave empty for default):",
		"",
	)
	if err != nil {
		return "", "", err
	}

	// Check if branch exists on origin but not locally
	if !git.BranchExists(ws.Root, branch) && git.RemoteBranchExists(ws.Root, branch) {
		if err := handleRemoteBranch(ws, branch); err != nil {
			return "", "", err
		}
	}

	return branch, customPath, nil
}

func handleRemoteBranch(ws *workspace.Workspace, branch string) error {
	fmt.Println()
	ui.PrintStep(ui.IconBolt, fmt.Sprintf("Creating worktree for '%s'...", branch))
	fmt.Println()
	ui.Info.Printf("  Branch '%s' exists on origin but not locally.\n", branch)

	if ws.UI.Confirm("Checkout from origin?") {
		fmt.Println()
		spinner := ui.NewSpinner(fmt.Sprintf("Fetching origin/%s...", branch))
		spinner.Start()

		if err := git.FetchBranch(ws.Root, branch); err != nil {
			spinner.Fail(fmt.Sprintf("Failed to fetch branch: %v", err))
			return err
		}

		spinner.Success(fmt.Sprintf("Fetched origin/%s", branch))
		fmt.Println()
	} else {
		ui.PrintWarning("Aborted.")
		return fmt.Errorf("user canceled")
	}

	return nil
}
