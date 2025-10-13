package requests

// CreateCropRequest represents a request to create a new crop
type CreateCropRequest struct {
	BaseRequest
	Name           string                 `json:"name" validate:"required,min=2,max=255" example:"Wheat"`
	ScientificName *string                `json:"scientific_name,omitempty" validate:"omitempty,max=255" example:"Triticum aestivum"`
	Category       string                 `json:"category" validate:"required,oneof=CEREALS PULSES VEGETABLES FRUITS OIL_SEEDS SPICES CASH_CROPS FODDER MEDICINAL OTHER" example:"CEREALS"`
	DurationDays   *int                   `json:"duration_days,omitempty" validate:"omitempty,min=1,max=365" example:"120"`
	Seasons        []string               `json:"seasons" validate:"required,min=1,dive,oneof=RABI KHARIF ZAID" example:"RABI"`
	Unit           string                 `json:"unit" validate:"required,min=1,max=50" example:"kg"`
	Properties     map[string]interface{} `json:"properties,omitempty" example:"water_requirement:medium,climate:temperate"`
}

// UpdateCropRequest represents a request to update an existing crop
type UpdateCropRequest struct {
	BaseRequest
	ID             string                 `json:"id" validate:"required" example:"crop_123e4567-e89b-12d3-a456-426614174000"`
	Name           *string                `json:"name,omitempty" validate:"omitempty,min=2,max=255" example:"Wheat - HD2967"`
	ScientificName *string                `json:"scientific_name,omitempty" validate:"omitempty,max=255" example:"Triticum aestivum L."`
	Category       *string                `json:"category,omitempty" validate:"omitempty,oneof=CEREALS PULSES VEGETABLES FRUITS OIL_SEEDS SPICES CASH_CROPS FODDER MEDICINAL OTHER" example:"CEREALS"`
	DurationDays   *int                   `json:"duration_days,omitempty" validate:"omitempty,min=1,max=365" example:"125"`
	Seasons        []string               `json:"seasons,omitempty" validate:"omitempty,min=1,dive,oneof=RABI KHARIF ZAID" example:"RABI"`
	Unit           *string                `json:"unit,omitempty" validate:"omitempty,min=1,max=50" example:"quintal"`
	Properties     map[string]interface{} `json:"properties,omitempty" example:"irrigation:required,fertilizer:high"`
	IsActive       *bool                  `json:"is_active,omitempty" example:"true"`
}

// DeleteCropRequest represents a request to delete a crop
type DeleteCropRequest struct {
	BaseRequest
	ID string `json:"id" validate:"required" example:"crop_123e4567-e89b-12d3-a456-426614174000"`
}

// GetCropRequest represents a request to retrieve a crop
type GetCropRequest struct {
	BaseRequest
	ID string `json:"id" validate:"required" example:"crop_123e4567-e89b-12d3-a456-426614174000"`
}

// ListCropsRequest represents a request to list crops with filtering
type ListCropsRequest struct {
	FilterRequest
	Category string   `json:"category,omitempty" validate:"omitempty,oneof=CEREALS PULSES VEGETABLES FRUITS OIL_SEEDS SPICES CASH_CROPS FODDER MEDICINAL OTHER" example:"CEREALS"`
	Season   string   `json:"season,omitempty" validate:"omitempty,oneof=RABI KHARIF ZAID" example:"RABI"`
	IsActive *bool    `json:"is_active,omitempty" example:"true"`
	Search   string   `json:"search,omitempty" example:"wheat"` // Search in name and scientific_name
	Seasons  []string `json:"seasons,omitempty" validate:"omitempty,dive,oneof=RABI KHARIF ZAID" example:"RABI,KHARIF"`
}

// YieldByAgeRequest represents yield information for a specific tree age range
type YieldByAgeRequest struct {
	AgeFrom      int     `json:"age_from" validate:"min=0" example:"5"`
	AgeTo        int     `json:"age_to" validate:"gtfield=AgeFrom" example:"10"`
	YieldPerTree float64 `json:"yield_per_tree" validate:"min=0" example:"50.5"`
}

// CreateCropVarietyRequest represents a request to create a new crop variety
type CreateCropVarietyRequest struct {
	BaseRequest
	CropID       string                 `json:"crop_id" validate:"required" example:"crop_123e4567-e89b-12d3-a456-426614174000"`
	Name         string                 `json:"name" validate:"required,min=2,max=255" example:"HD-2967"`
	Description  *string                `json:"description,omitempty" example:"High yielding wheat variety suitable for irrigated conditions"`
	DurationDays *int                   `json:"duration_days,omitempty" validate:"omitempty,min=1,max=365" example:"120"`
	YieldPerAcre *float64               `json:"yield_per_acre,omitempty" validate:"omitempty,min=0" example:"25.5"`
	YieldPerTree *float64               `json:"yield_per_tree,omitempty" validate:"omitempty,min=0" example:"0"`
	YieldByAge   []YieldByAgeRequest    `json:"yield_by_age,omitempty"`
	Properties   map[string]interface{} `json:"properties,omitempty" example:"disease_resistance:high,recommended_for:punjab_haryana"`
}

