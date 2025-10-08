package aaa

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Kisanlink/farmers-module/internal/auth"
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
	conn          *grpc.ClientConn
	config        *config.Config
	userClient    proto.UserServiceV2Client
	authzClient   proto.AuthorizationServiceClient
	orgClient     proto.OrganizationServiceClient
	groupClient   proto.GroupServiceClient
	roleClient    proto.RoleServiceClient
	permClient    proto.PermissionServiceClient
	catalogClient proto.CatalogServiceClient
	tokenClient   proto.TokenServiceClient
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
	log.Printf("AAA Client: Connecting to gRPC endpoint: %s", cfg.AAA.GRPCEndpoint)

	conn, err := grpc.NewClient(
		cfg.AAA.GRPCEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to AAA service: %w", err)
	}

	// Initialize gRPC clients
	userClient := proto.NewUserServiceV2Client(conn)
	authzClient := proto.NewAuthorizationServiceClient(conn)
	orgClient := proto.NewOrganizationServiceClient(conn)
	groupClient := proto.NewGroupServiceClient(conn)
	roleClient := proto.NewRoleServiceClient(conn)
	permClient := proto.NewPermissionServiceClient(conn)
	catalogClient := proto.NewCatalogServiceClient(conn)
	tokenClient := proto.NewTokenServiceClient(conn)

	log.Printf("AAA Client: Successfully initialized with endpoint %s", cfg.AAA.GRPCEndpoint)

	return &Client{
		conn:          conn,
		config:        cfg,
		userClient:    userClient,
		authzClient:   authzClient,
		orgClient:     orgClient,
		groupClient:   groupClient,
		roleClient:    roleClient,
		permClient:    permClient,
		catalogClient: catalogClient,
		tokenClient:   tokenClient,
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

	if c.userClient == nil {
		return nil, fmt.Errorf("user service client not initialized")
	}

	// Get all users and filter by email
	// Note: This is inefficient but works until AAA service adds GetUserByEmail
	req := &proto.GetAllUsersRequestV2{
		// No pagination fields available in current proto
	}

	resp, err := c.userClient.GetAllUsers(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	// Find user by email
	for _, user := range resp.Users {
		if user.Email == email {
			// Convert to UserData
			userData := &UserData{
				ID:          user.Id,
				Username:    user.Username,
				Email:       user.Email,
				PhoneNumber: user.PhoneNumber,
				CountryCode: user.CountryCode,
				FullName:    user.FullName,
				Status:      user.Status,
				CreatedAt:   user.CreatedAt,
				UpdatedAt:   user.UpdatedAt,
			}

			// Extract roles from UserRoles (UserRoleV2 doesn't have direct role field)
			// userData.Roles would need to be added to UserData struct
			// For now, we'll skip roles extraction since the proto doesn't expose them properly

			log.Printf("Found user by email: %s", user.Id)
			return userData, nil
		}
	}

	return nil, fmt.Errorf("user not found with email: %s", email)
}

// CreateOrganization creates a new organization in AAA service
func (c *Client) CreateOrganization(ctx context.Context, req *CreateOrganizationRequest) (*CreateOrganizationResponse, error) {
	log.Printf("AAA CreateOrganization: name=%s, type=%s", req.Name, req.Type)

	// Validate request
	if req.Name == "" {
		return nil, fmt.Errorf("organization name is required")
	}
	if req.Type == "" {
		return nil, fmt.Errorf("organization type is required")
	}

	// Create gRPC request
	grpcReq := &proto.CreateOrganizationRequest{
		Name:        req.Name,
		DisplayName: req.Name, // Use name as display name if not provided
		Description: req.Description,
		Type:        req.Type,
		OwnerId:     req.CEOUserID,
	}

	// Call the AAA service
	response, err := c.orgClient.CreateOrganization(ctx, grpcReq)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.AlreadyExists:
				return nil, fmt.Errorf("organization already exists")
			case codes.InvalidArgument:
				return nil, fmt.Errorf("invalid request: %s", st.Message())
			default:
				return nil, fmt.Errorf("failed to create organization: %s", st.Message())
			}
		}
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	if response.StatusCode != 200 && response.StatusCode != 201 {
		return nil, fmt.Errorf("unexpected status code: %d - %s", response.StatusCode, response.Message)
	}

	log.Printf("Organization created successfully with ID: %s", response.Organization.Id)

	return &CreateOrganizationResponse{
		OrgID:     response.Organization.Id,
		Name:      response.Organization.Name,
		Status:    response.Organization.Status,
		CreatedAt: time.Now(), // Use current time since proto uses string for timestamps
	}, nil
}

