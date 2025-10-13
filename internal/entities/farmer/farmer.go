package farmer

import (
	"github.com/Kisanlink/farmers-module/internal/entities"
	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// FarmerLink represents the link between a farmer and an FPO
type FarmerLink struct {
	base.BaseModel
	AAAUserID        string  `json:"aaa_user_id" gorm:"type:varchar(255);not null"`
	AAAOrgID         string  `json:"aaa_org_id" gorm:"type:varchar(255);not null"`
	KisanSathiUserID *string `json:"kisan_sathi_user_id" gorm:"type:varchar(255)"`
	Status           string  `json:"status" gorm:"type:link_status;not null;default:'ACTIVE'"`
}

// TableName returns the table name for the FarmerLink model
func (fl *FarmerLink) TableName() string {
	return "farmer_links"
}

// GetTableIdentifier returns the table identifier for ID generation
func (fl *FarmerLink) GetTableIdentifier() string {
	return "FMLK"
}

// GetTableSize returns the table size for ID generation
func (fl *FarmerLink) GetTableSize() hash.TableSize {
	return hash.Medium
}

// FarmerProfile represents a farmer's profile with linked farms
type FarmerProfile struct {
	AAAUserID        string  `json:"aaa_user_id"`
	AAAOrgID         string  `json:"aaa_org_id"`
	KisanSathiUserID *string `json:"kisan_sathi_user_id,omitempty"`
	Farms            []Farm  `json:"farms"`
}

// Farm represents a minimal farm reference for the farmer profile
type Farm struct {
	ID        string                 `json:"id"`
	AAAUserID string                 `json:"aaa_user_id"`
	AAAOrgID  string                 `json:"aaa_org_id"`
	AreaHa    float64                `json:"area_ha"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt string                 `json:"created_at"`
	UpdatedAt string                 `json:"updated_at"`
}

// FarmerLegacy represents the legacy farmer model with denormalized address fields
// DEPRECATED: This model is deprecated. Use Farmer (from farmer_normalized.go) instead.
// This model embeds address fields directly instead of using a foreign key relationship.
type FarmerLegacy struct {
	base.BaseModel
	AAAUserID         string         `json:"aaa_user_id" gorm:"type:varchar(255);not null;uniqueIndex:idx_farmer_legacy_unique"`
	AAAOrgID          string         `json:"aaa_org_id" gorm:"type:varchar(255);not null;uniqueIndex:idx_farmer_legacy_unique"`
	KisanSathiUserID  *string        `json:"kisan_sathi_user_id" gorm:"type:varchar(255)"`
	FirstName         string         `json:"first_name" gorm:"type:varchar(255);not null"`
	LastName          string         `json:"last_name" gorm:"type:varchar(255);not null"`
	PhoneNumber       string         `json:"phone_number" gorm:"type:varchar(50)"`
	Email             string         `json:"email" gorm:"type:varchar(255)"`
	DateOfBirth       *string        `json:"date_of_birth" gorm:"type:date"`
	Gender            string         `json:"gender" gorm:"type:varchar(50)"`
	StreetAddress     string         `json:"street_address" gorm:"type:text"`
	City              string         `json:"city" gorm:"type:varchar(255)"`
	State             string         `json:"state" gorm:"type:varchar(255)"`
	PostalCode        string         `json:"postal_code" gorm:"type:varchar(50)"`
	Country           string         `json:"country" gorm:"type:varchar(255);default:'India'"`
	Coordinates       string         `json:"coordinates" gorm:"type:text"`
	LandOwnershipType string         `json:"land_ownership_type" gorm:"type:varchar(100)"`
	Status            string         `json:"status" gorm:"type:farmer_status;not null;default:'ACTIVE'"`
	Preferences       entities.JSONB `json:"preferences" gorm:"type:jsonb;default:'{}'"`
	Metadata          entities.JSONB `json:"metadata" gorm:"type:jsonb;default:'{}'"`
}

// TableName returns the table name for the FarmerLegacy model
func (f *FarmerLegacy) TableName() string {
	return "farmers_legacy"
}

// GetTableIdentifier returns the table identifier for ID generation
func (f *FarmerLegacy) GetTableIdentifier() string {
	return "FMRL"
}

// GetTableSize returns the table size for ID generation
func (f *FarmerLegacy) GetTableSize() hash.TableSize {
	return hash.Large
}

// NewFarmerLegacy creates a new legacy farmer model with proper initialization
// DEPRECATED: Use NewFarmer() from farmer_normalized.go instead
func NewFarmerLegacy() *FarmerLegacy {
	baseModel := base.NewBaseModel("FMRL", hash.Large)
	return &FarmerLegacy{
		BaseModel:   *baseModel,
		Preferences: make(entities.JSONB),
		Metadata:    make(entities.JSONB),
	}
}

// Validate validates the legacy farmer model
func (f *FarmerLegacy) Validate() error {
	if f.AAAUserID == "" {
		return common.ErrInvalidInput
	}
	if f.AAAOrgID == "" {
		return common.ErrInvalidInput
	}
	if f.FirstName == "" {
		return common.ErrInvalidInput
	}
	if f.LastName == "" {
		return common.ErrInvalidInput
	}
	return nil
}

// Validate validates the farmer link model
func (fl *FarmerLink) Validate() error {
	if fl.AAAUserID == "" {
		return common.ErrInvalidInput
	}
	if fl.AAAOrgID == "" {
		return common.ErrInvalidInput
	}
	return nil
}
