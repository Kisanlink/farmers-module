package requests

// RegisterFPORefRequest represents a request to register an FPO reference
type RegisterFPORefRequest struct {
	BaseRequest
	AAAOrgID       string            `json:"aaa_org_id" validate:"required"`
	BusinessConfig string            `json:"business_config" validate:"required"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// NewRegisterFPORefRequest creates a new register FPO reference request
func NewRegisterFPORefRequest() RegisterFPORefRequest {
	return RegisterFPORefRequest{
		BaseRequest: NewBaseRequest(),
		Metadata:    make(map[string]string),
	}
}
