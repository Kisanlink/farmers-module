package common

import (
	"errors"
	"time"
)

var (
	// Farm errors
	ErrInvalidFarmData     = errors.New("invalid farm data")
	ErrInvalidFarmGeometry = errors.New("invalid farm geometry")

	// Crop cycle errors
	ErrInvalidCropCycleData = errors.New("invalid crop cycle data")

	// Farm activity errors
	ErrInvalidFarmActivityData = errors.New("invalid farm activity data")

	// General errors
	ErrNotFound       = errors.New("resource not found")
	ErrUnauthorized   = errors.New("unauthorized")
	ErrForbidden      = errors.New("forbidden")
	ErrInvalidInput   = errors.New("invalid input")
	ErrInternal       = errors.New("internal server error")
	ErrAlreadyExists  = errors.New("resource already exists")
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
