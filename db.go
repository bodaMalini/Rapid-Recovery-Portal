package database

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

// InitDB initializes the database connection
func InitDB() {
	var err error
	// Use the local MySQL connection string for local development
	DB, err = sql.Open("mysql", "root:sunny@2006@tcp(localhost:3306)/project_backend")
	if err != nil {
		log.Fatal("Failed to connect to MySQL database:", err)
	}

	// Setting connection pooling options
	DB.SetMaxOpenConns(10)
	DB.SetMaxIdleConns(5)
	log.Println("Connected to MySQL database")

	// Ensure the database connection is valid by pinging the DB
	if err := DB.Ping(); err != nil {
		log.Fatal("Error pinging MySQL database:", err)
	}
}
