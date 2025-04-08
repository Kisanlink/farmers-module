package models

import (
	"time"

	"github.com/Kisanlink/farmers-module/utils"
	"gorm.io/gorm"
)

// Base model for common fields
type Base struct {
	ID        string    `gorm:"type:varchar(10);primaryKey"`      // Store as string (10 digits)
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"` // Automatically set during creation
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"` // Automatically set during updates
}

// BeforeCreate hook to generate a 10-digit ID
func (b *Base) BeforeCreate(tx *gorm.DB) (err error) {
	if b.ID == "" { // Check if ID is empty
		b.ID = utils.Generate10DigitID() // Generate a 10-digit ID
	}
	return
}
