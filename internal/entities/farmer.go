package entities

import (
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// FarmerProfile represents a farmer profile in the domain
type FarmerProfile struct {
	base.BaseModel
	AAAUserID        string            `json:"aaa_user_id"`
	AAAOrgID         string            `json:"aaa_org_id"`
	KisanSathiUserID *string           `json:"kisan_sathi_user_id,omitempty"`
	FirstName        string            `json:"first_name,omitempty"`
	LastName         string            `json:"last_name,omitempty"`
	PhoneNumber      string            `json:"phone_number,omitempty"`
	Email            string            `json:"email,omitempty"`
	DateOfBirth      string            `json:"date_of_birth,omitempty"`
	Gender           string            `json:"gender,omitempty"`
	Address          Address           `json:"address,omitempty"`
	Preferences      map[string]string `json:"preferences,omitempty"`
	Metadata         map[string]string `json:"metadata,omitempty"`
	Status           string            `json:"status"`
}

// TableName returns the table name for the FarmerProfile model
func (fp *FarmerProfile) TableName() string {
	return "farmer_profiles"
}

// GetTableIdentifier returns the table identifier for ID generation
func (fp *FarmerProfile) GetTableIdentifier() string {
	return "farmer_profile"
}

// GetTableSize returns the table size for ID generation
func (fp *FarmerProfile) GetTableSize() hash.TableSize {
	return hash.Medium
}

// Address represents address information in the domain
type Address struct {
	StreetAddress string `json:"street_address,omitempty"`
	City          string `json:"city,omitempty"`
	State         string `json:"state,omitempty"`
	PostalCode    string `json:"postal_code,omitempty"`
	Country       string `json:"country,omitempty"`
	Coordinates   string `json:"coordinates,omitempty"`
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
	return "farmer_link"
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
	return "farmer_linkage"
}

// GetTableSize returns the table size for ID generation
func (fl *FarmerLinkage) GetTableSize() hash.TableSize {
	return hash.Medium
}
