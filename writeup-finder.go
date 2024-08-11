package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/mmcdole/gofeed"
)

type TelegramMessage struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

var useFile bool
var useDatabase bool

func init() {
	// Load the .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}

func updateLastCheckTime() {
	file, err := os.Create("last-check.txt")
	if err != nil {
		fmt.Println(color.RedString("Error creating file: %s", err))
		return
	}
	defer file.Close()

	currentTime := time.Now().Format("2006-01-02 15:04:05")
	file.WriteString(currentTime)
}

func readUrls() []string {
	file, err := os.Open("url.txt")
	if err != nil {
		fmt.Println(color.RedString("Error reading URL file: %s", err))
		return nil
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			urls = append(urls, line)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(color.RedString("Error scanning URL file: %s", err))
	}
	return urls
}

// Database functions

func connectDB() (*sql.DB, error) {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func readFoundUrlsFromDB(db *sql.DB) (map[string]struct{}, error) {
	foundUrls := make(map[string]struct{})

	rows, err := db.Query("SELECT url FROM found_urls")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			return nil, err
		}
		foundUrls[url] = struct{}{}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return foundUrls, nil
}

func saveUrlToDB(db *sql.DB, url string) error {
	_, err := db.Exec("INSERT INTO found_urls (url) VALUES ($1)", url)
	return err
}

// File-based functions

func readFoundUrlsFromFile() map[string]struct{} {
	foundUrls := make(map[string]struct{})
	file, err := os.Open("found-url.txt")
	if err != nil {
		return foundUrls
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		foundUrls[scanner.Text()] = struct{}{}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(color.RedString("Error scanning found URL file: %s", err))
	}
	return foundUrls
}

func saveUrlToFile(url string) {
	file, err := os.OpenFile("found-url.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println(color.RedString("Error opening found URL file: %s", err))
		return
	}
	defer file.Close()

	file.WriteString(url + "\n")
}

// Common functions

func fetchArticles(feedUrl string) ([]*gofeed.Item, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(feedUrl)
	if err != nil {
		return nil, err
	}
	return feed.Items, nil
}

func parseDate(dateString string) (time.Time, error) {
	t, err := time.Parse(time.RFC1123Z, dateString)
	if err != nil {
		t, err = time.Parse(time.RFC1123, dateString)
	}
	return t, err
}

func printPretty(message string, colorAttr color.Attribute, isTitle bool) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	colored := color.New(colorAttr).SprintFunc()

	if isTitle {
		fmt.Println(colored(strings.Repeat("=", 80)))
		fmt.Println(colored(fmt.Sprintf("%80s", message)))
		fmt.Println(colored(strings.Repeat("=", 80)))
	} else {
		fmt.Println(color.CyanString(timestamp), "-", colored(message))
	}
}

func sendToTelegram(message string, botToken string, channelID string) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

	telegramMessage := TelegramMessage{
		ChatID: channelID,
		Text:   message,
	}

	jsonData, err := json.Marshal(telegramMessage)
	if err != nil {
		fmt.Println(color.RedString("Error marshalling Telegram message: %s", err))
		return
	}

	var resp *http.Response
	var retryCount int
	maxRetries := 5

	for {
		resp, err = http.Post(url, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Println(color.RedString("Error sending message to Telegram: %s", err))
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			break
		} else if resp.StatusCode == http.StatusTooManyRequests {
			retryAfter := time.Duration(2<<retryCount) * time.Second
			fmt.Println(color.YellowString("Rate limit exceeded, retrying after %v...", retryAfter))
			time.Sleep(retryAfter)
			retryCount++
			if retryCount > maxRetries {
				fmt.Println(color.RedString("Max retries reached. Failed to send message to Telegram."))
				return
			}
		} else {
			fmt.Println(color.RedString("Failed to send message to Telegram, status code: %d", resp.StatusCode))
			return
		}
	}
}

func main() {
	flag.BoolVar(&useFile, "f", false, "Save new articles in found-url.txt")
	flag.BoolVar(&useDatabase, "d", false, "Save new articles in the database")
	flag.Parse()

	if !useFile && !useDatabase {
		log.Fatal("You must specify either -f (file) or -d (database)")
	}

	TELEGRAM_BOT_TOKEN := os.Getenv("TELEGRAM_BOT_TOKEN")
	TELEGRAM_CHANNEL_ID := os.Getenv("TELEGRAM_CHANNEL_ID")
	printPretty("Starting Writeup Finder Script", color.FgGreen, true)

	urls := readUrls()
	threeDaysAgo := time.Now().AddDate(0, 0, -3)

	var foundUrls map[string]struct{}
	var db *sql.DB
	var err error

	if useFile {
		foundUrls = readFoundUrlsFromFile()
	} else if useDatabase {
		db, err = connectDB()
		if err != nil {
			log.Fatalf("Error connecting to database: %s", err)
		}
		defer db.Close()
		foundUrls, err = readFoundUrlsFromDB(db)
		if err != nil {
			log.Fatalf("Error reading found URLs from database: %s", err)
		}
	}

	articlesFound := 0

	for _, url := range urls {
		printPretty(fmt.Sprintf("Processing feed: %s", url), color.FgMagenta, false)
		articles, err := fetchArticles(url)
		if err != nil {
			fmt.Println(color.RedString("Error fetching feed from %s: %s", url, err))
			continue
		}

		for _, article := range articles {
			if _, exists := foundUrls[article.Link]; !exists {
				pubDate, err := parseDate(article.Published)
				if err != nil || pubDate.Before(threeDaysAgo) {
					message := fmt.Sprintf("▶ %s\nPublished: %s\nLink: %s",
						article.Title, article.Published, article.Link)

					sendToTelegram(message, TELEGRAM_BOT_TOKEN, TELEGRAM_CHANNEL_ID)

					if useFile {
						saveUrlToFile(article.Link)
					} else if useDatabase {
						err = saveUrlToDB(db, article.Link)
						if err != nil {
							fmt.Println(color.RedString("Error saving URL to database: %s", err))
						}
					}

					fmt.Println(color.GreenString(message))
					fmt.Println()
					articlesFound++
				}
			}
		}

	}

	printPretty(fmt.Sprintf("Total new articles found: %d", articlesFound), color.FgCyan, false)
	printPretty("Writeup Finder Script Completed", color.FgGreen, true)
	updateLastCheckTime()
}
