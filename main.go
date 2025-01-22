package main

import (
	"writeup-finder.go/command"

	_ "github.com/lib/pq" // Import PostgreSQL driver for database/sql
	log "github.com/sirupsen/logrus"
)

// init configures the logrus logger format.
// It disables timestamps, level truncation, and full timestamps for cleaner log output.
func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp:       true,  // Remove timestamps from log output
		DisableLevelTruncation: true,  // Prevent truncation of log levels
		FullTimestamp:          false, // Do not use full timestamps
	})
}

// main is the entry point of the application.
// It calls the Execute function from the command package to start the application.
func main() {
	command.Execute()
}
