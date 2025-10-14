package requests

import (
	"fmt"
	"time"
)

// StartCycleRequest represents a request to start a new crop cycle
type StartCycleRequest struct {
	BaseRequest
	FarmID    string    `json:"farm_id" validate:"required" example:"farm_123e4567-e89b-12d3-a456-426614174000"`
	AreaHa    *float64  `json:"area_ha" validate:"omitempty,gt=0" example:"5.5"`
	Season    string    `json:"season" validate:"required,oneof=RABI KHARIF ZAID" example:"RABI"`
	StartDate time.Time `json:"start_date" validate:"required" example:"2024-11-01T00:00:00Z"`
	CropID    string    `json:"crop_id" validate:"required" example:"crop_123e4567-e89b-12d3-a456-426614174000"`
	VarietyID *string   `json:"variety_id,omitempty" example:"variety_123e4567-e89b-12d3-a456-426614174000"`
}

// UpdateCycleRequest represents a request to update an existing crop cycle
type UpdateCycleRequest struct {
	BaseRequest
	ID        string     `json:"id" validate:"required" example:"cycle_123e4567-e89b-12d3-a456-426614174000"`
	AreaHa    *float64   `json:"area_ha,omitempty" validate:"omitempty,gt=0" example:"6.0"`
	Season    *string    `json:"season,omitempty" validate:"omitempty,oneof=RABI KHARIF ZAID" example:"RABI"`
	StartDate *time.Time `json:"start_date,omitempty" example:"2024-11-05T00:00:00Z"`
	CropID    *string    `json:"crop_id,omitempty" example:"crop_123e4567-e89b-12d3-a456-426614174000"`
	VarietyID *string    `json:"variety_id,omitempty" example:"variety_123e4567-e89b-12d3-a456-426614174000"`
}

// EndCycleRequest represents a request to end a crop cycle
type EndCycleRequest struct {
	BaseRequest
	ID      string                 `json:"id" validate:"required" example:"cycle_123e4567-e89b-12d3-a456-426614174000"`
	Status  string                 `json:"status" validate:"required,oneof=COMPLETED CANCELLED" example:"COMPLETED"`
	EndDate time.Time              `json:"end_date" validate:"required" example:"2024-03-15T00:00:00Z"`
	Outcome map[string]interface{} `json:"outcome,omitempty" example:"yield_kg:2500,quality:good,notes:good_harvest"`
}

// ListCyclesRequest represents a request to list crop cycles with filtering
type ListCyclesRequest struct {
	FilterRequest
	FarmID   string   `json:"farm_id,omitempty" example:"farm_123e4567-e89b-12d3-a456-426614174000"`
	FarmerID string   `json:"farmer_id,omitempty" example:"farmer_123e4567-e89b-12d3-a456-426614174000"`
	Season   string   `json:"season,omitempty" validate:"omitempty,oneof=RABI KHARIF ZAID" example:"RABI"`
	Status   string   `json:"status,omitempty" validate:"omitempty,oneof=PLANNED ACTIVE COMPLETED CANCELLED" example:"ACTIVE"`
	MinArea  *float64 `json:"min_area,omitempty" validate:"omitempty,gt=0" example:"1.0"`
	MaxArea  *float64 `json:"max_area,omitempty" validate:"omitempty,gt=0" example:"10.0"`
}

// GetCycleRequest represents a request to retrieve a crop cycle
type GetCycleRequest struct {
	BaseRequest
	ID string `json:"id" validate:"required" example:"cycle_123e4567-e89b-12d3-a456-426614174000"`
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
		Outcome:     make(map[string]interface{}),
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

// Validate validates the start cycle request
func (req *StartCycleRequest) Validate() error {
	if req.CropID == "" {
		return fmt.Errorf("crop_id is required")
	}
	return nil
}

// Validate validates the update cycle request
func (req *UpdateCycleRequest) Validate() error {
	// For updates, we don't enforce the crop requirement as the existing data might be valid
	return nil
}
