# FPO Lifecycle Management - Implementation Plan

## Overview

This document provides a detailed implementation plan for the FPO lifecycle management system, broken down into meaningful, committable tasks that maintain the system in a working state throughout development.

## Implementation Phases

### Phase 1: Foundation (Week 1)
Focus: Core data model and repository layer

### Phase 2: State Machine (Week 1-2)
Focus: Lifecycle state management and transitions

### Phase 3: Service Layer (Week 2)
Focus: Business logic and workflow orchestration

### Phase 4: API Layer (Week 2-3)
Focus: REST endpoints and handlers

### Phase 5: Integration (Week 3)
Focus: AAA service integration and synchronization

### Phase 6: Testing & Documentation (Week 3-4)
Focus: Comprehensive testing and documentation

## Detailed Task Breakdown

### Phase 1: Foundation Tasks

#### Task 1.1: Database Schema Updates
**Description**: Add lifecycle management columns to fpo_refs table
**Files to modify**:
- `/internal/entities/fpo/fpo.go`
- Create migration file

**Implementation**:
```go
// Add to FPORef struct
PreviousStatus  FPOStatus      `json:"previous_status" gorm:"type:varchar(50)"`
StatusReason    string         `json:"status_reason" gorm:"type:text"`
StatusChangedAt time.Time      `json:"status_changed_at"`
StatusChangedBy string         `json:"status_changed_by"`
VerificationStatus string      `json:"verification_status" gorm:"type:varchar(50)"`
VerifiedAt      *time.Time     `json:"verified_at"`
VerifiedBy      string         `json:"verified_by"`
SetupAttempts   int           `json:"setup_attempts" gorm:"default:0"`
LastSetupAt     *time.Time    `json:"last_setup_at"`
SetupProgress   entities.JSONB `json:"setup_progress" gorm:"type:jsonb"`
CEOUserID       string        `json:"ceo_user_id" gorm:"type:varchar(255)"`
```

**Commit message**: "feat: extend FPORef entity with lifecycle management fields"

---

#### Task 1.2: Create Audit Log Entity
**Description**: Implement FPO audit log entity and table
**Files to create**:
- `/internal/entities/fpo/audit_log.go`

**Implementation**:
```go
package fpo

import (
    "time"
    "github.com/Kisanlink/farmers-module/internal/entities"
    "github.com/Kisanlink/kisanlink-db/pkg/base"
)

type FPOAuditLog struct {
    base.BaseModel
    FPOID          string         `json:"fpo_id" gorm:"type:varchar(255);not null;index"`
    Action         string         `json:"action" gorm:"type:varchar(100);not null"`
    PreviousState  FPOStatus      `json:"previous_state" gorm:"type:varchar(50)"`
    NewState       FPOStatus      `json:"new_state" gorm:"type:varchar(50)"`
    Reason         string         `json:"reason" gorm:"type:text"`
    PerformedBy    string         `json:"performed_by" gorm:"type:varchar(255);not null"`
    PerformedAt    time.Time      `json:"performed_at" gorm:"not null"`
    Details        entities.JSONB `json:"details" gorm:"type:jsonb"`
    RequestID      string         `json:"request_id" gorm:"type:varchar(255)"`
}

func (a *FPOAuditLog) TableName() string {
    return "fpo_audit_logs"
}

func (a *FPOAuditLog) GetTableIdentifier() string {
    return "FPOA"
}
```

**Commit message**: "feat: add FPO audit log entity for lifecycle tracking"

---

#### Task 1.3: Extend FPO Status Enum
**Description**: Add new lifecycle states to FPOStatus enum
**Files to modify**:
- `/internal/entities/fpo/fpo.go`

**Implementation**:
```go
const (
    // Existing statuses...

    // New lifecycle statuses
    FPOStatusDraft              FPOStatus = "DRAFT"
    FPOStatusPendingVerification FPOStatus = "PENDING_VERIFICATION"
    FPOStatusVerified           FPOStatus = "VERIFIED"
    FPOStatusRejected           FPOStatus = "REJECTED"
    FPOStatusSetupFailed        FPOStatus = "SETUP_FAILED"
    FPOStatusArchived           FPOStatus = "ARCHIVED"
)

// Add validation for new statuses
func (s FPOStatus) IsValid() bool {
    switch s {
    case FPOStatusDraft, FPOStatusPendingVerification, FPOStatusVerified,
         FPOStatusRejected, FPOStatusPendingSetup, FPOStatusSetupFailed,
         FPOStatusActive, FPOStatusInactive, FPOStatusSuspended, FPOStatusArchived:
        return true
    default:
        return false
    }
}

// Add state transition validation
func (s FPOStatus) CanTransitionTo(target FPOStatus) bool {
    transitions := map[FPOStatus][]FPOStatus{
        FPOStatusDraft:                {FPOStatusPendingVerification},
        FPOStatusPendingVerification:  {FPOStatusVerified, FPOStatusRejected},
        FPOStatusVerified:             {FPOStatusPendingSetup},
        FPOStatusRejected:             {FPOStatusDraft},
        FPOStatusPendingSetup:         {FPOStatusActive, FPOStatusSetupFailed},
        FPOStatusSetupFailed:          {FPOStatusPendingSetup},
        FPOStatusActive:               {FPOStatusSuspended, FPOStatusInactive},
        FPOStatusSuspended:            {FPOStatusActive},
        FPOStatusInactive:             {FPOStatusArchived},
    }

    allowed, exists := transitions[s]
    if !exists {
        return false
    }

    for _, status := range allowed {
        if status == target {
            return true
        }
    }
    return false
}
```

