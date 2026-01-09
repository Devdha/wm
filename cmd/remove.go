package cmd

import (
	"github.com/donghun/wm/internal/ui"
	"github.com/donghun/wm/internal/workspace"
	"github.com/spf13/cobra"
)

var (
	removeForce        bool
	removeDeleteBranch bool
)

var removeCmd = &cobra.Command{
	Use:     "remove <path>",
	Aliases: []string{"rm"},
	Short:   "Remove a worktree",
	Long:    "Remove a git worktree. Optionally delete the associated branch.",
	Args:    cobra.ExactArgs(1),
	RunE:    runRemove,
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
	return ws.RemoveWorktree(args[0], removeDeleteBranch, removeForce)
}
