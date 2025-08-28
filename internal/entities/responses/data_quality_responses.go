package responses

// ValidateGeometryResponse represents the response from geometry validation
type ValidateGeometryResponse struct {
	BaseResponse
	WKT      string   `json:"wkt"`
	IsValid  bool     `json:"is_valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
	SRID     int      `json:"srid"`
	AreaHa   *float64 `json:"area_ha,omitempty"`
}

// ReconcileAAALinksResponse represents the response from AAA links reconciliation
type ReconcileAAALinksResponse struct {
	BaseResponse
	ProcessedLinks int      `json:"processed_links"`
	FixedLinks     int      `json:"fixed_links"`
	BrokenLinks    int      `json:"broken_links"`
	Errors         []string `json:"errors,omitempty"`
}

// RebuildSpatialIndexesResponse represents the response from spatial indexes rebuild
type RebuildSpatialIndexesResponse struct {
	BaseResponse
	RebuiltIndexes []string `json:"rebuilt_indexes"`
	Errors         []string `json:"errors,omitempty"`
}

// FarmOverlap represents a detected overlap between two farms
type FarmOverlap struct {
	Farm1ID                string  `json:"farm1_id"`
	Farm1Name              string  `json:"farm1_name"`
	Farm1FarmerID          string  `json:"farm1_farmer_id"`
	Farm2ID                string  `json:"farm2_id"`
	Farm2Name              string  `json:"farm2_name"`
	Farm2FarmerID          string  `json:"farm2_farmer_id"`
	OverlapAreaHa          float64 `json:"overlap_area_ha"`
	Farm1AreaHa            float64 `json:"farm1_area_ha"`
	Farm2AreaHa            float64 `json:"farm2_area_ha"`
	OverlapPercentageFarm1 float64 `json:"overlap_percentage_farm1"`
	OverlapPercentageFarm2 float64 `json:"overlap_percentage_farm2"`
}

// DetectFarmOverlapsResponse represents the response from farm overlap detection
type DetectFarmOverlapsResponse struct {
	BaseResponse
	Overlaps      []FarmOverlap `json:"overlaps"`
	TotalOverlaps int           `json:"total_overlaps"`
}
