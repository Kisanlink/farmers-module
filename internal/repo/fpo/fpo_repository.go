package fpo

import (
	"context"
	"fmt"
	"time"

	"github.com/Kisanlink/farmers-module/internal/entities/fpo"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"gorm.io/gorm"
)

// FPORepository provides data access methods for FPO management
type FPORepository struct {
	*base.BaseFilterableRepository[*fpo.FPORef]
	auditRepo *base.BaseFilterableRepository[*fpo.FPOAuditLog]
	db        *gorm.DB
}

// NewFPORepository creates a new FPO repository instance
func NewFPORepository(dbManager interface{}) *FPORepository {
	repo := &FPORepository{
		BaseFilterableRepository: base.NewBaseFilterableRepository[*fpo.FPORef](),
		auditRepo:                base.NewBaseFilterableRepository[*fpo.FPOAuditLog](),
	}
	repo.SetDBManager(dbManager)
	repo.auditRepo.SetDBManager(dbManager)

	// Get the GORM DB instance with proper interface signature
	if postgresManager, ok := dbManager.(interface {
		GetDB(context.Context, bool) (*gorm.DB, error)
	}); ok {
		if gormDB, err := postgresManager.GetDB(context.Background(), false); err == nil {
			repo.db = gormDB
		}
	}

	return repo
}

// FindByID finds an FPO by ID
func (r *FPORepository) FindByID(ctx context.Context, id string) (*fpo.FPORef, error) {
	filter := base.NewFilterBuilder().
		Where("id", base.OpEqual, id).
		Build()
	return r.FindOne(ctx, filter)
}

// FindByAAAOrgID finds an FPO by AAA organization ID
func (r *FPORepository) FindByAAAOrgID(ctx context.Context, aaaOrgID string) (*fpo.FPORef, error) {
	filter := base.NewFilterBuilder().
		Where("aaa_org_id", base.OpEqual, aaaOrgID).
		Build()
	return r.FindOne(ctx, filter)
}

// FindByRegistrationNo finds an FPO by registration number
func (r *FPORepository) FindByRegistrationNo(ctx context.Context, registrationNo string) (*fpo.FPORef, error) {
	filter := base.NewFilterBuilder().
		Where("registration_number", base.OpEqual, registrationNo).
		Build()
	return r.FindOne(ctx, filter)
}

// FindByStatus finds all FPOs with a specific status
func (r *FPORepository) FindByStatus(ctx context.Context, status fpo.FPOStatus) ([]*fpo.FPORef, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	var results []*fpo.FPORef
	err := r.db.WithContext(ctx).
		Where("status = ? AND deleted_at IS NULL", string(status)).
		Find(&results).Error
	return results, err
}

// UpdateStatus updates the FPO status with audit logging
func (r *FPORepository) UpdateStatus(ctx context.Context, fpoID string, newStatus fpo.FPOStatus, reason string, performedBy string) error {
	if r.db == nil {
		return fmt.Errorf("database connection not available")
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		var fpoRef fpo.FPORef
		if err := tx.First(&fpoRef, "id = ?", fpoID).Error; err != nil {
			return fmt.Errorf("failed to find FPO: %w", err)
		}

		// Validate state transition
		if !fpoRef.Status.CanTransitionTo(newStatus) {
			return fmt.Errorf("invalid state transition from %s to %s", fpoRef.Status, newStatus)
		}

		now := time.Now()

		// Create audit log
		auditLog := &fpo.FPOAuditLog{
			FPOID:         fpoID,
			Action:        "STATUS_CHANGE",
			PreviousState: fpoRef.Status,
			NewState:      newStatus,
			Reason:        reason,
			PerformedBy:   performedBy,
			PerformedAt:   now,
		}

		// Update FPO status
		updates := map[string]interface{}{
			"previous_status":   fpoRef.Status,
			"status":            newStatus,
			"status_reason":     reason,
			"status_changed_at": now,
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

// GetAuditHistory retrieves audit history for an FPO
func (r *FPORepository) GetAuditHistory(ctx context.Context, fpoID string) ([]*fpo.FPOAuditLog, error) {
	if r.db == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	var results []*fpo.FPOAuditLog
	err := r.db.WithContext(ctx).
		Where("fpo_id = ? AND deleted_at IS NULL", fpoID).
		Order("performed_at DESC").
		Find(&results).Error
	return results, err
}

// IncrementSetupAttempt increments the setup attempt counter
func (r *FPORepository) IncrementSetupAttempt(ctx context.Context, fpoID string) error {
	if r.db == nil {
		return fmt.Errorf("database connection not available")
	}

	return r.db.Model(&fpo.FPORef{}).
		Where("id = ?", fpoID).
		UpdateColumn("setup_attempts", gorm.Expr("setup_attempts + ?", 1)).
		Error
}

// UpdateCEO updates the CEO user ID for an FPO
func (r *FPORepository) UpdateCEO(ctx context.Context, fpoID string, ceoUserID string) error {
	if r.db == nil {
		return fmt.Errorf("database connection not available")
	}

	result := r.db.WithContext(ctx).
		Model(&fpo.FPORef{}).
		Where("id = ?", fpoID).
		Update("ceo_user_id", ceoUserID)

	if result.Error != nil {
		return fmt.Errorf("failed to update CEO: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("FPO not found: %s", fpoID)
	}

	return nil
}