**Commit message**: "feat: extend FPO status enum with lifecycle states and transition rules"

---

#### Task 1.4: Create Enhanced FPO Repository
**Description**: Implement FPO repository with lifecycle methods
**Files to create**:
- `/internal/repo/fpo/fpo_repository.go`

**Implementation**:
```go
package fpo

import (
    "context"
    "fmt"
    "time"

    "github.com/Kisanlink/farmers-module/internal/entities/fpo"
    "github.com/Kisanlink/kisanlink-db/pkg/base"
    "gorm.io/gorm"
)

type FPORepository struct {
    *base.BaseFilterableRepository[*fpo.FPORef]
    auditRepo *base.BaseFilterableRepository[*fpo.FPOAuditLog]
    db        *gorm.DB
}

func NewFPORepository(dbManager interface{}) *FPORepository {
    repo := &FPORepository{
        BaseFilterableRepository: base.NewBaseFilterableRepository[*fpo.FPORef](),
        auditRepo: base.NewBaseFilterableRepository[*fpo.FPOAuditLog](),
    }
    repo.SetDBManager(dbManager)
    repo.auditRepo.SetDBManager(dbManager)

    // Extract DB instance for transactions
    if mgr, ok := dbManager.(*base.DBManager); ok {
        repo.db = mgr.DB
    }

    return repo
}

func (r *FPORepository) FindByAAAOrgID(ctx context.Context, aaaOrgID string) (*fpo.FPORef, error) {
    filter := base.NewFilterBuilder().
        Where("aaa_org_id", base.OpEqual, aaaOrgID).
        Build()
    return r.FindOne(ctx, filter)
}

func (r *FPORepository) UpdateStatus(ctx context.Context, fpoID string, newStatus fpo.FPOStatus, reason string, performedBy string) error {
    return r.db.Transaction(func(tx *gorm.DB) error {
        var fpoRef fpo.FPORef
        if err := tx.First(&fpoRef, "id = ?", fpoID).Error; err != nil {
            return fmt.Errorf("failed to find FPO: %w", err)
        }

        // Validate state transition
        if !fpoRef.Status.CanTransitionTo(newStatus) {
            return fmt.Errorf("invalid state transition from %s to %s", fpoRef.Status, newStatus)
        }

        // Create audit log
        auditLog := &fpo.FPOAuditLog{
            FPOID:         fpoID,
            Action:        "STATUS_CHANGE",
            PreviousState: fpoRef.Status,
            NewState:      newStatus,
            Reason:        reason,
            PerformedBy:   performedBy,
            PerformedAt:   time.Now(),
        }

        // Update FPO status
        updates := map[string]interface{}{
            "previous_status":    fpoRef.Status,
            "status":            newStatus,
            "status_reason":     reason,
            "status_changed_at": time.Now(),
            "status_changed_by": performedBy,
        }

        if err := tx.Model(&fpoRef).Updates(updates).Error; err != nil {
            return fmt.Errorf("failed to update FPO status: %w", err)
        }

        if err := tx.Create(&auditLog).Error; err != nil {
            return fmt.Errorf("failed to create audit log: %w", err)
        }

        return nil
    })
}

func (r *FPORepository) GetAuditHistory(ctx context.Context, fpoID string) ([]*fpo.FPOAuditLog, error) {
    filter := base.NewFilterBuilder().
        Where("fpo_id", base.OpEqual, fpoID).
        OrderBy("performed_at", false).
        Build()

    return r.auditRepo.FindAll(ctx, filter)
}
```

**Commit message**: "feat: implement enhanced FPO repository with lifecycle methods"

---

### Phase 2: State Machine Tasks

#### Task 2.1: Create State Machine Service
**Description**: Implement FPO state machine logic
**Files to create**:
- `/internal/services/fpo_state_machine.go`

