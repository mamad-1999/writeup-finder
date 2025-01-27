package handler

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/mmcdole/gofeed"
	"writeup-finder.go/db"
	"writeup-finder.go/global"
	"writeup-finder.go/telegram"
	"writeup-finder.go/utils"
)

// isYouTubeFeed determines if a given URL corresponds to a YouTube RSS feed.
func IsYouTubeFeed(url string) bool {
	return strings.HasPrefix(url, "https://www.youtube.com/feeds/")
}

// IsNewArticle checks if an article is new by comparing its publication date and database presence.
func IsNewArticle(item *gofeed.Item, db *sql.DB, today time.Time) bool {
	pubDate, err := utils.ParseDate(item.Published)
	if err != nil {
		return false
	}

	yesterday := today.AddDate(0, 0, -1)
	isToday := pubDate.Format(global.DateFormat) == today.Format(global.DateFormat)
	isYesterday := pubDate.Format(global.DateFormat) == yesterday.Format(global.DateFormat)

	if !isToday && !isYesterday {
		return false
	}

	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM articles WHERE title = $1)"
	err = db.QueryRow(query, item.Title).Scan(&exists)
	utils.HandleError(err, "Error checking if article title exists in database", false)

	return !exists
}

// FormatArticleMessage creates a formatted string for an article's details.
func FormatArticleMessage(item *gofeed.Item) string {
	// Check if the GUID starts with "https://medium.com"
	if !strings.HasPrefix(item.GUID, "https://medium.com") {
		return fmt.Sprintf("\u25BA %s\nPublished: %s\nLink: %s", item.Title, item.Published, item.GUID)
	}

	premium, err := isPremium(item.GUID)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// If the article is premium, change the domain
	if premium {
		item.GUID = strings.Replace(item.GUID, "https://medium.com", "https://readmedium.com", 1)
	}

	return fmt.Sprintf("\u25BA %s\nPublished: %s\nLink: %s", item.Title, item.Published, item.GUID)
}

func isPremium(url string) (bool, error) {
	// Create a new allocator with custom user agent
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
		chromedp.Flag("no-sandbox", true), // Add this line
	)

	// Create a new context with the allocator
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// Create a new browser context
	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	// Set a timeout for the entire operation
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var isPremium bool

	// Run the browser tasks
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body"), // Wait for the body to load
		chromedp.Evaluate(`document.querySelectorAll('[aria-label="Close"]').forEach(btn => btn.click());`, nil), // Close popups
		chromedp.Evaluate(`{
			// Check for "Member-only story" text
			const xpathCheck = document.evaluate(
				'//*[contains(text(), "Member-only story")]',
				document,
				null,
				XPathResult.ANY_TYPE,
				null
			);
			const hasMemberText = xpathCheck.iterateNext() !== null;

			// Check for the golden star icon
			const hasGoldenStar = document.querySelector('svg[fill="#FFC017"]') !== null;

			// Return true if either condition is met
			hasMemberText || hasGoldenStar;
		}`, &isPremium),
	)

	if err != nil {
		return false, fmt.Errorf("error checking %s: %v", url, err)
	}

	return isPremium, nil
}

// HandleArticle manages sending an article to Telegram and saving it to the database if enabled.
func HandleArticle(item *gofeed.Item, message string, database *sql.DB, isYoutube bool) error {
	if global.SendToTelegramFlag {
		fmt.Println("Start Send to Telegram...")

		telegram.SendToTelegram(message, global.ProxyURL, item.Title, isYoutube)
	}

	if global.UseDatabase {
		db.SaveUrlToDB(database, item.GUID, item.Title)
	}

	return nil
}
