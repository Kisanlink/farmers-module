package responses

// SwaggerKisanSathiAssignmentResponse represents a KisanSathi assignment response for Swagger
type SwaggerKisanSathiAssignmentResponse struct {
	Success   bool                      `json:"success"`
	Message   string                    `json:"message"`
	RequestID string                    `json:"request_id"`
	Data      *KisanSathiAssignmentData `json:"data"`
}

// SwaggerKisanSathiUserResponse represents a KisanSathi user creation response for Swagger
type SwaggerKisanSathiUserResponse struct {
	Success   bool                `json:"success"`
	Message   string              `json:"message"`
	RequestID string              `json:"request_id"`
	Data      *KisanSathiUserData `json:"data"`
}

// SwaggerFarmerLinkageResponse represents a farmer linkage response for Swagger
type SwaggerFarmerLinkageResponse struct {
	Success   bool               `json:"success"`
	Message   string             `json:"message"`
	RequestID string             `json:"request_id"`
	Data      *FarmerLinkageData `json:"data"`
}

// SwaggerErrorResponse represents an error response for Swagger
type SwaggerErrorResponse struct {
	Error         string            `json:"error"`
	Message       string            `json:"message"`
	RequestID     string            `json:"request_id"`
	CorrelationID string            `json:"correlation_id"`
	Details       map[string]string `json:"details,omitempty"`
}

// SwaggerCreateFPOResponse represents a FPO creation response for Swagger
type SwaggerCreateFPOResponse struct {
	Success   bool           `json:"success"`
	Message   string         `json:"message"`
	RequestID string         `json:"request_id"`
	Data      *CreateFPOData `json:"data"`
}

// SwaggerFPORefResponse represents a FPO reference response for Swagger
type SwaggerFPORefResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	RequestID string      `json:"request_id"`
	Data      *FPORefData `json:"data"`
}

// SwaggerAdminSeedResponse represents an admin seed response for Swagger
type SwaggerAdminSeedResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Error     string `json:"error,omitempty"`
	Duration  string `json:"duration"`
	Timestamp string `json:"timestamp"`
}

// SwaggerAdminHealthResponse represents an admin health check response for Swagger
type SwaggerAdminHealthResponse struct {
	Status     string                     `json:"status"`
	Message    string                     `json:"message,omitempty"`
	Components map[string]ComponentHealth `json:"components"`
	Duration   string                     `json:"duration"`
	Timestamp  string                     `json:"timestamp"`
}

// SwaggerComponentHealth represents the health status of a system component for Swagger
type SwaggerComponentHealth struct {
	Name      string                 `json:"name"`
	Status    string                 `json:"status"`
	Message   string                 `json:"message,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp string                 `json:"timestamp"`
}

// SwaggerCheckPermissionResponse represents a permission check response for Swagger
type SwaggerCheckPermissionResponse struct {
	Message       string                     `json:"message"`
	Data          SwaggerCheckPermissionData `json:"data"`
	CorrelationID string                     `json:"correlation_id"`
	Timestamp     string                     `json:"timestamp"`
}

// SwaggerCheckPermissionData represents the permission check result data for Swagger
type SwaggerCheckPermissionData struct {
	Subject  string `json:"subject"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
	Object   string `json:"object"`
	OrgID    string `json:"org_id"`
	Allowed  bool   `json:"allowed"`
}

// SwaggerAuditTrailResponse represents an audit trail response for Swagger
type SwaggerAuditTrailResponse struct {
	Message       string                `json:"message"`
	Data          SwaggerAuditTrailData `json:"data"`
	CorrelationID string                `json:"correlation_id"`
	Timestamp     string                `json:"timestamp"`
}

// SwaggerAuditTrailData represents the audit trail data for Swagger
type SwaggerAuditTrailData struct {
	AuditLogs  []interface{}            `json:"audit_logs"`
	Filters    SwaggerAuditTrailFilters `json:"filters"`
	TotalCount int                      `json:"total_count"`
	Page       int                      `json:"page"`
	PageSize   int                      `json:"page_size"`
}

// SwaggerAuditTrailFilters represents the filters applied to audit trail for Swagger
type SwaggerAuditTrailFilters struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	UserID    string `json:"user_id"`
	Action    string `json:"action"`
}

// SwaggerFarmResponse represents a farm response for Swagger
type SwaggerFarmResponse struct {
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
	RequestID string    `json:"request_id"`
	Data      *FarmData `json:"data"`
}

// SwaggerFarmListResponse represents a farm list response for Swagger
type SwaggerFarmListResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	RequestID string      `json:"request_id"`
	Data      []*FarmData `json:"data"`
	Page      int         `json:"page"`
	PageSize  int         `json:"page_size"`
	Total     int         `json:"total"`
}

