package requests

// LinkFarmerRequest represents a request to link farmer to FPO
type LinkFarmerRequest struct {
	BaseRequest
	AAAUserID string `json:"aaa_user_id" validate:"required" example:"usr_123e4567-e89b-12d3-a456-426614174000"`
	AAAOrgID  string `json:"aaa_org_id" validate:"required" example:"org_123e4567-e89b-12d3-a456-426614174000"`
}

// UnlinkFarmerRequest represents a request to unlink farmer from FPO
type UnlinkFarmerRequest struct {
	BaseRequest
	AAAUserID string `json:"aaa_user_id" validate:"required" example:"usr_123e4567-e89b-12d3-a456-426614174000"`
	AAAOrgID  string `json:"aaa_org_id" validate:"required" example:"org_123e4567-e89b-12d3-a456-426614174000"`
}

// NewLinkFarmerRequest creates a new link farmer request
func NewLinkFarmerRequest() LinkFarmerRequest {
	return LinkFarmerRequest{
		BaseRequest: NewBaseRequest(),
	}
}

// NewUnlinkFarmerRequest creates a new unlink farmer request
func NewUnlinkFarmerRequest() UnlinkFarmerRequest {
	return UnlinkFarmerRequest{
		BaseRequest: NewBaseRequest(),
	}
}
