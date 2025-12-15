package responses

import "time"

// UserGroupData represents user group information in responses
type UserGroupData struct {
	GroupID     string   `json:"group_id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	OrgID       string   `json:"org_id"`
	Permissions []string `json:"permissions"`
	CreatedAt   string   `json:"created_at"`
}

// CreateFPOResponse represents the response for FPO creation
type CreateFPOResponse struct {
	*BaseResponse `json:",inline"`
	Data          *CreateFPOData `json:"data"`
}

// CreateFPOData represents FPO creation data in responses
type CreateFPOData struct {
	FPOID      string          `json:"fpo_id"`
	AAAOrgID   string          `json:"aaa_org_id"`
	Name       string          `json:"name"`
	CEOUserID  string          `json:"ceo_user_id"`
	UserGroups []UserGroupData `json:"user_groups"`
	Status     string          `json:"status"`
	CreatedAt  time.Time       `json:"created_at"`
}

// NewCreateFPOResponse creates a new FPO creation response
func NewCreateFPOResponse(fpoData *CreateFPOData, message string) CreateFPOResponse {
	baseResp := NewSuccessResponse(message, fpoData)
	return CreateFPOResponse{
		BaseResponse: baseResp,
		Data:         fpoData,
	}
}

// FPORefResponse represents a single FPO reference response
type FPORefResponse struct {
	*BaseResponse `json:",inline"`
	Data          *FPORefData `json:"data"`
}

// FPORefData represents FPO reference data in responses
type FPORefData struct {
	ID             string                 `json:"id"`
	AAAOrgID       string                 `json:"aaa_org_id"`
	Name           string                 `json:"name"`
	RegistrationNo string                 `json:"registration_number"`
	BusinessConfig map[string]interface{} `json:"business_config"`
	Status         string                 `json:"status"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedBy      string                 `json:"created_by,omitempty"`
	CreatedAt      string                 `json:"created_at,omitempty"`
	UpdatedAt      string                 `json:"updated_at,omitempty"`
}

// NewFPORefResponse creates a new FPO reference response
func NewFPORefResponse(fpoRef *FPORefData, message string) FPORefResponse {
	baseResp := NewSuccessResponse(message, fpoRef)
	return FPORefResponse{
		BaseResponse: baseResp,
		Data:         fpoRef,
	}
}

// SetRequestID sets the request ID for tracking
func (r *FPORefResponse) SetRequestID(requestID string) {
	r.BaseResponse.RequestID = requestID
}

// SetRequestID sets the request ID for tracking
func (r *CreateFPOResponse) SetRequestID(requestID string) {
	r.BaseResponse.RequestID = requestID
}

// UpdateCEOResponse represents the response for CEO update
type UpdateCEOResponse struct {
	*BaseResponse `json:",inline"`
	Data          *UpdateCEOData `json:"data"`
}

// UpdateCEOData represents CEO update data in responses
type UpdateCEOData struct {
	AAAOrgID     string    `json:"aaa_org_id"`
	OrgName      string    `json:"org_name"`
	NewCEOUserID string    `json:"new_ceo_user_id"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// NewUpdateCEOResponse creates a new CEO update response
func NewUpdateCEOResponse(data *UpdateCEOData, message string) UpdateCEOResponse {
	baseResp := NewSuccessResponse(message, data)
	return UpdateCEOResponse{
		BaseResponse: baseResp,
		Data:         data,
	}
}

// SetRequestID sets the request ID for tracking
func (r *UpdateCEOResponse) SetRequestID(requestID string) {
	r.BaseResponse.RequestID = requestID
}
