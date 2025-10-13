# Stages Management - Implementation Guide

## Implementation Order

Follow this order to ensure dependencies are satisfied:

1. **Database Entities** (internal/entities/stage/)
2. **Database Migrations** (internal/db/)
3. **Repository Layer** (internal/repo/stage/)
4. **Request/Response Models** (internal/entities/requests/, internal/entities/responses/)
5. **Service Layer** (internal/services/)
6. **Handler Layer** (internal/handlers/)
7. **Routes** (internal/routes/)
8. **Integration & Testing**

## Step 1: Create Database Entities

### 1.1 Create Stage Entity

```bash
mkdir -p internal/entities/stage
```

Create `internal/entities/stage/stage.go`:

```go
package stage

import (
    "github.com/Kisanlink/farmers-module/internal/entities"
    "github.com/Kisanlink/farmers-module/pkg/common"
    "github.com/Kisanlink/kisanlink-db/pkg/base"
    "github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// Stage represents a master growth stage for crops
type Stage struct {
    base.BaseModel
    StageName   string         `json:"stage_name" gorm:"type:varchar(100);not null;uniqueIndex"`
    Description *string        `json:"description" gorm:"type:text"`
    Properties  entities.JSONB `json:"properties" gorm:"type:jsonb;not null;default:'{}'"`
    IsActive    bool          `json:"is_active" gorm:"type:boolean;not null;default:true"`
}

// TableName returns the table name for the Stage model
func (s *Stage) TableName() string {
    return "stages"
}

// GetTableIdentifier returns the table identifier for ID generation
func (s *Stage) GetTableIdentifier() string {
    return "STGE"
}

// GetTableSize returns the table size for ID generation
func (s *Stage) GetTableSize() hash.TableSize {
    return hash.Medium
}

// NewStage creates a new stage model with proper initialization
func NewStage() *Stage {
    baseModel := base.NewBaseModel("STGE", hash.Medium)
    return &Stage{
        BaseModel:  *baseModel,
        Properties: make(entities.JSONB),
        IsActive:   true,
    }
}

// Validate validates the stage model
func (s *Stage) Validate() error {
    if s.StageName == "" {
        return common.ErrInvalidInput
    }
    if len(s.StageName) > 100 {
        return common.ErrInvalidInput
    }
    return nil
}
```

### 1.2 Create CropStage Entity

Create `internal/entities/stage/crop_stage.go`:

