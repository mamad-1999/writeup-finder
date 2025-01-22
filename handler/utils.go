package handler

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"writeup-finder.go/db"
	"writeup-finder.go/global"
	"writeup-finder.go/telegram"
	"writeup-finder.go/utils"
)

// isYouTubeFeed determines if a given URL corresponds to a YouTube RSS feed.
func IsYouTubeFeed(url string) bool {
	return strings.HasPrefix(url, "https://www.youtube.com/feeds/")
}

// IsNewArticle checks if an article is new by comparing its publication date and database presence.
func IsNewArticle(item *gofeed.Item, db *sql.DB, today time.Time) bool {
	pubDate, err := utils.ParseDate(item.Published)
	if err != nil {
		return false
	}

	yesterday := today.AddDate(0, 0, -1)
	isToday := pubDate.Format(global.DateFormat) == today.Format(global.DateFormat)
	isYesterday := pubDate.Format(global.DateFormat) == yesterday.Format(global.DateFormat)

	if !isToday && !isYesterday {
		return false
	}

	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM articles WHERE title = $1)"
	err = db.QueryRow(query, item.Title).Scan(&exists)
	utils.HandleError(err, "Error checking if article title exists in database", false)

	return !exists
}

// FormatArticleMessage creates a formatted string for an article's details.
func FormatArticleMessage(item *gofeed.Item) string {
	return fmt.Sprintf("\u25BA %s\nPublished: %s\nLink: %s", item.Title, item.Published, item.GUID)
}

// HandleArticle manages sending an article to Telegram and saving it to the database if enabled.
func HandleArticle(item *gofeed.Item, message string, database *sql.DB, isYoutube bool) error {
	if global.SendToTelegramFlag {
		fmt.Println("Start Send to Telegram...")

		telegram.SendToTelegram(message, global.ProxyURL, item.Title, isYoutube)
	}

	if global.UseDatabase {
		db.SaveUrlToDB(database, item.GUID, item.Title)
	}

	return nil
}
