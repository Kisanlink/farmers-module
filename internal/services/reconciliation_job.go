package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/Kisanlink/farmers-module/internal/constants"
	cropcycle "github.com/Kisanlink/farmers-module/internal/entities/crop_cycle"
	"github.com/Kisanlink/farmers-module/internal/entities/farm"
	farmactivity "github.com/Kisanlink/farmers-module/internal/entities/farm_activity"
	farmerentity "github.com/Kisanlink/farmers-module/internal/entities/farmer"
	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"gorm.io/gorm"
)

// ReconciliationJob handles periodic reconciliation of pending operations
type ReconciliationJob struct {
	db         *gorm.DB
	aaaService AAAService
	logger     interfaces.Logger
	interval   time.Duration
	stopCh     chan struct{}
	wg         sync.WaitGroup
	running    bool
	mu         sync.Mutex
}

// ReconciliationReport contains the results of a reconciliation run
type ReconciliationReport struct {
	StartTime            time.Time `json:"start_time"`
	EndTime              time.Time `json:"end_time"`
	Duration             string    `json:"duration"`
	RolesProcessed       int       `json:"roles_processed"`
	RolesFixed           int       `json:"roles_fixed"`
	RolesStillPending    int       `json:"roles_still_pending"`
	OrphanedDeleted      int       `json:"orphaned_deleted"`
	FPOLinksProcessed    int       `json:"fpo_links_processed"`
	FPOLinksFixed        int       `json:"fpo_links_fixed"`
	FPOLinksStillPending int       `json:"fpo_links_still_pending"`
	Errors               []string  `json:"errors,omitempty"`
}

// NewReconciliationJob creates a new reconciliation job
func NewReconciliationJob(db *gorm.DB, aaaService AAAService, logger interfaces.Logger, interval time.Duration) *ReconciliationJob {
	if interval == 0 {
		interval = 6 * time.Hour // Default: run 4 times per day
	}
	return &ReconciliationJob{
		db:         db,
		aaaService: aaaService,
		logger:     logger,
		interval:   interval,
		stopCh:     make(chan struct{}),
	}
}

// Start begins the reconciliation job
func (j *ReconciliationJob) Start() {
	j.mu.Lock()
	if j.running {
		j.mu.Unlock()
		return
	}
	j.running = true
	j.mu.Unlock()

	j.wg.Add(1)
	go j.run()
	log.Printf("Reconciliation job started (interval: %s)", j.interval)
}

// Stop gracefully stops the reconciliation job
func (j *ReconciliationJob) Stop() {
	j.mu.Lock()
	if !j.running {
		j.mu.Unlock()
		return
	}
	j.running = false
	j.mu.Unlock()

	close(j.stopCh)
	j.wg.Wait()
	log.Println("Reconciliation job stopped")
}

func (j *ReconciliationJob) run() {
	defer j.wg.Done()

	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			j.runOnce()
		case <-j.stopCh:
			return
		}
	}
}

func (j *ReconciliationJob) runOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	report, err := j.Reconcile(ctx)
	if err != nil {
		log.Printf("Reconciliation job failed: %v", err)
		return
	}

	// Only log if there was work to do
	if report.RolesProcessed > 0 || report.FPOLinksProcessed > 0 || report.OrphanedDeleted > 0 {
		log.Printf("Reconciliation completed: roles=%d/%d fixed, orphaned=%d deleted, fpo_links=%d/%d fixed, duration=%s",
			report.RolesFixed, report.RolesProcessed,
			report.OrphanedDeleted,
			report.FPOLinksFixed, report.FPOLinksProcessed,
			report.Duration)
	}
}

