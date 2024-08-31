package utils

import (
	"net/http"
	"net/url"
	"time"
)

func CreateHttpClient(proxyURL string) *http.Client {
	var client *http.Client

	if proxyURL != "" {
		proxy, err := url.Parse(proxyURL)
		HandleError(err, "Error parsing proxy URL", false)
		client = &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxy),
			},
		}
	} else {
		client = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	return client
}
