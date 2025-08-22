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
