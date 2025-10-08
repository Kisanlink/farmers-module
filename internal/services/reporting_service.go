package services

import (
	"context"
	"fmt"
	"time"

	"github.com/Kisanlink/farmers-module/internal/auth"
	"github.com/Kisanlink/farmers-module/internal/entities"
	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/farmers-module/internal/repo"
	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"gorm.io/gorm"
)

// ReportingServiceImpl implements the ReportingService interface
type ReportingServiceImpl struct {
	repoFactory *repo.RepositoryFactory
	db          *gorm.DB
	aaaService  AAAService
}

// NewReportingService creates a new reporting service
func NewReportingService(repoFactory *repo.RepositoryFactory, db *gorm.DB, aaaService AAAService) ReportingService {
	return &ReportingServiceImpl{
		repoFactory: repoFactory,
		db:          db,
		aaaService:  aaaService,
	}
}

// ExportFarmerPortfolio aggregates farms, cycles, and activities data for a farmer
func (s *ReportingServiceImpl) ExportFarmerPortfolio(ctx context.Context, req interface{}) (interface{}, error) {
	request, ok := req.(*requests.ExportFarmerPortfolioRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type")
	}

	// Extract authenticated user from context
	userCtx, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user context: %w", err)
	}

	// Check if authenticated user can export reports
	hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "report", "export", "", request.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Get farmer information
	farmer := &entities.FarmerProfile{}
	_, err = s.repoFactory.FarmerRepo.GetByID(ctx, request.FarmerID, farmer)
	if err != nil {
		return nil, fmt.Errorf("failed to get farmer: %w", err)
	}

	// Build filters for farms
	farmFilterBuilder := base.NewFilterBuilder().
		Where("aaa_farmer_user_id", base.OpEqual, farmer.AAAUserID).
		Where("aaa_org_id", base.OpEqual, request.OrgID)

	// Get farms
	farms, err := s.repoFactory.FarmRepo.Find(ctx, farmFilterBuilder.Build())
	if err != nil {
		return nil, fmt.Errorf("failed to get farms: %w", err)
	}

	// Prepare farm summaries and collect farm IDs
	farmSummaries := make([]responses.FarmSummary, 0, len(farms))
	farmIDs := make([]string, 0, len(farms))
	totalAreaHa := 0.0

	for _, farm := range farms {
		farmName := ""
		if farm.Name != nil {
			farmName = *farm.Name
		}
		farmSummaries = append(farmSummaries, responses.FarmSummary{
			FarmID: farm.ID,
			Name:   farmName,
			AreaHa: farm.AreaHa,
		})
		farmIDs = append(farmIDs, farm.ID)
		totalAreaHa += farm.AreaHa
	}

	// Get crop cycles for these farms
	cycleFilterBuilder := base.NewFilterBuilder().
		Where("farmer_id", base.OpEqual, request.FarmerID)

	if len(farmIDs) > 0 {
		cycleFilterBuilder = cycleFilterBuilder.Where("farm_id", base.OpIn, farmIDs)
	}
	if request.Season != "" {
		cycleFilterBuilder = cycleFilterBuilder.Where("season", base.OpEqual, request.Season)
	}
	if request.StartDate != nil {
		cycleFilterBuilder = cycleFilterBuilder.Where("start_date", base.OpGreaterEqual, *request.StartDate)
	}
	if request.EndDate != nil {
		cycleFilterBuilder = cycleFilterBuilder.Where("end_date", base.OpLessEqual, *request.EndDate)
	}

	cycles, err := s.repoFactory.CropCycleRepo.Find(ctx, cycleFilterBuilder.Build())
	if err != nil {
		return nil, fmt.Errorf("failed to get crop cycles: %w", err)
	}

	// Prepare cycle summaries and collect cycle IDs
	cycleSummaries := make([]responses.CycleSummary, 0, len(cycles))
	cycleIDs := make([]string, 0, len(cycles))
	activeCycles := 0
	completedCycles := 0

	for _, cycle := range cycles {
		cycleSummaries = append(cycleSummaries, responses.CycleSummary{
			CycleID:   cycle.ID,
			FarmID:    cycle.FarmID,
			Season:    cycle.Season,
			Status:    cycle.Status,
			StartDate: cycle.StartDate,
			EndDate:   cycle.EndDate,
			CropName:  cycle.GetCropName(),
		})
		cycleIDs = append(cycleIDs, cycle.ID)

		switch cycle.Status {
		case "ACTIVE", "PLANNED":
			activeCycles++
		case "COMPLETED":
			completedCycles++
		}
	}

	// Get farm activities for these cycles
	activityFilterBuilder := base.NewFilterBuilder()
	if len(cycleIDs) > 0 {
		activityFilterBuilder = activityFilterBuilder.Where("crop_cycle_id", base.OpIn, cycleIDs)
	}
	if request.StartDate != nil {
		activityFilterBuilder = activityFilterBuilder.Where("planned_at", base.OpGreaterEqual, *request.StartDate)
	}
	if request.EndDate != nil {
		activityFilterBuilder = activityFilterBuilder.Where("planned_at", base.OpLessEqual, *request.EndDate)
	}

	activities, err := s.repoFactory.FarmActivityRepo.Find(ctx, activityFilterBuilder.Build())
	if err != nil {
		return nil, fmt.Errorf("failed to get farm activities: %w", err)
	}

	// Prepare activity summaries
	activitySummaries := make([]responses.ActivitySummary, 0, len(activities))
	completedActivities := 0

	for _, activity := range activities {
		activitySummaries = append(activitySummaries, responses.ActivitySummary{
			ActivityID:   activity.ID,
			CycleID:      activity.CropCycleID,
			ActivityType: activity.ActivityType,
			Status:       activity.Status,
			PlannedAt:    activity.PlannedAt,
			CompletedAt:  activity.CompletedAt,
		})

		if activity.Status == "COMPLETED" {
			completedActivities++
		}
	}

	// Build portfolio data
	portfolioData := responses.FarmerPortfolioData{
		FarmerID:   request.FarmerID,
		FarmerName: fmt.Sprintf("%s %s", farmer.FirstName, farmer.LastName),
		OrgID:      request.OrgID,
		Farms:      farmSummaries,
		Cycles:     cycleSummaries,
		Activities: activitySummaries,
		Summary: responses.PortfolioSummary{
			TotalFarms:          len(farmSummaries),
			TotalAreaHa:         totalAreaHa,
			TotalCycles:         len(cycleSummaries),
			ActiveCycles:        activeCycles,
			CompletedCycles:     completedCycles,
			TotalActivities:     len(activitySummaries),
			CompletedActivities: completedActivities,
		},
	}

	return &responses.ExportFarmerPortfolioResponse{
		BaseResponse: responses.BaseResponse{
			Success:   true,
			Message:   "Farmer portfolio exported successfully",
			Timestamp: time.Now(),
		},
		Data: portfolioData,
	}, nil
}

