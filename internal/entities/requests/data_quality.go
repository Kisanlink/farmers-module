package requests

// ValidateGeometryRequest represents a request to validate WKT geometry
type ValidateGeometryRequest struct {
	BaseRequest
	WKT         string `json:"wkt" validate:"required"`
	CheckBounds bool   `json:"check_bounds,omitempty"` // Whether to check if geometry is within India bounds
}

// ReconcileAAALinksRequest represents a request to reconcile AAA links
type ReconcileAAALinksRequest struct {
	BaseRequest
	DryRun bool `json:"dry_run,omitempty"` // If true, only report what would be fixed without making changes
}

// RebuildSpatialIndexesRequest represents a request to rebuild spatial indexes
type RebuildSpatialIndexesRequest struct {
	BaseRequest
}

// DetectFarmOverlapsRequest represents a request to detect farm boundary overlaps
type DetectFarmOverlapsRequest struct {
	BaseRequest
	MinOverlapAreaHa *float64 `json:"min_overlap_area_ha,omitempty"` // Minimum overlap area in hectares to report
	Limit            *int     `json:"limit,omitempty"`               // Maximum number of overlaps to return
}
