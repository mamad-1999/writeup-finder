package command

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// completionCmd generates shell autocompletion scripts.
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh]",
	Short: "Generate autocompletion script",
	Long: `To load completions:

Bash:

  $ source <(writeup-finder completion bash)

Zsh:

  $ source <(writeup-finder completion zsh)

  # To load completions for each session, execute once:
  # Linux:
  $ writeup-finder completion zsh > "${fpath[1]}/_writeup-finder"
  # macOS:
  $ writeup-finder completion zsh > /usr/local/share/zsh/site-functions/_writeup-finder
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			rootCmd.GenZshCompletion(os.Stdout)
		default:
			fmt.Println("Unsupported shell type. Please specify bash or zsh.")
		}
	},
}
