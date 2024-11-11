package url

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/fatih/color"
	"github.com/mmcdole/gofeed"
	"writeup-finder.go/db"
	"writeup-finder.go/global"
	"writeup-finder.go/rss"
	"writeup-finder.go/telegram"
	"writeup-finder.go/utils"
)

func ProcessUrls(urlList []string, today time.Time, database *sql.DB) int {
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
			if IsNewArticle(article, database, today) {
				message := FormatArticleMessage(article)
				if err := HandleArticle(article, message, database); err != nil {
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

func IsNewArticle(item *gofeed.Item, db *sql.DB, today time.Time) bool {
	// Parse the publication date of the article
	pubDate, err := rss.ParseDate(item.Published)
	if err != nil || pubDate.Format(global.DateFormat) != today.Format(global.DateFormat) {
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

func FormatArticleMessage(item *gofeed.Item) string {
	return fmt.Sprintf("▶ %s\nPublished: %s\nLink: %s", item.Title, item.Published, item.GUID)
}

func HandleArticle(item *gofeed.Item, message string, database *sql.DB) error {
	if global.SendToTelegramFlag {
		telegram.SendToTelegram(message, global.ProxyURL, item.Title)
	}

	if global.UseDatabase {
		db.SaveUrlToDB(database, item.GUID, item.Title)

	}

	return nil
}
