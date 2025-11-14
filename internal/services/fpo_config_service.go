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
	// GetFPOConfig retrieves FPO configuration by FPO ID
	GetFPOConfig(ctx context.Context, fpoID string) (*responses.FPOConfigData, error)

	// CreateFPOConfig creates a new FPO configuration
	CreateFPOConfig(ctx context.Context, req *requests.CreateFPOConfigRequest) (*responses.FPOConfigData, error)

	// UpdateFPOConfig updates an existing FPO configuration
	UpdateFPOConfig(ctx context.Context, fpoID string, req *requests.UpdateFPOConfigRequest) (*responses.FPOConfigData, error)

	// DeleteFPOConfig deletes an FPO configuration (soft delete)
	DeleteFPOConfig(ctx context.Context, fpoID string) error

	// ListFPOConfigs lists all FPO configurations with pagination
	ListFPOConfigs(ctx context.Context, req *requests.ListFPOConfigsRequest) ([]*responses.FPOConfigData, *responses.PaginationInfo, error)

	// CheckERPHealth checks the health of FPO's ERP service
	CheckERPHealth(ctx context.Context, fpoID string) (*responses.FPOHealthCheckData, error)
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

// GetFPOConfig retrieves FPO configuration by FPO ID
func (s *fpoConfigService) GetFPOConfig(ctx context.Context, fpoID string) (*responses.FPOConfigData, error) {
	if fpoID == "" {
		return nil, common.ErrInvalidInput
	}

	// Build filter
	filter := base.NewFilterBuilder().
		Where("fpo_id", base.OpEqual, fpoID).
		Where("deleted_at", base.OpIsNull, nil).
		Build()

	// Find FPO config
	config, err := s.repo.FindOne(ctx, filter)
	if err != nil {
		if err == common.ErrNotFound {
			return nil, common.ErrNotFound
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
	filter := base.NewFilterBuilder().
		Where("fpo_id", base.OpEqual, req.FPOID).
		Where("deleted_at", base.OpIsNull, nil).
		Build()

	existing, err := s.repo.FindOne(ctx, filter)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("%w: FPO config already exists for fpo_id: %s", common.ErrAlreadyExists, req.FPOID)
	}

	// Create FPO config entity
	config := &fpo_config.FPOConfig{
		FPOID:           req.FPOID,
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
func (s *fpoConfigService) UpdateFPOConfig(ctx context.Context, fpoID string, req *requests.UpdateFPOConfigRequest) (*responses.FPOConfigData, error) {
	if fpoID == "" {
		return nil, common.ErrInvalidInput
	}

	// Set defaults
	req.SetDefaults()

	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", common.ErrInvalidInput, err)
	}

	// Find existing config
	filter := base.NewFilterBuilder().
		Where("fpo_id", base.OpEqual, fpoID).
		Where("deleted_at", base.OpIsNull, nil).
		Build()

	config, err := s.repo.FindOne(ctx, filter)
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
func (s *fpoConfigService) DeleteFPOConfig(ctx context.Context, fpoID string) error {
	if fpoID == "" {
		return common.ErrInvalidInput
	}

	// Find existing config
	filter := base.NewFilterBuilder().
		Where("fpo_id", base.OpEqual, fpoID).
		Where("deleted_at", base.OpIsNull, nil).
		Build()

	config, err := s.repo.FindOne(ctx, filter)
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

	// Add search filter - search in fpo_id OR fpo_name
	if req.Search != "" {
		// Note: kisanlink-db doesn't support OR conditions directly yet
		// For now, we'll search by fpo_id only
		// TODO: Update when kisanlink-db supports OR conditions
		fb.Where("fpo_id", base.OpLike, "%"+req.Search+"%")
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
func (s *fpoConfigService) CheckERPHealth(ctx context.Context, fpoID string) (*responses.FPOHealthCheckData, error) {
	if fpoID == "" {
		return nil, common.ErrInvalidInput
	}

	// Get FPO config
	configData, err := s.GetFPOConfig(ctx, fpoID)
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
		FPOID:          fpoID,
		ERPBaseURL:     configData.ERPBaseURL,
		LastChecked:    time.Now(),
		ResponseTimeMs: responseTime,
	}

	if err != nil {
		healthData.Status = "unhealthy"
		healthData.Error = err.Error()

		// Update config status
		s.updateHealthStatus(ctx, fpoID, "unhealthy")

		return healthData, nil
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusOK {
		healthData.Status = "healthy"

		// Update config status
		s.updateHealthStatus(ctx, fpoID, "healthy")
	} else {
		healthData.Status = "unhealthy"
		healthData.Error = fmt.Sprintf("HTTP status: %d", resp.StatusCode)

		// Update config status
		s.updateHealthStatus(ctx, fpoID, "unhealthy")
	}

	return healthData, nil
}

// updateHealthStatus updates the health status of FPO config
func (s *fpoConfigService) updateHealthStatus(ctx context.Context, fpoID string, status string) {
	filter := base.NewFilterBuilder().
		Where("fpo_id", base.OpEqual, fpoID).
		Where("deleted_at", base.OpIsNull, nil).
		Build()

	config, err := s.repo.FindOne(ctx, filter)
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
		FPOID:           config.FPOID,
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
