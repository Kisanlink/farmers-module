package responses

import (
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// FarmActivityResponse represents a single farm activity response
type FarmActivityResponse struct {
	*base.BaseResponse `json:",inline"`
	Data               *FarmActivityData `json:"data"`
}

// FarmActivityListResponse represents a list of farm activities response
type FarmActivityListResponse struct {
	*base.PaginatedResponse `json:",inline"`
	Data                    []*FarmActivityData `json:"data"`
}

// FarmActivityData represents farm activity data in responses
type FarmActivityData struct {
	ID           string                 `json:"id"`
	CropCycleID  string                 `json:"crop_cycle_id"`
	ActivityType string                 `json:"activity_type"`
	PlannedAt    *time.Time             `json:"planned_at"`
	CompletedAt  *time.Time             `json:"completed_at"`
	CreatedBy    string                 `json:"created_by"`
	Status       string                 `json:"status"`
	Output       map[string]interface{} `json:"output"`
	Metadata     map[string]interface{} `json:"metadata"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// NewFarmActivityResponse creates a new farm activity response
func NewFarmActivityResponse(data *FarmActivityData, message string) FarmActivityResponse {
	return FarmActivityResponse{
		BaseResponse: base.NewSuccessResponse(message, data),
		Data:         data,
	}
}

// NewFarmActivityListResponse creates a new farm activity list response
func NewFarmActivityListResponse(data []*FarmActivityData, page, pageSize int, totalCount int64) FarmActivityListResponse {
	// Convert to interface slice for pagination
	var interfaceData []any
	for _, a := range data {
		interfaceData = append(interfaceData, a)
	}

	paginationInfo := base.NewPaginationInfo(page, pageSize, int(totalCount))
	return FarmActivityListResponse{
		PaginatedResponse: base.NewPaginatedResponse("Farm activities retrieved successfully", interfaceData, paginationInfo),
		Data:              data,
	}
}

// SetRequestID sets the request ID for tracking
func (r *FarmActivityResponse) SetRequestID(requestID string) {
	r.BaseResponse.RequestID = requestID
}

// SetRequestID sets the request ID for tracking
func (r *FarmActivityListResponse) SetRequestID(requestID string) {
	r.PaginatedResponse.RequestID = requestID
}
