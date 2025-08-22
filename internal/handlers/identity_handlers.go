package handlers

import (
	"net/http"

	"github.com/Kisanlink/farmers-module/internal/entities/requests"
	"github.com/Kisanlink/farmers-module/internal/entities/responses"
	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/gin-gonic/gin"
)

// LinkFarmerToFPO handles W1: Link farmer to FPO
func LinkFarmerToFPO(service services.FarmerLinkageService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req requests.LinkFarmerRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request format",
				"details": err.Error(),
			})
			return
		}

		// Set request ID if not provided
		if req.RequestID == "" {
			req.RequestID = generateRequestID()
		}

		// Call the service
		err := service.LinkFarmerToFPO(c.Request.Context(), &req)
		if err != nil {
			statusCode := http.StatusInternalServerError
			if isValidationError(err) {
				statusCode = http.StatusBadRequest
			} else if isPermissionError(err) {
				statusCode = http.StatusForbidden
			} else if isNotFoundError(err) {
				statusCode = http.StatusNotFound
			}

			c.JSON(statusCode, gin.H{
				"error":          err.Error(),
				"request_id":     req.RequestID,
				"correlation_id": req.RequestID,
			})
			return
		}

		// Create proper response using response structure
		response := responses.NewFarmerLinkageResponse(&responses.FarmerLinkageData{
			AAAUserID: req.AAAUserID,
			AAAOrgID:  req.AAAOrgID,
			Status:    "ACTIVE",
		}, "Farmer linked to FPO successfully")
		response.SetRequestID(req.RequestID)

		c.JSON(http.StatusOK, response)
	}
}

// UnlinkFarmerFromFPO handles W2: Unlink farmer from FPO
func UnlinkFarmerFromFPO(service services.FarmerLinkageService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req requests.UnlinkFarmerRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request format",
				"details": err.Error(),
			})
			return
		}

		// Set request ID if not provided
		if req.RequestID == "" {
			req.RequestID = generateRequestID()
		}

		// Call the service
		err := service.UnlinkFarmerFromFPO(c.Request.Context(), &req)
		if err != nil {
			statusCode := http.StatusInternalServerError
			if isValidationError(err) {
				statusCode = http.StatusBadRequest
			} else if isPermissionError(err) {
				statusCode = http.StatusForbidden
			} else if isNotFoundError(err) {
				statusCode = http.StatusNotFound
			}

			c.JSON(statusCode, gin.H{
				"error":          err.Error(),
				"request_id":     req.RequestID,
				"correlation_id": req.RequestID,
			})
			return
		}

		// Create proper response using response structure
		response := responses.NewFarmerLinkageResponse(&responses.FarmerLinkageData{
			AAAUserID: req.AAAUserID,
			AAAOrgID:  req.AAAOrgID,
			Status:    "INACTIVE",
		}, "Farmer unlinked from FPO successfully")
		response.SetRequestID(req.RequestID)

		c.JSON(http.StatusOK, response)
	}
}

// GetFarmerLinkage handles getting farmer linkage status
func GetFarmerLinkage(service services.FarmerLinkageService) gin.HandlerFunc {
	return func(c *gin.Context) {
		farmerID := c.Param("farmer_id")
		orgID := c.Param("org_id")

		if farmerID == "" || orgID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "farmer_id and org_id are required"})
			return
		}

		// Implement the actual service call
		linkage, err := service.GetFarmerLinkage(c.Request.Context(), farmerID, orgID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Farmer linkage retrieved successfully",
			"data":    linkage,
		})
	}
}

// RegisterFPORef handles W3: Register FPO reference
func RegisterFPORef(service services.FPOService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req requests.RegisterFPORefRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Implement the actual service call
		_, err := service.RegisterFPORef(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Create proper response using response structure
		response := responses.NewFPORefResponse(&responses.FPORefData{
			AAAOrgID:       req.AAAOrgID,
			BusinessConfig: req.BusinessConfig,
			Status:         "ACTIVE",
		}, "FPO reference registered successfully")
		response.SetRequestID(req.RequestID)

		c.JSON(http.StatusOK, response)
	}
}

// GetFPORef handles getting FPO reference
func GetFPORef(service services.FPOService) gin.HandlerFunc {
	return func(c *gin.Context) {
		orgID := c.Param("org_id")

		if orgID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "org_id is required"})
			return
		}

		// TODO: Implement the actual service call
		// fpoRef, err := service.GetFPORef(c.Request.Context(), orgID)
		// if err != nil {
		//     c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		//     return
		// }

		c.JSON(http.StatusOK, gin.H{
			"message": "FPO reference retrieved successfully",
			"data": gin.H{
				"aaa_org_id":      orgID,
				"business_config": "Sample business config", // Placeholder
			},
		})
	}
}
