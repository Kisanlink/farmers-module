package controllers

import (
	"context"
	"log"
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

// GetAllCommodityPrices retrieves prices for all crops
func (cpc *CommodityPriceController) GetAllCommodityPrices(c *gin.Context) {
	// Context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Fetch all commodity prices
	commodityPrices, err := cpc.CommodityPriceRepo.GetAllPrices(ctx)
	if err != nil {
		log.Printf("ERROR: Failed to fetch commodity prices: %v", err) // Improved error logging
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch commodity prices"})
		return
	}

	if len(commodityPrices) == 0 {
		log.Println("DEBUG: No commodity prices found in the database.") // Check if no data is found
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"prices": commodityPrices,
	})
}

// GetCommodityPriceByName retrieves the price for a specific crop by name
func (cpc *CommodityPriceController) GetCommodityPriceByName(c *gin.Context) {
	cropName := c.Param("cropname")

	// Context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Fetch commodity price by crop name
	commodityPrice, err := cpc.CommodityPriceRepo.GetPriceByName(ctx, cropName)
	if err != nil {
		log.Printf("ERROR: Failed to fetch commodity price for crop name %s: %v", cropName, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch commodity price"})
		return
	}

	if commodityPrice == nil {
		log.Printf("DEBUG: No commodity price found for crop name: %s", cropName)
		c.JSON(http.StatusNotFound, gin.H{"error": "Commodity price not found"})
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"price": commodityPrice,
	})
}

// GetCommodityPriceByID retrieves the price for a specific crop by ID
func (cpc *CommodityPriceController) GetCommodityPriceByID(c *gin.Context) {
	id := c.Param("id")

	// Context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Fetch commodity price by ID
	commodityPrice, err := cpc.CommodityPriceRepo.GetPriceByID(ctx, id)
	if err != nil {
		log.Printf("ERROR: Failed to fetch commodity price for ID %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch commodity price"})
		return
	}

	if commodityPrice == nil {
		log.Printf("DEBUG: No commodity price found for ID: %s", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "Commodity price not found"})
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"price": commodityPrice,
	})
}
