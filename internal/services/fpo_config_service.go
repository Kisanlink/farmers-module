package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Kisanlink/farmers-module/internal/entities/fpo_config"
	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"gorm.io/gorm"
)

// Note: The FPO configuration has been simplified to:
// - API endpoint (erp_base_url)
// - UI link (erp_ui_base_url)
// - Contact information
// - Business hours
// - Metadata

// FPOConfigService defines the interface for FPO configuration operations
type FPOConfigService interface {
	// GetFPOConfig retrieves FPO configuration by AAA Org ID
	GetFPOConfig(ctx context.Context, aaaOrgID string) (*responses.FPOConfigData, error)

	// CreateFPOConfig creates a new FPO configuration
	CreateFPOConfig(ctx context.Context, req *requests.CreateFPOConfigRequest) (*responses.FPOConfigData, error)

	// UpdateFPOConfig updates an existing FPO configuration
	UpdateFPOConfig(ctx context.Context, aaaOrgID string, req *requests.UpdateFPOConfigRequest) (*responses.FPOConfigData, error)

	// DeleteFPOConfig deletes an FPO configuration (soft delete)
	DeleteFPOConfig(ctx context.Context, aaaOrgID string, deletedBy ...string) error

	// ListFPOConfigs lists all FPO configurations with pagination
	ListFPOConfigs(ctx context.Context, req *requests.ListFPOConfigsRequest) ([]*responses.FPOConfigData, *responses.PaginationInfo, error)

	// CheckERPHealth checks the health of FPO's ERP service
	CheckERPHealth(ctx context.Context, aaaOrgID string) (*responses.FPOHealthCheckData, error)
}

// fpoConfigService implements FPOConfigService
type fpoConfigService struct {
	repo *base.BaseFilterableRepository[*fpo_config.FPOConfig]
}

// NewFPOConfigService creates a new FPO configuration service
func NewFPOConfigService(repo *base.BaseFilterableRepository[*fpo_config.FPOConfig]) FPOConfigService {
	return &fpoConfigService{
		repo: repo,
	}
}

// GetFPOConfig retrieves FPO configuration by AAA Org ID
func (s *fpoConfigService) GetFPOConfig(ctx context.Context, aaaOrgID string) (*responses.FPOConfigData, error) {
	if aaaOrgID == "" {
		return nil, common.ErrInvalidInput
	}

	// Find by ID (which is same as aaa_org_id due to BeforeCreate hook)
	config := &fpo_config.FPOConfig{}
	config, err := s.repo.GetByID(ctx, aaaOrgID, config)
	if err != nil {
		// Check if this is a "not found" error using multiple methods
		// 1. Check if it's GORM's ErrRecordNotFound
		// 2. Check if it's wrapped common.ErrNotFound
		// 3. Check if error message contains "not found" or "record not found"
		isNotFound := errors.Is(err, gorm.ErrRecordNotFound) ||
			errors.Is(err, common.ErrNotFound) ||
			strings.Contains(err.Error(), "not found") ||
			strings.Contains(err.Error(), "record not found")

		if isNotFound {
			// Return default config instead of error
			// This allows the frontend to know the FPO exists but has no configuration yet
			metadata := make(map[string]interface{})
			metadata["config_status"] = "not_configured"
			metadata["message"] = "FPO configuration has not been set up yet"

			return &responses.FPOConfigData{
				ID:            aaaOrgID,
				AAAOrgID:      aaaOrgID,
				FPOName:       "",
				ERPBaseURL:    "",
				ERPUIBaseURL:  "",
				Contact:       make(map[string]interface{}),
				BusinessHours: make(map[string]interface{}),
				Metadata:      metadata,
			}, nil
		}
		return nil, fmt.Errorf("failed to fetch FPO config: %w", err)
	}

	return s.toResponseData(config), nil
}

