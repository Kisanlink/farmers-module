package repositories

import (
	"fmt"
	"log"
    "strings"

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
	CheckFarmOverlap( geoJSON models.GeoJSONPolygon) (bool, error)
	CreateFarmRecord(farm *models.Farm) error
  
}

func (r *FarmRepository) CheckFarmOverlap( geoJSON models.GeoJSONPolygon) (bool, error) {
    var count int64

    // Convert GeoJSON to JSON using the Value() method
    val, err := geoJSON.Value()
    if err != nil {
        return false, fmt.Errorf("failed to get GeoJSON value: %w", err)
    }

    // Type assert the driver.Value to []byte
    geoJSONBytes, ok := val.([]byte)
    if !ok {
        return false, fmt.Errorf("expected []byte from GeoJSON value, got %T", val)
    }

    // Build the query using GORM's expression builder
    query := r.db.
        Model(&models.Farm{}).
        Where(
            "ST_Intersects(location, ST_GeomFromGeoJSON(?))", 
            string(geoJSONBytes), // Now safely converted to string
        )

    // Count overlapping farms
    err = query.Count(&count).Error
    if err != nil {
        log.Printf("CheckFarmOverlap query failed: %v", err)
        return false, fmt.Errorf("error checking farm overlap: %w", err)
    }

    return count > 0, nil
}

// Alternative solution if you must use map[string]interface{}
// repositories/farm_repository.go

// repositories/farm_repository.go
func (r *FarmRepository) CreateFarmRecord( farm *models.Farm) error {
    // Convert GeoJSONPolygon to map
    // Use standard GORM Create
    err := r.db.Create(farm).Error
    if err != nil {
        log.Printf("CreateFarmRecord failed: %v", err)
        return fmt.Errorf("failed to create farm: %v", err)
    }
    return nil
}

func convertGeoJSONToWKT(geoJSON models.GeoJSONPolygon) string {
    if len(geoJSON.Coordinates) == 0 || len(geoJSON.Coordinates[0]) == 0 {
        return "POLYGON EMPTY"
    }
    
    var points []string
    for _, coord := range geoJSON.Coordinates[0] {
        points = append(points, fmt.Sprintf("%f %f", coord[0], coord[1]))
    }
    
    // Ensure polygon is closed
    if len(points) > 0 && points[0] != points[len(points)-1] {
        points = append(points, points[0])
    }
    
    return fmt.Sprintf("POLYGON((%s))", strings.Join(points, ", "))
}