package cmd

import (
	"fmt"

	"github.com/Devdha/wm/internal/ui"
	"github.com/Devdha/wm/internal/workspace"
	"github.com/spf13/cobra"
)

var (
	removeForce        bool
	removeDeleteBranch bool
)

var removeCmd = &cobra.Command{
	Use:     "remove [path]",
	Aliases: []string{"rm"},
	Short:   "Remove a worktree",
	Long: `Remove a git worktree. Optionally delete the associated branch. If no path is provided, an interactive selection menu will appear.

Examples:
  wm remove feature-login
  wm remove feature-login -b          # Also delete branch
  wm remove -f feature-login          # Skip confirmation
  wm remove                           # Interactive mode`,
	Args: cobra.MaximumNArgs(1),
	RunE: runRemove,
}

func init() {
	removeCmd.Flags().BoolVarP(&removeForce, "force", "f", false, "Force removal without confirmation")
	removeCmd.Flags().BoolVarP(&removeDeleteBranch, "branch", "b", false, "Also delete the branch")
	rootCmd.AddCommand(removeCmd)
}

func runRemove(cmd *cobra.Command, args []string) error {
	var prompter ui.Prompter = ui.NewConsole()
	if removeForce {
		prompter = ui.NewSilent(true)
	}

	ws, err := workspace.Open(prompter)
	if err != nil {
		return err
	}

	var path string
	var deleteBranch bool

	// Interactive mode if no path provided
	if len(args) == 0 {
		path, deleteBranch, err = runRemoveInteractive(ws)
		if err != nil {
			return err
		}
	} else {
		path = args[0]
		deleteBranch = removeDeleteBranch
	}

	// Show removal spinner
	fmt.Println()
	spinner := ui.NewSpinner("Removing worktree...")
	spinner.Start()

	result, err := ws.RemoveWorktree(path, deleteBranch, removeForce)
	if err != nil {
		spinner.Fail("Failed to remove worktree")
		return err
	}

	spinner.Success(fmt.Sprintf("Removed worktree: %s", result.Path))

	// Handle branch deletion
	if deleteBranch && result.Branch != "" {
		branchSpinner := ui.NewSpinner(fmt.Sprintf("Deleting branch '%s'...", result.Branch))
		branchSpinner.Start()

		if result.BranchDeleted {
			branchSpinner.Success(fmt.Sprintf("Deleted branch: %s", result.Branch))
		} else {
			branchSpinner.Fail(fmt.Sprintf("Failed to delete branch '%s'", result.Branch))
			ui.Muted.Println("  Tip: Use 'git branch -D' to force delete.")
		}
	}

	fmt.Println()
	return nil
}

func runRemoveInteractive(ws *workspace.Workspace) (string, bool, error) {
	worktrees, err := ws.ListWorktrees()
	if err != nil {
		return "", false, err
	}

	if len(worktrees) <= 1 {
		ui.PrintWarning("No additional worktrees to remove.")
		return "", false, fmt.Errorf("no worktrees to remove")
	}

	// Build options
	options := make([]ui.SelectOption, 0, len(worktrees))
	for _, wt := range worktrees {
		label := wt.Path
		if wt.Branch != "" {
			label += " (" + wt.Branch + ")"
		}

		// Mark main worktree as disabled and current
		disabled := wt.Path == ws.Root
		if disabled {
			label += " [main]"
		}

		options = append(options, ui.SelectOption{
			Label:    label,
			Value:    wt.Path,
			Disabled: disabled,
		})
	}

	selectedPath, err := ui.Select("Select worktree to remove:", options)
	if err != nil {
		return "", false, err
	}

	// Ask if branch should be deleted
	deleteBranch := false
	if !removeForce {
		deleteBranch = ui.NewConsole().Confirm("Also delete branch?")
	}

	return selectedPath, deleteBranch, nil
}
