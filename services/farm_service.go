package services

import (
	"fmt"

	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/repositories"
	
)

type FarmServiceInterface interface {
	CreateFarm(
		farmerID string,
		location models.GeoJSONPolygon,  
		area float64,
		locality string,
        pincode int,        
		ownerID string,      
	) (*models.Farm, error)
     GetAllFarms() ([]*models.Farm, error)
    GetFarmByID(id string) (*models.Farm, error)
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
    farmerID string,
    location models.GeoJSONPolygon,
    area float64,
    locality string,
    pincode int,          
	ownerID string,      
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
        FarmerId:    farmerID,
        Location:    location,
        Area:        area,
        Locality:    locality,
        OwnerId:     ownerID,
        Pincode:     pincode,
        IsOwner:     ownerID != "", // true if owner_id provided, else false        
    }

    err = s.repo.CreateFarmRecord(farm)
    if err != nil {
        return nil, fmt.Errorf("failed to create farm record: %w", err)
    }

    return farm, nil
}

func (s *FarmService) GetAllFarms() ([]*models.Farm, error) {
    farms, err := s.repo.GetAllFarms()
    if err != nil {
        return nil, fmt.Errorf("failed to get farms: %w", err)
    }
    return farms, nil
}

func (s *FarmService) GetFarmByID(id string) (*models.Farm, error) {
    farm, err := s.repo.GetFarmByID(id)
    if err != nil {
        return nil, fmt.Errorf("failed to get farm by id: %w", err)
    }
    return farm, nil
}