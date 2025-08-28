package responses

import (
	"time"
)

// FarmSummary represents a summary of farm data
type FarmSummary struct {
	FarmID   string  `json:"farm_id"`
	Name     string  `json:"name"`
	AreaHa   float64 `json:"area_ha"`
	Location string  `json:"location,omitempty"`
}

// CycleSummary represents a summary of crop cycle data
type CycleSummary struct {
	CycleID      string     `json:"cycle_id"`
	FarmID       string     `json:"farm_id"`
	Season       string     `json:"season"`
	Status       string     `json:"status"`
	StartDate    *time.Time `json:"start_date,omitempty"`
	EndDate      *time.Time `json:"end_date,omitempty"`
	PlannedCrops []string   `json:"planned_crops,omitempty"`
}

// ActivitySummary represents a summary of farm activity data
type ActivitySummary struct {
	ActivityID   string     `json:"activity_id"`
	CycleID      string     `json:"cycle_id"`
	ActivityType string     `json:"activity_type"`
	Status       string     `json:"status"`
	PlannedAt    *time.Time `json:"planned_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
}

// FarmerPortfolioData represents aggregated farmer portfolio data
type FarmerPortfolioData struct {
	FarmerID   string            `json:"farmer_id"`
	FarmerName string            `json:"farmer_name"`
	OrgID      string            `json:"org_id"`
	Farms      []FarmSummary     `json:"farms"`
	Cycles     []CycleSummary    `json:"cycles"`
	Activities []ActivitySummary `json:"activities"`
	Summary    PortfolioSummary  `json:"summary"`
}

// PortfolioSummary represents summary statistics for a farmer's portfolio
type PortfolioSummary struct {
	TotalFarms          int     `json:"total_farms"`
	TotalAreaHa         float64 `json:"total_area_ha"`
	TotalCycles         int     `json:"total_cycles"`
	ActiveCycles        int     `json:"active_cycles"`
	CompletedCycles     int     `json:"completed_cycles"`
	TotalActivities     int     `json:"total_activities"`
	CompletedActivities int     `json:"completed_activities"`
}

// ExportFarmerPortfolioResponse represents the response for farmer portfolio export
type ExportFarmerPortfolioResponse struct {
	BaseResponse
	Data FarmerPortfolioData `json:"data"`
}

// OrgCounters represents organizational KPI counters
type OrgCounters struct {
	TotalFarmers        int     `json:"total_farmers"`
	ActiveFarmers       int     `json:"active_farmers"`
	TotalFarms          int     `json:"total_farms"`
	TotalAreaHa         float64 `json:"total_area_ha"`
	TotalCycles         int     `json:"total_cycles"`
	ActiveCycles        int     `json:"active_cycles"`
	CompletedCycles     int     `json:"completed_cycles"`
	TotalActivities     int     `json:"total_activities"`
	CompletedActivities int     `json:"completed_activities"`
}

// SeasonalCounters represents counters broken down by season
type SeasonalCounters struct {
	Season     string  `json:"season"`
	Cycles     int     `json:"cycles"`
	AreaHa     float64 `json:"area_ha"`
	Activities int     `json:"activities"`
}

// StatusCounters represents counters broken down by status
type StatusCounters struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

// OrgDashboardData represents organizational dashboard data
type OrgDashboardData struct {
	OrgID                   string             `json:"org_id"`
	OrgName                 string             `json:"org_name,omitempty"`
	Counters                OrgCounters        `json:"counters"`
	SeasonalBreakdown       []SeasonalCounters `json:"seasonal_breakdown"`
	CycleStatusBreakdown    []StatusCounters   `json:"cycle_status_breakdown"`
	ActivityStatusBreakdown []StatusCounters   `json:"activity_status_breakdown"`
	GeneratedAt             time.Time          `json:"generated_at"`
}

// OrgDashboardCountersResponse represents the response for organizational dashboard counters
type OrgDashboardCountersResponse struct {
	BaseResponse
	Data OrgDashboardData `json:"data"`
}
