package command

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"writeup-finder.go/global"
	"writeup-finder.go/handler"
	"writeup-finder.go/utils"
)

// ManageAction processes the list of URLs, finds new articles, and logs the results.
func ManageAction() {
	urlList := utils.ReadUrls(global.UrlFile)
	today := time.Now()

	// Process the URLs and store new articles in the database if enabled
	articlesFound := handler.ProcessUrls(urlList, today, global.DB)

	utils.PrintPretty(fmt.Sprintf("Total new articles found: %d", articlesFound), color.FgYellow, false)
	utils.PrintPretty("Writeup Finder Script Completed", color.FgHiYellow, true)
}
