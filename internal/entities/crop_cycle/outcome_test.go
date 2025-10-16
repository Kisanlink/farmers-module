package crop_cycle

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatePerennialOutcome(t *testing.T) {
	tests := []struct {
		name          string
		outcome       map[string]interface{}
		expectedError string
	}{
		{
			name: "valid perennial outcome with all required fields",
			outcome: map[string]interface{}{
				"age_range_min":  5,
				"age_range_max":  10,
				"yield_per_tree": 50.5,
				"yield_unit":     "kg",
			},
			expectedError: "",
		},
		{
			name: "valid perennial outcome with optional fields",
			outcome: map[string]interface{}{
				"age_range_min":   5,
				"age_range_max":   10,
				"yield_per_tree":  50.5,
				"yield_unit":      "kg",
				"number_of_trees": 100,
				"total_yield":     5050.0,
				"quality_grade":   "A",
				"notes":           "Good harvest",
			},
			expectedError: "",
		},
		{
			name: "valid perennial outcome with float64 values",
			outcome: map[string]interface{}{
				"age_range_min":  5.0,
				"age_range_max":  10.0,
				"yield_per_tree": 50.5,
				"yield_unit":     "kg",
			},
			expectedError: "",
		},
		{
			name:          "empty outcome is valid",
			outcome:       map[string]interface{}{},
			expectedError: "",
		},
		{
			name: "missing age_range_min",
			outcome: map[string]interface{}{
				"age_range_max":  10,
				"yield_per_tree": 50.5,
				"yield_unit":     "kg",
			},
			expectedError: "perennial crop outcome must include age_range_min",
		},
		{
			name: "missing age_range_max",
			outcome: map[string]interface{}{
				"age_range_min":  5,
				"yield_per_tree": 50.5,
				"yield_unit":     "kg",
			},
			expectedError: "perennial crop outcome must include age_range_max",
		},
		{
			name: "missing yield_per_tree",
			outcome: map[string]interface{}{
				"age_range_min": 5,
				"age_range_max": 10,
				"yield_unit":    "kg",
			},
			expectedError: "perennial crop outcome must include yield_per_tree",
		},
		{
			name: "missing yield_unit",
			outcome: map[string]interface{}{
				"age_range_min":  5,
				"age_range_max":  10,
				"yield_per_tree": 50.5,
			},
			expectedError: "perennial crop outcome must include yield_unit",
		},
		{
			name: "negative age_range_min",
			outcome: map[string]interface{}{
				"age_range_min":  -5,
				"age_range_max":  10,
				"yield_per_tree": 50.5,
				"yield_unit":     "kg",
			},
			expectedError: "age_range_min must be non-negative",
		},
		{
			name: "negative age_range_max",
			outcome: map[string]interface{}{
				"age_range_min":  5,
				"age_range_max":  -10,
				"yield_per_tree": 50.5,
				"yield_unit":     "kg",
			},
			expectedError: "age_range_max must be non-negative",
		},
		{
			name: "age_range_max less than age_range_min",
			outcome: map[string]interface{}{
				"age_range_min":  10,
				"age_range_max":  5,
				"yield_per_tree": 50.5,
				"yield_unit":     "kg",
			},
			expectedError: "age_range_max must be greater than or equal to age_range_min",
		},
		{
			name: "zero yield_per_tree",
			outcome: map[string]interface{}{
				"age_range_min":  5,
				"age_range_max":  10,
				"yield_per_tree": 0,
				"yield_unit":     "kg",
			},
			expectedError: "yield_per_tree must be greater than 0",
		},
		{
			name: "negative yield_per_tree",
			outcome: map[string]interface{}{
				"age_range_min":  5,
				"age_range_max":  10,
				"yield_per_tree": -50.5,
				"yield_unit":     "kg",
			},
			expectedError: "yield_per_tree must be greater than 0",
		},
		{
			name: "invalid age_range_min type",
			outcome: map[string]interface{}{
				"age_range_min":  "five",
				"age_range_max":  10,
				"yield_per_tree": 50.5,
				"yield_unit":     "kg",
			},
			expectedError: "age_range_min must be a number",
		},
		{
			name: "invalid yield_unit type",
			outcome: map[string]interface{}{
				"age_range_min":  5,
				"age_range_max":  10,
				"yield_per_tree": 50.5,
				"yield_unit":     123,
			},
			expectedError: "yield_unit must be a string",
		},
		{
			name: "invalid number_of_trees",
			outcome: map[string]interface{}{
				"age_range_min":   5,
				"age_range_max":   10,
				"yield_per_tree":  50.5,
				"yield_unit":      "kg",
				"number_of_trees": -100,
			},
			expectedError: "number_of_trees must be greater than 0",
		},
		{
			name: "invalid total_yield",
			outcome: map[string]interface{}{
				"age_range_min":  5,
				"age_range_max":  10,
				"yield_per_tree": 50.5,
				"yield_unit":     "kg",
				"total_yield":    -5050.0,
			},
			expectedError: "total_yield must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePerennialOutcome(tt.outcome)

			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			}
		})
	}
}

