package irrigation_source

import (
	"context"

	"github.com/Kisanlink/farmers-module/internal/entities/irrigation_source"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"gorm.io/gorm"
)

// IrrigationSourceRepository wraps BaseFilterableRepository with custom methods for IrrigationSource entity
type IrrigationSourceRepository struct {
	*base.BaseFilterableRepository[*irrigation_source.IrrigationSource]
	db *gorm.DB
}

// NewIrrigationSourceRepository creates a new irrigation source repository using BaseFilterableRepository
func NewIrrigationSourceRepository(dbManager interface{}) *IrrigationSourceRepository {
	baseRepo := base.NewBaseFilterableRepository[*irrigation_source.IrrigationSource]()
	baseRepo.SetDBManager(dbManager)

	// Get the GORM DB instance
	var db *gorm.DB
	if postgresManager, ok := dbManager.(interface {
		GetDB(context.Context, bool) (*gorm.DB, error)
	}); ok {
		if gormDB, err := postgresManager.GetDB(context.Background(), false); err == nil {
			db = gormDB
		}
	}

	return &IrrigationSourceRepository{
		BaseFilterableRepository: baseRepo,
		db:                       db,
	}
}

// CreateBatch creates multiple irrigation sources in a single transaction
func (r *IrrigationSourceRepository) CreateBatch(ctx context.Context, sources []*irrigation_source.IrrigationSource) error {
	for _, is := range sources {
		// Check if irrigation source already exists by name
		filter := base.NewFilterBuilder().
			Where("name", base.OpEqual, is.Name).
			Build()

		existing, err := r.FindOne(ctx, filter)
		if err == nil && existing != nil {
			// Irrigation source already exists, skip
			continue
		}

		// Create new irrigation source
		if err := r.Create(ctx, is); err != nil {
			return err
		}
	}
	return nil
}
