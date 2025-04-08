package services

import (
	"fmt"

	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/repositories"
)

type FarmServiceInterface interface {
	CreateFarm(
		farmerId string,
		location models.GeoJSONPolygon,
		area float64,
		locality string,
		pincode int,
		ownerId string,
	) (*models.Farm, error)
	GetAllFarms(farmerId, pincode, date, id string) ([]*models.Farm, error)
	GetFarmsWithFilters(farmerId, pincode string) ([]*models.Farm, error)
}

type FarmService struct {
	repo repositories.FarmRepositoryInterface
}

func NewFarmService(repo repositories.FarmRepositoryInterface) *FarmService {
	return &FarmService{repo: repo}
}

// services/farm_service.go

// services/farm_service.go
func (s *FarmService) CreateFarm(
	farmerId string,
	location models.GeoJSONPolygon,
	area float64,
	locality string,
	pincode int,
	ownerId string,
) (*models.Farm, error) {
	// Validate polygon
	if location.Type != "Polygon" {
		return nil, fmt.Errorf("invalid geometry type, expected Polygon")
	}

	// Check for overlapping farms
	overlap, err := s.repo.CheckFarmOverlap(location)
	if err != nil {
		return nil, fmt.Errorf("error checking farm overlap: %w", err)
	}
	if overlap {
		return nil, fmt.Errorf("farm location overlaps with existing farm")
	}

	farm := &models.Farm{
		FarmerId: farmerId,
		Location: location,
		Area:     area,
		Locality: locality,
		OwnerId:  ownerId,
		Pincode:  pincode,
		IsOwner:  ownerId != "", // true if owner_id provided, else false
	}

	err = s.repo.CreateFarmRecord(farm)
	if err != nil {
		return nil, fmt.Errorf("failed to create farm record: %w", err)
	}

	return farm, nil
}

func (s *FarmService) GetAllFarms(farmerId, pincode, date, id string) ([]*models.Farm, error) {
	farms, err := s.repo.GetAllFarms(farmerId, pincode, date, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get farms: %w", err)
	}
	return farms, nil
}

func (s *FarmService) GetFarmsWithFilters(farmerId, pincode string) ([]*models.Farm, error) {
	return s.repo.GetFarmsWithFilters(farmerId, pincode)
}
