package database

import (
	// "fmt"
	"fmt"
	"sync"

	"github.com/Kisanlink/farmers-module/config"
	"github.com/Kisanlink/farmers-module/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	db_instance *gorm.DB
	once        sync.Once
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
		db_name := config.GetEnv("DB_NAME")
		ssl_mode := config.GetEnv("DB_SSLMODE")

		// PostgreSQL DSN
		dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?ssl_mode=%s", user, password, host, port, db_name, ssl_mode)

		// // PostgreSQL DSN
		// dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s db_name=%s ssl_mode=%s", host, port, user, password, db_name, ssl_mode)

		// Connect to PostgreSQL
		var err error
		db_instance, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			utils.Log.Fatal("Failed to connect to PostgreSQL:", err)
		}
		utils.Log.Info("Connected to PostgreSQL successfully")
	})
}

// GetDatabase returns the global database instance.
func GetDatabase() *gorm.DB {
	if db_instance == nil {
		utils.Log.Fatal("Database connection is not initialized. Call InitializeDatabase first.")
	}
	return db_instance
}
