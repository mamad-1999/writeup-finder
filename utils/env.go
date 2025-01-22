package utils

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// LoadEnv loads environment variables from a `.env` file located in the specified directory.
// If the `.env` file cannot be loaded, it logs a fatal error and exits the program.
// The `.env` file path is determined by joining the `GITHUB_WORKSPACE` environment variable with `.env`.
func LoadEnv() {
	envFile := filepath.Join(os.Getenv("GITHUB_WORKSPACE"), ".env")
	err := godotenv.Load(envFile)

	HandleError(err, "Error loading .env file:", true)
}
