package entities

import "time"

// CropCycle represents a crop cycle in the domain
type CropCycle struct {
	ID               string            `json:"id"`
	FarmID           string            `json:"farm_id"`
	CropType         string            `json:"crop_type"`
	Season           string            `json:"season"`
	Status           string            `json:"status"`
	PlannedStartDate time.Time         `json:"planned_start_date"`
	ActualStartDate  *time.Time        `json:"actual_start_date,omitempty"`
	PlannedEndDate   time.Time         `json:"planned_end_date"`
	ActualEndDate    *time.Time        `json:"actual_end_date,omitempty"`
	Metadata         map[string]string `json:"metadata,omitempty"`
	CreatedBy        string            `json:"created_by,omitempty"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

// StartCycleRequest represents a request to start a crop cycle
type StartCycleRequest struct {
	FarmID           string            `json:"farm_id"`
	CropType         string            `json:"crop_type"`
	Season           string            `json:"season"`
	PlannedStartDate time.Time         `json:"planned_start_date"`
	PlannedEndDate   time.Time         `json:"planned_end_date"`
	Metadata         map[string]string `json:"metadata,omitempty"`
}

// UpdateCycleRequest represents a request to update a crop cycle
type UpdateCycleRequest struct {
	ID               string            `json:"id"`
	CropType         *string           `json:"crop_type,omitempty"`
	Season           *string           `json:"season,omitempty"`
	PlannedStartDate *time.Time        `json:"planned_start_date,omitempty"`
	PlannedEndDate   *time.Time        `json:"planned_end_date,omitempty"`
	Metadata         map[string]string `json:"metadata,omitempty"`
}

// EndCycleRequest represents a request to end a crop cycle
type EndCycleRequest struct {
	ID            string    `json:"id"`
	ActualEndDate time.Time `json:"actual_end_date"`
}

// ListCyclesRequest represents a request to list crop cycles
type ListCyclesRequest struct {
	FarmID   string `json:"farm_id,omitempty"`
	CropType string `json:"crop_type,omitempty"`
	Season   string `json:"season,omitempty"`
	Status   string `json:"status,omitempty"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}