```go
package stage

import (
    "github.com/Kisanlink/farmers-module/internal/entities"
    cropEntity "github.com/Kisanlink/farmers-module/internal/entities/crop"
    "github.com/Kisanlink/farmers-module/pkg/common"
    "github.com/Kisanlink/kisanlink-db/pkg/base"
    "github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

// DurationUnit represents the unit of duration
type DurationUnit string

const (
    DurationUnitDays   DurationUnit = "DAYS"
    DurationUnitWeeks  DurationUnit = "WEEKS"
    DurationUnitMonths DurationUnit = "MONTHS"
)

// CropStage represents the relationship between crop and stage
type CropStage struct {
    base.BaseModel
    CropID       string         `json:"crop_id" gorm:"type:varchar(20);not null;index"`
    StageID      string         `json:"stage_id" gorm:"type:varchar(20);not null;index"`
    StageOrder   int            `json:"stage_order" gorm:"type:integer;not null;index"`
    DurationDays *int           `json:"duration_days" gorm:"type:integer"`
    DurationUnit DurationUnit   `json:"duration_unit" gorm:"type:varchar(20);not null;default:'DAYS'"`
    Properties   entities.JSONB `json:"properties" gorm:"type:jsonb;not null;default:'{}'"`
    IsActive     bool          `json:"is_active" gorm:"type:boolean;not null;default:true"`

    // Relationships
    Crop  *cropEntity.Crop `json:"crop,omitempty" gorm:"foreignKey:CropID"`
    Stage *Stage           `json:"stage,omitempty" gorm:"foreignKey:StageID"`
}

// TableName returns the table name for the CropStage model
func (cs *CropStage) TableName() string {
    return "crop_stages"
}

// GetTableIdentifier returns the table identifier for ID generation
func (cs *CropStage) GetTableIdentifier() string {
    return "CSTG"
}

// GetTableSize returns the table size for ID generation
func (cs *CropStage) GetTableSize() hash.TableSize {
    return hash.Medium
}

// NewCropStage creates a new crop stage model with proper initialization
func NewCropStage() *CropStage {
    baseModel := base.NewBaseModel("CSTG", hash.Medium)
    return &CropStage{
        BaseModel:    *baseModel,
        Properties:   make(entities.JSONB),
        DurationUnit: DurationUnitDays,
        IsActive:     true,
    }
}

// Validate validates the crop stage model
func (cs *CropStage) Validate() error {
    if cs.CropID == "" || cs.StageID == "" {
        return common.ErrInvalidInput
    }
    if cs.StageOrder < 1 {
        return common.ErrInvalidInput
    }
    if cs.DurationDays != nil && *cs.DurationDays <= 0 {
        return common.ErrInvalidInput
    }

    // Validate duration unit
    validUnits := map[DurationUnit]bool{
        DurationUnitDays:   true,
        DurationUnitWeeks:  true,
        DurationUnitMonths: true,
    }
    if !validUnits[cs.DurationUnit] {
        return common.ErrInvalidInput
    }

    return nil
}

// GetValidDurationUnits returns all valid duration units
func GetValidDurationUnits() []DurationUnit {
    return []DurationUnit{
        DurationUnitDays,
        DurationUnitWeeks,
        DurationUnitMonths,
    }
}
```

## Step 2: Create Repository Layer

### 2.1 Create Stage Repository

```bash
mkdir -p internal/repo/stage
```

Create `internal/repo/stage/stage_repository.go`:

```go
package stage

import (
    "context"
    "fmt"
    "strings"

    "github.com/Kisanlink/farmers-module/internal/entities/stage"
    "github.com/Kisanlink/kisanlink-db/pkg/base"
    "gorm.io/gorm"
)

// StageRepository provides data access methods for stages
type StageRepository struct {
    *base.BaseFilterableRepository[*stage.Stage]
    db *gorm.DB
}

// NewStageRepository creates a new stage repository
func NewStageRepository(db *gorm.DB) *StageRepository {
    baseRepo := base.NewBaseFilterableRepository[*stage.Stage]()
    return &StageRepository{
        BaseFilterableRepository: baseRepo,
        db:                      db,
    }
}

// FindByName finds a stage by its name (case-insensitive)
func (r *StageRepository) FindByName(ctx context.Context, name string) (*stage.Stage, error) {
    var stg stage.Stage
    err := r.db.WithContext(ctx).
        Where("LOWER(stage_name) = LOWER(?)", name).
        Where("deleted_at IS NULL").
        First(&stg).Error
    if err != nil {
        return nil, err
    }
    return &stg, nil
}

// Search finds stages by name or description
func (r *StageRepository) Search(ctx context.Context, searchTerm string, page, pageSize int) ([]*stage.Stage, int, error) {
    var stages []*stage.Stage
    var total int64

    searchPattern := "%" + strings.ToLower(searchTerm) + "%"
    query := r.db.WithContext(ctx).Model(&stage.Stage{}).
        Where("(LOWER(stage_name) LIKE ? OR LOWER(description) LIKE ?)", searchPattern, searchPattern).
        Where("deleted_at IS NULL")

    // Get total count
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    // Get paginated results
    offset := (page - 1) * pageSize
    if err := query.Offset(offset).Limit(pageSize).Order("stage_name ASC").Find(&stages).Error; err != nil {
        return nil, 0, err
    }

    return stages, int(total), nil
}

// GetActiveStagesForLookup gets simplified stage data for dropdown/lookup
func (r *StageRepository) GetActiveStagesForLookup(ctx context.Context) ([]*StageLookup, error) {
    var stages []*StageLookup

    err := r.db.WithContext(ctx).
        Model(&stage.Stage{}).
        Select("id, stage_name, description").
        Where("is_active = ?", true).
        Where("deleted_at IS NULL").
        Order("stage_name ASC").
        Find(&stages).Error

    if err != nil {
        return nil, err
    }

    return stages, nil
}

// ListWithFilters lists stages with filters
func (r *StageRepository) ListWithFilters(ctx context.Context, filters StageFilters) ([]*stage.Stage, int, error) {
    var stages []*stage.Stage
    var total int64

    query := r.db.WithContext(ctx).Model(&stage.Stage{})

    // Apply filters
    if filters.Search != "" {
        searchPattern := "%" + strings.ToLower(filters.Search) + "%"
        query = query.Where("(LOWER(stage_name) LIKE ? OR LOWER(description) LIKE ?)", searchPattern, searchPattern)
    }

    if filters.IsActive != nil {
        query = query.Where("is_active = ?", *filters.IsActive)
    } else {
        query = query.Where("is_active = ?", true)
    }

    query = query.Where("deleted_at IS NULL")

    // Get total count
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    // Get paginated results
    offset := (filters.Page - 1) * filters.PageSize
    if err := query.Offset(offset).Limit(filters.PageSize).Order("stage_name ASC").Find(&stages).Error; err != nil {
        return nil, 0, err
    }

    return stages, int(total), nil
}

// StageFilters represents filters for stage queries
type StageFilters struct {
    Search   string
    IsActive *bool
    Page     int
    PageSize int
}

// StageLookup represents simplified stage data for lookups
type StageLookup struct {
    ID          string  `json:"id"`
    StageName   string  `json:"stage_name"`
    Description *string `json:"description"`
}
```

