package commands

import (
	"os"

	"github.com/spf13/cobra"
)

// ajfs completion.
var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Generate shell completion scripts.",
	Long: `Generate shell completion scripts.

Zsh:
  Generate the script.

  $ mkdir -p ~/.zsh/completion
  $ ajfs completion zsh > ~/.zsh/completion/_ajfs

  Ensure shell completion is enabled for your environment and load the
  completion script. Add the following line to your ~/.zshrc file:

    fpath=(~/.zsh/completion $fpath)
    autoload -U compinit; compinit

  Restart your terminal.

Bash:
  Generate the script.
  
  $ mkdir -p ~/.local/share/bash-completion/completions
  $ ajfs completion bash > ~/.local/share/bash-completion/completions/ajfs.bash

  Restart your terminal or reload configuration: source ~/.bashrc

Fish:
  Generate the script.
  
  $ mkdir -p ~/.config/fish/completions
  $ ajfs completion fish > ~/.config/fish/completions/ajfs.fish

  Restart your terminal or reload the configuration: source ~/.config/fish/config.fish

PowerShell:
  Generate the script.

  PS> mkdir -p (Split-Path $PROFILE)
  PS> ./ajfs completion powershell > ajfs.ps1
  PS> Add-Content $PROFILE ". /path/to/ajfs.ps1"

  Restart your terminal or reload the shell: & $PROFILE
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"zsh", "bash", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
