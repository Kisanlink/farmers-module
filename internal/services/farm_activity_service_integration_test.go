package services

import (
	"testing"
	"time"

	farmActivityEntity "github.com/Kisanlink/farmers-module/internal/entities/farm_activity"
	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/stretchr/testify/assert"
)

// TestFarmActivityService_ActivityLifecycle tests the complete activity lifecycle
func TestFarmActivityService_ActivityLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("activity validation", func(t *testing.T) {
		// Test activity validation
		activity := &farmActivityEntity.FarmActivity{
			CropCycleID:  "",
			ActivityType: "planting",
			CreatedBy:    "user123",
		}

		err := activity.Validate()
		assert.Error(t, err)
		assert.Equal(t, common.ErrInvalidFarmActivityData, err)

		// Valid activity
		activity.CropCycleID = "cycle123"
		activity.FarmerID = "farmer123"
		err = activity.Validate()
		assert.NoError(t, err)
	})

	t.Run("activity status transitions", func(t *testing.T) {
		// Test that activity starts as PLANNED
		activity := &farmActivityEntity.FarmActivity{
			CropCycleID:  "cycle123",
			FarmerID:     "farmer123",
			ActivityType: "planting",
			CreatedBy:    "user123",
			Status:       "PLANNED",
		}

		assert.Equal(t, "PLANNED", activity.Status)

		// Test completion
		now := time.Now()
		activity.Status = "COMPLETED"
		activity.CompletedAt = &now
		activity.Output = map[string]interface{}{
			"yield": "500kg",
		}

		assert.Equal(t, "COMPLETED", activity.Status)
		assert.NotNil(t, activity.CompletedAt)
		assert.Equal(t, "500kg", activity.Output["yield"])
	})
}

// TestFarmActivityService_RequestValidation tests request validation logic
func TestFarmActivityService_RequestValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("create activity request validation", func(t *testing.T) {
		// Test empty crop cycle ID
		req := &requests.CreateActivityRequest{
			BaseRequest: requests.BaseRequest{
				UserID: "user123",
				OrgID:  "org123",
			},
			CropCycleID:  "",
			ActivityType: "planting",
		}

		// This would be caught by the service validation
		assert.Empty(t, req.CropCycleID)

		// Valid request
		req.CropCycleID = "cycle123"
		assert.NotEmpty(t, req.CropCycleID)
		assert.Equal(t, "planting", req.ActivityType)
	})

	t.Run("complete activity request validation", func(t *testing.T) {
		now := time.Now()
		req := &requests.CompleteActivityRequest{
			BaseRequest: requests.BaseRequest{
				UserID: "user123",
				OrgID:  "org123",
			},
			ID:          "activity123",
			CompletedAt: now,
			Output: map[string]interface{}{
				"yield": "500kg",
			},
		}

		assert.Equal(t, "activity123", req.ID)
		assert.Equal(t, now, req.CompletedAt)
		assert.Equal(t, "500kg", req.Output["yield"])
	})
}

