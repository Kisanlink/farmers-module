package requests

// CreateCropRequest represents a request to create a new crop
type CreateCropRequest struct {
	BaseRequest
	Name           string            `json:"name" validate:"required,min=2,max=255"`
	ScientificName *string           `json:"scientific_name,omitempty" validate:"omitempty,max=255"`
	Category       string            `json:"category" validate:"required,oneof=CEREALS PULSES VEGETABLES FRUITS OIL_SEEDS SPICES CASH_CROPS FODDER MEDICINAL OTHER"`
	DurationDays   *int              `json:"duration_days,omitempty" validate:"omitempty,min=1,max=365"`
	Seasons        []string          `json:"seasons" validate:"required,min=1,dive,oneof=RABI KHARIF ZAID"`
	Unit           string            `json:"unit" validate:"required,min=1,max=50"`
	Properties     map[string]string `json:"properties,omitempty"`
}

// UpdateCropRequest represents a request to update an existing crop
type UpdateCropRequest struct {
	BaseRequest
	ID             string            `json:"id" validate:"required"`
	Name           *string           `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	ScientificName *string           `json:"scientific_name,omitempty" validate:"omitempty,max=255"`
	Category       *string           `json:"category,omitempty" validate:"omitempty,oneof=CEREALS PULSES VEGETABLES FRUITS OIL_SEEDS SPICES CASH_CROPS FODDER MEDICINAL OTHER"`
	DurationDays   *int              `json:"duration_days,omitempty" validate:"omitempty,min=1,max=365"`
	Seasons        []string          `json:"seasons,omitempty" validate:"omitempty,min=1,dive,oneof=RABI KHARIF ZAID"`
	Unit           *string           `json:"unit,omitempty" validate:"omitempty,min=1,max=50"`
	Properties     map[string]string `json:"properties,omitempty"`
	IsActive       *bool             `json:"is_active,omitempty"`
}

// DeleteCropRequest represents a request to delete a crop
type DeleteCropRequest struct {
	BaseRequest
	ID string `json:"id" validate:"required"`
}

// GetCropRequest represents a request to retrieve a crop
type GetCropRequest struct {
	BaseRequest
	ID string `json:"id" validate:"required"`
}

// ListCropsRequest represents a request to list crops with filtering
type ListCropsRequest struct {
	FilterRequest
	Category string   `json:"category,omitempty" validate:"omitempty,oneof=CEREALS PULSES VEGETABLES FRUITS OIL_SEEDS SPICES CASH_CROPS FODDER MEDICINAL OTHER"`
	Season   string   `json:"season,omitempty" validate:"omitempty,oneof=RABI KHARIF ZAID"`
	IsActive *bool    `json:"is_active,omitempty"`
	Search   string   `json:"search,omitempty"` // Search in name and scientific_name
	Seasons  []string `json:"seasons,omitempty" validate:"omitempty,dive,oneof=RABI KHARIF ZAID"`
}

// CreateCropVarietyRequest represents a request to create a new crop variety
type CreateCropVarietyRequest struct {
	BaseRequest
	CropID       string            `json:"crop_id" validate:"required"`
	Name         string            `json:"name" validate:"required,min=2,max=255"`
	Description  *string           `json:"description,omitempty"`
	DurationDays *int              `json:"duration_days,omitempty" validate:"omitempty,min=1,max=365"`
	YieldPerAcre *float64          `json:"yield_per_acre,omitempty" validate:"omitempty,min=0"`
	YieldPerTree *float64          `json:"yield_per_tree,omitempty" validate:"omitempty,min=0"`
	Properties   map[string]string `json:"properties,omitempty"`
}

// UpdateCropVarietyRequest represents a request to update an existing crop variety
type UpdateCropVarietyRequest struct {
	BaseRequest
	ID           string            `json:"id" validate:"required"`
	Name         *string           `json:"name,omitempty" validate:"omitempty,min=2,max=255"`
	Description  *string           `json:"description,omitempty"`
	DurationDays *int              `json:"duration_days,omitempty" validate:"omitempty,min=1,max=365"`
	YieldPerAcre *float64          `json:"yield_per_acre,omitempty" validate:"omitempty,min=0"`
	YieldPerTree *float64          `json:"yield_per_tree,omitempty" validate:"omitempty,min=0"`
	Properties   map[string]string `json:"properties,omitempty"`
	IsActive     *bool             `json:"is_active,omitempty"`
}

// DeleteCropVarietyRequest represents a request to delete a crop variety
type DeleteCropVarietyRequest struct {
	BaseRequest
	ID string `json:"id" validate:"required"`
}

// GetCropVarietyRequest represents a request to retrieve a crop variety
type GetCropVarietyRequest struct {
	BaseRequest
	ID string `json:"id" validate:"required"`
}

// ListCropVarietiesRequest represents a request to list crop varieties
type ListCropVarietiesRequest struct {
	FilterRequest
	CropID   string `json:"crop_id,omitempty" validate:"omitempty"`
	IsActive *bool  `json:"is_active,omitempty"`
	Search   string `json:"search,omitempty"` // Search in name and description
}

// GetCropLookupRequest represents a request to get crop lookup data
type GetCropLookupRequest struct {
	BaseRequest
	Category string `json:"category,omitempty" validate:"omitempty,oneof=CEREALS PULSES VEGETABLES FRUITS OIL_SEEDS SPICES CASH_CROPS FODDER MEDICINAL OTHER"`
	Season   string `json:"season,omitempty" validate:"omitempty,oneof=RABI KHARIF ZAID"`
}

// GetVarietyLookupRequest represents a request to get variety lookup data for a specific crop
type GetVarietyLookupRequest struct {
	BaseRequest
	CropID string `json:"crop_id" validate:"required"`
}

// NewCreateCropRequest creates a new create crop request
func NewCreateCropRequest() CreateCropRequest {
	return CreateCropRequest{
		BaseRequest: NewBaseRequest(),
		Properties:  make(map[string]string),
		Unit:        "kg",
	}
}

// NewUpdateCropRequest creates a new update crop request
func NewUpdateCropRequest() UpdateCropRequest {
	return UpdateCropRequest{
		BaseRequest: NewBaseRequest(),
		Properties:  make(map[string]string),
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
		Properties:  make(map[string]string),
	}
}

// NewUpdateCropVarietyRequest creates a new update crop variety request
func NewUpdateCropVarietyRequest() UpdateCropVarietyRequest {
	return UpdateCropVarietyRequest{
		BaseRequest: NewBaseRequest(),
		Properties:  make(map[string]string),
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