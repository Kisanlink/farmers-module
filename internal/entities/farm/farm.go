package farm

import (
	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// Geometry represents a geometric shape (PostGIS)
type Geometry struct {
	WKT string // Well-Known Text format
	WKB []byte // Well-Known Binary format
}

// Farm represents a farm with geographic boundaries
type Farm struct {
	base.BaseModel
	AAAFarmerUserID string            `json:"aaa_farmer_user_id" gorm:"type:varchar(255);not null"`
	AAAOrgID        string            `json:"aaa_org_id" gorm:"type:varchar(255);not null"`
	Geometry        Geometry          `json:"geometry" gorm:"type:geometry(Polygon,4326);not null"`
	AreaHa          float64           `json:"area_ha" gorm:"type:numeric(12,4);generated"`
	Metadata        map[string]string `json:"metadata" gorm:"type:jsonb;default:'{}'"`
}

// TableName returns the table name for the Farm model
func (f *Farm) TableName() string {
	return "farms"
}

// GetTableIdentifier returns the table identifier for ID generation
func (f *Farm) GetTableIdentifier() string {
	return "farm"
}

// GetTableSize returns the table size for ID generation
func (f *Farm) GetTableSize() hash.TableSize {
	return hash.Medium
}

// Validation methods
func (f *Farm) Validate() error {
	if f.AAAFarmerUserID == "" {
		return common.ErrInvalidFarmData
	}
	if f.AAAOrgID == "" {
		return common.ErrInvalidFarmData
	}
	if f.Geometry.WKT == "" && len(f.Geometry.WKB) == 0 {
		return common.ErrInvalidFarmGeometry
	}
	return nil
}

// Helper methods for geometry
func (g *Geometry) ToWKT() string {
	return g.WKT
}

func (g *Geometry) ToWKB() []byte {
	return g.WKB
}

func (g *Geometry) IsValid() bool {
	return g.WKT != "" || len(g.WKB) > 0
}
