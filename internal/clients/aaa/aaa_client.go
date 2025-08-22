package aaa

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Kisanlink/farmers-module/internal/config"
	"github.com/Kisanlink/farmers-module/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Client represents the AAA gRPC client
type Client struct {
	conn        *grpc.ClientConn
	config      *config.Config
	userClient  proto.UserServiceV2Client
	authzClient proto.AuthorizationServiceClient
}

// UserData represents user information from AAA service
type UserData struct {
	ID          string            `json:"id"`
	Username    string            `json:"username"`
	PhoneNumber string            `json:"phone_number"`
	CountryCode string            `json:"country_code"`
	Email       string            `json:"email"`
	FullName    string            `json:"full_name"`
	IsValidated bool              `json:"is_validated"`
	Status      string            `json:"status"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
	Metadata    map[string]string `json:"metadata"`
}

// OrganizationData represents organization information
type OrganizationData struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Type        string            `json:"type"`
	Status      string            `json:"status"`
	Metadata    map[string]string `json:"metadata"`
}

// UserGroupData represents user group information
type UserGroupData struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	OrgID       string   `json:"org_id"`
	Permissions []string `json:"permissions"`
}

// CreateUserRequest represents a user creation request
type CreateUserRequest struct {
	Username    string            `json:"username"`
	PhoneNumber string            `json:"phone_number"`
	CountryCode string            `json:"country_code"`
	Email       string            `json:"email"`
	Password    string            `json:"password"`
	FullName    string            `json:"full_name"`
	Role        string            `json:"role"`
	Metadata    map[string]string `json:"metadata"`
}

// CreateUserResponse represents a user creation response
type CreateUserResponse struct {
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateOrganizationRequest represents an organization creation request
type CreateOrganizationRequest struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Type        string            `json:"type"`
	CEOUserID   string            `json:"ceo_user_id"`
	Metadata    map[string]string `json:"metadata"`
}

// CreateOrganizationResponse represents an organization creation response
type CreateOrganizationResponse struct {
	OrgID     string    `json:"org_id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateUserGroupRequest represents a user group creation request
type CreateUserGroupRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	OrgID       string   `json:"org_id"`
	Permissions []string `json:"permissions"`
}

// CreateUserGroupResponse represents a user group creation response
type CreateUserGroupResponse struct {
	GroupID   string    `json:"group_id"`
	Name      string    `json:"name"`
	OrgID     string    `json:"org_id"`
	CreatedAt time.Time `json:"created_at"`
}

// NewClient creates a new AAA gRPC client
func NewClient(cfg *config.Config) (*Client, error) {
	log.Printf("Connecting to AAA service at %s", cfg.AAA.GRPCEndpoint)

	conn, err := grpc.NewClient(cfg.AAA.GRPCEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to AAA service: %w", err)
	}

	log.Printf("Successfully connected to AAA service at %s", cfg.AAA.GRPCEndpoint)

	// Initialize gRPC clients
	userClient := proto.NewUserServiceV2Client(conn)
	authzClient := proto.NewAuthorizationServiceClient(conn)

	return &Client{
		conn:        conn,
		config:      cfg,
		userClient:  userClient,
		authzClient: authzClient,
	}, nil
}

// Close closes the gRPC connection
func (c *Client) Close() error {
	return c.conn.Close()
}

// CreateUser creates a user in the AAA service
func (c *Client) CreateUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error) {
	log.Printf("AAA CreateUser: username=%s, phone=%s", req.Username, req.PhoneNumber)

	// Create gRPC request
	grpcReq := &proto.RegisterRequestV2{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	}

	// Call the AAA service
	response, err := c.userClient.Register(ctx, grpcReq)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.AlreadyExists:
				return nil, fmt.Errorf("user already exists")
			case codes.InvalidArgument:
				return nil, fmt.Errorf("invalid request: %s", st.Message())
			default:
				return nil, fmt.Errorf("failed to create user: %s", st.Message())
			}
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if response.StatusCode != 201 {
		return nil, fmt.Errorf("unexpected status code: %d - %s", response.StatusCode, response.Message)
	}

	log.Printf("User created successfully with ID: %s", response.User.Id)

	return &CreateUserResponse{
		UserID:    response.User.Id,
		Username:  response.User.Username,
		Status:    response.User.Status,
		CreatedAt: time.Now(),
	}, nil
}

