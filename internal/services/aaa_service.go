package services

import (
	"context"
	"fmt"
	"log"

	"github.com/Kisanlink/farmers-module/internal/clients/aaa"
	"github.com/Kisanlink/farmers-module/internal/config"
)

// AAAServiceImpl implements AAAService
type AAAServiceImpl struct {
	config *config.Config
	client *aaa.Client
}

// NewAAAService creates a new AAA service
func NewAAAService(cfg *config.Config) AAAService {
	client, err := aaa.NewClient(cfg)
	if err != nil {
		log.Printf("Warning: Failed to create AAA client: %v", err)
		// Continue without client for now
	}

	return &AAAServiceImpl{
		config: cfg,
		client: client,
	}
}

// SeedRolesAndPermissions implements W18: Seed roles and permissions
func (s *AAAServiceImpl) SeedRolesAndPermissions(ctx context.Context) error {
	if s.client == nil {
		log.Println("AAA client not available, skipping seeding")
		return nil
	}

	return s.client.SeedRolesAndPermissions(ctx)
}

// CheckPermission implements W19: Check permission
func (s *AAAServiceImpl) CheckPermission(ctx context.Context, req interface{}) (bool, error) {
	if s.client == nil {
		log.Println("AAA client not available, allowing operation")
		return true, nil
	}

	// Type assert the request to get the required fields
	permissionReq, ok := req.(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("invalid request format")
	}

	subject, ok := permissionReq["subject"].(string)
	if !ok {
		return false, fmt.Errorf("subject is required")
	}

	resource, ok := permissionReq["resource"].(string)
	if !ok {
		return false, fmt.Errorf("resource is required")
	}

	action, ok := permissionReq["action"].(string)
	if !ok {
		return false, fmt.Errorf("action is required")
	}

	object, _ := permissionReq["object"].(string)
	orgID, _ := permissionReq["org_id"].(string)

	return s.client.CheckPermission(ctx, subject, resource, action, object, orgID)
}

// CreateUser creates a user in AAA
func (s *AAAServiceImpl) CreateUser(ctx context.Context, req interface{}) (interface{}, error) {
	if s.client == nil {
		log.Println("AAA client not available, returning mock user")
		// Type assert the request to get the required fields
		createReq, ok := req.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid request format")
		}

		username, _ := createReq["username"].(string)
		mobileNumber, _ := createReq["mobile_number"].(string)
		aadhaarNumber, _ := createReq["aadhaar_number"].(string)

		return map[string]interface{}{
			"id":             "temp-user-id",
			"username":       username,
			"mobile_number":  mobileNumber,
			"aadhaar_number": aadhaarNumber,
		}, nil
	}

	// Type assert the request to get the required fields
	createReq, ok := req.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid request format")
	}

	username, _ := createReq["username"].(string)
	mobileNumber, _ := createReq["mobile_number"].(string)
	password, _ := createReq["password"].(string)
	countryCode, _ := createReq["country_code"].(string)
	aadhaarNumber, _ := createReq["aadhaar_number"].(string)

	userID, err := s.client.CreateUser(ctx, username, mobileNumber, password, countryCode, &aadhaarNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to create user in AAA: %w", err)
	}

	return map[string]interface{}{
		"id":             userID,
		"username":       username,
		"mobile_number":  mobileNumber,
		"aadhaar_number": aadhaarNumber,
	}, nil
}

// GetUser gets a user from AAA
func (s *AAAServiceImpl) GetUser(ctx context.Context, userID string) (interface{}, error) {
	if s.client == nil {
		log.Println("AAA client not available, returning mock user")
		return map[string]interface{}{
			"id":            userID,
			"username":      "temp-username",
			"mobile_number": "temp-mobile",
		}, nil
	}

	userData, err := s.client.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user from AAA: %w", err)
	}

	// Extract fields from userData map
	username, _ := userData["username"].(string)
	mobileNumber, _ := userData["mobile_number"].(string)

	return map[string]interface{}{
		"id":            userID,
		"username":      username,
		"mobile_number": mobileNumber,
	}, nil
}
