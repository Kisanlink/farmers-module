package controllers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/Kisanlink/farmers-module/repositories"
	"github.com/gin-gonic/gin"
)

type CommodityPriceController struct {
	CommodityPriceRepo *repositories.CommodityPriceRepository
}

func NewCommodityPriceController(repo *repositories.CommodityPriceRepository) *CommodityPriceController {
	return &CommodityPriceController{
		CommodityPriceRepo: repo,
	}
}

// GetCommodityPricesByFarmer retrieves all crop prices for a given farmer
func (cpc *CommodityPriceController) GetCommodityPricesByFarmerID(c *gin.Context) {

	// Extract farmerID from request parameters
	farmerIDStr := c.Param("farmerId")
	farmerID, err := strconv.Atoi(farmerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid farmer ID"})
		return
	}

	// Context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Fetch crops grown by the farmer
	crops, err := cpc.CommodityPriceRepo.GetCropsByFarmerID(ctx, farmerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch crops"})
		return
	}

	// If no crops found
	if len(crops) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No crops found for this farmer"})
		return
	}

	// Fetch commodity prices for the crops
	commodityPrices, err := cpc.CommodityPriceRepo.GetPricesForCrops(ctx, crops)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch commodity prices"})
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"farmerId": farmerID,
		"prices":   commodityPrices,
	})
}