// UpdateCropVarietyRequest represents a request to update an existing crop variety
type UpdateCropVarietyRequest struct {
	BaseRequest
	ID           string                 `json:"id" validate:"required" example:"variety_123e4567-e89b-12d3-a456-426614174000"`
	Name         *string                `json:"name,omitempty" validate:"omitempty,min=2,max=255" example:"HD-2967 (Improved)"`
	Description  *string                `json:"description,omitempty" example:"Updated description with latest cultivation practices"`
	DurationDays *int                   `json:"duration_days,omitempty" validate:"omitempty,min=1,max=365" example:"118"`
	YieldPerAcre *float64               `json:"yield_per_acre,omitempty" validate:"omitempty,min=0" example:"27.0"`
	YieldPerTree *float64               `json:"yield_per_tree,omitempty" validate:"omitempty,min=0" example:"0"`
	YieldByAge   []YieldByAgeRequest    `json:"yield_by_age,omitempty"`
	Properties   map[string]interface{} `json:"properties,omitempty" example:"drought_tolerance:medium,market_demand:high"`
	IsActive     *bool                  `json:"is_active,omitempty" example:"true"`
}

// DeleteCropVarietyRequest represents a request to delete a crop variety
type DeleteCropVarietyRequest struct {
	BaseRequest
	ID string `json:"id" validate:"required" example:"variety_123e4567-e89b-12d3-a456-426614174000"`
}

// GetCropVarietyRequest represents a request to retrieve a crop variety
type GetCropVarietyRequest struct {
	BaseRequest
	ID string `json:"id" validate:"required" example:"variety_123e4567-e89b-12d3-a456-426614174000"`
}

// ListCropVarietiesRequest represents a request to list crop varieties
type ListCropVarietiesRequest struct {
	FilterRequest
	CropID   string `json:"crop_id,omitempty" validate:"omitempty" example:"crop_123e4567-e89b-12d3-a456-426614174000"`
	IsActive *bool  `json:"is_active,omitempty" example:"true"`
	Search   string `json:"search,omitempty" example:"HD-2967"` // Search in name and description
}

// GetCropLookupRequest represents a request to get crop lookup data
type GetCropLookupRequest struct {
	BaseRequest
	Category string `json:"category,omitempty" validate:"omitempty,oneof=CEREALS PULSES VEGETABLES FRUITS OIL_SEEDS SPICES CASH_CROPS FODDER MEDICINAL OTHER" example:"CEREALS"`
	Season   string `json:"season,omitempty" validate:"omitempty,oneof=RABI KHARIF ZAID" example:"RABI"`
}

// GetVarietyLookupRequest represents a request to get variety lookup data for a specific crop
type GetVarietyLookupRequest struct {
	BaseRequest
	CropID string `json:"crop_id" validate:"required" example:"crop_123e4567-e89b-12d3-a456-426614174000"`
}

// NewCreateCropRequest creates a new create crop request
func NewCreateCropRequest() CreateCropRequest {
	return CreateCropRequest{
		BaseRequest: NewBaseRequest(),
		Properties:  make(map[string]interface{}),
		Unit:        "kg",
	}
}

// NewUpdateCropRequest creates a new update crop request
func NewUpdateCropRequest() UpdateCropRequest {
	return UpdateCropRequest{
		BaseRequest: NewBaseRequest(),
		Properties:  make(map[string]interface{}),
	}
}

// NewDeleteCropRequest creates a new delete crop request
func NewDeleteCropRequest() DeleteCropRequest {
	return DeleteCropRequest{
		BaseRequest: NewBaseRequest(),
	}
}

// NewGetCropRequest creates a new get crop request
func NewGetCropRequest() GetCropRequest {
	return GetCropRequest{
		BaseRequest: NewBaseRequest(),
	}
}

// NewListCropsRequest creates a new list crops request
func NewListCropsRequest() ListCropsRequest {
	return ListCropsRequest{
		FilterRequest: FilterRequest{
			PaginationRequest: NewPaginationRequest(1, 20),
		},
	}
}

// NewCreateCropVarietyRequest creates a new create crop variety request
func NewCreateCropVarietyRequest() CreateCropVarietyRequest {
	return CreateCropVarietyRequest{
		BaseRequest: NewBaseRequest(),
		Properties:  make(map[string]interface{}),
	}
}

// NewUpdateCropVarietyRequest creates a new update crop variety request
func NewUpdateCropVarietyRequest() UpdateCropVarietyRequest {
	return UpdateCropVarietyRequest{
		BaseRequest: NewBaseRequest(),
		Properties:  make(map[string]interface{}),
	}
}

// NewDeleteCropVarietyRequest creates a new delete crop variety request
func NewDeleteCropVarietyRequest() DeleteCropVarietyRequest {
	return DeleteCropVarietyRequest{
		BaseRequest: NewBaseRequest(),
	}
}

// NewGetCropVarietyRequest creates a new get crop variety request
func NewGetCropVarietyRequest() GetCropVarietyRequest {
	return GetCropVarietyRequest{
		BaseRequest: NewBaseRequest(),
	}
}

// NewListCropVarietiesRequest creates a new list crop varieties request
func NewListCropVarietiesRequest() ListCropVarietiesRequest {
	return ListCropVarietiesRequest{
		FilterRequest: FilterRequest{
			PaginationRequest: NewPaginationRequest(1, 20),
		},
	}
}

// NewGetCropLookupRequest creates a new get crop lookup request
func NewGetCropLookupRequest() GetCropLookupRequest {
	return GetCropLookupRequest{
		BaseRequest: NewBaseRequest(),
	}
}

// NewGetVarietyLookupRequest creates a new get variety lookup request
func NewGetVarietyLookupRequest() GetVarietyLookupRequest {
	return GetVarietyLookupRequest{
		BaseRequest: NewBaseRequest(),
	}
}
