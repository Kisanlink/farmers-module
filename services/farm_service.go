package services

import (
	"fmt"

	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/repositories"
)

type FarmServiceInterface interface {
	CreateFarm(
		farmer_id string,
		location models.GeoJSONPolygon,
		area float64,
		locality string,
		pincode int,
		owner_id string,
	) (*models.Farm, error)
	GetAllFarms(farmer_id, pincode, date, id string) ([]*models.Farm, error)
	GetFarmsWithFilters(farmer_id, pincode string) ([]*models.Farm, error)
	GetFarmByID(farm_id string) (*models.Farm, error)
}

type FarmService struct {
	Repo repositories.FarmRepositoryInterface
}

func NewFarmService(repo repositories.FarmRepositoryInterface) *FarmService {
	return &FarmService{Repo: repo}
}

func (s *FarmService) CreateFarm(
	farmer_id string,
	location models.GeoJSONPolygon,
	area float64,
	locality string,
	pincode int,
	owner_id string,
) (*models.Farm, error) {
	// Validate polygon
	if location.Type != "Polygon" {
		return nil, fmt.Errorf("invalid geometry type, expected Polygon")
	}

	// Check for overlapping farms
	overlap, err := s.Repo.CheckFarmOverlap(location)
	if err != nil {
		return nil, fmt.Errorf("error checking farm overlap: %w", err)
	}
	if overlap {
		return nil, fmt.Errorf("farm location overlaps with existing farm")
	}

	farm := &models.Farm{
		FarmerId: farmer_id,
		Location: location,
		Area:     area,
		Locality: locality,
		OwnerId:  owner_id,
		Pincode:  pincode,
		IsOwner:  owner_id != "", // true if owner_id provided, else false
	}

	err = s.Repo.CreateFarmRecord(farm)
	if err != nil {
		return nil, fmt.Errorf("failed to create farm record: %w", err)
	}

	return farm, nil
}

func (s *FarmService) GetAllFarms(farmer_id, pincode, date, id string) ([]*models.Farm, error) {
	farms, err := s.Repo.GetAllFarms(farmer_id, pincode, date, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get farms: %w", err)
	}
	return farms, nil
}

func (s *FarmService) GetFarmsWithFilters(farmer_id, pincode string) ([]*models.Farm, error) {
	return s.Repo.GetFarmsWithFilters(farmer_id, pincode)
}

func (s *FarmService) GetFarmByID(farm_id string) (*models.Farm, error) {
	farm, err := s.Repo.GetFarmByID(farm_id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve farm: %w", err)
	}
	return farm, nil
}
