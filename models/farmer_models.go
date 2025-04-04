package models

import ("time")



// FarmerSignupRequest defines the request structure for farmer registration
type FarmerSignupRequest struct {
	UserID          *string  `json:"user_id" validate:"omitempty,uuid"`
	UserName        *string  `json:"username" validate:"omitempty,min=2,max=100"`
	Email           *string  `json:"email" validate:"omitempty,email"`
	CountryCode     string   `json:"country_code" validate:"required,numeric,len=3"` // Changed to numeric validation
	MobileNumber    uint64   `json:"mobile_number" validate:"required,numeric,len=10"` // Indian mobile numbers
	AadhaarNumber   *string  `json:"aadhaar_number" validate:"omitempty,numeric,len=12"`
	KisansathiUserID *string `json:"kisansathi_user_id" validate:"omitempty,uuid"`
}


// Farmer represents a farmer entity in the database
type Farmer struct {
	Base
	UserID           string  `gorm:"type:varchar(36);uniqueIndex" json:"user_id"`  // Added uniqueIndex
	KisansathiUserID *string `gorm:"type:varchar(36)" json:"kisansathi_user_id,omitempty"`
	IsActive         bool    `gorm:"default:true" json:"is_active"`
}

type FarmerFilter struct {
	UserID           *string
	UserName         *string
	Email           *string
	CountryCode     *string
	MobileNumber    *uint64
	AadhaarNumber   *string
	KisansathiUserID *string
	IsActive        *bool
	CreatedAfter    *time.Time
	CreatedBefore   *time.Time
}