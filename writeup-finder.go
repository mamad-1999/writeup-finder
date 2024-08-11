package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/mmcdole/gofeed"
)

type TelegramMessage struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

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

func readFoundUrls() map[string]struct{} {
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

func saveUrl(url string) {
	file, err := os.OpenFile("found-url.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println(color.RedString("Error opening found URL file: %s", err))
		return
	}
	defer file.Close()

	file.WriteString(url + "\n")
}

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

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println(color.RedString("Error sending message to Telegram: %s", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println(color.RedString("Failed to send message to Telegram, status code: %d", resp.StatusCode))
	}
}

func main() {
	TELEGRAM_BOT_TOKEN := os.Getenv("TELEGRAM_BOT_TOKEN")
	TELEGRAM_CHANNEL_ID := os.Getenv("TELEGRAM_CHANNEL_ID")
	printPretty("Starting Writeup Finder Script", color.FgGreen, true)

	urls := readUrls()
	foundUrls := readFoundUrls()
	tenDaysAgo := time.Now().AddDate(0, 0, -10)

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
				if err != nil || pubDate.Before(tenDaysAgo) {
					saveUrl(article.Link)

					message := fmt.Sprintf("▶ %s\nPublished: %s\nLink: %s",
						article.Title, article.Published, article.Link)

					sendToTelegram(message, TELEGRAM_BOT_TOKEN, TELEGRAM_CHANNEL_ID) // Send the article to the Telegram channel

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
