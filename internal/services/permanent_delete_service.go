package services

import (
	"context"
	"fmt"
	"log"
	"time"

	cropcycle "github.com/Kisanlink/farmers-module/internal/entities/crop_cycle"
	"github.com/Kisanlink/farmers-module/internal/entities/farm"
	farmactivity "github.com/Kisanlink/farmers-module/internal/entities/farm_activity"
	farmirrigation "github.com/Kisanlink/farmers-module/internal/entities/farm_irrigation_source"
	farmsoil "github.com/Kisanlink/farmers-module/internal/entities/farm_soil_type"
	farmerentity "github.com/Kisanlink/farmers-module/internal/entities/farmer"
	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"gorm.io/gorm"
)

// PermanentDeleteService handles hard deletes with cascade for super admins
type PermanentDeleteService struct {
	db         *gorm.DB
	aaaService AAAService
	logger     interfaces.Logger
}

// DeleteReport contains the results of a permanent delete operation
type DeleteReport struct {
	EntityType        string    `json:"entity_type"`
	EntityID          string    `json:"entity_id"`
	DeletedAt         time.Time `json:"deleted_at"`
	FarmersDeleted    int64     `json:"farmers_deleted,omitempty"`
	FarmsDeleted      int64     `json:"farms_deleted,omitempty"`
	CropCyclesDeleted int64     `json:"crop_cycles_deleted,omitempty"`
	ActivitiesDeleted int64     `json:"activities_deleted,omitempty"`
	LinksDeleted      int64     `json:"links_deleted,omitempty"`
	AddressesDeleted  int64     `json:"addresses_deleted,omitempty"`
	IrrigationDeleted int64     `json:"irrigation_deleted,omitempty"`
	SoilTypesDeleted  int64     `json:"soil_types_deleted,omitempty"`
	Errors            []string  `json:"errors,omitempty"`
}

// NewPermanentDeleteService creates a new permanent delete service
func NewPermanentDeleteService(db *gorm.DB, aaaService AAAService, logger interfaces.Logger) *PermanentDeleteService {
	return &PermanentDeleteService{
		db:         db,
		aaaService: aaaService,
		logger:     logger,
	}
}

// CanUserPerformPermanentDelete checks if the user has super admin privileges
func (s *PermanentDeleteService) CanUserPerformPermanentDelete(ctx context.Context, userID string) (bool, error) {
	// Use CheckPermission to verify admin-level access
	// Check for permanent_delete action on admin resource
	allowed, err := s.aaaService.CheckPermission(ctx, userID, "admin", "permanent_delete", "", "")
	if err != nil {
		log.Printf("Permission check failed for permanent_delete: %v", err)
		// Fallback: try checking for admin.maintain permission
		allowed, err = s.aaaService.CheckPermission(ctx, userID, "admin", "maintain", "", "")
		if err != nil {
			return false, fmt.Errorf("failed to check admin permission: %w", err)
		}
	}
	return allowed, nil
}

