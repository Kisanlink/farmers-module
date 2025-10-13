package crop

import (
	"strings"

	"github.com/Kisanlink/farmers-module/internal/entities/crop"
)

// containsSeason checks if a crop's seasons list contains the given season
func containsSeason(seasons []string, season string) bool {
	for _, s := range seasons {
		if strings.EqualFold(s, season) {
			return true
		}
	}
	return false
}

// filterBySeason filters crops by a single season
func filterBySeason(crops []*crop.Crop, season string) []*crop.Crop {
	filtered := make([]*crop.Crop, 0)
	for _, c := range crops {
		if containsSeason(c.Seasons, season) {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

// filterBySeasons filters crops that match ANY of the provided seasons
func filterBySeasons(crops []*crop.Crop, seasons []string) []*crop.Crop {
	filtered := make([]*crop.Crop, 0)
	for _, c := range crops {
		for _, season := range seasons {
			if containsSeason(c.Seasons, season) {
				filtered = append(filtered, c)
				break
			}
		}
	}
	return filtered
}

// filterBySearch filters crops by name or scientific name (case-insensitive)
func filterBySearch(crops []*crop.Crop, searchTerm string) []*crop.Crop {
	filtered := make([]*crop.Crop, 0)
	searchLower := strings.ToLower(searchTerm)

	for _, c := range crops {
		nameMatch := strings.Contains(strings.ToLower(c.Name), searchLower)
		sciNameMatch := c.ScientificName != nil && strings.Contains(strings.ToLower(*c.ScientificName), searchLower)

		if nameMatch || sciNameMatch {
			filtered = append(filtered, c)
		}
	}
	return filtered
}
