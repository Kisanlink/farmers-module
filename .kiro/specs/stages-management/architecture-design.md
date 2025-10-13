# Stages Management Feature - Architecture Design

## Executive Summary

This document outlines the complete architecture design for implementing a stages management feature in the farmers-module backend service. The feature will manage growth stages for crops, following the established patterns in the codebase.

## Table of Contents

1. [Overview](#overview)
2. [Database Design](#database-design)
3. [Domain Entities](#domain-entities)
4. [Repository Layer](#repository-layer)
5. [Service Layer](#service-layer)
6. [Handler Layer](#handler-layer)
7. [Routes](#routes)
8. [Error Handling](#error-handling)
9. [Security Considerations](#security-considerations)
10. [Migration Strategy](#migration-strategy)

## Overview

The stages management feature will provide:
- Master `stages` table for reusable growth stages
- `crop_stages` join table linking stages to crops with order and duration
- Full CRUD operations with pagination
- AAA service integration for authorization
- Standardized ID generation with STGE prefix

## Database Design

### 1. Stages Table

```sql
CREATE TABLE IF NOT EXISTS stages (
    id VARCHAR(20) PRIMARY KEY,
    stage_name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    properties JSONB NOT NULL DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Indexes
CREATE INDEX idx_stages_stage_name ON stages(stage_name);
CREATE INDEX idx_stages_is_active ON stages(is_active);
CREATE INDEX idx_stages_deleted_at ON stages(deleted_at);
```

### 2. Crop Stages Table

```sql
CREATE TABLE IF NOT EXISTS crop_stages (
    id VARCHAR(20) PRIMARY KEY,
    crop_id VARCHAR(20) NOT NULL,
    stage_id VARCHAR(20) NOT NULL,
    stage_order INTEGER NOT NULL,
    duration_days INTEGER,
    duration_unit VARCHAR(20) NOT NULL DEFAULT 'DAYS',
    properties JSONB NOT NULL DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,

    -- Foreign Keys
    CONSTRAINT fk_crop_stages_crop
        FOREIGN KEY (crop_id)
        REFERENCES crops(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_crop_stages_stage
        FOREIGN KEY (stage_id)
        REFERENCES stages(id)
        ON DELETE RESTRICT,

    -- Unique constraint for crop-stage combination
    CONSTRAINT uk_crop_stage UNIQUE(crop_id, stage_id),

    -- Unique constraint for order within a crop
    CONSTRAINT uk_crop_stage_order UNIQUE(crop_id, stage_order)
);

-- Indexes
CREATE INDEX idx_crop_stages_crop_id ON crop_stages(crop_id);
CREATE INDEX idx_crop_stages_stage_id ON crop_stages(stage_id);
CREATE INDEX idx_crop_stages_order ON crop_stages(stage_order);
CREATE INDEX idx_crop_stages_deleted_at ON crop_stages(deleted_at);
```

## Domain Entities

### Stage Entity

```go
// internal/entities/stage/stage.go
package stage

import (
    "github.com/Kisanlink/farmers-module/internal/entities"
    "github.com/Kisanlink/kisanlink-db/pkg/base"
    "github.com/Kisanlink/kisanlink-db/pkg/core/hash"
    "github.com/Kisanlink/farmers-module/pkg/common"
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
    return nil
}
```

### CropStage Entity

```go
// internal/entities/stage/crop_stage.go
package stage

import (
    "github.com/Kisanlink/farmers-module/internal/entities"
    cropEntity "github.com/Kisanlink/farmers-module/internal/entities/crop"
    "github.com/Kisanlink/kisanlink-db/pkg/base"
    "github.com/Kisanlink/kisanlink-db/pkg/core/hash"
    "github.com/Kisanlink/farmers-module/pkg/common"
)

// DurationUnit represents the unit of duration
type DurationUnit string

const (
    DurationUnitDays  DurationUnit = "DAYS"
    DurationUnitWeeks DurationUnit = "WEEKS"
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
```

## Repository Layer

### Stage Repository Interface

```go
// internal/repo/stage/stage_repository.go
package stage

import (
    "context"
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
        Where("is_active = ?", true).
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
        Where("is_active = ?", true)

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
        Order("stage_name ASC").
        Find(&stages).Error

    if err != nil {
        return nil, err
    }

    return stages, nil
}

// StageLookup represents simplified stage data for lookups
type StageLookup struct {
    ID          string  `json:"id"`
    StageName   string  `json:"stage_name"`
    Description *string `json:"description"`
}
```

### CropStage Repository

```go
// internal/repo/stage/crop_stage_repository.go
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
        Preload("Stage").
        Where("crop_id = ?", cropID).
        Where("is_active = ?", true).
        Order("stage_order ASC").
        Find(&cropStages).Error

    if err != nil {
        return nil, err
    }

    return cropStages, nil
}

// CheckCropStageExists checks if a crop-stage combination exists
func (r *CropStageRepository) CheckCropStageExists(ctx context.Context, cropID, stageID string, excludeID ...string) (bool, error) {
    query := r.db.WithContext(ctx).Model(&stage.CropStage{}).
        Where("crop_id = ?", cropID).
        Where("stage_id = ?", stageID).
        Where("is_active = ?", true)

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
        Where("is_active = ?", true)

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
        Where("is_active = ?", true).
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
                Update("stage_order", order).Error

            if err != nil {
                return fmt.Errorf("failed to update stage order for stage %s: %w", stageID, err)
            }
        }
        return nil
    })
}
```

## Service Layer

### Stage Service Interface

```go
// internal/services/interfaces.go additions
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

### Stage Service Implementation

```go
// internal/services/stage_service.go
package services

import (
    "context"
    "fmt"

    stageEntity "github.com/Kisanlink/farmers-module/internal/entities/stage"
    "github.com/Kisanlink/farmers-module/internal/entities/requests"
    "github.com/Kisanlink/farmers-module/internal/entities/responses"
    "github.com/Kisanlink/farmers-module/internal/repo/stage"
    "github.com/Kisanlink/farmers-module/pkg/common"
)

// StageServiceImpl implements StageService
type StageServiceImpl struct {
    stageRepo     *stage.StageRepository
    cropStageRepo *stage.CropStageRepository
    aaaService    AAAService
}

// NewStageService creates a new stage service
func NewStageService(
    stageRepo *stage.StageRepository,
    cropStageRepo *stage.CropStageRepository,
    aaaService AAAService,
) StageService {
    return &StageServiceImpl{
        stageRepo:     stageRepo,
        cropStageRepo: cropStageRepo,
        aaaService:    aaaService,
    }
}

// CreateStage implements StageService.CreateStage
func (s *StageServiceImpl) CreateStage(ctx context.Context, req interface{}) (interface{}, error) {
    createReq, ok := req.(*requests.CreateStageRequest)
    if !ok {
        return nil, common.ErrInvalidInput
    }

    // Check permission
    hasPermission, err := s.aaaService.CheckPermission(ctx, createReq.UserID, "stage", "create", "", createReq.OrgID)
    if err != nil {
        return nil, fmt.Errorf("failed to check permission: %w", err)
    }
    if !hasPermission {
        return nil, common.ErrForbidden
    }

    // Check if stage with same name already exists
    existingStage, err := s.stageRepo.FindByName(ctx, createReq.StageName)
    if err == nil && existingStage != nil {
        return nil, common.ErrAlreadyExists
    }

    // Create stage entity
    stageEnt := stageEntity.NewStage()
    stageEnt.StageName = createReq.StageName
    stageEnt.Description = createReq.Description
    if createReq.Properties != nil {
        stageEnt.Properties = createReq.Properties
    }

    // Validate the stage entity
    if err := stageEnt.Validate(); err != nil {
        return nil, err
    }

    // Save to database
    if err := s.stageRepo.Create(ctx, stageEnt); err != nil {
        return nil, fmt.Errorf("failed to create stage: %w", err)
    }

    // Convert to response
    stageData := &responses.StageData{
        ID:          stageEnt.ID,
        StageName:   stageEnt.StageName,
        Description: stageEnt.Description,
        Properties:  stageEnt.Properties,
        IsActive:    stageEnt.IsActive,
        CreatedAt:   stageEnt.CreatedAt,
        UpdatedAt:   stageEnt.UpdatedAt,
    }

    return &responses.StageResponse{
        Success:   true,
        Message:   "Stage created successfully",
        RequestID: createReq.RequestID,
        Data:      stageData,
    }, nil
}

// Additional service methods would follow the same pattern...
```

## Handler Layer

### Stage Handlers

```go
// internal/handlers/stage_handlers.go
package handlers

import (
    "net/http"

    "github.com/Kisanlink/farmers-module/internal/entities/requests"
    "github.com/Kisanlink/farmers-module/internal/entities/responses"
    "github.com/Kisanlink/farmers-module/internal/services"
    "github.com/gin-gonic/gin"
)

// StageHandler handles stage-related HTTP requests
type StageHandler struct {
    stageService services.StageService
}

// NewStageHandler creates a new stage handler
func NewStageHandler(stageService services.StageService) *StageHandler {
    return &StageHandler{
        stageService: stageService,
    }
}

// CreateStage godoc
// @Summary Create a new stage
// @Description Create a new growth stage
// @Tags Stages
// @Accept json
// @Produce json
// @Param stage body requests.CreateStageRequest true "Stage details"
// @Success 201 {object} responses.StageResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 409 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/stages [post]
func (h *StageHandler) CreateStage(c *gin.Context) {
    var req requests.CreateStageRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, responses.ErrorResponse{
            Success: false,
            Message: "Invalid request body",
            Error:   err.Error(),
        })
        return
    }

    // Add user context from middleware
    req.UserID = c.GetString("user_id")
    req.OrgID = c.GetString("org_id")
    req.RequestID = c.GetString("request_id")

    resp, err := h.stageService.CreateStage(c.Request.Context(), &req)
    if err != nil {
        handleServiceError(c, err)
        return
    }

    c.JSON(http.StatusCreated, resp)
}

// GetStage godoc
// @Summary Get a stage by ID
// @Description Get a growth stage by its ID
// @Tags Stages
// @Accept json
// @Produce json
// @Param id path string true "Stage ID"
// @Success 200 {object} responses.StageResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /api/v1/stages/{id} [get]
func (h *StageHandler) GetStage(c *gin.Context) {
    req := &requests.GetStageRequest{
        ID:        c.Param("id"),
        UserID:    c.GetString("user_id"),
        OrgID:     c.GetString("org_id"),
        RequestID: c.GetString("request_id"),
    }

    resp, err := h.stageService.GetStage(c.Request.Context(), req)
    if err != nil {
        handleServiceError(c, err)
        return
    }

    c.JSON(http.StatusOK, resp)
}

// Additional handler methods would follow the same pattern...
```

## Routes

### Stage Routes

```go
// internal/routes/stage_routes.go
package routes

import (
    "github.com/Kisanlink/farmers-module/internal/handlers"
    "github.com/Kisanlink/farmers-module/internal/middleware"
    "github.com/gin-gonic/gin"
)

// RegisterStageRoutes registers all stage-related routes
func RegisterStageRoutes(router *gin.RouterGroup, handler *handlers.StageHandler, authMiddleware gin.HandlerFunc) {
    stageGroup := router.Group("/stages")
    stageGroup.Use(authMiddleware)
    {
        // Stage CRUD operations
        stageGroup.POST("", handler.CreateStage)
        stageGroup.GET("", handler.ListStages)
        stageGroup.GET("/:id", handler.GetStage)
        stageGroup.PUT("/:id", handler.UpdateStage)
        stageGroup.DELETE("/:id", handler.DeleteStage)

        // Lookup endpoint
        stageGroup.GET("/lookup", handler.GetStageLookup)
    }

    // Crop-Stage relationship endpoints
    cropStageGroup := router.Group("/crops/:crop_id/stages")
    cropStageGroup.Use(authMiddleware)
    {
        cropStageGroup.GET("", handler.GetCropStages)
        cropStageGroup.POST("", handler.AssignStageToCrop)
        cropStageGroup.PUT("/:stage_id", handler.UpdateCropStage)
        cropStageGroup.DELETE("/:stage_id", handler.RemoveStageFromCrop)
        cropStageGroup.POST("/reorder", handler.ReorderCropStages)
    }
}
```

## Request/Response Models

### Request Models

```go
// internal/entities/requests/stage.go
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
    Search   string `json:"search,omitempty"`
    IsActive *bool  `json:"is_active,omitempty"`
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
}

// ReorderCropStagesRequest represents the request to reorder crop stages
type ReorderCropStagesRequest struct {
    BaseRequest
    CropID      string         `json:"-"`
    StageOrders map[string]int `json:"stage_orders" binding:"required"` // map[stage_id]order
}
```

### Response Models

```go
// internal/entities/responses/stage_responses.go
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

## Error Handling

The stages feature will follow the established error handling patterns:

1. **Input Validation Errors** (400): Invalid request data
2. **Authentication Errors** (401): Missing or invalid auth token
3. **Authorization Errors** (403): Insufficient permissions
4. **Not Found Errors** (404): Stage or crop not found
5. **Conflict Errors** (409): Duplicate stage name or order
6. **Internal Server Errors** (500): Database or system errors

## Security Considerations

1. **AAA Integration**: All operations require proper authentication and authorization through the AAA service
2. **Input Validation**: Strict validation of all inputs using struct tags and custom validators
3. **SQL Injection Prevention**: Use of parameterized queries through GORM
4. **XSS Prevention**: Proper escaping of user inputs in responses
5. **Rate Limiting**: Should be applied at the API gateway level
6. **Audit Logging**: All operations should be logged for audit purposes

## Migration Strategy

### Database Migration

```go
// internal/db/migrations/20250112_add_stages_tables.go
package migrations

import (
    "github.com/Kisanlink/farmers-module/internal/entities/stage"
    "gorm.io/gorm"
)

// AddStagesTables creates stages and crop_stages tables
func AddStagesTables(db *gorm.DB) error {
    // Auto-migrate the models
    err := db.AutoMigrate(
        &stage.Stage{},
        &stage.CropStage{},
    )
    if err != nil {
        return err
    }

    // Add additional indexes if needed
    if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_stages_stage_name_lower ON stages(LOWER(stage_name))").Error; err != nil {
        return err
    }

    return nil
}
```

### Seed Data

```go
// internal/db/seeds/stage_seeds.go
package seeds

import (
    "context"
    "github.com/Kisanlink/farmers-module/internal/entities/stage"
    "gorm.io/gorm"
)

// SeedStages creates initial stage data
func SeedStages(ctx context.Context, db *gorm.DB) error {
    stages := []stage.Stage{
        {StageName: "Land Preparation", Description: stringPtr("Preparing the land for planting")},
        {StageName: "Sowing/Planting", Description: stringPtr("Planting seeds or seedlings")},
        {StageName: "Germination", Description: stringPtr("Seeds sprouting and initial growth")},
        {StageName: "Vegetative Growth", Description: stringPtr("Plant develops leaves and stems")},
        {StageName: "Flowering", Description: stringPtr("Plant produces flowers")},
        {StageName: "Fruit Development", Description: stringPtr("Fruits begin to form and grow")},
        {StageName: "Maturity", Description: stringPtr("Crop reaches harvest readiness")},
        {StageName: "Harvesting", Description: stringPtr("Collecting the mature crop")},
        {StageName: "Post-Harvest", Description: stringPtr("Processing and storage activities")},
    }

    for _, stg := range stages {
        // Check if exists
        var existing stage.Stage
        err := db.Where("stage_name = ?", stg.StageName).First(&existing).Error
        if err == gorm.ErrRecordNotFound {
            // Create new stage
            newStage := stage.NewStage()
            newStage.StageName = stg.StageName
            newStage.Description = stg.Description

            if err := db.Create(newStage).Error; err != nil {
                return err
            }
        }
    }

    return nil
}

func stringPtr(s string) *string {
    return &s
}
```

## Integration Points

### 1. Service Factory Update

```go
// internal/services/service_factory.go additions
type ServiceFactory struct {
    // ... existing fields
    stageService StageService
}

func (f *ServiceFactory) GetStageService() StageService {
    if f.stageService == nil {
        stageRepo := stage.NewStageRepository(f.db)
        cropStageRepo := stage.NewCropStageRepository(f.db)
        f.stageService = NewStageService(stageRepo, cropStageRepo, f.GetAAAService())
    }
    return f.stageService
}
```

### 2. Repository Factory Update

```go
// internal/repo/repository_factory.go additions
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

### 3. Main Route Registration

```go
// internal/routes/routes.go additions
func SetupRoutes(router *gin.Engine, serviceFactory *services.ServiceFactory) {
    // ... existing routes

    // Stage routes
    stageHandler := handlers.NewStageHandler(serviceFactory.GetStageService())
    RegisterStageRoutes(v1, stageHandler, authMiddleware)
}
```

## Testing Strategy

### Unit Tests

1. **Entity Tests**: Validate model validation logic
2. **Repository Tests**: Test database operations with test containers
3. **Service Tests**: Mock repository and AAA service for business logic testing
4. **Handler Tests**: Mock service layer and test HTTP responses

### Integration Tests

1. **End-to-End Tests**: Full flow from HTTP request to database
2. **AAA Integration**: Test permission checks with AAA service
3. **Database Constraints**: Test unique constraints and foreign key relationships

### Example Test

```go
// internal/services/stage_service_test.go
package services_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestCreateStage_Success(t *testing.T) {
    // Setup mocks
    mockStageRepo := new(MockStageRepository)
    mockAAAService := new(MockAAAService)

    service := NewStageService(mockStageRepo, nil, mockAAAService)

    // Setup expectations
    mockAAAService.On("CheckPermission", mock.Anything, "user123", "stage", "create", "", "org123").Return(true, nil)
    mockStageRepo.On("FindByName", mock.Anything, "Test Stage").Return(nil, gorm.ErrRecordNotFound)
    mockStageRepo.On("Create", mock.Anything, mock.AnythingOfType("*stage.Stage")).Return(nil)

    // Execute
    req := &requests.CreateStageRequest{
        BaseRequest: requests.BaseRequest{
            UserID: "user123",
            OrgID:  "org123",
        },
        StageName:   "Test Stage",
        Description: stringPtr("Test Description"),
    }

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
```

## Performance Considerations

1. **Database Indexes**: Proper indexing on frequently queried fields
2. **Pagination**: Mandatory pagination for list endpoints
3. **Query Optimization**: Use of preloading for relationships where needed
4. **Caching Strategy**: Consider Redis caching for frequently accessed stages
5. **Batch Operations**: Support bulk operations for crop-stage assignments

## Documentation Requirements

1. **API Documentation**: Swagger annotations for all endpoints
2. **Database Schema**: ERD diagram for stages relationships
3. **Sequence Diagrams**: Flow diagrams for complex operations
4. **ADR**: Architecture Decision Records for design choices
5. **README Updates**: Include stages feature in module documentation

## Rollback Strategy

In case of deployment issues:

1. **Database Rollback**: Migration down scripts to remove tables
2. **Code Rollback**: Git revert to previous version
3. **Feature Flag**: Consider feature flag for gradual rollout
4. **Data Backup**: Backup existing data before migration

## Success Metrics

1. **API Response Time**: < 200ms for single entity operations
2. **Error Rate**: < 0.1% for valid requests
3. **Test Coverage**: > 80% code coverage
4. **AAA Integration**: 100% of operations properly authorized

## Next Steps

1. **Review & Approval**: Architecture review by team lead
2. **Implementation**: Backend engineer to implement based on this design
3. **Testing**: Comprehensive testing including unit and integration tests
4. **Documentation**: Update API documentation and user guides
5. **Deployment**: Deploy to staging environment first
6. **Monitoring**: Set up monitoring and alerting for the new feature