### 2.2 Create CropStage Repository

Create `internal/repo/stage/crop_stage_repository.go`:

```go
package stage

import (
    "context"
    "fmt"

    "github.com/Kisanlink/farmers-module/internal/entities/stage"
    "github.com/Kisanlink/kisanlink-db/pkg/base"
    "gorm.io/gorm"
)

// CropStageRepository provides data access methods for crop stages
type CropStageRepository struct {
    *base.BaseFilterableRepository[*stage.CropStage]
    db *gorm.DB
}

// NewCropStageRepository creates a new crop stage repository
func NewCropStageRepository(db *gorm.DB) *CropStageRepository {
    baseRepo := base.NewBaseFilterableRepository[*stage.CropStage]()
    return &CropStageRepository{
        BaseFilterableRepository: baseRepo,
        db:                      db,
    }
}

// GetCropStages gets all stages for a crop in order
func (r *CropStageRepository) GetCropStages(ctx context.Context, cropID string) ([]*stage.CropStage, error) {
    var cropStages []*stage.CropStage

    err := r.db.WithContext(ctx).
        Preload("Stage", "deleted_at IS NULL").
        Where("crop_id = ?", cropID).
        Where("is_active = ?", true).
        Where("deleted_at IS NULL").
        Order("stage_order ASC").
        Find(&cropStages).Error

    if err != nil {
        return nil, err
    }

    return cropStages, nil
}

// GetCropStageByID gets a specific crop stage
func (r *CropStageRepository) GetCropStageByID(ctx context.Context, id string) (*stage.CropStage, error) {
    var cropStage stage.CropStage

    err := r.db.WithContext(ctx).
        Preload("Stage").
        Preload("Crop").
        Where("id = ?", id).
        Where("deleted_at IS NULL").
        First(&cropStage).Error

    if err != nil {
        return nil, err
    }

    return &cropStage, nil
}

// GetCropStageByCropAndStage gets a crop stage by crop and stage IDs
func (r *CropStageRepository) GetCropStageByCropAndStage(ctx context.Context, cropID, stageID string) (*stage.CropStage, error) {
    var cropStage stage.CropStage

    err := r.db.WithContext(ctx).
        Where("crop_id = ?", cropID).
        Where("stage_id = ?", stageID).
        Where("deleted_at IS NULL").
        First(&cropStage).Error

    if err != nil {
        return nil, err
    }

    return &cropStage, nil
}

// CheckCropStageExists checks if a crop-stage combination exists
func (r *CropStageRepository) CheckCropStageExists(ctx context.Context, cropID, stageID string, excludeID ...string) (bool, error) {
    query := r.db.WithContext(ctx).Model(&stage.CropStage{}).
        Where("crop_id = ?", cropID).
        Where("stage_id = ?", stageID).
        Where("deleted_at IS NULL")

    if len(excludeID) > 0 && excludeID[0] != "" {
        query = query.Where("id != ?", excludeID[0])
    }

    var count int64
    err := query.Count(&count).Error
    if err != nil {
        return false, err
    }

    return count > 0, nil
}

// CheckStageOrderExists checks if a stage order already exists for a crop
func (r *CropStageRepository) CheckStageOrderExists(ctx context.Context, cropID string, stageOrder int, excludeID ...string) (bool, error) {
    query := r.db.WithContext(ctx).Model(&stage.CropStage{}).
        Where("crop_id = ?", cropID).
        Where("stage_order = ?", stageOrder).
        Where("deleted_at IS NULL")

    if len(excludeID) > 0 && excludeID[0] != "" {
        query = query.Where("id != ?", excludeID[0])
    }

    var count int64
    err := query.Count(&count).Error
    if err != nil {
        return false, err
    }

    return count > 0, nil
}

// GetMaxStageOrder gets the maximum stage order for a crop
func (r *CropStageRepository) GetMaxStageOrder(ctx context.Context, cropID string) (int, error) {
    var maxOrder int
    err := r.db.WithContext(ctx).
        Model(&stage.CropStage{}).
        Select("COALESCE(MAX(stage_order), 0)").
        Where("crop_id = ?", cropID).
        Where("deleted_at IS NULL").
        Scan(&maxOrder).Error

    if err != nil {
        return 0, err
    }

    return maxOrder, nil
}

// ReorderStages updates stage orders for a crop
func (r *CropStageRepository) ReorderStages(ctx context.Context, cropID string, stageOrders map[string]int) error {
    return r.db.Transaction(func(tx *gorm.DB) error {
        for stageID, order := range stageOrders {
            err := tx.Model(&stage.CropStage{}).
                Where("crop_id = ?", cropID).
                Where("stage_id = ?", stageID).
                Where("deleted_at IS NULL").
                Update("stage_order", order).Error

            if err != nil {
                return fmt.Errorf("failed to update stage order for stage %s: %w", stageID, err)
            }
        }
        return nil
    })
}
```

