package services

import (
	"fmt"

	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/repositories"
	
)

type FarmServiceInterface interface {
	CreateFarm(
		farmerID string,
		location models.GeoJSONPolygon,  // Changed from [][]float64 to GeoJSONPolygon
		area float64,
		locality string,
		// cropType string,
		// isKisansathi bool,
	) (*models.Farm, error)
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
) (*models.Farm, error) {
    // Validate polygon
    if location.Type != "Polygon" {
        return nil, fmt.Errorf("invalid geometry type, expected Polygon")
    }
    
    if len(location.Coordinates) == 0 || len(location.Coordinates[0]) < 4 {
        return nil, fmt.Errorf("polygon must have at least 4 points")
    }

    // Auto-close polygon if needed
    ring := location.Coordinates[0]
    first, last := ring[0], ring[len(ring)-1]
    if first[0] != last[0] || first[1] != last[1] {
        location.Coordinates[0] = append(ring, ring[0])
    }

    farm := &models.Farm{
        FarmerId:    farmerID,
        Location:    location,
        Area:        area,
        Locality:    locality,
        CurrentCycle: "Wheat", // Default value
        OwnerId:     farmerID,
        KisansathiId: nil,    // Explicitly set to nil
    }

    err := s.repo.CreateFarmRecord(farm)
    if err != nil {
        return nil, fmt.Errorf("failed to create farm record: %w", err)
    }

    return farm, nil
}