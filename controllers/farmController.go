package controllers

import (
	"context"
	"net/http"
	"time"
	"strconv"
	"github.com/Kisanlink/farmers-module/repositories"
	"github.com/gin-gonic/gin"
)

type FarmController struct {
	FarmRepository         *repositories.FarmRepository
	CommodityPriceRepo     *repositories.CommodityPriceRepository
}

func NewFarmController(farmRepo *repositories.FarmRepository, commodityRepo *repositories.CommodityPriceRepository) *FarmController {
	return &FarmController{
		FarmRepository:        farmRepo,
		CommodityPriceRepo:    commodityRepo,
	}
}

// GetFarmsByFarmerID retrieves farms by farmerID along with crop price
func (fc *FarmController) GetFarmsByFarmerID(c *gin.Context) {
	// Parse farmer ID from URL parameter
	farmerIDstr := c.Param("farmerId")
	farmerid, err := strconv.ParseInt(farmerIDstr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid farmer ID"})
		return
	}
	// Context and timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get farms from repository (now expects multiple farms)
	farms, err := fc.FarmRepository.GetFarms(ctx, farmerid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err})
		return
	}

	// Check if no farms were found
	if len(farms) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No farms found for this farmer"})
		return
	}

	// Prepare response array to hold farm data along with crop price
	var farmResponses []map[string]interface{}

	// Iterate through all farms and add crop price for each
	for _, farm := range farms {
		// Get commodity price for each farm's crop
		commodityPrice, err := fc.CommodityPriceRepo.GetCommodityPriceByCropID(ctx, farm.Crop)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err})
			return
		}

		// Default price is 0 if no price is found
		price := 0.0
		if commodityPrice != nil {
			price = commodityPrice.Price
		}

		// Exclude certain fields dynamically and prepare the response map for each farm
		farmResponse := map[string]interface{}{
			"farmID":       farm.FarmID,
			"acres":        farm.Acres,
			"harvestDate":  farm.HarvestDate,
			"crop":         farm.Crop,
			"cropImage":    farm.CropImage,
			"price":        price, // Adding the price
		}

		// Add the farm response to the response array
		farmResponses = append(farmResponses, farmResponse)
	}

	// Return response with all farms and their prices
	c.JSON(http.StatusOK, farmResponses)
}

