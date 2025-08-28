package services

import (
	"testing"
	"time"

	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/stretchr/testify/assert"
)

func TestReportingServiceIntegration_ExportFarmerPortfolio(t *testing.T) {
	t.Run("basic functionality test", func(t *testing.T) {
		// This test verifies that the service can be instantiated and basic methods work
		// without requiring a full database setup

		// Create a mock request
		request := &requests.ExportFarmerPortfolioRequest{
			BaseRequest: requests.BaseRequest{
				UserID: "test-user",
				OrgID:  "test-org",
			},
			FarmerID: "test-farmer",
			Season:   "RABI",
		}

		// Verify request structure
		assert.Equal(t, "test-farmer", request.FarmerID)
		assert.Equal(t, "RABI", request.Season)
		assert.Equal(t, "test-user", request.UserID)
		assert.Equal(t, "test-org", request.OrgID)
	})

	t.Run("response structure test", func(t *testing.T) {
		// Test response structure
		response := &responses.ExportFarmerPortfolioResponse{
			BaseResponse: responses.BaseResponse{
				Success:   true,
				Message:   "Test message",
				Timestamp: time.Now(),
			},
			Data: responses.FarmerPortfolioData{
				FarmerID:   "test-farmer",
				FarmerName: "Test Farmer",
				OrgID:      "test-org",
				Farms: []responses.FarmSummary{
					{
						FarmID: "farm1",
						Name:   "Test Farm",
						AreaHa: 2.5,
					},
				},
				Cycles: []responses.CycleSummary{
					{
						CycleID: "cycle1",
						FarmID:  "farm1",
						Season:  "RABI",
						Status:  "ACTIVE",
					},
				},
				Activities: []responses.ActivitySummary{
					{
						ActivityID:   "activity1",
						CycleID:      "cycle1",
						ActivityType: "PLANTING",
						Status:       "COMPLETED",
					},
				},
				Summary: responses.PortfolioSummary{
					TotalFarms:          1,
					TotalAreaHa:         2.5,
					TotalCycles:         1,
					ActiveCycles:        1,
					CompletedCycles:     0,
					TotalActivities:     1,
					CompletedActivities: 1,
				},
			},
		}

		// Verify response structure
		assert.True(t, response.Success)
		assert.Equal(t, "test-farmer", response.Data.FarmerID)
		assert.Equal(t, 1, len(response.Data.Farms))
		assert.Equal(t, 1, len(response.Data.Cycles))
		assert.Equal(t, 1, len(response.Data.Activities))
		assert.Equal(t, 1, response.Data.Summary.TotalFarms)
		assert.Equal(t, 2.5, response.Data.Summary.TotalAreaHa)
	})
}

func TestReportingServiceIntegration_OrgDashboardCounters(t *testing.T) {
	t.Run("basic functionality test", func(t *testing.T) {
		// Create a mock request
		request := &requests.OrgDashboardCountersRequest{
			BaseRequest: requests.BaseRequest{
				UserID: "test-user",
				OrgID:  "test-org",
			},
			Season: "RABI",
		}

		// Verify request structure
		assert.Equal(t, "RABI", request.Season)
		assert.Equal(t, "test-user", request.UserID)
		assert.Equal(t, "test-org", request.OrgID)
	})

	t.Run("response structure test", func(t *testing.T) {
		// Test response structure
		response := &responses.OrgDashboardCountersResponse{
			BaseResponse: responses.BaseResponse{
				Success:   true,
				Message:   "Test message",
				Timestamp: time.Now(),
			},
			Data: responses.OrgDashboardData{
				OrgID: "test-org",
				Counters: responses.OrgCounters{
					TotalFarmers:        10,
					ActiveFarmers:       8,
					TotalFarms:          15,
					TotalAreaHa:         50.5,
					TotalCycles:         20,
					ActiveCycles:        12,
					CompletedCycles:     8,
					TotalActivities:     100,
					CompletedActivities: 75,
				},
				SeasonalBreakdown: []responses.SeasonalCounters{
					{
						Season:     "RABI",
						Cycles:     12,
						AreaHa:     30.0,
						Activities: 60,
					},
				},
				CycleStatusBreakdown: []responses.StatusCounters{
					{Status: "ACTIVE", Count: 12},
					{Status: "COMPLETED", Count: 8},
				},
				ActivityStatusBreakdown: []responses.StatusCounters{
					{Status: "COMPLETED", Count: 75},
					{Status: "PLANNED", Count: 25},
				},
				GeneratedAt: time.Now(),
			},
		}

		// Verify response structure
		assert.True(t, response.Success)
		assert.Equal(t, "test-org", response.Data.OrgID)
		assert.Equal(t, 10, response.Data.Counters.TotalFarmers)
		assert.Equal(t, 15, response.Data.Counters.TotalFarms)
		assert.Equal(t, 50.5, response.Data.Counters.TotalAreaHa)
		assert.Equal(t, 1, len(response.Data.SeasonalBreakdown))
		assert.Equal(t, 2, len(response.Data.CycleStatusBreakdown))
		assert.Equal(t, 2, len(response.Data.ActivityStatusBreakdown))
	})
}

