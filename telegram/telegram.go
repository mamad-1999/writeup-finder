package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
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
	botToken := utils.GetEnv("TELEGRAM_BOT_TOKEN")
	channelID := utils.GetEnv("CHAT_ID")
	mainThreadID := utils.GetEnv("MAIN_THREAD_ID")
	fmt.Println("This log is in SendToTelegram...")
	fmt.Println(botToken, "Bot Token")

	// Load keywords from the JSON configuration
	keywords, err := utils.LoadKeywords("data/keywords.json")
	if err != nil {
		utils.HandleError(err, "Failed to load keyword patterns", true)
	}

	// Determine the message thread ID
	messageThreadID := utils.MatchKeyword(title, keywords, mainThreadID)

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
				log.Printf("Failed to send message to Telegram after %d retries: %v", maxRetries, err)
				return
			}
			log.Printf("Retrying request (%d/%d): %v", retryCount, maxRetries, err)
			time.Sleep(retryDelay) // Wait before retrying
			continue
		}
		log.Println("Message sent successfully!")
		break // Exit the loop if request was successful
	}
}

func sendRequest(client *http.Client, apiURL string, jsonData []byte, retryCount *int) error {
	fmt.Println("Request start to send to Telegram...")
	resp, err := client.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			fmt.Println(color.RedString("Network timeout, retrying..."))
			(*retryCount)++
			return err // Return the error to trigger a retry
		}
		(*retryCount)++
		return err // Increment retry count for non-retryable errors
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil // Success, no need to retry
	}

	// Handle rate limiting
	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := time.Duration(rateLimitBase<<*retryCount) * time.Second
		fmt.Println(color.YellowString("Rate limit exceeded, retrying after %v...", retryAfter))
		(*retryCount)++
		time.Sleep(retryAfter)
		return fmt.Errorf("rate limit exceeded")
	}

	// Log unexpected HTTP status codes
	log.Printf("Unexpected status code %d: retrying...", resp.StatusCode)
	(*retryCount)++
	return fmt.Errorf("failed to send message, status code: %d", resp.StatusCode)
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
