package responses

import (
	"time"

	"github.com/Kisanlink/farmers-module/internal/entities/crop"
	"github.com/Kisanlink/farmers-module/internal/entities/crop_stage"
	"github.com/Kisanlink/farmers-module/internal/entities/crop_variety"
)

// PaginationData represents pagination information
type PaginationData struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// CropData represents crop data in responses
type CropData struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Category         string                 `json:"category"`
	CropDurationDays *int                   `json:"crop_duration_days,omitempty"`
	TypicalUnits     []string               `json:"typical_units"`
	Seasons          []string               `json:"seasons"`
	ImageURL         *string                `json:"image_url,omitempty"`
	DocumentID       *string                `json:"document_id,omitempty"`
	Metadata         map[string]interface{} `json:"metadata"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// CropResponse represents a single crop response
type CropResponse struct {
	BaseResponse
	Data *CropData `json:"data,omitempty"`
}

// CropsListResponse represents a list of crops response
type CropsListResponse struct {
	BaseResponse
	Data       []*CropData     `json:"data"`
	Pagination *PaginationData `json:"pagination,omitempty"`
}

// CropVarietyData represents crop variety data in responses
type CropVarietyData struct {
	ID              string                 `json:"id"`
	CropID          string                 `json:"crop_id"`
	VarietyName     string                 `json:"variety_name"`
	DurationDays    *int                   `json:"duration_days,omitempty"`
	Characteristics *string                `json:"characteristics,omitempty"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// CropVarietyResponse represents a single crop variety response
type CropVarietyResponse struct {
	BaseResponse
	Data *CropVarietyData `json:"data,omitempty"`
}

// CropVarietiesListResponse represents a list of crop varieties response
type CropVarietiesListResponse struct {
	BaseResponse
	Data       []*CropVarietyData `json:"data"`
	Pagination *PaginationData    `json:"pagination,omitempty"`
}

// CropStageData represents crop stage data in responses
type CropStageData struct {
	ID                  string                 `json:"id"`
	CropID              string                 `json:"crop_id"`
	StageName           string                 `json:"stage_name"`
	StageOrder          int                    `json:"stage_order"`
	TypicalDurationDays *int                   `json:"typical_duration_days,omitempty"`
	Description         *string                `json:"description,omitempty"`
	Metadata            map[string]interface{} `json:"metadata"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`
}

// CropStageResponse represents a single crop stage response
type CropStageResponse struct {
	BaseResponse
	Data *CropStageData `json:"data,omitempty"`
}

// CropStagesListResponse represents a list of crop stages response
type CropStagesListResponse struct {
	BaseResponse
	Data       []*CropStageData `json:"data"`
	Pagination *PaginationData  `json:"pagination,omitempty"`
}

// LookupData represents lookup data in responses
type LookupData struct {
	Type  string   `json:"type"`
	Items []string `json:"items"`
}

// LookupResponse represents a lookup response
type LookupResponse struct {
	BaseResponse
	Data *LookupData `json:"data,omitempty"`
}

// Helper functions to create responses

// NewCropResponse creates a new crop response
func NewCropResponse(crop *crop.Crop, message string) *CropResponse {
	return &CropResponse{
		BaseResponse: BaseResponse{
			Success: true,
			Message: message,
		},
		Data: mapCropToData(crop),
	}
}

// NewCropsListResponse creates a new crops list response
func NewCropsListResponse(crops []*crop.Crop, message string) *CropsListResponse {
	data := make([]*CropData, len(crops))
	for i, c := range crops {
		data[i] = mapCropToData(c)
	}

	return &CropsListResponse{
		BaseResponse: BaseResponse{
			Success: true,
			Message: message,
		},
		Data: data,
	}
}

// NewCropVarietyResponse creates a new crop variety response
func NewCropVarietyResponse(variety *crop_variety.CropVariety, message string) *CropVarietyResponse {
	return &CropVarietyResponse{
		BaseResponse: BaseResponse{
			Success: true,
			Message: message,
		},
		Data: mapVarietyToData(variety),
	}
}

// NewCropVarietiesListResponse creates a new crop varieties list response
func NewCropVarietiesListResponse(varieties []*crop_variety.CropVariety, message string) *CropVarietiesListResponse {
	data := make([]*CropVarietyData, len(varieties))
	for i, v := range varieties {
		data[i] = mapVarietyToData(v)
	}

	return &CropVarietiesListResponse{
		BaseResponse: BaseResponse{
			Success: true,
			Message: message,
		},
		Data: data,
	}
}

// NewCropStageResponse creates a new crop stage response
func NewCropStageResponse(stage *crop_stage.CropStage, message string) *CropStageResponse {
	return &CropStageResponse{
		BaseResponse: BaseResponse{
			Success: true,
			Message: message,
		},
		Data: mapStageToData(stage),
	}
}

// NewCropStagesListResponse creates a new crop stages list response
func NewCropStagesListResponse(stages []*crop_stage.CropStage, message string) *CropStagesListResponse {
	data := make([]*CropStageData, len(stages))
	for i, s := range stages {
		data[i] = mapStageToData(s)
	}

	return &CropStagesListResponse{
		BaseResponse: BaseResponse{
			Success: true,
			Message: message,
		},
		Data: data,
	}
}

// NewLookupResponse creates a new lookup response
func NewLookupResponse(lookupType string, items []string, message string) *LookupResponse {
	return &LookupResponse{
		BaseResponse: BaseResponse{
			Success: true,
			Message: message,
		},
		Data: &LookupData{
			Type:  lookupType,
			Items: items,
		},
	}
}

// Mapping functions

func mapCropToData(c *crop.Crop) *CropData {
	if c == nil {
		return nil
	}

	units := make([]string, len(c.TypicalUnits))
	for i, unit := range c.TypicalUnits {
		units[i] = string(unit)
	}

	seasons := make([]string, len(c.Seasons))
	for i, season := range c.Seasons {
		seasons[i] = string(season)
	}

	return &CropData{
		ID:               c.ID,
		Name:             c.Name,
		Category:         string(c.Category),
		CropDurationDays: c.CropDurationDays,
		TypicalUnits:     units,
		Seasons:          seasons,
		ImageURL:         c.ImageURL,
		DocumentID:       c.DocumentID,
		Metadata:         c.Metadata,
		CreatedAt:        c.CreatedAt,
		UpdatedAt:        c.UpdatedAt,
	}
}

func mapVarietyToData(v *crop_variety.CropVariety) *CropVarietyData {
	if v == nil {
		return nil
	}

	return &CropVarietyData{
		ID:              v.ID,
		CropID:          v.CropID,
		VarietyName:     v.VarietyName,
		DurationDays:    v.DurationDays,
		Characteristics: v.Characteristics,
		Metadata:        v.Metadata,
		CreatedAt:       v.CreatedAt,
		UpdatedAt:       v.UpdatedAt,
	}
}

func mapStageToData(s *crop_stage.CropStage) *CropStageData {
	if s == nil {
		return nil
	}

	return &CropStageData{
		ID:                  s.ID,
		CropID:              s.CropID,
		StageName:           s.StageName,
		StageOrder:          s.StageOrder,
		TypicalDurationDays: s.TypicalDurationDays,
		Description:         s.Description,
		Metadata:            s.Metadata,
		CreatedAt:           s.CreatedAt,
		UpdatedAt:           s.UpdatedAt,
	}
}