// Reconcile performs the reconciliation and returns a report
func (j *ReconciliationJob) Reconcile(ctx context.Context) (*ReconciliationReport, error) {
	report := &ReconciliationReport{
		StartTime: time.Now(),
		Errors:    []string{},
	}

	// First, clean up orphaned farmers (users that don't exist in AAA)
	j.cleanupOrphanedFarmers(ctx, report)

	// Reconcile pending role assignments
	j.reconcileRoleAssignments(ctx, report)

	// Reconcile pending FPO config links
	j.reconcileFPOConfigLinks(ctx, report)

	report.EndTime = time.Now()
	report.Duration = report.EndTime.Sub(report.StartTime).String()

	return report, nil
}

// cleanupOrphanedFarmers checks all farmers and permanently deletes those whose AAA user doesn't exist
func (j *ReconciliationJob) cleanupOrphanedFarmers(ctx context.Context, report *ReconciliationReport) {
	// Query all farmers in batches
	var farmers []farmerentity.Farmer
	err := j.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Limit(100). // Process in batches
		Find(&farmers).Error

	if err != nil {
		report.Errors = append(report.Errors, fmt.Sprintf("failed to query farmers for orphan cleanup: %v", err))
		return
	}

	for _, farmer := range farmers {
		select {
		case <-ctx.Done():
			report.Errors = append(report.Errors, "context cancelled during orphan cleanup")
			return
		default:
		}

		// Check if the AAA user exists
		userExists, err := j.checkUserExists(ctx, farmer.AAAUserID)
		if err != nil {
			// Network/service error - skip this farmer
			continue
		}

		if !userExists {
			// User doesn't exist in AAA - permanently delete this orphaned farmer
			log.Printf("Orphaned farmer detected: %s (AAA user %s not found), permanently deleting", farmer.GetID(), farmer.AAAUserID)
			err := j.permanentlyDeleteOrphanedFarmer(ctx, &farmer)
			if err != nil {
				report.Errors = append(report.Errors, fmt.Sprintf("farmer %s: failed to delete orphaned record: %v", farmer.AAAUserID, err))
				continue
			}
			report.OrphanedDeleted++
		}
	}
}

// reconcileRoleAssignments retries failed role assignments
func (j *ReconciliationJob) reconcileRoleAssignments(ctx context.Context, report *ReconciliationReport) {
	// Query farmers with role_assignment_pending = true
	var farmers []farmerentity.Farmer
	err := j.db.WithContext(ctx).
		Where("metadata->>'role_assignment_pending' = ?", "true").
		Where("deleted_at IS NULL").
		Limit(100). // Process in batches
		Find(&farmers).Error

	if err != nil {
		report.Errors = append(report.Errors, fmt.Sprintf("failed to query pending role assignments: %v", err))
		return
	}

	report.RolesProcessed = len(farmers)

	for _, farmer := range farmers {
		select {
		case <-ctx.Done():
			report.Errors = append(report.Errors, "context cancelled during role reconciliation")
			return
		default:
		}

		// First check if the AAA user exists
		userExists, err := j.checkUserExists(ctx, farmer.AAAUserID)
		if err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("farmer %s: failed to check user existence: %v", farmer.AAAUserID, err))
			report.RolesStillPending++
			continue
		}

		if !userExists {
			// User doesn't exist in AAA - permanently delete this orphaned farmer
			log.Printf("Orphaned farmer detected: %s (AAA user %s not found), permanently deleting", farmer.GetID(), farmer.AAAUserID)
			err := j.permanentlyDeleteOrphanedFarmer(ctx, &farmer)
			if err != nil {
				report.Errors = append(report.Errors, fmt.Sprintf("farmer %s: failed to delete orphaned record: %v", farmer.AAAUserID, err))
				report.RolesStillPending++
				continue
			}
			report.OrphanedDeleted++
			continue
		}

		err = j.retryRoleAssignment(ctx, &farmer)
		if err != nil {
			report.RolesStillPending++
			report.Errors = append(report.Errors, fmt.Sprintf("farmer %s: %v", farmer.AAAUserID, err))
			continue
		}

		report.RolesFixed++
	}
}

