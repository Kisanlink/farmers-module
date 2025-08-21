package responses

import (
	"time"
)

// BaseResponse provides common fields for all API responses
type BaseResponse struct {
	RequestID   string            `json:"request_id,omitempty"`
	Timestamp   time.Time         `json:"timestamp"`
	Status      string            `json:"status"`
	Message     string            `json:"message,omitempty"`
	ErrorCode   string            `json:"error_code,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	RequestType string            `json:"request_type,omitempty"`
}

// PaginationResponse provides pagination information for list responses
type PaginationResponse struct {
	BaseResponse
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
	TotalCount int64 `json:"total_count"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// SuccessResponse provides a standard success response structure
type SuccessResponse struct {
	BaseResponse
	Data interface{} `json:"data"`
}

// ErrorResponse provides a standard error response structure
type ErrorResponse struct {
	BaseResponse
	Details    []string          `json:"details,omitempty"`
	Validation map[string]string `json:"validation,omitempty"`
}

// ListResponse provides a standard list response structure
type ListResponse struct {
	PaginationResponse
	Data []interface{} `json:"data"`
}

// NewBaseResponse creates a new base response with default values
func NewBaseResponse() BaseResponse {
	return BaseResponse{
		Timestamp:   time.Now(),
		Status:      "success",
		Metadata:    make(map[string]string),
		RequestType: "unknown",
	}
}

// NewSuccessResponse creates a new success response
func NewSuccessResponse(data interface{}, message string) SuccessResponse {
	return SuccessResponse{
		BaseResponse: BaseResponse{
			Timestamp:   time.Now(),
			Status:      "success",
			Message:     message,
			Metadata:    make(map[string]string),
			RequestType: "success",
		},
		Data: data,
	}
}

// NewErrorResponse creates a new error response
func NewErrorResponse(message, errorCode string, details ...string) ErrorResponse {
	return ErrorResponse{
		BaseResponse: BaseResponse{
			Timestamp:   time.Now(),
			Status:      "error",
			Message:     message,
			ErrorCode:   errorCode,
			Metadata:    make(map[string]string),
			RequestType: "error",
		},
		Details: details,
	}
}

// NewListResponse creates a new list response with pagination
func NewListResponse(data []interface{}, page, pageSize int, totalCount int64) ListResponse {
	totalPages := int((totalCount + int64(pageSize) - 1) / int64(pageSize))

	return ListResponse{
		PaginationResponse: PaginationResponse{
			BaseResponse: BaseResponse{
				Timestamp:   time.Now(),
				Status:      "success",
				Message:     "Data retrieved successfully",
				Metadata:    make(map[string]string),
				RequestType: "list",
			},
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
			TotalCount: totalCount,
			HasNext:    page < totalPages,
			HasPrev:    page > 1,
		},
		Data: data,
	}
}

// SetResponseID sets the response ID to match the request
func (r *BaseResponse) SetResponseID(requestID string) {
	r.RequestID = requestID
}

// SetStatus sets the response status
func (r *BaseResponse) SetStatus(status string) {
	r.Status = status
}

// SetMessage sets the response message
func (r *BaseResponse) SetMessage(message string) {
	r.Message = message
}

// SetError sets the error information
func (r *BaseResponse) SetError(errorCode, message string) {
	r.Status = "error"
	r.ErrorCode = errorCode
	r.Message = message
}

// AddMetadata adds metadata to the response
func (r *BaseResponse) AddMetadata(key, value string) {
	if r.Metadata == nil {
		r.Metadata = make(map[string]string)
	}
	r.Metadata[key] = value
}

// SetRequestType sets the type of response
func (r *BaseResponse) SetRequestType(requestType string) {
	r.RequestType = requestType
}
