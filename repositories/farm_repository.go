package repositories

import (
	"context"
	"fmt"
	"log"

	"github.com/Kisanlink/farmers-module/models"
	"gorm.io/gorm"
)

type FarmRepository struct {
	db *gorm.DB
}

func NewFarmRepository(db *gorm.DB) *FarmRepository {
	return &FarmRepository{db: db}
}

type FarmRepositoryInterface interface {
	CheckFarmOverlap(ctx context.Context, geoJSON map[string]interface{}) (bool, error)
	CreateFarmRecord(ctx context.Context, farm *models.Farm, geoJSON map[string]interface{}) error
}

func (r *FarmRepository) CheckFarmOverlap(ctx context.Context, geoJSON map[string]interface{}) (bool, error) {
	var exists bool
	err := r.db.WithContext(ctx).
		Raw(`
			SELECT EXISTS(
				SELECT 1 FROM farms 
				WHERE ST_Intersects(
					location, 
					ST_GeomFromGeoJSON(?)
				)
			)`, geoJSON).
		Scan(&exists).Error

	if err != nil {
		log.Printf("CheckFarmOverlap query failed: %v", err)
		return false, fmt.Errorf("error checking farm overlap")
	}
	return exists, nil
}

func (r *FarmRepository) CreateFarmRecord(ctx context.Context, farm *models.Farm, geoJSON map[string]interface{}) error {
	// Convert the geoJSON to a string representation
	geoJSONStr := fmt.Sprintf(`{"type":"Polygon","coordinates":%v}`, geoJSON["coordinates"])
	farm.Location = models.GeoJSONPolygon(geoJSONStr)
	
	// Use standard GORM Create
	err := r.db.WithContext(ctx).Create(farm).Error
	if err != nil {
		log.Printf("CreateFarmRecord failed: %v", err)
		return fmt.Errorf("failed to create farm: %v", err)
	}
	return nil
}