package handlers

import (
	"net/http"

	"github.com/Kisanlink/farmers-module/internal/interfaces"
	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AssignKisanSathi handles W4: Assign KisanSathi to farmer
// @Summary Assign KisanSathi to farmer
// @Description Assign a KisanSathi user to a specific farmer
// @Tags kisansathi
// @Accept json
// @Produce json
// @Param assignment body object true "KisanSathi assignment data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /kisansathi/assign [post]
func AssignKisanSathi(service services.KisanSathiService, logger interfaces.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			FarmerID         string `json:"farmer_id" binding:"required"`
			KisanSathiUserID string `json:"kisan_sathi_user_id" binding:"required"`
			FPOID            string `json:"fpo_id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Error("Failed to bind request", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		logger.Info("Assigning KisanSathi to farmer",
			zap.String("farmer_id", req.FarmerID),
			zap.String("kisan_sathi_user_id", req.KisanSathiUserID),
			zap.String("fpo_id", req.FPOID))

		// TODO: Implement the actual service call
		// err := service.AssignKisanSathi(c.Request.Context(), req)
		// if err != nil {
		//     c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		//     return
		// }

		logger.Info("KisanSathi assigned successfully",
			zap.String("farmer_id", req.FarmerID),
			zap.String("kisan_sathi_user_id", req.KisanSathiUserID),
			zap.String("fpo_id", req.FPOID))

		c.JSON(http.StatusOK, gin.H{
			"message": "KisanSathi assigned successfully",
			"data": gin.H{
				"farmer_id":           req.FarmerID,
				"kisan_sathi_user_id": req.KisanSathiUserID,
				"fpo_id":              req.FPOID,
			},
		})
	}
}

// ReassignOrRemoveKisanSathi handles W5: Reassign or remove KisanSathi
// @Summary Reassign or remove KisanSathi
// @Description Reassign a KisanSathi to a different farmer or remove the assignment
// @Tags kisansathi
// @Accept json
// @Produce json
// @Param assignment body object true "KisanSathi reassignment data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /kisansathi/reassign [post]
func ReassignOrRemoveKisanSathi(service services.KisanSathiService, logger interfaces.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			FarmerID         string  `json:"farmer_id" binding:"required"`
			KisanSathiUserID *string `json:"kisan_sathi_user_id,omitempty"`
			FPOID            string  `json:"fpo_id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Error("Failed to bind request", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		logger.Info("Reassigning or removing KisanSathi",
			zap.String("farmer_id", req.FarmerID),
			zap.String("fpo_id", req.FPOID))

		// TODO: Implement the actual service call
		// err := service.ReassignOrRemoveKisanSathi(c.Request.Context(), req)
		// if err != nil {
		//     c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		//     return
		// }

		action := "reassigned"
		if req.KisanSathiUserID == nil {
			action = "removed"
		}

		logger.Info("KisanSathi operation completed successfully",
			zap.String("action", action),
			zap.String("farmer_id", req.FarmerID),
			zap.String("fpo_id", req.FPOID))

		c.JSON(http.StatusOK, gin.H{
			"message": "KisanSathi " + action + " successfully",
			"data": gin.H{
				"farmer_id":           req.FarmerID,
				"kisan_sathi_user_id": req.KisanSathiUserID,
				"fpo_id":              req.FPOID,
			},
		})
	}
}

// GetKisanSathiAssignment handles getting KisanSathi assignment
// @Summary Get KisanSathi assignment
// @Description Retrieve the KisanSathi assignment for a specific farmer
// @Tags kisansathi
// @Accept json
// @Produce json
// @Param farmer_id path string true "Farmer ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /kisansathi/assignment/{farmer_id} [get]
func GetKisanSathiAssignment(service services.KisanSathiService, logger interfaces.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		farmerID := c.Param("farmer_id")

		if farmerID == "" {
			logger.Error("Missing farmer_id parameter")
			c.JSON(http.StatusBadRequest, gin.H{"error": "farmer_id is required"})
			return
		}

		logger.Info("Getting KisanSathi assignment",
			zap.String("farmer_id", farmerID))

		// TODO: Implement the actual service call
		// assignment, err := service.GetKisanSathiAssignment(c.Request.Context(), farmerID)
		// if err != nil {
		//     c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		//     return
		// }

		logger.Info("KisanSathi assignment retrieved successfully",
			zap.String("farmer_id", farmerID))

		c.JSON(http.StatusOK, gin.H{
			"message": "KisanSathi assignment retrieved successfully",
			"data": gin.H{
				"farmer_id":           farmerID,
				"kisan_sathi_user_id": "sample-kisan-sathi-id", // Placeholder
				"fpo_id":              "sample-fpo-id",         // Placeholder
				"status":              "ACTIVE",                // Placeholder
			},
		})
	}
}
