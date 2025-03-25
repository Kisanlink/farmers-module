package models

import (
	"time"
	"github.com/Kisanlink/farmers-module/utils" // Import the utils package
	"gorm.io/gorm"
)

// Base model for common fields
type Base struct {
    Id        string    `gorm:"type:varchar(10);primaryKey"` // Store as string (10 digits)
    CreatedAt time.Time `json:"createdAt" gorm:"column:created_at"`
    UpdatedAt time.Time `json:"updatedAt" gorm:"column:updated_at"`
}

// BeforeCreate hook to generate a 10-digit ID
func (b *Base) BeforeCreate(tx *gorm.DB) (err error) {
    if b.Id == "" { // Check if ID is empty
        b.Id = utils.Generate10DigitID() // Generate a 10-digit ID
    }
    return
}
