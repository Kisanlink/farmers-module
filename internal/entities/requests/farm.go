package requests

// CreateFarmRequest represents a request to create a new farm
type CreateFarmRequest struct {
	BaseRequest
	AAAUserID                 string                    `json:"aaa_user_id" validate:"required" example:"usr_123e4567-e89b-12d3-a456-426614174000"`
	AAAOrgID                  string                    `json:"aaa_org_id" validate:"required" example:"org_123e4567-e89b-12d3-a456-426614174000"`
	Name                      *string                   `json:"name,omitempty" example:"North Field Farm"`
	OwnershipType             string                    `json:"ownership_type,omitempty" validate:"omitempty,oneof=OWN LEASE SHARED" example:"OWN"`
	AreaHa                    float64                   `json:"area_ha" validate:"required,min=0.01" example:"2.5"`
	Geometry                  GeometryData              `json:"geometry,omitempty"`
	SoilTypeID                *string                   `json:"soil_type_id,omitempty" example:"soil_123e4567-e89b-12d3-a456-426614174000"`
	PrimaryIrrigationSourceID *string                   `json:"primary_irrigation_source_id,omitempty" example:"irr_123e4567-e89b-12d3-a456-426614174000"`
	BoreWellCount             int                       `json:"bore_well_count,omitempty" validate:"min=0" example:"2"`
	OtherIrrigationDetails    *string                   `json:"other_irrigation_details,omitempty" example:"Canal irrigation available during monsoon"`
	IrrigationSources         []IrrigationSourceRequest `json:"irrigation_sources,omitempty"`
	Metadata                  map[string]string         `json:"metadata,omitempty" example:"soil_test_date:2024-01-10,elevation:450m"`
}

// UpdateFarmRequest represents a request to update an existing farm
type UpdateFarmRequest struct {
	BaseRequest
	ID                        string                    `json:"id" validate:"required" example:"farm_123e4567-e89b-12d3-a456-426614174000"`
	AAAUserID                 string                    `json:"aaa_user_id,omitempty" example:"usr_123e4567-e89b-12d3-a456-426614174000"`
	AAAOrgID                  string                    `json:"aaa_org_id,omitempty" example:"org_123e4567-e89b-12d3-a456-426614174000"`
	Name                      *string                   `json:"name,omitempty" example:"North Field Farm - Updated"`
	OwnershipType             *string                   `json:"ownership_type,omitempty" validate:"omitempty,oneof=OWN LEASE SHARED" example:"LEASE"`
	AreaHa                    *float64                  `json:"area_ha,omitempty" example:"3.0"`
	Geometry                  *GeometryData             `json:"geometry,omitempty"`
	SoilTypeID                *string                   `json:"soil_type_id,omitempty" example:"soil_123e4567-e89b-12d3-a456-426614174000"`
	PrimaryIrrigationSourceID *string                   `json:"primary_irrigation_source_id,omitempty" example:"irr_123e4567-e89b-12d3-a456-426614174000"`
	BoreWellCount             *int                      `json:"bore_well_count,omitempty" validate:"omitempty,min=0" example:"3"`
	OtherIrrigationDetails    *string                   `json:"other_irrigation_details,omitempty" example:"Drip irrigation installed"`
	IrrigationSources         []IrrigationSourceRequest `json:"irrigation_sources,omitempty"`
	Metadata                  map[string]string         `json:"metadata,omitempty" example:"last_survey:2024-02-15,certification:organic"`
}

// DeleteFarmRequest represents a request to delete a farm
type DeleteFarmRequest struct {
	BaseRequest
	ID string `json:"id" validate:"required" example:"farm_123e4567-e89b-12d3-a456-426614174000"`
}

// GetFarmRequest represents a request to retrieve a farm
type GetFarmRequest struct {
	BaseRequest
	ID string `json:"id" validate:"required" example:"farm_123e4567-e89b-12d3-a456-426614174000"`
}

// ListFarmsRequest represents a request to list farms with filtering
type ListFarmsRequest struct {
	FilterRequest
	AAAUserID string   `json:"aaa_user_id,omitempty" example:"usr_123e4567-e89b-12d3-a456-426614174000"`
	AAAOrgID  string   `json:"aaa_org_id,omitempty" example:"org_123e4567-e89b-12d3-a456-426614174000"`
	MinArea   *float64 `json:"min_area,omitempty" example:"1.0"`
	MaxArea   *float64 `json:"max_area,omitempty" example:"10.0"`
}

