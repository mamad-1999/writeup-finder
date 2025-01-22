package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq" // Postgres driver
	"github.com/sirupsen/logrus"
	"writeup-finder.go/utils"
)

// ConnectDB establishes a connection to the PostgreSQL database using environment variables.
// It returns a pointer to the sql.DB object or logs a fatal error if the connection fails.
func ConnectDB() *sql.DB {
	connStr := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=require",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"),
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"))

	db, err := sql.Open("postgres", connStr)
	utils.HandleError(err, "Error in Connection to DB", true)

	logrus.Info("[+] Database connection established.")
	return db
}

// SaveUrlToDB inserts a URL and its corresponding title into the articles table.
// It logs an error if the operation fails but does not stop the program execution.
func SaveUrlToDB(db *sql.DB, url, title string) {
	_, err := db.Exec("INSERT INTO articles (url, title) VALUES ($1, $2)", url, title)
	utils.HandleError(err, "Error saving URL and title to database", false)
}

// CreateArticlesTable creates the articles table if it does not already exist.
// The table includes columns for id (primary key), url, and title.
// It logs a fatal error if the table creation fails.
func CreateArticlesTable(db *sql.DB) {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS articles (
			id SERIAL PRIMARY KEY, 
			url VARCHAR(1000), 
			title VARCHAR(1000)
		);
	`)

	utils.HandleError(err, "Error creating articles table", true)
	logrus.Info("[+] Articles table created successfully.")
}
