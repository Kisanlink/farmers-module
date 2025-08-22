package responses

import (
	"github.com/Kisanlink/farmers-module/internal/entities/farm"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// FarmResponse represents a single farm response
type FarmResponse struct {
	*base.BaseResponse `json:",inline"`
	Data               *FarmData `json:"data"`
}

// FarmListResponse represents a list of farms response
type FarmListResponse struct {
	*base.PaginatedResponse `json:",inline"`
	Data                    []*FarmData `json:"data"`
}

// FarmOverlapResponse represents a farm overlap check response
type FarmOverlapResponse struct {
	*base.BaseResponse `json:",inline"`
	Data               *FarmOverlapData `json:"data"`
}

// FarmData represents farm data in responses
// Note: FarmData is already defined in farmer_responses.go as farm.Farm

// GeometryData represents geometric data in responses
type GeometryData = farm.Geometry

// FarmOverlapData represents farm overlap check result
type FarmOverlapData struct {
	HasOverlap       bool     `json:"has_overlap"`
	OverlapArea      float64  `json:"overlap_area,omitempty"`
	OverlappingFarms []string `json:"overlapping_farms,omitempty"`
	Message          string   `json:"message,omitempty"`
}

// NewFarmResponse creates a new farm response
func NewFarmResponse(farm *FarmData, message string) FarmResponse {
	return FarmResponse{
		BaseResponse: base.NewSuccessResponse(message, farm),
		Data:         farm,
	}
}

// NewFarmListResponse creates a new farm list response
func NewFarmListResponse(farms []*FarmData, page, pageSize int, totalCount int64) FarmListResponse {
	// Convert to interface slice for pagination
	var data []interface{}
	for _, f := range farms {
		data = append(data, f)
	}

	paginationInfo := base.NewPaginationInfo(page, pageSize, int(totalCount))
	return FarmListResponse{
		PaginatedResponse: base.NewPaginatedResponse("Farms retrieved successfully", data, paginationInfo),
		Data:              farms,
	}
}

// NewFarmOverlapResponse creates a new farm overlap response
func NewFarmOverlapResponse(overlap *FarmOverlapData, message string) FarmOverlapResponse {
	return FarmOverlapResponse{
		BaseResponse: base.NewSuccessResponse(message, overlap),
		Data:         overlap,
	}
}

// SetRequestID sets the request ID for tracking
func (r *FarmResponse) SetRequestID(requestID string) {
	r.BaseResponse.RequestID = requestID
}

// SetRequestID sets the request ID for tracking
func (r *FarmListResponse) SetRequestID(requestID string) {
	r.PaginatedResponse.RequestID = requestID
}

// SetRequestID sets the request ID for tracking
func (r *FarmOverlapResponse) SetRequestID(requestID string) {
	r.BaseResponse.RequestID = requestID
}
