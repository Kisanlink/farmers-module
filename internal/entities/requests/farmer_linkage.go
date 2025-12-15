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

// BulkLinkFarmersRequest represents a request to link multiple farmers to an FPO
type BulkLinkFarmersRequest struct {
	BaseRequest
	AAAOrgID        string   `json:"aaa_org_id" validate:"required" example:"ORGN00000005"`
	AAAUserIDs      []string `json:"aaa_user_ids" validate:"required,min=1,max=1000" example:"[\"USR00000001\",\"USR00000002\"]"`
	ContinueOnError bool     `json:"continue_on_error" example:"true"` // Continue processing on individual failures
}

// BulkUnlinkFarmersRequest represents a request to unlink multiple farmers from an FPO
type BulkUnlinkFarmersRequest struct {
	BaseRequest
	AAAOrgID        string   `json:"aaa_org_id" validate:"required" example:"ORGN00000005"`
	AAAUserIDs      []string `json:"aaa_user_ids" validate:"required,min=1,max=1000" example:"[\"USR00000001\",\"USR00000002\"]"`
	ContinueOnError bool     `json:"continue_on_error" example:"true"` // Continue processing on individual failures
}

// BulkLinkResult represents the result of linking a single farmer
type BulkLinkResult struct {
	AAAUserID string `json:"aaa_user_id"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
	Status    string `json:"status,omitempty"` // ACTIVE, ALREADY_LINKED, FAILED
}

// NewBulkLinkFarmersRequest creates a new bulk link farmers request
func NewBulkLinkFarmersRequest() BulkLinkFarmersRequest {
	return BulkLinkFarmersRequest{
		BaseRequest:     NewBaseRequest(),
		ContinueOnError: true,
	}
}

// NewBulkUnlinkFarmersRequest creates a new bulk unlink farmers request
func NewBulkUnlinkFarmersRequest() BulkUnlinkFarmersRequest {
	return BulkUnlinkFarmersRequest{
		BaseRequest:     NewBaseRequest(),
		ContinueOnError: true,
	}
}
