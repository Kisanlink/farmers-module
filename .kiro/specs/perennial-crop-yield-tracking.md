# Perennial Crop Yield Tracking

## Overview

Perennial crops have different yield tracking requirements compared to annual crops (RABI, KHARIF, ZAID). Instead of measuring yield per acre/hectare for a single season, perennial crops are tracked based on tree age and yield per tree.

## Business Requirements

### Yield Data Structure

For perennial crops (Season = PERENNIAL), the outcome data should capture:

1. **Age Range**: The age range of trees during the harvest period
   - `age_range_min`: Minimum age of trees (in years)
   - `age_range_max`: Maximum age of trees (in years)

2. **Yield Per Tree**: Average yield produced per tree
   - `yield_per_tree`: Numeric value representing yield
   - `yield_unit`: Unit of measurement (e.g., "kg", "quintal", "ton")

3. **Optional Fields**:
   - `number_of_trees`: Total number of trees in the cycle
   - `total_yield`: Total yield from all trees (calculated or manually entered)
   - `quality_grade`: Quality classification (e.g., "A", "B", "C")
   - `notes`: Additional observations

### Comparison: Annual vs Perennial Crops

#### Annual Crops (RABI, KHARIF, ZAID)
```json
{
  "yield_per_hectare": 2500,
  "yield_unit": "kg",
  "total_yield": 12500,
  "quality_grade": "A",
  "notes": "Good weather conditions"
}
```

#### Perennial Crops (PERENNIAL)
```json
{
  "age_range_min": 5,
  "age_range_max": 10,
  "yield_per_tree": 50.5,
  "yield_unit": "kg",
  "number_of_trees": 100,
  "total_yield": 5050,
  "quality_grade": "A",
  "notes": "Trees in prime productive age"
}
```

## Technical Implementation

### Database Schema

The `crop_cycles` table already has an `outcome` column of type `jsonb`, which can store both annual and perennial outcome structures without schema changes.

```sql
-- No migration needed, existing schema supports this:
outcome jsonb DEFAULT '{}'
```

### Data Model

#### CropCycle Entity

The `Outcome` field (type `entities.JSONB`) will store season-specific data:

```go
type CropCycle struct {
    // ... other fields
    Season    string         `json:"season" gorm:"type:season;not null"`
    Outcome   entities.JSONB `json:"outcome" gorm:"type:jsonb;default:'{}';serializer:json"`
}
```

#### Outcome Structures

```go
// PerennialCropOutcome represents yield data for perennial crops
type PerennialCropOutcome struct {
    AgeRangeMin   int     `json:"age_range_min" validate:"required,gte=0"`
    AgeRangeMax   int     `json:"age_range_max" validate:"required,gte=0,gtefield=AgeRangeMin"`
    YieldPerTree  float64 `json:"yield_per_tree" validate:"required,gt=0"`
    YieldUnit     string  `json:"yield_unit" validate:"required"`
    NumberOfTrees *int    `json:"number_of_trees,omitempty" validate:"omitempty,gt=0"`
    TotalYield    *float64 `json:"total_yield,omitempty" validate:"omitempty,gt=0"`
    QualityGrade  string  `json:"quality_grade,omitempty"`
    Notes         string  `json:"notes,omitempty"`
}

// AnnualCropOutcome represents yield data for annual crops
type AnnualCropOutcome struct {
    YieldPerHectare float64 `json:"yield_per_hectare" validate:"required,gt=0"`
    YieldUnit       string  `json:"yield_unit" validate:"required"`
    TotalYield      *float64 `json:"total_yield,omitempty" validate:"omitempty,gt=0"`
    QualityGrade    string  `json:"quality_grade,omitempty"`
    Notes           string  `json:"notes,omitempty"`
}
```

### API Changes

#### EndCycleRequest

The `EndCycleRequest.Outcome` field remains `map[string]interface{}` to support both structures:

```go
type EndCycleRequest struct {
    BaseRequest
    ID      string                 `json:"id" validate:"required"`
    Status  string                 `json:"status" validate:"required,oneof=COMPLETED CANCELLED"`
    EndDate time.Time              `json:"end_date" validate:"required"`
    Outcome map[string]interface{} `json:"outcome,omitempty"`
}
```

#### Validation Logic

Add season-aware validation:

