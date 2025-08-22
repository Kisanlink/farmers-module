package aaa

import (
	"context"
	"fmt"
	"log"

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

// OrganizationData represents organization information
type OrganizationData struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Type        string            `json:"type"`
	Status      string            `json:"status"`
	Metadata    map[string]string `json:"metadata"`
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

// CreateOrganization creates a new organization in AAA service
func (c *Client) CreateOrganization(ctx context.Context, name, description, orgType string, metadata map[string]string) (*OrganizationData, error) {
	log.Printf("AAA CreateOrganization: name=%s, type=%s", name, orgType)

	// TODO: Implement actual gRPC call to AAA service's CreateOrganization endpoint
	// This would typically call a method like CreateOrganization on the OrganizationService
	return nil, fmt.Errorf("CreateOrganization not implemented yet")
}

// VerifyOrganization verifies if an organization exists and is active in AAA service
func (c *Client) VerifyOrganization(ctx context.Context, orgID string) (*OrganizationData, error) {
	log.Printf("AAA VerifyOrganization: orgID=%s", orgID)

	// TODO: Implement actual gRPC call to AAA service's GetOrganization endpoint
	// This would typically call a method like GetOrganization on the OrganizationService
	return nil, fmt.Errorf("VerifyOrganization not implemented yet")
}

// AssignRole assigns a role to a user in an organization
func (c *Client) AssignRole(ctx context.Context, userID, orgID, roleName string) error {
	log.Printf("AAA AssignRole: userID=%s, orgID=%s, role=%s", userID, orgID, roleName)

	// TODO: Implement actual gRPC call to AAA service's role assignment endpoint
	// This would typically call a method like AssignRole on the AuthorizationService
	return fmt.Errorf("AssignRole not implemented yet")
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

// CreateUser creates a user in the AAA service
func (c *Client) CreateUser(ctx context.Context, username, mobileNumber, password, countryCode string, aadhaarNumber *string) (string, error) {
	log.Printf("AAA CreateUser: username=%s, mobile=%s, country=%s", username, mobileNumber, countryCode)

	// Create gRPC request
	req := &proto.RegisterRequestV2{
		Username: username,
		Email:    mobileNumber, // Using mobile as email for now
		Password: password,
	}

	// Call the AAA service
	response, err := c.userClient.Register(ctx, req)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.AlreadyExists:
				return "", fmt.Errorf("user already exists")
			case codes.InvalidArgument:
				return "", fmt.Errorf("invalid request: %s", st.Message())
			default:
				return "", fmt.Errorf("failed to create user: %s", st.Message())
			}
		}
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	if response.StatusCode != 201 {
		return "", fmt.Errorf("unexpected status code: %d - %s", response.StatusCode, response.Message)
	}

	log.Printf("User created successfully with ID: %s", response.User.Id)
	return response.User.Id, nil
}

// GetUser retrieves a user from the AAA service
func (c *Client) GetUser(ctx context.Context, userID string) (map[string]interface{}, error) {
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

	// Convert protobuf response to map
	userData := map[string]interface{}{
		"id":            response.User.Id,
		"username":      response.User.Username,
		"mobile_number": response.User.PhoneNumber,
		"status":        response.User.Status,
		"created_at":    response.User.CreatedAt,
		"updated_at":    response.User.UpdatedAt,
	}

	log.Printf("User data retrieved successfully for ID: %s", userID)
	return userData, nil
}

// GetUserByMobile retrieves a user from the AAA service by mobile number
func (c *Client) GetUserByMobile(ctx context.Context, mobileNumber string) (map[string]interface{}, error) {
	log.Printf("AAA GetUserByMobile: mobileNumber=%s", mobileNumber)

	if mobileNumber == "" {
		return nil, fmt.Errorf("mobile number is required")
	}

	// Create gRPC request to get user by phone number
	req := &proto.GetUserByPhoneRequestV2{
		PhoneNumber:        mobileNumber,
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
				return nil, fmt.Errorf("user not found with mobile number: %s", mobileNumber)
			default:
				return nil, fmt.Errorf("failed to get user by mobile: %s", st.Message())
			}
		}
		return nil, fmt.Errorf("failed to get user by mobile: %w", err)
	}

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d - %s", response.StatusCode, response.Message)
	}

	// Convert protobuf response to map
	userData := map[string]interface{}{
		"id":            response.User.Id,
		"username":      response.User.Username,
		"mobile_number": response.User.PhoneNumber,
		"status":        response.User.Status,
		"created_at":    response.User.CreatedAt,
		"updated_at":    response.User.UpdatedAt,
	}

	log.Printf("User data retrieved successfully for mobile: %s", mobileNumber)
	return userData, nil
}

// SeedRolesAndPermissions seeds roles and permissions in AAA
func (c *Client) SeedRolesAndPermissions(ctx context.Context) error {
	log.Println("AAA SeedRolesAndPermissions: Seeding roles and permissions")

	// TODO: Implement actual gRPC call to AAA service's admin endpoints
	// This would typically call various methods on the AuthorizationService
	return fmt.Errorf("SeedRolesAndPermissions not implemented yet")
}

// ValidateToken validates a JWT token with the AAA service
func (c *Client) ValidateToken(ctx context.Context, token string) (map[string]interface{}, error) {
	log.Printf("AAA ValidateToken: token=%s...", token[:10])

	// TODO: Implement actual gRPC call to AAA service's token validation endpoint
	// This would typically call the ValidateToken method on the AuthService
	return nil, fmt.Errorf("ValidateToken not implemented yet")
}

// GetUserPermissions retrieves all permissions for a user
func (c *Client) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	log.Printf("AAA GetUserPermissions: userID=%s", userID)

	// TODO: Implement actual gRPC call to AAA service's permissions endpoint
	// This would typically call a method like GetUserPermissions on the AuthorizationService
	return nil, fmt.Errorf("GetUserPermissions not implemented yet")
}

// CheckUserRole checks if a user has a specific role
func (c *Client) CheckUserRole(ctx context.Context, userID, roleName string) (bool, error) {
	log.Printf("AAA CheckUserRole: userID=%s, role=%s", userID, roleName)

	// TODO: Implement actual gRPC call to AAA service's role checking endpoint
	// This would typically call the HasRole method on the AuthorizationService
	return false, fmt.Errorf("CheckUserRole not implemented yet")
}

// AddRequestMetadata adds common metadata to the context for AAA service calls
func (c *Client) AddRequestMetadata(ctx context.Context, requestID, userID string) context.Context {
	md := metadata.New(map[string]string{
		"request-id": requestID,
		"user-id":    userID,
		"source":     "farmers-module",
		"timestamp":  "2024-01-01T00:00:00Z", // TODO: Use actual timestamp
	})

	return metadata.NewOutgoingContext(ctx, md)
}

// HealthCheck checks if the AAA service is healthy
func (c *Client) HealthCheck(ctx context.Context) error {
	// TODO: Implement actual gRPC health check
	// This would typically call the health check endpoint on the AAA service
	return fmt.Errorf("HealthCheck not implemented yet")
}
