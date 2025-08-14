package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type Farm struct {
	Base
	FarmerId     string         `json:"farmer_id" gorm:"type:varchar(36);not null"`
	KisansathiId *string        `json:"kisansathi_id,omitempty" gorm:"type:uuid;default:null"`
	Verified     bool           `json:"verified"`
	IsOwner      bool           `json:"is_owner"`
	Location     GeoJSONPolygon `json:"location" gorm:"type:geometry(Polygon,4326);not null"`
	Area         float64        `json:"area"`
	Locality     string         `json:"locality"`
	CurrentCycle string         `json:"current_cycle"`
	OwnerId      string         `json:"owner_id" gorm:"type:varchar(36);;default:null"`
	Pincode      int            `json:"pincode"`
}

type GeoJSONPolygon struct {
	Type        string        `json:"type" default:"Polygon"`
	Coordinates [][][]float64 `json:"coordinates"`
}

type GeoJSONPoint struct {
	Type        string    `json:"type" default:"Point"`
	Coordinates []float64 `json:"coordinates"`
}

func (p *GeoJSONPoint) Scan(value interface{}) error {
	if value == nil {
		p.Type = "Point"
		p.Coordinates = make([]float64, 0)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unsupported type for GeoJSONPoint: %T", value)
	}

	return json.Unmarshal(bytes, p)
}

func (p GeoJSONPoint) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Implement a more robust Scan method
func (g *GeoJSONPolygon) Scan(value interface{}) error {
	if value == nil {
		g.Type = "Polygon"
		g.Coordinates = make([][][]float64, 0)
		return nil
	}

	switch v := value.(type) {
	case []byte:
		// Try to unmarshal as GeoJSON first
		if err := json.Unmarshal(v, g); err == nil {
			return nil
		}
		// If not GeoJSON, try to parse as WKB (PostGIS binary format)
		return parsePostGISBinary(v, g)
	case string:
		return json.Unmarshal([]byte(v), g)
	default:
		return fmt.Errorf("unsupported type for GeoJSONPolygon: %T", value)
	}
}

func parsePostGISBinary(data []byte, _ *GeoJSONPolygon) error {
	// You'll need to implement WKB parsing here
	// For now, we'll just return the raw data for debugging
	return fmt.Errorf("received PostGIS binary data: %x", data)
}

func (g GeoJSONPolygon) Value() (driver.Value, error) {
	return json.Marshal(g)
}
