package handler

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/fatih/color"
	"github.com/mmcdole/gofeed"
	"writeup-finder.go/global"
	"writeup-finder.go/utils"
)

// processYouTubeFeed fetches and processes videos from a YouTube RSS feed.
func ProcessYouTubeFeed(url string, today time.Time, database *sql.DB) int {
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
