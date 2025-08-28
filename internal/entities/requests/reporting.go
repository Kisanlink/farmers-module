package requests

import (
	"time"
)

// ExportFarmerPortfolioRequest represents a request to export farmer portfolio data
type ExportFarmerPortfolioRequest struct {
	BaseRequest
	FarmerID  string     `json:"farmer_id" validate:"required"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	Season    string     `json:"season,omitempty" validate:"omitempty,oneof=RABI KHARIF ZAID"`
	Format    string     `json:"format,omitempty" validate:"omitempty,oneof=json csv"`
}

// OrgDashboardCountersRequest represents a request for organizational dashboard counters
type OrgDashboardCountersRequest struct {
	BaseRequest
	Season    string     `json:"season,omitempty" validate:"omitempty,oneof=RABI KHARIF ZAID"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
}
