package crop_cycle

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCropCycleValidate(t *testing.T) {
	tests := []struct {
		name      string
		cropCycle *CropCycle
		wantErr   bool
	}{
		{
			name: "valid crop cycle",
			cropCycle: &CropCycle{
				FarmID:   "farm123",
				FarmerID: "farmer123",
				Season:   "RABI",
				Status:   "PLANNED",
				CropID:   "crop123",
			},
			wantErr: false,
		},
		{
			name: "missing farm ID",
			cropCycle: &CropCycle{
				FarmerID: "farmer123",
				Season:   "RABI",
				Status:   "PLANNED",
			},
			wantErr: true,
		},
		{
			name: "missing farmer ID",
			cropCycle: &CropCycle{
				FarmID: "farm123",
				Season: "RABI",
				Status: "PLANNED",
			},
			wantErr: true,
		},
		{
			name: "missing season",
			cropCycle: &CropCycle{
				FarmID:   "farm123",
				FarmerID: "farmer123",
				Status:   "PLANNED",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cropCycle.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCropCycleTableName(t *testing.T) {
	cropCycle := &CropCycle{}
	assert.Equal(t, "crop_cycles", cropCycle.TableName())
}

func TestCropCycleGetTableIdentifier(t *testing.T) {
	cropCycle := &CropCycle{}
	assert.Equal(t, "crop_cycle", cropCycle.GetTableIdentifier())
}

func TestCropCycleFields(t *testing.T) {
	now := time.Now()
	varietyID := "variety123"
	cropCycle := &CropCycle{
		FarmID:    "farm123",
		FarmerID:  "farmer123",
		Season:    "RABI",
		Status:    "ACTIVE",
		StartDate: &now,
		CropID:    "crop123",
		VarietyID: &varietyID,
		Outcome:   map[string]string{"yield": "good"},
	}

	assert.Equal(t, "farm123", cropCycle.FarmID)
	assert.Equal(t, "farmer123", cropCycle.FarmerID)
	assert.Equal(t, "RABI", cropCycle.Season)
	assert.Equal(t, "ACTIVE", cropCycle.Status)
	assert.Equal(t, &now, cropCycle.StartDate)
	assert.Equal(t, "crop123", cropCycle.CropID)
	assert.Equal(t, &varietyID, cropCycle.VarietyID)
	assert.Equal(t, map[string]string{"yield": "good"}, cropCycle.Outcome)
}
