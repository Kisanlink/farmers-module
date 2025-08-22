package farm_activity

import (
	"github.com/Kisanlink/farmers-module/internal/entities/farm_activity"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// NewFarmActivityRepository creates a new farm activity repository using BaseFilterableRepository
func NewFarmActivityRepository(dbManager interface{}) *base.BaseFilterableRepository[*farm_activity.FarmActivity] {
	repo := base.NewBaseFilterableRepository[*farm_activity.FarmActivity]()
	repo.SetDBManager(dbManager)
	return repo
}
