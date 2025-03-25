package services

import (
	"context"
	"fmt"

	"github.com/Kisanlink/farmers-module/models"
	"github.com/Kisanlink/farmers-module/repositories"
	"github.com/Kisanlink/farmers-module/utils"
)

type FarmServiceInterface interface {
	CreateFarm(
		ctx context.Context,
		farmerID string,
		coordinates [][]float64,
		area float64,
		locality string,
		cropType string,
		isKisansathi bool,
	) (*models.Farm, error)
}

type FarmService struct {
	repo repositories.FarmRepositoryInterface
}

func NewFarmService(repo repositories.FarmRepositoryInterface) *FarmService {
	return &FarmService{repo: repo}
}

func (s *FarmService) CreateFarm(
	ctx context.Context,
	farmerID string,
	coordinates [][]float64,
	area float64,
	locality string,
	cropType string,
	isKisansathi bool,
) (*models.Farm, error) {
	
	// Convert coordinates to GeoJSON Polygon format
	geoJSON := map[string]interface{}{
		"type": "Polygon",
		"coordinates": [][][]float64{coordinates}, // Note the extra wrapping array for Polygon
	}

	// Check for overlapping farms
	overlap, err := s.repo.CheckFarmOverlap(ctx, geoJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to check farm overlap: %w", err)
	}
	if overlap {
		return nil, fmt.Errorf("farm location overlaps with existing farm")
	}

	// Create farm model
	farm := &models.Farm{
		Id     :      utils.Generate10DigitID(),
		FarmerId:     farmerID,
		Verified:     isKisansathi, // Auto-verified if created by Kisansathi
		IsOwner:      true,
		Area:         area,
		Locality:     locality,
		CurrentCycle: cropType,
		OwnerId:      farmerID,
	}

	// Store the farm in database
	err = s.repo.CreateFarmRecord(ctx, farm, geoJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to create farm record: %w", err)
	}

	return farm, nil
}