## Step 3: Create Request/Response Models

### 3.1 Create Stage Request Models

Create `internal/entities/requests/stage.go`:

```go
package requests

import "github.com/Kisanlink/farmers-module/internal/entities"

// CreateStageRequest represents the request to create a stage
type CreateStageRequest struct {
    BaseRequest
    StageName   string         `json:"stage_name" binding:"required,min=1,max=100"`
    Description *string        `json:"description,omitempty"`
    Properties  entities.JSONB `json:"properties,omitempty"`
}

// UpdateStageRequest represents the request to update a stage
type UpdateStageRequest struct {
    BaseRequest
    ID          string         `json:"-"`
    StageName   *string        `json:"stage_name,omitempty" binding:"omitempty,min=1,max=100"`
    Description *string        `json:"description,omitempty"`
    Properties  entities.JSONB `json:"properties,omitempty"`
    IsActive    *bool          `json:"is_active,omitempty"`
}

// GetStageRequest represents the request to get a stage
type GetStageRequest struct {
    BaseRequest
    ID string `json:"-"`
}

// DeleteStageRequest represents the request to delete a stage
type DeleteStageRequest struct {
    BaseRequest
    ID string `json:"-"`
}

// ListStagesRequest represents the request to list stages
type ListStagesRequest struct {
    BaseRequest
    PaginationRequest
    Search   string `json:"search,omitempty" form:"search"`
    IsActive *bool  `json:"is_active,omitempty" form:"is_active"`
}

// GetStageLookupRequest represents the request to get stage lookup data
type GetStageLookupRequest struct {
    BaseRequest
}

// AssignStageToCropRequest represents the request to assign a stage to a crop
type AssignStageToCropRequest struct {
    BaseRequest
    CropID       string         `json:"-"`
    StageID      string         `json:"stage_id" binding:"required"`
    StageOrder   int            `json:"stage_order" binding:"required,min=1"`
    DurationDays *int           `json:"duration_days,omitempty" binding:"omitempty,min=1"`
    DurationUnit string         `json:"duration_unit,omitempty" binding:"omitempty,oneof=DAYS WEEKS MONTHS"`
    Properties   entities.JSONB `json:"properties,omitempty"`
}

// UpdateCropStageRequest represents the request to update a crop stage
type UpdateCropStageRequest struct {
    BaseRequest
    CropID       string         `json:"-"`
    StageID      string         `json:"-"`
    StageOrder   *int           `json:"stage_order,omitempty" binding:"omitempty,min=1"`
    DurationDays *int           `json:"duration_days,omitempty" binding:"omitempty,min=1"`
    DurationUnit *string        `json:"duration_unit,omitempty" binding:"omitempty,oneof=DAYS WEEKS MONTHS"`
    Properties   entities.JSONB `json:"properties,omitempty"`
    IsActive     *bool          `json:"is_active,omitempty"`
}

// RemoveStageFromCropRequest represents the request to remove a stage from a crop
type RemoveStageFromCropRequest struct {
    BaseRequest
    CropID  string `json:"-"`
    StageID string `json:"-"`
}

// GetCropStagesRequest represents the request to get crop stages
type GetCropStagesRequest struct {
    BaseRequest
    CropID string `json:"-"`
}

// ReorderCropStagesRequest represents the request to reorder crop stages
type ReorderCropStagesRequest struct {
    BaseRequest
    CropID      string         `json:"-"`
    StageOrders map[string]int `json:"stage_orders" binding:"required"` // map[stage_id]order
}
```

