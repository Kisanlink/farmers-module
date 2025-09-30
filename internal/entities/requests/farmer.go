package requests

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
	FarmerID         string            `json:"farmer_id,omitempty"`   // Primary key lookup
	AAAUserID        string            `json:"aaa_user_id,omitempty"` // User ID lookup (no org required)
	AAAOrgID         string            `json:"aaa_org_id,omitempty"`  // Optional org filter
	KisanSathiUserID *string           `json:"kisan_sathi_user_id,omitempty"`
	Profile          FarmerProfileData `json:"profile,omitempty"`
}

// DeleteFarmerRequest represents a request to delete a farmer
type DeleteFarmerRequest struct {
	BaseRequest
	FarmerID  string `json:"farmer_id,omitempty"`   // Primary key lookup
	AAAUserID string `json:"aaa_user_id,omitempty"` // User ID lookup (no org required)
	AAAOrgID  string `json:"aaa_org_id,omitempty"`  // Optional org filter
}

// GetFarmerRequest represents a request to retrieve a farmer
type GetFarmerRequest struct {
	BaseRequest
	FarmerID  string `json:"farmer_id,omitempty"`   // Primary key lookup
	AAAUserID string `json:"aaa_user_id,omitempty"` // User ID lookup (no org required)
	AAAOrgID  string `json:"aaa_org_id,omitempty"`  // Optional org filter
}

// ListFarmersRequest represents a request to list farmers with filtering
type ListFarmersRequest struct {
	FilterRequest
	AAAOrgID         string `json:"aaa_org_id,omitempty"`
	KisanSathiUserID string `json:"kisan_sathi_user_id,omitempty"`
	Page             int    `json:"page,omitempty"`
	PageSize         int    `json:"page_size,omitempty"`
}

// FarmerProfileData represents the profile data for a farmer
type FarmerProfileData struct {
	FirstName   string            `json:"first_name,omitempty"`
	LastName    string            `json:"last_name,omitempty"`
	PhoneNumber string            `json:"phone_number,omitempty"`
	Email       string            `json:"email,omitempty"`
	DateOfBirth string            `json:"date_of_birth,omitempty"`
	Gender      string            `json:"gender,omitempty"`
	Address     AddressData       `json:"address,omitempty"`
	Preferences map[string]string `json:"preferences,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// AddressData represents address information
type AddressData struct {
	StreetAddress string `json:"street_address,omitempty"`
	City          string `json:"city,omitempty"`
	State         string `json:"state,omitempty"`
	PostalCode    string `json:"postal_code,omitempty"`
	Country       string `json:"country,omitempty"`
	Coordinates   string `json:"coordinates,omitempty"` // WKT format for PostGIS
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