// OrgDashboardCounters provides org-level KPIs including counts and areas by season/status
func (s *ReportingServiceImpl) OrgDashboardCounters(ctx context.Context, req interface{}) (interface{}, error) {
	request, ok := req.(*requests.OrgDashboardCountersRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type")
	}

	// Extract authenticated user from context
	userCtx, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user context: %w", err)
	}

	// Check if authenticated user can read organization dashboard
	hasPermission, err := s.aaaService.CheckPermission(ctx, userCtx.AAAUserID, "report", "read", "", request.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}
	if !hasPermission {
		return nil, common.ErrForbidden
	}

	// Get total farmers count
	farmerFilterBuilder := base.NewFilterBuilder().
		Where("aaa_org_id", base.OpEqual, request.OrgID)
	totalFarmers, err := s.repoFactory.FarmerRepo.Count(ctx, farmerFilterBuilder.Build(), &entities.FarmerProfile{})
	if err != nil {
		return nil, fmt.Errorf("failed to count farmers: %w", err)
	}

	// Get active farmers (those with active linkages)
	linkageFilterBuilder := base.NewFilterBuilder().
		Where("aaa_org_id", base.OpEqual, request.OrgID).
		Where("status", base.OpEqual, "ACTIVE")
	activeFarmers, err := s.repoFactory.FarmerLinkageRepo.Count(ctx, linkageFilterBuilder.Build(), &entities.FarmerLink{})
	if err != nil {
		return nil, fmt.Errorf("failed to count active farmers: %w", err)
	}

	// Get farms data
	farmFilterBuilder2 := base.NewFilterBuilder().
		Where("aaa_org_id", base.OpEqual, request.OrgID)
	farms, err := s.repoFactory.FarmRepo.Find(ctx, farmFilterBuilder2.Build())
	if err != nil {
		return nil, fmt.Errorf("failed to get farms: %w", err)
	}

	totalFarms := len(farms)
	totalAreaHa := 0.0
	farmIDs := make([]string, 0, len(farms))

	for _, farm := range farms {
		totalAreaHa += farm.AreaHa
		farmIDs = append(farmIDs, farm.ID)
	}

	// Get crop cycles data
	cycleFilterBuilder2 := base.NewFilterBuilder()
	if len(farmIDs) > 0 {
		cycleFilterBuilder2 = cycleFilterBuilder2.Where("farm_id", base.OpIn, farmIDs)
	}
	if request.Season != "" {
		cycleFilterBuilder2 = cycleFilterBuilder2.Where("season", base.OpEqual, request.Season)
	}
	if request.StartDate != nil {
		cycleFilterBuilder2 = cycleFilterBuilder2.Where("start_date", base.OpGreaterEqual, *request.StartDate)
	}
	if request.EndDate != nil {
		cycleFilterBuilder2 = cycleFilterBuilder2.Where("end_date", base.OpLessEqual, *request.EndDate)
	}

	cycles, err := s.repoFactory.CropCycleRepo.Find(ctx, cycleFilterBuilder2.Build())
	if err != nil {
		return nil, fmt.Errorf("failed to get crop cycles: %w", err)
	}

	// Analyze cycles
	totalCycles := len(cycles)
	activeCycles := 0
	completedCycles := 0
	seasonalBreakdown := make(map[string]*responses.SeasonalCounters)
	cycleStatusBreakdown := make(map[string]int)
	cycleIDs := make([]string, 0, len(cycles))

	for _, cycle := range cycles {
		cycleIDs = append(cycleIDs, cycle.ID)

		// Count by status
		switch cycle.Status {
		case "ACTIVE", "PLANNED":
			activeCycles++
		case "COMPLETED":
			completedCycles++
		}
		cycleStatusBreakdown[cycle.Status]++

		// Seasonal breakdown
		if seasonal, exists := seasonalBreakdown[cycle.Season]; exists {
			seasonal.Cycles++
		} else {
			seasonalBreakdown[cycle.Season] = &responses.SeasonalCounters{
				Season: cycle.Season,
				Cycles: 1,
				AreaHa: 0, // Will be calculated from farm data
			}
		}
	}

	// Calculate area by season (from farms associated with cycles)
	for _, cycle := range cycles {
		for _, farm := range farms {
			if farm.ID == cycle.FarmID {
				if seasonal, exists := seasonalBreakdown[cycle.Season]; exists {
					seasonal.AreaHa += farm.AreaHa
				}
				break
			}
		}
	}

	// Get farm activities data
	activityFilterBuilder2 := base.NewFilterBuilder()
	if len(cycleIDs) > 0 {
		activityFilterBuilder2 = activityFilterBuilder2.Where("crop_cycle_id", base.OpIn, cycleIDs)
	}
	if request.StartDate != nil {
		activityFilterBuilder2 = activityFilterBuilder2.Where("planned_at", base.OpGreaterEqual, *request.StartDate)
	}
	if request.EndDate != nil {
		activityFilterBuilder2 = activityFilterBuilder2.Where("planned_at", base.OpLessEqual, *request.EndDate)
	}

	activities, err := s.repoFactory.FarmActivityRepo.Find(ctx, activityFilterBuilder2.Build())
	if err != nil {
		return nil, fmt.Errorf("failed to get farm activities: %w", err)
	}

	// Analyze activities
	totalActivities := len(activities)
	completedActivities := 0
	activityStatusBreakdown := make(map[string]int)

	for _, activity := range activities {
		if activity.Status == "COMPLETED" {
			completedActivities++
		}
		activityStatusBreakdown[activity.Status]++

		// Add to seasonal breakdown
		for _, cycle := range cycles {
			if cycle.ID == activity.CropCycleID {
				if seasonal, exists := seasonalBreakdown[cycle.Season]; exists {
					seasonal.Activities++
				}
				break
			}
		}
	}

	// Convert maps to slices
	seasonalSlice := make([]responses.SeasonalCounters, 0, len(seasonalBreakdown))
	for _, seasonal := range seasonalBreakdown {
		seasonalSlice = append(seasonalSlice, *seasonal)
	}

	cycleStatusSlice := make([]responses.StatusCounters, 0, len(cycleStatusBreakdown))
	for status, count := range cycleStatusBreakdown {
		cycleStatusSlice = append(cycleStatusSlice, responses.StatusCounters{
			Status: status,
			Count:  count,
		})
	}

	activityStatusSlice := make([]responses.StatusCounters, 0, len(activityStatusBreakdown))
	for status, count := range activityStatusBreakdown {
		activityStatusSlice = append(activityStatusSlice, responses.StatusCounters{
			Status: status,
			Count:  count,
		})
	}

	// Build dashboard data
	dashboardData := responses.OrgDashboardData{
		OrgID: request.OrgID,
		Counters: responses.OrgCounters{
			TotalFarmers:        int(totalFarmers),
			ActiveFarmers:       int(activeFarmers),
			TotalFarms:          totalFarms,
			TotalAreaHa:         totalAreaHa,
			TotalCycles:         totalCycles,
			ActiveCycles:        activeCycles,
			CompletedCycles:     completedCycles,
			TotalActivities:     totalActivities,
			CompletedActivities: completedActivities,
		},
		SeasonalBreakdown:       seasonalSlice,
		CycleStatusBreakdown:    cycleStatusSlice,
		ActivityStatusBreakdown: activityStatusSlice,
		GeneratedAt:             time.Now(),
	}

	return &responses.OrgDashboardCountersResponse{
		BaseResponse: responses.BaseResponse{
			Success:   true,
			Message:   "Organization dashboard counters retrieved successfully",
			Timestamp: time.Now(),
		},
		Data: dashboardData,
	}, nil
}
