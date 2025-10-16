package responses

import "time"

// SeedLookupsResponse represents a response for seeding lookup data
type SeedLookupsResponse struct {
	Success   bool                   `json:"success" example:"true"`
	Message   string                 `json:"message" example:"Successfully seeded lookup data"`
	Error     string                 `json:"error,omitempty"`
	Duration  time.Duration          `json:"duration" example:"500ms"`
	Timestamp time.Time              `json:"timestamp" example:"2024-01-15T10:30:00Z"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// SwaggerSeedLookupsResponse represents a seed lookups response for Swagger
type SwaggerSeedLookupsResponse struct {
	Success   bool                   `json:"success" example:"true"`
	Message   string                 `json:"message" example:"Successfully seeded lookup data: 6 soil types, 6 irrigation sources"`
	Error     string                 `json:"error,omitempty"`
	Duration  string                 `json:"duration" example:"500ms"`
	Timestamp string                 `json:"timestamp" example:"2024-01-15T10:30:00Z"`
	Details   map[string]interface{} `json:"details,omitempty"`
}
