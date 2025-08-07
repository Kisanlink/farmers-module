package database

import (
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

		dsn := config.GetEnv("DATABASE_URL")

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
