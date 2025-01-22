package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
)

// KeywordPattern represents a compiled regex pattern, its associated thread ID, and priority.
type KeywordPattern struct {
	Pattern  *regexp.Regexp
	ThreadID string
	Priority int
}

// RawKeyword represents a keyword pattern and its associated thread ID and priority as loaded from JSON.
type RawKeyword struct {
	Pattern  string `json:"pattern"`
	ThreadID string `json:"threadID"`
	Priority int    `json:"priority"`
}

// KeywordGroup represents a group of keywords with a common name.
type KeywordGroup struct {
	Name     string       `json:"name"`
	Keywords []RawKeyword `json:"keywords"`
}

// LoadKeywords loads keyword patterns from a JSON configuration file and compiles them into regex patterns.
// It also maps thread IDs from environment variables and sorts the keywords by priority.
// Returns a slice of KeywordPattern or an error if the file cannot be read or the regex cannot be compiled.
func LoadKeywords(configPath string) ([]KeywordPattern, error) {
	// Map thread IDs from environment variables
	threadIDMap := map[string]string{
		"MONEY_THREAD_ID":            GetEnv("MONEY_THREAD_ID"),
		"BYPASS_THREAD_ID":           GetEnv("BYPASS_THREAD_ID"),
		"PLATFORMS_THREAD_ID":        GetEnv("PLATFORMS_THREAD_ID"),
		"TRYHACKME_THREAD_ID":        GetEnv("TRYHACKME_THREAD_ID"),
		"HACKTHEBOX_THREAD_ID":       GetEnv("HACKTHEBOX_THREAD_ID"),
		"MOBILE_THREAD_ID":           GetEnv("MOBILE_THREAD_ID"),
		"RECON_THREAD_ID":            GetEnv("RECON_THREAD_ID"),
		"PORTSWIGGER_THREAD_ID":      GetEnv("PORTSWIGGER_THREAD_ID"),
		"BURPSUITE_THREAD_ID":        GetEnv("BURPSUITE_THREAD_ID"),
		"CTF_THREAD_ID":              GetEnv("CTF_THREAD_ID"),
		"OS_THREAD_ID":               GetEnv("OS_THREAD_ID"),
		"VULNERABILITIES_THREAD_ID":  GetEnv("VULNERABILITIES_THREAD_ID"),
		"TOOLS_THREAD_ID":            GetEnv("TOOLS_THREAD_ID"),
		"PROGRAMMINGLANGS_THREAD_ID": GetEnv("PROGRAMMINGLANGS_THREAD_ID"),
		"CVE_THREAD_ID":              GetEnv("CVE_THREAD_ID"),
		"OSINT_THREAD_ID":            GetEnv("OSINT_THREAD_ID"),
		"CRYPTOGRAPHIC_THREAD_ID":    GetEnv("CRYPTOGRAPHIC_THREAD_ID"),
		"STEGANOGRAPHY_THREAD_ID":    GetEnv("STEGANOGRAPHY_THREAD_ID"),
		"WEBSCRAPING_THREAD_ID":      GetEnv("WEBSCRAPING_THREAD_ID"),
	}

	// Load JSON configuration file
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var rawConfig struct {
		Groups []KeywordGroup `json:"groups"`
	}

	if err := json.NewDecoder(file).Decode(&rawConfig); err != nil {
		return nil, err
	}

	// Parse keywords and compile regex patterns
	var keywords []KeywordPattern
	for _, group := range rawConfig.Groups {
		for _, raw := range group.Keywords {
			compiledPattern, err := regexp.Compile("(?i)" + raw.Pattern)
			if err != nil {
				return nil, err // Return an error if regex compilation fails
			}
			threadID, ok := threadIDMap[raw.ThreadID]
			if !ok {
				return nil, fmt.Errorf("unknown thread ID: %s", raw.ThreadID)
			}
			keywords = append(keywords, KeywordPattern{
				Pattern:  compiledPattern,
				ThreadID: threadID,
				Priority: raw.Priority,
			})
		}
	}

	// Sort keywords by priority (ascending order)
	sort.Slice(keywords, func(i, j int) bool {
		return keywords[i].Priority < keywords[j].Priority
	})

	return keywords, nil
}

// MatchKeyword searches for the first keyword pattern that matches the given title.
// It returns the associated thread ID if a match is found, otherwise returns the default thread ID.
func MatchKeyword(title string, keywords []KeywordPattern, defaultThreadID string) string {
	for _, keyword := range keywords {
		if keyword.Pattern.MatchString(title) {
			return keyword.ThreadID
		}
	}
	return defaultThreadID
}
