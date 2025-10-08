package services

import (
	"context"
	"fmt"
	"time"

	"github.com/Kisanlink/farmers-module/internal/auth"
	cropCycleEntity "github.com/Kisanlink/farmers-module/internal/entities/crop_cycle"
	farmActivityEntity "github.com/Kisanlink/farmers-module/internal/entities/farm_activity"
	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// FarmActivityServiceImpl implements FarmActivityService
type FarmActivityServiceImpl struct {
	farmActivityRepo *base.BaseFilterableRepository[*farmActivityEntity.FarmActivity]
	cropCycleRepo    *base.BaseFilterableRepository[*cropCycleEntity.CropCycle]
	farmerLinkRepo   FarmerLinkRepository
	aaaService       AAAService
}

// NewFarmActivityService creates a new farm activity service
func NewFarmActivityService(
	farmActivityRepo *base.BaseFilterableRepository[*farmActivityEntity.FarmActivity],
	cropCycleRepo *base.BaseFilterableRepository[*cropCycleEntity.CropCycle],
	farmerLinkRepo FarmerLinkRepository,
	aaaService AAAService,
) FarmActivityService {
	return &FarmActivityServiceImpl{
		farmActivityRepo: farmActivityRepo,
		cropCycleRepo:    cropCycleRepo,
		farmerLinkRepo:   farmerLinkRepo,
		aaaService:       aaaService,
	}
}

// CreateActivity implements W14: Create farm activity
func (s *FarmActivityServiceImpl) CreateActivity(ctx context.Context, req interface{}) (interface{}, error) {
	createReq, ok := req.(*requests.CreateActivityRequest)
	if !ok {
		return nil, common.ErrInvalidInput
	}

	// Extract authenticated user from context
	userCtx, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user context: %w", err)
	}

	// Check if authenticated user can create activity
	hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "activity", "create", "", createReq.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Validate crop cycle exists and user has access
	cropCycle := &cropCycleEntity.CropCycle{}
	_, err = s.cropCycleRepo.GetByID(ctx, createReq.CropCycleID, cropCycle)
	if err != nil {
		return nil, fmt.Errorf("failed to get crop cycle: %w", err)
	}

	// Business Rule 9.2: KisanSathi permission scope check
	// KisanSathi users can only create activities for farmers they are assigned to
	isKisanSathi, err := s.aaaService.CheckUserRole(ctx, userCtx.AAAUserID, "kisansathi")
	if err != nil {
		return nil, fmt.Errorf("failed to check user role: %w", err)
	}

	if isKisanSathi {
		// Get farmer link to verify KisanSathi assignment
		farmerLinkFilter := base.NewFilterBuilder().
			Where("aaa_user_id", base.OpEqual, cropCycle.FarmerID).
			Where("status", base.OpEqual, "ACTIVE").
			Build()

		farmerLinks, err := s.farmerLinkRepo.Find(ctx, farmerLinkFilter)
		if err != nil || len(farmerLinks) == 0 {
			return nil, fmt.Errorf("farmer link not found for farmer %s", cropCycle.FarmerID)
		}

		farmerLink := farmerLinks[0]
		if farmerLink.KisanSathiUserID == nil || *farmerLink.KisanSathiUserID != userCtx.AAAUserID {
			return nil, fmt.Errorf("KisanSathi can only create activities for assigned farmers")
		}
	}

	// Create farm activity entity
	activity := &farmActivityEntity.FarmActivity{
		CropCycleID:  createReq.CropCycleID,
		ActivityType: createReq.ActivityType,
		PlannedAt:    createReq.PlannedAt,
		CreatedBy:    userCtx.AAAUserID,
		Status:       "PLANNED",
		Metadata:     createReq.Metadata,
	}

	// Initialize empty maps if nil
	if activity.Output == nil {
		activity.Output = make(map[string]string)
	}
	if activity.Metadata == nil {
		activity.Metadata = make(map[string]string)
	}

	// Validate the activity
	if err := activity.Validate(); err != nil {
		return nil, err
	}

	// Create the activity in database
	if err := s.farmActivityRepo.Create(ctx, activity); err != nil {
		return nil, fmt.Errorf("failed to create farm activity: %w", err)
	}

	// Convert to response data
	activityData := &responses.FarmActivityData{
		ID:           activity.ID,
		CropCycleID:  activity.CropCycleID,
		ActivityType: activity.ActivityType,
		PlannedAt:    activity.PlannedAt,
		CompletedAt:  activity.CompletedAt,
		CreatedBy:    activity.CreatedBy,
		Status:       activity.Status,
		Output:       activity.Output,
		Metadata:     activity.Metadata,
		CreatedAt:    activity.CreatedAt,
		UpdatedAt:    activity.UpdatedAt,
	}

	response := responses.NewFarmActivityResponse(activityData, "Farm activity created successfully")
	return &response, nil
}

