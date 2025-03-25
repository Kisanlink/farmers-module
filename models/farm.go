package models

import (
	"database/sql/driver"
)
// FarmRequest - Request model for farm registration
type FarmRequest struct {
	// Actor Fields (Optional)
	KisansathiUserID *string `json:"kisansathi_user_id,omitempty" validate:"omitempty,uuid"` 
	
	// Principal Fields (Mandatory)
	FarmerID  string      `json:"farmer_id" validate:"required,uuid"` // UUID of farmer who owns the farm
	Location  [][]float64 `json:"location" validate:"required,min=4"`  // Polygon coordinates [[lat,lon], [lat,lon], ...]
	
	// Farm Metadata
	Area      float64 `json:"area" validate:"required,gt=0"`      // Area in hectares (must be > 0)
	Locality  string  `json:"locality" validate:"required"`       // Village/town name
	CropType  string  `json:"crop_type" validate:"required"`      // Current crop
	IsVerified bool   `json:"is_verified"`                        // Default false for farmer-created
	
	// System Fields (Auto-populated)
	RequestedBy string `json:"-"` // Populated from user-id header
}

// Farm - Database model for storing farm details
type Farm struct {
	Base
	FarmerID     string       `json:"farmer_id" gorm:"type:varchar(36);not null"`
	KisansathiID *string      `json:"kisansathi_id,omitempty" gorm:"type:uuid"`
	Verified     bool         `json:"verified"`
	IsOwner      bool         `json:"is_owner"`
	Location     GeoJSONPolygon `json:"location" gorm:"type:geometry(Polygon);not null"`
	Area         float64      `json:"area"`
	Locality     string       `json:"locality"`
	CurrentCycle string       `json:"current_cycle"`
	OwnerID      string       `json:"owner_id" gorm:"type:uuid;not null"`
}
type GeoJSONPolygon string

// Value converts the GeoJSON to WKT format for PostGIS
func (g GeoJSONPolygon) Value() (driver.Value, error) {
	// The geoJSON parameter will already be properly formatted as a string
	// when passed to the repository
	return string(g), nil
}

// Scan implements the sql.Scanner interface (if you need to read from DB)
func (g *GeoJSONPolygon) Scan(value interface{}) error {
	// Implement if you need to scan from DB
	return nil
}