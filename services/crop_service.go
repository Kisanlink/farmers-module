package services

import (
	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/repositories"
)

type CropServiceInterface interface {
	CreateCrop(crop *models.Crop) error
	GetAllCrops(name string, page, pageSize int) ([]*models.Crop, int64, error) // Updated signature
	GetCropByID(id string) (*models.Crop, error)
	UpdateCrop(crop *models.Crop) error
	DeleteCrop(id string) error
}

type CropService struct {
	repo repositories.CropRepositoryInterface
}

// NewCropService creates a new instance of CropService
func NewCropService(repo repositories.CropRepositoryInterface) *CropService {
	return &CropService{repo: repo}
}

// CreateCrop handles crop creation business logic
func (s *CropService) CreateCrop(crop *models.Crop) error {
	// Add any business logic/validation here before creating
	return s.repo.CreateCrop(crop)
}

// GetAllCrops handles retrieving crops with optional filtering and pagination
func (s *CropService) GetAllCrops(name string, page, pageSize int) ([]*models.Crop, int64, error) {
	// Add any business logic here (e.g., validate pagination params)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	return s.repo.GetAllCrops(name, page, pageSize)
}

// GetCropByID handles retrieving a single crop by ID
func (s *CropService) GetCropByID(id string) (*models.Crop, error) {
	// Add any business logic here (e.g., validate ID format)
	return s.repo.GetCropByID(id)
}

// UpdateCrop handles crop update business logic
func (s *CropService) UpdateCrop(crop *models.Crop) error {
	// Add any business logic/validation here before updating
	return s.repo.UpdateCrop(crop)
}

// DeleteCrop handles crop deletion business logic
func (s *CropService) DeleteCrop(id string) error {
	// Add any business logic here (e.g., check if crop can be deleted)
	return s.repo.DeleteCrop(id)
}