```go
func (cc *CropCycle) ValidateOutcome() error {
    if len(cc.Outcome) == 0 {
        return nil // Outcome is optional
    }

    if cc.Season == "PERENNIAL" {
        return validatePerennialOutcome(cc.Outcome)
    }

    return validateAnnualOutcome(cc.Outcome)
}

func validatePerennialOutcome(outcome map[string]interface{}) error {
    // Validate required fields for perennial crops
    ageMin, hasMin := outcome["age_range_min"].(float64)
    ageMax, hasMax := outcome["age_range_max"].(float64)
    yieldPerTree, hasYield := outcome["yield_per_tree"].(float64)
    yieldUnit, hasUnit := outcome["yield_unit"].(string)

    if !hasMin || !hasMax || !hasYield || !hasUnit {
        return fmt.Errorf("perennial crop outcome must include age_range_min, age_range_max, yield_per_tree, and yield_unit")
    }

    if ageMin < 0 || ageMax < 0 {
        return fmt.Errorf("age range values must be non-negative")
    }

    if ageMax < ageMin {
        return fmt.Errorf("age_range_max must be greater than or equal to age_range_min")
    }

    if yieldPerTree <= 0 {
        return fmt.Errorf("yield_per_tree must be greater than 0")
    }

    return nil
}
```

### Response Structure

The API response should include the outcome data as-is:

```json
{
  "success": true,
  "message": "Crop cycle ended successfully",
  "data": {
    "id": "CRCY-123",
    "farm_id": "FARM-456",
    "season": "PERENNIAL",
    "status": "COMPLETED",
    "outcome": {
      "age_range_min": 5,
      "age_range_max": 10,
      "yield_per_tree": 50.5,
      "yield_unit": "kg",
      "number_of_trees": 100,
      "total_yield": 5050
    }
  }
}
```

## UI/Frontend Considerations

The frontend should:

1. **Detect Season Type**: Check the `season` field when ending a crop cycle
2. **Conditional Form Fields**:
   - If `season == "PERENNIAL"`: Show age range and yield per tree fields
   - Otherwise: Show yield per hectare/acre fields
3. **Field Validation**: Enforce business rules client-side for better UX

### Example Form Logic

```javascript
if (cropCycle.season === 'PERENNIAL') {
  // Show perennial form
  fields = [
    { name: 'age_range_min', label: 'Age Range (Min)', type: 'number', required: true },
    { name: 'age_range_max', label: 'Age Range (Max)', type: 'number', required: true },
    { name: 'yield_per_tree', label: 'Yield Per Tree', type: 'number', required: true },
    { name: 'yield_unit', label: 'Yield Unit', type: 'select', options: ['kg', 'quintal', 'ton'], required: true },
    { name: 'number_of_trees', label: 'Number of Trees', type: 'number', optional: true },
  ];
} else {
  // Show annual crop form
  fields = [
    { name: 'yield_per_hectare', label: 'Yield Per Hectare', type: 'number', required: true },
    { name: 'yield_unit', label: 'Yield Unit', type: 'select', options: ['kg', 'quintal', 'ton'], required: true },
  ];
}
```

## Implementation Checklist

- [ ] Add `PerennialCropOutcome` struct to `internal/entities/crop_cycle/outcome.go`
- [ ] Add `AnnualCropOutcome` struct to `internal/entities/crop_cycle/outcome.go`
- [ ] Add `ValidateOutcome()` method to `CropCycle` entity
- [ ] Add `validatePerennialOutcome()` helper function
- [ ] Add `validateAnnualOutcome()` helper function
- [ ] Update `EndCycle` service method to call outcome validation
- [ ] Add unit tests for perennial outcome validation
- [ ] Add unit tests for annual outcome validation
- [ ] Add integration test with complete perennial cycle
- [ ] Update API documentation with perennial outcome examples
- [ ] Document frontend requirements for conditional forms

## Testing Scenarios

### Test Case 1: Valid Perennial Outcome
```json
{
  "age_range_min": 5,
  "age_range_max": 10,
  "yield_per_tree": 50.5,
  "yield_unit": "kg"
}
```
**Expected**: ✅ Accepted

### Test Case 2: Invalid Age Range
```json
{
  "age_range_min": 10,
  "age_range_max": 5,
  "yield_per_tree": 50.5,
  "yield_unit": "kg"
}
```
**Expected**: ❌ Error: "age_range_max must be greater than or equal to age_range_min"

### Test Case 3: Missing Required Field
```json
{
  "age_range_min": 5,
  "age_range_max": 10,
  "yield_unit": "kg"
}
```
**Expected**: ❌ Error: "perennial crop outcome must include yield_per_tree"

### Test Case 4: Negative Yield
```json
{
  "age_range_min": 5,
  "age_range_max": 10,
  "yield_per_tree": -50.5,
  "yield_unit": "kg"
}
```
**Expected**: ❌ Error: "yield_per_tree must be greater than 0"

## Migration Notes

- **No database migration required**: The `outcome` JSONB field is already flexible
- **Backward compatible**: Existing annual crop cycles continue to work
- **Data validation**: Only applied when ending cycles, not retroactively

## References

- Season enum definition: `internal/entities/enums.go`
- Database schema: `internal/db/db.go`
- CropCycle entity: `internal/entities/crop_cycle/crop_cycle.go`
- End cycle service: `internal/services/crop_cycle_service.go`