// PermanentDeleteFarmer permanently deletes a farmer and all related data
func (s *PermanentDeleteService) PermanentDeleteFarmer(ctx context.Context, farmerID string, deletedBy string) (*DeleteReport, error) {
	report := &DeleteReport{
		EntityType: "farmer",
		EntityID:   farmerID,
		DeletedAt:  time.Now(),
		Errors:     []string{},
	}

	// Get farmer first to retrieve aaa_user_id and address_id
	var farmer farmerentity.Farmer
	if err := s.db.WithContext(ctx).Unscoped().First(&farmer, "id = ?", farmerID).Error; err != nil {
		return nil, fmt.Errorf("farmer not found: %w", err)
	}

	// Get all farm IDs for this farmer (needed for sub-deletions)
	var farmIDs []string
	s.db.WithContext(ctx).Model(&farm.Farm{}).Unscoped().
		Where("farmer_id = ?", farmerID).
		Pluck("id", &farmIDs)

	// Execute cascade delete in a transaction
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Delete farm activities (depends on crop_cycles and farmer)
		result := tx.Unscoped().Where("farmer_id = ?", farmerID).Delete(&farmactivity.FarmActivity{})
		if result.Error != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("farm_activities: %v", result.Error))
		}
		report.ActivitiesDeleted = result.RowsAffected

		// 2. Delete crop cycles (depends on farms and farmer)
		result = tx.Unscoped().Where("farmer_id = ?", farmerID).Delete(&cropcycle.CropCycle{})
		if result.Error != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("crop_cycles: %v", result.Error))
		}
		report.CropCyclesDeleted = result.RowsAffected

		// 3. Delete farm irrigation sources (depends on farms)
		if len(farmIDs) > 0 {
			result = tx.Unscoped().Where("farm_id IN ?", farmIDs).Delete(&farmirrigation.FarmIrrigationSource{})
			if result.Error != nil {
				report.Errors = append(report.Errors, fmt.Sprintf("farm_irrigation_sources: %v", result.Error))
			}
			report.IrrigationDeleted = result.RowsAffected

			// 4. Delete farm soil types (depends on farms)
			result = tx.Unscoped().Where("farm_id IN ?", farmIDs).Delete(&farmsoil.FarmSoilType{})
			if result.Error != nil {
				report.Errors = append(report.Errors, fmt.Sprintf("farm_soil_types: %v", result.Error))
			}
			report.SoilTypesDeleted = result.RowsAffected
		}

		// 5. Delete farms
		result = tx.Unscoped().Where("farmer_id = ?", farmerID).Delete(&farm.Farm{})
		if result.Error != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("farms: %v", result.Error))
		}
		report.FarmsDeleted = result.RowsAffected

		// 6. Delete farmer links
		result = tx.Unscoped().Where("aaa_user_id = ?", farmer.AAAUserID).Delete(&farmerentity.FarmerLink{})
		if result.Error != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("farmer_links: %v", result.Error))
		}
		report.LinksDeleted = result.RowsAffected

		// 7. Delete address if exists
		if farmer.AddressID != nil && *farmer.AddressID != "" {
			result = tx.Unscoped().Where("id = ?", *farmer.AddressID).Delete(&farmerentity.Address{})
			if result.Error != nil {
				report.Errors = append(report.Errors, fmt.Sprintf("addresses: %v", result.Error))
			}
			report.AddressesDeleted = result.RowsAffected
		}

		// 8. Delete farmer
		result = tx.Unscoped().Where("id = ?", farmerID).Delete(&farmerentity.Farmer{})
		if result.Error != nil {
			return fmt.Errorf("failed to delete farmer: %w", result.Error)
		}
		report.FarmersDeleted = result.RowsAffected

		return nil
	})

	if err != nil {
		return report, err
	}

	log.Printf("Permanent delete completed for farmer %s by user %s: %+v", farmerID, deletedBy, report)
	return report, nil
}

// PermanentDeleteFarm permanently deletes a farm and all related data
func (s *PermanentDeleteService) PermanentDeleteFarm(ctx context.Context, farmID string, deletedBy string) (*DeleteReport, error) {
	report := &DeleteReport{
		EntityType: "farm",
		EntityID:   farmID,
		DeletedAt:  time.Now(),
		Errors:     []string{},
	}

	// Get farm first to check it exists
	var farmEntity farm.Farm
	if err := s.db.WithContext(ctx).Unscoped().First(&farmEntity, "id = ?", farmID).Error; err != nil {
		return nil, fmt.Errorf("farm not found: %w", err)
	}

	// Execute cascade delete in a transaction
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Delete farm activities (via crop cycles)
		result := tx.Unscoped().
			Where("crop_cycle_id IN (?)",
				tx.Model(&cropcycle.CropCycle{}).Select("id").Where("farm_id = ?", farmID)).
			Delete(&farmactivity.FarmActivity{})
		if result.Error != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("farm_activities: %v", result.Error))
		}
		report.ActivitiesDeleted = result.RowsAffected

		// 2. Delete crop cycles
		result = tx.Unscoped().Where("farm_id = ?", farmID).Delete(&cropcycle.CropCycle{})
		if result.Error != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("crop_cycles: %v", result.Error))
		}
		report.CropCyclesDeleted = result.RowsAffected

		// 3. Delete farm irrigation sources
		result = tx.Unscoped().Where("farm_id = ?", farmID).Delete(&farmirrigation.FarmIrrigationSource{})
		if result.Error != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("farm_irrigation_sources: %v", result.Error))
		}
		report.IrrigationDeleted = result.RowsAffected

		// 4. Delete farm soil types
		result = tx.Unscoped().Where("farm_id = ?", farmID).Delete(&farmsoil.FarmSoilType{})
		if result.Error != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("farm_soil_types: %v", result.Error))
		}
		report.SoilTypesDeleted = result.RowsAffected

		// 5. Delete farm
		result = tx.Unscoped().Where("id = ?", farmID).Delete(&farm.Farm{})
		if result.Error != nil {
			return fmt.Errorf("failed to delete farm: %w", result.Error)
		}
		report.FarmsDeleted = result.RowsAffected

		return nil
	})

	if err != nil {
		return report, err
	}

	log.Printf("Permanent delete completed for farm %s by user %s: %+v", farmID, deletedBy, report)
	return report, nil
}

