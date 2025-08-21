package handlers

import (
	"net/http"

	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/gin-gonic/gin"
)

// AssignKisanSathi handles W4: Assign KisanSathi to farmer
func AssignKisanSathi(service services.KisanSathiService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			FarmerID         string `json:"farmer_id" binding:"required"`
			KisanSathiUserID string `json:"kisan_sathi_user_id" binding:"required"`
			FPOID            string `json:"fpo_id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// TODO: Implement the actual service call
		// err := service.AssignKisanSathi(c.Request.Context(), req)
		// if err != nil {
		//     c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		//     return
		// }

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
func ReassignOrRemoveKisanSathi(service services.KisanSathiService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			FarmerID         string  `json:"farmer_id" binding:"required"`
			KisanSathiUserID *string `json:"kisan_sathi_user_id,omitempty"`
			FPOID            string  `json:"fpo_id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

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
func GetKisanSathiAssignment(service services.KisanSathiService) gin.HandlerFunc {
	return func(c *gin.Context) {
		farmerID := c.Param("farmer_id")

		if farmerID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "farmer_id is required"})
			return
		}

		// TODO: Implement the actual service call
		// assignment, err := service.GetKisanSathiAssignment(c.Request.Context(), farmerID)
		// if err != nil {
		//     c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		//     return
		// }

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
