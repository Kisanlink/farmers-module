package responses

// FPORefResponse represents a single FPO reference response
type FPORefResponse struct {
	BaseResponse
	Data *FPORefData `json:"data"`
}

// FPORefData represents FPO reference data in responses
type FPORefData struct {
	ID             string            `json:"id"`
	AAAOrgID       string            `json:"aaa_org_id"`
	BusinessConfig string            `json:"business_config"`
	Status         string            `json:"status"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	CreatedBy      string            `json:"created_by,omitempty"`
	CreatedAt      string            `json:"created_at,omitempty"`
	UpdatedAt      string            `json:"updated_at,omitempty"`
}

// NewFPORefResponse creates a new FPO reference response
func NewFPORefResponse(fpoRef *FPORefData, message string) FPORefResponse {
	return FPORefResponse{
		BaseResponse: NewBaseResponse(),
		Data:         fpoRef,
	}
}

// SetRequestID sets the request ID for tracking
func (r *FPORefResponse) SetRequestID(requestID string) {
	r.BaseResponse.SetResponseID(requestID)
}
