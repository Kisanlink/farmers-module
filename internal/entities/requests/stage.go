package requests

import "github.com/Kisanlink/farmers-module/internal/entities"

// CreateStageRequest represents the request to create a stage
type CreateStageRequest struct {
	BaseRequest
	StageName   string         `json:"stage_name" binding:"required,min=1,max=100" example:"Germination"`
	Description *string        `json:"description,omitempty" example:"Seeds sprouting and initial growth"`
	Properties  entities.JSONB `json:"properties,omitempty" swaggertype:"object" example:"{\"color\":\"green\"}"`
}

// UpdateStageRequest represents the request to update a stage
type UpdateStageRequest struct {
	BaseRequest
	ID          string         `json:"-"`
	StageName   *string        `json:"stage_name,omitempty" binding:"omitempty,min=1,max=100" example:"Germination Updated"`
	Description *string        `json:"description,omitempty" example:"Updated description"`
	Properties  entities.JSONB `json:"properties,omitempty" swaggertype:"object" example:"{\"color\":\"blue\"}"`
	IsActive    *bool          `json:"is_active,omitempty" example:"true"`
}

// GetStageRequest represents the request to get a stage
type GetStageRequest struct {
	BaseRequest
	ID string `json:"-"`
}

// DeleteStageRequest represents the request to delete a stage
type DeleteStageRequest struct {
	BaseRequest
	ID string `json:"-"`
}

// ListStagesRequest represents the request to list stages
type ListStagesRequest struct {
	BaseRequest
	PaginationRequest
	Search   string `json:"search,omitempty" form:"search" example:"germination"`
	IsActive *bool  `json:"is_active,omitempty" form:"is_active" example:"true"`
}

// GetStageLookupRequest represents the request to get stage lookup data
type GetStageLookupRequest struct {
	BaseRequest
}

// AssignStageToCropRequest represents the request to assign a stage to a crop
type AssignStageToCropRequest struct {
	BaseRequest
	CropID       string         `json:"-"`
	StageID      string         `json:"stage_id" binding:"required" example:"STGE00000001"`
	StageOrder   int            `json:"stage_order" binding:"required,min=1" example:"1"`
	DurationDays *int           `json:"duration_days,omitempty" binding:"omitempty,min=1" example:"14"`
	DurationUnit string         `json:"duration_unit,omitempty" binding:"omitempty,oneof=DAYS WEEKS MONTHS" example:"DAYS"`
	Properties   entities.JSONB `json:"properties,omitempty" swaggertype:"object" example:"{\"notes\":\"Critical stage\"}"`
}

// UpdateCropStageRequest represents the request to update a crop stage
type UpdateCropStageRequest struct {
	BaseRequest
	CropID       string         `json:"-"`
	StageID      string         `json:"-"`
	StageOrder   *int           `json:"stage_order,omitempty" binding:"omitempty,min=1" example:"2"`
	DurationDays *int           `json:"duration_days,omitempty" binding:"omitempty,min=1" example:"21"`
	DurationUnit *string        `json:"duration_unit,omitempty" binding:"omitempty,oneof=DAYS WEEKS MONTHS" example:"DAYS"`
	Properties   entities.JSONB `json:"properties,omitempty" swaggertype:"object" example:"{\"notes\":\"Updated notes\"}"`
	IsActive     *bool          `json:"is_active,omitempty" example:"true"`
}

// RemoveStageFromCropRequest represents the request to remove a stage from a crop
type RemoveStageFromCropRequest struct {
	BaseRequest
	CropID  string `json:"-"`
	StageID string `json:"-"`
}

// GetCropStagesRequest represents the request to get crop stages
type GetCropStagesRequest struct {
	BaseRequest
	CropID string `json:"-"`
}

// ReorderCropStagesRequest represents the request to reorder crop stages
type ReorderCropStagesRequest struct {
	BaseRequest
	CropID      string         `json:"-"`
	StageOrders map[string]int `json:"stage_orders" binding:"required" example:"{\"STGE00000001\":1,\"STGE00000002\":2}"` // map[stage_id]order
}
