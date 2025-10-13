package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPermissionForRoute_StageRoutes(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		path         string
		wantResource string
		wantAction   string
		wantExists   bool
	}{
		{
			name:         "Create Stage",
			method:       "POST",
			path:         "/api/v1/stages",
			wantResource: "stage",
			wantAction:   "create",
			wantExists:   true,
		},
		{
			name:         "List Stages",
			method:       "GET",
			path:         "/api/v1/stages",
			wantResource: "stage",
			wantAction:   "list",
			wantExists:   true,
		},
		{
			name:         "List Stages with query params",
			method:       "GET",
			path:         "/api/v1/stages?page=1&page_size=20",
			wantResource: "stage",
			wantAction:   "list",
			wantExists:   true,
		},
		{
			name:         "Get Stage Lookup",
			method:       "GET",
			path:         "/api/v1/stages/lookup",
			wantResource: "stage",
			wantAction:   "list",
			wantExists:   true,
		},
		{
			name:         "Get Stage by ID",
			method:       "GET",
			path:         "/api/v1/stages/STGE123",
			wantResource: "stage",
			wantAction:   "read",
			wantExists:   true,
		},
		{
			name:         "Update Stage",
			method:       "PUT",
			path:         "/api/v1/stages/STGE123",
			wantResource: "stage",
			wantAction:   "update",
			wantExists:   true,
		},
		{
			name:         "Delete Stage",
			method:       "DELETE",
			path:         "/api/v1/stages/STGE123",
			wantResource: "stage",
			wantAction:   "delete",
			wantExists:   true,
		},
		{
			name:         "Assign Stage to Crop",
			method:       "POST",
			path:         "/api/v1/crops/CROP123/stages",
			wantResource: "crop_stage",
			wantAction:   "create",
			wantExists:   true,
		},
		{
			name:         "Get Crop Stages",
			method:       "GET",
			path:         "/api/v1/crops/CROP123/stages",
			wantResource: "crop_stage",
			wantAction:   "read",
			wantExists:   true,
		},
		{
			name:         "Reorder Crop Stages",
			method:       "POST",
			path:         "/api/v1/crops/CROP123/stages/reorder",
			wantResource: "crop_stage",
			wantAction:   "update",
			wantExists:   true,
		},
		{
			name:         "Update Crop Stage",
			method:       "PUT",
			path:         "/api/v1/crops/CROP123/stages/STGE456",
			wantResource: "crop_stage",
			wantAction:   "update",
			wantExists:   true,
		},
		{
			name:         "Remove Stage from Crop",
			method:       "DELETE",
			path:         "/api/v1/crops/CROP123/stages/STGE456",
			wantResource: "crop_stage",
			wantAction:   "delete",
			wantExists:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			permission, exists := GetPermissionForRoute(tt.method, tt.path)

			assert.Equal(t, tt.wantExists, exists, "Expected exists=%v for route %s %s", tt.wantExists, tt.method, tt.path)

			if tt.wantExists {
				assert.Equal(t, tt.wantResource, permission.Resource, "Expected resource=%s for route %s %s", tt.wantResource, tt.method, tt.path)
				assert.Equal(t, tt.wantAction, permission.Action, "Expected action=%s for route %s %s", tt.wantAction, tt.method, tt.path)
			}
		})
	}
}

func TestGetPermissionForRoute_StageNormalization(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		path     string
		expected string
	}{
		{
			name:     "Stage with query params",
			method:   "GET",
			path:     "/api/v1/stages?page=1&page_size=20&search=sowing",
			expected: "/api/v1/stages",
		},
		{
			name:     "Stage by ID",
			method:   "GET",
			path:     "/api/v1/stages/STGE-ABC123XYZ",
			expected: "/api/v1/stages/:id",
		},
		{
			name:     "Crop stages",
			method:   "GET",
			path:     "/api/v1/crops/CROP-123/stages",
			expected: "/api/v1/crops/:id/stages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalized := normalizePath(tt.path)
			assert.Equal(t, tt.expected, normalized, "Normalization failed for path %s", tt.path)
		})
	}
}
