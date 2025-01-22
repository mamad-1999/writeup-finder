package telegram

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"writeup-finder.go/utils"
)

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
		err := SendRequest(client, apiURL, jsonData, &retryCount)
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
