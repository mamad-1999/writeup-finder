package utils

import (
	"fmt"
	"net/http"
	"time"

	"github.com/mmcdole/gofeed"
)

// FetchArticles retrieves articles from the given RSS feed URL.
// It returns a list of feed items or an error if the request or parsing fails.
func FetchArticles(feedURL string) ([]*gofeed.Item, error) {
	client := &http.Client{
		Timeout: time.Second * 10, // Set a timeout for the request
	}

	req, err := http.NewRequest("GET", feedURL, nil)
	if err != nil {
		HandleError(err, "Error creating request", false)
		return nil, err
	}

	// Set headers to mimic a browser request
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:91.0) Gecko/20100101 Firefox/91.0")
	req.Header.Set("Accept", "application/rss+xml, application/xml;q=0.9, */*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		HandleError(err, "Error fetching feed", false)
		return nil, err
	}
	defer resp.Body.Close()

	// Check for non-2xx HTTP status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err := fmt.Errorf("HTTP error: %s, status code: %d", resp.Status, resp.StatusCode)
		HandleError(err, "Invalid HTTP response", false)
		return nil, err
	}

	parser := gofeed.NewParser()
	feed, err := parser.Parse(resp.Body)
	if err != nil {
		HandleError(err, "Error parsing feed", false)
		return nil, err
	}

	if feed == nil {
		err := fmt.Errorf("no feed data received from URL: %s", feedURL)
		HandleError(err, "Nil feed data", false)
		return nil, err
	}

	return feed.Items, nil
}

// ParseDate attempts to parse a date string using RFC1123Z or RFC1123 formats.
// It returns the parsed time.Time object and any error encountered during parsing.
func ParseDate(dateString string) (time.Time, error) {
	parsedTime, err := time.Parse(time.RFC1123Z, dateString)
	if err != nil {
		// Attempt parsing with an alternative format
		parsedTime, err = time.Parse(time.RFC1123, dateString)
	}

	HandleError(err, "Error parsing date", false)

	return parsedTime, err
}