// PermanentDeleteCropCycle permanently deletes a crop cycle and all related activities
func (s *PermanentDeleteService) PermanentDeleteCropCycle(ctx context.Context, cycleID string, deletedBy string) (*DeleteReport, error) {
	report := &DeleteReport{
		EntityType: "crop_cycle",
		EntityID:   cycleID,
		DeletedAt:  time.Now(),
		Errors:     []string{},
	}

	// Check crop cycle exists
	var cycle cropcycle.CropCycle
	if err := s.db.WithContext(ctx).Unscoped().First(&cycle, "id = ?", cycleID).Error; err != nil {
		return nil, fmt.Errorf("crop cycle not found: %w", err)
	}

	// Execute cascade delete in a transaction
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Delete farm activities
		result := tx.Unscoped().Where("crop_cycle_id = ?", cycleID).Delete(&farmactivity.FarmActivity{})
		if result.Error != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("farm_activities: %v", result.Error))
		}
		report.ActivitiesDeleted = result.RowsAffected

		// 2. Delete crop cycle
		result = tx.Unscoped().Where("id = ?", cycleID).Delete(&cropcycle.CropCycle{})
		if result.Error != nil {
			return fmt.Errorf("failed to delete crop cycle: %w", result.Error)
		}
		report.CropCyclesDeleted = result.RowsAffected

		return nil
	})

	if err != nil {
		return report, err
	}

	log.Printf("Permanent delete completed for crop_cycle %s by user %s: %+v", cycleID, deletedBy, report)
	return report, nil
}

// PermanentDeleteFarmerLink permanently deletes a farmer link
func (s *PermanentDeleteService) PermanentDeleteFarmerLink(ctx context.Context, linkID string, deletedBy string) (*DeleteReport, error) {
	report := &DeleteReport{
		EntityType: "farmer_link",
		EntityID:   linkID,
		DeletedAt:  time.Now(),
		Errors:     []string{},
	}

	result := s.db.WithContext(ctx).Unscoped().Where("id = ?", linkID).Delete(&farmerentity.FarmerLink{})
	if result.Error != nil {
		return nil, fmt.Errorf("failed to delete farmer link: %w", result.Error)
	}
	report.LinksDeleted = result.RowsAffected

	log.Printf("Permanent delete completed for farmer_link %s by user %s", linkID, deletedBy)
	return report, nil
}

// PermanentDeleteByOrg permanently deletes all data for an organization (dangerous!)
func (s *PermanentDeleteService) PermanentDeleteByOrg(ctx context.Context, orgID string, deletedBy string, dryRun bool) (*DeleteReport, error) {
	report := &DeleteReport{
		EntityType: "organization",
		EntityID:   orgID,
		DeletedAt:  time.Now(),
		Errors:     []string{},
	}

	// Get all farmer IDs for this org via farmer_links
	var farmerIDs []string
	s.db.WithContext(ctx).Model(&farmerentity.FarmerLink{}).Unscoped().
		Where("aaa_org_id = ?", orgID).
		Distinct().
		Pluck("aaa_user_id", &farmerIDs)

	// Get farmer internal IDs
	var farmerInternalIDs []string
	if len(farmerIDs) > 0 {
		s.db.WithContext(ctx).Model(&farmerentity.Farmer{}).Unscoped().
			Where("aaa_user_id IN ?", farmerIDs).
			Pluck("id", &farmerInternalIDs)
	}

	// Get all farm IDs
	var farmIDs []string
	s.db.WithContext(ctx).Model(&farm.Farm{}).Unscoped().
		Where("aaa_org_id = ?", orgID).
		Pluck("id", &farmIDs)

	if dryRun {
		// Count what would be deleted
		s.db.WithContext(ctx).Model(&farmactivity.FarmActivity{}).Unscoped().
			Where("farmer_id IN ?", farmerInternalIDs).Count(&report.ActivitiesDeleted)
		s.db.WithContext(ctx).Model(&cropcycle.CropCycle{}).Unscoped().
			Where("farm_id IN ?", farmIDs).Count(&report.CropCyclesDeleted)
		s.db.WithContext(ctx).Model(&farm.Farm{}).Unscoped().
			Where("aaa_org_id = ?", orgID).Count(&report.FarmsDeleted)
		s.db.WithContext(ctx).Model(&farmerentity.FarmerLink{}).Unscoped().
			Where("aaa_org_id = ?", orgID).Count(&report.LinksDeleted)
		report.FarmersDeleted = int64(len(farmerInternalIDs))
		return report, nil
	}

	// Execute cascade delete in a transaction
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Delete farm activities
		if len(farmerInternalIDs) > 0 {
			result := tx.Unscoped().Where("farmer_id IN ?", farmerInternalIDs).Delete(&farmactivity.FarmActivity{})
			report.ActivitiesDeleted = result.RowsAffected
		}

		// 2. Delete crop cycles
		if len(farmIDs) > 0 {
			result := tx.Unscoped().Where("farm_id IN ?", farmIDs).Delete(&cropcycle.CropCycle{})
			report.CropCyclesDeleted = result.RowsAffected

			// 3. Delete farm irrigation sources
			result = tx.Unscoped().Where("farm_id IN ?", farmIDs).Delete(&farmirrigation.FarmIrrigationSource{})
			report.IrrigationDeleted = result.RowsAffected

			// 4. Delete farm soil types
			result = tx.Unscoped().Where("farm_id IN ?", farmIDs).Delete(&farmsoil.FarmSoilType{})
			report.SoilTypesDeleted = result.RowsAffected
		}

		// 5. Delete farms
		result := tx.Unscoped().Where("aaa_org_id = ?", orgID).Delete(&farm.Farm{})
		report.FarmsDeleted = result.RowsAffected

		// 6. Delete farmer links for this org
		result = tx.Unscoped().Where("aaa_org_id = ?", orgID).Delete(&farmerentity.FarmerLink{})
		report.LinksDeleted = result.RowsAffected

		// Note: We don't delete farmers here as they might be linked to other orgs
		// Farmers are only deleted when they have no more links

		return nil
	})

	if err != nil {
		return report, err
	}

	log.Printf("Permanent delete completed for org %s by user %s: %+v", orgID, deletedBy, report)
	return report, nil
}

