package models

// FarmerSignupRequest represents the request body for farmer registration
type FarmerSignupRequest struct {
	Name             string  `json:"name" binding:"required"`
	Email            string  `json:"email" binding:"required,email"`
	CountryCode      string  `json:"country_code" binding:"required"`
	MobileNumber     string  `json:"mobile_number" binding:"required"`
	AadhaarNumber    string  `json:"aadhaar_number" binding:"required"`
	KisansathiUserID *string `json:"kisansathi_user_id"` //union of uuid and nil
}

// Farmer represents a farmer entity in the database
type Farmer struct {
	Base
	UserID           string  `gorm:"type:varchar(36)" json:"user_id"` // Use VARCHAR for custom IDs
	KisansathiUserID *string `gorm:"type:varchar(36)" json:"kisansathi_user_id,omitempty"`
	IsActive         bool    `gorm:"default:true" json:"is_active"`
}
