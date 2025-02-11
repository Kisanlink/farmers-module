package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/Kisanlink/farmers-module/repositories"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CommodityPriceController struct {
	CommodityPriceRepo *repositories.CommodityPriceRepository
}

func NewCommodityPriceController(repo *repositories.CommodityPriceRepository) *CommodityPriceController {
	return &CommodityPriceController{
		CommodityPriceRepo: repo,
	}
}

// GetAllCommodityPrices retrieves prices for all crops
func (cpc *CommodityPriceController) GetAllCommodityPrices(c *gin.Context) {
	// Context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Fetch all commodity prices
	commodityPrices, err := cpc.CommodityPriceRepo.GetAllPrices(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch commodity prices"})
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"prices": commodityPrices,
	})
}

// GetCommodityPriceByID retrieves a crop price by cropId
func (cpc *CommodityPriceController) GetCommodityPriceByID(c *gin.Context) {
	// Extract cropId from request parameters
	cropIDStr := c.Param("id")
	cropID, err := primitive.ObjectIDFromHex(cropIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid crop ID"})
		return
	}

	// Context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Fetch commodity price by cropId
	commodityPrice, err := cpc.CommodityPriceRepo.GetPriceByCropID(ctx, cropID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch commodity price"})
		return
	}

	// Return response
	c.JSON(http.StatusOK, commodityPrice)
}
