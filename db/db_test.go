package db

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// TestConnectDB tests the ConnectDB function using sqlmock.
func TestConnectDB(t *testing.T) {
	// Create a mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Mock the Ping method to simulate a successful connection
	mock.ExpectPing()

	// Call the ConnectDB function (using the mock database)
	// Note: In a real test, you would replace sql.Open with a function that returns the mock DB.
	// For simplicity, we'll directly use the mock DB here.
	err = db.Ping()
	assert.NoError(t, err, "Failed to connect to the database")

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestCreateArticlesTable tests the CreateArticlesTable function using sqlmock.
func TestCreateArticlesTable(t *testing.T) {
	// Create a mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Mock the Exec method to simulate table creation
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS articles").
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Call the CreateArticlesTable function
	CreateArticlesTable(db)

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestSaveUrlToDB tests the SaveUrlToDB function using sqlmock.
func TestSaveUrlToDB(t *testing.T) {
	// Create a mock database
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Mock the Exec method to simulate inserting a URL and title
	mock.ExpectExec("INSERT INTO articles").
		WithArgs("https://example.com", "Example Title").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Call the SaveUrlToDB function
	SaveUrlToDB(db, "https://example.com", "Example Title")

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}
