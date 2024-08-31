package utils

import (
	"net/http"
	"net/url"
	"time"
)

func CreateHTTPClient(proxyURL string) *http.Client {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	if proxyURL != "" {
		proxy, err := url.Parse(proxyURL)
		if err != nil {
			HandleError(err, "Error parsing proxy URL", false)
			// In case of an error, return the client without proxy settings
			return client
		}
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxy),
		}
	}

	return client
}
