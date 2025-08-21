package responses

import (
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
)

// FarmerResponse represents a single farmer response
type FarmerResponse struct {
	responses.BaseResponse
	Data *FarmerProfileData `json:"data"`
}

// FarmerListResponse represents a list of farmers response
type FarmerListResponse struct {
	responses.ListResponse
	Data []*FarmerProfileData `json:"data"`
}

// FarmerLinkResponse represents a farmer link response
type FarmerLinkResponse struct {
	responses.BaseResponse
	Data *FarmerLinkData `json:"data"`
}

// FarmerLinkListResponse represents a list of farmer links response
type FarmerLinkListResponse struct {
	responses.ListResponse
	Data []*FarmerLinkData `json:"data"`
}

// FarmerProfileResponse represents a farmer profile response
type FarmerProfileResponse struct {
	responses.BaseResponse
	Data *FarmerProfileData `json:"data"`
}

// FarmerProfileData represents the profile data in responses
type FarmerProfileData struct {
	AAAUserID        string            `json:"aaa_user_id"`
	AAAOrgID         string            `json:"aaa_org_id"`
	KisanSathiUserID *string           `json:"kisan_sathi_user_id,omitempty"`
	FirstName        string            `json:"first_name,omitempty"`
	LastName         string            `json:"last_name,omitempty"`
	PhoneNumber      string            `json:"phone_number,omitempty"`
	Email            string            `json:"email,omitempty"`
	DateOfBirth      string            `json:"date_of_birth,omitempty"`
	Gender           string            `json:"gender,omitempty"`
	Address          AddressData       `json:"address,omitempty"`
	Preferences      map[string]string `json:"preferences,omitempty"`
	Metadata         map[string]string `json:"metadata,omitempty"`
	Farms            []*FarmData       `json:"farms,omitempty"`
	CreatedAt        string            `json:"created_at,omitempty"`
	UpdatedAt        string            `json:"updated_at,omitempty"`
}

// AddressData represents address information in responses
type AddressData struct {
	StreetAddress string `json:"street_address,omitempty"`
	City          string `json:"city,omitempty"`
	State         string `json:"state,omitempty"`
	PostalCode    string `json:"postal_code,omitempty"`
	Country       string `json:"country,omitempty"`
	Coordinates   string `json:"coordinates,omitempty"`
}

// FarmerLinkData represents farmer link data in responses
type FarmerLinkData struct {
	ID               string  `json:"id"`
	AAAUserID        string  `json:"aaa_user_id"`
	AAAOrgID         string  `json:"aaa_org_id"`
	KisanSathiUserID *string `json:"kisan_sathi_user_id,omitempty"`
	Status           string  `json:"status"`
	CreatedAt        string  `json:"created_at,omitempty"`
	UpdatedAt        string  `json:"updated_at,omitempty"`
}

// FarmData represents farm data in responses
type FarmData struct {
	ID              string            `json:"id"`
	AAAFarmerUserID string            `json:"aaa_farmer_user_id"`
	AAAOrgID        string            `json:"aaa_org_id"`
	AreaHa          float64           `json:"area_ha"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	CreatedAt       string            `json:"created_at,omitempty"`
	UpdatedAt       string            `json:"updated_at,omitempty"`
}

// NewFarmerResponse creates a new farmer response
func NewFarmerResponse(farmer *FarmerProfileData, message string) FarmerResponse {
	return FarmerResponse{
		BaseResponse: responses.NewBaseResponse(),
		Data:         farmer,
	}
}

// NewFarmerListResponse creates a new farmer list response
func NewFarmerListResponse(farmers []*FarmerProfileData, page, pageSize int, totalCount int64) FarmerListResponse {
	// Convert to interface slice for NewListResponse
	var responseData []interface{}
	for _, f := range farmers {
		responseData = append(responseData, f)
	}

	return FarmerListResponse{
		ListResponse: responses.NewListResponse(responseData, page, pageSize, totalCount),
		Data:         farmers,
	}
}

// NewFarmerLinkResponse creates a new farmer link response
func NewFarmerLinkResponse(link *FarmerLinkData, message string) FarmerLinkResponse {
	return FarmerLinkResponse{
		BaseResponse: responses.NewBaseResponse(),
		Data:         link,
	}
}

// NewFarmerLinkListResponse creates a new farmer link list response
func NewFarmerLinkListResponse(links []*FarmerLinkData, page, pageSize int, totalCount int64) FarmerLinkListResponse {
	// Convert to interface slice for NewListResponse
	var responseData []interface{}
	for _, l := range links {
		responseData = append(responseData, l)
	}

	return FarmerLinkListResponse{
		ListResponse: responses.NewListResponse(responseData, page, pageSize, totalCount),
		Data:         links,
	}
}

// NewFarmerProfileResponse creates a new farmer profile response
func NewFarmerProfileResponse(profile *FarmerProfileData, message string) FarmerProfileResponse {
	return FarmerProfileResponse{
		BaseResponse: responses.NewBaseResponse(),
		Data:         profile,
	}
}

// SetRequestID sets the request ID for tracking
func (r *FarmerResponse) SetRequestID(requestID string) {
	r.BaseResponse.SetResponseID(requestID)
}

// SetRequestID sets the request ID for tracking
func (r *FarmerListResponse) SetRequestID(requestID string) {
	r.ListResponse.PaginationResponse.BaseResponse.SetResponseID(requestID)
}

// SetRequestID sets the request ID for tracking
func (r *FarmerLinkResponse) SetRequestID(requestID string) {
	r.BaseResponse.SetResponseID(requestID)
}

// SetRequestID sets the request ID for tracking
func (r *FarmerLinkListResponse) SetRequestID(requestID string) {
	r.ListResponse.PaginationResponse.BaseResponse.SetResponseID(requestID)
}

// SetRequestID sets the request ID for tracking
func (r *FarmerProfileResponse) SetRequestID(requestID string) {
	r.BaseResponse.SetResponseID(requestID)
}
