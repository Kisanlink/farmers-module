package models

import (
	pb "github.com/kisanlink/protobuf/pb-aaa"
)

// FarmerSignupRequest defines the request structure for farmer registration
type FarmerSignupRequest struct {
	UserId           *string `json:"user_id" validate:"omitempty,uuid"`
	UserName         *string `json:"username" validate:"omitempty,min=2,max=100"`
	Email            *string `json:"email" validate:"omitempty,email"`
	CountryCode      string  `json:"country_code" validate:"required,numeric,len=3"`   // Changed to numeric validation
	MobileNumber     uint64  `json:"mobile_number" validate:"required,numeric,len=10"` // Indian mobile numbers
	AadhaarNumber    *string `json:"aadhaar_number" validate:"omitempty,numeric,len=12"`
	KisansathiUserId *string `json:"kisansathi_user_id" validate:"omitempty,uuid"`

	IsSubscribed *bool `json:"is_subscribed" validate:"omitempty"`
}

// Farmer represents a farmer entity in the database
type Farmer struct {
	Base
	UserId           string  `gorm:"type:varchar(36);uniqueIndex" json:"user_id"`
	KisansathiUserId *string `gorm:"type:varchar(36)" json:"kisansathi_user_id,omitempty"`
	IsActive         bool    `gorm:"default:true" json:"is_active"`

	IsSubscribed bool `gorm:"default:false" json:"is_subscribed"`

	// Changed from json:"user,omitempty" to json:"user_details,omitempty"
	UserDetails *pb.User `json:"user_details,omitempty" gorm:"-"`
}
