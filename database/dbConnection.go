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

// InitializeDatabase sets up a singleton *gorm.DB.
func InitializeDatabase() {
	once.Do(func() {
		// 1. Load environment variables (.env and/or OS)
		config.LoadEnv()

		// 2. Prefer a pre-built DSN (simplest for managed hosts like Neon)
		dsn := config.GetEnv("DATABASE_URL")

		// 3. Connect via GORM
		var err error
		dbInstance, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatalf("failed to connect to PostgreSQL: %v", err)
		}
		log.Println("connected to PostgreSQL successfully")

		// 4. Auto-migrate (if you need it)
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
