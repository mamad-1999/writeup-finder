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
	"net/url"
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
var proxyURL string

const dataFolder = "data/"

func init() {
	err := godotenv.Load()
	handleError(err, "Error loading .env file", true)
}

func handleError(err error, message string, exit bool) {
	if err != nil {
		fmt.Println(color.RedString("%s: %s", message, err))
		if exit {
			os.Exit(1)
		}
	}
}

func updateLastCheckTime() {
	file, err := os.Create(dataFolder + "last-check.txt")
	handleError(err, "Error creating last-check.txt file", false)
	defer file.Close()

	currentTime := time.Now().Format("2006-01-02 15:04:05")
	_, err = file.WriteString(currentTime)
	handleError(err, "Error writing to last-check.txt file", false)
}

func readUrls() []string {
	file, err := os.Open(dataFolder + "url.txt")
	handleError(err, "Error reading URL file", false)
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			urls = append(urls, line)
		}
	}

	handleError(scanner.Err(), "Error scanning URL file", false)
	return urls
}

func connectDB() (*sql.DB, error) {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := sql.Open("postgres", connStr)
	handleError(err, "Error opening database connection", true)

	return db, nil
}

func readFoundUrlsFromDB(db *sql.DB) (map[string]struct{}, error) {
	foundUrls := make(map[string]struct{})

	rows, err := db.Query("SELECT url FROM articles")
	handleError(err, "Error querying database", true)
	defer rows.Close()

	for rows.Next() {
		var url string
		err := rows.Scan(&url)
		handleError(err, "Error scanning row in database", false)
		foundUrls[url] = struct{}{}
	}

	handleError(rows.Err(), "Error iterating over database rows", false)

	return foundUrls, nil
}

func saveUrlToDB(db *sql.DB, url string) {
	_, err := db.Exec("INSERT INTO articles (url) VALUES ($1)", url)
	handleError(err, "Error saving URL to database", false)
}

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
	if err != nil && err != io.EOF {
		handleError(err, "Error decoding found-url.json", false)
		return foundUrls
	}

	for _, entry := range data.Urls {
		foundUrls[entry.Url] = struct{}{}
	}

	return foundUrls
}

func saveUrlToFile(title string, url string) {
	file, err := os.OpenFile(dataFolder+"found-url.json", os.O_RDWR|os.O_CREATE, 0644)
	handleError(err, "Error opening found-url.json file", false)
	defer file.Close()

	var data FoundUrls
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&data)
	if err != nil && err != io.EOF {
		handleError(err, "Error decoding found-url.json file", false)
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
	handleError(err, "Error encoding to found-url.json file", false)
}

func fetchArticles(feedUrl string) ([]*gofeed.Item, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(feedUrl)
	handleError(err, fmt.Sprintf("Error fetching feed from %s", feedUrl), false)
	return feed.Items, err
}

func parseDate(dateString string) (time.Time, error) {
	t, err := time.Parse(time.RFC1123Z, dateString)
	if err != nil {
		t, err = time.Parse(time.RFC1123, dateString)
	}
	handleError(err, "Error parsing date", false)
	return t, err
}

func printPretty(message string, colorAttr color.Attribute, isTitle bool) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	colored := color.New(colorAttr).SprintFunc()

	if isTitle {
		width := 80
		padding := (width - len(message)) / 2
		fmt.Println(colored(strings.Repeat("=", width)))
		fmt.Printf("%s%s%s\n", strings.Repeat(" ", padding), colored(message), strings.Repeat(" ", width-len(message)-padding))
		fmt.Println(colored(strings.Repeat("=", width)))
	} else {
		fmt.Println(color.CyanString(timestamp), "-", colored(message))
	}
}

func sendToTelegram(message string, botToken string, channelID string, proxyURL string) {
	apiUrl := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

	telegramMessage := TelegramMessage{
		ChatID: channelID,
		Text:   message,
	}

	jsonData, err := json.Marshal(telegramMessage)
	handleError(err, "Error marshalling Telegram message", false)

	var client *http.Client
	if proxyURL != "" {
		proxy, err := url.Parse(proxyURL)
		handleError(err, "Error parsing proxy URL", false)
		client = &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxy),
			},
		}
	} else {
		client = &http.Client{}
	}

	var resp *http.Response
	var retryCount int
	maxRetries := 5

	for {
		resp, err = client.Post(apiUrl, "application/json", bytes.NewBuffer(jsonData))
		handleError(err, "Error sending message to Telegram", false)
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
	flag.StringVar(&proxyURL, "proxy", "", "Proxy URL to use for sending Telegram messages")
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
		handleError(err, "Error connecting to database", true)
		defer db.Close()
		foundUrls, err = readFoundUrlsFromDB(db)
		handleError(err, "Error reading found URLs from database", true)
	}

	articlesFound := 0

	for _, url := range urls {
		printPretty(fmt.Sprintf("Processing feed: %s", url), color.FgMagenta, false)
		articles, err := fetchArticles(url)
		if err != nil {
			continue
		}

		for _, article := range articles {
			pubDate, err := parseDate(article.Published)
			if err != nil || pubDate.Format("2006-01-02") != today.Format("2006-01-02") {
				continue
			}

			if _, exists := foundUrls[article.Link]; !exists {
				message := fmt.Sprintf("▶ %s\nPublished: %s\nLink: %s",
					article.Title, article.Published, article.GUID)

				if sendToTelegramFlag {
					sendToTelegram(message, TELEGRAM_BOT_TOKEN, TELEGRAM_CHANNEL_ID, proxyURL)
				}

				if useFile {
					saveUrlToFile(article.Title, article.Link)
				} else if useDatabase {
					saveUrlToDB(db, article.Link)
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
