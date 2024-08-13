package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
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

type UrlEntry struct {
	Title string `json:"title"`
	Url   string `json:"url"`
}

type FoundUrls struct {
	Urls []UrlEntry `json:"urls"`
}

var useFile bool
var useDatabase bool
var sendToTelegramFlag bool

const dataFolder = "data/"

func init() {
	// Load the .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}

func updateLastCheckTime() {
	file, err := os.Create(dataFolder + "last-check.txt")
	if err != nil {
		fmt.Println(color.RedString("Error creating file: %s", err))
		return
	}
	defer file.Close()

	currentTime := time.Now().Format("2006-01-02 15:04:05")
	file.WriteString(currentTime)
}

func readUrls() []string {
	file, err := os.Open(dataFolder + "url.txt")
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

	rows, err := db.Query("SELECT url FROM articles")
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
	_, err := db.Exec("INSERT INTO articles (url) VALUES ($1)", url)
	return err
}

// File-based functions

func readFoundUrlsFromFile() map[string]struct{} {
	foundUrls := make(map[string]struct{})
	file, err := os.Open(dataFolder + "found-url.json")
	if err != nil {
		return foundUrls
	}
	defer file.Close()

	var data FoundUrls
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&data)
	if err != nil {
		if err == io.EOF {
			return foundUrls
		}
		fmt.Println(color.RedString("Error decoding found URL JSON file: %s", err))
		return foundUrls
	}

	for _, entry := range data.Urls {
		foundUrls[entry.Url] = struct{}{}
	}

	return foundUrls
}

func saveUrlToFile(title string, url string) {
	file, err := os.OpenFile(dataFolder+"found-url.json", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println(color.RedString("Error opening found URL JSON file: %s", err))
		return
	}
	defer file.Close()

	var data FoundUrls
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&data)
	if err != nil && err != io.EOF {
		fmt.Println(color.RedString("Error decoding found URL JSON file: %s", err))
		return
	}

	if err == io.EOF {
		data.Urls = []UrlEntry{}
	}

	newEntry := UrlEntry{Title: title, Url: url}
	data.Urls = append(data.Urls, newEntry)

	file.Seek(0, 0)
	file.Truncate(0)
	encoder := json.NewEncoder(file)
	err = encoder.Encode(data)
	if err != nil {
		fmt.Println(color.RedString("Error encoding to found URL JSON file: %s", err))
	}
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
		width := 80 // Set the width for alignment
		// Calculate padding
		padding := (width - len(message)) / 2
		fmt.Println(colored(strings.Repeat("=", width)))
		fmt.Printf("%s%s%s\n", strings.Repeat(" ", padding), colored(message), strings.Repeat(" ", width-len(message)-padding))
		fmt.Println(colored(strings.Repeat("=", width)))
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
	flag.BoolVar(&sendToTelegramFlag, "t", false, "Send new articles to Telegram")
	flag.Parse()

	if !useFile && !useDatabase {
		log.Fatal("You must specify either -f (file) or -d (database)")
	}

	TELEGRAM_BOT_TOKEN := os.Getenv("TELEGRAM_BOT_TOKEN")
	TELEGRAM_CHANNEL_ID := os.Getenv("TELEGRAM_CHANNEL_ID")
	printPretty("Starting Writeup Finder Script", color.FgHiYellow, true)

	urls := readUrls()
	today := time.Now()

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
			pubDate, err := parseDate(article.Published)
			if err != nil || pubDate.Format("2006-01-02") != today.Format("2006-01-02") {
				continue // Skip articles not published today
			}

			if _, exists := foundUrls[article.Link]; !exists {
				message := fmt.Sprintf("▶ %s\nPublished: %s\nLink: %s",
					article.Title, article.Published, article.Link)

				if sendToTelegramFlag {
					sendToTelegram(message, TELEGRAM_BOT_TOKEN, TELEGRAM_CHANNEL_ID)
				}

				if useFile {
					saveUrlToFile(article.Title, article.Link)
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

	printPretty(fmt.Sprintf("Total new articles found: %d", articlesFound), color.FgYellow, false)
	printPretty("Writeup Finder Script Completed", color.FgHiYellow, true)
	updateLastCheckTime()
}
