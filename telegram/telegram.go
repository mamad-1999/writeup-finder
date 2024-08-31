package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/fatih/color"
	"writeup-finder.go/utils"
)

type TelegramMessage struct {
	ChatID          string `json:"chat_id"`
	Text            string `json:"text"`
	MessageThreadID string `json:"message_thread_id,omitempty"`
}

func SendToTelegram(message string, proxyURL string) {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	channelID := os.Getenv("CHAT_ID")
	messageThreadID := os.Getenv("MESSAGE_THREAD_ID")

	apiUrl := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

	telegramMessage := TelegramMessage{
		ChatID:          channelID,
		Text:            message,
		MessageThreadID: messageThreadID,
	}

	jsonData, err := json.Marshal(telegramMessage)
	utils.HandleError(err, "Error marshalling Telegram message", false)

	client := utils.CreateHttpClient(proxyURL)

	var resp *http.Response
	var retryCount int
	maxRetries := 5

	for {
		resp, err = client.Post(apiUrl, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				fmt.Println(color.RedString("Network timeout, retrying..."))
				retryCount++
				if retryCount > maxRetries {
					fmt.Println(color.RedString("Max retries reached. Failed to send message to Telegram."))
					return
				}
				time.Sleep(time.Second * 2)
				continue
			}
			utils.HandleError(err, "Error sending message to Telegram", false)
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

func ValidateProxyURL(proxyURL string) error {
	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		return err
	}

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