// GetOrganization retrieves an organization from AAA service
func (c *Client) GetOrganization(ctx context.Context, orgID string) (*OrganizationData, error) {
	log.Printf("AAA GetOrganization: orgID=%s", orgID)

	if orgID == "" {
		return nil, fmt.Errorf("organization ID is required")
	}

	// Create gRPC request
	grpcReq := &proto.GetOrganizationRequest{
		Id: orgID,
	}

	// Call the AAA service
	response, err := c.orgClient.GetOrganization(ctx, grpcReq)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return nil, fmt.Errorf("organization not found")
			default:
				return nil, fmt.Errorf("failed to get organization: %s", st.Message())
			}
		}
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d - %s", response.StatusCode, response.Message)
	}

	// Convert proto response to OrganizationData
	org := response.Organization
	orgData := &OrganizationData{
		ID:          org.Id,
		Name:        org.Name,
		Type:        org.Type,
		Status:      org.Status,
		Description: org.Description,
		Metadata:    make(map[string]string),
	}

	log.Printf("Organization data retrieved successfully for ID: %s", orgID)
	return orgData, nil
}

// CreateUserGroup creates a user group in AAA service
func (c *Client) CreateUserGroup(ctx context.Context, req *CreateUserGroupRequest) (*CreateUserGroupResponse, error) {
	log.Printf("AAA CreateUserGroup: name=%s, orgID=%s", req.Name, req.OrgID)

	// Validate request
	if req.Name == "" {
		return nil, fmt.Errorf("group name is required")
	}
	if req.OrgID == "" {
		return nil, fmt.Errorf("organization ID is required")
	}

	// Create gRPC request
	grpcReq := &proto.CreateGroupRequest{
		Name:           req.Name,
		Description:    req.Description,
		OrganizationId: req.OrgID,
	}

	// Call the AAA service
	response, err := c.groupClient.CreateGroup(ctx, grpcReq)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.AlreadyExists:
				return nil, fmt.Errorf("group already exists")
			case codes.InvalidArgument:
				return nil, fmt.Errorf("invalid request: %s", st.Message())
			default:
				return nil, fmt.Errorf("failed to create group: %s", st.Message())
			}
		}
		return nil, fmt.Errorf("failed to create group: %w", err)
	}

	log.Printf("Group created successfully with ID: %s", response.Group.Id)

	return &CreateUserGroupResponse{
		GroupID:   response.Group.Id,
		Name:      response.Group.Name,
		OrgID:     response.Group.OrganizationId,
		CreatedAt: time.Now(),
	}, nil
}

// AddUserToGroup adds a user to a group
func (c *Client) AddUserToGroup(ctx context.Context, userID, groupID string) error {
	log.Printf("AAA AddUserToGroup: userID=%s, groupID=%s", userID, groupID)

	// Validate input
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}
	if groupID == "" {
		return fmt.Errorf("group ID is required")
	}

	// Create gRPC request
	grpcReq := &proto.AddGroupMemberRequest{
		GroupId:       groupID,
		PrincipalId:   userID,
		PrincipalType: "user",
	}

	// Call the AAA service
	response, err := c.groupClient.AddGroupMember(ctx, grpcReq)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			return fmt.Errorf("failed to add user to group: %s", st.Message())
		}
		return fmt.Errorf("failed to add user to group: %w", err)
	}

	if response.StatusCode != 200 && response.StatusCode != 201 {
		return fmt.Errorf("failed to add user to group: %s", response.Message)
	}

	log.Printf("User %s added to group %s successfully", userID, groupID)
	return nil
}