// CreateFPOConfig creates a new FPO configuration
func (s *fpoConfigService) CreateFPOConfig(ctx context.Context, req *requests.CreateFPOConfigRequest) (*responses.FPOConfigData, error) {
	// Set defaults
	req.SetDefaults()

	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", common.ErrInvalidInput, err)
	}

	// Check if FPO config already exists
	existing := &fpo_config.FPOConfig{}
	existing, err := s.repo.GetByID(ctx, req.AAAOrgID, existing)
	if err == nil && existing != nil && existing.ID != "" {
		return nil, fmt.Errorf("%w: FPO config already exists for aaa_org_id: %s", common.ErrAlreadyExists, req.AAAOrgID)
	}

	// Create FPO config entity using constructor
	// Constructor ensures ID is set correctly from the start
	config := fpo_config.NewFPOConfig(req.AAAOrgID)

	// Set user-provided fields
	config.FPOName = req.FPOName
	config.ERPBaseURL = req.ERPBaseURL
	config.ERPUIBaseURL = req.ERPUIBaseURL
	config.Contact = req.Contact
	config.BusinessHours = req.BusinessHours
	config.Metadata = req.Metadata

	// Validate entity
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// Create in database
	if err := s.repo.Create(ctx, config); err != nil {
		return nil, fmt.Errorf("failed to create FPO config: %w", err)
	}

	return s.toResponseData(config), nil
}

// UpdateFPOConfig updates an existing FPO configuration
func (s *fpoConfigService) UpdateFPOConfig(ctx context.Context, aaaOrgID string, req *requests.UpdateFPOConfigRequest) (*responses.FPOConfigData, error) {
	if aaaOrgID == "" {
		return nil, common.ErrInvalidInput
	}

	// Set defaults
	req.SetDefaults()

	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", common.ErrInvalidInput, err)
	}

	// Find existing config
	config := &fpo_config.FPOConfig{}
	config, err := s.repo.GetByID(ctx, aaaOrgID, config)
	if err != nil {
		if err == common.ErrNotFound {
			return nil, common.ErrNotFound
		}
		return nil, fmt.Errorf("failed to fetch FPO config: %w", err)
	}

	// Update fields
	if req.FPOName != nil {
		config.FPOName = *req.FPOName
	}
	if req.ERPBaseURL != nil {
		config.ERPBaseURL = *req.ERPBaseURL
	}
	if req.ERPUIBaseURL != nil {
		config.ERPUIBaseURL = *req.ERPUIBaseURL
	}
	if req.Contact != nil {
		config.Contact = req.Contact
	}
	if req.BusinessHours != nil {
		config.BusinessHours = req.BusinessHours
	}
	if req.Metadata != nil {
		config.Metadata = req.Metadata
	}

	// Validate entity
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// Update in database
	if err := s.repo.Update(ctx, config); err != nil {
		return nil, fmt.Errorf("failed to update FPO config: %w", err)
	}

	return s.toResponseData(config), nil
}

// DeleteFPOConfig deletes an FPO configuration (soft delete)
func (s *fpoConfigService) DeleteFPOConfig(ctx context.Context, aaaOrgID string, deletedBy ...string) error {
	if aaaOrgID == "" {
		return common.ErrInvalidInput
	}

	// Find existing config
	config := &fpo_config.FPOConfig{}
	config, err := s.repo.GetByID(ctx, aaaOrgID, config)
	if err != nil {
		if err == common.ErrNotFound {
			return common.ErrNotFound
		}
		return fmt.Errorf("failed to fetch FPO config: %w", err)
	}

	// Soft delete with audit trail
	deletedByUser := ""
	if len(deletedBy) > 0 {
		deletedByUser = deletedBy[0]
	}
	if err := s.repo.SoftDelete(ctx, config.ID, deletedByUser); err != nil {
		return fmt.Errorf("failed to delete FPO config: %w", err)
	}

	return nil
}

// ListFPOConfigs lists all FPO configurations with pagination
func (s *fpoConfigService) ListFPOConfigs(ctx context.Context, req *requests.ListFPOConfigsRequest) ([]*responses.FPOConfigData, *responses.PaginationInfo, error) {
	// Set defaults
	req.SetDefaults()

	// Validate request
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("%w: %v", common.ErrInvalidInput, err)
	}

	// Build filter
	fb := base.NewFilterBuilder().
		Where("deleted_at", base.OpIsNull, nil)

	// Add search filter - search in aaa_org_id OR fpo_name
	if req.Search != "" {
		// Note: kisanlink-db doesn't support OR conditions directly yet
		// For now, we'll search by aaa_org_id only
		// TODO: Update when kisanlink-db supports OR conditions
		fb.Where("aaa_org_id", base.OpLike, "%"+req.Search+"%")
	}

	// Add status filter
	if req.Status != "" {
		fb.Where("api_health_status", base.OpEqual, req.Status)
	}

	filter := fb.Build()
	filter.Page = req.Page
	filter.Limit = req.PageSize

	// Fetch configs
	configs, err := s.repo.Find(ctx, filter)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list FPO configs: %w", err)
	}

	// Count total
	totalCount, err := s.repo.Count(ctx, filter, &fpo_config.FPOConfig{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count FPO configs: %w", err)
	}

	// Convert to response
	data := make([]*responses.FPOConfigData, len(configs))
	for i, config := range configs {
		data[i] = s.toResponseData(config)
	}

	// Build pagination info using kisanlink-db helper
	pagination := responses.NewPaginationInfo(req.Page, req.PageSize, int(totalCount))

	return data, pagination, nil
}

