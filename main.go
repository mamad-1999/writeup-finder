package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
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

	var foundUrls map[string]struct{}
	var database *sql.DB
	var err error

	if useFile {
		foundUrls = utils.ReadFoundUrlsFromFile(foundUrlFile)
	} else if useDatabase {
		database, err = db.ConnectDB()
		if err != nil {
			log.Fatalf("Error connecting to database: %v", err)
		}
		defer database.Close()

		foundUrls, err = db.ReadFoundUrlsFromDB(database)
		if err != nil {
			log.Fatalf("Error reading found URLs from database: %v", err)
		}
	}

	articlesFound := processUrls(urlList, foundUrls, today, database)

	utils.PrintPretty(fmt.Sprintf("Total new articles found: %d", articlesFound), color.FgYellow, false)
	utils.PrintPretty("Writeup Finder Script Completed", color.FgHiYellow, true)
}

func parseFlags() {
	flag.BoolVar(&useFile, "f", false, "Save new articles in found-url.txt")
	flag.BoolVar(&useDatabase, "d", false, "Save new articles in the database")
	flag.BoolVar(&sendToTelegramFlag, "t", false, "Send new articles to Telegram")
	flag.StringVar(&proxyURL, "proxy", "", "Proxy URL to use for sending Telegram messages")
	flag.Parse()
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

	if err := telegram.ValidateProxyURL(proxyURL); err != nil {
		log.Fatalf("Error: Invalid proxy URL: %s", err)
	}
}

func processUrls(urlList []string, foundUrls map[string]struct{}, today time.Time, database *sql.DB) int {
	articlesFound := 0

	for _, url := range urlList {
		utils.PrintPretty(fmt.Sprintf("Processing feed: %s", url), color.FgMagenta, false)
		articles, err := rss.FetchArticles(url)
		if err != nil {
			log.Printf("Error fetching articles from %s: %v", url, err)
			continue
		}

		for _, article := range articles {
			if isNewArticle(article, foundUrls, today) {
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

func isNewArticle(item *gofeed.Item, foundUrls map[string]struct{}, today time.Time) bool {
	pubDate, err := rss.ParseDate(item.Published)
	if err != nil || pubDate.Format(dateFormat) != today.Format(dateFormat) {
		return false
	}

	_, exists := foundUrls[item.GUID]
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
		db.SaveUrlToDB(database, item.GUID)

	}

	return nil
}