// RemoveUserFromGroup removes a user from a group
func (c *Client) RemoveUserFromGroup(ctx context.Context, userID, groupID string) error {
	log.Printf("AAA RemoveUserFromGroup: userID=%s, groupID=%s", userID, groupID)

	// Validate input
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}
	if groupID == "" {
		return fmt.Errorf("group ID is required")
	}

	// Create gRPC request
	grpcReq := &proto.RemoveGroupMemberRequest{
		GroupId:     groupID,
		PrincipalId: userID,
	}

	// Call the AAA service
	response, err := c.groupClient.RemoveGroupMember(ctx, grpcReq)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			return fmt.Errorf("failed to remove user from group: %s", st.Message())
		}
		return fmt.Errorf("failed to remove user from group: %w", err)
	}

	if response.StatusCode != 200 {
		return fmt.Errorf("failed to remove user from group: %s", response.Message)
	}

	log.Printf("User %s removed from group %s successfully", userID, groupID)
	return nil
}

// AssignRole assigns a role to a user in an organization
func (c *Client) AssignRole(ctx context.Context, userID, orgID, roleName string) error {
	log.Printf("AAA AssignRole: userID=%s, orgID=%s, role=%s", userID, orgID, roleName)

	// Validate input
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}
	if orgID == "" {
		return fmt.Errorf("organization ID is required")
	}
	if roleName == "" {
		return fmt.Errorf("role name is required")
	}

	// Validate role name against known roles
	validRoles := map[string]bool{
		"admin":       true,
		"farmer":      true,
		"kisansathi":  true,
		"fpo_manager": true,
		"readonly":    true,
	}
	if !validRoles[strings.ToLower(roleName)] {
		return fmt.Errorf("invalid role name: %s", roleName)
	}

	// Create gRPC request
	grpcReq := &proto.AssignRoleRequest{
		UserId:   userID,
		OrgId:    orgID,
		RoleName: roleName,
	}

	// Call the AAA service
	response, err := c.roleClient.AssignRole(ctx, grpcReq)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			return fmt.Errorf("failed to assign role: %s", st.Message())
		}
		return fmt.Errorf("failed to assign role: %w", err)
	}

	if response.StatusCode != 200 && response.StatusCode != 201 {
		return fmt.Errorf("failed to assign role: %s", response.Message)
	}

	log.Printf("Role %s assigned to user %s in org %s successfully", roleName, userID, orgID)
	return nil
}

// CheckUserRole checks if a user has a specific role
func (c *Client) CheckUserRole(ctx context.Context, userID, roleName string) (bool, error) {
	log.Printf("AAA CheckUserRole: userID=%s, role=%s", userID, roleName)

	// Validate input
	if userID == "" {
		return false, fmt.Errorf("user ID is required")
	}
	if roleName == "" {
		return false, fmt.Errorf("role name is required")
	}

	if c.userClient == nil {
		// NOTE: Role service is not yet available, using stub response
		log.Printf("STUB: CheckUserRole called - RoleService not yet available")
		// Default to false for security
		return false, nil
	}

	// Try to get user and check roles from UserServiceV2
	req := &proto.GetUserRequestV2{Id: userID}
	resp, err := c.userClient.GetUser(ctx, req)
	if err != nil {
		log.Printf("Failed to get user for role check: %v", err)
		return false, nil // Return false on error for security
	}

	// Check if user has the role
	for _, userRole := range resp.User.UserRoles {
		if strings.EqualFold(userRole.RoleName, roleName) {
			log.Printf("User %s has role %s", userID, roleName)
			return true, nil
		}
	}

	log.Printf("User %s does not have role %s", userID, roleName)
	return false, nil
}