### 3.2 Create Stage Response Models

Create `internal/entities/responses/stage_responses.go`:

```go
package responses

import (
    "time"
    "github.com/Kisanlink/farmers-module/internal/entities"
)

// StageData represents stage data in responses
type StageData struct {
    ID          string         `json:"id"`
    StageName   string         `json:"stage_name"`
    Description *string        `json:"description,omitempty"`
    Properties  entities.JSONB `json:"properties,omitempty"`
    IsActive    bool          `json:"is_active"`
    CreatedAt   time.Time     `json:"created_at"`
    UpdatedAt   time.Time     `json:"updated_at"`
}

// StageResponse represents a single stage response
type StageResponse struct {
    BaseResponse
    Data *StageData `json:"data,omitempty"`
}

// StageListResponse represents a list of stages response
type StageListResponse struct {
    BaseResponse
    Data     []*StageData `json:"data"`
    Page     int         `json:"page"`
    PageSize int         `json:"page_size"`
    Total    int         `json:"total"`
}

// CropStageData represents crop stage data in responses
type CropStageData struct {
    ID           string         `json:"id"`
    CropID       string         `json:"crop_id"`
    StageID      string         `json:"stage_id"`
    StageName    string         `json:"stage_name"`
    Description  *string        `json:"description,omitempty"`
    StageOrder   int           `json:"stage_order"`
    DurationDays *int          `json:"duration_days,omitempty"`
    DurationUnit string        `json:"duration_unit"`
    Properties   entities.JSONB `json:"properties,omitempty"`
    IsActive     bool          `json:"is_active"`
    CreatedAt    time.Time     `json:"created_at"`
    UpdatedAt    time.Time     `json:"updated_at"`
}

// CropStageResponse represents a single crop stage response
type CropStageResponse struct {
    BaseResponse
    Data *CropStageData `json:"data,omitempty"`
}

// CropStagesResponse represents a list of crop stages response
type CropStagesResponse struct {
    BaseResponse
    Data []*CropStageData `json:"data"`
}

// StageLookupData represents simplified stage data for lookups
type StageLookupData struct {
    ID          string  `json:"id"`
    StageName   string  `json:"stage_name"`
    Description *string `json:"description,omitempty"`
}

// StageLookupResponse represents stage lookup response
type StageLookupResponse struct {
    BaseResponse
    Data []*StageLookupData `json:"data"`
}
```

