package handlers

import (
	"net/http"
	"time"

	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"

	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/Kisanlink/farmers-module/pkg/common"
	"github.com/gin-gonic/gin"
)

// ReportingHandlers handles HTTP requests for reporting and analytics operations
type ReportingHandlers struct {
	reportingService services.ReportingService
}

// NewReportingHandlers creates new reporting handlers
func NewReportingHandlers(reportingService services.ReportingService) *ReportingHandlers {
	return &ReportingHandlers{
		reportingService: reportingService,
	}
}

// ExportFarmerPortfolio aggregates farms, cycles, and activities data for a farmer
// @Summary Export farmer portfolio
// @Description Aggregates farms, cycles, and activities data for a farmer with proper scope validation
// @Tags Reporting
// @Accept json
// @Produce json
// @Param request body requests.ExportFarmerPortfolioRequest true "Export farmer portfolio request"
// @Success 200 {object} responses.SwaggerExportFarmerPortfolioResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 401 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 404 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Router /api/v1/reports/farmer-portfolio [post]
func (h *ReportingHandlers) ExportFarmerPortfolio(c *gin.Context) {
	var req requests.ExportFarmerPortfolioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
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

	// Validate required fields
	if req.FarmerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "farmer_id is required"})
		return
	}

	response, err := h.reportingService.ExportFarmerPortfolio(c.Request.Context(), &req)
	if err != nil {
		common.HandleServiceError(c, err)
		return
	}

	portfolioResponse, ok := response.(*responses.ExportFarmerPortfolioResponse)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid response type"})
		return
	}

	c.JSON(http.StatusOK, portfolioResponse)
}

// OrgDashboardCounters provides org-level KPIs including counts and areas by season/status
// @Summary Get organization dashboard counters
// @Description Provides org-level KPIs including counts and areas by season/status with proper scope validation
// @Tags Reporting
// @Accept json
// @Produce json
// @Param request body requests.OrgDashboardCountersRequest true "Organization dashboard counters request"
// @Success 200 {object} responses.SwaggerOrgDashboardCountersResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 401 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Router /api/v1/reports/org-dashboard [post]
func (h *ReportingHandlers) OrgDashboardCounters(c *gin.Context) {
	var req requests.OrgDashboardCountersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
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

	response, err := h.reportingService.OrgDashboardCounters(c.Request.Context(), &req)
	if err != nil {
		common.HandleServiceError(c, err)
		return
	}

	dashboardResponse, ok := response.(*responses.OrgDashboardCountersResponse)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid response type"})
		return
	}

	c.JSON(http.StatusOK, dashboardResponse)
}

// ExportFarmerPortfolioByID is a convenience endpoint that exports farmer portfolio by farmer ID from URL path
// @Summary Export farmer portfolio by ID
// @Description Exports farmer portfolio data for a specific farmer ID from URL path
// @Tags Reporting
// @Accept json
// @Produce json
// @Param farmer_id path string true "Farmer ID"
// @Param season query string false "Season filter" Enums(RABI, KHARIF, ZAID)
// @Param start_date query string false "Start date filter (RFC3339 format)"
// @Param end_date query string false "End date filter (RFC3339 format)"
// @Param format query string false "Export format" Enums(json, csv) default(json)
// @Success 200 {object} responses.SwaggerExportFarmerPortfolioResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 401 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 404 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Router /api/v1/reports/farmer-portfolio/{farmer_id} [get]
func (h *ReportingHandlers) ExportFarmerPortfolioByID(c *gin.Context) {
	farmerID := c.Param("farmer_id")
	if farmerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "farmer_id is required"})
		return
	}

	// Build request from query parameters
	req := requests.ExportFarmerPortfolioRequest{
		FarmerID: farmerID,
		Season:   c.Query("season"),
		Format:   c.DefaultQuery("format", "json"),
	}

	// Parse date filters if provided
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := parseRFC3339Date(startDateStr); err == nil {
			req.StartDate = &startDate
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format, use RFC3339"})
			return
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := parseRFC3339Date(endDateStr); err == nil {
			req.EndDate = &endDate
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format, use RFC3339"})
			return
		}
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

	response, err := h.reportingService.ExportFarmerPortfolio(c.Request.Context(), &req)
	if err != nil {
		common.HandleServiceError(c, err)
		return
	}

	portfolioResponse, ok := response.(*responses.ExportFarmerPortfolioResponse)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid response type"})
		return
	}

	c.JSON(http.StatusOK, portfolioResponse)
}

// OrgDashboardCountersByID is a convenience endpoint that gets org dashboard counters from URL path
// @Summary Get organization dashboard counters by ID
// @Description Gets organization dashboard counters for a specific org ID from URL path
// @Tags Reporting
// @Accept json
// @Produce json
// @Param org_id path string true "Organization ID"
// @Param season query string false "Season filter" Enums(RABI, KHARIF, ZAID)
// @Param start_date query string false "Start date filter (RFC3339 format)"
// @Param end_date query string false "End date filter (RFC3339 format)"
// @Success 200 {object} responses.SwaggerOrgDashboardCountersResponse
// @Failure 400 {object} responses.SwaggerErrorResponse
// @Failure 401 {object} responses.SwaggerErrorResponse
// @Failure 403 {object} responses.SwaggerErrorResponse
// @Failure 500 {object} responses.SwaggerErrorResponse
// @Router /api/v1/reports/org-dashboard/{org_id} [get]
func (h *ReportingHandlers) OrgDashboardCountersByID(c *gin.Context) {
	orgID := c.Param("org_id")
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "org_id is required"})
		return
	}

	// Build request from query parameters
	req := requests.OrgDashboardCountersRequest{
		Season: c.Query("season"),
	}

	// Parse date filters if provided
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := parseRFC3339Date(startDateStr); err == nil {
			req.StartDate = &startDate
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format, use RFC3339"})
			return
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := parseRFC3339Date(endDateStr); err == nil {
			req.EndDate = &endDate
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format, use RFC3339"})
			return
		}
	}

	// Set user context from middleware
	if userID, exists := c.Get("aaa_subject"); exists {
		req.UserID = userID.(string)
	}
	// Override org ID from path parameter
	req.OrgID = orgID

	// Set request ID for tracing
	if requestID, exists := c.Get("request_id"); exists {
		req.RequestID = requestID.(string)
	}

	response, err := h.reportingService.OrgDashboardCounters(c.Request.Context(), &req)
	if err != nil {
		common.HandleServiceError(c, err)
		return
	}

	dashboardResponse, ok := response.(*responses.OrgDashboardCountersResponse)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid response type"})
		return
	}

	c.JSON(http.StatusOK, dashboardResponse)
}

// parseRFC3339Date parses a date string in RFC3339 format
func parseRFC3339Date(dateStr string) (time.Time, error) {
	return time.Parse(time.RFC3339, dateStr)
}
