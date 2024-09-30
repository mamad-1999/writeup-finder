package rss

import (
	"time"

	"github.com/mmcdole/gofeed"
	"writeup-finder.go/utils"
)

// FetchArticles retrieves articles from the given feed URL.
func FetchArticles(feedURL string) ([]*gofeed.Item, error) {
	parser := gofeed.NewParser()
	feed, err := parser.ParseURL(feedURL)
	utils.HandleError(err, "Error fetching feed", false)
	return feed.Items, nil
}

// ParseDate parses a date string and returns the corresponding time.Time object.
func ParseDate(dateString string) (time.Time, error) {
	parsedTime, err := time.Parse(time.RFC1123Z, dateString)
	if err != nil {
		parsedTime, err = time.Parse(time.RFC1123, dateString)
	}
	utils.HandleError(err, "Error parsing date", false)
	return parsedTime, nil
}
