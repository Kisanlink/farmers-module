package farmer

import (
	"time"

	"github.com/Kisanlink/farmers-module/internal/entities"
	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// Address represents an address entity (normalized, reusable)
// This allows multiple entities (farmers, farms, etc.) to reference the same address
// Note: Coordinates stored as text (lat,lng). PostGIS geometry only used in Farm entity.
type Address struct {
	base.BaseModel
	StreetAddress string `json:"street_address" gorm:"type:text"`
	City          string `json:"city" gorm:"type:varchar(255)"`
	State         string `json:"state" gorm:"type:varchar(255)"`
	PostalCode    string `json:"postal_code" gorm:"type:varchar(50)"`
	Country       string `json:"country" gorm:"type:varchar(255);default:'India'"`
	Coordinates   string `json:"coordinates" gorm:"type:text"` // Stored as "lat,lng" text
}

// TableName returns the table name for the Address model
func (a *Address) TableName() string {
	return "addresses"
}

// GetTableIdentifier returns the table identifier for ID generation
func (a *Address) GetTableIdentifier() string {
	return "ADDR"
}

// GetTableSize returns the table size for ID generation
func (a *Address) GetTableSize() hash.TableSize {
	return hash.Medium
}

// NewAddress creates a new address with proper initialization
func NewAddress() *Address {
	baseModel := base.NewBaseModel("ADDR", hash.Medium)
	return &Address{
		BaseModel: *baseModel,
		Country:   "India", // Default country
	}
}

// Farmer represents a farmer's database model with normalized address relationship
// This is the recommended entity to use for all farmer operations
type Farmer struct {
	base.BaseModel

	// AAA Integration (External System IDs)
	AAAUserID        string  `json:"aaa_user_id" gorm:"type:varchar(255);not null;uniqueIndex:idx_farmer_unique"`
	AAAOrgID         string  `json:"aaa_org_id" gorm:"type:varchar(255);not null;uniqueIndex:idx_farmer_unique"`
	KisanSathiUserID *string `json:"kisan_sathi_user_id" gorm:"type:varchar(255)"`

	// Personal Information
	FirstName   string  `json:"first_name" gorm:"type:varchar(255);not null"`
	LastName    string  `json:"last_name" gorm:"type:varchar(255);not null"`
	PhoneNumber string  `json:"phone_number" gorm:"type:varchar(50)"`
	Email       string  `json:"email" gorm:"type:varchar(255)"`
	DateOfBirth *string `json:"date_of_birth" gorm:"type:date"`
	Gender      string  `json:"gender" gorm:"type:varchar(50)"`

	// Address (Normalized via Foreign Key)
	AddressID *string  `json:"address_id" gorm:"type:varchar(255)"`
	Address   *Address `json:"address,omitempty" gorm:"foreignKey:AddressID;constraint:OnDelete:SET NULL"`

	// Additional Fields
	LandOwnershipType string `json:"land_ownership_type" gorm:"type:varchar(100)"`
	SocialCategory    string `json:"social_category" gorm:"type:varchar(50)"` // e.g., General, OBC, SC, ST
	AreaType          string `json:"area_type" gorm:"type:varchar(50)"`       // e.g., Rural, Urban, Semi-Urban
	Status            string `json:"status" gorm:"type:farmer_status;not null;default:'ACTIVE'"`

	// Relationships - FPO Linkage (optional, preloaded when needed)
	FPOLinkages []*FarmerLink `json:"fpo_linkages,omitempty" gorm:"foreignKey:AAAUserID;references:AAAUserID"`
	Farms       []FarmRef     `json:"farms,omitempty" gorm:"foreignKey:FarmerID;references:ID"`

	// Flexible Data (JSONB for extensibility)
	Preferences entities.JSONB `json:"preferences" gorm:"type:jsonb;default:'{}'::jsonb;serializer:json"`
	Metadata    entities.JSONB `json:"metadata" gorm:"type:jsonb;default:'{}'::jsonb;serializer:json"`
}

// FarmRef represents a minimal farm reference (to avoid circular imports with farm package)
type FarmRef struct {
	ID        string    `json:"id" gorm:"column:id;primaryKey"`
	FarmerID  string    `json:"farmer_id" gorm:"column:farmer_id"`
	Name      *string   `json:"name,omitempty" gorm:"column:name"`
	AreaHa    float64   `json:"area_ha" gorm:"column:area_ha_computed"`
	CreatedAt time.Time `json:"created_at,omitempty" gorm:"column:created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty" gorm:"column:updated_at"`
}

// TableName for FarmRef points to the farms table
func (FarmRef) TableName() string {
	return "farms"
}

// TableName returns the table name for the Farmer model
func (f *Farmer) TableName() string {
	return "farmers"
}

// GetTableIdentifier returns the table identifier for ID generation
func (f *Farmer) GetTableIdentifier() string {
	return "FMRR"
}

// GetTableSize returns the table size for ID generation
func (f *Farmer) GetTableSize() hash.TableSize {
	return hash.Large
}

// NewFarmer creates a new farmer model with proper initialization
func NewFarmer() *Farmer {
	baseModel := base.NewBaseModel("FMRR", hash.Large)
	return &Farmer{
		BaseModel:   *baseModel,
		Status:      "ACTIVE",
		Preferences: make(entities.JSONB),
		Metadata:    make(entities.JSONB),
	}
}

// Validate validates the farmer model
func (f *Farmer) Validate() error {
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

// SetAddress sets the address for the farmer and establishes the foreign key relationship
func (f *Farmer) SetAddress(address *Address) {
	if address != nil {
		f.Address = address
		addressID := address.GetID()
		f.AddressID = &addressID
	}
}

// ClearAddress removes the address relationship
func (f *Farmer) ClearAddress() {
	f.Address = nil
	f.AddressID = nil
}
