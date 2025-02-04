package controllers

import (
	"context"
	"net/http"
	"time"
	"fmt"
	"strconv"
	"github.com/Kisanlink/farmers-module/repositories"
	"github.com/gin-gonic/gin"
)

type FarmController struct {
	FarmRepository         *repositories.FarmRepository
	CommodityPriceRepo     *repositories.CommodityPriceRepository
	SoilTestReportRepo     *repositories.SoilTestReportRepository
}

func NewFarmController(farmRepo *repositories.FarmRepository, commodityRepo *repositories.CommodityPriceRepository, soilTestRepo *repositories.SoilTestReportRepository) *FarmController {
	return &FarmController{
		FarmRepository:        farmRepo,
		CommodityPriceRepo:    commodityRepo,
		SoilTestReportRepo:    soilTestRepo,
	}
}

// GetFarmsByFarmerID retrieves farms by farmerID along with crop price and soil test reports
func (fc *FarmController) GetFarmsByFarmerID(c *gin.Context) {
	farmerIDstr := c.Param("farmerId")
	farmerid, err := strconv.ParseInt(farmerIDstr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid farmer ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	farms, err := fc.FarmRepository.GetFarms(ctx, farmerid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if len(farms) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No farms found for this farmer"})
		return
	}

	var farmResponses []map[string]interface{}

	for _, farm := range farms {
		commodityPrice, err := fc.CommodityPriceRepo.GetCommodityPriceByCropID(ctx, farm.Crop)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		price := 0.0
		if commodityPrice != nil {
			price = commodityPrice.Price
		}

		// Fetch soil test reports for each farm
		soilReports, err := fc.SoilTestReportRepo.GetSoilTestReports(ctx, farm.FarmID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var soilTestData []map[string]interface{}
		for _, report := range soilReports {
			soilTestData = append(soilTestData, map[string]interface{}{
				"reportDate": fmt.Sprintf("%04d-%02d-%02d", report.ReportDate.Year, report.ReportDate.Month, report.ReportDate.Day),
				"reports":    report.Reports,
			})
		}

		farmResponse := map[string]interface{}{
			"farmID":       farm.FarmID,
			"acres":        farm.Acres,
			"harvestDate": fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d.%s",
		                 farm.HarvestDate.Year, farm.HarvestDate.Month, farm.HarvestDate.Day,
		                 farm.HarvestDate.Hour, farm.HarvestDate.Minute, farm.HarvestDate.Second,
		                 farm.HarvestDate.FractionalSecond),
			"crop":         farm.Crop,
			"cropImage":    farm.CropImage,
			"price":        price,
			"soilTests":    soilTestData, // Adding soil test reports
		}

		farmResponses = append(farmResponses, farmResponse)
	}

	c.JSON(http.StatusOK, farmResponses)
}
