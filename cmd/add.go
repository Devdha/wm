package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/Devdha/wm/internal/git"
	"github.com/Devdha/wm/internal/ui"
	"github.com/Devdha/wm/internal/workspace"
	"github.com/spf13/cobra"
)

var (
	addPath string
	addYes  bool
)

var addCmd = &cobra.Command{
	Use:   "add [branch]",
	Short: "Create a new worktree",
	Long: `Create a new git worktree with file sync and optional background tasks. If no branch is provided, an interactive selection menu will appear.

Examples:
  wm add feature-login
  wm add feature-login -p ~/custom/path
  wm add                              # Interactive mode`,
	Args: cobra.MaximumNArgs(1),
	RunE: runAdd,
}

func init() {
	addCmd.Flags().StringVarP(&addPath, "path", "p", "", "Custom path for the worktree")
	addCmd.Flags().BoolVarP(&addYes, "yes", "y", false, "Auto-confirm prompts (useful for CI/CD)")
	rootCmd.AddCommand(addCmd)
}

func isInteractive() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func runAdd(cmd *cobra.Command, args []string) error {
	var prompter ui.Prompter = ui.NewConsole()
	if addYes || !isInteractive() {
		prompter = ui.NewSilent(true)
	}

	ws, err := workspace.Open(prompter)
	if err != nil {
		return err
	}

	var branch string
	var customPath string

	if len(args) == 0 {
		if !isInteractive() {
			return fmt.Errorf("branch name required in non-interactive mode")
		}
		branch, customPath, err = runAddInteractive(ws)
		if err != nil {
			return err
		}
	} else {
		branch = args[0]
		customPath = addPath

		if !git.BranchExists(ws.Root, branch) && git.RemoteBranchExists(ws.Root, branch) {
			if err := handleRemoteBranch(ws, branch); err != nil {
				return err
			}
		}
	}

	result, err := ws.AddWorktree(branch, customPath)
	if err != nil {
		return err
	}

	printAddResult(result)

	if isInteractive() || addYes {
		if prompter.ConfirmYes("Navigate to worktree directory?") {
			return openShellAt(result.Path)
		}
	}

	return nil
}

func printAddResult(result *workspace.AddResult) {
	fmt.Println()
	ui.PrintStep(ui.IconBolt, "Creating worktree...")
	ui.PrintSubStep(fmt.Sprintf("Branch: %s", result.Branch))
	ui.PrintSubStepEnd(fmt.Sprintf("Path: %s", result.Path))
	fmt.Println()

	if len(result.SyncedFiles) > 0 {
		ui.PrintStep(ui.IconPackage, "Syncing files...")
		for _, file := range result.SyncedFiles {
			ui.PrintSubStep(fmt.Sprintf("%s %s", file, ui.Success.Sprint(ui.IconCheck)))
		}
		fmt.Println()
	}

	if len(result.PostInstall) > 0 {
		ui.PrintStep(ui.IconRocket, "Running post-install...")
		for _, c := range result.PostInstall {
			suffix := ""
			if result.IsBackground {
				suffix = " " + ui.Muted.Sprint("(background)")
			}
			ui.PrintSubStep(c + suffix)
		}
		fmt.Println()
	}

	fmt.Println()
	ui.Success.Print(ui.IconCheck + " ")
	ui.Bold.Print("Worktree ready: ")
	fmt.Println(result.Path)
	fmt.Println()
}

func runAddInteractive(ws *workspace.Workspace) (string, string, error) {
	// Get remote branches
	remoteBranches, err := git.ListRemoteBranches(ws.Root)
	if err != nil {
		remoteBranches = []string{} // Continue even if remote fetch fails
	}

	// Handle empty remote branches with a helpful placeholder
	if len(remoteBranches) == 0 {
		remoteBranches = []string{"(No remote branches found - type a branch name above)"}
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

func detectShell() string {
	if shell := os.Getenv("SHELL"); shell != "" {
		return shell
	}
	if runtime.GOOS == "windows" {
		if comspec := os.Getenv("COMSPEC"); comspec != "" {
			return comspec
		}
		return "cmd.exe"
	}
	return "/bin/sh"
}

func openShellAt(dir string) error {
	shell := detectShell()

	fmt.Println()
	ui.Muted.Printf("  Starting shell in %s\n", dir)
	ui.Muted.Println("  Type 'exit' to return to the previous directory.")
	fmt.Println()

	cmd := exec.Command(shell)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
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