**Implementation**:
```go
package services

import (
    "context"
    "fmt"

    "github.com/Kisanlink/farmers-module/internal/entities/fpo"
)

type FPOStateMachine struct {
    repo FPORepository
}

func NewFPOStateMachine(repo FPORepository) *FPOStateMachine {
    return &FPOStateMachine{repo: repo}
}

func (sm *FPOStateMachine) Transition(ctx context.Context, fpoID string, targetState fpo.FPOStatus, reason string, performedBy string) error {
    // Get current FPO
    fpoRef, err := sm.repo.FindByID(ctx, fpoID)
    if err != nil {
        return fmt.Errorf("failed to find FPO: %w", err)
    }

    // Validate transition
    if !fpoRef.Status.CanTransitionTo(targetState) {
        return fmt.Errorf("cannot transition from %s to %s", fpoRef.Status, targetState)
    }

    // Perform transition-specific actions
    switch targetState {
    case fpo.FPOStatusVerified:
        if err := sm.onVerified(ctx, fpoRef); err != nil {
            return err
        }
    case fpo.FPOStatusActive:
        if err := sm.onActivated(ctx, fpoRef); err != nil {
            return err
        }
    case fpo.FPOStatusSuspended:
        if err := sm.onSuspended(ctx, fpoRef); err != nil {
            return err
        }
    }

    // Update status with audit
    return sm.repo.UpdateStatus(ctx, fpoID, targetState, reason, performedBy)
}

func (sm *FPOStateMachine) onVerified(ctx context.Context, fpo *fpo.FPORef) error {
    // Set verification timestamp
    now := time.Now()
    fpo.VerifiedAt = &now
    return nil
}

func (sm *FPOStateMachine) onActivated(ctx context.Context, fpo *fpo.FPORef) error {
    // Clear setup errors
    fpo.SetupErrors = nil
    return nil
}

func (sm *FPOStateMachine) onSuspended(ctx context.Context, fpo *fpo.FPORef) error {
    // Log suspension reason
    return nil
}
```

**Commit message**: "feat: implement FPO state machine for lifecycle transitions"

---

### Phase 3: Service Layer Tasks

#### Task 3.1: Create FPO Lifecycle Service
**Description**: Implement main lifecycle service with business logic
**Files to create**:
- `/internal/services/fpo_lifecycle_service.go`

