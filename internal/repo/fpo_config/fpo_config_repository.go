package fpo_config

import (
	"github.com/Kisanlink/farmers-module/internal/entities/fpo_config"
	"github.com/Kisanlink/kisanlink-db/pkg/base"
)

// NewFPOConfigRepository creates a new FPO configuration repository using BaseFilterableRepository
func NewFPOConfigRepository(dbManager interface{}) *base.BaseFilterableRepository[*fpo_config.FPOConfig] {
	repo := base.NewBaseFilterableRepository[*fpo_config.FPOConfig]()
	repo.SetDBManager(dbManager)
	return repo
}
