package farmer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFarmer(t *testing.T) {
	farmer := NewFarmer()

	assert.NotNil(t, farmer)
	assert.NotEmpty(t, farmer.ID)
	assert.NotNil(t, farmer.Preferences)
	assert.NotNil(t, farmer.Metadata)
	assert.Equal(t, "FMRR", farmer.GetTableIdentifier())
}

func TestFarmerValidate(t *testing.T) {
	tests := []struct {
		name    string
		farmer  *Farmer
		wantErr bool
	}{
		{
			name: "valid farmer",
			farmer: &Farmer{
				AAAUserID: "user123",
				AAAOrgID:  "org123",
				FirstName: "John",
				LastName:  "Doe",
			},
			wantErr: false,
		},
		{
			name: "missing AAA user ID",
			farmer: &Farmer{
				AAAOrgID:  "org123",
				FirstName: "John",
				LastName:  "Doe",
			},
			wantErr: true,
		},
		{
			name: "missing AAA org ID - now optional",
			farmer: &Farmer{
				AAAUserID: "user123",
				FirstName: "John",
				LastName:  "Doe",
			},
			wantErr: false, // AAA org ID is now optional - farmer is identified by user ID only
		},
		{
			name: "missing first name",
			farmer: &Farmer{
				AAAUserID: "user123",
				AAAOrgID:  "org123",
				LastName:  "Doe",
			},
			wantErr: true,
		},
		{
			name: "missing last name",
			farmer: &Farmer{
				AAAUserID: "user123",
				AAAOrgID:  "org123",
				FirstName: "John",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.farmer.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFarmerLinkValidate(t *testing.T) {
	tests := []struct {
		name       string
		farmerLink *FarmerLink
		wantErr    bool
	}{
		{
			name: "valid farmer link",
			farmerLink: &FarmerLink{
				AAAUserID: "user123",
				AAAOrgID:  "org123",
				Status:    "ACTIVE",
			},
			wantErr: false,
		},
		{
			name: "missing AAA user ID",
			farmerLink: &FarmerLink{
				AAAOrgID: "org123",
				Status:   "ACTIVE",
			},
			wantErr: true,
		},
		{
			name: "missing AAA org ID",
			farmerLink: &FarmerLink{
				AAAUserID: "user123",
				Status:    "ACTIVE",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.farmerLink.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFarmerTableName(t *testing.T) {
	farmer := &Farmer{}
	assert.Equal(t, "farmers", farmer.TableName())
}

func TestFarmerLinkTableName(t *testing.T) {
	farmerLink := &FarmerLink{}
	assert.Equal(t, "farmer_links", farmerLink.TableName())
}

func TestFarmerGetTableIdentifier(t *testing.T) {
	farmer := &Farmer{}
	assert.Equal(t, "FMRR", farmer.GetTableIdentifier())
}

func TestFarmerLinkGetTableIdentifier(t *testing.T) {
	farmerLink := &FarmerLink{}
	assert.Equal(t, "FMLK", farmerLink.GetTableIdentifier())
}
