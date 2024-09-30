package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"writeup-finder.go/config"
	"writeup-finder.go/db"
	"writeup-finder.go/rss"
	"writeup-finder.go/telegram"
	"writeup-finder.go/utils"

	"github.com/fatih/color"
	"github.com/mmcdole/gofeed"
)

const (
	dataFolder   = "data/"
	urlFile      = dataFolder + "url.txt"
	foundUrlFile = dataFolder + "found-url.json"
	dateFormat   = "2006-01-02"
)

var (
	useFile            bool
	useDatabase        bool
	sendToTelegramFlag bool
	proxyURL           string
	help               bool
)

func init() {
	config.LoadEnv()
}

func main() {
	parseFlags()
	validateFlags()

	utils.PrintPretty("Starting Writeup Finder Script", color.FgHiYellow, true)

	urlList := utils.ReadUrls(urlFile)
	today := time.Now()

	var foundTitles map[string]struct{}
	var database *sql.DB
	var err error

	if useFile {
		foundTitles = utils.ReadFoundUrlsFromFile(foundUrlFile)
	} else if useDatabase {
		database, err = db.ConnectDB()
		db.CreateArticlesTable(database)
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}
		defer database.Close()

		foundTitles, err = db.ReadFoundTitlesFromDB(database)
		if err != nil {
			log.Fatalf("Error reading found URLs from database: %v", err)
		}
	}

	articlesFound := processUrls(urlList, foundTitles, today, database)

	utils.PrintPretty(fmt.Sprintf("Total new articles found: %d", articlesFound), color.FgYellow, false)
	utils.PrintPretty("Writeup Finder Script Completed", color.FgHiYellow, true)
}

func parseFlags() {
	flag.BoolVar(&useFile, "f", false, "Save new articles in found-url.json")
	flag.BoolVar(&useFile, "file", false, "Save new articles in found-url.json")
	flag.BoolVar(&useDatabase, "d", false, "Save new articles in the database")
	flag.BoolVar(&useDatabase, "database", false, "Save new articles in the database")
	flag.BoolVar(&sendToTelegramFlag, "t", false, "Send new articles to Telegram")
	flag.BoolVar(&sendToTelegramFlag, "telegram", false, "Send new articles to Telegram")
	flag.StringVar(&proxyURL, "proxy", "", "Proxy URL to use for sending Telegram messages")
	flag.BoolVar(&help, "h", false, "Show help")
	flag.BoolVar(&help, "help", false, "Show help")
	flag.Parse()

	if help {
		printHelp()
		os.Exit(0) // Exit after printing help
	}
}

func printHelp() {
	fmt.Println("Usage: writeup-finder [OPTIONS]")
	fmt.Println("\nOptions:")
	fmt.Println("  -f            Save new articles in found-url.txt")
	fmt.Println("  -d            Save new articles in the database")
	fmt.Println("  -t            Send new articles to Telegram")
	fmt.Println("  --proxy URL   Proxy URL to use for sending Telegram messages (only valid with -t)")
	fmt.Println("  -h, --help    Show this help message")
	fmt.Println("\nNote: You must specify either -f (file) or -d (database), but not both.")
	fmt.Println("      The --proxy option is only valid when using the -t (Telegram) flag.")
}

func validateFlags() {
	if !useFile && !useDatabase {
		log.Fatal("You must specify either -f (file) or -d (database)")
	}

	if useFile && useDatabase {
		log.Fatal("Error: Please specify only one of -f (file) or -d (database), not both.")
	}

	if proxyURL != "" && !sendToTelegramFlag {
		log.Fatal("Error: --proxy option is only valid when used with -t (send to Telegram).")
	}
}

func processUrls(urlList []string, foundTitles map[string]struct{}, today time.Time, database *sql.DB) int {
	articlesFound := 0

	for _, url := range urlList {
		utils.PrintPretty(fmt.Sprintf("Processing feed: %s", url), color.FgMagenta, false)
		articles, err := rss.FetchArticles(url)
		if err != nil {
			log.Printf("Error fetching articles from %s: %v", url, err)
			continue
		}

		for _, article := range articles {
			// Check if the article's title is in foundTitles (instead of URL)
			if isNewArticle(article, foundTitles, today) {
				message := formatArticleMessage(article)
				if err := handleArticle(article, message, database); err != nil {
					log.Printf("Error handling article %s: %v", article.GUID, err)
					continue
				}
				fmt.Println(color.GreenString(message))
				fmt.Println()
				articlesFound++
			}
		}
	}

	return articlesFound
}

func isNewArticle(item *gofeed.Item, foundTitles map[string]struct{}, today time.Time) bool {
	// Parse the publication date of the article
	pubDate, err := rss.ParseDate(item.Published)
	if err != nil || pubDate.Format(dateFormat) != today.Format(dateFormat) {
		return false
	}

	// Check if the title already exists in foundTitles
	_, exists := foundTitles[item.Title]
	return !exists
}

func formatArticleMessage(item *gofeed.Item) string {
	return fmt.Sprintf("▶ %s\nPublished: %s\nLink: %s", item.Title, item.Published, item.GUID)
}

func handleArticle(item *gofeed.Item, message string, database *sql.DB) error {
	if sendToTelegramFlag {
		telegram.SendToTelegram(message, proxyURL)
	}
	if useFile {
		utils.SaveUrlToFile(foundUrlFile, item.Title, item.GUID)

	} else if useDatabase {
		db.SaveUrlToDB(database, item.GUID, item.Title)

	}

	return nil
}
