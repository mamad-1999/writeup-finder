package handler

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/fatih/color"
	"writeup-finder.go/utils"
)

// processMediumFeed fetches and processes articles from a Medium RSS feed.
func ProcessMediumFeed(url string, today time.Time, database *sql.DB) int {
	articlesFound := 0
	articles, err := utils.FetchArticles(url)
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
			articlesFound++
		}
	}
	return articlesFound
}