// CompleteActivity implements W15: Complete farm activity
func (s *FarmActivityServiceImpl) CompleteActivity(ctx context.Context, req interface{}) (interface{}, error) {
	completeReq, ok := req.(*requests.CompleteActivityRequest)
	if !ok {
		return nil, common.ErrInvalidInput
	}

	// Extract authenticated user from context
	userCtx, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user context: %w", err)
	}

	// Check if authenticated user can complete activity
	hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "activity", "complete", completeReq.ID, completeReq.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Get existing activity
	activity := &farmActivityEntity.FarmActivity{}
	_, err = s.farmActivityRepo.GetByID(ctx, completeReq.ID, activity)
	if err != nil {
		return nil, fmt.Errorf("failed to get farm activity: %w", err)
	}

	// Check if activity is already completed
	if activity.Status == "COMPLETED" {
		return nil, fmt.Errorf("activity is already completed")
	}

	// Update activity with completion details
	activity.Status = "COMPLETED"
	activity.CompletedAt = &completeReq.CompletedAt
	if completeReq.Output != nil {
		activity.Output = completeReq.Output
	}

	// Update the activity in database
	if err := s.farmActivityRepo.Update(ctx, activity); err != nil {
		return nil, fmt.Errorf("failed to complete farm activity: %w", err)
	}

	// Convert to response data
	activityData := &responses.FarmActivityData{
		ID:           activity.ID,
		CropCycleID:  activity.CropCycleID,
		ActivityType: activity.ActivityType,
		PlannedAt:    activity.PlannedAt,
		CompletedAt:  activity.CompletedAt,
		CreatedBy:    activity.CreatedBy,
		Status:       activity.Status,
		Output:       activity.Output,
		Metadata:     activity.Metadata,
		CreatedAt:    activity.CreatedAt,
		UpdatedAt:    activity.UpdatedAt,
	}

	response := responses.NewFarmActivityResponse(activityData, "Farm activity completed successfully")
	return &response, nil
}

// UpdateActivity implements W16: Update farm activity
func (s *FarmActivityServiceImpl) UpdateActivity(ctx context.Context, req interface{}) (interface{}, error) {
	updateReq, ok := req.(*requests.UpdateActivityRequest)
	if !ok {
		return nil, common.ErrInvalidInput
	}

	// Extract authenticated user from context
	userCtx, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user context: %w", err)
	}

	// Check if authenticated user can update activity
	hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "activity", "update", updateReq.ID, updateReq.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Get existing activity
	activity := &farmActivityEntity.FarmActivity{}
	_, err = s.farmActivityRepo.GetByID(ctx, updateReq.ID, activity)
	if err != nil {
		return nil, fmt.Errorf("failed to get farm activity: %w", err)
	}

	// Check if activity is completed (restrict updates to completed activities)
	if activity.Status == "COMPLETED" {
		return nil, fmt.Errorf("cannot update completed activity")
	}

	// Update fields if provided
	if updateReq.ActivityType != nil {
		activity.ActivityType = *updateReq.ActivityType
	}
	if updateReq.PlannedAt != nil {
		activity.PlannedAt = updateReq.PlannedAt
	}
	if updateReq.Metadata != nil {
		activity.Metadata = updateReq.Metadata
	}

	// Validate the updated activity
	if err := activity.Validate(); err != nil {
		return nil, err
	}

	// Update the activity in database
	if err := s.farmActivityRepo.Update(ctx, activity); err != nil {
		return nil, fmt.Errorf("failed to update farm activity: %w", err)
	}

	// Convert to response data
	activityData := &responses.FarmActivityData{
		ID:           activity.ID,
		CropCycleID:  activity.CropCycleID,
		ActivityType: activity.ActivityType,
		PlannedAt:    activity.PlannedAt,
		CompletedAt:  activity.CompletedAt,
		CreatedBy:    activity.CreatedBy,
		Status:       activity.Status,
		Output:       activity.Output,
		Metadata:     activity.Metadata,
		CreatedAt:    activity.CreatedAt,
		UpdatedAt:    activity.UpdatedAt,
	}

	response := responses.NewFarmActivityResponse(activityData, "Farm activity updated successfully")
	return &response, nil
}

