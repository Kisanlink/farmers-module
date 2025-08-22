package crop_cycle

import (
	"github.com/Kisanlink/farmers-module/internal/entities/crop_cycle"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// NewRepository creates a new crop cycle repository using BaseFilterableRepository
func NewRepository(dbManager interface{}) *base.BaseFilterableRepository[*crop_cycle.CropCycle] {
	repo := base.NewBaseFilterableRepository[*crop_cycle.CropCycle]()
	repo.SetDBManager(dbManager)
	return repo
}
