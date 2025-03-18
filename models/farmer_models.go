package models
import (
	"github.com/google/uuid"
)
// FarmerSignupRequest represents the request body for farmer registration
type FarmerSignupRequest struct {
	Name          string `json:"name" binding:"required"`
	Email         string `json:"email" binding:"required,email"`
	CountryCode   string `json:"country_code" binding:"required"`
	MobileNumber  int    `json:"mobile_number" binding:"required"`
	AadhaarNumber int64  `json:"aadhaar_number" binding:"required"`
	KisansathiUserID uuid.UUID `json:"kisansathi_user_id"` //union of uuid and nil
}
// Farmer represents a farmer entity in the database
type Farmer struct {
	ID              int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID          uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	KisanSathiUserID uuid.UUID `gorm:"type:uuid;not null" json:"kisansathi_user_id"`
	IsActive        bool      `gorm:"default:true" json:"is_active"`
}
