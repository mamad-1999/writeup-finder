package command

import (
	"flag"
	"fmt"
	"log"
	"os"

	"writeup-finder.go/global"
)

func ManageFlags() {
	ParseFlags()
	ValidateFlags()
}

func ParseFlags() {
	flag.BoolVar(&global.UseDatabase, "d", false, "Save new articles in the database")
	flag.BoolVar(&global.UseDatabase, "database", false, "Save new articles in the database")
	flag.BoolVar(&global.SendToTelegramFlag, "t", false, "Send new articles to Telegram")
	flag.BoolVar(&global.SendToTelegramFlag, "telegram", false, "Send new articles to Telegram")
	flag.StringVar(&global.ProxyURL, "proxy", "", "Proxy URL to use for sending Telegram messages")
	flag.BoolVar(&global.Help, "h", false, "Show help")
	flag.BoolVar(&global.Help, "help", false, "Show help")
	flag.Parse()

	if global.Help {
		PrintHelp()
		os.Exit(0) // Exit after printing help
	}
}

func PrintHelp() {
	fmt.Println("Usage: writeup-finder [OPTIONS]")
	fmt.Println("\nOptions:")
	fmt.Println("  -d            Save new articles in the database")
	fmt.Println("  -t            Send new articles to Telegram")
	fmt.Println("  --proxy URL   Proxy URL to use for sending Telegram messages (only valid with -t)")
	fmt.Println("  -h, --help    Show this help message")
	fmt.Println("\nNote: You must specify either -f (file) or -d (database), but not both.")
	fmt.Println("      The --proxy option is only valid when using the -t (Telegram) flag.")
}

func ValidateFlags() {
	if !global.UseDatabase {
		log.Fatal("You must specify -d (database)")
	}

	if global.ProxyURL != "" && !global.SendToTelegramFlag {
		log.Fatal("Error: --proxy option is only valid when used with -t (send to Telegram).")
	}
}
