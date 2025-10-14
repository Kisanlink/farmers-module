package services

import (
	"context"
	"fmt"

	"github.com/Kisanlink/farmers-module/internal/auth"
	cropCycleEntity "github.com/Kisanlink/farmers-module/internal/entities/crop_cycle"
	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/farmers-module/internal/repo/crop_cycle"
	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// CropCycleServiceImpl implements CropCycleService
type CropCycleServiceImpl struct {
	cropCycleRepo *crop_cycle.CropCycleRepository
	aaaService    AAAService
}

// NewCropCycleService creates a new crop cycle service
func NewCropCycleService(cropCycleRepo *crop_cycle.CropCycleRepository, aaaService AAAService) CropCycleService {
	return &CropCycleServiceImpl{
		cropCycleRepo: cropCycleRepo,
		aaaService:    aaaService,
	}
}

// StartCycle implements W10: Start crop cycle
func (s *CropCycleServiceImpl) StartCycle(ctx context.Context, req interface{}) (interface{}, error) {
	startReq, ok := req.(*requests.StartCycleRequest)
	if !ok {
		return nil, common.ErrInvalidInput
	}

	// Extract authenticated user from context
	userCtx, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user context: %w", err)
	}

	// Check if authenticated user can start crop cycle
	hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "cycle", "start", "", startReq.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Validate farm exists and user has access
	// This would typically involve checking the farm service
	// For now, we'll assume the farm validation is done by the permission check

	// Validate area allocation if provided
	if startReq.AreaHa != nil {
		if err := s.cropCycleRepo.ValidateAreaAllocation(ctx, startReq.FarmID, "", *startReq.AreaHa); err != nil {
			return nil, err
		}
	}

	// Create crop cycle entity
	cycle := &cropCycleEntity.CropCycle{
		FarmID:    startReq.FarmID,
		FarmerID:  startReq.UserID,
		AreaHa:    startReq.AreaHa,
		Season:    startReq.Season,
		Status:    "PLANNED",
		StartDate: &startReq.StartDate,
		CropID:    startReq.CropID,
		VarietyID: startReq.VarietyID,
	}

	// Validate the cycle
	if err := cycle.Validate(); err != nil {
		return nil, err
	}

	// Create the cycle in database
	if err := s.cropCycleRepo.Create(ctx, cycle); err != nil {
		return nil, fmt.Errorf("failed to create crop cycle: %w", err)
	}

	// Convert to response data
	cycleData := &responses.CropCycleData{
		ID:        cycle.ID,
		FarmID:    cycle.FarmID,
		FarmerID:  cycle.FarmerID,
		AreaHa:    cycle.AreaHa,
		Season:    cycle.Season,
		Status:    cycle.Status,
		StartDate: cycle.StartDate,
		EndDate:   cycle.EndDate,
		CropID:    cycle.CropID,
		VarietyID: cycle.VarietyID,
		CropName:  cycle.GetCropName(),
		VarietyName: func() *string {
			if name := cycle.GetVarietyName(); name != "" {
				return &name
			}
			return nil
		}(),
		Outcome:   cycle.Outcome,
		CreatedAt: cycle.CreatedAt,
		UpdatedAt: cycle.UpdatedAt,
	}

	return responses.NewCropCycleResponse(cycleData, "Crop cycle started successfully"), nil
}

