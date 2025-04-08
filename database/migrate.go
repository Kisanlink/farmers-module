package database

import (
	"fmt"
	"log"

	"github.com/Kisanlink/farmers-module/models"
)

// RunMigrations applies migrations for all models
func RunMigrations() {
	db := GetDatabase()

	// Auto migrate the Farmer model
	err := db.AutoMigrate(&models.Farmer{}, &models.Farm{})
	if err != nil {
		log.Fatalf("❌ Failed to run migrations: %v", err)
	}

	fmt.Println("✅ Database migrations completed successfully!")
}

