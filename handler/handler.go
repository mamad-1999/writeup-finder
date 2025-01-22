package handler

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/fatih/color"
	"writeup-finder.go/utils"
)

// ProcessUrls iterates over a list of URLs and processes each one based on its type (Medium or YouTube).
func ProcessUrls(urlList []string, today time.Time, database *sql.DB) int {
	articlesFound := 0

	for i, url := range urlList {
		utils.PrintPretty(fmt.Sprintf("Processing feed: %s", url), color.FgMagenta, false)

		// Determine the type of feed and process accordingly
		if IsYouTubeFeed(url) {
			videosFound := ProcessYouTubeFeed(url, today, database)
			articlesFound += videosFound
		} else {
			articlesFound += ProcessMediumFeed(url, today, database)
		}

		// Delay processing of the next URL to prevent rate-limiting or server overload
		if i < len(urlList)-1 {
			time.Sleep(3 * time.Second)
		}
	}

	return articlesFound
}
