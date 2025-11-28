package requests

// CEOUserData represents CEO user information for FPO creation
// Note: first_name, last_name, and password are only required when creating a new user.
// If the user already exists in AAA (by phone_number), these fields are optional.
type CEOUserData struct {
	FirstName   string `json:"first_name,omitempty" example:"Rajesh"`
	LastName    string `json:"last_name,omitempty" example:"Sharma"`
	PhoneNumber string `json:"phone_number" validate:"required" example:"+91-9876543210"`
	Email       string `json:"email,omitempty" validate:"omitempty,email" example:"rajesh.sharma@fpo.com"`
	Password    string `json:"password,omitempty" validate:"omitempty,min=8" example:"SecurePass@123"`
}

// CreateFPORequest represents a request to create an FPO organization
type CreateFPORequest struct {
	BaseRequest
	Name           string                 `json:"name" validate:"required" example:"Rampur Farmers Producer Company"`
	RegistrationNo string                 `json:"registration_number" validate:"required" example:"FPO/MP/2024/001234"`
	Description    string                 `json:"description" example:"A farmer producer organization serving 500+ farmers in Rampur region"`
	CEOUser        CEOUserData            `json:"ceo_user" validate:"required"`
	BusinessConfig map[string]interface{} `json:"business_config" example:"max_farmers:1000,procurement_enabled:true"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// NewCreateFPORequest creates a new create FPO request
func NewCreateFPORequest() CreateFPORequest {
	return CreateFPORequest{
		BaseRequest:    NewBaseRequest(),
		BusinessConfig: make(map[string]interface{}),
		Metadata:       make(map[string]interface{}),
	}
}

// RegisterFPORefRequest represents a request to register an FPO reference
type RegisterFPORefRequest struct {
	BaseRequest
	AAAOrgID       string                 `json:"aaa_org_id" validate:"required" example:"org_123e4567-e89b-12d3-a456-426614174000"`
	Name           string                 `json:"name" validate:"required" example:"Rampur Farmers Producer Company"`
	RegistrationNo string                 `json:"registration_number" example:"FPO/MP/2024/001234"`
	BusinessConfig map[string]interface{} `json:"business_config" example:"credit_limit:500000,payment_terms:net30"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// NewRegisterFPORefRequest creates a new register FPO reference request
func NewRegisterFPORefRequest() RegisterFPORefRequest {
	return RegisterFPORefRequest{
		BaseRequest:    NewBaseRequest(),
		BusinessConfig: make(map[string]interface{}),
		Metadata:       make(map[string]interface{}),
	}
}

// SyncFPORequest represents a request to sync FPO from AAA
type SyncFPORequest struct {
	BaseRequest
	AAAOrgID string `json:"aaa_org_id" validate:"required" example:"org_123e4567-e89b-12d3-a456-426614174000"`
}

// RetrySetupRequest represents a request to retry failed FPO setup
type RetrySetupRequest struct {
	BaseRequest
	FPOID string `json:"fpo_id" validate:"required" example:"FPOR_1234567890"`
}

// SuspendFPORequest represents a request to suspend an FPO
type SuspendFPORequest struct {
	BaseRequest
	Reason string `json:"reason" validate:"required" example:"Compliance violation"`
}

// DeactivateFPORequest represents a request to deactivate an FPO
type DeactivateFPORequest struct {
	BaseRequest
	Reason string `json:"reason" validate:"required" example:"Business closure"`
}
