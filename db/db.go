package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"writeup-finder.go/utils"
)

// ConnectDB establishes a connection to the PostgreSQL database using environment variables.
func ConnectDB() (*sql.DB, error) {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	if dbUser == "" || dbPassword == "" || dbHost == "" || dbPort == "" || dbName == "" {
		return nil, fmt.Errorf("one or more environment variables are not set")
	}

	connStr := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening database connection: %w", err)
	}

	// Ping the database to ensure connectivity
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging database: %w", err)
	}

	return db, nil
}

func ReadFoundUrlsFromDB(db *sql.DB) (map[string]struct{}, error) {
	foundUrls := make(map[string]struct{})

	rows, err := db.Query("SELECT url FROM articles")
	if err != nil {
		utils.HandleError(err, "Error querying database", true)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			utils.HandleError(err, "Error scanning row in database", false)
			continue
		}
		foundUrls[url] = struct{}{}
	}

	if err := rows.Err(); err != nil {
		utils.HandleError(err, "Error iterating over database rows", false)
		return nil, err
	}

	return foundUrls, nil
}

func SaveUrlToDB(db *sql.DB, url string) {
	_, err := db.Exec("INSERT INTO articles (url) VALUES ($1)", url)
	if err != nil {
		utils.HandleError(err, "Error saving URL to database", false)
	}
}
