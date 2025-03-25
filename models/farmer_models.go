package models



// FarmerSignupRequest defines the request structure for farmer registration
type FarmerSignupRequest struct {
	UserID          *string  `json:"user_id" validate:"omitempty,uuid"`
	Name            *string  `json:"name" validate:"omitempty,min=2,max=100"`
	Email           *string  `json:"email" validate:"omitempty,email"`
	CountryCode     string   `json:"country_code" validate:"required,numeric,len=3"` // Changed to numeric validation
	MobileNumber    string   `json:"mobile_number" validate:"required,numeric,len=10"` // Indian mobile numbers
	AadhaarNumber   *string  `json:"aadhaar_number" validate:"omitempty,numeric,len=12"`
	KisansathiUserID *string `json:"kisansathi_user_id" validate:"omitempty,uuid"`
	Roles           *[]string `json:"roles" validate:"omitempty"`
	Actions         *[]string `json:"actions" validate:"omitempty"`
}

// Farmer represents a farmer entity in the database
type Farmer struct {
	Base
	UserID           string  `gorm:"type:varchar(36)" json:"user_id"`
	KisansathiUserID *string `gorm:"type:varchar(36)" json:"kisansathi_user_id,omitempty"`
	IsActive         bool    `gorm:"default:true" json:"is_active"`
}