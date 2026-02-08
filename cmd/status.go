package cmd

import (
	"fmt"
	"os"

	"github.com/Devdha/wm/internal/git"
	"github.com/Devdha/wm/internal/ui"
	"github.com/Devdha/wm/internal/workspace"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"st"},
	Short:   "Show status of all worktrees",
	Long: `Show git status for all worktrees in the current repository.

Examples:
  wm status
  wm st`,
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
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

	currentDir, _ := os.Getwd()

	ui.PrintStep(ui.IconBolt, "Worktree Status")
	fmt.Println()

	for _, wt := range worktrees {
		branch := wt.Branch
		if branch == "" {
			branch = ui.Mute("detached")
		}

		isCurrent := ""
		if wt.Path == currentDir {
			isCurrent = ui.Accent.Sprint(" [current]")
		}

		fmt.Printf("  %s (%s)%s\n", wt.Path, branch, isCurrent)

		modified, untracked, err := git.WorktreeStatus(wt.Path)
		if err != nil {
			ui.Muted.Printf("    %s Unable to get status\n", ui.IconWarning)
		} else {
			if modified == 0 && untracked == 0 {
				ui.Success.Printf("    %s Clean\n", ui.IconCheck)
			} else {
				statusParts := []string{}
				if modified > 0 {
					statusParts = append(statusParts, fmt.Sprintf("%d modified", modified))
				}
				if untracked > 0 {
					statusParts = append(statusParts, fmt.Sprintf("%d untracked", untracked))
				}
				ui.Error.Printf("    %s %s\n", ui.IconCross, statusParts[0])
				if len(statusParts) > 1 {
					ui.Error.Printf("       %s\n", statusParts[1])
				}
			}
		}

		ahead, behind, err := git.WorktreeAheadBehind(wt.Path)
		if err == nil && (ahead > 0 || behind > 0) {
			if ahead > 0 {
				ui.Info.Printf("    ↑%d ahead of origin\n", ahead)
			}
			if behind > 0 {
				ui.Warning.Printf("    ↓%d behind origin\n", behind)
			}
		}

		fmt.Println()
	}

	return nil
}
