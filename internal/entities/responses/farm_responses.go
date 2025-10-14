package responses

import (
	"time"

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
type FarmData struct {
	ID        string                 `json:"id" example:"farm_123e4567-e89b-12d3-a456-426614174000"`
	FarmerID  string                 `json:"farmer_id" example:"FMRR0000000001"`
	AAAUserID string                 `json:"aaa_user_id" example:"usr_123e4567-e89b-12d3-a456-426614174000"`
	AAAOrgID  string                 `json:"aaa_org_id" example:"org_123e4567-e89b-12d3-a456-426614174000"`
	Name      string                 `json:"name" example:"North Field Farm"`
	Geometry  string                 `json:"geometry" example:"POLYGON((75.85 22.71, 75.85663 22.71, 75.85663 22.71663, 75.85 22.71663, 75.85 22.71))"`
	AreaHa    float64                `json:"area_ha" example:"2.5"`
	Metadata  map[string]interface{} `json:"metadata" example:"soil_type:loamy,irrigation:drip"`
	CreatedAt time.Time              `json:"created_at" example:"2024-01-15T10:30:00Z"`
	UpdatedAt time.Time              `json:"updated_at" example:"2024-01-20T15:45:00Z"`
}

// FarmOverlapData represents farm overlap check result
type FarmOverlapData struct {
	HasOverlap       bool     `json:"has_overlap" example:"true"`
	OverlapArea      float64  `json:"overlap_area,omitempty" example:"0.25"`
	OverlappingFarms []string `json:"overlapping_farms,omitempty" example:"farm_123e4567-e89b-12d3-a456-426614174001,farm_123e4567-e89b-12d3-a456-426614174002"`
	Message          string   `json:"message,omitempty" example:"Farm boundary overlaps with 2 existing farms"`
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
