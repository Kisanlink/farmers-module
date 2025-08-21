package responses

import (
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
)

// FarmResponse represents a single farm response
type FarmResponse struct {
	responses.BaseResponse
	Data *FarmData `json:"data"`
}

// FarmListResponse represents a list of farms response
type FarmListResponse struct {
	responses.ListResponse
	Data []*FarmData `json:"data"`
}

// FarmOverlapResponse represents a farm overlap check response
type FarmOverlapResponse struct {
	responses.BaseResponse
	Data *FarmOverlapData `json:"data"`
}

// FarmData represents farm data in responses
type FarmData struct {
	ID              string            `json:"id"`
	AAAFarmerUserID string            `json:"aaa_farmer_user_id"`
	AAAOrgID        string            `json:"aaa_org_id"`
	Geometry        GeometryData      `json:"geometry,omitempty"`
	AreaHa          float64           `json:"area_ha"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	CreatedBy       string            `json:"created_by,omitempty"`
	CreatedAt       string            `json:"created_at,omitempty"`
	UpdatedAt       string            `json:"updated_at,omitempty"`
}

// GeometryData represents geometric data in responses
type GeometryData struct {
	WKT string `json:"wkt,omitempty"` // Well-Known Text format
	WKB []byte `json:"wkb,omitempty"` // Well-Known Binary format
}

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
		BaseResponse: responses.NewBaseResponse(),
		Data:         farm,
	}
}

// NewFarmListResponse creates a new farm list response
func NewFarmListResponse(farms []*FarmData, page, pageSize int, totalCount int64) FarmListResponse {
	// Convert to interface slice for NewListResponse
	var responseData []interface{}
	for _, f := range farms {
		responseData = append(responseData, f)
	}

	return FarmListResponse{
		ListResponse: responses.NewListResponse(responseData, page, pageSize, totalCount),
		Data:         farms,
	}
}

// NewFarmOverlapResponse creates a new farm overlap response
func NewFarmOverlapResponse(overlap *FarmOverlapData, message string) FarmOverlapResponse {
	return FarmOverlapResponse{
		BaseResponse: responses.NewBaseResponse(),
		Data:         overlap,
	}
}

// SetRequestID sets the request ID for tracking
func (r *FarmResponse) SetRequestID(requestID string) {
	r.BaseResponse.SetResponseID(requestID)
}

// SetRequestID sets the request ID for tracking
func (r *FarmListResponse) SetRequestID(requestID string) {
	r.ListResponse.PaginationResponse.BaseResponse.SetResponseID(requestID)
}

// SetRequestID sets the request ID for tracking
func (r *FarmOverlapResponse) SetRequestID(requestID string) {
	r.BaseResponse.SetResponseID(requestID)
}
