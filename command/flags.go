package command

import (
	log "github.com/sirupsen/logrus"
	"writeup-finder.go/global"
)

// ManageFlags validates and logs the parsed flags.
func ManageFlags() {
	ValidateFlags()

	log.Infof("[+] Use Database: %v", global.UseDatabase)
	log.Infof("[+] Send to Telegram: %v", global.SendToTelegramFlag)

	if global.ProxyURL != "" {
		log.Infof("[+] Proxy URL: %v", global.ProxyURL)
	} else {
		log.Info("[+] No Proxy URL set.")
	}
}

// ValidateFlags ensures that flag combinations are valid and throws errors for invalid input.
func ValidateFlags() {
	if !global.UseDatabase {
		log.Fatal("You must specify --database to save articles in the database.")
	}

	if global.ProxyURL != "" && !global.SendToTelegramFlag {
		log.Fatal("Error: --proxy option is only valid when used with --telegram.")
	}
}