// retryRoleAssignment attempts to assign the farmer role and clears pending flag on success
func (j *ReconciliationJob) retryRoleAssignment(ctx context.Context, farmer *farmerentity.Farmer) error {
	// Check if user already has the role
	hasRole, err := j.aaaService.CheckUserRole(ctx, farmer.AAAUserID, constants.RoleFarmer)
	if err != nil {
		return fmt.Errorf("failed to check role: %w", err)
	}

	if !hasRole {
		// Try to assign the role
		err = j.aaaService.AssignRole(ctx, farmer.AAAUserID, farmer.AAAOrgID, constants.RoleFarmer)
		if err != nil {
			return fmt.Errorf("failed to assign role: %w", err)
		}

		// Wait briefly for role assignment to propagate in AAA service
		time.Sleep(500 * time.Millisecond)

		// Verify assignment with retry
		var verificationErr error
		for attempt := 0; attempt < 3; attempt++ {
			hasRole, err = j.aaaService.CheckUserRole(ctx, farmer.AAAUserID, constants.RoleFarmer)
			if err != nil {
				verificationErr = fmt.Errorf("failed to verify role: %w", err)
				time.Sleep(200 * time.Millisecond)
				continue
			}
			if hasRole {
				log.Printf("Role farmer assigned successfully to user %s (verified on attempt %d)", farmer.AAAUserID, attempt+1)
				break
			}
			verificationErr = fmt.Errorf("role assignment verification failed")
			time.Sleep(200 * time.Millisecond)
		}

		if !hasRole {
			return verificationErr
		}
	}

	// Clear the pending flag
	return j.clearRolePendingFlag(ctx, farmer)
}

func (j *ReconciliationJob) clearRolePendingFlag(ctx context.Context, farmer *farmerentity.Farmer) error {
	// Update metadata to remove pending flags
	updates := map[string]interface{}{
		"metadata": gorm.Expr(`
			metadata - 'role_assignment_pending' - 'role_assignment_error' - 'role_assignment_attempted_at'
			|| jsonb_build_object('role_assignment_fixed_at', ?)
		`, time.Now().Format(time.RFC3339)),
	}

	return j.db.WithContext(ctx).
		Model(&farmerentity.Farmer{}).
		Where("id = ?", farmer.GetID()).
		Updates(updates).Error
}

// reconcileFPOConfigLinks retries failed FPO config links
func (j *ReconciliationJob) reconcileFPOConfigLinks(ctx context.Context, report *ReconciliationReport) {
	// Query farmers with fpo_config_link_pending = true
	var farmers []farmerentity.Farmer
	err := j.db.WithContext(ctx).
		Where("metadata->>'fpo_config_link_pending' = ?", "true").
		Where("deleted_at IS NULL").
		Limit(100). // Process in batches
		Find(&farmers).Error

	if err != nil {
		report.Errors = append(report.Errors, fmt.Sprintf("failed to query pending FPO config links: %v", err))
		return
	}

	report.FPOLinksProcessed = len(farmers)

	for _, farmer := range farmers {
		select {
		case <-ctx.Done():
			report.Errors = append(report.Errors, "context cancelled during FPO link reconciliation")
			return
		default:
		}

		// For FPO config links, we just clear the pending flag for now
		// The actual linking can be retried via the API
		err := j.clearFPOLinkPendingFlag(ctx, &farmer)
		if err != nil {
			report.FPOLinksStillPending++
			report.Errors = append(report.Errors, fmt.Sprintf("farmer %s: failed to clear FPO link pending: %v", farmer.AAAUserID, err))
			continue
		}

		report.FPOLinksFixed++
	}
}

func (j *ReconciliationJob) clearFPOLinkPendingFlag(ctx context.Context, farmer *farmerentity.Farmer) error {
	// Just mark as acknowledged - actual retry can be done via API
	updates := map[string]interface{}{
		"metadata": gorm.Expr(`
			metadata - 'fpo_config_link_pending'
			|| jsonb_build_object('fpo_config_link_acknowledged_at', ?)
		`, time.Now().Format(time.RFC3339)),
	}

	return j.db.WithContext(ctx).
		Model(&farmerentity.Farmer{}).
		Where("id = ?", farmer.GetID()).
		Updates(updates).Error
}

