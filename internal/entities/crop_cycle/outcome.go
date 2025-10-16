package crop_cycle

import "fmt"

// PerennialCropOutcome represents yield data for perennial crops
// Used when Season = PERENNIAL
type PerennialCropOutcome struct {
	AgeRangeMin   int      `json:"age_range_min" validate:"required,gte=0" example:"5"`
	AgeRangeMax   int      `json:"age_range_max" validate:"required,gte=0,gtefield=AgeRangeMin" example:"10"`
	YieldPerTree  float64  `json:"yield_per_tree" validate:"required,gt=0" example:"50.5"`
	YieldUnit     string   `json:"yield_unit" validate:"required" example:"kg"`
	NumberOfTrees *int     `json:"number_of_trees,omitempty" validate:"omitempty,gt=0" example:"100"`
	TotalYield    *float64 `json:"total_yield,omitempty" validate:"omitempty,gt=0" example:"5050"`
	QualityGrade  string   `json:"quality_grade,omitempty" example:"A"`
	Notes         string   `json:"notes,omitempty" example:"Trees in prime productive age"`
}

// AnnualCropOutcome represents yield data for annual crops
// Used when Season = RABI, KHARIF, ZAID, or OTHER
type AnnualCropOutcome struct {
	YieldPerHectare float64  `json:"yield_per_hectare" validate:"required,gt=0" example:"2500"`
	YieldUnit       string   `json:"yield_unit" validate:"required" example:"kg"`
	TotalYield      *float64 `json:"total_yield,omitempty" validate:"omitempty,gt=0" example:"12500"`
	QualityGrade    string   `json:"quality_grade,omitempty" example:"A"`
	Notes           string   `json:"notes,omitempty" example:"Good weather conditions"`
}

// ValidatePerennialOutcome validates outcome data for perennial crops
func ValidatePerennialOutcome(outcome map[string]interface{}) error {
	if len(outcome) == 0 {
		return nil // Outcome is optional
	}

	// Validate required fields
	ageMin, hasMin := outcome["age_range_min"]
	ageMax, hasMax := outcome["age_range_max"]
	yieldPerTree, hasYield := outcome["yield_per_tree"]
	yieldUnit, hasUnit := outcome["yield_unit"]

	if !hasMin {
		return fmt.Errorf("perennial crop outcome must include age_range_min")
	}
	if !hasMax {
		return fmt.Errorf("perennial crop outcome must include age_range_max")
	}
	if !hasYield {
		return fmt.Errorf("perennial crop outcome must include yield_per_tree")
	}
	if !hasUnit {
		return fmt.Errorf("perennial crop outcome must include yield_unit")
	}

	// Convert and validate age_range_min
	var minAge float64
	switch v := ageMin.(type) {
	case float64:
		minAge = v
	case int:
		minAge = float64(v)
	default:
		return fmt.Errorf("age_range_min must be a number")
	}

	if minAge < 0 {
		return fmt.Errorf("age_range_min must be non-negative")
	}

	// Convert and validate age_range_max
	var maxAge float64
	switch v := ageMax.(type) {
	case float64:
		maxAge = v
	case int:
		maxAge = float64(v)
	default:
		return fmt.Errorf("age_range_max must be a number")
	}

	if maxAge < 0 {
		return fmt.Errorf("age_range_max must be non-negative")
	}

	if maxAge < minAge {
		return fmt.Errorf("age_range_max must be greater than or equal to age_range_min")
	}

	// Convert and validate yield_per_tree
	var yieldValue float64
	switch v := yieldPerTree.(type) {
	case float64:
		yieldValue = v
	case int:
		yieldValue = float64(v)
	default:
		return fmt.Errorf("yield_per_tree must be a number")
	}

	if yieldValue <= 0 {
		return fmt.Errorf("yield_per_tree must be greater than 0")
	}

	// Validate yield_unit is a string
	if _, ok := yieldUnit.(string); !ok {
		return fmt.Errorf("yield_unit must be a string")
	}

	// Validate optional fields if present
	if numTrees, hasNumTrees := outcome["number_of_trees"]; hasNumTrees {
		var treesCount float64
		switch v := numTrees.(type) {
		case float64:
			treesCount = v
		case int:
			treesCount = float64(v)
		default:
			return fmt.Errorf("number_of_trees must be a number")
		}

		if treesCount <= 0 {
			return fmt.Errorf("number_of_trees must be greater than 0")
		}
	}

	if totalYield, hasTotalYield := outcome["total_yield"]; hasTotalYield {
		var totalValue float64
		switch v := totalYield.(type) {
		case float64:
			totalValue = v
		case int:
			totalValue = float64(v)
		default:
			return fmt.Errorf("total_yield must be a number")
		}

		if totalValue <= 0 {
			return fmt.Errorf("total_yield must be greater than 0")
		}
	}

	return nil
}

// ValidateAnnualOutcome validates outcome data for annual crops
func ValidateAnnualOutcome(outcome map[string]interface{}) error {
	if len(outcome) == 0 {
		return nil // Outcome is optional
	}

	// Validate required fields
	yieldPerHectare, hasYield := outcome["yield_per_hectare"]
	yieldUnit, hasUnit := outcome["yield_unit"]

	if !hasYield {
		return fmt.Errorf("annual crop outcome must include yield_per_hectare")
	}
	if !hasUnit {
		return fmt.Errorf("annual crop outcome must include yield_unit")
	}

	// Convert and validate yield_per_hectare
	var yieldValue float64
	switch v := yieldPerHectare.(type) {
	case float64:
		yieldValue = v
	case int:
		yieldValue = float64(v)
	default:
		return fmt.Errorf("yield_per_hectare must be a number")
	}

	if yieldValue <= 0 {
		return fmt.Errorf("yield_per_hectare must be greater than 0")
	}

	// Validate yield_unit is a string
	if _, ok := yieldUnit.(string); !ok {
		return fmt.Errorf("yield_unit must be a string")
	}

	// Validate optional total_yield if present
	if totalYield, hasTotalYield := outcome["total_yield"]; hasTotalYield {
		var totalValue float64
		switch v := totalYield.(type) {
		case float64:
			totalValue = v
		case int:
			totalValue = float64(v)
		default:
			return fmt.Errorf("total_yield must be a number")
		}

		if totalValue <= 0 {
			return fmt.Errorf("total_yield must be greater than 0")
		}
	}

	return nil
}
