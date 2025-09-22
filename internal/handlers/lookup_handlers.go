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

// GetSoilTypes handles GET /api/v1/lookup/soil-types
func (h *LookupHandlers) GetSoilTypes(c *gin.Context) {
	soilTypes, err := h.lookupService.GetSoilTypes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    soilTypes,
		"success": true,
	})
}

// GetIrrigationSources handles GET /api/v1/lookup/irrigation-sources
func (h *LookupHandlers) GetIrrigationSources(c *gin.Context) {
	irrigationSources, err := h.lookupService.GetIrrigationSources(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    irrigationSources,
		"success": true,
	})
}
