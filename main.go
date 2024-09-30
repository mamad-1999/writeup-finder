package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"time"

	"writeup-finder.go/config"
	"writeup-finder.go/db"
	"writeup-finder.go/rss"
	"writeup-finder.go/telegram"
	"writeup-finder.go/utils"

	"github.com/fatih/color"
	"github.com/mmcdole/gofeed"
	log "github.com/sirupsen/logrus"
)

const (
	dataFolder = "data/"
	urlFile    = dataFolder + "url.txt"
	dateFormat = "2006-01-02"
)

var (
	useDatabase        bool
	sendToTelegramFlag bool
	proxyURL           string
	help               bool
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp:       true,
		DisableLevelTruncation: true,
		FullTimestamp:          false,
	})
	config.LoadEnv()
}

func main() {
	parseFlags()
	validateFlags()

	utils.PrintPretty("Starting Writeup Finder Script", color.FgHiYellow, true)

	urlList := utils.ReadUrls(urlFile)
	today := time.Now()

	var database *sql.DB

	if useDatabase {
		database = db.ConnectDB()
		db.CreateArticlesTable(database)
		defer database.Close()
	}

	articlesFound := processUrls(urlList, today, database)

	utils.PrintPretty(fmt.Sprintf("Total new articles found: %d", articlesFound), color.FgYellow, false)
	utils.PrintPretty("Writeup Finder Script Completed", color.FgHiYellow, true)
}

func parseFlags() {
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
	fmt.Println("  -d            Save new articles in the database")
	fmt.Println("  -t            Send new articles to Telegram")
	fmt.Println("  --proxy URL   Proxy URL to use for sending Telegram messages (only valid with -t)")
	fmt.Println("  -h, --help    Show this help message")
	fmt.Println("\nNote: You must specify either -f (file) or -d (database), but not both.")
	fmt.Println("      The --proxy option is only valid when using the -t (Telegram) flag.")
}

func validateFlags() {
	if !useDatabase {
		log.Fatal("You must specify -d (database)")
	}

	if proxyURL != "" && !sendToTelegramFlag {
		log.Fatal("Error: --proxy option is only valid when used with -t (send to Telegram).")
	}
}

func processUrls(urlList []string, today time.Time, database *sql.DB) int {
	articlesFound := 0

	for _, url := range urlList {
		utils.PrintPretty(fmt.Sprintf("Processing feed: %s", url), color.FgMagenta, false)
		articles, err := rss.FetchArticles(url)
		if err != nil {
			log.Printf("Error fetching articles from %s: %v", url, err)
			continue
		}

		for _, article := range articles {
			// Check if the article's title is new and if it was published today
			if isNewArticle(article, database, today) {
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

func isNewArticle(item *gofeed.Item, db *sql.DB, today time.Time) bool {
	// Parse the publication date of the article
	pubDate, err := rss.ParseDate(item.Published)
	if err != nil || pubDate.Format(dateFormat) != today.Format(dateFormat) {
		// Article is not from today
		return false
	}

	// Query the database to check if the title already exists
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM articles WHERE title = $1)"
	err = db.QueryRow(query, item.Title).Scan(&exists)
	utils.HandleError(err, "Error checking if article title exists in database", false)

	// Return true only if the article is from today and does not exist in the database
	return !exists
}

func formatArticleMessage(item *gofeed.Item) string {
	return fmt.Sprintf("▶ %s\nPublished: %s\nLink: %s", item.Title, item.Published, item.GUID)
}

func handleArticle(item *gofeed.Item, message string, database *sql.DB) error {
	if sendToTelegramFlag {
		telegram.SendToTelegram(message, proxyURL)
	}

	if useDatabase {
		db.SaveUrlToDB(database, item.GUID, item.Title)

	}

	return nil
}
