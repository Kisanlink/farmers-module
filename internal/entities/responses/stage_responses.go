package responses

import (
	"time"

	"github.com/Kisanlink/farmers-module/internal/entities"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// StageData represents stage data in responses
type StageData struct {
	ID          string         `json:"id" example:"STGE00000001"`
	StageName   string         `json:"stage_name" example:"Germination"`
	Description *string        `json:"description,omitempty" example:"Seeds sprouting and initial growth"`
	Properties  entities.JSONB `json:"properties,omitempty" swaggertype:"object"`
	IsActive    bool           `json:"is_active" example:"true"`
	CreatedAt   time.Time      `json:"created_at" example:"2025-01-15T10:30:00Z"`
	UpdatedAt   time.Time      `json:"updated_at" example:"2025-01-15T10:30:00Z"`
}

// StageResponse represents a single stage response
type StageResponse struct {
	*base.BaseResponse `json:",inline"`
	Data               *StageData `json:"data,omitempty"`
}

// StageListResponse represents a list of stages response
type StageListResponse struct {
	*base.BaseResponse `json:",inline"`
	Data               []*StageData `json:"data"`
	Page               int          `json:"page" example:"1"`
	PageSize           int          `json:"page_size" example:"20"`
	Total              int          `json:"total" example:"50"`
}

// CropStageData represents crop stage data in responses
type CropStageData struct {
	ID           string         `json:"id" example:"CSTG00000001"`
	CropID       string         `json:"crop_id" example:"CROP00000001"`
	StageID      string         `json:"stage_id" example:"STGE00000001"`
	StageName    string         `json:"stage_name" example:"Germination"`
	Description  *string        `json:"description,omitempty" example:"Seeds sprouting and initial growth"`
	StageOrder   int            `json:"stage_order" example:"1"`
	DurationDays *int           `json:"duration_days,omitempty" example:"14"`
	DurationUnit string         `json:"duration_unit" example:"DAYS"`
	Properties   entities.JSONB `json:"properties,omitempty" swaggertype:"object"`
	IsActive     bool           `json:"is_active" example:"true"`
	CreatedAt    time.Time      `json:"created_at" example:"2025-01-15T10:30:00Z"`
	UpdatedAt    time.Time      `json:"updated_at" example:"2025-01-15T10:30:00Z"`
}

// CropStageResponse represents a single crop stage response
type CropStageResponse struct {
	*base.BaseResponse `json:",inline"`
	Data               *CropStageData `json:"data,omitempty"`
}

// CropStagesResponse represents a list of crop stages response
type CropStagesResponse struct {
	*base.BaseResponse `json:",inline"`
	Data               []*CropStageData `json:"data"`
}

// StageLookupData represents simplified stage data for lookups
type StageLookupData struct {
	ID          string  `json:"id" example:"STGE00000001"`
	StageName   string  `json:"stage_name" example:"Germination"`
	Description *string `json:"description,omitempty" example:"Seeds sprouting and initial growth"`
}

// StageLookupResponse represents stage lookup response
type StageLookupResponse struct {
	*base.BaseResponse `json:",inline"`
	Data               []*StageLookupData `json:"data"`
}