**Implementation**:
```go
package services

import (
    "context"
    "fmt"
    "log"

    "github.com/Kisanlink/farmers-module/internal/entities/fpo"
    "github.com/Kisanlink/farmers-module/internal/entities/requests"
)

type FPOLifecycleService struct {
    repo         FPORepository
    stateMachine *FPOStateMachine
    aaaService   AAAService
}

func NewFPOLifecycleService(repo FPORepository, aaaService AAAService) *FPOLifecycleService {
    return &FPOLifecycleService{
        repo:         repo,
        stateMachine: NewFPOStateMachine(repo),
        aaaService:   aaaService,
    }
}

// Create draft FPO
func (s *FPOLifecycleService) CreateDraftFPO(ctx context.Context, req *requests.CreateDraftFPORequest) (*fpo.FPORef, error) {
    fpoRef := &fpo.FPORef{
        Name:           req.Name,
        RegistrationNo: req.RegistrationNo,
        Status:         fpo.FPOStatusDraft,
        BusinessConfig: req.BusinessConfig,
        Metadata:       req.Metadata,
    }

    if err := s.repo.Create(ctx, fpoRef); err != nil {
        return nil, fmt.Errorf("failed to create draft FPO: %w", err)
    }

    log.Printf("Created draft FPO: %s", fpoRef.ID)
    return fpoRef, nil
}

// Submit for verification
func (s *FPOLifecycleService) SubmitForVerification(ctx context.Context, fpoID string) error {
    userID := GetUserIDFromContext(ctx)
    return s.stateMachine.Transition(ctx, fpoID, fpo.FPOStatusPendingVerification, "Submitted for verification", userID)
}

// Verify FPO
func (s *FPOLifecycleService) VerifyFPO(ctx context.Context, fpoID string, req *requests.VerifyFPORequest) error {
    // Update verification details
    fpoRef, err := s.repo.FindByID(ctx, fpoID)
    if err != nil {
        return err
    }

    now := time.Now()
    fpoRef.VerificationStatus = "VERIFIED"
    fpoRef.VerifiedAt = &now
    fpoRef.VerifiedBy = GetUserIDFromContext(ctx)
    fpoRef.VerificationNotes = req.Notes

    if err := s.repo.Update(ctx, fpoRef); err != nil {
        return err
    }

    // Transition state
    return s.stateMachine.Transition(ctx, fpoID, fpo.FPOStatusVerified, req.Notes, fpoRef.VerifiedBy)
}

// Initialize FPO setup in AAA
func (s *FPOLifecycleService) InitializeFPOSetup(ctx context.Context, fpoID string) error {
    fpoRef, err := s.repo.FindByID(ctx, fpoID)
    if err != nil {
        return err
    }

    if fpoRef.Status != fpo.FPOStatusVerified {
        return fmt.Errorf("FPO must be verified before setup")
    }

    // Transition to pending setup
    if err := s.stateMachine.Transition(ctx, fpoID, fpo.FPOStatusPendingSetup, "Starting AAA setup", "system"); err != nil {
        return err
    }

    // Trigger AAA setup asynchronously
    go s.performAAASetup(context.Background(), fpoID)

    return nil
}

// Sync FPO from AAA
func (s *FPOLifecycleService) SyncFPOFromAAA(ctx context.Context, aaaOrgID string) (*fpo.FPORef, error) {
    // Check if already exists
    existing, _ := s.repo.FindByAAAOrgID(ctx, aaaOrgID)
    if existing != nil {
        log.Printf("FPO already exists for org %s", aaaOrgID)
        return existing, nil
    }

    // Get from AAA
    org, err := s.aaaService.GetOrganization(ctx, aaaOrgID)
    if err != nil {
        return nil, fmt.Errorf("failed to get organization from AAA: %w", err)
    }

    orgMap, ok := org.(map[string]interface{})
    if !ok {
        return nil, fmt.Errorf("invalid organization response")
    }

    // Create local reference
    fpoRef := &fpo.FPORef{
        AAAOrgID: aaaOrgID,
        Name:     orgMap["name"].(string),
        Status:   fpo.FPOStatusActive,
        Metadata: orgMap["metadata"].(map[string]interface{}),
    }

    if err := s.repo.Create(ctx, fpoRef); err != nil {
        return nil, fmt.Errorf("failed to create FPO reference: %w", err)
    }

    log.Printf("Synchronized FPO %s from AAA", fpoRef.ID)
    return fpoRef, nil
}

// Get or sync FPO
func (s *FPOLifecycleService) GetOrSyncFPO(ctx context.Context, aaaOrgID string) (*fpo.FPORef, error) {
    // Try local first
    fpoRef, err := s.repo.FindByAAAOrgID(ctx, aaaOrgID)
    if err == nil && fpoRef != nil {
        return fpoRef, nil
    }

    // Try sync from AAA
    log.Printf("FPO not found locally for org %s, attempting sync", aaaOrgID)
    return s.SyncFPOFromAAA(ctx, aaaOrgID)
}
```

**Commit message**: "feat: implement FPO lifecycle service with core operations"

---

#### Task 3.2: Add Setup Retry Logic
**Description**: Implement retry mechanism for failed setups
**Files to modify**:
- `/internal/services/fpo_lifecycle_service.go`

**Implementation**:
```go
const MaxSetupRetries = 3

func (s *FPOLifecycleService) RetryFailedSetup(ctx context.Context, fpoID string) error {
    fpoRef, err := s.repo.FindByID(ctx, fpoID)
    if err != nil {
        return err
    }

    if fpoRef.Status != fpo.FPOStatusSetupFailed {
        return fmt.Errorf("FPO is not in SETUP_FAILED status")
    }

    // Check retry limit
    fpoRef.SetupAttempts++
    if fpoRef.SetupAttempts > MaxSetupRetries {
        return fmt.Errorf("maximum setup retries (%d) exceeded", MaxSetupRetries)
    }

    // Transition back to pending setup
    if err := s.stateMachine.Transition(ctx, fpoID, fpo.FPOStatusPendingSetup,
        fmt.Sprintf("Retry attempt %d", fpoRef.SetupAttempts), GetUserIDFromContext(ctx)); err != nil {
        return err
    }

    // Update retry tracking
    now := time.Now()
    fpoRef.LastSetupAt = &now
    if err := s.repo.Update(ctx, fpoRef); err != nil {
        return err
    }

    // Trigger setup
    go s.performAAASetup(context.Background(), fpoID)

    return nil
}

func (s *FPOLifecycleService) performAAASetup(ctx context.Context, fpoID string) {
    fpoRef, err := s.repo.FindByID(ctx, fpoID)
    if err != nil {
        log.Printf("Failed to find FPO for setup: %v", err)
        return
    }

    setupErrors := make(map[string]interface{})

    // Create organization in AAA
    if fpoRef.AAAOrgID == "" {
        orgResp, err := s.aaaService.CreateOrganization(ctx, map[string]interface{}{
            "name":        fpoRef.Name,
            "type":        "FPO",
            "metadata":    fpoRef.Metadata,
        })

        if err != nil {
            setupErrors["organization"] = err.Error()
        } else {
            orgMap := orgResp.(map[string]interface{})
            fpoRef.AAAOrgID = orgMap["org_id"].(string)
        }
    }

    // Create user groups
    if fpoRef.AAAOrgID != "" {
        groupErrors := s.createUserGroups(ctx, fpoRef.AAAOrgID)
        for k, v := range groupErrors {
            setupErrors[k] = v
        }
    }

    // Update status based on results
    var newStatus fpo.FPOStatus
    var reason string

    if len(setupErrors) == 0 {
        newStatus = fpo.FPOStatusActive
        reason = "Setup completed successfully"
        fpoRef.SetupErrors = nil
    } else {
        newStatus = fpo.FPOStatusSetupFailed
        reason = fmt.Sprintf("Setup failed with %d errors", len(setupErrors))
        fpoRef.SetupErrors = setupErrors
    }

    // Update FPO
    now := time.Now()
    fpoRef.LastSetupAt = &now
    if err := s.repo.Update(ctx, fpoRef); err != nil {
        log.Printf("Failed to update FPO after setup: %v", err)
        return
    }

    // Transition state
    if err := s.stateMachine.Transition(ctx, fpoID, newStatus, reason, "system"); err != nil {
        log.Printf("Failed to transition FPO state after setup: %v", err)
    }
}
```

