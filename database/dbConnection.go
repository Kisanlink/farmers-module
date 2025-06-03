package database

import (
	// "fmt"

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

// InitializeDatabase initializes the PostgreSQL connection and sets the global database instance.
func InitializeDatabase() {
	once.Do(func() {
		// Load environment variables
		config.LoadEnv()

		// Get PostgreSQL connection details
		host := config.GetEnv("DB_HOST")
		port := config.GetEnv("DB_PORT")
		user := config.GetEnv("DB_USER")
		password := config.GetEnv("DB_PASSWORD")
		dbName := config.GetEnv("DB_NAME")
		sslMode := config.GetEnv("DB_SSLMODE")

		// PostgreSQL DSN
		dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s", user, password, host, port, dbName, sslMode)

		// // PostgreSQL DSN
		// dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, dbName, sslMode)

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
