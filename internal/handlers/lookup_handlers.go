package handlers

import (
	"net/http"

	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/gin-gonic/gin"
)

// LookupHandlers handles lookup-related HTTP requests
type LookupHandlers struct {
	lookupService services.LookupService
}

// NewLookupHandlers creates new lookup handlers
func NewLookupHandlers(lookupService services.LookupService) *LookupHandlers {
	return &LookupHandlers{
		lookupService: lookupService,
	}
}

// GetSoilTypes handles GET /api/v1/lookups/soil-types
// @Summary Get all soil types
// @Description Retrieve a list of all available soil types for farm management
// @Tags lookups
// @Accept json
// @Produce json
// @Success 200 {object} SoilTypesResponse
// @Failure 500 {object} ErrorResponse
// @Router /lookups/soil-types [get]
func (h *LookupHandlers) GetSoilTypes(c *gin.Context) {
	soilTypes, err := h.lookupService.GetSoilTypes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    soilTypes,
		"success": true,
		"message": "Soil types retrieved successfully",
	})
}

// GetIrrigationSources handles GET /api/v1/lookups/irrigation-sources
// @Summary Get all irrigation sources
// @Description Retrieve a list of all available irrigation sources for farm management
// @Tags lookups
// @Accept json
// @Produce json
// @Success 200 {object} IrrigationSourcesResponse
// @Failure 500 {object} ErrorResponse
// @Router /lookups/irrigation-sources [get]
func (h *LookupHandlers) GetIrrigationSources(c *gin.Context) {
	irrigationSources, err := h.lookupService.GetIrrigationSources(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    irrigationSources,
		"success": true,
		"message": "Irrigation sources retrieved successfully",
	})
}

// Response types for Swagger documentation

// SoilTypesResponse represents the soil types response
type SoilTypesResponse struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message" example:"Soil types retrieved successfully"`
	Data    interface{} `json:"data"`
}

// IrrigationSourcesResponse represents the irrigation sources response
type IrrigationSourcesResponse struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message" example:"Irrigation sources retrieved successfully"`
	Data    interface{} `json:"data"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error" example:"Failed to retrieve data"`
}
