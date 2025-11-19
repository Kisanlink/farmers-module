package services

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Kisanlink/farmers-module/internal/entities/fpo_config"
	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// FPOConfigService defines the interface for FPO configuration operations
type FPOConfigService interface {
	// GetFPOConfig retrieves FPO configuration by AAA Org ID
	GetFPOConfig(ctx context.Context, aaaOrgID string) (*responses.FPOConfigData, error)

	// CreateFPOConfig creates a new FPO configuration
	CreateFPOConfig(ctx context.Context, req *requests.CreateFPOConfigRequest) (*responses.FPOConfigData, error)

	// UpdateFPOConfig updates an existing FPO configuration
	UpdateFPOConfig(ctx context.Context, aaaOrgID string, req *requests.UpdateFPOConfigRequest) (*responses.FPOConfigData, error)

	// DeleteFPOConfig deletes an FPO configuration (soft delete)
	DeleteFPOConfig(ctx context.Context, aaaOrgID string) error

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
		if err == common.ErrNotFound {
			// Return default config instead of error
			// This allows the frontend to know the FPO exists but has no configuration yet
			metadata := make(map[string]interface{})
			metadata["config_status"] = "not_configured"
			metadata["message"] = "FPO configuration has not been set up yet"

			return &responses.FPOConfigData{
				ID:              aaaOrgID,
				AAAOrgID:        aaaOrgID,
				FPOName:         "",
				ERPBaseURL:      "",
				ERPAPIVersion:   "",
				Features:        make(map[string]interface{}),
				Contact:         make(map[string]interface{}),
				BusinessHours:   make(map[string]interface{}),
				Metadata:        metadata,
				APIHealthStatus: "not_configured",
				LastSyncedAt:    nil,
				SyncInterval:    0,
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

	// Create FPO config entity
	config := &fpo_config.FPOConfig{
		AAAOrgID:        req.AAAOrgID,
		FPOName:         req.FPOName,
		ERPBaseURL:      req.ERPBaseURL,
		ERPAPIVersion:   req.ERPAPIVersion,
		Features:        req.Features,
		Contact:         req.Contact,
		BusinessHours:   req.BusinessHours,
		Metadata:        make(map[string]interface{}),
		APIHealthStatus: "unknown",
		SyncInterval:    req.SyncInterval,
	}

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
	if req.ERPAPIVersion != nil {
		config.ERPAPIVersion = *req.ERPAPIVersion
	}
	if req.Features != nil {
		config.Features = req.Features
	}
	if req.Contact != nil {
		config.Contact = req.Contact
	}
	if req.BusinessHours != nil {
		config.BusinessHours = req.BusinessHours
	}
	if req.SyncInterval != nil {
		config.SyncInterval = *req.SyncInterval
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
func (s *fpoConfigService) DeleteFPOConfig(ctx context.Context, aaaOrgID string) error {
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

	// Soft delete
	if err := s.repo.Delete(ctx, config.ID, config); err != nil {
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
func (s *fpoConfigService) CheckERPHealth(ctx context.Context, aaaOrgID string) (*responses.FPOHealthCheckData, error) {
	if aaaOrgID == "" {
		return nil, common.ErrInvalidInput
	}

	// Get FPO config
	configData, err := s.GetFPOConfig(ctx, aaaOrgID)
	if err != nil {
		return nil, err
	}

	// Perform health check
	startTime := time.Now()
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	healthURL := configData.ERPBaseURL + "/health"
	resp, err := client.Get(healthURL)
	responseTime := time.Since(startTime).Milliseconds()

	healthData := &responses.FPOHealthCheckData{
		AAAOrgID:       aaaOrgID,
		ERPBaseURL:     configData.ERPBaseURL,
		LastChecked:    time.Now(),
		ResponseTimeMs: responseTime,
	}

	if err != nil {
		healthData.Status = "unhealthy"
		healthData.Error = err.Error()

		// Update config status
		s.updateHealthStatus(ctx, aaaOrgID, "unhealthy")

		return healthData, nil
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusOK {
		healthData.Status = "healthy"

		// Update config status
		s.updateHealthStatus(ctx, aaaOrgID, "healthy")
	} else {
		healthData.Status = "unhealthy"
		healthData.Error = fmt.Sprintf("HTTP status: %d", resp.StatusCode)

		// Update config status
		s.updateHealthStatus(ctx, aaaOrgID, "unhealthy")
	}

	return healthData, nil
}

// updateHealthStatus updates the health status of FPO config
func (s *fpoConfigService) updateHealthStatus(ctx context.Context, aaaOrgID string, status string) {
	config := &fpo_config.FPOConfig{}
	config, err := s.repo.GetByID(ctx, aaaOrgID, config)
	if err != nil {
		return
	}

	config.APIHealthStatus = status
	now := time.Now()
	config.LastSyncedAt = &now

	_ = s.repo.Update(ctx, config)
}

// toResponseData converts FPOConfig entity to response data
func (s *fpoConfigService) toResponseData(config *fpo_config.FPOConfig) *responses.FPOConfigData {
	// Build metadata
	metadata := config.Metadata
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["last_synced_at"] = config.LastSyncedAt
	metadata["sync_interval_minutes"] = config.SyncInterval
	metadata["api_health_status"] = config.APIHealthStatus

	return &responses.FPOConfigData{
		ID:              config.ID,
		AAAOrgID:        config.AAAOrgID,
		FPOName:         config.FPOName,
		ERPBaseURL:      config.ERPBaseURL,
		ERPAPIVersion:   config.ERPAPIVersion,
		Features:        config.Features,
		Contact:         config.Contact,
		BusinessHours:   config.BusinessHours,
		Metadata:        metadata,
		APIHealthStatus: config.APIHealthStatus,
		LastSyncedAt:    config.LastSyncedAt,
		SyncInterval:    config.SyncInterval,
		CreatedAt:       config.CreatedAt,
		UpdatedAt:       config.UpdatedAt,
	}
}