// CleanupOrphanedRecords removes records that are orphaned (soft-deleted but referenced data is gone)
func (s *PermanentDeleteService) CleanupOrphanedRecords(ctx context.Context, dryRun bool) (*DeleteReport, error) {
	report := &DeleteReport{
		EntityType: "orphaned_records",
		EntityID:   "cleanup",
		DeletedAt:  time.Now(),
		Errors:     []string{},
	}

	if dryRun {
		// Count orphaned farm activities (crop_cycle doesn't exist)
		s.db.WithContext(ctx).Model(&farmactivity.FarmActivity{}).Unscoped().
			Where("deleted_at IS NOT NULL").
			Where("crop_cycle_id NOT IN (?)",
				s.db.Model(&cropcycle.CropCycle{}).Unscoped().Select("id")).
			Count(&report.ActivitiesDeleted)

		// Count orphaned crop cycles (farm doesn't exist)
		s.db.WithContext(ctx).Model(&cropcycle.CropCycle{}).Unscoped().
			Where("deleted_at IS NOT NULL").
			Where("farm_id NOT IN (?)",
				s.db.Model(&farm.Farm{}).Unscoped().Select("id")).
			Count(&report.CropCyclesDeleted)

		// Count orphaned farms (farmer doesn't exist)
		s.db.WithContext(ctx).Model(&farm.Farm{}).Unscoped().
			Where("deleted_at IS NOT NULL").
			Where("farmer_id NOT IN (?)",
				s.db.Model(&farmerentity.Farmer{}).Unscoped().Select("id")).
			Count(&report.FarmsDeleted)

		return report, nil
	}

	// Execute cleanup in a transaction
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete orphaned farm activities
		result := tx.Unscoped().
			Where("deleted_at IS NOT NULL").
			Where("crop_cycle_id NOT IN (?)",
				tx.Model(&cropcycle.CropCycle{}).Unscoped().Select("id")).
			Delete(&farmactivity.FarmActivity{})
		report.ActivitiesDeleted = result.RowsAffected

		// Delete orphaned crop cycles
		result = tx.Unscoped().
			Where("deleted_at IS NOT NULL").
			Where("farm_id NOT IN (?)",
				tx.Model(&farm.Farm{}).Unscoped().Select("id")).
			Delete(&cropcycle.CropCycle{})
		report.CropCyclesDeleted = result.RowsAffected

		// Delete orphaned farms
		result = tx.Unscoped().
			Where("deleted_at IS NOT NULL").
			Where("farmer_id NOT IN (?)",
				tx.Model(&farmerentity.Farmer{}).Unscoped().Select("id")).
			Delete(&farm.Farm{})
		report.FarmsDeleted = result.RowsAffected

		return nil
	})

	if err != nil {
		return report, err
	}

	log.Printf("Orphaned records cleanup completed: %+v", report)
	return report, nil
}
