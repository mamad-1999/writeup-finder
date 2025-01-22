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

// TelegramMessage represents the structure of a message to be sent to Telegram.
type TelegramMessage struct {
	ChatID          string `json:"chat_id"`
	Text            string `json:"text"`
	MessageThreadID string `json:"message_thread_id,omitempty"`
}

const (
	maxRetries    = 5               // Maximum number of retries for sending a message
	retryDelay    = 2 * time.Second // Delay between retries
	rateLimitBase = 2               // Base multiplier for rate limit backoff
)

// SendToTelegram sends a message to a Telegram channel using the provided proxy.
// It handles retries, rate limiting, and thread selection based on the message type (YouTube or keyword-based).
func SendToTelegram(message string, proxyURL string, title string, isYoutube bool) {
	botToken := utils.GetEnv("TELEGRAM_BOT_TOKEN")
	channelID := utils.GetEnv("CHAT_ID")
	mainThreadID := utils.GetEnv("MAIN_THREAD_ID")
	youtubeThreadID := utils.GetEnv("YOUTUBE_THREAD_ID")

	var messageThreadID string

	if isYoutube {
		messageThreadID = youtubeThreadID
	} else {
		// Load keywords from the JSON configuration
		keywords, err := utils.LoadKeywords("data/keywords.json")
		if err != nil {
			utils.HandleError(err, "Failed to load keyword patterns", true)
		}

		// Determine the message thread ID based on title keywords
		messageThreadID = utils.MatchKeyword(title, keywords, mainThreadID)
	}

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

// sendRequest sends an HTTP POST request to the Telegram API.
// It handles retries for network errors, rate limiting, and unexpected status codes.
func sendRequest(client *http.Client, apiURL string, jsonData []byte, retryCount *int) error {
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

// ValidateProxyURL checks if the provided proxy URL is valid and supported.
// It returns an error if the scheme is unsupported or the hostname is missing.
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
