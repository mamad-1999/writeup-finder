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
	// Create a custom HTTP client with a timeout
	client := &http.Client{
		Timeout: time.Second * 10, // Set a timeout for the request
	}

	// Create a new HTTP request
	req, err := http.NewRequest("GET", feedURL, nil)
	if err != nil {
		utils.HandleError(err, "Error creating request", false)
		return nil, err // Return the error after handling it
	}

	// Set headers for the request
	req.Header.Set("User-Agent", "WriteupFinder/1.0")
	req.Header.Set("Accept", "application/rss+xml, application/xml;q=0.9, */*;q=0.8")

	// Use the custom HTTP client to execute the request
	resp, err := client.Do(req)
	if err != nil {
		utils.HandleError(err, "Error fetching feed", false)
		return nil, err // Return the error after handling it
	}
	defer resp.Body.Close()

	// Check for HTTP status code errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err := fmt.Errorf("HTTP error: %s, status code: %d", resp.Status, resp.StatusCode)
		utils.HandleError(err, "Invalid HTTP response", false)
		return nil, err // Return the error with details
	}

	// Parse the RSS feed
	parser := gofeed.NewParser()
	feed, err := parser.Parse(resp.Body)
	if err != nil {
		utils.HandleError(err, "Error parsing feed", false)
		return nil, err // Return the error after handling it
	}

	// Ensure the parsed feed is not nil
	if feed == nil {
		err := fmt.Errorf("no feed data received from URL: %s", feedURL)
		utils.HandleError(err, "Nil feed data", false)
		return nil, err
	}

	return feed.Items, nil
}

// ParseDate parses a date string and returns the corresponding time.Time object.
func ParseDate(dateString string) (time.Time, error) {
	parsedTime, err := time.Parse(time.RFC1123Z, dateString)
	if err != nil {
		// Attempt parsing with an alternative format
		parsedTime, err = time.Parse(time.RFC1123, dateString)
	}

	// Handle any parsing error, but allow the program to continue
	utils.HandleError(err, "Error parsing date", false)

	return parsedTime, err // Return both the parsed time and any error encountered
}
