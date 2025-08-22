package fpo_ref

import (
	"github.com/Kisanlink/farmers-module/internal/entities/fpo"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// NewFPORefRepository creates a new FPO reference repository using BaseFilterableRepository
func NewFPORefRepository(dbManager interface{}) *base.BaseFilterableRepository[*fpo.FPORef] {
	repo := base.NewBaseFilterableRepository[*fpo.FPORef]()
	repo.SetDBManager(dbManager)
	return repo
}
