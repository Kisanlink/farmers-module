package db

import (
	farmEntity "github.com/Kisanlink/farmers-module/internal/entities/farm"
	"github.com/Kisanlink/farmers-module/internal/entities/farm_irrigation_source"
	"github.com/Kisanlink/farmers-module/internal/entities/farm_soil_type"
	"github.com/Kisanlink/farmers-module/internal/entities/irrigation_source"
	"github.com/Kisanlink/farmers-module/internal/entities/soil_type"
	"gorm.io/gorm"
)

// MigrationService handles database migrations
type MigrationService struct {
	db *gorm.DB
}

// NewMigrationService creates a new migration service
func NewMigrationService(db *gorm.DB) *MigrationService {
	return &MigrationService{
		db: db,
	}
}

// AutoMigrate runs auto migration for all models
func (m *MigrationService) AutoMigrate() error {
	// Migrate in the correct order to handle foreign key dependencies
	models := []interface{}{
		&soil_type.SoilType{},
		&irrigation_source.IrrigationSource{},
		&farmEntity.Farm{},
		&farm_irrigation_source.FarmIrrigationSource{},
		&farm_soil_type.FarmSoilType{},
	}

	for _, model := range models {
		if err := m.db.AutoMigrate(model); err != nil {
			return err
		}
	}

	return nil
}
