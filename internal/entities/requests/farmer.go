package requests

// CreateFarmerRequest represents a request to create a new farmer
// Supports two workflows:
// 1. Provide aaa_user_id + aaa_org_id: Use existing AAA user
// 2. Provide country_code + phone_number + aaa_org_id: Create/find AAA user automatically
//   - If user doesn't exist, creates new user in AAA
//   - If user exists (conflict), retrieves existing user ID from AAA
type CreateFarmerRequest struct {
	BaseRequest
	AAAUserID        string            `json:"aaa_user_id,omitempty" example:"USER00000001"`          // Optional: AAA User ID (if known)
	AAAOrgID         string            `json:"aaa_org_id" validate:"required" example:"ORGN00000001"` // Required: AAA Org ID
	KisanSathiUserID *string           `json:"kisan_sathi_user_id,omitempty" example:"USER00000002"`
	LinkFPOConfig    bool              `json:"link_fpo_config,omitempty" example:"false"` // Optional: Link FPO configuration to farmer (default: false)
	Profile          FarmerProfileData `json:"profile" validate:"required"`               // Required: Farmer profile (must include country_code + phone_number if aaa_user_id not provided)
}

// UpdateFarmerRequest represents a request to update an existing farmer
type UpdateFarmerRequest struct {
	BaseRequest
	FarmerID         string            `json:"farmer_id,omitempty" example:"FMRR0000000001"` // Primary key lookup
	AAAUserID        string            `json:"aaa_user_id,omitempty" example:"USER00000001"` // User ID lookup (no org required)
	AAAOrgID         string            `json:"aaa_org_id,omitempty" example:"ORGN00000001"`  // Optional org filter
	KisanSathiUserID *string           `json:"kisan_sathi_user_id,omitempty" example:"USER00000002"`
	Profile          FarmerProfileData `json:"profile,omitempty"`
}

// DeleteFarmerRequest represents a request to delete a farmer
type DeleteFarmerRequest struct {
	BaseRequest
	FarmerID  string `json:"farmer_id,omitempty" example:"FMRR0000000001"` // Primary key lookup
	AAAUserID string `json:"aaa_user_id,omitempty" example:"USER00000001"` // User ID lookup (no org required)
	AAAOrgID  string `json:"aaa_org_id,omitempty" example:"ORGN00000001"`  // Optional org filter
}

// GetFarmerRequest represents a request to retrieve a farmer
type GetFarmerRequest struct {
	BaseRequest
	FarmerID  string `json:"farmer_id,omitempty" example:"FMRR0000000001"` // Primary key lookup
	AAAUserID string `json:"aaa_user_id,omitempty" example:"USER00000001"` // User ID lookup (no org required)
	AAAOrgID  string `json:"aaa_org_id,omitempty" example:"ORGN00000001"`  // Optional org filter
}

// ListFarmersRequest represents a request to list farmers with filtering
type ListFarmersRequest struct {
	FilterRequest
	AAAOrgID         string `json:"aaa_org_id,omitempty" example:"ORGN00000001"`
	KisanSathiUserID string `json:"kisan_sathi_user_id,omitempty" example:"USER00000002"`
	PhoneNumber      string `json:"phone_number,omitempty" example:"9876543210"`
	Page             int    `json:"page,omitempty" example:"1"`
	PageSize         int    `json:"page_size,omitempty" example:"20"`
}

// FarmerProfileData represents the profile data for a farmer
type FarmerProfileData struct {
	Username    string                 `json:"username,omitempty" example:"ramesh_kumar"`
	FirstName   string                 `json:"first_name,omitempty" example:"Ramesh"`
	LastName    string                 `json:"last_name,omitempty" example:"Kumar"`
	PhoneNumber string                 `json:"phone_number,omitempty" example:"9876543210"`
	CountryCode string                 `json:"country_code,omitempty" example:"+91"`
	Email       string                 `json:"email,omitempty" example:"ramesh.kumar@example.com"`
	DateOfBirth string                 `json:"date_of_birth,omitempty" example:"1980-05-15"`
	Gender      string                 `json:"gender,omitempty" example:"male"`
	Address     AddressData            `json:"address,omitempty"`
	Preferences map[string]interface{} `json:"preferences,omitempty" example:"language:hindi,notification:sms"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// AddressData represents address information
type AddressData struct {
	StreetAddress string `json:"street_address,omitempty" example:"Village Rampur, Post Khandwa"`
	City          string `json:"city,omitempty" example:"Indore"`
	State         string `json:"state,omitempty" example:"Madhya Pradesh"`
	PostalCode    string `json:"postal_code,omitempty" example:"452001"`
	Country       string `json:"country,omitempty" example:"India"`
	Coordinates   string `json:"coordinates,omitempty" example:"POINT(75.8577 22.7196)"` // WKT format for PostGIS
}

// NewCreateFarmerRequest creates a new create farmer request
func NewCreateFarmerRequest() CreateFarmerRequest {
	return CreateFarmerRequest{
		BaseRequest: NewBaseRequest(),
		Profile:     FarmerProfileData{},
	}
}

// NewUpdateFarmerRequest creates a new update farmer request
func NewUpdateFarmerRequest() UpdateFarmerRequest {
	return UpdateFarmerRequest{
		BaseRequest: NewBaseRequest(),
		Profile:     FarmerProfileData{},
	}
}

// NewDeleteFarmerRequest creates a new delete farmer request
func NewDeleteFarmerRequest() DeleteFarmerRequest {
	return DeleteFarmerRequest{
		BaseRequest: NewBaseRequest(),
	}
}

// NewGetFarmerRequest creates a new get farmer request
func NewGetFarmerRequest() GetFarmerRequest {
	return GetFarmerRequest{
		BaseRequest: NewBaseRequest(),
	}
}

// NewListFarmersRequest creates a new list farmers request
func NewListFarmersRequest() ListFarmersRequest {
	return ListFarmersRequest{
		FilterRequest: FilterRequest{
			PaginationRequest: NewPaginationRequest(1, 10),
		},
	}
}
