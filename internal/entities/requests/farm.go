package requests

// CreateFarmRequest represents a request to create a new farm
type CreateFarmRequest struct {
	BaseRequest
	AAAFarmerUserID string            `json:"aaa_farmer_user_id" validate:"required"`
	AAAOrgID        string            `json:"aaa_org_id" validate:"required"`
	AreaHa          float64           `json:"area_ha" validate:"required,min=0.01"`
	Geometry        GeometryData      `json:"geometry,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

// UpdateFarmRequest represents a request to update an existing farm
type UpdateFarmRequest struct {
	BaseRequest
	ID              string            `json:"id" validate:"required"`
	AAAFarmerUserID string            `json:"aaa_farmer_user_id,omitempty"`
	AAAOrgID        string            `json:"aaa_org_id,omitempty"`
	AreaHa          *float64          `json:"area_ha,omitempty"`
	Geometry        *GeometryData     `json:"geometry,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

// DeleteFarmRequest represents a request to delete a farm
type DeleteFarmRequest struct {
	BaseRequest
	ID string `json:"id" validate:"required"`
}

// GetFarmRequest represents a request to retrieve a farm
type GetFarmRequest struct {
	BaseRequest
	ID string `json:"id" validate:"required"`
}

// ListFarmsRequest represents a request to list farms with filtering
type ListFarmsRequest struct {
	FilterRequest
	AAAFarmerUserID string   `json:"aaa_farmer_user_id,omitempty"`
	AAAOrgID        string   `json:"aaa_org_id,omitempty"`
	MinArea         *float64 `json:"min_area,omitempty"`
	MaxArea         *float64 `json:"max_area,omitempty"`
}

// GetFarmsByFarmerRequest represents a request to get farms by farmer
type GetFarmsByFarmerRequest struct {
	PaginationRequest
	AAAFarmerUserID string `json:"aaa_farmer_user_id" validate:"required"`
	AAAOrgID        string `json:"aaa_org_id,omitempty"`
}

// GetFarmsByOrgRequest represents a request to get farms by organization
type GetFarmsByOrgRequest struct {
	PaginationRequest
	AAAOrgID string `json:"aaa_org_id,omitempty"`
}

// CheckFarmOverlapRequest represents a request to check farm overlap
type CheckFarmOverlapRequest struct {
	BaseRequest
	Geometry string `json:"geometry" validate:"required"` // WKT format
	FarmID   string `json:"farm_id,omitempty"`            // Exclude this farm from overlap check
}

// BoundingBox represents a geographic bounding box
type BoundingBox struct {
	MinLat float64 `json:"min_lat" validate:"required,min=-90,max=90"`
	MaxLat float64 `json:"max_lat" validate:"required,min=-90,max=90"`
	MinLon float64 `json:"min_lon" validate:"required,min=-180,max=180"`
	MaxLon float64 `json:"max_lon" validate:"required,min=-180,max=180"`
}

// GeometryData represents geometric data for farms
type GeometryData struct {
	WKT string `json:"wkt,omitempty"` // Well-Known Text format
	WKB []byte `json:"wkb,omitempty"` // Well-Known Binary format
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
