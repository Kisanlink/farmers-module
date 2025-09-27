package farmer

import (
	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// FarmerLinkResponse represents the link between a farmer and an FPO (Response DTO)
type FarmerLinkResponse struct {
	ID               string  `json:"id"`
	AAAUserID        string  `json:"aaa_user_id"`
	AAAOrgID         string  `json:"aaa_org_id"`
	KisanSathiUserID *string `json:"kisan_sathi_user_id,omitempty"`
	Status           string  `json:"status"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

// FarmerProfileResponse represents a farmer's profile with linked farms (Response DTO)
type FarmerProfileResponse struct {
	ID                string            `json:"id"`
	AAAUserID         string            `json:"aaa_user_id"`
	AAAOrgID          string            `json:"aaa_org_id"`
	KisanSathiUserID  *string           `json:"kisan_sathi_user_id,omitempty"`
	FirstName         string            `json:"first_name"`
	LastName          string            `json:"last_name"`
	PhoneNumber       string            `json:"phone_number"`
	Email             string            `json:"email"`
	DateOfBirth       *string           `json:"date_of_birth"`
	Gender            string            `json:"gender"`
	StreetAddress     string            `json:"street_address"`
	City              string            `json:"city"`
	State             string            `json:"state"`
	PostalCode        string            `json:"postal_code"`
	Country           string            `json:"country"`
	Coordinates       string            `json:"coordinates"`
	LandOwnershipType string            `json:"land_ownership_type"`
	Status            string            `json:"status"`
	Preferences       map[string]string `json:"preferences"`
	Metadata          map[string]string `json:"metadata"`
	Farms             []FarmResponse    `json:"farms"`
	CreatedAt         string            `json:"created_at"`
	UpdatedAt         string            `json:"updated_at"`
}

// FarmResponse represents a minimal farm reference for the farmer profile (Response DTO)
type FarmResponse struct {
	ID              string            `json:"id"`
	AAAFarmerUserID string            `json:"aaa_farmer_user_id"`
	AAAOrgID        string            `json:"aaa_org_id"`
	AreaHa          float64           `json:"area_ha"`
	Metadata        map[string]string `json:"metadata"`
	CreatedAt       string            `json:"created_at"`
	UpdatedAt       string            `json:"updated_at"`
}

// FarmerResponse represents a farmer's response DTO (no GORM tags)
type FarmerResponse struct {
	ID                string            `json:"id"`
	AAAUserID         string            `json:"aaa_user_id"`
	AAAOrgID          string            `json:"aaa_org_id"`
	KisanSathiUserID  *string           `json:"kisan_sathi_user_id,omitempty"`
	FirstName         string            `json:"first_name"`
	LastName          string            `json:"last_name"`
	PhoneNumber       string            `json:"phone_number"`
	Email             string            `json:"email"`
	DateOfBirth       *string           `json:"date_of_birth"`
	Gender            string            `json:"gender"`
	StreetAddress     string            `json:"street_address"`
	City              string            `json:"city"`
	State             string            `json:"state"`
	PostalCode        string            `json:"postal_code"`
	Country           string            `json:"country"`
	Coordinates       string            `json:"coordinates"`
	LandOwnershipType string            `json:"land_ownership_type"`
	Status            string            `json:"status"`
	Preferences       map[string]string `json:"preferences"`
	Metadata          map[string]string `json:"metadata"`
	CreatedAt         string            `json:"created_at"`
	UpdatedAt         string            `json:"updated_at"`
}

// Database Models (Keep these for backward compatibility and database operations)

// FarmerLink represents the link between a farmer and an FPO (Database Model)
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
	return "farmer_link"
}

// GetTableSize returns the table size for ID generation
func (fl *FarmerLink) GetTableSize() hash.TableSize {
	return hash.Medium
}

// Farmer represents a farmer's database model that embeds base.BaseModel
type Farmer struct {
	base.BaseModel
	AAAUserID         string            `json:"aaa_user_id" gorm:"type:varchar(255);not null;uniqueIndex:idx_farmer_unique"`
	AAAOrgID          string            `json:"aaa_org_id" gorm:"type:varchar(255);not null;uniqueIndex:idx_farmer_unique"`
	KisanSathiUserID  *string           `json:"kisan_sathi_user_id" gorm:"type:varchar(255)"`
	FirstName         string            `json:"first_name" gorm:"type:varchar(255);not null"`
	LastName          string            `json:"last_name" gorm:"type:varchar(255);not null"`
	PhoneNumber       string            `json:"phone_number" gorm:"type:varchar(50)"`
	Email             string            `json:"email" gorm:"type:varchar(255)"`
	DateOfBirth       *string           `json:"date_of_birth" gorm:"type:date"`
	Gender            string            `json:"gender" gorm:"type:varchar(50)"`
	StreetAddress     string            `json:"street_address" gorm:"type:text"`
	City              string            `json:"city" gorm:"type:varchar(255)"`
	State             string            `json:"state" gorm:"type:varchar(255)"`
	PostalCode        string            `json:"postal_code" gorm:"type:varchar(50)"`
	Country           string            `json:"country" gorm:"type:varchar(255);default:'India'"`
	Coordinates       string            `json:"coordinates" gorm:"type:text"`
	LandOwnershipType string            `json:"land_ownership_type" gorm:"type:varchar(100)"`
	Status            string            `json:"status" gorm:"type:farmer_status;not null;default:'ACTIVE'"`
	Preferences       map[string]string `json:"preferences" gorm:"type:jsonb;default:'{}'"`
	Metadata          map[string]string `json:"metadata" gorm:"type:jsonb;default:'{}'"`
}

// TableName returns the table name for the Farmer model
func (f *Farmer) TableName() string {
	return "farmers"
}

// GetTableIdentifier returns the table identifier for ID generation
func (f *Farmer) GetTableIdentifier() string {
	return "farmer"
}

// GetTableSize returns the table size for ID generation
func (f *Farmer) GetTableSize() hash.TableSize {
	return hash.Large
}

// NewFarmer creates a new farmer model with proper initialization
func NewFarmer() *Farmer {
	baseModel := base.NewBaseModel("farmer", hash.Large)
	return &Farmer{
		BaseModel:   *baseModel,
		Preferences: make(map[string]string),
		Metadata:    make(map[string]string),
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

// Conversion Functions

// ToFarmerResponse converts Farmer database model to FarmerResponse DTO
func (f *Farmer) ToFarmerResponse() *FarmerResponse {
	return &FarmerResponse{
		ID:                f.ID,
		AAAUserID:         f.AAAUserID,
		AAAOrgID:          f.AAAOrgID,
		KisanSathiUserID:  f.KisanSathiUserID,
		FirstName:         f.FirstName,
		LastName:          f.LastName,
		PhoneNumber:       f.PhoneNumber,
		Email:             f.Email,
		DateOfBirth:       f.DateOfBirth,
		Gender:            f.Gender,
		StreetAddress:     f.StreetAddress,
		City:              f.City,
		State:             f.State,
		PostalCode:        f.PostalCode,
		Country:           f.Country,
		Coordinates:       f.Coordinates,
		LandOwnershipType: f.LandOwnershipType,
		Status:            f.Status,
		Preferences:       f.Preferences,
		Metadata:          f.Metadata,
		CreatedAt:         f.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:         f.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// ToFarmerLinkResponse converts FarmerLink database model to FarmerLinkResponse DTO
func (fl *FarmerLink) ToFarmerLinkResponse() *FarmerLinkResponse {
	return &FarmerLinkResponse{
		ID:               fl.ID,
		AAAUserID:        fl.AAAUserID,
		AAAOrgID:         fl.AAAOrgID,
		KisanSathiUserID: fl.KisanSathiUserID,
		Status:           fl.Status,
		CreatedAt:        fl.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        fl.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
