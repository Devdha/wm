package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/donghun/wm/internal/ui"
	"github.com/donghun/wm/internal/workspace"
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
		fmt.Println("No worktrees found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PATH\tBRANCH\tHEAD")
	fmt.Fprintln(w, "----\t------\t----")

	for _, wt := range worktrees {
		branch := wt.Branch
		if branch == "" {
			branch = "(detached)"
		}
		shortHead := wt.HEAD
		if len(shortHead) > 7 {
			shortHead = shortHead[:7]
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", wt.Path, branch, shortHead)
	}

	w.Flush()
	return nil
}
