package utils

import (
	"encoding/json"
	"os"
	"regexp"
	"sort"
)

type KeywordPattern struct {
	Pattern  *regexp.Regexp
	ThreadID string
	Priority int
}

type RawKeyword struct {
	Pattern  string `json:"pattern"`
	ThreadID string `json:"threadID"`
	Priority int    `json:"priority"`
}

type KeywordGroup struct {
	Name     string       `json:"name"`
	Keywords []RawKeyword `json:"keywords"`
}

// LoadKeywords loads keyword patterns from a JSON file and resolves thread IDs
func LoadKeywords(configPath string) ([]KeywordPattern, error) {
	// Prepopulate thread IDs from environment variables
	threadIDMap := map[string]string{
		"MONEY_THREAD_ID":            GetEnv("MONEY_THREAD_ID"),
		"BYPASS_THREAD_ID":           GetEnv("BYPASS_THREAD_ID"),
		"PLATFORMS_THREAD_ID":        GetEnv("PLATFORMS_THREAD_ID"),
		"TRYHACKME_THREAD_ID":        GetEnv("TRYHACKME_THREAD_ID"),
		"HACKTHEBOX_THREAD_ID":       GetEnv("HACKTHEBOX_THREAD_ID"),
		"MOBILE_THREAD_ID":           GetEnv("MOBILE_THREAD_ID"),
		"PORTSWIGGER_THREAD_ID":      GetEnv("PORTSWIGGER_THREAD_ID"),
		"BURPSUITE_THREAD_ID":        GetEnv("BURPSUITE_THREAD_ID"),
		"CTF_THREAD_ID":              GetEnv("CTF_THREAD_ID"),
		"OS_THREAD_ID":               GetEnv("OS_THREAD_ID"),
		"VULNERABILITIES_THREAD_ID":  GetEnv("VULNERABILITIES_THREAD_ID"),
		"TOOLS_THREAD_ID":            GetEnv("TOOLS_THREAD_ID"),
		"PROGRAMMINGLANGS_THREAD_ID": GetEnv("PROGRAMMINGLANGS_THREAD_ID"),
	}

	// Load JSON config
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var groups []KeywordGroup
	if err := json.NewDecoder(file).Decode(&groups); err != nil {
		return nil, err
	}

	// Parse keywords and compile patterns
	var keywords []KeywordPattern
	for _, group := range groups {
		for _, raw := range group.Keywords {
			compiledPattern, err := regexp.Compile(raw.Pattern)
			if err != nil {
				return nil, err // Return an error if regex compilation fails
			}
			keywords = append(keywords, KeywordPattern{
				Pattern:  compiledPattern,
				ThreadID: threadIDMap[raw.ThreadID],
				Priority: raw.Priority,
			})
		}
	}

	// Sort by priority
	sort.Slice(keywords, func(i, j int) bool {
		return keywords[i].Priority < keywords[j].Priority
	})

	return keywords, nil
}

// MatchKeyword finds the thread ID for the first matching keyword in the title
func MatchKeyword(title string, keywords []KeywordPattern, defaultThreadID string) string {
	for _, keyword := range keywords {
		if keyword.Pattern.MatchString(title) {
			return keyword.ThreadID
		}
	}
	return defaultThreadID
}
