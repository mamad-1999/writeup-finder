package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"

	"encoding/json"
	"io"
)

type UrlEntry struct {
	Title string `json:"title"`
	Url   string `json:"url"`
}

type FoundUrls struct {
	Urls []UrlEntry `json:"urls"`
}

func ReadFoundUrlsFromFile(filePath string) map[string]struct{} {
	foundUrls := make(map[string]struct{})
	file, err := os.Open(filePath)
	if err != nil {
		return foundUrls
	}
	defer file.Close()

	var data FoundUrls
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&data)
	if err != nil && err != io.EOF {
		HandleError(err, "Error decoding found-url.json", false)
		return foundUrls
	}

	for _, entry := range data.Urls {
		foundUrls[entry.Url] = struct{}{}
	}

	return foundUrls
}

func SaveUrlToFile(filePath, title, url string) {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	HandleError(err, "Error opening found-url.json file", false)
	defer file.Close()

	var data FoundUrls
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&data)
	if err != nil && err != io.EOF {
		HandleError(err, "Error decoding found-url.json file", false)
		return
	}

	if err == io.EOF {
		data.Urls = []UrlEntry{}
	}

	newEntry := UrlEntry{Title: title, Url: url}
	data.Urls = append(data.Urls, newEntry)

	file.Seek(0, 0)
	file.Truncate(0)
	encoder := json.NewEncoder(file)
	err = encoder.Encode(data)
	HandleError(err, "Error encoding to found-url.json file", false)
}

func ReadUrls(filePath string) []string {
	file, err := os.Open(filePath)
	HandleError(err, "Error reading URL file", false)
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			urls = append(urls, line)
		}
	}

	HandleError(scanner.Err(), "Error scanning URL file", false)
	return urls
}

func HandleError(err error, message string, exit bool) {
	if err != nil {
		fmt.Println(color.RedString("%s: %s", message, err))
		if exit {
			os.Exit(1)
		}
	}
}

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
