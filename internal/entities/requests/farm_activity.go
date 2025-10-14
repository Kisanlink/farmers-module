package requests

import (
	"time"
)

// CreateActivityRequest represents the request to create a farm activity
type CreateActivityRequest struct {
	BaseRequest
	CropCycleID  string                 `json:"crop_cycle_id" validate:"required" example:"cycle_123e4567-e89b-12d3-a456-426614174000"`
	CropStageID  *string                `json:"crop_stage_id,omitempty" example:"CSTG_GERMINATION"`
	ActivityType string                 `json:"activity_type" validate:"required" example:"SOWING"`
	PlannedAt    *time.Time             `json:"planned_at" example:"2024-11-10T09:00:00Z"`
	Metadata     map[string]interface{} `json:"metadata" example:"seed_type:HD2967,seed_rate:100kg_per_acre"`
}

// CompleteActivityRequest represents the request to complete a farm activity
type CompleteActivityRequest struct {
	BaseRequest
	ID          string                 `json:"id" validate:"required" example:"activity_123e4567-e89b-12d3-a456-426614174000"`
	CompletedAt time.Time              `json:"completed_at" validate:"required" example:"2024-11-11T16:30:00Z"`
	Output      map[string]interface{} `json:"output" example:"area_covered:2.5ha,workers:4,notes:completed_successfully"`
}

// UpdateActivityRequest represents the request to update a farm activity
type UpdateActivityRequest struct {
	BaseRequest
	ID           string                 `json:"id" validate:"required" example:"activity_123e4567-e89b-12d3-a456-426614174000"`
	CropStageID  *string                `json:"crop_stage_id,omitempty" example:"CSTG_FLOWERING"`
	ActivityType *string                `json:"activity_type,omitempty" example:"IRRIGATION"`
	PlannedAt    *time.Time             `json:"planned_at,omitempty" example:"2024-12-01T08:00:00Z"`
	Metadata     map[string]interface{} `json:"metadata,omitempty" example:"water_source:borewell,duration:2hours"`
}

// ListActivitiesRequest represents the request to list farm activities
type ListActivitiesRequest struct {
	BaseRequest
	CropCycleID  string `json:"crop_cycle_id,omitempty" example:"cycle_123e4567-e89b-12d3-a456-426614174000"`
	CropStageID  string `json:"crop_stage_id,omitempty" example:"CSTG_GERMINATION"`
	ActivityType string `json:"activity_type,omitempty" example:"SOWING"`
	Status       string `json:"status,omitempty" example:"COMPLETED"`
	DateFrom     string `json:"date_from,omitempty" example:"2024-11-01"`
	DateTo       string `json:"date_to,omitempty" example:"2024-12-31"`
	Page         int    `json:"page" validate:"min=1" example:"1"`
	PageSize     int    `json:"page_size" validate:"min=1,max=100" example:"20"`
}
