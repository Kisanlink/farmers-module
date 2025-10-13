package requests

// CEOUserData represents CEO user information for FPO creation
type CEOUserData struct {
	FirstName   string `json:"first_name" validate:"required" example:"Rajesh"`
	LastName    string `json:"last_name" validate:"required" example:"Sharma"`
	PhoneNumber string `json:"phone_number" validate:"required,phone" example:"+91-9876543210"`
	Email       string `json:"email" validate:"email" example:"rajesh.sharma@fpo.com"`
	Password    string `json:"password" validate:"required,min=8" example:"SecurePass@123"`
}

// CreateFPORequest represents a request to create an FPO organization
type CreateFPORequest struct {
	BaseRequest
	Name           string                 `json:"name" validate:"required" example:"Rampur Farmers Producer Company"`
	RegistrationNo string                 `json:"registration_no" validate:"required" example:"FPO/MP/2024/001234"`
	Description    string                 `json:"description" example:"A farmer producer organization serving 500+ farmers in Rampur region"`
	CEOUser        CEOUserData            `json:"ceo_user" validate:"required"`
	BusinessConfig map[string]interface{} `json:"business_config" example:"max_farmers:1000,procurement_enabled:true"`
	Metadata       map[string]interface{} `json:"metadata,omitempty" example:"district:Indore,state:MP,established:2024"`
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
	RegistrationNo string                 `json:"registration_no" example:"FPO/MP/2024/001234"`
	BusinessConfig map[string]interface{} `json:"business_config" example:"credit_limit:500000,payment_terms:net30"`
	Metadata       map[string]interface{} `json:"metadata,omitempty" example:"region:central_india,crop_focus:wheat_soybean"`
}

// NewRegisterFPORefRequest creates a new register FPO reference request
func NewRegisterFPORefRequest() RegisterFPORefRequest {
	return RegisterFPORefRequest{
		BaseRequest:    NewBaseRequest(),
		BusinessConfig: make(map[string]interface{}),
		Metadata:       make(map[string]interface{}),
	}
}
