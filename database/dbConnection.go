package database

import (
	"fmt"
	"log"
	"sync"

	"github.com/Kisanlink/farmers-module/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	dbInstance *gorm.DB
	once       sync.Once
)

func InitializeDatabase() {
	once.Do(func() {
		// Load environment variables
		config.LoadEnv()

		// Get individual database connection variables from the environment
		host := config.GetEnv("DB_HOST")
		user := config.GetEnv("DB_USER")
		password := config.GetEnv("DB_PASSWORD")
		dbname := config.GetEnv("DB_NAME")
		port := config.GetEnv("DB_PORT")
		sslmode := config.GetEnv("DB_SSLMODE")

		// Check if essential variables are set
		if host == "" || user == "" || dbname == "" || port == "" {
			log.Fatal("Error: One or more required database environment variables (DB_HOST, DB_USER, DB_NAME, DB_PORT) are not set.")
		}

		// Construct the DSN (Data Source Name) string
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
			host, user, password, dbname, port, sslmode)

		// Connect to PostgreSQL
		var err error
		dbInstance, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatalf("Failed to connect to PostgreSQL: %v", err)
		}
		log.Println("Connected to PostgreSQL successfully")
		RunMigrations()
	})
}

// GetDatabase returns the global database instance.
func GetDatabase() *gorm.DB {
	if dbInstance == nil {
		log.Fatal("Database connection is not initialized. Call InitializeDatabase first.")
	}
	return dbInstance
}