// SwaggerFarmActivityResponse represents a farm activity response for Swagger
type SwaggerFarmActivityResponse struct {
	Success   bool              `json:"success"`
	Message   string            `json:"message"`
	RequestID string            `json:"request_id"`
	Data      *FarmActivityData `json:"data"`
}

// SwaggerFarmActivityListResponse represents a farm activity list response for Swagger
type SwaggerFarmActivityListResponse struct {
	Success   bool                `json:"success"`
	Message   string              `json:"message"`
	RequestID string              `json:"request_id"`
	Data      []*FarmActivityData `json:"data"`
	Page      int                 `json:"page"`
	PageSize  int                 `json:"page_size"`
	Total     int                 `json:"total"`
}

// SwaggerCropCycleResponse represents a crop cycle response for Swagger
type SwaggerCropCycleResponse struct {
	Success   bool           `json:"success"`
	Message   string         `json:"message"`
	RequestID string         `json:"request_id"`
	Data      *CropCycleData `json:"data"`
}

// SwaggerCropCycleListResponse represents a crop cycle list response for Swagger
type SwaggerCropCycleListResponse struct {
	Success   bool             `json:"success"`
	Message   string           `json:"message"`
	RequestID string           `json:"request_id"`
	Data      []*CropCycleData `json:"data"`
	Page      int              `json:"page"`
	PageSize  int              `json:"page_size"`
	Total     int              `json:"total"`
}

// SwaggerExportFarmerPortfolioResponse represents an export farmer portfolio response for Swagger
type SwaggerExportFarmerPortfolioResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	RequestID string      `json:"request_id"`
	Data      interface{} `json:"data"`
}

// SwaggerOrgDashboardCountersResponse represents an org dashboard counters response for Swagger
type SwaggerOrgDashboardCountersResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	RequestID string      `json:"request_id"`
	Data      interface{} `json:"data"`
}

// SwaggerKisanSathiListResponse represents a KisanSathi list response for Swagger
type SwaggerKisanSathiListResponse struct {
	Success   bool              `json:"success" example:"true"`
	Message   string            `json:"message" example:"KisanSathis retrieved successfully"`
	Data      []*KisanSathiData `json:"data"`
	Page      int               `json:"page" example:"1"`
	PageSize  int               `json:"page_size" example:"50"`
	Total     int64             `json:"total" example:"100"`
	RequestID string            `json:"request_id,omitempty" example:"req_123456789"`
	Timestamp string            `json:"timestamp,omitempty" example:"2024-01-15T10:30:00Z"`
}

// SwaggerValidateGeometryResponse represents a geometry validation response for Swagger
type SwaggerValidateGeometryResponse struct {
	Success   bool     `json:"success" example:"true"`
	Message   string   `json:"message" example:"Geometry validated successfully"`
	RequestID string   `json:"request_id,omitempty" example:"req_123456789"`
	WKT       string   `json:"wkt" example:"POLYGON((...))"`
	IsValid   bool     `json:"is_valid" example:"true"`
	Errors    []string `json:"errors,omitempty"`
	Warnings  []string `json:"warnings,omitempty"`
	SRID      int      `json:"srid" example:"4326"`
	AreaHa    *float64 `json:"area_ha,omitempty" example:"2.5"`
}

// SwaggerReconcileAAALinksResponse represents an AAA links reconciliation response for Swagger
type SwaggerReconcileAAALinksResponse struct {
	Success        bool     `json:"success" example:"true"`
	Message        string   `json:"message" example:"AAA links reconciled successfully"`
	RequestID      string   `json:"request_id,omitempty" example:"req_123456789"`
	ProcessedLinks int      `json:"processed_links" example:"150"`
	FixedLinks     int      `json:"fixed_links" example:"10"`
	BrokenLinks    int      `json:"broken_links" example:"5"`
	Errors         []string `json:"errors,omitempty"`
}

// SwaggerRebuildSpatialIndexesResponse represents a spatial indexes rebuild response for Swagger
type SwaggerRebuildSpatialIndexesResponse struct {
	Success        bool     `json:"success" example:"true"`
	Message        string   `json:"message" example:"Spatial indexes rebuilt successfully"`
	RequestID      string   `json:"request_id,omitempty" example:"req_123456789"`
	RebuiltIndexes []string `json:"rebuilt_indexes" example:"idx_farms_boundary,idx_plots_location"`
	Errors         []string `json:"errors,omitempty"`
}

// SwaggerDetectFarmOverlapsResponse represents a farm overlaps detection response for Swagger
type SwaggerDetectFarmOverlapsResponse struct {
	Success       bool          `json:"success" example:"true"`
	Message       string        `json:"message" example:"Farm overlaps detected successfully"`
	RequestID     string        `json:"request_id,omitempty" example:"req_123456789"`
	Overlaps      []FarmOverlap `json:"overlaps"`
	TotalOverlaps int           `json:"total_overlaps" example:"3"`
}

