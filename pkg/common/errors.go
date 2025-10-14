package common

import (
	"errors"
	"fmt"
	"time"
)

var (
	// Farm errors
	ErrInvalidFarmData     = errors.New("invalid farm data")
	ErrInvalidFarmGeometry = errors.New("invalid farm geometry")

	// Crop cycle errors
	ErrInvalidCropCycleData = errors.New("invalid crop cycle data")
	ErrInvalidAreaValue     = errors.New("invalid area value")
	ErrStatusNotModifiable  = errors.New("crop cycle status does not allow modification")

	// Farm activity errors
	ErrInvalidFarmActivityData = errors.New("invalid farm activity data")
	ErrInvalidStageForCrop     = errors.New("crop stage does not belong to the crop")

	// General errors
	ErrNotFound      = errors.New("resource not found")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
	ErrInvalidInput  = errors.New("invalid input")
	ErrInternal      = errors.New("internal server error")
	ErrAlreadyExists = errors.New("resource already exists")
)

// AreaExceededError represents an error when area allocation exceeds farm capacity
type AreaExceededError struct {
	FarmID        string
	FarmArea      float64
	RequestedArea float64
	AvailableArea float64
	AllocatedArea float64
}

func (e *AreaExceededError) Error() string {
	return fmt.Sprintf("requested area %.4f ha exceeds available area %.4f ha for farm %s (total: %.4f ha, allocated: %.4f ha)",
		e.RequestedArea, e.AvailableArea, e.FarmID, e.FarmArea, e.AllocatedArea)
}

// ConcurrentModificationError represents an error when a resource was modified by another request
type ConcurrentModificationError struct {
	ResourceID   string
	ResourceType string
	Operation    string
}

func (e *ConcurrentModificationError) Error() string {
	return fmt.Sprintf("%s %s was modified by another request during %s operation",
		e.ResourceType, e.ResourceID, e.Operation)
}

// ErrorResponse represents a structured error response
type ErrorResponse struct {
	Error         string            `json:"error"`
	Message       string            `json:"message"`
	Code          string            `json:"code,omitempty"`
	CorrelationID string            `json:"correlation_id,omitempty"`
	Details       map[string]string `json:"details,omitempty"`
	Timestamp     time.Time         `json:"timestamp"`
}