// ListActivities implements W17: List farm activities
func (s *FarmActivityServiceImpl) ListActivities(ctx context.Context, req interface{}) (interface{}, error) {
	listReq, ok := req.(*requests.ListActivitiesRequest)
	if !ok {
		return nil, common.ErrInvalidInput
	}

	// Extract authenticated user from context
	userCtx, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user context: %w", err)
	}

	// Check if authenticated user can list activities
	hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "activity", "list", "", listReq.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Build filters
	filters := make(map[string]interface{})
	if listReq.CropCycleID != "" {
		filters["crop_cycle_id"] = listReq.CropCycleID
	}
	if listReq.ActivityType != "" {
		filters["activity_type"] = listReq.ActivityType
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

	// Add date range filters if provided
	if listReq.DateFrom != "" {
		dateFrom, err := time.Parse("2006-01-02", listReq.DateFrom)
		if err != nil {
			return nil, fmt.Errorf("invalid date_from format: %w", err)
		}
		filterBuilder.Where("planned_at", base.OpGreaterEqual, dateFrom)
	}
	if listReq.DateTo != "" {
		dateTo, err := time.Parse("2006-01-02", listReq.DateTo)
		if err != nil {
			return nil, fmt.Errorf("invalid date_to format: %w", err)
		}
		filterBuilder.Where("planned_at", base.OpLessEqual, dateTo)
	}

	filter := filterBuilder.
		Limit(listReq.PageSize, (listReq.Page-1)*listReq.PageSize).
		Build()

	// Get activities from database
	activities, err := s.farmActivityRepo.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list farm activities: %w", err)
	}

	// Get total count for pagination
	totalCount, err := s.farmActivityRepo.Count(ctx, filter, &farmActivityEntity.FarmActivity{})
	if err != nil {
		return nil, fmt.Errorf("failed to count farm activities: %w", err)
	}

	// Convert to response data
	var activityDataList []*responses.FarmActivityData
	for _, activity := range activities {
		activityData := &responses.FarmActivityData{
			ID:           activity.ID,
			CropCycleID:  activity.CropCycleID,
			ActivityType: activity.ActivityType,
			PlannedAt:    activity.PlannedAt,
			CompletedAt:  activity.CompletedAt,
			CreatedBy:    activity.CreatedBy,
			Status:       activity.Status,
			Output:       activity.Output,
			Metadata:     activity.Metadata,
			CreatedAt:    activity.CreatedAt,
			UpdatedAt:    activity.UpdatedAt,
		}
		activityDataList = append(activityDataList, activityData)
	}

	response := responses.NewFarmActivityListResponse(activityDataList, listReq.Page, listReq.PageSize, totalCount)
	return &response, nil
}

// GetFarmActivity gets farm activity by ID
func (s *FarmActivityServiceImpl) GetFarmActivity(ctx context.Context, activityID string) (interface{}, error) {
	// Get activity from database
	activity := &farmActivityEntity.FarmActivity{}
	_, err := s.farmActivityRepo.GetByID(ctx, activityID, activity)
	if err != nil {
		return nil, fmt.Errorf("failed to get farm activity: %w", err)
	}

	// Convert to response data
	activityData := &responses.FarmActivityData{
		ID:           activity.ID,
		CropCycleID:  activity.CropCycleID,
		ActivityType: activity.ActivityType,
		PlannedAt:    activity.PlannedAt,
		CompletedAt:  activity.CompletedAt,
		CreatedBy:    activity.CreatedBy,
		Status:       activity.Status,
		Output:       activity.Output,
		Metadata:     activity.Metadata,
		CreatedAt:    activity.CreatedAt,
		UpdatedAt:    activity.UpdatedAt,
	}

	response := responses.NewFarmActivityResponse(activityData, "Farm activity retrieved successfully")
	return &response, nil
}