// UpdateCycle implements W11: Update crop cycle
func (s *CropCycleServiceImpl) UpdateCycle(ctx context.Context, req interface{}) (interface{}, error) {
	updateReq, ok := req.(*requests.UpdateCycleRequest)
	if !ok {
		return nil, common.ErrInvalidInput
	}

	// Extract authenticated user from context
	userCtx, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user context: %w", err)
	}

	// Check if authenticated user can update crop cycle
	hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "cycle", "update", updateReq.ID, updateReq.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Get existing cycle
	cycle := &cropCycleEntity.CropCycle{}
	_, err = s.cropCycleRepo.GetByID(ctx, updateReq.ID, cycle)
	if err != nil {
		return nil, fmt.Errorf("failed to get crop cycle: %w", err)
	}

	// Check if cycle is in terminal state
	if cycle.Status == "COMPLETED" || cycle.Status == "CANCELLED" {
		return nil, fmt.Errorf("cannot update cycle in terminal state: %s", cycle.Status)
	}

	// Validate area allocation if area is being updated
	if updateReq.AreaHa != nil {
		if !cycle.CanModifyArea() {
			return nil, common.ErrStatusNotModifiable
		}
		if err := s.cropCycleRepo.ValidateAreaAllocation(ctx, cycle.FarmID, cycle.ID, *updateReq.AreaHa); err != nil {
			return nil, err
		}
	}

	// Update fields if provided
	if updateReq.AreaHa != nil {
		cycle.AreaHa = updateReq.AreaHa
	}
	if updateReq.Season != nil {
		cycle.Season = *updateReq.Season
	}
	if updateReq.StartDate != nil {
		cycle.StartDate = updateReq.StartDate
	}
	if updateReq.CropID != nil {
		cycle.CropID = *updateReq.CropID
	}
	if updateReq.VarietyID != nil {
		cycle.VarietyID = updateReq.VarietyID
	}

	// Validate the updated cycle
	if err := cycle.Validate(); err != nil {
		return nil, err
	}

	// Update the cycle in database
	if err := s.cropCycleRepo.Update(ctx, cycle); err != nil {
		return nil, fmt.Errorf("failed to update crop cycle: %w", err)
	}

	// Convert to response data
	cycleData := &responses.CropCycleData{
		ID:        cycle.ID,
		FarmID:    cycle.FarmID,
		FarmerID:  cycle.FarmerID,
		AreaHa:    cycle.AreaHa,
		Season:    cycle.Season,
		Status:    cycle.Status,
		StartDate: cycle.StartDate,
		EndDate:   cycle.EndDate,
		CropID:    cycle.CropID,
		VarietyID: cycle.VarietyID,
		CropName:  cycle.GetCropName(),
		VarietyName: func() *string {
			if name := cycle.GetVarietyName(); name != "" {
				return &name
			}
			return nil
		}(),
		Outcome:   cycle.Outcome,
		CreatedAt: cycle.CreatedAt,
		UpdatedAt: cycle.UpdatedAt,
	}

	return responses.NewCropCycleResponse(cycleData, "Crop cycle updated successfully"), nil
}

// EndCycle implements W12: End crop cycle
func (s *CropCycleServiceImpl) EndCycle(ctx context.Context, req interface{}) (interface{}, error) {
	endReq, ok := req.(*requests.EndCycleRequest)
	if !ok {
		return nil, common.ErrInvalidInput
	}

	// Extract authenticated user from context
	userCtx, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user context: %w", err)
	}

	// Check if authenticated user can end crop cycle
	hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "cycle", "end", endReq.ID, endReq.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Get existing cycle
	cycle := &cropCycleEntity.CropCycle{}
	_, err = s.cropCycleRepo.GetByID(ctx, endReq.ID, cycle)
	if err != nil {
		return nil, fmt.Errorf("failed to get crop cycle: %w", err)
	}

	// Check if cycle is already in terminal state
	if cycle.Status == "COMPLETED" || cycle.Status == "CANCELLED" {
		return nil, fmt.Errorf("cycle is already in terminal state: %s", cycle.Status)
	}

	// Update cycle with end details
	cycle.Status = endReq.Status
	cycle.EndDate = &endReq.EndDate
	if endReq.Outcome != nil {
		cycle.Outcome = endReq.Outcome
	}

	// Update the cycle in database
	if err := s.cropCycleRepo.Update(ctx, cycle); err != nil {
		return nil, fmt.Errorf("failed to end crop cycle: %w", err)
	}

	// Convert to response data
	cycleData := &responses.CropCycleData{
		ID:        cycle.ID,
		FarmID:    cycle.FarmID,
		FarmerID:  cycle.FarmerID,
		AreaHa:    cycle.AreaHa,
		Season:    cycle.Season,
		Status:    cycle.Status,
		StartDate: cycle.StartDate,
		EndDate:   cycle.EndDate,
		CropID:    cycle.CropID,
		VarietyID: cycle.VarietyID,
		CropName:  cycle.GetCropName(),
		VarietyName: func() *string {
			if name := cycle.GetVarietyName(); name != "" {
				return &name
			}
			return nil
		}(),
		Outcome:   cycle.Outcome,
		CreatedAt: cycle.CreatedAt,
		UpdatedAt: cycle.UpdatedAt,
	}

	return responses.NewCropCycleResponse(cycleData, "Crop cycle ended successfully"), nil
}

