package requests

// LinkFarmerRequest represents a request to link farmer to FPO
type LinkFarmerRequest struct {
	BaseRequest
	AAAUserID string `json:"aaa_user_id" validate:"required"`
	AAAOrgID  string `json:"aaa_org_id" validate:"required"`
}

// UnlinkFarmerRequest represents a request to unlink farmer from FPO
type UnlinkFarmerRequest struct {
	BaseRequest
	AAAUserID string `json:"aaa_user_id" validate:"required"`
	AAAOrgID  string `json:"aaa_org_id" validate:"required"`
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