**Commit message**: "feat: add setup retry logic with attempt tracking"

---

### Phase 4: API Layer Tasks

#### Task 4.1: Create Lifecycle Request/Response DTOs
**Description**: Define request and response structures for lifecycle operations
**Files to create**:
- `/internal/entities/requests/fpo_lifecycle.go`
- `/internal/entities/responses/fpo_lifecycle.go`

**Implementation**:
```go
// requests/fpo_lifecycle.go
package requests

import "github.com/Kisanlink/farmers-module/internal/entities"

type CreateDraftFPORequest struct {
    Name           string         `json:"name" binding:"required"`
    RegistrationNo string         `json:"registration_number" binding:"required"`
    Description    string         `json:"description"`
    BusinessConfig entities.JSONB `json:"business_config"`
    Metadata       entities.JSONB `json:"metadata"`
}

type VerifyFPORequest struct {
    Approved bool   `json:"approved" binding:"required"`
    Notes    string `json:"notes"`
}

type SuspendFPORequest struct {
    Reason string `json:"reason" binding:"required"`
}

type RejectFPORequest struct {
    Reason string `json:"reason" binding:"required"`
}

// responses/fpo_lifecycle.go
package responses

type FPOLifecycleResponse struct {
    BaseResponse
    Data *FPOLifecycleData `json:"data"`
}

type FPOLifecycleData struct {
    FPOID          string    `json:"fpo_id"`
    Status         string    `json:"status"`
    PreviousStatus string    `json:"previous_status,omitempty"`
    StatusReason   string    `json:"status_reason,omitempty"`
    StatusChangedAt string   `json:"status_changed_at"`
    StatusChangedBy string   `json:"status_changed_by"`
}

type FPOHistoryResponse struct {
    BaseResponse
    Data []*FPOAuditEntry `json:"data"`
}

type FPOAuditEntry struct {
    Action        string    `json:"action"`
    PreviousState string    `json:"previous_state"`
    NewState      string    `json:"new_state"`
    Reason        string    `json:"reason"`
    PerformedBy   string    `json:"performed_by"`
    PerformedAt   string    `json:"performed_at"`
}
```

**Commit message**: "feat: add request/response DTOs for FPO lifecycle operations"

---

#### Task 4.2: Create FPO Lifecycle Handlers
**Description**: Implement HTTP handlers for lifecycle endpoints
**Files to create**:
- `/internal/handlers/fpo_lifecycle_handlers.go`