// AssignPermissionToGroup assigns a permission to a group
func (c *Client) AssignPermissionToGroup(ctx context.Context, groupID, resource, action string) error {
	log.Printf("AAA AssignPermissionToGroup: groupID=%s, resource=%s, action=%s", groupID, resource, action)

	// Validate input
	if groupID == "" {
		return fmt.Errorf("group ID is required")
	}
	if resource == "" {
		return fmt.Errorf("resource is required")
	}
	if action == "" {
		return fmt.Errorf("action is required")
	}

	// Validate action against known actions
	validActions := map[string]bool{
		"create": true,
		"read":   true,
		"update": true,
		"delete": true,
		"list":   true,
		"manage": true,
	}
	if !validActions[strings.ToLower(action)] {
		return fmt.Errorf("invalid action: %s", action)
	}

	// Create gRPC request
	grpcReq := &proto.AssignPermissionToGroupRequest{
		GroupId:  groupID,
		Resource: resource,
		Action:   action,
	}

	// Call the AAA service
	response, err := c.permClient.AssignPermissionToGroup(ctx, grpcReq)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			return fmt.Errorf("failed to assign permission to group: %s", st.Message())
		}
		return fmt.Errorf("failed to assign permission to group: %w", err)
	}

	if response.StatusCode != 200 && response.StatusCode != 201 {
		return fmt.Errorf("failed to assign permission to group: %s", response.Message)
	}

	log.Printf("Permission %s:%s assigned to group %s successfully", resource, action, groupID)
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

	if c.authzClient == nil {
		log.Printf("Authorization client not initialized; allowing by default")
		return true, nil
	}

	// Convert empty object to wildcard for list operations
	resourceID := object
	if resourceID == "" {
		resourceID = "*"
	}

	req := &proto.CheckRequest{
		PrincipalId:  subject,
		ResourceType: resource,
		ResourceId:   resourceID,
		Action:       action,
	}

	// Use a longer timeout for permission checks
	rpcTimeout := 10 * time.Second
	if d, err := time.ParseDuration(c.config.AAA.RequestTimeout); err == nil && d > 0 {
		rpcTimeout = d
	}
	rpcCtx, cancel := context.WithTimeout(ctx, rpcTimeout)
	defer cancel()

	// Extract auth token from context and add to gRPC metadata
	if token := auth.GetTokenFromContext(ctx); token != "" {
		md := metadata.New(map[string]string{
			"authorization": "Bearer " + token,
		})
		rpcCtx = metadata.NewOutgoingContext(rpcCtx, md)
		log.Printf("AAA CheckPermission: Added authorization token to gRPC metadata (token length: %d)", len(token))
	} else {
		log.Printf("AAA CheckPermission: Warning - no auth token found in context")
	}

	log.Printf("AAA CheckPermission: Calling authz service at %s", c.config.AAA.GRPCEndpoint)
	resp, err := c.authzClient.Check(rpcCtx, req)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			log.Printf("AAA CheckPermission: gRPC error code=%s, message=%s", st.Code(), st.Message())
			switch st.Code() {
			case codes.Unimplemented, codes.Unavailable:
				// Maintain permissive fallback when authz is not served
				log.Printf("Authz service %v; allowing by default", st.Code())
				return true, nil
			case codes.DeadlineExceeded:
				// Handle timeout - allow by default to prevent blocking
				log.Printf("Warning: Permission check timed out; allowing by default")
				return true, nil
			case codes.PermissionDenied:
				return false, nil
			default:
				return false, fmt.Errorf("permission check failed: %s", st.Message())
			}
		}
		log.Printf("AAA CheckPermission: Non-gRPC error: %v", err)
		return false, fmt.Errorf("permission check failed: %w", err)
	}

	log.Printf("AAA CheckPermission: Success, allowed=%v", resp.GetAllowed())
	return resp.GetAllowed(), nil
}

