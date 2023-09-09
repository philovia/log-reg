package storage

import (
	"fmt"
	"log"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/joho/godotenv"
)

var db *gorm.DB

func init() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Initialize the database connection
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	dbURI := fmt.Sprintf("localhost=%s 5432=%s postgres=%s postgres=%s postgres=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbName, dbPassword)

	var err error
	db, err = gorm.Open("postgres", dbURI)
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	// Uncomment the following line if you want to enable detailed SQL logging
	// db.LogMode(true)
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return db
}
