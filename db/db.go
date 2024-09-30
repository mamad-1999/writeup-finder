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

	fmt.Println("[+] Database Successfully Established.")

	return db, nil
}

func SaveUrlToDB(db *sql.DB, url, title string) {
	_, err := db.Exec("INSERT INTO articles (url, title) VALUES ($1, $2)", url, title)
	if err != nil {
		utils.HandleError(err, "Error saving URL and title to database", false)
	}
}

func CreateArticlesTable(db *sql.DB) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS articles (
			id SERIAL PRIMARY KEY, 
			url VARCHAR(1000), 
			title VARCHAR(1000)
		);
	`)
	if err != nil {
		utils.HandleError(err, "Error creating articles table", true)
		return
	}

	fmt.Println("[+] Articles table created successfully.")
}