// GetUser retrieves a user from the AAA service
func (c *Client) GetUser(ctx context.Context, userID string) (*UserData, error) {
	log.Printf("AAA GetUser: userID=%s", userID)

	// Create gRPC request
	req := &proto.GetUserRequestV2{
		Id: userID,
	}

	// Call the AAA service
	response, err := c.userClient.GetUser(ctx, req)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return nil, fmt.Errorf("user not found")
			default:
				return nil, fmt.Errorf("failed to get user: %s", st.Message())
			}
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d - %s", response.StatusCode, response.Message)
	}

	// Convert protobuf response to UserData
	userData := &UserData{
		ID:          response.User.Id,
		Username:    response.User.Username,
		PhoneNumber: response.User.PhoneNumber,
		CountryCode: response.User.CountryCode,
		Email:       response.User.Email,
		FullName:    response.User.FullName,
		IsValidated: response.User.IsValidated,
		Status:      response.User.Status,
		CreatedAt:   response.User.CreatedAt,
		UpdatedAt:   response.User.UpdatedAt,
	}

	log.Printf("User data retrieved successfully for ID: %s", userID)
	return userData, nil
}

// GetUserByPhone retrieves a user from the AAA service by phone number
func (c *Client) GetUserByPhone(ctx context.Context, phone string) (*UserData, error) {
	log.Printf("AAA GetUserByPhone: phone=%s", phone)

	if phone == "" {
		return nil, fmt.Errorf("phone number is required")
	}

	// Create gRPC request to get user by phone number
	req := &proto.GetUserByPhoneRequestV2{
		PhoneNumber:        phone,
		CountryCode:        "+91", // Default to India, should be configurable
		IncludeRoles:       false,
		IncludePermissions: false,
	}

	// Call the AAA service
	response, err := c.userClient.GetUserByPhone(ctx, req)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return nil, fmt.Errorf("user not found with phone number: %s", phone)
			default:
				return nil, fmt.Errorf("failed to get user by phone: %s", st.Message())
			}
		}
		return nil, fmt.Errorf("failed to get user by phone: %w", err)
	}

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d - %s", response.StatusCode, response.Message)
	}

	// Convert protobuf response to UserData
	userData := &UserData{
		ID:          response.User.Id,
		Username:    response.User.Username,
		PhoneNumber: response.User.PhoneNumber,
		CountryCode: response.User.CountryCode,
		Email:       response.User.Email,
		FullName:    response.User.FullName,
		IsValidated: response.User.IsValidated,
		Status:      response.User.Status,
		CreatedAt:   response.User.CreatedAt,
		UpdatedAt:   response.User.UpdatedAt,
	}

	log.Printf("User data retrieved successfully for phone: %s", phone)
	return userData, nil
}

// GetUserByEmail retrieves a user from the AAA service by email
func (c *Client) GetUserByEmail(ctx context.Context, email string) (*UserData, error) {
	log.Printf("AAA GetUserByEmail: email=%s", email)

	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	// For now, we'll use a placeholder implementation since the proto doesn't have GetUserByEmail
	// In a real implementation, this would call the appropriate gRPC method
	log.Printf("GetUserByEmail not fully implemented - would need AAA service support")
	return nil, fmt.Errorf("GetUserByEmail not implemented in AAA service")
}

// CreateOrganization creates a new organization in AAA service
func (c *Client) CreateOrganization(ctx context.Context, req *CreateOrganizationRequest) (*CreateOrganizationResponse, error) {
	log.Printf("AAA CreateOrganization: name=%s, type=%s", req.Name, req.Type)

	// For now, this is a placeholder implementation
	// In a real implementation, this would call the OrganizationService gRPC method
	log.Printf("CreateOrganization not fully implemented - would need OrganizationService proto")
	return nil, fmt.Errorf("CreateOrganization not implemented - missing OrganizationService proto")
}

// GetOrganization retrieves an organization from AAA service
func (c *Client) GetOrganization(ctx context.Context, orgID string) (*OrganizationData, error) {
	log.Printf("AAA GetOrganization: orgID=%s", orgID)

	// For now, this is a placeholder implementation
	// In a real implementation, this would call the OrganizationService gRPC method
	log.Printf("GetOrganization not fully implemented - would need OrganizationService proto")
	return nil, fmt.Errorf("GetOrganization not implemented - missing OrganizationService proto")
}

