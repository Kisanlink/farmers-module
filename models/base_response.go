package models

type Response struct {
	StatusCode int    `json:"status_code"`
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	Data       any    `json:"data"`
	Error      any    `json:"error"`
	TimeStamp  string `json:"timestamp"`
}

// BatchRequest represents a request with multiple farm IDs
type BatchRequest struct {
	FarmIDs []string               `json:"farm_ids" binding:"required,min=1,max=50"`
	Filters map[string]interface{} `json:"filters,omitempty"`
}

// BatchResponse represents a response with data organized by farm ID
type BatchResponse struct {
	StatusCode int                    `json:"status_code"`
	Success    bool                   `json:"success"`
	Message    string                 `json:"message"`
	Data       map[string]interface{} `json:"data"`
	Errors     map[string]string      `json:"errors,omitempty"`
	TimeStamp  string                 `json:"timestamp"`
}
