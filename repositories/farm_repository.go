package repositories

import (
	"encoding/json"
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
func (r *FarmRepository) CreateFarmRecord(farm *models.Farm) error {
    // Marshal the GeoJSONPolygon to a JSON string.
    geoJSONBytes, err := json.Marshal(farm.Location)
    if err != nil {
        log.Printf("failed to marshal location: %v", err)
        return fmt.Errorf("failed to marshal location: %v", err)
    }
    geoJSONString := string(geoJSONBytes)

    // Build a map for insertion so we can use a raw SQL expression for the location field.
    farmData := map[string]interface{}{
        "farmer_id":     farm.FarmerId,
        "kisansathi_id": farm.KisansathiId,
        "verified":      farm.Verified,
        "is_owner":      farm.IsOwner,
        "location":      gorm.Expr("ST_SetSRID(ST_GeomFromGeoJSON(?),4326)", geoJSONString),
        "area":          farm.Area,
        "locality":      farm.Locality,
        "current_cycle": farm.CurrentCycle,
        "owner_id":      farm.OwnerId,
    }

    err = r.db.Model(&models.Farm{}).Create(farmData).Error
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