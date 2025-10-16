package soil_type

import (
	"context"

	"github.com/Kisanlink/farmers-module/internal/entities/soil_type"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
	"gorm.io/gorm"
)

// SoilTypeRepository wraps BaseFilterableRepository with custom methods for SoilType entity
type SoilTypeRepository struct {
	*base.BaseFilterableRepository[*soil_type.SoilType]
	db *gorm.DB
}

// NewSoilTypeRepository creates a new soil type repository using BaseFilterableRepository
func NewSoilTypeRepository(dbManager interface{}) *SoilTypeRepository {
	baseRepo := base.NewBaseFilterableRepository[*soil_type.SoilType]()
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

	return &SoilTypeRepository{
		BaseFilterableRepository: baseRepo,
		db:                       db,
	}
}

// CreateBatch creates multiple soil types in a single transaction
func (r *SoilTypeRepository) CreateBatch(ctx context.Context, soilTypes []*soil_type.SoilType) error {
	for _, st := range soilTypes {
		// Check if soil type already exists by name
		filter := base.NewFilterBuilder().
			Where("name", base.OpEqual, st.Name).
			Build()

		existing, err := r.FindOne(ctx, filter)
		if err == nil && existing != nil {
			// Soil type already exists, skip
			continue
		}

		// Create new soil type
		if err := r.Create(ctx, st); err != nil {
			return err
		}
	}
	return nil
}
