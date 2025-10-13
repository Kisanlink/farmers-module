package responses

import "time"

// CropData represents crop data in responses
type CropData struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	ScientificName *string                `json:"scientific_name,omitempty"`
	Category       string                 `json:"category"`
	DurationDays   *int                   `json:"duration_days,omitempty"`
	Seasons        []string               `json:"seasons"`
	Unit           string                 `json:"unit"`
	Properties     map[string]interface{} `json:"properties,omitempty"`
	IsActive       bool                   `json:"is_active"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	VarietyCount   int                    `json:"variety_count,omitempty"` // Count of active varieties
}

// YieldByAgeData represents yield information for a specific tree age range in responses
type YieldByAgeData struct {
	AgeFrom      int     `json:"age_from"`
	AgeTo        int     `json:"age_to"`
	YieldPerTree float64 `json:"yield_per_tree"`
}

// CropVarietyData represents crop variety data in responses
type CropVarietyData struct {
	ID           string                 `json:"id"`
	CropID       string                 `json:"crop_id"`
	CropName     string                 `json:"crop_name,omitempty"`
	Name         string                 `json:"name"`
	Description  *string                `json:"description,omitempty"`
	DurationDays *int                   `json:"duration_days,omitempty"`
	YieldPerAcre *float64               `json:"yield_per_acre,omitempty"`
	YieldPerTree *float64               `json:"yield_per_tree,omitempty"`
	YieldByAge   []YieldByAgeData       `json:"yield_by_age,omitempty"`
	Properties   map[string]interface{} `json:"properties,omitempty"`
	IsActive     bool                   `json:"is_active"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// CropResponse represents a single crop response
type CropResponse struct {
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
	RequestID string    `json:"request_id"`
	Data      *CropData `json:"data,omitempty"`
}

// CropListResponse represents a list of crops response
type CropListResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	RequestID string      `json:"request_id"`
	Data      []*CropData `json:"data"`
	Page      int         `json:"page"`
	PageSize  int         `json:"page_size"`
	Total     int         `json:"total"`
}

// CropVarietyResponse represents a single crop variety response
type CropVarietyResponse struct {
	Success   bool             `json:"success"`
	Message   string           `json:"message"`
	RequestID string           `json:"request_id"`
	Data      *CropVarietyData `json:"data,omitempty"`
}

// CropVarietyListResponse represents a list of crop varieties response
type CropVarietyListResponse struct {
	Success   bool               `json:"success"`
	Message   string             `json:"message"`
	RequestID string             `json:"request_id"`
	Data      []*CropVarietyData `json:"data"`
	Page      int                `json:"page"`
	PageSize  int                `json:"page_size"`
	Total     int                `json:"total"`
}

// CropLookupData represents simplified crop data for dropdowns
type CropLookupData struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Category string   `json:"category"`
	Seasons  []string `json:"seasons"`
	Unit     string   `json:"unit"`
}

// CropVarietyLookupData represents simplified variety data for dropdowns
type CropVarietyLookupData struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	DurationDays *int   `json:"duration_days,omitempty"`
}

// CropLookupResponse represents crop lookup data response
type CropLookupResponse struct {
	Success   bool              `json:"success"`
	Message   string            `json:"message"`
	RequestID string            `json:"request_id"`
	Data      []*CropLookupData `json:"data"`
}

// CropVarietyLookupResponse represents variety lookup data response
type CropVarietyLookupResponse struct {
	Success   bool                     `json:"success"`
	Message   string                   `json:"message"`
	RequestID string                   `json:"request_id"`
	Data      []*CropVarietyLookupData `json:"data"`
}

// CropCategoriesResponse represents available crop categories
type CropCategoriesResponse struct {
	Success   bool     `json:"success"`
	Message   string   `json:"message"`
	RequestID string   `json:"request_id"`
	Data      []string `json:"data"`
}

// CropSeasonsResponse represents available crop seasons
type CropSeasonsResponse struct {
	Success   bool     `json:"success"`
	Message   string   `json:"message"`
	RequestID string   `json:"request_id"`
	Data      []string `json:"data"`
}

// Enhanced crop cycle data including crop master data
type EnhancedCropCycleData struct {
	ID        string                 `json:"id"`
	FarmID    string                 `json:"farm_id"`
	FarmerID  string                 `json:"farmer_id"`
	Season    string                 `json:"season"`
	Status    string                 `json:"status"`
	StartDate *time.Time             `json:"start_date,omitempty"`
	EndDate   *time.Time             `json:"end_date,omitempty"`
	CropID    string                 `json:"crop_id"`
	VarietyID *string                `json:"variety_id,omitempty"`
	Crop      *CropLookupData        `json:"crop,omitempty"`
	Variety   *CropVarietyLookupData `json:"variety,omitempty"`
	Outcome   map[string]interface{} `json:"outcome,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// Swagger response types for documentation

// SwaggerCropResponse represents a crop response for Swagger
type SwaggerCropResponse struct {
	Success   bool      `json:"success" example:"true"`
	Message   string    `json:"message" example:"Crop created successfully"`
	RequestID string    `json:"request_id" example:"req_123456789"`
	Data      *CropData `json:"data"`
}

// SwaggerCropListResponse represents a crop list response for Swagger
type SwaggerCropListResponse struct {
	Success   bool        `json:"success" example:"true"`
	Message   string      `json:"message" example:"Crops retrieved successfully"`
	RequestID string      `json:"request_id" example:"req_123456789"`
	Data      []*CropData `json:"data"`
	Page      int         `json:"page" example:"1"`
	PageSize  int         `json:"page_size" example:"20"`
	Total     int         `json:"total" example:"50"`
}

// SwaggerCropVarietyResponse represents a crop variety response for Swagger
type SwaggerCropVarietyResponse struct {
	Success   bool             `json:"success" example:"true"`
	Message   string           `json:"message" example:"Crop variety created successfully"`
	RequestID string           `json:"request_id" example:"req_123456789"`
	Data      *CropVarietyData `json:"data"`
}

// SwaggerCropLookupResponse represents a crop lookup response for Swagger
type SwaggerCropLookupResponse struct {
	Success   bool              `json:"success" example:"true"`
	Message   string            `json:"message" example:"Crop lookup data retrieved successfully"`
	RequestID string            `json:"request_id" example:"req_123456789"`
	Data      []*CropLookupData `json:"data"`
}
