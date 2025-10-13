package responses

import (
	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// KisanSathiAssignmentResponse represents a KisanSathi assignment response
type KisanSathiAssignmentResponse struct {
	*base.BaseResponse
	Data *KisanSathiAssignmentData `json:"data"`
}

// KisanSathiAssignmentData represents KisanSathi assignment data in responses
type KisanSathiAssignmentData struct {
	ID               string  `json:"id"`
	AAAUserID        string  `json:"aaa_user_id"`
	AAAOrgID         string  `json:"aaa_org_id"`
	KisanSathiUserID *string `json:"kisan_sathi_user_id,omitempty"`
	Status           string  `json:"status"`
	AssignedAt       string  `json:"assigned_at,omitempty"`
	UnassignedAt     string  `json:"unassigned_at,omitempty"`
	CreatedAt        string  `json:"created_at,omitempty"`
	UpdatedAt        string  `json:"updated_at,omitempty"`
}

// KisanSathiUserResponse represents a KisanSathi user creation response
type KisanSathiUserResponse struct {
	*base.BaseResponse
	Data *KisanSathiUserData `json:"data"`
}

// KisanSathiUserData represents KisanSathi user data in responses
type KisanSathiUserData struct {
	ID          string                 `json:"id"`
	Username    string                 `json:"username"`
	PhoneNumber string                 `json:"phone_number"`
	Email       string                 `json:"email"`
	FullName    string                 `json:"full_name"`
	Role        string                 `json:"role"`
	Status      string                 `json:"status"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   string                 `json:"created_at,omitempty"`
}

// NewKisanSathiAssignmentResponse creates a new KisanSathi assignment response
func NewKisanSathiAssignmentResponse(assignment *KisanSathiAssignmentData, message string) *KisanSathiAssignmentResponse {
	return &KisanSathiAssignmentResponse{
		BaseResponse: base.NewSuccessResponse(message, assignment),
		Data:         assignment,
	}
}

// NewKisanSathiUserResponse creates a new KisanSathi user response
func NewKisanSathiUserResponse(user *KisanSathiUserData, message string) *KisanSathiUserResponse {
	return &KisanSathiUserResponse{
		BaseResponse: base.NewSuccessResponse(message, user),
		Data:         user,
	}
}

// SetRequestID sets the request ID for tracking
func (r *KisanSathiAssignmentResponse) SetRequestID(requestID string) {
	if r.BaseResponse != nil {
		r.BaseResponse.RequestID = requestID
	}
}

// SetRequestID sets the request ID for tracking
func (r *KisanSathiUserResponse) SetRequestID(requestID string) {
	if r.BaseResponse != nil {
		r.BaseResponse.RequestID = requestID
	}
}

// KisanSathiListResponse represents a list of KisanSathis response
type KisanSathiListResponse struct {
	Success   bool              `json:"success"`
	Message   string            `json:"message"`
	Data      []*KisanSathiData `json:"data"`
	Page      int               `json:"page"`
	PageSize  int               `json:"page_size"`
	Total     int64             `json:"total"`
	RequestID string            `json:"request_id,omitempty"`
	Timestamp string            `json:"timestamp,omitempty"`
}

// KisanSathiData represents KisanSathi user data in list responses
type KisanSathiData struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email"`
	FullName    string `json:"full_name"`
	Status      string `json:"status"`
}