// CreateUserGroup creates a user group in AAA service
func (c *Client) CreateUserGroup(ctx context.Context, req *CreateUserGroupRequest) (*CreateUserGroupResponse, error) {
	log.Printf("AAA CreateUserGroup: name=%s, orgID=%s", req.Name, req.OrgID)

	// For now, this is a placeholder implementation
	// In a real implementation, this would call the GroupService gRPC method
	log.Printf("CreateUserGroup not fully implemented - would need GroupService proto")
	return nil, fmt.Errorf("CreateUserGroup not implemented - missing GroupService proto")
}

// AddUserToGroup adds a user to a group
func (c *Client) AddUserToGroup(ctx context.Context, userID, groupID string) error {
	log.Printf("AAA AddUserToGroup: userID=%s, groupID=%s", userID, groupID)

	// For now, this is a placeholder implementation
	// In a real implementation, this would call the GroupService gRPC method
	log.Printf("AddUserToGroup not fully implemented - would need GroupService proto")
	return fmt.Errorf("AddUserToGroup not implemented - missing GroupService proto")
}

// RemoveUserFromGroup removes a user from a group
func (c *Client) RemoveUserFromGroup(ctx context.Context, userID, groupID string) error {
	log.Printf("AAA RemoveUserFromGroup: userID=%s, groupID=%s", userID, groupID)

	// For now, this is a placeholder implementation
	// In a real implementation, this would call the GroupService gRPC method
	log.Printf("RemoveUserFromGroup not fully implemented - would need GroupService proto")
	return fmt.Errorf("RemoveUserFromGroup not implemented - missing GroupService proto")
}

// AssignRole assigns a role to a user in an organization
func (c *Client) AssignRole(ctx context.Context, userID, orgID, roleName string) error {
	log.Printf("AAA AssignRole: userID=%s, orgID=%s, role=%s", userID, orgID, roleName)

	// For now, this is a placeholder implementation
	// In a real implementation, this would call the RoleService gRPC method
	log.Printf("AssignRole not fully implemented - would need RoleService proto")
	return fmt.Errorf("AssignRole not implemented - missing RoleService proto")
}

// CheckUserRole checks if a user has a specific role
func (c *Client) CheckUserRole(ctx context.Context, userID, roleName string) (bool, error) {
	log.Printf("AAA CheckUserRole: userID=%s, role=%s", userID, roleName)

	// For now, this is a placeholder implementation
	// In a real implementation, this would call the RoleService gRPC method
	log.Printf("CheckUserRole not fully implemented - would need RoleService proto")
	return false, fmt.Errorf("CheckUserRole not implemented - missing RoleService proto")
}

// AssignPermissionToGroup assigns a permission to a group
func (c *Client) AssignPermissionToGroup(ctx context.Context, groupID, resource, action string) error {
	log.Printf("AAA AssignPermissionToGroup: groupID=%s, resource=%s, action=%s", groupID, resource, action)

	// For now, this is a placeholder implementation
	// In a real implementation, this would call the PermissionService gRPC method
	log.Printf("AssignPermissionToGroup not fully implemented - would need PermissionService proto")
	return fmt.Errorf("AssignPermissionToGroup not implemented - missing PermissionService proto")
}

// CheckPermission checks if a user has permission to perform an action
func (c *Client) CheckPermission(ctx context.Context, subject, resource, action, object, orgID string) (bool, error) {
	log.Printf("AAA CheckPermission: subject=%s, resource=%s, action=%s, object=%s, orgID=%s",
		subject, resource, action, object, orgID)

	// Validate input parameters
	if subject == "" || resource == "" || action == "" {
		log.Printf("Warning: Missing permission parameters")
		return false, fmt.Errorf("missing permission parameters")
	}

	// Create the authorization request
	resourceID := object
	if resourceID == "" {
		resourceID = "*" // Wildcard for resource-level permissions
	}

	req := &proto.CheckRequest{
		PrincipalId:  subject,
		ResourceType: resource,
		ResourceId:   resourceID,
		Action:       action,
	}

	// Call the AAA authorization service
	response, err := c.authzClient.Check(ctx, req)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.InvalidArgument:
				return false, fmt.Errorf("invalid permission request: %s", st.Message())
			case codes.PermissionDenied:
				return false, nil
			default:
				return false, fmt.Errorf("authorization check failed: %s", st.Message())
			}
		}
		return false, fmt.Errorf("authorization check failed: %w", err)
	}

	log.Printf("Permission check completed: %s wants to %s on %s/%s in org %s - result: %t",
		subject, action, resource, object, orgID, response.Allowed)

	return response.Allowed, nil
}

