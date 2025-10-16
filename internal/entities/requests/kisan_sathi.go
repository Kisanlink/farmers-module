package requests

// AssignKisanSathiRequest represents a request to assign KisanSathi to farmer
type AssignKisanSathiRequest struct {
	BaseRequest
	AAAUserID        string `json:"aaa_user_id" validate:"required" example:"usr_123e4567-e89b-12d3-a456-426614174000"`
	AAAOrgID         string `json:"aaa_org_id" validate:"required" example:"org_123e4567-e89b-12d3-a456-426614174000"`
	KisanSathiUserID string `json:"kisan_sathi_user_id" validate:"required" example:"ks_123e4567-e89b-12d3-a456-426614174001"`
}

// ReassignKisanSathiRequest represents a request to reassign or remove KisanSathi
type ReassignKisanSathiRequest struct {
	BaseRequest
	AAAUserID           string  `json:"aaa_user_id" validate:"required" example:"usr_123e4567-e89b-12d3-a456-426614174000"`
	AAAOrgID            string  `json:"aaa_org_id" validate:"required" example:"org_123e4567-e89b-12d3-a456-426614174000"`
	NewKisanSathiUserID *string `json:"new_kisan_sathi_user_id,omitempty" example:"ks_123e4567-e89b-12d3-a456-426614174002"` // nil means remove
}

// CreateKisanSathiUserRequest represents a request to create a new KisanSathi user
type CreateKisanSathiUserRequest struct {
	BaseRequest
	Username    string                 `json:"username" validate:"required" example:"pradeep.ks"`
	PhoneNumber string                 `json:"phone_number" validate:"required" example:"+91-9876543211"`
	Email       string                 `json:"email" validate:"email" example:"pradeep.ks@fpo.com"`
	Password    string                 `json:"password" validate:"required,min=8" example:"SecureKS@123"`
	FullName    string                 `json:"full_name" validate:"required" example:"Pradeep Kumar"`
	CountryCode string                 `json:"country_code" example:"+91"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ListKisanSathisRequest represents a request to list all KisanSathis
type ListKisanSathisRequest struct {
	FilterRequest
	Page     int `json:"page,omitempty" example:"1"`
	PageSize int `json:"page_size,omitempty" example:"50"`
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
