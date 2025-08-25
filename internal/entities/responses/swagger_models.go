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
