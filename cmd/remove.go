package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/donghun/wm/internal/git"
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
	wtPath := args[0]

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	repoRoot, err := git.GetRepoRoot(cwd)
	if err != nil {
		return err
	}

	// Find the worktree to get branch info
	worktrees, err := git.ListWorktrees(repoRoot)
	if err != nil {
		return err
	}

	var targetWT *git.Worktree
	for i, wt := range worktrees {
		if wt.Path == wtPath || strings.HasSuffix(wt.Path, "/"+wtPath) {
			targetWT = &worktrees[i]
			wtPath = wt.Path // Use full path
			break
		}
	}

	if targetWT == nil {
		return fmt.Errorf("worktree '%s' not found", wtPath)
	}

	// Prevent removing main worktree
	if targetWT.Path == repoRoot {
		return fmt.Errorf("cannot remove the main worktree")
	}

	// Confirmation prompt
	if !removeForce {
		fmt.Printf("Remove worktree at '%s'", wtPath)
		if removeDeleteBranch && targetWT.Branch != "" {
			fmt.Printf(" and branch '%s'", targetWT.Branch)
		}
		fmt.Print("? [y/N]: ")

		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))

		if answer != "y" && answer != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	// Check if branch is used by another worktree
	if removeDeleteBranch && targetWT.Branch != "" {
		for _, wt := range worktrees {
			if wt.Path != wtPath && wt.Branch == targetWT.Branch {
				return fmt.Errorf("cannot delete branch '%s': used by worktree at '%s'", targetWT.Branch, wt.Path)
			}
		}
	}

	// Remove worktree
	fmt.Printf("Removing worktree...")
	if err := git.RemoveWorktree(repoRoot, wtPath, removeForce); err != nil {
		return err
	}
	fmt.Println(" done.")

	// Delete branch if requested
	if removeDeleteBranch && targetWT.Branch != "" {
		fmt.Printf("Deleting branch '%s'...", targetWT.Branch)
		if err := git.DeleteBranch(repoRoot, targetWT.Branch, false); err != nil {
			fmt.Printf(" failed: %v\n", err)
			fmt.Println("Tip: Use 'git branch -D' to force delete.")
		} else {
			fmt.Println(" done.")
		}
	}

	return nil
}
