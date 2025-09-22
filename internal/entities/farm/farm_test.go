package farm

import (
	"testing"

	"github.com/Kisanlink/farmers-module/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestFarmValidate(t *testing.T) {
	tests := []struct {
		name    string
		farm    *Farm
		wantErr bool
	}{
		{
			name: "valid farm",
			farm: &Farm{
				AAAFarmerUserID: "user123",
				AAAOrgID:        "org123",
				Name:            testutils.StringPtr("Test Farm"),
				Geometry:        "POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))",
			},
			wantErr: false,
		},
		{
			name: "missing AAA farmer user ID",
			farm: &Farm{
				AAAOrgID: "org123",
				Name:     testutils.StringPtr("Test Farm"),
				Geometry: "POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))",
			},
			wantErr: true,
		},
		{
			name: "missing AAA org ID",
			farm: &Farm{
				AAAFarmerUserID: "user123",
				Name:            testutils.StringPtr("Test Farm"),
				Geometry:        "POLYGON((0 0, 1 0, 1 1, 0 1, 0 0))",
			},
			wantErr: true,
		},
		{
			name: "missing geometry",
			farm: &Farm{
				AAAFarmerUserID: "user123",
				AAAOrgID:        "org123",
				Name:            testutils.StringPtr("Test Farm"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.farm.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFarmTableName(t *testing.T) {
	farm := &Farm{}
	assert.Equal(t, "farms", farm.TableName())
}

func TestFarmGetTableIdentifier(t *testing.T) {
	farm := &Farm{}
	assert.Equal(t, "farm", farm.GetTableIdentifier())
}
