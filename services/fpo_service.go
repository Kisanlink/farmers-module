package services

import (
	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/repositories"
)

type FPOServiceInterface interface {
	Create(fpo *models.FPO) error
	Get(reg string) (*models.FPO, error)
	List() ([]models.FPO, error)
	Update(fpo *models.FPO) error
	Delete(reg string) error
}

type FPOService struct {
	repo repositories.FPORepositoryInterface
}

func NewFPOService(r repositories.FPORepositoryInterface) *FPOService {
	return &FPOService{r}
}
func (s *FPOService) Create(f *models.FPO) error          { return s.repo.Create(f) }
func (s *FPOService) Get(reg string) (*models.FPO, error) { return s.repo.Get(reg) }
func (s *FPOService) Update(f *models.FPO) error          { return s.repo.Update(f) }
func (s *FPOService) Delete(reg string) error             { return s.repo.Delete(reg) }
func (s *FPOService) List() ([]models.FPO, error) {
	return s.repo.List()
}
