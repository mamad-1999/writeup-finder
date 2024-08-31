package rss

import (
	"time"

	"github.com/mmcdole/gofeed"
	"writeup-finder.go/utils"
)

func FetchArticles(feedUrl string) ([]*gofeed.Item, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(feedUrl)
	if err != nil {
		utils.HandleError(err, "Error fetching feed", false)
		return nil, err
	}
	return feed.Items, nil
}

func ParseDate(dateString string) (time.Time, error) {
	t, err := time.Parse(time.RFC1123Z, dateString)
	if err != nil {
		t, err = time.Parse(time.RFC1123, dateString)
	}
	if err != nil {
		utils.HandleError(err, "Error parsing date", false)
	}
	return t, err
}
