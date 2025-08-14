package database

import (
	"fmt"
	"log"
	"strings"

	"github.com/Kisanlink/farmers-module/models"
)

// RunMigrations applies migrations for all models
func RunMigrations() {
	db := GetDatabase()

	// Auto migrate all models except CropStage first
	err := db.AutoMigrate(
		&models.Farmer{},
		&models.Farm{},
		&models.FarmActivity{},
		&models.CropCycle{},
		&models.Crop{},
		&models.Stage{},
		&models.FPO{},
	)
	if err != nil {
		log.Fatalf("❌ Failed to run migrations: %v", err)
	}

	// Handle CropStage migration separately to avoid conflicts
	err = db.AutoMigrate(&models.CropStage{})
	if err != nil {
		// Check if it's just the column exists error
		if strings.Contains(err.Error(), "already exists") {
			log.Printf("⚠️ Migration note: %v", err)
			log.Println("Column already exists, continuing...")
		} else {
			log.Fatalf("❌ Failed to run CropStage migration: %v", err)
		}
	}

	fmt.Println("✅ Database migrations completed successfully!")
}
