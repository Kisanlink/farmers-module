package responses

import (
	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// FarmerLinkageResponse represents a single farmer linkage response
type FarmerLinkageResponse struct {
	*base.BaseResponse
	Data *FarmerLinkageData `json:"data"`
}

// FarmerLinkageData represents farmer linkage data in responses
type FarmerLinkageData struct {
	ID               string  `json:"id"`
	AAAUserID        string  `json:"aaa_user_id"`
	AAAOrgID         string  `json:"aaa_org_id"`
	KisanSathiUserID *string `json:"kisan_sathi_user_id,omitempty"`
	Status           string  `json:"status"`
	LinkedAt         string  `json:"linked_at,omitempty"`
	UnlinkedAt       string  `json:"unlinked_at,omitempty"`
	CreatedAt        string  `json:"created_at,omitempty"`
	UpdatedAt        string  `json:"updated_at,omitempty"`
}

// NewFarmerLinkageResponse creates a new farmer linkage response
func NewFarmerLinkageResponse(linkage *FarmerLinkageData, message string) *FarmerLinkageResponse {
	return &FarmerLinkageResponse{
		BaseResponse: base.NewSuccessResponse(message, linkage),
		Data:         linkage,
	}
}

// SetRequestID sets the request ID for tracking
func (r *FarmerLinkageResponse) SetRequestID(requestID string) {
	if r.BaseResponse != nil {
		r.BaseResponse.RequestID = requestID
	}
}

// BulkLinkFarmersResponse represents the response for bulk farmer linkage operations
type BulkLinkFarmersResponse struct {
	*base.BaseResponse
	Data *BulkLinkFarmersData `json:"data"`
}

// BulkLinkFarmersData represents bulk farmer linkage result data
type BulkLinkFarmersData struct {
	AAAOrgID     string           `json:"aaa_org_id"`
	TotalCount   int              `json:"total_count"`
	SuccessCount int              `json:"success_count"`
	FailureCount int              `json:"failure_count"`
	SkippedCount int              `json:"skipped_count"` // Already linked
	Results      []BulkLinkResult `json:"results"`
}

// BulkLinkResult represents the result of linking a single farmer
type BulkLinkResult struct {
	AAAUserID string `json:"aaa_user_id"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
	Status    string `json:"status"` // LINKED, ALREADY_LINKED, FAILED, UNLINKED
}

// NewBulkLinkFarmersResponse creates a new bulk link farmers response
func NewBulkLinkFarmersResponse(data *BulkLinkFarmersData, message string) *BulkLinkFarmersResponse {
	return &BulkLinkFarmersResponse{
		BaseResponse: base.NewSuccessResponse(message, data),
		Data:         data,
	}
}

// SetRequestID sets the request ID for tracking
func (r *BulkLinkFarmersResponse) SetRequestID(requestID string) {
	if r.BaseResponse != nil {
		r.BaseResponse.RequestID = requestID
	}
}
