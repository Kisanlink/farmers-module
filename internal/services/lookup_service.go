package services

import (
	"context"

	"github.com/Kisanlink/farmers-module/internal/entities/irrigation_source"
	"github.com/Kisanlink/farmers-module/internal/entities/soil_type"
	"gorm.io/gorm"
)

// LookupService handles lookup data operations
type LookupService interface {
	InitializeLookupData(ctx context.Context) error
	GetSoilTypes(ctx context.Context) ([]soil_type.SoilType, error)
	GetIrrigationSources(ctx context.Context) ([]irrigation_source.IrrigationSource, error)
}

// LookupServiceImpl implements LookupService
type LookupServiceImpl struct {
	db *gorm.DB
}

// NewLookupService creates a new lookup service
func NewLookupService(db *gorm.DB) LookupService {
	return &LookupServiceImpl{
		db: db,
	}
}

// InitializeLookupData initializes predefined lookup data
func (s *LookupServiceImpl) InitializeLookupData(ctx context.Context) error {
	// Initialize soil types
	for _, soilType := range soil_type.PredefinedSoilTypes {
		var existing soil_type.SoilType
		if err := s.db.Where("name = ?", soilType.Name).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := s.db.Create(&soilType).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	// Initialize irrigation sources
	for _, irrigationSource := range irrigation_source.PredefinedIrrigationSources {
		var existing irrigation_source.IrrigationSource
		if err := s.db.Where("name = ?", irrigationSource.Name).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := s.db.Create(&irrigationSource).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	return nil
}

// GetSoilTypes retrieves all soil types
func (s *LookupServiceImpl) GetSoilTypes(ctx context.Context) ([]soil_type.SoilType, error) {
	var soilTypes []soil_type.SoilType
	if err := s.db.Find(&soilTypes).Error; err != nil {
		return nil, err
	}
	return soilTypes, nil
}

// GetIrrigationSources retrieves all irrigation sources
func (s *LookupServiceImpl) GetIrrigationSources(ctx context.Context) ([]irrigation_source.IrrigationSource, error) {
	var irrigationSources []irrigation_source.IrrigationSource
	if err := s.db.Find(&irrigationSources).Error; err != nil {
		return nil, err
	}
	return irrigationSources, nil
}
