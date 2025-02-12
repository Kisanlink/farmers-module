package controllers

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Kisanlink/farmers-module/repositories"
	"github.com/gin-gonic/gin"
)

type FarmController struct {
	FarmRepository *repositories.FarmRepository
}

func NewFarmController(farmRepo *repositories.FarmRepository) *FarmController {
	return &FarmController{
		FarmRepository: farmRepo,
	}
}

// GetFarmsByFarmerID retrieves farms by farmerID and allows filtering fields via query parameters
func (fc *FarmController) GetFarmsByFarmerID(c *gin.Context) {
	farmerIDstr := c.Param("id")
	if farmerIDstr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Farmer ID is required"})
		return
	}

	farmerid, err := strconv.ParseInt(farmerIDstr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid farmer ID"})
		return
	}

	// Extract the "fields" query parameter to decide which fields to return
	fields := c.DefaultQuery("fields", "farmID, acres, harvestDate, crop, cropImage, cropCategory, cropVariety, locality")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	farms, err := fc.FarmRepository.GetFarms(ctx, farmerid, fields)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if len(farms) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No farms found for this farmer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": farms,
		"success": true,
		"message": "Farms fetched successfully for the farmerId",
		"statusCode": 200,
	})
}

// GetFarmByFarmID retrieves a farm by its farmID
func (fc *FarmController) GetFarmByFarmID(c *gin.Context) {
	farmIDstr := c.Param("id")
	if farmIDstr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Farm ID is required"})
		return
	}

	farmID, err := strconv.ParseInt(farmIDstr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid farm ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	farm, err := fc.FarmRepository.GetFarmByID(ctx, farmID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, farm)
}

// Helper function to check if a field exists in the fields string
func contains(fields, field string) bool {
	return strings.Contains(fields, field)
}
