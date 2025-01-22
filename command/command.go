package command

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"writeup-finder.go/db"
	"writeup-finder.go/global"
	"writeup-finder.go/handler"
	"writeup-finder.go/utils"
)

// Package command defines the CLI commands for the writeup-finder tool,
// including the root command and subcommands for completion.

// rootCmd is the main command for the writeup-finder CLI tool.
var rootCmd = &cobra.Command{
	Use:   "writeup-finder",
	Short: "A tool to find writeups and manage articles",
	Long:  `Writeup-finder is a tool to search for writeups and manage article data, including sending notifications.`,
	Run: func(cmd *cobra.Command, args []string) {
		if cmd.CalledAs() != "completion" {
			// Load environment variables and flags
			utils.LoadEnv()
			ManageFlags()
			utils.PrintPretty("Starting Writeup Finder Script", color.FgHiYellow, true)

			// Connect to the database if enabled
			if global.UseDatabase {
				global.DB = db.ConnectDB()
				db.CreateArticlesTable(global.DB)
				defer global.DB.Close()
			}

			// Execute main logic of the script
			ManageAction()
		}
	},
}

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

// Execute runs the root command, to be called in main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// init initializes global flags and subcommands.
func init() {
	rootCmd.PersistentFlags().BoolVar(&global.UseDatabase, "database", false, "Save new articles in the database")
	rootCmd.PersistentFlags().BoolVar(&global.SendToTelegramFlag, "telegram", false, "Send new articles to Telegram")
	rootCmd.PersistentFlags().StringVar(&global.ProxyURL, "proxy", "", "Proxy URL to use for sending Telegram messages")
	rootCmd.PersistentFlags().BoolVar(&global.Help, "help", false, "Show help")

	rootCmd.AddCommand(completionCmd)
}

// ManageFlags validates and logs the parsed flags.
func ManageFlags() {
	ValidateFlags()

	log.Infof("[+] Use Database: %v", global.UseDatabase)
	log.Infof("[+] Send to Telegram: %v", global.SendToTelegramFlag)

	if global.ProxyURL != "" {
		log.Infof("[+] Proxy URL: %v", global.ProxyURL)
	} else {
		log.Info("[+] No Proxy URL set.")
	}
}

// ValidateFlags ensures that flag combinations are valid and throws errors for invalid input.
func ValidateFlags() {
	if !global.UseDatabase {
		log.Fatal("You must specify --database to save articles in the database.")
	}

	if global.ProxyURL != "" && !global.SendToTelegramFlag {
		log.Fatal("Error: --proxy option is only valid when used with --telegram.")
	}
}

// ManageAction processes the list of URLs, finds new articles, and logs the results.
func ManageAction() {
	urlList := utils.ReadUrls(global.UrlFile)
	today := time.Now()

	// Process the URLs and store new articles in the database if enabled
	articlesFound := handler.ProcessUrls(urlList, today, global.DB)

	utils.PrintPretty(fmt.Sprintf("Total new articles found: %d", articlesFound), color.FgYellow, false)
	utils.PrintPretty("Writeup Finder Script Completed", color.FgHiYellow, true)
}
