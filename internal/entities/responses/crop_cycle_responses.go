package responses

import (
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// CropCycleResponse represents a single crop cycle response
type CropCycleResponse struct {
	*base.BaseResponse `json:",inline"`
	Data               *CropCycleData `json:"data"`
}

// CropCycleListResponse represents a list of crop cycles response
type CropCycleListResponse struct {
	*base.PaginatedResponse `json:",inline"`
	Data                    []*CropCycleData `json:"data"`
}

// CropCycleData represents crop cycle data in responses
type CropCycleData struct {
	ID           string            `json:"id"`
	FarmID       string            `json:"farm_id"`
	FarmerID     string            `json:"farmer_id"`
	Season       string            `json:"season"`
	Status       string            `json:"status"`
	StartDate    *time.Time        `json:"start_date"`
	EndDate      *time.Time        `json:"end_date"`
	PlannedCrops []string          `json:"planned_crops"`
	Outcome      map[string]string `json:"outcome"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// NewCropCycleResponse creates a new crop cycle response
func NewCropCycleResponse(cycle *CropCycleData, message string) CropCycleResponse {
	return CropCycleResponse{
		BaseResponse: base.NewSuccessResponse(message, cycle),
		Data:         cycle,
	}
}

// NewCropCycleListResponse creates a new crop cycle list response
func NewCropCycleListResponse(cycles []*CropCycleData, page, pageSize int, totalCount int64) CropCycleListResponse {
	// Convert to interface slice for pagination
	var data []any
	for _, c := range cycles {
		data = append(data, c)
	}

	paginationInfo := base.NewPaginationInfo(page, pageSize, int(totalCount))
	return CropCycleListResponse{
		PaginatedResponse: base.NewPaginatedResponse("Crop cycles retrieved successfully", data, paginationInfo),
		Data:              cycles,
	}
}

// SetRequestID sets the request ID for tracking
func (r *CropCycleResponse) SetRequestID(requestID string) {
	r.BaseResponse.RequestID = requestID
}

// SetRequestID sets the request ID for tracking
func (r *CropCycleListResponse) SetRequestID(requestID string) {
	r.PaginatedResponse.RequestID = requestID
}
