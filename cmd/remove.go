package cmd

import (
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

	// Interactive mode if no path provided
	if len(args) == 0 {
		return runRemoveInteractive(ws)
	}

	return ws.RemoveWorktree(args[0], removeDeleteBranch, removeForce)
}

func runRemoveInteractive(ws *workspace.Workspace) error {
	worktrees, err := ws.ListWorktrees()
	if err != nil {
		return err
	}

	if len(worktrees) <= 1 {
		ui.PrintWarning("No additional worktrees to remove.")
		return nil
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
		return err
	}

	// Ask if branch should be deleted
	deleteBranch := false
	if !removeForce {
		deleteBranch = ui.NewConsole().Confirm("Also delete branch?")
	}

	return ws.RemoveWorktree(selectedPath, deleteBranch, removeForce)
}