## Step 4: Update Service Interfaces

Add to `internal/services/interfaces.go`:

```go
// StageService handles stage-related operations
type StageService interface {
    // Stage CRUD
    CreateStage(ctx context.Context, req interface{}) (interface{}, error)
    GetStage(ctx context.Context, req interface{}) (interface{}, error)
    UpdateStage(ctx context.Context, req interface{}) (interface{}, error)
    DeleteStage(ctx context.Context, req interface{}) (interface{}, error)
    ListStages(ctx context.Context, req interface{}) (interface{}, error)

    // CropStage operations
    AssignStageToCrop(ctx context.Context, req interface{}) (interface{}, error)
    RemoveStageFromCrop(ctx context.Context, req interface{}) (interface{}, error)
    UpdateCropStage(ctx context.Context, req interface{}) (interface{}, error)
    GetCropStages(ctx context.Context, req interface{}) (interface{}, error)
    ReorderCropStages(ctx context.Context, req interface{}) (interface{}, error)

    // Lookup operations
    GetStageLookup(ctx context.Context, req interface{}) (interface{}, error)
}
```

## Step 5: Database Migration

Update `internal/db/migrate.go` to include stage tables:

```go
// Add to the AutoMigrate function
func AutoMigrate(db *gorm.DB) error {
    // ... existing migrations

    // Stage tables
    if err := db.AutoMigrate(
        &stage.Stage{},
        &stage.CropStage{},
    ); err != nil {
        return fmt.Errorf("failed to migrate stage tables: %w", err)
    }

    // Create indexes
    if err := createStageIndexes(db); err != nil {
        return fmt.Errorf("failed to create stage indexes: %w", err)
    }

    return nil
}

func createStageIndexes(db *gorm.DB) error {
    // Create composite unique index for crop_stages
    if err := db.Exec(`
        CREATE UNIQUE INDEX IF NOT EXISTS uk_crop_stage
        ON crop_stages(crop_id, stage_id)
        WHERE deleted_at IS NULL
    `).Error; err != nil {
        return err
    }

    // Create unique index for stage order within a crop
    if err := db.Exec(`
        CREATE UNIQUE INDEX IF NOT EXISTS uk_crop_stage_order
        ON crop_stages(crop_id, stage_order)
        WHERE deleted_at IS NULL
    `).Error; err != nil {
        return err
    }

    return nil
}
```

## Step 6: Seed Initial Data

Create `internal/db/seeds/stage_seeds.go`:

```go
package seeds

import (
    "context"
    "github.com/Kisanlink/farmers-module/internal/entities/stage"
    "gorm.io/gorm"
)

// SeedStages creates initial stage data
func SeedStages(ctx context.Context, db *gorm.DB) error {
    stages := []struct {
        name        string
        description string
    }{
        {"Land Preparation", "Preparing the land for planting"},
        {"Sowing/Planting", "Planting seeds or seedlings"},
        {"Germination", "Seeds sprouting and initial growth"},
        {"Vegetative Growth", "Plant develops leaves and stems"},
        {"Flowering", "Plant produces flowers"},
        {"Fruit Development", "Fruits begin to form and grow"},
        {"Maturity", "Crop reaches harvest readiness"},
        {"Harvesting", "Collecting the mature crop"},
        {"Post-Harvest", "Processing and storage activities"},
    }

    for _, stgData := range stages {
        // Check if exists
        var existing stage.Stage
        err := db.Where("stage_name = ?", stgData.name).First(&existing).Error
        if err == gorm.ErrRecordNotFound {
            // Create new stage
            newStage := stage.NewStage()
            newStage.StageName = stgData.name
            desc := stgData.description
            newStage.Description = &desc

            if err := db.Create(newStage).Error; err != nil {
                return err
            }
        }
    }

    return nil
}
```