// ListCycles implements W13: List crop cycles
func (s *CropCycleServiceImpl) ListCycles(ctx context.Context, req interface{}) (interface{}, error) {
	listReq, ok := req.(*requests.ListCyclesRequest)
	if !ok {
		return nil, common.ErrInvalidInput
	}

	// Extract authenticated user from context
	userCtx, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user context: %w", err)
	}

	// Check if authenticated user can list crop cycles
	hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "cycle", "list", "", listReq.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Build filters
	filters := make(map[string]interface{})
	if listReq.FarmID != "" {
		filters["farm_id"] = listReq.FarmID
	}
	if listReq.FarmerID != "" {
		filters["farmer_id"] = listReq.FarmerID
	}
	if listReq.Season != "" {
		filters["season"] = listReq.Season
	}
	if listReq.Status != "" {
		filters["status"] = listReq.Status
	}

	// Build filter for database query
	filterBuilder := base.NewFilterBuilder()

	// Add filter conditions
	for key, value := range filters {
		filterBuilder.Where(key, base.OpEqual, value)
	}

	filter := filterBuilder.
		Limit(listReq.PageSize, (listReq.Page-1)*listReq.PageSize).
		Build()

	// Get cycles from database
	cycles, err := s.cropCycleRepo.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list crop cycles: %w", err)
	}

	// Get total count for pagination
	totalCount, err := s.cropCycleRepo.Count(ctx, filter, &cropCycleEntity.CropCycle{})
	if err != nil {
		return nil, fmt.Errorf("failed to count crop cycles: %w", err)
	}

	// Convert to response data
	var cycleDataList []*responses.CropCycleData
	for _, cycle := range cycles {
		cycleData := &responses.CropCycleData{
			ID:        cycle.ID,
			FarmID:    cycle.FarmID,
			FarmerID:  cycle.FarmerID,
			AreaHa:    cycle.AreaHa,
			Season:    cycle.Season,
			Status:    cycle.Status,
			StartDate: cycle.StartDate,
			EndDate:   cycle.EndDate,
			CropID:    cycle.CropID,
			VarietyID: cycle.VarietyID,
			CropName:  cycle.GetCropName(),
			VarietyName: func() *string {
				if name := cycle.GetVarietyName(); name != "" {
					return &name
				}
				return nil
			}(),
			Outcome:   cycle.Outcome,
			CreatedAt: cycle.CreatedAt,
			UpdatedAt: cycle.UpdatedAt,
		}
		cycleDataList = append(cycleDataList, cycleData)
	}

	return responses.NewCropCycleListResponse(cycleDataList, listReq.Page, listReq.PageSize, totalCount), nil
}

// GetCropCycle gets crop cycle by ID
func (s *CropCycleServiceImpl) GetCropCycle(ctx context.Context, cycleID string) (interface{}, error) {
	// Get cycle from database
	cycle := &cropCycleEntity.CropCycle{}
	_, err := s.cropCycleRepo.GetByID(ctx, cycleID, cycle)
	if err != nil {
		return nil, fmt.Errorf("failed to get crop cycle: %w", err)
	}

	// Convert to response data
	cycleData := &responses.CropCycleData{
		ID:        cycle.ID,
		FarmID:    cycle.FarmID,
		FarmerID:  cycle.FarmerID,
		AreaHa:    cycle.AreaHa,
		Season:    cycle.Season,
		Status:    cycle.Status,
		StartDate: cycle.StartDate,
		EndDate:   cycle.EndDate,
		CropID:    cycle.CropID,
		VarietyID: cycle.VarietyID,
		CropName:  cycle.GetCropName(),
		VarietyName: func() *string {
			if name := cycle.GetVarietyName(); name != "" {
				return &name
			}
			return nil
		}(),
		Outcome:   cycle.Outcome,
		CreatedAt: cycle.CreatedAt,
		UpdatedAt: cycle.UpdatedAt,
	}

	return responses.NewCropCycleResponse(cycleData, "Crop cycle retrieved successfully"), nil
}

// GetAreaAllocationSummary retrieves area allocation summary for a farm
func (s *CropCycleServiceImpl) GetAreaAllocationSummary(ctx context.Context, farmID string) (interface{}, error) {
	// Get summary from repository
	summary, err := s.cropCycleRepo.GetAreaAllocationSummary(ctx, farmID)
	if err != nil {
		return nil, fmt.Errorf("failed to get area allocation summary: %w", err)
	}

	// Calculate utilization percentage
	utilizationPercent := 0.0
	if summary.TotalAreaHa > 0 {
		utilizationPercent = (summary.AllocatedAreaHa / summary.TotalAreaHa) * 100
	}

	// Build response
	summaryData := &responses.AreaAllocationSummaryData{
		FarmID:             summary.FarmID,
		TotalAreaHa:        summary.TotalAreaHa,
		AllocatedAreaHa:    summary.AllocatedAreaHa,
		AvailableAreaHa:    summary.AvailableAreaHa,
		UtilizationPercent: utilizationPercent,
		ActiveCyclesCount:  summary.ActiveCyclesCount,
		PlannedCyclesCount: summary.PlannedCyclesCount,
	}

	return responses.NewAreaAllocationSummaryResponse(summaryData, "Area allocation summary retrieved successfully"), nil
}
