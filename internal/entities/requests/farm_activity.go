package requests

import (
	"time"
)

// CreateActivityRequest represents the request to create a farm activity
type CreateActivityRequest struct {
	BaseRequest
	CropCycleID  string            `json:"crop_cycle_id" validate:"required"`
	ActivityType string            `json:"activity_type" validate:"required"`
	PlannedAt    *time.Time        `json:"planned_at"`
	Metadata     map[string]string `json:"metadata"`
}

// CompleteActivityRequest represents the request to complete a farm activity
type CompleteActivityRequest struct {
	BaseRequest
	ID          string            `json:"id" validate:"required"`
	CompletedAt time.Time         `json:"completed_at" validate:"required"`
	Output      map[string]string `json:"output"`
}

// UpdateActivityRequest represents the request to update a farm activity
type UpdateActivityRequest struct {
	BaseRequest
	ID           string            `json:"id" validate:"required"`
	ActivityType *string           `json:"activity_type,omitempty"`
	PlannedAt    *time.Time        `json:"planned_at,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// ListActivitiesRequest represents the request to list farm activities
type ListActivitiesRequest struct {
	BaseRequest
	CropCycleID  string `json:"crop_cycle_id,omitempty"`
	ActivityType string `json:"activity_type,omitempty"`
	Status       string `json:"status,omitempty"`
	DateFrom     string `json:"date_from,omitempty"`
	DateTo       string `json:"date_to,omitempty"`
	Page         int    `json:"page" validate:"min=1"`
	PageSize     int    `json:"page_size" validate:"min=1,max=100"`
}
