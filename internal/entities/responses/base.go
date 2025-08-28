package responses

import (
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// Re-export the standardized response types from kisanlink-db
type BaseResponse = base.BaseResponse
type PaginatedResponse = base.PaginatedResponse
type PaginationInfo = base.PaginationInfo
type BaseError = base.BaseError
type ErrorInterface = base.ErrorInterface
type ResponseInterface = base.ResponseInterface

// Re-export the standardized request types
type BaseRequest = base.BaseRequest
type CreateRequest = base.CreateRequest
type UpdateRequest = base.UpdateRequest
type DeleteRequest = base.DeleteRequest
type ListRequest = base.ListRequest
type GetByIDRequest = base.GetByIDRequest

// Re-export the standardized error constructors
var (
	NewValidationError         = base.NewValidationError
	NewNotFoundError           = base.NewNotFoundError
	NewUnauthorizedError       = base.NewUnauthorizedError
	NewForbiddenError          = base.NewForbiddenError
	NewConflictError           = base.NewConflictError
	NewInternalServerError     = base.NewInternalServerError
	NewServiceUnavailableError = base.NewServiceUnavailableError
	NewTooManyRequestsError    = base.NewTooManyRequestsError
)

// Re-export the standardized response constructors
var (
	NewSuccessResponse   = base.NewSuccessResponse
	NewErrorResponse     = base.NewErrorResponse
	NewPaginatedResponse = base.NewPaginatedResponse
	NewPaginationInfo    = base.NewPaginationInfo
)

// ErrorResponse represents a structured error response
type ErrorResponse struct {
	Error         string            `json:"error"`
	Message       string            `json:"message"`
	Code          string            `json:"code,omitempty"`
	CorrelationID string            `json:"correlation_id,omitempty"`
	Details       map[string]string `json:"details,omitempty"`
	Timestamp     time.Time         `json:"timestamp"`
}
