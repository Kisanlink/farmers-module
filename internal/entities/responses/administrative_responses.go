package responses

import "time"

// SeedRolesAndPermissionsResponse represents the response from seeding roles and permissions
type SeedRolesAndPermissionsResponse struct {
	BaseResponse
	Success   bool          `json:"success"`
	Message   string        `json:"message"`
	Error     string        `json:"error,omitempty"`
	Duration  time.Duration `json:"duration"`
	Timestamp time.Time     `json:"timestamp"`
}

// HealthCheckResponse represents the response from a health check
type HealthCheckResponse struct {
	BaseResponse
	Status     string                     `json:"status"`
	Message    string                     `json:"message,omitempty"`
	Components map[string]ComponentHealth `json:"components"`
	Duration   time.Duration              `json:"duration"`
	Timestamp  time.Time                  `json:"timestamp"`
}

// ComponentHealth represents the health status of a system component
type ComponentHealth struct {
	Name      string                 `json:"name"`
	Status    string                 `json:"status"`
	Message   string                 `json:"message,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}