**Implementation**:
```go
package handlers

import (
    "net/http"

    "github.com/Kisanlink/farmers-module/internal/entities/requests"
    "github.com/Kisanlink/farmers-module/internal/entities/responses"
    "github.com/Kisanlink/farmers-module/internal/services"
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

type FPOLifecycleHandler struct {
    service services.FPOLifecycleService
    logger  interfaces.Logger
}

func NewFPOLifecycleHandler(service services.FPOLifecycleService, logger interfaces.Logger) *FPOLifecycleHandler {
    return &FPOLifecycleHandler{
        service: service,
        logger:  logger,
    }
}

// @Summary Create Draft FPO
// @Description Create a new FPO in draft status
// @Tags FPO Lifecycle
// @Accept json
// @Produce json
// @Param request body requests.CreateDraftFPORequest true "Draft FPO Request"
// @Success 201 {object} responses.FPOLifecycleResponse
// @Router /api/v1/fpo/draft [post]
func (h *FPOLifecycleHandler) CreateDraft(c *gin.Context) {
    var req requests.CreateDraftFPORequest
    if err := c.ShouldBindJSON(&req); err != nil {
        h.logger.Error("Invalid request", zap.Error(err))
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    fpo, err := h.service.CreateDraftFPO(c.Request.Context(), &req)
    if err != nil {
        h.logger.Error("Failed to create draft FPO", zap.Error(err))
        handleServiceError(c, err)
        return
    }

    response := &responses.FPOLifecycleResponse{
        BaseResponse: responses.BaseResponse{
            Success: true,
            Message: "Draft FPO created successfully",
        },
        Data: &responses.FPOLifecycleData{
            FPOID:  fpo.ID,
            Status: fpo.Status.String(),
        },
    }

    c.JSON(http.StatusCreated, response)
}

// @Summary Submit FPO for Verification
// @Description Submit a draft FPO for verification
// @Tags FPO Lifecycle
// @Param id path string true "FPO ID"
// @Success 200 {object} responses.FPOLifecycleResponse
// @Router /api/v1/fpo/{id}/submit [post]
func (h *FPOLifecycleHandler) SubmitForVerification(c *gin.Context) {
    fpoID := c.Param("id")

    if err := h.service.SubmitForVerification(c.Request.Context(), fpoID); err != nil {
        h.logger.Error("Failed to submit FPO", zap.Error(err))
        handleServiceError(c, err)
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "FPO submitted for verification",
    })
}

// @Summary Verify FPO
// @Description Verify or reject an FPO
// @Tags FPO Lifecycle
// @Accept json
// @Produce json
// @Param id path string true "FPO ID"
// @Param request body requests.VerifyFPORequest true "Verification Request"
// @Success 200 {object} responses.FPOLifecycleResponse
// @Router /api/v1/fpo/{id}/verify [post]
func (h *FPOLifecycleHandler) VerifyFPO(c *gin.Context) {
    fpoID := c.Param("id")

    var req requests.VerifyFPORequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if req.Approved {
        if err := h.service.VerifyFPO(c.Request.Context(), fpoID, &req); err != nil {
            handleServiceError(c, err)
            return
        }
        c.JSON(http.StatusOK, gin.H{
            "success": true,
            "message": "FPO verified successfully",
        })
    } else {
        if err := h.service.RejectFPO(c.Request.Context(), fpoID, req.Notes); err != nil {
            handleServiceError(c, err)
            return
        }
        c.JSON(http.StatusOK, gin.H{
            "success": true,
            "message": "FPO rejected",
        })
    }
}

// @Summary Sync FPO from AAA
// @Description Synchronize FPO reference from AAA service
// @Tags FPO Lifecycle
// @Param aaa_org_id path string true "AAA Organization ID"
// @Success 200 {object} responses.FPORefResponse
// @Router /api/v1/fpo/sync/{aaa_org_id} [post]
func (h *FPOLifecycleHandler) SyncFromAAA(c *gin.Context) {
    aaaOrgID := c.Param("aaa_org_id")

    fpo, err := h.service.SyncFPOFromAAA(c.Request.Context(), aaaOrgID)
    if err != nil {
        h.logger.Error("Failed to sync FPO", zap.Error(err))
        handleServiceError(c, err)
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "FPO synchronized successfully",
        "data": fpo,
    })
}

// @Summary Get FPO History
// @Description Get audit history for an FPO
// @Tags FPO Lifecycle
// @Param id path string true "FPO ID"
// @Success 200 {object} responses.FPOHistoryResponse
// @Router /api/v1/fpo/{id}/history [get]
func (h *FPOLifecycleHandler) GetHistory(c *gin.Context) {
    fpoID := c.Param("id")

    history, err := h.service.GetFPOHistory(c.Request.Context(), fpoID)
    if err != nil {
        handleServiceError(c, err)
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data": history,
    })
}
```

**Commit message**: "feat: implement FPO lifecycle HTTP handlers"

---

