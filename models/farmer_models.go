package models


// FarmerSignupRequest defines the request structure for farmer registration
type FarmerSignupRequest struct {
	UserID          *string `json:"user_id"`           // Optional
	Name            *string  `json:"name"` // Optional
	Email           *string `json:"email"`             // Optional
	CountryCode     string  `json:"country_code" binding:"required"` // Mandatory
	MobileNumber    string  `json:"mobile_number" binding:"required"` // Mandatory
	AadhaarNumber   *string `json:"aadhaar_number"`    // Optional
	KisansathiUserID *string `json:"kisansathi_user_id"` // Optional
	Roles            []string `json:"roles"` // Optional
	Actions       []string `json:"actions"` // Optional
}

// Farmer represents a farmer entity in the database
type Farmer struct {
	Base
	UserID           string  `gorm:"type:varchar(36)" json:"user_id"` // Use VARCHAR for custom IDs
	KisansathiUserID *string `gorm:"type:varchar(36)" json:"kisansathi_user_id,omitempty"`
	IsActive         bool    `gorm:"default:true" json:"is_active"`
}
