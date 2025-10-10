package handlers

import (
	"net/http"

	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/gin-gonic/gin"
)

// ValidateGeometryResponse represents a simple geometry validation response
type ValidateGeometryResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	RequestID string      `json:"request_id"`
	Data      interface{} `json:"data"`
}

// ReconcileAAALinksResponse represents a simple AAA links reconciliation response
type ReconcileAAALinksResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	RequestID string      `json:"request_id"`
	Data      interface{} `json:"data"`
}

// RebuildSpatialIndexesResponse represents a simple spatial indexes rebuild response
type RebuildSpatialIndexesResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	RequestID string      `json:"request_id"`
	Data      interface{} `json:"data"`
}

// DetectFarmOverlapsResponse represents a simple farm overlaps detection response
type DetectFarmOverlapsResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	RequestID string      `json:"request_id"`
	Data      interface{} `json:"data"`
}

// DataQualityHandlers handles HTTP requests for data quality operations
type DataQualityHandlers struct {
	dataQualityService services.DataQualityService
}

// NewDataQualityHandlers creates new data quality handlers
func NewDataQualityHandlers(dataQualityService services.DataQualityService) *DataQualityHandlers {
	return &DataQualityHandlers{
		dataQualityService: dataQualityService,
	}
}

// ValidateGeometry validates WKT geometry with PostGIS validation and SRID checks
// @Summary Validate geometry
// @Description Validates WKT geometry using PostGIS with SRID enforcement and integrity checks
// @Tags Data Quality
// @Accept json
// @Produce json
// @Param request body requests.ValidateGeometryRequest true "Validate geometry request"
// @Success 200 {object} ValidateGeometryResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /data-quality/validate-geometry [post]
func (h *DataQualityHandlers) ValidateGeometry(c *gin.Context) {
	var req requests.ValidateGeometryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Set user context from middleware
	if userID, exists := c.Get("aaa_subject"); exists {
		req.UserID = userID.(string)
	}
	if orgID, exists := c.Get("aaa_org"); exists {
		req.OrgID = orgID.(string)
	}

	// Set request ID for tracing
	if requestID, exists := c.Get("request_id"); exists {
		req.RequestID = requestID.(string)
	}

	response, err := h.dataQualityService.ValidateGeometry(c.Request.Context(), &req)
	if err != nil {
		common.HandleServiceError(c, err)
		return
	}

	validateResponse, ok := response.(*responses.ValidateGeometryResponse)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid response type"})
		return
	}

	c.JSON(http.StatusOK, validateResponse)
}

// ReconcileAAALinks heals broken AAA references in farmer_links
// @Summary Reconcile AAA links
// @Description Heals broken AAA references in farmer_links by checking against AAA service
// @Tags Data Quality
// @Accept json
// @Produce json
// @Param request body requests.ReconcileAAALinksRequest true "Reconcile AAA links request"
// @Success 200 {object} ReconcileAAALinksResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /data-quality/reconcile-aaa-links [post]
func (h *DataQualityHandlers) ReconcileAAALinks(c *gin.Context) {
	var req requests.ReconcileAAALinksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Set user context from middleware
	if userID, exists := c.Get("aaa_subject"); exists {
		req.UserID = userID.(string)
	}
	if orgID, exists := c.Get("aaa_org"); exists {
		req.OrgID = orgID.(string)
	}

	// Set request ID for tracing
	if requestID, exists := c.Get("request_id"); exists {
		req.RequestID = requestID.(string)
	}

	response, err := h.dataQualityService.ReconcileAAALinks(c.Request.Context(), &req)
	if err != nil {
		common.HandleServiceError(c, err)
		return
	}

	reconcileResponse, ok := response.(*responses.ReconcileAAALinksResponse)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid response type"})
		return
	}

	c.JSON(http.StatusOK, reconcileResponse)
}

// RebuildSpatialIndexes rebuilds GIST indexes for database maintenance
// @Summary Rebuild spatial indexes
// @Description Rebuilds GIST indexes for spatial tables for database maintenance
// @Tags Data Quality
// @Accept json
// @Produce json
// @Param request body requests.RebuildSpatialIndexesRequest true "Rebuild spatial indexes request"
// @Success 200 {object} RebuildSpatialIndexesResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /data-quality/rebuild-spatial-indexes [post]
func (h *DataQualityHandlers) RebuildSpatialIndexes(c *gin.Context) {
	var req requests.RebuildSpatialIndexesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Set user context from middleware
	if userID, exists := c.Get("aaa_subject"); exists {
		req.UserID = userID.(string)
	}
	if orgID, exists := c.Get("aaa_org"); exists {
		req.OrgID = orgID.(string)
	}

	// Set request ID for tracing
	if requestID, exists := c.Get("request_id"); exists {
		req.RequestID = requestID.(string)
	}

	response, err := h.dataQualityService.RebuildSpatialIndexes(c.Request.Context(), &req)
	if err != nil {
		common.HandleServiceError(c, err)
		return
	}

	rebuildResponse, ok := response.(*responses.RebuildSpatialIndexesResponse)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid response type"})
		return
	}

	c.JSON(http.StatusOK, rebuildResponse)
}

// DetectFarmOverlaps detects spatial intersections between farm boundaries
// @Summary Detect farm overlaps
// @Description Detects spatial intersections between farm boundaries within an organization
// @Tags Data Quality
// @Accept json
// @Produce json
// @Param request body requests.DetectFarmOverlapsRequest true "Detect farm overlaps request"
// @Success 200 {object} DetectFarmOverlapsResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 403 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /data-quality/detect-farm-overlaps [post]
func (h *DataQualityHandlers) DetectFarmOverlaps(c *gin.Context) {
	var req requests.DetectFarmOverlapsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Set user context from middleware
	if userID, exists := c.Get("aaa_subject"); exists {
		req.UserID = userID.(string)
	}
	if orgID, exists := c.Get("aaa_org"); exists {
		req.OrgID = orgID.(string)
	}

	// Set request ID for tracing
	if requestID, exists := c.Get("request_id"); exists {
		req.RequestID = requestID.(string)
	}

	response, err := h.dataQualityService.DetectFarmOverlaps(c.Request.Context(), &req)
	if err != nil {
		common.HandleServiceError(c, err)
		return
	}

	detectResponse, ok := response.(*responses.DetectFarmOverlapsResponse)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid response type"})
		return
	}

	c.JSON(http.StatusOK, detectResponse)
}
