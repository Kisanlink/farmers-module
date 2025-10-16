package responses

import (
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// FarmResponse represents a single farm response
type FarmResponse struct {
	*base.BaseResponse `json:",inline"`
	Data               *FarmData `json:"data"`
}

// FarmListResponse represents a list of farms response
type FarmListResponse struct {
	*base.PaginatedResponse `json:",inline"`
	Data                    []*FarmData `json:"data"`
}

// FarmOverlapResponse represents a farm overlap check response
type FarmOverlapResponse struct {
	*base.BaseResponse `json:",inline"`
	Data               *FarmOverlapData `json:"data"`
}

// FarmData represents farm data in responses
type FarmData struct {
	ID                        string                 `json:"id" example:"FARM00000001"`
	FarmerID                  string                 `json:"farmer_id" example:"FMRR0000000001"`
	AAAUserID                 string                 `json:"aaa_user_id" example:"usr_123e4567-e89b-12d3-a456-426614174000"`
	AAAOrgID                  string                 `json:"aaa_org_id" example:"org_123e4567-e89b-12d3-a456-426614174000"`
	Name                      string                 `json:"name" example:"North Field Farm"`
	OwnershipType             string                 `json:"ownership_type" example:"OWN"`
	Geometry                  string                 `json:"geometry" example:"POLYGON((75.85 22.71, 75.85663 22.71, 75.85663 22.71663, 75.85 22.71663, 75.85 22.71))"`
	AreaHa                    float64                `json:"area_ha" example:"2.5"`             // User-provided or stored area
	AreaHaComputed            float64                `json:"area_ha_computed" example:"2.4876"` // Computed from geometry using PostGIS
	SoilTypeID                *string                `json:"soil_type_id,omitempty"`
	PrimaryIrrigationSourceID *string                `json:"primary_irrigation_source_id,omitempty"`
	BoreWellCount             int                    `json:"bore_well_count" example:"2"`
	OtherIrrigationDetails    *string                `json:"other_irrigation_details,omitempty"`
	Metadata                  map[string]interface{} `json:"metadata"`
	CreatedAt                 time.Time              `json:"created_at" example:"2024-01-15T10:30:00Z"`
	UpdatedAt                 time.Time              `json:"updated_at" example:"2024-01-20T15:45:00Z"`
	Farmer                    *FarmerBasicData       `json:"farmer,omitempty"`
	SoilType                  *SoilTypeData          `json:"soil_type,omitempty"`
	PrimaryIrrigationSource   *IrrigationSourceData  `json:"primary_irrigation_source,omitempty"`
	IrrigationSources         []IrrigationSourceData `json:"irrigation_sources,omitempty"`
	SoilTypes                 []SoilTypeData         `json:"soil_types,omitempty"`
}

// SoilTypeData represents soil type data in farm responses
type SoilTypeData struct {
	ID          string    `json:"id" example:"SOIL00000001"`
	Name        string    `json:"name" example:"Loamy Soil"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// IrrigationSourceData represents irrigation source data in farm responses
type IrrigationSourceData struct {
	ID          string    `json:"id" example:"IRRG00000001"`
	Name        string    `json:"name" example:"Bore Well"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// FarmerBasicData represents basic farmer information in farm responses
type FarmerBasicData struct {
	ID             string    `json:"id" example:"FMRR0000000001"`
	AAAUserID      string    `json:"aaa_user_id" example:"USER00000001"`
	AAAOrgID       string    `json:"aaa_org_id" example:"ORGN00000001"`
	FirstName      string    `json:"first_name,omitempty" example:"Ramesh"`
	LastName       string    `json:"last_name,omitempty" example:"Kumar"`
	PhoneNumber    string    `json:"phone_number,omitempty" example:"9876543210"`
	Email          string    `json:"email,omitempty" example:"ramesh.kumar@example.com"`
	TotalAcreageHa float64   `json:"total_acreage_ha" example:"15.75"`
	Status         string    `json:"status" example:"ACTIVE"`
	CreatedAt      time.Time `json:"created_at,omitempty"`
	UpdatedAt      time.Time `json:"updated_at,omitempty"`
}

// FarmOverlapData represents farm overlap check result
type FarmOverlapData struct {
	HasOverlap       bool     `json:"has_overlap" example:"true"`
	OverlapArea      float64  `json:"overlap_area,omitempty" example:"0.25"`
	OverlappingFarms []string `json:"overlapping_farms,omitempty" example:"farm_123e4567-e89b-12d3-a456-426614174001,farm_123e4567-e89b-12d3-a456-426614174002"`
	Message          string   `json:"message,omitempty" example:"Farm boundary overlaps with 2 existing farms"`
}

// NewFarmResponse creates a new farm response
func NewFarmResponse(farm *FarmData, message string) FarmResponse {
	return FarmResponse{
		BaseResponse: base.NewSuccessResponse(message, farm),
		Data:         farm,
	}
}

// NewFarmListResponse creates a new farm list response
func NewFarmListResponse(farms []*FarmData, page, pageSize int, totalCount int64) FarmListResponse {
	// Convert to interface slice for pagination
	var data []interface{}
	for _, f := range farms {
		data = append(data, f)
	}

	paginationInfo := base.NewPaginationInfo(page, pageSize, int(totalCount))
	return FarmListResponse{
		PaginatedResponse: base.NewPaginatedResponse("Farms retrieved successfully", data, paginationInfo),
		Data:              farms,
	}
}

// NewFarmOverlapResponse creates a new farm overlap response
func NewFarmOverlapResponse(overlap *FarmOverlapData, message string) FarmOverlapResponse {
	return FarmOverlapResponse{
		BaseResponse: base.NewSuccessResponse(message, overlap),
		Data:         overlap,
	}
}

// SetRequestID sets the request ID for tracking
func (r *FarmResponse) SetRequestID(requestID string) {
	r.BaseResponse.RequestID = requestID
}

// SetRequestID sets the request ID for tracking
func (r *FarmListResponse) SetRequestID(requestID string) {
	r.PaginatedResponse.RequestID = requestID
}

// SetRequestID sets the request ID for tracking
func (r *FarmOverlapResponse) SetRequestID(requestID string) {
	r.BaseResponse.RequestID = requestID
}
