package services

import (
	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/repositories"
)

type StageServiceInterface interface {
	CreateStage(stage *models.Stage) error
	GetAllStages() ([]models.Stage, error)
	GetStageByID(id string) (*models.Stage, error)
	UpdateStage(stage *models.Stage) error
	DeleteStage(id string) error
}

type StageService struct {
	repo repositories.StageRepositoryInterface
}

func NewStageService(repo repositories.StageRepositoryInterface) StageServiceInterface {
	return &StageService{repo: repo}
}

func (s *StageService) CreateStage(stage *models.Stage) error {
	return s.repo.CreateStage(stage)
}

func (s *StageService) GetAllStages() ([]models.Stage, error) {
	return s.repo.GetAllStages()
}

func (s *StageService) GetStageByID(id string) (*models.Stage, error) {
	return s.repo.GetStageByID(id)
}

func (s *StageService) UpdateStage(stage *models.Stage) error {
	return s.repo.UpdateStage(stage)
}

func (s *StageService) DeleteStage(id string) error {
	return s.repo.DeleteStage(id)
}
