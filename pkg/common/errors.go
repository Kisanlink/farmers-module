package common

import "errors"

var (
	// Farm errors
	ErrInvalidFarmData     = errors.New("invalid farm data")
	ErrInvalidFarmGeometry = errors.New("invalid farm geometry")

	// Crop cycle errors
	ErrInvalidCropCycleData = errors.New("invalid crop cycle data")

	// Farm activity errors
	ErrInvalidFarmActivityData = errors.New("invalid farm activity data")

	// General errors
	ErrNotFound     = errors.New("resource not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrInvalidInput = errors.New("invalid input")
	ErrInternal     = errors.New("internal server error")
)
