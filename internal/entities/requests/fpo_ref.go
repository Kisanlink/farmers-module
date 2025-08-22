package requests

// CEOUserData represents CEO user information for FPO creation
type CEOUserData struct {
	FirstName   string `json:"first_name" validate:"required"`
	LastName    string `json:"last_name" validate:"required"`
	PhoneNumber string `json:"phone_number" validate:"required,phone"`
	Email       string `json:"email" validate:"email"`
	Password    string `json:"password" validate:"required,min=8"`
}

// CreateFPORequest represents a request to create an FPO organization
type CreateFPORequest struct {
	BaseRequest
	Name           string            `json:"name" validate:"required"`
	RegistrationNo string            `json:"registration_no" validate:"required"`
	Description    string            `json:"description"`
	CEOUser        CEOUserData       `json:"ceo_user" validate:"required"`
	BusinessConfig map[string]string `json:"business_config"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// NewCreateFPORequest creates a new create FPO request
func NewCreateFPORequest() CreateFPORequest {
	return CreateFPORequest{
		BaseRequest:    NewBaseRequest(),
		BusinessConfig: make(map[string]string),
		Metadata:       make(map[string]string),
	}
}

// RegisterFPORefRequest represents a request to register an FPO reference
type RegisterFPORefRequest struct {
	BaseRequest
	AAAOrgID       string            `json:"aaa_org_id" validate:"required"`
	Name           string            `json:"name" validate:"required"`
	RegistrationNo string            `json:"registration_no"`
	BusinessConfig map[string]string `json:"business_config"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// NewRegisterFPORefRequest creates a new register FPO reference request
func NewRegisterFPORefRequest() RegisterFPORefRequest {
	return RegisterFPORefRequest{
		BaseRequest:    NewBaseRequest(),
		BusinessConfig: make(map[string]string),
		Metadata:       make(map[string]string),
	}
}
