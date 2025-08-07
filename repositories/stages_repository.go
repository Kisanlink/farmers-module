package repositories

import (
	"time"

	"github.com/Kisanlink/farmers-module/models"
	"gorm.io/gorm"
)

type StageRepositoryInterface interface {
	CreateStage(stage *models.Stage) error
	GetAllStages() ([]models.Stage, error)
	GetStageByID(id string) (*models.Stage, error)
	UpdateStage(stage *models.Stage) error
	DeleteStage(id string) error
}

type StageRepository struct {
	db *gorm.DB
}

func NewStageRepository(db *gorm.DB) StageRepositoryInterface {
	return &StageRepository{db: db}
}

func (r *StageRepository) CreateStage(stage *models.Stage) error {
	// BeforeCreate hook handles ID and timestamps
	return r.db.Create(stage).Error
}

func (r *StageRepository) GetAllStages() ([]models.Stage, error) {
	var stages []models.Stage
	err := r.db.Find(&stages).Error
	return stages, err
}

func (r *StageRepository) GetStageByID(id string) (*models.Stage, error) {
	var stage models.Stage
	err := r.db.First(&stage, "id = ?", id).Error
	return &stage, err
}

func (r *StageRepository) UpdateStage(stage *models.Stage) error {
	stage.UpdatedAt = time.Now()
	return r.db.Save(stage).Error
}

func (r *StageRepository) DeleteStage(id string) error {
	// Note: ON DELETE CASCADE will handle deleting associations in crop_stages
	return r.db.Delete(&models.Stage{}, "id = ?", id).Error
}
