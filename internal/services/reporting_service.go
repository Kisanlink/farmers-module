package services

import (
	"context"
	"fmt"
	"time"

	"github.com/Kisanlink/farmers-module/internal/entities/crop_cycle"
	"github.com/Kisanlink/farmers-module/internal/entities/farm"
	"github.com/Kisanlink/farmers-module/internal/entities/farm_activity"
	"github.com/Kisanlink/farmers-module/internal/entities/farmer"
	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/Kisanlink/farmers-module/internal/repo"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// ReportingServiceImpl implements ReportingService
type ReportingServiceImpl struct {
	repoFactory *repo.RepositoryFactory
	aaaService  interfaces.AAAService
}

// NewReportingService creates a new reporting service
func NewReportingService(repoFactory *repo.RepositoryFactory, aaaService interfaces.AAAService) *ReportingServiceImpl {
	return &ReportingServiceImpl{
		repoFactory: repoFactory,
		aaaService:  aaaService,
	}
}

// ExportFarmerPortfolio aggregates farms, cycles, and activities data for a farmer
func (s *ReportingServiceImpl) ExportFarmerPortfolio(ctx context.Context, req interface{}) (interface{}, error) {
	// Type assertion to get the actual request
	request, ok := req.(*requests.ExportFarmerPortfolioRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type")
	}

	// Check permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, request.UserID, "farmer", "read", request.FarmerID, request.OrgID)
	if err != nil {
		return nil, fmt.Errorf("permission check failed: %w", err)
	}
	if !hasPermission {
		return nil, fmt.Errorf("insufficient permissions to read farmer portfolio")
	}

	// Get farmer information
	farmerEntity := &farmer.Farmer{}
	_, err = s.repoFactory.FarmerRepo.GetByID(ctx, request.FarmerID, farmerEntity)
	if err != nil {
		return nil, fmt.Errorf("failed to get farmer: %w", err)
	}

	// Get farm information
	farmFilter := base.NewFilterBuilder().
		Where("aaa_farmer_user_id", base.OpEqual, request.FarmerID).
		Build()

	farms, err := s.repoFactory.FarmRepo.Find(ctx, farmFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to get farms: %w", err)
	}

	// Get crop cycles for all farms
	var allCropCycles []*crop_cycle.CropCycle
	for _, farmEntity := range farms {
		cycleFilter := base.NewFilterBuilder().
			Where("farm_id", base.OpEqual, farmEntity.ID).
			Build()

		cycles, err := s.repoFactory.CropCycleRepo.Find(ctx, cycleFilter)
		if err != nil {
			return nil, fmt.Errorf("failed to get crop cycles: %w", err)
		}
		allCropCycles = append(allCropCycles, cycles...)
	}

	// Get farm activities for all crop cycles
	var allActivities []*farm_activity.FarmActivity
	for _, cycle := range allCropCycles {
		activityFilter := base.NewFilterBuilder().
			Where("crop_cycle_id", base.OpEqual, cycle.ID).
			Build()

		activities, err := s.repoFactory.FarmActivityRepo.Find(ctx, activityFilter)
		if err != nil {
			return nil, fmt.Errorf("failed to get farm activities: %w", err)
		}
		allActivities = append(allActivities, activities...)
	}

	// Create response
	response := &responses.ExportFarmerPortfolioResponse{
		BaseResponse: responses.BaseResponse{
			RequestID: request.RequestID,
		},
		Data: responses.FarmerPortfolioData{
			FarmerID:   farmerEntity.ID,
			FarmerName: farmerEntity.FirstName + " " + farmerEntity.LastName,
			OrgID:      farmerEntity.AAAOrgID,
			Farms:      convertFarmsToSummary(farms),
			Cycles:     convertCropCyclesToSummary(allCropCycles),
			Activities: convertActivitiesToSummary(allActivities),
			Summary: responses.PortfolioSummary{
				TotalFarms:          len(farms),
				TotalAreaHa:         calculateTotalArea(farms),
				TotalCycles:         len(allCropCycles),
				ActiveCycles:        countActiveCycles(allCropCycles),
				CompletedCycles:     countCompletedCycles(allCropCycles),
				TotalActivities:     len(allActivities),
				CompletedActivities: countCompletedActivities(allActivities),
			},
		},
	}

	return response, nil
}

// OrgDashboardCounters provides org-level KPIs including counts and areas by season/status
func (s *ReportingServiceImpl) OrgDashboardCounters(ctx context.Context, req interface{}) (interface{}, error) {
	// Type assertion to get the actual request
	request, ok := req.(*requests.OrgDashboardCountersRequest)
	if !ok {
		return nil, fmt.Errorf("invalid request type")
	}

	// Check permission
	hasPermission, err := s.aaaService.CheckPermission(ctx, request.UserID, "fpo", "read", "", request.OrgID)
	if err != nil {
		return nil, fmt.Errorf("permission check failed: %w", err)
	}
	if !hasPermission {
		return nil, fmt.Errorf("insufficient permissions to read FPO analytics")
	}

	// Get farmer count
	farmerFilterBuilder := base.NewFilterBuilder().
		Where("aaa_org_id", base.OpEqual, request.OrgID)
	totalFarmers, err := s.repoFactory.FarmerRepo.Count(ctx, farmerFilterBuilder.Build(), &farmer.Farmer{})
	if err != nil {
		return nil, fmt.Errorf("failed to count farmers: %w", err)
	}

	// Get farm count and total area
	farmFilterBuilder := base.NewFilterBuilder().
		Where("aaa_org_id", base.OpEqual, request.OrgID)
	totalFarms, err := s.repoFactory.FarmRepo.Count(ctx, farmFilterBuilder.Build(), &farm.Farm{})
	if err != nil {
		return nil, fmt.Errorf("failed to count farms: %w", err)
	}

	// Get total area
	farms, err := s.repoFactory.FarmRepo.Find(ctx, farmFilterBuilder.Build())
	if err != nil {
		return nil, fmt.Errorf("failed to get farms: %w", err)
	}
	totalArea := calculateTotalArea(farms)

	// Get crop cycle counts by status
	cycleFilterBuilder := base.NewFilterBuilder().
		Where("aaa_org_id", base.OpEqual, request.OrgID)
	allCycles, err := s.repoFactory.CropCycleRepo.Find(ctx, cycleFilterBuilder.Build())
	if err != nil {
		return nil, fmt.Errorf("failed to get crop cycles: %w", err)
	}

	// Create response
	response := &responses.OrgDashboardCountersResponse{
		BaseResponse: responses.BaseResponse{
			RequestID: request.RequestID,
		},
		Data: responses.OrgDashboardData{
			OrgID: request.OrgID,
			Counters: responses.OrgCounters{
				TotalFarmers:    int(totalFarmers),
				TotalFarms:      int(totalFarms),
				TotalAreaHa:     totalArea,
				ActiveCycles:    countActiveCycles(allCycles),
				CompletedCycles: countCompletedCycles(allCycles),
				TotalCycles:     len(allCycles),
			},
			GeneratedAt: time.Now(),
		},
	}

	return response, nil
}

// Helper functions

func convertFarmsToSummary(farms []*farm.Farm) []responses.FarmSummary {
	var farmSummaryList []responses.FarmSummary
	for _, farmEntity := range farms {
		farmSummary := responses.FarmSummary{
			FarmID: farmEntity.ID,
			Name:   *farmEntity.Name,
			AreaHa: farmEntity.AreaHa,
		}
		farmSummaryList = append(farmSummaryList, farmSummary)
	}
	return farmSummaryList
}

func convertCropCyclesToSummary(cycles []*crop_cycle.CropCycle) []responses.CycleSummary {
	var cycleSummaryList []responses.CycleSummary
	for _, cycle := range cycles {
		cycleSummary := responses.CycleSummary{
			CycleID:   cycle.ID,
			FarmID:    cycle.FarmID,
			Season:    cycle.Season,
			Status:    cycle.Status,
			StartDate: cycle.StartDate,
			EndDate:   cycle.EndDate,
		}
		cycleSummaryList = append(cycleSummaryList, cycleSummary)
	}
	return cycleSummaryList
}

func convertActivitiesToSummary(activities []*farm_activity.FarmActivity) []responses.ActivitySummary {
	var activitySummaryList []responses.ActivitySummary
	for _, activity := range activities {
		activitySummary := responses.ActivitySummary{
			ActivityID:   activity.ID,
			CycleID:      activity.CropCycleID,
			ActivityType: activity.ActivityType,
			Status:       activity.Status,
			PlannedAt:    activity.PlannedAt,
			CompletedAt:  activity.CompletedAt,
		}
		activitySummaryList = append(activitySummaryList, activitySummary)
	}
	return activitySummaryList
}

func calculateTotalArea(farms []*farm.Farm) float64 {
	total := 0.0
	for _, farmEntity := range farms {
		total += farmEntity.AreaHa
	}
	return total
}

func countActiveCycles(cycles []*crop_cycle.CropCycle) int {
	count := 0
	for _, cycle := range cycles {
		if cycle.Status == "ACTIVE" {
			count++
		}
	}
	return count
}

func countCompletedCycles(cycles []*crop_cycle.CropCycle) int {
	count := 0
	for _, cycle := range cycles {
		if cycle.Status == "COMPLETED" {
			count++
		}
	}
	return count
}

func countCompletedActivities(activities []*farm_activity.FarmActivity) int {
	count := 0
	for _, activity := range activities {
		if activity.Status == "COMPLETED" {
			count++
		}
	}
	return count
}