#### Task 4.3: Register Lifecycle Routes
**Description**: Add lifecycle endpoints to router
**Files to modify**:
- `/internal/routes/fpo_routes.go` (create if doesn't exist)

**Implementation**:
```go
package routes

import (
    "github.com/gin-gonic/gin"
    "github.com/Kisanlink/farmers-module/internal/handlers"
)

func RegisterFPOLifecycleRoutes(router *gin.RouterGroup, handler *handlers.FPOLifecycleHandler) {
    fpo := router.Group("/fpo")
    {
        // Lifecycle endpoints
        fpo.POST("/draft", handler.CreateDraft)
        fpo.POST("/:id/submit", handler.SubmitForVerification)
        fpo.POST("/:id/verify", handler.VerifyFPO)
        fpo.POST("/:id/reject", handler.RejectFPO)
        fpo.POST("/:id/setup", handler.InitializeSetup)
        fpo.POST("/:id/complete-setup", handler.CompleteSetup)
        fpo.POST("/:id/retry-setup", handler.RetrySetup)
        fpo.PUT("/:id/suspend", handler.SuspendFPO)
        fpo.PUT("/:id/reactivate", handler.ReactivateFPO)
        fpo.DELETE("/:id/deactivate", handler.DeactivateFPO)
        fpo.POST("/:id/archive", handler.ArchiveFPO)

        // Sync and recovery
        fpo.POST("/sync/:aaa_org_id", handler.SyncFromAAA)

        // Queries
        fpo.GET("/:id", handler.GetFPO)
        fpo.GET("/by-org/:aaa_org_id", handler.GetByOrgID)
        fpo.GET("/:id/history", handler.GetHistory)
        fpo.GET("/list", handler.ListFPOs)
    }
}
```

**Commit message**: "feat: register FPO lifecycle routes"

---

### Phase 5: Integration Tasks

#### Task 5.1: Update Service Factory
**Description**: Wire up new services in factory
**Files to modify**:
- `/internal/services/service_factory.go`

**Implementation**:
```go
// Add to ServiceFactory struct
FPOLifecycle FPOLifecycleService

// Add to NewServiceFactory
fpoRepo := fpo.NewFPORepository(dbManager)
factory.FPOLifecycle = NewFPOLifecycleService(fpoRepo, factory.AAA)
```

**Commit message**: "feat: integrate FPO lifecycle service in factory"

---

#### Task 5.2: Update Existing FPO Service
**Description**: Modify existing service to use lifecycle service
**Files to modify**:
- `/internal/services/fpo_ref_service.go`

**Implementation**:
```go
// Modify GetFPORef to use GetOrSyncFPO
func (s *FPOServiceImpl) GetFPORef(ctx context.Context, orgID string) (interface{}, error) {
    log.Printf("FPOService: Getting FPO reference for org ID: %s", orgID)

    // Use lifecycle service for get or sync
    fpoRef, err := s.lifecycleService.GetOrSyncFPO(ctx, orgID)
    if err != nil {
        return nil, fmt.Errorf("failed to get FPO reference: %w", err)
    }

    // Convert to response
    responseData := &responses.FPORefData{
        ID:             fpoRef.ID,
        AAAOrgID:       fpoRef.AAAOrgID,
        Name:           fpoRef.Name,
        RegistrationNo: fpoRef.RegistrationNo,
        BusinessConfig: fpoRef.BusinessConfig,
        Status:         fpoRef.Status.String(),
        CreatedAt:      fpoRef.CreatedAt.Format(time.RFC3339),
        UpdatedAt:      fpoRef.UpdatedAt.Format(time.RFC3339),
    }

    log.Printf("Successfully retrieved FPO reference: %s", fpoRef.ID)
    return responseData, nil
}
```

**Commit message**: "feat: integrate lifecycle service with existing FPO operations"

---

### Phase 6: Testing Tasks

#### Task 6.1: Unit Tests for State Machine
**Description**: Test state transitions and validation
**Files to create**:
- `/internal/services/fpo_state_machine_test.go`

**Commit message**: "test: add unit tests for FPO state machine"

---

#### Task 6.2: Integration Tests for Lifecycle
**Description**: Test complete lifecycle workflows
**Files to create**:
- `/internal/services/fpo_lifecycle_service_test.go`

**Commit message**: "test: add integration tests for FPO lifecycle"

---

#### Task 6.3: API Tests for Endpoints
**Description**: Test HTTP endpoints
**Files to create**:
- `/internal/handlers/fpo_lifecycle_handlers_test.go`

**Commit message**: "test: add API tests for lifecycle endpoints"

---

## Database Migration Script

```sql
-- Migration: Add FPO lifecycle fields
BEGIN;

-- Add new columns to fpo_refs
ALTER TABLE fpo_refs
ADD COLUMN IF NOT EXISTS previous_status VARCHAR(50),
ADD COLUMN IF NOT EXISTS status_reason TEXT,
ADD COLUMN IF NOT EXISTS status_changed_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS status_changed_by VARCHAR(255),
ADD COLUMN IF NOT EXISTS verification_status VARCHAR(50),
ADD COLUMN IF NOT EXISTS verified_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS verified_by VARCHAR(255),
ADD COLUMN IF NOT EXISTS verification_notes TEXT,
ADD COLUMN IF NOT EXISTS setup_attempts INT DEFAULT 0,
ADD COLUMN IF NOT EXISTS last_setup_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS setup_progress JSONB DEFAULT '{}',
ADD COLUMN IF NOT EXISTS ceo_user_id VARCHAR(255),
ADD COLUMN IF NOT EXISTS parent_fpo_id VARCHAR(255);

-- Create audit log table
CREATE TABLE IF NOT EXISTS fpo_audit_logs (
    id VARCHAR(255) PRIMARY KEY,
    fpo_id VARCHAR(255) NOT NULL,
    action VARCHAR(100) NOT NULL,
    previous_state VARCHAR(50),
    new_state VARCHAR(50),
    reason TEXT,
    performed_by VARCHAR(255) NOT NULL,
    performed_at TIMESTAMP NOT NULL,
    details JSONB,
    request_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Add indexes
CREATE INDEX IF NOT EXISTS idx_fpo_refs_status ON fpo_refs(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_fpo_refs_aaa_org_id ON fpo_refs(aaa_org_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_fpo_refs_ceo_user_id ON fpo_refs(ceo_user_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_fpo_audit_logs_fpo_id ON fpo_audit_logs(fpo_id);
CREATE INDEX IF NOT EXISTS idx_fpo_audit_logs_performed_at ON fpo_audit_logs(performed_at DESC);

-- Update existing records
UPDATE fpo_refs
SET status = 'ACTIVE'
WHERE status = 'ACTIVE' AND previous_status IS NULL;

UPDATE fpo_refs
SET status = 'SETUP_FAILED',
    status_reason = 'Legacy pending setup - requires retry'
WHERE status = 'PENDING_SETUP';

COMMIT;
```

## Rollback Script

```sql
-- Rollback: Remove FPO lifecycle fields
BEGIN;

-- Drop audit log table
DROP TABLE IF EXISTS fpo_audit_logs;

-- Remove added columns
ALTER TABLE fpo_refs
DROP COLUMN IF EXISTS previous_status,
DROP COLUMN IF EXISTS status_reason,
DROP COLUMN IF EXISTS status_changed_at,
DROP COLUMN IF EXISTS status_changed_by,
DROP COLUMN IF EXISTS verification_status,
DROP COLUMN IF EXISTS verified_at,
DROP COLUMN IF EXISTS verified_by,
DROP COLUMN IF EXISTS verification_notes,
DROP COLUMN IF EXISTS setup_attempts,
DROP COLUMN IF EXISTS last_setup_at,
DROP COLUMN IF EXISTS setup_progress,
DROP COLUMN IF EXISTS ceo_user_id,
DROP COLUMN IF EXISTS parent_fpo_id;

-- Restore original status values
UPDATE fpo_refs
SET status = 'PENDING_SETUP'
WHERE status = 'SETUP_FAILED';

COMMIT;
```

## Testing Plan

### Unit Test Coverage
- State machine transitions (all valid and invalid paths)
- Repository methods (CRUD, status updates, audit logging)
- Service layer business logic
- Request/response validation

### Integration Test Scenarios
1. Complete lifecycle: Draft → Verified → Active
2. Rejection and resubmission flow
3. Setup failure and retry mechanism
4. FPO synchronization from AAA
5. Concurrent state transitions
6. Audit log generation

### E2E Test Cases
1. Create draft FPO via API
2. Submit and verify FPO
3. Initialize and complete setup
4. Handle setup failures
5. Sync missing FPO from AAA
6. Query FPO history

## Deployment Checklist

- [ ] Database migration executed successfully
- [ ] All unit tests passing
- [ ] Integration tests passing
- [ ] API documentation updated in Swagger
- [ ] Environment variables configured
- [ ] AAA service integration tested
- [ ] Monitoring metrics configured
- [ ] Rollback plan tested in staging
- [ ] Performance benchmarks met
- [ ] Security review completed

## Success Metrics

1. **Functional Metrics**
   - All lifecycle states functioning: 100%
   - State transition accuracy: 100%
   - Audit log completeness: 100%
   - AAA sync success rate: > 95%

2. **Performance Metrics**
   - API response time (P95): < 500ms
   - Database query time (P95): < 100ms
   - Sync operation time: < 2s
   - Concurrent request handling: > 100 req/s

3. **Quality Metrics**
   - Code coverage: > 80%
   - Zero critical bugs in production
   - All endpoints documented
   - Zero security vulnerabilities

## Risk Mitigation

1. **Data Migration Risk**
   - Test migration on staging data copy
   - Keep backup before migration
   - Prepare rollback script

2. **AAA Integration Risk**
   - Implement circuit breaker
   - Add retry with exponential backoff
   - Cache AAA responses where possible

3. **Performance Risk**
   - Add database indexes
   - Implement caching layer
   - Monitor query performance

4. **Backward Compatibility**
   - Keep existing endpoints functional
   - Use feature flags for gradual rollout
   - Version new APIs appropriately
