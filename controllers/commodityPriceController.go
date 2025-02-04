package controllers

import (
	"context"
	"net/http"
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

// GetCommodityPrice retrieves the price for a given crop
func (cpc *CommodityPriceController) GetCommodityPrice(c *gin.Context) {
	// Extract crop name from request parameters
	crop := c.Param("crop")
	if crop == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Crop name is required"})
		return
	}

	// Context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Fetch price from repository
	commodityPrice, err := cpc.CommodityPriceRepo.GetCommodityPriceByCropID(ctx, crop)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch commodity price"})
		return
	}

	// If price not found
	if commodityPrice == nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Price not found for this crop"})
		return
	}

	// Return price
	c.JSON(http.StatusOK, gin.H{
		"crop":  crop,
		"price": commodityPrice.Price,
	})
}
