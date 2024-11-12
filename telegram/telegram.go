package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"time"

	"github.com/fatih/color"
	"writeup-finder.go/utils"
)

type TelegramMessage struct {
	ChatID          string `json:"chat_id"`
	Text            string `json:"text"`
	MessageThreadID string `json:"message_thread_id,omitempty"`
}

const (
	maxRetries    = 5
	retryDelay    = 2 * time.Second
	rateLimitBase = 2
)

func SendToTelegram(message string, proxyURL string, title string) {
	botToken := getEnv("TELEGRAM_BOT_TOKEN")
	channelID := getEnv("CHAT_ID")
	mainThreadID := getEnv("MAIN_THREAD_ID")

	// Define keywords with priority levels and regex patterns
	type keywordPattern struct {
		pattern  *regexp.Regexp
		threadID string
		priority int
	}

	keywords := []keywordPattern{
		{regexp.MustCompile(`(?i)\$[0-9]*|\bMoney\b|bounty|bounties`), getEnv("MONEY_THREAD_ID"), 1},
		{regexp.MustCompile(`(?i)Bypass|waf|firewall(?:-bypass)?|waf-bypass`), getEnv("BYPASS_THREAD_ID"), 2},
		{regexp.MustCompile(`(?i)Recon|reconnaissance|osint`), getEnv("RECON_THREAD_ID"), 3},
		{regexp.MustCompile(`(?i)THM|TryHackMe`), getEnv("TRYHACKME_THREAD_ID"), 4},
		{regexp.MustCompile(`(?i)HTB|HackTheBox`), getEnv("HACKTHEBOX_THREAD_ID"), 4},
		{regexp.MustCompile(`(?i)Mobile|Android|iOS|iPhone|iPad|Phone|Tablet`), getEnv("MOBILE_THREAD_ID"), 4},
		{regexp.MustCompile(`(?i)Portswigger`), getEnv("PORTSWIGGER_THREAD_ID"), 4},
		{regexp.MustCompile(`(?i)Burp|Burp\s?suite|Burpsuite-Pro`), getEnv("BURPSUITE_THREAD_ID"), 4},
		{regexp.MustCompile(`(?i)CTF|Capture\s?The\s?Flag`), getEnv("CTF_THREAD_ID"), 5},
		{regexp.MustCompile(`(?i)hackerone|bugcrowd|yeswehack|intigriti`), getEnv("PLATFORMS_THREAD_ID"), 5},
	}

	// Sort keywords by priority (lower number means higher priority)
	sort.Slice(keywords, func(i, j int) bool {
		return keywords[i].priority < keywords[j].priority
	})

	// Default to the main thread ID
	messageThreadID := mainThreadID

	// Find the first matching keyword in order of priority
	for _, keyword := range keywords {
		if keyword.pattern.MatchString(title) {
			messageThreadID = keyword.threadID
			break
		}
	}

	// Send the message
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	telegramMessage := TelegramMessage{
		ChatID:          channelID,
		Text:            message,
		MessageThreadID: messageThreadID,
	}

	jsonData, err := json.Marshal(telegramMessage)
	utils.HandleError(err, "Error marshalling Telegram message", false)

	client := utils.CreateHTTPClient(proxyURL)
	retryCount := 0

	for {
		err := sendRequest(client, apiURL, jsonData, &retryCount)
		if err != nil {
			if retryCount >= maxRetries {
				utils.HandleError(err, "Failed to send message to Telegram after retries", false)
				return
			}
			time.Sleep(retryDelay) // Wait before retrying
			continue
		}
		break // Exit the loop if request was successful
	}
}

func sendRequest(client *http.Client, apiURL string, jsonData []byte, retryCount *int) error {
	resp, err := client.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			fmt.Println(color.RedString("Network timeout, retrying..."))
			(*retryCount)++
			return err // Return the error to trigger a retry
		}
		return err // Return the error for non-retryable errors
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil // Success, no need to retry
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := time.Duration(rateLimitBase<<*retryCount) * time.Second
		fmt.Println(color.YellowString("Rate limit exceeded, retrying after %v...", retryAfter))
		(*retryCount)++
		time.Sleep(retryAfter)
		return fmt.Errorf("rate limit exceeded")
	}

	return fmt.Errorf("failed to send message, status code: %d", resp.StatusCode)
}

func getEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		utils.HandleError(fmt.Errorf("environment variable %s not set", key), "Missing environment variable", false)
	}
	return value
}

func ValidateProxyURL(proxyURL string) error {
	parsedURL, err := url.Parse(proxyURL)
	utils.HandleError(err, "Error:", true)

	switch parsedURL.Scheme {
	case "http", "https", "socks5":
	default:
		return fmt.Errorf("unsupported proxy scheme: %s", parsedURL.Scheme)
	}

	if parsedURL.Hostname() == "" {
		return fmt.Errorf("missing hostname or IP address in proxy URL")
	}

	return nil
}
