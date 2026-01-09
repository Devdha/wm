package cmd

import (
	"github.com/Devdha/wm/internal/ui"
	"github.com/Devdha/wm/internal/workspace"
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
	ws, err := workspace.Open(ui.NewConsole())
	if err != nil {
		return err
	}
	return ws.AddWorktree(args[0], addPath)
}
