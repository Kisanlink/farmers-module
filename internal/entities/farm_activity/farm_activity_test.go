package farm_activity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFarmActivityValidate(t *testing.T) {
	tests := []struct {
		name         string
		farmActivity *FarmActivity
		wantErr      bool
	}{
		{
			name: "valid farm activity",
			farmActivity: &FarmActivity{
				CropCycleID:  "cycle123",
				ActivityType: "PLANTING",
				CreatedBy:    "user123",
				Status:       "PLANNED",
			},
			wantErr: false,
		},
		{
			name: "missing crop cycle ID",
			farmActivity: &FarmActivity{
				ActivityType: "PLANTING",
				CreatedBy:    "user123",
				Status:       "PLANNED",
			},
			wantErr: true,
		},
		{
			name: "missing activity type",
			farmActivity: &FarmActivity{
				CropCycleID: "cycle123",
				CreatedBy:   "user123",
				Status:      "PLANNED",
			},
			wantErr: true,
		},
		{
			name: "missing created by",
			farmActivity: &FarmActivity{
				CropCycleID:  "cycle123",
				ActivityType: "PLANTING",
				Status:       "PLANNED",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.farmActivity.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFarmActivityTableName(t *testing.T) {
	farmActivity := &FarmActivity{}
	assert.Equal(t, "farm_activities", farmActivity.TableName())
}

func TestFarmActivityGetTableIdentifier(t *testing.T) {
	farmActivity := &FarmActivity{}
	assert.Equal(t, "FACT", farmActivity.GetTableIdentifier())
}

func TestFarmActivityFields(t *testing.T) {
	now := time.Now()
	farmActivity := &FarmActivity{
		CropCycleID:  "cycle123",
		ActivityType: "PLANTING",
		PlannedAt:    &now,
		CompletedAt:  &now,
		CreatedBy:    "user123",
		Status:       "COMPLETED",
		Output:       map[string]string{"seeds_used": "10kg"},
		Metadata:     map[string]string{"weather": "sunny"},
	}

	assert.Equal(t, "cycle123", farmActivity.CropCycleID)
	assert.Equal(t, "PLANTING", farmActivity.ActivityType)
	assert.Equal(t, &now, farmActivity.PlannedAt)
	assert.Equal(t, &now, farmActivity.CompletedAt)
	assert.Equal(t, "user123", farmActivity.CreatedBy)
	assert.Equal(t, "COMPLETED", farmActivity.Status)
	assert.Equal(t, map[string]string{"seeds_used": "10kg"}, farmActivity.Output)
	assert.Equal(t, map[string]string{"weather": "sunny"}, farmActivity.Metadata)
}
