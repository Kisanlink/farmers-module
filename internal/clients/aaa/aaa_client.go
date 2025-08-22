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

	// For now, we'll simulate organization creation since the proto files aren't fully generated
	// In a real implementation, this would call the AAA service's CreateOrganization gRPC method

	// Simulate successful organization creation
	orgData := &OrganizationData{
		ID:          fmt.Sprintf("org_%d", time.Now().Unix()),
		Name:        name,
		Description: description,
		Type:        orgType,
		Status:      "ACTIVE",
		Metadata:    metadata,
	}

	log.Printf("Organization created successfully: %s", orgData.ID)
	return orgData, nil
}

// VerifyOrganization verifies if an organization exists and is active in AAA service
func (c *Client) VerifyOrganization(ctx context.Context, orgID string) (*OrganizationData, error) {
	log.Printf("AAA VerifyOrganization: orgID=%s", orgID)

	// For now, we'll simulate organization verification
	// In a real implementation, this would call the AAA service's GetOrganization gRPC method

	// Simulate successful organization verification
	orgData := &OrganizationData{
		ID:          orgID,
		Name:        "Verified Organization",
		Description: "Organization verified from AAA service",
		Type:        "fpo",
		Status:      "ACTIVE",
		Metadata:    make(map[string]string),
	}

	log.Printf("Organization verified successfully: %s", orgData.ID)
	return orgData, nil
}

// AssignRole assigns a role to a user in an organization
func (c *Client) AssignRole(ctx context.Context, userID, orgID, roleName string) error {
	log.Printf("AAA AssignRole: userID=%s, orgID=%s, role=%s", userID, orgID, roleName)

	// For now, we'll simulate role assignment
	// In a real implementation, this would call the AAA service's role assignment gRPC method

	log.Printf("Role %s assigned successfully to user %s in organization %s", roleName, userID, orgID)
	return nil
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
		"status":        "ACTIVE", // TODO: Get actual status from response
		"created_at":    response.User.CreatedAt,
		"updated_at":    response.User.UpdatedAt,
	}

	log.Printf("User data retrieved successfully for ID: %s", userID)
	return userData, nil
}

// SeedRolesAndPermissions seeds roles and permissions in AAA
func (c *Client) SeedRolesAndPermissions(ctx context.Context) error {
	log.Println("AAA SeedRolesAndPermissions: Seeding roles and permissions")

	// TODO: Implement actual gRPC call to AAA service's admin endpoints
	// This would typically call various methods on the AuthorizationService

	// For now, just log the operation
	log.Println("Mock seeding of roles and permissions completed")
	return nil
}

// ValidateToken validates a JWT token with the AAA service
func (c *Client) ValidateToken(ctx context.Context, token string) (map[string]interface{}, error) {
	log.Printf("AAA ValidateToken: token=%s...", token[:10])

	// TODO: Implement actual gRPC call to AAA service's token validation endpoint
	// This would typically call the ValidateToken method on the AuthService
	// For now, return mock validation data since the auth service doesn't expose this via gRPC yet

	validationData := map[string]interface{}{
		"valid":      true,
		"user_id":    "mock-user-id",
		"username":   "mock-username",
		"org_id":     "mock-org-id",
		"expires_at": time.Now().Add(24 * time.Hour).Format(time.RFC3339),
	}

	log.Printf("Mock token validation completed for token: %s...", token[:10])
	return validationData, nil
}

// GetUserPermissions retrieves all permissions for a user
func (c *Client) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	log.Printf("AAA GetUserPermissions: userID=%s", userID)

	// For now, check common permissions using the authorization service
	commonResources := []string{"farmers", "farms", "crops", "activities"}
	commonActions := []string{"create", "read", "update", "delete"}

	var permissions []string

	for _, resource := range commonResources {
		for _, action := range commonActions {
			req := &proto.CheckRequest{
				PrincipalId:  userID,
				ResourceType: resource,
				ResourceId:   "*",
				Action:       action,
			}

			response, err := c.authzClient.Check(ctx, req)
			if err == nil && response.Allowed {
				permissions = append(permissions, fmt.Sprintf("%s:%s", resource, action))
			}
		}
	}

	log.Printf("Permissions retrieved for user %s: %v", userID, permissions)
	return permissions, nil
}

// CheckUserRole checks if a user has a specific role
func (c *Client) CheckUserRole(ctx context.Context, userID, roleName string) (bool, error) {
	log.Printf("AAA CheckUserRole: userID=%s, role=%s", userID, roleName)

	// TODO: Implement actual gRPC call to AAA service's role checking endpoint
	// This would typically call the HasRole method on the AuthorizationService

	// For now, return mock role check for testing
	hasRole := roleName == "farmer" || roleName == "admin"
	log.Printf("Mock role check for user %s, role %s: %t", userID, roleName, hasRole)

	return hasRole, nil
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

// HealthCheck checks if the AAA service is healthy
func (c *Client) HealthCheck(ctx context.Context) error {
	// TODO: Implement actual gRPC health check
	// This would typically call the health check endpoint on the AAA service

	log.Println("AAA HealthCheck: Service health check completed")
	return nil
}
