package handlers

import (
	"net/http"

	"github.com/Kisanlink/farmers-module/internal/services"
	"github.com/gin-gonic/gin"
)

// CreateFarm handles W6: Create farm
func CreateFarm(service services.FarmService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			FarmerID string `json:"farmer_id" binding:"required"`
			FPOID    string `json:"fpo_id" binding:"required"`
			Name     string `json:"name" binding:"required"`
			Geometry string `json:"geometry" binding:"required"`
			Location string `json:"location" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// TODO: Implement the actual service call
		c.JSON(http.StatusOK, gin.H{
			"message": "Farm created successfully",
			"data": gin.H{
				"farmer_id": req.FarmerID,
				"fpo_id":    req.FPOID,
				"name":      req.Name,
				"geometry":  req.Geometry,
				"location":  req.Location,
			},
		})
	}
}

// UpdateFarm handles W7: Update farm
func UpdateFarm(service services.FarmService) gin.HandlerFunc {
	return func(c *gin.Context) {
		farmID := c.Param("farm_id")
		var req struct {
			Name     *string `json:"name,omitempty"`
			Geometry *string `json:"geometry,omitempty"`
			Location *string `json:"location,omitempty"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// TODO: Implement the actual service call
		c.JSON(http.StatusOK, gin.H{
			"message": "Farm updated successfully",
			"data": gin.H{
				"farm_id": farmID,
			},
		})
	}
}

// DeleteFarm handles W8: Delete farm
func DeleteFarm(service services.FarmService) gin.HandlerFunc {
	return func(c *gin.Context) {
		farmID := c.Param("farm_id")

		// TODO: Implement the actual service call
		c.JSON(http.StatusOK, gin.H{
			"message": "Farm deleted successfully",
			"data": gin.H{
				"farm_id": farmID,
			},
		})
	}
}

// ListFarms handles W9: List farms
func ListFarms(service services.FarmService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement the actual service call
		c.JSON(http.StatusOK, gin.H{
			"message": "Farms retrieved successfully",
			"data":    []interface{}{},
		})
	}
}

// GetFarm handles getting farm by ID
func GetFarm(service services.FarmService) gin.HandlerFunc {
	return func(c *gin.Context) {
		farmID := c.Param("farm_id")

		// TODO: Implement the actual service call
		c.JSON(http.StatusOK, gin.H{
			"message": "Farm retrieved successfully",
			"data": gin.H{
				"farm_id": farmID,
			},
		})
	}
}
