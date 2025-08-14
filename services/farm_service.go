package services

import (
	"fmt"
	"math"

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
	GetAllFarms(farmerId, pincode, date, id, fpoRegNo string) ([]*models.Farm, error)
	GetFarmsWithFilters(farmerId, pincode string) ([]*models.Farm, error)

	GetFarmByID(farmId string) (*models.Farm, error)
	GetFarmCentroids() ([]*repositories.FarmCentroid, error)
	GetFarmHeatmap(radius float64) ([]*HeatmapPoint, error)
}

type FarmService struct {
	repo repositories.FarmRepositoryInterface
}

func NewFarmService(repo repositories.FarmRepositoryInterface) *FarmService {
	return &FarmService{repo: repo}
}

type HeatmapPoint struct {
	Center models.GeoJSONPoint `json:"center"`
	Weight int                 `json:"weight"` // Number of farms in this cluster
	Color  string              `json:"color"`  // Hex color code for visualization
}

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

func (s *FarmService) GetAllFarms(
	farmerId, pincode, date, id, fpoRegNo string,
) ([]*models.Farm, error) {

	farms, err := s.repo.GetAllFarms(farmerId, pincode, date, id, fpoRegNo)
	if err != nil {
		return nil, fmt.Errorf("failed to get farms: %w", err)
	}
	return farms, nil
}

func (s *FarmService) GetFarmsWithFilters(farmerId, pincode string) ([]*models.Farm, error) {
	return s.repo.GetFarmsWithFilters(farmerId, pincode)
}

func (s *FarmService) GetFarmByID(farmId string) (*models.Farm, error) {
	farm, err := s.repo.GetFarmByID(farmId)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve farm: %w", err)
	}
	return farm, nil
}

func (s *FarmService) GetFarmCentroids() ([]*repositories.FarmCentroid, error) {
	centroids, err := s.repo.GetAllFarmCentroids()
	if err != nil {
		// The error from the repository is already well-described.
		return nil, err
	}
	return centroids, nil
}

const earthRadiusKm = 6371 // Radius of the Earth in kilometers

// haversineDistance calculates the distance between two lat/lon points in kilometers.
func haversineDistance(p1, p2 models.GeoJSONPoint) float64 {
	lat1 := p1.Coordinates[1] * math.Pi / 180
	lon1 := p1.Coordinates[0] * math.Pi / 180
	lat2 := p2.Coordinates[1] * math.Pi / 180
	lon2 := p2.Coordinates[0] * math.Pi / 180

	dLon := lon2 - lon1
	dLat := lat2 - lat1

	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1)*math.Cos(lat2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}

// getColorForWeight generates a hex color from yellow to red based on weight.
func getColorForWeight(weight, maxWeight int) string {
	if maxWeight == 0 || weight == 0 {
		return "#FFFF00" // Yellow for single points
	}
	// Intensity from 0.0 (low) to 1.0 (high)
	intensity := float64(weight) / float64(maxWeight)

	// Keep Red at max (255), decrease Green from 255 to 0
	green := 255 - int(intensity*255)
	if green < 0 {
		green = 0
	}

	return fmt.Sprintf("#FF%02X00", green)
}

func (s *FarmService) GetFarmHeatmap(radius float64) ([]*HeatmapPoint, error) {
	// 1. Fetch all farm centroids
	centroids, err := s.repo.GetAllFarmCentroids()
	if err != nil {
		return nil, fmt.Errorf("failed to get farm centroids: %w", err)
	}

	if len(centroids) == 0 {
		return []*HeatmapPoint{}, nil
	}

	// 2. Cluster the centroids
	var clusters [][]*repositories.FarmCentroid
	processed := make(map[string]bool) // To track centroids that are already in a cluster

	for _, centroid := range centroids {
		if processed[centroid.FarmID] {
			continue
		}

		currentCluster := []*repositories.FarmCentroid{centroid}
		processed[centroid.FarmID] = true

		// Find other centroids within the radius of the current one
		for _, other := range centroids {
			if !processed[other.FarmID] {
				if haversineDistance(centroid.Centroid, other.Centroid) <= radius {
					currentCluster = append(currentCluster, other)
					processed[other.FarmID] = true
				}
			}
		}
		clusters = append(clusters, currentCluster)
	}

	// 3. Process clusters to create heatmap points
	var heatmapPoints []*HeatmapPoint
	maxWeight := 0
	// First pass: find the max weight for color scaling
	for _, cluster := range clusters {
		if len(cluster) > maxWeight {
			maxWeight = len(cluster)
		}
	}

	// Second pass: build the heatmap points
	for _, cluster := range clusters {
		weight := len(cluster)
		if weight == 0 {
			continue
		}

		// Calculate the center of the cluster (average lat/lon)
		var totalLat, totalLon float64
		for _, point := range cluster {
			totalLon += point.Centroid.Coordinates[0]
			totalLat += point.Centroid.Coordinates[1]
		}
		centerLon := totalLon / float64(weight)
		centerLat := totalLat / float64(weight)

		heatmapPoints = append(heatmapPoints, &HeatmapPoint{
			Center: models.GeoJSONPoint{
				Type:        "Point",
				Coordinates: []float64{centerLon, centerLat},
			},
			Weight: weight,
			Color:  getColorForWeight(weight, maxWeight),
		})
	}

	return heatmapPoints, nil
}
