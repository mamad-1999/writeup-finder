package rss

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mmcdole/gofeed"
	"writeup-finder.go/utils"
)

// FetchArticles retrieves articles from the given feed URL.
func FetchArticles(feedURL string) ([]*gofeed.Item, error) {
	// Create a custom HTTP client with the User-Agent header
	client := &http.Client{
		Timeout: time.Second * 1, // Set a timeout for the request
	}

	// Create a new request
	req, err := http.NewRequest("GET", feedURL, nil)
	if err != nil {
		utils.HandleError(err, "Error creating request", false)
		return nil, err
	}

	// Set a valid, commonly accepted User-Agent header for Linux (Ubuntu/Firefox)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:91.0) Gecko/20100101 Firefox/91.0")

	// Use the custom HTTP client to fetch the feed
	resp, err := client.Do(req)
	if err != nil {
		utils.HandleError(err, "Error fetching feed", false)
		return nil, err // Return early to avoid nil dereference
	}
	defer resp.Body.Close()

	parser := gofeed.NewParser()
	feed, err := parser.Parse(resp.Body)

	// Handle error when fetching feed fails
	if err != nil {
		utils.HandleError(err, "Error parsing feed", false)
		return nil, err // Return early to avoid nil dereference
	}

	// Check if feed is nil
	if feed == nil {
		return nil, fmt.Errorf("no feed data received from URL: %s", feedURL)
	}

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
