package fpo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFPORefValidate(t *testing.T) {
	tests := []struct {
		name    string
		fpoRef  *FPORef
		wantErr bool
	}{
		{
			name: "valid FPO ref",
			fpoRef: &FPORef{
				AAAOrgID:       "org123",
				Name:           "Test FPO",
				RegistrationNo: "REG123",
				Status:         "ACTIVE",
			},
			wantErr: false,
		},
		{
			name: "missing AAA org ID",
			fpoRef: &FPORef{
				Name:           "Test FPO",
				RegistrationNo: "REG123",
				Status:         "ACTIVE",
			},
			wantErr: true,
		},
		{
			name: "missing name",
			fpoRef: &FPORef{
				AAAOrgID:       "org123",
				RegistrationNo: "REG123",
				Status:         "ACTIVE",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fpoRef.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFPORefTableName(t *testing.T) {
	fpoRef := &FPORef{}
	assert.Equal(t, "fpo_refs", fpoRef.TableName())
}

func TestFPORefGetTableIdentifier(t *testing.T) {
	fpoRef := &FPORef{}
	assert.Equal(t, "fpo_ref", fpoRef.GetTableIdentifier())
}

func TestFPORefFields(t *testing.T) {
	fpoRef := &FPORef{
		AAAOrgID:       "org123",
		Name:           "Test FPO",
		RegistrationNo: "REG123",
		Status:         FPOStatusActive,
		BusinessConfig: map[string]string{"type": "agricultural"},
	}

	assert.Equal(t, "org123", fpoRef.AAAOrgID)
	assert.Equal(t, "Test FPO", fpoRef.Name)
	assert.Equal(t, "REG123", fpoRef.RegistrationNo)
	assert.Equal(t, FPOStatusActive, fpoRef.Status)
	assert.Equal(t, map[string]string{"type": "agricultural"}, fpoRef.BusinessConfig)
}
