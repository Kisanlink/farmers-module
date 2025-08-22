package docs

import (
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// Swagger type definitions to avoid generic types

// CreateFarmerRequest represents the request structure for creating a farmer
// @Description Request structure for creating a new farmer
type CreateFarmerRequest struct {
	RequestID        string                  `json:"request_id,omitempty" example:"req_123" description:"Unique request identifier for tracking"`
	Timestamp        time.Time               `json:"timestamp" example:"2024-01-01T00:00:00Z" description:"Request timestamp"`
	UserID           string                  `json:"user_id,omitempty" example:"user_123" description:"User who initiated the request"`
	OrgID            string                  `json:"org_id,omitempty" example:"org_123" description:"Organization ID"`
	Metadata         map[string]string       `json:"metadata,omitempty" description:"Additional metadata for the request"`
	RequestType      string                  `json:"request_type,omitempty" example:"create_farmer" description:"Type of request"`
	AAAUserID        string                  `json:"aaa_user_id" validate:"required" example:"aaa_user_123" description:"AAA service user ID (required)"`
	AAAOrgID         string                  `json:"aaa_org_id" validate:"required" example:"aaa_org_123" description:"AAA service organization ID (required)"`
	KisanSathiUserID *string                 `json:"kisan_sathi_user_id,omitempty" example:"ks_user_123" description:"Optional KisanSathi user ID"`
	Profile          CreateFarmerProfileData `json:"profile,omitempty" description:"Farmer profile information"`
}

// UpdateFarmerRequest represents the request structure for updating a farmer
// @Description Request structure for updating an existing farmer
type UpdateFarmerRequest struct {
	RequestID        string                  `json:"request_id,omitempty" example:"req_123" description:"Unique request identifier for tracking"`
	Timestamp        time.Time               `json:"timestamp" example:"2024-01-01T00:00:00Z" description:"Request timestamp"`
	UserID           string                  `json:"user_id,omitempty" example:"user_123" description:"User who initiated the request"`
	OrgID            string                  `json:"org_id,omitempty" example:"org_123" description:"Organization ID"`
	Metadata         map[string]string       `json:"metadata,omitempty" description:"Additional metadata for the request"`
	RequestType      string                  `json:"request_type,omitempty" example:"update_farmer" description:"Type of request"`
	AAAUserID        string                  `json:"aaa_user_id" validate:"required" example:"aaa_user_123" description:"AAA service user ID (required)"`
	AAAOrgID         string                  `json:"aaa_org_id" validate:"required" example:"aaa_org_123" description:"AAA service organization ID (required)"`
	KisanSathiUserID *string                 `json:"kisan_sathi_user_id,omitempty" example:"ks_user_123" description:"Optional KisanSathi user ID"`
	Profile          CreateFarmerProfileData `json:"profile,omitempty" description:"Updated farmer profile information"`
}

// DeleteFarmerRequest represents the request structure for deleting a farmer
// @Description Request structure for deleting a farmer
type DeleteFarmerRequest struct {
	RequestID   string            `json:"request_id,omitempty" example:"req_123"`
	Timestamp   time.Time         `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	UserID      string            `json:"user_id,omitempty" example:"user_123"`
	OrgID       string            `json:"org_id,omitempty" example:"org_123"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	RequestType string            `json:"request_type,omitempty" example:"delete_farmer"`
	AAAUserID   string            `json:"aaa_user_id" validate:"required" example:"aaa_user_123"`
	AAAOrgID    string            `json:"aaa_org_id" validate:"required" example:"aaa_org_123"`
}

// GetFarmerRequest represents the request structure for getting a farmer
// @Description Request structure for retrieving a farmer
type GetFarmerRequest struct {
	RequestID   string            `json:"request_id,omitempty" example:"req_123"`
	Timestamp   time.Time         `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	UserID      string            `json:"user_id,omitempty" example:"user_123"`
	OrgID       string            `json:"org_id,omitempty" example:"org_123"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	RequestType string            `json:"request_type,omitempty" example:"get_farmer"`
	AAAUserID   string            `json:"aaa_user_id" validate:"required" example:"aaa_user_123"`
	AAAOrgID    string            `json:"aaa_org_id" validate:"required" example:"aaa_org_123"`
}

// ListFarmersRequest represents the request structure for listing farmers
// @Description Request structure for listing farmers with filtering and pagination
type ListFarmersRequest struct {
	RequestID        string                 `json:"request_id,omitempty" example:"req_123"`
	Timestamp        time.Time              `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	UserID           string                 `json:"user_id,omitempty" example:"user_123"`
	OrgID            string                 `json:"org_id,omitempty" example:"org_123"`
	Metadata         map[string]string      `json:"metadata,omitempty"`
	RequestType      string                 `json:"request_type,omitempty" example:"list_farmers"`
	Page             int                    `json:"page" validate:"min=1" example:"1"`
	PageSize         int                    `json:"page_size" validate:"min=1,max=100" example:"10"`
	Filters          map[string]interface{} `json:"filters,omitempty"`
	SortBy           string                 `json:"sort_by,omitempty" example:"created_at"`
	SortDir          string                 `json:"sort_dir,omitempty" validate:"oneof=asc desc" example:"desc"`
	AAAOrgID         string                 `json:"aaa_org_id,omitempty" example:"aaa_org_123"`
	KisanSathiUserID string                 `json:"kisan_sathi_user_id,omitempty" example:"ks_user_123"`
}

// CreateFarmerProfileData represents the profile data structure for creating farmers
// @Description Profile data structure for farmers (create request)
type CreateFarmerProfileData struct {
	FirstName   string            `json:"first_name,omitempty" example:"John" description:"Farmer's first name"`
	LastName    string            `json:"last_name,omitempty" example:"Doe" description:"Farmer's last name"`
	PhoneNumber string            `json:"phone_number,omitempty" example:"+91-9876543210" description:"Farmer's phone number"`
	Email       string            `json:"email,omitempty" example:"john.doe@example.com" description:"Farmer's email address"`
	DateOfBirth string            `json:"date_of_birth,omitempty" example:"1990-01-01" description:"Farmer's date of birth (YYYY-MM-DD format)"`
	Gender      string            `json:"gender,omitempty" example:"male" description:"Farmer's gender" enums:"male,female,other"`
	Address     AddressData       `json:"address,omitempty" description:"Farmer's address information"`
	Preferences map[string]string `json:"preferences,omitempty" description:"Farmer's preferences as key-value pairs"`
	Metadata    map[string]string `json:"metadata,omitempty" description:"Additional metadata as key-value pairs"`
}

// FarmerProfileData represents the profile data structure for responses
// @Description Profile data structure for farmers (response)
type FarmerProfileData struct {
	AAAUserID        string            `json:"aaa_user_id" example:"aaa_user_123" description:"AAA service user ID"`
	AAAOrgID         string            `json:"aaa_org_id" example:"aaa_org_123" description:"AAA service organization ID"`
	KisanSathiUserID *string           `json:"kisan_sathi_user_id,omitempty" example:"ks_user_123" description:"KisanSathi user ID if assigned"`
	FirstName        string            `json:"first_name,omitempty" example:"John" description:"Farmer's first name"`
	LastName         string            `json:"last_name,omitempty" example:"Doe" description:"Farmer's last name"`
	PhoneNumber      string            `json:"phone_number,omitempty" example:"+91-9876543210" description:"Farmer's phone number"`
	Email            string            `json:"email,omitempty" example:"john.doe@example.com" description:"Farmer's email address"`
	DateOfBirth      string            `json:"date_of_birth,omitempty" example:"1990-01-01" description:"Farmer's date of birth"`
	Gender           string            `json:"gender,omitempty" example:"male" description:"Farmer's gender"`
	Address          AddressData       `json:"address,omitempty" description:"Farmer's address information"`
	Preferences      map[string]string `json:"preferences,omitempty" description:"Farmer's preferences"`
	Metadata         map[string]string `json:"metadata,omitempty" description:"Additional metadata"`
	CreatedAt        string            `json:"created_at,omitempty" example:"2024-01-01T00:00:00Z" description:"Record creation timestamp"`
	UpdatedAt        string            `json:"updated_at,omitempty" example:"2024-01-01T00:00:00Z" description:"Record last update timestamp"`
}

// AddressData represents address information structure
// @Description Address data structure
type AddressData struct {
	StreetAddress string `json:"street_address,omitempty" example:"123 Main Street" description:"Street address line"`
	City          string `json:"city,omitempty" example:"Mumbai" description:"City name"`
	State         string `json:"state,omitempty" example:"Maharashtra" description:"State or province"`
	PostalCode    string `json:"postal_code,omitempty" example:"400001" description:"Postal or ZIP code"`
	Country       string `json:"country,omitempty" example:"India" description:"Country name"`
	Coordinates   string `json:"coordinates,omitempty" example:"POINT(72.8777 19.0760)" description:"Geographic coordinates in WKT format"`
}

// FarmerResponse represents the response structure for a single farmer
// @Description Response structure for a single farmer
type FarmerResponse struct {
	RequestID string             `json:"request_id,omitempty" example:"req_123" description:"Request ID for tracking"`
	Success   bool               `json:"success" example:"true" description:"Indicates if the operation was successful"`
	Message   string             `json:"message" example:"Farmer created successfully" description:"Human-readable response message"`
	Data      *FarmerProfileData `json:"data" description:"Farmer profile data"`
	Timestamp time.Time          `json:"timestamp" example:"2024-01-01T00:00:00Z" description:"Response timestamp"`
}

// FarmerListResponse represents the response structure for a list of farmers
// @Description Response structure for a list of farmers
type FarmerListResponse struct {
	RequestID  string               `json:"request_id,omitempty" example:"req_123" description:"Request ID for tracking"`
	Success    bool                 `json:"success" example:"true" description:"Indicates if the operation was successful"`
	Message    string               `json:"message" example:"Farmers retrieved successfully" description:"Human-readable response message"`
	Data       []*FarmerProfileData `json:"data" description:"List of farmer profiles"`
	Timestamp  time.Time            `json:"timestamp" example:"2024-01-01T00:00:00Z" description:"Response timestamp"`
	Pagination PaginationInfo       `json:"pagination" description:"Pagination information"`
}

// BaseResponse represents the base response structure
// @Description Base response structure for all API responses
type BaseResponse struct {
	RequestID string      `json:"request_id,omitempty" example:"req_123" description:"Request ID for tracking"`
	Success   bool        `json:"success" example:"true" description:"Indicates if the operation was successful"`
	Message   string      `json:"message" example:"Operation completed successfully" description:"Human-readable response message"`
	Data      interface{} `json:"data,omitempty" description:"Response data payload"`
	Timestamp time.Time   `json:"timestamp" example:"2024-01-01T00:00:00Z" description:"Response timestamp"`
}

// PaginationInfo represents pagination information
// @Description Pagination information structure
type PaginationInfo struct {
	Page       int   `json:"page" example:"1" description:"Current page number"`
	PageSize   int   `json:"page_size" example:"10" description:"Number of items per page"`
	TotalCount int64 `json:"total_count" example:"100" description:"Total number of items"`
	TotalPages int   `json:"total_pages" example:"10" description:"Total number of pages"`
}

// ErrorResponse represents error response structure
// @Description Error response structure
type ErrorResponse struct {
	RequestID string          `json:"request_id,omitempty" example:"req_123" description:"Request ID for tracking"`
	Success   bool            `json:"success" example:"false" description:"Always false for error responses"`
	Message   string          `json:"message" example:"Operation failed" description:"Human-readable error message"`
	Error     *base.BaseError `json:"error" description:"Detailed error information"`
	Timestamp time.Time       `json:"timestamp" example:"2024-01-01T00:00:00Z" description:"Response timestamp"`
}

// Crop Cycle Types

// StartCycleRequest represents a request to start a crop cycle
// @Description Request structure for starting a new crop cycle
type StartCycleRequest struct {
	FarmID           string            `json:"farm_id" example:"farm_123"`
	CropType         string            `json:"crop_type" example:"wheat"`
	Season           string            `json:"season" example:"rabi"`
	PlannedStartDate time.Time         `json:"planned_start_date" example:"2024-01-01T00:00:00Z"`
	PlannedEndDate   time.Time         `json:"planned_end_date" example:"2024-06-01T00:00:00Z"`
	Metadata         map[string]string `json:"metadata,omitempty"`
}

// UpdateCycleRequest represents a request to update a crop cycle
// @Description Request structure for updating an existing crop cycle
type UpdateCycleRequest struct {
	ID               string            `json:"id" example:"cc_123"`
	CropType         *string           `json:"crop_type,omitempty" example:"wheat"`
	Season           *string           `json:"season,omitempty" example:"rabi"`
	PlannedStartDate *time.Time        `json:"planned_start_date,omitempty" example:"2024-01-01T00:00:00Z"`
	PlannedEndDate   *time.Time        `json:"planned_end_date,omitempty" example:"2024-06-01T00:00:00Z"`
	Metadata         map[string]string `json:"metadata,omitempty"`
}

// EndCycleRequest represents a request to end a crop cycle
// @Description Request structure for ending a crop cycle
type EndCycleRequest struct {
	ID            string    `json:"id" example:"cc_123"`
	ActualEndDate time.Time `json:"actual_end_date" example:"2024-06-01T00:00:00Z"`
}

// ListCyclesRequest represents a request to list crop cycles
// @Description Request structure for listing crop cycles with filtering and pagination
type ListCyclesRequest struct {
	FarmID   string `json:"farm_id,omitempty" example:"farm_123"`
	CropType string `json:"crop_type,omitempty" example:"wheat"`
	Season   string `json:"season,omitempty" example:"rabi"`
	Status   string `json:"status,omitempty" example:"ACTIVE"`
	Page     int    `json:"page" example:"1"`
	PageSize int    `json:"page_size" example:"10"`
}

// CropCycleResponse represents the response structure for a crop cycle
// @Description Response structure for a single crop cycle
type CropCycleResponse struct {
	ID               string            `json:"id" example:"cc_123"`
	FarmID           string            `json:"farm_id" example:"farm_123"`
	CropType         string            `json:"crop_type" example:"wheat"`
	Season           string            `json:"season" example:"rabi"`
	Status           string            `json:"status" example:"ACTIVE"`
	PlannedStartDate time.Time         `json:"planned_start_date" example:"2024-01-01T00:00:00Z"`
	ActualStartDate  *time.Time        `json:"actual_start_date,omitempty" example:"2024-01-01T00:00:00Z"`
	PlannedEndDate   time.Time         `json:"planned_end_date" example:"2024-06-01T00:00:00Z"`
	ActualEndDate    *time.Time        `json:"actual_end_date,omitempty" example:"2024-06-01T00:00:00Z"`
	Metadata         map[string]string `json:"metadata,omitempty"`
	CreatedBy        string            `json:"created_by,omitempty" example:"user_123"`
	CreatedAt        time.Time         `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt        time.Time         `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// CropCycleListResponse represents the response structure for a list of crop cycles
// @Description Response structure for a list of crop cycles
type CropCycleListResponse struct {
	RequestID  string               `json:"request_id,omitempty" example:"req_123"`
	Success    bool                 `json:"success" example:"true"`
	Message    string               `json:"message" example:"Crop cycles retrieved successfully"`
	Data       []*CropCycleResponse `json:"data"`
	Timestamp  time.Time            `json:"timestamp" example:"2024-01-01T00:00:00Z"`
	Pagination PaginationInfo       `json:"pagination"`
}