// CheckERPHealth checks the health of FPO's ERP service
// This is an on-demand health check and does not persist status to the database
func (s *fpoConfigService) CheckERPHealth(ctx context.Context, aaaOrgID string) (*responses.FPOHealthCheckData, error) {
	if aaaOrgID == "" {
		return nil, common.ErrInvalidInput
	}

	// Get FPO config
	configData, err := s.GetFPOConfig(ctx, aaaOrgID)
	if err != nil {
		return nil, err
	}

	// ERP URL to test - prioritizing override from context if added later, 
	// or fallback to saved config
	erpURL := configData.ERPBaseURL
	
	// Check if we have an override URL in context (added by handler for "Test Connection" button)
	if overrideURL, ok := ctx.Value("override_erp_url").(string); ok && overrideURL != "" {
		erpURL = overrideURL
	}

	if erpURL == "" {
		return &responses.FPOHealthCheckData{
			AAAOrgID:     aaaOrgID,
			ERPBaseURL:   "",
			ERPUIBaseURL: "",
			Status:       "not_configured",
			LastChecked:  time.Now(),
			Error:        "FPO configuration has not been set up yet",
		}, nil
	}

	// Perform health check
	startTime := time.Now()
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Try multiple health check paths
	// 1. URL + /health (as provided)
	// 2. If URL has /v1 or /v2, try stripping it and adding /health (root health check)
	
	healthPaths := []string{erpURL + "/health"}
	
	// If URL ends with /v1, /v2, etc, try root /health
	if strings.Contains(erpURL, "/v") {
		// Attempt to find the base domain (strip anything after the last / before /v)
		// Example: http://localhost:8002/v1 -> http://localhost:8002/health
		baseDomain := erpURL
		if idx := strings.LastIndex(erpURL, "/v"); idx != -1 {
			baseDomain = erpURL[:idx]
			healthPaths = append(healthPaths, baseDomain+"/health")
		}
	}

	var lastErr error
	var finalStatus string = "unhealthy"
	
	for _, healthURL := range healthPaths {
		resp, err := client.Get(healthURL)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			finalStatus = "healthy"
			lastErr = nil
			break
		} else {
			lastErr = fmt.Errorf("HTTP status: %d", resp.StatusCode)
		}
	}

	responseTime := time.Since(startTime).Milliseconds()

	healthData := &responses.FPOHealthCheckData{
		AAAOrgID:       aaaOrgID,
		ERPBaseURL:     erpURL,
		ERPUIBaseURL:   configData.ERPUIBaseURL,
		LastChecked:    time.Now(),
		ResponseTimeMs: responseTime,
		Status:         finalStatus,
	}

	if lastErr != nil {
		healthData.Error = lastErr.Error()
	}

	return healthData, nil
}

// toResponseData converts FPOConfig entity to response data
func (s *fpoConfigService) toResponseData(config *fpo_config.FPOConfig) *responses.FPOConfigData {
	// Initialize metadata if nil
	metadata := config.Metadata
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	return &responses.FPOConfigData{
		ID:            config.ID,
		AAAOrgID:      config.AAAOrgID,
		FPOName:       config.FPOName,
		ERPBaseURL:    config.ERPBaseURL,
		ERPUIBaseURL:  config.ERPUIBaseURL,
		Contact:       config.Contact,
		BusinessHours: config.BusinessHours,
		Metadata:      metadata,
		CreatedAt:     config.CreatedAt,
		UpdatedAt:     config.UpdatedAt,
	}
}