// ValidateToken validates a JWT token with the AAA service
func (c *Client) ValidateToken(ctx context.Context, token string) (map[string]interface{}, error) {
	log.Printf("AAA ValidateToken: token=%s...", token[:min(10, len(token))])

	if token == "" {
		return nil, fmt.Errorf("token is required")
	}

	// For now, this is a placeholder implementation
	// In a real implementation, this would call the AuthService ValidateToken method
	log.Printf("ValidateToken not fully implemented - would need AuthService proto")
	return nil, fmt.Errorf("ValidateToken not implemented - missing AuthService proto")
}

// SeedRolesAndPermissions seeds roles and permissions in AAA
func (c *Client) SeedRolesAndPermissions(ctx context.Context) error {
	log.Println("AAA SeedRolesAndPermissions: Seeding roles and permissions")

	// For now, this is a placeholder implementation
	// In a real implementation, this would call various methods on the CatalogService
	log.Printf("SeedRolesAndPermissions not fully implemented - would need CatalogService proto")
	return fmt.Errorf("SeedRolesAndPermissions not implemented - missing CatalogService proto")
}

// HealthCheck checks if the AAA service is healthy
func (c *Client) HealthCheck(ctx context.Context) error {
	log.Println("AAA HealthCheck: Checking service health")

	// Simple health check by trying to make a basic call
	// We'll use a simple authorization check as a health indicator
	req := &proto.CheckRequest{
		PrincipalId:  "health-check",
		ResourceType: "system",
		ResourceId:   "health",
		Action:       "check",
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := c.authzClient.Check(ctx, req)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			// Even if the check fails due to permissions, the service is healthy
			if st.Code() == codes.PermissionDenied || st.Code() == codes.InvalidArgument {
				log.Println("AAA service is healthy (responded to health check)")
				return nil
			}
		}
		return fmt.Errorf("AAA service health check failed: %w", err)
	}

	log.Println("AAA service is healthy")
	return nil
}

// AddRequestMetadata adds common metadata to the context for AAA service calls
func (c *Client) AddRequestMetadata(ctx context.Context, requestID, userID string) context.Context {
	md := metadata.New(map[string]string{
		"request-id": requestID,
		"user-id":    userID,
		"source":     "farmers-module",
		"timestamp":  time.Now().Format(time.RFC3339),
	})

	return metadata.NewOutgoingContext(ctx, md)
}

// CreateUserLegacy creates a user with the old method signature for backward compatibility
func (c *Client) CreateUserLegacy(ctx context.Context, username, mobileNumber, password, countryCode string, aadhaarNumber *string) (string, error) {
	log.Printf("AAA CreateUserLegacy: username=%s, mobile=%s, country=%s", username, mobileNumber, countryCode)

	req := &CreateUserRequest{
		Username:    username,
		PhoneNumber: mobileNumber,
		CountryCode: countryCode,
		Email:       mobileNumber, // Using mobile as email for backward compatibility
		Password:    password,
		FullName:    username, // Using username as full name for backward compatibility
	}

	response, err := c.CreateUser(ctx, req)
	if err != nil {
		return "", err
	}

	return response.UserID, nil
}

// GetUserByMobile is an alias for GetUserByPhone for backward compatibility
func (c *Client) GetUserByMobile(ctx context.Context, mobileNumber string) (map[string]interface{}, error) {
	userData, err := c.GetUserByPhone(ctx, mobileNumber)
	if err != nil {
		return nil, err
	}

	// Convert UserData to map for backward compatibility
	return map[string]interface{}{
		"id":            userData.ID,
		"username":      userData.Username,
		"mobile_number": userData.PhoneNumber,
		"status":        userData.Status,
		"created_at":    userData.CreatedAt,
		"updated_at":    userData.UpdatedAt,
	}, nil
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