// SwaggerSoilTypesResponse represents a soil types lookup response for Swagger
type SwaggerSoilTypesResponse struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message" example:"Soil types retrieved successfully"`
	Data    interface{} `json:"data"`
}

// SwaggerIrrigationSourcesResponse represents an irrigation sources lookup response for Swagger
type SwaggerIrrigationSourcesResponse struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message" example:"Irrigation sources retrieved successfully"`
	Data    interface{} `json:"data"`
}

// SwaggerBaseResponse represents a generic success response for Swagger
type SwaggerBaseResponse struct {
	Success   bool        `json:"success" example:"true"`
	Message   string      `json:"message" example:"Operation completed successfully"`
	RequestID string      `json:"request_id,omitempty" example:"req_123456789"`
	Data      interface{} `json:"data,omitempty"`
}

// SwaggerCropVarietyListResponse represents a crop variety list response for Swagger
type SwaggerCropVarietyListResponse struct {
	Success   bool               `json:"success" example:"true"`
	Message   string             `json:"message" example:"Crop varieties retrieved successfully"`
	RequestID string             `json:"request_id" example:"req_123456789"`
	Data      []*CropVarietyData `json:"data"`
	Page      int                `json:"page" example:"1"`
	PageSize  int                `json:"page_size" example:"20"`
	Total     int                `json:"total" example:"50"`
}

// SwaggerCropVarietyLookupResponse represents a crop variety lookup response for Swagger
type SwaggerCropVarietyLookupResponse struct {
	Success   bool                     `json:"success" example:"true"`
	Message   string                   `json:"message" example:"Crop variety lookup data retrieved successfully"`
	RequestID string                   `json:"request_id" example:"req_123456789"`
	Data      []*CropVarietyLookupData `json:"data"`
}

// SwaggerFarmerResponse represents a farmer response for Swagger
type SwaggerFarmerResponse struct {
	Success   bool               `json:"success" example:"true"`
	Message   string             `json:"message" example:"Farmer created successfully"`
	RequestID string             `json:"request_id" example:"req_123456789"`
	Data      *FarmerProfileData `json:"data"`
}

// SwaggerFarmerListResponse represents a farmer list response for Swagger
type SwaggerFarmerListResponse struct {
	Success   bool                 `json:"success" example:"true"`
	Message   string               `json:"message" example:"Farmers retrieved successfully"`
	RequestID string               `json:"request_id" example:"req_123456789"`
	Data      []*FarmerProfileData `json:"data"`
	Page      int                  `json:"page" example:"1"`
	PageSize  int                  `json:"page_size" example:"10"`
	Total     int                  `json:"total" example:"100"`
}

// SwaggerStageResponse represents a stage response for Swagger
type SwaggerStageResponse struct {
	Success   bool       `json:"success" example:"true"`
	Message   string     `json:"message" example:"Stage created successfully"`
	RequestID string     `json:"request_id" example:"req_123456789"`
	Data      *StageData `json:"data"`
}

// SwaggerStageListResponse represents a stage list response for Swagger
type SwaggerStageListResponse struct {
	Success   bool         `json:"success" example:"true"`
	Message   string       `json:"message" example:"Stages retrieved successfully"`
	RequestID string       `json:"request_id" example:"req_123456789"`
	Data      []*StageData `json:"data"`
	Page      int          `json:"page" example:"1"`
	PageSize  int          `json:"page_size" example:"20"`
	Total     int          `json:"total" example:"50"`
}

// SwaggerCropStageResponse represents a crop stage response for Swagger
type SwaggerCropStageResponse struct {
	Success   bool           `json:"success" example:"true"`
	Message   string         `json:"message" example:"Stage assigned to crop successfully"`
	RequestID string         `json:"request_id" example:"req_123456789"`
	Data      *CropStageData `json:"data"`
}

// SwaggerCropStagesResponse represents a crop stages list response for Swagger
type SwaggerCropStagesResponse struct {
	Success   bool             `json:"success" example:"true"`
	Message   string           `json:"message" example:"Crop stages retrieved successfully"`
	RequestID string           `json:"request_id" example:"req_123456789"`
	Data      []*CropStageData `json:"data"`
}

// SwaggerStageLookupResponse represents a stage lookup response for Swagger
type SwaggerStageLookupResponse struct {
	Success   bool               `json:"success" example:"true"`
	Message   string             `json:"message" example:"Stage lookup data retrieved successfully"`
	RequestID string             `json:"request_id" example:"req_123456789"`
	Data      []*StageLookupData `json:"data"`
}

// SwaggerAreaAllocationSummaryResponse represents an area allocation summary response for Swagger
type SwaggerAreaAllocationSummaryResponse struct {
	Success   bool                       `json:"success" example:"true"`
	Message   string                     `json:"message" example:"Area allocation summary retrieved successfully"`
	RequestID string                     `json:"request_id" example:"req_123456789"`
	Data      *AreaAllocationSummaryData `json:"data"`
}
