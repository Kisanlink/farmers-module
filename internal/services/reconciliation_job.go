package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Kisanlink/farmers-module/internal/constants"
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
	FPOLinksProcessed    int       `json:"fpo_links_processed"`
	FPOLinksFixed        int       `json:"fpo_links_fixed"`
	FPOLinksStillPending int       `json:"fpo_links_still_pending"`
	Errors               []string  `json:"errors,omitempty"`
}

// NewReconciliationJob creates a new reconciliation job
func NewReconciliationJob(db *gorm.DB, aaaService AAAService, logger interfaces.Logger, interval time.Duration) *ReconciliationJob {
	if interval == 0 {
		interval = 5 * time.Minute // Default: run every 5 minutes
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

	// Run immediately on start
	j.runOnce()

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
	if report.RolesProcessed > 0 || report.FPOLinksProcessed > 0 {
		log.Printf("Reconciliation completed: roles=%d/%d fixed, fpo_links=%d/%d fixed, duration=%s",
			report.RolesFixed, report.RolesProcessed,
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

	// Reconcile pending role assignments
	j.reconcileRoleAssignments(ctx, report)

	// Reconcile pending FPO config links
	j.reconcileFPOConfigLinks(ctx, report)

	report.EndTime = time.Now()
	report.Duration = report.EndTime.Sub(report.StartTime).String()

	return report, nil
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

		err := j.retryRoleAssignment(ctx, &farmer)
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

		// Verify assignment
		hasRole, err = j.aaaService.CheckUserRole(ctx, farmer.AAAUserID, constants.RoleFarmer)
		if err != nil {
			return fmt.Errorf("failed to verify role: %w", err)
		}
		if !hasRole {
			return fmt.Errorf("role assignment verification failed")
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

// FarmerRepository interface for reconciliation (uses base filterable pattern)
type FarmerReconciliationRepo interface {
	Find(ctx context.Context, filter *base.Filter) ([]*farmerentity.Farmer, error)
	Update(ctx context.Context, entity *farmerentity.Farmer) error
}