// ValidateToken validates a JWT token with the AAA service
func (c *Client) ValidateToken(ctx context.Context, token string) (map[string]interface{}, error) {
	log.Printf("AAA ValidateToken: validating token via remote service")

	if token == "" {
		return nil, fmt.Errorf("token is required")
	}

	// Call AAA service for token validation
	req := &proto.ValidateTokenRequest{
		Token:               token,
		IncludeUserDetails:  true,
		IncludePermissions:  true,
		IncludeOrganization: true,
		StrictValidation:    true,
	}

	// Use a timeout for the validation call
	rpcTimeout := 5 * time.Second
	if d, err := time.ParseDuration(c.config.AAA.RequestTimeout); err == nil && d > 0 {
		rpcTimeout = d
	}
	rpcCtx, cancel := context.WithTimeout(ctx, rpcTimeout)
	defer cancel()

	// Add token to gRPC metadata for AAA service authentication
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + token,
	})
	rpcCtx = metadata.NewOutgoingContext(rpcCtx, md)

	resp, err := c.tokenClient.ValidateToken(rpcCtx, req)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			log.Printf("AAA ValidateToken: gRPC error code=%s, message=%s", st.Code(), st.Message())
			return nil, fmt.Errorf("token validation failed: %s", st.Message())
		}
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	if !resp.Valid {
		log.Printf("AAA ValidateToken: Token invalid - %s", resp.Message)
		return nil, fmt.Errorf("invalid token: %s", resp.Message)
	}

	if resp.Claims == nil {
		return nil, fmt.Errorf("no claims returned from token validation")
	}

	// Convert claims to map for backward compatibility
	result := map[string]interface{}{
		"user_id":     resp.Claims.UserId,
		"username":    resp.Claims.Username,
		"email":       resp.Claims.Email,
		"org_id":      resp.Claims.OrganizationId,
		"org_name":    resp.Claims.OrganizationName,
		"roles":       resp.Claims.Roles,
		"permissions": resp.Claims.Permissions,
		"token_type":  resp.Claims.TokenType,
	}

	if resp.Claims.ExpiresAt != nil {
		result["exp"] = resp.Claims.ExpiresAt.Seconds
	}
	if resp.Claims.IssuedAt != nil {
		result["iat"] = resp.Claims.IssuedAt.Seconds
	}

	log.Printf("Token validated successfully for user: %s", resp.Claims.UserId)
	return result, nil
}

// SeedRolesAndPermissions seeds roles and permissions in AAA
func (c *Client) SeedRolesAndPermissions(ctx context.Context) error {
	log.Println("AAA SeedRolesAndPermissions: Seeding roles and permissions")

	// Create gRPC request
	grpcReq := &proto.SeedRolesAndPermissionsRequest{
		Force: false, // Don't force reseed if data exists
	}

	// Call the AAA service
	response, err := c.catalogClient.SeedRolesAndPermissions(ctx, grpcReq)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			return fmt.Errorf("failed to seed roles and permissions: %s", st.Message())
		}
		return fmt.Errorf("failed to seed roles and permissions: %w", err)
	}

	if response.StatusCode != 200 && response.StatusCode != 201 {
		return fmt.Errorf("failed to seed roles and permissions: %s", response.Message)
	}

	log.Printf("Roles and permissions seeded successfully: %d roles, %d permissions created",
		response.RolesCreated, response.PermissionsCreated)
	return nil
}

// HealthCheck checks if the AAA service is healthy
func (c *Client) HealthCheck(ctx context.Context) error {
	log.Println("AAA HealthCheck: Checking service health")

	// Use a simple call to UserServiceV2 which we know exists
	// Try to get a non-existent user - if service responds (even with NotFound), it's healthy
	req := &proto.GetUserRequestV2{
		Id: "health-check-user-id",
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := c.userClient.GetUser(ctx, req)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			// If we get NotFound, InvalidArgument, or similar, the service is responding
			switch st.Code() {
			case codes.NotFound, codes.InvalidArgument, codes.PermissionDenied:
				log.Println("AAA service is healthy (UserService responded)")
				return nil
			case codes.Unavailable, codes.DeadlineExceeded:
				return fmt.Errorf("AAA service health check failed: service unavailable")
			default:
				// For other errors, consider service healthy if it responded
				log.Printf("AAA service responded with code %v, considering healthy", st.Code())
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
