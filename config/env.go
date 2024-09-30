package config

import (
	"github.com/joho/godotenv"
	"writeup-finder.go/utils"
)

func LoadEnv() {
	err := godotenv.Load()

	utils.HandleError(err, "Error loading .env file:", true)
}
