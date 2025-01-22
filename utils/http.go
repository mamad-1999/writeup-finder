package utils

import (
	"net/http"
	"net/url"
	"time"
)

// CreateHTTPClient creates and returns an HTTP client with a 30-second timeout.
// If a proxy URL is provided, it configures the client to use the proxy.
// If the proxy URL is invalid, the function logs an error and returns a client without proxy settings.
func CreateHTTPClient(proxyURL string) *http.Client {
	client := &http.Client{
		Timeout: 30 * time.Second, // Set a 30-second timeout for all requests
	}

	if proxyURL != "" {
		proxy, err := url.Parse(proxyURL)
		if err != nil {
			HandleError(err, "Error parsing proxy URL", false)
			return client
		}
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxy),
		}
	}

	return client
}
