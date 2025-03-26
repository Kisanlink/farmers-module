package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// FarmRequest - Request model for farm registration
type FarmRequest struct {
	KisansathiUserID *string `json:"kisansathi_user_id,omitempty"` // Remove UUID validation if not needed
	FarmerID        string  `json:"farmer_id" validate:"required"` // Changed from UUID/numeric to simple string
	Location        [][][]float64 `json:"location" validate:"required,min=4"`
	Area            float64 `json:"area" validate:"required,gt=0"`
	Locality        string  `json:"locality" validate:"required"`
	CropType        string  `json:"crop_type" validate:"required"`
	IsVerified      bool    `json:"is_verified"`
	RequestedBy     string  `json:"-"`
}

// models/farm.go
type Farm struct {
	Base
	FarmerId     string         `json:"farmer_id" gorm:"type:varchar(36);not null"`
	KisansathiId *string        `json:"kisansathi_id,omitempty" gorm:"type:uuid;default:null"` // Changed to pointer and default null
	Verified     bool           `json:"verified"`
	IsOwner      bool           `json:"is_owner"`
	Location     GeoJSONPolygon `json:"location" gorm:"type:geometry(Polygon,4326);not null"`
	Area         float64        `json:"area"`
	Locality     string         `json:"locality"`
	CurrentCycle string         `json:"current_cycle"`
	OwnerId      string         `json:"owner_id" gorm:"type:uuid;not null"`
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