package farm

import (
	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// Farm represents a farm with geographic boundaries
type Farm struct {
	base.BaseModel
	AAAFarmerUserID string            `json:"aaa_farmer_user_id" gorm:"type:varchar(255);not null"`
	AAAOrgID        string            `json:"aaa_org_id" gorm:"type:varchar(255);not null"`
	Name            string            `json:"name" gorm:"type:varchar(255)"`
	Geometry        string            `json:"geometry" gorm:"type:geometry(POLYGON,4326)"`
	AreaHa          float64           `json:"area_ha" gorm:"type:numeric(12,4);->"`
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
	if f.Geometry == "" {
		return common.ErrInvalidFarmGeometry
	}
	return nil
}
