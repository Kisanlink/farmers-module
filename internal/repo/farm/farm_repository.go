package farm

import (
	"github.com/Kisanlink/farmers-module/internal/entities/farm"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// NewFarmRepository creates a new farm repository using BaseFilterableRepository
func NewFarmRepository(dbManager interface{}) *base.BaseFilterableRepository[*farm.Farm] {
	repo := base.NewBaseFilterableRepository[*farm.Farm]()
	repo.SetDBManager(dbManager)
	return repo
}