func TestValidateAnnualOutcome(t *testing.T) {
	tests := []struct {
		name          string
		outcome       map[string]interface{}
		expectedError string
	}{
		{
			name: "valid annual outcome with all required fields",
			outcome: map[string]interface{}{
				"yield_per_hectare": 2500.0,
				"yield_unit":        "kg",
			},
			expectedError: "",
		},
		{
			name: "valid annual outcome with optional fields",
			outcome: map[string]interface{}{
				"yield_per_hectare": 2500.0,
				"yield_unit":        "kg",
				"total_yield":       12500.0,
				"quality_grade":     "A",
				"notes":             "Good harvest",
			},
			expectedError: "",
		},
		{
			name: "valid annual outcome with integer values",
			outcome: map[string]interface{}{
				"yield_per_hectare": 2500,
				"yield_unit":        "kg",
			},
			expectedError: "",
		},
		{
			name:          "empty outcome is valid",
			outcome:       map[string]interface{}{},
			expectedError: "",
		},
		{
			name: "missing yield_per_hectare",
			outcome: map[string]interface{}{
				"yield_unit": "kg",
			},
			expectedError: "annual crop outcome must include yield_per_hectare",
		},
		{
			name: "missing yield_unit",
			outcome: map[string]interface{}{
				"yield_per_hectare": 2500.0,
			},
			expectedError: "annual crop outcome must include yield_unit",
		},
		{
			name: "zero yield_per_hectare",
			outcome: map[string]interface{}{
				"yield_per_hectare": 0,
				"yield_unit":        "kg",
			},
			expectedError: "yield_per_hectare must be greater than 0",
		},
		{
			name: "negative yield_per_hectare",
			outcome: map[string]interface{}{
				"yield_per_hectare": -2500.0,
				"yield_unit":        "kg",
			},
			expectedError: "yield_per_hectare must be greater than 0",
		},
		{
			name: "invalid yield_per_hectare type",
			outcome: map[string]interface{}{
				"yield_per_hectare": "lots",
				"yield_unit":        "kg",
			},
			expectedError: "yield_per_hectare must be a number",
		},
		{
			name: "invalid yield_unit type",
			outcome: map[string]interface{}{
				"yield_per_hectare": 2500.0,
				"yield_unit":        123,
			},
			expectedError: "yield_unit must be a string",
		},
		{
			name: "invalid total_yield",
			outcome: map[string]interface{}{
				"yield_per_hectare": 2500.0,
				"yield_unit":        "kg",
				"total_yield":       -12500.0,
			},
			expectedError: "total_yield must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAnnualOutcome(tt.outcome)

			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			}
		})
	}
}

func TestCropCycle_ValidateOutcome(t *testing.T) {
	tests := []struct {
		name          string
		season        string
		outcome       map[string]interface{}
		expectedError string
	}{
		{
			name:   "perennial crop with valid perennial outcome",
			season: "PERENNIAL",
			outcome: map[string]interface{}{
				"age_range_min":  5,
				"age_range_max":  10,
				"yield_per_tree": 50.5,
				"yield_unit":     "kg",
			},
			expectedError: "",
		},
		{
			name:   "perennial crop with invalid outcome (missing yield_per_tree)",
			season: "PERENNIAL",
			outcome: map[string]interface{}{
				"age_range_min": 5,
				"age_range_max": 10,
				"yield_unit":    "kg",
			},
			expectedError: "perennial crop outcome must include yield_per_tree",
		},
		{
			name:   "annual crop (RABI) with valid annual outcome",
			season: "RABI",
			outcome: map[string]interface{}{
				"yield_per_hectare": 2500.0,
				"yield_unit":        "kg",
			},
			expectedError: "",
		},
		{
			name:   "annual crop (KHARIF) with valid annual outcome",
			season: "KHARIF",
			outcome: map[string]interface{}{
				"yield_per_hectare": 2500.0,
				"yield_unit":        "kg",
			},
			expectedError: "",
		},
		{
			name:   "annual crop (ZAID) with valid annual outcome",
			season: "ZAID",
			outcome: map[string]interface{}{
				"yield_per_hectare": 2500.0,
				"yield_unit":        "kg",
			},
			expectedError: "",
		},
		{
			name:   "annual crop with invalid outcome (missing yield_per_hectare)",
			season: "RABI",
			outcome: map[string]interface{}{
				"yield_unit": "kg",
			},
			expectedError: "annual crop outcome must include yield_per_hectare",
		},
		{
			name:          "empty outcome is valid for any season",
			season:        "RABI",
			outcome:       map[string]interface{}{},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cropCycle := &CropCycle{
				Season:  tt.season,
				Outcome: tt.outcome,
			}

			err := cropCycle.ValidateOutcome()

			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			}
		})
	}
}

func TestCropCycle_IsPerennial(t *testing.T) {
	tests := []struct {
		name     string
		season   string
		expected bool
	}{
		{"perennial crop", "PERENNIAL", true},
		{"rabi crop", "RABI", false},
		{"kharif crop", "KHARIF", false},
		{"zaid crop", "ZAID", false},
		{"other crop", "OTHER", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cropCycle := &CropCycle{Season: tt.season}
			assert.Equal(t, tt.expected, cropCycle.IsPerennial())
		})
	}
}
