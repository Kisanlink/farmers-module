package services

import (
	"context"
	"fmt"
	"log"

	"github.com/Kisanlink/farmers-module/internal/clients/aaa"
	"github.com/Kisanlink/farmers-module/internal/config"
	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AAAClientInterface defines the interface for AAA client operations
type AAAClientInterface interface {
	CreateUser(ctx context.Context, req *aaa.CreateUserRequest) (*aaa.CreateUserResponse, error)
	GetUser(ctx context.Context, userID string) (*aaa.UserData, error)
	GetUserByPhone(ctx context.Context, phone string) (*aaa.UserData, error)
	GetUserByEmail(ctx context.Context, email string) (*aaa.UserData, error)
	GetUserByMobile(ctx context.Context, mobile string) (map[string]interface{}, error)
	CreateOrganization(ctx context.Context, req *aaa.CreateOrganizationRequest) (*aaa.CreateOrganizationResponse, error)
	GetOrganization(ctx context.Context, orgID string) (*aaa.OrganizationData, error)
	CreateUserGroup(ctx context.Context, req *aaa.CreateUserGroupRequest) (*aaa.CreateUserGroupResponse, error)
	AddUserToGroup(ctx context.Context, userID, groupID string) error
	RemoveUserFromGroup(ctx context.Context, userID, groupID string) error
	AssignRole(ctx context.Context, userID, orgID, roleName string) error
	CheckUserRole(ctx context.Context, userID, roleName string) (bool, error)
	AssignPermissionToGroup(ctx context.Context, groupID, resource, action string) error
	CheckPermission(ctx context.Context, subject, resource, action, object, orgID string) (bool, error)
	ValidateToken(ctx context.Context, token string) (map[string]interface{}, error)
	SeedRolesAndPermissions(ctx context.Context, force bool) error
	HealthCheck(ctx context.Context) error
	Close() error
}

// AAAServiceImpl implements AAAService
type AAAServiceImpl struct {
	config *config.Config
	client AAAClientInterface
}

// NewAAAService creates a new AAA service
func NewAAAService(cfg *config.Config) AAAService {
	var client AAAClientInterface
	var err error

	// Only create client if AAA is enabled
	if cfg.AAA.Enabled {
		client, err = aaa.NewClient(cfg)
		if err != nil {
			log.Printf("Warning: Failed to create AAA client: %v", err)
			log.Printf("AAA service will run in degraded mode without external integration")
			client = nil
		}
	} else {
		log.Printf("AAA service integration is disabled")
		client = nil
	}

	return &AAAServiceImpl{
		config: cfg,
		client: client,
	}
}

// SeedRolesAndPermissions implements W18: Seed roles and permissions
func (s *AAAServiceImpl) SeedRolesAndPermissions(ctx context.Context, force bool) error {
	if s.client == nil {
		log.Println("AAA client not available, skipping seeding")
		return nil
	}

	return s.client.SeedRolesAndPermissions(ctx, force)
}

// CheckPermission implements W19: Check permission
func (s *AAAServiceImpl) CheckPermission(ctx context.Context, subject, resource, action, object, orgID string) (bool, error) {
	if s.client == nil {
		log.Println("AAA client not available, allowing operation")
		return true, nil
	}

	return s.client.CheckPermission(ctx, subject, resource, action, object, orgID)
}

// CreateUser creates a user in AAA
func (s *AAAServiceImpl) CreateUser(ctx context.Context, req interface{}) (interface{}, error) {
	if s.client == nil {
		return nil, fmt.Errorf("AAA client not available")
	}

	// Type assert the request to get the required fields
	createReq, ok := req.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid request format")
	}

	username, _ := createReq["username"].(string)
	phoneNumber, _ := createReq["phone_number"].(string)
	email, _ := createReq["email"].(string)
	password, _ := createReq["password"].(string)
	countryCode, _ := createReq["country_code"].(string)
	fullName, _ := createReq["full_name"].(string)
	role, _ := createReq["role"].(string)

	// Create structured request
	userReq := &aaa.CreateUserRequest{
		Username:    username,
		PhoneNumber: phoneNumber,
		CountryCode: countryCode,
		Email:       email,
		Password:    password,
		FullName:    fullName,
		Role:        role,
	}

	response, err := s.client.CreateUser(ctx, userReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create user in AAA: %w", err)
	}

	return map[string]interface{}{
		"id":         response.UserID,
		"username":   response.Username,
		"status":     response.Status,
		"created_at": response.CreatedAt,
	}, nil
}

