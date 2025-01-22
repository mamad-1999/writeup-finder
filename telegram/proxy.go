package telegram

import (
	"fmt"
	"net/url"

	"writeup-finder.go/utils"
)

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
