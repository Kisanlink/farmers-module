package farmer

import (
	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// Address represents an address entity (normalized, reusable)
// This allows multiple entities (farmers, farms, etc.) to reference the same address
type Address struct {
	base.BaseModel
	StreetAddress string `json:"street_address" gorm:"type:text"`
	City          string `json:"city" gorm:"type:varchar(255)"`
	State         string `json:"state" gorm:"type:varchar(255)"`
	PostalCode    string `json:"postal_code" gorm:"type:varchar(50)"`
	Country       string `json:"country" gorm:"type:varchar(255);default:'India'"`
	Coordinates   string `json:"coordinates" gorm:"type:geometry(Point,4326)"` // PostGIS for spatial queries
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

// FarmerNormalized represents a farmer's database model with normalized address relationship
// This is the RECOMMENDED entity to use for all farmer operations
type FarmerNormalized struct {
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
	Status            string `json:"status" gorm:"type:varchar(50);not null;default:'ACTIVE'"`

	// Flexible Data (JSONB for extensibility)
	Preferences map[string]string `json:"preferences" gorm:"type:jsonb;default:'{}'"`
	Metadata    map[string]string `json:"metadata" gorm:"type:jsonb;default:'{}'"`
}

// TableName returns the table name for the FarmerNormalized model
func (f *FarmerNormalized) TableName() string {
	return "farmers"
}

// GetTableIdentifier returns the table identifier for ID generation
func (f *FarmerNormalized) GetTableIdentifier() string {
	return "FMRR"
}

// GetTableSize returns the table size for ID generation
func (f *FarmerNormalized) GetTableSize() hash.TableSize {
	return hash.Large
}

// NewFarmerNormalized creates a new normalized farmer model with proper initialization
func NewFarmerNormalized() *FarmerNormalized {
	baseModel := base.NewBaseModel("FMRR", hash.Large)
	return &FarmerNormalized{
		BaseModel:   *baseModel,
		Status:      "ACTIVE",
		Preferences: make(map[string]string),
		Metadata:    make(map[string]string),
	}
}

// Validate validates the normalized farmer model
func (f *FarmerNormalized) Validate() error {
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
func (f *FarmerNormalized) SetAddress(address *Address) {
	if address != nil {
		f.Address = address
		addressID := address.GetID()
		f.AddressID = &addressID
	}
}

// ClearAddress removes the address relationship
func (f *FarmerNormalized) ClearAddress() {
	f.Address = nil
	f.AddressID = nil
}