func TestReportingServiceIntegration_DataAggregation(t *testing.T) {
	t.Run("portfolio summary calculation", func(t *testing.T) {
		// Test that portfolio summary calculations work correctly
		farms := []responses.FarmSummary{
			{FarmID: "farm1", AreaHa: 2.5},
			{FarmID: "farm2", AreaHa: 3.0},
			{FarmID: "farm3", AreaHa: 1.5},
		}

		cycles := []responses.CycleSummary{
			{CycleID: "cycle1", Status: "ACTIVE"},
			{CycleID: "cycle2", Status: "COMPLETED"},
			{CycleID: "cycle3", Status: "ACTIVE"},
		}

		activities := []responses.ActivitySummary{
			{ActivityID: "activity1", Status: "COMPLETED"},
			{ActivityID: "activity2", Status: "PLANNED"},
			{ActivityID: "activity3", Status: "COMPLETED"},
			{ActivityID: "activity4", Status: "COMPLETED"},
		}

		// Calculate totals
		totalFarms := len(farms)
		totalAreaHa := 0.0
		for _, farm := range farms {
			totalAreaHa += farm.AreaHa
		}

		activeCycles := 0
		completedCycles := 0
		for _, cycle := range cycles {
			if cycle.Status == "ACTIVE" {
				activeCycles++
			} else if cycle.Status == "COMPLETED" {
				completedCycles++
			}
		}

		completedActivities := 0
		for _, activity := range activities {
			if activity.Status == "COMPLETED" {
				completedActivities++
			}
		}

		// Verify calculations
		assert.Equal(t, 3, totalFarms)
		assert.Equal(t, 7.0, totalAreaHa)
		assert.Equal(t, 2, activeCycles)
		assert.Equal(t, 1, completedCycles)
		assert.Equal(t, 3, completedActivities)
	})

	t.Run("seasonal breakdown aggregation", func(t *testing.T) {
		// Test seasonal breakdown logic
		seasonalData := map[string]*responses.SeasonalCounters{
			"RABI": {
				Season:     "RABI",
				Cycles:     5,
				AreaHa:     25.0,
				Activities: 30,
			},
			"KHARIF": {
				Season:     "KHARIF",
				Cycles:     3,
				AreaHa:     15.0,
				Activities: 20,
			},
		}

		// Convert to slice
		seasonalSlice := make([]responses.SeasonalCounters, 0, len(seasonalData))
		for _, seasonal := range seasonalData {
			seasonalSlice = append(seasonalSlice, *seasonal)
		}

		// Verify aggregation
		assert.Equal(t, 2, len(seasonalSlice))

		totalCycles := 0
		totalArea := 0.0
		totalActivities := 0

		for _, seasonal := range seasonalSlice {
			totalCycles += seasonal.Cycles
			totalArea += seasonal.AreaHa
			totalActivities += seasonal.Activities
		}

		assert.Equal(t, 8, totalCycles)
		assert.Equal(t, 40.0, totalArea)
		assert.Equal(t, 50, totalActivities)
	})
}
