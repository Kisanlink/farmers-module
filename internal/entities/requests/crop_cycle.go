package requests

import (
	"time"

	"github.com/Kisanlink/farmers-module/internal/entities/crop_cycle"
)

// StartCycleRequest represents a request to start a new crop cycle
type StartCycleRequest struct {
	BaseRequest
	FarmID    string     `json:"farm_id" validate:"required"`
	Season    string     `json:"season" validate:"required,oneof=RABI KHARIF ZAID"`
	StartDate *time.Time `json:"start_date,omitempty"`

	// Enhanced crop details
	CropID      *string             `json:"crop_id,omitempty"`
	VarietyID   *string             `json:"variety_id,omitempty"`
	CropType    crop_cycle.CropType `json:"crop_type" validate:"required,oneof=ANNUAL PERENNIAL"`
	CropName    *string             `json:"crop_name,omitempty"`
	VarietyName *string             `json:"variety_name,omitempty"`

	// Area and planting details
	Acreage       *float64 `json:"acreage,omitempty" validate:"omitempty,min=0"`
	NumberOfTrees *int     `json:"number_of_trees,omitempty" validate:"omitempty,min=0"`

	// Date fields
	SowingTransplantingDate *time.Time `json:"sowing_transplanting_date,omitempty"`

	// Additional data
	PlannedCrops []string               `json:"planned_crops,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateCycleRequest represents a request to update an existing crop cycle
type UpdateCycleRequest struct {
	BaseRequest
	ID        string     `json:"id" validate:"required"`
	Season    *string    `json:"season,omitempty" validate:"omitempty,oneof=RABI KHARIF ZAID"`
	StartDate *time.Time `json:"start_date,omitempty"`

	// Enhanced crop details
	CropID      *string              `json:"crop_id,omitempty"`
	VarietyID   *string              `json:"variety_id,omitempty"`
	CropType    *crop_cycle.CropType `json:"crop_type,omitempty" validate:"omitempty,oneof=ANNUAL PERENNIAL"`
	CropName    *string              `json:"crop_name,omitempty"`
	VarietyName *string              `json:"variety_name,omitempty"`

	// Area and planting details
	Acreage       *float64 `json:"acreage,omitempty" validate:"omitempty,min=0"`
	NumberOfTrees *int     `json:"number_of_trees,omitempty" validate:"omitempty,min=0"`

	// Date fields
	SowingTransplantingDate *time.Time `json:"sowing_transplanting_date,omitempty"`
	HarvestDate             *time.Time `json:"harvest_date,omitempty"`

	// Yield information
	YieldPerAcre *float64 `json:"yield_per_acre,omitempty" validate:"omitempty,min=0"`
	YieldPerTree *float64 `json:"yield_per_tree,omitempty" validate:"omitempty,min=0"`
	TotalYield   *float64 `json:"total_yield,omitempty" validate:"omitempty,min=0"`
	YieldUnit    *string  `json:"yield_unit,omitempty"`

	// Tree age information
	TreeAgeRangeMin *int `json:"tree_age_range_min,omitempty" validate:"omitempty,min=0"`
	TreeAgeRangeMax *int `json:"tree_age_range_max,omitempty" validate:"omitempty,min=0"`

	// Media and documentation
	ImageURL   *string `json:"image_url,omitempty"`
	DocumentID *string `json:"document_id,omitempty"`

	// Additional data
	PlannedCrops []string               `json:"planned_crops,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// EndCycleRequest represents a request to end a crop cycle
type EndCycleRequest struct {
	BaseRequest
	ID      string            `json:"id" validate:"required"`
	Status  string            `json:"status" validate:"required,oneof=COMPLETED CANCELLED"`
	EndDate time.Time         `json:"end_date" validate:"required"`
	Outcome map[string]string `json:"outcome,omitempty"`
}

// ListCyclesRequest represents a request to list crop cycles with filtering
type ListCyclesRequest struct {
	FilterRequest
	FarmID   string `json:"farm_id,omitempty"`
	FarmerID string `json:"farmer_id,omitempty"`
	Season   string `json:"season,omitempty" validate:"omitempty,oneof=RABI KHARIF ZAID"`
	Status   string `json:"status,omitempty" validate:"omitempty,oneof=PLANNED ACTIVE COMPLETED CANCELLED"`
}

// GetCycleRequest represents a request to retrieve a crop cycle
type GetCycleRequest struct {
	BaseRequest
	ID string `json:"id" validate:"required"`
}

// RecordHarvestRequest represents a request to record harvest data
type RecordHarvestRequest struct {
	BaseRequest
	CycleID      string                 `json:"cycle_id" validate:"required"`
	HarvestDate  time.Time              `json:"harvest_date" validate:"required"`
	YieldPerAcre *float64               `json:"yield_per_acre,omitempty" validate:"omitempty,min=0"`
	YieldPerTree *float64               `json:"yield_per_tree,omitempty" validate:"omitempty,min=0"`
	TotalYield   *float64               `json:"total_yield,omitempty" validate:"omitempty,min=0"`
	YieldUnit    *string                `json:"yield_unit,omitempty"`
	ImageURL     *string                `json:"image_url,omitempty"`
	DocumentID   *string                `json:"document_id,omitempty"`
	ReportData   map[string]interface{} `json:"report_data,omitempty"`
}

// UploadReportRequest represents a request to upload a report for a crop cycle
type UploadReportRequest struct {
	BaseRequest
	CycleID    string                 `json:"cycle_id" validate:"required"`
	ReportType string                 `json:"report_type" validate:"required,oneof=PROGRESS HARVEST FINAL"`
	ReportData map[string]interface{} `json:"report_data" validate:"required"`
	ImageURL   *string                `json:"image_url,omitempty"`
	DocumentID *string                `json:"document_id,omitempty"`
}

// NewStartCycleRequest creates a new start cycle request
func NewStartCycleRequest() StartCycleRequest {
	return StartCycleRequest{
		BaseRequest: NewBaseRequest(),
	}
}

// NewUpdateCycleRequest creates a new update cycle request
func NewUpdateCycleRequest() UpdateCycleRequest {
	return UpdateCycleRequest{
		BaseRequest: NewBaseRequest(),
	}
}

// NewEndCycleRequest creates a new end cycle request
func NewEndCycleRequest() EndCycleRequest {
	return EndCycleRequest{
		BaseRequest: NewBaseRequest(),
		Outcome:     make(map[string]string),
	}
}

// NewListCyclesRequest creates a new list cycles request
func NewListCyclesRequest() ListCyclesRequest {
	return ListCyclesRequest{
		FilterRequest: FilterRequest{
			PaginationRequest: NewPaginationRequest(1, 10),
		},
	}
}

// NewGetCycleRequest creates a new get cycle request
func NewGetCycleRequest() GetCycleRequest {
	return GetCycleRequest{
		BaseRequest: NewBaseRequest(),
	}
}
