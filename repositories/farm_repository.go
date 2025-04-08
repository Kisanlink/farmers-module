package repositories

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/utils"
	"gorm.io/gorm"
)

type FarmRepository struct {
	db *gorm.DB
}

func NewFarmRepository(db *gorm.DB) *FarmRepository {
	return &FarmRepository{db: db}
}

type FarmRepositoryInterface interface {
	CheckFarmOverlap(geoJSON models.GeoJSONPolygon) (bool, error)
	CreateFarmRecord(farm *models.Farm) error
	GetAllFarms(farmerID, pincode, date string) ([]*models.Farm, error)
	GetFarmsWithFilters(farmerID, pincode string) ([]*models.Farm, error)
}

func (r *FarmRepository) CheckFarmOverlap(geoJSON models.GeoJSONPolygon) (bool, error) {
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

func (r *FarmRepository) CreateFarmRecord(farm *models.Farm) error {
    // Set ID and timestamps on the model
    farm.ID = utils.Generate10DigitID()
    farm.CreatedAt = time.Now()
    farm.UpdatedAt = time.Now()
    
    // Marshal location
    geoJSONBytes, err := json.Marshal(farm.Location)
    if err != nil {
        return fmt.Errorf("failed to marshal location: %v", err)
    }
    
    // Create with raw SQL for location only
    err = r.db.Model(farm).
        Create(map[string]interface{}{
            "id":            farm.ID,
            "farmer_id":     farm.FarmerId,
            "kisansathi_id": farm.KisansathiId,
            "is_owner":      farm.IsOwner,
            "location":      gorm.Expr("ST_SetSRID(ST_GeomFromGeoJSON(?),4326)", string(geoJSONBytes)),
            "area":          farm.Area,
            "locality":      farm.Locality,
            "owner_id":      farm.OwnerId,
            "pincode":       farm.Pincode,
            "created_at":    farm.CreatedAt,
            "updated_at":    farm.UpdatedAt,
        }).Error
    
    if err != nil {
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

// Implement the methods in FarmRepository
func (r *FarmRepository) GetAllFarms(farmerID, pincode, date string) ([]*models.Farm, error) {
    var farms []*models.Farm

    // Build the base query
    query := `
        SELECT 
            id,
            farmer_id,
            kisansathi_id,
            verified,
            is_owner,
            ST_AsGeoJSON(location)::jsonb as location,
            area,
            pincode,
            locality,
            current_cycle,
            owner_id,
            created_at
        FROM farms
        WHERE 1=1
    `

    // Add filters dynamically
    var args []interface{}
    if farmerID != "" {
        query += " AND farmer_id = ?"
        args = append(args, farmerID)
    }
    if pincode != "" {
        query += " AND pincode = ?"
        args = append(args, pincode)
    }
    if date != "" {
        // Filter for a specific date (ignoring time portion)
        query += " AND DATE(created_at) = ?"
        args = append(args, date)
    }

    // Execute the query with filters
    err := r.db.Raw(query, args...).Scan(&farms).Error
    if err != nil {
        return nil, fmt.Errorf("database error: %w", err)
    }

    return farms, nil
}
func (r *FarmRepository) GetFarmsWithFilters(farmerID, pincode string) ([]*models.Farm, error) {
	var farms []*models.Farm
	query := r.db.Model(&models.Farm{})

	// Apply filters if query parameters are provided
	if farmerID != "" {
		query = query.Where("farmer_id = ?", farmerID)
	}
	if pincode != "" {
		query = query.Where("pincode = ?", pincode)
	}

	// Execute the query
	if err := query.Find(&farms).Error; err != nil {
		return nil, err
	}
	return farms, nil
}
