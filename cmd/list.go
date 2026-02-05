package cmd

import (
	"fmt"

	"github.com/Devdha/wm/internal/ui"
	"github.com/Devdha/wm/internal/workspace"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all worktrees",
	Long:    "List all git worktrees in the current repository with their branches and status.",
	RunE:    runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	ws, err := workspace.Open(ui.NewSilent(false))
	if err != nil {
		return err
	}

	worktrees, err := ws.ListWorktrees()
	if err != nil {
		return err
	}

	if len(worktrees) == 0 {
		ui.PrintInfo("No worktrees found.")
		return nil
	}

	table := ui.NewTable("PATH", "BRANCH", "HEAD")

	for _, wt := range worktrees {
		branch := wt.Branch
		if branch == "" {
			branch = ui.Mute("(detached)")
		}
		shortHead := wt.HEAD
		if len(shortHead) > 7 {
			shortHead = shortHead[:7]
		}
		table.AddRow(wt.Path, branch, shortHead)
	}

	fmt.Println()
	table.Render()
	fmt.Println()
	ui.Muted.Printf("Total: %d worktree(s)\n", len(worktrees))

	return nil
}
