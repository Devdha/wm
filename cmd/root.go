package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "wm",
	Short: "Git worktree manager",
	Long:  "WM is a CLI tool that makes git worktree easier to use with file sync and background tasks.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