## Step 7: Update Factory Pattern

### 7.1 Update Repository Factory

Add to `internal/repo/repository_factory.go`:

```go
import (
    // ... existing imports
    "github.com/Kisanlink/farmers-module/internal/repo/stage"
)

type RepositoryFactory struct {
    // ... existing fields
    stageRepo     *stage.StageRepository
    cropStageRepo *stage.CropStageRepository
}

func (f *RepositoryFactory) GetStageRepository() *stage.StageRepository {
    if f.stageRepo == nil {
        f.stageRepo = stage.NewStageRepository(f.db)
    }
    return f.stageRepo
}

func (f *RepositoryFactory) GetCropStageRepository() *stage.CropStageRepository {
    if f.cropStageRepo == nil {
        f.cropStageRepo = stage.NewCropStageRepository(f.db)
    }
    return f.cropStageRepo
}
```

### 7.2 Update Service Factory

Add to `internal/services/service_factory.go`:

```go
type ServiceFactory struct {
    // ... existing fields
    stageService StageService
}

func (f *ServiceFactory) GetStageService() StageService {
    if f.stageService == nil {
        repoFactory := repo.NewRepositoryFactory(f.db)
        f.stageService = NewStageService(
            repoFactory.GetStageRepository(),
            repoFactory.GetCropStageRepository(),
            f.GetAAAService(),
        )
    }
    return f.stageService
}
```

## Step 8: Testing

### 8.1 Create Unit Test for Stage Entity

Create `internal/entities/stage/stage_test.go`:

```go
package stage_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/Kisanlink/farmers-module/internal/entities/stage"
)

func TestNewStage(t *testing.T) {
    stg := stage.NewStage()

    assert.NotNil(t, stg)
    assert.NotEmpty(t, stg.ID)
    assert.True(t, strings.HasPrefix(stg.ID, "STGE"))
    assert.True(t, stg.IsActive)
    assert.NotNil(t, stg.Properties)
}

func TestStageValidation(t *testing.T) {
    tests := []struct {
        name      string
        stage     *stage.Stage
        wantError bool
    }{
        {
            name: "valid stage",
            stage: &stage.Stage{
                StageName: "Test Stage",
            },
            wantError: false,
        },
        {
            name: "empty stage name",
            stage: &stage.Stage{
                StageName: "",
            },
            wantError: true,
        },
        {
            name: "stage name too long",
            stage: &stage.Stage{
                StageName: strings.Repeat("a", 101),
            },
            wantError: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.stage.Validate()
            if tt.wantError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### 8.2 Create Integration Test

Create `internal/services/stage_service_test.go`:

```go
package services_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/Kisanlink/farmers-module/internal/entities/requests"
    "github.com/Kisanlink/farmers-module/internal/entities/responses"
)

func TestStageService_CreateStage(t *testing.T) {
    // Setup
    mockStageRepo := new(MockStageRepository)
    mockAAAService := new(MockAAAService)

    service := NewStageService(mockStageRepo, nil, mockAAAService)

    // Test data
    req := &requests.CreateStageRequest{
        BaseRequest: requests.BaseRequest{
            UserID:    "USER123",
            OrgID:     "ORG123",
            RequestID: "REQ123",
        },
        StageName:   "Test Stage",
        Description: stringPtr("Test Description"),
    }

    // Setup expectations
    mockAAAService.On("CheckPermission",
        mock.Anything, "USER123", "stage", "create", "", "ORG123",
    ).Return(true, nil)

    mockStageRepo.On("FindByName",
        mock.Anything, "Test Stage",
    ).Return(nil, gorm.ErrRecordNotFound)

    mockStageRepo.On("Create",
        mock.Anything, mock.AnythingOfType("*stage.Stage"),
    ).Return(nil)

    // Execute
    resp, err := service.CreateStage(context.Background(), req)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, resp)

    stageResp := resp.(*responses.StageResponse)
    assert.True(t, stageResp.Success)
    assert.Equal(t, "Stage created successfully", stageResp.Message)
    assert.Equal(t, "Test Stage", stageResp.Data.StageName)

    mockAAAService.AssertExpectations(t)
    mockStageRepo.AssertExpectations(t)
}

