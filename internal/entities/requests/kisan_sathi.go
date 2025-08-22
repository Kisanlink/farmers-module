package requests

// AssignKisanSathiRequest represents a request to assign KisanSathi to farmer
type AssignKisanSathiRequest struct {
	BaseRequest
	AAAUserID        string `json:"aaa_user_id" validate:"required"`
	AAAOrgID         string `json:"aaa_org_id" validate:"required"`
	KisanSathiUserID string `json:"kisan_sathi_user_id" validate:"required"`
}

// ReassignKisanSathiRequest represents a request to reassign or remove KisanSathi
type ReassignKisanSathiRequest struct {
	BaseRequest
	AAAUserID           string  `json:"aaa_user_id" validate:"required"`
	AAAOrgID            string  `json:"aaa_org_id" validate:"required"`
	NewKisanSathiUserID *string `json:"new_kisan_sathi_user_id,omitempty"` // nil means remove
}

// CreateKisanSathiUserRequest represents a request to create a new KisanSathi user
type CreateKisanSathiUserRequest struct {
	BaseRequest
	Username    string            `json:"username" validate:"required"`
	PhoneNumber string            `json:"phone_number" validate:"required"`
	Email       string            `json:"email" validate:"email"`
	Password    string            `json:"password" validate:"required,min=8"`
	FullName    string            `json:"full_name" validate:"required"`
	CountryCode string            `json:"country_code"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// NewAssignKisanSathiRequest creates a new assign KisanSathi request
func NewAssignKisanSathiRequest() AssignKisanSathiRequest {
	return AssignKisanSathiRequest{
		BaseRequest: NewBaseRequest(),
	}
}

// NewReassignKisanSathiRequest creates a new reassign KisanSathi request
func NewReassignKisanSathiRequest() ReassignKisanSathiRequest {
	return ReassignKisanSathiRequest{
		BaseRequest: NewBaseRequest(),
	}
}

// NewCreateKisanSathiUserRequest creates a new create KisanSathi user request
func NewCreateKisanSathiUserRequest() CreateKisanSathiUserRequest {
	return CreateKisanSathiUserRequest{
		BaseRequest: NewBaseRequest(),
	}
}
