package config

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"writeup-finder.go/utils"
)

func LoadEnv() {
	// err := godotenv.Load("/home/mohammad/Videos/go/writeup-finder/.env")

	envFile := filepath.Join(os.Getenv("GITHUB_WORKSPACE"), ".env")
	err := godotenv.Load(envFile)

	utils.HandleError(err, "Error loading .env file:", true)
}
