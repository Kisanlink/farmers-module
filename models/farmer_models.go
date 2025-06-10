package models

import (
	"fmt"

	"github.com/Kisanlink/farmers-module/entities"
	pb "github.com/kisanlink/protobuf/pb-aaa"
	"gorm.io/gorm"
)

// FarmerSignupRequest defines the request structure for farmer registration
type FarmerSignupRequest struct {
	UserId             *string `json:"user_id" validate:"omitempty,uuid"`
	UserName           *string `json:"username" validate:"omitempty,min=2,max=100"`
	Email              *string `json:"email" validate:"omitempty,email"`
	CountryCode        string  `json:"country_code" validate:"required,numeric,len=3"`
	MobileNumberString string  `json:"mobile_number" validate:"required,numeric,len=10"`
	MobileNumber       uint64  `json:"-"`
	AadhaarNumber      *string `json:"aadhaar_number" validate:"omitempty,numeric,len=12"`
	KisansathiUserId   *string `json:"kisansathi_user_id" validate:"omitempty,uuid"`
	Type               string  `json:"type" validate:"omitempty"`
}

// Farmer represents a farmer entity in the database
type Farmer struct {
	Base
	UserId           string              `gorm:"type:varchar(36);uniqueIndex" json:"user_id"`
	KisansathiUserId *string             `gorm:"type:varchar(36)" json:"kisansathi_user_id,omitempty"`
	IsActive         bool                `gorm:"default:true" json:"is_active"`
	UserDetails      *pb.User            `json:"user_details,omitempty" gorm:"-"`
	IsSubscribed     bool                `gorm:"default:false" json:"is_subscribed"`
	Type             entities.FarmerType `gorm:"type:varchar(10);not null;default:'OTHER'" json:"type"`
}

func (f *Farmer) BeforeCreate(tx *gorm.DB) (err error) {
	// First, generate/validate Base.Id
	if err = f.Base.BeforeCreate(tx); err != nil {
		return err
	}
	// Then validate Farmer-specific fields (type, etc.)
	if !entities.FARMER_TYPES.IsValid(string(f.Type)) {
		return fmt.Errorf(
			"invalid farmer type: %s. Valid values are: %v",
			f.Type,
			entities.FARMER_TYPES.StringValues(),
		)
	}
	return nil
}
