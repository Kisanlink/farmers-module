package requests

import "time"

// StartCycleRequest represents a request to start a new crop cycle
type StartCycleRequest struct {
	BaseRequest
	FarmID       string    `json:"farm_id" validate:"required"`
	Season       string    `json:"season" validate:"required,oneof=RABI KHARIF ZAID"`
	StartDate    time.Time `json:"start_date" validate:"required"`
	PlannedCrops []string  `json:"planned_crops" validate:"required,min=1"`
}

// UpdateCycleRequest represents a request to update an existing crop cycle
type UpdateCycleRequest struct {
	BaseRequest
	ID           string     `json:"id" validate:"required"`
	Season       *string    `json:"season,omitempty" validate:"omitempty,oneof=RABI KHARIF ZAID"`
	StartDate    *time.Time `json:"start_date,omitempty"`
	PlannedCrops []string   `json:"planned_crops,omitempty"`
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
