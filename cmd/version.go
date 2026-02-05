package cmd

import (
	"github.com/Devdha/wm/internal/ui"
	"github.com/Devdha/wm/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		ui.Primary.Print("wm ")
		ui.Bold.Println(version.String())
		ui.Muted.Println("Git Worktree Manager")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
