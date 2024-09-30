package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

// ReadUrls reads a list of URLs from a text file and returns them as a slice of strings.
func ReadUrls(filePath string) []string {
	file, err := os.Open(filePath)
	HandleError(err, "Error opening URL file", false)
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
