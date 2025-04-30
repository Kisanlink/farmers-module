package repositories

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/utils"
	"gorm.io/gorm"
)

type FarmRepository struct {
	DB *gorm.DB
}

func NewFarmRepository(db *gorm.DB) *FarmRepository {
	return &FarmRepository{DB: db}
}

type FarmRepositoryInterface interface {
	CheckFarmOverlap(geoJSON models.GeoJSONPolygon) (bool, error)
	CreateFarmRecord(farm *models.Farm) error
	GetAllFarms(farmer_id, pincode, date, id string) ([]*models.Farm, error)
	GetFarmsWithFilters(farmer_id, pincode string) ([]*models.Farm, error)
	// New method to get a farm by its ID
	GetFarmByID(farm_id string) (*models.Farm, error)

	GetFarmAreaByID(farm_id string) (float64, error)
}

func (r *FarmRepository) CheckFarmOverlap(geoJSON models.GeoJSONPolygon) (bool, error) {
	var count int64

	// Convert GeoJSON to JSON using the Value() method
	val, err := geoJSON.Value()
	if err != nil {
		utils.Log.Error("Failed to get GeoJSON value", "error", err.Error()) // Replaced log with utils.Log
		return false, fmt.Errorf("failed to get GeoJSON value: %w", err)
	}

	// Type assert the driver.Value to []byte
	geoJSONBytes, ok := val.([]byte)
	if !ok {
		utils.Log.Error("Expected []byte from GeoJSON value, got unexpected type", "expected_type", "[]byte", "actual_type", fmt.Sprintf("%T", val)) // Replaced log with utils.Log
		return false, fmt.Errorf("expected []byte from GeoJSON value, got %T", val)
	}

	// Build the query using GORM's expression builder
	query := r.DB.
		Model(&models.Farm{}).
		Where(
			"ST_Intersects(location, ST_GeomFromGeoJSON(?))",
			string(geoJSONBytes), // Now safely converted to string
		)

	// Count overlapping farms
	err = query.Count(&count).Error
	if err != nil {
		utils.Log.Error("CheckFarmOverlap query failed", "error", err.Error()) // Replaced log with utils.Log
		return false, fmt.Errorf("error checking farm overlap: %w", err)
	}

	return count > 0, nil
}

func (r *FarmRepository) CreateFarmRecord(farm *models.Farm) error {
	// Set Id and timestamps on the model
	farm.Id = utils.Generate10DigitId()
	farm.CreatedAt = time.Now()
	farm.UpdatedAt = time.Now()

	// Marshal location
	geoJSONBytes, err := json.Marshal(farm.Location)
	if err != nil {
		utils.Log.Error("Failed to marshal farm location", "error", err.Error()) // Replaced log with utils.Log
		return fmt.Errorf("failed to marshal location: %v", err)
	}

	// Create with raw SQL for location only
	err = r.DB.Model(farm).
		Create(map[string]interface{}{
			"id":            farm.Id,
			"farmer_id":     farm.FarmerId,
			"kisansathi_id": farm.KisansathiId,
			"is_owner":      farm.IsOwner,
			"location":      gorm.Expr("ST_SetSRId(ST_GeomFromGeoJSON(?),4326)", string(geoJSONBytes)),
			"area":          farm.Area,
			"locality":      farm.Locality,
			"owner_id":      farm.OwnerId,
			"pincode":       farm.Pincode,
			"created_at":    farm.CreatedAt,
			"updated_at":    farm.UpdatedAt,
		}).Error

	if err != nil {
		utils.Log.Error("Failed to create farm record", "error", err.Error()) // Replaced log with utils.Log
		return fmt.Errorf("failed to create farm: %v", err)
	}
	return nil
}

// Implement the methods in FarmRepository
func (r *FarmRepository) GetAllFarms(farmer_id, pincode, date, id string) ([]*models.Farm, error) {
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
            created_at,
            updated_at
        FROM farms
        WHERE 1=1
    `

	// Add filters dynamically
	var args []interface{}
	if farmer_id != "" {
		query += " AND farmer_id = ?"
		args = append(args, farmer_id)
	}
	if pincode != "" {
		query += " AND pincode = ?"
		args = append(args, pincode)
	}
	if date != "" {
		// Filter for a specific date (ignoring time portion)
		query += " AND DATE(created_at) >= ?"
		args = append(args, date)
	}
	if id != "" {
		// Filter for a specific Id
		query += " AND id = ?"
		args = append(args, id)
	}

	// Execute the query with filters
	err := r.DB.Raw(query, args...).Scan(&farms).Error
	if err != nil {
		utils.Log.Error("Database error when retrieving all farms", "error", err.Error()) // Replaced log with utils.Log
		return nil, fmt.Errorf("database error: %w", err)
	}

	return farms, nil
}

func (r *FarmRepository) GetFarmsWithFilters(farmer_id, pincode string) ([]*models.Farm, error) {
	var farms []*models.Farm
	query := r.DB.Model(&models.Farm{})

	// Apply filters if query parameters are provided
	if farmer_id != "" {
		query = query.Where("farmer_id = ?", farmer_id)
	}
	if pincode != "" {
		query = query.Where("pincode = ?", pincode)
	}

	// Execute the query
	if err := query.Find(&farms).Error; err != nil {
		utils.Log.Error("Error fetching farms with filters", "farmer_id", farmer_id, "pincode", pincode, "error", err.Error()) // Replaced log with utils.Log
		return nil, err
	}
	return farms, nil
}

func (r *FarmRepository) GetFarmByID(farm_id string) (*models.Farm, error) {
	var farm models.Farm

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
            created_at,
            updated_at
        FROM farms
        WHERE id = ? 
    `

	if err := r.DB.Raw(query, farm_id).Scan(&farm).Error; err != nil {
		utils.Log.Error("Failed to retrieve farm by ID", "farm_id", farm_id, "error", err.Error()) // Replaced log with utils.Log
		return nil, fmt.Errorf("failed to retrieve farm: %w", err)
	}

	// 🚨 Check if the ID is still empty => no rows were found
	if farm.Base.Id == "" {
		utils.Log.Warn("Farm not found", "farm_id", farm_id) // Added a warning log
		return nil, fmt.Errorf("farm not found")
	}

	return &farm, nil
}

func (r *FarmRepository) GetFarmAreaByID(farm_id string) (float64, error) {
	var area float64
	err := r.DB.
		Model(&models.Farm{}).
		Select("area").
		Where("id = ?", farm_id).
		Scan(&area).Error

	if err != nil {
		utils.Log.Error("Failed to get farm area by ID", "farm_id", farm_id, "error", err.Error()) // Replaced log with utils.Log
		return 0, fmt.Errorf("failed to get farm area: %w", err)
	}
	return area, nil
}
