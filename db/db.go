package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq" // Postgres driver
	"github.com/sirupsen/logrus"
	"writeup-finder.go/utils"
)

func ConnectDB() *sql.DB {
	connStr := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"),
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"))

	db, err := sql.Open("postgres", connStr)
	utils.HandleError(err, "Error in Connection to DB", true)

	logrus.Info("[+] Database connection established.")
	return db
}

func SaveUrlToDB(db *sql.DB, url, title string) {
	_, err := db.Exec("INSERT INTO articles (url, title) VALUES ($1, $2)", url, title)
	utils.HandleError(err, "Error saving URL and title to database", false)
}

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