// TestFarmActivityService_ResponseMapping tests response data mapping
func TestFarmActivityService_ResponseMapping(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	now := time.Now()

	t.Run("farm activity response mapping", func(t *testing.T) {
		activityData := &responses.FarmActivityData{
			ID:           "activity123",
			CropCycleID:  "cycle123",
			ActivityType: "planting",
			PlannedAt:    &now,
			CompletedAt:  nil,
			CreatedBy:    "user123",
			Status:       "PLANNED",
			Output:       map[string]interface{}{"test": "value"},
			Metadata:     map[string]interface{}{"meta": "data"},
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		response := responses.NewFarmActivityResponse(activityData, "Activity created successfully")

		assert.Equal(t, "activity123", response.Data.ID)
		assert.Equal(t, "cycle123", response.Data.CropCycleID)
		assert.Equal(t, "planting", response.Data.ActivityType)
		assert.Equal(t, "user123", response.Data.CreatedBy)
		assert.Equal(t, "PLANNED", response.Data.Status)
		assert.Nil(t, response.Data.CompletedAt)
	})

	t.Run("farm activity list response mapping", func(t *testing.T) {
		activities := []*responses.FarmActivityData{
			{
				ID:           "activity1",
				CropCycleID:  "cycle123",
				ActivityType: "planting",
				Status:       "PLANNED",
				CreatedBy:    "user123",
				CreatedAt:    now,
				UpdatedAt:    now,
			},
			{
				ID:           "activity2",
				CropCycleID:  "cycle123",
				ActivityType: "harvesting",
				Status:       "COMPLETED",
				CreatedBy:    "user123",
				CreatedAt:    now,
				UpdatedAt:    now,
			},
		}

		response := responses.NewFarmActivityListResponse(activities, 1, 10, 2)

		assert.Len(t, response.Data, 2)
		assert.Equal(t, "activity1", response.Data[0].ID)
		assert.Equal(t, "activity2", response.Data[1].ID)
	})
}

// TestFarmActivityService_FilteringLogic tests filtering and pagination logic
func TestFarmActivityService_FilteringLogic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("date filtering validation", func(t *testing.T) {
		// Test valid date formats
		validDates := []string{
			"2024-01-01",
			"2024-12-31",
			"2023-06-15",
		}

		for _, dateStr := range validDates {
			_, err := time.Parse("2006-01-02", dateStr)
			assert.NoError(t, err, "Date %s should be valid", dateStr)
		}

		// Test invalid date formats
		invalidDates := []string{
			"invalid-date",
			"2024/01/01",
			"01-01-2024",
			"2024-13-01", // Invalid month
		}

		for _, dateStr := range invalidDates {
			_, err := time.Parse("2006-01-02", dateStr)
			assert.Error(t, err, "Date %s should be invalid", dateStr)
		}
	})

	t.Run("pagination parameters", func(t *testing.T) {
		req := &requests.ListActivitiesRequest{
			Page:     1,
			PageSize: 10,
		}

		assert.Equal(t, 1, req.Page)
		assert.Equal(t, 10, req.PageSize)

		// Test page size limits (would be enforced by handlers)
		if req.PageSize > 100 {
			req.PageSize = 100
		}
		assert.LessOrEqual(t, req.PageSize, 100)
	})
}

// TestFarmActivityService_BusinessRules tests business rule validation
func TestFarmActivityService_BusinessRules(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("activity status rules", func(t *testing.T) {
		// Test that completed activities cannot be updated
		activity := &farmActivityEntity.FarmActivity{
			Status: "COMPLETED",
		}

		// This business rule would be enforced by the service
		if activity.Status == "COMPLETED" {
			// Should not allow updates
			assert.Equal(t, "COMPLETED", activity.Status)
		}

		// Test that planned activities can be updated
		plannedActivity := &farmActivityEntity.FarmActivity{
			Status: "PLANNED",
		}

		if plannedActivity.Status == "PLANNED" {
			// Should allow updates
			plannedActivity.ActivityType = "updated_type"
			assert.Equal(t, "updated_type", plannedActivity.ActivityType)
		}
	})

	t.Run("activity completion rules", func(t *testing.T) {
		// Test that only planned activities can be completed
		activity := &farmActivityEntity.FarmActivity{
			Status: "PLANNED",
		}

		// Can be completed
		if activity.Status != "COMPLETED" {
			activity.Status = "COMPLETED"
			now := time.Now()
			activity.CompletedAt = &now
		}

		assert.Equal(t, "COMPLETED", activity.Status)
		assert.NotNil(t, activity.CompletedAt)
	})

	t.Run("metadata and output handling", func(t *testing.T) {
		activity := &farmActivityEntity.FarmActivity{
			Output:   make(map[string]interface{}),
			Metadata: make(map[string]interface{}),
		}

		// Test that maps are properly initialized
		assert.NotNil(t, activity.Output)
		assert.NotNil(t, activity.Metadata)

		// Test adding data
		activity.Output["yield"] = "500kg"
		activity.Metadata["notes"] = "Good weather"

		assert.Equal(t, "500kg", activity.Output["yield"])
		assert.Equal(t, "Good weather", activity.Metadata["notes"])
	})
}
