package farmer

import (
	"context"
	"testing"
	"time"

	farmerentity "github.com/Kisanlink/farmers-module/internal/entities/farmer"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Create simplified tables compatible with SQLite
	// Farmers table
	err = db.Exec(`
		CREATE TABLE farmers (
			id VARCHAR(255) PRIMARY KEY,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			created_by VARCHAR(255),
			updated_by VARCHAR(255),
			deleted_at DATETIME,
			deleted_by VARCHAR(255),
			aaa_user_id VARCHAR(255) NOT NULL,
			aaa_org_id VARCHAR(255),
			kisan_sathi_user_id VARCHAR(255),
			first_name VARCHAR(255) NOT NULL,
			last_name VARCHAR(255) NOT NULL,
			phone_number VARCHAR(50),
			email VARCHAR(255),
			date_of_birth DATE,
			gender VARCHAR(50),
			address_id VARCHAR(255),
			land_ownership_type VARCHAR(100),
			social_category VARCHAR(50),
			area_type VARCHAR(50),
			status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
			total_acreage_ha REAL NOT NULL DEFAULT 0,
			farm_count INTEGER NOT NULL DEFAULT 0,
			preferences TEXT DEFAULT '{}',
			metadata TEXT DEFAULT '{}'
		);
	`).Error
	require.NoError(t, err)

	// Farmer links table
	err = db.Exec(`
		CREATE TABLE farmer_links (
			id VARCHAR(255) PRIMARY KEY,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			created_by VARCHAR(255),
			updated_by VARCHAR(255),
			deleted_at DATETIME,
			deleted_by VARCHAR(255),
			farmer_id VARCHAR(255),
			aaa_user_id VARCHAR(255) NOT NULL,
			aaa_org_id VARCHAR(255) NOT NULL,
			kisan_sathi_user_id VARCHAR(255),
			status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE'
		);
	`).Error
	require.NoError(t, err)

	// Create unique index on farmer_links
	err = db.Exec(`
		CREATE UNIQUE INDEX idx_farmer_link_user_org ON farmer_links (aaa_user_id, aaa_org_id);
	`).Error
	require.NoError(t, err)

	return db
}

