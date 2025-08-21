package requests

import (
	"time"
)

// BaseRequest provides common fields for all API requests
type BaseRequest struct {
	RequestID   string            `json:"request_id,omitempty"`
	Timestamp   time.Time         `json:"timestamp"`
	UserID      string            `json:"user_id,omitempty"`
	OrgID       string            `json:"org_id,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	RequestType string            `json:"request_type,omitempty"`
}

// PaginationRequest provides pagination parameters for list requests
type PaginationRequest struct {
	BaseRequest
	Page     int `json:"page" validate:"min=1"`
	PageSize int `json:"page_size" validate:"min=1,max=100"`
}

// FilterRequest provides filtering parameters for search requests
type FilterRequest struct {
	PaginationRequest
	Filters map[string]interface{} `json:"filters,omitempty"`
	SortBy  string                 `json:"sort_by,omitempty"`
	SortDir string                 `json:"sort_dir,omitempty" validate:"oneof=asc desc"`
}

// NewBaseRequest creates a new base request with default values
func NewBaseRequest() BaseRequest {
	return BaseRequest{
		Timestamp:   time.Now(),
		Metadata:    make(map[string]string),
		RequestType: "unknown",
	}
}

// NewPaginationRequest creates a new pagination request
func NewPaginationRequest(page, pageSize int) PaginationRequest {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	return PaginationRequest{
		BaseRequest: NewBaseRequest(),
		Page:        page,
		PageSize:    pageSize,
	}
}

// SetRequestID sets the request ID for tracking
func (r *BaseRequest) SetRequestID(requestID string) {
	r.RequestID = requestID
}

// SetUserContext sets the user context for the request
func (r *BaseRequest) SetUserContext(userID, orgID string) {
	r.UserID = userID
	r.OrgID = orgID
}

// AddMetadata adds metadata to the request
func (r *BaseRequest) AddMetadata(key, value string) {
	if r.Metadata == nil {
		r.Metadata = make(map[string]string)
	}
	r.Metadata[key] = value
}

// SetRequestType sets the type of request
func (r *BaseRequest) SetRequestType(requestType string) {
	r.RequestType = requestType
}
