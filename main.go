package main

import (
	"fmt"
	"time"

	"writeup-finder.go/command"
	"writeup-finder.go/config"
	"writeup-finder.go/db"
	"writeup-finder.go/global"
	"writeup-finder.go/url"
	"writeup-finder.go/utils"

	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp:       true,
		DisableLevelTruncation: true,
		FullTimestamp:          false,
	})
	config.LoadEnv()
}

func main() {
	command.ManageFlags()
	utils.PrintPretty("Starting Writeup Finder Script", color.FgHiYellow, true)

	if global.UseDatabase {
		global.DB = db.ConnectDB()
		db.CreateArticlesTable(global.DB)
		defer global.DB.Close()
	}

	urlList := utils.ReadUrls(global.UrlFile)
	today := time.Now()

	articlesFound := url.ProcessUrls(urlList, today, global.DB)

	utils.PrintPretty(fmt.Sprintf("Total new articles found: %d", articlesFound), color.FgYellow, false)
	utils.PrintPretty("Writeup Finder Script Completed", color.FgHiYellow, true)
}