// seedTestData creates test farmers and farmer links
func seedTestData(t *testing.T, db *gorm.DB) (map[string]*farmerentity.Farmer, map[string]*farmerentity.FarmerLink) {
	farmers := make(map[string]*farmerentity.Farmer)
	farmerLinks := make(map[string]*farmerentity.FarmerLink)

	// Create test farmers
	farmer1 := &farmerentity.Farmer{
		BaseModel: base.BaseModel{
			Model: base.Model{
				ID:        "FMRR-001",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
		AAAUserID:   "user-001",
		FirstName:   "John",
		LastName:    "Doe",
		PhoneNumber: "1234567890",
		Email:       "john@example.com",
		Status:      "ACTIVE",
	}
	require.NoError(t, db.Create(farmer1).Error)
	farmers["farmer1"] = farmer1

	farmer2 := &farmerentity.Farmer{
		BaseModel: base.BaseModel{
			Model: base.Model{
				ID:        "FMRR-002",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
		AAAUserID:   "user-002",
		FirstName:   "Jane",
		LastName:    "Smith",
		PhoneNumber: "9876543210",
		Email:       "jane@example.com",
		Status:      "ACTIVE",
	}
	require.NoError(t, db.Create(farmer2).Error)
	farmers["farmer2"] = farmer2

	farmer3 := &farmerentity.Farmer{
		BaseModel: base.BaseModel{
			Model: base.Model{
				ID:        "FMRR-003",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
		AAAUserID:   "user-003",
		FirstName:   "Bob",
		LastName:    "Johnson",
		PhoneNumber: "5555555555",
		Email:       "bob@example.com",
		Status:      "ACTIVE",
	}
	require.NoError(t, db.Create(farmer3).Error)
	farmers["farmer3"] = farmer3

	// Create farmer links (farmer1 and farmer2 linked to org-001, farmer3 to org-002)
	link1 := &farmerentity.FarmerLink{
		BaseModel: base.BaseModel{
			Model: base.Model{
				ID:        "FMLK-001",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
		AAAUserID: "user-001",
		AAAOrgID:  "org-001",
		Status:    "ACTIVE",
	}
	require.NoError(t, db.Create(link1).Error)
	farmerLinks["link1"] = link1

	link2 := &farmerentity.FarmerLink{
		BaseModel: base.BaseModel{
			Model: base.Model{
				ID:        "FMLK-002",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
		AAAUserID: "user-002",
		AAAOrgID:  "org-001",
		Status:    "ACTIVE",
	}
	require.NoError(t, db.Create(link2).Error)
	farmerLinks["link2"] = link2

	link3 := &farmerentity.FarmerLink{
		BaseModel: base.BaseModel{
			Model: base.Model{
				ID:        "FMLK-003",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
		AAAUserID: "user-003",
		AAAOrgID:  "org-002",
		Status:    "ACTIVE",
	}
	require.NoError(t, db.Create(link3).Error)
	farmerLinks["link3"] = link3

	// Create additional link for farmer1 to org-002 (multi-org scenario)
	link4 := &farmerentity.FarmerLink{
		BaseModel: base.BaseModel{
			Model: base.Model{
				ID:        "FMLK-004",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
		AAAUserID: "user-001",
		AAAOrgID:  "org-002",
		Status:    "ACTIVE",
	}
	require.NoError(t, db.Create(link4).Error)
	farmerLinks["link4"] = link4

	return farmers, farmerLinks
}

func TestFarmerRepository_FindByOrgID(t *testing.T) {
	db := setupTestDB(t)
	_, _ = seedTestData(t, db)

	// Create repository
	repo := &FarmerRepository{
		db: db,
	}

	tests := []struct {
		name            string
		aaaOrgID        string
		filter          *base.Filter
		expectedCount   int
		expectedUserIDs []string
		expectError     bool
	}{
		{
			name:            "Find farmers for org-001",
			aaaOrgID:        "org-001",
			filter:          base.NewFilterBuilder().Build(),
			expectedCount:   2,
			expectedUserIDs: []string{"user-001", "user-002"},
			expectError:     false,
		},
		{
			name:            "Find farmers for org-002",
			aaaOrgID:        "org-002",
			filter:          base.NewFilterBuilder().Build(),
			expectedCount:   2,
			expectedUserIDs: []string{"user-001", "user-003"},
			expectError:     false,
		},
		{
			name:            "Find farmers for non-existent org",
			aaaOrgID:        "org-999",
			filter:          base.NewFilterBuilder().Build(),
			expectedCount:   0,
			expectedUserIDs: []string{},
			expectError:     false,
		},
		{
			name:     "Find farmers with phone number filter",
			aaaOrgID: "org-001",
			filter: base.NewFilterBuilder().
				Where("phone_number", base.OpEqual, "1234567890").
				Build(),
			expectedCount:   1,
			expectedUserIDs: []string{"user-001"},
			expectError:     false,
		},
		{
			name:     "Find farmers with pagination",
			aaaOrgID: "org-001",
			filter: base.NewFilterBuilder().
				Page(1, 1).
				Build(),
			expectedCount:   1,
			expectedUserIDs: []string{}, // Don't check specific user IDs with pagination
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			farmers, err := repo.FindByOrgID(ctx, tt.aaaOrgID, tt.filter)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedCount, len(farmers))

			if len(tt.expectedUserIDs) > 0 {
				// Verify that the returned farmers have the expected user IDs
				userIDs := make([]string, len(farmers))
				for i, farmer := range farmers {
					userIDs[i] = farmer.AAAUserID
				}
				assert.ElementsMatch(t, tt.expectedUserIDs, userIDs)
			}
		})
	}
}

func TestFarmerRepository_CountByOrgID(t *testing.T) {
	db := setupTestDB(t)
	_, _ = seedTestData(t, db)

	// Create repository
	repo := &FarmerRepository{
		db: db,
	}

	tests := []struct {
		name          string
		aaaOrgID      string
		filter        *base.Filter
		expectedCount int64
		expectError   bool
	}{
		{
			name:          "Count farmers for org-001",
			aaaOrgID:      "org-001",
			filter:        base.NewFilterBuilder().Build(),
			expectedCount: 2,
			expectError:   false,
		},
		{
			name:          "Count farmers for org-002",
			aaaOrgID:      "org-002",
			filter:        base.NewFilterBuilder().Build(),
			expectedCount: 2,
			expectError:   false,
		},
		{
			name:          "Count farmers for non-existent org",
			aaaOrgID:      "org-999",
			filter:        base.NewFilterBuilder().Build(),
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:     "Count farmers with phone number filter",
			aaaOrgID: "org-001",
			filter: base.NewFilterBuilder().
				Where("phone_number", base.OpEqual, "1234567890").
				Build(),
			expectedCount: 1,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			count, err := repo.CountByOrgID(ctx, tt.aaaOrgID, tt.filter)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedCount, count)
		})
	}
}

func TestFarmerRepository_FindByOrgID_WithSorting(t *testing.T) {
	db := setupTestDB(t)
	_, _ = seedTestData(t, db)

	// Create repository
	repo := &FarmerRepository{
		db: db,
	}

	ctx := context.Background()

	// Test ascending sort by first name
	filter := base.NewFilterBuilder().
		Sort("first_name", "asc").
		Build()

	farmers, err := repo.FindByOrgID(ctx, "org-001", filter)
	require.NoError(t, err)
	require.Equal(t, 2, len(farmers))
	assert.Equal(t, "Jane", farmers[0].FirstName) // Jane comes before John
	assert.Equal(t, "John", farmers[1].FirstName)

	// Test descending sort by first name
	filter = base.NewFilterBuilder().
		Sort("first_name", "desc").
		Build()

	farmers, err = repo.FindByOrgID(ctx, "org-001", filter)
	require.NoError(t, err)
	require.Equal(t, 2, len(farmers))
	assert.Equal(t, "John", farmers[0].FirstName) // John comes before Jane in desc
	assert.Equal(t, "Jane", farmers[1].FirstName)
}

func TestFarmerRepository_FindByOrgID_MultiOrgScenario(t *testing.T) {
	db := setupTestDB(t)
	_, _ = seedTestData(t, db)

	// Create repository
	repo := &FarmerRepository{
		db: db,
	}

	ctx := context.Background()

	// Verify farmer1 appears in both org-001 and org-002
	filter := base.NewFilterBuilder().
		Where("aaa_user_id", base.OpEqual, "user-001").
		Build()

	// Check org-001
	farmersOrg1, err := repo.FindByOrgID(ctx, "org-001", filter)
	require.NoError(t, err)
	assert.Equal(t, 1, len(farmersOrg1))
	assert.Equal(t, "user-001", farmersOrg1[0].AAAUserID)

	// Check org-002
	farmersOrg2, err := repo.FindByOrgID(ctx, "org-002", filter)
	require.NoError(t, err)
	assert.Equal(t, 1, len(farmersOrg2))
	assert.Equal(t, "user-001", farmersOrg2[0].AAAUserID)

	// Verify it's the same farmer in both queries
	assert.Equal(t, farmersOrg1[0].ID, farmersOrg2[0].ID)
}
