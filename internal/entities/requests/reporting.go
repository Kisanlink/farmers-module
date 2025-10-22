package requests

import (
	"time"
)

// ExportFarmerPortfolioRequest represents a request to export farmer portfolio data
type ExportFarmerPortfolioRequest struct {
	BaseRequest
	FarmerID  string     `json:"farmer_id" validate:"required" example:"farmer_123e4567-e89b-12d3-a456-426614174000"`
	StartDate *time.Time `json:"start_date,omitempty" example:"2024-01-01T00:00:00Z"`
	EndDate   *time.Time `json:"end_date,omitempty" example:"2024-12-31T23:59:59Z"`
	Season    string     `json:"season,omitempty" validate:"omitempty,oneof=RABI KHARIF ZAID PERENNIAL OTHER" example:"RABI"`
	Format    string     `json:"format,omitempty" validate:"omitempty,oneof=json csv" example:"json"`
}

// OrgDashboardCountersRequest represents a request for organizational dashboard counters
type OrgDashboardCountersRequest struct {
	BaseRequest
	Season    string     `json:"season,omitempty" validate:"omitempty,oneof=RABI KHARIF ZAID PERENNIAL OTHER" example:"RABI"`
	StartDate *time.Time `json:"start_date,omitempty" example:"2024-01-01T00:00:00Z"`
	EndDate   *time.Time `json:"end_date,omitempty" example:"2024-12-31T23:59:59Z"`
}
