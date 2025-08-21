package handlers

import (
	"net/http"
	"strconv"

	farmerReq "github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/gin-gonic/gin"
)

// FarmerHandler handles HTTP requests for farmer operations
type FarmerHandler struct {
	farmerService services.FarmerService
}

// NewFarmerHandler creates a new farmer handler
func NewFarmerHandler(farmerService services.FarmerService) *FarmerHandler {
	return &FarmerHandler{
		farmerService: farmerService,
	}
}

// CreateFarmer handles POST /api/v1/identity/farmers
// @Summary Create a new farmer
// @Description Create a new farmer profile
// @Tags identity
// @Accept json
// @Produce json
// @Param farmer body farmerReq.CreateFarmerRequest true "Farmer data"
// @Success 201 {object} farmerResp.FarmerResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /identity/farmers [post]
func (h *FarmerHandler) CreateFarmer(c *gin.Context) {
	var req farmerReq.CreateFarmerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	response, err := h.farmerService.CreateFarmer(c.Request.Context(), &req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "farmer already exists" {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{
			"error":   "Failed to create farmer",
			"message": err.Error(),
		})
		return
	}

	// Set request ID if available
	if req.RequestID != "" {
		response.SetRequestID(req.RequestID)
	}

	c.JSON(http.StatusCreated, response)
}

// GetFarmer handles GET /api/v1/identity/farmers/:aaa_user_id/:aaa_org_id
// @Summary Get farmer by ID
// @Description Retrieve a farmer profile by AAA user ID and org ID
// @Tags identity
// @Accept json
// @Produce json
// @Param aaa_user_id path string true "AAA User ID"
// @Param aaa_org_id path string true "AAA Org ID"
// @Success 200 {object} farmerResp.FarmerProfileResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /identity/farmers/{aaa_user_id}/{aaa_org_id} [get]
func (h *FarmerHandler) GetFarmer(c *gin.Context) {
	aaaUserID := c.Param("aaa_user_id")
	aaaOrgID := c.Param("aaa_org_id")

	req := farmerReq.GetFarmerRequest{
		AAAUserID: aaaUserID,
		AAAOrgID:  aaaOrgID,
	}

	response, err := h.farmerService.GetFarmer(c.Request.Context(), &req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "farmer not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{
			"error":   "Failed to get farmer",
			"message": err.Error(),
		})
		return
	}

	// Set request ID if available
	if req.RequestID != "" {
		response.SetRequestID(req.RequestID)
	}

	c.JSON(http.StatusOK, response)
}

// ListFarmers handles GET /api/v1/identity/farmers
// @Summary List farmers
// @Description List farmers with filtering and pagination
// @Tags identity
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param aaa_org_id query string false "AAA Org ID filter"
// @Param kisan_sathi_user_id query string false "KisanSathi User ID filter"
// @Success 200 {object} farmerResp.FarmerListResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /identity/farmers [get]
func (h *FarmerHandler) ListFarmers(c *gin.Context) {
	req := farmerReq.NewListFarmersRequest()

	// Parse query parameters
	if page := c.Query("page"); page != "" {
		if pageNum, err := strconv.Atoi(page); err == nil && pageNum > 0 {
			req.Page = pageNum
		}
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		if size, err := strconv.Atoi(pageSize); err == nil && size > 0 && size <= 100 {
			req.PageSize = size
		}
	}
	if orgID := c.Query("aaa_org_id"); orgID != "" {
		req.AAAOrgID = orgID
	}
	if kisanSathiID := c.Query("kisan_sathi_user_id"); kisanSathiID != "" {
		req.KisanSathiUserID = kisanSathiID
	}

	response, err := h.farmerService.ListFarmers(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list farmers",
			"message": err.Error(),
		})
		return
	}

	// Set request ID if available
	if req.RequestID != "" {
		response.SetRequestID(req.RequestID)
	}

	c.JSON(http.StatusOK, response)
}
