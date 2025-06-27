package database

import (
	"fmt"
	"log"

	"github.com/Kisanlink/farmers-module/models"
)

// RunMigrations applies migrations for all models
func RunMigrations() {
	db := GetDatabase()

	// Auto migrate all models
	err := db.AutoMigrate(
		&models.Farmer{},
		&models.Farm{},
		&models.FarmActivity{},
		&models.CropCycle{},
		&models.Crop{},
		&models.FPO{},
	)
	if err != nil {
		log.Fatalf("❌ Failed to run migrations: %v", err)
	}

	fmt.Println("✅ Database migrations completed successfully!")
}