// GetUser gets a user from AAA
func (s *AAAServiceImpl) GetUser(ctx context.Context, userID string) (interface{}, error) {
	if s.client == nil {
		return nil, fmt.Errorf("AAA client not available")
	}

	userData, err := s.client.GetUser(ctx, userID)
	if err != nil {
		return nil, s.mapGRPCError(err, "get user")
	}

	return userData, nil
}

// GetUserByMobile gets a user from AAA by mobile number
func (s *AAAServiceImpl) GetUserByMobile(ctx context.Context, mobileNumber string) (interface{}, error) {
	if s.client == nil {
		return nil, fmt.Errorf("AAA client not available")
	}

	userData, err := s.client.GetUserByMobile(ctx, mobileNumber)
	if err != nil {
		return nil, s.mapGRPCError(err, "get user by mobile")
	}

	return userData, nil
}

// GetUserByEmail gets a user from AAA by email
func (s *AAAServiceImpl) GetUserByEmail(ctx context.Context, email string) (interface{}, error) {
	if s.client == nil {
		return nil, fmt.Errorf("AAA client not available")
	}

	userData, err := s.client.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, s.mapGRPCError(err, "get user by email")
	}

	return userData, nil
}

// CreateOrganization creates an organization in AAA
func (s *AAAServiceImpl) CreateOrganization(ctx context.Context, req interface{}) (interface{}, error) {
	if s.client == nil {
		return nil, fmt.Errorf("AAA client not available")
	}

	// Type assert the request to get the required fields
	createReq, ok := req.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid request format")
	}

	name, _ := createReq["name"].(string)
	description, _ := createReq["description"].(string)
	orgType, _ := createReq["type"].(string)
	ceoUserID, _ := createReq["ceo_user_id"].(string)
	metadata, _ := createReq["metadata"].(map[string]string)

	// Create structured request
	orgReq := &aaa.CreateOrganizationRequest{
		Name:        name,
		Description: description,
		Type:        orgType,
		CEOUserID:   ceoUserID,
		Metadata:    metadata,
	}

	response, err := s.client.CreateOrganization(ctx, orgReq)
	if err != nil {
		return nil, s.mapGRPCError(err, "create organization")
	}

	return map[string]interface{}{
		"org_id":     response.OrgID,
		"name":       response.Name,
		"status":     response.Status,
		"created_at": response.CreatedAt,
	}, nil
}

// GetOrganization gets an organization from AAA
func (s *AAAServiceImpl) GetOrganization(ctx context.Context, orgID string) (interface{}, error) {
	if s.client == nil {
		return nil, fmt.Errorf("AAA client not available")
	}

	orgData, err := s.client.GetOrganization(ctx, orgID)
	if err != nil {
		return nil, s.mapGRPCError(err, "get organization")
	}

	return orgData, nil
}

// CreateUserGroup creates a user group in AAA
func (s *AAAServiceImpl) CreateUserGroup(ctx context.Context, req interface{}) (interface{}, error) {
	if s.client == nil {
		return nil, fmt.Errorf("AAA client not available")
	}

	// Type assert the request to get the required fields
	createReq, ok := req.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid request format")
	}

	name, _ := createReq["name"].(string)
	description, _ := createReq["description"].(string)
	orgID, _ := createReq["org_id"].(string)
	permissions, _ := createReq["permissions"].([]string)

	// Create structured request
	groupReq := &aaa.CreateUserGroupRequest{
		Name:        name,
		Description: description,
		OrgID:       orgID,
		Permissions: permissions,
	}

	response, err := s.client.CreateUserGroup(ctx, groupReq)
	if err != nil {
		return nil, s.mapGRPCError(err, "create user group")
	}

	return map[string]interface{}{
		"group_id":   response.GroupID,
		"name":       response.Name,
		"org_id":     response.OrgID,
		"created_at": response.CreatedAt,
	}, nil
}

// AddUserToGroup adds a user to a group
func (s *AAAServiceImpl) AddUserToGroup(ctx context.Context, userID, groupID string) error {
	if s.client == nil {
		return fmt.Errorf("AAA client not available")
	}

	err := s.client.AddUserToGroup(ctx, userID, groupID)
	if err != nil {
		return s.mapGRPCError(err, "add user to group")
	}

	return nil
}

// RemoveUserFromGroup removes a user from a group
func (s *AAAServiceImpl) RemoveUserFromGroup(ctx context.Context, userID, groupID string) error {
	if s.client == nil {
		return fmt.Errorf("AAA client not available")
	}

	err := s.client.RemoveUserFromGroup(ctx, userID, groupID)
	if err != nil {
		return s.mapGRPCError(err, "remove user from group")
	}

	return nil
}

