package main

import (
	"writeup-finder.go/command"

	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp:       true,
		DisableLevelTruncation: true,
		FullTimestamp:          false,
	})
}

func main() {
	command.Execute()
}
