package command

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"writeup-finder.go/db"
	"writeup-finder.go/global"
	"writeup-finder.go/utils"
)

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
