package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for rift.

To load completions:

Bash:
  $ source <(rift completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ rift completion bash > /etc/bash_completion.d/rift
  # macOS:
  $ rift completion bash > $(brew --prefix)/etc/bash_completion.d/rift

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc
  # To load completions for each session, execute once:
  $ rift completion zsh > "${fpath[1]}/_rift"
  # You will need to start a new shell for this setup to take effect.

Fish:
  $ rift completion fish | source
  # To load completions for each session, execute once:
  $ rift completion fish > ~/.config/fish/completions/rift.fish

PowerShell:
  PS> rift completion powershell | Out-String | Invoke-Expression
  # To load completions for every new session, add the output to your profile:
  PS> rift completion powershell > rift.ps1
`,
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
