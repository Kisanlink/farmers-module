package entities

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// JSONB is a custom type for handling JSONB columns
type JSONB map[string]interface{}

// Value implements the driver.Valuer interface
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal JSONB value")
	}

	result := make(map[string]interface{})
	if err := json.Unmarshal(bytes, &result); err != nil {
		return err
	}

	*j = result
	return nil
}

// FarmerProfile represents a farmer profile in the domain
// DEPRECATED: This entity is deprecated and should not be used for new development.
// Use internal/entities/farmer/farmer_normalized.go Farmer entity instead.
// This entity uses denormalized address fields and is NOT included in active migrations.
type FarmerProfile struct {
	base.BaseModel
	AAAUserID        string   `json:"aaa_user_id"`
	AAAOrgID         string   `json:"aaa_org_id"`
	KisanSathiUserID *string  `json:"kisan_sathi_user_id,omitempty"`
	FirstName        string   `json:"first_name,omitempty"`
	LastName         string   `json:"last_name,omitempty"`
	PhoneNumber      string   `json:"phone_number,omitempty"`
	Email            string   `json:"email,omitempty"`
	DateOfBirth      string   `json:"date_of_birth,omitempty"`
	Gender           string   `json:"gender,omitempty"`
	AddressID        *string  `gorm:"column:address_id" json:"-"`
	Address          *Address `gorm:"foreignKey:AddressID" json:"address,omitempty"`
	Preferences      JSONB    `gorm:"type:jsonb;serializer:json" json:"preferences,omitempty"`
	Metadata         JSONB    `gorm:"type:jsonb;serializer:json" json:"metadata,omitempty"`
	Status           string   `json:"status"`
}

// TableName returns the table name for the FarmerProfile model
func (fp *FarmerProfile) TableName() string {
	return "farmer_profiles"
}

// GetTableIdentifier returns the table identifier for ID generation
func (fp *FarmerProfile) GetTableIdentifier() string {
	return "FMRP"
}

// GetTableSize returns the table size for ID generation
func (fp *FarmerProfile) GetTableSize() hash.TableSize {
	return hash.Medium
}

// Address represents address information in the domain
// DEPRECATED: This Address entity is deprecated and should not be migrated.
// Use internal/entities/farmer/farmer_normalized.go Address entity instead.
// This type exists only for backward compatibility with FarmerProfile entity.
type Address struct {
	base.BaseModel
	StreetAddress string `json:"street_address,omitempty" gorm:"type:text"`
	City          string `json:"city,omitempty" gorm:"type:varchar(255)"`
	State         string `json:"state,omitempty" gorm:"type:varchar(255)"`
	PostalCode    string `json:"postal_code,omitempty" gorm:"type:varchar(50)"`
	Country       string `json:"country,omitempty" gorm:"type:varchar(255);default:'India'"`
	Coordinates   string `json:"coordinates,omitempty" gorm:"type:text"`
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

// FarmerLink represents a farmer link in the domain
type FarmerLink struct {
	base.BaseModel
	AAAUserID        string     `json:"aaa_user_id"`
	AAAOrgID         string     `json:"aaa_org_id"`
	KisanSathiUserID *string    `json:"kisan_sathi_user_id,omitempty"`
	Status           string     `json:"status"`
	LinkedAt         *time.Time `json:"linked_at,omitempty"`
	UnlinkedAt       *time.Time `json:"unlinked_at,omitempty"`
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

// LinkFarmerRequest represents a request to link farmer to FPO
type LinkFarmerRequest struct {
	AAAUserID string `json:"aaa_user_id"`
	AAAOrgID  string `json:"aaa_org_id"`
}

// UnlinkFarmerRequest represents a request to unlink farmer from FPO
type UnlinkFarmerRequest struct {
	AAAUserID string `json:"aaa_user_id"`
	AAAOrgID  string `json:"aaa_org_id"`
}

// FarmerLinkage represents farmer linkage information
type FarmerLinkage struct {
	base.BaseModel
	AAAUserID        string     `json:"aaa_user_id"`
	AAAOrgID         string     `json:"aaa_org_id"`
	KisanSathiUserID *string    `json:"kisan_sathi_user_id,omitempty"`
	Status           string     `json:"status"`
	LinkedAt         *time.Time `json:"linked_at,omitempty"`
	UnlinkedAt       *time.Time `json:"unlinked_at,omitempty"`
}

// TableName returns the table name for the FarmerLinkage model
func (fl *FarmerLinkage) TableName() string {
	return "farmer_linkages"
}

// GetTableIdentifier returns the table identifier for ID generation
func (fl *FarmerLinkage) GetTableIdentifier() string {
	return "FMLG"
}

// GetTableSize returns the table size for ID generation
func (fl *FarmerLinkage) GetTableSize() hash.TableSize {
	return hash.Medium
}
