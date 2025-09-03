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

// SwaggerFarmerResponse represents a farmer response for Swagger
type SwaggerFarmerResponse struct {
	Success   bool               `json:"success"`
	Message   string             `json:"message"`
	RequestID string             `json:"request_id"`
	Data      *FarmerProfileData `json:"data"`
}

// SwaggerFarmerListResponse represents a farmer list response for Swagger
type SwaggerFarmerListResponse struct {
	Success   bool                 `json:"success"`
	Message   string               `json:"message"`
	RequestID string               `json:"request_id"`
	Data      []*FarmerProfileData `json:"data"`
	Page      int                  `json:"page"`
	PageSize  int                  `json:"page_size"`
	Total     int                  `json:"total"`
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
