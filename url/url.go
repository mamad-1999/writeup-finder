package url

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
	"writeup-finder.go/rss"
	"writeup-finder.go/telegram"
	"writeup-finder.go/utils"
)

// ProcessUrls handles both Medium and YouTube feeds
func ProcessUrls(urlList []string, today time.Time, database *sql.DB) int {
	articlesFound := 0

	for i, url := range urlList {
		utils.PrintPretty(fmt.Sprintf("Processing feed: %s", url), color.FgMagenta, false)

		// Check if the URL is a YouTube RSS feed
		if isYouTubeFeed(url) {
			videosFound := processYouTubeFeed(url, today, database)
			articlesFound += videosFound
		} else {
			articlesFound += processMediumFeed(url, today, database)
		}

		// Introduce a delay of 2 seconds between processing URLs, except for the last one
		if i < len(urlList)-1 {
			time.Sleep(3 * time.Second)
		}
	}

	return articlesFound
}

// processMediumFeed processes Medium RSS feeds
func processMediumFeed(url string, today time.Time, database *sql.DB) int {
	articlesFound := 0
	articles, err := rss.FetchArticles(url)
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
			fmt.Println()
			articlesFound++
		}
	}
	return articlesFound
}

// processYouTubeFeed processes YouTube RSS feeds
func processYouTubeFeed(url string, today time.Time, database *sql.DB) int {
	articlesFound := 0
	feedParser := gofeed.NewParser()

	// Fetch and parse the YouTube RSS feed
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

		// Check if the video is new and was published today or yesterday
		yesterday := today.AddDate(0, 0, -1)
		isToday := pubDate.Format(global.DateFormat) == today.Format(global.DateFormat)
		isYesterday := pubDate.Format(global.DateFormat) == yesterday.Format(global.DateFormat)

		if !isToday && !isYesterday {
			continue
		}

		// Check if the video link already exists in the database
		var exists bool
		query := "SELECT EXISTS(SELECT 1 FROM articles WHERE url = $1)"
		err = database.QueryRow(query, item.Link).Scan(&exists)
		utils.HandleError(err, "Error checking if YouTube video link exists in database", false)

		if exists {
			continue
		}

		article := &gofeed.Item{
			GUID:      item.Link, // Use the YouTube video link as the GUID
			Title:     item.Title,
			Published: item.Published,
		}
		message := FormatArticleMessage(article)

		if err := HandleArticle(article, message, database, true); err != nil {
			log.Printf("Error handling YouTube video %s: %v", item.Link, err)
			continue
		}
		fmt.Println(color.GreenString(message))
		fmt.Println()

		articlesFound++
	}
	return articlesFound
}

// isYouTubeFeed checks if the URL is a YouTube RSS feed
func isYouTubeFeed(url string) bool {
	return strings.HasPrefix(url, "https://www.youtube.com/feeds/")
}

func IsNewArticle(item *gofeed.Item, db *sql.DB, today time.Time) bool {
	// Parse the publication date of the article
	pubDate, err := rss.ParseDate(item.Published)
	if err != nil {
		// If parsing fails, treat it as not a new article
		return false
	}

	// Check if the publication date matches today or yesterday
	yesterday := today.AddDate(0, 0, -1)
	isToday := pubDate.Format(global.DateFormat) == today.Format(global.DateFormat)
	isYesterday := pubDate.Format(global.DateFormat) == yesterday.Format(global.DateFormat)

	if !isToday && !isYesterday {
		// Article is neither from today nor yesterday
		return false
	}

	// Query the database to check if the title already exists
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM articles WHERE title = $1)"
	err = db.QueryRow(query, item.Title).Scan(&exists)
	utils.HandleError(err, "Error checking if article title exists in database", false)

	// Return true only if the article does not exist in the database
	return !exists
}

func FormatArticleMessage(item *gofeed.Item) string {
	return fmt.Sprintf("▶ %s\nPublished: %s\nLink: %s", item.Title, item.Published, item.GUID)
}

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
