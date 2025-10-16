package requests

// SeedLookupsRequest represents a request to seed lookup data
type SeedLookupsRequest struct {
	BaseRequest
	SeedSoilTypes         bool `json:"seed_soil_types" example:"true"`         // Seed soil types
	SeedIrrigationSources bool `json:"seed_irrigation_sources" example:"true"` // Seed irrigation sources
	Force                 bool `json:"force" example:"false"`                  // Force reseed even if data exists
}