// AssignRole assigns a role to a user in an organization
func (s *AAAServiceImpl) AssignRole(ctx context.Context, userID, orgID, roleName string) error {
	if s.client == nil {
		return fmt.Errorf("AAA client not available")
	}

	err := s.client.AssignRole(ctx, userID, orgID, roleName)
	if err != nil {
		return s.mapGRPCError(err, "assign role")
	}

	return nil
}

// CheckUserRole checks if a user has a specific role
func (s *AAAServiceImpl) CheckUserRole(ctx context.Context, userID, roleName string) (bool, error) {
	if s.client == nil {
		log.Println("AAA client not available, returning false for role check")
		return false, nil
	}

	hasRole, err := s.client.CheckUserRole(ctx, userID, roleName)
	if err != nil {
		return false, s.mapGRPCError(err, "check user role")
	}

	return hasRole, nil
}

// AssignPermissionToGroup assigns a permission to a group
func (s *AAAServiceImpl) AssignPermissionToGroup(ctx context.Context, groupID, resource, action string) error {
	if s.client == nil {
		return fmt.Errorf("AAA client not available")
	}

	err := s.client.AssignPermissionToGroup(ctx, groupID, resource, action)
	if err != nil {
		return s.mapGRPCError(err, "assign permission to group")
	}

	return nil
}

// ValidateToken validates a JWT token with the AAA service
func (s *AAAServiceImpl) ValidateToken(ctx context.Context, token string) (*interfaces.UserInfo, error) {
	if s.client == nil {
		return nil, fmt.Errorf("AAA client not available")
	}

	tokenData, err := s.client.ValidateToken(ctx, token)
	if err != nil {
		return nil, s.mapGRPCError(err, "validate token")
	}

	// Convert the map response to UserInfo struct
	userInfo := &interfaces.UserInfo{}

	if userID, ok := tokenData["user_id"].(string); ok {
		userInfo.UserID = userID
	}
	if username, ok := tokenData["username"].(string); ok {
		userInfo.Username = username
	}
	if email, ok := tokenData["email"].(string); ok {
		userInfo.Email = email
	}
	if phone, ok := tokenData["phone"].(string); ok {
		userInfo.Phone = phone
	}
	if roles, ok := tokenData["roles"].([]interface{}); ok {
		userInfo.Roles = make([]string, len(roles))
		for i, role := range roles {
			if roleStr, ok := role.(string); ok {
				userInfo.Roles[i] = roleStr
			}
		}
	}
	if orgID, ok := tokenData["org_id"].(string); ok {
		userInfo.OrgID = orgID
	}
	if orgName, ok := tokenData["org_name"].(string); ok {
		userInfo.OrgName = orgName
	}
	if orgType, ok := tokenData["org_type"].(string); ok {
		userInfo.OrgType = orgType
	}

	return userInfo, nil
}

// ValidateTokenRaw validates a JWT token and returns raw map data (for backward compatibility)
func (s *AAAServiceImpl) ValidateTokenRaw(ctx context.Context, token string) (map[string]interface{}, error) {
	if s.client == nil {
		return nil, fmt.Errorf("AAA client not available")
	}

	tokenData, err := s.client.ValidateToken(ctx, token)
	if err != nil {
		return nil, s.mapGRPCError(err, "validate token")
	}

	return tokenData, nil
}

// HealthCheck checks if the AAA service is healthy
func (s *AAAServiceImpl) HealthCheck(ctx context.Context) error {
	if s.client == nil {
		return fmt.Errorf("AAA client not available")
	}

	err := s.client.HealthCheck(ctx)
	if err != nil {
		return s.mapGRPCError(err, "health check")
	}

	return nil
}

// mapGRPCError maps gRPC errors to appropriate application errors
func (s *AAAServiceImpl) mapGRPCError(err error, operation string) error {
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.NotFound:
			return fmt.Errorf("%s failed: resource not found", operation)
		case codes.AlreadyExists:
			return fmt.Errorf("%s failed: resource already exists", operation)
		case codes.InvalidArgument:
			return fmt.Errorf("%s failed: invalid argument - %s", operation, st.Message())
		case codes.PermissionDenied:
			return fmt.Errorf("%s failed: permission denied", operation)
		case codes.Unauthenticated:
			return fmt.Errorf("%s failed: authentication required", operation)
		case codes.Unavailable:
			return fmt.Errorf("%s failed: AAA service unavailable", operation)
		case codes.DeadlineExceeded:
			return fmt.Errorf("%s failed: request timeout", operation)
		default:
			return fmt.Errorf("%s failed: %s", operation, st.Message())
		}
	}
	return fmt.Errorf("%s failed: %w", operation, err)
}
