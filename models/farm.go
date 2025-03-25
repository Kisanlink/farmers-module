package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
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
	FarmerId    string       `json:"farmer_id" gorm:"type:varchar(36);not null"`
	KisansathiId *string      `json:"kisansathi_id,omitempty" gorm:"type:uuid"`
	Verified     bool         `json:"verified"`
	IsOwner      bool         `json:"is_owner"`
	Location     GeoJSONPolygon `json:"location" gorm:"type:geometry(Polygon);not null"`
	Area         float64      `json:"area"`
	Locality     string       `json:"locality"`
	CurrentCycle string       `json:"current_cycle"`
	OwnerId      string       `json:"owner_id" gorm:"type:uuid;not null"`
}

// GeoJSONPolygon represents a GeoJSON Polygon.
type GeoJSONPolygon struct {
    Type        string          `json:"type"`       // should be "Polygon"
    Coordinates [][][]float64   `json:"coordinates"` // array of linear rings
}

// Value marshals the GeoJSONPolygon into JSON for storage.
func (g GeoJSONPolygon) Value() (driver.Value, error) {
    return json.Marshal(g)
}

// Scan unmarshals a JSON-encoded value from the database into GeoJSONPolygon.
func (g *GeoJSONPolygon) Scan(value interface{}) error {
    bytes, ok := value.([]byte)
    if !ok {
        return fmt.Errorf("failed to convert value to []byte")
    }
    return json.Unmarshal(bytes, g)
}