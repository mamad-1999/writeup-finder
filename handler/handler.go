package handler

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/mmcdole/gofeed"
	"writeup-finder.go/db"
	"writeup-finder.go/global"
	"writeup-finder.go/telegram"
	"writeup-finder.go/utils"
)

// ProcessUrls iterates over a list of URLs and processes each one based on its type (Medium or YouTube).
func ProcessUrls(urlList []string, today time.Time, database *sql.DB) int {
	articlesFound := 0

	for i, url := range urlList {
		utils.PrintPretty(fmt.Sprintf("Processing feed: %s", url), color.FgMagenta, false)

		// Determine the type of feed and process accordingly
		if isYouTubeFeed(url) {
			videosFound := processYouTubeFeed(url, today, database)
			articlesFound += videosFound
		} else {
			articlesFound += processMediumFeed(url, today, database)
		}

		// Delay processing of the next URL to prevent rate-limiting or server overload
		if i < len(urlList)-1 {
			time.Sleep(3 * time.Second)
		}
	}

	return articlesFound
}

// processMediumFeed fetches and processes articles from a Medium RSS feed.
func processMediumFeed(url string, today time.Time, database *sql.DB) int {
	articlesFound := 0
	articles, err := utils.FetchArticles(url)
	if err != nil {
		log.Printf("Error fetching articles from %s: %v", url, err)
		return 0
	}

	for _, article := range articles {
		if IsNewArticle(article, database, today) {
			message := FormatArticleMessage(article)
			if err := HandleArticle(article, message, database, false); err != nil {
				log.Printf("Error handling article %s: %v", article.GUID, err)
				continue
			}
			fmt.Println(color.GreenString(message))
			articlesFound++
		}
	}
	return articlesFound
}

// processYouTubeFeed fetches and processes videos from a YouTube RSS feed.
func processYouTubeFeed(url string, today time.Time, database *sql.DB) int {
	articlesFound := 0
	feedParser := gofeed.NewParser()

	feed, err := feedParser.ParseURL(url)
	if err != nil {
		log.Printf("Error fetching YouTube feed from %s: %v", url, err)
		return 0
	}

	for _, item := range feed.Items {
		pubDate, err := time.Parse(time.RFC3339, item.Published)
		if err != nil {
			log.Printf("Error parsing publication date for YouTube video: %v", err)
			continue
		}

		// Determine if the video is new based on publication date
		yesterday := today.AddDate(0, 0, -1)
		isToday := pubDate.Format(global.DateFormat) == today.Format(global.DateFormat)
		isYesterday := pubDate.Format(global.DateFormat) == yesterday.Format(global.DateFormat)

		if !isToday && !isYesterday {
			continue
		}

		// Check for the video's existence in the database
		var exists bool
		query := "SELECT EXISTS(SELECT 1 FROM articles WHERE url = $1)"
		err = database.QueryRow(query, item.Link).Scan(&exists)
		utils.HandleError(err, "Error checking if YouTube video link exists in database", false)

		if exists {
			continue
		}

		article := &gofeed.Item{
			GUID:      item.Link,
			Title:     item.Title,
			Published: item.Published,
		}
		message := FormatArticleMessage(article)

		if err := HandleArticle(article, message, database, true); err != nil {
			log.Printf("Error handling YouTube video %s: %v", item.Link, err)
			continue
		}
		fmt.Println(color.GreenString(message))
		articlesFound++
	}
	return articlesFound
}

// isYouTubeFeed determines if a given URL corresponds to a YouTube RSS feed.
func isYouTubeFeed(url string) bool {
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
	fmt.Println("Start Send to Telegram...")

	if global.SendToTelegramFlag {
		telegram.SendToTelegram(message, global.ProxyURL, item.Title, isYoutube)
	}

	if global.UseDatabase {
		db.SaveUrlToDB(database, item.GUID, item.Title)
	}

	return nil
}
