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
)

var (
	useFile            bool
	useDatabase        bool
	sendToTelegramFlag bool
	proxyURL           string
)

const dataFolder = "data/"

func init() {
	config.LoadEnv()
}

func main() {
	flag.BoolVar(&useFile, "f", false, "Save new articles in found-url.txt")
	flag.BoolVar(&useDatabase, "d", false, "Save new articles in the database")
	flag.BoolVar(&sendToTelegramFlag, "t", false, "Send new articles to Telegram")
	flag.StringVar(&proxyURL, "proxy", "", "Proxy URL to use for sending Telegram messages")
	flag.Parse()

	if !useFile && !useDatabase {
		log.Fatal("You must specify either -f (file) or -d (database)")
	}

	if useFile && useDatabase {
		log.Fatal("Error: Please specify only one of -f (file) or -d (database), not both.")
	}

	if proxyURL != "" {
		if !sendToTelegramFlag {
			log.Fatal("Error: --proxy option is only valid when used with -t (send to Telegram).")
		}
		err := telegram.ValidateProxyURL(proxyURL)
		if err != nil {
			log.Fatalf("Error: Invalid proxy URL: %s", err)
		}
	}

	utils.PrintPretty("Starting Writeup Finder Script", color.FgHiYellow, true)

	urlList := utils.ReadUrls(dataFolder + "url.txt")
	today := time.Now()

	var foundUrls map[string]struct{}
	var database *sql.DB
	var err error

	if useFile {
		foundUrls = utils.ReadFoundUrlsFromFile(dataFolder + "found-url.json")
	} else if useDatabase {
		database, err = db.ConnectDB()

		utils.HandleError(err, "Error connecting to database", true)
		defer database.Close()
		foundUrls, err = db.ReadFoundUrlsFromDB(database)
		utils.HandleError(err, "Error reading found URLs from database", true)
	}

	articlesFound := 0

	for _, url := range urlList {
		utils.PrintPretty(fmt.Sprintf("Processing feed: %s", url), color.FgMagenta, false)
		articles, err := rss.FetchArticles(url)
		if err != nil {
			continue
		}

		for _, article := range articles {
			pubDate, err := rss.ParseDate(article.Published)
			if err != nil || pubDate.Format("2006-01-02") != today.Format("2006-01-02") {
				continue
			}

			if _, exists := foundUrls[article.GUID]; !exists {
				message := fmt.Sprintf("▶ %s\nPublished: %s\nLink: %s",
					article.Title, article.Published, article.GUID)

				if sendToTelegramFlag {
					telegram.SendToTelegram(message, proxyURL)
				}

				if useFile {
					utils.SaveUrlToFile(dataFolder+"found-url.json", article.Title, article.GUID)
				} else if useDatabase {
					db.SaveUrlToDB(database, article.GUID)
				}

				fmt.Println(color.GreenString(message))
				fmt.Println()
				articlesFound++
			}
		}
	}

	utils.PrintPretty(fmt.Sprintf("Total new articles found: %d", articlesFound), color.FgYellow, false)
	utils.PrintPretty("Writeup Finder Script Completed", color.FgHiYellow, true)
}
