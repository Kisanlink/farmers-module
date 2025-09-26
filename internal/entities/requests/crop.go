package requests

import (
	"github.com/Kisanlink/farmers-module/internal/entities/crop"
)

// CreateCropRequest represents a request to create a new crop
type CreateCropRequest struct {
	BaseRequest
	Name             string                 `json:"name" validate:"required,min=2,max=100"`
	Category         crop.CropCategory      `json:"category" validate:"required,oneof=CEREAL LEGUME VEGETABLE OIL_SEEDS FRUIT SPICE"`
	CropDurationDays *int                   `json:"crop_duration_days" validate:"omitempty,min=1"`
	TypicalUnits     []crop.CropUnit        `json:"typical_units" validate:"omitempty,dive,oneof=KG QUINTAL TONNES PIECES"`
	Seasons          []crop.CropSeason      `json:"seasons" validate:"omitempty,dive,oneof=KHARIF RABI SUMMER PERENNIAL"`
	ImageURL         *string                `json:"image_url" validate:"omitempty,url"`
	DocumentID       *string                `json:"document_id" validate:"omitempty,min=1,max=255"`
	Metadata         map[string]interface{} `json:"metadata" validate:"omitempty"`
}

// UpdateCropRequest represents a request to update an existing crop
type UpdateCropRequest struct {
	BaseRequest
	CropID           string                 `json:"crop_id" validate:"required"`
	Name             *string                `json:"name" validate:"omitempty,min=2,max=100"`
	Category         *crop.CropCategory     `json:"category" validate:"omitempty,oneof=CEREAL LEGUME VEGETABLE OIL_SEEDS FRUIT SPICE"`
	CropDurationDays *int                   `json:"crop_duration_days" validate:"omitempty,min=1"`
	TypicalUnits     []crop.CropUnit        `json:"typical_units" validate:"omitempty,dive,oneof=KG QUINTAL TONNES PIECES"`
	Seasons          []crop.CropSeason      `json:"seasons" validate:"omitempty,dive,oneof=KHARIF RABI SUMMER PERENNIAL"`
	ImageURL         *string                `json:"image_url" validate:"omitempty,url"`
	DocumentID       *string                `json:"document_id" validate:"omitempty,min=1,max=255"`
	Metadata         map[string]interface{} `json:"metadata" validate:"omitempty"`
}

// ListCropsRequest represents a request to list crops
type ListCropsRequest struct {
	BaseRequest
	Category *crop.CropCategory `json:"category" validate:"omitempty,oneof=CEREAL LEGUME VEGETABLE OIL_SEEDS FRUIT SPICE"`
	Season   *crop.CropSeason   `json:"season" validate:"omitempty,oneof=KHARIF RABI SUMMER PERENNIAL"`
	Search   *string            `json:"search" validate:"omitempty,min=1,max=100"`
	Limit    *int               `json:"limit" validate:"omitempty,min=1,max=100"`
	Offset   *int               `json:"offset" validate:"omitempty,min=0"`
}

// GetCropRequest represents a request to get a specific crop
type GetCropRequest struct {
	BaseRequest
	CropID string `json:"crop_id" validate:"required"`
}

// DeleteCropRequest represents a request to delete a crop
type DeleteCropRequest struct {
	BaseRequest
	CropID string `json:"crop_id" validate:"required"`
}

// CreateVarietyRequest represents a request to create a new crop variety
type CreateVarietyRequest struct {
	BaseRequest
	CropID          string                 `json:"crop_id" validate:"required"`
	VarietyName     string                 `json:"variety_name" validate:"required,min=2,max=100"`
	DurationDays    *int                   `json:"duration_days" validate:"omitempty,min=1"`
	Characteristics *string                `json:"characteristics" validate:"omitempty,max=1000"`
	Metadata        map[string]interface{} `json:"metadata" validate:"omitempty"`
}

// UpdateVarietyRequest represents a request to update an existing crop variety
type UpdateVarietyRequest struct {
	BaseRequest
	VarietyID       string                 `json:"variety_id" validate:"required"`
	VarietyName     *string                `json:"variety_name" validate:"omitempty,min=2,max=100"`
	DurationDays    *int                   `json:"duration_days" validate:"omitempty,min=1"`
	Characteristics *string                `json:"characteristics" validate:"omitempty,max=1000"`
	Metadata        map[string]interface{} `json:"metadata" validate:"omitempty"`
}

// ListVarietiesRequest represents a request to list varieties for a crop
type ListVarietiesRequest struct {
	BaseRequest
	CropID string  `json:"crop_id" validate:"required"`
	Search *string `json:"search" validate:"omitempty,min=1,max=100"`
	Limit  *int    `json:"limit" validate:"omitempty,min=1,max=100"`
	Offset *int    `json:"offset" validate:"omitempty,min=0"`
}

// GetVarietyRequest represents a request to get a specific variety
type GetVarietyRequest struct {
	BaseRequest
	VarietyID string `json:"variety_id" validate:"required"`
}

// DeleteVarietyRequest represents a request to delete a variety
type DeleteVarietyRequest struct {
	BaseRequest
	VarietyID string `json:"variety_id" validate:"required"`
}

// CreateStageRequest represents a request to create a new crop stage
type CreateStageRequest struct {
	BaseRequest
	CropID              string                 `json:"crop_id" validate:"required"`
	StageName           string                 `json:"stage_name" validate:"required,min=2,max=100"`
	StageOrder          int                    `json:"stage_order" validate:"required,min=0"`
	TypicalDurationDays *int                   `json:"typical_duration_days" validate:"omitempty,min=1"`
	Description         *string                `json:"description" validate:"omitempty,max=1000"`
	Metadata            map[string]interface{} `json:"metadata" validate:"omitempty"`
}

// UpdateStageRequest represents a request to update an existing crop stage
type UpdateStageRequest struct {
	BaseRequest
	StageID             string                 `json:"stage_id" validate:"required"`
	StageName           *string                `json:"stage_name" validate:"omitempty,min=2,max=100"`
	StageOrder          *int                   `json:"stage_order" validate:"omitempty,min=0"`
	TypicalDurationDays *int                   `json:"typical_duration_days" validate:"omitempty,min=1"`
	Description         *string                `json:"description" validate:"omitempty,max=1000"`
	Metadata            map[string]interface{} `json:"metadata" validate:"omitempty"`
}

// ListStagesRequest represents a request to list stages for a crop
type ListStagesRequest struct {
	BaseRequest
	CropID string `json:"crop_id" validate:"required"`
	Limit  *int   `json:"limit" validate:"omitempty,min=1,max=100"`
	Offset *int   `json:"offset" validate:"omitempty,min=0"`
}

// GetStageRequest represents a request to get a specific stage
type GetStageRequest struct {
	BaseRequest
	StageID string `json:"stage_id" validate:"required"`
}

// DeleteStageRequest represents a request to delete a stage
type DeleteStageRequest struct {
	BaseRequest
	StageID string `json:"stage_id" validate:"required"`
}

// LookupRequest represents a request for lookup data
type LookupRequest struct {
	BaseRequest
	Type string `json:"type" validate:"required,oneof=categories units seasons"`
}