func stringPtr(s string) *string {
    return &s
}
```

## Step 9: API Documentation

Add Swagger annotations to handlers. Example:

```go
// CreateStage godoc
// @Summary Create a new stage
// @Description Create a new growth stage for crops
// @Tags Stages
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param stage body requests.CreateStageRequest true "Stage details"
// @Success 201 {object} responses.StageResponse
// @Failure 400 {object} responses.ErrorResponse "Invalid input"
// @Failure 401 {object} responses.ErrorResponse "Unauthorized"
// @Failure 403 {object} responses.ErrorResponse "Forbidden"
// @Failure 409 {object} responses.ErrorResponse "Stage already exists"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /api/v1/stages [post]
func (h *StageHandler) CreateStage(c *gin.Context) {
    // Implementation...
}
```

## Step 10: Makefile Updates

Add commands to the Makefile:

```makefile
# Seed stages data
seed-stages:
	@echo "Seeding stages data..."
	@go run cmd/farmers-service/main.go seed stages

# Run stage tests
test-stages:
	@echo "Running stage tests..."
	@go test ./internal/entities/stage/... -v
	@go test ./internal/repo/stage/... -v
	@go test ./internal/services/... -run Stage -v

# Generate stage mocks
mock-stages:
	@echo "Generating stage mocks..."
	@mockery --dir=internal/repo/stage --name=StageRepository --output=internal/services/mocks
	@mockery --dir=internal/repo/stage --name=CropStageRepository --output=internal/services/mocks
```

## Implementation Checklist

- [ ] Create stage entity files
- [ ] Create crop_stage entity files
- [ ] Create stage repository
- [ ] Create crop_stage repository
- [ ] Create request models
- [ ] Create response models
- [ ] Update service interfaces
- [ ] Implement stage service
- [ ] Create stage handlers
- [ ] Setup routes
- [ ] Update repository factory
- [ ] Update service factory
- [ ] Add database migrations
- [ ] Create seed data
- [ ] Write unit tests
- [ ] Write integration tests
- [ ] Add Swagger documentation
- [ ] Update Makefile
- [ ] Run tests
- [ ] Generate docs

## Common Pitfalls to Avoid

1. **ID Generation**: Always use the kisanlink-db base model for consistent ID generation
2. **Soft Deletes**: Use `deleted_at IS NULL` in queries to respect soft deletes
3. **AAA Integration**: Never skip permission checks
4. **Validation**: Always validate at both handler and service levels
5. **Transaction Management**: Use transactions for multi-step operations
6. **Error Handling**: Return appropriate HTTP status codes
7. **Pagination**: Always paginate list endpoints
8. **Preloading**: Use GORM's Preload for relationships to avoid N+1 queries
9. **Indexes**: Create appropriate indexes for frequently queried fields
10. **Testing**: Write tests for all new functionality

## Verification Steps

1. Run migrations: `make migrate`
2. Seed data: `make seed-stages`
3. Run tests: `make test-stages`
4. Generate docs: `make docs`
5. Start server: `make run`
6. Test endpoints with Postman/curl

## Sample API Calls

### Create Stage
```bash
curl -X POST http://localhost:8000/api/v1/stages \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "stage_name": "Seedling",
    "description": "Young plant stage"
  }'
```

### List Stages
```bash
curl -X GET "http://localhost:8000/api/v1/stages?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN"
```

### Assign Stage to Crop
```bash
curl -X POST http://localhost:8000/api/v1/crops/CROP123/stages \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "stage_id": "STGE123",
    "stage_order": 1,
    "duration_days": 14,
    "duration_unit": "DAYS"
  }'
```
