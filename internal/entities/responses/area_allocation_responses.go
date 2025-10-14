package responses

import (
	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// AreaAllocationSummaryResponse represents farm area allocation summary
type AreaAllocationSummaryResponse struct {
	*base.BaseResponse `json:",inline"`
	Data               *AreaAllocationSummaryData `json:"data"`
}

// AreaAllocationSummaryData contains area allocation details
type AreaAllocationSummaryData struct {
	FarmID             string             `json:"farm_id"`
	FarmName           string             `json:"farm_name,omitempty"`
	TotalAreaHa        float64            `json:"total_area_ha"`
	AllocatedAreaHa    float64            `json:"allocated_area_ha"`
	AvailableAreaHa    float64            `json:"available_area_ha"`
	UtilizationPercent float64            `json:"utilization_percentage"`
	ActiveCyclesCount  int64              `json:"active_cycles_count"`
	PlannedCyclesCount int64              `json:"planned_cycles_count"`
	Allocations        []*CycleAllocation `json:"allocations,omitempty"`
}

// CycleAllocation represents a single crop cycle allocation
type CycleAllocation struct {
	CropCycleID string  `json:"crop_cycle_id"`
	CropName    string  `json:"crop_name"`
	VarietyName string  `json:"variety_name,omitempty"`
	AreaHa      float64 `json:"area_ha"`
	Status      string  `json:"status"`
	Season      string  `json:"season"`
	StartDate   string  `json:"start_date,omitempty"`
}

// NewAreaAllocationSummaryResponse creates a new area allocation summary response
func NewAreaAllocationSummaryResponse(data *AreaAllocationSummaryData, message string) AreaAllocationSummaryResponse {
	return AreaAllocationSummaryResponse{
		BaseResponse: base.NewSuccessResponse(message, data),
		Data:         data,
	}
}

// SetRequestID sets the request ID for tracking
func (r *AreaAllocationSummaryResponse) SetRequestID(requestID string) {
	r.BaseResponse.RequestID = requestID
}
