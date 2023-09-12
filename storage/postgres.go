package storage

import (
	"fmt"
	"log"

	// "net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/joho/godotenv"
	// "golang.org/x/crypto/bcrypt"
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

	dbURI := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbName, dbPassword)

	var err error
	db, err = gorm.Open("postgres", dbURI)
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	// Uncomment the following line if you want to enable detailed SQL logging
	// db.LogMode(true)

	// Auto-migrate your database models here
	// db.AutoMigrate(&Account{}, &Product{}, &Order{}, &CartItem{})
}

func main() {
	app := fiber.New()

	// Middleware: CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
	}))

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	address := ":" + port

	go func() {
		// Wait for a few seconds to ensure the database is initialized
		time.Sleep(5 * time.Second)
		fmt.Printf("Server is running on %s\n", address)
	}()

	log.Fatal(app.Listen(address))
}

// package storage

// import (
// 	"fmt"
// 	"log"
// 	"os"
//

// 	"github.com/jinzhu/gorm"
// 	_ "github.com/jinzhu/gorm/dialects/postgres"
// 	"github.com/joho/godotenv"

// )

// var db *gorm.DB

// func init() {
// 	// Load environment variables from .env file
// 	if err := godotenv.Load(); err != nil {
// 		log.Fatalf("Error loading .env file: %v", err)
// 	}

// 	// Initialize the database connection
// 	dbHost := os.Getenv("DB_HOST")
// 	dbPort := os.Getenv("DB_PORT")
// 	dbUser := os.Getenv("DB_USER")
// 	dbPassword := os.Getenv("DB_PASSWORD")
// 	dbName := os.Getenv("DB_NAME")

// 	dbURI := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
// 		dbHost, dbPort, dbUser, dbName, dbPassword)

// 	var err error
// 	db, err = gorm.Open("postgres", dbURI)
// 	if err != nil {
// 		log.Fatalf("Error connecting to the database: %v", err)
// 	}
// 	// Uncomment the following line if you want to enable detailed SQL logging
// 	db.LogMode(true)
// }

// // GetDB returns the database instance
// func GetDB() *gorm.DB {
// 	return db
// }
