package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

// UrlEntry represents a single URL entry with a title and URL.
type UrlEntry struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

// FoundUrls holds a slice of URL entries.
type FoundUrls struct {
	Urls []UrlEntry `json:"urls"`
}

// ReadFoundUrlsFromFile reads URL entries from a JSON file and returns a map of URLs.
func ReadFoundUrlsFromFile(filePath string) map[string]struct{} {
	foundUrls := make(map[string]struct{})

	file, err := os.Open(filePath)
	if err != nil {
		HandleError(err, "Error opening found-url.json file", false)
		return foundUrls
	}
	defer file.Close()

	var data FoundUrls
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil && err != io.EOF {
		HandleError(err, "Error decoding found-url.json file", false)
		return foundUrls
	}

	for _, entry := range data.Urls {
		foundUrls[entry.URL] = struct{}{}
	}

	return foundUrls
}

// SaveUrlToFile appends a new URL entry to the JSON file.
func SaveUrlToFile(filePath, title, url string) {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		HandleError(err, "Error opening found-url.json file", false)
		return
	}
	defer file.Close()

	var data FoundUrls
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil && err != io.EOF {
		HandleError(err, "Error decoding found-url.json file", false)
		return
	}

	if err == io.EOF {
		data.Urls = []UrlEntry{}
	}

	data.Urls = append(data.Urls, UrlEntry{Title: title, URL: url})

	file.Seek(0, io.SeekStart)
	file.Truncate(0)
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(data); err != nil {
		HandleError(err, "Error encoding to found-url.json file", false)
	}
}

// ReadUrls reads a list of URLs from a text file and returns them as a slice of strings.
func ReadUrls(filePath string) []string {
	file, err := os.Open(filePath)
	if err != nil {
		HandleError(err, "Error opening URL file", false)
		return nil
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			urls = append(urls, line)
		}
	}

	if err := scanner.Err(); err != nil {
		HandleError(err, "Error scanning URL file", false)
	}

	return urls
}

// HandleError prints an error message and optionally exits the program.
func HandleError(err error, message string, exit bool) {
	if err != nil {
		fmt.Println(color.RedString("%s: %s", message, err))
		if exit {
			os.Exit(1)
		}
	}
}

// PrintPretty prints a message with optional coloring and formatting.
func PrintPretty(message string, colorAttr color.Attribute, isTitle bool) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	colored := color.New(colorAttr).SprintFunc()

	if isTitle {
		width := 80
		padding := (width - len(message)) / 2
		fmt.Println(colored(strings.Repeat("=", width)))
		fmt.Printf("%s%s%s\n", strings.Repeat(" ", padding), colored(message), strings.Repeat(" ", width-len(message)-padding))
		fmt.Println(colored(strings.Repeat("=", width)))
	} else {
		fmt.Println(color.CyanString(timestamp), "-", colored(message))
	}
}
