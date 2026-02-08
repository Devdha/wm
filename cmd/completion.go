package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `Generate shell completion script for wm.

To load completions:

Bash:
  $ source <(wm completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ wm completion bash > /etc/bash_completion.d/wm
  # macOS:
  $ wm completion bash > $(brew --prefix)/etc/bash_completion.d/wm

Zsh:
  $ source <(wm completion zsh)
  # To load completions for each session, execute once:
  $ wm completion zsh > "${fpath[1]}/_wm"

Fish:
  $ wm completion fish | source
  # To load completions for each session, execute once:
  $ wm completion fish > ~/.config/fish/completions/wm.fish

PowerShell:
  PS> wm completion powershell | Out-String | Invoke-Expression`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
