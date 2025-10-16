package responses

import (
	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// FarmerResponse represents a single farmer response
type FarmerResponse struct {
	*base.BaseResponse `json:",inline"`
	Data               *FarmerProfileData `json:"data"`
}

// FarmerListResponse represents a list of farmers response
type FarmerListResponse struct {
	*base.PaginatedResponse `json:",inline"`
	Data                    []*FarmerProfileData `json:"data"`
}

// FarmerLinkResponse represents a farmer link response
type FarmerLinkResponse struct {
	*base.BaseResponse `json:",inline"`
	Data               *FarmerLinkData `json:"data"`
}

// FarmerLinkListResponse represents a list of farmer links response
type FarmerLinkListResponse struct {
	*base.PaginatedResponse `json:",inline"`
	Data                    []*FarmerLinkData `json:"data"`
}

// FarmerProfileResponse represents a farmer profile response
type FarmerProfileResponse struct {
	*base.BaseResponse `json:",inline"`
	Data               *FarmerProfileData `json:"data"`
}

// FarmerProfileData represents the profile data in responses
type FarmerProfileData struct {
	ID               string                 `json:"id" example:"FMRR0000000001"` // Farmer ID (primary key)
	AAAUserID        string                 `json:"aaa_user_id" example:"USER00000001"`
	AAAOrgID         string                 `json:"aaa_org_id" example:"ORGN00000001"`
	KisanSathiUserID *string                `json:"kisan_sathi_user_id,omitempty" example:"USER00000002"`
	FirstName        string                 `json:"first_name,omitempty" example:"Ramesh"`
	LastName         string                 `json:"last_name,omitempty" example:"Kumar"`
	PhoneNumber      string                 `json:"phone_number,omitempty" example:"9876543210"`
	Email            string                 `json:"email,omitempty" example:"ramesh.kumar@example.com"`
	DateOfBirth      string                 `json:"date_of_birth,omitempty" example:"1980-05-15"`
	Gender           string                 `json:"gender,omitempty" example:"male"`
	SocialCategory   string                 `json:"social_category,omitempty" example:"OBC"`
	AreaType         string                 `json:"area_type,omitempty" example:"Rural"`
	TotalAcreageHa   float64                `json:"total_acreage_ha" example:"15.75"`
	Address          AddressData            `json:"address,omitempty"`
	FPOLinkages      []*FarmerLinkData      `json:"fpo_linkages,omitempty"`
	Preferences      map[string]interface{} `json:"preferences,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	Farms            []*FarmData            `json:"farms,omitempty"`
	CreatedAt        string                 `json:"created_at,omitempty" example:"2024-01-15T10:30:00Z"`
	UpdatedAt        string                 `json:"updated_at,omitempty" example:"2024-01-20T15:45:00Z"`
}

// AddressData represents address information in responses
type AddressData struct {
	StreetAddress string `json:"street_address,omitempty" example:"Village Rampur, Post Khandwa"`
	City          string `json:"city,omitempty" example:"Indore"`
	State         string `json:"state,omitempty" example:"Madhya Pradesh"`
	PostalCode    string `json:"postal_code,omitempty" example:"452001"`
	Country       string `json:"country,omitempty" example:"India"`
	Coordinates   string `json:"coordinates,omitempty" example:"POINT(75.8577 22.7196)"`
}

// FarmerLinkData represents farmer link data in responses
type FarmerLinkData struct {
	ID               string  `json:"id" example:"FMLK0000000001"`
	AAAUserID        string  `json:"aaa_user_id" example:"USER00000001"`
	AAAOrgID         string  `json:"aaa_org_id" example:"ORGN00000001"`
	KisanSathiUserID *string `json:"kisan_sathi_user_id,omitempty" example:"USER00000002"`
	Status           string  `json:"status" example:"ACTIVE"`
	CreatedAt        string  `json:"created_at,omitempty" example:"2024-01-15T10:30:00Z"`
	UpdatedAt        string  `json:"updated_at,omitempty" example:"2024-01-20T15:45:00Z"`
}

// FarmData is defined in farm_responses.go

// NewFarmerResponse creates a new farmer response
func NewFarmerResponse(farmer *FarmerProfileData, message string) FarmerResponse {
	return FarmerResponse{
		BaseResponse: base.NewSuccessResponse(message, farmer),
		Data:         farmer,
	}
}

// NewFarmerListResponse creates a new farmer list response
func NewFarmerListResponse(farmers []*FarmerProfileData, page, pageSize int, totalCount int64) FarmerListResponse {
	// Convert to interface slice for pagination
	var data []interface{}
	for _, f := range farmers {
		data = append(data, f)
	}

	paginationInfo := base.NewPaginationInfo(page, pageSize, int(totalCount))
	return FarmerListResponse{
		PaginatedResponse: base.NewPaginatedResponse("Farmers retrieved successfully", data, paginationInfo),
		Data:              farmers,
	}
}

// NewFarmerLinkResponse creates a new farmer link response
func NewFarmerLinkResponse(link *FarmerLinkData, message string) FarmerLinkResponse {
	return FarmerLinkResponse{
		BaseResponse: base.NewSuccessResponse(message, link),
		Data:         link,
	}
}

// NewFarmerLinkListResponse creates a new farmer link list response
func NewFarmerLinkListResponse(links []*FarmerLinkData, page, pageSize int, totalCount int64) FarmerLinkListResponse {
	// Convert to interface slice for pagination
	var data []interface{}
	for _, l := range links {
		data = append(data, l)
	}

	paginationInfo := base.NewPaginationInfo(page, pageSize, int(totalCount))
	return FarmerLinkListResponse{
		PaginatedResponse: base.NewPaginatedResponse("Farmer links retrieved successfully", data, paginationInfo),
		Data:              links,
	}
}

// NewFarmerProfileResponse creates a new farmer profile response
func NewFarmerProfileResponse(profile *FarmerProfileData, message string) FarmerProfileResponse {
	return FarmerProfileResponse{
		BaseResponse: base.NewSuccessResponse(message, profile),
		Data:         profile,
	}
}

// SetRequestID sets the request ID for tracking
func (r *FarmerResponse) SetRequestID(requestID string) {
	r.BaseResponse.RequestID = requestID
}

// SetRequestID sets the request ID for tracking
func (r *FarmerListResponse) SetRequestID(requestID string) {
	r.PaginatedResponse.RequestID = requestID
}

// SetRequestID sets the request ID for tracking
func (r *FarmerLinkResponse) SetRequestID(requestID string) {
	r.BaseResponse.RequestID = requestID
}

// SetRequestID sets the request ID for tracking
func (r *FarmerLinkListResponse) SetRequestID(requestID string) {
	r.PaginatedResponse.RequestID = requestID
}

// SetRequestID sets the request ID for tracking
func (r *FarmerProfileResponse) SetRequestID(requestID string) {
	r.BaseResponse.RequestID = requestID
}
