package responses

// Request types for Swagger documentation
// These are type aliases to the actual request types in the requests package

// CreateFarmerRequest represents a request to create a new farmer
type CreateFarmerRequest struct {
	BaseRequest
	AAAUserID        string            `json:"aaa_user_id" validate:"required"`
	AAAOrgID         string            `json:"aaa_org_id" validate:"required"`
	KisanSathiUserID *string           `json:"kisan_sathi_user_id,omitempty"`
	Profile          FarmerProfileData `json:"profile,omitempty"`
}

// UpdateFarmerRequest represents a request to update an existing farmer
type UpdateFarmerRequest struct {
	BaseRequest
	AAAUserID        string            `json:"aaa_user_id" validate:"required"`
	AAAOrgID         string            `json:"aaa_org_id" validate:"required"`
	KisanSathiUserID *string           `json:"kisan_sathi_user_id,omitempty"`
	Profile          FarmerProfileData `json:"profile,omitempty"`
}

// DeleteFarmerRequest represents a request to delete a farmer
type DeleteFarmerRequest struct {
	BaseRequest
	AAAUserID string `json:"aaa_user_id" validate:"required"`
	AAAOrgID  string `json:"aaa_org_id" validate:"required"`
}

// GetFarmerRequest represents a request to retrieve a farmer
type GetFarmerRequest struct {
	BaseRequest
	AAAUserID string `json:"aaa_user_id" validate:"required"`
	AAAOrgID  string `json:"aaa_org_id" validate:"required"`
}

// ListFarmersRequest represents a request to list farmers with filtering
type ListFarmersRequest struct {
	FilterRequest
	AAAOrgID         string `json:"aaa_org_id,omitempty"`
	KisanSathiUserID string `json:"kisan_sathi_user_id,omitempty"`
	Page             int    `json:"page,omitempty"`
	PageSize         int    `json:"page_size,omitempty"`
}

// FilterRequest represents base filtering request
type FilterRequest struct {
	BaseRequest
}

// BaseRequest is already defined in base.go
