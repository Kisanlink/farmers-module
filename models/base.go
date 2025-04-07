package models

import (
	"time"
	"github.com/Kisanlink/farmers-module/utils" // Import the utils package
	"gorm.io/gorm"
)

// Base model for common fields
type Base struct {
    id        string    `gorm:"type:varchar(10);primaryKey"` // Store as string (10 digits)
    CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
    UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`
}

// BeforeCreate hook to generate a 10-digit ID
func (b *Base) BeforeCreate(tx *gorm.DB) (err error) {
    if b.id == "" { // Check if ID is empty
        b.id = utils.Generate10DigitID() // Generate a 10-digit ID
    }
    return
}
