package models


// FarmRequest - Request model for creating a farm
type FarmRequest struct {
	KisansathiUserID *string  `json:"kisansathi_user_id,omitempty"` // If Kisansathi is creating the farm
	FarmerID         string   `json:"farmer_id" binding:"required"` // Mandatory: Farmer who owns the farm
	Location         string   `json:"location" binding:"required"`  // GeoJSON or WKT format
	Area             float64  `json:"area" binding:"required"`      // Must be > 0
	Locality         string   `json:"locality" binding:"required"`  // Village or city name
	Verified         bool     `json:"verified"`                     // Whether the farm is verified
	Actions       []string `json:"actions"` // Optional
}

// Farm - Database model for storing farm details
type Farm struct {
	Base
	FarmerID string `json:"farmer_id" gorm:"type:varchar(36);not null"`
	KisansathiID *string `json:"kisansathi_id,omitempty" gorm:"type:uuid"`     // Nullable: If created by Kisansathi
	Verified     bool    `json:"verified"`                                     // Verified by admin
	IsOwner      bool    `json:"is_owner"`                                     // If the farmer is the owner
	Location     string  `json:"location" gorm:"type:geometry(Polygon);not null"` // Stored as spatial data
	Area         float64 `json:"area"`                                         // Farm area in hectares/acres
	Locality     string  `json:"locality"`                                     // Name of village/city
	CurrentCycle string  `json:"current_cycle"`                                // Crop cycle
	OwnerID      string  `json:"owner_id" gorm:"type:uuid;not null"`           // References User.ID (Farmer or Kisansathi
}