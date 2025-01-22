package telegram

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/fatih/color"
)

// sendRequest sends an HTTP POST request to the Telegram API.
// It handles retries for network errors, rate limiting, and unexpected status codes.
func SendRequest(client *http.Client, apiURL string, jsonData []byte, retryCount *int) error {
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