// GetFarmsByFarmerRequest represents a request to get farms by farmer
type GetFarmsByFarmerRequest struct {
	PaginationRequest
	AAAUserID string `json:"aaa_user_id" validate:"required" example:"usr_123e4567-e89b-12d3-a456-426614174000"`
	AAAOrgID  string `json:"aaa_org_id,omitempty" example:"org_123e4567-e89b-12d3-a456-426614174000"`
}

// GetFarmsByOrgRequest represents a request to get farms by organization
type GetFarmsByOrgRequest struct {
	PaginationRequest
	AAAOrgID string `json:"aaa_org_id,omitempty" example:"org_123e4567-e89b-12d3-a456-426614174000"`
}

// CheckFarmOverlapRequest represents a request to check farm overlap
type CheckFarmOverlapRequest struct {
	BaseRequest
	Geometry string `json:"geometry" validate:"required" example:"POLYGON((75.85 22.71, 75.85663 22.71, 75.85663 22.71663, 75.85 22.71663, 75.85 22.71))"` // WKT format (~50 ha)
	FarmID   string `json:"farm_id,omitempty" example:"farm_123e4567-e89b-12d3-a456-426614174000"`                                                         // Exclude this farm from overlap check
}

// BoundingBox represents a geographic bounding box
type BoundingBox struct {
	MinLat float64 `json:"min_lat" validate:"required,min=-90,max=90" example:"22.71"`
	MaxLat float64 `json:"max_lat" validate:"required,min=-90,max=90" example:"22.73"`
	MinLon float64 `json:"min_lon" validate:"required,min=-180,max=180" example:"75.85"`
	MaxLon float64 `json:"max_lon" validate:"required,min=-180,max=180" example:"75.87"`
}

// GeometryData represents geometric data for farms
type GeometryData struct {
	WKT string `json:"wkt,omitempty" example:"POLYGON((75.85 22.71, 75.85663 22.71, 75.85663 22.71663, 75.85 22.71663, 75.85 22.71))"` // Well-Known Text format (~50 ha)
	WKB []byte `json:"wkb,omitempty"`                                                                                                  // Well-Known Binary format
}

// IrrigationSourceRequest represents irrigation source data in farm requests
type IrrigationSourceRequest struct {
	IrrigationSourceID string  `json:"irrigation_source_id" validate:"required" example:"irr_123e4567-e89b-12d3-a456-426614174000"`
	Count              int     `json:"count,omitempty" validate:"min=0" example:"2"`
	Details            *string `json:"details,omitempty" example:"Bore well depth 150ft"`
	IsPrimary          bool    `json:"is_primary,omitempty" example:"true"`
}

// NewCreateFarmRequest creates a new create farm request
func NewCreateFarmRequest() CreateFarmRequest {
	return CreateFarmRequest{
		BaseRequest: NewBaseRequest(),
		Metadata:    make(map[string]string),
	}
}

// NewUpdateFarmRequest creates a new update farm request
func NewUpdateFarmRequest() UpdateFarmRequest {
	return UpdateFarmRequest{
		BaseRequest: NewBaseRequest(),
		Metadata:    make(map[string]string),
	}
}

// NewDeleteFarmRequest creates a new delete farm request
func NewDeleteFarmRequest() DeleteFarmRequest {
	return DeleteFarmRequest{
		BaseRequest: NewBaseRequest(),
	}
}

// NewGetFarmRequest creates a new get farm request
func NewGetFarmRequest() GetFarmRequest {
	return GetFarmRequest{
		BaseRequest: NewBaseRequest(),
	}
}

// NewListFarmsRequest creates a new list farms request
func NewListFarmsRequest() ListFarmsRequest {
	return ListFarmsRequest{
		FilterRequest: FilterRequest{
			PaginationRequest: NewPaginationRequest(1, 10),
		},
	}
}

// NewGetFarmsByFarmerRequest creates a new get farms by farmer request
func NewGetFarmsByFarmerRequest() GetFarmsByFarmerRequest {
	return GetFarmsByFarmerRequest{
		PaginationRequest: NewPaginationRequest(1, 10),
	}
}

// NewGetFarmsByOrgRequest creates a new get farms by org request
func NewGetFarmsByOrgRequest() GetFarmsByOrgRequest {
	return GetFarmsByOrgRequest{
		PaginationRequest: NewPaginationRequest(1, 10),
	}
}

// NewCheckFarmOverlapRequest creates a new check farm overlap request
func NewCheckFarmOverlapRequest() CheckFarmOverlapRequest {
	return CheckFarmOverlapRequest{
		BaseRequest: NewBaseRequest(),
	}
}
