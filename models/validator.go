package models

import "github.com/go-playground/validator/v10"

var (
	// Validator is shared across handlers / services
	Validator = validator.New()
)