// RunNow triggers an immediate reconciliation run (for manual/API triggers)
func (j *ReconciliationJob) RunNow(ctx context.Context) (*ReconciliationReport, error) {
	return j.Reconcile(ctx)
}

// GetPendingCounts returns counts of pending items without processing them
func (j *ReconciliationJob) GetPendingCounts(ctx context.Context) (rolesPending, fpoLinksPending int64, err error) {
	err = j.db.WithContext(ctx).
		Model(&farmerentity.Farmer{}).
		Where("metadata->>'role_assignment_pending' = ?", "true").
		Where("deleted_at IS NULL").
		Count(&rolesPending).Error
	if err != nil {
		return 0, 0, fmt.Errorf("failed to count pending roles: %w", err)
	}

	err = j.db.WithContext(ctx).
		Model(&farmerentity.Farmer{}).
		Where("metadata->>'fpo_config_link_pending' = ?", "true").
		Where("deleted_at IS NULL").
		Count(&fpoLinksPending).Error
	if err != nil {
		return 0, 0, fmt.Errorf("failed to count pending FPO links: %w", err)
	}

	return rolesPending, fpoLinksPending, nil
}

// checkUserExists checks if the user exists in AAA service
func (j *ReconciliationJob) checkUserExists(ctx context.Context, userID string) (bool, error) {
	_, err := j.aaaService.GetUser(ctx, userID)
	if err != nil {
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "NotFound") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// permanentlyDeleteOrphanedFarmer permanently deletes a farmer and all related data
func (j *ReconciliationJob) permanentlyDeleteOrphanedFarmer(ctx context.Context, farmer *farmerentity.Farmer) error {
	return j.db.Transaction(func(tx *gorm.DB) error {
		farmerID := farmer.GetID()

		// Get all farms for this farmer
		var farms []farm.Farm
		if err := tx.Unscoped().Where("farmer_id = ?", farmerID).Find(&farms).Error; err != nil {
			return fmt.Errorf("failed to find farms: %w", err)
		}

		// Delete farm activities for each farm
		for _, f := range farms {
			if err := tx.Unscoped().Where("farm_id = ?", f.GetID()).Delete(&farmactivity.FarmActivity{}).Error; err != nil {
				return fmt.Errorf("failed to delete farm activities: %w", err)
			}
		}

		// Delete crop cycles for each farm
		for _, f := range farms {
			if err := tx.Unscoped().Where("farm_id = ?", f.GetID()).Delete(&cropcycle.CropCycle{}).Error; err != nil {
				return fmt.Errorf("failed to delete crop cycles: %w", err)
			}
		}

		// Delete farms
		if err := tx.Unscoped().Where("farmer_id = ?", farmerID).Delete(&farm.Farm{}).Error; err != nil {
			return fmt.Errorf("failed to delete farms: %w", err)
		}

		// Delete farmer links
		if err := tx.Unscoped().Where("farmer_id = ?", farmerID).Delete(&farmerentity.FarmerLink{}).Error; err != nil {
			return fmt.Errorf("failed to delete farmer links: %w", err)
		}

		// Delete the farmer
		if err := tx.Unscoped().Where("id = ?", farmerID).Delete(&farmerentity.Farmer{}).Error; err != nil {
			return fmt.Errorf("failed to delete farmer: %w", err)
		}

		log.Printf("Permanently deleted orphaned farmer %s and all related data", farmerID)
		return nil
	})
}

// FarmerRepository interface for reconciliation (uses base filterable pattern)
type FarmerReconciliationRepo interface {
	Find(ctx context.Context, filter *base.Filter) ([]*farmerentity.Farmer, error)
	Update(ctx context.Context, entity *farmerentity.Farmer) error
}
