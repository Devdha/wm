package cmd

import (
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var noColor bool

var rootCmd = &cobra.Command{
	Use:   "wm",
	Short: "Git worktree manager",
	Long:  "WM is a CLI tool that makes git worktree easier to use with file sync and background tasks.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if noColor {
			color.NoColor = true
		}
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
