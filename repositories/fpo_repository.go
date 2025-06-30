package repositories

import (
	"github.com/Kisanlink/farmers-module/models"
	"gorm.io/gorm"
)

// ---- interface & struct ---------------------------------------------

type FPORepositoryInterface interface {
	Create(fpo *models.FPO) error
	Get(reg string) (*models.FPO, error)
	GetByCEO(ceoID string) (*models.FPO, error)
	List() ([]models.FPO, error)
	Update(fpo *models.FPO) error
	Delete(reg string) error
}

type FPORepository struct{ db *gorm.DB }

// ---- constructor -----------------------------------------------------

func NewFPORepository(db *gorm.DB) *FPORepository {
	return &FPORepository{db: db}
}

// ---- CRUD methods ----------------------------------------------------

func (r *FPORepository) Create(fpo *models.FPO) error {
	return r.db.Create(fpo).Error
}

func (r *FPORepository) Get(reg string) (*models.FPO, error) {
	var f models.FPO
	if err := r.db.First(&f, "fpo_reg_no = ?", reg).Error; err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *FPORepository) GetByCEO(ceoID string) (*models.FPO, error) {
	var f models.FPO
	if err := r.db.First(&f, "ceo_id = ?", ceoID).Error; err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *FPORepository) Update(fpo *models.FPO) error {
	return r.db.Save(fpo).Error
}

func (r *FPORepository) Delete(reg string) error {
	return r.db.Delete(&models.FPO{}, "fpo_reg_no = ?", reg).Error
}

func (r *FPORepository) List() ([]models.FPO, error) {
	var all []models.FPO
	return all, r.db.Find(&all).Error
}
